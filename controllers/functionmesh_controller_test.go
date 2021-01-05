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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("FunctionMesh Controller", func() {
	var objectMeta = metav1.ObjectMeta{
		Name:      "test-function-mesh",
		Namespace: "default",
		UID:       "dead-beef", // uid not generate automatically with fake k8s
	}
	pulsarConfig := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pulsar",
			Namespace: "default",
		},
		Data: map[string]string{
			"webServiceURL":    "http://test-pulsar-broker.default.svc.cluster.local:8080",
			"brokerServiceURL": "pulsar://test-pulsar-broker.default.svc.cluster.local:6650",
		},
	}
	var pulsar = &v1alpha1.PulsarMessaging{
		PulsarConfig: pulsarConfig.Name,
		//AuthConfig: "test-auth",
	}

	Context("Simple FunctionMesh", func() {
		maxPending := int32(1000)
		replicas := int32(1)
		maxReplicas := int32(1)
		trueVal := true

		mesh := &v1alpha1.FunctionMesh{
			TypeMeta: metav1.TypeMeta{
				Kind:       "FunctionMesh",
				APIVersion: "cloud.streamnative.io/v1alpha1",
			},
			ObjectMeta: objectMeta,
			Spec: v1alpha1.FunctionMeshSpec{
				Functions: []v1alpha1.FunctionSpec{
					{
						Name:        "ex1",
						ClassName:   "org.apache.pulsar.functions.api.examples.ExclamationFunction",
						Tenant:      "public",
						ClusterName: "test-pulsar",
						SourceType:  "java.lang.String",
						SinkType:    "java.lang.String",
						Input: v1alpha1.InputConf{
							Topics: []string{
								"persistent://public/default/functionmesh-input-topic",
							},
						},
						Output: v1alpha1.OutputConf{
							Topic: "persistent://public/default/mid-topic",
						},
						LogTopic:                     "persistent://public/default/logging-function-logs",
						Timeout:                      0,
						MaxMessageRetry:              0,
						ForwardSourceMessageProperty: &trueVal,
						Replicas:                     &replicas,
						MaxReplicas:                  &maxReplicas,
						AutoAck:                      &trueVal,
						MaxPendingAsyncRequests:      &maxPending,
						Messaging: v1alpha1.Messaging{
							Pulsar: pulsar,
						},
						Runtime: v1alpha1.Runtime{
							Java: &v1alpha1.JavaRuntime{
								Jar:         "pulsar-functions-api-examples.jar",
								JarLocation: "public/default/nlu-test-functionmesh-ex1",
							},
						},
					},
					{
						Name:        "ex2",
						ClassName:   "org.apache.pulsar.functions.api.examples.ExclamationFunction",
						Tenant:      "public",
						ClusterName: "test-pulsar",
						SourceType:  "java.lang.String",
						SinkType:    "java.lang.String",
						Input: v1alpha1.InputConf{
							Topics: []string{
								"persistent://public/default/mid-topic",
							},
						},
						Output: v1alpha1.OutputConf{
							Topic: "persistent://public/default/functionmesh-output-topic",
						},
						LogTopic:                     "persistent://public/default/logging-function-logs",
						Timeout:                      0,
						MaxMessageRetry:              0,
						ForwardSourceMessageProperty: &trueVal,
						Replicas:                     &replicas,
						MaxReplicas:                  &maxReplicas,
						AutoAck:                      &trueVal,
						MaxPendingAsyncRequests:      &maxPending,
						Messaging: v1alpha1.Messaging{
							Pulsar: pulsar,
						},
						Runtime: v1alpha1.Runtime{
							Java: &v1alpha1.JavaRuntime{
								Jar:         "pulsar-functions-api-examples.jar",
								JarLocation: "public/default/nlu-test-functionmesh-ex2",
							},
						},
					},
				},
			},
		}

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
