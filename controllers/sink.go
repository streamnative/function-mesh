package controllers

import (
	"context"

	"github.com/streamnative/function-mesh/api/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/spec"
	appsv1 "k8s.io/api/apps/v1"
	autov1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *SinkReconciler) ObserveSinkStatefulSet(ctx context.Context, req ctrl.Request, sink *v1alpha1.Sink) error {
	condition, ok := sink.Status.Conditions[v1alpha1.StatefulSet]
	if !ok {
		sink.Status.Conditions[v1alpha1.StatefulSet] = v1alpha1.ResourceCondition{
			Condition: v1alpha1.StatefulSetReady,
			Status:    metav1.ConditionFalse,
			Action:    v1alpha1.Create,
		}
		return nil
	}

	statefulSet := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: sink.Namespace,
		Name:      sink.Name,
	}, statefulSet)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("sink is not ready yet...")
			return nil
		}
		return err
	}

	selector, err := metav1.LabelSelectorAsSelector(statefulSet.Spec.Selector)
	if err != nil {
		r.Log.Error(err, "error retrieving statefulSet selector")
		return err
	}
	sink.Status.Selector = selector.String()

	if *statefulSet.Spec.Replicas != *sink.Spec.Replicas {
		condition.Status = metav1.ConditionFalse
		condition.Action = v1alpha1.Update
		sink.Status.Conditions[v1alpha1.StatefulSet] = condition
		return nil
	}

	if statefulSet.Status.ReadyReplicas == *sink.Spec.Replicas {
		condition.Action = v1alpha1.NoAction
		condition.Status = metav1.ConditionTrue
	} else {
		condition.Action = v1alpha1.Wait
	}
	sink.Status.Replicas = *statefulSet.Spec.Replicas
	sink.Status.Conditions[v1alpha1.StatefulSet] = condition

	return nil
}

func (r *SinkReconciler) ApplySinkStatefulSet(ctx context.Context, req ctrl.Request, sink *v1alpha1.Sink) error {
	condition := sink.Status.Conditions[v1alpha1.StatefulSet]

	if condition.Status == metav1.ConditionTrue {
		return nil
	}

	switch condition.Action {
	case v1alpha1.Create:
		statefulSet := spec.MakeSinkStatefulSet(sink)
		if err := r.Create(ctx, statefulSet); err != nil {
			r.Log.Error(err, "failed to create new sink statefulSet")
			return err
		}
	case v1alpha1.Update:
		statefulSet := spec.MakeSinkStatefulSet(sink)
		if err := r.Update(ctx, statefulSet); err != nil {
			r.Log.Error(err, "failed to update the sink statefulSet")
			return err
		}
	case v1alpha1.Wait:
		// do nothing
	}

	return nil
}

func (r *SinkReconciler) ObserveSinkService(ctx context.Context, req ctrl.Request, sink *v1alpha1.Sink) error {
	condition, ok := sink.Status.Conditions[v1alpha1.Service]
	if !ok {
		sink.Status.Conditions[v1alpha1.Service] = v1alpha1.ResourceCondition{
			Condition: v1alpha1.ServiceReady,
			Status:    metav1.ConditionFalse,
			Action:    v1alpha1.Create,
		}
		return nil
	}

	if condition.Status == metav1.ConditionTrue {
		return nil
	}

	svc := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Namespace: sink.Namespace, Name: sink.Name}, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("service is not created...", "Name", sink.Name)
			return nil
		}
		return err
	}

	condition.Action = v1alpha1.NoAction
	condition.Status = metav1.ConditionTrue
	sink.Status.Conditions[v1alpha1.Service] = condition
	return nil
}

func (r *SinkReconciler) ApplySinkService(ctx context.Context, req ctrl.Request, sink *v1alpha1.Sink) error {
	condition := sink.Status.Conditions[v1alpha1.Service]

	if condition.Status == metav1.ConditionTrue {
		return nil
	}

	switch condition.Action {
	case v1alpha1.Create:
		svc := spec.MakeSinkService(sink)
		if err := r.Create(ctx, svc); err != nil {
			r.Log.Error(err, "failed to expose service for sink", "name", sink.Name)
			return err
		}
	case v1alpha1.Wait:
		// do nothing
	}

	return nil
}

func (r *SinkReconciler) ObserveSinkHPA(ctx context.Context, req ctrl.Request, sink *v1alpha1.Sink) error {
	if *sink.Spec.MaxReplicas == 0 {
		// HPA not enabled, skip further action
		return nil
	}

	condition, ok := sink.Status.Conditions[v1alpha1.HPA]
	if !ok {
		sink.Status.Conditions[v1alpha1.HPA] = v1alpha1.ResourceCondition{
			Condition: v1alpha1.HPAReady,
			Status:    metav1.ConditionFalse,
			Action:    v1alpha1.Create,
		}
		return nil
	}

	if condition.Status == metav1.ConditionTrue {
		return nil
	}

	hpa := &autov1.HorizontalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{Namespace: sink.Namespace, Name: sink.Name}, hpa)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("hpa is not created for sink...", "name", sink.Name)
			return nil
		}
		return err
	}

	condition.Action = v1alpha1.NoAction
	condition.Status = metav1.ConditionTrue
	sink.Status.Conditions[v1alpha1.HPA] = condition
	return nil
}

func (r *SinkReconciler) ApplySinkHPA(ctx context.Context, req ctrl.Request, sink *v1alpha1.Sink) error {
	if *sink.Spec.MaxReplicas == 0 {
		// HPA not enabled, skip further action
		return nil
	}

	condition := sink.Status.Conditions[v1alpha1.HPA]

	if condition.Status == metav1.ConditionTrue {
		return nil
	}

	switch condition.Action {
	case v1alpha1.Create:
		hpa := spec.MakeSinkHPA(sink)
		if err := r.Create(ctx, hpa); err != nil {
			r.Log.Error(err, "failed to create pod autoscaler for sink", "name", sink.Name)
			return err
		}
	case v1alpha1.Wait:
		// do nothing
	}

	return nil
}
