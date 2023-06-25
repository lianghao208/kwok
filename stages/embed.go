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

// Package stages contains node and pod stages used by controllers.
package stages

import (
	_ "embed"
)

var (
	// DefaultNodeStages is the default node stages.
	//go:embed node-fast.yaml
	DefaultNodeStages string

	// DefaultNodeHeartbeatStages is the default node heartbeat stages.
	//go:embed node-heartbeat.yaml
	DefaultNodeHeartbeatStages string

	// DefaultPodStages is the default pod stages.
	//go:embed pod-fast.yaml
	DefaultPodStages string
)
