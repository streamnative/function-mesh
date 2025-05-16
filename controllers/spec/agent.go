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
)

type AuthenticationConfig struct {
	AuthPlugin string `yaml:"auth_plugin,omitempty"`
	AuthParams string `yaml:"auth_params,omitempty"`
}

type AgentTLSConfig struct {
	UseTLS                      bool   `yaml:"use_tls,omitempty"`
	TlsAllowInsecureConnection  bool   `yaml:"tls_allow_insecure_connection,omitempty"`
	TlsTrustCertPath            string `yaml:"tls_trust_cert_path,omitempty"`
	HostnameVerificationEnabled bool   `yaml:"hostname_verification_enabled,omitempty"`
}

type PulsarClientConfig struct {
	OperationTimeoutSeconds  int `yaml:"operation_timeout_seconds,omitempty"`
	IOThreads                int `yaml:"io_threads,omitempty"`
	MessageListenerThreads   int `yaml:"message_listener_threads,omitempty"`
	ConcurrentLookupRequests int `yaml:"concurrent_lookup_requests,omitempty"`
}

type PulsarClusterConfig struct {
	ClusterName    string                `yaml:"cluster_name,omitempty"`
	ServiceUrl     string                `yaml:"service_url,omitempty"`
	AdminUrl       string                `yaml:"admin_url,omitempty"`
	Client         *PulsarClientConfig   `yaml:"client,omitempty"`
	TLS            *AgentTLSConfig       `yaml:"tls,omitempty"`
	Authentication *AuthenticationConfig `yaml:"authentication,omitempty"`
}

type LoggingConfig struct {
	Directory  string `yaml:"directory,omitempty"`
	FilePrefix string `yaml:"file_prefix,omitempty"`
	Level      string `yaml:"level,omitempty"`
	ConfigFile string `yaml:"config_file,omitempty"`
}

type StateStorageConfig struct {
	ServiceUrl string `yaml:"service_url,omitempty"`
}

type SecretsProviderConfig struct {
	Provider string  `yaml:"provider,omitempty"`
	Config   *string `yaml:"config,omitempty"`
}

type ServerConfig struct {
	Port                        int32 `yaml:"port,omitempty"`
	MetricsPort                 int32 `yaml:"metrics_port,omitempty"`
	ExpectedHealthcheckInterval int   `yaml:"expected_healthcheck_interval,omitempty"`
}

type InstanceConfig struct {
	MaxBufferedTuples int `yaml:"max_buffered_tuples,omitempty"`

	Logging *LoggingConfig `yaml:"logging,omitempty"`

	StateStorage *StateStorageConfig `yaml:"state_storage,omitempty"`

	SecretsProvider *SecretsProviderConfig `yaml:"secrets_provider,omitempty"`

	PulsarCluster *PulsarClusterConfig `yaml:"pulsar_cluster,omitempty"`

	Server *ServerConfig `yaml:"server,omitempty"`
}

type PulsarConsumerSpec struct {
	SchemaType string `yaml:"schema_type,omitempty"`

	SerdeClassName string `yaml:"serde_class_name,omitempty"`

	IsRegexSubscription bool `yaml:"is_regex_subscription,omitempty"`

	ReceiverQueueSize *int32 `yaml:"receiver_queue_size,omitempty"`

	SchemaProperties map[string]string `yaml:"schema_properties,omitempty"`

	ConsumerProperties map[string]string `yaml:"consumer_properties,omitempty"`

	PoolMessages *bool `yaml:"pool_messages,omitempty"`
}

type PulsarProducerSpec struct {
	SchemaType                         string                   `yaml:"schema_type,omitempty"`
	SerdeClassName                     string                   `yaml:"serde_class_name,omitempty"`
	MaxPendingMessages                 *int32                   `yaml:"max_pending_messages,omitempty"`
	MaxPendingMessagesAcrossPartitions *int32                   `yaml:"max_pending_messages_across_partitions,omitempty"`
	UseThreadLocalProducers            *bool                    `yaml:"use_thread_local_producers,omitempty"`
	BatchBuilder                       string                   `yaml:"batch_builder,omitempty"`
	CompressionType                    v1alpha1.CompressionType `yaml:"compression_type,omitempty"`
	SchemaProperties                   map[string]string        `yaml:"schema_properties,omitempty"`
	ConsumerProperties                 map[string]string        `yaml:"producer_properties,omitempty"`
}

type PulsarSourceSpec struct {
	Configs                      map[string]interface{}        `yaml:"configs,omitempty"`
	SubscriptionName             string                        `yaml:"subscriptionName,omitempty"`
	SubscriptionType             string                        `yaml:"subscriptionType,omitempty"`
	InputSpecs                   map[string]PulsarConsumerSpec `yaml:"input_specs,omitempty"`
	TimeoutMs                    int32                         `yaml:"timeout_ms,omitempty"`
	CleanUpSubscription          bool                          `yaml:"cleanup_subscription,omitempty"`
	NegativeAckRedeliveryDelayMs *int                          `yaml:"negative_ack_redelivery_delay_ms,omitempty"`
	SkipToLatest                 bool                          `yaml:"skip_to_latest,omitempty"`
	SkipToEarliest               bool                          `yaml:"skip_to_earliest,omitempty"`
}

type PulsarSinkSpec struct {
	Configs                        map[string]interface{}        `yaml:"configs,omitempty"`
	OutputSpecs                    map[string]PulsarProducerSpec `yaml:"output_specs,omitempty"`
	ForwardSourceMessageProperties *bool                         `yaml:"forward_source_message_properties,omitempty"`
}

type SourceSpec struct {
	Pulsar *PulsarSourceSpec `yaml:"pulsar,omitempty"`
}

type SinkSpec struct {
	Pulsar *PulsarSinkSpec `yaml:"pulsar,omitempty"`
}

type RuntimeSpec struct {
	Tenant              string                        `yaml:"tenant,omitempty"`
	Namespace           string                        `yaml:"namespace,omitempty"`
	Language            string                        `yaml:"language,omitempty"`
	UserConfig          map[string]interface{}        `yaml:"user_config,omitempty"`
	SecretsMap          map[string]v1alpha1.SecretRef `yaml:"secrets_map,omitempty"`
	LogTopic            string                        `yaml:"log_topic,omitempty"`
	Parallelism         int                           `yaml:"parallelism,omitempty"`
	ProcessingGuarantee v1alpha1.ProcessGuarantee     `json:"processingGuarantee,omitempty"`
}

type MetaSpec struct {
	Name string `yaml:"name,omitempty"`

	Description string `yaml:"description,omitempty"`
}

type AgentFunctionSpec struct {
	Meta    MetaSpec    `yaml:"meta,omitempty"`
	Runtime RuntimeSpec `yaml:"runtime,omitempty"`
	Source  SourceSpec  `yaml:"source,omitempty"`
	Sink    SinkSpec    `yaml:"sink,omitempty"`
}
