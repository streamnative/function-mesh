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

// Package spec define the specs
package spec

import (
	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	autoscaling "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

func MakeVPA(objectMeta *metav1.ObjectMeta, targetRef *autoscaling.CrossVersionObjectReference, vpa *v1alpha1.VPASpec) *vpav1.VerticalPodAutoscaler {
	return &vpav1.VerticalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "autoscaling.k8s.io/v1",
			Kind:       "VerticalPodAutoscaler",
		},
		ObjectMeta: *objectMeta,
		Spec: vpav1.VerticalPodAutoscalerSpec{
			TargetRef:      targetRef,
			UpdatePolicy:   vpa.UpdatePolicy,
			ResourcePolicy: vpa.ResourcePolicy,
		},
	}
}
