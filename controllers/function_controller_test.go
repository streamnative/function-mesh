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

package controllers

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/streamnative/function-mesh/api/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/spec"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Function Controller", func() {
	var objectMeta = metav1.ObjectMeta{
		Name:      "test-function",
		Namespace: "default",
		UID:       "dead-beef", // uid not generate automatically with fake k8s
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

	Context("Simple Function Item", func() {
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

		It("Should create pulsar configmap successfully", func() {
			Expect(k8sClient.Create(context.Background(), pulsarConfig)).Should(Succeed())
		})

		It("Should create successfully", func() {
			Expect(k8sClient.Create(context.Background(), statefulSet)).Should(Succeed())
		})

		It("Should delete successfully", func() {
			Expect(k8sClient.Delete(context.Background(), statefulSet)).Should(Succeed())
		})

		It("Should delete pulsar configmap successfully", func() {
			Expect(k8sClient.Delete(context.Background(), pulsarConfig)).Should(Succeed())
		})
	})

})
var _ = Describe("Function Controller (E2E)", func() {
	var objectMeta = metav1.ObjectMeta{
		Name:      "test-function",
		Namespace: "default",
		UID:       "dead-beef", // uid not generate automatically with fake k8s
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

	Context("Function With E2E Crypto Item", func() {
		maxPending := int32(1000)
		replicas := int32(1)
		maxReplicas := int32(5)
		trueVal := true

		cryptosecrets := &v1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "java-function-crypto-sample-crypto-secret",
				Namespace: "default",
				UID:       "dead-beef-secret",
			},
			Data: map[string][]byte{
				"test_ecdsa_privkey.pem": []byte{0x00, 0x01, 0x02},
				"test_ecdsa_pubkey.pem":  []byte{0x02, 0x01, 0x00},
			},
			Type: "Opaque",
		}

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
					SourceSpecs: map[string]v1alpha1.ConsumerConfig{
						"persistent://public/default/java-function-input-topic": v1alpha1.ConsumerConfig{
							CryptoConfig: &v1alpha1.CryptoConfig{
								CryptoKeyReaderClassName: "org.apache.pulsar.functions.api.examples.RawFileKeyReader",
								CryptoKeyReaderConfig: map[string]string{
									"PUBLIC":  "/keys/test_ecdsa_pubkey.pem",
									"PRIVATE": "/keys/test_ecdsa_privkey.pem",
								},
								EncryptionKeys:              []string{"myapp1"},
								ProducerCryptoFailureAction: "FAIL",
								CryptoSecrets: []v1alpha1.CryptoSecret{
									{
										SecretName: "java-function-crypto-sample-crypto-secret",
										SecretKey:  "test_ecdsa_privkey.pem",
										AsVolume:   "/keys/test_ecdsa_privkey.pem",
									},
									{
										SecretName: "java-function-crypto-sample-crypto-secret",
										SecretKey:  "test_ecdsa_pubkey.pem",
										AsVolume:   "/keys/test_ecdsa_pubkey.pem",
									},
								},
							},
						},
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

		It("Should create pulsar configmap successfully", func() {
			Expect(k8sClient.Create(context.Background(), pulsarConfig)).Should(Succeed())
		})

		It("Should create crypto secret successfully", func() {
			Expect(k8sClient.Create(context.Background(), cryptosecrets)).Should(Succeed())
		})

		It("Should create successfully", func() {
			Expect(k8sClient.Create(context.Background(), statefulSet)).Should(Succeed())
		})

		It("Should delete successfully", func() {
			Expect(k8sClient.Delete(context.Background(), statefulSet)).Should(Succeed())
		})

		It("Should delete crypto secret successfully", func() {
			Expect(k8sClient.Delete(context.Background(), cryptosecrets)).Should(Succeed())
		})

		It("Should delete pulsar configmap successfully", func() {
			Expect(k8sClient.Delete(context.Background(), pulsarConfig)).Should(Succeed())
		})

	})

})
