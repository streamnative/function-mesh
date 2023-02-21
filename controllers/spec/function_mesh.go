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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	computeapi "github.com/streamnative/function-mesh/api/compute/v1alpha2"
)

func MakeFunctionComponent(functionName string, mesh *computeapi.FunctionMesh,
	spec *computeapi.FunctionSpec) *computeapi.Function {
	return &computeapi.Function{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "compute.functionmesh.io/v1alpha2",
			Kind:       "Function",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      functionName,
			Namespace: mesh.Namespace,
			Labels:    makeMeshLabels(mesh, functionName),
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mesh, mesh.GroupVersionKind()),
			},
		},
		Spec: *spec,
	}
}

func MakeSourceComponent(sourceName string, mesh *computeapi.FunctionMesh, spec *computeapi.SourceSpec) *computeapi.Source {
	return &computeapi.Source{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "compute.functionmesh.io/v1alpha2",
			Kind:       "Source",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      sourceName,
			Namespace: mesh.Namespace,
			Labels:    makeMeshLabels(mesh, sourceName),
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mesh, mesh.GroupVersionKind()),
			},
		},
		Spec: *spec,
	}
}

func MakeSinkComponent(sinkName string, mesh *computeapi.FunctionMesh, spec *computeapi.SinkSpec) *computeapi.Sink {
	return &computeapi.Sink{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "compute.functionmesh.io/v1alpha2",
			Kind:       "Sink",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      sinkName,
			Namespace: mesh.Namespace,
			Labels:    makeMeshLabels(mesh, sinkName),
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mesh, mesh.GroupVersionKind()),
			},
		},
		Spec: *spec,
	}
}

func makeMeshLabels(mesh *computeapi.FunctionMesh, componentName string) map[string]string {
	labels := map[string]string{
		"app.kubernetes.io/name":            componentName,
		"app.kubernetes.io/instance":        componentName,
		"compute.functionmesh.io/app":       AppFunctionMesh,
		"compute.functionmesh.io/part-of":   mesh.Name,
		"compute.functionmesh.io/name":      componentName,
		"compute.functionmesh.io/namespace": mesh.Namespace,
	}
	return labels
}
