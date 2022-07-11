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
	"strconv"

	"fmt"
	"strings"

	autov2beta2 "k8s.io/api/autoscaling/v2beta2"

	pctlutil "github.com/streamnative/pulsarctl/pkg/pulsar/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Messaging struct {
	Pulsar *PulsarMessaging `json:"pulsar,omitempty"`
}

type Stateful struct {
	Pulsar *PulsarStateStore `json:"pulsar,omitempty"`
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

	// To replace the TLSSecret
	TLSConfig *PulsarTLSConfig `json:"tlsConfig,omitempty"`
}

type TLSConfig struct {
	Enabled              bool   `json:"enabled,omitempty"`
	AllowInsecure        bool   `json:"allowInsecure,omitempty"`
	HostnameVerification bool   `json:"hostnameVerification,omitempty"`
	CertSecretName       string `json:"certSecretName,omitempty"`
	CertSecretKey        string `json:"certSecretKey,omitempty"`
}

type PulsarTLSConfig struct {
	TLSConfig `json:",inline"`
}

func (c *PulsarTLSConfig) IsEnabled() bool {
	return c.Enabled
}

func (c *PulsarTLSConfig) AllowInsecureConnection() string {
	return strconv.FormatBool(c.AllowInsecure)
}

func (c *PulsarTLSConfig) EnableHostnameVerification() string {
	return strconv.FormatBool(c.HostnameVerification)
}

func (c *PulsarTLSConfig) SecretName() string {
	return c.CertSecretName
}

func (c *PulsarTLSConfig) SecretKey() string {
	return c.CertSecretKey
}

func (c *PulsarTLSConfig) HasSecretVolume() bool {
	return c.CertSecretName != "" && c.CertSecretKey != ""
}

func (c *PulsarTLSConfig) GetMountPath() string {
	return "/etc/tls/pulsar-functions"
}

type PulsarStateStore struct {
	// The service url points to the state store service
	// By default, the state store service is bookkeeper table service
	ServiceURL string `json:"serviceUrl"`

	// The state store config for Java runtime
	JavaProvider *PulsarStateStoreJavaProvider `json:"javaProvider,omitempty"`
}

type PulsarStateStoreJavaProvider struct {
	// The java class name of the state store provider implementation
	// The class must implement `org.apache.pulsar.functions.instance.state.StateStoreProvider` interface
	// If not set, `org.apache.pulsar.functions.instance.state.BKStateStoreProviderImpl` will be used
	ClassName string `json:"className"`

	// The configmap of the configuration for the state store provider
	Config *Config `json:"config,omitempty"`
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

	// BuiltinAutoscaler refers to the built-in autoscaling rules
	// Available values: AverageUtilizationCPUPercent80, AverageUtilizationCPUPercent50, AverageUtilizationCPUPercent20
	// AverageUtilizationMemoryPercent80, AverageUtilizationMemoryPercent50, AverageUtilizationMemoryPercent20
	// +optional
	// TODO: validate the rules, user may provide duplicate rules, should check with webhook
	BuiltinAutoscaler []BuiltinHPARule `json:"builtinAutoscaler,omitempty"`

	// AutoScalingMetrics contains the specifications for which to use to calculate the
	// desired replica count (the maximum replica count across all metrics will
	// be used).
	// More info: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#metricspec-v2beta2-autoscaling
	// +optional
	AutoScalingMetrics []autov2beta2.MetricSpec `json:"autoScalingMetrics,omitempty"`

	// AutoScalingBehavior configures the scaling behavior of the target
	// in both Up and Down directions (scaleUp and scaleDown fields respectively).
	// If not set, the default HPAScalingRules for scale up and scale down are used.
	// +optional
	AutoScalingBehavior *autov2beta2.HorizontalPodAutoscalerBehavior `json:"autoScalingBehavior,omitempty"`
}

type Runtime struct {
	Java   *JavaRuntime   `json:"java,omitempty"`
	Python *PythonRuntime `json:"python,omitempty"`
	Golang *GoRuntime     `json:"golang,omitempty"`
}

// JavaRuntime contains the java runtime configs
// +kubebuilder:validation:Optional
type JavaRuntime struct {
	// +kubebuilder:validation:Required
	Jar                  string `json:"jar"`
	JarLocation          string `json:"jarLocation,omitempty"`
	ExtraDependenciesDir string `json:"extraDependenciesDir,omitempty"`
}

// PythonRuntime contains the python runtime configs
// +kubebuilder:validation:Optional
type PythonRuntime struct {
	// +kubebuilder:validation:Required
	Py         string `json:"py"`
	PyLocation string `json:"pyLocation,omitempty"`
}

// GoRuntime contains the golang runtime configs
// +kubebuilder:validation:Optional
type GoRuntime struct {
	// +kubebuilder:validation:Required
	Go         string `json:"go"`
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

// SubscribePosition enum type
// +kubebuilder:validation:Enum=latest;earliest
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

// ProcessGuarantee enum type
// +kubebuilder:validation:Enum=atleast_once;atmost_once;effectively_once
type ProcessGuarantee string

const (
	AtleastOnce     ProcessGuarantee = "atleast_once"
	AtmostOnce      ProcessGuarantee = "atmost_once"
	EffectivelyOnce ProcessGuarantee = "effectively_once"

	DefaultTenant    string = "public"
	DefaultNamespace string = "default"
	DefaultCluster   string = "kubernetes"

	DefaultResourceCPU    int64 = 1
	DefaultResourceMemory int64 = 1073741824
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

	PackageURLHTTP     string = "http://"
	PackageURLHTTPS    string = "https://"
	PackageURLFunction string = "function://"
	PackageURLSource   string = "source://"
	PackageURLSink     string = "sink://"
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
	out.Data = runtime.DeepCopyJSON(c.Data)
}

func validPackageLocation(packageLocation string) error {
	if hasPackageTypePrefix(packageLocation) {
		err := isValidPulsarPackageURL(packageLocation)
		if err != nil {
			return err
		}
	} else {
		if !isFunctionPackageURLSupported(packageLocation) {
			return fmt.Errorf("invalid function package url %s, supported url (http/https)", packageLocation)
		}
	}

	return nil
}

func hasPackageTypePrefix(packageLocation string) bool {
	lowerCase := strings.ToLower(packageLocation)
	return strings.HasPrefix(lowerCase, PackageURLFunction) ||
		strings.HasPrefix(lowerCase, PackageURLSource) ||
		strings.HasPrefix(lowerCase, PackageURLSink)
}

func isValidPulsarPackageURL(packageLocation string) error {
	parts := strings.Split(packageLocation, "://")
	if len(parts) != 2 {
		return fmt.Errorf("invalid package name %s", packageLocation)
	}
	if !hasPackageTypePrefix(packageLocation) {
		return fmt.Errorf("invalid package name %s", packageLocation)
	}
	rest := parts[1]
	if !strings.Contains(rest, "@") {
		rest += "@"
	}
	packageParts := strings.Split(rest, "@")
	if len(packageParts) != 2 {
		return fmt.Errorf("invalid package name %s", packageLocation)
	}
	partsWithoutVersion := strings.Split(packageParts[0], "/")
	if len(partsWithoutVersion) != 3 {
		return fmt.Errorf("invalid package name %s", packageLocation)
	}
	return nil
}

func isFunctionPackageURLSupported(packageLocation string) bool {
	// TODO: support file:// schema
	lowerCase := strings.ToLower(packageLocation)
	return strings.HasPrefix(lowerCase, PackageURLHTTP) ||
		strings.HasPrefix(lowerCase, PackageURLHTTPS)
}

func collectAllInputTopics(inputs InputConf) []string {
	ret := []string{}
	if len(inputs.Topics) > 0 {
		ret = append(ret, inputs.Topics...)
	}
	if inputs.TopicPattern != "" {
		ret = append(ret, inputs.TopicPattern)
	}
	if len(inputs.CustomSerdeSources) > 0 {
		for k := range inputs.CustomSerdeSources {
			ret = append(ret, k)
		}
	}
	if len(inputs.CustomSchemaSources) > 0 {
		for k := range inputs.CustomSchemaSources {
			ret = append(ret, k)
		}
	}
	if len(inputs.SourceSpecs) > 0 {
		for k := range inputs.SourceSpecs {
			ret = append(ret, k)
		}
	}
	return ret
}

func isValidTopicName(topicName string) error {
	_, err := pctlutil.GetTopicName(topicName)
	return err
}
