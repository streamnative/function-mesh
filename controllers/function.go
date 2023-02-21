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

	computeapi "github.com/streamnative/function-mesh/api/compute/v1alpha2"
	"github.com/streamnative/function-mesh/controllers/spec"
)

func (r *FunctionReconciler) ObserveFunctionStatefulSet(
	ctx context.Context, function *computeapi.Function, helper ReconciliationHelper) error {
	statefulSet := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: function.Namespace,
		Name:      spec.MakeFunctionObjectMeta(function).Name,
	}, statefulSet)
	if err != nil {
		helper.GetState().StatefulSetState = "unready"
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error fetching statefulSet [%w]", err)
	}

	selector, err := metav1.LabelSelectorAsSelector(statefulSet.Spec.Selector)
	if err != nil {
		helper.GetState().StatefulSetState = "unready"
		return fmt.Errorf("error retrieving statefulSet selector [%w]", err)
	}
	function.Status.Selector = selector.String()

	helper.GetState().StatefulSetState = "created"
	if !r.isGenerationChanged(function) && statefulSet.Status.ReadyReplicas == *function.Spec.Replicas {
		helper.GetState().StatefulSetState = "ready"
	}
	function.Status.Replicas = *statefulSet.Spec.Replicas
	return nil
}

func (r *FunctionReconciler) ApplyFunctionStatefulSet(
	ctx context.Context, function *computeapi.Function, helper ReconciliationHelper) error {
	if helper.GetState().StatefulSetState == "ready" {
		return nil
	}
	desiredStatefulSet := spec.MakeFunctionStatefulSet(function)
	desiredStatefulSetSpec := desiredStatefulSet.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredStatefulSet, func() error {
		// function statefulSet mutate logic
		desiredStatefulSet.Spec = desiredStatefulSetSpec
		return nil
	}); err != nil {
		return fmt.Errorf("error creating or updating statefulSet [%w]", err)
	}
	helper.GetState().StatefulSetState = "created"
	return nil
}

func (r *FunctionReconciler) ObserveFunctionService(
	ctx context.Context, function *computeapi.Function, helper ReconciliationHelper) error {
	svc := &corev1.Service{}
	svcName := spec.MakeHeadlessServiceName(spec.MakeFunctionObjectMeta(function).Name)
	err := r.Get(ctx, types.NamespacedName{Namespace: function.Namespace,
		Name: svcName}, svc)
	if err != nil {
		helper.GetState().ServiceState = "unready"
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error fetching service [%w]", err)
	}

	helper.GetState().ServiceState = "created"
	if !r.isGenerationChanged(function) {
		helper.GetState().ServiceState = "ready"
	}
	return nil
}

func (r *FunctionReconciler) ApplyFunctionService(
	ctx context.Context, function *computeapi.Function, helper ReconciliationHelper) error {
	if helper.GetState().ServiceState == "ready" {
		return nil
	}
	desiredService := spec.MakeFunctionService(function)
	desiredServiceSpec := desiredService.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredService, func() error {
		// function service mutate logic
		desiredService.Spec = desiredServiceSpec
		return nil
	}); err != nil {
		return fmt.Errorf("error creating or updating service [%w]", err)
	}
	helper.GetState().ServiceState = "ready"
	return nil
}

func (r *FunctionReconciler) ObserveFunctionHPA(
	ctx context.Context, function *computeapi.Function, helper ReconciliationHelper) error {
	if function.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
		helper.GetState().HPAState = "delete"
		return nil
	}

	hpa := &autov2beta2.HorizontalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: function.Namespace,
		Name:      spec.MakeFunctionObjectMeta(function).Name}, hpa)
	if err != nil {
		helper.GetState().HPAState = "unready"
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error fetching hpa [%w]", err)
	}

	helper.GetState().HPAState = "created"
	if !r.isGenerationChanged(function) {
		helper.GetState().HPAState = "ready"
	}
	return nil
}

func (r *FunctionReconciler) ApplyFunctionHPA(
	ctx context.Context, function *computeapi.Function, helper ReconciliationHelper) error {
	if function.Spec.MaxReplicas == nil {
		// HPA not enabled, clear the exists HPA
		hpa := &autov2beta2.HorizontalPodAutoscaler{}
		hpa.Namespace = function.Namespace
		hpa.Name = spec.MakeFunctionObjectMeta(function).Name
		if err := r.Delete(ctx, hpa); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return fmt.Errorf("error deleting hpa [%w]", err)
		}
		return nil
	}
	if helper.GetState().HPAState == "ready" {
		return nil
	}
	desiredHPA := spec.MakeFunctionHPA(function)
	desiredHPASpec := desiredHPA.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredHPA, func() error {
		// function hpa mutate logic
		desiredHPA.Spec = desiredHPASpec
		return nil
	}); err != nil {
		return fmt.Errorf("error creating or updating hpa [%w]", err)
	}
	helper.GetState().HPAState = "ready"
	return nil
}

func (r *FunctionReconciler) ObserveFunctionVPA(
	ctx context.Context, function *computeapi.Function, helper ReconciliationHelper) error {
	if function.Spec.Pod.VPA == nil {
		// VPA not enabled, skip further action
		helper.GetState().VPAState = "delete"
		return nil
	}

	vpa := &vpav1.VerticalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: function.Namespace,
		Name:      spec.MakeFunctionObjectMeta(function).Name}, vpa)
	if err != nil {
		helper.GetState().VPAState = "unready"
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error fetching vpa [%w]", err)
	}

	helper.GetState().VPAState = "created"
	if !r.isGenerationChanged(function) {
		helper.GetState().VPAState = "ready"
	}
	return nil
}

func (r *FunctionReconciler) ApplyFunctionVPA(
	ctx context.Context, function *computeapi.Function, helper ReconciliationHelper) error {
	if function.Spec.Pod.VPA == nil {
		// VPA not enabled, clear the exists VPA
		vpa := &vpav1.VerticalPodAutoscaler{}
		vpa.Namespace = function.Namespace
		vpa.Name = spec.MakeFunctionObjectMeta(function).Name
		if err := r.Delete(ctx, vpa); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return fmt.Errorf("error deleting vpa [%w]", err)
		}
		return nil
	}
	if helper.GetState().VPAState == "ready" {
		return nil
	}
	desiredVPA := spec.MakeFunctionVPA(function)
	desiredVPASpec := desiredVPA.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredVPA, func() error {
		// function vpa mutate logic
		desiredVPA.Spec = desiredVPASpec
		return nil
	}); err != nil {
		return fmt.Errorf("error creating or updating vpa [%w]", err)
	}
	helper.GetState().VPAState = "ready"
	return nil
}

func (r *FunctionReconciler) isGenerationChanged(function *computeapi.Function) bool {
	// if the generation has not changed, we do not need to update the component
	return function.Generation != function.Status.ObservedGeneration
}
