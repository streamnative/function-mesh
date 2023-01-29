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

	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"

	"k8s.io/apimachinery/pkg/api/meta"

	appsv1 "k8s.io/api/apps/v1"
	autov2beta2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/spec"
)

func (r *SourceReconciler) ObserveSourceStatefulSet(ctx context.Context, source *v1alpha1.Source) error {
	statefulSet := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: source.Namespace,
		Name:      spec.MakeSourceObjectMeta(source).Name,
	}, statefulSet)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("source statefulSet is not ready yet...",
				"namespace", source.Namespace, "name", source.Name,
				"statefulSet name", statefulSet.Name)
			return nil
		}
		source.SetCondition(v1alpha1.Error, metav1.ConditionTrue, v1alpha1.StatefulSetError,
			fmt.Sprintf("error fetching source statefulSet: %v", err))
		return err
	}

	selector, err := metav1.LabelSelectorAsSelector(statefulSet.Spec.Selector)
	if err != nil {
		r.Log.Error(err, "error retrieving statefulSet selector")
		source.SetCondition(v1alpha1.Error, metav1.ConditionTrue, v1alpha1.StatefulSetError,
			fmt.Sprintf("error retrieving statefulSet selector: %v", err))
		return err
	}
	source.Status.Selector = selector.String()

	if statefulSet.Status.ReadyReplicas == *source.Spec.Replicas {
		source.SetCondition(v1alpha1.StatefulSetReady, metav1.ConditionTrue, v1alpha1.StatefulSetIsReady, "")
	} else {
		source.SetCondition(v1alpha1.StatefulSetReady, metav1.ConditionFalse, v1alpha1.PendingCreation,
			"wait for the number of replicas of statefulSet to be ready")
	}
	source.Status.Replicas = *statefulSet.Spec.Replicas
	return nil
}

func (r *SourceReconciler) ApplySourceStatefulSet(ctx context.Context, source *v1alpha1.Source) error {
	if !r.checkIfStatefulSetNeedUpdate(source) {
		return nil
	}
	desiredStatefulSet := spec.MakeSourceStatefulSet(source)
	desiredStatefulSetSpec := desiredStatefulSet.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredStatefulSet, func() error {
		// source statefulSet mutate logic
		desiredStatefulSet.Spec = desiredStatefulSetSpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error creating or updating statefulSet workload for source",
			"namespace", source.Namespace, "name", source.Name,
			"statefulSet name", desiredStatefulSet.Name)
		source.SetCondition(v1alpha1.StatefulSetReady, metav1.ConditionFalse, v1alpha1.ErrorCreatingStatefulSet,
			fmt.Sprintf("error creating or updating statefulSet for source: %v", err))
		return err
	}
	source.SetCondition(v1alpha1.StatefulSetReady, metav1.ConditionFalse, v1alpha1.PendingCreation,
		"creating or updating statefulSet for source...")
	return nil
}

func (r *SourceReconciler) ObserveSourceService(ctx context.Context, source *v1alpha1.Source) error {
	svc := &corev1.Service{}
	svcName := spec.MakeHeadlessServiceName(spec.MakeSourceObjectMeta(source).Name)
	err := r.Get(ctx, types.NamespacedName{Namespace: source.Namespace,
		Name: svcName}, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("source service is not created...",
				"namespace", source.Namespace, "name", source.Name,
				"service name", svcName)
			return nil
		}
		source.SetCondition(v1alpha1.Error, metav1.ConditionTrue, v1alpha1.ServiceError,
			fmt.Sprintf("error fetching source service: %v", err))
		return err
	}
	source.SetCondition(v1alpha1.ServiceReady, metav1.ConditionTrue, v1alpha1.ServiceIsReady, "")
	return nil
}

func (r *SourceReconciler) ApplySourceService(ctx context.Context, source *v1alpha1.Source) error {
	if !r.checkIfServiceNeedUpdate(source) {
		return nil
	}
	desiredService := spec.MakeSourceService(source)
	desiredServiceSpec := desiredService.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredService, func() error {
		// source service mutate logic
		desiredService.Spec = desiredServiceSpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error creating or updating service for source",
			"namespace", source.Namespace, "name", source.Name,
			"service name", desiredService.Name)
		source.SetCondition(v1alpha1.ServiceReady, metav1.ConditionFalse, v1alpha1.ErrorCreatingService,
			fmt.Sprintf("error creating or updating service for source: %v", err))
		return err
	}
	source.SetCondition(v1alpha1.ServiceReady, metav1.ConditionTrue, v1alpha1.ServiceIsReady, "")
	return nil
}

func (r *SourceReconciler) ObserveSourceHPA(ctx context.Context, source *v1alpha1.Source) error {
	if source.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
		source.RemoveCondition(v1alpha1.HPAReady)
		return nil
	}

	hpa := &autov2beta2.HorizontalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{Namespace: source.Namespace,
		Name: spec.MakeSourceObjectMeta(source).Name}, hpa)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("hpa is not created for source...",
				"namespace", source.Namespace, "name", source.Name,
				"hpa name", hpa.Name)
			return nil
		}
		source.SetCondition(v1alpha1.Error, metav1.ConditionTrue, v1alpha1.HPAError,
			fmt.Sprintf("error fetching source hpa: %v", err))
		return err
	}

	source.SetCondition(v1alpha1.HPAReady, metav1.ConditionTrue, v1alpha1.HPAIsReady, "")
	return nil
}

func (r *SourceReconciler) ApplySourceHPA(ctx context.Context, source *v1alpha1.Source) error {
	if source.Spec.MaxReplicas == nil {
		// HPA not enabled, clear the exists HPA
		hpa := &autov2beta2.HorizontalPodAutoscaler{}
		hpa.Namespace = source.Namespace
		hpa.Name = spec.MakeSourceObjectMeta(source).Name
		if err := r.Delete(ctx, hpa); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			source.SetCondition(v1alpha1.Error, metav1.ConditionTrue, v1alpha1.HPAError,
				fmt.Sprintf("error deleting hpa for source: %v", err))
			return err
		}
		return nil
	}
	if !r.checkIfHPANeedUpdate(source) {
		return nil
	}
	desiredHPA := spec.MakeSourceHPA(source)
	desiredHPASpec := desiredHPA.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredHPA, func() error {
		// source hpa mutate logic
		desiredHPA.Spec = desiredHPASpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error creating or updating hpa for source",
			"namespace", source.Namespace, "name", source.Name,
			"hpa name", desiredHPA.Name)
		source.SetCondition(v1alpha1.HPAReady, metav1.ConditionFalse, v1alpha1.ErrorCreatingHPA,
			fmt.Sprintf("error creating or updating hpa for source: %v", err))
		return err
	}
	source.SetCondition(v1alpha1.HPAReady, metav1.ConditionTrue, v1alpha1.HPAIsReady, "")
	return nil
}

func (r *SourceReconciler) ObserveSourceVPA(ctx context.Context, source *v1alpha1.Source) error {
	if source.Spec.Pod.VPA == nil {
		// VPA not enabled, skip further action
		source.RemoveCondition(v1alpha1.VPAReady)
		return nil
	}

	vpa := &vpav1.VerticalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: source.Namespace,
		Name:      spec.MakeSourceObjectMeta(source).Name}, vpa)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("vpa is not created for source...",
				"namespace", source.Namespace, "name", source.Name,
				"vpa name", vpa.Name)
			return nil
		}
		source.SetCondition(v1alpha1.Error, metav1.ConditionTrue, v1alpha1.VPAError,
			fmt.Sprintf("error fetching source vpa: %v", err))
		return err
	}

	source.SetCondition(v1alpha1.VPAReady, metav1.ConditionTrue, v1alpha1.VPAIsReady, "")
	return nil
}

func (r *SourceReconciler) ApplySourceVPA(ctx context.Context, source *v1alpha1.Source) error {
	if source.Spec.Pod.VPA == nil {
		// VPA not enabled, clear the exists VPA
		vpa := &vpav1.VerticalPodAutoscaler{}
		vpa.Namespace = source.Namespace
		vpa.Name = spec.MakeSourceObjectMeta(source).Name
		if err := r.Delete(ctx, vpa); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			source.SetCondition(v1alpha1.Error, metav1.ConditionTrue, v1alpha1.VPAError,
				fmt.Sprintf("error deleting vpa for source: %v", err))
			return err
		}
		return nil
	}

	if !r.checkIfVPANeedUpdate(source) {
		return nil
	}
	desiredVPA := spec.MakeSourceVPA(source)
	desiredVPASpec := desiredVPA.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredVPA, func() error {
		// function vpa mutate logic
		desiredVPA.Spec = desiredVPASpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error creating or updating vpa for source",
			"namespace", source.Namespace, "name", source.Name,
			"vpa name", desiredVPA.Name)
		source.SetCondition(v1alpha1.VPAReady, metav1.ConditionFalse, v1alpha1.ErrorCreatingVPA,
			fmt.Sprintf("error creating or updating vpa for source: %v", err))
		return err
	}
	source.SetCondition(v1alpha1.VPAReady, metav1.ConditionTrue, v1alpha1.VPAIsReady, "")
	return nil
}

func (r *SourceReconciler) checkIfStatefulSetNeedUpdate(source *v1alpha1.Source) bool {
	if statefulSetStatus := meta.FindStatusCondition(source.Status.Conditions, string(v1alpha1.StatefulSetReady)); statefulSetStatus != nil {
		if statefulSetStatus.ObservedGeneration != source.Generation {
			return true
		}
		if statefulSetStatus.Status == metav1.ConditionTrue {
			return false
		}
	}
	return true
}

func (r *SourceReconciler) checkIfServiceNeedUpdate(source *v1alpha1.Source) bool {
	if serviceStatus := meta.FindStatusCondition(source.Status.Conditions, string(v1alpha1.ServiceReady)); serviceStatus != nil {
		if serviceStatus.ObservedGeneration != source.Generation {
			return true
		}
		if serviceStatus.Status == metav1.ConditionTrue {
			return false
		}
	}
	return true
}

func (r *SourceReconciler) checkIfHPANeedUpdate(source *v1alpha1.Source) bool {
	if hpaStatus := meta.FindStatusCondition(source.Status.Conditions, string(v1alpha1.HPAReady)); hpaStatus != nil {
		if hpaStatus.ObservedGeneration != source.Generation {
			return true
		}
		if hpaStatus.Status == metav1.ConditionTrue {
			return false
		}
	}
	return true
}

func (r *SourceReconciler) checkIfVPANeedUpdate(source *v1alpha1.Source) bool {
	if vpaStatus := meta.FindStatusCondition(source.Status.Conditions, string(v1alpha1.VPAReady)); vpaStatus != nil {
		if vpaStatus.ObservedGeneration != source.Generation {
			return true
		}
		if vpaStatus.Status == metav1.ConditionTrue {
			return false
		}
	}
	return true
}
