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

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	v1alpha1 "sigs.k8s.io/kwok/pkg/apis/v1alpha1"
	scheme "sigs.k8s.io/kwok/pkg/client/clientset/versioned/scheme"
)

// ExecsGetter has a method to return a ExecInterface.
// A group's client should implement this interface.
type ExecsGetter interface {
	Execs(namespace string) ExecInterface
}

// ExecInterface has methods to work with Exec resources.
type ExecInterface interface {
	Create(ctx context.Context, exec *v1alpha1.Exec, opts v1.CreateOptions) (*v1alpha1.Exec, error)
	Update(ctx context.Context, exec *v1alpha1.Exec, opts v1.UpdateOptions) (*v1alpha1.Exec, error)
	UpdateStatus(ctx context.Context, exec *v1alpha1.Exec, opts v1.UpdateOptions) (*v1alpha1.Exec, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Exec, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.ExecList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Exec, err error)
	ExecExpansion
}

// execs implements ExecInterface
type execs struct {
	client rest.Interface
	ns     string
}

// newExecs returns a Execs
func newExecs(c *KwokV1alpha1Client, namespace string) *execs {
	return &execs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the exec, and returns the corresponding exec object, and an error if there is any.
func (c *execs) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Exec, err error) {
	result = &v1alpha1.Exec{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("execs").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Execs that match those selectors.
func (c *execs) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ExecList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.ExecList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("execs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested execs.
func (c *execs) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("execs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a exec and creates it.  Returns the server's representation of the exec, and an error, if there is any.
func (c *execs) Create(ctx context.Context, exec *v1alpha1.Exec, opts v1.CreateOptions) (result *v1alpha1.Exec, err error) {
	result = &v1alpha1.Exec{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("execs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(exec).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a exec and updates it. Returns the server's representation of the exec, and an error, if there is any.
func (c *execs) Update(ctx context.Context, exec *v1alpha1.Exec, opts v1.UpdateOptions) (result *v1alpha1.Exec, err error) {
	result = &v1alpha1.Exec{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("execs").
		Name(exec.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(exec).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *execs) UpdateStatus(ctx context.Context, exec *v1alpha1.Exec, opts v1.UpdateOptions) (result *v1alpha1.Exec, err error) {
	result = &v1alpha1.Exec{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("execs").
		Name(exec.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(exec).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the exec and deletes it. Returns an error if one occurs.
func (c *execs) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("execs").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *execs) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("execs").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched exec.
func (c *execs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Exec, err error) {
	result = &v1alpha1.Exec{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("execs").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
