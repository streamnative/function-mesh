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

package cmdutils

import "fmt"

// Version information.
var (
	ReleaseVersion = "None"
	BuildTS        = "None"
	GitHash        = "None"
	GitBranch      = "None"
	GoVersion      = "None"
)

// Print out the version information.
func PrintVersionInfo() {
	fmt.Printf("Release Version: %s\n", ReleaseVersion)
	fmt.Printf("Git Commit Hash: %s\n", GitHash)
	fmt.Printf("Git Branch: %s\n", GitBranch)
	fmt.Printf("UTC Build Time: %s\n", BuildTS)
	fmt.Printf("Go Version: %s\n", GoVersion)
}
