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
	"testing"

	"github.com/streamnative/function-mesh/api/compute/v1alpha1"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetValFromPtrOrDefault(t *testing.T) {
	boolVal := true
	boolPtr := &boolVal
	assert.Equal(t, getBoolFromPtrOrDefault(boolPtr, false), boolVal)
	assert.Equal(t, getBoolFromPtrOrDefault(nil, boolVal), boolVal)

	var int32Val int32 = 100
	int32Ptr := &int32Val
	assert.Equal(t, getInt32FromPtrOrDefault(int32Ptr, 200), int32Val)
	assert.Equal(t, getInt32FromPtrOrDefault(nil, int32Val), int32Val)
}

func TestMarshalSecretsMap(t *testing.T) {
	secrets := map[string]v1alpha1.SecretRef{
		"foo": {
			Path: "path",
		},
	}
	marshaledSecrets := marshalSecretsMap(secrets)
	assert.Equal(t, marshaledSecrets, `{"foo":{"path":"path"}}`)

	marshaledSecretsNil := marshalSecretsMap(nil)
	assert.Equal(t, marshaledSecretsNil, `{}`)
}

func TestBatchSource(t *testing.T) {
	source := &v1alpha1.Source{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Sink",
			APIVersion: "compute.functionmesh.io/v1alpha1",
		},
		ObjectMeta: *makeSampleObjectMeta("test-source"),
		Spec: v1alpha1.SourceSpec{
			Name:        "test-suorce",
			ClassName:   "org.apache.pulsar.ecosystem.io.bigquery.BigQuerySource",
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
			BatchSourceConfig: &v1alpha1.BatchSourceConfig{
				DiscoveryTriggererClassName: "test-trigger-class",
				DiscoveryTriggererConfig: &v1alpha1.Config{
					Data: map[string]interface{}{
						"test-key": "test-value",
					},
				},
			},
			SourceConfig: &v1alpha1.Config{
				Data: map[string]interface{}{
					"tableName": "test-table",
				},
			},
			Messaging: v1alpha1.Messaging{
				Pulsar: &v1alpha1.PulsarMessaging{
					PulsarConfig: TestClusterName,
					//AuthConfig: "test-auth",
				},
			},
			Image: "test-image",
			Runtime: v1alpha1.Runtime{
				Java: &v1alpha1.JavaRuntime{
					Jar:         "connectors/test.jar",
					JarLocation: "",
				},
			},
		},
	}
	sourceSpec := generateSourceInputSpec(source)
	assert.Equal(t, v1alpha1.BatchSourceClass, sourceSpec.ClassName)
	assert.Equal(t, `{"__BATCHSOURCECLASSNAME__":"org.apache.pulsar.ecosystem.io.bigquery.BigQuerySource","__BATCHSOURCECONFIGS__":"{\"discoveryTriggererClassName\":\"test-trigger-class\",\"discoveryTriggererConfig\":{\"test-key\":\"test-value\"}}","tableName":"test-table"}`, sourceSpec.Configs)
}

func TestGenerateFunctionInputSpecWithConnector(t *testing.T) {
	function := makeFunctionSample("connector-source")
	function.Spec.Input.Topics = nil
	function.Spec.Input.CustomSerdeSources = nil
	function.Spec.Input.CustomSchemaSources = nil
	function.Spec.Input.SourceSpecs = nil
	configs := &v1alpha1.Config{
		Data: map[string]interface{}{
			"bootstrapServers": "kafka:9092",
			"groupId":          "test-consumer",
		},
	}
	function.Spec.SourceConfig = &v1alpha1.SourceConnectorSpec{
		Archive:   "builtin://kafka",
		ClassName: "org.apache.pulsar.io.kafka.KafkaSource",
		Configs:   configs,
	}

	sourceSpec := generateFunctionInputSpec(function)
	assert.Equal(t, "kafka", sourceSpec.Builtin)
	assert.Equal(t, "org.apache.pulsar.io.kafka.KafkaSource", sourceSpec.ClassName)
	assert.Equal(t, `{"bootstrapServers":"kafka:9092","groupId":"test-consumer"}`, sourceSpec.Configs)
}

func TestGenerateFunctionOutputSpecWithConnector(t *testing.T) {
	function := makeFunctionSample("connector-sink")
	function.Spec.Output.Topic = ""
	configs := &v1alpha1.Config{
		Data: map[string]interface{}{
			"bootstrapServers": "kafka:9092",
			"topic":            "kafka-output",
		},
	}
	function.Spec.SinkConfig = &v1alpha1.SinkConnectorSpec{
		SinkType:  "kafka",
		ClassName: "org.apache.pulsar.io.kafka.KafkaSink",
		Configs:   configs,
	}

	sinkSpec := generateFunctionOutputSpec(function)
	assert.Equal(t, "kafka", sinkSpec.Builtin)
	assert.Equal(t, "org.apache.pulsar.io.kafka.KafkaSink", sinkSpec.ClassName)
	assert.Equal(t, `{"bootstrapServers":"kafka:9092","sinkType":"kafka","topic":"kafka-output"}`, sinkSpec.Configs)
}
