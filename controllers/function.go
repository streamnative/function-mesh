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
	"reflect"

	"github.com/streamnative/function-mesh/api/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/spec"
	appsv1 "k8s.io/api/apps/v1"
	autov2beta2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *FunctionReconciler) ObserveFunctionStatefulSet(ctx context.Context, req ctrl.Request,
	function *v1alpha1.Function) error {
	condition, ok := function.Status.Conditions[v1alpha1.StatefulSet]
	if !ok {
		function.Status.Conditions[v1alpha1.StatefulSet] = v1alpha1.ResourceCondition{
			Condition: v1alpha1.StatefulSetReady,
			Status:    metav1.ConditionFalse,
			Action:    v1alpha1.Create,
		}
		return nil
	}

	statefulSet := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: function.Namespace,
		Name:      spec.MakeFunctionObjectMeta(function).Name,
	}, statefulSet)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("function is not ready yet...")
			return nil
		}
		return err
	}

	selector, err := metav1.LabelSelectorAsSelector(statefulSet.Spec.Selector)
	if err != nil {
		r.Log.Error(err, "error retrieving statefulSet selector")
		return err
	}
	function.Status.Selector = selector.String()

	if *statefulSet.Spec.Replicas != *function.Spec.Replicas {
		condition.Status = metav1.ConditionFalse
		condition.Action = v1alpha1.Update
		function.Status.Conditions[v1alpha1.StatefulSet] = condition
		return nil
	}

	if statefulSet.Status.ReadyReplicas == *function.Spec.Replicas {
		condition.Action = v1alpha1.NoAction
		condition.Status = metav1.ConditionTrue
	} else {
		condition.Action = v1alpha1.Wait
	}
	function.Status.Replicas = *statefulSet.Spec.Replicas
	function.Status.Conditions[v1alpha1.StatefulSet] = condition

	return nil
}

func (r *FunctionReconciler) ApplyFunctionStatefulSet(ctx context.Context, req ctrl.Request,
	function *v1alpha1.Function) error {
	condition := function.Status.Conditions[v1alpha1.StatefulSet]
	if condition.Status == metav1.ConditionTrue {
		return nil
	}

	switch condition.Action {
	case v1alpha1.Create:
		statefulSet := spec.MakeFunctionStatefulSet(function)
		if err := r.Create(ctx, statefulSet); err != nil {
			r.Log.Error(err, "Failed to create new function statefulSet")
			return err
		}
	case v1alpha1.Update:
		statefulSet := spec.MakeFunctionStatefulSet(function)
		if err := r.Update(ctx, statefulSet); err != nil {
			r.Log.Error(err, "Failed to update the function statefulSet")
			return err
		}
	case v1alpha1.Wait:
		// do nothing
	}

	return nil
}

func (r *FunctionReconciler) ObserveFunctionService(ctx context.Context, req ctrl.Request,
	function *v1alpha1.Function) error {
	condition, ok := function.Status.Conditions[v1alpha1.Service]
	if !ok {
		function.Status.Conditions[v1alpha1.Service] = v1alpha1.ResourceCondition{
			Condition: v1alpha1.ServiceReady,
			Status:    metav1.ConditionFalse,
			Action:    v1alpha1.Create,
		}
		return nil
	}

	if condition.Status == metav1.ConditionTrue {
		return nil
	}

	svc := &corev1.Service{}
	svcName := spec.MakeHeadlessServiceName(spec.MakeFunctionObjectMeta(function).Name)
	err := r.Get(ctx, types.NamespacedName{Namespace: function.Namespace,
		Name: svcName}, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("service is not created...", "Name", function.Name, "ServiceName", svcName)
			return nil
		}
		return err
	}

	condition.Action = v1alpha1.NoAction
	condition.Status = metav1.ConditionTrue
	function.Status.Conditions[v1alpha1.Service] = condition
	return nil
}

func (r *FunctionReconciler) ApplyFunctionService(ctx context.Context, req ctrl.Request,
	function *v1alpha1.Function) error {
	condition := function.Status.Conditions[v1alpha1.Service]

	if condition.Status == metav1.ConditionTrue {
		return nil
	}

	switch condition.Action {
	case v1alpha1.Create:
		svc := spec.MakeFunctionService(function)
		if err := r.Create(ctx, svc); err != nil {
			r.Log.Error(err, "failed to expose service for function", "name", function.Name)
			return err
		}
	case v1alpha1.Wait:
		// do nothing
	}

	return nil
}

func (r *FunctionReconciler) ObserveFunctionHPA(ctx context.Context, req ctrl.Request,
	function *v1alpha1.Function) error {
	if function.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
		return nil
	}

	condition, ok := function.Status.Conditions[v1alpha1.HPA]
	if !ok {
		function.Status.Conditions[v1alpha1.HPA] = v1alpha1.ResourceCondition{
			Condition: v1alpha1.HPAReady,
			Status:    metav1.ConditionFalse,
			Action:    v1alpha1.Create,
		}
		return nil
	}

	if condition.Status == metav1.ConditionTrue {
		return nil
	}

	hpa := &autov2beta2.HorizontalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{Namespace: function.Namespace,
		Name: spec.MakeFunctionObjectMeta(function).Name}, hpa)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("hpa is not created for function...", "Name", function.Name)
			return nil
		}
		return err
	}

	if hpa.Spec.MaxReplicas != *function.Spec.MaxReplicas ||
		!reflect.DeepEqual(hpa.Spec.Metrics, function.Spec.Pod.AutoScalingMetrics) ||
		(function.Spec.Pod.AutoScalingBehavior != nil && hpa.Spec.Behavior == nil) ||
		(function.Spec.Pod.AutoScalingBehavior != nil && hpa.Spec.Behavior != nil &&
			!reflect.DeepEqual(*hpa.Spec.Behavior, *function.Spec.Pod.AutoScalingBehavior)) {
		condition.Status = metav1.ConditionFalse
		condition.Action = v1alpha1.Update
		function.Status.Conditions[v1alpha1.HPA] = condition
		return nil
	}

	condition.Action = v1alpha1.NoAction
	condition.Status = metav1.ConditionTrue
	function.Status.Conditions[v1alpha1.HPA] = condition
	return nil
}

func (r *FunctionReconciler) ApplyFunctionHPA(ctx context.Context, req ctrl.Request,
	function *v1alpha1.Function) error {
	if function.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
		return nil
	}

	condition := function.Status.Conditions[v1alpha1.HPA]

	if condition.Status == metav1.ConditionTrue {
		return nil
	}

	switch condition.Action {
	case v1alpha1.Create:
		hpa := spec.MakeFunctionHPA(function)
		if err := r.Create(ctx, hpa); err != nil {
			r.Log.Error(err, "failed to create pod autoscaler for function", "name", function.Name)
			return err
		}
	case v1alpha1.Update:
		hpa := &autov2beta2.HorizontalPodAutoscaler{}
		err := r.Get(ctx, types.NamespacedName{Namespace: function.Namespace,
			Name: spec.MakeFunctionObjectMeta(function).Name}, hpa)
		if err != nil {
			r.Log.Error(err, "failed to update pod autoscaler for function, cannot find hpa", "name", function.Name)
			return err
		}
		if hpa.Spec.MaxReplicas != *function.Spec.MaxReplicas {
			hpa.Spec.MaxReplicas = *function.Spec.MaxReplicas
		}
		if len(function.Spec.Pod.AutoScalingMetrics) > 0 && !reflect.DeepEqual(hpa.Spec.Metrics, function.Spec.Pod.AutoScalingMetrics) {
			hpa.Spec.Metrics = function.Spec.Pod.AutoScalingMetrics
		}
		if function.Spec.Pod.AutoScalingBehavior != nil {
			hpa.Spec.Behavior = function.Spec.Pod.AutoScalingBehavior
		}
		if len(function.Spec.Pod.BuiltinAutoscaler) > 0 {
			metrics := spec.MakeMetricsFromBuiltinHPARules(function.Spec.Pod.BuiltinAutoscaler)
			if !reflect.DeepEqual(hpa.Spec.Metrics, metrics) {
				hpa.Spec.Metrics = metrics
			}
		}
		if err := r.Update(ctx, hpa); err != nil {
			r.Log.Error(err, "failed to update pod autoscaler for function", "name", function.Name)
			return err
		}
	case v1alpha1.Wait, v1alpha1.NoAction:
		// do nothing
	}

	return nil
}
