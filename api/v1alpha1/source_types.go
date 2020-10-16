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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SourceSpec defines the desired state of Source
type SourceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Name         string               `json:"name,omitempty"`
	ClassName    string               `json:"className,omitempty"`
	Tenant       string               `json:"tentant,omitempty"`
	ClusterName  string               `json:"clusterName,omitempty"`
	SourceType   string               `json:"sourceType,omitempty"`
	SinkType     string               `json:"sinkType,omitempty"`
	Replicas     *int32               `json:"replicas,omitempty"`
	MaxReplicas  *int32               `json:"maxReplicas,omitempty"` // if provided, turn on autoscaling
	Sink         string               `json:"sink,omitempty"`
	SchemaType   string               `json:"schemaType,omitempty"`
	LogTopic     string               `json:"logTopic,omitempty"`
	SourceConfig map[string]string    `json:"sourceConfig,omitempty"`
	Resources    corev1.ResourceList  `json:"resources,omitempty"`
	SecretsMap   map[string]SecretRef `json:"secretsMap,omitempty"`

	Timeout                      int32  `json:"timeout,omitempty"`
	AutoAck                      *bool  `json:"autoAck,omitempty"`
	MaxMessageRetry              int32  `json:"maxMessageRetry,omitempty"`
	ProcessingGuarantee          string `json:"processingGuarantee,omitempty"`
	RetainOrdering               bool   `json:"retainOrdering,omitempty"`
	RetainKeyOrdering            bool   `json:"retainKeyOrdering,omitempty"`
	DeadLetterTopic              string `json:"deadLetterTopic,omitempty"`
	ForwardSourceMessageProperty *bool  `json:"forwardSourceMessageProperty,omitempty"`
	MaxPendingAsyncRequests      *int32 `json:"maxPendingAsyncRequests,omitempty"`

	Messaging `json:",inline"`
	Runtime   `json:",inline"`
}

// SourceStatus defines the observed state of Source
type SourceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions map[Component]ResourceCondition `json:"conditions"`
	Replicas   int32                           `json:"replicas"`
	Selector   string                          `json:"selector"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
//+kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector

// Source is the Schema for the sources API
type Source struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SourceSpec   `json:"spec,omitempty"`
	Status SourceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SourceList contains a list of Source
type SourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Source `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Source{}, &SourceList{})
}
