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

	"github.com/streamnative/function-mesh/utils"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/go-logr/logr"
	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/spec"
	appsv1 "k8s.io/api/apps/v1"
	autov2 "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	CleanUpFinalizerName = "cleanup.subscription.finalizer"
)

func deleteHPAV2Beta2(ctx context.Context, r client.Client, name types.NamespacedName) error {
	hpa := &autoscalingv2beta2.HorizontalPodAutoscaler{}
	err := r.Get(ctx, name, hpa)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	err = r.Delete(ctx, hpa)
	if err != nil {
		return err
	}
	return nil
}

func deleteHPA(ctx context.Context, r client.Client, name types.NamespacedName) error {
	hpa := &autov2.HorizontalPodAutoscaler{}
	err := r.Get(ctx, name, hpa)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	err = r.Delete(ctx, hpa)
	if err != nil {
		return err
	}
	return nil
}

func observeVPA(ctx context.Context, r client.Reader, name types.NamespacedName, vpaSpec *v1alpha1.VPASpec,
	conditions map[v1alpha1.Component]v1alpha1.ResourceCondition) (*corev1.ResourceRequirements, error) {
	_, ok := conditions[v1alpha1.VPA]
	condition := v1alpha1.ResourceCondition{Condition: v1alpha1.VPAReady}
	if !ok {
		if vpaSpec != nil {
			condition.Status = metav1.ConditionFalse
			condition.Action = v1alpha1.Create
			conditions[v1alpha1.VPA] = condition
			return nil, nil
		}
		// VPA is not enabled, skip further action
		return nil, nil
	}

	vpa := &vpav1.VerticalPodAutoscaler{}
	err := r.Get(ctx, name, vpa)
	if err != nil {
		if errors.IsNotFound(err) {
			if vpaSpec == nil { // VPA is deleted, delete the status
				delete(conditions, v1alpha1.VPA)
				return nil, nil
			}
			condition.Status = metav1.ConditionFalse
			condition.Action = v1alpha1.Create
			conditions[v1alpha1.VPA] = condition
			return nil, nil
		}
		return nil, err
	}

	// old VPA exists while new Spec removes it, delete the old one
	if vpaSpec == nil {
		condition.Status = metav1.ConditionFalse
		condition.Action = v1alpha1.Delete
		conditions[v1alpha1.VPA] = condition
		return nil, nil
	}

	// compare exists VPA with new Spec
	updatePolicy := spec.UpdateVPAUpdatePolicy(vpaSpec.UpdatePolicy, vpaSpec.ResourceUnit)
	containerName := spec.GetVPAContainerName(&vpa.ObjectMeta)
	resourcePolicy := spec.UpdateResourcePolicy(vpaSpec.ResourcePolicy, containerName)
	if !reflect.DeepEqual(updatePolicy, vpa.Spec.UpdatePolicy) ||
		!reflect.DeepEqual(resourcePolicy, vpa.Spec.ResourcePolicy) {
		condition.Status = metav1.ConditionFalse
		condition.Action = v1alpha1.Update
		conditions[v1alpha1.VPA] = condition
		return nil, nil
	}

	condition.Action = v1alpha1.NoAction
	condition.Status = metav1.ConditionTrue
	conditions[v1alpha1.VPA] = condition

	// if object has special resource unit, we need to manually apply the recommendation to it
	resources := calculateVPARecommendation(vpa, vpaSpec)
	return resources, nil
}

func calculateVPARecommendation(vpa *vpav1.VerticalPodAutoscaler, vpaSpec *v1alpha1.VPASpec) *corev1.ResourceRequirements {
	var multiple int64 = 0
	if vpaSpec.ResourceUnit == nil || vpaSpec.ResourceUnit.CPU.MilliValue() == 0 && vpaSpec.ResourceUnit.Memory.MilliValue() == 0 {
		return nil
	}
	containerName := spec.GetVPAContainerName(&vpa.ObjectMeta)
	if vpa.Status.Recommendation != nil && vpa.Status.Recommendation.ContainerRecommendations != nil {
		for _, recommend := range vpa.Status.Recommendation.ContainerRecommendations {
			if containerName == recommend.ContainerName {
				// set resource based on CPU
				if recommend.Target.Cpu() != nil && recommend.Target.Cpu().Value() != 0 {
					multiple = recommend.Target.Cpu().MilliValue() / vpaSpec.ResourceUnit.CPU.MilliValue()
					if recommend.Target.Cpu().MilliValue()%vpaSpec.ResourceUnit.CPU.MilliValue() != 0 {
						multiple += 1
					}
				}
				if recommend.Target.Memory() != nil && recommend.Target.Memory().Value() != 0 {
					multipleMemory := recommend.Target.Memory().MilliValue() / vpaSpec.ResourceUnit.Memory.MilliValue()
					if recommend.Target.Memory().MilliValue()%vpaSpec.ResourceUnit.Memory.MilliValue() != 0 {
						multipleMemory += 1
					}
					// use the larger multiple
					if multipleMemory > multiple {
						multiple = multipleMemory
					}
				}
			}
		}
	}

	if multiple == 0 {
		return nil
	}
	targetCPU := *resource.NewScaledQuantity(multiple*vpaSpec.ResourceUnit.CPU.MilliValue(), resource.Milli)
	targetMemory := *resource.NewScaledQuantity(multiple*vpaSpec.ResourceUnit.Memory.MilliValue(), resource.Milli)

	if vpa.Spec.ResourcePolicy == nil || len(vpa.Spec.ResourcePolicy.ContainerPolicies) == 0 ||
		vpa.Spec.ResourcePolicy.ContainerPolicies[0].ControlledValues == nil ||
		*vpa.Spec.ResourcePolicy.ContainerPolicies[0].ControlledValues == vpav1.ContainerControlledValuesRequestsAndLimits {
		return &corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    targetCPU,
				corev1.ResourceMemory: targetMemory,
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    targetCPU,
				corev1.ResourceMemory: targetMemory,
			},
		}
	} else {
		return &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    targetCPU,
				corev1.ResourceMemory: targetMemory,
			},
		}
	}
}

func applyVPA(ctx context.Context, r client.Client, logger logr.Logger, condition v1alpha1.ResourceCondition,
	meta *metav1.ObjectMeta,
	targetRef *autov2.CrossVersionObjectReference, vpaSpec *v1alpha1.VPASpec, component string, namespace string,
	name string) error {
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
			logger.Error(err, "failed to update vertical pod autoscaler, cannot find vpa", "name", name, "component",
				component)
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
			logger.Error(err, "failed to delete vertical pod autoscaler, cannot find vpa", "name", name, "component",
				component)
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

func containsCleanupFinalizer(arr []string) bool {
	for _, str := range arr {
		if str == CleanUpFinalizerName {
			return true
		}
	}
	return false
}

func removeCleanupFinalizer(arr []string) []string {
	var result []string
	for _, str := range arr {
		if str != CleanUpFinalizerName {
			result = append(result, str)
		}
	}
	return result
}

func keepStatefulSetUnchangeableFields(ctx context.Context, reader client.Reader, logger logr.Logger,
	desiredStatefulSet *appsv1.StatefulSet) {
	existingStatefulSet := &appsv1.StatefulSet{}
	err := reader.Get(ctx, types.NamespacedName{
		Namespace: desiredStatefulSet.Namespace,
		Name:      desiredStatefulSet.Name,
	}, existingStatefulSet)
	// ignore get error
	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "error get statefulSet workload",
				"namespace", desiredStatefulSet.Namespace, "name", desiredStatefulSet.Name)
		}
		existingStatefulSet = nil
	}

	// below fields are not modifiable, so keep same with the original statefulSet if existing
	if existingStatefulSet != nil {
		desiredStatefulSet.Spec.Selector = existingStatefulSet.Spec.Selector
		// ensure the labels in template match with selector
		for key, val := range desiredStatefulSet.Spec.Selector.MatchLabels {
			desiredStatefulSet.Spec.Template.Labels[key] = val
		}
		desiredStatefulSet.Spec.PodManagementPolicy = existingStatefulSet.Spec.PodManagementPolicy
		desiredStatefulSet.Spec.ServiceName = existingStatefulSet.Spec.ServiceName
		desiredStatefulSet.Spec.VolumeClaimTemplates = existingStatefulSet.Spec.VolumeClaimTemplates
	}
}

func AddControllerBuilderOwn(b *builder.Builder, gv string) *builder.Builder {
	switch gv {
	case utils.GroupVersionV2:
		return b.Owns(&autov2.HorizontalPodAutoscaler{},
			builder.WithPredicates(predicate.GenerationChangedPredicate{}))
	case utils.GroupVersionV2Beta2:
		return b.Owns(&autoscalingv2beta2.HorizontalPodAutoscaler{},
			builder.WithPredicates(predicate.GenerationChangedPredicate{}))
	default:
		panic("Invalid autoscaling group version [" + gv + "]")
	}
}

func ConvertHPAV2ToV2beta2(hpa *autov2.HorizontalPodAutoscaler) *autoscalingv2beta2.HorizontalPodAutoscaler {
	if hpa == nil {
		return nil
	}

	result := &autoscalingv2beta2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: hpa.Namespace,
			Name:      hpa.Name,
		},
		Spec: autoscalingv2beta2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2beta2.CrossVersionObjectReference{
				APIVersion: hpa.Spec.ScaleTargetRef.APIVersion,
				Kind:       hpa.Spec.ScaleTargetRef.Kind,
				Name:       hpa.Spec.ScaleTargetRef.Name,
			},
			MinReplicas: hpa.Spec.MinReplicas,
			MaxReplicas: hpa.Spec.MaxReplicas,
			Metrics:     make([]autoscalingv2beta2.MetricSpec, len(hpa.Spec.Metrics)),
		},
	}

	result.Spec.Behavior = &autoscalingv2beta2.HorizontalPodAutoscalerBehavior{}
	if hpa.Spec.Behavior != nil {
		if hpa.Spec.Behavior.ScaleUp != nil {
			result.Spec.Behavior.ScaleUp.StabilizationWindowSeconds = hpa.Spec.Behavior.ScaleUp.StabilizationWindowSeconds
			result.Spec.Behavior.ScaleUp.SelectPolicy = (*autoscalingv2beta2.ScalingPolicySelect)(hpa.Spec.Behavior.ScaleUp.SelectPolicy)
			result.Spec.Behavior.ScaleUp.Policies = make([]autoscalingv2beta2.HPAScalingPolicy, len(hpa.Spec.Behavior.ScaleUp.Policies))
			for i, policy := range hpa.Spec.Behavior.ScaleUp.Policies {
				result.Spec.Behavior.ScaleUp.Policies[i] = autoscalingv2beta2.HPAScalingPolicy{
					Type:          autoscalingv2beta2.HPAScalingPolicyType(policy.Type),
					Value:         policy.Value,
					PeriodSeconds: policy.PeriodSeconds,
				}
			}
		}
		if hpa.Spec.Behavior.ScaleDown != nil {
			result.Spec.Behavior.ScaleDown.StabilizationWindowSeconds = hpa.Spec.Behavior.ScaleDown.StabilizationWindowSeconds
			result.Spec.Behavior.ScaleDown.SelectPolicy = (*autoscalingv2beta2.ScalingPolicySelect)(hpa.Spec.Behavior.ScaleDown.SelectPolicy)
			result.Spec.Behavior.ScaleDown.Policies = make([]autoscalingv2beta2.HPAScalingPolicy, len(hpa.Spec.Behavior.ScaleDown.Policies))
			for i, policy := range hpa.Spec.Behavior.ScaleDown.Policies {
				result.Spec.Behavior.ScaleDown.Policies[i] = autoscalingv2beta2.HPAScalingPolicy{
					Type:          autoscalingv2beta2.HPAScalingPolicyType(policy.Type),
					Value:         policy.Value,
					PeriodSeconds: policy.PeriodSeconds,
				}
			}
		}
	}

	for i, metric := range hpa.Spec.Metrics {
		ms := autoscalingv2beta2.MetricSpec{Type: autoscalingv2beta2.MetricSourceType(metric.Type)}
		switch metric.Type {
		case autov2.ResourceMetricSourceType:
			ms.Resource = &autoscalingv2beta2.ResourceMetricSource{
				Name: metric.Resource.Name,
				Target: autoscalingv2beta2.MetricTarget{
					Type:               autoscalingv2beta2.MetricTargetType(metric.Resource.Target.Type),
					Value:              metric.Resource.Target.Value,
					AverageValue:       metric.Resource.Target.AverageValue,
					AverageUtilization: metric.Resource.Target.AverageUtilization,
				},
			}
		case autov2.PodsMetricSourceType:
			ms.Pods = &autoscalingv2beta2.PodsMetricSource{
				Metric: autoscalingv2beta2.MetricIdentifier{
					Name:     metric.Pods.Metric.Name,
					Selector: metric.Pods.Metric.Selector,
				},
				Target: autoscalingv2beta2.MetricTarget{
					Type:               autoscalingv2beta2.MetricTargetType(metric.Pods.Target.Type),
					Value:              metric.Pods.Target.Value,
					AverageValue:       metric.Pods.Target.AverageValue,
					AverageUtilization: metric.Pods.Target.AverageUtilization,
				},
			}
		case autov2.ObjectMetricSourceType:
			ms.Object = &autoscalingv2beta2.ObjectMetricSource{
				DescribedObject: autoscalingv2beta2.CrossVersionObjectReference{
					Kind:       metric.Object.DescribedObject.Kind,
					Name:       metric.Object.DescribedObject.Name,
					APIVersion: metric.Object.DescribedObject.APIVersion,
				},
				Metric: autoscalingv2beta2.MetricIdentifier{
					Name:     metric.Object.Metric.Name,
					Selector: metric.Object.Metric.Selector,
				},
				Target: autoscalingv2beta2.MetricTarget{
					Type:               autoscalingv2beta2.MetricTargetType(metric.Object.Target.Type),
					Value:              metric.Object.Target.Value,
					AverageValue:       metric.Object.Target.AverageValue,
					AverageUtilization: metric.Object.Target.AverageUtilization,
				},
			}
		case autov2.ContainerResourceMetricSourceType:
			ms.ContainerResource = &autoscalingv2beta2.ContainerResourceMetricSource{
				Name:      metric.ContainerResource.Name,
				Container: metric.ContainerResource.Container,
				Target: autoscalingv2beta2.MetricTarget{
					Type:               autoscalingv2beta2.MetricTargetType(metric.ContainerResource.Target.Type),
					Value:              metric.ContainerResource.Target.Value,
					AverageValue:       metric.ContainerResource.Target.AverageValue,
					AverageUtilization: metric.ContainerResource.Target.AverageUtilization,
				},
			}
		case autov2.ExternalMetricSourceType:
			ms.External = &autoscalingv2beta2.ExternalMetricSource{
				Metric: autoscalingv2beta2.MetricIdentifier{
					Name:     metric.External.Metric.Name,
					Selector: metric.External.Metric.Selector,
				},
				Target: autoscalingv2beta2.MetricTarget{
					Type:               autoscalingv2beta2.MetricTargetType(metric.External.Target.Type),
					Value:              metric.External.Target.Value,
					AverageValue:       metric.External.Target.AverageValue,
					AverageUtilization: metric.External.Target.AverageUtilization,
				},
			}
		}
		result.Spec.Metrics[i] = ms
	}

	return result
}

func getBackgroundDeletionPolicy() client.DeleteOption {
	backgroundDeletion := metav1.DeletePropagationBackground
	var deleteOptions client.DeleteOption = &client.DeleteOptions{
		PropagationPolicy: &backgroundDeletion,
	}
	return deleteOptions
}
