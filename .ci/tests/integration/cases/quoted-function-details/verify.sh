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

set -e

E2E_DIR=$(dirname "$0")
BASE_DIR=$(cd "${E2E_DIR}"/../../../../..; pwd)
PULSAR_NAMESPACE=${PULSAR_NAMESPACE:-"default"}
PULSAR_RELEASE_NAME=${PULSAR_RELEASE_NAME:-"sn-platform"}
E2E_KUBECONFIG=${E2E_KUBECONFIG:-"/tmp/e2e-k8s.config"}
MANIFESTS_FILE="${BASE_DIR}"/.ci/tests/integration/cases/quoted-function-details/manifests.yaml
FUNCTION_NAME=function-details-quoted-sample
STS_NAME=${FUNCTION_NAME}-function

source "${BASE_DIR}"/.ci/helm.sh

if [ ! "$KUBECONFIG" ]; then
  export KUBECONFIG=${E2E_KUBECONFIG}
fi

cleanup() {
  kubectl delete -f "${MANIFESTS_FILE}" > /dev/null 2>&1 || true
}

require_fragment() {
  fragment=$1
  message=$2

  printf '%s' "${START_COMMAND}" | grep -F -- "${fragment}" > /dev/null 2>&1 || {
    echo "${message}"
    echo "${START_COMMAND}"
    exit 1
  }
}

trap cleanup EXIT

kubectl apply -f "${MANIFESTS_FILE}" > /dev/null 2>&1

START_COMMAND=""
for _ in $(seq 1 24); do
  START_COMMAND=$(kubectl get statefulset "${STS_NAME}" -o jsonpath='{.spec.template.spec.containers[0].command[2]}' 2> /dev/null || true)
  if [ -n "${START_COMMAND}" ]; then
    break
  fi
  sleep 5
done

if [ -z "${START_COMMAND}" ]; then
  echo "statefulset command is not available"
  exit 1
fi

unsafe_pattern_fragment=$(cat <<'EOF'
timePartitionPattern\":\"yyyy-MM-dd/HH'h'-mm'm'
EOF
)
escaped_h_fragment=$(cat <<'EOF'
'"'"'h'"'"'
EOF
)
escaped_m_fragment=$(cat <<'EOF'
'"'"'m'"'"'
EOF
)

require_fragment "timePartitionPattern" "function details command does not include the quoted config key"
if printf '%s' "${START_COMMAND}" | grep -F -- "${unsafe_pattern_fragment}" > /dev/null 2>&1; then
  echo "function details command still contains raw single quotes"
  echo "${START_COMMAND}"
  exit 1
fi
require_fragment "${escaped_h_fragment}" "function details command does not escape the quoted h segment"
require_fragment "${escaped_m_fragment}" "function details command does not escape the quoted m segment"

verify_fm_result=$(ci::verify_function_mesh "${FUNCTION_NAME}" 2>&1)
if [ $? -ne 0 ]; then
  echo "${verify_fm_result}"
  exit 1
fi

verify_java_result=$(NAMESPACE=${PULSAR_NAMESPACE} CLUSTER=${PULSAR_RELEASE_NAME} ci::verify_exclamation_function \
  persistent://public/default/input-function-details-quoted-topic \
  persistent://public/default/output-function-details-quoted-topic \
  test-message \
  test-message! \
  10 2>&1)
if [ $? -eq 0 ]; then
  echo "e2e-test: ok" | yq eval -
else
  echo "${verify_java_result}"
  exit 1
fi
