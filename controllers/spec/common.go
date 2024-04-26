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
	"bytes"
	"context"

	// used for template
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	autov2 "k8s.io/api/autoscaling/v2"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/proto"
	"github.com/streamnative/function-mesh/utils"
	pctlutil "github.com/streamnative/pulsarctl/pkg/pulsar/utils"
)

const (
	EnvShardID                      = "SHARD_ID"
	FunctionsInstanceClasspath      = "pulsar.functions.instance.classpath"
	DefaultRunnerTag                = "2.10.0.0-rc10"
	DefaultGenericRunnerTag         = "0.1.0"
	DefaultRunnerPrefix             = "streamnative/"
	DefaultRunnerImage              = DefaultRunnerPrefix + "pulsar-all:" + DefaultRunnerTag
	DefaultJavaRunnerImage          = DefaultRunnerPrefix + "pulsar-functions-java-runner:" + DefaultRunnerTag
	DefaultPythonRunnerImage        = DefaultRunnerPrefix + "pulsar-functions-python-runner:" + DefaultRunnerTag
	DefaultGoRunnerImage            = DefaultRunnerPrefix + "pulsar-functions-go-runner:" + DefaultRunnerTag
	DefaultGenericNodejsRunnerImage = DefaultRunnerPrefix + "pulsar-functions-generic-nodejs-runner:" + DefaultGenericRunnerTag
	DefaultGenericPythonRunnerImage = DefaultRunnerPrefix + "pulsar-functions-generic-python-runner:" + DefaultGenericRunnerTag
	DefaultGenericRunnerImage       = DefaultRunnerPrefix + "pulsar-functions-generic-base-runner:" + DefaultGenericRunnerTag
	PulsarAdminExecutableFile       = "/pulsar/bin/pulsar-admin"
	WorkDir                         = "/pulsar/"

	RunnerImageHasPulsarctl = "pulsar-functions-(pulsarctl|sn|generic)-(java|python|go|nodejs|base)-runner"

	PulsarctlExecutableFile = "pulsarctl"
	DownloaderName          = "downloader"
	DownloaderVolume        = "downloader-volume"
	DownloaderImage         = DefaultRunnerPrefix + "pulsarctl:2.10.2.3"
	DownloadDir             = "/pulsar/download"

	CleanupContainerName = "cleanup"

	WindowFunctionConfigKeyName = "__WINDOWCONFIGS__"
	WindowFunctionExecutorClass = "org.apache.pulsar.functions.windowing.WindowFunctionExecutor"

	DefaultForAllowInsecure              = "false"
	DefaultForEnableHostNameVerification = "true"

	AppFunctionMesh   = "function-mesh"
	ComponentSource   = "source"
	ComponentSink     = "sink"
	ComponentFunction = "function"

	PackageNameFunctionPrefix = "function://"
	PackageNameSinkPrefix     = "sink://"
	PackageNameSourcePrefix   = "source://"

	HTTPPrefix  = "http://"
	HTTPSPrefix = "https://"

	AnnotationPrometheusScrape = "prometheus.io/scrape"
	AnnotationPrometheusPort   = "prometheus.io/port"
	AnnotationManaged          = "compute.functionmesh.io/managed"
	AnnotationNeedCleanup      = "compute.functionmesh.io/need-cleanup"

	// if labels contains below, we think it comes from function-mesh-worker-service
	LabelPulsarCluster           = "compute.functionmesh.io/pulsar-cluster"
	LabelPulsarClusterDeprecated = "pulsar-cluster"

	EnvGoFunctionConfigs = "GO_FUNCTION_CONF"

	DefaultRunnerUserID  int64 = 10000
	DefaultRunnerGroupID int64 = 10001

	OAuth2AuthenticationPlugin = "org.apache.pulsar.client.impl.auth.oauth2.AuthenticationOAuth2"
	TokenAuthenticationPlugin  = "org.apache.pulsar.client.impl.auth.AuthenticationToken"

	JavaLogConfigDirectory       = "/pulsar/conf/java-log/"
	JavaLogConfigFileXML         = "java_instance_log4j.xml"
	JavaLogConfigFileYAML        = "java_instance_log4j.yaml"
	DefaultJavaLogConfigPath     = JavaLogConfigDirectory + JavaLogConfigFileXML
	DefaultJavaLogConfigPathYAML = JavaLogConfigDirectory + JavaLogConfigFileYAML
	PythonLogConifgDirectory     = "/pulsar/conf/python-log/"
	PythonLogConfigFile          = "python_instance_logging.ini"
	DefaultPythonLogConfigPath   = PythonLogConifgDirectory + PythonLogConfigFile

	DefaultFilebeatConfig = "/usr/share/filebeat/config/filebeat.yaml"
	DefaultFilebeatImage  = "streamnative/filebeat:v0.6.0-rc7"

	EnvGoFunctionLogLevel = "LOGGING_LEVEL"
)

//go:embed template/java-runtime-log4j.xml.tmpl
var javaLog4jXMLTemplate string

//go:embed template/java-runtime-log4j.yaml.tmpl
var javaLog4jYAMLTemplate string

//go:embed template/python-runtime-log-config.ini.tmpl
var pythonLoggingINITemplate string

//go:embed template/filebeat-config.yaml.tmpl
var filebeatYAMLTemplate string

var GRPCPort = corev1.ContainerPort{
	Name:          "tcp-grpc",
	ContainerPort: 9093,
	Protocol:      corev1.ProtocolTCP,
}

var MetricsPort = corev1.ContainerPort{
	Name:          "http-metrics",
	ContainerPort: 9094,
	Protocol:      corev1.ProtocolTCP,
}

type TLSConfig interface {
	IsEnabled() bool
	AllowInsecureConnection() string
	EnableHostnameVerification() string
	SecretName() string
	SecretKey() string
	HasSecretVolume() bool
	GetMountPath() string
}

func IsManaged(object metav1.Object) bool {
	managed, exists := object.GetAnnotations()[AnnotationManaged]
	return !exists || managed != "false"
}

func NeedCleanup(object metav1.Object) bool {
	// don't cleanup if it's managed by function-mesh-worker-service
	_, exists := object.GetLabels()[LabelPulsarCluster]
	if exists {
		return false
	}
	_, exists = object.GetLabels()[LabelPulsarClusterDeprecated]
	if exists {
		return false
	}

	needCleanup, exists := object.GetAnnotations()[AnnotationNeedCleanup]
	return !exists || needCleanup != "false"
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

func MakeStatefulSet(objectMeta *metav1.ObjectMeta, replicas *int32, downloaderImage string,
	container *corev1.Container, filebeatContainer *corev1.Container,
	volumes []corev1.Volume, labels map[string]string, policy v1alpha1.PodPolicy, pulsar v1alpha1.PulsarMessaging,
	javaRuntime *v1alpha1.JavaRuntime, pythonRuntime *v1alpha1.PythonRuntime,
	goRuntime *v1alpha1.GoRuntime, definedVolumeMounts []corev1.VolumeMount,
	volumeClaimTemplates []corev1.PersistentVolumeClaim,
	persistentVolumeClaimRetentionPolicy *appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy) *appsv1.StatefulSet {

	volumeMounts := generateDownloaderVolumeMountsForDownloader(javaRuntime, pythonRuntime, goRuntime)
	var downloaderContainer *corev1.Container
	var podVolumes = volumes
	// there must be a download path specified, we need to create an init container and emptyDir volume
	if len(volumeMounts) > 0 {
		var downloadPath, componentPackage string
		if javaRuntime != nil {
			downloadPath = javaRuntime.JarLocation
			componentPackage = javaRuntime.Jar
		} else if pythonRuntime != nil {
			downloadPath = pythonRuntime.PyLocation
			componentPackage = pythonRuntime.Py
		} else {
			downloadPath = goRuntime.GoLocation
			componentPackage = goRuntime.Go
		}

		// mount auth and tls related VolumeMounts when download package from pulsar
		if !hasHTTPPrefix(downloadPath) {
			if pulsar.AuthConfig != nil && pulsar.AuthConfig.OAuth2Config != nil {
				volumeMounts = append(volumeMounts, generateVolumeMountFromOAuth2Config(pulsar.AuthConfig.OAuth2Config))
			}

			if !reflect.ValueOf(pulsar.TLSConfig).IsNil() && pulsar.TLSConfig.HasSecretVolume() {
				volumeMounts = append(volumeMounts, generateVolumeMountFromTLSConfig(pulsar.TLSConfig))
			}
		}
		volumeMounts = append(volumeMounts, definedVolumeMounts...)

		image := downloaderImage
		if image == "" {
			image = DownloaderImage
		}

		componentPackage = fmt.Sprintf("%s/%s", DownloadDir, getFilenameOfComponentPackage(componentPackage))

		downloaderContainer = &corev1.Container{
			Name:  DownloaderName,
			Image: image,
			Command: []string{"sh", "-c",
				strings.Join(getDownloadCommand(downloadPath, componentPackage, true, true,
					pulsar.AuthSecret != "", pulsar.TLSSecret != "", pulsar.TLSConfig, pulsar.AuthConfig), " ")},
			VolumeMounts:    volumeMounts,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Env: []corev1.EnvVar{{
				Name:  "HOME",
				Value: "/tmp",
			}},
			EnvFrom: generateContainerEnvFrom(pulsar.PulsarConfig, pulsar.AuthSecret, pulsar.TLSSecret),
		}
		podVolumes = append(podVolumes, corev1.Volume{
			Name: DownloaderVolume,
		})
	}
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: *objectMeta,
		Spec: *MakeStatefulSetSpec(replicas, container, filebeatContainer, podVolumes, labels, policy,
			MakeHeadlessServiceName(objectMeta.Name), downloaderContainer, volumeClaimTemplates,
			persistentVolumeClaimRetentionPolicy),
	}
}

// PatchStatefulSet Apply global and namespaced configs to StatefulSet
func PatchStatefulSet(ctx context.Context, cli client.Client, namespace string, statefulSet *appsv1.StatefulSet) (string, string, error) {
	globalBackendConfigVersion := ""
	namespacedBackendConfigVersion := ""
	envData := make(map[string]string)

	if utils.GlobalBackendConfig != "" && utils.GlobalBackendConfigNamespace != "" {
		globalBackendConfig := &v1alpha1.BackendConfig{}
		err := cli.Get(ctx, types.NamespacedName{
			Namespace: utils.GlobalBackendConfigNamespace,
			Name:      utils.GlobalBackendConfig,
		}, globalBackendConfig)
		if err != nil {
			// ignore not found error
			if !k8serrors.IsNotFound(err) {
				return "", "", err
			}
		} else {
			globalBackendConfigVersion = globalBackendConfig.ResourceVersion
			for key, val := range globalBackendConfig.Spec.Env {
				envData[key] = val
			}
		}
	}

	// patch namespaced configs
	if utils.NamespacedBackendConfig != "" {
		namespacedBackendConfig := &v1alpha1.BackendConfig{}
		err := cli.Get(ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      utils.NamespacedBackendConfig,
		}, namespacedBackendConfig)
		if err != nil {
			// ignore not found error
			if !k8serrors.IsNotFound(err) {
				return "", "", err
			}
		} else {
			namespacedBackendConfigVersion = namespacedBackendConfig.ResourceVersion
			for key, val := range namespacedBackendConfig.Spec.Env {
				envData[key] = val
			}
		}
	}

	// merge env
	if len(envData) == 0 {
		return globalBackendConfigVersion, namespacedBackendConfigVersion, nil
	}
	globalEnvs := make([]corev1.EnvVar, 0, len(envData))
	for key, val := range envData {
		globalEnvs = append(globalEnvs, corev1.EnvVar{
			Name:  key,
			Value: val,
		})
	}
	for i := range statefulSet.Spec.Template.Spec.Containers {
		statefulSet.Spec.Template.Spec.Containers[i].Env = append(statefulSet.Spec.Template.Spec.Containers[i].Env,
			globalEnvs...)
	}

	return globalBackendConfigVersion, namespacedBackendConfigVersion, nil
}

func MakeStatefulSetSpec(replicas *int32, container *corev1.Container, filebeatContainer *corev1.Container,
	volumes []corev1.Volume, labels map[string]string, policy v1alpha1.PodPolicy,
	serviceName string, downloaderContainer *corev1.Container, volumeClaimTemplates []corev1.PersistentVolumeClaim,
	persistentVolumeClaimRetentionPolicy *appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy) *appsv1.StatefulSetSpec {
	spec := &appsv1.StatefulSetSpec{
		Replicas: replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		Template:            *makePodTemplate(container, filebeatContainer, volumes, labels, policy, downloaderContainer),
		PodManagementPolicy: appsv1.ParallelPodManagement,
		UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
			Type: appsv1.RollingUpdateStatefulSetStrategyType,
		},
		ServiceName: serviceName,
	}
	if len(volumeClaimTemplates) > 0 {
		spec.VolumeClaimTemplates = volumeClaimTemplates
	}
	if persistentVolumeClaimRetentionPolicy != nil {
		spec.PersistentVolumeClaimRetentionPolicy = persistentVolumeClaimRetentionPolicy
	}
	return spec
}

func makePodTemplate(container *corev1.Container, filebeatContainer *corev1.Container, volumes []corev1.Volume,
	labels map[string]string, policy v1alpha1.PodPolicy,
	downloaderContainer *corev1.Container) *corev1.PodTemplateSpec {
	podSecurityContext := getDefaultRunnerPodSecurityContext(DefaultRunnerUserID, DefaultRunnerGroupID,
		getEnvOrDefault("RUN_AS_NON_ROOT", "false"))
	if policy.SecurityContext != nil {
		podSecurityContext = policy.SecurityContext
	}
	initContainers := policy.InitContainers
	if downloaderContainer != nil {
		initContainers = append(initContainers, *downloaderContainer)
	}
	containers := append(policy.Sidecars, *container)
	if filebeatContainer != nil {
		containers = append(containers, *filebeatContainer)
	}
	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      mergeLabels(labels, Configs.ResourceLabels, policy.Labels),
			Annotations: generateAnnotations(Configs.ResourceAnnotations, policy.Annotations),
		},
		Spec: corev1.PodSpec{
			InitContainers:                initContainers,
			Containers:                    containers,
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

func MakeJavaFunctionCommand(downloadPath, packageFile, name, clusterName, generateLogConfigCommand, logLevel, details, extraDependenciesDir, uid string,
	javaOpts []string, hasPulsarctl, hasWget, authProvided, tlsProvided bool, secretMaps map[string]v1alpha1.SecretRef,
	state *v1alpha1.Stateful,
	tlsConfig TLSConfig, authConfig *v1alpha1.AuthConfig,
	maxPendingAsyncRequests *int32, logConfigFileName string) []string {
	processCommand := setShardIDEnvironmentVariableCommand() + " && " + generateLogConfigCommand +
		strings.Join(getProcessJavaRuntimeArgs(name, packageFile, clusterName, logLevel, details,
			extraDependenciesDir, uid, javaOpts, authProvided, tlsProvided, secretMaps, state, tlsConfig,
			authConfig, maxPendingAsyncRequests, logConfigFileName), " ")
	if downloadPath != "" && !utils.EnableInitContainers {
		// prepend download command if the downPath is provided
		downloadCommand := strings.Join(getDownloadCommand(downloadPath, packageFile, hasPulsarctl, hasWget,
			authProvided, tlsProvided, tlsConfig, authConfig), " ")
		processCommand = downloadCommand + " && " + processCommand
	}
	return []string{"bash", "-c", processCommand}
}

func MakePythonFunctionCommand(downloadPath, packageFile, name, clusterName, generateLogConfigCommand, details, uid string,
	hasPulsarctl, hasWget, authProvided, tlsProvided bool, secretMaps map[string]v1alpha1.SecretRef,
	state *v1alpha1.Stateful,
	tlsConfig TLSConfig, authConfig *v1alpha1.AuthConfig) []string {
	processCommand := setShardIDEnvironmentVariableCommand() + " && " + generateLogConfigCommand +
		strings.Join(getProcessPythonRuntimeArgs(name, packageFile, clusterName,
			details, uid, authProvided, tlsProvided, secretMaps, state, tlsConfig, authConfig), " ")
	if downloadPath != "" && !utils.EnableInitContainers {
		// prepend download command if the downPath is provided
		downloadCommand := strings.Join(getDownloadCommand(downloadPath, packageFile, hasPulsarctl, hasWget,
			authProvided,
			tlsProvided, tlsConfig, authConfig), " ")
		processCommand = downloadCommand + " && " + processCommand
	}
	return []string{"bash", "-c", processCommand}
}

func MakeGoFunctionCommand(downloadPath, goExecFilePath string, function *v1alpha1.Function) []string {
	processCommand := setShardIDEnvironmentVariableCommand() + " && " +
		strings.Join(getProcessGoRuntimeArgs(goExecFilePath, function), " ")
	if downloadPath != "" && !utils.EnableInitContainers {
		// prepend download command if the downPath is provided
		hasPulsarctl := function.Spec.ImageHasPulsarctl
		hasWget := function.Spec.ImageHasWget
		if match, _ := regexp.MatchString(RunnerImageHasPulsarctl, function.Spec.Image); match {
			hasPulsarctl = true
			hasWget = true
		}
		downloadCommand := strings.Join(getDownloadCommand(downloadPath, goExecFilePath,
			hasPulsarctl, hasWget, function.Spec.Pulsar.AuthSecret != "",
			function.Spec.Pulsar.TLSSecret != "", function.Spec.Pulsar.TLSConfig, function.Spec.Pulsar.AuthConfig), " ")
		processCommand = downloadCommand + " && ls -al && pwd &&" + processCommand
	}
	return []string{"bash", "-c", processCommand}
}

func MakeGenericFunctionCommand(downloadPath, functionFile, language, clusterName, details, uid string, authProvided, tlsProvided bool, secretMaps map[string]v1alpha1.SecretRef,
	state *v1alpha1.Stateful,
	tlsConfig TLSConfig, authConfig *v1alpha1.AuthConfig) []string {
	processCommand := setShardIDEnvironmentVariableCommand() + " && " +
		strings.Join(getProcessGenericRuntimeArgs(language, functionFile, clusterName,
			details, uid, authProvided, tlsProvided, secretMaps, state, tlsConfig, authConfig), " ")
	if downloadPath != "" && !utils.EnableInitContainers {
		// prepend download command if the downPath is provided
		downloadCommand := strings.Join(getDownloadCommand(downloadPath, functionFile, true, true,
			authProvided,
			tlsProvided, tlsConfig, authConfig), " ")
		processCommand = downloadCommand + " && " + processCommand
	}
	return []string{"sh", "-c", processCommand}
}

func MakeLivenessProbe(liveness *v1alpha1.Liveness) *corev1.Probe {
	if liveness == nil || liveness.PeriodSeconds <= 0 {
		return nil
	}
	var initialDelay int32
	if liveness.InitialDelaySeconds > initialDelay {
		initialDelay = liveness.InitialDelaySeconds
	}
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/",
				Port: intstr.FromInt(int(MetricsPort.ContainerPort)),
			},
		},
		InitialDelaySeconds: initialDelay,
		TimeoutSeconds:      liveness.PeriodSeconds,
		PeriodSeconds:       liveness.PeriodSeconds,
		SuccessThreshold:    liveness.SuccessThreshold,
		FailureThreshold:    liveness.FailureThreshold,
	}
}

func makeCleanUpJob(objectMeta *metav1.ObjectMeta, container *corev1.Container, volumes []corev1.Volume,
	labels map[string]string, policy v1alpha1.PodPolicy) *v1.Job {
	temp := makePodTemplate(container, nil, volumes, labels, policy, nil)
	temp.Spec.RestartPolicy = corev1.RestartPolicyOnFailure
	var ttlSecond int32
	return &v1.Job{
		ObjectMeta: *objectMeta,
		Spec: v1.JobSpec{
			Template:                *temp,
			TTLSecondsAfterFinished: &ttlSecond,
		},
	}
}

func TriggerCleanup(ctx context.Context, k8sclient client.Client, restClient rest.Interface, config *rest.Config,
	job *v1.Job) error {
	pods := &corev1.PodList{}
	err := k8sclient.List(ctx, pods, client.InNamespace(job.Namespace), client.MatchingLabels{
		"job-name": job.Name,
	})
	if err != nil || len(pods.Items) == 0 {
		return errors.New("error list cleanup pod")
	}
	command := []string{"sh", "-c", "kill -s INT 1"}
	res := restClient.Post().
		Resource("pods").
		Namespace(job.Namespace).
		Name(pods.Items[0].Name).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command:   command,
			Container: CleanupContainerName,
			Stdin:     false,
			Stderr:    true,
			Stdout:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(config, "POST", res.URL())
	if err != nil {
		return err
	}
	err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	if err != nil {
		return err
	}
	return nil
}

func getPulsarAdminCommand(authProvided, tlsProvided bool, tlsConfig TLSConfig,
	authConfig *v1alpha1.AuthConfig) []string {
	args := []string{
		PulsarAdminExecutableFile,
		"--admin-url",
		"$webServiceURL",
	}
	if authConfig != nil {
		if authConfig.OAuth2Config != nil {
			args = append(args, []string{
				"--auth-plugin",
				OAuth2AuthenticationPlugin,
				"--auth-params",
				authConfig.OAuth2Config.AuthenticationParameters(),
			}...)
		} else if authConfig.GenericAuth != nil {
			args = append(args, []string{
				"--auth-plugin",
				authConfig.GenericAuth.ClientAuthenticationPlugin,
				"--auth-params",
				"'" + authConfig.GenericAuth.ClientAuthenticationParameters + "'",
			}...)
		}
	} else if authProvided {
		args = append(args, []string{
			"--auth-plugin",
			"$clientAuthenticationPlugin",
			"--auth-params",
			"$clientAuthenticationParameters"}...)
	}

	// Use traditional way
	if reflect.ValueOf(tlsConfig).IsNil() {
		if tlsProvided {
			args = append(args, []string{
				"--tls-allow-insecure",
				"--tls-enable-hostname-verification",
				"--tls-trust-cert-path",
				"$tlsTrustCertsFilePath",
			}...)
		}
	} else {
		if tlsConfig.IsEnabled() {
			if tlsConfig.AllowInsecureConnection() == "true" {
				args = append(args, "--tls-allow-insecure")
			}

			if tlsConfig.EnableHostnameVerification() == "true" {
				args = append(args, "--tls-enable-hostname-verification")
			}

			if tlsConfig.HasSecretVolume() {
				args = append(args, []string{
					"--tls-trust-cert-path",
					getTLSTrustCertPath(tlsConfig, tlsConfig.SecretKey()),
				}...)
			}
		}
	}

	return args
}

func getPulsarctlCommand(authProvided, tlsProvided bool, tlsConfig TLSConfig,
	authConfig *v1alpha1.AuthConfig) []string {
	args := []string{
		"export PATH=$PATH:/pulsar/bin && ",
		PulsarctlExecutableFile,
		"--admin-service-url",
		"$webServiceURL",
	}
	// activate oauth2 for pulsarctl
	if authConfig != nil {
		if authConfig.OAuth2Config != nil {
			args = []string{
				"export PATH=$PATH:/pulsar/bin && ",
				PulsarctlExecutableFile,
				"context",
				"set",
				"downloader",
				"--admin-service-url",
				"$webServiceURL",
				"--issuer-endpoint",
				authConfig.OAuth2Config.IssuerURL,
				"--audience",
				authConfig.OAuth2Config.Audience,
				"--key-file",
				authConfig.OAuth2Config.GetMountFile(),
			}
			if authConfig.OAuth2Config.Scope != "" {
				args = append(args, []string{
					"--scope",
					authConfig.OAuth2Config.Scope,
				}...)
			}
			args = append(args, []string{
				"&& " + PulsarctlExecutableFile,
				"oauth2",
				"activate",
				"&& " + PulsarctlExecutableFile,
			}...)
		} else if authConfig.GenericAuth != nil {
			args = []string{
				"export PATH=$PATH:/pulsar/bin && ",
				"( " + PulsarctlExecutableFile,
				"oauth2",
				"activate",
				"--auth-params",
				"'" + authConfig.GenericAuth.ClientAuthenticationParameters + "'",
				"|| true ) &&",
				PulsarctlExecutableFile,
				"--auth-plugin",
				authConfig.GenericAuth.ClientAuthenticationPlugin,
				"--auth-params",
				"'" + authConfig.GenericAuth.ClientAuthenticationParameters + "'",
				"--admin-service-url",
				"$webServiceURL",
			}
		}
	} else if authProvided {
		args = []string{
			"export PATH=$PATH:/pulsar/bin && ",
			"( " + PulsarctlExecutableFile,
			"oauth2",
			"activate",
			"--auth-params",
			"$clientAuthenticationParameters || true",
			") &&",
			PulsarctlExecutableFile,
			"--auth-plugin",
			"$clientAuthenticationPlugin",
			"--auth-params",
			"$clientAuthenticationParameters",
			"--admin-service-url",
			"$webServiceURL",
		}
	}

	// Use traditional way
	if reflect.ValueOf(tlsConfig).IsNil() {
		if tlsProvided {
			args = append(args, []string{
				"--tls-allow-insecure=${tlsAllowInsecureConnection:-" + DefaultForAllowInsecure + "}",
				"--tls-enable-hostname-verification=${tlsHostnameVerificationEnable:-" + DefaultForEnableHostNameVerification + "}",
				"--tls-trust-cert-path",
				"$tlsTrustCertsFilePath",
			}...)
		}
	} else {
		if tlsConfig.IsEnabled() {
			args = append(args, []string{
				"--tls-allow-insecure=" + tlsConfig.AllowInsecureConnection(),
				"--tls-enable-hostname-verification=" + tlsConfig.EnableHostnameVerification(),
			}...)

			if tlsConfig.HasSecretVolume() {
				args = append(args, []string{
					"--tls-trust-cert-path",
					getTLSTrustCertPath(tlsConfig, tlsConfig.SecretKey()),
				}...)
			}
		}
	}

	return args
}

func getCleanUpCommand(hasPulsarctl, authProvided, tlsProvided bool, tlsConfig TLSConfig,
	authConfig *v1alpha1.AuthConfig, inputTopics []string,
	topicPattern, subscriptionName, tenant, namespace, name string, deleteTopic bool) []string {
	var adminArgs []string
	if hasPulsarctl {
		adminArgs = getPulsarctlCommand(authProvided, tlsProvided, tlsConfig, authConfig)
	} else {
		adminArgs = getPulsarAdminCommand(authProvided, tlsProvided, tlsConfig, authConfig)
	}

	var cleanupArgs []string
	if topicPattern != "" {
		topicName, _ := pctlutil.GetTopicName(topicPattern)
		if hasPulsarctl {
			cleanupArgs = append(adminArgs, []string{
				"namespaces",
				"unsubscribe",
				topicName.GetTenant() + "/" + topicName.GetNamespace(),
				getSubscriptionNameOrDefault(subscriptionName, tenant, namespace, name),
			}...)
		} else {
			cleanupArgs = append(adminArgs, []string{
				"namespaces",
				"unsubscribe",
				"-s",
				getSubscriptionNameOrDefault(subscriptionName, tenant, namespace, name),
				topicName.GetTenant() + "/" + topicName.GetNamespace(),
			}...)
		}
	} else {
		for idx, topic := range inputTopics {
			var singleCleanupArg []string
			if hasPulsarctl {
				singleCleanupArg = append(adminArgs, []string{
					"subscriptions",
					"delete",
					topic,
					getSubscriptionNameOrDefault(subscriptionName, tenant, namespace, name),
				}...)
			} else {
				singleCleanupArg = append(adminArgs, []string{
					"topics",
					"unsubscribe",
					"-s",
					getSubscriptionNameOrDefault(subscriptionName, tenant, namespace, name),
					topic,
				}...)
			}
			cleanupArgs = append(cleanupArgs, singleCleanupArg...)
			if idx < len(inputTopics)-1 {
				cleanupArgs = append(cleanupArgs, "&& ")
			}
		}
	}

	if deleteTopic {
		deleteTopicArgs := append(adminArgs, []string{
			"topics",
			"delete",
			"-f",
			inputTopics[0],
		}...)
		cleanupArgs = append(cleanupArgs, "&& ")
		cleanupArgs = append(cleanupArgs, deleteTopicArgs...)
	}

	return []string{"sh", "-c",
		"sleep infinity & pid=$!; echo $pid; trap \"kill $pid\" INT; trap 'exit 0' TERM; echo 'waiting...'; wait; sleep 10s; echo 'cleaning...'; " + strings.Join(cleanupArgs,
			" ")}
}

func getDownloadCommand(downloadPath, componentPackage string, hasPulsarctl, hasWget, authProvided, tlsProvided bool,
	tlsConfig TLSConfig,
	authConfig *v1alpha1.AuthConfig) []string {
	var args []string
	if hasHTTPPrefix(downloadPath) && hasWget {
		args = append(args, "wget", downloadPath, "-O", componentPackage)
		return args
	}
	if hasPulsarctl {
		args = getPulsarctlCommand(authProvided, tlsProvided, tlsConfig, authConfig)
	} else {
		args = getPulsarAdminCommand(authProvided, tlsProvided, tlsConfig, authConfig)
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

func generateJavaLogConfigCommand(runtime *v1alpha1.JavaRuntime, agent v1alpha1.LogTopicAgent) string {
	if runtime == nil || (runtime.Log != nil && runtime.Log.LogConfig != nil) {
		return ""
	}
	configFileType := v1alpha1.XML
	if runtime != nil && runtime.Log != nil && runtime.Log.JavaLog4JConfigFileType != nil {
		configFileType = *runtime.Log.JavaLog4JConfigFileType
	}
	switch configFileType {
	case v1alpha1.XML:
		{
			if log4jXML, err := renderJavaInstanceLog4jXMLTemplate(runtime, agent); err == nil {
				generateConfigFileCommand := []string{
					"mkdir", "-p", JavaLogConfigDirectory, "&&",
					"echo", fmt.Sprintf("\"%s\"", log4jXML), ">", DefaultJavaLogConfigPath,
					"&& ",
				}
				return strings.Join(generateConfigFileCommand, " ")
			}
		}
	case v1alpha1.YAML:
		{
			if log4jYAML, err := renderJavaInstanceLog4jYAMLTemplate(runtime, agent); err == nil {
				generateConfigFileCommand := []string{
					"mkdir", "-p", JavaLogConfigDirectory, "&&",
					"echo", fmt.Sprintf("\"%s\"", log4jYAML), ">", DefaultJavaLogConfigPathYAML,
					"&& ",
				}
				return strings.Join(generateConfigFileCommand, " ")
			}
		}
	}
	return ""
}

func generateJavaLogConfigFileName(runtime *v1alpha1.JavaRuntime) string {
	if runtime == nil || (runtime.Log != nil && runtime.Log.LogConfig != nil) {
		return DefaultJavaLogConfigPath
	}
	configFileType := v1alpha1.XML
	if runtime != nil && runtime.Log != nil && runtime.Log.JavaLog4JConfigFileType != nil {
		configFileType = *runtime.Log.JavaLog4JConfigFileType
	}
	switch configFileType {
	case v1alpha1.XML:
		{
			return DefaultJavaLogConfigPath
		}
	case v1alpha1.YAML:
		{
			return DefaultJavaLogConfigPathYAML
		}
	}
	return DefaultJavaLogConfigPath
}

func renderJavaInstanceLog4jYAMLTemplate(runtime *v1alpha1.JavaRuntime, agent v1alpha1.LogTopicAgent) (string, error) {
	tmpl := template.Must(template.New("log4j-yaml-template").Parse(javaLog4jYAMLTemplate))
	var tpl bytes.Buffer
	type logConfig struct {
		RollingEnabled bool
		Level          string
		Format         string
		Policy         string
	}
	lc := &logConfig{}
	lc.Level = "INFO"
	lc.Format = "text"
	if runtime.Log != nil && runtime.Log.Level != "" {
		if level := parseJavaLogLevel(runtime); level != "" {
			lc.Level = level
		}
	}
	if runtime.Log != nil && runtime.Log.RotatePolicy != nil {
		lc.RollingEnabled = true
		lc.Policy = string(*runtime.Log.RotatePolicy)
	} else if agent == v1alpha1.SIDECAR {
		lc.RollingEnabled = true
		lc.Policy = string(v1alpha1.SizedPolicyWith10MB)
	}
	if runtime.Log != nil && runtime.Log.Format != nil {
		lc.Format = string(*runtime.Log.Format)
	}
	if err := tmpl.Execute(&tpl, lc); err != nil {
		log.Error(err, "failed to render java instance log4j yaml template")
		return "", err
	}
	return tpl.String(), nil
}

func renderJavaInstanceLog4jXMLTemplate(runtime *v1alpha1.JavaRuntime, agent v1alpha1.LogTopicAgent) (string, error) {
	tmpl := template.Must(template.New("spec").Parse(javaLog4jXMLTemplate))
	var tpl bytes.Buffer
	type logConfig struct {
		RollingEnabled bool
		Level          string
		Policy         template.HTML
		Format         string
	}
	lc := &logConfig{}
	lc.Level = "INFO"
	lc.Format = "text"
	if runtime.Log != nil && runtime.Log.Level != "" {
		if level := parseJavaLogLevel(runtime); level != "" {
			lc.Level = level
		}
	}
	if runtime.Log != nil && runtime.Log.RotatePolicy != nil {
		lc.RollingEnabled = true
		switch *runtime.Log.RotatePolicy {
		case v1alpha1.TimedPolicyWithDaily:
			lc.Policy = template.HTML(`<CronTriggeringPolicy>
                    <schedule>"0 0 0 \* \* \? \*"</schedule>
                </CronTriggeringPolicy>`)
		case v1alpha1.TimedPolicyWithWeekly:
			lc.Policy = template.HTML(`<CronTriggeringPolicy>
                    <schedule>"0 0 0 \? \* 1 *"</schedule>
                </CronTriggeringPolicy>`)
		case v1alpha1.TimedPolicyWithMonthly:
			lc.Policy = template.HTML(`<CronTriggeringPolicy>
                    <schedule>"0 0 0 1 \* \? \*"</schedule>
                </CronTriggeringPolicy>`)
		case v1alpha1.SizedPolicyWith10MB:
			lc.Policy = template.HTML(`<SizeBasedTriggeringPolicy>
                    <size>10MB</size>
                </SizeBasedTriggeringPolicy>`)
		case v1alpha1.SizedPolicyWith50MB:
			lc.Policy = template.HTML(`<SizeBasedTriggeringPolicy>
                    <size>50MB</size>
                </SizeBasedTriggeringPolicy>`)
		case v1alpha1.SizedPolicyWith100MB:
			lc.Policy = template.HTML(`<SizeBasedTriggeringPolicy>
                    <size>100MB</size>
                </SizeBasedTriggeringPolicy>`)
		}
	} else if agent == v1alpha1.SIDECAR {
		lc.RollingEnabled = true
		lc.Policy = template.HTML(`<SizeBasedTriggeringPolicy>
                    <size>10MB</size>
                </SizeBasedTriggeringPolicy>`)
	}
	if runtime.Log != nil && runtime.Log.Format != nil {
		lc.Format = string(*runtime.Log.Format)
	}
	if err := tmpl.Execute(&tpl, lc); err != nil {
		log.Error(err, "failed to render java instance log4j xml template")
		return "", err
	}
	return tpl.String(), nil
}

func generatePythonLogConfigCommand(name string, runtime *v1alpha1.PythonRuntime, agent v1alpha1.LogTopicAgent) string {
	commands := "sed -i.bak 's/^  Log.setLevel/#&/' /pulsar/instances/python-instance/log.py && "
	if runtime == nil || (runtime.Log != nil && runtime.Log.LogConfig != nil) {
		return commands
	}
	if loggingINI, err := renderPythonInstanceLoggingINITemplate(name, runtime, agent); err == nil {
		generateConfigFileCommand := []string{
			"mkdir", "-p", PythonLogConifgDirectory, "logs/functions", "&&",
			"echo", fmt.Sprintf("\"%s\"", loggingINI), ">", DefaultPythonLogConfigPath,
			"&& ",
		}
		level := parsePythonLogLevel(runtime)
		hackCmd := ""
		if level != "" {
			hackCmd = fmt.Sprintf("sed -i.bak 's/^level=.*/level=%s/g' %s && ", level, DefaultPythonLogConfigPath)
		}
		return commands + strings.Join(generateConfigFileCommand, " ") + hackCmd
	}
	return ""
}

func renderPythonInstanceLoggingINITemplate(name string, runtime *v1alpha1.PythonRuntime, agent v1alpha1.LogTopicAgent) (string, error) {
	tmpl := template.Must(template.New("spec").Parse(pythonLoggingINITemplate))
	var tpl bytes.Buffer
	type logConfig struct {
		RollingEnabled bool
		Level          string
		Policy         template.HTML
		Handlers       string
		Format         string
	}
	lc := &logConfig{}
	lc.Level = "INFO"
	lc.Format = "text"
	lc.Handlers = "stream_handler"
	if runtime.Log != nil && runtime.Log.Level != "" {
		if level := parsePythonLogLevel(runtime); level != "" {
			lc.Level = level
		}
	}
	if runtime.Log != nil && runtime.Log.Format != nil {
		lc.Format = string(*runtime.Log.Format)
	}
	if runtime.Log != nil && runtime.Log.RotatePolicy != nil {
		lc.RollingEnabled = true
		logFile := fmt.Sprintf("logs/functions/%s-${%s}.log", name, EnvShardID)
		switch *runtime.Log.RotatePolicy {
		case v1alpha1.TimedPolicyWithDaily:
			lc.Handlers = "stream_handler,timed_rotating_file_handler"
			lc.Policy = template.HTML(fmt.Sprintf(`[handler_timed_rotating_file_handler]
args=(\"%s\", 'D', 1, 5,)
class=handlers.TimedRotatingFileHandler
level=%s 
formatter=formatter`, logFile, lc.Level))
		case v1alpha1.TimedPolicyWithWeekly:
			lc.Handlers = "stream_handler,timed_rotating_file_handler"
			lc.Policy = template.HTML(fmt.Sprintf(`[handler_timed_rotating_file_handler]
args=(\"%s\", 'W0', 1, 5,)
class=handlers.TimedRotatingFileHandler
level=%s
formatter=formatter`, logFile, lc.Level))
		case v1alpha1.TimedPolicyWithMonthly:
			lc.Handlers = "stream_handler,timed_rotating_file_handler"
			lc.Policy = template.HTML(fmt.Sprintf(`[handler_timed_rotating_file_handler]
args=(\"%s\", 'D', 30, 5,)
class=handlers.TimedRotatingFileHandler
level=%s
formatter=formatter`, logFile, lc.Level))
		case v1alpha1.SizedPolicyWith10MB:
			lc.Handlers = "stream_handler,rotating_file_handler"
			lc.Policy = template.HTML(fmt.Sprintf(`[handler_rotating_file_handler]
args=(\"%s\", 'a', 10485760, 5,)
class=handlers.RotatingFileHandler
level=%s
formatter=formatter`, logFile, lc.Level))
		case v1alpha1.SizedPolicyWith50MB:
			lc.Handlers = "handler_stream_handler,rotating_file_handler"
			lc.Policy = template.HTML(fmt.Sprintf(`[handler_rotating_file_handler]
args=(%s, 'a', 52428800, 5,)
class=handlers.RotatingFileHandler
level=%s
formatter=formatter`, logFile, lc.Level))
		case v1alpha1.SizedPolicyWith100MB:
			lc.Handlers = "handler_stream_handler,rotating_file_handler"
			lc.Policy = template.HTML(fmt.Sprintf(`[handler_rotating_file_handler]
args=(%s, 'a', 104857600, 5,)
class=handlers.RotatingFileHandler
level=%s
formatter=formatter`, logFile, lc.Level))
		}
	} else if agent == v1alpha1.SIDECAR { // sidecar mode needs the rotated log file
		lc.RollingEnabled = true
		logFile := fmt.Sprintf("logs/functions/%s-${%s}.log", name, EnvShardID)
		lc.Handlers = "stream_handler,rotating_file_handler"
		lc.Policy = template.HTML(fmt.Sprintf(`[handler_rotating_file_handler]
args=(\"%s\", 'a', 10485760, 5,)
class=handlers.RotatingFileHandler
level=%s
formatter=formatter`, logFile, lc.Level))
	}
	if err := tmpl.Execute(&tpl, lc); err != nil {
		log.Error(err, "failed to render python instance logging template")
		return "", err
	}
	return tpl.String(), nil
}

func parseJavaLogLevel(runtime *v1alpha1.JavaRuntime) string {
	var levelMap = map[v1alpha1.LogLevel]string{
		v1alpha1.LogLevelAll:   "ALL",
		v1alpha1.LogLevelDebug: "DEBUG",
		v1alpha1.LogLevelTrace: "TRACE",
		v1alpha1.LogLevelInfo:  "INFO",
		v1alpha1.LogLevelWarn:  "WARN",
		v1alpha1.LogLevelError: "ERROR",
		v1alpha1.LogLevelFatal: "FATAL",
		v1alpha1.LogLevelOff:   "OFF",
	}
	if runtime.Log != nil && runtime.Log.Level != "" && runtime.Log.LogConfig == nil {
		if level, exist := levelMap[runtime.Log.Level]; exist {
			return level
		}
		return levelMap[v1alpha1.LogLevelInfo]
	}
	return ""
}

func parsePythonLogLevel(runtime *v1alpha1.PythonRuntime) string {
	var levelMap = map[v1alpha1.LogLevel]string{
		v1alpha1.LogLevelDebug: "DEBUG",
		v1alpha1.LogLevelInfo:  "INFO",
		v1alpha1.LogLevelWarn:  "WARNING",
		v1alpha1.LogLevelError: "ERROR",
		v1alpha1.LogLevelFatal: "CRITICAL",
	}
	if runtime.Log != nil && runtime.Log.Level != "" && runtime.Log.LogConfig == nil {
		if level, exist := levelMap[runtime.Log.Level]; exist {
			return level
		}
		return levelMap[v1alpha1.LogLevelInfo]
	}
	return ""
}

func parseGolangLogLevel(runtime *v1alpha1.GoRuntime) string {
	if runtime == nil || (runtime.Log != nil && runtime.Log.LogConfig != nil) {
		return ""
	}
	var levelMap = map[v1alpha1.LogLevel]string{
		v1alpha1.LogLevelDebug: "debug",
		v1alpha1.LogLevelTrace: "trace",
		v1alpha1.LogLevelInfo:  "info",
		v1alpha1.LogLevelWarn:  "warn",
		v1alpha1.LogLevelError: "error",
		v1alpha1.LogLevelFatal: "fatal",
		v1alpha1.LogLevelPanic: "panic",
	}
	if runtime.Log != nil && runtime.Log.Level != "" {
		if level, exist := levelMap[runtime.Log.Level]; exist {
			return level
		}
	}
	return levelMap[v1alpha1.LogLevelInfo]
}

// TODO: do a more strict check for the package name https://github.com/streamnative/function-mesh/issues/49
func hasPackageNamePrefix(packagesName string) bool {
	return strings.HasPrefix(packagesName, PackageNameFunctionPrefix) ||
		strings.HasPrefix(packagesName, PackageNameSinkPrefix) ||
		strings.HasPrefix(packagesName, PackageNameSourcePrefix)
}

func hasHTTPPrefix(packageName string) bool {
	return strings.HasPrefix(packageName, HTTPPrefix) ||
		strings.HasPrefix(packageName, HTTPSPrefix)
}

func setShardIDEnvironmentVariableCommand() string {
	return fmt.Sprintf("%s=${POD_NAME##*-} && echo shardId=${%s}", EnvShardID, EnvShardID)
}

func getProcessJavaRuntimeArgs(name, packageName, clusterName, logLevel, details, extraDependenciesDir, uid string,
	javaOpts []string, authProvided, tlsProvided bool, secretMaps map[string]v1alpha1.SecretRef,
	state *v1alpha1.Stateful,
	tlsConfig TLSConfig, authConfig *v1alpha1.AuthConfig,
	maxPendingAsyncRequests *int32, logConfigFileName string) []string {
	classPath := "/pulsar/instances/java-instance.jar:/pulsar/lib/*"
	javaLogConfigPath := logConfigFileName
	if javaLogConfigPath == "" {
		javaLogConfigPath = DefaultJavaLogConfigPath
	}

	if extraDependenciesDir != "" {
		classPath = fmt.Sprintf("%s:%s/*", classPath, extraDependenciesDir)
	}
	setLogLevel := ""
	if logLevel != "" {
		setLogLevel = strings.Join(
			[]string{
				fmt.Sprintf("-Dpulsar.log.level=%s", logLevel),
				fmt.Sprintf("-Dbk.log.level=%s", logLevel),
			},
			" ")
	}
	args := []string{
		"exec",
		"java",
		"-cp",
		classPath,
		fmt.Sprintf("-D%s=%s", FunctionsInstanceClasspath, "/pulsar/lib/*"),
		fmt.Sprintf("-Dlog4j.configurationFile=%s", javaLogConfigPath),
		"-Dpulsar.function.log.dir=logs/functions",
		"-Dpulsar.function.log.file=" + fmt.Sprintf("%s-${%s}", name, EnvShardID),
		setLogLevel,
		"-XX:MaxRAMPercentage=40",
		"-XX:+UseG1GC",
		"-XX:+HeapDumpOnOutOfMemoryError",
		"-XX:HeapDumpPath=/pulsar/tmp/heapdump-%p.hprof",
		"-Xlog:gc*:file=/pulsar/logs/gc.log:time,level,tags:filecount=5,filesize=10M",
		strings.Join(javaOpts, " "),
		"org.apache.pulsar.functions.instance.JavaInstanceMain",
		"--jar",
		packageName,
	}
	sharedArgs := getSharedArgs(details, clusterName, uid, authProvided, tlsProvided, tlsConfig, authConfig)
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
	if maxPendingAsyncRequests != nil {
		args = append(args, []string{
			"--pending_async_requests",
			strconv.Itoa(int(*maxPendingAsyncRequests)),
		}...)
	}
	return args
}

func getProcessPythonRuntimeArgs(name, packageName, clusterName, details, uid string, authProvided, tlsProvided bool,
	secretMaps map[string]v1alpha1.SecretRef, state *v1alpha1.Stateful, tlsConfig TLSConfig,
	authConfig *v1alpha1.AuthConfig) []string {
	args := []string{
		"exec",
		"python",
		"/pulsar/instances/python-instance/python_instance_main.py",
		"--py",
		packageName,
		"--logging_directory",
		"logs/functions",
		"--logging_file",
		fmt.Sprintf("%s-${%s}.log", name, EnvShardID),
		"--logging_config_file",
		DefaultPythonLogConfigPath,
		"--install_usercode_dependencies",
		"true",
		// TODO: Maybe we don't need installUserCodeDependencies, dependency_repository, and pythonExtraDependencyRepository
	}
	sharedArgs := getSharedArgs(details, clusterName, uid, authProvided, tlsProvided, tlsConfig, authConfig)
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

func getProcessGenericRuntimeArgs(language, functionFile, clusterName, details, uid string, authProvided, tlsProvided bool,
	secretMaps map[string]v1alpha1.SecretRef, state *v1alpha1.Stateful, tlsConfig TLSConfig,
	authConfig *v1alpha1.AuthConfig) []string {

	args := []string{
		"exec",
		"pulsar_rust_instance",
		"--function_file",
		functionFile,
		"--language",
		language,
	}
	sharedArgs := getSharedArgs(details, clusterName, uid, authProvided, tlsProvided, tlsConfig, authConfig)
	args = append(args, sharedArgs...)
	if len(secretMaps) > 0 {
		secretProviderArgs := getGenericSecretProviderArgs(secretMaps, language)
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
func getSharedArgs(details, clusterName, uid string, authProvided bool, tlsProvided bool,
	tlsConfig TLSConfig, authConfig *v1alpha1.AuthConfig) []string {
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
		"0", // TurnOff BuiltIn HealthCheck to avoid instance exit
		"--cluster_name",
		clusterName,
	}

	if authConfig != nil {
		if authConfig.OAuth2Config != nil {
			args = append(args, []string{
				"--client_auth_plugin",
				OAuth2AuthenticationPlugin,
				"--client_auth_params",
				authConfig.OAuth2Config.AuthenticationParameters()}...)
		} else if authConfig.GenericAuth != nil {
			args = append(args, []string{
				"--client_auth_plugin",
				authConfig.GenericAuth.ClientAuthenticationPlugin,
				"--client_auth_params",
				"'" + authConfig.GenericAuth.ClientAuthenticationParameters + "'"}...)
		}
	} else if authProvided {
		args = append(args, []string{
			"--client_auth_plugin",
			"$clientAuthenticationPlugin",
			"--client_auth_params",
			"$clientAuthenticationParameters"}...)
	}

	// Use traditional way
	if reflect.ValueOf(tlsConfig).IsNil() {
		if tlsProvided {
			args = append(args, []string{
				"--use_tls",
				"true",
				"--tls_allow_insecure",
				"${tlsAllowInsecureConnection:-" + DefaultForAllowInsecure + "}",
				"--hostname_verification_enabled",
				"${tlsHostnameVerificationEnable:-" + DefaultForEnableHostNameVerification + "}",
				"--tls_trust_cert_path",
				"$tlsTrustCertsFilePath",
			}...)
		} else {
			args = append(args, []string{
				"--use_tls",
				"false",
			}...)
		}
	} else {
		args = append(args, []string{
			"--use_tls",
			strconv.FormatBool(tlsConfig.IsEnabled()),
		}...)

		if tlsConfig.IsEnabled() {
			args = append(args, []string{
				"--tls_allow_insecure",
				tlsConfig.AllowInsecureConnection(),
				"--hostname_verification_enabled",
				tlsConfig.EnableHostnameVerification(),
			}...)

			if tlsConfig.HasSecretVolume() {
				args = append(args, []string{
					"--tls_trust_cert_path",
					getTLSTrustCertPath(tlsConfig, tlsConfig.SecretKey()),
				}...)
			}
		}
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
	case v1alpha1.Manual:
		return proto.ProcessingGuarantees_MANUAL
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

func convertCompressionType(compressionType v1alpha1.CompressionType) proto.CompressionType {
	switch compressionType {
	case v1alpha1.LZ4:
		return proto.CompressionType_LZ4
	case v1alpha1.ZLIB:
		return proto.CompressionType_ZLIB
	case v1alpha1.ZSTD:
		return proto.CompressionType_ZSTD
	case v1alpha1.NONE:
		return proto.CompressionType_NONE
	case v1alpha1.SNAPPY:
		return proto.CompressionType_SNAPPY
	default:
		return proto.CompressionType_LZ4
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
		Cpu:  resources.Cpu().AsApproximateFloat64(),
		Ram:  int64(resources.Memory().AsApproximateFloat64()),
		Disk: int64(resources.Storage().AsApproximateFloat64()),
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

func generateContainerEnv(function *v1alpha1.Function) []corev1.EnvVar {
	envs := generateBasicContainerEnv(function.Spec.SecretsMap, function.Spec.Pod.Env)

	// add env to set logging level for Go runtime
	if level := parseGolangLogLevel(function.Spec.Golang); level != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  EnvGoFunctionLogLevel,
			Value: level,
		})
	}
	return envs
}

func generateBasicContainerEnv(secrets map[string]v1alpha1.SecretRef, env []corev1.EnvVar) []corev1.EnvVar {
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

	vars = append(vars, env...)

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

func generateContainerVolumesFromLogConfigs(confs map[int32]*v1alpha1.RuntimeLogConfig) []corev1.Volume {
	volumes := []corev1.Volume{}
	if len(confs) > 0 {
		if conf, exist := confs[javaRuntimeLog]; exist {
			filePath := JavaLogConfigFileXML
			if conf.JavaLog4JConfigFileType != nil {
				switch *conf.JavaLog4JConfigFileType {
				case v1alpha1.XML:
					filePath = JavaLogConfigFileXML
				case v1alpha1.YAML:
					filePath = JavaLogConfigFileYAML
				default:
					filePath = JavaLogConfigFileXML
				}
			}
			javaLogConfigVolume := &corev1.Volume{
				Name: generateVolumeNameFromLogConfigs(conf.LogConfig.Name, "java"),
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: conf.LogConfig.Name,
						},
						Items: []corev1.KeyToPath{
							{
								Key:  conf.LogConfig.Key,
								Path: filePath,
							},
						},
					},
				},
			}
			volumes = append(volumes, *javaLogConfigVolume)
		}
		if conf, exist := confs[pythonRuntimeLog]; exist {
			pythonLogConfigVolume := &corev1.Volume{
				Name: generateVolumeNameFromLogConfigs(conf.LogConfig.Name, "python"),
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: conf.LogConfig.Name,
						},
						Items: []corev1.KeyToPath{
							{
								Key:  conf.LogConfig.Key,
								Path: PythonLogConfigFile,
							},
						},
					},
				},
			}
			volumes = append(volumes, *pythonLogConfigVolume)
		}
	}
	return volumes
}

func generateFilebeatVolumes() []corev1.Volume {
	return []corev1.Volume{{
		Name: "filebeat-logs",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}}
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

func generateVolumeFromTLSConfig(tlsConfig TLSConfig) corev1.Volume {
	return corev1.Volume{
		Name: generateVolumeNameFromTLSConfig(tlsConfig),
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: tlsConfig.SecretName(),
				Items: []corev1.KeyToPath{
					{
						Key:  tlsConfig.SecretKey(),
						Path: tlsConfig.SecretKey(),
					},
				},
			},
		},
	}
}

func generateVolumeFromOAuth2Config(config *v1alpha1.OAuth2Config) corev1.Volume {
	return corev1.Volume{
		Name: generateVolumeNameFromOAuth2Config(config),
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: config.KeySecretName,
				Items: []corev1.KeyToPath{
					{
						Key:  config.KeySecretKey,
						Path: config.KeySecretKey,
					},
				},
			},
		},
	}
}

func generateVolumeMountFromLogConfigs(confs map[int32]*v1alpha1.RuntimeLogConfig) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{}
	if len(confs) > 0 {
		if conf, exist := confs[javaRuntimeLog]; exist {
			javaLogConfigVolumeMount := &corev1.VolumeMount{
				Name:      generateVolumeNameFromLogConfigs(conf.LogConfig.Name, "java"),
				MountPath: JavaLogConfigDirectory,
			}
			volumeMounts = append(volumeMounts, *javaLogConfigVolumeMount)
		}
		if conf, exist := confs[pythonRuntimeLog]; exist {
			pythonLogConfigVolumeMount := &corev1.VolumeMount{
				Name:      generateVolumeNameFromLogConfigs(conf.LogConfig.Name, "python"),
				MountPath: PythonLogConifgDirectory,
			}
			volumeMounts = append(volumeMounts, *pythonLogConfigVolumeMount)
		}
	}
	return volumeMounts
}

func generateVolumeMountFromCryptoSecret(secret *v1alpha1.CryptoSecret) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      generateVolumeNameFromCryptoSecrets(secret),
		MountPath: secret.AsVolume,
	}
}

func generateVolumeMountFromTLSConfig(tlsConfig TLSConfig) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      generateVolumeNameFromTLSConfig(tlsConfig),
		MountPath: tlsConfig.GetMountPath(),
	}
}

func generateVolumeMountFromOAuth2Config(config *v1alpha1.OAuth2Config) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      generateVolumeNameFromOAuth2Config(config),
		MountPath: config.GetMountPath(),
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

func generateDownloaderVolumeMountsForDownloader(javaRuntime *v1alpha1.JavaRuntime,
	pythonRuntime *v1alpha1.PythonRuntime, goRuntime *v1alpha1.GoRuntime) []corev1.VolumeMount {
	if !utils.EnableInitContainers {
		return nil
	}
	if (javaRuntime != nil && javaRuntime.JarLocation != "") ||
		(pythonRuntime != nil && pythonRuntime.PyLocation != "") ||
		(goRuntime != nil && goRuntime.GoLocation != "") {
		return []corev1.VolumeMount{{
			Name:      DownloaderVolume,
			MountPath: DownloadDir,
		}}
	}
	return nil
}

func generateDownloaderVolumeMountsForRuntime(javaRuntime *v1alpha1.JavaRuntime, pythonRuntime *v1alpha1.PythonRuntime,
	goRuntime *v1alpha1.GoRuntime) []corev1.VolumeMount {
	downloadPath := ""
	if javaRuntime != nil && javaRuntime.JarLocation != "" {
		downloadPath = javaRuntime.Jar
	} else if pythonRuntime != nil && pythonRuntime.PyLocation != "" {
		downloadPath = pythonRuntime.Py
	} else if goRuntime != nil && goRuntime.GoLocation != "" {
		downloadPath = goRuntime.Go
	}

	if downloadPath != "" {
		subPath := getFilenameOfComponentPackage(downloadPath)
		mountPath := downloadPath
		// for relative path, volume should be mounted to the WorkDir
		// and path also should be under the $WorkDir dir
		if !strings.HasPrefix(downloadPath, "/") {
			mountPath = WorkDir + downloadPath
		} else if !strings.HasPrefix(downloadPath, WorkDir) {
			mountPath = strings.Replace(downloadPath, "/", WorkDir, 1)
		}
		// if mount path is in the $WorkDir instead of its sub path
		// we should use SubPath to avoid the mounted volume overwrite the $WorkDir
		if !strings.Contains(strings.Replace(mountPath, WorkDir, "", 1), "/") {
			return []corev1.VolumeMount{{
				Name:      DownloaderVolume,
				MountPath: mountPath,
				SubPath:   subPath,
			}}
		}
		return []corev1.VolumeMount{{
			Name:      DownloaderVolume,
			MountPath: mountPath[:len(mountPath)-len(subPath)],
		}}
	}
	return nil
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
	consumerConfs map[string]v1alpha1.ConsumerConfig, tlsConfig TLSConfig, authConfig *v1alpha1.AuthConfig,
	logConfigs map[int32]*v1alpha1.RuntimeLogConfig, agent v1alpha1.LogTopicAgent) []corev1.VolumeMount {
	mounts := []corev1.VolumeMount{}
	mounts = append(mounts, volumeMounts...)
	if agent == v1alpha1.SIDECAR {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      "filebeat-logs",
			MountPath: "/pulsar/logs/functions",
		})
	}
	if !reflect.ValueOf(tlsConfig).IsNil() && tlsConfig.HasSecretVolume() {
		mounts = append(mounts, generateVolumeMountFromTLSConfig(tlsConfig))
	}
	if authConfig != nil {
		if authConfig.OAuth2Config != nil {
			mounts = append(mounts, generateVolumeMountFromOAuth2Config(authConfig.OAuth2Config))
		}
	}
	mounts = append(mounts, generateContainerVolumeMountsFromProducerConf(producerConf)...)
	mounts = append(mounts, generateContainerVolumeMountsFromConsumerConfigs(consumerConfs)...)
	mounts = append(mounts, generateVolumeMountFromLogConfigs(logConfigs)...)
	return mounts
}

func generatePodVolumes(podVolumes []corev1.Volume, producerConf *v1alpha1.ProducerConfig,
	consumerConfs map[string]v1alpha1.ConsumerConfig, tlsConfig TLSConfig, authConfig *v1alpha1.AuthConfig,
	logConfigs map[int32]*v1alpha1.RuntimeLogConfig,
	agent v1alpha1.LogTopicAgent) []corev1.Volume {
	volumes := []corev1.Volume{}
	volumes = append(volumes, podVolumes...)
	if agent == v1alpha1.SIDECAR {
		volumes = append(volumes, generateFilebeatVolumes()...)
	}
	if !reflect.ValueOf(tlsConfig).IsNil() && tlsConfig.HasSecretVolume() {
		volumes = append(volumes, generateVolumeFromTLSConfig(tlsConfig))
	}
	if authConfig != nil {
		if authConfig.OAuth2Config != nil {
			volumes = append(volumes, generateVolumeFromOAuth2Config(authConfig.OAuth2Config))
		}
	}
	volumes = append(volumes, generateContainerVolumesFromProducerConf(producerConf)...)
	volumes = append(volumes, generateContainerVolumesFromConsumerConfigs(consumerConfs)...)
	volumes = append(volumes, generateContainerVolumesFromLogConfigs(logConfigs)...)
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
	} else if runtime.GenericRuntime != nil && runtime.GenericRuntime.Language != "" {
		return Configs.RunnerImages.GenericRuntime[runtime.GenericRuntime.Language]
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
func getDefaultRunnerPodSecurityContext(uid, gid int64, runAsNonRootEnv string) *corev1.PodSecurityContext {
	nonRoot, err := strconv.ParseBool(runAsNonRootEnv)
	if err != nil {
		nonRoot = true
	}
	return &corev1.PodSecurityContext{
		RunAsUser:    &uid,
		RunAsGroup:   &gid,
		RunAsNonRoot: &nonRoot,
		FSGroup:      &gid,
		SeccompProfile: &corev1.SeccompProfile{
			Type: corev1.SeccompProfileTypeRuntimeDefault,
		},
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

func getGenericSecretProviderArgs(secretMaps map[string]v1alpha1.SecretRef, language string) []string {
	var ret []string
	if len(secretMaps) > 0 {
		if language == "python" {
			ret = []string{
				"--secrets_provider",
				"secrets_provider.EnvironmentBasedSecretsProvider",
			}
		} else if language == "nodejs" {
			ret = []string{
				"--secrets_provider",
				"EnvironmentBasedSecretsProvider",
			}
		}
	}
	return ret
}

func getTLSTrustCertPath(tlsVolume TLSConfig, path string) string {
	return fmt.Sprintf("%s/%s", tlsVolume.GetMountPath(), path)
}

const (
	javaRuntimeLog = iota
	pythonRuntimeLog
	golangRuntimeLog
)

func getRuntimeLogConfigNames(java *v1alpha1.JavaRuntime, python *v1alpha1.PythonRuntime,
	golang *v1alpha1.GoRuntime) map[int32]*v1alpha1.RuntimeLogConfig {

	var configs = map[int32]*v1alpha1.RuntimeLogConfig{}

	if java != nil && java.Log != nil && java.Log.LogConfig != nil {
		configs[javaRuntimeLog] = java.Log
	}
	if python != nil && python.Log != nil && python.Log.LogConfig != nil {
		configs[pythonRuntimeLog] = python.Log
	}
	if golang != nil && golang.Log != nil && golang.Log.LogConfig != nil {
		configs[golangRuntimeLog] = golang.Log
	}
	return configs
}

func CheckIfStatefulSetSpecIsEqual(spec *appsv1.StatefulSetSpec, desiredSpec *appsv1.StatefulSetSpec) bool {
	if *spec.Replicas != *desiredSpec.Replicas {
		return false
	}

	if len(spec.Template.Spec.Containers) != len(desiredSpec.Template.Spec.Containers) {
		return false
	}

	for _, container := range spec.Template.Spec.Containers {
		containerMatch := false
		for _, desiredContainer := range desiredSpec.Template.Spec.Containers {
			if container.Name == desiredContainer.Name {
				containerMatch = true
				// sort container ports
				ports := container.Ports
				desiredPorts := desiredContainer.Ports
				sort.Slice(ports, func(i, j int) bool {
					return ports[i].Name < ports[j].Name
				})
				sort.Slice(desiredPorts, func(i, j int) bool {
					return desiredPorts[i].Name < desiredPorts[j].Name
				})
				// sort container envFrom
				containerEnvFrom := container.EnvFrom
				desiredContainerEnvFrom := desiredContainer.EnvFrom
				sort.Slice(containerEnvFrom, func(i, j int) bool {
					return containerEnvFrom[i].Prefix < containerEnvFrom[j].Prefix
				})
				sort.Slice(desiredContainerEnvFrom, func(i, j int) bool {
					return desiredContainerEnvFrom[i].Prefix < desiredContainerEnvFrom[j].Prefix
				})

				if desiredContainer.ImagePullPolicy == "" {
					desiredContainer.ImagePullPolicy = corev1.PullIfNotPresent
				}

				if !reflect.DeepEqual(container.Command, desiredContainer.Command) ||
					container.Image != desiredContainer.Image ||
					container.ImagePullPolicy != desiredContainer.ImagePullPolicy ||
					!reflect.DeepEqual(ports, desiredPorts) ||
					!reflect.DeepEqual(containerEnvFrom, desiredContainerEnvFrom) ||
					!reflect.DeepEqual(container.Resources, desiredContainer.Resources) {
					return false
				}

				if len(container.Env) != len(desiredContainer.Env) {
					return false
				}
			}
		}
		if !containerMatch {
			return false
		}
	}
	return true
}

func CheckIfHPASpecIsEqual(spec *autov2.HorizontalPodAutoscalerSpec,
	desiredSpec *autov2.HorizontalPodAutoscalerSpec) bool {
	if spec.MaxReplicas != desiredSpec.MaxReplicas || *spec.MinReplicas != *desiredSpec.MinReplicas {
		return false
	}
	if !reflect.DeepEqual(spec.Metrics, desiredSpec.Metrics) {
		return false
	}
	if objectXOROperator(spec.Behavior, desiredSpec.Behavior) {
		return false
	}
	if desiredSpec.Behavior != nil {
		if objectXOROperator(spec.Behavior.ScaleUp, desiredSpec.Behavior.ScaleUp) ||
			objectXOROperator(spec.Behavior.ScaleDown, desiredSpec.Behavior.ScaleDown) {
			return false
		}
		if desiredSpec.Behavior.ScaleUp != nil {
			if objectXOROperator(spec.Behavior.ScaleUp.StabilizationWindowSeconds,
				desiredSpec.Behavior.ScaleUp.StabilizationWindowSeconds) ||
				objectXOROperator(spec.Behavior.ScaleUp.SelectPolicy, desiredSpec.Behavior.ScaleUp.SelectPolicy) ||
				objectXOROperator(spec.Behavior.ScaleUp.Policies, desiredSpec.Behavior.ScaleUp.Policies) {
				return false
			}
			if desiredSpec.Behavior.ScaleUp.StabilizationWindowSeconds != nil && *desiredSpec.Behavior.ScaleUp.StabilizationWindowSeconds != *spec.Behavior.ScaleUp.StabilizationWindowSeconds {
				return false
			}
			if desiredSpec.Behavior.ScaleUp.SelectPolicy != nil && *desiredSpec.Behavior.ScaleUp.SelectPolicy != *spec.Behavior.ScaleUp.SelectPolicy {
				return false
			}
			// sort policies
			desiredPolicies := desiredSpec.Behavior.ScaleUp.Policies
			specPolicies := spec.Behavior.ScaleUp.Policies
			sort.Slice(desiredPolicies, func(i, j int) bool {
				return desiredPolicies[i].Type < desiredPolicies[j].Type
			})
			sort.Slice(specPolicies, func(i, j int) bool {
				return specPolicies[i].Type < specPolicies[j].Type
			})
			if !reflect.DeepEqual(desiredPolicies, specPolicies) {
				return false
			}
		}
		if desiredSpec.Behavior.ScaleDown != nil {
			if objectXOROperator(spec.Behavior.ScaleDown.StabilizationWindowSeconds,
				desiredSpec.Behavior.ScaleDown.StabilizationWindowSeconds) ||
				objectXOROperator(spec.Behavior.ScaleDown.SelectPolicy, desiredSpec.Behavior.ScaleDown.SelectPolicy) ||
				objectXOROperator(spec.Behavior.ScaleDown.Policies, desiredSpec.Behavior.ScaleDown.Policies) {
				return false
			}
			if desiredSpec.Behavior.ScaleDown.StabilizationWindowSeconds != nil && *desiredSpec.Behavior.ScaleDown.StabilizationWindowSeconds != *spec.Behavior.ScaleDown.StabilizationWindowSeconds {
				return false
			}
			if desiredSpec.Behavior.ScaleDown.SelectPolicy != nil && *desiredSpec.Behavior.ScaleDown.SelectPolicy != *spec.Behavior.ScaleDown.SelectPolicy {
				return false
			}
			// sort policies
			desiredPolicies := desiredSpec.Behavior.ScaleDown.Policies
			specPolicies := spec.Behavior.ScaleDown.Policies
			sort.Slice(desiredPolicies, func(i, j int) bool {
				return desiredPolicies[i].Type < desiredPolicies[j].Type
			})
			sort.Slice(specPolicies, func(i, j int) bool {
				return specPolicies[i].Type < specPolicies[j].Type
			})
			if !reflect.DeepEqual(desiredPolicies, specPolicies) {
				return false
			}
		}
	}
	return true
}

func CheckIfHPAV2Beta2SpecIsEqual(spec *autoscalingv2beta2.HorizontalPodAutoscalerSpec,
	desiredSpec *autoscalingv2beta2.HorizontalPodAutoscalerSpec) bool {
	if spec.MaxReplicas != desiredSpec.MaxReplicas || *spec.MinReplicas != *desiredSpec.MinReplicas {
		return false
	}
	if !reflect.DeepEqual(spec.Metrics, desiredSpec.Metrics) {
		return false
	}
	if objectXOROperator(spec.Behavior, desiredSpec.Behavior) {
		return false
	}
	if desiredSpec.Behavior != nil {
		if objectXOROperator(spec.Behavior.ScaleUp, desiredSpec.Behavior.ScaleUp) ||
			objectXOROperator(spec.Behavior.ScaleDown, desiredSpec.Behavior.ScaleDown) {
			return false
		}
		if desiredSpec.Behavior.ScaleUp != nil {
			if objectXOROperator(spec.Behavior.ScaleUp.StabilizationWindowSeconds,
				desiredSpec.Behavior.ScaleUp.StabilizationWindowSeconds) ||
				objectXOROperator(spec.Behavior.ScaleUp.SelectPolicy, desiredSpec.Behavior.ScaleUp.SelectPolicy) ||
				objectXOROperator(spec.Behavior.ScaleUp.Policies, desiredSpec.Behavior.ScaleUp.Policies) {
				return false
			}
			if desiredSpec.Behavior.ScaleUp.StabilizationWindowSeconds != nil && *desiredSpec.Behavior.ScaleUp.StabilizationWindowSeconds != *spec.Behavior.ScaleUp.StabilizationWindowSeconds {
				return false
			}
			if desiredSpec.Behavior.ScaleUp.SelectPolicy != nil && *desiredSpec.Behavior.ScaleUp.SelectPolicy != *spec.Behavior.ScaleUp.SelectPolicy {
				return false
			}
			// sort policies
			desiredPolicies := desiredSpec.Behavior.ScaleUp.Policies
			specPolicies := spec.Behavior.ScaleUp.Policies
			sort.Slice(desiredPolicies, func(i, j int) bool {
				return desiredPolicies[i].Type < desiredPolicies[j].Type
			})
			sort.Slice(specPolicies, func(i, j int) bool {
				return specPolicies[i].Type < specPolicies[j].Type
			})
			if !reflect.DeepEqual(desiredPolicies, specPolicies) {
				return false
			}
		}
		if desiredSpec.Behavior.ScaleDown != nil {
			if objectXOROperator(spec.Behavior.ScaleDown.StabilizationWindowSeconds,
				desiredSpec.Behavior.ScaleDown.StabilizationWindowSeconds) ||
				objectXOROperator(spec.Behavior.ScaleDown.SelectPolicy, desiredSpec.Behavior.ScaleDown.SelectPolicy) ||
				objectXOROperator(spec.Behavior.ScaleDown.Policies, desiredSpec.Behavior.ScaleDown.Policies) {
				return false
			}
			if desiredSpec.Behavior.ScaleDown.StabilizationWindowSeconds != nil && *desiredSpec.Behavior.ScaleDown.StabilizationWindowSeconds != *spec.Behavior.ScaleDown.StabilizationWindowSeconds {
				return false
			}
			if desiredSpec.Behavior.ScaleDown.SelectPolicy != nil && *desiredSpec.Behavior.ScaleDown.SelectPolicy != *spec.Behavior.ScaleDown.SelectPolicy {
				return false
			}
			// sort policies
			desiredPolicies := desiredSpec.Behavior.ScaleDown.Policies
			specPolicies := spec.Behavior.ScaleDown.Policies
			sort.Slice(desiredPolicies, func(i, j int) bool {
				return desiredPolicies[i].Type < desiredPolicies[j].Type
			})
			sort.Slice(specPolicies, func(i, j int) bool {
				return specPolicies[i].Type < specPolicies[j].Type
			})
			if !reflect.DeepEqual(desiredPolicies, specPolicies) {
				return false
			}
		}
	}
	return true
}

func objectXOROperator(first interface{}, second interface{}) bool {
	return (first != nil && second == nil) || (first == nil && second != nil)
}

func getFilenameOfComponentPackage(componentPackage string) string {
	data := strings.Split(componentPackage, "/")
	if len(data) > 0 {
		return data[len(data)-1]
	}
	return componentPackage
}

func getSubscriptionNameOrDefault(subscription, tenant, namespace, name string) string {
	if subscription == "" {
		return fmt.Sprintf("%s/%s/%s", tenant, namespace, name)
	}
	return subscription
}

func makeFilebeatContainer(volumeMounts []corev1.VolumeMount, envVar []corev1.EnvVar, name string, logTopic string,
	agent v1alpha1.LogTopicAgent, tlsConfig TLSConfig, authConfig *v1alpha1.AuthConfig,
	pulsarConfig string, authSecret string, tlsSecret string, image string) *corev1.Container {
	if logTopic == "" || agent != v1alpha1.SIDECAR {
		return nil
	}
	filebeatImage := image
	if filebeatImage == "" {
		filebeatImage = DefaultFilebeatImage
	}
	imagePullPolicy := corev1.PullIfNotPresent
	allowPrivilegeEscalation := false
	mounts := generateContainerVolumeMounts(volumeMounts, nil, nil, tlsConfig, authConfig, nil, agent)

	var uid int64 = 1000

	envs := append(envVar, corev1.EnvVar{
		Name:  "logTopic",
		Value: logTopic,
	}, corev1.EnvVar{
		Name:  "logName",
		Value: name,
	})

	if authConfig != nil {
		if authConfig.GenericAuth != nil {
			// it only supports jwt token authentication and oauth2 authentication
			if authConfig.GenericAuth.ClientAuthenticationPlugin == TokenAuthenticationPlugin {
				envs = append(envs, corev1.EnvVar{
					Name:  "clientAuthenticationParameters",
					Value: authConfig.GenericAuth.ClientAuthenticationParameters,
				})
			} else if authConfig.GenericAuth.ClientAuthenticationPlugin == OAuth2AuthenticationPlugin {
				// we should unmarshal auth params to an OAuth2 object to get issuerUrl, audience, privateKey and scope
				var oauth2Params map[string]string
				err := json.Unmarshal([]byte(authConfig.GenericAuth.ClientAuthenticationParameters), &oauth2Params)
				if err != nil {
					log.Error(err, "failed to unmarshal auth params to an OAuth2 object")
					return nil
				}
				envs = append(envs, corev1.EnvVar{
					Name:  "oauth2Enabled",
					Value: "true",
				}, corev1.EnvVar{
					Name:  "oauth2IssuerUrl",
					Value: oauth2Params["issuerUrl"],
				}, corev1.EnvVar{
					Name:  "oauth2Audience",
					Value: oauth2Params["audience"],
				}, corev1.EnvVar{
					Name:  "oauth2PrivateKey",
					Value: oauth2Params["privateKey"],
				})
				if oauth2Params["scope"] != "" {
					envs = append(envs, corev1.EnvVar{
						Name:  "oauth2Scope",
						Value: oauth2Params["scope"],
					})
				}
			}
		} else if authConfig.OAuth2Config != nil {
			envs = append(envs, corev1.EnvVar{
				Name:  "oauth2Enabled",
				Value: "true",
			}, corev1.EnvVar{
				Name:  "oauth2IssuerUrl",
				Value: authConfig.OAuth2Config.IssuerURL,
			}, corev1.EnvVar{
				Name:  "oauth2Audience",
				Value: authConfig.OAuth2Config.Audience,
			}, corev1.EnvVar{
				Name:  "oauth2PrivateKey",
				Value: authConfig.OAuth2Config.GetMountFile(),
			})
			if authConfig.OAuth2Config.Scope != "" {
				envs = append(envs, corev1.EnvVar{
					Name:  "oauth2Scope",
					Value: authConfig.OAuth2Config.Scope,
				})
			}
		}
	}

	tmpl := template.Must(template.New("filebeat-yaml-template").Parse(filebeatYAMLTemplate))
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, nil); err != nil {
		log.Error(err, "failed to render filebeat instance log4j yaml template")
	}

	return &corev1.Container{
		Name:            "filebeat",
		Image:           filebeatImage,
		Command:         []string{"/bin/sh", "-c", "--", "echo " + fmt.Sprintf("\"%s\"", tpl.String()) + " > " + DefaultFilebeatConfig + " && /usr/share/filebeat/filebeat -e -c " + DefaultFilebeatConfig},
		Env:             envs,
		ImagePullPolicy: imagePullPolicy,
		EnvFrom:         generateContainerEnvFrom(pulsarConfig, authSecret, tlsSecret),
		VolumeMounts:    mounts,
		SecurityContext: &corev1.SecurityContext{
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"ALL"},
			},
			AllowPrivilegeEscalation: &allowPrivilegeEscalation,
			RunAsUser:                &uid,
		},
	}
}
