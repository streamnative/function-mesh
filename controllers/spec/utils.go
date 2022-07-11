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
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/streamnative/function-mesh/api/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/proto"
)

func convertFunctionDetails(function *v1alpha1.Function) *proto.FunctionDetails {
	fd := &proto.FunctionDetails{
		Tenant:               function.Spec.Tenant,
		Namespace:            function.Spec.Namespace,
		Name:                 function.Spec.Name,
		ClassName:            function.Spec.ClassName,
		LogTopic:             function.Spec.LogTopic,
		ProcessingGuarantees: convertProcessingGuarantee(function.Spec.ProcessingGuarantee),
		UserConfig:           getUserConfig(function.Spec.FuncConfig),
		Runtime:              proto.FunctionDetails_JAVA,
		AutoAck:              getBoolFromPtrOrDefault(function.Spec.AutoAck, true),
		Parallelism:          getInt32FromPtrOrDefault(function.Spec.Replicas, 1),
		Source:               generateFunctionInputSpec(function),
		Sink:                 generateFunctionOutputSpec(function),
		Resources:            generateResource(function.Spec.Resources.Requests),
		PackageUrl:           "",
		RetryDetails:         generateRetryDetails(function.Spec.MaxMessageRetry, function.Spec.DeadLetterTopic),
		RuntimeFlags:         function.Spec.RuntimeFlags,
		ComponentType:        proto.FunctionDetails_FUNCTION,
		CustomRuntimeOptions: "",
		Builtin:              "",
		RetainOrdering:       function.Spec.RetainOrdering,
		RetainKeyOrdering:    function.Spec.RetainKeyOrdering,
	}

	if function.Spec.SecretsMap != nil {
		fd.SecretsMap = marshalSecretsMap(function.Spec.SecretsMap)
	}

	return fd
}

func convertGoFunctionConfs(function *v1alpha1.Function) *GoFunctionConf {
	return &GoFunctionConf{
		FuncID:               fmt.Sprintf("${%s}-%s", EnvShardID, string(function.UID)),
		PulsarServiceURL:     "${brokerServiceURL}",
		FuncVersion:          "0",
		MaxBufTuples:         100, //TODO
		Port:                 int(GRPCPort.ContainerPort),
		ClusterName:          function.Spec.ClusterName,
		Tenant:               function.Spec.Tenant,
		NameSpace:            function.Spec.Namespace,
		Name:                 function.Spec.Name,
		LogTopic:             function.Spec.LogTopic,
		ProcessingGuarantees: int32(convertProcessingGuarantee(function.Spec.ProcessingGuarantee)),
		//SecretsMap:                  marshalSecretsMap(function.Spec.SecretsMap),
		Runtime:                     int32(proto.FunctionDetails_GO),
		AutoACK:                     getBoolFromPtrOrDefault(function.Spec.AutoAck, true),
		Parallelism:                 getInt32FromPtrOrDefault(function.Spec.Replicas, 1),
		TimeoutMs:                   uint64(function.Spec.Timeout),
		SubscriptionName:            function.Spec.SubscriptionName,
		CleanupSubscription:         function.Spec.CleanupSubscription,
		SourceSpecTopic:             function.Spec.Input.Topics[0],
		SourceSchemaType:            "", // TODO: map schema type
		IsRegexPatternSubscription:  function.Spec.Input.TopicPattern != "",
		SinkSpecTopic:               function.Spec.Output.Topic,
		SinkSchemaType:              "", // TODO: map schema type
		CPU:                         float64(function.Spec.Resources.Requests.Cpu().Value()),
		RAM:                         function.Spec.Resources.Requests.Memory().Value(),
		Disk:                        function.Spec.Resources.Requests.Storage().Value(),
		MaxMessageRetries:           function.Spec.MaxMessageRetry,
		DeadLetterTopic:             function.Spec.DeadLetterTopic,
		UserConfig:                  getUserConfig(function.Spec.FuncConfig),
		MetricsPort:                 int(MetricsPort.ContainerPort),
		ExpectedHealthCheckInterval: -1, // TurnOff BuiltIn HealthCheck to avoid instance exit
	}
}

func generateInputSpec(sourceConf v1alpha1.InputConf) map[string]*proto.ConsumerSpec {
	inputSpecs := make(map[string]*proto.ConsumerSpec)

	for _, source := range sourceConf.Topics {
		inputSpecs[source] = &proto.ConsumerSpec{
			IsRegexPattern: false,
		}
	}

	if sourceConf.TopicPattern != "" {
		inputSpecs[sourceConf.TopicPattern] = &proto.ConsumerSpec{
			IsRegexPattern: true,
		}
	}

	for topicName, serdeClassName := range sourceConf.CustomSerdeSources {
		inputSpecs[topicName] = &proto.ConsumerSpec{
			SerdeClassName: serdeClassName,
			IsRegexPattern: false,
		}
	}

	for topicName, schemaType := range sourceConf.CustomSchemaSources {
		inputSpecs[topicName] = &proto.ConsumerSpec{
			SchemaType:     schemaType,
			IsRegexPattern: false,
		}
	}

	if sourceConf.SourceSpecs != nil && len(sourceConf.SourceSpecs) > 0 {
		for topicName, conf := range sourceConf.SourceSpecs {
			var receiverQueueSize *proto.ConsumerSpec_ReceiverQueueSize
			if conf.ReceiverQueueSize != nil {
				receiverQueueSize = &proto.ConsumerSpec_ReceiverQueueSize{Value: *conf.ReceiverQueueSize}
			} else {
				receiverQueueSize = nil
			}
			inputSpecs[topicName] = &proto.ConsumerSpec{
				SchemaType:         conf.SchemaType,
				SerdeClassName:     conf.SerdeClassName,
				IsRegexPattern:     conf.IsRegexPattern,
				ReceiverQueueSize:  receiverQueueSize,
				SchemaProperties:   conf.SchemaProperties,
				ConsumerProperties: conf.ConsumerProperties,
				CryptoSpec:         generateCryptoSpec(conf.CryptoConfig),
			}
		}
	}

	return inputSpecs
}

func generateFunctionInputSpec(function *v1alpha1.Function) *proto.SourceSpec {
	inputSpecs := generateInputSpec(function.Spec.Input)

	return &proto.SourceSpec{
		ClassName:                    "",
		Configs:                      "",
		TypeClassName:                function.Spec.Input.TypeClassName,
		SubscriptionType:             proto.SubscriptionType_SHARED,
		InputSpecs:                   inputSpecs,
		TimeoutMs:                    uint64(function.Spec.Timeout),
		Builtin:                      "",
		SubscriptionName:             function.Spec.SubscriptionName,
		CleanupSubscription:          function.Spec.CleanupSubscription,
		SubscriptionPosition:         convertSubPosition(function.Spec.SubscriptionPosition),
		NegativeAckRedeliveryDelayMs: uint64(function.Spec.Timeout),
	}
}

func generateFunctionOutputSpec(function *v1alpha1.Function) *proto.SinkSpec {
	sinkSpec := &proto.SinkSpec{
		ClassName:                    "",
		Configs:                      "",
		TypeClassName:                function.Spec.Output.TypeClassName,
		Topic:                        function.Spec.Output.Topic,
		ProducerSpec:                 nil,
		SerDeClassName:               function.Spec.Output.SinkSerdeClassName,
		Builtin:                      "",
		SchemaType:                   function.Spec.Output.SinkSchemaType,
		ForwardSourceMessageProperty: *function.Spec.ForwardSourceMessageProperty,
		SchemaProperties:             nil,
		ConsumerProperties:           nil,
	}

	if function.Spec.Output.CustomSchemaSinks != nil && function.Spec.Output.Topic != "" {
		conf := function.Spec.Output.CustomSchemaSinks[function.Spec.Output.Topic]
		config := unmarshalConsumerConfig(conf)
		sinkSpec.SchemaProperties = config.SchemaProperties
		sinkSpec.ConsumerProperties = config.ConsumerProperties
	}

	if function.Spec.Output.ProducerConf != nil {
		producerConfig := &proto.ProducerSpec{
			MaxPendingMessages:                 function.Spec.Output.ProducerConf.MaxPendingMessages,
			MaxPendingMessagesAcrossPartitions: function.Spec.Output.ProducerConf.MaxPendingMessagesAcrossPartitions,
			UseThreadLocalProducers:            function.Spec.Output.ProducerConf.UseThreadLocalProducers,
			CryptoSpec:                         generateCryptoSpec(function.Spec.Output.ProducerConf.CryptoConfig),
			BatchBuilder:                       function.Spec.Output.ProducerConf.BatchBuilder,
		}

		sinkSpec.ProducerSpec = producerConfig
	}

	return sinkSpec
}

func convertSourceDetails(source *v1alpha1.Source) *proto.FunctionDetails {
	fd := &proto.FunctionDetails{
		Tenant:               source.Spec.Tenant,
		Namespace:            source.Spec.Namespace,
		Name:                 source.Name,
		ClassName:            "org.apache.pulsar.functions.api.utils.IdentityFunction",
		ProcessingGuarantees: convertProcessingGuarantee(source.Spec.ProcessingGuarantee),
		UserConfig:           getUserConfig(source.Spec.SourceConfig),
		Runtime:              proto.FunctionDetails_JAVA,
		AutoAck:              true,
		Parallelism:          getInt32FromPtrOrDefault(source.Spec.Replicas, 1),
		Source:               generateSourceInputSpec(source),
		Sink:                 generateSourceOutputSpec(source),
		Resources:            generateResource(source.Spec.Resources.Requests),
		RuntimeFlags:         source.Spec.RuntimeFlags,
		ComponentType:        proto.FunctionDetails_SOURCE,
	}

	if source.Spec.SecretsMap != nil {
		fd.SecretsMap = marshalSecretsMap(source.Spec.SecretsMap)
	}

	return fd
}

func generateSourceInputSpec(source *v1alpha1.Source) *proto.SourceSpec {
	configs := getUserConfig(source.Spec.SourceConfig)
	return &proto.SourceSpec{
		ClassName:     source.Spec.ClassName,
		Configs:       configs, // TODO handle batch source
		TypeClassName: source.Spec.Output.TypeClassName,
	}
}

func generateSourceOutputSpec(source *v1alpha1.Source) *proto.SinkSpec {
	var producerSpec proto.ProducerSpec
	var cryptoSpec *proto.CryptoSpec
	if source.Spec.Output.ProducerConf != nil {
		if source.Spec.Output.ProducerConf.CryptoConfig != nil {
			cryptoSpec = generateCryptoSpec(source.Spec.Output.ProducerConf.CryptoConfig)
		}
		producerSpec = proto.ProducerSpec{
			MaxPendingMessages:                 source.Spec.Output.ProducerConf.MaxPendingMessages,
			MaxPendingMessagesAcrossPartitions: source.Spec.Output.ProducerConf.MaxPendingMessagesAcrossPartitions,
			UseThreadLocalProducers:            source.Spec.Output.ProducerConf.UseThreadLocalProducers,
			CryptoSpec:                         cryptoSpec,
			BatchBuilder:                       source.Spec.Output.ProducerConf.BatchBuilder,
		}
	}
	var forward = true
	if source.Spec.ForwardSourceMessageProperty != nil {
		forward = *source.Spec.ForwardSourceMessageProperty
	}
	return &proto.SinkSpec{
		TypeClassName:                source.Spec.Output.TypeClassName,
		Topic:                        source.Spec.Output.Topic,
		ProducerSpec:                 &producerSpec,
		SerDeClassName:               source.Spec.Output.SinkSerdeClassName,
		SchemaType:                   source.Spec.Output.SinkSchemaType,
		ForwardSourceMessageProperty: forward,
	}
}

func convertSinkDetails(sink *v1alpha1.Sink) *proto.FunctionDetails {
	fd := &proto.FunctionDetails{
		Tenant:               sink.Spec.Tenant,
		Namespace:            sink.Spec.Namespace,
		Name:                 sink.Name,
		ClassName:            "org.apache.pulsar.functions.api.utils.IdentityFunction",
		ProcessingGuarantees: convertProcessingGuarantee(sink.Spec.ProcessingGuarantee),
		Runtime:              proto.FunctionDetails_JAVA,
		AutoAck:              getBoolFromPtrOrDefault(sink.Spec.AutoAck, true),
		Parallelism:          getInt32FromPtrOrDefault(sink.Spec.Replicas, 1),
		Source:               generateSinkInputSpec(sink),
		Sink:                 generateSinkOutputSpec(sink),
		Resources:            generateResource(sink.Spec.Resources.Requests),
		RetryDetails:         generateRetryDetails(sink.Spec.MaxMessageRetry, sink.Spec.DeadLetterTopic),
		RuntimeFlags:         sink.Spec.RuntimeFlags,
		ComponentType:        proto.FunctionDetails_SINK,
	}

	if sink.Spec.SecretsMap != nil {
		fd.SecretsMap = marshalSecretsMap(sink.Spec.SecretsMap)
	}

	return fd
}

func generateSinkInputSpec(sink *v1alpha1.Sink) *proto.SourceSpec {
	inputSpecs := generateInputSpec(sink.Spec.Input)

	return &proto.SourceSpec{
		TypeClassName:                sink.Spec.Input.TypeClassName,
		SubscriptionType:             getSubscriptionType(sink.Spec.RetainOrdering, sink.Spec.ProcessingGuarantee),
		InputSpecs:                   inputSpecs,
		TimeoutMs:                    uint64(sink.Spec.Timeout),
		SubscriptionName:             sink.Spec.SubscriptionName,
		CleanupSubscription:          sink.Spec.CleanupSubscription,
		SubscriptionPosition:         convertSubPosition(sink.Spec.SubscriptionPosition),
		NegativeAckRedeliveryDelayMs: uint64(sink.Spec.NegativeAckRedeliveryDelayMs),
	}
}

func getSubscriptionType(retainOrdering bool, processingGuarantee v1alpha1.ProcessGuarantee) proto.SubscriptionType {
	if retainOrdering || processingGuarantee == v1alpha1.EffectivelyOnce {
		return proto.SubscriptionType_FAILOVER
	}

	return proto.SubscriptionType_SHARED
}

func generateSinkOutputSpec(sink *v1alpha1.Sink) *proto.SinkSpec {
	configs := getUserConfig(sink.Spec.SinkConfig)
	return &proto.SinkSpec{
		ClassName:     sink.Spec.ClassName,
		Configs:       configs,
		TypeClassName: sink.Spec.Input.TypeClassName,
	}
}

func marshalSecretsMap(secrets map[string]v1alpha1.SecretRef) string {
	bytes, err := json.Marshal(secrets)
	if err != nil || string(bytes) == "null" {
		return "{}"
	}
	return string(bytes)
}

func unmarshalConsumerConfig(conf string) v1alpha1.ConsumerConfig {
	var config v1alpha1.ConsumerConfig
	// TODO: check unmarshel error in admission hook
	json.Unmarshal([]byte(conf), &config)
	return config
}

func generateCryptoSpec(conf *v1alpha1.CryptoConfig) *proto.CryptoSpec {
	if conf == nil {
		return nil
	}
	cryptoSpec := &proto.CryptoSpec{
		CryptoKeyReaderClassName:    conf.CryptoKeyReaderClassName,
		ProducerEncryptionKeyName:   conf.EncryptionKeys,
		ProducerCryptoFailureAction: getProducerProtoFailureAction(conf.ProducerCryptoFailureAction),
		ConsumerCryptoFailureAction: getConsumerProtoFailureAction(conf.ConsumerCryptoFailureAction),
	}
	if conf.CryptoKeyReaderConfig != nil {
		configs, err := json.Marshal(conf.CryptoKeyReaderConfig)
		if err == nil {
			cryptoSpec.CryptoKeyReaderConfig = string(configs)
		}
	}
	return cryptoSpec
}

func getConsumerProtoFailureAction(action string) proto.CryptoSpec_FailureAction {
	if r, has := proto.CryptoSpec_FailureAction_value[action]; has {
		return proto.CryptoSpec_FailureAction(r)
	}
	return proto.CryptoSpec_FAIL
}

func getProducerProtoFailureAction(action string) proto.CryptoSpec_FailureAction {
	if r, has := proto.CryptoSpec_FailureAction_value[action]; has {
		return proto.CryptoSpec_FailureAction(r)
	}
	return proto.CryptoSpec_FAIL
}

func generateVolumeNameFromCryptoSecrets(c *v1alpha1.CryptoSecret) string {
	return sanitizeVolumeName(c.SecretName + "-" + c.SecretKey)
}

func generateVolumeNameFromTLSConfig(c TLSConfig) string {
	return sanitizeVolumeName(c.SecretName() + "-" + c.SecretKey())
}

var invalidDNS1123Characters = regexp.MustCompile("[^-a-z0-9]+")

// sanitizeVolumeName ensures that the given volume name is a valid DNS-1123 label
// accepted by Kubernetes.
func sanitizeVolumeName(name string) string {
	name = strings.ToLower(name)
	name = invalidDNS1123Characters.ReplaceAllString(name, "-")
	if len(name) > validation.DNS1123LabelMaxLength {
		name = name[0:validation.DNS1123LabelMaxLength]
	}
	return strings.Trim(name, "-")
}

func makeJobName(name, suffix string) string {
	return fmt.Sprintf("%s-%s", name, suffix)
}

func getBoolFromPtrOrDefault(ptr *bool, val bool) bool {
	ret := val
	if ptr != nil {
		ret = *ptr
	}
	return ret
}

func getInt32FromPtrOrDefault(ptr *int32, val int32) int32 {
	ret := val
	if ptr != nil {
		ret = *ptr
	}
	return ret
}

func toServicePort(port *corev1.ContainerPort) corev1.ServicePort {
	return corev1.ServicePort{
		Name:       port.Name,
		Port:       port.ContainerPort,
		TargetPort: intstr.FromInt(int(port.ContainerPort)),
	}
}
