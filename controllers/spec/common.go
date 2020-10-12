package spec

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	autov1 "k8s.io/api/autoscaling/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ENV_SHARD_ID = "SHARD_ID"
const FUNCTIONS_INSTANCE_CLASSPATH = "pulsar.functions.instance.classpath"
const PULSAR_CONFIG = "pulsar-config"
const PathPulsarClusterConfigs = "/pulsar/cluster/configs"

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

func MakeHPA(objectMeta *metav1.ObjectMeta, minReplicas, maxReplicas int32, kind string) *autov1.HorizontalPodAutoscaler {
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

func MakeStatefulSet(objectMeta *metav1.ObjectMeta, replicas *int32, container *corev1.Container, labels map[string]string, config string) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: *objectMeta,
		Spec:       *MakeStatefulSetSpec(replicas, container, labels, config),
	}
}

func MakeStatefulSetSpec(replicas *int32, container *corev1.Container, labels map[string]string, config string) *appsv1.StatefulSetSpec {
	return &appsv1.StatefulSetSpec{
		Replicas: replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		Template:            *MakePodTemplate(container, labels, config),
		PodManagementPolicy: appsv1.ParallelPodManagement,
		UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
			Type: appsv1.RollingUpdateStatefulSetStrategyType,
		},
	}
}

func MakePodTemplate(container *corev1.Container, labels map[string]string, configName string) *corev1.PodTemplateSpec {
	ZeroGracePeriod := int64(0)
	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			// Tolerations: nil TODO
			Containers:                    []corev1.Container{*container},
			TerminationGracePeriodSeconds: &ZeroGracePeriod,
			Volumes: []corev1.Volume{{
				Name: PULSAR_CONFIG,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: configName,
						},
					},
				},
			}},
		},
	}
}

func MakeCommand(downloadPath, packageFile, name, clusterName, details string) []string {
	processCommand := setShardIdEnvironmentVariableCommand() + " && " +
		strings.Join(getProcessArgs(name, packageFile, clusterName, details), " ")
	if downloadPath != "" {
		// prepend download command if the downPath is provided
		downloadCommand := strings.Join(getDownloadCommand(downloadPath, packageFile), " ")
		processCommand = downloadCommand + " && " + processCommand
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

func setShardIdEnvironmentVariableCommand() string {
	return fmt.Sprintf("%s=${POD_NAME##*-} && echo shardId=${%s}", ENV_SHARD_ID, ENV_SHARD_ID)
}

func getProcessArgs(name string, packageName string, clusterName string, details string) []string {
	// TODO support multiple runtime
	return []string{
		"exec",
		"java",
		"-cp",
		"/pulsar/instances/java-instance.jar",
		fmt.Sprintf("-D%s=%s", FUNCTIONS_INSTANCE_CLASSPATH, "/pulsar/lib/*"),
		"-Dlog4j.configurationFile=kubernetes_instance_log4j2.xml", // todo
		"-Dpulsar.function.log.dir=logs/functions",
		"-Dpulsar.function.log.file=" + fmt.Sprintf("%s-${%s}", name, ENV_SHARD_ID),
		"-Xmx1G", // TODO
		"org.apache.pulsar.functions.instance.JavaInstanceMain",
		"--jar",
		packageName,
		"--instance_id",
		"${" + ENV_SHARD_ID + "}",
		"--function_id",
		fmt.Sprintf("${%s}-%d", ENV_SHARD_ID, time.Now().Unix()),
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
}
