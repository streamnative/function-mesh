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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/streamnative/function-mesh/api/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/proto"

	appsv1 "k8s.io/api/apps/v1"
	autov1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const EnvShardID = "SHARD_ID"
const FunctionsInstanceClasspath = "pulsar.functions.instance.classpath"
const DefaultRunnerImage = "streamnative/pulsar-all:2.7.0-rc-pm-3"

const ComponentSource = "source"
const ComponentSink = "sink"
const ComponentFunction = "function"

var GRPCPort = corev1.ContainerPort{
	Name:          "grpc",
	ContainerPort: 9093,
	Protocol:      corev1.ProtocolTCP,
}

var MetricsPort = corev1.ContainerPort{
	Name:          "metrics",
	ContainerPort: 9094,
	Protocol:      corev1.ProtocolTCP,
}

func MakeService(objectMeta *metav1.ObjectMeta, labels map[string]string) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "core/v1",
		},
		ObjectMeta: *objectMeta,
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:     "grpc",
				Protocol: corev1.ProtocolTCP,
				Port:     GRPCPort.ContainerPort,
			}},
			Selector:  labels,
			ClusterIP: "None",
		},
	}
}

func MakeHPA(objectMeta *metav1.ObjectMeta, minReplicas, maxReplicas int32,
	kind string) *autov1.HorizontalPodAutoscaler {
	// TODO: configurable cpu percentage
	cpuPercentage := int32(80)
	return &autov1.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "autoscaling/v1",
			APIVersion: "HorizontalPodAutoscaler",
		},
		ObjectMeta: *objectMeta,
		Spec: autov1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autov1.CrossVersionObjectReference{
				Kind:       kind,
				Name:       objectMeta.Name,
				APIVersion: "cloud.streamnative.io/v1alpha1",
			},
			MinReplicas:                    &minReplicas,
			MaxReplicas:                    maxReplicas,
			TargetCPUUtilizationPercentage: &cpuPercentage,
		},
	}
}

func MakeStatefulSet(objectMeta *metav1.ObjectMeta, replicas *int32, container *corev1.Container,
	labels map[string]string) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: *objectMeta,
		Spec:       *MakeStatefulSetSpec(replicas, container, labels),
	}
}

func MakeStatefulSetSpec(replicas *int32, container *corev1.Container,
	labels map[string]string) *appsv1.StatefulSetSpec {
	return &appsv1.StatefulSetSpec{
		Replicas: replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		Template:            *MakePodTemplate(container, labels),
		PodManagementPolicy: appsv1.ParallelPodManagement,
		UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
			Type: appsv1.RollingUpdateStatefulSetStrategyType,
		},
	}
}

func MakePodTemplate(container *corev1.Container, labels map[string]string) *corev1.PodTemplateSpec {
	ZeroGracePeriod := int64(0)
	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			// Tolerations: nil TODO
			Containers:                    []corev1.Container{*container},
			TerminationGracePeriodSeconds: &ZeroGracePeriod,
		},
	}
}

func MakeJavaFunctionCommand(downloadPath, packageFile, name, clusterName, details, memory string, authProvided bool) []string {
	processCommand := setShardIDEnvironmentVariableCommand() + " && " +
		strings.Join(getProcessJavaRuntimeArgs(name, packageFile, clusterName, details, memory, authProvided), " ")
	if downloadPath != "" {
		// prepend download command if the downPath is provided
		downloadCommand := strings.Join(getDownloadCommand(downloadPath, packageFile), " ")
		processCommand = downloadCommand + " && " + processCommand
	}
	return []string{"sh", "-c", processCommand}
}

func MakeGoFunctionCommand(downloadPath, goExecFilePath string, function *v1alpha1.Function) []string {
	processCommand := setShardIDEnvironmentVariableCommand() + " && " +
		strings.Join(getProcessGoRuntimeArgs(goExecFilePath, function), " ")
	if downloadPath != "" {
		// prepend download command if the downPath is provided
		downloadCommand := strings.Join(getDownloadCommand(downloadPath, goExecFilePath), " ")
		processCommand = downloadCommand + " && ls -al && pwd &&" + processCommand
	}
	return []string{"sh", "-c", processCommand}
}

func getDownloadCommand(downloadPath, componentPackage string) []string {
	return []string{
		"/pulsar/bin/pulsar-admin", // TODO configurable pulsar ROOTDIR and adminCLI
		"--admin-url",
		"$webServiceURL",
		"functions",
		"download",
		"--path",
		downloadPath,
		"--destination-file",
		"/pulsar/" + componentPackage,
	}
}

func setShardIDEnvironmentVariableCommand() string {
	return fmt.Sprintf("%s=${POD_NAME##*-} && echo shardId=${%s}", EnvShardID, EnvShardID)
}

func getProcessJavaRuntimeArgs(name string, packageName string, clusterName string, details string, memory string, authProvided bool) []string {
	args := []string{
		"exec",
		"java",
		"-cp",
		"/pulsar/instances/java-instance.jar",
		fmt.Sprintf("-D%s=%s", FunctionsInstanceClasspath, "/pulsar/lib/*"),
		"-Dlog4j.configurationFile=kubernetes_instance_log4j2.xml", // todo
		"-Dpulsar.function.log.dir=logs/functions",
		"-Dpulsar.function.log.file=" + fmt.Sprintf("%s-${%s}", name, EnvShardID),
		"-Xmx" + memory,
		"org.apache.pulsar.functions.instance.JavaInstanceMain",
		"--jar",
		packageName,
	}
	sharedArgs := getSharedArgs(details, clusterName, authProvided)
	args = append(args, sharedArgs...)
	return args
}

// This method is suitable for Java and Python runtime, not include Go runtime.
func getSharedArgs(details, clusterName string, authProvided bool) []string {
	args := []string{
		"--instance_id",
		"${" + EnvShardID + "}",
		"--function_id",
		fmt.Sprintf("${%s}-%d", EnvShardID, time.Now().Unix()),
		"--function_version",
		"0",
		"--function_details",
		"'" + details + "'", //in json format
		"--pulsar_serviceurl",
		"$brokerServiceURL",
		"--max_buffered_tuples",
		"100", // TODO
		"--port",
		strconv.Itoa(int(GRPCPort.ContainerPort)),
		"--metrics_port",
		strconv.Itoa(int(MetricsPort.ContainerPort)),
		"--expected_healthcheck_interval",
		"-1", // TurnOff BuiltIn HealthCheck to avoid instance exit
		"--cluster_name",
		clusterName,
	}

	if authProvided {
		args = append(args, []string{
			"--client_auth_plugin",
			"$clientAuthenticationPlugin",
			"--client_auth_params",
			"$clientAuthenticationParameters",
			"--use_tls",
			"$useTls",
			"--tls_allow_insecure",
			"$tlsAllowInsecureConnection",
			"--hostname_verification_enabled",
			"$tlsHostnameVerificationEnable",
			"--tls_trust_cert_path",
			"$tlsTrustCertsFilePath"}...)
	}

	return args
}

func generateGoFunctionDetailsInJSON(function *v1alpha1.Function) string {
	functionDetails := convertFunctionDetails(function)
	marshaler := &jsonpb.Marshaler{}
	json, err := marshaler.MarshalToString(functionDetails)
	if err != nil {
		// TODO
		panic(err)
	}
	return json
}

func getProcessGoRuntimeArgs(goExecFilePath string, function *v1alpha1.Function) []string {
	str := generateGoFunctionDetailsInJSON(function)
	tmpStr := strings.TrimSuffix(str, "}")

	inputTopic := function.Spec.Input.Topics[0]
	outputTopic := function.Spec.Output.Topic

	configContent := fmt.Sprintf("%s, \"pulsarServiceURL\": \"pulsar://test-pulsar-broker.default.svc.cluster.local:6650\", "+
		"\"sourceSpecsTopic\": \"%s\", \"sinkSpecsTopic\": \"%s\"}", tmpStr, inputTopic, outputTopic)

	goPath := fmt.Sprintf("/pulsar/%s", goExecFilePath)
	conf := fmt.Sprintf("'%s'", configContent)

	args := []string{
		"chmod +x",
		goPath,
		"&&",
		"exec",
		goPath,
		"-instance-conf",
		conf,
	}

	return args
}

func convertProcessingGuarantee(input v1alpha1.ProcessGuarantee) proto.ProcessingGuarantees {
	switch input {
	case v1alpha1.AtmostOnce:
		return proto.ProcessingGuarantees_ATMOST_ONCE
	case v1alpha1.AtleastOnce:
		return proto.ProcessingGuarantees_ATLEAST_ONCE
	case v1alpha1.EffectivelyOnce:
		return proto.ProcessingGuarantees_EFFECTIVELY_ONCE
	default:
		// should never reach here
		return proto.ProcessingGuarantees_ATLEAST_ONCE
	}
}

func convertSubPosition(pos v1alpha1.SubscribePosition) proto.SubscriptionPosition {
	switch pos {
	case v1alpha1.Earliest:
		return proto.SubscriptionPosition_EARLIEST
	case v1alpha1.Latest:
		return proto.SubscriptionPosition_LATEST
	default:
		return proto.SubscriptionPosition_EARLIEST
	}
}

func generateRetryDetails(maxMessageRetry int32, deadLetterTopic string) *proto.RetryDetails {
	return &proto.RetryDetails{
		MaxMessageRetries: maxMessageRetry,
		DeadLetterTopic:   deadLetterTopic,
	}
}

func generateResource(resources corev1.ResourceList) *proto.Resources {
	return &proto.Resources{
		Cpu:  float64(resources.Cpu().Value()),
		Ram:  resources.Memory().Value(),
		Disk: resources.Storage().Value(),
	}
}

func getUserConfig(configs map[string]string) string {
	// validated in admission web hook
	bytes, _ := json.Marshal(configs)
	return string(bytes)
}

func generateContainerEnv(secrets map[string]v1alpha1.SecretRef) []corev1.EnvVar {
	vars := []corev1.EnvVar{{
		Name:      "POD_NAME",
		ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}},
	}}

	for secretName, secretRef := range secrets {
		vars = append(vars, corev1.EnvVar{
			Name: secretName,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: secretRef.Path},
					Key:                  secretRef.Key,
				},
			},
		})
	}

	return vars
}

func generateContainerEnvFrom(messagingConfig string, authConfig string) []corev1.EnvFromSource {
	envs := []corev1.EnvFromSource{{
		ConfigMapRef: &corev1.ConfigMapEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: messagingConfig},
		},
	}}

	if authConfig != "" {
		envs = append(envs, corev1.EnvFromSource{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: authConfig},
			},
		})
	}

	return envs
}
