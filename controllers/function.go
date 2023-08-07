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

package controllers

import (
	"context"

	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"

	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/spec"
	appsv1 "k8s.io/api/apps/v1"
	autov2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *FunctionReconciler) ObserveFunctionStatefulSet(ctx context.Context, function *v1alpha1.Function) error {
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
		Name:      spec.MakeFunctionObjectMeta(function).Name,
	}, statefulSet)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("function statefulSet is not ready yet...",
				"namespace", function.Namespace, "name", function.Name,
				"statefulSet name", statefulSet.Name)
			condition.Status = metav1.ConditionFalse
			condition.Action = v1alpha1.Create
			function.Status.Conditions[v1alpha1.StatefulSet] = condition
			return nil
		}
		return err
	}

	selector, err := metav1.LabelSelectorAsSelector(statefulSet.Spec.Selector)
	if err != nil {
		r.Log.Error(err, "error retrieving statefulSet selector")
		return err
	}
	function.Status.Selector = selector.String()

	if r.checkIfStatefulSetNeedUpdate(statefulSet, function) {
		condition.Status = metav1.ConditionFalse
		condition.Action = v1alpha1.Update
		function.Status.Conditions[v1alpha1.StatefulSet] = condition
		return nil
	}

	if statefulSet.Status.ReadyReplicas == *function.Spec.Replicas {
		condition.Action = v1alpha1.NoAction
	} else {
		condition.Action = v1alpha1.Wait
	}
	condition.Status = metav1.ConditionTrue
	function.Status.Replicas = *statefulSet.Spec.Replicas
	function.Status.Conditions[v1alpha1.StatefulSet] = condition
	return nil
}

func (r *FunctionReconciler) ApplyFunctionStatefulSet(ctx context.Context, function *v1alpha1.Function,
	newGeneration bool) error {
	condition := function.Status.Conditions[v1alpha1.StatefulSet]
	if condition.Status == metav1.ConditionTrue && !newGeneration {
		return nil
	}
	desiredStatefulSet := spec.MakeFunctionStatefulSet(function)
	keepStatefulSetUnchangeableFields(ctx, r, r.Log, desiredStatefulSet)
	desiredStatefulSetSpec := desiredStatefulSet.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredStatefulSet, func() error {
		// function statefulSet mutate logic
		desiredStatefulSet.Spec = desiredStatefulSetSpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error create or update statefulSet workload for function",
			"namespace", function.Namespace, "name", function.Name,
			"statefulSet name", desiredStatefulSet.Name)
		return err
	}
	return nil
}

func (r *FunctionReconciler) ObserveFunctionService(ctx context.Context, function *v1alpha1.Function) error {
	condition, ok := function.Status.Conditions[v1alpha1.Service]
	if !ok {
		function.Status.Conditions[v1alpha1.Service] = v1alpha1.ResourceCondition{
			Condition: v1alpha1.ServiceReady,
			Status:    metav1.ConditionFalse,
			Action:    v1alpha1.Create,
		}
		return nil
	}

	svc := &corev1.Service{}
	svcName := spec.MakeHeadlessServiceName(spec.MakeFunctionObjectMeta(function).Name)
	err := r.Get(ctx, types.NamespacedName{Namespace: function.Namespace,
		Name: svcName}, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			condition.Status = metav1.ConditionFalse
			condition.Action = v1alpha1.Create
			function.Status.Conditions[v1alpha1.Service] = condition
			r.Log.Info("function service is not created...",
				"namespace", function.Namespace, "name", function.Name,
				"service name", svcName)
			return nil
		}
		return err
	}

	condition.Action = v1alpha1.NoAction
	condition.Status = metav1.ConditionTrue
	function.Status.Conditions[v1alpha1.Service] = condition
	return nil
}

func (r *FunctionReconciler) ApplyFunctionService(ctx context.Context, function *v1alpha1.Function,
	newGeneration bool) error {
	condition := function.Status.Conditions[v1alpha1.Service]
	if condition.Status == metav1.ConditionTrue && !newGeneration {
		return nil
	}
	desiredService := spec.MakeFunctionService(function)
	desiredServiceSpec := desiredService.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredService, func() error {
		// function service mutate logic
		desiredService.Spec = desiredServiceSpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error create or update service for function",
			"namespace", function.Namespace, "name", function.Name,
			"service name", desiredService.Name)
		return err
	}
	return nil
}

func (r *FunctionReconciler) ObserveFunctionHPA(ctx context.Context, function *v1alpha1.Function) error {
	if function.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
		return nil
	}

	condition, ok := function.Status.Conditions[v1alpha1.HPA]
	if !ok {
		function.Status.Conditions[v1alpha1.HPA] = v1alpha1.ResourceCondition{
			Condition: v1alpha1.HPAReady,
			Status:    metav1.ConditionFalse,
			Action:    v1alpha1.Create,
		}
		return nil
	}

	hpa := &autov2.HorizontalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{Namespace: function.Namespace,
		Name: spec.MakeFunctionObjectMeta(function).Name}, hpa)
	if err != nil {
		if errors.IsNotFound(err) {
			condition.Status = metav1.ConditionFalse
			condition.Action = v1alpha1.Create
			function.Status.Conditions[v1alpha1.HPA] = condition
			r.Log.Info("hpa is not created for function...",
				"namespace", function.Namespace, "name", function.Name,
				"hpa name", hpa.Name)
			return nil
		}
		return err
	}

	if r.checkIfHPANeedUpdate(hpa, function) {
		condition.Status = metav1.ConditionFalse
		condition.Action = v1alpha1.Update
		function.Status.Conditions[v1alpha1.HPA] = condition
		return nil
	}

	condition.Action = v1alpha1.NoAction
	condition.Status = metav1.ConditionTrue
	function.Status.Conditions[v1alpha1.HPA] = condition
	return nil
}

func (r *FunctionReconciler) ApplyFunctionHPA(ctx context.Context, function *v1alpha1.Function,
	newGeneration bool) error {
	if function.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
		return nil
	}
	condition := function.Status.Conditions[v1alpha1.HPA]
	if condition.Status == metav1.ConditionTrue && !newGeneration {
		return nil
	}
	desiredHPA := spec.MakeFunctionHPA(function)
	desiredHPASpec := desiredHPA.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredHPA, func() error {
		// function hpa mutate logic
		desiredHPA.Spec = desiredHPASpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error create or update hpa for function",
			"namespace", function.Namespace, "name", function.Name,
			"hpa name", desiredHPA.Name)
		return err
	}
	return nil
}

func (r *FunctionReconciler) ObserveFunctionHPAV2Beta2(ctx context.Context, function *v1alpha1.Function) error {
	if function.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
		return nil
	}

	condition, ok := function.Status.Conditions[v1alpha1.HPA]
	if !ok {
		function.Status.Conditions[v1alpha1.HPA] = v1alpha1.ResourceCondition{
			Condition: v1alpha1.HPAReady,
			Status:    metav1.ConditionFalse,
			Action:    v1alpha1.Create,
		}
		return nil
	}

	hpa := &autoscalingv2beta2.HorizontalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{Namespace: function.Namespace,
		Name: spec.MakeFunctionObjectMeta(function).Name}, hpa)
	if err != nil {
		if errors.IsNotFound(err) {
			condition.Status = metav1.ConditionFalse
			condition.Action = v1alpha1.Create
			function.Status.Conditions[v1alpha1.HPA] = condition
			r.Log.Info("hpa is not created for function...",
				"namespace", function.Namespace, "name", function.Name,
				"hpa name", hpa.Name)
			return nil
		}
		return err
	}

	if r.checkIfHPAV2Beta2NeedUpdate(hpa, function) {
		condition.Status = metav1.ConditionFalse
		condition.Action = v1alpha1.Update
		function.Status.Conditions[v1alpha1.HPA] = condition
		return nil
	}

	condition.Action = v1alpha1.NoAction
	condition.Status = metav1.ConditionTrue
	function.Status.Conditions[v1alpha1.HPA] = condition
	return nil
}

func (r *FunctionReconciler) ApplyFunctionHPAV2Beta2(ctx context.Context, function *v1alpha1.Function,
	newGeneration bool) error {
	if function.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
		return nil
	}
	condition := function.Status.Conditions[v1alpha1.HPA]
	if condition.Status == metav1.ConditionTrue && !newGeneration {
		return nil
	}
	desiredHPA := ConvertHPAV2ToV2beta2(spec.MakeFunctionHPA(function))
	desiredHPASpec := desiredHPA.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredHPA, func() error {
		// function hpa mutate logic
		desiredHPA.Spec = desiredHPASpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error create or update hpa for function",
			"namespace", function.Namespace, "name", function.Name,
			"hpa name", desiredHPA.Name)
		return err
	}
	return nil
}

func (r *FunctionReconciler) ObserveFunctionVPA(ctx context.Context, function *v1alpha1.Function) error {
	return observeVPA(ctx, r, types.NamespacedName{Namespace: function.Namespace,
		Name: spec.MakeFunctionObjectMeta(function).Name}, function.Spec.Pod.VPA, function.Status.Conditions)
}

func (r *FunctionReconciler) ApplyFunctionVPA(ctx context.Context, function *v1alpha1.Function) error {

	condition, ok := function.Status.Conditions[v1alpha1.VPA]

	if !ok || condition.Status == metav1.ConditionTrue {
		return nil
	}

	objectMeta := spec.MakeFunctionObjectMeta(function)
	targetRef := &autov2.CrossVersionObjectReference{
		Kind:       function.Kind,
		Name:       function.Name,
		APIVersion: function.APIVersion,
	}

	err := applyVPA(ctx, r.Client, r.Log, condition, objectMeta, targetRef, function.Spec.Pod.VPA, "function",
		function.Namespace, function.Name)
	if err != nil {
		return err
	}

	return nil
}

func (r *FunctionReconciler) ApplyFunctionCleanUpJob(ctx context.Context, function *v1alpha1.Function) error {
	if !spec.NeedCleanup(function) {
		desiredJob := spec.MakeFunctionCleanUpJob(function)
		if err := r.Delete(ctx, desiredJob); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			r.Log.Error(err, "error delete cleanup job for function",
				"namespace", function.Namespace, "name", function.Name,
				"job name", desiredJob.Name)
			return err
		}
		return nil
	}
	hasCleanupFinalizer := containsCleanupFinalizer(function.ObjectMeta.Finalizers)
	if function.Spec.CleanupSubscription {
		// add finalizer if function is updated to clean up subscription
		if function.ObjectMeta.DeletionTimestamp.IsZero() {
			if !hasCleanupFinalizer {
				desiredJob := spec.MakeFunctionCleanUpJob(function)
				if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredJob, func() error {
					return nil
				}); err != nil {
					r.Log.Error(err, "error create or update clean up job for function",
						"namespace", function.Namespace, "name", function.Name,
						"job name", desiredJob.Name)
					return err
				}
				function.ObjectMeta.Finalizers = append(function.ObjectMeta.Finalizers, CleanUpFinalizerName)
				if err := r.Update(ctx, function); err != nil {
					return err
				}
			}
		} else {
			desiredJob := spec.MakeFunctionCleanUpJob(function)
			// if function is deleting, send an "INT" signal to the cleanup job to clean up subscription
			if hasCleanupFinalizer {
				if err := spec.TriggerCleanup(ctx, r.Client, r.RestClient, r.Config, desiredJob); err != nil {
					r.Log.Error(err, "error send signal to clean up job for function",
						"namespace", function.Namespace, "name", function.Name)
				}
				function.ObjectMeta.Finalizers = removeCleanupFinalizer(function.ObjectMeta.Finalizers)
				if err := r.Update(ctx, function); err != nil {
					return err
				}
			} else {
				// delete the cleanup job
				if err := r.Delete(ctx, desiredJob); err != nil {
					return err
				}
			}
		}
	} else {
		// remove finalizer if function is updated to not cleanup subscription
		if hasCleanupFinalizer {
			function.ObjectMeta.Finalizers = removeCleanupFinalizer(function.ObjectMeta.Finalizers)
			if err := r.Update(ctx, function); err != nil {
				return err
			}

			desiredJob := spec.MakeFunctionCleanUpJob(function)
			// delete the cleanup job
			if err := r.Delete(ctx, desiredJob); err != nil {
				return err
			}

		}
	}
	return nil
}

func (r *FunctionReconciler) checkIfStatefulSetNeedUpdate(statefulSet *appsv1.StatefulSet,
	function *v1alpha1.Function) bool {
	return !spec.CheckIfStatefulSetSpecIsEqual(&statefulSet.Spec, &spec.MakeFunctionStatefulSet(function).Spec)
}

func (r *FunctionReconciler) checkIfHPANeedUpdate(hpa *autov2.HorizontalPodAutoscaler,
	function *v1alpha1.Function) bool {
	return !spec.CheckIfHPASpecIsEqual(&hpa.Spec, &spec.MakeFunctionHPA(function).Spec)
}

func (r *FunctionReconciler) checkIfHPAV2Beta2NeedUpdate(hpa *autoscalingv2beta2.HorizontalPodAutoscaler,
	function *v1alpha1.Function) bool {
	return !spec.CheckIfHPAV2Beta2SpecIsEqual(&hpa.Spec, &ConvertHPAV2ToV2beta2(spec.MakeFunctionHPA(function)).Spec)
}
