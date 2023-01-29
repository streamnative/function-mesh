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

func (r *SinkReconciler) ObserveSinkStatefulSet(ctx context.Context, sink *v1alpha1.Sink) error {
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
			sink.SetCondition(v1alpha1.StatefulSetReady, metav1.ConditionFalse, v1alpha1.PendingCreation,
				"sink statefulSet is not ready yet...")
			return nil
		}
		sink.SetCondition(v1alpha1.Error, metav1.ConditionTrue, v1alpha1.StatefulSetError,
			fmt.Sprintf("error fetching sink statefulSet: %v", err))
		return err
	}

	selector, err := metav1.LabelSelectorAsSelector(statefulSet.Spec.Selector)
	if err != nil {
		r.Log.Error(err, "error retrieving statefulSet selector")
		sink.SetCondition(v1alpha1.Error, metav1.ConditionTrue, v1alpha1.StatefulSetError,
			fmt.Sprintf("error retrieving statefulSet selector: %v", err))
		return err
	}
	sink.Status.Selector = selector.String()

	if r.checkIfStatefulSetNeedToUpdate(sink, statefulSet) {
		sink.SetCondition(v1alpha1.StatefulSetReady, metav1.ConditionFalse, v1alpha1.PendingCreation,
			"wait for the sink statefulSet to be ready")
	} else {
		sink.SetCondition(v1alpha1.StatefulSetReady, metav1.ConditionTrue, v1alpha1.StatefulSetIsReady, "")
	}
	sink.Status.Replicas = *statefulSet.Spec.Replicas
	return nil
}

func (r *SinkReconciler) ApplySinkStatefulSet(ctx context.Context, sink *v1alpha1.Sink) error {
	if meta.IsStatusConditionTrue(sink.Status.Conditions, string(v1alpha1.StatefulSetReady)) {
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
		sink.SetCondition(v1alpha1.StatefulSetReady, metav1.ConditionFalse, v1alpha1.ErrorCreatingStatefulSet,
			fmt.Sprintf("error creating or updating statefulSet for sink: %v", err))
		return err
	}
	sink.SetCondition(v1alpha1.StatefulSetReady, metav1.ConditionFalse, v1alpha1.PendingCreation,
		"creating or updating statefulSet for sink...")
	desiredStatefulSetSpecBytes, _ := json.Marshal(desiredStatefulSetSpec)
	sink.SetComponentHash(v1alpha1.StatefulSet, spec.GenerateSpecHash(desiredStatefulSetSpecBytes))
	return nil
}

func (r *SinkReconciler) ObserveSinkService(ctx context.Context, sink *v1alpha1.Sink) error {
	svc := &corev1.Service{}
	svcName := spec.MakeHeadlessServiceName(spec.MakeSinkObjectMeta(sink).Name)
	err := r.Get(ctx, types.NamespacedName{Namespace: sink.Namespace,
		Name: svcName}, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("sink service is not created...",
				"namespace", sink.Namespace, "name", sink.Name,
				"service name", svcName)
			sink.SetCondition(v1alpha1.ServiceReady, metav1.ConditionFalse, v1alpha1.PendingCreation,
				"sink service is not created...")
			return nil
		}
		sink.SetCondition(v1alpha1.Error, metav1.ConditionTrue, v1alpha1.ServiceError,
			fmt.Sprintf("error fetching sink service: %v", err))
		return err
	}
	if r.checkIfServiceNeedToUpdate(sink) {
		sink.SetCondition(v1alpha1.ServiceReady, metav1.ConditionFalse, v1alpha1.PendingCreation,
			"wait for the sink service to be ready")
	} else {
		sink.SetCondition(v1alpha1.ServiceReady, metav1.ConditionTrue, v1alpha1.ServiceIsReady, "")
	}
	return nil
}

func (r *SinkReconciler) ApplySinkService(ctx context.Context, sink *v1alpha1.Sink) error {
	if meta.IsStatusConditionTrue(sink.Status.Conditions, string(v1alpha1.ServiceReady)) {
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
		sink.SetCondition(v1alpha1.ServiceReady, metav1.ConditionFalse, v1alpha1.ErrorCreatingService,
			fmt.Sprintf("error creating or updating service for sink: %v", err))
		return err
	}
	sink.SetCondition(v1alpha1.ServiceReady, metav1.ConditionTrue, v1alpha1.ServiceIsReady, "")
	desiredServiceSpecBytes, _ := json.Marshal(desiredServiceSpec)
	sink.SetComponentHash(v1alpha1.Service, spec.GenerateSpecHash(desiredServiceSpecBytes))
	return nil
}

func (r *SinkReconciler) ObserveSinkHPA(ctx context.Context, sink *v1alpha1.Sink) error {
	if sink.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
		sink.RemoveCondition(v1alpha1.HPAReady)
		sink.RemoveComponentHash(v1alpha1.HPA)
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
			sink.SetCondition(v1alpha1.HPAReady, metav1.ConditionFalse, v1alpha1.PendingCreation,
				"sink hpa is not created...")
			return nil
		}
		sink.SetCondition(v1alpha1.Error, metav1.ConditionTrue, v1alpha1.HPAError,
			fmt.Sprintf("error fetching sink hpa: %v", err))
		return err
	}
	if r.checkIfHPANeedUpdate(sink) {
		sink.SetCondition(v1alpha1.HPAReady, metav1.ConditionFalse, v1alpha1.PendingCreation,
			"wait for the sink hpa to be ready")
	} else {
		sink.SetCondition(v1alpha1.HPAReady, metav1.ConditionTrue, v1alpha1.HPAIsReady, "")
	}
	return nil
}

func (r *SinkReconciler) ApplySinkHPA(ctx context.Context, sink *v1alpha1.Sink) error {
	if sink.Spec.MaxReplicas == nil {
		// HPA not enabled, clear the exists HPA
		hpa := &autov2beta2.HorizontalPodAutoscaler{}
		hpa.Namespace = sink.Namespace
		hpa.Name = spec.MakeSinkObjectMeta(sink).Name
		if err := r.Delete(ctx, hpa); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			sink.SetCondition(v1alpha1.Error, metav1.ConditionTrue, v1alpha1.HPAError,
				fmt.Sprintf("error deleting hpa for sink: %v", err))
			return err
		}
		return nil
	}
	if meta.IsStatusConditionTrue(sink.Status.Conditions, string(v1alpha1.HPAReady)) {
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
		sink.SetCondition(v1alpha1.HPAReady, metav1.ConditionFalse, v1alpha1.ErrorCreatingHPA,
			fmt.Sprintf("error creating or updating hpa for sink: %v", err))
		return err
	}
	sink.SetCondition(v1alpha1.HPAReady, metav1.ConditionTrue, v1alpha1.HPAIsReady, "")
	desiredHPASpecBytes, _ := json.Marshal(desiredHPASpec)
	sink.SetComponentHash(v1alpha1.HPA, spec.GenerateSpecHash(desiredHPASpecBytes))
	return nil
}

func (r *SinkReconciler) ObserveSinkVPA(ctx context.Context, sink *v1alpha1.Sink) error {
	if sink.Spec.Pod.VPA == nil {
		// VPA not enabled, skip further action
		sink.RemoveCondition(v1alpha1.VPAReady)
		sink.RemoveComponentHash(v1alpha1.VPA)
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
			sink.SetCondition(v1alpha1.VPAReady, metav1.ConditionFalse, v1alpha1.PendingCreation,
				"sink vpa is not created...")
			return nil
		}
		sink.SetCondition(v1alpha1.Error, metav1.ConditionTrue, v1alpha1.VPAError,
			fmt.Sprintf("error fetching sink vpa: %v", err))
		return err
	}
	if r.checkIfVPANeedUpdate(sink) {
		sink.SetCondition(v1alpha1.VPAReady, metav1.ConditionFalse, v1alpha1.PendingCreation,
			"wait for the sink vpa to be ready")
	} else {
		sink.SetCondition(v1alpha1.VPAReady, metav1.ConditionTrue, v1alpha1.VPAIsReady, "")
	}
	return nil
}

func (r *SinkReconciler) ApplySinkVPA(ctx context.Context, sink *v1alpha1.Sink) error {
	if sink.Spec.Pod.VPA == nil {
		// VPA not enabled, clear the exists VPA
		vpa := &vpav1.VerticalPodAutoscaler{}
		vpa.Namespace = sink.Namespace
		vpa.Name = spec.MakeSinkObjectMeta(sink).Name
		if err := r.Delete(ctx, vpa); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			sink.SetCondition(v1alpha1.Error, metav1.ConditionTrue, v1alpha1.VPAError,
				fmt.Sprintf("error deleting vpa for sink: %v", err))
			return err
		}
		return nil
	}
	if meta.IsStatusConditionTrue(sink.Status.Conditions, string(v1alpha1.VPAReady)) {
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
		sink.SetCondition(v1alpha1.VPAReady, metav1.ConditionFalse, v1alpha1.ErrorCreatingVPA,
			fmt.Sprintf("error creating or updating vpa for sink: %v", err))
		return err
	}
	sink.SetCondition(v1alpha1.VPAReady, metav1.ConditionTrue, v1alpha1.VPAIsReady, "")
	desiredVPASpecBytes, _ := json.Marshal(desiredVPASpec)
	sink.SetComponentHash(v1alpha1.VPA, spec.GenerateSpecHash(desiredVPASpecBytes))
	return nil
}

func (r *SinkReconciler) checkIfStatefulSetNeedToUpdate(sink *v1alpha1.Sink, statefulSet *appsv1.StatefulSet) bool {
	desiredObject := spec.MakeSinkStatefulSet(sink)
	desiredSpecBytes, _ := json.Marshal(desiredObject.Spec)
	return r.checkIfComponentNeedToUpdate(sink, v1alpha1.StatefulSetReady, v1alpha1.StatefulSet, desiredSpecBytes) ||
		statefulSet.Status.ReadyReplicas != *sink.Spec.Replicas
}

func (r *SinkReconciler) checkIfServiceNeedToUpdate(sink *v1alpha1.Sink) bool {
	desiredObject := spec.MakeSinkService(sink)
	desiredSpecBytes, _ := json.Marshal(desiredObject.Spec)
	return r.checkIfComponentNeedToUpdate(sink, v1alpha1.ServiceReady, v1alpha1.Service, desiredSpecBytes)
}

func (r *SinkReconciler) checkIfHPANeedUpdate(sink *v1alpha1.Sink) bool {
	desiredObject := spec.MakeSinkHPA(sink)
	desiredSpecBytes, _ := json.Marshal(desiredObject.Spec)
	return r.checkIfComponentNeedToUpdate(sink, v1alpha1.HPAReady, v1alpha1.HPA, desiredSpecBytes)
}

func (r *SinkReconciler) checkIfVPANeedUpdate(sink *v1alpha1.Sink) bool {
	desiredObject := spec.MakeSinkVPA(sink)
	desiredSpecBytes, _ := json.Marshal(desiredObject.Spec)
	return r.checkIfComponentNeedToUpdate(sink, v1alpha1.VPAReady, v1alpha1.VPA, desiredSpecBytes)
}

func (r *SinkReconciler) checkIfComponentNeedToUpdate(sink *v1alpha1.Sink, condType v1alpha1.ResourceConditionType,
	componentType v1alpha1.Component, desiredSpecBytes []byte) bool {
	if cond := meta.FindStatusCondition(sink.Status.Conditions, string(condType)); cond != nil {
		// if the generation has not changed, we do not need to update the component
		if cond.ObservedGeneration == sink.Generation {
			return false
		}
		// if the desired specification has not changed, we do not need to update the component
		if specHash := sink.GetComponentHash(componentType); specHash != nil {
			if *specHash == spec.GenerateSpecHash(desiredSpecBytes) {
				return false
			}
		}
	}
	return true
}
