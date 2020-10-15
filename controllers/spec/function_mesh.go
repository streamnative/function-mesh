package spec

import (
	"github.com/streamnative/function-mesh/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MakeFunctionComponent(functionName string, mesh *v1alpha1.FunctionMesh,
	spec *v1alpha1.FunctionSpec) *v1alpha1.Function {
	return &v1alpha1.Function{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cloud.streamnative.io/v1alpha1",
			Kind:       "Function",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      functionName,
			Namespace: mesh.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mesh, mesh.GroupVersionKind()),
			},
		},
		Spec: *spec,
	}
}

func MakeSourceComponent(sourceName string, mesh *v1alpha1.FunctionMesh, spec *v1alpha1.SourceSpec) *v1alpha1.Source {
	return &v1alpha1.Source{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cloud.streamnative.io/v1alpha1",
			Kind:       "Source",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      sourceName,
			Namespace: mesh.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mesh, mesh.GroupVersionKind()),
			},
		},
		Spec: *spec,
	}
}

func MakeSinkComponent(sinkName string, mesh *v1alpha1.FunctionMesh, spec *v1alpha1.SinkSpec) *v1alpha1.Sink {
	return &v1alpha1.Sink{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cloud.streamnative.io/v1alpha1",
			Kind:       "Sink",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      sinkName,
			Namespace: mesh.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mesh, mesh.GroupVersionKind()),
			},
		},
		Spec: *spec,
	}
}
