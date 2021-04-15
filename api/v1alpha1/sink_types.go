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

// SinkSpec defines the desired state of Topic
type SinkSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Name         string                      `json:"name,omitempty"`
	ClassName    string                      `json:"className,omitempty"`
	ClusterName  string                      `json:"clusterName,omitempty"`
	Tenant       string                      `json:"tenant,omitempty"`
	SourceType   string                      `json:"sourceType,omitempty"`
	SinkType     string                      `json:"sinkType,omitempty"`
	Replicas     *int32                      `json:"replicas,omitempty"`
	MaxReplicas  *int32                      `json:"maxReplicas,omitempty"` // if provided, turn on autoscaling
	Input        InputConf                   `json:"input,omitempty"`
	LogTopic     string                      `json:"logTopic,omitempty"`
	SinkConfig   map[string]string           `json:"sinkConfig,omitempty"`
	Resources    corev1.ResourceRequirements `json:"resources,omitempty"`
	SecretsMap   map[string]SecretRef        `json:"secretsMap,omitempty"`
	VolumeMounts []corev1.VolumeMount        `json:"volumeMounts,omitempty"`

	Timeout                      int32            `json:"timeout,omitempty"`
	NegativeAckRedeliveryDelayMs int32            `json:"negativeAckRedeliveryDelayMs,omitempty"`
	AutoAck                      *bool            `json:"autoAck,omitempty"`
	MaxMessageRetry              int32            `json:"maxMessageRetry,omitempty"`
	ProcessingGuarantee          ProcessGuarantee `json:"processingGuarantee,omitempty"`
	RetainOrdering               bool             `json:"retainOrdering,omitempty"`
	DeadLetterTopic              string           `json:"deadLetterTopic,omitempty"`

	RuntimeFlags         string            `json:"runtimeFlags,omitempty"`
	SubscriptionName     string            `json:"subscriptionName,omitempty"`
	CleanupSubscription  bool              `json:"cleanupSubscription,omitempty"`
	SubscriptionPosition SubscribePosition `json:"subscriptionPosition,omitempty"`

	Pod PodPolicy `json:"pod,omitempty"`

	Messaging `json:",inline"`
	Runtime   `json:",inline"`

	// Image is the container image used to run sink pods.
	// default is streamnative/pulsar-functions-java-runner
	Image string `json:"image,omitempty"`

	// Image pull policy, one of Always, Never, IfNotPresent, default to Always.
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
}

// SinkStatus defines the observed state of Topic
type SinkStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions map[Component]ResourceCondition `json:"conditions"`
	Replicas   int32                           `json:"replicas"`
	Selector   string                          `json:"selector"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
//+kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector

// Topic is the Schema for the sinks API
type Sink struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SinkSpec   `json:"spec,omitempty"`
	Status SinkStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SinkList contains a list of Topic
type SinkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Sink `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Sink{}, &SinkList{})
}
