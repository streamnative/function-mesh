#!/usr/bin/env bash
#
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.
#

# copied from: https://github.com/OpenFunction/OpenFunction/blob/main/hack/update-codegen.sh

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(git rev-parse --show-toplevel)

# Grab code-generator version from go.sum.
CODEGEN_VERSION=$(grep 'k8s.io/code-generator' go.sum | awk '{print $2}' | head -1 |cut -b 1-7)
CODEGEN_PKG=$(echo `go env GOPATH`"/pkg/mod/k8s.io/code-generator@${CODEGEN_VERSION}")

# code-generator does work with go.mod but makes assumptions about
# the project living in `$GOPATH/src`. To work around this and support
# any location; create a temporary directory, use this as an output
# base, and copy everything back once generated.
TEMP_DIR=$(mktemp -d)
cleanup() {
    echo ">> Removing ${TEMP_DIR}"
    rm -rf ${TEMP_DIR}
}
trap "cleanup" EXIT SIGINT

echo ">> Temporary output directory ${TEMP_DIR}"

# Ensure we can execute.
chmod +x ${CODEGEN_PKG}/generate-groups.sh

${CODEGEN_PKG}/generate-groups.sh "client" \
    github.com/streamnative/function-mesh/api/generated github.com/streamnative/function-mesh/api \
    "compute:v1alpha1" \
    --output-base "${TEMP_DIR}" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.go.txt

# Copy everything back.
cp -r "${TEMP_DIR}/github.com/streamnative/function-mesh/api/." "${SCRIPT_ROOT}/api/"