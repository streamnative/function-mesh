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

	autoscaling "k8s.io/api/autoscaling/v1"

	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/spec"
	appsv1 "k8s.io/api/apps/v1"
	autov2beta2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *SinkReconciler) ObserveSinkStatefulSet(ctx context.Context, sink *v1alpha1.Sink) error {
	condition, ok := sink.Status.Conditions[v1alpha1.StatefulSet]
	if !ok {
		sink.Status.Conditions[v1alpha1.StatefulSet] = v1alpha1.ResourceCondition{
			Condition: v1alpha1.StatefulSetReady,
			Status:    metav1.ConditionFalse,
			Action:    v1alpha1.Create,
		}
		return nil
	}

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
			condition.Status = metav1.ConditionFalse
			condition.Action = v1alpha1.Create
			sink.Status.Conditions[v1alpha1.StatefulSet] = condition
			return nil
		}
		return err
	}

	selector, err := metav1.LabelSelectorAsSelector(statefulSet.Spec.Selector)
	if err != nil {
		r.Log.Error(err, "error retrieving statefulSet selector")
		return err
	}
	sink.Status.Selector = selector.String()

	if r.checkIfStatefulSetNeedUpdate(statefulSet, sink) {
		condition.Status = metav1.ConditionFalse
		condition.Action = v1alpha1.Update
		sink.Status.Conditions[v1alpha1.StatefulSet] = condition
		return nil
	}

	if statefulSet.Status.ReadyReplicas == *sink.Spec.Replicas {
		condition.Action = v1alpha1.NoAction
	} else {
		condition.Action = v1alpha1.Wait
	}
	condition.Status = metav1.ConditionTrue
	sink.Status.Replicas = *statefulSet.Spec.Replicas
	sink.Status.Conditions[v1alpha1.StatefulSet] = condition
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
		r.Log.Error(err, "error create or update statefulSet workload for sink",
			"namespace", sink.Namespace, "name", sink.Name,
			"statefulSet name", desiredStatefulSet.Name)
		return err
	}
	return nil
}

func (r *SinkReconciler) ObserveSinkService(ctx context.Context, sink *v1alpha1.Sink) error {
	condition, ok := sink.Status.Conditions[v1alpha1.Service]
	if !ok {
		sink.Status.Conditions[v1alpha1.Service] = v1alpha1.ResourceCondition{
			Condition: v1alpha1.ServiceReady,
			Status:    metav1.ConditionFalse,
			Action:    v1alpha1.Create,
		}
		return nil
	}

	svc := &corev1.Service{}
	svcName := spec.MakeHeadlessServiceName(spec.MakeSinkObjectMeta(sink).Name)
	err := r.Get(ctx, types.NamespacedName{Namespace: sink.Namespace,
		Name: svcName}, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			condition.Status = metav1.ConditionFalse
			condition.Action = v1alpha1.Create
			sink.Status.Conditions[v1alpha1.Service] = condition
			r.Log.Info("sink service is not created...",
				"namespace", sink.Namespace, "name", sink.Name,
				"service name", svcName)
			return nil
		}
		return err
	}

	condition.Action = v1alpha1.NoAction
	condition.Status = metav1.ConditionTrue
	sink.Status.Conditions[v1alpha1.Service] = condition
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
		return err
	}
	return nil
}

func (r *SinkReconciler) ObserveSinkHPA(ctx context.Context, sink *v1alpha1.Sink) error {
	if sink.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
		return nil
	}

	condition, ok := sink.Status.Conditions[v1alpha1.HPA]
	if !ok {
		sink.Status.Conditions[v1alpha1.HPA] = v1alpha1.ResourceCondition{
			Condition: v1alpha1.HPAReady,
			Status:    metav1.ConditionFalse,
			Action:    v1alpha1.Create,
		}
		return nil
	}

	hpa := &autov2beta2.HorizontalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{Namespace: sink.Namespace,
		Name: spec.MakeSinkObjectMeta(sink).Name}, hpa)
	if err != nil {
		if errors.IsNotFound(err) {
			condition.Status = metav1.ConditionFalse
			condition.Action = v1alpha1.Create
			sink.Status.Conditions[v1alpha1.HPA] = condition
			r.Log.Info("sink hpa is not created for sink...",
				"namespace", sink.Namespace, "name", sink.Name,
				"hpa name", hpa.Name)
			return nil
		}
		return err
	}

	if r.checkIfHPANeedUpdate(hpa, sink) {
		condition.Status = metav1.ConditionFalse
		condition.Action = v1alpha1.Update
		sink.Status.Conditions[v1alpha1.HPA] = condition
		return nil
	}

	condition.Action = v1alpha1.NoAction
	condition.Status = metav1.ConditionTrue
	sink.Status.Conditions[v1alpha1.HPA] = condition
	return nil
}

func (r *SinkReconciler) ApplySinkHPA(ctx context.Context, sink *v1alpha1.Sink, newGeneration bool) error {
	if sink.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
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
		return err
	}
	return nil
}

func (r *SinkReconciler) ObserveSinkVPA(ctx context.Context, sink *v1alpha1.Sink) error {
	return observeVPA(ctx, r, types.NamespacedName{Namespace: sink.Namespace,
		Name: spec.MakeSinkObjectMeta(sink).Name}, sink.Spec.Pod.VPA, sink.Status.Conditions)
}

func (r *SinkReconciler) ApplySinkVPA(ctx context.Context, sink *v1alpha1.Sink) error {

	condition, ok := sink.Status.Conditions[v1alpha1.VPA]

	if !ok || condition.Status == metav1.ConditionTrue {
		return nil
	}

	objectMeta := spec.MakeSinkObjectMeta(sink)
	targetRef := &autoscaling.CrossVersionObjectReference{
		Kind:       sink.Kind,
		Name:       sink.Name,
		APIVersion: sink.APIVersion,
	}

	err := applyVPA(ctx, r.Client, r.Log, condition, objectMeta, targetRef, sink.Spec.Pod.VPA, "sink", sink.Namespace, sink.Name)
	if err != nil {
		return err
	}

	return nil
}

func (r *SinkReconciler) checkIfStatefulSetNeedUpdate(statefulSet *appsv1.StatefulSet, sink *v1alpha1.Sink) bool {
	return !spec.CheckIfStatefulSetSpecIsEqual(&statefulSet.Spec, &spec.MakeSinkStatefulSet(sink).Spec)
}

func (r *SinkReconciler) checkIfHPANeedUpdate(hpa *autov2beta2.HorizontalPodAutoscaler, sink *v1alpha1.Sink) bool {
	return !spec.CheckIfHPASpecIsEqual(&hpa.Spec, &spec.MakeSinkHPA(sink).Spec)
}
