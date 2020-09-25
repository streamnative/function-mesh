package spec

import (
	"github.com/gogo/protobuf/jsonpb"
	"github.com/streamnative/mesh-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MakeSinkService(sink *v1alpha1.Sink) *corev1.Service {
	labels := MakeSinkLabels(sink)
	objectMeta := MakeSinkObjectMeta(sink)
	return MakeService(objectMeta, labels)
}

func MakeSinkStatefulSet(sink *v1alpha1.Sink) *appsv1.StatefulSet {
	objectMeta := MakeSinkObjectMeta(sink)
	return MakeStatefulSet(objectMeta, &sink.Spec.Replicas, MakeSinkContainer(sink),
		MakeSinkLabels(sink), sink.Spec.Pulsar.PulsarConfig)
}

func MakeSinkObjectMeta(sink *v1alpha1.Sink) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		Name:      sink.Name,
		Namespace: sink.Namespace,
		OwnerReferences: []metav1.OwnerReference{
			*metav1.NewControllerRef(sink, sink.GroupVersionKind()),
		},
	}
}

func MakeSinkContainer(sink *v1alpha1.Sink) *corev1.Container {
	return &corev1.Container{
		// TODO new container to pull user code image and upload jars into bookkeeper
		Name:    "sink-instance",
		Image:   "apachepulsar/pulsar-all",
		Command: MakeSinkCommand(sink),
		Ports:   []corev1.ContainerPort{GRPCPort, MetricsPort},
		Env: []corev1.EnvVar{{
			Name:      "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}},
		}},
		// TODO calculate resource precisely
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("0.2"),
				corev1.ResourceMemory: resource.MustParse("2G")},
			Limits: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("0.2"),
				corev1.ResourceMemory: resource.MustParse("2G")},
		},
		ImagePullPolicy: corev1.PullIfNotPresent,
		VolumeMounts: []corev1.VolumeMount{{
			Name:      PULSAR_CONFIG,
			ReadOnly:  true,
			MountPath: PathPulsarClusterConfigs,
		}},
		EnvFrom: []corev1.EnvFromSource{{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: v1.LocalObjectReference{Name: sink.Spec.Pulsar.PulsarConfig},
			},
		}},
	}
}

func MakeSinkLabels(sink *v1alpha1.Sink) map[string]string {
	labels := make(map[string]string)
	labels["component"] = "sink"
	labels["name"] = sink.Spec.Name
	labels["namespace"] = sink.Spec.Namespace
	labels["tenant"] = sink.Spec.Tenant

	return labels
}

func MakeSinkCommand(sink *v1alpha1.Sink) []string {
	return MakeCommand(sink.Spec.PackageDownloadPath, sink.Spec.SinkPackage,
		sink.Spec.Name, sink.Spec.Pulsar.PulsarConfig, generateSinkDetailsInJson(sink))
}

func generateSinkDetailsInJson(sink *v1alpha1.Sink) string {
	sourceDetails := convertSinkDetails(sink)
	marshaler := &jsonpb.Marshaler{}
	json, error := marshaler.MarshalToString(sourceDetails)
	if error != nil {
		// TODO
		panic(error)
	}
	return json
}
