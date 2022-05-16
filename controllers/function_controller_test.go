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
	"fmt"
	"regexp"
	"strings"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	autov2beta2 "k8s.io/api/autoscaling/v2beta2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/streamnative/function-mesh/api/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/spec"
	v1 "k8s.io/api/core/v1"
)

var log = logf.Log.WithName("function-resource-test")

const (
	timeout  = time.Second * 10
	interval = time.Millisecond * 250
)

var _ = Describe("Function Controller", func() {
	Context("Simple Function Item", func() {
		function := makeFunctionSample(TestFunctionName)

		createFunction(function)
	})

	Context("FunctionConfig", func() {
		function := makeFunctionSample(fmt.Sprintf("%s-streamnative", TestFunctionName))

		It("Should update function statefulset container command", func() {
			err := k8sClient.Create(context.Background(), function)
			Expect(err).NotTo(HaveOccurred())
			functionSts := &appsv1.StatefulSet{}

			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      spec.MakeFunctionObjectMeta(function).Name,
					Namespace: spec.MakeFunctionObjectMeta(function).Namespace,
				}, functionSts)
				return err == nil && len(functionSts.Spec.Template.Spec.Containers[0].Command[2]) > 0
			}, 10*time.Second, 1*time.Second).Should(BeTrue())
			err = k8sClient.Get(context.Background(), types.NamespacedName{
				Name:      fmt.Sprintf("%s-streamnative", TestFunctionName),
				Namespace: TestNameSpace,
			}, function)
			Expect(err).NotTo(HaveOccurred())

			oldFunctionConfig := &v1alpha1.Config{}
			oldFunctionConfig.Data = map[string]interface{}{
				"configkey1": "configvalue1",
				"configkey2": "configvalue2",
				"configkey3": "configvalue3",
			}
			function.Spec.FuncConfig = oldFunctionConfig
			err = k8sClient.Update(context.Background(), function)
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() bool {
				funcReconciler.Reconcile(
					ctrl.Request{
						NamespacedName: types.NamespacedName{
							Name:      fmt.Sprintf("%s-streamnative", TestFunctionName),
							Namespace: TestNameSpace,
						},
					})
				err := k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      spec.MakeFunctionObjectMeta(function).Name,
					Namespace: spec.MakeFunctionObjectMeta(function).Namespace,
				}, functionSts)
				return err == nil && functionSts.ObjectMeta.Generation > 1
			}, 10*time.Second, 1*time.Second).Should(BeTrue())
			re := regexp.MustCompile("{\"configkey1\":\"configvalue1\",\"configkey2\":\"configvalue2\",\"configkey3\":\"configvalue3\"}")
			// Verify new config synced to pod spec
			Expect(len(re.FindAllString(strings.ReplaceAll(functionSts.Spec.Template.Spec.Containers[0].Command[2],
				"\\", ""), -1))).To(Equal(1))
		})
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
		function := makeFunctionSample(TestFunctionHPAName)
		cpuPercentage := int32(20)
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
			{
				Type: autov2beta2.ResourceMetricSourceType,
				Resource: &autov2beta2.ResourceMetricSource{
					Name: v1.ResourceMemory,
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

var _ = Describe("Function Controller (builtin HPA)", func() {
	Context("Simple Function Item with builtin HPA", func() {
		r := int32(100)
		function := makeFunctionSample(TestFunctionBuiltinHPAName)
		function.Spec.MaxReplicas = &r
		function.Spec.Pod.BuiltinAutoscaler = []v1alpha1.BuiltinHPARule{
			v1alpha1.AverageUtilizationCPUPercent20,
			v1alpha1.AverageUtilizationMemoryPercent20,
		}

		createFunction(function)
	})
})

var _ = Describe("Function Controller (Stateful Function)", func() {
	Context("Simple Function Item with builtin HPA", func() {
		r := int32(100)
		function := makeFunctionSample(TestFunctionBuiltinHPAName)
		function.Spec.MaxReplicas = &r
		function.Spec.Pod.BuiltinAutoscaler = []v1alpha1.BuiltinHPARule{
			v1alpha1.AverageUtilizationCPUPercent20,
			v1alpha1.AverageUtilizationMemoryPercent20,
		}

		createFunction(function)
	})
})

func createFunction(function *v1alpha1.Function) {

	It("Function should be created", func() {
		Eventually(func() bool {
			err := k8sClient.Create(context.Background(), function)
			return err == nil
		}, timeout, interval).Should(BeTrue())
	})

	It("StatefulSet should be created", func() {
		statefulSet := &appsv1.StatefulSet{}

		Eventually(func() bool {
			err := k8sClient.Get(context.Background(), types.NamespacedName{
				Namespace: function.Namespace,
				Name:      spec.MakeFunctionObjectMeta(function).Name,
			}, statefulSet)
			return err == nil
		}, timeout, interval).Should(BeTrue())

		Expect(statefulSet.Name).Should(Equal(spec.MakeFunctionObjectMeta(function).Name))
		Expect(*statefulSet.Spec.Replicas).Should(Equal(int32(1)))
	})

	It("Service should be created", func() {
		srv := &v1.Service{}
		svcName := spec.MakeHeadlessServiceName(spec.MakeFunctionObjectMeta(function).Name)
		Eventually(func() bool {
			err := k8sClient.Get(context.Background(), types.NamespacedName{Namespace: function.Namespace,
				Name: svcName}, srv)
			return err == nil
		}, timeout, interval).Should(BeTrue())
	})

	It("HPA should be created", func() {
		if function.Spec.MaxReplicas != nil {
			hpa := &autov2beta2.HorizontalPodAutoscaler{}
			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), types.NamespacedName{Namespace: function.Namespace,
					Name: spec.MakeFunctionObjectMeta(function).Name}, hpa)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			log.Info("HPA should be created", "hpa", hpa, "name", spec.MakeFunctionObjectMeta(function).Name)

			if len(function.Spec.Pod.AutoScalingMetrics) > 0 {
				Expect(len(hpa.Spec.Metrics)).Should(Equal(len(function.Spec.Pod.AutoScalingMetrics)))
				for _, metric := range function.Spec.Pod.AutoScalingMetrics {
					Expect(hpa.Spec.Metrics).Should(ContainElement(metric))
				}
			}

			if len(function.Spec.Pod.BuiltinAutoscaler) > 0 {
				Expect(len(hpa.Spec.Metrics)).Should(Equal(len(function.Spec.Pod.BuiltinAutoscaler)))
				for _, rule := range function.Spec.Pod.BuiltinAutoscaler {
					autoscaler := spec.GetBuiltinAutoScaler(rule)
					Expect(autoscaler).Should(Not(BeNil()))
					Expect(hpa.Spec.Metrics).Should(ContainElement(autoscaler.Metrics()[0]))
				}
			}
		}
	})

	It("Function should be deleted", func() {
		key, err := client.ObjectKeyFromObject(function)
		Expect(err).To(Succeed())

		log.Info("deleting resource", "namespace", key.Namespace, "name", key.Name, "test",
			CurrentGinkgoTestDescription().FullTestText)
		Expect(k8sClient.Delete(context.Background(), function)).To(Succeed())

		log.Info("waiting for resource to disappear", "namespace", key.Namespace, "name", key.Name, "test",
			CurrentGinkgoTestDescription().FullTestText)
		Eventually(func() error {
			return k8sClient.Get(context.Background(), key, function)
		}, timeout, interval).Should(HaveOccurred())
		log.Info("deleted resource", "namespace", key.Namespace, "name", key.Name, "test",
			CurrentGinkgoTestDescription().FullTestText)

		statefulsets := new(appsv1.StatefulSetList)
		err = k8sClient.List(context.Background(), statefulsets, client.InNamespace(function.Namespace))
		Expect(err).Should(BeNil())
		for _, item := range statefulsets.Items {
			log.Info("deleting statefulset resource", "namespace", key.Namespace, "name", key.Name, "test",
				CurrentGinkgoTestDescription().FullTestText)
			Expect(k8sClient.Delete(context.Background(), &item)).To(Succeed())
		}
		log.Info("waiting for StatefulSet resource to disappear", "namespace", key.Namespace, "name", key.Name, "test",
			CurrentGinkgoTestDescription().FullTestText)
		Eventually(func() bool {
			err := k8sClient.List(context.Background(), statefulsets, client.InNamespace(function.Namespace))
			log.Info("delete statefulset result", "err", err, "statefulsets", statefulsets)
			containsFunction := false
			for _, item := range statefulsets.Items {
				if item.Name == function.Name {
					containsFunction = true
				}
			}
			return err == nil && !containsFunction
		}, timeout, interval).Should(BeTrue())
		log.Info("StatefulSet resource deleted", "namespace", key.Namespace, "name", key.Name, "test",
			CurrentGinkgoTestDescription().FullTestText)

		hpas := new(autov2beta2.HorizontalPodAutoscalerList)
		err = k8sClient.List(context.Background(), hpas, client.InNamespace(function.Namespace))
		Expect(err).Should(BeNil())
		for _, item := range hpas.Items {
			log.Info("deleting HPA resource", "namespace", key.Namespace, "name", key.Name, "test",
				CurrentGinkgoTestDescription().FullTestText)
			log.Info("deleting HPA", "item", item)
			Expect(k8sClient.Delete(context.Background(), &item)).To(Succeed())
		}
		log.Info("waiting for HPA resource to disappear", "namespace", key.Namespace, "name", key.Name, "test",
			CurrentGinkgoTestDescription().FullTestText)
		Eventually(func() bool {
			err := k8sClient.List(context.Background(), hpas, client.InNamespace(function.Namespace))
			return err == nil && len(hpas.Items) == 0
		}, timeout, interval).Should(BeTrue())
		log.Info("HPA resource deleted", "namespace", key.Namespace, "name", key.Name, "test",
			CurrentGinkgoTestDescription().FullTestText)

	})
}
