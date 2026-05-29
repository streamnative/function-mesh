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
BASE_DIR=$(cd "${E2E_DIR}"/../../../../..;pwd)
PULSAR_NAMESPACE=${PULSAR_NAMESPACE:-"default"}
PULSAR_RELEASE_NAME=${PULSAR_RELEASE_NAME:-"sn-platform"}
E2E_KUBECONFIG=${E2E_KUBECONFIG:-"/tmp/e2e-k8s.config"}

source "${BASE_DIR}"/.ci/helm.sh

if [ ! "$KUBECONFIG" ]; then
  export KUBECONFIG=${E2E_KUBECONFIG}
fi

manifests_file="${BASE_DIR}"/.ci/tests/integration/cases/crypto-function/manifests.yaml
input_topic="persistent://public/default/java-function-crypto-input-topic"
output_topic="persistent://public/default/java-function-crypto-output-topic"

function delete_topic() {
  topic=$1
  kubectl exec -n "${PULSAR_NAMESPACE}" "${PULSAR_RELEASE_NAME}"-pulsar-broker-0 -- bin/pulsar-admin topics delete "${topic}" -f > /dev/null 2>&1 || true
  kubectl exec -n "${PULSAR_NAMESPACE}" "${PULSAR_RELEASE_NAME}"-pulsar-broker-0 -- bin/pulsar-admin topics delete-partitioned-topic "${topic}" -f > /dev/null 2>&1 || true
  kubectl exec -n "${PULSAR_NAMESPACE}" "${PULSAR_RELEASE_NAME}"-pulsar-broker-0 -- bin/pulsar-admin topics delete "${topic}-partition-0" -f > /dev/null 2>&1 || true
  kubectl exec -n "${PULSAR_NAMESPACE}" "${PULSAR_RELEASE_NAME}"-pulsar-broker-0 -- bin/pulsar-admin schemas delete "${topic}" > /dev/null 2>&1 || true
}

function cleanup() {
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  delete_topic "${input_topic}"
  delete_topic "${output_topic}"
}

function upload_string_schema() {
  kubectl exec -n "${PULSAR_NAMESPACE}" "${PULSAR_RELEASE_NAME}"-pulsar-broker-0 -- sh -c \
    'printf "%s\n" "{\"type\":\"STRING\",\"schema\":\"\",\"properties\":{}}" > /tmp/function-mesh-string-schema.json && bin/pulsar-admin schemas upload --filename /tmp/function-mesh-string-schema.json "$1"' \
    sh "${input_topic}" > /dev/null
}

trap cleanup EXIT
cleanup

if ! create_topic_result=$(NAMESPACE=${PULSAR_NAMESPACE} CLUSTER=${PULSAR_RELEASE_NAME} ci::create_topic "${input_topic}" 2>&1); then
  echo "$create_topic_result"
  exit 1
fi

if ! upload_schema_result=$(upload_string_schema 2>&1); then
  echo "$upload_schema_result"
  exit 1
fi

kubectl apply -f "${manifests_file}" > /dev/null 2>&1

if ! verify_fm_result=$(ci::verify_function_mesh java-function-crypto-sample 2>&1); then
  echo "$verify_fm_result"
  exit 1
fi

if verify_crypto_result=$(NAMESPACE=${PULSAR_NAMESPACE} CLUSTER=${PULSAR_RELEASE_NAME} ci::verify_crypto_function 2>&1); then
  echo "e2e-test: ok" | yq eval -
else
  echo "$verify_crypto_result"
  exit 1
fi
