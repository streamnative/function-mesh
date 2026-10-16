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
	"testing"

	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	"github.com/stretchr/testify/require"
	autov2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFixedReplicaHPADoesNotDriftAfterAPIDefaulting(t *testing.T) {
	replicas := int32(1)
	desired := MakeHPA(&metav1.ObjectMeta{Name: "fixed"}, autov2.CrossVersionObjectReference{
		APIVersion: "compute.functionmesh.io/v1alpha1", Kind: "Function", Name: "fixed",
	}, &replicas, &replicas, v1alpha1.PodPolicy{}, corev1.ResourceRequirements{})

	// Kubernetes supplies CPU utilization 80 when an HPA has no metrics.
	persisted := desired.DeepCopy()
	if len(persisted.Spec.Metrics) == 0 {
		utilization := int32(80)
		persisted.Spec.Metrics = []autov2.MetricSpec{{
			Type: autov2.ResourceMetricSourceType,
			Resource: &autov2.ResourceMetricSource{
				Name:   corev1.ResourceCPU,
				Target: autov2.MetricTarget{Type: autov2.UtilizationMetricType, AverageUtilization: &utilization},
			},
		}}
	}
	require.True(t, CheckIfHPASpecIsEqual(&persisted.Spec, &desired.Spec))
	require.Equal(t, int32(1), desired.Spec.MaxReplicas)
}
