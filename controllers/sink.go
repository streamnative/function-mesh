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

	needUpdate, err := r.checkIfStatefulSetNeedUpdate(ctx, statefulSet, sink)
	if err != nil {
		r.Log.Error(err, "error comparing statefulSet")
		return err
	}
	if needUpdate {
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
	desiredStatefulSet, err := spec.MakeSinkStatefulSet(ctx, r.Client, sink)
	if err != nil {
		return err
	}
	keepStatefulSetUnchangeableFields(ctx, r, r.Log, desiredStatefulSet)
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
		delete(sink.Status.Conditions, v1alpha1.HPA)
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

	hpa := &autov2.HorizontalPodAutoscaler{}
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
		// HPA not enabled, delete HPA if it exists
		err := deleteHPA(ctx, r.Client, types.NamespacedName{Namespace: sink.Namespace, Name: sink.Name})
		if err != nil {
			r.Log.Error(err, "failed to delete HPA for sink", "namespace", sink.Namespace, "name", sink.Name)
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
		return err
	}
	return nil
}

func (r *SinkReconciler) ObserveSinkHPAV2Beta2(ctx context.Context, sink *v1alpha1.Sink) error {
	if sink.Spec.MaxReplicas == nil {
		// HPA not enabled, skip further action
		delete(sink.Status.Conditions, v1alpha1.HPA)
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

	hpa := &autoscalingv2beta2.HorizontalPodAutoscaler{}
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

	if r.checkIfHPAV2Beta2NeedUpdate(hpa, sink) {
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

func (r *SinkReconciler) ApplySinkHPAV2Beta2(ctx context.Context, sink *v1alpha1.Sink, newGeneration bool) error {
	if sink.Spec.MaxReplicas == nil {
		// HPA not enabled, delete HPA if it exists
		err := deleteHPAV2Beta2(ctx, r.Client, types.NamespacedName{Namespace: sink.Namespace, Name: sink.Name})
		if err != nil {
			r.Log.Error(err, "failed to delete HPA for sink", "namespace", sink.Namespace, "name", sink.Name)
			return err
		}
		return nil
	}
	condition := sink.Status.Conditions[v1alpha1.HPA]
	if condition.Status == metav1.ConditionTrue && !newGeneration {
		return nil
	}
	desiredHPA := ConvertHPAV2ToV2beta2(spec.MakeSinkHPA(sink))
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
	targetRef := &autov2.CrossVersionObjectReference{
		Kind:       sink.Kind,
		Name:       sink.Name,
		APIVersion: sink.APIVersion,
	}

	err := applyVPA(ctx, r.Client, r.Log, condition, objectMeta, targetRef, sink.Spec.Pod.VPA, "sink", sink.Namespace,
		sink.Name)
	if err != nil {
		return err
	}

	return nil
}

func (r *SinkReconciler) ApplySinkCleanUpJob(ctx context.Context, sink *v1alpha1.Sink) error {
	if !spec.NeedCleanup(sink) {
		desiredJob := spec.MakeSinkCleanUpJob(sink)
		if err := r.Delete(ctx, desiredJob, getBackgroundDeletionPolicy()); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			r.Log.Error(err, "error delete cleanup job for sink",
				"namespace", sink.Namespace, "name", sink.Name,
				"job name", desiredJob.Name)
			return err
		}
		return nil
	}
	hasCleanupFinalizer := containsCleanupFinalizer(sink.ObjectMeta.Finalizers)
	if sink.Spec.CleanupSubscription {
		// add finalizer if sink is updated to clean up subscription
		if sink.ObjectMeta.DeletionTimestamp.IsZero() {
			if !hasCleanupFinalizer {
				desiredJob := spec.MakeSinkCleanUpJob(sink)
				if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredJob, func() error {
					return nil
				}); err != nil {
					r.Log.Error(err, "error create or update clean up job for sink",
						"namespace", sink.Namespace, "name", sink.Name,
						"job name", desiredJob.Name)
					return err
				}
				sink.ObjectMeta.Finalizers = append(sink.ObjectMeta.Finalizers, CleanUpFinalizerName)
				if err := r.Update(ctx, sink); err != nil {
					return err
				}
			}
		} else {
			desiredJob := spec.MakeSinkCleanUpJob(sink)
			// if sink is deleting, send an "INT" signal to the cleanup job to clean up subscription
			if hasCleanupFinalizer {
				if err := spec.TriggerCleanup(ctx, r.Client, r.RestClient, r.Config, desiredJob); err != nil {
					r.Log.Error(err, "error send signal to clean up job for sink",
						"namespace", sink.Namespace, "name", sink.Name)
				}
				sink.ObjectMeta.Finalizers = removeCleanupFinalizer(sink.ObjectMeta.Finalizers)
				if err := r.Update(ctx, sink); err != nil {
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
		// remove finalizer if sink is updated to not cleanup subscription
		if hasCleanupFinalizer {
			sink.ObjectMeta.Finalizers = removeCleanupFinalizer(sink.ObjectMeta.Finalizers)
			if err := r.Update(ctx, sink); err != nil {
				return err
			}

			desiredJob := spec.MakeSinkCleanUpJob(sink)
			// delete the cleanup job
			if err := r.Delete(ctx, desiredJob, getBackgroundDeletionPolicy()); err != nil {
				return err
			}

		}
	}
	return nil
}

func (r *SinkReconciler) checkIfStatefulSetNeedUpdate(ctx context.Context, statefulSet *appsv1.StatefulSet, sink *v1alpha1.Sink) (bool, error) {
	desiredStatefulSet, err := spec.MakeSinkStatefulSet(ctx, r.Client, sink)
	if err != nil {
		return false, err
	}
	diff, err := spec.CreateDiff(statefulSet, desiredStatefulSet)
	if err != nil {
		return false, err
	}
	sink.Status.PendingChange = diff
	return !spec.CheckIfStatefulSetSpecIsEqual(&statefulSet.Spec, &desiredStatefulSet.Spec), nil
}

func (r *SinkReconciler) checkIfHPANeedUpdate(hpa *autov2.HorizontalPodAutoscaler, sink *v1alpha1.Sink) bool {
	return !spec.CheckIfHPASpecIsEqual(&hpa.Spec, &spec.MakeSinkHPA(sink).Spec)
}

func (r *SinkReconciler) checkIfHPAV2Beta2NeedUpdate(hpa *autoscalingv2beta2.HorizontalPodAutoscaler,
	sink *v1alpha1.Sink) bool {
	return !spec.CheckIfHPAV2Beta2SpecIsEqual(&hpa.Spec, &ConvertHPAV2ToV2beta2(spec.MakeSinkHPA(sink)).Spec)
}
