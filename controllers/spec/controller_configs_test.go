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
	"testing"

	"gotest.tools/assert"
)

func TestParseConfigFiles(t *testing.T) {
	err := ParseControllerConfigs("testdata/controller_configs.yaml")
	if err != nil {
		t.Errorf("ParseControllerConfigs failed: %v", err)
	}
	assert.Assert(t, Configs != nil)
	assert.Assert(t, Configs.RunnerImages.Java == "streamnative/pulsar-functions-java-runner:latest")
	assert.Assert(t, Configs.RunnerImages.Python == "streamnative/pulsar-functions-python-runner:latest")
	assert.Assert(t, Configs.RunnerImages.Go == "streamnative/pulsar-functions-go-runner:latest")
	assert.Assert(t, len(Configs.ResourceLabels) == 2)
	assert.Assert(t, len(Configs.ResourceAnnotations) == 1)
	assert.Assert(t, Configs.ResourceLabels["functionmesh.io/managedBy"] == "function-mesh")
	assert.Assert(t, Configs.ResourceLabels["foo"] == "bar")
	assert.Assert(t, Configs.ResourceAnnotations["fooAnnotation"] == "barAnnotation")
}
