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

	"github.com/streamnative/function-mesh/api/v1alpha1"

	"github.com/stretchr/testify/assert"
)

func TestGetDownloadCommand(t *testing.T) {
	doTest := func(downloadPath, componentPackage string, expectedCommand []string) {
		actualResult := getDownloadCommand(downloadPath, componentPackage)
		assert.Equal(t, expectedCommand, actualResult)
	}

	testData := []struct {
		downloadPath     string
		componentPackage string
		expectedCommand  []string
	}{
		// test get the download command with package name
		{"function://public/default/test@v1", "function-package.jar",
			[]string{
				PulsarAdminExecutableFile,
				"--admin-url", "$webServiceURL",
				"packages", "download", "function://public/default/test@v1", "--path", PulsarDownloadRootDir + "/function-package.jar",
			},
		},
		{"sink://public/default/test@v1", "sink-package.jar",
			[]string{
				PulsarAdminExecutableFile,
				"--admin-url", "$webServiceURL",
				"packages", "download", "sink://public/default/test@v1", "--path", PulsarDownloadRootDir + "/sink-package.jar",
			},
		},
		{"source://public/default/test@v1", "source-package.jar",
			[]string{
				PulsarAdminExecutableFile,
				"--admin-url", "$webServiceURL",
				"packages", "download", "source://public/default/test@v1", "--path", PulsarDownloadRootDir + "/source-package.jar",
			},
		},
		// test get the download command with normal name
		{"/test", "test.jar",
			[]string{
				PulsarAdminExecutableFile,
				"--admin-url", "$webServiceURL",
				"functions", "download", "--path", "/test", "--destination-file", PulsarDownloadRootDir + "/test.jar",
			},
		},
		// test get the download command with a wrong package name
		{"source/public/default/test@v1", "source-package.jar",
			[]string{
				PulsarAdminExecutableFile,
				"--admin-url", "$webServiceURL",
				"functions", "download", "--path", "source/public/default/test@v1", "--destination-file", PulsarDownloadRootDir + "/source-package.jar",
			},
		},
		{"source:/public/default/test@v1", "source-package.jar",
			[]string{
				PulsarAdminExecutableFile,
				"--admin-url", "$webServiceURL",
				"functions", "download", "--path", "source:/public/default/test@v1", "--destination-file", PulsarDownloadRootDir + "/source-package.jar",
			},
		},
	}

	for _, v := range testData {
		doTest(v.downloadPath, v.componentPackage, v.expectedCommand)
	}
}

func TestGetFunctionRunnerImage(t *testing.T) {
	javaRuntime := v1alpha1.Runtime{Java: &v1alpha1.JavaRuntime{
		Jar:         "test.jar",
		JarLocation: "test",
	}}
	image := getFunctionRunnerImage(&v1alpha1.FunctionSpec{Runtime: javaRuntime})
	assert.Equal(t, image, DefaultJavaRunnerImage)

	pythonRuntime := v1alpha1.Runtime{Python: &v1alpha1.PythonRuntime{
		Py:         "test.py",
		PyLocation: "test",
	}}
	image = getFunctionRunnerImage(&v1alpha1.FunctionSpec{Runtime: pythonRuntime})
	assert.Equal(t, image, DefaultPythonRunnerImage)

	goRuntime := v1alpha1.Runtime{Golang: &v1alpha1.GoRuntime{
		Go:         "test",
		GoLocation: "test",
	}}
	image = getFunctionRunnerImage(&v1alpha1.FunctionSpec{Runtime: goRuntime})
	assert.Equal(t, image, DefaultGoRunnerImage)
}
