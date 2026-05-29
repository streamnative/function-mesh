package webhook

import (
	"context"
	"strings"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
)

func TestSourceWebhookValidateUpdateRejectsKafkaMessaging(t *testing.T) {
	ctx := admission.NewContextWithRequest(context.Background(), admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Kind: metav1.GroupVersionKind{Kind: sourceKind},
		},
	})

	_, err := (&SourceWebhook{}).ValidateUpdate(ctx, &v1alpha1.Source{}, &v1alpha1.Source{
		ObjectMeta: metav1.ObjectMeta{Name: "source-kafka"},
		Spec: v1alpha1.SourceSpec{
			Messaging: v1alpha1.Messaging{
				Kafka: &v1alpha1.KafkaMessaging{BootstrapServers: "kafka:9092"},
			},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "source does not support kafka messaging") {
		t.Fatalf("expected source kafka unsupported error, got %v", err)
	}
}

func TestSinkWebhookValidateUpdateRejectsKafkaMessaging(t *testing.T) {
	ctx := admission.NewContextWithRequest(context.Background(), admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Kind: metav1.GroupVersionKind{Kind: sinkKind},
		},
	})

	_, err := (&SinkWebhook{}).ValidateUpdate(ctx, &v1alpha1.Sink{}, &v1alpha1.Sink{
		ObjectMeta: metav1.ObjectMeta{Name: "sink-kafka"},
		Spec: v1alpha1.SinkSpec{
			Messaging: v1alpha1.Messaging{
				Kafka: &v1alpha1.KafkaMessaging{BootstrapServers: "kafka:9092"},
			},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "sink does not support kafka messaging") {
		t.Fatalf("expected sink kafka unsupported error, got %v", err)
	}
}
