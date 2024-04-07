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

	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"

	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/spec"
	appsv1 "k8s.io/api/apps/v1"
	autov2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *SourceReconciler) ObserveSourceStatefulSet(ctx context.Context, source *v1alpha1.Source) error {
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
			r.Log.Info("source statefulSet is not ready yet...",
				"namespace", source.Namespace, "name", source.Name,
				"statefulSet name", statefulSet.Name)
			condition.Status = metav1.ConditionFalse
			condition.Action = v1alpha1.Create
			source.Status.Conditions[v1alpha1.StatefulSet] = condition
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

	needUpdate, err := r.checkIfStatefulSetNeedUpdate(ctx, statefulSet, source)
	if err != nil {
		r.Log.Error(err, "error comparing statefulSet")
		return err
	}
	if needUpdate {
		condition.Status = metav1.ConditionFalse
		condition.Action = v1alpha1.Update
		source.Status.Conditions[v1alpha1.StatefulSet] = condition
		return nil
	}

	if statefulSet.Status.ReadyReplicas == *source.Spec.Replicas {
		condition.Action = v1alpha1.NoAction
	} else {
		condition.Action = v1alpha1.Wait
	}
	condition.Status = metav1.ConditionTrue
	source.Status.Replicas = *statefulSet.Spec.Replicas
	source.Status.Conditions[v1alpha1.StatefulSet] = condition
	return nil
}

func (r *SourceReconciler) ApplySourceStatefulSet(ctx context.Context, source *v1alpha1.Source,
	newGeneration bool) error {
	condition := source.Status.Conditions[v1alpha1.StatefulSet]
	if condition.Status == metav1.ConditionTrue && !newGeneration {
		return nil
	}
	desiredStatefulSet, err := spec.MakeSourceStatefulSet(ctx, r.Client, source)
	if err != nil {
		return err
	}
	keepStatefulSetUnchangeableFields(ctx, r, r.Log, desiredStatefulSet)
	desiredStatefulSetSpec := desiredStatefulSet.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredStatefulSet, func() error {
		// source statefulSet mutate logic
		desiredStatefulSet.Spec = desiredStatefulSetSpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error create or update statefulSet workload for source",
			"namespace", source.Namespace, "name", source.Name,
			"statefulSet name", desiredStatefulSet.Name)
		return err
	}
	return nil
}

func (r *SourceReconciler) ObserveSourceService(ctx context.Context, source *v1alpha1.Source) error {
	condition, ok := source.Status.Conditions[v1alpha1.Service]
	if !ok {
		source.Status.Conditions[v1alpha1.Service] = v1alpha1.ResourceCondition{
			Condition: v1alpha1.ServiceReady,
			Status:    metav1.ConditionFalse,
			Action:    v1alpha1.Create,
		}
		return nil
	}

	svc := &corev1.Service{}
	svcName := spec.MakeHeadlessServiceName(spec.MakeSourceObjectMeta(source).Name)
	err := r.Get(ctx, types.NamespacedName{Namespace: source.Namespace,
		Name: svcName}, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			condition.Status = metav1.ConditionFalse
			condition.Action = v1alpha1.Create
			source.Status.Conditions[v1alpha1.Service] = condition
			r.Log.Info("source service is not created...",
				"namespace", source.Namespace, "name", source.Name,
				"service name", svcName)
			return nil
		}
		return err
	}

	condition.Action = v1alpha1.NoAction
	condition.Status = metav1.ConditionTrue
	source.Status.Conditions[v1alpha1.Service] = condition
	return nil
}

func (r *SourceReconciler) ApplySourceService(ctx context.Context, source *v1alpha1.Source, newGeneration bool) error {
	condition := source.Status.Conditions[v1alpha1.Service]
	if condition.Status == metav1.ConditionTrue && !newGeneration {
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
		return err
	}
	return nil
}

func (r *SourceReconciler) ObserveSourceHPA(ctx context.Context, source *v1alpha1.Source) error {
	if source.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
		delete(source.Status.Conditions, v1alpha1.HPA)
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

	hpa := &autov2.HorizontalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{Namespace: source.Namespace,
		Name: spec.MakeSourceObjectMeta(source).Name}, hpa)
	if err != nil {
		if errors.IsNotFound(err) {
			condition.Status = metav1.ConditionFalse
			condition.Action = v1alpha1.Create
			source.Status.Conditions[v1alpha1.HPA] = condition
			r.Log.Info("hpa is not created for source...",
				"namespace", source.Namespace, "name", source.Name,
				"hpa name", hpa.Name)
			return nil
		}
		return err
	}

	if r.checkIfHPANeedUpdate(hpa, source) {
		condition.Status = metav1.ConditionFalse
		condition.Action = v1alpha1.Update
		source.Status.Conditions[v1alpha1.HPA] = condition
		return nil
	}

	condition.Action = v1alpha1.NoAction
	condition.Status = metav1.ConditionTrue
	source.Status.Conditions[v1alpha1.HPA] = condition
	return nil
}

func (r *SourceReconciler) ApplySourceHPA(ctx context.Context, source *v1alpha1.Source, newGeneration bool) error {
	if source.Spec.MaxReplicas == nil {
		// HPA not enabled, delete HPA if it exists
		err := deleteHPA(ctx, r.Client, types.NamespacedName{Namespace: source.Namespace, Name: source.Name})
		if err != nil {
			r.Log.Error(err, "failed to delete HPA for source", "namespace", source.Namespace, "name", source.Name)
			return err
		}
		return nil
	}
	condition := source.Status.Conditions[v1alpha1.HPA]
	if condition.Status == metav1.ConditionTrue && !newGeneration {
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
		return err
	}
	return nil
}

func (r *SourceReconciler) ObserveSourceHPAV2Beta2(ctx context.Context, source *v1alpha1.Source) error {
	if source.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
		delete(source.Status.Conditions, v1alpha1.HPA)
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

	hpa := &autoscalingv2beta2.HorizontalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{Namespace: source.Namespace,
		Name: spec.MakeSourceObjectMeta(source).Name}, hpa)
	if err != nil {
		if errors.IsNotFound(err) {
			condition.Status = metav1.ConditionFalse
			condition.Action = v1alpha1.Create
			source.Status.Conditions[v1alpha1.HPA] = condition
			r.Log.Info("hpa is not created for source...",
				"namespace", source.Namespace, "name", source.Name,
				"hpa name", hpa.Name)
			return nil
		}
		return err
	}

	if r.checkIfHPAV2Beta2NeedUpdate(hpa, source) {
		condition.Status = metav1.ConditionFalse
		condition.Action = v1alpha1.Update
		source.Status.Conditions[v1alpha1.HPA] = condition
		return nil
	}

	condition.Action = v1alpha1.NoAction
	condition.Status = metav1.ConditionTrue
	source.Status.Conditions[v1alpha1.HPA] = condition
	return nil
}

func (r *SourceReconciler) ApplySourceHPAV2Beta2(ctx context.Context, source *v1alpha1.Source,
	newGeneration bool) error {
	if source.Spec.MaxReplicas == nil {
		// HPA not enabled, delete HPA if it exists
		err := deleteHPAV2Beta2(ctx, r.Client, types.NamespacedName{Namespace: source.Namespace, Name: source.Name})
		if err != nil {
			r.Log.Error(err, "failed to delete HPA for source", "namespace", source.Namespace, "name", source.Name)
			return err
		}
		return nil
	}
	condition := source.Status.Conditions[v1alpha1.HPA]
	if condition.Status == metav1.ConditionTrue && !newGeneration {
		return nil
	}
	desiredHPA := ConvertHPAV2ToV2beta2(spec.MakeSourceHPA(source))
	desiredHPASpec := desiredHPA.Spec
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredHPA, func() error {
		// source hpa mutate logic
		desiredHPA.Spec = desiredHPASpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error create or update hpa for source",
			"namespace", source.Namespace, "name", source.Name,
			"hpa name", desiredHPA.Name)
		return err
	}
	return nil
}

func (r *SourceReconciler) ObserveSourceVPA(ctx context.Context, source *v1alpha1.Source) error {
	return observeVPA(ctx, r, types.NamespacedName{Namespace: source.Namespace,
		Name: spec.MakeSourceObjectMeta(source).Name}, source.Spec.Pod.VPA, source.Status.Conditions)
}

func (r *SourceReconciler) ApplySourceVPA(ctx context.Context, source *v1alpha1.Source) error {

	condition, ok := source.Status.Conditions[v1alpha1.VPA]

	if !ok || condition.Status == metav1.ConditionTrue {
		return nil
	}

	objectMeta := spec.MakeSourceObjectMeta(source)
	targetRef := &autov2.CrossVersionObjectReference{
		Kind:       source.Kind,
		Name:       source.Name,
		APIVersion: source.APIVersion,
	}

	err := applyVPA(ctx, r.Client, r.Log, condition, objectMeta, targetRef, source.Spec.Pod.VPA, "source",
		source.Namespace, source.Name)
	if err != nil {
		return err
	}

	return nil
}

func (r *SourceReconciler) ApplySourceCleanUpJob(ctx context.Context, source *v1alpha1.Source) error {
	if !spec.NeedCleanup(source) {
		desiredJob := spec.MakeSourceCleanUpJob(source)
		if err := r.Delete(ctx, desiredJob, getBackgroundDeletionPolicy()); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			r.Log.Error(err, "error delete cleanup job for source",
				"namespace", source.Namespace, "name", source.Name,
				"job name", desiredJob.Name)
			return err
		}
		return nil
	}
	hasCleanupFinalizer := containsCleanupFinalizer(source.ObjectMeta.Finalizers)
	if source.Spec.BatchSourceConfig != nil {
		// add finalizer if source is updated to clean up subscription
		if source.ObjectMeta.DeletionTimestamp.IsZero() {
			if !hasCleanupFinalizer {
				desiredJob := spec.MakeSourceCleanUpJob(source)
				if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredJob, func() error {
					return nil
				}); err != nil {
					r.Log.Error(err, "error create or update clean up job for source",
						"namespace", source.Namespace, "name", source.Name,
						"job name", desiredJob.Name)
					return err
				}
				source.ObjectMeta.Finalizers = append(source.ObjectMeta.Finalizers, CleanUpFinalizerName)
				if err := r.Update(ctx, source); err != nil {
					return err
				}
			}
		} else {
			desiredJob := spec.MakeSourceCleanUpJob(source)
			// if source is deleting, send an "INT" signal to the cleanup job to clean up subscription
			if hasCleanupFinalizer {
				if err := spec.TriggerCleanup(ctx, r.Client, r.RestClient, r.Config, desiredJob); err != nil {
					r.Log.Error(err, "error send signal to clean up job for source",
						"namespace", source.Namespace, "name", source.Name)
				}
				source.ObjectMeta.Finalizers = removeCleanupFinalizer(source.ObjectMeta.Finalizers)
				if err := r.Update(ctx, source); err != nil {
					return err
				}
			} else {
				// delete the cleanup job
				if err := r.Delete(ctx, desiredJob, getBackgroundDeletionPolicy()); err != nil {
					return err
				}
			}
		}
	} else {
		// remove finalizer if source is updated to not cleanup subscription
		if hasCleanupFinalizer {
			source.ObjectMeta.Finalizers = removeCleanupFinalizer(source.ObjectMeta.Finalizers)
			if err := r.Update(ctx, source); err != nil {
				return err
			}

			desiredJob := spec.MakeSourceCleanUpJob(source)
			// delete the cleanup job
			if err := r.Delete(ctx, desiredJob, getBackgroundDeletionPolicy()); err != nil {
				return err
			}

		}
	}
	return nil
}

func (r *SourceReconciler) checkIfStatefulSetNeedUpdate(ctx context.Context, statefulSet *appsv1.StatefulSet, source *v1alpha1.Source) (bool, error) {
	desiredStatefulSet, err := spec.MakeSourceStatefulSet(ctx, r.Client, source)
	if err != nil {
		return false, err
	}
	return !spec.CheckIfStatefulSetSpecIsEqual(&statefulSet.Spec, &desiredStatefulSet.Spec), nil
}

func (r *SourceReconciler) checkIfHPANeedUpdate(hpa *autov2.HorizontalPodAutoscaler, source *v1alpha1.Source) bool {
	return !spec.CheckIfHPASpecIsEqual(&hpa.Spec, &spec.MakeSourceHPA(source).Spec)
}

func (r *SourceReconciler) checkIfHPAV2Beta2NeedUpdate(hpa *autoscalingv2beta2.HorizontalPodAutoscaler,
	source *v1alpha1.Source) bool {
	return !spec.CheckIfHPAV2Beta2SpecIsEqual(&hpa.Spec, &ConvertHPAV2ToV2beta2(spec.MakeSourceHPA(source)).Spec)
}
