package spec

import (
	"encoding/json"

	"github.com/streamnative/mesh-operator/api/v1alpha1"
	"github.com/streamnative/mesh-operator/controllers/proto"
)

func convertFunctionDetails(function *v1alpha1.Function) *proto.FunctionDetails {
	return &proto.FunctionDetails{
		// TODO: default tenant value
		Tenant:               "public",
		Namespace:            function.Namespace,
		Name:                 function.Spec.Name,
		ClassName:            function.Spec.ClassName,
		LogTopic:             "",
		ProcessingGuarantees: 0,
		UserConfig:           "",
		SecretsMap:           "",
		Runtime:              proto.FunctionDetails_JAVA,
		AutoAck:              false,
		Parallelism:          function.Spec.Replicas,
		Source:               generateFunctionInputSpec(function.Spec.Sources, function.Spec.SourceType),
		Sink:                 generateFunctionOutputSpec(function.Spec.Sink, function.Spec.SinkType),
		Resources: &proto.Resources{
			Cpu:  1,
			Ram:  102400,
			Disk: 102400,
		},
		PackageUrl:           "",
		RetryDetails:         nil,
		RuntimeFlags:         "",
		ComponentType:        proto.FunctionDetails_FUNCTION,
		CustomRuntimeOptions: "",
		Builtin:              "",
		RetainOrdering:       false,
		RetainKeyOrdering:    false,
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

func generateFunctionOutputSpec(topic, sinkTypeClass string) *proto.SinkSpec {
	return &proto.SinkSpec{
		Topic:         topic,
		TypeClassName: sinkTypeClass,
	}
}

func convertSourceDetails(source *v1alpha1.Source) *proto.FunctionDetails {
	return &proto.FunctionDetails{
		Tenant:               "public",
		Namespace:            source.Namespace,
		Name:                 source.Name,
		ClassName:            "org.apache.pulsar.functions.api.utils.IdentityFunction", // TODO
		LogTopic:             "",
		ProcessingGuarantees: 0,
		UserConfig:           "",
		SecretsMap:           "",
		Runtime:              proto.FunctionDetails_JAVA,
		AutoAck:              true,
		Parallelism:          source.Spec.Replicas,
		Source:               generateSourceInputSpec(source),
		Sink:                 generateSourceOutputSpec(source.Spec.Destination, source.Spec.SinkType),
		Resources: &proto.Resources{
			Cpu:  1,
			Ram:  102400,
			Disk: 102400,
		},
		PackageUrl:           "",
		RetryDetails:         nil,
		RuntimeFlags:         "",
		ComponentType:        proto.FunctionDetails_SOURCE,
		CustomRuntimeOptions: "",
		Builtin:              "",
		RetainOrdering:       false,
		RetainKeyOrdering:    false,
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
		TimeoutMs:                    0,
		Builtin:                      "",
		SubscriptionName:             "",
		CleanupSubscription:          false,
		SubscriptionPosition:         0,
		NegativeAckRedeliveryDelayMs: 0,
	}
}

func generateSourceOutputSpec(topic, sinkTypeClass string) *proto.SinkSpec {
	return &proto.SinkSpec{
		Topic:         topic,
		TypeClassName: sinkTypeClass, //"java.lang.String", // TODO resolve it
	}
}

func convertSinkDetails(sink *v1alpha1.Sink) *proto.FunctionDetails {
	return &proto.FunctionDetails{
		Tenant:               "public",
		Namespace:            sink.Namespace,
		Name:                 sink.Name,
		ClassName:            "org.apache.pulsar.functions.api.utils.IdentityFunction", // TODO
		LogTopic:             "",
		ProcessingGuarantees: 0,
		UserConfig:           "",
		SecretsMap:           "",
		Runtime:              proto.FunctionDetails_JAVA,
		AutoAck:              true,
		Parallelism:          sink.Spec.Replicas,
		Source:               generateSinkInputSpec(sink.Spec.Inputs, sink.Spec.SourceType),
		Sink:                 generateSinkOutputSpec(sink),
		Resources: &proto.Resources{
			Cpu:  1,
			Ram:  102400,
			Disk: 102400,
		},
		PackageUrl:           "",
		RetryDetails:         nil,
		RuntimeFlags:         "",
		ComponentType:        proto.FunctionDetails_SINK,
		CustomRuntimeOptions: "",
		Builtin:              "",
		RetainOrdering:       false,
		RetainKeyOrdering:    false,
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
		ForwardSourceMessageProperty: false,
		SchemaProperties:             nil,
		ConsumerProperties:           nil,
	}
}
