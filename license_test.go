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
	"os"
	"os/exec"
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
	"hack/webhook-create-signed-cert.sh": true,
	"hack/webhooks/certs/csr.conf":       true,
	"pulsar-charts":                      true,
}

func TestLicense(t *testing.T) {
	trackedFiles, trackedDirs := gitTrackedPaths()

	err := filepath.Walk(".", func(path string, fi os.FileInfo, err error) error {
		if skip[path] {
			return nil
		}

		if err != nil {
			return err
		}

		cleanPath := filepath.Clean(path)
		if fi.IsDir() {
			if cleanPath == "vendor" || cleanPath == ".git" {
				return filepath.SkipDir
			}
			if len(trackedDirs) > 0 && cleanPath != "." {
				if _, ok := trackedDirs[cleanPath]; !ok {
					return filepath.SkipDir
				}
			}
			return nil
		}

		if len(trackedFiles) > 0 {
			if _, ok := trackedFiles[cleanPath]; !ok {
				return nil
			}
		}

		switch filepath.Ext(path) {
		case ".go":
			if strings.Contains(path, "zz_generated") {
				return nil
			}
			if strings.Contains(path, "vendor") {
				return nil
			}

			src, err := os.ReadFile(path)
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
			if strings.Contains(path, "vendor") {
				return nil
			}
			src, err := os.ReadFile(path)
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

func gitTrackedPaths() (map[string]struct{}, map[string]struct{}) {
	cmd := exec.Command("git", "ls-files", "-z")
	output, err := cmd.Output()
	if err != nil {
		return nil, nil
	}

	files := make(map[string]struct{})
	dirs := map[string]struct{}{
		".": {},
	}

	for _, file := range strings.Split(string(output), "\x00") {
		if file == "" {
			continue
		}

		cleanFile := filepath.Clean(file)
		files[cleanFile] = struct{}{}

		dir := filepath.Dir(cleanFile)
		for dir != "." && dir != string(filepath.Separator) {
			dirs[dir] = struct{}{}
			next := filepath.Dir(dir)
			if next == dir {
				break
			}
			dir = next
		}
	}

	return files, dirs
}
