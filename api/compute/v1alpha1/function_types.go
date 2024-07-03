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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FunctionSpec defines the desired state of Function
// +kubebuilder:validation:Optional
type FunctionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Name        string `json:"name,omitempty"`
	ClassName   string `json:"className,omitempty"`
	Tenant      string `json:"tenant,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
	ClusterName string `json:"clusterName,omitempty"`
	// +kubebuilder:validation:Minimum=0
	Replicas *int32 `json:"replicas,omitempty"`
	// Whether show the precise parallelism, if true, the `Parallelism` will be equal to the `Replicas`,
	// in such case, update the `Replicas` will cause all pods being recreated since the command of pod is updated.
	// else, the `Parallelism` will be 1, default to false.
	// It just affects the result of context.getNumInstances, there will be only 1 process and 1 thread in each pod in any cases.
	ShowPreciseParallelism bool `json:"showPreciseParallelism,omitempty"`
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=1
	MinReplicas *int32 `json:"minReplicas,omitempty"`

	DownloaderImage string `json:"downloaderImage,omitempty"`
	// the image used to clean up subscription, if empty, the runner image will be used
	CleanupImage string `json:"cleanupImage,omitempty"`

	// MaxReplicas indicates the maximum number of replicas and enables the HorizontalPodAutoscaler
	// If provided, a default HPA with CPU at average of 80% will be used.
	// For complex HPA strategies, please refer to Pod.HPAutoscaler.
	MaxReplicas   *int32        `json:"maxReplicas,omitempty"` // if provided, turn on autoscaling
	Input         InputConf     `json:"input,omitempty"`
	Output        OutputConf    `json:"output,omitempty"`
	LogTopic      string        `json:"logTopic,omitempty"`
	LogTopicAgent LogTopicAgent `json:"logTopicAgent,omitempty"`
	FilebeatImage string        `json:"filebeatImage,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:pruning:PreserveUnknownFields
	FuncConfig   *Config                     `json:"funcConfig,omitempty"`
	Resources    corev1.ResourceRequirements `json:"resources,omitempty"`
	SecretsMap   map[string]SecretRef        `json:"secretsMap,omitempty"`
	VolumeMounts []corev1.VolumeMount        `json:"volumeMounts,omitempty"`

	Timeout                      int32            `json:"timeout,omitempty"`
	AutoAck                      *bool            `json:"autoAck,omitempty"`
	MaxMessageRetry              int32            `json:"maxMessageRetry,omitempty"`
	ProcessingGuarantee          ProcessGuarantee `json:"processingGuarantee,omitempty"`
	RetainOrdering               bool             `json:"retainOrdering,omitempty"`
	RetainKeyOrdering            bool             `json:"retainKeyOrdering,omitempty"`
	DeadLetterTopic              string           `json:"deadLetterTopic,omitempty"`
	ForwardSourceMessageProperty *bool            `json:"forwardSourceMessageProperty,omitempty"`
	MaxPendingAsyncRequests      *int32           `json:"maxPendingAsyncRequests,omitempty"`

	RuntimeFlags         string            `json:"runtimeFlags,omitempty"`
	SubscriptionName     string            `json:"subscriptionName,omitempty"`
	CleanupSubscription  bool              `json:"cleanupSubscription,omitempty"`
	SubscriptionPosition SubscribePosition `json:"subscriptionPosition,omitempty"`
	SkipToLatest         bool              `json:"skipToLatest,omitempty"`

	Pod PodPolicy `json:"pod,omitempty"`

	WindowConfig *WindowConfig `json:"windowConfig,omitempty"`

	// +kubebuilder:validation:Required
	Messaging `json:",inline"`

	// +kubebuilder:validation:Required
	Runtime `json:",inline"`

	// Image is the container image used to run function pods.
	// default is streamnative/pulsar-functions-java-runner
	Image string `json:"image,omitempty"`

	// Image pull policy, one of Always, Never, IfNotPresent, default to IfNotPresent.
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// Image which has pulsarctl will use pulsarctl to download package and do cleanup
	ImageHasPulsarctl bool `json:"imageHasPulsarctl,omitempty"`
	// Image which has wget will use wget to download http package
	ImageHasWget bool `json:"imageHasGet,omitempty"`

	// +kubebuilder:validation:Optional
	StateConfig *Stateful `json:"statefulConfig,omitempty"`

	// +kubebuilder:validation:Optional
	VolumeClaimTemplates []corev1.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty"`

	// +kubebuilder:validation:Optional
	PersistentVolumeClaimRetentionPolicy *appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy `json:"persistentVolumeClaimRetentionPolicy,omitempty"`
}

// FunctionStatus defines the observed state of Function
type FunctionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions                      map[Component]ResourceCondition `json:"conditions"`
	Replicas                        int32                           `json:"replicas"`
	Selector                        string                          `json:"selector"`
	ObservedGeneration              int64                           `json:"observedGeneration,omitempty"`
	GlobalBackendConfigRevision     string                          `json:"globalBackendConfigRevision,omitempty"`
	NamespacedBackendConfigRevision string                          `json:"namespacedBackendConfigRevision,omitempty"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
//+kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector

// Function is the Schema for the functions API
// +kubebuilder:pruning:PreserveUnknownFields
type Function struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FunctionSpec   `json:"spec,omitempty"`
	Status FunctionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FunctionList contains a list of Function
type FunctionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Function `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Function{}, &FunctionList{})
}
