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

func (r *SourceReconciler) ObserveSourceStatefulSet(ctx context.Context, req ctrl.Request,
	source *v1alpha1.Source) error {
	condition, ok := source.Status.Conditions[v1alpha1.StatefulSet]
	if !ok {
		source.Status.Conditions[v1alpha1.StatefulSet] = v1alpha1.ResourceCondition{
			Condition: v1alpha1.StatefulSetReady,
			Status:    metav1.ConditionFalse,
			Action:    v1alpha1.Create,
		}
		return nil
	}

	statefulSet := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: source.Namespace,
		Name:      spec.MakeSourceObjectMeta(source).Name,
	}, statefulSet)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("source is not ready yet...")
			return nil
		}
		return err
	}

	selector, err := metav1.LabelSelectorAsSelector(statefulSet.Spec.Selector)
	if err != nil {
		r.Log.Error(err, "error retrieving statefulSet selector")
		return err
	}
	source.Status.Selector = selector.String()

	if *statefulSet.Spec.Replicas != *source.Spec.Replicas {
		condition.Status = metav1.ConditionFalse
		condition.Action = v1alpha1.Update
		source.Status.Conditions[v1alpha1.StatefulSet] = condition
		return nil
	}

	if statefulSet.Status.ReadyReplicas == *source.Spec.Replicas {
		condition.Action = v1alpha1.NoAction
		condition.Status = metav1.ConditionTrue
	} else {
		condition.Action = v1alpha1.Wait
	}
	source.Status.Replicas = *statefulSet.Spec.Replicas
	source.Status.Conditions[v1alpha1.StatefulSet] = condition

	return nil
}

func (r *SourceReconciler) ApplySourceStatefulSet(ctx context.Context, req ctrl.Request,
	source *v1alpha1.Source) error {
	condition := source.Status.Conditions[v1alpha1.StatefulSet]

	if condition.Status == metav1.ConditionTrue {
		return nil
	}

	switch condition.Action {
	case v1alpha1.Create:
		statefulSet := spec.MakeSourceStatefulSet(source)
		if err := r.Create(ctx, statefulSet); err != nil {
			r.Log.Error(err, "failed to create new source statefulSet")
			return err
		}
	case v1alpha1.Update:
		statefulSet := spec.MakeSourceStatefulSet(source)
		if err := r.Update(ctx, statefulSet); err != nil {
			r.Log.Error(err, "failed to update the source statefulSet")
			return err
		}
	case v1alpha1.Wait:
		// do nothing
	}

	return nil
}

func (r *SourceReconciler) ObserveSourceService(ctx context.Context, req ctrl.Request, source *v1alpha1.Source) error {
	condition, ok := source.Status.Conditions[v1alpha1.Service]
	if !ok {
		source.Status.Conditions[v1alpha1.Service] = v1alpha1.ResourceCondition{
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
	err := r.Get(ctx, types.NamespacedName{Namespace: source.Namespace,
		Name: spec.MakeSourceObjectMeta(source).Name}, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("service is not created...", "Name", source.Name)
			return nil
		}
		return err
	}

	condition.Action = v1alpha1.NoAction
	condition.Status = metav1.ConditionTrue
	source.Status.Conditions[v1alpha1.Service] = condition
	return nil
}

func (r *SourceReconciler) ApplySourceService(ctx context.Context, req ctrl.Request, source *v1alpha1.Source) error {
	condition := source.Status.Conditions[v1alpha1.Service]

	if condition.Status == metav1.ConditionTrue {
		return nil
	}

	switch condition.Action {
	case v1alpha1.Create:
		svc := spec.MakeSourceService(source)
		if err := r.Create(ctx, svc); err != nil {
			r.Log.Error(err, "failed to expose service for source", "name", source.Name)
			return err
		}
	case v1alpha1.Wait:
		// do nothing
	}

	return nil
}

func (r *SourceReconciler) ObserveSourceHPA(ctx context.Context, req ctrl.Request, source *v1alpha1.Source) error {
	if source.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
		return nil
	}

	condition, ok := source.Status.Conditions[v1alpha1.HPA]
	if !ok {
		source.Status.Conditions[v1alpha1.HPA] = v1alpha1.ResourceCondition{
			Condition: v1alpha1.HPAReady,
			Status:    metav1.ConditionFalse,
			Action:    v1alpha1.Create,
		}
		return nil
	}

	if condition.Status == metav1.ConditionTrue {
		return nil
	}

	hpa := &autov1.HorizontalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{Namespace: source.Namespace,
		Name: spec.MakeSourceObjectMeta(source).Name}, hpa)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("hpa is not created for source...", "name", source.Name)
			return nil
		}
		return err
	}

	condition.Action = v1alpha1.NoAction
	condition.Status = metav1.ConditionTrue
	source.Status.Conditions[v1alpha1.HPA] = condition
	return nil
}

func (r *SourceReconciler) ApplySourceHPA(ctx context.Context, req ctrl.Request, source *v1alpha1.Source) error {
	if source.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
		return nil
	}

	condition := source.Status.Conditions[v1alpha1.HPA]

	if condition.Status == metav1.ConditionTrue {
		return nil
	}

	switch condition.Action {
	case v1alpha1.Create:
		hpa := spec.MakeSourceHPA(source)
		if err := r.Create(ctx, hpa); err != nil {
			r.Log.Error(err, "failed to create pod autoscaler for source", "name", source.Name)
			return err
		}
	case v1alpha1.Wait:
		// do nothing
	}

	return nil
}
