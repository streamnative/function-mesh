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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/streamnative/function-mesh/api/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/spec"
)

var _ = Describe("Source Controller", func() {
	Context("Simple Source Item", func() {
		pulsarConfig := makeSamplePulsarConfig()
		source := makeSourceSample()
		if source.Status.Conditions == nil {
			source.Status.Conditions = make(map[v1alpha1.Component]v1alpha1.ResourceCondition)
		}
		statefulSet := spec.MakeSourceStatefulSet(source)

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
})
