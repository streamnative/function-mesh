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

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	computeapi "github.com/streamnative/function-mesh/api/compute/v1alpha2"
	apispec "github.com/streamnative/function-mesh/pkg/spec"
)

const TestClusterName string = "test-pulsar"
const TestFunctionName string = "test-function"
const TestFunctionE2EName string = "test-function-e2e"
const TestFunctionKeyBatcherName string = "test-function-key-batcher"
const TestFunctionHPAName string = "test-function-hpa"
const TestFunctionBuiltinHPAName string = "test-function-builtin-hpa"
const TestFunctionVPAName string = "test-function-vpa"
const TestFunctionStatefulJavaName string = "test-function-stateful-java"
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

func makeFunctionSample(functionName string) *computeapi.Function {
	maxPending := int32(1000)
	replicas := int32(1)
	minReplicas := int32(1)
	maxReplicas := int32(5)
	trueVal := true
	return &computeapi.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: "compute.functionmesh.io/v1alpha2",
		},
		ObjectMeta: *makeSampleObjectMeta(functionName),
		Spec: computeapi.FunctionSpec{
			Name:        functionName,
			ClassName:   "org.apache.pulsar.functions.spec.examples.ExclamationFunction",
			Tenant:      "public",
			ClusterName: TestClusterName,
			Input: apispec.InputConf{
				Topics: []string{
					"persistent://public/default/java-function-input-topic",
				},
				TypeClassName: "java.lang.String",
			},
			Output: apispec.OutputConf{
				Topic:         "persistent://public/default/java-function-output-topic",
				TypeClassName: "java.lang.String",
			},
			LogTopic:                     "persistent://public/default/logging-function-logs",
			Timeout:                      0,
			MaxMessageRetry:              0,
			ForwardSourceMessageProperty: &trueVal,
			Replicas:                     &replicas,
			MinReplicas:                  &minReplicas,
			MaxReplicas:                  &maxReplicas,
			AutoAck:                      &trueVal,
			MaxPendingAsyncRequests:      &maxPending,
			Messaging: apispec.Messaging{
				Pulsar: &apispec.PulsarMessaging{
					PulsarConfig: TestClusterName,
					//AuthConfig: "test-auth",
				},
			},
			Runtime: apispec.Runtime{
				Java: &apispec.JavaRuntime{
					Jar:         "pulsar-functions-spec-examples.jar",
					JarLocation: "public/default/nlu-test-java-function",
				},
			},
		},
	}
}

func makeFunctionSampleWithCryptoEnabled() *computeapi.Function {
	function := makeFunctionSample(TestFunctionE2EName)
	function.Spec.Input = apispec.InputConf{
		Topics: []string{
			"persistent://public/default/java-function-input-topic",
		},
		SourceSpecs: map[string]apispec.ConsumerConfig{
			"persistent://public/default/java-function-input-topic": apispec.ConsumerConfig{
				CryptoConfig: &apispec.CryptoConfig{
					CryptoKeyReaderClassName: "org.apache.pulsar.functions.spec.examples.RawFileKeyReader",
					CryptoKeyReaderConfig: map[string]string{
						"PUBLIC":  "/keys/test_ecdsa_pubkey.pem",
						"PRIVATE": "/keys/test_ecdsa_privkey.pem",
					},
					EncryptionKeys:              []string{"myapp1"},
					ProducerCryptoFailureAction: "FAIL",
					CryptoSecrets: []apispec.CryptoSecret{
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

func makeFunctionSampleWithKeyBasedBatcher() *computeapi.Function {
	function := makeFunctionSample(TestFunctionKeyBatcherName)
	function.Spec.Output = apispec.OutputConf{
		Topic: "persistent://public/default/java-function-output-topic",
		ProducerConf: &apispec.ProducerConfig{
			BatchBuilder: "KEY_BASED",
		},
	}
	return function
}

func makeFunctionMeshSample() *computeapi.FunctionMesh {
	inputTopic := "persistent://public/default/functionmesh-input-topic"
	middleTopic := "persistent://public/default/mid-topic"
	outputTopic := "persistent://public/default/functionmesh-output-topic"

	function1 := makeFunctionSample("ex1")
	function2 := makeFunctionSample("ex2")

	overwriteFunctionInputOutput(function1, inputTopic, middleTopic)
	overwriteFunctionInputOutput(function2, middleTopic, outputTopic)

	return &computeapi.FunctionMesh{
		TypeMeta: metav1.TypeMeta{
			Kind:       "FunctionMesh",
			APIVersion: "compute.functionmesh.io/v1alpha2",
		},
		ObjectMeta: *makeSampleObjectMeta(TestFunctionMeshName),
		Spec: computeapi.FunctionMeshSpec{
			Functions: []computeapi.FunctionSpec{
				function1.Spec,
				function2.Spec,
			},
		},
	}
}

func makeSinkSample() *computeapi.Sink {
	replicas := int32(1)
	minReplicas := int32(1)
	maxReplicas := int32(1)
	trueVal := true
	sinkConfig := apispec.NewConfig(map[string]interface{}{
		"elasticSearchUrl": "http://quickstart-es-http.default.svc.cluster.local:9200",
		"indexName":        "my_index",
		"typeName":         "doc",
		"username":         "elastic",
		"password":         "wJ757TmoXEd941kXm07Z2GW3",
	})
	return &computeapi.Sink{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Sink",
			APIVersion: "compute.functionmesh.io/v1alpha2",
		},
		ObjectMeta: *makeSampleObjectMeta(TestSinkName),
		Spec: computeapi.SinkSpec{
			Name:        TestSinkName,
			ClassName:   "org.apache.pulsar.io.elasticsearch.ElasticSearchSink",
			Tenant:      "public",
			ClusterName: TestClusterName,
			Input: apispec.InputConf{
				Topics: []string{
					"persistent://public/default/input",
				},
				TypeClassName: "[B",
			},
			SinkConfig:      &sinkConfig,
			Timeout:         0,
			MaxMessageRetry: 0,
			Replicas:        &replicas,
			MinReplicas:     &minReplicas,
			MaxReplicas:     &maxReplicas,
			AutoAck:         &trueVal,
			Messaging: apispec.Messaging{
				Pulsar: &apispec.PulsarMessaging{
					PulsarConfig: TestClusterName,
					//AuthConfig: "test-auth",
				},
			},
			Image: "streamnative/pulsar-io-elastic-search:2.10.0.0-rc10",
			Runtime: apispec.Runtime{
				Java: &apispec.JavaRuntime{
					Jar:         "connectors/pulsar-io-elastic-search-2.10.0.0-rc10.nar",
					JarLocation: "",
				},
			},
		},
	}
}

func makeSourceSample() *computeapi.Source {
	replicas := int32(1)
	minReplicas := int32(1)
	maxReplicas := int32(1)
	sourceConfig := apispec.NewConfig(map[string]interface{}{
		"mongodb.hosts":      "rs0/mongo-dbz-0.mongo.default.svc.cluster.local:27017,rs0/mongo-dbz-1.mongo.default.svc.cluster.local:27017,rs0/mongo-dbz-2.mongo.default.svc.cluster.local:27017",
		"mongodb.name":       "dbserver1",
		"mongodb.user":       "debezium",
		"mongodb.password":   "dbz",
		"mongodb.task.id":    "1",
		"database.whitelist": "inventory",
		"pulsar.service.url": "pulsar://test-pulsar-broker.default.svc.cluster.local:6650",
	})
	return &computeapi.Source{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Sink",
			APIVersion: "compute.functionmesh.io/v1alpha2",
		},
		ObjectMeta: *makeSampleObjectMeta(TestSourceName),
		Spec: computeapi.SourceSpec{
			Name:        TestSourceName,
			ClassName:   "org.apache.pulsar.io.debezium.mongodb.DebeziumMongoDbSource",
			Tenant:      "public",
			ClusterName: TestClusterName,
			Output: apispec.OutputConf{
				Topic:         "persistent://public/default/destination",
				TypeClassName: "org.apache.pulsar.common.schema.KeyValue",
				ProducerConf: &apispec.ProducerConfig{
					MaxPendingMessages:                 1000,
					MaxPendingMessagesAcrossPartitions: 50000,
					UseThreadLocalProducers:            true,
				},
			},
			SourceConfig: &sourceConfig,
			Replicas:     &replicas,
			MinReplicas:  &minReplicas,
			MaxReplicas:  &maxReplicas,
			Messaging: apispec.Messaging{
				Pulsar: &apispec.PulsarMessaging{
					PulsarConfig: TestClusterName,
					//AuthConfig: "test-auth",
				},
			},
			Image: "streamnative/pulsar-io-debezium-mongodb:2.10.0.0-rc10",
			Runtime: apispec.Runtime{
				Java: &apispec.JavaRuntime{
					Jar:         "connectors/pulsar-io-debezium-mongodb-2.10.0.0-rc10.nar",
					JarLocation: "",
				},
			},
		},
	}
}

func overwriteFunctionInputOutput(function *computeapi.Function, input string, output string) {
	function.Spec.Input = apispec.InputConf{
		Topics: []string{
			input,
		},
	}

	function.Spec.Output = apispec.OutputConf{
		Topic: output,
	}
}
