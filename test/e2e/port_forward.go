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

package e2e

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"sigs.k8s.io/kwok/pkg/utils/exec"
	"sigs.k8s.io/kwok/test/e2e/helper"
)

// CasePortForward creates a feature that tests port forward
func CasePortForward(kwokctlPath, clusterName, nodeName, namespace string) *features.FeatureBuilder {
	node := helper.NewNodeBuilder(nodeName).
		Build()
	pod0 := helper.NewPodBuilder("pod0").
		WithNamespace(namespace).
		WithNodeName(nodeName).
		Build()

	return features.New("Pod Port Forward").
		Setup(helper.CreateNode(node)).
		Setup(helper.CreatePod(pod0)).
		Teardown(helper.DeletePod(pod0)).
		Teardown(helper.DeleteNode(node)).
		Assess("test port forward", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Log("test port forward")

			cmd, err := exec.Command(exec.WithFork(ctx, true), kwokctlPath, "--name", clusterName, "kubectl", "port-forward", "-n", namespace, "pod0", "8080:8080")
			if err != nil {
				t.Fatal(err)
			}

			defer func() {
				err = cmd.Process.Kill()
				if err != nil {
					t.Fatal(err)
				}
			}()

			var resp *http.Response
			for i := 0; i != 30; i++ {
				resp, err = http.Get("http://localhost:8080/healthz")
				if err == nil {
					break
				}
				time.Sleep(1 * time.Second)
			}
			if err != nil {
				t.Fatal(err)
			}

			defer func() {
				_ = resp.Body.Close()
			}()

			if resp.StatusCode != http.StatusOK {
				t.Fatal("port forward failed", resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			if string(body) != "ok" {
				t.Fatal("port forward failed", string(body))
			}

			return ctx
		})
}
