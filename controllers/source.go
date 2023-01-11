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

func (r *SourceReconciler) ObserveSourceStatefulSet(ctx context.Context, source *v1alpha1.Source) error {
	statefulSet := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: source.Namespace,
		Name:      spec.MakeSourceObjectMeta(source).Name,
	}, statefulSet)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("source statefulset is not ready yet...",
				"namespace", source.Namespace, "name", source.Name,
				"statefulset name", statefulSet.Name)
			source.SetCondition(v1alpha1.StatefulSet, v1alpha1.CreateCondition(
				v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
				"source statefulset is not ready yet..."))
			return nil
		}
		source.SetCondition(v1alpha1.StatefulSet, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.StatefulSetError,
			fmt.Sprintf("failed to fetch source statefulset: %v", err)))
		return err
	}

	selector, err := metav1.LabelSelectorAsSelector(statefulSet.Spec.Selector)
	if err != nil {
		r.Log.Error(err, "error retrieving statefulset selector")
		source.SetCondition(v1alpha1.StatefulSet, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.StatefulSetError,
			fmt.Sprintf("error retrieving statefulset selector: %v", err)))
		return err
	}
	source.Status.Selector = selector.String()
	source.Status.Replicas = *source.Spec.Replicas

	if r.checkIfStatefulSetNeedUpdate(statefulSet, source) {
		source.SetCondition(v1alpha1.StatefulSet, v1alpha1.CreateCondition(
			v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
			"need to update source statefulset..."))
		return nil
	}

	if statefulSet.Status.ReadyReplicas == *source.Spec.Replicas {
		source.SetCondition(v1alpha1.StatefulSet, v1alpha1.CreateCondition(
			v1alpha1.Ready, metav1.ConditionTrue, v1alpha1.StatefulSetIsReady,
			""))
		return nil
	}
	source.SetCondition(v1alpha1.StatefulSet, v1alpha1.CreateCondition(
		v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
		"source statefulset is pending creation..."))
	return nil
}

func (r *SourceReconciler) ApplySourceStatefulSet(ctx context.Context, source *v1alpha1.Source) error {
	condition := source.Status.Conditions[v1alpha1.StatefulSet]
	if condition.Status == metav1.ConditionTrue &&
		source.Generation == source.Status.ObservedGeneration {
		return nil
	}

	desiredStatefulSet := spec.MakeSourceStatefulSet(source)
	desiredStatefulSetSpec := desiredStatefulSet.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredStatefulSet, func() error {
		// source statefulSet mutate logic
		desiredStatefulSet.Spec = desiredStatefulSetSpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error create or update statefulset workload for source",
			"namespace", source.Namespace, "name", source.Name,
			"statefulset name", desiredStatefulSet.Name)
		source.SetCondition(v1alpha1.StatefulSet, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.ErrorCreatingStatefulSet,
			fmt.Sprintf("error create or update statefulset workload for source: %v", err)))
		return err
	}
	source.SetCondition(v1alpha1.StatefulSet, v1alpha1.CreateCondition(
		v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
		"creating or updating statefulset workload for source..."))
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
			source.SetCondition(v1alpha1.Service, v1alpha1.CreateCondition(
				v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
				"source service is not created..."))
			return nil
		}
		source.SetCondition(v1alpha1.Service, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.ServiceError,
			fmt.Sprintf("failed to fetch source service: %v", err)))
		return err
	}

	source.SetCondition(v1alpha1.Service, v1alpha1.CreateCondition(
		v1alpha1.Ready, metav1.ConditionTrue, v1alpha1.ServiceIsReady,
		""))
	return nil
}

func (r *SourceReconciler) ApplySourceService(ctx context.Context, source *v1alpha1.Source) error {
	condition := source.Status.Conditions[v1alpha1.Service]
	if condition.Status == metav1.ConditionTrue &&
		source.Generation == source.Status.ObservedGeneration {
		return nil
	}

	desiredService := spec.MakeSourceService(source)
	desiredServiceSpec := desiredService.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredService, func() error {
		// source service mutate logic
		desiredService.Spec = desiredServiceSpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error create or update service for source",
			"namespace", source.Namespace, "name", source.Name,
			"service name", desiredService.Name)
		source.SetCondition(v1alpha1.Service, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.ErrorCreatingService,
			fmt.Sprintf("error create or update service for source: %v", err)))
		return err
	}
	source.SetCondition(v1alpha1.Service, v1alpha1.CreateCondition(
		v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
		"creating or updating service for source..."))
	return nil
}

func (r *SourceReconciler) ObserveSourceHPA(ctx context.Context, source *v1alpha1.Source) error {
	if source.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
		delete(source.Status.Conditions, v1alpha1.HPA)
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
			source.SetCondition(v1alpha1.HPA, v1alpha1.CreateCondition(
				v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
				"hpa is not created for source..."))
			return nil
		}
		source.SetCondition(v1alpha1.HPA, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.HPAError,
			fmt.Sprintf("failed to fetch source hpa: %v", err)))
		return err
	}

	if r.checkIfHPANeedUpdate(hpa, source) {
		source.SetCondition(v1alpha1.HPA, v1alpha1.CreateCondition(
			v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
			"need to update source hpa..."))
		return nil
	}

	source.SetCondition(v1alpha1.HPA, v1alpha1.CreateCondition(
		v1alpha1.Ready, metav1.ConditionTrue, v1alpha1.HPAIsReady,
		""))
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
			source.SetCondition(v1alpha1.HPA, v1alpha1.CreateCondition(
				v1alpha1.Error, metav1.ConditionFalse, v1alpha1.HPAError,
				fmt.Sprintf("failed to delete source hpa: %v", err)))
			return err
		}
		return nil
	}

	condition := source.Status.Conditions[v1alpha1.HPA]
	if condition.Status == metav1.ConditionTrue &&
		source.Generation == source.Status.ObservedGeneration {
		return nil
	}

	desiredHPA := spec.MakeSourceHPA(source)
	desiredHPASpec := desiredHPA.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredHPA, func() error {
		// source hpa mutate logic
		desiredHPA.Spec = desiredHPASpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error create or update hpa for source",
			"namespace", source.Namespace, "name", source.Name,
			"hpa name", desiredHPA.Name)
		source.SetCondition(v1alpha1.HPA, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.ErrorCreatingHPA,
			fmt.Sprintf("error create or update hpa for source: %v", err)))
		return err
	}
	source.SetCondition(v1alpha1.HPA, v1alpha1.CreateCondition(
		v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
		"creating or updating hpa for source..."))
	return nil
}

func (r *SourceReconciler) ObserveSourceVPA(ctx context.Context, source *v1alpha1.Source) error {
	if source.Spec.Pod.VPA == nil {
		delete(source.Status.Conditions, v1alpha1.VPA)
		return nil
	}

	vpa := &vpav1.VerticalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: source.Namespace,
		Name:      spec.MakeSourceObjectMeta(source).Name,
	}, vpa)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("vpa is not created for source...",
				"namespace", source.Namespace, "name", source.Name,
				"vpa name", vpa.Name)
			source.SetCondition(v1alpha1.VPA, v1alpha1.CreateCondition(
				v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
				"vpa is not created for source..."))
			return nil
		}
		source.SetCondition(v1alpha1.VPA, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.VPAError,
			fmt.Sprintf("failed to fetch source vpa: %v", err)))
		return err
	}

	if r.checkIfVPANeedUpdate(vpa, source) {
		source.SetCondition(v1alpha1.VPA, v1alpha1.CreateCondition(
			v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
			"need to update source vpa..."))
		return nil
	}

	source.SetCondition(v1alpha1.VPA, v1alpha1.CreateCondition(
		v1alpha1.Ready, metav1.ConditionTrue, v1alpha1.VPAIsReady,
		""))
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
			source.SetCondition(v1alpha1.VPA, v1alpha1.CreateCondition(
				v1alpha1.Error, metav1.ConditionFalse, v1alpha1.VPAError,
				fmt.Sprintf("failed to delete source vpa: %v", err)))
			return err
		}
		return nil
	}

	condition := source.Status.Conditions[v1alpha1.VPA]
	if condition.Status == metav1.ConditionTrue &&
		source.Generation == source.Status.ObservedGeneration {
		return nil
	}

	desiredVPA := spec.MakeSourceVPA(source)
	desiredVPASpec := desiredVPA.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredVPA, func() error {
		// source vpa mutate logic
		desiredVPA.Spec = desiredVPASpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error create or update vpa for source",
			"namespace", source.Namespace, "name", source.Name,
			"vpa name", desiredVPA.Name)
		source.SetCondition(v1alpha1.VPA, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.ErrorCreatingVPA,
			fmt.Sprintf("error create or update vpa for source: %v", err)))
		return err
	}
	source.SetCondition(v1alpha1.VPA, v1alpha1.CreateCondition(
		v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
		"creating or updating vpa for source..."))
	return nil
}

func (r *SourceReconciler) checkIfStatefulSetNeedUpdate(statefulSet *appsv1.StatefulSet, source *v1alpha1.Source) bool {
	return !spec.CheckIfStatefulSetSpecIsEqual(&statefulSet.Spec, &spec.MakeSourceStatefulSet(source).Spec)
}

func (r *SourceReconciler) checkIfHPANeedUpdate(hpa *autov2beta2.HorizontalPodAutoscaler, source *v1alpha1.Source) bool {
	return !spec.CheckIfHPASpecIsEqual(&hpa.Spec, &spec.MakeSourceHPA(source).Spec)
}

func (r *SourceReconciler) checkIfVPANeedUpdate(vpa *vpav1.VerticalPodAutoscaler, source *v1alpha1.Source) bool {
	return !spec.CheckIfVPASpecIsEqual(&vpa.Spec, &spec.MakeSourceVPA(source).Spec)
}
