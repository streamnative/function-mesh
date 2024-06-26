// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/streamnative/function-mesh/api/compute/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeFunctionMeshes implements FunctionMeshInterface
type FakeFunctionMeshes struct {
	Fake *FakeComputeV1alpha1
	ns   string
}

var functionmeshesResource = schema.GroupVersionResource{Group: "compute", Version: "v1alpha1", Resource: "functionmeshes"}

var functionmeshesKind = schema.GroupVersionKind{Group: "compute", Version: "v1alpha1", Kind: "FunctionMesh"}

// Get takes name of the functionMesh, and returns the corresponding functionMesh object, and an error if there is any.
func (c *FakeFunctionMeshes) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.FunctionMesh, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(functionmeshesResource, c.ns, name), &v1alpha1.FunctionMesh{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.FunctionMesh), err
}

// List takes label and field selectors, and returns the list of FunctionMeshes that match those selectors.
func (c *FakeFunctionMeshes) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.FunctionMeshList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(functionmeshesResource, functionmeshesKind, c.ns, opts), &v1alpha1.FunctionMeshList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.FunctionMeshList{ListMeta: obj.(*v1alpha1.FunctionMeshList).ListMeta}
	for _, item := range obj.(*v1alpha1.FunctionMeshList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested functionMeshes.
func (c *FakeFunctionMeshes) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(functionmeshesResource, c.ns, opts))

}

// Create takes the representation of a functionMesh and creates it.  Returns the server's representation of the functionMesh, and an error, if there is any.
func (c *FakeFunctionMeshes) Create(ctx context.Context, functionMesh *v1alpha1.FunctionMesh, opts v1.CreateOptions) (result *v1alpha1.FunctionMesh, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(functionmeshesResource, c.ns, functionMesh), &v1alpha1.FunctionMesh{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.FunctionMesh), err
}

// Update takes the representation of a functionMesh and updates it. Returns the server's representation of the functionMesh, and an error, if there is any.
func (c *FakeFunctionMeshes) Update(ctx context.Context, functionMesh *v1alpha1.FunctionMesh, opts v1.UpdateOptions) (result *v1alpha1.FunctionMesh, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(functionmeshesResource, c.ns, functionMesh), &v1alpha1.FunctionMesh{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.FunctionMesh), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeFunctionMeshes) UpdateStatus(ctx context.Context, functionMesh *v1alpha1.FunctionMesh, opts v1.UpdateOptions) (*v1alpha1.FunctionMesh, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(functionmeshesResource, "status", c.ns, functionMesh), &v1alpha1.FunctionMesh{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.FunctionMesh), err
}

// Delete takes name of the functionMesh and deletes it. Returns an error if one occurs.
func (c *FakeFunctionMeshes) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(functionmeshesResource, c.ns, name, opts), &v1alpha1.FunctionMesh{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeFunctionMeshes) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(functionmeshesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.FunctionMeshList{})
	return err
}

// Patch applies the patch and returns the patched functionMesh.
func (c *FakeFunctionMeshes) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.FunctionMesh, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(functionmeshesResource, c.ns, name, pt, data, subresources...), &v1alpha1.FunctionMesh{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.FunctionMesh), err
}
