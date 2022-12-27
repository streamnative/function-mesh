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
	autov2beta2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BuiltinAutoScaler interface {
	Metrics() []autov2beta2.MetricSpec
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

func (H *HPARuleAverageUtilizationResourceMemoryPercent) Metrics() []autov2beta2.MetricSpec {
	return []autov2beta2.MetricSpec{
		{
			Type: autov2beta2.ResourceMetricSourceType,
			Resource: &autov2beta2.ResourceMetricSource{
				Name: corev1.ResourceMemory,
				Target: autov2beta2.MetricTarget{
					Type:               autov2beta2.UtilizationMetricType,
					AverageUtilization: &H.memoryPercentage,
				},
			},
		},
	}
}

func (H *HPARuleAverageUtilizationCPUPercent) Metrics() []autov2beta2.MetricSpec {
	return []autov2beta2.MetricSpec{
		{
			Type: autov2beta2.ResourceMetricSourceType,
			Resource: &autov2beta2.ResourceMetricSource{
				Name: corev1.ResourceCPU,
				Target: autov2beta2.MetricTarget{
					Type:               autov2beta2.UtilizationMetricType,
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

func GetBuiltinAutoScaler(builtinRule v1alpha1.BuiltinHPARule) (BuiltinAutoScaler, int) {
	switch builtinRule {
	case v1alpha1.AverageUtilizationCPUPercent80:
		return NewHPARuleAverageUtilizationCPUPercent(80), cpuRuleIdx
	case v1alpha1.AverageUtilizationCPUPercent50:
		return NewHPARuleAverageUtilizationCPUPercent(50), cpuRuleIdx
	case v1alpha1.AverageUtilizationCPUPercent20:
		return NewHPARuleAverageUtilizationCPUPercent(20), cpuRuleIdx
	case v1alpha1.AverageUtilizationMemoryPercent80:
		return NewHPARuleAverageUtilizationMemoryPercent(80), memoryRuleIdx
	case v1alpha1.AverageUtilizationMemoryPercent50:
		return NewHPARuleAverageUtilizationMemoryPercent(50), memoryRuleIdx
	case v1alpha1.AverageUtilizationMemoryPercent20:
		return NewHPARuleAverageUtilizationMemoryPercent(20), memoryRuleIdx
	default:
		return nil, 2
	}
}

// defaultHPAMetrics generates a default HPA Metrics settings based on CPU usage and utilized on 80%.
func defaultHPAMetrics() []autov2beta2.MetricSpec {
	return NewHPARuleAverageUtilizationCPUPercent(80).Metrics()
}

func makeDefaultHPA(objectMeta *metav1.ObjectMeta, minReplicas, maxReplicas int32, targetRef autov2beta2.CrossVersionObjectReference) *autov2beta2.HorizontalPodAutoscaler {
	return &autov2beta2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "autoscaling/v2beta2",
			Kind:       "HorizontalPodAutoscaler",
		},
		ObjectMeta: *objectMeta,
		Spec: autov2beta2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: targetRef,
			MinReplicas:    &minReplicas,
			MaxReplicas:    maxReplicas,
			Metrics:        defaultHPAMetrics(),
		},
	}
}

const (
	cpuRuleIdx = iota
	memoryRuleIdx
)

func MakeMetricsFromBuiltinHPARules(builtinRules []v1alpha1.BuiltinHPARule) []autov2beta2.MetricSpec {
	isRuleExists := map[int]bool{}
	metrics := []autov2beta2.MetricSpec{}
	for _, r := range builtinRules {
		s, idx := GetBuiltinAutoScaler(r)
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

func makeBuiltinHPA(objectMeta *metav1.ObjectMeta, minReplicas, maxReplicas int32, targetRef autov2beta2.CrossVersionObjectReference, builtinRules []v1alpha1.BuiltinHPARule) *autov2beta2.HorizontalPodAutoscaler {
	metrics := MakeMetricsFromBuiltinHPARules(builtinRules)
	return &autov2beta2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "autoscaling/v2beta2",
			Kind:       "HorizontalPodAutoscaler",
		},
		ObjectMeta: *objectMeta,
		Spec: autov2beta2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: targetRef,
			MinReplicas:    &minReplicas,
			MaxReplicas:    maxReplicas,
			Metrics:        metrics,
		},
	}
}

func makeHPA(objectMeta *metav1.ObjectMeta, minReplicas, maxReplicas int32, podPolicy v1alpha1.PodPolicy, targetRef autov2beta2.CrossVersionObjectReference) *autov2beta2.HorizontalPodAutoscaler {
	spec := autov2beta2.HorizontalPodAutoscalerSpec{
		ScaleTargetRef: targetRef,
		MinReplicas:    &minReplicas,
		MaxReplicas:    maxReplicas,
		Metrics:        podPolicy.AutoScalingMetrics,
		Behavior:       podPolicy.AutoScalingBehavior,
	}
	return &autov2beta2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "autoscaling/v2beta2",
			Kind:       "HorizontalPodAutoscaler",
		},
		ObjectMeta: *objectMeta,
		Spec:       spec,
	}
}
