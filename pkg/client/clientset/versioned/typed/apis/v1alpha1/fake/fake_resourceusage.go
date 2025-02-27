/*
Copyright The Kubernetes Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	gentype "k8s.io/client-go/gentype"
	v1alpha1 "sigs.k8s.io/kwok/pkg/apis/v1alpha1"
	apisv1alpha1 "sigs.k8s.io/kwok/pkg/client/clientset/versioned/typed/apis/v1alpha1"
)

// fakeResourceUsages implements ResourceUsageInterface
type fakeResourceUsages struct {
	*gentype.FakeClientWithList[*v1alpha1.ResourceUsage, *v1alpha1.ResourceUsageList]
	Fake *FakeKwokV1alpha1
}

func newFakeResourceUsages(fake *FakeKwokV1alpha1, namespace string) apisv1alpha1.ResourceUsageInterface {
	return &fakeResourceUsages{
		gentype.NewFakeClientWithList[*v1alpha1.ResourceUsage, *v1alpha1.ResourceUsageList](
			fake.Fake,
			namespace,
			v1alpha1.SchemeGroupVersion.WithResource("resourceusages"),
			v1alpha1.SchemeGroupVersion.WithKind("ResourceUsage"),
			func() *v1alpha1.ResourceUsage { return &v1alpha1.ResourceUsage{} },
			func() *v1alpha1.ResourceUsageList { return &v1alpha1.ResourceUsageList{} },
			func(dst, src *v1alpha1.ResourceUsageList) { dst.ListMeta = src.ListMeta },
			func(list *v1alpha1.ResourceUsageList) []*v1alpha1.ResourceUsage {
				return gentype.ToPointerSlice(list.Items)
			},
			func(list *v1alpha1.ResourceUsageList, items []*v1alpha1.ResourceUsage) {
				list.Items = gentype.FromPointerSlice(items)
			},
		),
		fake,
	}
}
