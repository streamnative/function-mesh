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
	"fmt"
	"strconv"

	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"

	autov2beta2 "k8s.io/api/autoscaling/v2beta2"

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

	// To replace the AuthSecret
	AuthConfig *AuthConfig `json:"authConfig,omitempty"`
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

type AuthConfig struct {
	OAuth2Config *OAuth2Config `json:"oauth2Config,omitempty"`
	GenericAuth  *GenericAuth  `json:"genericAuth,omitempty"`
	// TODO: support other auth providers
}

type OAuth2Config struct {
	Audience  string `json:"audience"`
	IssuerURL string `json:"issuerUrl"`
	Scope     string `json:"scope,omitempty"`
	// the secret name of the OAuth2 private key file
	KeySecretName string `json:"keySecretName"`
	// the secret key of the OAuth2 private key file, such as `auth.json`
	KeySecretKey string `json:"keySecretKey"`
}

func (o *OAuth2Config) GetMountPath() string {
	return "/etc/oauth2"
}

func (o *OAuth2Config) GetMountFile() string {
	return fmt.Sprintf("%s/%s", o.GetMountPath(), o.KeySecretKey)
}

func (o *OAuth2Config) AuthenticationParameters() string {
	return fmt.Sprintf(`'{"privateKey":"%s","private_key":"%s","issuerUrl":"%s","issuer_url":"%s","audience":"%s","scope":"%s"}'`, o.GetMountFile(), o.GetMountFile(), o.IssuerURL, o.IssuerURL, o.Audience, o.Scope)
}

type GenericAuth struct {
	ClientAuthenticationPlugin     string `json:"clientAuthenticationPlugin"`
	ClientAuthenticationParameters string `json:"clientAuthenticationParameters"`
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

	// VPA indicates whether to enable the VerticalPodAutoscaler, it should not be used with HPA
	VPA *VPASpec `json:"vpa,omitempty"`

	// Env Environment variables to expose on the pulsar-function containers
	Env []corev1.EnvVar `json:"env,omitempty"`

	Liveness *Liveness `json:"liveness,omitempty"`
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
	Jar                  string            `json:"jar"`
	JarLocation          string            `json:"jarLocation,omitempty"`
	ExtraDependenciesDir string            `json:"extraDependenciesDir,omitempty"`
	Log                  *RuntimeLogConfig `json:"log,omitempty"`
	JavaOpts             []string          `json:"javaOpts,omitempty"`
}

// PythonRuntime contains the python runtime configs
// +kubebuilder:validation:Optional
type PythonRuntime struct {
	// +kubebuilder:validation:Required
	Py         string            `json:"py"`
	PyLocation string            `json:"pyLocation,omitempty"`
	Log        *RuntimeLogConfig `json:"log,omitempty"`
}

// GoRuntime contains the golang runtime configs
// +kubebuilder:validation:Optional
type GoRuntime struct {
	// +kubebuilder:validation:Required
	Go         string            `json:"go"`
	GoLocation string            `json:"goLocation,omitempty"`
	Log        *RuntimeLogConfig `json:"log,omitempty"`
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
	VPA         Component = "VerticalPodAutoscaler"
)

// ResourceCondition The `Status` of a given `Condition` and the `Action` needed to reach the `Status`
type ResourceCondition struct {
	Condition ResourceConditionType  `json:"condition,omitempty"`
	Status    metav1.ConditionStatus `json:"status,omitempty"`
	Action    ReconcileAction        `json:"action,omitempty"`
}

type ResourceConditionType string

const (
	Orphaned ResourceConditionType = "Orphaned"

	MeshReady     ResourceConditionType = "MeshReady"
	FunctionReady ResourceConditionType = "FunctionReady"
	SourceReady   ResourceConditionType = "SourceReady"
	SinkReady     ResourceConditionType = "SinkReady"

	StatefulSetReady ResourceConditionType = "StatefulSetReady"
	ServiceReady     ResourceConditionType = "ServiceReady"
	HPAReady         ResourceConditionType = "HPAReady"
	VPAReady         ResourceConditionType = "VPAReady"
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
)

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
	out.Data = runtime.DeepCopyJSON(c.Data)
}

func CreateCondition(condType ResourceConditionType, status metav1.ConditionStatus, action ReconcileAction) ResourceCondition {
	condition := ResourceCondition{
		Condition: condType,
		Status:    status,
		Action:    action,
	}
	return condition
}

const (
	// LogLevelOff indicates no logging and is only available for the Java runtime
	LogLevelOff LogLevel = "off"
	// LogLevelTrace indicates the detailed debugging purposes, available for Python, Go and Java runtime
	LogLevelTrace LogLevel = "trace"
	// LogLevelDebug indicates the debugging purposes, available for Python, Go and Java runtime
	LogLevelDebug LogLevel = "debug"
	// LogLevelInfo indicates the normal purposes, available for Python, Go and Java runtime
	LogLevelInfo LogLevel = "info"
	// LogLevelWarn indicates the unexpected purposes, available for Python, Go and Java runtime
	LogLevelWarn LogLevel = "warn"
	// LogLevelError indicates the errors have occurred, available for Python, Go and Java runtime
	LogLevelError LogLevel = "error"
	// LogLevelFatal indicates the server is unusable, available for Python, Go and Java runtime
	LogLevelFatal LogLevel = "fatal"
	// LogLevelAll indicates that all logs are logged and is only available for the Java runtime
	LogLevelAll LogLevel = "all"
	// LogLevelPanic indicates the server is panic and is only available for the Go runtime
	LogLevelPanic LogLevel = "panic"
)

// LogLevel describes the level of the logging
// +kubebuilder:validation:Enum=off;trace;debug;info;warn;error;fatal;all;panic
type LogLevel string

const (
	TimedPolicyWithDaily   TriggeringPolicy = "TimedPolicyWithDaily"
	TimedPolicyWithWeekly  TriggeringPolicy = "TimedPolicyWithWeekly"
	TimedPolicyWithMonthly TriggeringPolicy = "TimedPolicyWithMonthly"
	SizedPolicyWith10MB    TriggeringPolicy = "SizedPolicyWith10MB"
	SizedPolicyWith50MB    TriggeringPolicy = "SizedPolicyWith50MB"
	SizedPolicyWith100MB   TriggeringPolicy = "SizedPolicyWith100MB"
)

// TriggeringPolicy is using to determine if a rollover should occur.
// +kubebuilder:validation:Enum=TimedPolicyWithDaily;TimedPolicyWithWeekly;TimedPolicyWithMonthly;SizedPolicyWith10MB;SizedPolicyWith50MB;SizedPolicyWith100MB
type TriggeringPolicy string

type RuntimeLogConfig struct {
	Level        LogLevel          `json:"level,omitempty"`
	RotatePolicy *TriggeringPolicy `json:"rotatePolicy,omitempty"`
	LogConfig    *LogConfig        `json:"logConfig,omitempty"`
}

type LogConfig struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

type WindowConfig struct {
	ActualWindowFunctionClassName string  `json:"actualWindowFunctionClassName"`
	WindowLengthCount             *int32  `json:"windowLengthCount,omitempty"`
	WindowLengthDurationMs        *int64  `json:"windowLengthDurationMs,omitempty"`
	SlidingIntervalCount          *int32  `json:"slidingIntervalCount,omitempty"`
	SlidingIntervalDurationMs     *int64  `json:"slidingIntervalDurationMs,omitempty"`
	LateDataTopic                 string  `json:"lateDataTopic,omitempty"`
	MaxLagMs                      *int64  `json:"maxLagMs,omitempty"`
	WatermarkEmitIntervalMs       *int64  `json:"watermarkEmitIntervalMs,omitempty"`
	TimestampExtractorClassName   *string `json:"timestampExtractorClassName,omitempty"`
}

type VPASpec struct {
	// Describes the rules on how changes are applied to the pods.
	// If not specified, all fields in the `PodUpdatePolicy` are set to their
	// default values.
	// +optional
	UpdatePolicy *vpav1.PodUpdatePolicy `json:"updatePolicy,omitempty"`

	// Controls how the autoscaler computes recommended resources.
	// +optional
	ResourcePolicy *vpav1.PodResourcePolicy `json:"resourcePolicy,omitempty"`
}

type Liveness struct {
	// +kubebuilder:validation:Optional
	PeriodSeconds int32 `json:"periodSeconds,omitempty"`

	// some functions may take a long time to start up(like download packages), so we need to set the initial delay
	// +kubebuilder:validation:Optional
	InitialDelaySeconds int32 `json:"initialDelaySeconds,omitempty"`

	// Minimum consecutive successes for the probe to be considered successful after having failed.
	// Defaults to 1. Must be 1 for liveness and startup. Minimum value is 1.
	// +optional
	SuccessThreshold int32 `json:"successThreshold,omitempty"`
	// Minimum consecutive failures for the probe to be considered failed after having succeeded.
	// Defaults to 3. Minimum value is 1.
	// +optional
	FailureThreshold int32 `json:"failureThreshold,omitempty"`
}

func (rc *ResourceCondition) SetCondition(condition ResourceConditionType, action ReconcileAction, status metav1.ConditionStatus) {
	rc.Condition = condition
	rc.Action = action
	rc.Status = status
}
