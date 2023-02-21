package v1alpha1

import (
	"time"

	"k8s.io/apimachinery/pkg/api/meta"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	"github.com/streamnative/function-mesh/api/compute/v1alpha2"
	apispec "github.com/streamnative/function-mesh/pkg/spec"
)

// ConvertTo converts this FunctionMesh to the Hub version (v1alpha2).
func (src *FunctionMesh) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha2.FunctionMesh)
	dst.ObjectMeta = src.ObjectMeta
	if err := src.convertSpecTo(dst); err != nil {
		return err
	}
	if err := src.convertStatusTo(dst); err != nil {
		return err
	}
	return nil
}

func (src *FunctionMesh) convertSpecTo(dst *v1alpha2.FunctionMesh) error {
	for _, function := range src.Spec.Functions {
		desiredSpec := &v1alpha2.FunctionSpec{}
		if err := function.convertSpecTo(desiredSpec); err != nil {
			return err
		}
		dst.Spec.Functions = append(dst.Spec.Functions, *desiredSpec)
	}
	for _, source := range src.Spec.Sources {
		desiredSpec := &v1alpha2.SourceSpec{}
		if err := source.convertSpecTo(desiredSpec); err != nil {
			return err
		}
		dst.Spec.Sources = append(dst.Spec.Sources, *desiredSpec)
	}
	for _, sink := range src.Spec.Sinks {
		desiredSpec := &v1alpha2.SinkSpec{}
		if err := sink.convertSpecTo(desiredSpec); err != nil {
			return err
		}
		dst.Spec.Sinks = append(dst.Spec.Sinks, *desiredSpec)
	}
	return nil
}

func (src *FunctionMesh) convertStatusTo(dst *v1alpha2.FunctionMesh) error {
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration

	meshCondition := &metav1.Condition{
		Type:               string(apispec.Ready),
		Status:             metav1.ConditionTrue,
		Reason:             "MeshReady",
		LastTransitionTime: metav1.Time{Time: time.Now()},
		ObservedGeneration: src.Generation,
	}
	if src.Status.Condition != nil &&
		src.Status.Condition.Status == metav1.ConditionFalse &&
		src.Status.Condition.Action != NoAction {
		meshCondition.Status = metav1.ConditionFalse
		meshCondition.Reason = "MeshPending"
	}
	dst.Status.Conditions = []metav1.Condition{
		*meshCondition,
	}

	if len(src.Status.SourceConditions) > 0 {
		dstSourceStatus := []v1alpha2.MeshComponentStatus{}
		for name, cond := range src.Status.SourceConditions {
			s := &v1alpha2.MeshComponentStatus{
				Name:               name,
				Type:               apispec.SourceComponent,
				Message:            "",
				LastTransitionTime: metav1.Time{Time: time.Now()},
			}
			if cond.Status == metav1.ConditionTrue &&
				cond.Action == NoAction {
				s.Status = metav1.ConditionTrue
			} else {
				s.Status = metav1.ConditionFalse
			}
			dstSourceStatus = append(dstSourceStatus, *s)
		}
		dst.Status.SourceStatus = dstSourceStatus
	}

	if len(src.Status.SinkConditions) > 0 {
		dstSinkStatus := []v1alpha2.MeshComponentStatus{}
		for name, cond := range src.Status.SinkConditions {
			s := &v1alpha2.MeshComponentStatus{
				Name:               name,
				Type:               apispec.SinkComponent,
				Message:            "",
				LastTransitionTime: metav1.Time{Time: time.Now()},
			}
			if cond.Status == metav1.ConditionTrue &&
				cond.Action == NoAction {
				s.Status = metav1.ConditionTrue
			} else {
				s.Status = metav1.ConditionFalse
			}
			dstSinkStatus = append(dstSinkStatus, *s)
		}
		dst.Status.SinkStatus = dstSinkStatus
	}

	if len(src.Status.FunctionConditions) > 0 {
		dstFunctionStatus := []v1alpha2.MeshComponentStatus{}
		for name, cond := range src.Status.FunctionConditions {
			s := &v1alpha2.MeshComponentStatus{
				Name:               name,
				Type:               apispec.FunctionComponent,
				Message:            "",
				LastTransitionTime: metav1.Time{Time: time.Now()},
			}
			if cond.Status == metav1.ConditionTrue &&
				cond.Action == NoAction {
				s.Status = metav1.ConditionTrue
			} else {
				s.Status = metav1.ConditionFalse
			}
			dstFunctionStatus = append(dstFunctionStatus, *s)
		}
		dst.Status.FunctionStatus = dstFunctionStatus
	}
	return nil
}

// ConvertFrom converts from the Hub version (v1alpha2) to this version.
func (dst *FunctionMesh) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha2.FunctionMesh)
	src.ObjectMeta = dst.ObjectMeta
	if err := dst.convertSpecFrom(src); err != nil {
		return err
	}
	if err := dst.convertStatusFrom(src); err != nil {
		return err
	}
	return nil
}

func (dst *FunctionMesh) convertSpecFrom(src *v1alpha2.FunctionMesh) error {
	for _, function := range src.Spec.Functions {
		desiredSpec := &FunctionSpec{}
		if err := desiredSpec.convertSpecFrom(&function); err != nil {
			return err
		}
		dst.Spec.Functions = append(dst.Spec.Functions, *desiredSpec)
	}
	for _, source := range src.Spec.Sources {
		desiredSpec := &SourceSpec{}
		if err := desiredSpec.convertSpecFrom(&source); err != nil {
			return err
		}
		dst.Spec.Sources = append(dst.Spec.Sources, *desiredSpec)
	}
	for _, sink := range src.Spec.Sinks {
		desiredSpec := &SinkSpec{}
		if err := desiredSpec.convertSpecFrom(&sink); err != nil {
			return err
		}
		dst.Spec.Sinks = append(dst.Spec.Sinks, *desiredSpec)
	}
	return nil
}

func (dst *FunctionMesh) convertStatusFrom(src *v1alpha2.FunctionMesh) error {
	dst.Status.Condition = &ResourceCondition{
		Condition: MeshReady,
		Status:    metav1.ConditionTrue,
		Action:    NoAction,
	}
	if srcMeshCondition := meta.FindStatusCondition(src.Status.Conditions, string(apispec.Ready)); srcMeshCondition != nil {
		if srcMeshCondition.Status == metav1.ConditionFalse {
			dst.Status.Condition.Action = Wait
			dst.Status.Condition.Status = metav1.ConditionFalse
		}
	}

	if len(src.Status.FunctionStatus) > 0 {
		dst.Status.FunctionConditions = map[string]ResourceCondition{}
		for _, s := range src.Status.FunctionStatus {
			rc := &ResourceCondition{
				Condition: FunctionReady,
				Status:    metav1.ConditionTrue,
				Action:    NoAction,
			}
			if s.Status == metav1.ConditionFalse {
				rc.Status = metav1.ConditionFalse
				rc.Action = Wait
			}
			dst.Status.FunctionConditions[s.Name] = *rc
		}
	}

	if len(src.Status.SinkStatus) > 0 {
		dst.Status.SinkConditions = map[string]ResourceCondition{}
		for _, s := range src.Status.SinkStatus {
			rc := &ResourceCondition{
				Condition: SinkReady,
				Status:    metav1.ConditionTrue,
				Action:    NoAction,
			}
			if s.Status == metav1.ConditionFalse {
				rc.Status = metav1.ConditionFalse
				rc.Action = Wait
			}
			dst.Status.SinkConditions[s.Name] = *rc
		}
	}

	if len(src.Status.SourceStatus) > 0 {
		dst.Status.SourceConditions = map[string]ResourceCondition{}
		for _, s := range src.Status.SourceStatus {
			rc := &ResourceCondition{
				Condition: SourceReady,
				Status:    metav1.ConditionTrue,
				Action:    NoAction,
			}
			if s.Status == metav1.ConditionFalse {
				rc.Status = metav1.ConditionFalse
				rc.Action = Wait
			}
			dst.Status.SourceConditions[s.Name] = *rc
		}
	}
	return nil
}
