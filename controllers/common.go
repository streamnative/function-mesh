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

// Package controllers define k8s operator controllers
package controllers

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/spec"
	autoscaling "k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	CleanUpFinalizerName = "cleanup.subscription.finalizer"
)

func observeVPA(ctx context.Context, r client.Reader, name types.NamespacedName, vpaSpec *v1alpha1.VPASpec, conditions map[v1alpha1.Component]v1alpha1.ResourceCondition) error {
	_, ok := conditions[v1alpha1.VPA]
	condition := v1alpha1.ResourceCondition{Condition: v1alpha1.VPAReady}
	if !ok {
		if vpaSpec != nil {
			condition.Status = metav1.ConditionFalse
			condition.Action = v1alpha1.Create
			conditions[v1alpha1.VPA] = condition
			return nil
		}
		// VPA is not enabled, skip further action
		return nil
	}

	vpa := &vpav1.VerticalPodAutoscaler{}
	err := r.Get(ctx, name, vpa)
	if err != nil {
		if errors.IsNotFound(err) {
			if vpaSpec == nil { // VPA is deleted, delete the status
				delete(conditions, v1alpha1.VPA)
				return nil
			}
			condition.Status = metav1.ConditionFalse
			condition.Action = v1alpha1.Create
			conditions[v1alpha1.VPA] = condition
			return nil
		}
		return err
	}

	// old VPA exists while new Spec removes it, delete the old one
	if vpaSpec == nil {
		condition.Status = metav1.ConditionFalse
		condition.Action = v1alpha1.Delete
		conditions[v1alpha1.VPA] = condition
		return nil
	}

	// compare exists VPA with new Spec
	if !reflect.DeepEqual(vpa.Spec.UpdatePolicy, vpaSpec.UpdatePolicy) ||
		!reflect.DeepEqual(vpa.Spec.ResourcePolicy, vpaSpec.ResourcePolicy) {
		condition.Status = metav1.ConditionFalse
		condition.Action = v1alpha1.Update
		conditions[v1alpha1.VPA] = condition
		return nil
	}

	condition.Action = v1alpha1.NoAction
	condition.Status = metav1.ConditionTrue
	conditions[v1alpha1.VPA] = condition
	return nil
}

func applyVPA(ctx context.Context, r client.Client, logger logr.Logger, condition v1alpha1.ResourceCondition, meta *metav1.ObjectMeta,
	targetRef *autoscaling.CrossVersionObjectReference, vpaSpec *v1alpha1.VPASpec, component string, namespace string, name string) error {
	switch condition.Action {
	case v1alpha1.Create:
		vpa := spec.MakeVPA(meta, targetRef, vpaSpec)
		if err := r.Create(ctx, vpa); err != nil {
			logger.Error(err, "failed to create vertical pod autoscaler", "name", name, "component", component)
			return err
		}

	case v1alpha1.Update:
		vpa := &vpav1.VerticalPodAutoscaler{}
		err := r.Get(ctx, types.NamespacedName{Namespace: namespace,
			Name: meta.Name}, vpa)
		if err != nil {
			logger.Error(err, "failed to update vertical pod autoscaler, cannot find vpa", "name", name, "component", component)
			return err
		}
		newVpa := spec.MakeVPA(meta, targetRef, vpaSpec)
		vpa.Spec = newVpa.Spec
		if err := r.Update(ctx, vpa); err != nil {
			logger.Error(err, "failed to update vertical pod autoscaler", "name", name, "component", component)
			return err
		}

	case v1alpha1.Delete:
		vpa := &vpav1.VerticalPodAutoscaler{}
		err := r.Get(ctx, types.NamespacedName{Namespace: namespace,
			Name: meta.Name}, vpa)
		if err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			logger.Error(err, "failed to delete vertical pod autoscaler, cannot find vpa", "name", name, "component", component)
			return err
		}
		err = r.Delete(ctx, vpa)
		if err != nil {
			logger.Error(err, "failed to delete vertical pod autoscaler", "name", name, "component", component)
			return err
		}

	case v1alpha1.Wait, v1alpha1.NoAction:
		// do nothing
	}
	return nil
}

func containsString(arr []string, target string) bool {
	for _, str := range arr {
		if str == target {
			return true
		}
	}
	return false
}

func removeString(arr []string, target string) []string {
	var result []string
	for _, str := range arr {
		if str != target {
			result = append(result, str)
		}
	}
	return result
}
