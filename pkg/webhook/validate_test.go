package webhook

import (
	"strings"
	"testing"

	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
)

func TestValidateFunctionMessagingAllowsKafka(t *testing.T) {
	err := validateFunctionMessaging(&v1alpha1.FunctionSpec{
		Messaging: v1alpha1.Messaging{
			Kafka: &v1alpha1.KafkaMessaging{BootstrapServers: "kafka:9092"},
		},
	})
	if err != nil {
		t.Fatalf("expected kafka messaging to be valid, got %v", err)
	}
}

func TestValidateFunctionMessagingRejectsMissingKafkaBootstrapServers(t *testing.T) {
	err := validateFunctionMessaging(&v1alpha1.FunctionSpec{
		Messaging: v1alpha1.Messaging{
			Kafka: &v1alpha1.KafkaMessaging{},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "kafka.bootstrapServers needs to be set") {
		t.Fatalf("expected missing kafka bootstrapServers error, got %v", err)
	}
}

func TestValidateKafkaMessagingUnsupportedRejectsSourceAndSink(t *testing.T) {
	messaging := &v1alpha1.Messaging{
		Pulsar: &v1alpha1.PulsarMessaging{PulsarConfig: "pulsar-config"},
		Kafka:  &v1alpha1.KafkaMessaging{BootstrapServers: "kafka:9092"},
	}

	for _, component := range []string{"source", "sink"} {
		err := validateKafkaMessagingUnsupported(component, messaging)
		if err == nil || !strings.Contains(err.Error(), component+" does not support kafka messaging") {
			t.Fatalf("expected %s kafka unsupported error, got %v", component, err)
		}
	}
}

func TestValidateKafkaMessagingRuntimeRequiresGenericRuntime(t *testing.T) {
	err := validateKafkaMessagingRuntime(v1alpha1.Runtime{
		Java: &v1alpha1.JavaRuntime{Jar: "function.jar"},
	}, &v1alpha1.KafkaMessaging{BootstrapServers: "kafka:9092"})
	if err == nil || !strings.Contains(err.Error(), "only genericRuntime supports kafka messaging") {
		t.Fatalf("expected genericRuntime-only error, got %v", err)
	}
}

func TestValidateKafkaInputOutputAllowsKafkaTopicNames(t *testing.T) {
	errs := validateKafkaInputOutput(&v1alpha1.InputConf{
		Topics: []string{"orders"},
	}, &v1alpha1.OutputConf{
		Topic: "enriched-orders",
	}, false, false)
	if len(errs) > 0 {
		t.Fatalf("expected kafka topic names to be valid, got %v", errs)
	}
}
