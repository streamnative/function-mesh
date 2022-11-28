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
	"encoding/json"
	"fmt"
	"html/template"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/proto"
	"github.com/streamnative/function-mesh/utils"
	appsv1 "k8s.io/api/apps/v1"
	autov2beta2 "k8s.io/api/autoscaling/v2beta2"
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
	WorkDir                    = "/pulsar/"

	// for init container
	PulsarctlExecutableFile = "/usr/local/bin/pulsarctl"
	DownloaderName          = "downloader"
	DownloaderVolume        = "downloader-volume"
	DownloaderImage         = DefaultRunnerPrefix + "pulsarctl:2.10.2.1"
	DownloadDir             = "/pulsar/download"

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

	AnnotationPrometheusScrape = "prometheus.io/scrape"
	AnnotationPrometheusPort   = "prometheus.io/port"
	AnnotationManaged          = "compute.functionmesh.io/managed"

	EnvGoFunctionConfigs = "GO_FUNCTION_CONF"

	DefaultRunnerUserID  int64 = 10000
	DefaultRunnerGroupID int64 = 10001

	OAuth2AuthenticationPlugin = "org.apache.pulsar.client.impl.auth.oauth2.AuthenticationOAuth2"

	JavaLogConfigDirectory     = "/pulsar/conf/java-log/"
	JavaLogConfigFile          = "java_instance_log4j.xml"
	DefaultJavaLogConfigPath   = JavaLogConfigDirectory + JavaLogConfigFile
	PythonLogConifgDirectory   = "/pulsar/conf/python-log/"
	PythonLogConfigFile        = "python_instance_logging.ini"
	DefaultPythonLogConfigPath = PythonLogConifgDirectory + PythonLogConfigFile

	EnvGoFunctionLogLevel = "LOGGING_LEVEL"

	javaLog4jXMLTemplate = `<Configuration>
    <name>pulsar-functions-kubernetes-instance</name>
    <monitorInterval>30</monitorInterval>
    <Properties>
        <Property>
            <name>pulsar.log.level</name>
            <value>{{ .Level }}</value>
        </Property>
        <Property>
            <name>bk.log.level</name>
            <value>{{ .Level }}</value>
        </Property>
    </Properties>
    <Appenders>
        <Console>
            <name>Console</name>
            <target>SYSTEM_OUT</target>
            <PatternLayout>
                <Pattern>%d{ISO8601_OFFSET_DATE_TIME_HHMM} [%t] %-5level %logger{36} - %msg%n</Pattern>
            </PatternLayout>
        </Console>
        {{- if .RollingEnabled }}
        <RollingRandomAccessFile>
            <name>RollingRandomAccessFile</name>
            <fileName>\${sys:pulsar.function.log.dir}/\${sys:pulsar.function.log.file}.log</fileName>
            <filePattern>\${sys:pulsar.function.log.dir}/\${sys:pulsar.function.log.file}.%d{yyyy-MM-dd-hh-mm}-%i.log.gz</filePattern>
            <PatternLayout>
                <Pattern>%d{yyyy-MMM-dd HH:mm:ss a} [%t] %-5level %logger{36} - %msg%n</Pattern>
            </PatternLayout>
            <Policies>
                {{ .Policy }}
            </Policies>
            <DefaultRolloverStrategy>
                <max>5</max>
            </DefaultRolloverStrategy>
        </RollingRandomAccessFile>
       {{- end }}
    </Appenders>
    <Loggers>
        <Logger>
            <name>org.apache.pulsar.functions.runtime.shaded.org.apache.bookkeeper</name>
            <level>\${sys:bk.log.level}</level>
            <additivity>false</additivity>
            <AppenderRef>
                <ref>Console</ref>
            </AppenderRef>
            {{- if .RollingEnabled }}
            <AppenderRef>
                <ref>RollingRandomAccessFile</ref>
            </AppenderRef>
            {{- end }}
        </Logger>
        <Root>
            <level>\${sys:pulsar.log.level}</level>
            <AppenderRef>
                <ref>Console</ref>
                <level>\${sys:pulsar.log.level}</level>
            </AppenderRef>
            {{- if .RollingEnabled }}
            <AppenderRef>
                <ref>RollingRandomAccessFile</ref>
            </AppenderRef>
            {{- end }}
        </Root>
    </Loggers>
</Configuration>`
	pythonLoggingINITemplate = `[loggers]
keys=root

[handlers]
keys={{ .Handlers }}

[formatters]
keys=formatter

[logger_root]
level={{ .Level }}
handlers={{ .Handlers }}

{{- if .RollingEnabled }}
{{ .Policy }}
{{- end }}

[handler_stream_handler]
class=StreamHandler
level={{ .Level }}
formatter=formatter
args=(sys.stdout,)

[formatter_formatter]
format=[%(asctime)s] [%(levelname)s] aa %(filename)s: %(message)s
datefmt=%Y-%m-%d %H:%M:%S %z`
)

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
	container *corev1.Container,
	volumes []corev1.Volume, labels map[string]string, policy v1alpha1.PodPolicy, pulsar v1alpha1.PulsarMessaging,
	javaRuntime *v1alpha1.JavaRuntime, pythonRuntime *v1alpha1.PythonRuntime,
	goRuntime *v1alpha1.GoRuntime, definedVolumeMounts []corev1.VolumeMount) *appsv1.StatefulSet {

	volumeMounts := generateDownloaderVolumeMountsForDownloader(javaRuntime, pythonRuntime, goRuntime)
	var downloaderContainer *corev1.Container
	var podVolumes = volumes
	// there must be a download path specified, we need to create an init container and emptyDir volume
	if len(volumeMounts) > 0 {
		if pulsar.AuthConfig != nil && pulsar.AuthConfig.OAuth2Config != nil {
			volumeMounts = append(volumeMounts, generateVolumeMountFromOAuth2Config(pulsar.AuthConfig.OAuth2Config))
		}

		if !reflect.ValueOf(pulsar.TLSConfig).IsNil() && pulsar.TLSConfig.HasSecretVolume() {
			volumeMounts = append(volumeMounts, generateVolumeMountFromTLSConfig(pulsar.TLSConfig))
		}

		volumeMounts = append(volumeMounts, definedVolumeMounts...)

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

		image := downloaderImage
		if image == "" {
			image = DownloaderImage
		}

		componentPackage = fmt.Sprintf("%s/%s", DownloadDir, getFilenameOfComponentPackage(componentPackage))

		downloaderContainer = &corev1.Container{
			Name:  DownloaderName,
			Image: image,
			Command: []string{"sh", "-c",
				strings.Join(getDownloadCommand(downloadPath, componentPackage, pulsar.TLSSecret != "",
					pulsar.AuthSecret != "", pulsar.TLSConfig, pulsar.AuthConfig), " ")},
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
		Spec: *MakeStatefulSetSpec(replicas, container, podVolumes, labels, policy,
			MakeHeadlessServiceName(objectMeta.Name), downloaderContainer),
	}
}

func MakeStatefulSetSpec(replicas *int32, container *corev1.Container,
	volumes []corev1.Volume, labels map[string]string, policy v1alpha1.PodPolicy,
	serviceName string, downloaderContainer *corev1.Container) *appsv1.StatefulSetSpec {
	return &appsv1.StatefulSetSpec{
		Replicas: replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		Template:            *MakePodTemplate(container, volumes, labels, policy, downloaderContainer),
		PodManagementPolicy: appsv1.ParallelPodManagement,
		UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
			Type: appsv1.RollingUpdateStatefulSetStrategyType,
		},
		ServiceName: serviceName,
	}
}

func MakePodTemplate(container *corev1.Container, volumes []corev1.Volume,
	labels map[string]string, policy v1alpha1.PodPolicy,
	downloaderContainer *corev1.Container) *corev1.PodTemplateSpec {
	podSecurityContext := getDefaultRunnerPodSecurityContext(DefaultRunnerUserID, DefaultRunnerGroupID, false)
	if policy.SecurityContext != nil {
		podSecurityContext = policy.SecurityContext
	}
	initContainers := policy.InitContainers
	if downloaderContainer != nil {
		initContainers = append(initContainers, *downloaderContainer)
	}
	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      mergeLabels(labels, Configs.ResourceLabels, policy.Labels),
			Annotations: generateAnnotations(Configs.ResourceAnnotations, policy.Annotations),
		},
		Spec: corev1.PodSpec{
			InitContainers:                initContainers,
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

func MakeJavaFunctionCommand(downloadPath, packageFile, name, clusterName, generateLogConfigCommand, logLevel, details, memory, extraDependenciesDir, uid string,
	javaOpts []string, authProvided, tlsProvided bool, secretMaps map[string]v1alpha1.SecretRef,
	state *v1alpha1.Stateful,
	tlsConfig TLSConfig, authConfig *v1alpha1.AuthConfig, healthCheckInterval *int32) []string {
	processCommand := setShardIDEnvironmentVariableCommand() + " && " + generateLogConfigCommand +
		strings.Join(getProcessJavaRuntimeArgs(name, packageFile, clusterName, logLevel, details,
			memory, extraDependenciesDir, uid, javaOpts, authProvided, tlsProvided, secretMaps, state, tlsConfig,
			authConfig, healthCheckInterval), " ")
	if downloadPath != "" && !utils.EnableInitContainers {
		// prepend download command if the downPath is provided
		downloadCommand := strings.Join(getLegacyDownloadCommand(downloadPath, packageFile, authProvided, tlsProvided,
			tlsConfig, authConfig), " ")
		processCommand = downloadCommand + " && " + processCommand
	}
	return []string{"sh", "-c", processCommand}
}

func MakePythonFunctionCommand(downloadPath, packageFile, name, clusterName, generateLogConfigCommand, details, uid string,
	authProvided, tlsProvided bool, secretMaps map[string]v1alpha1.SecretRef, state *v1alpha1.Stateful,
	tlsConfig TLSConfig, authConfig *v1alpha1.AuthConfig, healthCheckInterval *int32) []string {
	processCommand := setShardIDEnvironmentVariableCommand() + " && " + generateLogConfigCommand +
		strings.Join(getProcessPythonRuntimeArgs(name, packageFile, clusterName,
			details, uid, authProvided, tlsProvided, secretMaps, state, tlsConfig, authConfig, healthCheckInterval), " ")
	if downloadPath != "" && !utils.EnableInitContainers {
		// prepend download command if the downPath is provided
		downloadCommand := strings.Join(getLegacyDownloadCommand(downloadPath, packageFile, authProvided, tlsProvided,
			tlsConfig, authConfig), " ")
		processCommand = downloadCommand + " && " + processCommand
	}
	return []string{"sh", "-c", processCommand}
}

func MakeGoFunctionCommand(downloadPath, goExecFilePath string, function *v1alpha1.Function) []string {
	processCommand := setShardIDEnvironmentVariableCommand() + " && " +
		strings.Join(getProcessGoRuntimeArgs(goExecFilePath, function), " ")
	if downloadPath != "" && !utils.EnableInitContainers {
		// prepend download command if the downPath is provided
		downloadCommand := strings.Join(getLegacyDownloadCommand(downloadPath, goExecFilePath,
			function.Spec.Pulsar.AuthSecret != "", function.Spec.Pulsar.TLSSecret != "",
			function.Spec.Pulsar.TLSConfig, function.Spec.Pulsar.AuthConfig), " ")
		processCommand = downloadCommand + " && ls -al && pwd &&" + processCommand
	}
	return []string{"sh", "-c", processCommand}
}

func MakeLivenessProbe(interval int32) *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: &corev1.ExecAction{
				Command: []string{"/pulsar/bin/grpcurl",
					"-plaintext",
					"-proto",
					"conf/InstanceCommunication.proto",
					"localhost:" + string(GRPCPort.ContainerPort),
					"proto.InstanceControl/HealthCheck",
				},
			},
		},
		TimeoutSeconds: interval,
		PeriodSeconds:  interval,
	}
}

func getLegacyDownloadCommand(downloadPath, componentPackage string, authProvided, tlsProvided bool,
	tlsConfig TLSConfig, authConfig *v1alpha1.AuthConfig) []string {
	args := []string{
		PulsarAdminExecutableFile,
		"--admin-url",
		"$webServiceURL",
	}
	if authConfig != nil && authConfig.OAuth2Config != nil {
		args = append(args, []string{
			"--auth-plugin",
			OAuth2AuthenticationPlugin,
			"--auth-params",
			authConfig.OAuth2Config.AuthenticationParameters(),
		}...)
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

func getDownloadCommand(downloadPath, componentPackage string, tlsProvided, authProvided bool, tlsConfig TLSConfig,
	authConfig *v1alpha1.AuthConfig) []string {
	var args []string
	// activate oauth2 for pulsarctl
	if authConfig != nil && authConfig.OAuth2Config != nil {
		args = []string{
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
	} else if authProvided {
		args = []string{
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
	} else {
		args = []string{
			PulsarctlExecutableFile,
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

func generateJavaLogConfigCommand(runtime *v1alpha1.JavaRuntime) string {
	if runtime == nil || (runtime.Log != nil && runtime.Log.LogConfig != nil) {
		return ""
	}
	if log4jXML, err := renderJavaInstanceLog4jXMLTemplate(runtime); err == nil {
		generateConfigFileCommand := []string{
			"mkdir", "-p", JavaLogConfigDirectory, "&&",
			"echo", fmt.Sprintf("\"%s\"", log4jXML), ">", DefaultJavaLogConfigPath,
			"&& ",
		}
		return strings.Join(generateConfigFileCommand, " ")
	}
	return ""
}

func renderJavaInstanceLog4jXMLTemplate(runtime *v1alpha1.JavaRuntime) (string, error) {
	tmpl := template.Must(template.New("spec").Parse(javaLog4jXMLTemplate))
	var tpl bytes.Buffer
	type logConfig struct {
		RollingEnabled bool
		Level          string
		Policy         template.HTML
	}
	lc := &logConfig{}
	lc.Level = "INFO"
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
	}
	if err := tmpl.Execute(&tpl, lc); err != nil {
		log.Error(err, "failed to render java instance log4j template")
		return "", err
	}
	return tpl.String(), nil
}

func generatePythonLogConfigCommand(name string, runtime *v1alpha1.PythonRuntime) string {
	commands := "sed -i.bak 's/^  Log.setLevel/#&/' /pulsar/instances/python-instance/log.py && "
	if runtime == nil || (runtime.Log != nil && runtime.Log.LogConfig != nil) {
		return commands
	}
	if loggingINI, err := renderPythonInstanceLoggingINITemplate(name, runtime); err == nil {
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

func renderPythonInstanceLoggingINITemplate(name string, runtime *v1alpha1.PythonRuntime) (string, error) {
	tmpl := template.Must(template.New("spec").Parse(pythonLoggingINITemplate))
	var tpl bytes.Buffer
	type logConfig struct {
		RollingEnabled bool
		Level          string
		Policy         template.HTML
		Handlers       string
	}
	lc := &logConfig{}
	lc.Level = "INFO"
	lc.Handlers = "stream_handler"
	if runtime.Log != nil && runtime.Log.Level != "" {
		if level := parsePythonLogLevel(runtime); level != "" {
			lc.Level = level
		}
	}
	if runtime.Log != nil && runtime.Log.RotatePolicy != nil {
		lc.RollingEnabled = true
		logFile := fmt.Sprintf("logs/functions/%s-${%s}", name, EnvShardID)
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

func setShardIDEnvironmentVariableCommand() string {
	return fmt.Sprintf("%s=${POD_NAME##*-} && echo shardId=${%s}", EnvShardID, EnvShardID)
}

func getProcessJavaRuntimeArgs(name, packageName, clusterName, logLevel, details, memory, extraDependenciesDir, uid string,
	javaOpts []string, authProvided, tlsProvided bool, secretMaps map[string]v1alpha1.SecretRef,
	state *v1alpha1.Stateful,
	tlsConfig TLSConfig, authConfig *v1alpha1.AuthConfig,
	healthCheckInterval *int32) []string {
	classPath := "/pulsar/instances/java-instance.jar"
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
		fmt.Sprintf("-Dlog4j.configurationFile=%s", DefaultJavaLogConfigPath),
		"-Dpulsar.function.log.dir=logs/functions",
		"-Dpulsar.function.log.file=" + fmt.Sprintf("%s-${%s}", name, EnvShardID),
		setLogLevel,
		"-Xmx" + memory,
		strings.Join(javaOpts, " "),
		"org.apache.pulsar.functions.instance.JavaInstanceMain",
		"--jar",
		packageName,
	}
	sharedArgs := getSharedArgs(details, clusterName, uid, authProvided, tlsProvided, tlsConfig, authConfig, healthCheckInterval)
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
	secretMaps map[string]v1alpha1.SecretRef, state *v1alpha1.Stateful, tlsConfig TLSConfig,
	authConfig *v1alpha1.AuthConfig, healthCheckInterval *int32) []string {
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
		DefaultPythonLogConfigPath,
		"--install_usercode_dependencies",
		"true",
		// TODO: Maybe we don't need installUserCodeDependencies, dependency_repository, and pythonExtraDependencyRepository
	}
	sharedArgs := getSharedArgs(details, clusterName, uid, authProvided, tlsProvided, tlsConfig, authConfig, healthCheckInterval)
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
func getSharedArgs(details, clusterName, uid string, authProvided bool, tlsProvided bool,
	tlsConfig TLSConfig, authConfig *v1alpha1.AuthConfig, healthCheckInterval *int32) []string {
	var hInterval int32 = -1
	if healthCheckInterval != nil && *healthCheckInterval > 0 {
		hInterval = *healthCheckInterval
	}
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
		strconv.Itoa(int(hInterval)),
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

func generateContainerVolumesFromLogConfigs(confs map[int32]*v1alpha1.LogConfig) []corev1.Volume {
	volumes := []corev1.Volume{}
	if len(confs) > 0 {
		if conf, exist := confs[javaRuntimeLog]; exist {
			javaLogConfigVolume := &corev1.Volume{
				Name: generateVolumeNameFromLogConfigs(conf.Name, "java"),
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: conf.Name,
						},
						Items: []corev1.KeyToPath{
							{
								Key:  conf.Key,
								Path: JavaLogConfigFile,
							},
						},
					},
				},
			}
			volumes = append(volumes, *javaLogConfigVolume)
		}
		if conf, exist := confs[pythonRuntimeLog]; exist {
			pythonLogConfigVolume := &corev1.Volume{
				Name: generateVolumeNameFromLogConfigs(conf.Name, "python"),
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: conf.Name,
						},
						Items: []corev1.KeyToPath{
							{
								Key:  conf.Key,
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

func generateVolumeMountFromLogConfigs(confs map[int32]*v1alpha1.LogConfig) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{}
	if len(confs) > 0 {
		if conf, exist := confs[javaRuntimeLog]; exist {
			javaLogConfigVolumeMount := &corev1.VolumeMount{
				Name:      generateVolumeNameFromLogConfigs(conf.Name, "java"),
				MountPath: JavaLogConfigDirectory,
			}
			volumeMounts = append(volumeMounts, *javaLogConfigVolumeMount)
		}
		if conf, exist := confs[pythonRuntimeLog]; exist {
			pythonLogConfigVolumeMount := &corev1.VolumeMount{
				Name:      generateVolumeNameFromLogConfigs(conf.Name, "python"),
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
		if !strings.HasPrefix(downloadPath, "/") {
			mountPath = WorkDir + downloadPath
		}
		return []corev1.VolumeMount{{
			Name:      DownloaderVolume,
			MountPath: mountPath,
			SubPath:   subPath,
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
	logConfs map[int32]*v1alpha1.LogConfig, javaRuntime *v1alpha1.JavaRuntime,
	pythonRuntime *v1alpha1.PythonRuntime,
	goRuntime *v1alpha1.GoRuntime) []corev1.VolumeMount {
	mounts := []corev1.VolumeMount{}
	mounts = append(mounts, volumeMounts...)
	if !reflect.ValueOf(tlsConfig).IsNil() && tlsConfig.HasSecretVolume() {
		mounts = append(mounts, generateVolumeMountFromTLSConfig(tlsConfig))
	}
	if authConfig != nil {
		if authConfig.OAuth2Config != nil {
			mounts = append(mounts, generateVolumeMountFromOAuth2Config(authConfig.OAuth2Config))
		}
	}
	if utils.EnableInitContainers {
		mounts = append(mounts, generateDownloaderVolumeMountsForRuntime(javaRuntime, pythonRuntime, goRuntime)...)
	}
	mounts = append(mounts, generateContainerVolumeMountsFromProducerConf(producerConf)...)
	mounts = append(mounts, generateContainerVolumeMountsFromConsumerConfigs(consumerConfs)...)
	mounts = append(mounts, generateVolumeMountFromLogConfigs(logConfs)...)
	return mounts
}

func generatePodVolumes(podVolumes []corev1.Volume, producerConf *v1alpha1.ProducerConfig,
	consumerConfs map[string]v1alpha1.ConsumerConfig, tlsConfig TLSConfig, authConfig *v1alpha1.AuthConfig,
	logConf map[int32]*v1alpha1.LogConfig) []corev1.Volume {
	volumes := []corev1.Volume{}
	volumes = append(volumes, podVolumes...)
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
	volumes = append(volumes, generateContainerVolumesFromLogConfigs(logConf)...)
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

func getTLSTrustCertPath(tlsVolume TLSConfig, path string) string {
	return fmt.Sprintf("%s/%s", tlsVolume.GetMountPath(), path)
}

const (
	javaRuntimeLog = iota
	pythonRuntimeLog
	golangRuntimeLog
)

func getRuntimeLogConfigNames(java *v1alpha1.JavaRuntime, python *v1alpha1.PythonRuntime,
	golang *v1alpha1.GoRuntime) map[int32]*v1alpha1.LogConfig {
	logConfMap := map[int32]*v1alpha1.LogConfig{}

	if java != nil && java.Log != nil && java.Log.LogConfig != nil {
		logConfMap[javaRuntimeLog] = java.Log.LogConfig
	}
	if python != nil && python.Log != nil && python.Log.LogConfig != nil {
		logConfMap[pythonRuntimeLog] = python.Log.LogConfig
	}
	if golang != nil && golang.Log != nil && golang.Log.LogConfig != nil {
		logConfMap[golangRuntimeLog] = golang.Log.LogConfig
	}
	return logConfMap
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

func CheckIfHPASpecIsEqual(spec *autov2beta2.HorizontalPodAutoscalerSpec,
	desiredSpec *autov2beta2.HorizontalPodAutoscalerSpec) bool {
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
