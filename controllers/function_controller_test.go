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

	"k8s.io/apimachinery/pkg/api/resource"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"

	ctrl "sigs.k8s.io/controller-runtime"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	autov2 "k8s.io/api/autoscaling/v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
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
					context.Background(),
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
		function.Spec.Pod.AutoScalingMetrics = []autov2.MetricSpec{
			{
				Type: autov2.ResourceMetricSourceType,
				Resource: &autov2.ResourceMetricSource{
					Name: v1.ResourceCPU,
					Target: autov2.MetricTarget{
						Type:               autov2.UtilizationMetricType,
						AverageUtilization: &cpuPercentage,
					},
				},
			},
			{
				Type: autov2.ResourceMetricSourceType,
				Resource: &autov2.ResourceMetricSource{
					Name: v1.ResourceMemory,
					Target: autov2.MetricTarget{
						Type:               autov2.UtilizationMetricType,
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

var _ = Describe("Function Controller (VPA)", func() {
	Context("Simple Function Item with VPA", func() {
		mode := vpav1.UpdateModeAuto
		controlledValues := vpav1.ContainerControlledValuesRequestsAndLimits
		function := makeFunctionSample(TestFunctionVPAName)
		function.Spec.Pod.VPA = &v1alpha1.VPASpec{
			UpdatePolicy: &vpav1.PodUpdatePolicy{
				UpdateMode: &mode,
			},
			ResourcePolicy: &vpav1.PodResourcePolicy{
				ContainerPolicies: []vpav1.ContainerResourcePolicy{
					{
						ContainerName: "*",
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
		}

		createFunction(function)
	})
})

var _ = Describe("Function Controller (Stateful Function)", func() {
	Context("Simple Function Item with Stateful store config", func() {
		function := makeFunctionSample(TestFunctionStatefulJavaName)
		function.Spec.StateConfig = &v1alpha1.Stateful{
			Pulsar: &v1alpha1.PulsarStateStore{
				ServiceURL: "bk://localhost:4181",
			},
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

		if function.Spec.StateConfig != nil {
			containers := statefulSet.Spec.Template.Spec.Containers
			Expect(len(containers) > 0).Should(BeTrue())
			for _, container := range containers {
				if container.Name == "pulsar-function" {
					fullCommand := strings.Join(container.Command, " ")
					Expect(fullCommand).Should(ContainSubstring("--state_storage_serviceurl"),
						"--state_storage_serviceurl should be set in [%s]", fullCommand)
					Expect(fullCommand).Should(ContainSubstring("bk://localhost:4181"))
				}
			}
		}
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
			hpa := &autov2.HorizontalPodAutoscaler{}
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
					autoscaler, _ := spec.GetBuiltinAutoScaler(rule)
					Expect(autoscaler).Should(Not(BeNil()))
					Expect(hpa.Spec.Metrics).Should(ContainElement(autoscaler.Metrics()[0]))
				}
			}
		}
	})

	It("VPA should be created", func() {
		if function.Spec.Pod.VPA != nil {
			vpa := &vpav1.VerticalPodAutoscaler{}
			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), types.NamespacedName{Namespace: function.Namespace,
					Name: spec.MakeFunctionObjectMeta(function).Name}, vpa)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			log.Info("VPA should be created", "vpa", vpa, "name", spec.MakeFunctionObjectMeta(function).Name)

			Expect(vpa.Spec.UpdatePolicy).Should(Equal(function.Spec.Pod.VPA.UpdatePolicy))
			Expect(vpa.Spec.ResourcePolicy).Should(Equal(function.Spec.Pod.VPA.ResourcePolicy))
		}
	})

	It("Function should be deleted", func() {
		key := client.ObjectKeyFromObject(function)
		//fmt.Println(key)
		//Expect(key).To(nil)

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

		statefulset := new(appsv1.StatefulSet)
		statefulsetKey := key
		statefulsetKey.Name = spec.MakeFunctionObjectMeta(function).Name
		err := k8sClient.Get(context.Background(), statefulsetKey, statefulset)
		Expect(err).Should(BeNil())
		Expect(k8sClient.Delete(context.Background(), statefulset)).To(Succeed())
		Eventually(func() bool {
			err := k8sClient.Get(context.Background(), statefulsetKey, statefulset)
			log.Info("delete statefulset result", "err", err, "statefulset", statefulset)
			return err != nil && strings.Contains(err.Error(), "not found")
		}, timeout, interval).Should(BeTrue())
		log.Info("StatefulSet resource deleted", "namespace", key.Namespace, "name", key.Name, "test",
			CurrentGinkgoTestDescription().FullTestText)

		hpa := new(autov2.HorizontalPodAutoscaler)
		hpaKey := key
		hpaKey.Name = spec.MakeFunctionObjectMeta(function).Name
		err = k8sClient.Get(context.Background(), hpaKey, hpa)
		Expect(err).Should(BeNil())
		Expect(k8sClient.Delete(context.Background(), hpa)).To(Succeed())
		Eventually(func() bool {
			err := k8sClient.Get(context.Background(), hpaKey, hpa)
			log.Info("delete hpa result", "err", err, "hpa", hpa)
			return err != nil && strings.Contains(err.Error(), "not found")
		}, timeout, interval).Should(BeTrue())
		log.Info("HPA resource deleted", "namespace", key.Namespace, "name", key.Name, "test",
			CurrentGinkgoTestDescription().FullTestText)

		vpas := new(vpav1.VerticalPodAutoscalerList)
		err = k8sClient.List(context.Background(), vpas, client.InNamespace(function.Namespace))
		Expect(err).Should(BeNil())
		for _, item := range vpas.Items {
			log.Info("deleting VPA resource", "namespace", key.Namespace, "name", key.Name, "test",
				CurrentGinkgoTestDescription().FullTestText)
			log.Info("deleting VPA", "item", item)
			Expect(k8sClient.Delete(context.Background(), &item)).To(Succeed())
		}
		log.Info("waiting for VPA resource to disappear", "namespace", key.Namespace, "name", key.Name, "test",
			CurrentGinkgoTestDescription().FullTestText)
		Eventually(func() bool {
			err := k8sClient.List(context.Background(), vpas, client.InNamespace(function.Namespace))
			return err == nil && len(vpas.Items) == 0
		}, timeout, interval).Should(BeTrue())
		log.Info("VPA resource deleted", "namespace", key.Namespace, "name", key.Name, "test",
			CurrentGinkgoTestDescription().FullTestText)

	})
}
