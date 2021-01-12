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

var _ = Describe("FunctionMesh Controller", func() {
	Context("Simple FunctionMesh", func() {
		pulsarConfig := makeSamplePulsarConfig()
		mesh := makeFunctionMeshSample()
		if mesh.Status.FunctionConditions == nil {
			mesh.Status.FunctionConditions = make(map[string]v1alpha1.ResourceCondition)
		}
		if mesh.Status.SourceConditions == nil {
			mesh.Status.SourceConditions = make(map[string]v1alpha1.ResourceCondition)
		}
		if mesh.Status.SinkConditions == nil {
			mesh.Status.SinkConditions = make(map[string]v1alpha1.ResourceCondition)
		}

		It("Should create pulsar configmap successfully", func() {
			Expect(k8sClient.Create(context.Background(), pulsarConfig)).Should(Succeed())
		})

		for _, functionSpec := range mesh.Spec.Functions {
			function := spec.MakeFunctionComponent(makeComponentName(mesh.Name, functionSpec.Name), mesh, &functionSpec)
			It("Should create function successfully", func() {
				Expect(k8sClient.Create(context.Background(), function)).Should(Succeed())
			})
		}

		for _, functionSpec := range mesh.Spec.Functions {
			function := spec.MakeFunctionComponent(makeComponentName(mesh.Name, functionSpec.Name), mesh, &functionSpec)
			It("Should delete function successfully", func() {
				Expect(k8sClient.Delete(context.Background(), function)).Should(Succeed())
			})
		}

		It("Should delete pulsar configmap successfully", func() {
			Expect(k8sClient.Delete(context.Background(), pulsarConfig)).Should(Succeed())
		})
	})

})
