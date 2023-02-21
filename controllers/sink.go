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

func (r *SinkReconciler) ObserveSinkStatefulSet(
	ctx context.Context, sink *computeapi.Sink, helper ReconciliationHelper) error {
	statefulSet := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: sink.Namespace,
		Name:      spec.MakeSinkObjectMeta(sink).Name,
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
	sink.Status.Selector = selector.String()

	helper.GetState().StatefulSetState = "created"
	if !r.isGenerationChanged(sink) && statefulSet.Status.ReadyReplicas == *sink.Spec.Replicas {
		helper.GetState().StatefulSetState = "ready"
	}
	sink.Status.Replicas = *statefulSet.Spec.Replicas
	return nil
}

func (r *SinkReconciler) ApplySinkStatefulSet(
	ctx context.Context, sink *computeapi.Sink, helper ReconciliationHelper) error {
	if helper.GetState().StatefulSetState == "ready" {
		return nil
	}
	desiredStatefulSet := spec.MakeSinkStatefulSet(sink)
	desiredStatefulSetSpec := desiredStatefulSet.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredStatefulSet, func() error {
		// sink statefulSet mutate logic
		desiredStatefulSet.Spec = desiredStatefulSetSpec
		return nil
	}); err != nil {
		return fmt.Errorf("error creating or updating statefulSet [%w]", err)
	}
	helper.GetState().StatefulSetState = "created"
	return nil
}

func (r *SinkReconciler) ObserveSinkService(
	ctx context.Context, sink *computeapi.Sink, helper ReconciliationHelper) error {
	svc := &corev1.Service{}
	svcName := spec.MakeHeadlessServiceName(spec.MakeSinkObjectMeta(sink).Name)
	err := r.Get(ctx, types.NamespacedName{Namespace: sink.Namespace,
		Name: svcName}, svc)
	if err != nil {
		helper.GetState().ServiceState = "unready"
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error fetching service [%w]", err)
	}

	helper.GetState().ServiceState = "created"
	if !r.isGenerationChanged(sink) {
		helper.GetState().ServiceState = "ready"
	}
	return nil
}

func (r *SinkReconciler) ApplySinkService(
	ctx context.Context, sink *computeapi.Sink, helper ReconciliationHelper) error {
	if helper.GetState().ServiceState == "ready" {
		return nil
	}
	desiredService := spec.MakeSinkService(sink)
	desiredServiceSpec := desiredService.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredService, func() error {
		// sink service mutate logic
		desiredService.Spec = desiredServiceSpec
		return nil
	}); err != nil {
		return fmt.Errorf("error creating or updating service [%w]", err)
	}
	helper.GetState().ServiceState = "ready"
	return nil
}

func (r *SinkReconciler) ObserveSinkHPA(
	ctx context.Context, sink *computeapi.Sink, helper ReconciliationHelper) error {
	if sink.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
		helper.GetState().HPAState = "delete"
		return nil
	}

	hpa := &autov2beta2.HorizontalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: sink.Namespace,
		Name:      spec.MakeSinkObjectMeta(sink).Name}, hpa)
	if err != nil {
		helper.GetState().HPAState = "unready"
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error fetching hpa [%w]", err)
	}

	helper.GetState().HPAState = "created"
	if !r.isGenerationChanged(sink) {
		helper.GetState().HPAState = "ready"
	}
	return nil
}

func (r *SinkReconciler) ApplySinkHPA(
	ctx context.Context, sink *computeapi.Sink, helper ReconciliationHelper) error {
	if sink.Spec.MaxReplicas == nil {
		// HPA not enabled, clear the exists HPA
		hpa := &autov2beta2.HorizontalPodAutoscaler{}
		hpa.Namespace = sink.Namespace
		hpa.Name = spec.MakeSinkObjectMeta(sink).Name
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
	desiredHPA := spec.MakeSinkHPA(sink)
	desiredHPASpec := desiredHPA.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredHPA, func() error {
		// sink hpa mutate logic
		desiredHPA.Spec = desiredHPASpec
		return nil
	}); err != nil {
		return fmt.Errorf("error creating or updating hpa [%w]", err)
	}
	helper.GetState().HPAState = "ready"
	return nil
}

func (r *SinkReconciler) ObserveSinkVPA(
	ctx context.Context, sink *computeapi.Sink, helper ReconciliationHelper) error {
	if sink.Spec.Pod.VPA == nil {
		// VPA not enabled, skip further action
		helper.GetState().VPAState = "delete"
		return nil
	}

	vpa := &vpav1.VerticalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: sink.Namespace,
		Name:      spec.MakeSinkObjectMeta(sink).Name}, vpa)
	if err != nil {
		helper.GetState().VPAState = "unready"
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error fetching vpa [%w]", err)
	}

	helper.GetState().VPAState = "created"
	if !r.isGenerationChanged(sink) {
		helper.GetState().VPAState = "ready"
	}
	return nil
}

func (r *SinkReconciler) ApplySinkVPA(
	ctx context.Context, sink *computeapi.Sink, helper ReconciliationHelper) error {
	if sink.Spec.Pod.VPA == nil {
		// VPA not enabled, clear the exists VPA
		vpa := &vpav1.VerticalPodAutoscaler{}
		vpa.Namespace = sink.Namespace
		vpa.Name = spec.MakeSinkObjectMeta(sink).Name
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
	desiredVPA := spec.MakeSinkVPA(sink)
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

func (r *SinkReconciler) isGenerationChanged(sink *computeapi.Sink) bool {
	// if the generation has not changed, we do not need to update the component
	return sink.Generation != sink.Status.ObservedGeneration
}
