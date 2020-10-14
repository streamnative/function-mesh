package spec

import (
	"github.com/gogo/protobuf/jsonpb"
	"github.com/streamnative/function-mesh/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	autov1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MakeSinkHPA(sink *v1alpha1.Sink) *autov1.HorizontalPodAutoscaler {
	objectMeta := MakeSinkObjectMeta(sink)
	return MakeHPA(objectMeta, *sink.Spec.Replicas, *sink.Spec.MaxReplicas, sink.Kind)
}

func MakeSinkService(sink *v1alpha1.Sink) *corev1.Service {
	labels := MakeSinkLabels(sink)
	objectMeta := MakeSinkObjectMeta(sink)
	return MakeService(objectMeta, labels)
}

func MakeSinkStatefulSet(sink *v1alpha1.Sink) *appsv1.StatefulSet {
	objectMeta := MakeSinkObjectMeta(sink)
	return MakeStatefulSet(objectMeta, sink.Spec.Replicas, MakeSinkContainer(sink),
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
		Resources:       *generateContainerResourceRequest(sink.Spec.Resources),
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
	labels["name"] = sink.Name
	labels["namespace"] = sink.Namespace

	return labels
}

func MakeSinkCommand(sink *v1alpha1.Sink) []string {
	return MakeCommand(sink.Spec.Java.JarLocation, sink.Spec.Java.Jar,
		sink.Spec.Name, sink.Spec.ClusterName, generateSinkDetailsInJson(sink))
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
