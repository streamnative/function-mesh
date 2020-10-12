package controllers

import (
	"context"

	"github.com/streamnative/mesh-operator/api/v1alpha1"
	"github.com/streamnative/mesh-operator/controllers/spec"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *SourceReconciler) ObserveSourceStatefulSet(ctx context.Context, req ctrl.Request, source *v1alpha1.Source) error {
	condition, ok := source.Status.Conditions[v1alpha1.StatefulSet]
	if !ok {
		source.Status.Conditions[v1alpha1.StatefulSet] = v1alpha1.ResourceCondition{
			Condition: v1alpha1.StatefulSetReady,
			Status:    metav1.ConditionFalse,
			Action:    v1alpha1.Create,
		}
		return nil
	}

	statefulSet := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: source.Namespace,
		Name:      source.Name,
	}, statefulSet)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("source is not ready yet...")
			return nil
		}
		return err
	}

	if *statefulSet.Spec.Replicas != source.Spec.Parallelism {
		condition.Status = metav1.ConditionFalse
		condition.Action = v1alpha1.Update
		source.Status.Conditions[v1alpha1.StatefulSet] = condition
		return nil
	}

	if statefulSet.Status.ReadyReplicas == source.Spec.Parallelism {
		condition.Action = v1alpha1.NoAction
		condition.Status = metav1.ConditionTrue
	} else {
		condition.Action = v1alpha1.Wait
	}
	source.Status.Conditions[v1alpha1.StatefulSet] = condition

	return nil
}

func (r *SourceReconciler) ApplySourceStatefulSet(ctx context.Context, req ctrl.Request, source *v1alpha1.Source) error {
	condition := source.Status.Conditions[v1alpha1.StatefulSet]

	if condition.Status == metav1.ConditionTrue {
		return nil
	}

	switch condition.Action {
	case v1alpha1.Create:
		statefulSet := spec.MakeSourceStatefulSet(source)
		if err := r.Create(ctx, statefulSet); err != nil {
			r.Log.Error(err, "failed to create new source statefulSet")
			return err
		}
	case v1alpha1.Update:
		statefulSet := spec.MakeSourceStatefulSet(source)
		if err := r.Update(ctx, statefulSet); err != nil {
			r.Log.Error(err, "failed to update the source statefulSet")
			return err
		}
	case v1alpha1.Wait:
		// do nothing
	}

	return nil
}

func (r *SourceReconciler) ObserveSourceService(ctx context.Context, req ctrl.Request, source *v1alpha1.Source) error {
	condition, ok := source.Status.Conditions[v1alpha1.Service]
	if !ok {
		source.Status.Conditions[v1alpha1.Service] = v1alpha1.ResourceCondition{
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
	err := r.Get(ctx, types.NamespacedName{Namespace: source.Namespace, Name: source.Name}, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("service is not created...", "Name", source.Name)
			return nil
		}
		return err
	}

	condition.Action = v1alpha1.NoAction
	condition.Status = metav1.ConditionTrue
	source.Status.Conditions[v1alpha1.Service] = condition
	return nil
}

func (r *SourceReconciler) ApplySourceService(ctx context.Context, req ctrl.Request, source *v1alpha1.Source) error {
	condition := source.Status.Conditions[v1alpha1.Service]

	if condition.Status == metav1.ConditionTrue {
		return nil
	}

	switch condition.Action {
	case v1alpha1.Create:
		svc := spec.MakeSourceService(source)
		if err := r.Create(ctx, svc); err != nil {
			r.Log.Error(err, "failed to expose service for source", "name", source.Name)
			return err
		}
	case v1alpha1.Wait:
		// do nothing
	}

	return nil
}
