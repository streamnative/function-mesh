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

manifests_file="${BASE_DIR}"/.ci/tests/integration/cases/java-log-level/manifests.yaml

kubectl apply -f "${manifests_file}" > /dev/null 2>&1

verify_fm_result=$(ci::verify_function_mesh java-log-level 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_fm_result"
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

verify_java_result=$(NAMESPACE=${PULSAR_NAMESPACE} CLUSTER=${PULSAR_RELEASE_NAME} ci::verify_exclamation_function "persistent://public/default/input-java-log-level-topic" "persistent://public/default/output-java-log-level-topic" "test-message" "test-message!" 10 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_java_result"
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

verify_log_level_result=$(kubectl logs -l compute.functionmesh.io/name=java-log-level --tail=-1 | grep "Got result: object:" 2>&1)
if [ $? -eq 0 ]; then
  echo "e2e-test: ok" | yq eval -
else
  echo "$verify_log_level_result"
fi
kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
