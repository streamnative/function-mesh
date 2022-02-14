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

package config

import (
	"github.com/streamnative/function-mesh/controllers/spec"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type RunnerImages struct {
	Java   string `yaml:"java,omitempty"`
	Python string `yaml:"python,omitempty"`
	Go     string `yaml:"go,omitempty"`
}

type ControllerConfigs struct {
	RunnerImages        RunnerImages      `yaml:"runnerImages,omitempty"`
	ResourceLabels      map[string]string `yaml:"resourceLabels,omitempty"`
	ResourceAnnotations map[string]string `yaml:"resourceAnnotations,omitempty"`
}

var Configs = DefaultConfigs()

func DefaultConfigs() *ControllerConfigs {
	return &ControllerConfigs{
		RunnerImages: RunnerImages{
			Java:   spec.DefaultJavaRunnerImage,
			Python: spec.DefaultPythonRunnerImage,
			Go:     spec.DefaultGoRunnerImage,
		},
	}
}

func ParseControllerConfigs(configFilePath string) error {
	yamlFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, Configs)
	if err != nil {
		return err
	}

	return nil
}
