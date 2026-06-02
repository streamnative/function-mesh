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
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	"github.com/streamnative/function-mesh/utils"
	"google.golang.org/protobuf/encoding/protojson"
	appsv1 "k8s.io/api/apps/v1"
	autov2 "k8s.io/api/autoscaling/v2"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// log is for logging in this package.
var log = logf.Log.WithName("function-resource")

func MakeFunctionHPA(function *v1alpha1.Function) *autov2.HorizontalPodAutoscaler {
	objectMeta := MakeFunctionObjectMeta(function)
	targetRef := autov2.CrossVersionObjectReference{
		Kind:       function.Kind,
		Name:       function.Name,
		APIVersion: function.APIVersion,
	}
	return MakeHPA(objectMeta, targetRef, function.Spec.MinReplicas, function.Spec.MaxReplicas, function.Spec.Pod, function.Spec.Resources)
}

func MakeFunctionService(function *v1alpha1.Function) *corev1.Service {
	labels := makeFunctionLabels(function)
	objectMeta := MakeFunctionObjectMeta(function)
	return MakeService(objectMeta, labels)
}

func MakeFunctionStatefulSet(ctx context.Context, cli client.Client, function *v1alpha1.Function) (*appsv1.StatefulSet, error) {
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
	downloadConfig := NewDownloadServiceConfig(function.Spec.PulsarPackageService, function.Spec.Pulsar)
	statefulSet := MakeStatefulSet(objectMeta, function.Spec.Replicas, function.Spec.DownloaderImage,
		makeFunctionContainer(function), makeFunctionVolumes(function, functionRuntimeAuthConfig(function)), labels, function.Spec.Pod,
		function.Spec.Pulsar, downloadConfig, function.Spec.Java, function.Spec.Python, function.Spec.Golang, function.Spec.Pod.Env,
		function.Spec.LogTopic, function.Spec.FilebeatImage, function.Spec.LogTopicAgent, function.Spec.VolumeMounts,
		function.Spec.VolumeClaimTemplates, function.Spec.PersistentVolumeClaimRetentionPolicy)

	globalBackendConfigVersion, namespacedBackendConfigVersion, err := PatchStatefulSet(ctx, cli, function.Namespace, statefulSet)
	if err != nil {
		return nil, err
	}
	if globalBackendConfigVersion != "" {
		function.Status.GlobalBackendConfigRevision = globalBackendConfigVersion
	}
	if namespacedBackendConfigVersion != "" {
		function.Status.NamespacedBackendConfigRevision = namespacedBackendConfigVersion
	}

	return statefulSet, nil
}

func MakeFunctionObjectMeta(function *v1alpha1.Function) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		Name:      makeJobName(function.Name, v1alpha1.FunctionComponent),
		Namespace: function.Namespace,
		Labels:    makeFunctionLabels(function),
		OwnerReferences: []metav1.OwnerReference{
			*metav1.NewControllerRef(function, function.GroupVersionKind()),
		},
	}
}

func MakeFunctionCleanUpJob(function *v1alpha1.Function) *v1.Job {
	labels := makeFunctionLabels(function)
	labels["owner"] = string(function.GetUID())
	objectMeta := &metav1.ObjectMeta{
		Name:      makeJobName(function.Name, v1alpha1.FunctionComponent) + "-cleanup",
		Namespace: function.Namespace,
		Labels:    labels,
	}
	container := makeFunctionContainer(function)
	container.Name = CleanupContainerName
	container.LivenessProbe = nil
	container.StartupProbe = nil
	authConfig := function.Spec.Pulsar.CleanupAuthConfig
	if authConfig == nil {
		authConfig = function.Spec.Pulsar.AuthConfig
	}
	volumeMounts := makeFunctionVolumeMounts(function, authConfig)
	container.VolumeMounts = volumeMounts
	if function.Spec.CleanupImage != "" {
		container.Image = function.Spec.CleanupImage
	}
	topicPattern := function.Spec.Input.TopicPattern
	inputSpecs := generateInputSpec(function.Spec.Input)
	inputTopics := make([]string, 0, len(inputSpecs))
	for topic, spec := range inputSpecs {
		if spec.IsRegexPattern {
			topicPattern = topic
		} else {
			inputTopics = append(inputTopics, topic)
		}
	}
	hasPulsarctl := function.Spec.ImageHasPulsarctl
	if match, _ := regexp.MatchString(RunnerImageHasPulsarctl, function.Spec.Image); match {
		hasPulsarctl = true
	}
	command := getCleanUpCommand(hasPulsarctl,
		function.Spec.Pulsar.AuthSecret != "",
		function.Spec.Pulsar.TLSSecret != "",
		function.Spec.Pulsar.TLSConfig,
		authConfig,
		inputTopics,
		topicPattern,
		function.Spec.SubscriptionName,
		function.Spec.Tenant,
		function.Spec.Namespace,
		function.Spec.Name, false)
	container.Command = command
	return makeCleanUpJob(objectMeta, container, makeFunctionVolumes(function, authConfig),
		makeFunctionLabels(function), function.Spec.Pod)
}

func makeFunctionVolumes(function *v1alpha1.Function, authConfig *v1alpha1.AuthConfig) []corev1.Volume {
	tlsConfig := functionPulsarTLSConfig(function)
	volumes := GeneratePodVolumes(function.Spec.Pod.Volumes,
		function.Spec.Output.ProducerConf,
		function.Spec.Input.SourceSpecs,
		tlsConfig,
		authConfig,
		GetRuntimeLogConfigNames(function.Spec.Java, function.Spec.Python, function.Spec.Golang),
		function.Spec.LogTopicAgent)
	return AppendPackageServiceVolumes(volumes, function.Spec.PulsarPackageService)
}

func makeFunctionVolumeMounts(function *v1alpha1.Function, authConfig *v1alpha1.AuthConfig) []corev1.VolumeMount {
	tlsConfig := functionPulsarTLSConfig(function)
	return GenerateContainerVolumeMounts(function.Spec.VolumeMounts,
		function.Spec.Output.ProducerConf,
		function.Spec.Input.SourceSpecs,
		tlsConfig,
		authConfig,
		GetRuntimeLogConfigNames(function.Spec.Java, function.Spec.Python, function.Spec.Golang),
		function.Spec.LogTopicAgent)
}

func makeFunctionContainer(function *v1alpha1.Function) *corev1.Container {
	imagePullPolicy := function.Spec.ImagePullPolicy
	if imagePullPolicy == "" {
		imagePullPolicy = corev1.PullIfNotPresent
	}
	livenessProbe := MakeLivenessProbe(function.Spec.Pod.Liveness)
	startupProbe := function.Spec.Pod.StartupProbe.DeepCopy()
	allowPrivilegeEscalation := false
	mounts := makeFunctionVolumeMounts(function, functionRuntimeAuthConfig(function))
	mounts = AppendPackageServiceVolumeMounts(mounts, function.Spec.PulsarPackageService)
	if utils.EnableInitContainers {
		mounts = append(mounts,
			generateDownloaderVolumeMountsForRuntime(function.Spec.Java, function.Spec.Python, function.Spec.Golang, function.Spec.GenericRuntime)...)
	}
	envFrom := functionPulsarEnvFrom(function)
	if function.Spec.PulsarPackageService != nil {
		envFrom = append(envFrom, GenerateContainerEnvFromWithPrefix(function.Spec.PulsarPackageService.PulsarConfig,
			function.Spec.PulsarPackageService.AuthSecret, function.Spec.PulsarPackageService.TLSSecret, PackageServiceEnvPrefix)...)
	}
	env := generateContainerEnv(function)
	env = append(env, generateKafkaAuthEnv(function.Spec.Kafka)...)
	return &corev1.Container{
		// TODO new container to pull user code image and upload jars into bookkeeper
		Name:            FunctionContainerName,
		Image:           getFunctionRunnerImage(&function.Spec),
		Command:         makeFunctionCommand(function),
		Ports:           []corev1.ContainerPort{GRPCPort, MetricsPort},
		Env:             env,
		Resources:       function.Spec.Resources,
		ImagePullPolicy: imagePullPolicy,
		EnvFrom:         envFrom,
		VolumeMounts:    mounts,
		LivenessProbe:   livenessProbe,
		StartupProbe:    startupProbe,
		SecurityContext: &corev1.SecurityContext{
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"ALL"},
			},
			AllowPrivilegeEscalation: &allowPrivilegeEscalation,
		},
	}
}

func makeFunctionLabels(function *v1alpha1.Function) map[string]string {
	jobName := makeJobName(function.Name, v1alpha1.FunctionComponent)
	labels := map[string]string{
		"app.kubernetes.io/name":            jobName,
		"app.kubernetes.io/instance":        jobName,
		"compute.functionmesh.io/app":       AppFunctionMesh,
		"compute.functionmesh.io/component": ComponentFunction,
		"compute.functionmesh.io/name":      function.Name,
		"compute.functionmesh.io/namespace": function.Namespace,
		// The following will be deprecated after two releases
		"app":       AppFunctionMesh,
		"component": ComponentFunction,
		"name":      function.Name,
		"namespace": function.Namespace,
	}
	return labels
}

func makeFunctionCommand(function *v1alpha1.Function) []string {
	spec := function.Spec
	if isKafkaFunction(function) && (spec.GenericRuntime == nil || spec.GenericRuntime.FunctionFile == "") {
		return nil
	}
	downloadConfig := NewDownloadServiceConfig(spec.PulsarPackageService, spec.Pulsar)
	pulsarAuthConfig := functionPulsarAuthConfig(function)
	pulsarTLSConfig := functionPulsarTLSConfig(function)

	connectorsDirectory := ""
	if spec.SourceConfig != nil || spec.SinkConfig != nil {
		connectorsDirectory = DefaultConnectorsDirectory
	}

	hasPulsarctl := function.Spec.ImageHasPulsarctl
	hasWget := function.Spec.ImageHasWget
	if match, _ := regexp.MatchString(RunnerImageHasPulsarctl, function.Spec.Image); match {
		hasPulsarctl = true
		hasWget = true
	}
	if spec.Java != nil {
		if spec.Java.Jar != "" {
			mountPath := extractMountPath(spec.Java.Jar)
			instancePath := DefaultPulsarFunctionsJavaInstancePath
			if spec.Java.InstancePath != nil && *spec.Java.InstancePath != "" {
				instancePath = *spec.Java.InstancePath
			}
			entryClass := DefaultPulsarFunctionsJavaInstanceEntryClass
			if spec.Java.EntryClass != nil && *spec.Java.EntryClass != "" {
				entryClass = *spec.Java.EntryClass
			}
			return MakeJavaFunctionCommand(spec.Java.JarLocation, mountPath,
				spec.Name, spec.ClusterName,
				GenerateJavaLogConfigCommand(spec.Java, spec.LogTopicAgent),
				parseJavaLogLevel(spec.Java),
				generateFunctionDetailsInJSON(function),
				spec.Java.ExtraDependenciesDir,
				connectorsDirectory,
				string(function.UID),
				spec.Resources.Limits.Memory(),
				spec.Java.JavaOpts, hasPulsarctl, hasWget, downloadConfig,
				pulsarAuthSecret(function) != "", pulsarTLSSecret(function) != "",
				spec.SecretsMap, spec.StateConfig, pulsarTLSConfig,
				pulsarAuthConfig, spec.MaxPendingAsyncRequests,
				GenerateJavaLogConfigFileName(function.Spec.Java), instancePath, entryClass)
		}
	} else if spec.Python != nil {
		if spec.Python.Py != "" {
			mountPath := extractMountPath(spec.Python.Py)
			return MakePythonFunctionCommand(spec.Python.PyLocation, mountPath,
				spec.Name, spec.ClusterName,
				GeneratePythonLogConfigCommand(spec.Name, spec.Python, spec.LogTopicAgent),
				generateFunctionDetailsInJSON(function), string(function.UID), hasPulsarctl, hasWget, downloadConfig,
				pulsarAuthSecret(function) != "", pulsarTLSSecret(function) != "", spec.SecretsMap,
				spec.StateConfig, pulsarTLSConfig, pulsarAuthConfig)
		}
	} else if spec.Golang != nil {
		if spec.Golang.Go != "" {
			mountPath := extractMountPath(spec.Golang.Go)
			return MakeGoFunctionCommand(spec.Golang.GoLocation, mountPath, function, downloadConfig)
		}
	} else if spec.GenericRuntime != nil {
		if spec.GenericRuntime.FunctionFile != "" {
			mountPath := extractMountPath(spec.GenericRuntime.FunctionFile)
			return MakeGenericFunctionCommand(spec.GenericRuntime.FunctionFileLocation, mountPath,
				spec.GenericRuntime.Language, spec.ClusterName,
				generateFunctionDetailsInJSON(function), string(function.UID), downloadConfig,
				pulsarAuthSecret(function) != "", pulsarTLSSecret(function) != "", function.Spec.SecretsMap,
				function.Spec.StateConfig, genericRuntimeTLSConfig(function), genericRuntimeAuthConfig(function),
				genericMessagingServiceType(function), genericRuntimeClientAuthArgs(function))
		}
	}

	return nil
}

func functionPulsarTLSConfig(function *v1alpha1.Function) *v1alpha1.PulsarTLSConfig {
	if function.Spec.Pulsar == nil {
		return nil
	}
	return function.Spec.Pulsar.TLSConfig
}

func functionPulsarAuthConfig(function *v1alpha1.Function) *v1alpha1.AuthConfig {
	if function.Spec.Pulsar == nil {
		return nil
	}
	return function.Spec.Pulsar.AuthConfig
}

func functionRuntimeAuthConfig(function *v1alpha1.Function) *v1alpha1.AuthConfig {
	if isKafkaFunction(function) {
		return kafkaAuthConfigAsAuthConfig(function.Spec.Kafka.AuthConfig)
	}
	return functionPulsarAuthConfig(function)
}

func genericRuntimeAuthConfig(function *v1alpha1.Function) *v1alpha1.AuthConfig {
	if isKafkaFunction(function) {
		return nil
	}
	return functionPulsarAuthConfig(function)
}

func genericRuntimeTLSConfig(function *v1alpha1.Function) TLSConfig {
	if isKafkaFunction(function) {
		return nil
	}
	return functionPulsarTLSConfig(function)
}

func functionPulsarEnvFrom(function *v1alpha1.Function) []corev1.EnvFromSource {
	if function.Spec.Pulsar == nil {
		return nil
	}
	return GenerateContainerEnvFrom(function.Spec.Pulsar.PulsarConfig, function.Spec.Pulsar.AuthSecret,
		function.Spec.Pulsar.TLSSecret)
}

func pulsarAuthSecret(function *v1alpha1.Function) string {
	if function.Spec.Pulsar == nil {
		return ""
	}
	return function.Spec.Pulsar.AuthSecret
}

func pulsarTLSSecret(function *v1alpha1.Function) string {
	if function.Spec.Pulsar == nil {
		return ""
	}
	return function.Spec.Pulsar.TLSSecret
}

func isKafkaFunction(function *v1alpha1.Function) bool {
	return function.Spec.Kafka != nil
}

func genericMessagingServiceType(function *v1alpha1.Function) string {
	if isKafkaFunction(function) {
		return "kafka"
	}
	return ""
}

func kafkaAuthConfigAsAuthConfig(auth *v1alpha1.KafkaAuthConfig) *v1alpha1.AuthConfig {
	if auth == nil {
		return nil
	}
	return &v1alpha1.AuthConfig{
		OAuth2Config: auth.OAuth2Config,
		GenericAuth:  auth.GenericAuth,
	}
}

func genericRuntimeClientAuthArgs(function *v1alpha1.Function) []string {
	if !isKafkaFunction(function) || function.Spec.Kafka.AuthConfig == nil {
		return nil
	}
	auth := function.Spec.Kafka.AuthConfig
	switch {
	case auth.OAuth2Config != nil:
		return []string{
			"--client_auth_plugin",
			"oauth2",
			"--client_auth_params",
			shellQuoteLiteral(mustJSON(map[string]string{
				"private_key": auth.OAuth2Config.GetMountFile(),
				"issuer_url":  auth.OAuth2Config.IssuerURL,
				"audience":    auth.OAuth2Config.Audience,
				"scope":       auth.OAuth2Config.Scope,
			})),
		}
	case auth.GenericAuth != nil:
		return []string{
			"--client_auth_plugin",
			auth.GenericAuth.ClientAuthenticationPlugin,
			"--client_auth_params",
			shellQuoteLiteral(auth.GenericAuth.ClientAuthenticationParameters),
		}
	case auth.PlainAuthConfig != nil:
		return []string{
			"--client_auth_plugin",
			"token",
			"--client_auth_params",
			fmt.Sprintf(`"${%s}:${%s}"`, KafkaAuthUsernameEnv, KafkaAuthPasswordEnv),
		}
	default:
		return nil
	}
}

func generateKafkaAuthEnv(kafka *v1alpha1.KafkaMessaging) []corev1.EnvVar {
	if kafka == nil || kafka.AuthConfig == nil || kafka.AuthConfig.PlainAuthConfig == nil {
		return nil
	}
	plainAuth := kafka.AuthConfig.PlainAuthConfig
	usernameKey := plainAuth.UsernameKey
	if usernameKey == "" {
		usernameKey = "username"
	}
	passwordKey := plainAuth.PasswordKey
	if passwordKey == "" {
		passwordKey = "password"
	}
	return []corev1.EnvVar{
		{
			Name: KafkaAuthUsernameEnv,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: plainAuth.SecretName},
					Key:                  usernameKey,
				},
			},
		},
		{
			Name: KafkaAuthPasswordEnv,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: plainAuth.SecretName},
					Key:                  passwordKey,
				},
			},
		},
	}
}

func mustJSON(value interface{}) string {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func generateFunctionDetailsInJSON(function *v1alpha1.Function) string {
	functionDetails := convertFunctionDetails(function)
	json, err := protojson.Marshal(functionDetails)
	if err != nil {
		// TODO
		panic(err)
	}
	log.Info(string(json))
	return string(json)
}

func extractMountPath(p string) string {
	if utils.EnableInitContainers {
		mountPath := p
		// for relative path, volume should be mounted to the WorkDir
		// and path also should be under the $WorkDir dir
		if !strings.HasPrefix(p, "/") {
			mountPath = WorkDir + p
		} else if !strings.HasPrefix(p, WorkDir) {
			mountPath = strings.Replace(p, "/", WorkDir, 1)
		}
		return mountPath
	}
	return p
}
