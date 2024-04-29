/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"reflect"
	"testing"

	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

func Test_calculateVPARecommendation(t *testing.T) {
	type args struct {
		vpa     *vpav1.VerticalPodAutoscaler
		vpaSpec *v1alpha1.VPASpec
	}
	requestsOnly := vpav1.ContainerControlledValuesRequestsOnly
	tests := []struct {
		name string
		args args
		want *corev1.ResourceRequirements
	}{
		{
			name: "Use the target cpu when target cpu is equal to the resource unit cpu",
			args: args{
				vpa: &vpav1.VerticalPodAutoscaler{
					Status: vpav1.VerticalPodAutoscalerStatus{
						Recommendation: &vpav1.RecommendedPodResources{
							ContainerRecommendations: []vpav1.RecommendedContainerResources{
								{
									Target: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("0.2"),
										corev1.ResourceMemory: resource.MustParse("1Gi"),
									},
								},
							},
						},
					},
				},
				vpaSpec: &v1alpha1.VPASpec{
					ResourceUnit: &v1alpha1.ResourceUnit{
						Cpu:    resource.MustParse("200m"),
						Memory: resource.MustParse("800Mi"),
					},
				},
			},
			want: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewScaledQuantity(200, resource.Milli),
					corev1.ResourceMemory: *resource.NewScaledQuantity(800*1024*1024*1000, resource.Milli),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewScaledQuantity(200, resource.Milli),
					corev1.ResourceMemory: *resource.NewScaledQuantity(800*1024*1024*1000, resource.Milli),
				},
			},
		},
		{
			name: "Increase to resource unit when target cpu is smaller than the resource unit cpu",
			args: args{
				vpa: &vpav1.VerticalPodAutoscaler{
					Status: vpav1.VerticalPodAutoscalerStatus{
						Recommendation: &vpav1.RecommendedPodResources{
							ContainerRecommendations: []vpav1.RecommendedContainerResources{
								{
									Target: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("100m"),
										corev1.ResourceMemory: resource.MustParse("1Gi"),
									},
								},
							},
						},
					},
				},
				vpaSpec: &v1alpha1.VPASpec{
					ResourceUnit: &v1alpha1.ResourceUnit{
						Cpu:    resource.MustParse("0.2"),       // 200m
						Memory: resource.MustParse("838860800"), // 800Mi
					},
				},
			},
			want: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewScaledQuantity(200, resource.Milli),
					corev1.ResourceMemory: *resource.NewScaledQuantity(800*1024*1024*1000, resource.Milli),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewScaledQuantity(200, resource.Milli),
					corev1.ResourceMemory: *resource.NewScaledQuantity(800*1024*1024*1000, resource.Milli),
				},
			},
		},
		{
			name: "Use the target cpu when target cpu can evenly divide by the resource unit cpu",
			args: args{
				vpa: &vpav1.VerticalPodAutoscaler{
					Status: vpav1.VerticalPodAutoscalerStatus{
						Recommendation: &vpav1.RecommendedPodResources{
							ContainerRecommendations: []vpav1.RecommendedContainerResources{
								{
									Target: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("400m"),
										corev1.ResourceMemory: resource.MustParse("1Gi"),
									},
								},
							},
						},
					},
				},
				vpaSpec: &v1alpha1.VPASpec{
					ResourceUnit: &v1alpha1.ResourceUnit{
						Cpu:    resource.MustParse("0.2"),
						Memory: resource.MustParse("838860800"),
					},
				},
			},
			want: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewScaledQuantity(400, resource.Milli),
					corev1.ResourceMemory: *resource.NewScaledQuantity(2*800*1024*1024*1000, resource.Milli),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewScaledQuantity(400, resource.Milli),
					corev1.ResourceMemory: *resource.NewScaledQuantity(2*800*1024*1024*1000, resource.Milli),
				},
			},
		},
		{
			name: "Increase to the smallest multiples of resource unit cpu when target cpu cannot evenly divided by the resource unit cpu",
			args: args{
				vpa: &vpav1.VerticalPodAutoscaler{
					Status: vpav1.VerticalPodAutoscalerStatus{
						Recommendation: &vpav1.RecommendedPodResources{
							ContainerRecommendations: []vpav1.RecommendedContainerResources{
								{
									Target: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("0.45"),
										corev1.ResourceMemory: resource.MustParse("1Gi"),
									},
								},
							},
						},
					},
				},
				vpaSpec: &v1alpha1.VPASpec{
					ResourceUnit: &v1alpha1.ResourceUnit{
						Cpu:    resource.MustParse("200m"),
						Memory: resource.MustParse("800Mi"),
					},
				},
			},
			want: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewScaledQuantity(600, resource.Milli),
					corev1.ResourceMemory: *resource.NewScaledQuantity(3*800*1024*1024*1000, resource.Milli),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewScaledQuantity(600, resource.Milli),
					corev1.ResourceMemory: *resource.NewScaledQuantity(3*800*1024*1024*1000, resource.Milli),
				},
			},
		},
		{
			name: "Use the target memory when target memory is equal to the resource unit memory",
			args: args{
				vpa: &vpav1.VerticalPodAutoscaler{
					Status: vpav1.VerticalPodAutoscalerStatus{
						Recommendation: &vpav1.RecommendedPodResources{
							ContainerRecommendations: []vpav1.RecommendedContainerResources{
								{
									Target: corev1.ResourceList{
										corev1.ResourceMemory: resource.MustParse("800Mi"),
									},
								},
							},
						},
					},
				},
				vpaSpec: &v1alpha1.VPASpec{
					ResourceUnit: &v1alpha1.ResourceUnit{
						Cpu:    resource.MustParse("200m"),
						Memory: resource.MustParse("838860800"),
					},
				},
			},
			want: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewScaledQuantity(200, resource.Milli),
					corev1.ResourceMemory: *resource.NewScaledQuantity(800*1024*1024*1000, resource.Milli),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewScaledQuantity(200, resource.Milli),
					corev1.ResourceMemory: *resource.NewScaledQuantity(800*1024*1024*1000, resource.Milli),
				},
			},
		},
		{
			name: "Increase to resource unit when target memory is smaller than the resource unit memory",
			args: args{
				vpa: &vpav1.VerticalPodAutoscaler{
					Status: vpav1.VerticalPodAutoscalerStatus{
						Recommendation: &vpav1.RecommendedPodResources{
							ContainerRecommendations: []vpav1.RecommendedContainerResources{
								{
									Target: corev1.ResourceList{
										corev1.ResourceMemory: resource.MustParse("629145600"), // 600Mi
									},
								},
							},
						},
					},
				},
				vpaSpec: &v1alpha1.VPASpec{
					ResourceUnit: &v1alpha1.ResourceUnit{
						Cpu:    resource.MustParse("200m"),
						Memory: resource.MustParse("800Mi"),
					},
				},
			},
			want: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewScaledQuantity(200, resource.Milli),
					corev1.ResourceMemory: *resource.NewScaledQuantity(800*1024*1024*1000, resource.Milli),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewScaledQuantity(200, resource.Milli),
					corev1.ResourceMemory: *resource.NewScaledQuantity(800*1024*1024*1000, resource.Milli),
				},
			},
		},
		{
			name: "Use the target memory when target memory can evenly divide by the resource unit memory",
			args: args{
				vpa: &vpav1.VerticalPodAutoscaler{
					Status: vpav1.VerticalPodAutoscalerStatus{
						Recommendation: &vpav1.RecommendedPodResources{
							ContainerRecommendations: []vpav1.RecommendedContainerResources{
								{
									Target: corev1.ResourceList{
										corev1.ResourceMemory: resource.MustParse("1600Mi"),
									},
								},
							},
						},
					},
				},
				vpaSpec: &v1alpha1.VPASpec{
					ResourceUnit: &v1alpha1.ResourceUnit{
						Cpu:    resource.MustParse("200m"),
						Memory: resource.MustParse("800Mi"),
					},
				},
			},
			want: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewScaledQuantity(400, resource.Milli),
					corev1.ResourceMemory: *resource.NewScaledQuantity(2*800*1024*1024*1000, resource.Milli),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewScaledQuantity(400, resource.Milli),
					corev1.ResourceMemory: *resource.NewScaledQuantity(2*800*1024*1024*1000, resource.Milli),
				},
			},
		},
		{
			name: "Increase to the smallest multiples of resource unit memory when target memory cannot evenly divided by the resource unit memory",
			args: args{
				vpa: &vpav1.VerticalPodAutoscaler{
					Status: vpav1.VerticalPodAutoscalerStatus{
						Recommendation: &vpav1.RecommendedPodResources{
							ContainerRecommendations: []vpav1.RecommendedContainerResources{
								{
									Target: corev1.ResourceList{
										corev1.ResourceMemory: resource.MustParse("1.7Gi"),
									},
								},
							},
						},
					},
				},
				vpaSpec: &v1alpha1.VPASpec{
					ResourceUnit: &v1alpha1.ResourceUnit{
						Cpu:    resource.MustParse("200m"),
						Memory: resource.MustParse("800Mi"),
					},
				},
			},
			want: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewScaledQuantity(600, resource.Milli),
					corev1.ResourceMemory: *resource.NewScaledQuantity(3*800*1024*1024*1000, resource.Milli),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewScaledQuantity(600, resource.Milli),
					corev1.ResourceMemory: *resource.NewScaledQuantity(3*800*1024*1024*1000, resource.Milli),
				},
			},
		},
		{
			name: "Only set request resources when it set requestsOnly",
			args: args{
				vpa: &vpav1.VerticalPodAutoscaler{
					Status: vpav1.VerticalPodAutoscalerStatus{
						Recommendation: &vpav1.RecommendedPodResources{
							ContainerRecommendations: []vpav1.RecommendedContainerResources{
								{
									Target: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("200m"),
										corev1.ResourceMemory: resource.MustParse("1Gi"),
									},
								},
							},
						},
					},
					Spec: vpav1.VerticalPodAutoscalerSpec{
						ResourcePolicy: &vpav1.PodResourcePolicy{
							ContainerPolicies: []vpav1.ContainerResourcePolicy{
								{
									ContainerName:    "pulsar-function",
									ControlledValues: &requestsOnly,
								},
							},
						},
					},
				},
				vpaSpec: &v1alpha1.VPASpec{
					ResourceUnit: &v1alpha1.ResourceUnit{
						Cpu:    resource.MustParse("200m"),
						Memory: resource.MustParse("800Mi"),
					},
				},
			},
			want: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewScaledQuantity(200, resource.Milli),
					corev1.ResourceMemory: *resource.NewScaledQuantity(800*1024*1024*1000, resource.Milli),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateVPARecommendation(tt.args.vpa, tt.args.vpaSpec); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateVPARecommendation() = %v, want %v", got, tt.want)
			}
		})
	}
}
