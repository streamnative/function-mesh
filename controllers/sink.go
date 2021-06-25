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

	"github.com/streamnative/function-mesh/api/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/spec"
	appsv1 "k8s.io/api/apps/v1"
	autov1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *SinkReconciler) ObserveSinkStatefulSet(ctx context.Context, req ctrl.Request, sink *v1alpha1.Sink) error {
	condition := v1alpha1.ResourceCondition{
		Condition: v1alpha1.StatefulSetReady,
		Status:    metav1.ConditionFalse,
		Action:    v1alpha1.NoAction,
	}

	statefulSet := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: sink.Namespace,
		Name:      spec.MakeSinkObjectMeta(sink).Name,
	}, statefulSet)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("sink statefulset is not found...")
			condition.Action = v1alpha1.Create
			sink.Status.Conditions[v1alpha1.StatefulSet] = condition
			return nil
		}

		sink.Status.Conditions[v1alpha1.StatefulSet] = condition
		return err
	} else {
		// statefulset created, waiting it to be ready
		condition.Action = v1alpha1.Wait
	}

	if *statefulSet.Spec.Replicas != *sink.Spec.Replicas {
		condition.Action = v1alpha1.Update
	}

	if statefulSet.Status.ReadyReplicas == *sink.Spec.Replicas {
		condition.Status = metav1.ConditionTrue
	}

	sink.Status.Replicas = statefulSet.Status.Replicas
	sink.Status.Conditions[v1alpha1.StatefulSet] = condition

	selector, err := metav1.LabelSelectorAsSelector(statefulSet.Spec.Selector)
	if err != nil {
		r.Log.Error(err, "error retrieving statefulSet selector")
		return err
	}
	sink.Status.Selector = selector.String()
	return nil
}

func (r *SinkReconciler) ApplySinkStatefulSet(ctx context.Context, req ctrl.Request, sink *v1alpha1.Sink) error {
	condition := sink.Status.Conditions[v1alpha1.StatefulSet]

	if condition.Status == metav1.ConditionTrue {
		return nil
	}

	switch condition.Action {
	case v1alpha1.Create:
		statefulSet := spec.MakeSinkStatefulSet(sink)
		if err := r.Create(ctx, statefulSet); err != nil {
			r.Log.Error(err, "failed to create new sink statefulSet")
			return err
		}
	case v1alpha1.Update:
		statefulSet := spec.MakeSinkStatefulSet(sink)
		if err := r.Update(ctx, statefulSet); err != nil {
			r.Log.Error(err, "failed to update the sink statefulSet")
			return err
		}
	case v1alpha1.Wait, v1alpha1.NoAction:
		// do nothing
	}

	return nil
}

func (r *SinkReconciler) ObserveSinkService(ctx context.Context, req ctrl.Request, sink *v1alpha1.Sink) error {
	condition := v1alpha1.ResourceCondition{
		Condition: v1alpha1.ServiceReady,
		Status:    metav1.ConditionFalse,
		Action:    v1alpha1.NoAction,
	}

	svc := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Namespace: sink.Namespace,
		Name: spec.MakeSinkObjectMeta(sink).Name}, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			condition.Action = v1alpha1.Create
			r.Log.Info("service is not created...", "Name", sink.Name)
		} else {
			sink.Status.Conditions[v1alpha1.Service] = condition
			return err
		}
	} else {
		// service object doesn't have status, so once it's created just consider it's ready
		condition.Status = metav1.ConditionTrue
	}

	sink.Status.Conditions[v1alpha1.Service] = condition
	return nil
}

func (r *SinkReconciler) ApplySinkService(ctx context.Context, req ctrl.Request, sink *v1alpha1.Sink) error {
	condition := sink.Status.Conditions[v1alpha1.Service]

	if condition.Status == metav1.ConditionTrue {
		return nil
	}

	switch condition.Action {
	case v1alpha1.Create:
		svc := spec.MakeSinkService(sink)
		if err := r.Create(ctx, svc); err != nil {
			r.Log.Error(err, "failed to expose service for sink", "name", sink.Name)
			return err
		}
	case v1alpha1.Wait, v1alpha1.NoAction:
		// do nothing
	}

	return nil
}

func (r *SinkReconciler) ObserveSinkHPA(ctx context.Context, req ctrl.Request, sink *v1alpha1.Sink) error {
	if sink.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
		return nil
	}

	condition := v1alpha1.ResourceCondition{
		Condition: v1alpha1.HPAReady,
		Status:    metav1.ConditionFalse,
		Action:    v1alpha1.NoAction,
	}

	hpa := &autov1.HorizontalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{Namespace: sink.Namespace,
		Name: spec.MakeSinkObjectMeta(sink).Name}, hpa)
	if err != nil {
		if errors.IsNotFound(err) {
			condition.Action = v1alpha1.Create
			r.Log.Info("hpa is not created for sink...", "name", sink.Name)
		} else {
			sink.Status.Conditions[v1alpha1.HPA] = condition
			return err
		}
	} else {
		// HPA's status doesn't show its readiness, , so once it's created just consider it's ready
		condition.Status = metav1.ConditionTrue
	}

	sink.Status.Conditions[v1alpha1.HPA] = condition
	return nil
}

func (r *SinkReconciler) ApplySinkHPA(ctx context.Context, req ctrl.Request, sink *v1alpha1.Sink) error {
	if sink.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
		return nil
	}

	condition := sink.Status.Conditions[v1alpha1.HPA]

	if condition.Status == metav1.ConditionTrue {
		return nil
	}

	switch condition.Action {
	case v1alpha1.Create:
		hpa := spec.MakeSinkHPA(sink)
		if err := r.Create(ctx, hpa); err != nil {
			r.Log.Error(err, "failed to create pod autoscaler for sink", "name", sink.Name)
			return err
		}
	case v1alpha1.Wait, v1alpha1.NoAction:
		// do nothing
	}

	return nil
}
