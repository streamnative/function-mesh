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

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BackendConfigSpec defines the desired state of BackendConfig
// +kubebuilder:validation:Optional
type BackendConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Name string `json:"name,omitempty"`

	// +kubebuilder:validation:Optional
	Env map[string]string `json:"env,omitempty"`

	// +kubebuilder:validation:Optional
	Pod *BackendConfigPodPolicy `json:"pod,omitempty"`

	// +kubebuilder:validation:Optional
	// indicate whether auto update functions&sinks&source when the BackendConfig is updated
	AutoUpdate bool `json:"autoUpdate,omitempty"`
}

// BackendConfigPodPolicy defines the policy for the pod
type BackendConfigPodPolicy struct {
	// Labels specify the labels to attach to pod the operator creates for the cluster.
	Labels map[string]string `json:"labels,omitempty"`

	// NodeSelector specifies a map of key-value pairs. For a pod to be eligible to run
	// on a node, the node must have each of the indicated key-value pairs as labels.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Affinity specifies the scheduling constraints of a pod
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// Tolerations specifies the tolerations of a Pod
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`

	// Annotations specifies the annotations to attach to pods the operator creates
	Annotations map[string]string `json:"annotations,omitempty"`

	// SecurityContext specifies the security context for the entire pod
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/security-context
	SecurityContext *v1.PodSecurityContext `json:"securityContext,omitempty"`

	// TerminationGracePeriodSeconds is the amount of time that kubernetes will give
	// for a pod before terminating it.
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty"`

	// ImagePullSecrets is an optional list of references to secrets in the same
	// namespace to use for pulling any of the images used by this PodSpec.
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// ServiceAccountName is the name of the ServiceAccount to use to run this pod.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// +kubebuilder:validation:Optional
	Liveness *Liveness `json:"liveness,omitempty"`
}

// BackendConfigStatus defines the observed state of BackendConfig
type BackendConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector

// BackendConfig is the Schema of the global configs for all functions, sinks and sources
// +kubebuilder:pruning:PreserveUnknownFields
type BackendConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackendConfigSpec   `json:"spec,omitempty"`
	Status BackendConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BackendConfigList contains a list of BackendConfig
type BackendConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackendConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BackendConfig{}, &BackendConfigList{})
}
