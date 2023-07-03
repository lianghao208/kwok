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

package snapshot

import (
	"context"
	"fmt"
	"io"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/net"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/pager"
	"k8s.io/client-go/util/retry"

	"sigs.k8s.io/kwok/pkg/log"
	"sigs.k8s.io/kwok/pkg/utils/client"
	"sigs.k8s.io/kwok/pkg/utils/yaml"
)

// PagerConfig is the configuration of the list pager.
// It defines the page size and the page buffer size of the list pager.
type PagerConfig struct {
	PageSize       int64
	PageBufferSize int32
}

// SaveConfig is the a combination of the impersonation config
// and the PagerConfig.
type SaveConfig struct {
	PagerConfig         *PagerConfig
	ImpersonationConfig rest.ImpersonationConfig
}

// Save saves the snapshot of cluster
func Save(ctx context.Context, kubeconfigPath string, w io.Writer, resources []string, saveConfig SaveConfig) error {
	clientset, err := client.NewClientset("", kubeconfigPath,
		client.WithImpersonate(saveConfig.ImpersonationConfig),
	)
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}

	restMapper, err := clientset.ToRESTMapper()
	if err != nil {
		return fmt.Errorf("failed to get rest mapper: %w", err)
	}
	dynamicClient, err := clientset.ToDynamicClient()
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	logger := log.FromContext(ctx)

	gvrs := make([]schema.GroupVersionResource, 0, len(resources))
	for _, resource := range resources {
		mapping, err := mappingFor(restMapper, resource)
		if err != nil {
			logger.Warn("Failed to get mapping for resource", "resource", resource, "err", err)
			continue
		}
		gvrs = append(gvrs, mapping.Resource)
	}

	encoder := yaml.NewEncoder(w)
	totalCount := 0
	start := time.Now()
	for _, gvr := range gvrs {
		nri := dynamicClient.Resource(gvr)
		logger := logger.With("resource", gvr.Resource)

		start := time.Now()
		page := 0
		listPager := pager.New(func(ctx context.Context, opts metav1.ListOptions) (runtime.Object, error) {
			var list runtime.Object
			var err error
			page++
			logger := logger.With("page", page, "limit", opts.Limit)
			logger.Debug("Listing resource")
			err = retry.OnError(retry.DefaultBackoff, retriable, func() error {
				list, err = nri.List(ctx, opts)
				if err != nil {
					logger.Error("failed to list resource", err)
				}
				return err
			})
			return list, err
		})

		pagerConfig := saveConfig.PagerConfig

		if pagerConfig != nil {
			if pagerConfig.PageSize > 0 {
				listPager.PageSize = pagerConfig.PageSize
			}
			if pagerConfig.PageBufferSize > 0 {
				listPager.PageBufferSize = pagerConfig.PageBufferSize
			}
		}

		count := 0
		if err := listPager.EachListItem(ctx, metav1.ListOptions{}, func(obj runtime.Object) error {
			if o, ok := obj.(metav1.Object); ok {
				clearUnstructured(o)
			}
			count++
			return encoder.Encode(obj)
		}); err != nil {
			return fmt.Errorf("failed to list resource %q: %w", gvr.Resource, err)
		}

		logger.Debug("Listed resource",
			"count", count,
			"elapsed", time.Since(start),
		)
		totalCount += count
	}

	logger.Info("Saved resources",
		"count", totalCount,
		"elapsed", time.Since(start),
	)
	return nil
}

func retriable(err error) bool {
	return apierrors.IsInternalError(err) ||
		apierrors.IsServiceUnavailable(err) ||
		apierrors.IsTooManyRequests(err) ||
		apierrors.IsTimeout(err) ||
		apierrors.IsServerTimeout(err) ||
		net.IsConnectionRefused(err)
}
