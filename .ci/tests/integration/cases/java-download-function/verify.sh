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

USE_TLS=${1:-"false"}

source "${BASE_DIR}"/.ci/helm.sh

if [ ! "$KUBECONFIG" ]; then
  export KUBECONFIG=${E2E_KUBECONFIG}
fi

# TODO: for consumer using topic pattern, at least one topic should be created in advance, else consumer cannot
# subscribe to new created topic, error is: "Topic does not have schema to check"
ci::create_topic persistent://public/default/input-download-java-topic || true

kubectl apply -f "${BASE_DIR}"/.ci/tests/integration/cases/java-download-function/manifests.yaml > /dev/null 2>&1

verify_fm_result=$(ci::verify_function_mesh function-download-sample 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_fm_result"
  kubectl delete -f "${BASE_DIR}"/.ci/tests/integration/cases/java-download-function/manifests.yaml > /dev/null 2>&1 || true
  exit 1
fi

verify_java_result=$(NAMESPACE=${PULSAR_NAMESPACE} CLUSTER=${PULSAR_RELEASE_NAME} ci::verify_download_java_function 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_java_result"
  kubectl delete -f "${BASE_DIR}"/.ci/tests/integration/cases/java-download-function/manifests.yaml > /dev/null 2>&1 || true
  exit 1
fi

if [ $USE_TLS == "false" ]; then
  verify_log_topic=$(ci::verify_log_topic persistent://public/default/logging-function-logs-sidecar "it is not a NAR file" 10 2>&1)
  if [ $? -ne 0 ]; then
    echo "$verify_log_topic"
    kubectl delete -f "${BASE_DIR}"/.ci/tests/integration/cases/java-download-function/manifests.yaml > /dev/null 2>&1 || true
    exit 1
  fi
fi

kubectl delete -f "${BASE_DIR}"/.ci/tests/integration/cases/java-download-function/manifests.yaml > /dev/null 2>&1 || true
verify_cleanup_result=$(NAMESPACE=${PULSAR_NAMESPACE} CLUSTER=${PULSAR_RELEASE_NAME} ci::verify_cleanup_subscription persistent://public/default/input-download-java-topic java-download-subscription 2>&1)
if [ $? -eq 0 ]; then
  echo "e2e-test: ok" | yq eval -
else
  echo "$verify_cleanup_result"
fi
