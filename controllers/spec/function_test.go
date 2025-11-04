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
	"regexp"
	"strings"
	"testing"

	"github.com/streamnative/function-mesh/utils"
	corev1 "k8s.io/api/core/v1"

	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/proto"

	"google.golang.org/protobuf/encoding/protojson"

	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateFunctionDetailsForStatefulFunction(t *testing.T) {
	fnc := makeFunctionSample("test")
	commands := makeFunctionCommand(fnc)
	assert.Assert(t, len(commands) == 3, "commands should be 3 but got %d", len(commands))
	startCommands := commands[2]
	assert.Assert(t, strings.Contains(startCommands, "--state_storage_serviceurl"),
		"start command should contain state_storage_serviceurl")
	assert.Assert(t, strings.Contains(startCommands, "bk://localhost:4181"),
		"start command should contain bk://localhost:4181")
}

func makeFunctionSample(functionName string) *v1alpha1.Function {
	maxPending := int32(1000)
	replicas := int32(1)
	minReplicas := int32(1)
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
			Namespace:   "default",
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
			MinReplicas:                  &minReplicas,
			MaxReplicas:                  &maxReplicas,
			AutoAck:                      &trueVal,
			MaxPendingAsyncRequests:      &maxPending,
			Messaging: v1alpha1.Messaging{
				Pulsar: &v1alpha1.PulsarMessaging{
					PulsarConfig: TestClusterName,
				},
			},
			Runtime: v1alpha1.Runtime{
				Java: &v1alpha1.JavaRuntime{
					Jar:         "pulsar-functions-api-examples.jar",
					JarLocation: "public/default/nlu-test-java-function",
				},
			},
			StateConfig: &v1alpha1.Stateful{
				Pulsar: &v1alpha1.PulsarStateStore{
					ServiceURL: "bk://localhost:4181",
				},
			},
		},
	}
}

func makeFunctionSamplePackageURL(functionName string) *v1alpha1.Function {
	f := makeFunctionSample(functionName)
	f.Spec.Java.JarLocation = "function://public/default/java-function"
	f.Spec.Java.Jar = "/tmp/java-function.jar"
	return f
}

func TestInitContainerDownloader(t *testing.T) {
	utils.EnableInitContainers = true
	function := makeFunctionSamplePackageURL("test")

	objectMeta := MakeFunctionObjectMeta(function)

	runnerImagePullSecrets := getFunctionRunnerImagePullSecret()
	for _, mapSecret := range runnerImagePullSecrets {
		if value, ok := mapSecret["name"]; ok {
			function.Spec.Pod.ImagePullSecrets = append(function.Spec.Pod.ImagePullSecrets, corev1.LocalObjectReference{Name: value})
		}
	}
	runnerImagePullPolicy := getFunctionRunnerImagePullPolicy()
	function.Spec.ImagePullPolicy = runnerImagePullPolicy

	labels := makeFunctionLabels(function)
	statefulSet := MakeStatefulSet(objectMeta, function.Spec.Replicas, function.Spec.DownloaderImage,
		makeFunctionContainer(function), makeFunctionVolumes(function, function.Spec.Pulsar.AuthConfig), labels, function.Spec.Pod,
		function.Spec.Pulsar.AuthConfig, function.Spec.Pulsar.TLSConfig, function.Spec.Pulsar.PulsarConfig, function.Spec.Pulsar.AuthSecret,
		function.Spec.Pulsar.TLSSecret, function.Spec.Java, function.Spec.Python, function.Spec.Golang, function.Spec.Pod.Env,
		function.Spec.LogTopic, function.Spec.FilebeatImage, function.Spec.LogTopicAgent, function.Spec.VolumeMounts,
		function.Spec.VolumeClaimTemplates, function.Spec.PersistentVolumeClaimRetentionPolicy)

	assert.Assert(t, statefulSet != nil, "statefulSet should not be nil")
	assert.Assert(t, len(statefulSet.Spec.Template.Spec.InitContainers) == 1, "init container should be 1 but got %d", len(statefulSet.Spec.Template.Spec.InitContainers))
	assert.Assert(t, statefulSet.Spec.Template.Spec.InitContainers[0].Name == "downloader", "init container name should be downloader but got %s", statefulSet.Spec.Template.Spec.InitContainers[0].Name)
	downloaderCommands := statefulSet.Spec.Template.Spec.InitContainers[0].Command
	functionCommands := makeFunctionCommand(function)
	assert.Assert(t, len(downloaderCommands) == 3, "downloader commands should be 3 but got %d", len(downloaderCommands))
	assert.Assert(t, len(functionCommands) == 3, "function commands should be 3 but got %d", len(functionCommands))
	assert.Assert(t, len(statefulSet.Spec.Template.Spec.InitContainers[0].VolumeMounts) == 1, "volume mounts should be 1 but got %d", len(statefulSet.Spec.Template.Spec.InitContainers[0].VolumeMounts))
	assert.Assert(t, len(statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts) == 1, "volume mounts should be 1 but got %d", len(statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts))
	startDownloadCommands := downloaderCommands[2]
	downloadVolumeMount := statefulSet.Spec.Template.Spec.InitContainers[0].VolumeMounts[0]
	assert.Assert(t, downloadVolumeMount.Name == "downloader-volume", "volume mount name should be download-volume but got %s", downloadVolumeMount.Name)
	assert.Assert(t, downloadVolumeMount.MountPath == "/pulsar/download", "volume mount path should be /pulsar/download but got %s", downloadVolumeMount.MountPath)
	startFunctionCommands := functionCommands[2]
	functionVolumeMount := statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts[0]
	assert.Assert(t, functionVolumeMount.Name == "downloader-volume", "volume mount name should be downloader-volume but got %s", functionVolumeMount.Name)
	assert.Assert(t, functionVolumeMount.MountPath == "/pulsar/tmp", "volume mount path should be /pulsar/tmp but got %s", functionVolumeMount.MountPath)
	assert.Assert(t, strings.Contains(startDownloadCommands, "/pulsar/download/java-function.jar"), "download command should contain /pulsar/download/java-function.jar: %s", startDownloadCommands)
	assert.Assert(t, strings.Contains(startFunctionCommands, "/pulsar/tmp/java-function.jar"), "function command should contain /pulsar/tmp/java-function.jar: %s", startFunctionCommands)
}

func TestJavaFunctionCommandWithConnectorOverrides(t *testing.T) {
	function := makeFunctionSample("connector-test")
	function.Spec.Input = v1alpha1.InputConf{}
	function.Spec.Output = v1alpha1.OutputConf{}
	function.Spec.LogTopic = ""
	function.Spec.StateConfig = nil

	sourceConfigs := v1alpha1.NewConfig(map[string]interface{}{
		"awsEndpoint":              "http://localstack:4566",
		"cloudwatchEndpoint":       "http://localstack:4566",
		"dynamoEndpoint":           "http://localstack:4566",
		"awsRegion":                "us-east-1",
		"awsKinesisStreamName":     "demo-stream",
		"awsCredentialPluginParam": `{"accessKey":"test","secretKey":"test"}`,
		"useEnhancedFanOut":        false,
		"applicationName":          "pulsar-kinesis-demo",
		"initialPositionInStream":  "TRIM_HORIZON",
	})
	function.Spec.SourceConfig = &v1alpha1.SourceConnectorSpec{
		SourceType:    "kinesis",
		Configs:       &sourceConfigs,
		TypeClassName: "java.lang.Object",
	}

	sinkConfigs := v1alpha1.NewConfig(map[string]interface{}{
		"bootstrapServers": "kafka:9092",
		"acks":             "all",
		"topic":            "kinesis-demo",
		"producerConfigProperties": map[string]interface{}{
			"enable.idempotence": true,
		},
	})
	function.Spec.SinkConfig = &v1alpha1.SinkConnectorSpec{
		SinkType:      "kafka",
		Configs:       &sinkConfigs,
		TypeClassName: "java.lang.Object",
	}

	commands := makeFunctionCommand(function)
	assert.Assert(t, len(commands) == 3, "commands should be 3 but got %d", len(commands))

	startCommand := commands[2]
	assert.Assert(t, strings.Contains(startCommand, "--connectors_directory "+DefaultConnectorsDirectory),
		"start command should include connectors directory but got %s", startCommand)
	re := regexp.MustCompile(`--function_details '([^']+)'`)
	matches := re.FindStringSubmatch(startCommand)
	assert.Assert(t, len(matches) == 2, "unable to locate function details in command: %s", startCommand)

	functionDetailsJSON := matches[1]
	details := &proto.FunctionDetails{}
	err := protojson.Unmarshal([]byte(functionDetailsJSON), details)
	assert.NilError(t, err)

	assert.Equal(t, details.GetTenant(), "public")
	assert.Equal(t, details.GetNamespace(), "default")
	assert.Equal(t, details.GetName(), "connector-test")

	sourceSpec := details.GetSource()
	assert.Equal(t, sourceSpec.GetBuiltin(), "kinesis")
	assert.Equal(t, sourceSpec.GetTypeClassName(), "java.lang.Object")
	sourceConfigsMap := map[string]interface{}{}
	err = json.Unmarshal([]byte(sourceSpec.GetConfigs()), &sourceConfigsMap)
	assert.NilError(t, err)
	assert.Equal(t, sourceConfigsMap["awsEndpoint"], "http://localstack:4566")
	assert.Equal(t, sourceConfigsMap["awsKinesisStreamName"], "demo-stream")
	assert.Equal(t, sourceConfigsMap["initialPositionInStream"], "TRIM_HORIZON")

	sinkSpec := details.GetSink()
	sinkConfigsMap := map[string]interface{}{}
	err = json.Unmarshal([]byte(sinkSpec.GetConfigs()), &sinkConfigsMap)
	assert.NilError(t, err)
	assert.Equal(t, sinkConfigsMap["sinkType"], "kafka")
	assert.Equal(t, sinkConfigsMap["bootstrapServers"], "kafka:9092")
	assert.Equal(t, sinkConfigsMap["acks"], "all")
	producerConfig := sinkConfigsMap["producerConfigProperties"].(map[string]interface{})
	assert.Equal(t, producerConfig["enable.idempotence"], true)
}
