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

	"github.com/streamnative/function-mesh/api/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/proto"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	EnvShardID                 = "SHARD_ID"
	FunctionsInstanceClasspath = "pulsar.functions.instance.classpath"
	DefaultRunnerTag           = "2.10.0.0-rc10"
	DefaultRunnerPrefix        = "streamnative/"
	DefaultRunnerImage         = DefaultRunnerPrefix + "pulsar-all:" + DefaultRunnerTag
	DefaultJavaRunnerImage     = DefaultRunnerPrefix + "pulsar-functions-java-runner:" + DefaultRunnerTag
	DefaultPythonRunnerImage   = DefaultRunnerPrefix + "pulsar-functions-python-runner:" + DefaultRunnerTag
	DefaultGoRunnerImage       = DefaultRunnerPrefix + "pulsar-functions-go-runner:" + DefaultRunnerTag
	PulsarAdminExecutableFile  = "/pulsar/bin/pulsar-admin"

	DefaultForAllowInsecure              = "false"
	DefaultForEnableHostNameVerification = "true"

	AppFunctionMesh   = "function-mesh"
	ComponentSource   = "source"
	ComponentSink     = "sink"
	ComponentFunction = "function"

	PackageNameFunctionPrefix = "function://"
	PackageNameSinkPrefix     = "sink://"
	PackageNameSourcePrefix   = "source://"

	AnnotationPrometheusScrape = "prometheus.io/scrape"
	AnnotationPrometheusPort   = "prometheus.io/port"
	AnnotationManaged          = "compute.functionmesh.io/managed"

	EnvGoFunctionConfigs = "GO_FUNCTION_CONF"

	DefaultRunnerUserID  int64 = 10000
	DefaultRunnerGroupID int64 = 10001
)

var GRPCPort = corev1.ContainerPort{
	Name:          "tcp-grpc",
	ContainerPort: 9093,
	Protocol:      corev1.ProtocolTCP,
}

var MetricsPort = corev1.ContainerPort{
	Name:          "tcp-metrics",
	ContainerPort: 9094,
	Protocol:      corev1.ProtocolTCP,
}

func IsManaged(object metav1.Object) bool {
	managed, exists := object.GetAnnotations()[AnnotationManaged]
	return !exists || managed != "false"
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
			Ports:     []corev1.ServicePort{toServicePort(&GRPCPort), toServicePort(&MetricsPort)},
			Selector:  labels,
			ClusterIP: corev1.ClusterIPNone,
		},
	}
}

// MakeHeadlessServiceName changes the name of service to headless style
func MakeHeadlessServiceName(serviceName string) string {
	return fmt.Sprintf("%s-headless", serviceName)
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
	podSecurityContext := getDefaultRunnerPodSecurityContext(DefaultRunnerUserID, DefaultRunnerGroupID, false)
	if policy.SecurityContext != nil {
		podSecurityContext = policy.SecurityContext
	}
	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      mergeLabels(labels, Configs.ResourceLabels, policy.Labels),
			Annotations: generateAnnotations(Configs.ResourceAnnotations, policy.Annotations),
		},
		Spec: corev1.PodSpec{
			InitContainers:                policy.InitContainers,
			Containers:                    append(policy.Sidecars, *container),
			TerminationGracePeriodSeconds: &policy.TerminationGracePeriodSeconds,
			Volumes:                       volumes,
			NodeSelector:                  policy.NodeSelector,
			Affinity:                      policy.Affinity,
			Tolerations:                   policy.Tolerations,
			SecurityContext:               podSecurityContext,
			ImagePullSecrets:              policy.ImagePullSecrets,
			ServiceAccountName:            policy.ServiceAccountName,
		},
	}
}

func MakeJavaFunctionCommand(downloadPath, packageFile, name, clusterName, details, memory, extraDependenciesDir, uid string,
	authProvided, tlsProvided bool, secretMaps map[string]v1alpha1.SecretRef, state *v1alpha1.Stateful, trustCert v1alpha1.CryptoSecret) []string {
	processCommand := setShardIDEnvironmentVariableCommand() + " && " +
		strings.Join(getProcessJavaRuntimeArgs(name, packageFile, clusterName, details,
			memory, extraDependenciesDir, uid, authProvided, tlsProvided, secretMaps, state, trustCert), " ")
	if downloadPath != "" {
		// prepend download command if the downPath is provided
		downloadCommand := strings.Join(getDownloadCommand(downloadPath, packageFile, authProvided, tlsProvided, trustCert), " ")
		processCommand = downloadCommand + " && " + processCommand
	}
	return []string{"sh", "-c", processCommand}
}

func MakePythonFunctionCommand(downloadPath, packageFile, name, clusterName, details, uid string,
	authProvided, tlsProvided bool, secretMaps map[string]v1alpha1.SecretRef, state *v1alpha1.Stateful, trustCert v1alpha1.CryptoSecret) []string {
	processCommand := setShardIDEnvironmentVariableCommand() + " && " +
		strings.Join(getProcessPythonRuntimeArgs(name, packageFile, clusterName,
			details, uid, authProvided, tlsProvided, secretMaps, state, trustCert), " ")
	if downloadPath != "" {
		// prepend download command if the downPath is provided
		downloadCommand := strings.Join(getDownloadCommand(downloadPath, packageFile, authProvided, tlsProvided, trustCert), " ")
		processCommand = downloadCommand + " && " + processCommand
	}
	return []string{"sh", "-c", processCommand}
}

func MakeGoFunctionCommand(downloadPath, goExecFilePath string, function *v1alpha1.Function) []string {
	processCommand := setShardIDEnvironmentVariableCommand() + " && " +
		strings.Join(getProcessGoRuntimeArgs(goExecFilePath, function), " ")
	if downloadPath != "" {
		// prepend download command if the downPath is provided
		downloadCommand := strings.Join(getDownloadCommand(downloadPath, goExecFilePath,
			function.Spec.Pulsar.AuthSecret != "", function.Spec.Pulsar.TLSSecret != "", function.Spec.TLSTrustCert), " ")
		processCommand = downloadCommand + " && ls -al && pwd &&" + processCommand
	}
	return []string{"sh", "-c", processCommand}
}

func getDownloadCommand(downloadPath, componentPackage string, authProvided, tlsProvided bool, trustCert v1alpha1.CryptoSecret) []string {
	// The download path is the path that the package saved in the pulsar.
	// By default, it's the path that the package saved in the pulsar, we can use package name
	// to replace it for downloading packages from packages management service.
	args := []string{
		PulsarAdminExecutableFile,
		"--admin-url",
		"$webServiceURL",
	}
	if authProvided {
		args = append(args, []string{
			"--auth-plugin",
			"$clientAuthenticationPlugin",
			"--auth-params",
			"$clientAuthenticationParameters"}...)
	}

	if tlsProvided {
		args = append(args, []string{
			"--tls-allow-insecure",
			"${tlsAllowInsecureConnection:-" + DefaultForAllowInsecure + "}",
			"--tls-enable-hostname-verification",
			"${tlsHostnameVerificationEnable:-" + DefaultForEnableHostNameVerification + "}",
		}...)

		if trustCert.SecretName != "" {
			args = append(args, []string{
				"--tls-trust-cert-path",
				getTLSTrustCertPath(trustCert),
			}...)
		}
	}
	if hasPackageNamePrefix(downloadPath) {
		args = append(args, []string{
			"packages",
			"download",
			downloadPath,
			"--path",
			componentPackage,
		}...)
		return args
	}
	args = append(args, []string{
		"functions",
		"download",
		"--path",
		downloadPath,
		"--destination-file",
		componentPackage,
	}...)
	return args
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

func getProcessJavaRuntimeArgs(name, packageName, clusterName, details, memory, extraDependenciesDir, uid string,
	authProvided, tlsProvided bool, secretMaps map[string]v1alpha1.SecretRef, state *v1alpha1.Stateful, trustCert v1alpha1.CryptoSecret) []string {
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
	sharedArgs := getSharedArgs(details, clusterName, uid, authProvided, tlsProvided, trustCert)
	args = append(args, sharedArgs...)
	if len(secretMaps) > 0 {
		secretProviderArgs := getJavaSecretProviderArgs(secretMaps)
		args = append(args, secretProviderArgs...)
	}
	if state != nil && state.Pulsar != nil && state.Pulsar.ServiceURL != "" {
		statefulArgs := []string{
			"--state_storage_serviceurl",
			state.Pulsar.ServiceURL,
		}
		if state.Pulsar.JavaProvider != nil {
			statefulArgs = append(statefulArgs, "--state_storage_impl_class", state.Pulsar.JavaProvider.ClassName)
		}
		args = append(args, statefulArgs...)
	}
	return args
}

func getProcessPythonRuntimeArgs(name, packageName, clusterName, details, uid string, authProvided, tlsProvided bool,
	secretMaps map[string]v1alpha1.SecretRef, state *v1alpha1.Stateful, trustCert v1alpha1.CryptoSecret) []string {
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
	sharedArgs := getSharedArgs(details, clusterName, uid, authProvided, tlsProvided, trustCert)
	args = append(args, sharedArgs...)
	if len(secretMaps) > 0 {
		secretProviderArgs := getPythonSecretProviderArgs(secretMaps)
		args = append(args, secretProviderArgs...)
	}
	if state != nil && state.Pulsar != nil && state.Pulsar.ServiceURL != "" {
		statefulArgs := []string{
			"--state_storage_serviceurl",
			state.Pulsar.ServiceURL,
		}
		args = append(args, statefulArgs...)
	}
	return args
}

// This method is suitable for Java and Python runtime, not include Go runtime.
func getSharedArgs(details, clusterName, uid string, authProvided bool, tlsProvided bool, trustCert v1alpha1.CryptoSecret) []string {
	args := []string{
		"--instance_id",
		"${" + EnvShardID + "}",
		"--function_id",
		fmt.Sprintf("${%s}-%s", EnvShardID, uid),
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
			"${tlsAllowInsecureConnection:-" + DefaultForAllowInsecure + "}",
			"--hostname_verification_enabled",
			"${tlsHostnameVerificationEnable:-" + DefaultForEnableHostNameVerification + "}",
		}...)

		// only set --tls_trust_cert_path when it's mounted
		if trustCert.SecretName != "" {
			args = append(args, []string{
				"--tls_trust_cert_path",
				getTLSTrustCertPath(trustCert),
			}...)
		}
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
	j, err := json.Marshal(goFunctionConfs)
	if err != nil {
		// TODO
		panic(err)
	}
	ret := string(j)
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
		"ls -l",
		goExecFilePath,
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
	consumerConfs map[string]v1alpha1.ConsumerConfig, trustCert v1alpha1.CryptoSecret) []corev1.VolumeMount {
	mounts := []corev1.VolumeMount{}
	mounts = append(mounts, volumeMounts...)
	if trustCert.SecretName != "" && trustCert.AsVolume != "" {
		mounts = append(mounts, generateVolumeMountFromCryptoSecret(&trustCert))
	}
	mounts = append(mounts, generateContainerVolumeMountsFromProducerConf(producerConf)...)
	mounts = append(mounts, generateContainerVolumeMountsFromConsumerConfigs(consumerConfs)...)
	return mounts
}

func generatePodVolumes(podVolumes []corev1.Volume, producerConf *v1alpha1.ProducerConfig,
	consumerConfs map[string]v1alpha1.ConsumerConfig, trustCert v1alpha1.CryptoSecret) []corev1.Volume {
	volumes := []corev1.Volume{}
	volumes = append(volumes, podVolumes...)
	if trustCert.SecretName != "" {
		volumes = append(volumes, generateVolumeFromCryptoSecret(&trustCert))
	}
	volumes = append(volumes, generateContainerVolumesFromProducerConf(producerConf)...)
	volumes = append(volumes, generateContainerVolumesFromConsumerConfigs(consumerConfs)...)
	return volumes
}

func mergeLabels(labels ...map[string]string) map[string]string {
	merged := make(map[string]string)

	for _, m := range labels {
		for k, v := range m {
			merged[k] = v
		}
	}

	return merged
}

func generateAnnotations(customAnnotations ...map[string]string) map[string]string {
	annotations := make(map[string]string)

	// controlled annotations
	annotations[AnnotationPrometheusScrape] = "true"
	annotations[AnnotationPrometheusPort] = strconv.Itoa(int(MetricsPort.ContainerPort))

	// customized annotations which may override any previous set annotations
	for _, custom := range customAnnotations {
		for k, v := range custom {
			annotations[k] = v
		}
	}

	return annotations
}

func getFunctionRunnerImage(spec *v1alpha1.FunctionSpec) string {
	runtime := &spec.Runtime
	img := spec.Image
	if img != "" {
		return img
	} else if runtime.Java != nil && runtime.Java.Jar != "" {
		return Configs.RunnerImages.Java
	} else if runtime.Python != nil && runtime.Python.Py != "" {
		return Configs.RunnerImages.Python
	} else if runtime.Golang != nil && runtime.Golang.Go != "" {
		return Configs.RunnerImages.Go
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
		return Configs.RunnerImages.Java
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
		return Configs.RunnerImages.Java
	}
	return DefaultRunnerImage
}

// getDefaultRunnerPodSecurityContext returns a default PodSecurityContext that runs as non-root
func getDefaultRunnerPodSecurityContext(uid, gid int64, nonRoot bool) *corev1.PodSecurityContext {
	return &corev1.PodSecurityContext{
		RunAsUser:    &uid,
		RunAsGroup:   &gid,
		RunAsNonRoot: &nonRoot,
		FSGroup:      &gid,
	}
}

func getJavaSecretProviderArgs(secretMaps map[string]v1alpha1.SecretRef) []string {
	var ret []string
	if len(secretMaps) > 0 {
		ret = []string{
			"--secrets_provider",
			"org.apache.pulsar.functions.secretsprovider.EnvironmentBasedSecretsProvider",
		}
	}
	return ret
}

func getPythonSecretProviderArgs(secretMaps map[string]v1alpha1.SecretRef) []string {
	var ret []string
	if len(secretMaps) > 0 {
		ret = []string{
			"--secrets_provider",
			"secretsprovider.EnvironmentBasedSecretsProvider",
		}
	}
	return ret
}

// Java command requires memory values in resource.DecimalSI format
func getDecimalSIMemory(quantity *resource.Quantity) string {
	if quantity.Format == resource.DecimalSI {
		return quantity.String()
	}
	return resource.NewQuantity(quantity.Value(), resource.DecimalSI).String()
}

func getTLSTrustCertPath(trustCert v1alpha1.CryptoSecret) string {
	return fmt.Sprintf("%s/%s", trustCert.AsVolume, trustCert.SecretKey)
}
