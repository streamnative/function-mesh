package controllers

import (
	"github.com/streamnative/function-mesh/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const TestClusterName string = "test-pulsar"
const TestFunctionName string = "test-function"
const TestFunctionMeshName = "test-function-mesh"
const TestSinkName string = "test-sink"
const TestSourceName string = "test-source"
const TestNameSpace string = "default"

func makeSampleObjectMeta(name string) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		Name:      name,
		Namespace: TestNameSpace,
		UID:       "dead-beef", // uid not generate automatically with fake k8s
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
			APIVersion: "cloud.streamnative.io/v1alpha1",
		},
		ObjectMeta: *makeSampleObjectMeta(functionName),
		Spec: v1alpha1.FunctionSpec{
			Name:        functionName,
			ClassName:   "org.apache.pulsar.functions.api.examples.ExclamationFunction",
			Tenant:      "public",
			ClusterName: TestClusterName,
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
	function := makeFunctionSample(TestFunctionName)
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
			APIVersion: "cloud.streamnative.io/v1alpha1",
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
	return &v1alpha1.Sink{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Sink",
			APIVersion: "cloud.streamnative.io/v1alpha1",
		},
		ObjectMeta: *makeSampleObjectMeta(TestSinkName),
		Spec: v1alpha1.SinkSpec{
			Name:        TestSinkName,
			ClassName:   "org.apache.pulsar.io.elasticsearch.ElasticSearchSink",
			Tenant:      "public",
			ClusterName: TestClusterName,
			SourceType:  "[B",
			SinkType:    "[B",
			Input: v1alpha1.InputConf{
				Topics: []string{
					"persistent://public/default/input",
				},
			},
			SinkConfig: map[string]string{
				"elasticSearchUrl": "http://quickstart-es-http.default.svc.cluster.local:9200",
				"indexName":        "my_index",
				"typeName":         "doc",
				"username":         "elastic",
				"password":         "wJ757TmoXEd941kXm07Z2GW3",
			},
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
			Runtime: v1alpha1.Runtime{
				Java: &v1alpha1.JavaRuntime{
					Jar:         "connectors/pulsar-io-elastic-search-2.7.0-rc-pm-3.nar",
					JarLocation: "",
				},
			},
		},
	}
}

func makeSourceSample() *v1alpha1.Source {
	replicas := int32(1)
	maxReplicas := int32(1)
	return &v1alpha1.Source{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Sink",
			APIVersion: "cloud.streamnative.io/v1alpha1",
		},
		ObjectMeta: *makeSampleObjectMeta(TestSourceName),
		Spec: v1alpha1.SourceSpec{
			Name:        TestSourceName,
			ClassName:   "org.apache.pulsar.io.debezium.mongodb.DebeziumMongoDbSource",
			Tenant:      "public",
			ClusterName: TestClusterName,
			SourceType:  "org.apache.pulsar.common.schema.KeyValue",
			SinkType:    "org.apache.pulsar.common.schema.KeyValue",
			Output: v1alpha1.OutputConf{
				Topic: "persistent://public/default/destination",
				ProducerConf: &v1alpha1.ProducerConfig{
					MaxPendingMessages:                 1000,
					MaxPendingMessagesAcrossPartitions: 50000,
					UseThreadLocalProducers:            true,
				},
			},
			SourceConfig: map[string]string{
				"mongodb.hosts":      "rs0/mongo-dbz-0.mongo.default.svc.cluster.local:27017,rs0/mongo-dbz-1.mongo.default.svc.cluster.local:27017,rs0/mongo-dbz-2.mongo.default.svc.cluster.local:27017",
				"mongodb.name":       "dbserver1",
				"mongodb.user":       "debezium",
				"mongodb.password":   "dbz",
				"mongodb.task.id":    "1",
				"database.whitelist": "inventory",
				"pulsar.service.url": "pulsar://test-pulsar-broker.default.svc.cluster.local:6650",
			},
			Replicas:    &replicas,
			MaxReplicas: &maxReplicas,
			Messaging: v1alpha1.Messaging{
				Pulsar: &v1alpha1.PulsarMessaging{
					PulsarConfig: TestClusterName,
					//AuthConfig: "test-auth",
				},
			},
			Runtime: v1alpha1.Runtime{
				Java: &v1alpha1.JavaRuntime{
					Jar:         "connectors/pulsar-io-elastic-search-2.7.0-rc-pm-3.nar",
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
