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
	autov2beta2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
)

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
