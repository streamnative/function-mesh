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

func (r *FunctionReconciler) ObserveFunctionStatefulSet(ctx context.Context, req ctrl.Request, function *v1alpha1.Function) error {
	condition, ok := function.Status.Conditions[v1alpha1.StatefulSet]
	if !ok {
		function.Status.Conditions[v1alpha1.StatefulSet] = v1alpha1.ResourceCondition{
			Condition: v1alpha1.StatefulSetReady,
			Status:    metav1.ConditionFalse,
			Action:    v1alpha1.Create,
		}
		return nil
	}

	statefulSet := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: function.Namespace,
		Name:      function.Name,
	}, statefulSet)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("function is not ready yet...")
			return nil
		}
		return err
	}

	if *statefulSet.Spec.Replicas != function.Spec.Replicas {
		condition.Status = metav1.ConditionFalse
		condition.Action = v1alpha1.Update
		function.Status.Conditions[v1alpha1.StatefulSet] = condition
		return nil
	}

	if statefulSet.Status.ReadyReplicas == function.Spec.Replicas {
		condition.Action = v1alpha1.NoAction
		condition.Status = metav1.ConditionTrue
	} else {
		condition.Action = v1alpha1.Wait
	}
	function.Status.Conditions[v1alpha1.StatefulSet] = condition

	return nil
}

func (r *FunctionReconciler) ApplyFunctionStatefulSet(ctx context.Context, req ctrl.Request, function *v1alpha1.Function) error {
	condition := function.Status.Conditions[v1alpha1.StatefulSet]

	if condition.Status == metav1.ConditionTrue {
		return nil
	}

	switch condition.Action {
	case v1alpha1.Create:
		statefulSet := spec.MakeFunctionStatefulSet(function)
		if err := r.Create(ctx, statefulSet); err != nil {
			r.Log.Error(err, "Failed to create new function statefulSet")
			return err
		}
	case v1alpha1.Update:
		statefulSet := spec.MakeFunctionStatefulSet(function)
		if err := r.Update(ctx, statefulSet); err != nil {
			r.Log.Error(err, "Failed to update the function statefulSet")
			return err
		}
	case v1alpha1.Wait:
		// do nothing
	}

	return nil
}

func (r *FunctionReconciler) ObserveFunctionService(ctx context.Context, req ctrl.Request, function *v1alpha1.Function) error {
	condition, ok := function.Status.Conditions[v1alpha1.Service]
	if !ok {
		function.Status.Conditions[v1alpha1.Service] = v1alpha1.ResourceCondition{
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
	err := r.Get(ctx, types.NamespacedName{Namespace: function.Namespace, Name: function.Name}, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("service is not created...", "Name", function.Name)
			return nil
		}
		return err
	}

	condition.Action = v1alpha1.NoAction
	condition.Status = metav1.ConditionTrue
	function.Status.Conditions[v1alpha1.Service] = condition
	return nil
}

func (r *FunctionReconciler) ApplyFunctionService(ctx context.Context, req ctrl.Request, function *v1alpha1.Function) error {
	condition := function.Status.Conditions[v1alpha1.Service]

	if condition.Status == metav1.ConditionTrue {
		return nil
	}

	switch condition.Action {
	case v1alpha1.Create:
		svc := spec.MakeFunctionService(function)
		if err := r.Create(ctx, svc); err != nil {
			r.Log.Error(err, "failed to expose service for function", "name", function.Name)
			return err
		}
	case v1alpha1.Wait:
		// do nothing
	}

	return nil
}
