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
		ProcessingGuarantees: getProcessingGuarantee(function.Spec.ProcessingGuarantee),
		UserConfig:           getUserConfig(function.Spec.FuncConfig),
		SecretsMap:           marshalSecretsMap(function.Spec.SecretsMap),
		Runtime:              proto.FunctionDetails_JAVA,
		AutoAck:              *function.Spec.AutoAck,
		Parallelism:          *function.Spec.Replicas,
		Source:               generateFunctionInputSpec(function.Spec.Sources, function.Spec.SourceType),
		Sink:                 generateFunctionOutputSpec(function.Spec.Sink, function.Spec.SinkType, *function.Spec.ForwardSourceMessageProperty),
		Resources:            generateResource(function.Spec.Resources),
		PackageUrl:           "",
		RetryDetails:         generateRetryDetails(function.Spec.MaxMessageRetry, function.Spec.DeadLetterTopic),
		RuntimeFlags:         "",
		ComponentType:        proto.FunctionDetails_FUNCTION,
		CustomRuntimeOptions: "",
		Builtin:              "",
		RetainOrdering:       function.Spec.RetainOrdering,
		RetainKeyOrdering:    function.Spec.RetainKeyOrdering,
	}
}

func generateFunctionInputSpec(sources []string, sourceTypeClass string) *proto.SourceSpec {
	inputSpecs := make(map[string]*proto.ConsumerSpec)

	for _, source := range sources {
		inputSpecs[source] = &proto.ConsumerSpec{
			IsRegexPattern: false,
		}
	}

	return &proto.SourceSpec{
		InputSpecs:           inputSpecs,
		SubscriptionType:     proto.SubscriptionType_SHARED,
		SubscriptionPosition: proto.SubscriptionPosition_EARLIEST,
		TypeClassName:        sourceTypeClass,
		CleanupSubscription:  true,
	}
}

func generateFunctionOutputSpec(topic, sinkTypeClass string, forwardMessageProperty bool) *proto.SinkSpec {
	return &proto.SinkSpec{
		Topic:                        topic,
		TypeClassName:                sinkTypeClass,
		ForwardSourceMessageProperty: forwardMessageProperty,
	}
}

func convertSourceDetails(source *v1alpha1.Source) *proto.FunctionDetails {
	return &proto.FunctionDetails{
		Tenant:               source.Spec.Tenant,
		Namespace:            source.Namespace,
		Name:                 source.Name,
		ClassName:            "org.apache.pulsar.functions.api.utils.IdentityFunction",
		LogTopic:             source.Spec.LogTopic,
		ProcessingGuarantees: getProcessingGuarantee(source.Spec.ProcessingGuarantee),
		UserConfig:           getUserConfig(source.Spec.SourceConfig),
		SecretsMap:           marshalSecretsMap(source.Spec.SecretsMap),
		Runtime:              proto.FunctionDetails_JAVA,
		AutoAck:              *source.Spec.AutoAck,
		Parallelism:          *source.Spec.Replicas,
		Source:               generateSourceInputSpec(source),
		Sink:                 generateSourceOutputSpec(source.Spec.Sink, source.Spec.SinkType, *source.Spec.ForwardSourceMessageProperty),
		Resources:            generateResource(source.Spec.Resources),
		PackageUrl:           "",
		RetryDetails:         generateRetryDetails(source.Spec.MaxMessageRetry, source.Spec.DeadLetterTopic),
		RuntimeFlags:         "",
		ComponentType:        proto.FunctionDetails_SOURCE,
		CustomRuntimeOptions: "",
		Builtin:              "",
		RetainOrdering:       source.Spec.RetainOrdering,
		RetainKeyOrdering:    source.Spec.RetainKeyOrdering,
	}
}

func generateSourceInputSpec(source *v1alpha1.Source) *proto.SourceSpec {
	configs, _ := json.Marshal(source.Spec.SourceConfig)
	return &proto.SourceSpec{
		ClassName:                    source.Spec.ClassName,
		Configs:                      string(configs), // TODO handle batch source
		TypeClassName:                source.Spec.SourceType,
		SubscriptionType:             0,
		InputSpecs:                   nil,
		TimeoutMs:                    uint64(source.Spec.Timeout),
		Builtin:                      "",
		SubscriptionName:             "",
		CleanupSubscription:          false,
		SubscriptionPosition:         0,
		NegativeAckRedeliveryDelayMs: 0,
	}
}

func generateSourceOutputSpec(topic, sinkTypeClass string, forwardMessageProperty bool) *proto.SinkSpec {
	return &proto.SinkSpec{
		Topic:                        topic,
		TypeClassName:                sinkTypeClass,
		ForwardSourceMessageProperty: forwardMessageProperty,
	}
}

func convertSinkDetails(sink *v1alpha1.Sink) *proto.FunctionDetails {
	return &proto.FunctionDetails{
		Tenant:               "public",
		Namespace:            sink.Namespace,
		Name:                 sink.Name,
		ClassName:            "org.apache.pulsar.functions.api.utils.IdentityFunction",
		LogTopic:             sink.Spec.LogTopic,
		ProcessingGuarantees: getProcessingGuarantee(sink.Spec.ProcessingGuarantee),
		UserConfig:           getUserConfig(sink.Spec.SinkConfig),
		SecretsMap:           marshalSecretsMap(sink.Spec.SecretsMap),
		Runtime:              proto.FunctionDetails_JAVA,
		AutoAck:              *sink.Spec.AutoAck,
		Parallelism:          *sink.Spec.Replicas,
		Source:               generateSinkInputSpec(sink.Spec.Sources, sink.Spec.SourceType),
		Sink:                 generateSinkOutputSpec(sink),
		Resources:            generateResource(sink.Spec.Resources),
		PackageUrl:           "",
		RetryDetails:         generateRetryDetails(sink.Spec.MaxMessageRetry, sink.Spec.DeadLetterTopic),
		RuntimeFlags:         "",
		ComponentType:        proto.FunctionDetails_SINK,
		CustomRuntimeOptions: "",
		Builtin:              "",
		RetainOrdering:       sink.Spec.RetainOrdering,
		RetainKeyOrdering:    sink.Spec.RetainKeyOrdering,
	}
}

func generateSinkInputSpec(sources []string, sourceTypeClass string) *proto.SourceSpec {
	inputSpecs := make(map[string]*proto.ConsumerSpec)

	for _, source := range sources {
		inputSpecs[source] = &proto.ConsumerSpec{
			IsRegexPattern: false,
		}
	}

	return &proto.SourceSpec{
		InputSpecs:           inputSpecs,
		SubscriptionType:     proto.SubscriptionType_SHARED,
		SubscriptionPosition: proto.SubscriptionPosition_EARLIEST,
		TypeClassName:        sourceTypeClass,
		CleanupSubscription:  true,
	}
}

func generateSinkOutputSpec(sink *v1alpha1.Sink) *proto.SinkSpec {
	configs, _ := json.Marshal(sink.Spec.SinkConfig)
	return &proto.SinkSpec{
		ClassName:                    sink.Spec.ClassName,
		Configs:                      string(configs),
		TypeClassName:                sink.Spec.SinkType,
		Topic:                        "",
		ProducerSpec:                 nil,
		SerDeClassName:               "",
		Builtin:                      "",
		SchemaType:                   "",
		ForwardSourceMessageProperty: *sink.Spec.ForwardSourceMessageProperty,
		SchemaProperties:             nil,
		ConsumerProperties:           nil,
	}
}

func marshalSecretsMap(secrets map[string]v1alpha1.SecretRef) string {
	// validated in admission webhook
	bytes, _ := json.Marshal(secrets)
	return string(bytes)
}
