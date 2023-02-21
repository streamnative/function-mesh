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

func (r *SourceReconciler) ObserveSourceStatefulSet(
	ctx context.Context, source *computeapi.Source, helper ReconciliationHelper) error {
	statefulSet := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: source.Namespace,
		Name:      spec.MakeSourceObjectMeta(source).Name,
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
	source.Status.Selector = selector.String()

	helper.GetState().StatefulSetState = "created"
	if !r.isGenerationChanged(source) && statefulSet.Status.ReadyReplicas == *source.Spec.Replicas {
		helper.GetState().StatefulSetState = "ready"
	}
	source.Status.Replicas = *statefulSet.Spec.Replicas
	return nil
}

func (r *SourceReconciler) ApplySourceStatefulSet(
	ctx context.Context, source *computeapi.Source, helper ReconciliationHelper) error {
	if helper.GetState().StatefulSetState == "ready" {
		return nil
	}
	desiredStatefulSet := spec.MakeSourceStatefulSet(source)
	desiredStatefulSetSpec := desiredStatefulSet.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredStatefulSet, func() error {
		// source statefulSet mutate logic
		desiredStatefulSet.Spec = desiredStatefulSetSpec
		return nil
	}); err != nil {
		return fmt.Errorf("error creating or updating statefulSet [%w]", err)
	}
	helper.GetState().StatefulSetState = "created"
	return nil
}

func (r *SourceReconciler) ObserveSourceService(
	ctx context.Context, source *computeapi.Source, helper ReconciliationHelper) error {
	svc := &corev1.Service{}
	svcName := spec.MakeHeadlessServiceName(spec.MakeSourceObjectMeta(source).Name)
	err := r.Get(ctx, types.NamespacedName{Namespace: source.Namespace,
		Name: svcName}, svc)
	if err != nil {
		helper.GetState().ServiceState = "unready"
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error fetching service [%w]", err)
	}

	helper.GetState().ServiceState = "created"
	if !r.isGenerationChanged(source) {
		helper.GetState().ServiceState = "ready"
	}
	return nil
}

func (r *SourceReconciler) ApplySourceService(
	ctx context.Context, source *computeapi.Source, helper ReconciliationHelper) error {
	if helper.GetState().ServiceState == "ready" {
		return nil
	}
	desiredService := spec.MakeSourceService(source)
	desiredServiceSpec := desiredService.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredService, func() error {
		// source service mutate logic
		desiredService.Spec = desiredServiceSpec
		return nil
	}); err != nil {
		return fmt.Errorf("error creating or updating service [%w]", err)
	}
	helper.GetState().ServiceState = "ready"
	return nil
}

func (r *SourceReconciler) ObserveSourceHPA(
	ctx context.Context, source *computeapi.Source, helper ReconciliationHelper) error {
	if source.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
		helper.GetState().HPAState = "delete"
		return nil
	}

	hpa := &autov2beta2.HorizontalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: source.Namespace,
		Name:      spec.MakeSourceObjectMeta(source).Name}, hpa)
	if err != nil {
		helper.GetState().HPAState = "unready"
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error fetching hpa [%w]", err)
	}

	helper.GetState().HPAState = "created"
	if !r.isGenerationChanged(source) {
		helper.GetState().HPAState = "ready"
	}
	return nil
}

func (r *SourceReconciler) ApplySourceHPA(
	ctx context.Context, source *computeapi.Source, helper ReconciliationHelper) error {
	if source.Spec.MaxReplicas == nil {
		// HPA not enabled, clear the exists HPA
		hpa := &autov2beta2.HorizontalPodAutoscaler{}
		hpa.Namespace = source.Namespace
		hpa.Name = spec.MakeSourceObjectMeta(source).Name
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
	desiredHPA := spec.MakeSourceHPA(source)
	desiredHPASpec := desiredHPA.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredHPA, func() error {
		// source hpa mutate logic
		desiredHPA.Spec = desiredHPASpec
		return nil
	}); err != nil {
		return fmt.Errorf("error creating or updating hpa [%w]", err)
	}
	helper.GetState().HPAState = "ready"
	return nil
}

func (r *SourceReconciler) ObserveSourceVPA(
	ctx context.Context, source *computeapi.Source, helper ReconciliationHelper) error {
	if source.Spec.Pod.VPA == nil {
		// VPA not enabled, skip further action
		helper.GetState().VPAState = "delete"
		return nil
	}

	vpa := &vpav1.VerticalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: source.Namespace,
		Name:      spec.MakeSourceObjectMeta(source).Name}, vpa)
	if err != nil {
		helper.GetState().VPAState = "unready"
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error fetching vpa [%w]", err)
	}

	helper.GetState().VPAState = "created"
	if !r.isGenerationChanged(source) {
		helper.GetState().VPAState = "ready"
	}
	return nil
}

func (r *SourceReconciler) ApplySourceVPA(
	ctx context.Context, source *computeapi.Source, helper ReconciliationHelper) error {
	if source.Spec.Pod.VPA == nil {
		// VPA not enabled, clear the exists VPA
		vpa := &vpav1.VerticalPodAutoscaler{}
		vpa.Namespace = source.Namespace
		vpa.Name = spec.MakeSourceObjectMeta(source).Name
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
	desiredVPA := spec.MakeSourceVPA(source)
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

func (r *SourceReconciler) isGenerationChanged(source *computeapi.Source) bool {
	// if the generation has not changed, we do not need to update the component
	return source.Generation != source.Status.ObservedGeneration
}
