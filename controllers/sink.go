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

func (r *SinkReconciler) ObserveSinkStatefulSet(ctx context.Context, sink *computeapi.Sink) error {
	statefulSet := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: sink.Namespace,
		Name:      spec.MakeSinkObjectMeta(sink).Name,
	}, statefulSet)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("sink statefulSet is not found...",
				"namespace", sink.Namespace, "name", sink.Name,
				"statefulSet name", statefulSet.Name)
			sink.SetCondition(apispec.StatefulSetReady, metav1.ConditionFalse, apispec.PendingCreation,
				"sink statefulSet is not ready yet...")
			return nil
		}
		sink.SetCondition(apispec.Error, metav1.ConditionTrue, apispec.StatefulSetError,
			fmt.Sprintf("error fetching sink statefulSet: %v", err))
		return err
	}

	selector, err := metav1.LabelSelectorAsSelector(statefulSet.Spec.Selector)
	if err != nil {
		r.Log.Error(err, "error retrieving statefulSet selector")
		sink.SetCondition(apispec.Error, metav1.ConditionTrue, apispec.StatefulSetError,
			fmt.Sprintf("error retrieving statefulSet selector: %v", err))
		return err
	}
	sink.Status.Selector = selector.String()

	if r.checkIfStatefulSetNeedToUpdate(sink, statefulSet) {
		sink.SetCondition(apispec.StatefulSetReady, metav1.ConditionFalse, apispec.PendingCreation,
			"wait for the sink statefulSet to be ready")
	} else {
		sink.SetCondition(apispec.StatefulSetReady, metav1.ConditionTrue, apispec.StatefulSetIsReady, "")
	}
	sink.Status.Replicas = *statefulSet.Spec.Replicas
	return nil
}

func (r *SinkReconciler) ApplySinkStatefulSet(ctx context.Context, sink *computeapi.Sink) error {
	if meta.IsStatusConditionTrue(sink.Status.Conditions, string(apispec.StatefulSetReady)) {
		return nil
	}
	desiredStatefulSet := spec.MakeSinkStatefulSet(sink)
	desiredStatefulSetSpec := desiredStatefulSet.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredStatefulSet, func() error {
		// sink statefulSet mutate logic
		desiredStatefulSet.Spec = desiredStatefulSetSpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error creating or updating statefulSet workload for sink",
			"namespace", sink.Namespace, "name", sink.Name,
			"statefulSet name", desiredStatefulSet.Name)
		sink.SetCondition(apispec.StatefulSetReady, metav1.ConditionFalse, apispec.ErrorCreatingStatefulSet,
			fmt.Sprintf("error creating or updating statefulSet for sink: %v", err))
		return err
	}
	sink.SetCondition(apispec.StatefulSetReady, metav1.ConditionFalse, apispec.PendingCreation,
		"creating or updating statefulSet for sink...")
	return nil
}

func (r *SinkReconciler) ObserveSinkService(ctx context.Context, sink *computeapi.Sink) error {
	svc := &corev1.Service{}
	svcName := spec.MakeHeadlessServiceName(spec.MakeSinkObjectMeta(sink).Name)
	err := r.Get(ctx, types.NamespacedName{Namespace: sink.Namespace,
		Name: svcName}, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("sink service is not created...",
				"namespace", sink.Namespace, "name", sink.Name,
				"service name", svcName)
			sink.SetCondition(apispec.ServiceReady, metav1.ConditionFalse, apispec.PendingCreation,
				"sink service is not created...")
			return nil
		}
		sink.SetCondition(apispec.Error, metav1.ConditionTrue, apispec.ServiceError,
			fmt.Sprintf("error fetching sink service: %v", err))
		return err
	}
	if r.checkIfServiceNeedToUpdate(sink) {
		sink.SetCondition(apispec.ServiceReady, metav1.ConditionFalse, apispec.PendingCreation,
			"wait for the sink service to be ready")
	} else {
		sink.SetCondition(apispec.ServiceReady, metav1.ConditionTrue, apispec.ServiceIsReady, "")
	}
	return nil
}

func (r *SinkReconciler) ApplySinkService(ctx context.Context, sink *computeapi.Sink) error {
	if meta.IsStatusConditionTrue(sink.Status.Conditions, string(apispec.ServiceReady)) {
		return nil
	}
	desiredService := spec.MakeSinkService(sink)
	desiredServiceSpec := desiredService.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredService, func() error {
		// sink service mutate logic
		desiredService.Spec = desiredServiceSpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error creating or updating service for sink",
			"namespace", sink.Namespace, "name", sink.Name,
			"service name", desiredService.Name)
		sink.SetCondition(apispec.ServiceReady, metav1.ConditionFalse, apispec.ErrorCreatingService,
			fmt.Sprintf("error creating or updating service for sink: %v", err))
		return err
	}
	sink.SetCondition(apispec.ServiceReady, metav1.ConditionTrue, apispec.ServiceIsReady, "")
	return nil
}

func (r *SinkReconciler) ObserveSinkHPA(ctx context.Context, sink *computeapi.Sink) error {
	if sink.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
		sink.RemoveCondition(apispec.HPAReady)
		return nil
	}

	hpa := &autov2beta2.HorizontalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: sink.Namespace,
		Name:      spec.MakeSinkObjectMeta(sink).Name}, hpa)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("sink hpa is not created...",
				"namespace", sink.Namespace, "name", sink.Name,
				"hpa name", hpa.Name)
			sink.SetCondition(apispec.HPAReady, metav1.ConditionFalse, apispec.PendingCreation,
				"sink hpa is not created...")
			return nil
		}
		sink.SetCondition(apispec.Error, metav1.ConditionTrue, apispec.HPAError,
			fmt.Sprintf("error fetching sink hpa: %v", err))
		return err
	}
	if r.checkIfHPANeedUpdate(sink) {
		sink.SetCondition(apispec.HPAReady, metav1.ConditionFalse, apispec.PendingCreation,
			"wait for the sink hpa to be ready")
	} else {
		sink.SetCondition(apispec.HPAReady, metav1.ConditionTrue, apispec.HPAIsReady, "")
	}
	return nil
}

func (r *SinkReconciler) ApplySinkHPA(ctx context.Context, sink *computeapi.Sink) error {
	if sink.Spec.MaxReplicas == nil {
		// HPA not enabled, clear the exists HPA
		hpa := &autov2beta2.HorizontalPodAutoscaler{}
		hpa.Namespace = sink.Namespace
		hpa.Name = spec.MakeSinkObjectMeta(sink).Name
		if err := r.Delete(ctx, hpa); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			sink.SetCondition(apispec.Error, metav1.ConditionTrue, apispec.HPAError,
				fmt.Sprintf("error deleting hpa for sink: %v", err))
			return err
		}
		return nil
	}
	if meta.IsStatusConditionTrue(sink.Status.Conditions, string(apispec.HPAReady)) {
		return nil
	}
	desiredHPA := spec.MakeSinkHPA(sink)
	desiredHPASpec := desiredHPA.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredHPA, func() error {
		// sink hpa mutate logic
		desiredHPA.Spec = desiredHPASpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error creating or updating hpa for sink",
			"namespace", sink.Namespace, "name", sink.Name,
			"hpa name", desiredHPA.Name)
		sink.SetCondition(apispec.HPAReady, metav1.ConditionFalse, apispec.ErrorCreatingHPA,
			fmt.Sprintf("error creating or updating hpa for sink: %v", err))
		return err
	}
	sink.SetCondition(apispec.HPAReady, metav1.ConditionTrue, apispec.HPAIsReady, "")
	return nil
}

func (r *SinkReconciler) ObserveSinkVPA(ctx context.Context, sink *computeapi.Sink) error {
	if sink.Spec.Pod.VPA == nil {
		// VPA not enabled, skip further action
		sink.RemoveCondition(apispec.VPAReady)
		return nil
	}

	vpa := &vpav1.VerticalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: sink.Namespace,
		Name:      spec.MakeSinkObjectMeta(sink).Name}, vpa)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("sink vpa is not created...",
				"namespace", sink.Namespace, "name", sink.Name,
				"vpa name", vpa.Name)
			sink.SetCondition(apispec.VPAReady, metav1.ConditionFalse, apispec.PendingCreation,
				"sink vpa is not created...")
			return nil
		}
		sink.SetCondition(apispec.Error, metav1.ConditionTrue, apispec.VPAError,
			fmt.Sprintf("error fetching sink vpa: %v", err))
		return err
	}
	if r.checkIfVPANeedUpdate(sink) {
		sink.SetCondition(apispec.VPAReady, metav1.ConditionFalse, apispec.PendingCreation,
			"wait for the sink vpa to be ready")
	} else {
		sink.SetCondition(apispec.VPAReady, metav1.ConditionTrue, apispec.VPAIsReady, "")
	}
	return nil
}

func (r *SinkReconciler) ApplySinkVPA(ctx context.Context, sink *computeapi.Sink) error {
	if sink.Spec.Pod.VPA == nil {
		// VPA not enabled, clear the exists VPA
		vpa := &vpav1.VerticalPodAutoscaler{}
		vpa.Namespace = sink.Namespace
		vpa.Name = spec.MakeSinkObjectMeta(sink).Name
		if err := r.Delete(ctx, vpa); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			sink.SetCondition(apispec.Error, metav1.ConditionTrue, apispec.VPAError,
				fmt.Sprintf("error deleting vpa for sink: %v", err))
			return err
		}
		return nil
	}
	if meta.IsStatusConditionTrue(sink.Status.Conditions, string(apispec.VPAReady)) {
		return nil
	}
	desiredVPA := spec.MakeSinkVPA(sink)
	desiredVPASpec := desiredVPA.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredVPA, func() error {
		// function vpa mutate logic
		desiredVPA.Spec = desiredVPASpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error creating or updating vpa for sink",
			"namespace", sink.Namespace, "name", sink.Name,
			"vpa name", desiredVPA.Name)
		sink.SetCondition(apispec.VPAReady, metav1.ConditionFalse, apispec.ErrorCreatingVPA,
			fmt.Sprintf("error creating or updating vpa for sink: %v", err))
		return err
	}
	sink.SetCondition(apispec.VPAReady, metav1.ConditionTrue, apispec.VPAIsReady, "")
	return nil
}

func (r *SinkReconciler) checkIfStatefulSetNeedToUpdate(sink *computeapi.Sink, statefulSet *appsv1.StatefulSet) bool {
	return r.checkIfComponentNeedToUpdate(sink, apispec.StatefulSetReady) ||
		statefulSet.Status.ReadyReplicas != *sink.Spec.Replicas
}

func (r *SinkReconciler) checkIfServiceNeedToUpdate(sink *computeapi.Sink) bool {
	return r.checkIfComponentNeedToUpdate(sink, apispec.ServiceReady)
}

func (r *SinkReconciler) checkIfHPANeedUpdate(sink *computeapi.Sink) bool {
	return r.checkIfComponentNeedToUpdate(sink, apispec.HPAReady)
}

func (r *SinkReconciler) checkIfVPANeedUpdate(sink *computeapi.Sink) bool {
	return r.checkIfComponentNeedToUpdate(sink, apispec.VPAReady)
}

func (r *SinkReconciler) checkIfComponentNeedToUpdate(sink *computeapi.Sink, condType apispec.ResourceConditionType) bool {
	if cond := meta.FindStatusCondition(sink.Status.Conditions, string(condType)); cond != nil {
		// if the generation has not changed, we do not need to update the component
		if cond.ObservedGeneration == sink.Generation {
			return false
		}
	}
	return true
}
