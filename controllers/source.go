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
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	computeapi "github.com/streamnative/function-mesh/api/compute/v1alpha2"
	"github.com/streamnative/function-mesh/controllers/spec"
	apispec "github.com/streamnative/function-mesh/pkg/spec"
)

func (r *SourceReconciler) ObserveSourceStatefulSet(ctx context.Context, source *computeapi.Source) error {
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
			source.SetCondition(apispec.StatefulSetReady, metav1.ConditionFalse, apispec.PendingCreation,
				"source statefulSet is not ready yet...")
			return nil
		}
		source.SetCondition(apispec.Error, metav1.ConditionTrue, apispec.StatefulSetError,
			fmt.Sprintf("error fetching source statefulSet: %v", err))
		return err
	}

	selector, err := metav1.LabelSelectorAsSelector(statefulSet.Spec.Selector)
	if err != nil {
		r.Log.Error(err, "error retrieving statefulSet selector")
		source.SetCondition(apispec.Error, metav1.ConditionTrue, apispec.StatefulSetError,
			fmt.Sprintf("error retrieving statefulSet selector: %v", err))
		return err
	}
	source.Status.Selector = selector.String()

	if r.checkIfStatefulSetNeedToUpdate(source, statefulSet) {
		source.SetCondition(apispec.StatefulSetReady, metav1.ConditionFalse, apispec.PendingCreation,
			"wait for the source statefulSet to be ready")
	} else {
		source.SetCondition(apispec.StatefulSetReady, metav1.ConditionTrue, apispec.StatefulSetIsReady, "")
	}
	source.Status.Replicas = *statefulSet.Spec.Replicas
	return nil
}

func (r *SourceReconciler) ApplySourceStatefulSet(ctx context.Context, source *computeapi.Source) error {
	if meta.IsStatusConditionTrue(source.Status.Conditions, string(apispec.StatefulSetReady)) {
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
		source.SetCondition(apispec.StatefulSetReady, metav1.ConditionFalse, apispec.ErrorCreatingStatefulSet,
			fmt.Sprintf("error creating or updating statefulSet for source: %v", err))
		return err
	}
	source.SetCondition(apispec.StatefulSetReady, metav1.ConditionFalse, apispec.PendingCreation,
		"creating or updating statefulSet for source...")
	return nil
}

func (r *SourceReconciler) ObserveSourceService(ctx context.Context, source *computeapi.Source) error {
	svc := &corev1.Service{}
	svcName := spec.MakeHeadlessServiceName(spec.MakeSourceObjectMeta(source).Name)
	err := r.Get(ctx, types.NamespacedName{Namespace: source.Namespace,
		Name: svcName}, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("source service is not created...",
				"namespace", source.Namespace, "name", source.Name,
				"service name", svcName)
			source.SetCondition(apispec.ServiceReady, metav1.ConditionFalse, apispec.PendingCreation,
				"source service is not created...")
			return nil
		}
		source.SetCondition(apispec.Error, metav1.ConditionTrue, apispec.ServiceError,
			fmt.Sprintf("error fetching source service: %v", err))
		return err
	}
	if r.checkIfServiceNeedToUpdate(source) {
		source.SetCondition(apispec.ServiceReady, metav1.ConditionFalse, apispec.PendingCreation,
			"wait for the source service to be ready")
	} else {
		source.SetCondition(apispec.ServiceReady, metav1.ConditionTrue, apispec.ServiceIsReady, "")
	}
	return nil
}

func (r *SourceReconciler) ApplySourceService(ctx context.Context, source *computeapi.Source) error {
	if meta.IsStatusConditionTrue(source.Status.Conditions, string(apispec.ServiceReady)) {
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
		source.SetCondition(apispec.ServiceReady, metav1.ConditionFalse, apispec.ErrorCreatingService,
			fmt.Sprintf("error creating or updating service for source: %v", err))
		return err
	}
	source.SetCondition(apispec.ServiceReady, metav1.ConditionTrue, apispec.ServiceIsReady, "")
	return nil
}

func (r *SourceReconciler) ObserveSourceHPA(ctx context.Context, source *computeapi.Source) error {
	if source.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
		source.RemoveCondition(apispec.HPAReady)
		return nil
	}

	hpa := &autov2beta2.HorizontalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: source.Namespace,
		Name:      spec.MakeSourceObjectMeta(source).Name}, hpa)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("source hpa is not created...",
				"namespace", source.Namespace, "name", source.Name,
				"hpa name", hpa.Name)
			source.SetCondition(apispec.HPAReady, metav1.ConditionFalse, apispec.PendingCreation,
				"source hpa is not created...")
			return nil
		}
		source.SetCondition(apispec.Error, metav1.ConditionTrue, apispec.HPAError,
			fmt.Sprintf("error fetching source hpa: %v", err))
		return err
	}
	if r.checkIfHPANeedUpdate(source) {
		source.SetCondition(apispec.HPAReady, metav1.ConditionFalse, apispec.PendingCreation,
			"wait for the source hpa to be ready")
	} else {
		source.SetCondition(apispec.HPAReady, metav1.ConditionTrue, apispec.HPAIsReady, "")
	}
	return nil
}

func (r *SourceReconciler) ApplySourceHPA(ctx context.Context, source *computeapi.Source) error {
	if source.Spec.MaxReplicas == nil {
		// HPA not enabled, clear the exists HPA
		hpa := &autov2beta2.HorizontalPodAutoscaler{}
		hpa.Namespace = source.Namespace
		hpa.Name = spec.MakeSourceObjectMeta(source).Name
		if err := r.Delete(ctx, hpa); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			source.SetCondition(apispec.Error, metav1.ConditionTrue, apispec.HPAError,
				fmt.Sprintf("error deleting hpa for source: %v", err))
			return err
		}
		return nil
	}
	if meta.IsStatusConditionTrue(source.Status.Conditions, string(apispec.HPAReady)) {
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
		source.SetCondition(apispec.HPAReady, metav1.ConditionFalse, apispec.ErrorCreatingHPA,
			fmt.Sprintf("error creating or updating hpa for source: %v", err))
		return err
	}
	source.SetCondition(apispec.HPAReady, metav1.ConditionTrue, apispec.HPAIsReady, "")
	return nil
}

func (r *SourceReconciler) ObserveSourceVPA(ctx context.Context, source *computeapi.Source) error {
	if source.Spec.Pod.VPA == nil {
		// VPA not enabled, skip further action
		source.RemoveCondition(apispec.VPAReady)
		return nil
	}

	vpa := &vpav1.VerticalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: source.Namespace,
		Name:      spec.MakeSourceObjectMeta(source).Name}, vpa)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("source vpa is not created...",
				"namespace", source.Namespace, "name", source.Name,
				"vpa name", vpa.Name)
			source.SetCondition(apispec.VPAReady, metav1.ConditionFalse, apispec.PendingCreation,
				"source vpa is not created...")
			return nil
		}
		source.SetCondition(apispec.Error, metav1.ConditionTrue, apispec.VPAError,
			fmt.Sprintf("error fetching source vpa: %v", err))
		return err
	}
	if r.checkIfVPANeedUpdate(source) {
		source.SetCondition(apispec.VPAReady, metav1.ConditionFalse, apispec.PendingCreation,
			"wait for the source vpa to be ready")
	} else {
		source.SetCondition(apispec.VPAReady, metav1.ConditionTrue, apispec.VPAIsReady, "")
	}
	return nil
}

func (r *SourceReconciler) ApplySourceVPA(ctx context.Context, source *computeapi.Source) error {
	if source.Spec.Pod.VPA == nil {
		// VPA not enabled, clear the exists VPA
		vpa := &vpav1.VerticalPodAutoscaler{}
		vpa.Namespace = source.Namespace
		vpa.Name = spec.MakeSourceObjectMeta(source).Name
		if err := r.Delete(ctx, vpa); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			source.SetCondition(apispec.Error, metav1.ConditionTrue, apispec.VPAError,
				fmt.Sprintf("error deleting vpa for source: %v", err))
			return err
		}
		return nil
	}
	if meta.IsStatusConditionTrue(source.Status.Conditions, string(apispec.VPAReady)) {
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
		source.SetCondition(apispec.VPAReady, metav1.ConditionFalse, apispec.ErrorCreatingVPA,
			fmt.Sprintf("error creating or updating vpa for source: %v", err))
		return err
	}
	source.SetCondition(apispec.VPAReady, metav1.ConditionTrue, apispec.VPAIsReady, "")
	return nil
}

func (r *SourceReconciler) checkIfStatefulSetNeedToUpdate(source *computeapi.Source, statefulSet *appsv1.StatefulSet) bool {
	return r.checkIfComponentNeedToUpdate(source, apispec.StatefulSetReady) ||
		statefulSet.Status.ReadyReplicas != *source.Spec.Replicas
}

func (r *SourceReconciler) checkIfServiceNeedToUpdate(source *computeapi.Source) bool {
	return r.checkIfComponentNeedToUpdate(source, apispec.ServiceReady)
}

func (r *SourceReconciler) checkIfHPANeedUpdate(source *computeapi.Source) bool {
	return r.checkIfComponentNeedToUpdate(source, apispec.HPAReady)
}

func (r *SourceReconciler) checkIfVPANeedUpdate(source *computeapi.Source) bool {
	return r.checkIfComponentNeedToUpdate(source, apispec.VPAReady)
}

func (r *SourceReconciler) checkIfComponentNeedToUpdate(source *computeapi.Source, condType apispec.ResourceConditionType) bool {
	if cond := meta.FindStatusCondition(source.Status.Conditions, string(condType)); cond != nil {
		// if the generation has not changed, we do not need to update the component
		if cond.ObservedGeneration == source.Generation {
			return false
		}
	}
	return true
}
