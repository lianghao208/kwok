//go:build windows

/*
Copyright 2023 The Kubernetes Authors.

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

package server

import (
	"context"
	"io"

	clientremotecommand "k8s.io/client-go/tools/remotecommand"

	"sigs.k8s.io/kwok/pkg/log"
)

func (s *Server) execInContainerWithTTY(ctx context.Context, cmd []string, in io.Reader, out io.WriteCloser, resize <-chan clientremotecommand.TerminalSize) error {
	logger := log.FromContext(ctx)
	logger.Warn("execInContainerWithTTY is not supported on windows, fallback to execInContainer")
	return s.execInContainer(ctx, cmd, in, out, out)
}
