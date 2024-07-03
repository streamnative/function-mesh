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
	"regexp"
	"strings"
	"time"

	"github.com/streamnative/function-mesh/controllers/spec"

	appv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
)

var _ = Describe("Sink Controller", func() {
	Context("Simple Sink Item", func() {
		pulsarConfig := makeSamplePulsarConfig()
		sink := makeSinkSample()
		if sink.Status.Conditions == nil {
			sink.Status.Conditions = make(map[v1alpha1.Component]v1alpha1.ResourceCondition)
		}
		statefulSet, _ := spec.MakeSinkStatefulSet(context.Background(), k8sClient, sink)

		It("Should create pulsar configmap successfully", func() {
			Expect(k8sClient.Create(context.Background(), pulsarConfig)).Should(Succeed())
		})

		It("Should create successfully", func() {
			Expect(k8sClient.Create(context.Background(), statefulSet)).Should(Succeed())
		})

		It("Should delete successfully", func() {
			Expect(k8sClient.Delete(context.Background(), statefulSet)).Should(Succeed())
		})

		It("Should delete pulsar configmap successfully", func() {
			Expect(k8sClient.Delete(context.Background(), pulsarConfig)).Should(Succeed())
		})
	})

	Context("Sink Controller", func() {
		sink := makeSinkSample()

		It("Should update sink statefulset container command", func() {
			err := k8sClient.Create(context.Background(), sink)
			Expect(err).NotTo(HaveOccurred())
			sinkSts := &appv1.StatefulSet{}

			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      spec.MakeSinkObjectMeta(sink).Name,
					Namespace: spec.MakeSinkObjectMeta(sink).Namespace,
				}, sinkSts)
				return err == nil && len(sinkSts.Spec.Template.Spec.Containers[0].Command[2]) > 0
			}, 10*time.Second, 1*time.Second).Should(BeTrue())
			err = k8sClient.Get(context.Background(), types.NamespacedName{
				Name:      TestSinkName,
				Namespace: TestNameSpace,
			}, sink)
			Expect(err).NotTo(HaveOccurred())

			oldSinkConfig := sink.Spec.SinkConfig
			oldSinkConfig.Data = map[string]interface{}{
				"configkey1": "configvalue1",
				"configkey2": "configvalue2",
				"configkey3": "configvalue3",
			}
			sink.Spec.SinkConfig = oldSinkConfig
			err = k8sClient.Update(context.Background(), sink)
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() bool {
				sinkReconciler.Reconcile(
					context.Background(),
					ctrl.Request{
						NamespacedName: types.NamespacedName{
							Name:      TestSinkName,
							Namespace: TestNameSpace,
						},
					})
				err := k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      spec.MakeSinkObjectMeta(sink).Name,
					Namespace: spec.MakeSinkObjectMeta(sink).Namespace,
				}, sinkSts)
				return err == nil && sinkSts.ObjectMeta.Generation > 1
			}, 10*time.Second, 1*time.Second).Should(BeTrue())
			re := regexp.MustCompile("{\"configkey1\":\"configvalue1\",\"configkey2\":\"configvalue2\",\"configkey3\":\"configvalue3\"}")
			// Verify new config synced to pod spec
			Expect(len(re.FindAllString(strings.ReplaceAll(sinkSts.Spec.Template.Spec.Containers[0].Command[2], "\\", ""), -1))).To(Equal(1))
			// cleanup
			Expect(k8sClient.Delete(context.Background(), sink)).Should(Succeed())
		})

	})
})
