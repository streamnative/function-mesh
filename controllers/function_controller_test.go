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

	autov2beta2 "k8s.io/api/autoscaling/v2beta2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/streamnative/function-mesh/api/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/spec"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Function Controller", func() {
	Context("Simple Function Item", func() {
		configs := makeSamplePulsarConfig()
		function := makeFunctionSample(TestFunctionName)

		createFunctionConfigMap(configs)
		createFunction(function)
		deleteFunction(function)
		deleteFunctionConfigMap(configs)
	})
})

var _ = Describe("Function Controller (E2E)", func() {
	Context("Function With E2E Crypto Item", func() {
		cryptosecrets := &v1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "java-function-crypto-sample-crypto-secret",
				Namespace: "default",
				UID:       "dead-beef-secret",
			},
			Data: map[string][]byte{
				"test_ecdsa_privkey.pem": []byte{0x00, 0x01, 0x02},
				"test_ecdsa_pubkey.pem":  []byte{0x02, 0x01, 0x00},
			},
			Type: "Opaque",
		}
		configs := makeSamplePulsarConfig()
		function := makeFunctionSampleWithCryptoEnabled()

		createFunctionConfigMap(configs)
		createFunctionSecret(cryptosecrets)
		createFunction(function)
		deleteFunction(function)
		deleteFunctionSecret(cryptosecrets)
		deleteFunctionConfigMap(configs)
	})
})

var _ = Describe("Function Controller (Batcher)", func() {
	Context("Function With Batcher Item", func() {
		configs := makeSamplePulsarConfig()
		function := makeFunctionSampleWithKeyBasedBatcher()

		createFunctionConfigMap(configs)
		createFunction(function)
		deleteFunction(function)
		deleteFunctionConfigMap(configs)
	})
})

var _ = Describe("Function Controller (HPA)", func() {
	Context("Simple Function Item with HPA", func() {
		configs := makeSamplePulsarConfig()
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

		createFunctionConfigMap(configs)
		createFunction(function)
		createFunctionHPA(function)
		deleteFunctionHPA(function)
		deleteFunction(function)
		deleteFunctionConfigMap(configs)
	})
})

func createFunction(function *v1alpha1.Function) {
	if function.Status.Conditions == nil {
		function.Status.Conditions = make(map[v1alpha1.Component]v1alpha1.ResourceCondition)
	}

	It("StatefulSet should be created", func() {
		statefulSet := spec.MakeFunctionStatefulSet(function)
		Expect(k8sClient.Create(context.Background(), statefulSet)).Should(Succeed())
	})
}

func createFunctionHPA(function *v1alpha1.Function) {
	if function.Status.Conditions == nil {
		function.Status.Conditions = make(map[v1alpha1.Component]v1alpha1.ResourceCondition)
	}

	It("HPA should be created", func() {
		hpa := spec.MakeFunctionHPA(function)
		Expect(k8sClient.Create(context.Background(), hpa)).Should(Succeed())
	})
}

func createFunctionConfigMap(configs *v1.ConfigMap) {
	It("Should create pulsar configmap successfully", func() {
		Expect(k8sClient.Create(context.Background(), configs)).Should(Succeed())
	})
}

func createFunctionSecret(secret *v1.Secret) {
	It("Should create crypto secret successfully", func() {
		Expect(k8sClient.Create(context.Background(), secret)).Should(Succeed())
	})
}

func deleteFunction(function *v1alpha1.Function) {
	It("StatefulSet should be deleted", func() {
		statefulSet := spec.MakeFunctionStatefulSet(function)
		Expect(k8sClient.Delete(context.Background(), statefulSet)).Should(Succeed())
	})
}

func deleteFunctionConfigMap(configs *v1.ConfigMap) {
	It("Should create pulsar configmap successfully", func() {
		Expect(k8sClient.Delete(context.Background(), configs)).Should(Succeed())
	})
}

func deleteFunctionSecret(secret *v1.Secret) {
	It("Should delete crypto secret successfully", func() {
		Expect(k8sClient.Delete(context.Background(), secret)).Should(Succeed())
	})
}

func deleteFunctionHPA(function *v1alpha1.Function) {
	hpa := spec.MakeFunctionHPA(function)
	It("Should delete HPA successfully", func() {
		Expect(k8sClient.Delete(context.Background(), hpa)).Should(Succeed())
	})
}
