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

	"github.com/streamnative/function-mesh/api/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/proto"
)

func convertFunctionDetails(function *v1alpha1.Function) *proto.FunctionDetails {
	return &proto.FunctionDetails{
		Tenant:               function.Spec.Tenant,
		Namespace:            function.Namespace,
		Name:                 function.Spec.Name,
		ClassName:            function.Spec.ClassName,
		LogTopic:             function.Spec.LogTopic,
		ProcessingGuarantees: convertProcessingGuarantee(function.Spec.ProcessingGuarantee),
		UserConfig:           getUserConfig(function.Spec.FuncConfig),
		SecretsMap:           marshalSecretsMap(function.Spec.SecretsMap),
		Runtime:              proto.FunctionDetails_JAVA,
		AutoAck:              *function.Spec.AutoAck,
		Parallelism:          *function.Spec.Replicas,
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

	for topicName, conf := range sourceConf.CustomSchemaSources {
		consumerConf := unmarshalConsumerConfig(conf)
		inputSpecs[topicName] = &proto.ConsumerSpec{
			SchemaType:         consumerConf.SchemaType,
			IsRegexPattern:     false,
			SchemaProperties:   consumerConf.SchemaProperties,
			ConsumerProperties: consumerConf.ConsumerProperties,
		}
	}

	for topicName, conf := range sourceConf.SourceSpecs {
		inputSpecs[topicName] = &proto.ConsumerSpec{
			SchemaType:         conf.SchemaType,
			SerdeClassName:     conf.SerdeClassName,
			IsRegexPattern:     conf.IsRegexPattern,
			ReceiverQueueSize:  &proto.ConsumerSpec_ReceiverQueueSize{Value: conf.ReceiverQueueSize},
			SchemaProperties:   conf.SchemaProperties,
			ConsumerProperties: conf.ConsumerProperties,
		}
	}

	return inputSpecs
}

func generateFunctionInputSpec(function *v1alpha1.Function) *proto.SourceSpec {
	inputSpecs := generateInputSpec(function.Spec.Input)

	return &proto.SourceSpec{
		ClassName:                    "",
		Configs:                      "",
		TypeClassName:                function.Spec.SourceType,
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
		TypeClassName:                function.Spec.SinkType,
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
		}

		sinkSpec.ProducerSpec = producerConfig
	}

	return sinkSpec
}

func convertSourceDetails(source *v1alpha1.Source) *proto.FunctionDetails {
	return &proto.FunctionDetails{
		Tenant:               source.Spec.Tenant,
		Namespace:            source.Namespace,
		Name:                 source.Name,
		ClassName:            "org.apache.pulsar.functions.api.utils.IdentityFunction",
		ProcessingGuarantees: convertProcessingGuarantee(source.Spec.ProcessingGuarantee),
		UserConfig:           getUserConfig(source.Spec.SourceConfig),
		SecretsMap:           marshalSecretsMap(source.Spec.SecretsMap),
		Runtime:              proto.FunctionDetails_JAVA,
		AutoAck:              true,
		Parallelism:          *source.Spec.Replicas,
		Source:               generateSourceInputSpec(source),
		Sink:                 generateSourceOutputSpec(source),
		Resources:            generateResource(source.Spec.Resources.Requests),
		RuntimeFlags:         source.Spec.RuntimeFlags,
		ComponentType:        proto.FunctionDetails_SOURCE,
	}
}

func generateSourceInputSpec(source *v1alpha1.Source) *proto.SourceSpec {
	configs, _ := json.Marshal(source.Spec.SourceConfig)
	return &proto.SourceSpec{
		ClassName:     source.Spec.ClassName,
		Configs:       string(configs), // TODO handle batch source
		TypeClassName: source.Spec.SourceType,
	}
}

func generateSourceOutputSpec(source *v1alpha1.Source) *proto.SinkSpec {
	return &proto.SinkSpec{
		TypeClassName: source.Spec.SinkType,
		Topic:         source.Spec.Output.Topic,
		ProducerSpec: &proto.ProducerSpec{
			MaxPendingMessages:                 source.Spec.Output.ProducerConf.MaxPendingMessages,
			MaxPendingMessagesAcrossPartitions: source.Spec.Output.ProducerConf.MaxPendingMessagesAcrossPartitions,
			UseThreadLocalProducers:            source.Spec.Output.ProducerConf.UseThreadLocalProducers,
		},
		SerDeClassName: source.Spec.Output.SinkSerdeClassName,
		SchemaType:     source.Spec.Output.SinkSchemaType,
	}
}

func convertSinkDetails(sink *v1alpha1.Sink) *proto.FunctionDetails {
	return &proto.FunctionDetails{
		Tenant:               sink.Spec.Tenant,
		Namespace:            sink.Namespace,
		Name:                 sink.Name,
		ClassName:            "org.apache.pulsar.functions.api.utils.IdentityFunction",
		ProcessingGuarantees: convertProcessingGuarantee(sink.Spec.ProcessingGuarantee),
		SecretsMap:           marshalSecretsMap(sink.Spec.SecretsMap),
		Runtime:              proto.FunctionDetails_JAVA,
		AutoAck:              *sink.Spec.AutoAck,
		Parallelism:          *sink.Spec.Replicas,
		Source:               generateSinkInputSpec(sink),
		Sink:                 generateSinkOutputSpec(sink),
		Resources:            generateResource(sink.Spec.Resources.Requests),
		RetryDetails:         generateRetryDetails(sink.Spec.MaxMessageRetry, sink.Spec.DeadLetterTopic),
		RuntimeFlags:         sink.Spec.RuntimeFlags,
		ComponentType:        proto.FunctionDetails_SINK,
	}
}

func generateSinkInputSpec(sink *v1alpha1.Sink) *proto.SourceSpec {
	inputSpecs := generateInputSpec(sink.Spec.Input)

	return &proto.SourceSpec{
		TypeClassName:                sink.Spec.SourceType,
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
	} else {
		return proto.SubscriptionType_SHARED
	}
}

func generateSinkOutputSpec(sink *v1alpha1.Sink) *proto.SinkSpec {
	configs, _ := json.Marshal(sink.Spec.SinkConfig)
	return &proto.SinkSpec{
		ClassName:     sink.Spec.ClassName,
		Configs:       string(configs),
		TypeClassName: sink.Spec.SinkType,
	}
}

func marshalSecretsMap(secrets map[string]v1alpha1.SecretRef) string {
	// validated in admission webhook
	bytes, _ := json.Marshal(secrets)
	return string(bytes)
}

func unmarshalConsumerConfig(conf string) v1alpha1.ConsumerConfig {
	var config v1alpha1.ConsumerConfig
	// TODO: check unmarshel error in admission hook
	json.Unmarshal([]byte(conf), &config)
	return config
}
