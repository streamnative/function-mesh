package spec

import (
	"github.com/gogo/protobuf/jsonpb"
	"github.com/streamnative/function-mesh/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	autov1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MakeFunctionHPA(function *v1alpha1.Function) *autov1.HorizontalPodAutoscaler {
	objectMeta := MakeFunctionObjectMeta(function)
	return MakeHPA(objectMeta, function.Spec.Replicas, function.Spec.MaxReplicas, function.Kind)
}

func MakeFunctionService(function *v1alpha1.Function) *corev1.Service {
	labels := makeFunctionLabels(function)
	objectMeta := MakeFunctionObjectMeta(function)
	return MakeService(objectMeta, labels)
}

func MakeFunctionStatefulSet(function *v1alpha1.Function) *appsv1.StatefulSet {
	objectMeta := MakeFunctionObjectMeta(function)
	return MakeStatefulSet(objectMeta, &function.Spec.Replicas, MakeFunctionContainer(function),
		makeFunctionLabels(function), function.Spec.Pulsar.PulsarConfig)
}

func MakeFunctionObjectMeta(function *v1alpha1.Function) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		Name:      function.Name,
		Namespace: function.Namespace,
		OwnerReferences: []metav1.OwnerReference{
			*metav1.NewControllerRef(function, function.GroupVersionKind()),
		},
	}
}

func MakeFunctionContainer(function *v1alpha1.Function) *corev1.Container {
	return &corev1.Container{
		// TODO new container to pull user code image and upload jars into bookkeeper
		Name:    "function-instance",
		Image:   "apachepulsar/pulsar-all",
		Command: makeFunctionCommand(function),
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
				LocalObjectReference: v1.LocalObjectReference{Name: function.Spec.Pulsar.PulsarConfig},
			},
		}},
	}
}

func makeFunctionLabels(function *v1alpha1.Function) map[string]string {
	labels := make(map[string]string)
	labels["component"] = "function"
	labels["name"] = function.Name
	labels["namespace"] = function.Namespace

	return labels
}

func makeFunctionCommand(function *v1alpha1.Function) []string {
	return MakeCommand(function.Spec.Java.JarLocation, function.Spec.Java.Jar,
		function.Spec.Name, function.Spec.Pulsar.PulsarConfig, generateFunctionDetailsInJson(function))
}

func generateFunctionDetailsInJson(function *v1alpha1.Function) string {
	functionDetails := convertFunctionDetails(function)
	marshaler := &jsonpb.Marshaler{}
	json, error := marshaler.MarshalToString(functionDetails)
	if error != nil {
		// TODO
		panic(error)
	}
	return json
}
