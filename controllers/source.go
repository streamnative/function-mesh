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
	"encoding/json"
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
			source.SetCondition(v1alpha1.StatefulSetReady, metav1.ConditionFalse, v1alpha1.PendingCreation,
				"source statefulSet is not ready yet...")
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

	if r.checkIfStatefulSetNeedToUpdate(source, statefulSet) {
		source.SetCondition(v1alpha1.StatefulSetReady, metav1.ConditionFalse, v1alpha1.PendingCreation,
			"wait for the source statefulSet to be ready")
	} else {
		source.SetCondition(v1alpha1.StatefulSetReady, metav1.ConditionTrue, v1alpha1.StatefulSetIsReady, "")
	}
	source.Status.Replicas = *statefulSet.Spec.Replicas
	return nil
}

func (r *SourceReconciler) ApplySourceStatefulSet(ctx context.Context, source *v1alpha1.Source) error {
	if meta.IsStatusConditionTrue(source.Status.Conditions, string(v1alpha1.StatefulSetReady)) {
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
	desiredStatefulSetSpecBytes, _ := json.Marshal(desiredStatefulSetSpec)
	source.SetComponentHash(v1alpha1.StatefulSet, spec.GenerateSpecHash(desiredStatefulSetSpecBytes))
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
			source.SetCondition(v1alpha1.ServiceReady, metav1.ConditionFalse, v1alpha1.PendingCreation,
				"source service is not created...")
			return nil
		}
		source.SetCondition(v1alpha1.Error, metav1.ConditionTrue, v1alpha1.ServiceError,
			fmt.Sprintf("error fetching source service: %v", err))
		return err
	}
	if r.checkIfServiceNeedToUpdate(source) {
		source.SetCondition(v1alpha1.ServiceReady, metav1.ConditionFalse, v1alpha1.PendingCreation,
			"wait for the source service to be ready")
	} else {
		source.SetCondition(v1alpha1.ServiceReady, metav1.ConditionTrue, v1alpha1.ServiceIsReady, "")
	}
	return nil
}

func (r *SourceReconciler) ApplySourceService(ctx context.Context, source *v1alpha1.Source) error {
	if meta.IsStatusConditionTrue(source.Status.Conditions, string(v1alpha1.ServiceReady)) {
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
	desiredServiceSpecBytes, _ := json.Marshal(desiredServiceSpec)
	source.SetComponentHash(v1alpha1.Service, spec.GenerateSpecHash(desiredServiceSpecBytes))
	return nil
}

func (r *SourceReconciler) ObserveSourceHPA(ctx context.Context, source *v1alpha1.Source) error {
	if source.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
		source.RemoveCondition(v1alpha1.HPAReady)
		source.RemoveComponentHash(v1alpha1.HPA)
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
			source.SetCondition(v1alpha1.HPAReady, metav1.ConditionFalse, v1alpha1.PendingCreation,
				"source hpa is not created...")
			return nil
		}
		source.SetCondition(v1alpha1.Error, metav1.ConditionTrue, v1alpha1.HPAError,
			fmt.Sprintf("error fetching source hpa: %v", err))
		return err
	}
	if r.checkIfHPANeedUpdate(source) {
		source.SetCondition(v1alpha1.HPAReady, metav1.ConditionFalse, v1alpha1.PendingCreation,
			"wait for the source hpa to be ready")
	} else {
		source.SetCondition(v1alpha1.HPAReady, metav1.ConditionTrue, v1alpha1.HPAIsReady, "")
	}
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
	if meta.IsStatusConditionTrue(source.Status.Conditions, string(v1alpha1.HPAReady)) {
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
	desiredHPASpecBytes, _ := json.Marshal(desiredHPASpec)
	source.SetComponentHash(v1alpha1.HPA, spec.GenerateSpecHash(desiredHPASpecBytes))
	return nil
}

func (r *SourceReconciler) ObserveSourceVPA(ctx context.Context, source *v1alpha1.Source) error {
	if source.Spec.Pod.VPA == nil {
		// VPA not enabled, skip further action
		source.RemoveCondition(v1alpha1.VPAReady)
		source.RemoveComponentHash(v1alpha1.VPA)
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
			source.SetCondition(v1alpha1.VPAReady, metav1.ConditionFalse, v1alpha1.PendingCreation,
				"source vpa is not created...")
			return nil
		}
		source.SetCondition(v1alpha1.Error, metav1.ConditionTrue, v1alpha1.VPAError,
			fmt.Sprintf("error fetching source vpa: %v", err))
		return err
	}
	if r.checkIfVPANeedUpdate(source) {
		source.SetCondition(v1alpha1.VPAReady, metav1.ConditionFalse, v1alpha1.PendingCreation,
			"wait for the source vpa to be ready")
	} else {
		source.SetCondition(v1alpha1.VPAReady, metav1.ConditionTrue, v1alpha1.VPAIsReady, "")
	}
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
	if meta.IsStatusConditionTrue(source.Status.Conditions, string(v1alpha1.VPAReady)) {
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
	desiredVPASpecBytes, _ := json.Marshal(desiredVPASpec)
	source.SetComponentHash(v1alpha1.VPA, spec.GenerateSpecHash(desiredVPASpecBytes))
	return nil
}

func (r *SourceReconciler) checkIfStatefulSetNeedToUpdate(source *v1alpha1.Source, statefulSet *appsv1.StatefulSet) bool {
	desiredObject := spec.MakeSourceStatefulSet(source)
	desiredSpecBytes, _ := json.Marshal(desiredObject.Spec)
	return r.checkIfComponentNeedToUpdate(source, v1alpha1.StatefulSetReady, v1alpha1.StatefulSet, desiredSpecBytes) ||
		statefulSet.Status.ReadyReplicas != *source.Spec.Replicas
}

func (r *SourceReconciler) checkIfServiceNeedToUpdate(source *v1alpha1.Source) bool {
	desiredObject := spec.MakeSourceService(source)
	desiredSpecBytes, _ := json.Marshal(desiredObject.Spec)
	return r.checkIfComponentNeedToUpdate(source, v1alpha1.ServiceReady, v1alpha1.Service, desiredSpecBytes)
}

func (r *SourceReconciler) checkIfHPANeedUpdate(source *v1alpha1.Source) bool {
	desiredObject := spec.MakeSourceHPA(source)
	desiredSpecBytes, _ := json.Marshal(desiredObject.Spec)
	return r.checkIfComponentNeedToUpdate(source, v1alpha1.HPAReady, v1alpha1.HPA, desiredSpecBytes)
}

func (r *SourceReconciler) checkIfVPANeedUpdate(source *v1alpha1.Source) bool {
	desiredObject := spec.MakeSourceVPA(source)
	desiredSpecBytes, _ := json.Marshal(desiredObject.Spec)
	return r.checkIfComponentNeedToUpdate(source, v1alpha1.VPAReady, v1alpha1.VPA, desiredSpecBytes)
}

func (r *SourceReconciler) checkIfComponentNeedToUpdate(source *v1alpha1.Source, condType v1alpha1.ResourceConditionType,
	componentType v1alpha1.Component, desiredSpecBytes []byte) bool {
	if cond := meta.FindStatusCondition(source.Status.Conditions, string(condType)); cond != nil {
		// if the generation has not changed, we do not need to update the component
		if cond.ObservedGeneration == source.Generation {
			return false
		}
		// if the desired specification has not changed, we do not need to update the component
		if specHash := source.GetComponentHash(componentType); specHash != nil {
			if *specHash == spec.GenerateSpecHash(desiredSpecBytes) {
				return false
			}
		}
	}
	return true
}
