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
	"fmt"

	"github.com/streamnative/function-mesh/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const TestClusterName string = "test-pulsar"
const TestFunctionName string = "test-function"
const TestFunctionE2EName string = "test-function-e2e"
const TestFunctionKeyBatcherName string = "test-function-key-batcher"
const TestFunctionHPAName string = "test-function-hpa"
const TestFunctionBuiltinHPAName string = "test-function-builtin-hpa"
const TestFunctionMeshName = "test-function-mesh"
const TestSinkName string = "test-sink"
const TestSourceName string = "test-source"
const TestNameSpace string = "default"

func makeSampleObjectMeta(name string) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		Name:      name,
		Namespace: TestNameSpace,
		UID:       types.UID(fmt.Sprintf("dead-beef-%s", name)), // uid not generate automatically with fake k8s
	}
}

func makeSamplePulsarConfig() *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      TestClusterName,
			Namespace: TestNameSpace,
		},
		Data: map[string]string{
			"webServiceURL":    "http://test-pulsar-broker.default.svc.cluster.local:8080",
			"brokerServiceURL": "pulsar://test-pulsar-broker.default.svc.cluster.local:6650",
		},
	}
}

func makeFunctionSample(functionName string) *v1alpha1.Function {
	maxPending := int32(1000)
	replicas := int32(1)
	maxReplicas := int32(5)
	trueVal := true
	return &v1alpha1.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: "compute.functionmesh.io/v1alpha1",
		},
		ObjectMeta: *makeSampleObjectMeta(functionName),
		Spec: v1alpha1.FunctionSpec{
			Name:        functionName,
			ClassName:   "org.apache.pulsar.functions.api.examples.ExclamationFunction",
			Tenant:      "public",
			ClusterName: TestClusterName,
			Input: v1alpha1.InputConf{
				Topics: []string{
					"persistent://public/default/java-function-input-topic",
				},
				TypeClassName: "java.lang.String",
			},
			Output: v1alpha1.OutputConf{
				Topic:         "persistent://public/default/java-function-output-topic",
				TypeClassName: "java.lang.String",
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
				Pulsar: &v1alpha1.PulsarMessaging{
					PulsarConfig: TestClusterName,
					//AuthConfig: "test-auth",
				},
			},
			Runtime: v1alpha1.Runtime{
				Java: &v1alpha1.JavaRuntime{
					Jar:         "pulsar-functions-api-examples.jar",
					JarLocation: "public/default/nlu-test-java-function",
				},
			},
		},
	}
}

func makeFunctionSampleWithCryptoEnabled() *v1alpha1.Function {
	function := makeFunctionSample(TestFunctionE2EName)
	function.Spec.Input = v1alpha1.InputConf{
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
	}
	return function
}

func makeFunctionSampleWithKeyBasedBatcher() *v1alpha1.Function {
	function := makeFunctionSample(TestFunctionKeyBatcherName)
	function.Spec.Output = v1alpha1.OutputConf{
		Topic: "persistent://public/default/java-function-output-topic",
		ProducerConf: &v1alpha1.ProducerConfig{
			BatchBuilder: "KEY_BASED",
		},
	}
	return function
}

func makeFunctionMeshSample() *v1alpha1.FunctionMesh {
	inputTopic := "persistent://public/default/functionmesh-input-topic"
	middleTopic := "persistent://public/default/mid-topic"
	outputTopic := "persistent://public/default/functionmesh-output-topic"

	function1 := makeFunctionSample("ex1")
	function2 := makeFunctionSample("ex2")

	overwriteFunctionInputOutput(function1, inputTopic, middleTopic)
	overwriteFunctionInputOutput(function2, middleTopic, outputTopic)

	return &v1alpha1.FunctionMesh{
		TypeMeta: metav1.TypeMeta{
			Kind:       "FunctionMesh",
			APIVersion: "compute.functionmesh.io/v1alpha1",
		},
		ObjectMeta: *makeSampleObjectMeta(TestFunctionMeshName),
		Spec: v1alpha1.FunctionMeshSpec{
			Functions: []v1alpha1.FunctionSpec{
				function1.Spec,
				function2.Spec,
			},
		},
	}
}

func makeSinkSample() *v1alpha1.Sink {
	replicas := int32(1)
	maxReplicas := int32(1)
	trueVal := true
	sinkConfig := v1alpha1.NewConfig(map[string]interface{}{
		"elasticSearchUrl": "http://quickstart-es-http.default.svc.cluster.local:9200",
		"indexName":        "my_index",
		"typeName":         "doc",
		"username":         "elastic",
		"password":         "wJ757TmoXEd941kXm07Z2GW3",
	})
	return &v1alpha1.Sink{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Sink",
			APIVersion: "compute.functionmesh.io/v1alpha1",
		},
		ObjectMeta: *makeSampleObjectMeta(TestSinkName),
		Spec: v1alpha1.SinkSpec{
			Name:        TestSinkName,
			ClassName:   "org.apache.pulsar.io.elasticsearch.ElasticSearchSink",
			Tenant:      "public",
			ClusterName: TestClusterName,
			Input: v1alpha1.InputConf{
				Topics: []string{
					"persistent://public/default/input",
				},
				TypeClassName: "[B",
			},
			SinkConfig:      &sinkConfig,
			Timeout:         0,
			MaxMessageRetry: 0,
			Replicas:        &replicas,
			MaxReplicas:     &maxReplicas,
			AutoAck:         &trueVal,
			Messaging: v1alpha1.Messaging{
				Pulsar: &v1alpha1.PulsarMessaging{
					PulsarConfig: TestClusterName,
					//AuthConfig: "test-auth",
				},
			},
			Image: "streamnative/pulsar-io-elastic-search:2.10.0.0-rc10",
			Runtime: v1alpha1.Runtime{
				Java: &v1alpha1.JavaRuntime{
					Jar:         "connectors/pulsar-io-elastic-search-2.10.0.0-rc10.nar",
					JarLocation: "",
				},
			},
		},
	}
}

func makeSourceSample() *v1alpha1.Source {
	replicas := int32(1)
	maxReplicas := int32(1)
	sourceConfig := v1alpha1.NewConfig(map[string]interface{}{
		"mongodb.hosts":      "rs0/mongo-dbz-0.mongo.default.svc.cluster.local:27017,rs0/mongo-dbz-1.mongo.default.svc.cluster.local:27017,rs0/mongo-dbz-2.mongo.default.svc.cluster.local:27017",
		"mongodb.name":       "dbserver1",
		"mongodb.user":       "debezium",
		"mongodb.password":   "dbz",
		"mongodb.task.id":    "1",
		"database.whitelist": "inventory",
		"pulsar.service.url": "pulsar://test-pulsar-broker.default.svc.cluster.local:6650",
	})
	return &v1alpha1.Source{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Sink",
			APIVersion: "compute.functionmesh.io/v1alpha1",
		},
		ObjectMeta: *makeSampleObjectMeta(TestSourceName),
		Spec: v1alpha1.SourceSpec{
			Name:        TestSourceName,
			ClassName:   "org.apache.pulsar.io.debezium.mongodb.DebeziumMongoDbSource",
			Tenant:      "public",
			ClusterName: TestClusterName,
			Output: v1alpha1.OutputConf{
				Topic:         "persistent://public/default/destination",
				TypeClassName: "org.apache.pulsar.common.schema.KeyValue",
				ProducerConf: &v1alpha1.ProducerConfig{
					MaxPendingMessages:                 1000,
					MaxPendingMessagesAcrossPartitions: 50000,
					UseThreadLocalProducers:            true,
				},
			},
			SourceConfig: &sourceConfig,
			Replicas:     &replicas,
			MaxReplicas:  &maxReplicas,
			Messaging: v1alpha1.Messaging{
				Pulsar: &v1alpha1.PulsarMessaging{
					PulsarConfig: TestClusterName,
					//AuthConfig: "test-auth",
				},
			},
			Image: "streamnative/pulsar-io-debezium-mongodb:2.10.0.0-rc10",
			Runtime: v1alpha1.Runtime{
				Java: &v1alpha1.JavaRuntime{
					Jar:         "connectors/pulsar-io-debezium-mongodb-2.10.0.0-rc10.nar",
					JarLocation: "",
				},
			},
		},
	}
}

func overwriteFunctionInputOutput(function *v1alpha1.Function, input string, output string) {
	function.Spec.Input = v1alpha1.InputConf{
		Topics: []string{
			input,
		},
	}

	function.Spec.Output = v1alpha1.OutputConf{
		Topic: output,
	}
}
