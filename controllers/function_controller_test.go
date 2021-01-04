package controllers

import (
	"context"
	"github.com/streamnative/function-mesh/controllers/spec"
	v1 "k8s.io/api/core/v1"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/streamnative/function-mesh/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Function Controller", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1
	var objectMeta = metav1.ObjectMeta{
		Name:      "test-function",
		Namespace: "default",
		UID:       "dead-beef",
	}
	pulsarConfig := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pulsar",
			Namespace: "default",
		},
		Data: map[string]string{
			"webServiceURL":    "http://test-pulsar-broker.default.svc.cluster.local:8080",
			"brokerServiceURL": "pulsar://test-pulsar-broker.default.svc.cluster.local:6650",
		},
	}
	var pulsar = &v1alpha1.PulsarMessaging{
		PulsarConfig: pulsarConfig.Name,
		//AuthConfig: "test-auth",
	}

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
		Expect(k8sClient.Create(context.Background(), pulsarConfig)).Should(Succeed())
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
		Expect(k8sClient.Delete(context.Background(), pulsarConfig)).Should(Succeed())
	})

	Context("Function Item", func() {
		It("Should create successfully", func() {
			maxPending := int32(1000)
			replicas := int32(1)
			maxReplicas := int32(5)
			trueVal := true
			function := &v1alpha1.Function{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Function",
					APIVersion: "cloud.streamnative.io/v1alpha1",
				},
				ObjectMeta: objectMeta,
				Spec: v1alpha1.FunctionSpec{
					Name:        objectMeta.Name,
					ClassName:   "org.apache.pulsar.functions.api.examples.ExclamationFunction",
					Tenant:      "public",
					ClusterName: "test-pulsar",
					SourceType:  "java.lang.String",
					SinkType:    "java.lang.String",
					Input: v1alpha1.InputConf{
						Topics: []string{
							"persistent://public/default/java-function-input-topic",
						},
					},
					Output: v1alpha1.OutputConf{
						Topic: "persistent://public/default/java-function-output-topic",
					},
					LogTopic:                     "persistent://public/default/logging-function-logs",
					Timeout:                      0,
					MaxMessageRetry:              0,
					ForwardSourceMessageProperty: &trueVal,
					Replicas:                     &replicas,
					MaxReplicas:                  &maxReplicas,
					AutoAck:                      &trueVal,
					MaxPendingAsyncRequests:      &maxPending,
					Messaging: v1alpha1.Messaging{
						Pulsar: pulsar,
					},
					Runtime: v1alpha1.Runtime{
						Java: &v1alpha1.JavaRuntime{
							Jar:         "pulsar-functions-api-examples.jar",
							JarLocation: "public/default/nlu-test-java-function",
						},
					},
				},
			}
			if function.Status.Conditions == nil {
				function.Status.Conditions = make(map[v1alpha1.Component]v1alpha1.ResourceCondition)
			}
			statefulSet := spec.MakeFunctionStatefulSet(function)
			Expect(k8sClient.Create(context.Background(), statefulSet)).Should(Succeed())
		})
	})
})
