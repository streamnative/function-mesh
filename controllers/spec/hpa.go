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

package spec

import (
	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	autov2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BuiltinAutoScaler interface {
	Metrics() []autov2.MetricSpec
}

func isDefaultHPAEnabled(minReplicas, maxReplicas *int32, podPolicy v1alpha1.PodPolicy) bool {
	return minReplicas != nil && maxReplicas != nil && podPolicy.AutoScalingBehavior == nil && len(podPolicy.AutoScalingMetrics) == 0 && len(podPolicy.BuiltinAutoscaler) == 0 && *maxReplicas > *minReplicas
}

func isBuiltinHPAEnabled(minReplicas, maxReplicas *int32, podPolicy v1alpha1.PodPolicy) bool {
	return minReplicas != nil && maxReplicas != nil && len(podPolicy.BuiltinAutoscaler) > 0 && *maxReplicas > *minReplicas
}

type HPARuleAverageUtilizationCPUPercent struct {
	cpuPercentage int32
}

type HPARuleAverageUtilizationResourceMemoryPercent struct {
	memoryPercentage int32
}

func (H *HPARuleAverageUtilizationResourceMemoryPercent) Metrics() []autov2.MetricSpec {
	return []autov2.MetricSpec{
		{
			Type: autov2.ResourceMetricSourceType,
			Resource: &autov2.ResourceMetricSource{
				Name: corev1.ResourceMemory,
				Target: autov2.MetricTarget{
					Type:               autov2.UtilizationMetricType,
					AverageUtilization: &H.memoryPercentage,
				},
			},
		},
	}
}

func (H *HPARuleAverageUtilizationCPUPercent) Metrics() []autov2.MetricSpec {
	return []autov2.MetricSpec{
		{
			Type: autov2.ResourceMetricSourceType,
			Resource: &autov2.ResourceMetricSource{
				Name: corev1.ResourceCPU,
				Target: autov2.MetricTarget{
					Type:               autov2.UtilizationMetricType,
					AverageUtilization: &H.cpuPercentage,
				},
			},
		},
	}
}

func NewHPARuleAverageUtilizationCPUPercent(cpuPercentage int32) BuiltinAutoScaler {
	return &HPARuleAverageUtilizationCPUPercent{
		cpuPercentage: cpuPercentage,
	}
}

func NewHPARuleAverageUtilizationMemoryPercent(memoryPercentage int32) BuiltinAutoScaler {
	return &HPARuleAverageUtilizationResourceMemoryPercent{
		memoryPercentage: memoryPercentage,
	}
}

func GetBuiltinAutoScaler(builtinRule v1alpha1.BuiltinHPARule, res corev1.ResourceRequirements) (BuiltinAutoScaler, int) {
	switch builtinRule {
	case v1alpha1.AverageUtilizationCPUPercent80:
		return NewHPARuleAverageUtilizationCPUPercent(getUtilizationPercentage(80, getResourceCpuFactor(res))), cpuRuleIdx
	case v1alpha1.AverageUtilizationCPUPercent50:
		return NewHPARuleAverageUtilizationCPUPercent(getUtilizationPercentage(50, getResourceCpuFactor(res))), cpuRuleIdx
	case v1alpha1.AverageUtilizationCPUPercent20:
		return NewHPARuleAverageUtilizationCPUPercent(getUtilizationPercentage(20, getResourceCpuFactor(res))), cpuRuleIdx
	case v1alpha1.AverageUtilizationMemoryPercent80:
		return NewHPARuleAverageUtilizationMemoryPercent(getUtilizationPercentage(80, getResourceMemoryFactor(res))), memoryRuleIdx
	case v1alpha1.AverageUtilizationMemoryPercent50:
		return NewHPARuleAverageUtilizationMemoryPercent(getUtilizationPercentage(50, getResourceMemoryFactor(res))), memoryRuleIdx
	case v1alpha1.AverageUtilizationMemoryPercent20:
		return NewHPARuleAverageUtilizationMemoryPercent(getUtilizationPercentage(20, getResourceMemoryFactor(res))), memoryRuleIdx
	default:
		return nil, 2
	}
}

// defaultHPAMetrics generates a default HPA Metrics settings based on CPU usage and utilized on 80%.
func defaultHPAMetrics(res corev1.ResourceRequirements) []autov2.MetricSpec {
	return NewHPARuleAverageUtilizationCPUPercent(getUtilizationPercentage(80, getResourceCpuFactor(res))).Metrics()
}

func makeDefaultHPA(objectMeta *metav1.ObjectMeta, minReplicas, maxReplicas int32, targetRef autov2.CrossVersionObjectReference, res corev1.ResourceRequirements) *autov2.HorizontalPodAutoscaler {
	return &autov2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "autoscaling/v2",
			Kind:       "HorizontalPodAutoscaler",
		},
		ObjectMeta: *objectMeta,
		Spec: autov2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: targetRef,
			MinReplicas:    &minReplicas,
			MaxReplicas:    maxReplicas,
			Metrics:        defaultHPAMetrics(res),
		},
	}
}

const (
	cpuRuleIdx = iota
	memoryRuleIdx
)

func MakeMetricsFromBuiltinHPARules(builtinRules []v1alpha1.BuiltinHPARule, res corev1.ResourceRequirements) []autov2.MetricSpec {
	isRuleExists := map[int]bool{}
	metrics := []autov2.MetricSpec{}
	for _, r := range builtinRules {
		s, idx := GetBuiltinAutoScaler(r, res)
		if s != nil {
			if isRuleExists[idx] {
				continue
			}
			isRuleExists[idx] = true
			metrics = append(metrics, s.Metrics()...)
		}
	}
	return metrics
}

func makeBuiltinHPA(objectMeta *metav1.ObjectMeta, minReplicas, maxReplicas int32, targetRef autov2.CrossVersionObjectReference, builtinRules []v1alpha1.BuiltinHPARule, res corev1.ResourceRequirements) *autov2.HorizontalPodAutoscaler {
	metrics := MakeMetricsFromBuiltinHPARules(builtinRules, res)
	return &autov2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "autoscaling/v2",
			Kind:       "HorizontalPodAutoscaler",
		},
		ObjectMeta: *objectMeta,
		Spec: autov2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: targetRef,
			MinReplicas:    &minReplicas,
			MaxReplicas:    maxReplicas,
			Metrics:        metrics,
		},
	}
}

func makeHPA(objectMeta *metav1.ObjectMeta, minReplicas, maxReplicas int32, podPolicy v1alpha1.PodPolicy, targetRef autov2.CrossVersionObjectReference) *autov2.HorizontalPodAutoscaler {
	spec := autov2.HorizontalPodAutoscalerSpec{
		ScaleTargetRef: targetRef,
		MinReplicas:    &minReplicas,
		MaxReplicas:    maxReplicas,
		Metrics:        podPolicy.AutoScalingMetrics,
		Behavior:       podPolicy.AutoScalingBehavior,
	}
	return &autov2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "autoscaling/v2",
			Kind:       "HorizontalPodAutoscaler",
		},
		ObjectMeta: *objectMeta,
		Spec:       spec,
	}
}

func MakeHPA(objectMeta *metav1.ObjectMeta, targetRef autov2.CrossVersionObjectReference, minReplicas, maxReplicas *int32, policy v1alpha1.PodPolicy, res corev1.ResourceRequirements) *autov2.HorizontalPodAutoscaler {
	if isBuiltinHPAEnabled(minReplicas, maxReplicas, policy) {
		return makeBuiltinHPA(objectMeta, *minReplicas, *maxReplicas, targetRef, policy.BuiltinAutoscaler, res)
	} else if !isDefaultHPAEnabled(minReplicas, maxReplicas, policy) {
		return makeHPA(objectMeta, *minReplicas, *maxReplicas, policy, targetRef)
	}
	return makeDefaultHPA(objectMeta, *minReplicas, *maxReplicas, targetRef, res)
}

func getResourceCpuFactor(res corev1.ResourceRequirements) float64 {
	if res.Requests.Cpu() != nil && res.Limits.Cpu() != nil {
		if !res.Requests.Cpu().IsZero() && !res.Limits.Cpu().IsZero() {
			return float64(res.Limits.Cpu().MilliValue()) / float64(res.Requests.Cpu().MilliValue())
		}
	}
	return 1.0
}

func getResourceMemoryFactor(res corev1.ResourceRequirements) float64 {
	if res.Requests.Memory() != nil && res.Limits.Memory() != nil {
		if !res.Requests.Memory().IsZero() && !res.Limits.Memory().IsZero() {
			return float64(res.Limits.Memory().Value()) / float64(res.Requests.Memory().Value())
		}
	}
	return 1.0
}

func getUtilizationPercentage(val int32, factor float64) int32 {
	return int32(float64(val) * factor)
}
