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

package spec

import (
	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func MakeFunctionComponent(functionName string, mesh *v1alpha1.FunctionMesh,
	spec *v1alpha1.FunctionSpec) *v1alpha1.Function {
	return &v1alpha1.Function{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "compute.functionmesh.io/v1alpha1",
			Kind:       "Function",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      functionName,
			Namespace: mesh.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mesh, mesh.GroupVersionKind()),
			},
		},
		Spec: *spec,
	}
}

func MakeSourceComponent(sourceName string, mesh *v1alpha1.FunctionMesh, spec *v1alpha1.SourceSpec) *v1alpha1.Source {
	return &v1alpha1.Source{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "compute.functionmesh.io/v1alpha1",
			Kind:       "Source",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      sourceName,
			Namespace: mesh.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mesh, mesh.GroupVersionKind()),
			},
		},
		Spec: *spec,
	}
}

func MakeSinkComponent(sinkName string, mesh *v1alpha1.FunctionMesh, spec *v1alpha1.SinkSpec) *v1alpha1.Sink {
	return &v1alpha1.Sink{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "compute.functionmesh.io/v1alpha1",
			Kind:       "Sink",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      sinkName,
			Namespace: mesh.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mesh, mesh.GroupVersionKind()),
			},
		},
		Spec: *spec,
	}
}

func MakeGenericResourceComponent(resourceName string, mesh *v1alpha1.FunctionMesh, spec *v1alpha1.GenericResourceSpec) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetAPIVersion(spec.APIVersion)
	obj.SetKind(spec.Kind)
	obj.SetName(resourceName)
	obj.SetNamespace(mesh.Namespace)

	specFieldName := "spec"
	if spec.SpecFieldName != "" {
		specFieldName = spec.SpecFieldName
	}
	obj.Object[specFieldName] = spec.Spec

	// Set owner reference for garbage collection
	obj.SetOwnerReferences([]metav1.OwnerReference{*metav1.NewControllerRef(mesh, mesh.GroupVersionKind())})
	return obj
}
