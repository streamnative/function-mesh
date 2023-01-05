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
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	autov2beta2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/spec"
)

func (r *FunctionReconciler) ObserveFunctionStatefulSet(ctx context.Context, function *v1alpha1.Function) error {
	defer function.SaveStatus(ctx, r.Log, r.Client)

	statefulSet := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: function.Namespace,
		Name:      spec.MakeFunctionObjectMeta(function).Name,
	}, statefulSet)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("function statefulset is not ready yet...",
				"namespace", function.Namespace, "name", function.Name,
				"statefulset name", statefulSet.Name)
			function.SetCondition(v1alpha1.StatefulSet, v1alpha1.CreateCondition(
				v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
				"function statefulset is not ready yet..."))
			return nil
		}
		function.SetCondition(v1alpha1.StatefulSet, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.StatefulSetError,
			fmt.Sprintf("failed to fetch function statefulset: %v", err)))
		return err
	}

	selector, err := metav1.LabelSelectorAsSelector(statefulSet.Spec.Selector)
	if err != nil {
		r.Log.Error(err, "error retrieving statefulset selector")
		function.SetCondition(v1alpha1.StatefulSet, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.StatefulSetError,
			fmt.Sprintf("error retrieving statefulset selector: %v", err)))
		return err
	}
	function.Status.Selector = selector.String()
	function.Status.Replicas = *function.Spec.Replicas

	if r.checkIfStatefulSetNeedUpdate(statefulSet, function) {
		function.SetCondition(v1alpha1.StatefulSet, v1alpha1.CreateCondition(
			v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
			"need to update function statefulset..."))
		return nil
	}

	if statefulSet.Status.ReadyReplicas == *function.Spec.Replicas {
		function.SetCondition(v1alpha1.StatefulSet, v1alpha1.CreateCondition(
			v1alpha1.Ready, metav1.ConditionTrue, v1alpha1.StatefulSetIsReady,
			""))
		return nil
	}
	function.SetCondition(v1alpha1.StatefulSet, v1alpha1.CreateCondition(
		v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
		"function statefulset is pending creation..."))
	return nil
}

func (r *FunctionReconciler) ApplyFunctionStatefulSet(ctx context.Context, function *v1alpha1.Function, newGeneration bool) error {
	defer function.SaveStatus(ctx, r.Log, r.Client)

	condition := function.Status.Conditions[v1alpha1.StatefulSet]
	if condition.Status == metav1.ConditionTrue && !newGeneration {
		return nil
	}

	desiredStatefulSet := spec.MakeFunctionStatefulSet(function)
	desiredStatefulSetSpec := desiredStatefulSet.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredStatefulSet, func() error {
		// function statefulSet mutate logic
		desiredStatefulSet.Spec = desiredStatefulSetSpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error create or update statefulset workload for function",
			"namespace", function.Namespace, "name", function.Name,
			"statefulset name", desiredStatefulSet.Name)
		function.SetCondition(v1alpha1.StatefulSet, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.ErrorCreatingStatefulSet,
			fmt.Sprintf("error create or update statefulset workload for function: %v", err)))
		return err
	}
	function.SetCondition(v1alpha1.StatefulSet, v1alpha1.CreateCondition(
		v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
		"creating or updating statefulset workload for function..."))
	return nil
}

func (r *FunctionReconciler) ObserveFunctionService(ctx context.Context, function *v1alpha1.Function) error {
	defer function.SaveStatus(ctx, r.Log, r.Client)

	svc := &corev1.Service{}
	svcName := spec.MakeHeadlessServiceName(spec.MakeFunctionObjectMeta(function).Name)
	err := r.Get(ctx, types.NamespacedName{Namespace: function.Namespace,
		Name: svcName}, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("function service is not created...",
				"namespace", function.Namespace, "name", function.Name,
				"service name", svcName)
			function.SetCondition(v1alpha1.Service, v1alpha1.CreateCondition(
				v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
				"function service is not created..."))
			return nil
		}
		function.SetCondition(v1alpha1.Service, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.ServiceError,
			fmt.Sprintf("failed to fetch function service: %v", err)))
		return err
	}

	function.SetCondition(v1alpha1.Service, v1alpha1.CreateCondition(
		v1alpha1.Ready, metav1.ConditionTrue, v1alpha1.ServiceIsReady,
		""))
	return nil
}

func (r *FunctionReconciler) ApplyFunctionService(ctx context.Context, function *v1alpha1.Function, newGeneration bool) error {
	defer function.SaveStatus(ctx, r.Log, r.Client)

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
		function.SetCondition(v1alpha1.Service, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.ErrorCreatingService,
			fmt.Sprintf("error create or update service for function: %v", err)))
		return err
	}
	function.SetCondition(v1alpha1.Service, v1alpha1.CreateCondition(
		v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
		"creating or updating service for function..."))
	return nil
}

func (r *FunctionReconciler) ObserveFunctionHPA(ctx context.Context, function *v1alpha1.Function) error {
	defer function.SaveStatus(ctx, r.Log, r.Client)

	if function.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
		delete(function.Status.Conditions, v1alpha1.HPA)
		return nil
	}

	hpa := &autov2beta2.HorizontalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{Namespace: function.Namespace,
		Name: spec.MakeFunctionObjectMeta(function).Name}, hpa)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("hpa is not created for function...",
				"namespace", function.Namespace, "name", function.Name,
				"hpa name", hpa.Name)
			function.SetCondition(v1alpha1.HPA, v1alpha1.CreateCondition(
				v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
				"hpa is not created for function..."))
			return nil
		}
		function.SetCondition(v1alpha1.HPA, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.HPAError,
			fmt.Sprintf("failed to fetch function hpa: %v", err)))
		return err
	}

	if r.checkIfHPANeedUpdate(hpa, function) {
		function.SetCondition(v1alpha1.HPA, v1alpha1.CreateCondition(
			v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
			"need to update function hpa..."))
		return nil
	}

	function.SetCondition(v1alpha1.HPA, v1alpha1.CreateCondition(
		v1alpha1.Ready, metav1.ConditionTrue, v1alpha1.HPAIsReady,
		""))
	return nil
}

func (r *FunctionReconciler) ApplyFunctionHPA(ctx context.Context, function *v1alpha1.Function, newGeneration bool) error {
	defer function.SaveStatus(ctx, r.Log, r.Client)

	if function.Spec.MaxReplicas == nil {
		// HPA not enabled, clear the exists HPA
		hpa := &autov2beta2.HorizontalPodAutoscaler{}
		hpa.Namespace = function.Namespace
		hpa.Name = spec.MakeFunctionObjectMeta(function).Name
		if err := r.Delete(ctx, hpa); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			function.SetCondition(v1alpha1.HPA, v1alpha1.CreateCondition(
				v1alpha1.Error, metav1.ConditionFalse, v1alpha1.HPAError,
				fmt.Sprintf("failed to delete function hpa: %v", err)))
			return err
		}
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
		function.SetCondition(v1alpha1.HPA, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.ErrorCreatingHPA,
			fmt.Sprintf("error create or update hpa for function: %v", err)))
		return err
	}
	function.SetCondition(v1alpha1.HPA, v1alpha1.CreateCondition(
		v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
		"creating or updating hpa for function..."))
	return nil
}

func (r *FunctionReconciler) ObserveFunctionVPA(ctx context.Context, function *v1alpha1.Function) error {
	defer function.SaveStatus(ctx, r.Log, r.Client)

	if function.Spec.Pod.VPA == nil {
		delete(function.Status.Conditions, v1alpha1.VPA)
		return nil
	}

	vpa := &vpav1.VerticalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: function.Namespace,
		Name:      spec.MakeFunctionObjectMeta(function).Name,
	}, vpa)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("vpa is not created for function...",
				"namespace", function.Namespace, "name", function.Name,
				"vpa name", vpa.Name)
			function.SetCondition(v1alpha1.VPA, v1alpha1.CreateCondition(
				v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
				"vpa is not created for function..."))
			return nil
		}
		function.SetCondition(v1alpha1.VPA, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.VPAError,
			fmt.Sprintf("failed to fetch function vpa: %v", err)))
		return err
	}

	if r.checkIfVPANeedUpdate(vpa, function) {
		function.SetCondition(v1alpha1.VPA, v1alpha1.CreateCondition(
			v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
			"need to update function vpa..."))
		return nil
	}

	function.SetCondition(v1alpha1.VPA, v1alpha1.CreateCondition(
		v1alpha1.Ready, metav1.ConditionTrue, v1alpha1.VPAIsReady,
		""))
	return nil
}

func (r *FunctionReconciler) ApplyFunctionVPA(ctx context.Context, function *v1alpha1.Function, newGeneration bool) error {
	defer function.SaveStatus(ctx, r.Log, r.Client)

	if function.Spec.Pod.VPA == nil {
		// VPA not enabled, clear the exists VPA
		vpa := &vpav1.VerticalPodAutoscaler{}
		vpa.Namespace = function.Namespace
		vpa.Name = spec.MakeFunctionObjectMeta(function).Name
		if err := r.Delete(ctx, vpa); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			function.SetCondition(v1alpha1.VPA, v1alpha1.CreateCondition(
				v1alpha1.Error, metav1.ConditionFalse, v1alpha1.VPAError,
				fmt.Sprintf("failed to delete function vpa: %v", err)))
			return err
		}
		return nil
	}

	condition := function.Status.Conditions[v1alpha1.VPA]
	if condition.Status == metav1.ConditionTrue && !newGeneration {
		return nil
	}

	desiredVPA := spec.MakeFunctionVPA(function)
	desiredVPASpec := desiredVPA.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredVPA, func() error {
		// function vpa mutate logic
		desiredVPA.Spec = desiredVPASpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error create or update vpa for function",
			"namespace", function.Namespace, "name", function.Name,
			"vpa name", desiredVPA.Name)
		function.SetCondition(v1alpha1.VPA, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.ErrorCreatingVPA,
			fmt.Sprintf("error create or update vpa for function: %v", err)))
		return err
	}
	function.SetCondition(v1alpha1.VPA, v1alpha1.CreateCondition(
		v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
		"creating or updating vpa for function..."))
	return nil
}

func (r *FunctionReconciler) UpdateObservedGeneration(ctx context.Context, function *v1alpha1.Function) {
	defer function.SaveStatus(ctx, r.Log, r.Client)
	function.Status.ObservedGeneration = function.Generation
}

func (r *FunctionReconciler) checkIfStatefulSetNeedUpdate(statefulSet *appsv1.StatefulSet, function *v1alpha1.Function) bool {
	return !spec.CheckIfStatefulSetSpecIsEqual(&statefulSet.Spec, &spec.MakeFunctionStatefulSet(function).Spec)
}

func (r *FunctionReconciler) checkIfHPANeedUpdate(hpa *autov2beta2.HorizontalPodAutoscaler, function *v1alpha1.Function) bool {
	return !spec.CheckIfHPASpecIsEqual(&hpa.Spec, &spec.MakeFunctionHPA(function).Spec)
}

func (r *FunctionReconciler) checkIfVPANeedUpdate(vpa *vpav1.VerticalPodAutoscaler, function *v1alpha1.Function) bool {
	return !spec.CheckIfVPASpecIsEqual(&vpa.Spec, &spec.MakeFunctionVPA(function).Spec)
}
