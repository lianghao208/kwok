/*
Copyright 2022 The Kubernetes Authors.

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

package runtime

import (
	"context"
	"fmt"
	"sort"
)

// BuildRuntime is a function to build a runtime
type BuildRuntime func(name, workdir string) (Runtime, error)

// DefaultRegistry is the default registry
var DefaultRegistry = NewRegistry()

// Registry is a registry of runtime
type Registry struct {
	items map[string]BuildRuntime
}

// NewRegistry create a new registry
func NewRegistry() *Registry {
	return &Registry{
		items: map[string]BuildRuntime{},
	}
}

// Register a runtime
func (r *Registry) Register(name string, buildRuntime BuildRuntime) {
	r.items[name] = buildRuntime
}

// Get a runtime
func (r *Registry) Get(name string) (BuildRuntime, bool) {
	buildRuntime, ok := r.items[name]
	return buildRuntime, ok
}

// Load a runtime
func (r *Registry) Load(ctx context.Context, name, workdir string) (Runtime, error) {
	cluster := NewCluster(name, workdir)
	config, err := cluster.Load(ctx)
	if err != nil {
		return nil, err
	}
	conf := &config.Options

	buildRuntime, ok := r.Get(conf.Runtime)
	if !ok {
		return nil, fmt.Errorf("not found runtime %q", conf.Runtime)
	}
	return buildRuntime(name, workdir)
}

// List all registered runtime
func (r *Registry) List() []string {
	items := make([]string, 0, len(r.items))
	for name := range r.items {
		items = append(items, name)
	}
	sort.Strings(items)
	return items
}
