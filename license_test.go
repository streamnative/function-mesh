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

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var licenseCheck = regexp.MustCompile(`// Licensed to the Apache Software Foundation \(ASF\) under one
// or more contributor license agreements\.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership\.  The ASF licenses this file
// to you under the Apache License, Version 2\.0 \(the
// "License"\); you may not use this file except in compliance
// with the License\.  You may obtain a copy of the License at
//
//   http://www\.apache\.org/licenses/LICENSE-2\.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied\.  See the License for the
// specific language governing permissions and limitations
// under the License\.

`)

var scriptLicenseCheck = regexp.MustCompile(`#
# Licensed to the Apache Software Foundation \(ASF\) under one
# or more contributor license agreements\.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership\.  The ASF licenses this file
# to you under the Apache License, Version 2\.0 \(the
# "License"\); you may not use this file except in compliance
# with the License\.  You may obtain a copy of the License at
#
#   http://www\.apache\.org/licenses/LICENSE-2\.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied\.  See the License for the
# specific language governing permissions and limitations
# under the License\.
#
`)

var skip = map[string]bool{
	"controllers/proto/Function.pb.go":   true,
	".github/workflows/project.yml":      true,
	".github/workflows/release.yml":      true,
	".github/workflows/release-node.yml": true,
	".github/release-drafter.yml":        true,
}

func TestLicense(t *testing.T) {
	err := filepath.Walk(".", func(path string, fi os.FileInfo, err error) error {
		if skip[path] {
			return nil
		}

		if err != nil {
			return err
		}

		switch filepath.Ext(path) {
		case ".go":
			if strings.Contains(path, "zz_generated") {
				return nil
			}
			if strings.Contains(path, "vendor") {
				return nil
			}

			src, err := ioutil.ReadFile(path)
			if err != nil {
				return nil
			}

			// Find license
			if !licenseCheck.Match(src) {
				t.Errorf("%v: license header not present", path)
				return nil
			}
		case ".sh":
			if strings.Contains(path, "vendor") {
				return nil
			}
			fallthrough
		case ".conf":
			src, err := ioutil.ReadFile(path)
			if err != nil {
				return nil
			}

			// Find license
			if !scriptLicenseCheck.Match(src) {
				t.Errorf("%v: license header not present", path)
				return nil
			}

		default:
			return nil
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
