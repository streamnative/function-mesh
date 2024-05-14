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
	autov2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

func MakeVPA(objectMeta *metav1.ObjectMeta, targetRef *autov2.CrossVersionObjectReference, vpa *v1alpha1.VPASpec) *vpav1.VerticalPodAutoscaler {
	containerName := GetVPAContainerName(objectMeta)
	updatePolicy := UpdateVPAUpdatePolicy(vpa.UpdatePolicy, vpa.ResourceUnit)
	if vpa.ResourceUnit != nil {
		objectMeta.Labels[LabelCustomResourceUnit] = "true"
	}
	resourcePolicy := UpdateResourcePolicy(vpa.ResourcePolicy, containerName)

	return &vpav1.VerticalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "autoscaling.k8s.io/v1",
			Kind:       "VerticalPodAutoscaler",
		},
		ObjectMeta: *objectMeta,
		Spec: vpav1.VerticalPodAutoscalerSpec{
			TargetRef: &autoscaling.CrossVersionObjectReference{
				Kind:       targetRef.Kind,
				Name:       targetRef.Name,
				APIVersion: targetRef.APIVersion,
			},
			UpdatePolicy:   updatePolicy,
			ResourcePolicy: resourcePolicy,
		},
	}
}

func GetVPAContainerName(objectMeta *metav1.ObjectMeta) string {
	containerName := "*"
	label := objectMeta.GetLabels()[LabelComponent]
	if label == ComponentFunction {
		containerName = FunctionContainerName
	} else if label == ComponentSink {
		containerName = SinkContainerName
	} else if label == ComponentSource {
		containerName = SourceContainerName
	}
	return containerName
}

func UpdateVPAUpdatePolicy(updatePolicy *vpav1.PodUpdatePolicy, resourceUnit *v1alpha1.ResourceUnit) *vpav1.PodUpdatePolicy {
	resultUpdatePolicy := updatePolicy
	if resourceUnit != nil {
		if resultUpdatePolicy == nil {
			resultUpdatePolicy = &vpav1.PodUpdatePolicy{}
		}
		off := vpav1.UpdateModeOff
		resultUpdatePolicy.UpdateMode = &off
	}
	return resultUpdatePolicy
}

func UpdateResourcePolicy(resourcePolicy *vpav1.PodResourcePolicy, containerName string) *vpav1.PodResourcePolicy {
	var containerPolicies []vpav1.ContainerResourcePolicy
	containerScalingMode := vpav1.ContainerScalingModeAuto
	if resourcePolicy != nil {
		if resourcePolicy.ContainerPolicies == nil {
			containerPolicies = []vpav1.ContainerResourcePolicy{}
		}
		for _, policy := range resourcePolicy.ContainerPolicies {
			if policy.ContainerName == containerName {
				policy.Mode = &containerScalingMode
				containerPolicies = []vpav1.ContainerResourcePolicy{policy}
				break
			}
		}
	}

	// if resource policy is not set, set the default policy, so the vpa policy won't be applied to other containers
	if len(containerPolicies) == 0 {
		containerPolicies = []vpav1.ContainerResourcePolicy{
			{
				ContainerName: containerName,
				Mode:          &containerScalingMode,
			},
		}
	}
	resultResourcePolicy := resourcePolicy
	if resultResourcePolicy == nil {
		resultResourcePolicy = &vpav1.PodResourcePolicy{}
	}
	resultResourcePolicy.ContainerPolicies = containerPolicies
	return resultResourcePolicy
}
