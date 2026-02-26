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
MANIFESTS_FILE="${BASE_DIR}"/.ci/tests/integration-oauth2/cases/java-download-function-with-package-service/manifests.yaml

source "${BASE_DIR}"/.ci/helm.sh

FUNCTION_NAME=function-download-sample-package-service
STS_NAME=${FUNCTION_NAME}-function

if [ ! "$KUBECONFIG" ]; then
  export KUBECONFIG=${E2E_KUBECONFIG}
fi

kubectl apply -f "${MANIFESTS_FILE}" > /dev/null 2>&1

verify_fm_result=$(ci::verify_function_mesh ${FUNCTION_NAME} 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_fm_result"
  kubectl delete -f "${MANIFESTS_FILE}" > /dev/null 2>&1 || true
  exit 1
fi

sts_yaml=$(kubectl get statefulset "${STS_NAME}" -o yaml 2>&1)
echo "${sts_yaml}" | grep "prefix: PACKAGE_" > /dev/null 2>&1 || {
  echo "packageService env prefix is not injected"
  kubectl delete -f "${MANIFESTS_FILE}" > /dev/null 2>&1 || true
  exit 1
}
echo "${sts_yaml}" | grep "name: test-pulsar-package-service" > /dev/null 2>&1 || {
  echo "packageService configmap is not injected"
  kubectl delete -f "${MANIFESTS_FILE}" > /dev/null 2>&1 || true
  exit 1
}
echo "${sts_yaml}" | grep "/etc/oauth2-package-service" > /dev/null 2>&1 || {
  echo "packageService oauth2 volume mount is not injected"
  kubectl delete -f "${MANIFESTS_FILE}" > /dev/null 2>&1 || true
  exit 1
}

verify_java_result=$(NAMESPACE=${PULSAR_NAMESPACE} CLUSTER=${PULSAR_RELEASE_NAME} ci::verify_exclamation_function_with_auth \
  persistent://public/default/input-download-java-package-service-topic \
  persistent://public/default/output-download-java-package-service-topic \
  test-message test-message! 10 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_java_result"
  kubectl delete -f "${MANIFESTS_FILE}" > /dev/null 2>&1 || true
  exit 1
fi

verify_log_topic=$(ci::verify_log_topic_with_auth persistent://public/default/logging-function-logs-package-service "it is not a NAR file" 10 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_log_topic"
  kubectl delete -f "${MANIFESTS_FILE}" > /dev/null 2>&1 || true
  exit 1
fi

kubectl delete -f "${MANIFESTS_FILE}" > /dev/null 2>&1 || true
echo "e2e-test: ok" | yq eval -
