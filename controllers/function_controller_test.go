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

package controllers

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"

	autov2beta2 "k8s.io/api/autoscaling/v2beta2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/streamnative/function-mesh/api/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/spec"
	v1 "k8s.io/api/core/v1"
)

const (
	timeout  = time.Second * 10
	interval = time.Millisecond * 250
)

var _ = Describe("Function Controller", func() {
	Context("Simple Function Item", func() {
		function := makeFunctionSample(TestFunctionName)

		createFunction(function)
	})
})

var _ = Describe("Function Controller (E2E)", func() {
	Context("Function With E2E Crypto Item", func() {
		function := makeFunctionSampleWithCryptoEnabled()

		createFunction(function)
	})
})

var _ = Describe("Function Controller (Batcher)", func() {
	Context("Function With Batcher Item", func() {
		function := makeFunctionSampleWithKeyBasedBatcher()

		createFunction(function)
	})
})

var _ = Describe("Function Controller (HPA)", func() {
	Context("Simple Function Item with HPA", func() {
		function := makeFunctionSample(TestFunctionName)
		cpuPercentage := int32(80)
		function.Spec.Pod.AutoScalingMetrics = []autov2beta2.MetricSpec{
			{
				Type: autov2beta2.ResourceMetricSourceType,
				Resource: &autov2beta2.ResourceMetricSource{
					Name: v1.ResourceCPU,
					Target: autov2beta2.MetricTarget{
						Type:               autov2beta2.UtilizationMetricType,
						AverageUtilization: &cpuPercentage,
					},
				},
			},
		}

		createFunction(function)
	})
})

func createFunction(function *v1alpha1.Function) {

	It("Function should be created", func() {
		Expect(k8sClient.Create(context.Background(), function)).Should(Succeed())
	})

	It("StatefulSet should be created", func() {
		statefulSet := &appsv1.StatefulSet{}

		Eventually(func() bool {
			err := k8sClient.Get(context.Background(), types.NamespacedName{
				Namespace: function.Namespace,
				Name:      spec.MakeFunctionObjectMeta(function).Name,
			}, statefulSet)
			if err != nil {
				return false
			}
			return true
		}, timeout, interval).Should(BeTrue())

		Expect(*statefulSet.Spec.Replicas).Should(Equal(int32(1)))
	})

	It("Service should be created", func() {
		srv := &v1.Service{}
		svcName := spec.MakeHeadlessServiceName(spec.MakeFunctionObjectMeta(function).Name)
		Eventually(func() bool {
			err := k8sClient.Get(context.Background(), types.NamespacedName{Namespace: function.Namespace,
				Name: svcName}, srv)
			if err != nil {
				return false
			}
			return true
		}, timeout, interval).Should(BeTrue())
	})

	It("HPA should be created", func() {
		if function.Spec.MaxReplicas != nil {
			hpa := &autov2beta2.HorizontalPodAutoscaler{}
			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), types.NamespacedName{Namespace: function.Namespace,
					Name: spec.MakeFunctionObjectMeta(function).Name}, hpa)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
		}
	})

	It("Function should be deleted", func() {
		Expect(k8sClient.Delete(context.Background(), function)).Should(Succeed())
	})
}
