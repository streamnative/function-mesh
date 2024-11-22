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
	"os"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

type RunnerImages struct {
	Java           string            `yaml:"java,omitempty"`
	Python         string            `yaml:"python,omitempty"`
	Go             string            `yaml:"go,omitempty"`
	GenericRuntime map[string]string `yaml:"genericRuntime,omitempty"`
}

type ControllerConfigs struct {
	RunnerImages           RunnerImages        `yaml:"runnerImages,omitempty"`
	RunnerImagePullSecrets []map[string]string `yaml:"runnerImagePullSecrets,omitempty"`
	RunnerImagePullPolicy  corev1.PullPolicy   `yaml:"imagePullPolicy,omitempty"`
	ResourceLabels         map[string]string   `yaml:"resourceLabels,omitempty"`
	ResourceAnnotations    map[string]string   `yaml:"resourceAnnotations,omitempty"`
}

var Configs = DefaultConfigs()

func DefaultConfigs() *ControllerConfigs {
	return &ControllerConfigs{
		RunnerImages: RunnerImages{
			Java:   DefaultJavaRunnerImage,
			Python: DefaultPythonRunnerImage,
			Go:     DefaultGoRunnerImage,
			GenericRuntime: map[string]string{
				"nodejs":     DefaultGenericNodejsRunnerImage,
				"python":     DefaultGenericPythonRunnerImage,
				"executable": DefaultGenericRunnerImage,
				"wasm":       DefaultGenericRunnerImage,
			},
		},
	}
}

func ParseControllerConfigs(configFilePath string) error {
	yamlFile, err := os.ReadFile(configFilePath)
	if err != nil {
		return err
	}
	if len(yamlFile) == 0 {
		return nil
	}
	err = yaml.Unmarshal(yamlFile, Configs)
	if err != nil {
		return err
	}

	return nil
}
