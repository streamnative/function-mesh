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

func (r *SinkReconciler) ObserveSinkStatefulSet(ctx context.Context, sink *v1alpha1.Sink) error {
	statefulSet := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: sink.Namespace,
		Name:      spec.MakeSinkObjectMeta(sink).Name,
	}, statefulSet)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("sink statefulset is not ready yet...",
				"namespace", sink.Namespace, "name", sink.Name,
				"statefulset name", statefulSet.Name)
			sink.SetCondition(v1alpha1.StatefulSet, v1alpha1.CreateCondition(
				v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
				"sink statefulset is not ready yet..."))
			return nil
		}
		sink.SetCondition(v1alpha1.StatefulSet, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.StatefulSetError,
			fmt.Sprintf("failed to fetch sink statefulset: %v", err)))
		return err
	}

	selector, err := metav1.LabelSelectorAsSelector(statefulSet.Spec.Selector)
	if err != nil {
		r.Log.Error(err, "error retrieving statefulset selector")
		sink.SetCondition(v1alpha1.StatefulSet, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.StatefulSetError,
			fmt.Sprintf("error retrieving statefulset selector: %v", err)))
		return err
	}
	sink.Status.Selector = selector.String()
	sink.Status.Replicas = *sink.Spec.Replicas

	if r.checkIfStatefulSetNeedUpdate(statefulSet, sink) {
		sink.SetCondition(v1alpha1.StatefulSet, v1alpha1.CreateCondition(
			v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
			"need to update function statefulset..."))
		return nil
	}

	if statefulSet.Status.ReadyReplicas == *sink.Spec.Replicas {
		sink.SetCondition(v1alpha1.StatefulSet, v1alpha1.CreateCondition(
			v1alpha1.Ready, metav1.ConditionTrue, v1alpha1.StatefulSetIsReady,
			""))
		return nil
	}
	sink.SetCondition(v1alpha1.StatefulSet, v1alpha1.CreateCondition(
		v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
		"sink statefulset is pending creation..."))
	return nil
}

func (r *SinkReconciler) ApplySinkStatefulSet(ctx context.Context, sink *v1alpha1.Sink, newGeneration bool) error {
	condition := sink.Status.Conditions[v1alpha1.StatefulSet]
	if condition.Status == metav1.ConditionTrue && !newGeneration {
		return nil
	}

	desiredStatefulSet := spec.MakeSinkStatefulSet(sink)
	desiredStatefulSetSpec := desiredStatefulSet.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredStatefulSet, func() error {
		// sink statefulSet mutate logic
		desiredStatefulSet.Spec = desiredStatefulSetSpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error create or update statefulset workload for sink",
			"namespace", sink.Namespace, "name", sink.Name,
			"statefulset name", desiredStatefulSet.Name)
		sink.SetCondition(v1alpha1.StatefulSet, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.ErrorCreatingStatefulSet,
			fmt.Sprintf("error create or update statefulset workload for sink: %v", err)))
		return err
	}
	sink.SetCondition(v1alpha1.StatefulSet, v1alpha1.CreateCondition(
		v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
		"creating or updating statefulset workload for sink..."))
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
			sink.SetCondition(v1alpha1.Service, v1alpha1.CreateCondition(
				v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
				"sink service is not created..."))
			return nil
		}
		sink.SetCondition(v1alpha1.Service, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.ServiceError,
			fmt.Sprintf("failed to fetch sink service: %v", err)))
		return err
	}

	sink.SetCondition(v1alpha1.Service, v1alpha1.CreateCondition(
		v1alpha1.Ready, metav1.ConditionTrue, v1alpha1.ServiceIsReady,
		""))
	return nil
}

func (r *SinkReconciler) ApplySinkService(ctx context.Context, sink *v1alpha1.Sink, newGeneration bool) error {
	condition := sink.Status.Conditions[v1alpha1.Service]
	if condition.Status == metav1.ConditionTrue && !newGeneration {
		return nil
	}

	desiredService := spec.MakeSinkService(sink)
	desiredServiceSpec := desiredService.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredService, func() error {
		// sink service mutate logic
		desiredService.Spec = desiredServiceSpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error create or update service for sink",
			"namespace", sink.Namespace, "name", sink.Name,
			"service name", desiredService.Name)
		sink.SetCondition(v1alpha1.Service, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.ErrorCreatingService,
			fmt.Sprintf("error create or update service for sink: %v", err)))
		return err
	}
	sink.SetCondition(v1alpha1.Service, v1alpha1.CreateCondition(
		v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
		"creating or updating service for sink..."))
	return nil
}

func (r *SinkReconciler) ObserveSinkHPA(ctx context.Context, sink *v1alpha1.Sink) error {
	if sink.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
		delete(sink.Status.Conditions, v1alpha1.HPA)
		return nil
	}

	hpa := &autov2beta2.HorizontalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{Namespace: sink.Namespace,
		Name: spec.MakeSinkObjectMeta(sink).Name}, hpa)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("sink hpa is not created for sink...",
				"namespace", sink.Namespace, "name", sink.Name,
				"hpa name", hpa.Name)
			sink.SetCondition(v1alpha1.HPA, v1alpha1.CreateCondition(
				v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
				"hpa is not created for sink..."))
			return nil
		}
		sink.SetCondition(v1alpha1.HPA, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.HPAError,
			fmt.Sprintf("failed to fetch sink hpa: %v", err)))
		return err
	}

	if r.checkIfHPANeedUpdate(hpa, sink) {
		sink.SetCondition(v1alpha1.HPA, v1alpha1.CreateCondition(
			v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
			"need to update sink hpa..."))
		return nil
	}

	sink.SetCondition(v1alpha1.HPA, v1alpha1.CreateCondition(
		v1alpha1.Ready, metav1.ConditionTrue, v1alpha1.HPAIsReady,
		""))
	return nil
}

func (r *SinkReconciler) ApplySinkHPA(ctx context.Context, sink *v1alpha1.Sink, newGeneration bool) error {
	if sink.Spec.MaxReplicas == nil {
		// HPA not enabled, clear the exists HPA
		hpa := &autov2beta2.HorizontalPodAutoscaler{}
		hpa.Namespace = sink.Namespace
		hpa.Name = spec.MakeSinkObjectMeta(sink).Name
		if err := r.Delete(ctx, hpa); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			sink.SetCondition(v1alpha1.HPA, v1alpha1.CreateCondition(
				v1alpha1.Error, metav1.ConditionFalse, v1alpha1.HPAError,
				fmt.Sprintf("failed to delete sink hpa: %v", err)))
			return err
		}
		return nil
	}

	condition := sink.Status.Conditions[v1alpha1.HPA]
	if condition.Status == metav1.ConditionTrue && !newGeneration {
		return nil
	}

	desiredHPA := spec.MakeSinkHPA(sink)
	desiredHPASpec := desiredHPA.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredHPA, func() error {
		// sink hpa mutate logic
		desiredHPA.Spec = desiredHPASpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error create or update hpa for sink",
			"namespace", sink.Namespace, "name", sink.Name,
			"hpa name", desiredHPA.Name)
		sink.SetCondition(v1alpha1.HPA, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.ErrorCreatingHPA,
			fmt.Sprintf("error create or update hpa for sink: %v", err)))
		return err
	}
	sink.SetCondition(v1alpha1.HPA, v1alpha1.CreateCondition(
		v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
		"creating or updating hpa for sink..."))
	return nil
}

func (r *SinkReconciler) ObserveSinkVPA(ctx context.Context, sink *v1alpha1.Sink) error {
	if sink.Spec.Pod.VPA == nil {
		delete(sink.Status.Conditions, v1alpha1.VPA)
		return nil
	}

	vpa := &vpav1.VerticalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: sink.Namespace,
		Name:      spec.MakeSinkObjectMeta(sink).Name,
	}, vpa)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("vpa is not created for sink...",
				"namespace", sink.Namespace, "name", sink.Name,
				"vpa name", vpa.Name)
			sink.SetCondition(v1alpha1.VPA, v1alpha1.CreateCondition(
				v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
				"vpa is not created for sink..."))
			return nil
		}
		sink.SetCondition(v1alpha1.VPA, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.VPAError,
			fmt.Sprintf("failed to fetch sink vpa: %v", err)))
		return err
	}

	if r.checkIfVPANeedUpdate(vpa, sink) {
		sink.SetCondition(v1alpha1.VPA, v1alpha1.CreateCondition(
			v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
			"need to update sink vpa..."))
		return nil
	}

	sink.SetCondition(v1alpha1.VPA, v1alpha1.CreateCondition(
		v1alpha1.Ready, metav1.ConditionTrue, v1alpha1.VPAIsReady,
		""))
	return nil
}

func (r *SinkReconciler) ApplySinkVPA(ctx context.Context, sink *v1alpha1.Sink, newGeneration bool) error {
	if sink.Spec.Pod.VPA == nil {
		// VPA not enabled, clear the exists VPA
		vpa := &vpav1.VerticalPodAutoscaler{}
		vpa.Namespace = sink.Namespace
		vpa.Name = spec.MakeSinkObjectMeta(sink).Name
		if err := r.Delete(ctx, vpa); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			sink.SetCondition(v1alpha1.VPA, v1alpha1.CreateCondition(
				v1alpha1.Error, metav1.ConditionFalse, v1alpha1.VPAError,
				fmt.Sprintf("failed to delete sink vpa: %v", err)))
			return err
		}
		return nil
	}

	condition := sink.Status.Conditions[v1alpha1.VPA]
	if condition.Status == metav1.ConditionTrue && !newGeneration {
		return nil
	}

	desiredVPA := spec.MakeSinkVPA(sink)
	desiredVPASpec := desiredVPA.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredVPA, func() error {
		// sink vpa mutate logic
		desiredVPA.Spec = desiredVPASpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error create or update vpa for sink",
			"namespace", sink.Namespace, "name", sink.Name,
			"vpa name", desiredVPA.Name)
		sink.SetCondition(v1alpha1.VPA, v1alpha1.CreateCondition(
			v1alpha1.Error, metav1.ConditionFalse, v1alpha1.ErrorCreatingVPA,
			fmt.Sprintf("error create or update vpa for sink: %v", err)))
		return err
	}
	sink.SetCondition(v1alpha1.VPA, v1alpha1.CreateCondition(
		v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
		"creating or updating vpa for sink..."))
	return nil
}

func (r *SinkReconciler) UpdateObservedGeneration(ctx context.Context, sink *v1alpha1.Sink) {
	sink.Status.ObservedGeneration = sink.Generation
}

func (r *SinkReconciler) checkIfStatefulSetNeedUpdate(statefulSet *appsv1.StatefulSet, sink *v1alpha1.Sink) bool {
	return !spec.CheckIfStatefulSetSpecIsEqual(&statefulSet.Spec, &spec.MakeSinkStatefulSet(sink).Spec)
}

func (r *SinkReconciler) checkIfHPANeedUpdate(hpa *autov2beta2.HorizontalPodAutoscaler, sink *v1alpha1.Sink) bool {
	return !spec.CheckIfHPASpecIsEqual(&hpa.Spec, &spec.MakeSinkHPA(sink).Spec)
}

func (r *SinkReconciler) checkIfVPANeedUpdate(vpa *vpav1.VerticalPodAutoscaler, sink *v1alpha1.Sink) bool {
	return !spec.CheckIfVPASpecIsEqual(&vpa.Spec, &spec.MakeSinkVPA(sink).Spec)
}
