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
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Messaging struct {
	Pulsar *PulsarMessaging `json:"pulsar,omitempty"`
}

type PulsarMessaging struct {
	// The config map need to contain the following fields
	// webServiceURL
	// brokerServiceURL
	PulsarConfig string `json:"pulsarConfig,omitempty"`

	// The auth secret should contain the following fields
	// clientAuthenticationPlugin
	// clientAuthenticationParameters
	AuthSecret string `json:"authSecret,omitempty"`

	// The TLS secret should contain the following fields
	// use_tls
	// tls_allow_insecure
	// hostname_verification_enabled
	// tls_trust_cert_path
	TLSSecret string `json:"tlsSecret,omitempty"`
}

type PodPolicy struct {
	// Labels specifies the labels to attach to pod the operator creates for the cluster.
	Labels map[string]string `json:"labels,omitempty"`

	// NodeSelector specifies a map of key-value pairs. For a pod to be eligible to run
	// on a node, the node must have each of the indicated key-value pairs as labels.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Affinity specifies the scheduling constraints of a pod
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// Tolerations specifies the tolerations of a Pod
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// Annotations specifies the annotations to attach to pods the operator creates
	Annotations map[string]string `json:"annotations,omitempty"`

	// SecurityContext specifies the security context for the entire pod
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/security-context
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`

	// TerminationGracePeriodSeconds is the amount of time that kubernetes will give
	// for a pod before terminating it.
	TerminationGracePeriodSeconds int64 `json:"terminationGracePeriodSeconds,omitempty"`

	// List of volumes that can be mounted by containers belonging to the pod.
	// More info: https://kubernetes.io/docs/concepts/storage/volumes
	Volumes []corev1.Volume `json:"volumes,omitempty"`

	// ImagePullSecrets is an optional list of references to secrets in the same
	// namespace to use for pulling any of the images used by this PodSpec.
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// Init containers of the pod. A typical use case could be using an init
	// container to download a remote jar to a local path.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
	InitContainers []corev1.Container `json:"initContainers,omitempty"`

	// Sidecar containers running alongside with the main function container in the
	// pod.
	Sidecars []corev1.Container `json:"sidecars,omitempty"`

	// ServiceAccountName is the name of the ServiceAccount to use to run this pod.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
}

type Runtime struct {
	Java   *JavaRuntime   `json:"java,omitempty"`
	Python *PythonRuntime `json:"python,omitempty"`
	Golang *GoRuntime     `json:"golang,omitempty"`
}

type JavaRuntime struct {
	Jar                  string `json:"jar,omitempty"`
	JarLocation          string `json:"jarLocation,omitempty"`
	ExtraDependenciesDir string `json:"extraDependenciesDir,omitempty"`
}

type PythonRuntime struct {
	Py         string `json:"py,omitempty"`
	PyLocation string `json:"pyLocation,omitempty"`
}

type GoRuntime struct {
	Go         string `json:"go,omitempty"`
	GoLocation string `json:"goLocation,omitempty"`
}

type SecretRef struct {
	Path string `json:"path,omitempty"`
	Key  string `json:"key,omitempty"`
}

type InputConf struct {
	TypeClassName       string                    `json:"typeClassName,omitempty"`
	Topics              []string                  `json:"topics,omitempty"`
	TopicPattern        string                    `json:"topicPattern,omitempty"`
	CustomSerdeSources  map[string]string         `json:"customSerdeSources,omitempty"`
	CustomSchemaSources map[string]string         `json:"customSchemaSources,omitempty"`
	SourceSpecs         map[string]ConsumerConfig `json:"sourceSpecs,omitempty"`
}

type ConsumerConfig struct {
	SchemaType         string            `json:"schemaType,omitempty"`
	SerdeClassName     string            `json:"serdeClassname,omitempty"`
	IsRegexPattern     bool              `json:"isRegexPattern,omitempty"`
	SchemaProperties   map[string]string `json:"schemaProperties,omitempty"`
	ConsumerProperties map[string]string `json:"consumerProperties,omitempty"`
	ReceiverQueueSize  *int32            `json:"receiverQueueSize,omitempty"`
	CryptoConfig       *CryptoConfig     `json:"cryptoConfig,omitempty"`
}

type OutputConf struct {
	TypeClassName      string            `json:"typeClassName,omitempty"`
	Topic              string            `json:"topic,omitempty"`
	SinkSerdeClassName string            `json:"sinkSerdeClassName,omitempty"`
	SinkSchemaType     string            `json:"sinkSchemaType,omitempty"`
	ProducerConf       *ProducerConfig   `json:"producerConf,omitempty"`
	CustomSchemaSinks  map[string]string `json:"customSchemaSinks,omitempty"`
}

type ProducerConfig struct {
	MaxPendingMessages                 int32         `json:"maxPendingMessages,omitempty"`
	MaxPendingMessagesAcrossPartitions int32         `json:"maxPendingMessagesAcrossPartitions,omitempty"`
	UseThreadLocalProducers            bool          `json:"useThreadLocalProducers,omitempty"`
	CryptoConfig                       *CryptoConfig `json:"cryptoConfig,omitempty"`
	BatchBuilder                       string        `json:"batchBuilder,omitempty"`
}

type CryptoConfig struct {
	CryptoKeyReaderClassName    string            `json:"cryptoKeyReaderClassName,omitempty"`
	CryptoKeyReaderConfig       map[string]string `json:"cryptoKeyReaderConfig,omitempty"`
	EncryptionKeys              []string          `json:"encryptionKeys,omitempty"`
	ProducerCryptoFailureAction string            `json:"producerCryptoFailureAction,omitempty"`
	ConsumerCryptoFailureAction string            `json:"consumerCryptoFailureAction,omitempty"`
	CryptoSecrets               []CryptoSecret    `json:"cryptoSecrets,omitempty"`
}

type CryptoSecret struct {
	SecretName string `json:"secretName"`
	SecretKey  string `json:"secretKey"`
	AsVolume   string `json:"asVolume,omitempty"`
	//AsEnv      string `json:"asEnv,omitempty"`
}

type SubscribePosition string

const (
	Latest   SubscribePosition = "latest"
	Earliest SubscribePosition = "earliest"
)

type Component string

const (
	StatefulSet Component = "StatefulSet"
	Service     Component = "Service"
	HPA         Component = "HorizontalPodAutoscaler"
)

// The `Status` of a given `Condition` and the `Action` needed to reach the `Status`
type ResourceCondition struct {
	Condition ResourceConditionType  `json:"condition,omitempty"`
	Status    metav1.ConditionStatus `json:"status,omitempty"`
	Action    ReconcileAction        `json:"action,omitempty"`
}

type ResourceConditionType string

const (
	FunctionReady ResourceConditionType = "FunctionReady"
	SourceReady   ResourceConditionType = "SourceReady"
	SinkReady     ResourceConditionType = "SinkReady"

	StatefulSetReady ResourceConditionType = "StatefulSetReady"
	ServiceReady     ResourceConditionType = "ServiceReady"
	HPAReady         ResourceConditionType = "HPAReady"
)

type ReconcileAction string

const (
	Create   ReconcileAction = "Create"
	Delete   ReconcileAction = "Delete"
	Update   ReconcileAction = "Update"
	Wait     ReconcileAction = "Wait"
	NoAction ReconcileAction = "NoAction"
)

type ProcessGuarantee string

const (
	AtleastOnce     ProcessGuarantee = "atleast_once"
	AtmostOnce      ProcessGuarantee = "atmost_once"
	EffectivelyOnce ProcessGuarantee = "effectively_once"

	DefaultTenant  string = "public"
	DefaultCluster string = "kubernetes"
)

func validResourceRequirement(requirements corev1.ResourceRequirements) bool {
	return validResource(requirements.Requests) && validResource(requirements.Limits) &&
		requirements.Requests.Memory().Cmp(*requirements.Limits.Memory()) <= 0 &&
		requirements.Requests.Cpu().Cmp(*requirements.Limits.Cpu()) <= 0
}

func validResource(resources corev1.ResourceList) bool {
	// cpu & memory > 0 and storage >= 0
	return resources.Cpu().Sign() == 1 &&
		resources.Memory().Sign() == 1 &&
		resources.Storage().Sign() >= 0
}

func paddingResourceLimit(requirement *corev1.ResourceRequirements) {
	// TODO: better padding calculation
	requirement.Limits.Cpu().Set(requirement.Requests.Cpu().Value())
	requirement.Limits.Memory().Set(requirement.Requests.Memory().Value())
	requirement.Limits.Storage().Set(requirement.Requests.Storage().Value())
}

const (
	FunctionComponent string = "function"
	SourceComponent   string = "source"
	SinkComponent     string = "sink"
)

// Config represents untyped YAML configuration.
type Config struct {
	// Data holds the configuration keys and values.
	// This field exists to work around https://github.com/kubernetes-sigs/kubebuilder/issues/528
	Data map[string]interface{} `json:"-"`
}

// NewConfig constructs a Config with the given unstructured configuration data.
func NewConfig(cfg map[string]interface{}) Config {
	return Config{Data: cfg}
}

// MarshalJSON implements the Marshaler interface.
func (c *Config) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Data)
}

// UnmarshalJSON implements the Unmarshaler interface.
func (c *Config) UnmarshalJSON(data []byte) error {
	var out map[string]interface{}
	err := json.Unmarshal(data, &out)
	if err != nil {
		return err
	}
	c.Data = out
	return nil
}

// DeepCopyInto is an ~autogenerated~ deepcopy function, copying the receiver, writing into out. in must be non-nil.
// This exists here to work around https://github.com/kubernetes/code-generator/issues/50
func (c *Config) DeepCopyInto(out *Config) {
	bytes, err := json.Marshal(c.Data)
	if err != nil {
		// we assume that it marshals cleanly because otherwise the resource would not have been
		// created in the API server
		panic(err)
	}
	var clone map[string]interface{}
	err = json.Unmarshal(bytes, &clone)
	if err != nil {
		// we assume again optimistically because we just marshalled that the round trip works as well
		panic(err)
	}
	out.Data = clone
}
