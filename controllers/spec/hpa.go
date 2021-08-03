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
	"github.com/streamnative/function-mesh/api/v1alpha1"
	autov2beta2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func isDefaultHPAEnabled(minReplicas, maxReplicas *int32, podPolicy v1alpha1.PodPolicy) bool {
	return minReplicas != nil && maxReplicas != nil && podPolicy.AutoScalingBehavior == nil && len(podPolicy.AutoScalingMetrics) == 0 && *maxReplicas > *minReplicas
}

// defaultHPAMetrics generates a default HPA metrics settings based on CPU usage and utilized on 80%.
func defaultHPAMetrics() []autov2beta2.MetricSpec {
	// TODO: configurable cpu percentage
	cpuPercentage := int32(80)
	return []autov2beta2.MetricSpec{
		{
			Type: autov2beta2.ResourceMetricSourceType,
			Resource: &autov2beta2.ResourceMetricSource{
				Name: corev1.ResourceCPU,
				Target: autov2beta2.MetricTarget{
					Type:               autov2beta2.UtilizationMetricType,
					AverageUtilization: &cpuPercentage,
				},
			},
		},
	}
}

func makeDefaultHPA(objectMeta *metav1.ObjectMeta, minReplicas, maxReplicas int32) *autov2beta2.HorizontalPodAutoscaler {
	return &autov2beta2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "autoscaling/v2beta2",
			Kind:       "HorizontalPodAutoscaler",
		},
		ObjectMeta: *objectMeta,
		Spec: autov2beta2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autov2beta2.CrossVersionObjectReference{
				Kind:       "StatefulSet",
				Name:       objectMeta.Name,
				APIVersion: "apps/v1",
			},
			MinReplicas: &minReplicas,
			MaxReplicas: maxReplicas,
			Metrics:     defaultHPAMetrics(),
		},
	}
}

func makeHPA(objectMeta *metav1.ObjectMeta, minReplicas, maxReplicas int32, podPolicy v1alpha1.PodPolicy) *autov2beta2.HorizontalPodAutoscaler {
	spec := autov2beta2.HorizontalPodAutoscalerSpec{
		ScaleTargetRef: autov2beta2.CrossVersionObjectReference{
			Kind:       "StatefulSet",
			Name:       objectMeta.Name,
			APIVersion: "apps/v1",
		},
		MinReplicas: &minReplicas,
		MaxReplicas: maxReplicas,
		Metrics:     podPolicy.AutoScalingMetrics,
		Behavior:    podPolicy.AutoScalingBehavior,
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
