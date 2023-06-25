/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package portforward

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/server/httplog"
	"k8s.io/apiserver/pkg/util/wsstream"

	"sigs.k8s.io/kwok/pkg/log"
)

const (
	dataChannel = iota
	errorChannel

	v4BinaryWebsocketProtocol = "v4." + wsstream.ChannelWebSocketProtocol
	v4Base64WebsocketProtocol = "v4." + wsstream.Base64ChannelWebSocketProtocol
)

// V4Options contains details about which streams are required for port
// forwarding.
// All fields included in V4Options need to be expressed explicitly in the
// CRI (k8s.io/cri-corev1/pkg/apis/{version}/corev1.proto) PortForwardRequest.
type V4Options struct {
	Ports []int32
}

// NewV4Options creates a new options from the Request.
func NewV4Options(req *http.Request) (*V4Options, error) {
	if !wsstream.IsWebSocketRequest(req) {
		return &V4Options{}, nil
	}

	portStrings := req.URL.Query()[corev1.PortHeader]
	if len(portStrings) == 0 {
		return nil, fmt.Errorf("query parameter %q is required", corev1.PortHeader)
	}

	ports := make([]int32, 0, len(portStrings))
	for _, portString := range portStrings {
		if len(portString) == 0 {
			return nil, fmt.Errorf("query parameter %q cannot be empty", corev1.PortHeader)
		}
		for _, p := range strings.Split(portString, ",") {
			port, err := strconv.ParseUint(p, 10, 16)
			if err != nil {
				return nil, fmt.Errorf("unable to parse %q as a port: %w", portString, err)
			}
			if port < 1 {
				return nil, fmt.Errorf("port %q must be > 0", portString)
			}
			ports = append(ports, int32(port))
		}
	}

	return &V4Options{
		Ports: ports,
	}, nil
}

// handleWebSocketStreams handles requests to forward ports to a pod via
// a PortForwarder. A pair of streams are created per port (DATA n,
// ERROR n+1). The associated port is written to each stream as a unsigned 16
// bit integer in little endian format.
func handleWebSocketStreams(ctx context.Context, req *http.Request, w http.ResponseWriter, portForwarder PortForwarder, podName, podNamespace string, uid types.UID, opts *V4Options, idleTimeout time.Duration) error {
	channels := make([]wsstream.ChannelType, 0, len(opts.Ports)*2)
	for i := 0; i < len(opts.Ports); i++ {
		channels = append(channels, wsstream.ReadWriteChannel, wsstream.WriteChannel)
	}
	conn := wsstream.NewConn(map[string]wsstream.ChannelProtocolConfig{
		"": {
			Binary:   true,
			Channels: channels,
		},
		v4BinaryWebsocketProtocol: {
			Binary:   true,
			Channels: channels,
		},
		v4Base64WebsocketProtocol: {
			Binary:   false,
			Channels: channels,
		},
	})
	conn.SetIdleTimeout(idleTimeout)
	_, streams, err := conn.Open(httplog.Unlogged(req, w), req)
	if err != nil {
		return fmt.Errorf("unable to upgrade websocket connection: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()
	streamPairs := make([]*websocketStreamPair, len(opts.Ports))
	for i := range streamPairs {
		streamPair := websocketStreamPair{
			port:        opts.Ports[i],
			dataStream:  streams[i*2+dataChannel],
			errorStream: streams[i*2+errorChannel],
		}
		streamPairs[i] = &streamPair

		portBytes := make([]byte, 2)
		// port is always positive so conversion is allowable
		binary.LittleEndian.PutUint16(portBytes, uint16(streamPair.port))
		_, _ = streamPair.dataStream.Write(portBytes)
		_, _ = streamPair.errorStream.Write(portBytes)
	}

	logger := log.FromContext(ctx)
	h := &websocketStreamHandler{
		logger:       logger,
		conn:         conn,
		streamPairs:  streamPairs,
		podName:      podName,
		podNamespace: podNamespace,
		uid:          uid,
		forwarder:    portForwarder,
	}
	h.run(ctx)

	return nil
}

// websocketStreamPair represents the error and data streams for a port
// forwarding request.
type websocketStreamPair struct {
	port        int32
	dataStream  io.ReadWriteCloser
	errorStream io.WriteCloser
}

// websocketStreamHandler is capable of processing a single port forward
// request over a websocket connection
type websocketStreamHandler struct {
	logger       *log.Logger
	conn         *wsstream.Conn
	streamPairs  []*websocketStreamPair
	podName      string
	podNamespace string
	uid          types.UID
	forwarder    PortForwarder
}

// run invokes the websocketStreamHandler's forwarder.PortForward
// function for the given stream pair.
func (h *websocketStreamHandler) run(ctx context.Context) {
	wg := sync.WaitGroup{}
	wg.Add(len(h.streamPairs))

	for _, pair := range h.streamPairs {
		p := pair
		go func() {
			defer wg.Done()
			h.portForward(ctx, p)
		}()
	}

	wg.Wait()
}

func (h *websocketStreamHandler) portForward(ctx context.Context, p *websocketStreamPair) {
	defer func() {
		_ = p.errorStream.Close()
		_ = p.dataStream.Close()
	}()

	h.logger.Debug("Connection invoking forwarder.PortForward for port", "connection", h.conn, "port", p.port)
	err := h.forwarder.PortForward(ctx, h.podName, h.podNamespace, h.uid, p.port, p.dataStream)
	h.logger.Debug("Connection done invoking forwarder.PortForward for port", "connection", h.conn, "port", p.port)

	if err != nil {
		err := fmt.Errorf("error forwarding port %d to pod %s/%s: %w", p.port, h.podNamespace, h.podName, err)
		logger := log.FromContext(ctx)
		logger.Error("PortForward", err)
		fmt.Fprint(p.errorStream, err.Error())
	}
}
