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
	"github.com/stretchr/testify/assert"
	autoscaling "k8s.io/api/autoscaling/v1"
	autov2 "k8s.io/api/autoscaling/v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

func TestMakeVPA(t *testing.T) {
	mode := vpav1.UpdateModeAuto
	off := vpav1.UpdateModeOff
	containerMode := vpav1.ContainerScalingModeAuto
	containerOffMode := vpav1.ContainerScalingModeOff
	controlledValues := vpav1.ContainerControlledValuesRequestsAndLimits
	var minReplicas int32 = 3
	type args struct {
		objectMeta *metav1.ObjectMeta
		targetRef  *autov2.CrossVersionObjectReference
		vpa        *v1alpha1.VPASpec
	}
	tests := []struct {
		name string
		args args
		want *vpav1.VerticalPodAutoscaler
	}{
		{
			name: "Generate VPA successfully",
			args: args{
				objectMeta: &metav1.ObjectMeta{
					Name: "test-vpa",
					Labels: map[string]string{
						LabelComponent: "function",
					},
				},
				targetRef: &autov2.CrossVersionObjectReference{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "test-deployment",
				},
				vpa: &v1alpha1.VPASpec{
					UpdatePolicy: &vpav1.PodUpdatePolicy{
						UpdateMode: &mode,
					},
					ResourcePolicy: &vpav1.PodResourcePolicy{
						ContainerPolicies: []vpav1.ContainerResourcePolicy{{
							ContainerName: "pulsar-function",
							MinAllowed: v1.ResourceList{
								v1.ResourceCPU:    resource.MustParse("100m"),
								v1.ResourceMemory: resource.MustParse("100Mi"),
							},
							MaxAllowed: v1.ResourceList{
								v1.ResourceCPU:    resource.MustParse("1000m"),
								v1.ResourceMemory: resource.MustParse("1000Mi"),
							},
							ControlledResources: &[]v1.ResourceName{
								v1.ResourceCPU, v1.ResourceMemory,
							},
							ControlledValues: &controlledValues,
						}},
					},
				},
			},
			want: &vpav1.VerticalPodAutoscaler{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-vpa",
					Labels: map[string]string{
						LabelComponent: "function",
					},
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: "autoscaling.k8s.io/v1",
					Kind:       "VerticalPodAutoscaler",
				},
				Spec: vpav1.VerticalPodAutoscalerSpec{
					TargetRef: &autoscaling.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "test-deployment",
					},
					UpdatePolicy: &vpav1.PodUpdatePolicy{
						UpdateMode: &mode,
					},
					ResourcePolicy: &vpav1.PodResourcePolicy{
						ContainerPolicies: []vpav1.ContainerResourcePolicy{
							{
								ContainerName: "pulsar-function",
								MinAllowed: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("100m"),
									v1.ResourceMemory: resource.MustParse("100Mi"),
								},
								MaxAllowed: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("1000m"),
									v1.ResourceMemory: resource.MustParse("1000Mi"),
								},
								ControlledResources: &[]v1.ResourceName{
									v1.ResourceCPU, v1.ResourceMemory,
								},
								ControlledValues: &controlledValues,
								Mode:             &containerMode,
							},
						},
					},
				},
			},
		},
		{
			name: "Generate VPA without resourcePolicy successfully",
			args: args{
				objectMeta: &metav1.ObjectMeta{
					Name: "test-vpa",
					Labels: map[string]string{
						LabelComponent: "function",
					},
				},
				targetRef: &autov2.CrossVersionObjectReference{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "test-deployment",
				},
				vpa: &v1alpha1.VPASpec{
					UpdatePolicy: &vpav1.PodUpdatePolicy{
						UpdateMode: &mode,
					},
				},
			},
			want: &vpav1.VerticalPodAutoscaler{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-vpa",
					Labels: map[string]string{
						LabelComponent: "function",
					},
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: "autoscaling.k8s.io/v1",
					Kind:       "VerticalPodAutoscaler",
				},
				Spec: vpav1.VerticalPodAutoscalerSpec{
					TargetRef: &autoscaling.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "test-deployment",
					},
					UpdatePolicy: &vpav1.PodUpdatePolicy{
						UpdateMode: &mode,
					},
					ResourcePolicy: &vpav1.PodResourcePolicy{
						ContainerPolicies: []vpav1.ContainerResourcePolicy{
							{
								ContainerName: "pulsar-function",
								Mode:          &containerMode,
							},
						},
					},
				},
			},
		},
		{
			name: "Generate VPA without updatePolicy successfully",
			args: args{
				objectMeta: &metav1.ObjectMeta{
					Name: "test-vpa",
					Labels: map[string]string{
						LabelComponent: "sink",
					},
				},
				targetRef: &autov2.CrossVersionObjectReference{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "test-deployment",
				},
				vpa: &v1alpha1.VPASpec{
					ResourcePolicy: &vpav1.PodResourcePolicy{
						ContainerPolicies: []vpav1.ContainerResourcePolicy{{
							ContainerName: "pulsar-sink",
							MinAllowed: v1.ResourceList{
								v1.ResourceCPU:    resource.MustParse("100m"),
								v1.ResourceMemory: resource.MustParse("100Mi"),
							},
							MaxAllowed: v1.ResourceList{
								v1.ResourceCPU:    resource.MustParse("1000m"),
								v1.ResourceMemory: resource.MustParse("1000Mi"),
							},
							ControlledResources: &[]v1.ResourceName{
								v1.ResourceCPU, v1.ResourceMemory,
							},
							ControlledValues: &controlledValues,
						},
						},
					},
				},
			},
			want: &vpav1.VerticalPodAutoscaler{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-vpa",
					Labels: map[string]string{
						LabelComponent: "sink",
					},
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: "autoscaling.k8s.io/v1",
					Kind:       "VerticalPodAutoscaler",
				},
				Spec: vpav1.VerticalPodAutoscalerSpec{
					TargetRef: &autoscaling.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "test-deployment",
					},
					UpdatePolicy: nil,
					ResourcePolicy: &vpav1.PodResourcePolicy{
						ContainerPolicies: []vpav1.ContainerResourcePolicy{
							{
								ContainerName: "pulsar-sink",
								MinAllowed: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("100m"),
									v1.ResourceMemory: resource.MustParse("100Mi"),
								},
								MaxAllowed: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("1000m"),
									v1.ResourceMemory: resource.MustParse("1000Mi"),
								},
								ControlledResources: &[]v1.ResourceName{
									v1.ResourceCPU, v1.ResourceMemory,
								},
								ControlledValues: &controlledValues,
								Mode:             &containerMode,
							},
						},
					},
				},
			},
		},
		{
			name: "Generate VPA without updatePolicy and resourcePolicy successfully",
			args: args{
				objectMeta: &metav1.ObjectMeta{
					Name: "test-vpa",
					Labels: map[string]string{
						LabelComponent: "source",
					},
				},
				targetRef: &autov2.CrossVersionObjectReference{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "test-deployment",
				},
				vpa: &v1alpha1.VPASpec{},
			},
			want: &vpav1.VerticalPodAutoscaler{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-vpa",
					Labels: map[string]string{
						LabelComponent: "source",
					},
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: "autoscaling.k8s.io/v1",
					Kind:       "VerticalPodAutoscaler",
				},
				Spec: vpav1.VerticalPodAutoscalerSpec{
					TargetRef: &autoscaling.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "test-deployment",
					},
					UpdatePolicy: nil,
					ResourcePolicy: &vpav1.PodResourcePolicy{
						ContainerPolicies: []vpav1.ContainerResourcePolicy{
							{
								ContainerName: "pulsar-source",
								Mode:          &containerMode,
							},
						},
					},
				},
			},
		},
		{
			name: "Generate VPA with resource unit successfully",
			args: args{
				objectMeta: &metav1.ObjectMeta{
					Name: "test-vpa",
					Labels: map[string]string{
						LabelComponent: "source",
					},
				},
				targetRef: &autov2.CrossVersionObjectReference{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "test-deployment",
				},
				vpa: &v1alpha1.VPASpec{
					ResourceUnit: &v1alpha1.ResourceUnit{
						CPU:    resource.MustParse("200m"),
						Memory: resource.MustParse("800Mi"),
					},
				},
			},
			want: &vpav1.VerticalPodAutoscaler{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-vpa",
					Labels: map[string]string{
						LabelComponent:          "source",
						LabelCustomResourceUnit: "true",
					},
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: "autoscaling.k8s.io/v1",
					Kind:       "VerticalPodAutoscaler",
				},
				Spec: vpav1.VerticalPodAutoscalerSpec{
					TargetRef: &autoscaling.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "test-deployment",
					},
					UpdatePolicy: &vpav1.PodUpdatePolicy{
						UpdateMode: &off,
					},
					ResourcePolicy: &vpav1.PodResourcePolicy{
						ContainerPolicies: []vpav1.ContainerResourcePolicy{
							{
								ContainerName: "pulsar-source",
								Mode:          &containerMode,
							},
						},
					},
				},
			},
		},
		{
			name: "Generate VPA with resource unit and resource policy successfully",
			args: args{
				objectMeta: &metav1.ObjectMeta{
					Name: "test-vpa",
					Labels: map[string]string{
						LabelComponent: "source",
					},
				},
				targetRef: &autov2.CrossVersionObjectReference{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "test-deployment",
				},
				vpa: &v1alpha1.VPASpec{
					ResourceUnit: &v1alpha1.ResourceUnit{
						CPU:    resource.MustParse("200m"),
						Memory: resource.MustParse("800Mi"),
					},
					UpdatePolicy: &vpav1.PodUpdatePolicy{
						UpdateMode:  &mode, // should be set to 'Off'
						MinReplicas: &minReplicas,
					},
					ResourcePolicy: &vpav1.PodResourcePolicy{
						ContainerPolicies: []vpav1.ContainerResourcePolicy{
							{
								ContainerName: "pulsar-source",
								MinAllowed: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("100m"),
									v1.ResourceMemory: resource.MustParse("100Mi"),
								},
								MaxAllowed: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("1000m"),
									v1.ResourceMemory: resource.MustParse("1000Mi"),
								},
								Mode:                &containerOffMode, // should be converted to 'Auto'
								ControlledValues:    &controlledValues,
								ControlledResources: &[]v1.ResourceName{v1.ResourceCPU},
							},
							{
								ContainerName: "istio", // should be ignored
								Mode:          &containerMode,
							},
						},
					},
				},
			},
			want: &vpav1.VerticalPodAutoscaler{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-vpa",
					Labels: map[string]string{
						LabelComponent:          "source",
						LabelCustomResourceUnit: "true",
					},
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: "autoscaling.k8s.io/v1",
					Kind:       "VerticalPodAutoscaler",
				},
				Spec: vpav1.VerticalPodAutoscalerSpec{
					TargetRef: &autoscaling.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "test-deployment",
					},
					UpdatePolicy: &vpav1.PodUpdatePolicy{
						UpdateMode:  &off,
						MinReplicas: &minReplicas,
					},
					ResourcePolicy: &vpav1.PodResourcePolicy{
						ContainerPolicies: []vpav1.ContainerResourcePolicy{
							{
								ContainerName: "pulsar-source",
								Mode:          &containerMode,
								MinAllowed: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("100m"),
									v1.ResourceMemory: resource.MustParse("100Mi"),
								},
								MaxAllowed: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("1000m"),
									v1.ResourceMemory: resource.MustParse("1000Mi"),
								},
								ControlledValues:    &controlledValues,
								ControlledResources: &[]v1.ResourceName{v1.ResourceCPU},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, MakeVPA(tt.args.objectMeta, tt.args.targetRef, tt.args.vpa), "MakeVPA(%v, %v, %v)", tt.args.objectMeta, tt.args.targetRef, tt.args.vpa)
		})
	}
}
