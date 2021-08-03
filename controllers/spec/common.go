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

	autov2beta2 "k8s.io/api/autoscaling/v2beta2"

	"github.com/streamnative/function-mesh/api/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/proto"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	EnvShardID                 = "SHARD_ID"
	FunctionsInstanceClasspath = "pulsar.functions.instance.classpath"
	DefaultRunnerTag           = "2.8.0.7"
	DefaultRunnerPrefix        = "streamnative/"
	DefaultRunnerImage         = DefaultRunnerPrefix + "pulsar-all:" + DefaultRunnerTag
	DefaultJavaRunnerImage     = DefaultRunnerPrefix + "pulsar-functions-java-runner:" + DefaultRunnerTag
	DefaultPythonRunnerImage   = DefaultRunnerPrefix + "pulsar-functions-python-runner:" + DefaultRunnerTag
	DefaultGoRunnerImage       = DefaultRunnerPrefix + "pulsar-functions-go-runner:" + DefaultRunnerTag
	PulsarAdminExecutableFile  = "/pulsar/bin/pulsar-admin"
	PulsarDownloadRootDir      = "/pulsar"

	ComponentSource   = "source"
	ComponentSink     = "sink"
	ComponentFunction = "function"

	PackageNameFunctionPrefix = "function://"
	PackageNameSinkPrefix     = "sink://"
	PackageNameSourcePrefix   = "source://"

	AnnotationPrometheusScrape = "prometheus.io/scrape"
	AnnotationPrometheusPort   = "prometheus.io/port"

	EnvGoFunctionConfigs = "GO_FUNCTION_CONF"
)

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
	objectMeta.Name = MakeHeadlessServiceName(objectMeta.Name)
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

// MakeHeadlessServiceName changes the name of service to headless style
func MakeHeadlessServiceName(serviceName string) string {
	return fmt.Sprintf("%s-headless", serviceName)
}

func MakeDefaultHPA(objectMeta *metav1.ObjectMeta, minReplicas, maxReplicas int32,
	kind string) *autov2beta2.HorizontalPodAutoscaler {
	return &autov2beta2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "autoscaling/v1",
			APIVersion: "HorizontalPodAutoscaler",
		},
		ObjectMeta: *objectMeta,
		Spec: autov2beta2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autov2beta2.CrossVersionObjectReference{
				Kind:       kind,
				Name:       objectMeta.Name,
				APIVersion: "compute.functionmesh.io/v1alpha1",
			},
			MinReplicas: &minReplicas,
			MaxReplicas: maxReplicas,
			Metrics:     defaultHPAMetrics(),
		},
	}
}

func MakeHPA(objectMeta *metav1.ObjectMeta, autoscalerSpec *autov2beta2.HorizontalPodAutoscalerSpec,
	kind string) *autov2beta2.HorizontalPodAutoscaler {
	return &autov2beta2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "autoscaling/v1",
			APIVersion: "HorizontalPodAutoscaler",
		},
		ObjectMeta: *objectMeta,
		Spec: autov2beta2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autov2beta2.CrossVersionObjectReference{
				Kind:       kind,
				Name:       objectMeta.Name,
				APIVersion: "compute.functionmesh.io/v1alpha1",
			},
			MinReplicas: autoscalerSpec.MinReplicas,
			MaxReplicas: autoscalerSpec.MaxReplicas,
			Metrics:     autoscalerSpec.Metrics,
			Behavior:    autoscalerSpec.Behavior,
		},
	}
}

func MakeStatefulSet(objectMeta *metav1.ObjectMeta, replicas *int32, container *corev1.Container,
	volumes []corev1.Volume, labels map[string]string, policy v1alpha1.PodPolicy) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: *objectMeta,
		Spec: *MakeStatefulSetSpec(replicas, container, volumes, labels, policy,
			MakeHeadlessServiceName(objectMeta.Name)),
	}
}

func MakeStatefulSetSpec(replicas *int32, container *corev1.Container,
	volumes []corev1.Volume, labels map[string]string, policy v1alpha1.PodPolicy,
	serviceName string) *appsv1.StatefulSetSpec {
	return &appsv1.StatefulSetSpec{
		Replicas: replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		Template:            *MakePodTemplate(container, volumes, labels, policy),
		PodManagementPolicy: appsv1.ParallelPodManagement,
		UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
			Type: appsv1.RollingUpdateStatefulSetStrategyType,
		},
		ServiceName: serviceName,
	}
}

func MakePodTemplate(container *corev1.Container, volumes []corev1.Volume,
	labels map[string]string, policy v1alpha1.PodPolicy) *corev1.PodTemplateSpec {
	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      mergeLabels(labels, policy.Labels),
			Annotations: generateAnnotations(policy.Annotations),
		},
		Spec: corev1.PodSpec{
			InitContainers:                policy.InitContainers,
			Containers:                    append(policy.Sidecars, *container),
			TerminationGracePeriodSeconds: &policy.TerminationGracePeriodSeconds,
			Volumes:                       volumes,
			NodeSelector:                  policy.NodeSelector,
			Affinity:                      policy.Affinity,
			Tolerations:                   policy.Tolerations,
			SecurityContext:               policy.SecurityContext,
			ImagePullSecrets:              policy.ImagePullSecrets,
			ServiceAccountName:            policy.ServiceAccountName,
		},
	}
}

func MakeJavaFunctionCommand(downloadPath, packageFile, name, clusterName, details, memory, extraDependenciesDir string,
	authProvided, tlsProvided bool) []string {
	processCommand := setShardIDEnvironmentVariableCommand() + " && " +
		strings.Join(getProcessJavaRuntimeArgs(name, packageFile, clusterName, details,
			memory, extraDependenciesDir, authProvided, tlsProvided), " ")
	if downloadPath != "" {
		// prepend download command if the downPath is provided
		downloadCommand := strings.Join(getDownloadCommand(downloadPath, packageFile), " ")
		processCommand = downloadCommand + " && " + processCommand
	}
	return []string{"sh", "-c", processCommand}
}

func MakePythonFunctionCommand(downloadPath, packageFile, name, clusterName, details string,
	authProvided, tlsProvided bool) []string {
	processCommand := setShardIDEnvironmentVariableCommand() + " && " +
		strings.Join(getProcessPythonRuntimeArgs(name, packageFile, clusterName,
			details, authProvided, tlsProvided), " ")
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
	// The download path is the path that the package saved in the pulsar.
	// By default, it's the path that the package saved in the pulsar, we can use package name
	// to replace it for downloading packages from packages management service.
	if hasPackageNamePrefix(downloadPath) {
		return []string{
			PulsarAdminExecutableFile,
			"--admin-url",
			"$webServiceURL",
			"packages",
			"download",
			downloadPath,
			"--path",
			PulsarDownloadRootDir + "/" + componentPackage,
		}
	}
	return []string{
		PulsarAdminExecutableFile, // TODO configurable pulsar ROOTDIR and adminCLI
		"--admin-url",
		"$webServiceURL",
		"functions",
		"download",
		"--path",
		downloadPath,
		"--destination-file",
		PulsarDownloadRootDir + "/" + componentPackage,
	}
}

// TODO: do a more strict check for the package name https://github.com/streamnative/function-mesh/issues/49
func hasPackageNamePrefix(packagesName string) bool {
	return strings.HasPrefix(packagesName, PackageNameFunctionPrefix) ||
		strings.HasPrefix(packagesName, PackageNameSinkPrefix) ||
		strings.HasPrefix(packagesName, PackageNameSourcePrefix)
}

func setShardIDEnvironmentVariableCommand() string {
	return fmt.Sprintf("%s=${POD_NAME##*-} && echo shardId=${%s}", EnvShardID, EnvShardID)
}

func getProcessJavaRuntimeArgs(name, packageName, clusterName, details, memory, extraDependenciesDir string, authProvided, tlsProvided bool) []string {
	classPath := "/pulsar/instances/java-instance.jar"
	if extraDependenciesDir != "" {
		classPath = fmt.Sprintf("%s:%s/*", classPath, extraDependenciesDir)
	}
	args := []string{
		"exec",
		"java",
		"-cp",
		classPath,
		fmt.Sprintf("-D%s=%s", FunctionsInstanceClasspath, "/pulsar/lib/*"),
		"-Dlog4j.configurationFile=kubernetes_instance_log4j2.xml", // todo
		"-Dpulsar.function.log.dir=logs/functions",
		"-Dpulsar.function.log.file=" + fmt.Sprintf("%s-${%s}", name, EnvShardID),
		"-Xmx" + memory,
		"org.apache.pulsar.functions.instance.JavaInstanceMain",
		"--jar",
		packageName,
	}
	sharedArgs := getSharedArgs(details, clusterName, authProvided, tlsProvided)
	args = append(args, sharedArgs...)
	return args
}

func getProcessPythonRuntimeArgs(name, packageName, clusterName, details string, authProvided, tlsProvided bool) []string {
	args := []string{
		"exec",
		"python",
		"/pulsar/instances/python-instance/python_instance_main.py",
		"--py",
		packageName,
		"--logging_directory",
		"logs/functions",
		"--logging_file",
		fmt.Sprintf("%s-${%s}", name, EnvShardID),
		"--logging_config_file",
		"/pulsar/conf/functions-logging/console_logging_config.ini",
		// TODO: Maybe we don't need installUserCodeDependencies, dependency_repository, and pythonExtraDependencyRepository
	}
	sharedArgs := getSharedArgs(details, clusterName, authProvided, tlsProvided)
	args = append(args, sharedArgs...)
	return args
}

// This method is suitable for Java and Python runtime, not include Go runtime.
func getSharedArgs(details, clusterName string, authProvided bool, tlsProvided bool) []string {
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
			"$clientAuthenticationParameters"}...)
	}

	if tlsProvided {
		args = append(args, []string{
			"--use_tls",
			"true",
			"--tls_allow_insecure",
			"$tlsAllowInsecureConnection",
			"--hostname_verification_enabled",
			"$tlsHostnameVerificationEnable",
			"--tls_trust_cert_path",
			"$tlsTrustCertsFilePath",
		}...)
	} else {
		args = append(args, []string{
			"--use_tls",
			"false",
		}...)
	}

	return args
}

func generateGoFunctionConf(function *v1alpha1.Function) string {
	goFunctionConfs := convertGoFunctionConfs(function)
	json, err := json.Marshal(goFunctionConfs)
	if err != nil {
		// TODO
		panic(err)
	}
	ret := string(json)
	ret = strings.ReplaceAll(ret, "\"instanceID\":0", "\"instanceID\":${"+EnvShardID+"}")
	return ret
}

func getProcessGoRuntimeArgs(goExecFilePath string, function *v1alpha1.Function) []string {
	str := generateGoFunctionConf(function)
	str = strings.ReplaceAll(str, "\"", "\\\"")
	args := []string{
		fmt.Sprintf("%s=%s", EnvGoFunctionConfigs, str),
		"&&",
		fmt.Sprintf("goFunctionConfigs=${%s}", EnvGoFunctionConfigs),
		"&&",
		"echo goFunctionConfigs=\"'${goFunctionConfigs}'\"",
		"&&",
		"chmod +x",
		goExecFilePath,
		"&&",
		"exec",
		goExecFilePath,
		"-instance-conf",
		"${goFunctionConfigs}",
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
	if maxMessageRetry <= 0 && deadLetterTopic == "" {
		return nil
	}
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

func getUserConfig(configs *v1alpha1.Config) string {
	if configs == nil {
		return ""
	}
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

func generateContainerEnvFrom(messagingConfig string, authSecret string, tlsSecret string) []corev1.EnvFromSource {
	envs := []corev1.EnvFromSource{{
		ConfigMapRef: &corev1.ConfigMapEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: messagingConfig},
		},
	}}

	if authSecret != "" {
		envs = append(envs, corev1.EnvFromSource{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: authSecret},
			},
		})
	}

	if tlsSecret != "" {
		envs = append(envs, corev1.EnvFromSource{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: tlsSecret},
			},
		})
	}

	return envs
}

func generateContainerVolumesFromConsumerConfigs(confs map[string]v1alpha1.ConsumerConfig) []corev1.Volume {
	volumes := []corev1.Volume{}
	if len(confs) > 0 {
		for _, conf := range confs {
			if conf.CryptoConfig != nil && len(conf.CryptoConfig.CryptoSecrets) > 0 {
				for _, c := range conf.CryptoConfig.CryptoSecrets {
					volumes = append(volumes, generateVolumeFromCryptoSecret(&c))
				}
			}
		}
	}
	return volumes
}

func generateContainerVolumesFromProducerConf(conf *v1alpha1.ProducerConfig) []corev1.Volume {
	volumes := []corev1.Volume{}
	if conf != nil && conf.CryptoConfig != nil && len(conf.CryptoConfig.CryptoSecrets) > 0 {
		for _, c := range conf.CryptoConfig.CryptoSecrets {
			volumes = append(volumes, generateVolumeFromCryptoSecret(&c))
		}
	}
	return volumes
}

func generateVolumeFromCryptoSecret(secret *v1alpha1.CryptoSecret) corev1.Volume {
	return corev1.Volume{
		Name: generateVolumeNameFromCryptoSecrets(secret),
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secret.SecretName,
				Items: []corev1.KeyToPath{
					{
						Key:  secret.SecretKey,
						Path: secret.SecretKey,
					},
				},
			},
		},
	}
}

func generateVolumeMountFromCryptoSecret(secret *v1alpha1.CryptoSecret) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      generateVolumeNameFromCryptoSecrets(secret),
		MountPath: secret.AsVolume,
	}
}

func generateContainerVolumeMountsFromConsumerConfigs(confs map[string]v1alpha1.ConsumerConfig) []corev1.VolumeMount {
	mounts := []corev1.VolumeMount{}
	if len(confs) > 0 {
		for _, conf := range confs {
			if conf.CryptoConfig != nil && len(conf.CryptoConfig.CryptoSecrets) > 0 {
				for _, c := range conf.CryptoConfig.CryptoSecrets {
					if c.AsVolume != "" {
						mounts = append(mounts, generateVolumeMountFromCryptoSecret(&c))
					}
				}
			}
		}
	}
	return mounts
}

func generateContainerVolumeMountsFromProducerConf(conf *v1alpha1.ProducerConfig) []corev1.VolumeMount {
	mounts := []corev1.VolumeMount{}
	if conf != nil && conf.CryptoConfig != nil && len(conf.CryptoConfig.CryptoSecrets) > 0 {
		for _, c := range conf.CryptoConfig.CryptoSecrets {
			if c.AsVolume != "" {
				mounts = append(mounts, generateVolumeMountFromCryptoSecret(&c))
			}
		}
	}
	return mounts
}

func generateContainerVolumeMounts(volumeMounts []corev1.VolumeMount, producerConf *v1alpha1.ProducerConfig,
	consumerConfs map[string]v1alpha1.ConsumerConfig) []corev1.VolumeMount {
	mounts := []corev1.VolumeMount{}
	mounts = append(mounts, volumeMounts...)
	mounts = append(mounts, generateContainerVolumeMountsFromProducerConf(producerConf)...)
	mounts = append(mounts, generateContainerVolumeMountsFromConsumerConfigs(consumerConfs)...)
	return mounts
}

func generatePodVolumes(podVolumes []corev1.Volume, producerConf *v1alpha1.ProducerConfig,
	consumerConfs map[string]v1alpha1.ConsumerConfig) []corev1.Volume {
	volumes := []corev1.Volume{}
	volumes = append(volumes, podVolumes...)
	volumes = append(volumes, generateContainerVolumesFromProducerConf(producerConf)...)
	volumes = append(volumes, generateContainerVolumesFromConsumerConfigs(consumerConfs)...)
	return volumes
}

func mergeLabels(label1, label2 map[string]string) map[string]string {
	label := make(map[string]string)

	for k, v := range label1 {
		label[k] = v
	}

	for k, v := range label2 {
		label[k] = v
	}

	return label
}

func generateAnnotations(customAnnotations map[string]string) map[string]string {
	annotations := make(map[string]string)

	// controlled annotations
	annotations[AnnotationPrometheusScrape] = "true"
	annotations[AnnotationPrometheusPort] = strconv.Itoa(int(MetricsPort.ContainerPort))

	// customized annotations which may override any previous set annotations
	for k, v := range customAnnotations {
		annotations[k] = v
	}

	return annotations
}

func getFunctionRunnerImage(spec *v1alpha1.FunctionSpec) string {
	runtime := &spec.Runtime
	img := spec.Image
	if img != "" {
		return img
	} else if runtime.Java != nil && runtime.Java.Jar != "" {
		return DefaultJavaRunnerImage
	} else if runtime.Python != nil && runtime.Python.Py != "" {
		return DefaultPythonRunnerImage
	} else if runtime.Golang != nil && runtime.Golang.Go != "" {
		return DefaultGoRunnerImage
	}
	return DefaultRunnerImage
}

func getSinkRunnerImage(spec *v1alpha1.SinkSpec) string {
	img := spec.Image
	if img != "" {
		return img
	}
	if spec.Runtime.Java.Jar != "" && spec.Runtime.Java.JarLocation != "" &&
		hasPackageNamePrefix(spec.Runtime.Java.JarLocation) {
		return DefaultJavaRunnerImage
	}
	return DefaultRunnerImage
}

func getSourceRunnerImage(spec *v1alpha1.SourceSpec) string {
	img := spec.Image
	if img != "" {
		return img
	}
	if spec.Runtime.Java.Jar != "" && spec.Runtime.Java.JarLocation != "" &&
		hasPackageNamePrefix(spec.Runtime.Java.JarLocation) {
		return DefaultJavaRunnerImage
	}
	return DefaultRunnerImage
}
