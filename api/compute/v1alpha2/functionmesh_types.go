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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apispec "github.com/streamnative/function-mesh/pkg/spec"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FunctionMeshSpec defines the desired state of FunctionMesh
type FunctionMeshSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Sources   []SourceSpec   `json:"sources,omitempty"`
	Sinks     []SinkSpec     `json:"sinks,omitempty"`
	Functions []FunctionSpec `json:"functions,omitempty"`
}

// FunctionMeshStatus defines the observed state of FunctionMesh
type FunctionMeshStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions         []metav1.Condition             `json:"conditions,omitempty"`
	SourceConditions   map[string]*ComponentCondition `json:"sourceConditions,omitempty"`
	SinkConditions     map[string]*ComponentCondition `json:"sinkConditions,omitempty"`
	FunctionConditions map[string]*ComponentCondition `json:"functionConditions,omitempty"`
}

type ComponentCondition struct {
	Status apispec.ResourceConditionType `json:"status"`
	Hash   *string                       `json:"hash"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// FunctionMesh is the Schema for the functionmeshes API
type FunctionMesh struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FunctionMeshSpec   `json:"spec,omitempty"`
	Status FunctionMeshStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FunctionMeshList contains a list of FunctionMesh
type FunctionMeshList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FunctionMesh `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FunctionMesh{}, &FunctionMeshList{})
}
