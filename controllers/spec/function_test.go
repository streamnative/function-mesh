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
	"strings"
	"testing"

	"github.com/streamnative/function-mesh/api/v1alpha1"
	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateFunctionDetailsForStatefulFunction(t *testing.T) {
	fnc := makeFunctionSample("test")
	commands := makeFunctionCommand(fnc)
	assert.Assert(t, len(commands) == 3, "commands should be 3 but got %d", len(commands))
	startCommands := commands[2]
	assert.Assert(t, strings.Contains(startCommands, "--state_storage_serviceurl"),
		"start command should contain state_storage_serviceurl")
	assert.Assert(t, strings.Contains(startCommands, "bk://localhost:4181"),
		"start command should contain bk://localhost:4181")
}

func makeFunctionSample(functionName string) *v1alpha1.Function {
	maxPending := int32(1000)
	replicas := int32(1)
	minReplicas := int32(1)
	maxReplicas := int32(5)
	trueVal := true
	return &v1alpha1.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: "compute.functionmesh.io/v1alpha1",
		},
		ObjectMeta: *makeSampleObjectMeta(functionName),
		Spec: v1alpha1.FunctionSpec{
			Name:        functionName,
			ClassName:   "org.apache.pulsar.functions.api.examples.ExclamationFunction",
			Tenant:      "public",
			ClusterName: TestClusterName,
			Input: v1alpha1.InputConf{
				Topics: []string{
					"persistent://public/default/java-function-input-topic",
				},
				TypeClassName: "java.lang.String",
			},
			Output: v1alpha1.OutputConf{
				Topic:         "persistent://public/default/java-function-output-topic",
				TypeClassName: "java.lang.String",
			},
			LogTopic:                     "persistent://public/default/logging-function-logs",
			Timeout:                      0,
			MaxMessageRetry:              0,
			ForwardSourceMessageProperty: &trueVal,
			Replicas:                     &replicas,
			MinReplicas:                  &minReplicas,
			MaxReplicas:                  &maxReplicas,
			AutoAck:                      &trueVal,
			MaxPendingAsyncRequests:      &maxPending,
			Messaging: v1alpha1.Messaging{
				Pulsar: &v1alpha1.PulsarMessaging{
					PulsarConfig: TestClusterName,
				},
			},
			Runtime: v1alpha1.Runtime{
				Java: &v1alpha1.JavaRuntime{
					Jar:         "pulsar-functions-api-examples.jar",
					JarLocation: "public/default/nlu-test-java-function",
				},
			},
			StateConfig: &v1alpha1.Stateful{
				Pulsar: &v1alpha1.PulsarStateStore{
					ServiceURL: "bk://localhost:4181",
				},
			},
		},
	}
}
