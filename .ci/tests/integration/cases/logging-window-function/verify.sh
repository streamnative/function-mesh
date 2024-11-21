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

manifests_file="${BASE_DIR}"/.ci/tests/integration/cases/logging-window-function/manifests.yaml

kubectl apply -f "${manifests_file}" > /dev/null 2>&1

verify_fm_result=$(ci::verify_function_mesh window-function-sample 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_fm_result"
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

verify_java_result=$(NAMESPACE=${PULSAR_NAMESPACE} CLUSTER=${PULSAR_RELEASE_NAME} ci::send_test_data "persistent://public/default/window-function-input-topic" "test-message" 3 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_java_result"
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

sleep 3

# the 3 messages will not be processed, so backlog should be 3
verify_backlog_result=$(NAMESPACE=${PULSAR_NAMESPACE} CLUSTER=${PULSAR_RELEASE_NAME} ci::verify_backlog "persistent://public/default/window-function-input-topic-partition-0" "public/default/window-function-sample" 3 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_backlog_result"
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

# it will fire the window with first 5 messages when get the 5th message, and then fire again with 10 messages when get 10th message
verify_java_result=$(NAMESPACE=${PULSAR_NAMESPACE} CLUSTER=${PULSAR_RELEASE_NAME} ci::send_test_data "persistent://public/default/window-function-input-topic" "test-message" 7 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_java_result"
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

sleep 3

verify_backlog_result=$(NAMESPACE=${PULSAR_NAMESPACE} CLUSTER=${PULSAR_RELEASE_NAME} ci::verify_backlog "persistent://public/default/window-function-input-topic-partition-0" "public/default/window-function-sample" 0 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_backlog_result"
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

verify_log_result=$(kubectl logs -l compute.functionmesh.io/name=window-function-sample --tail=-1 | grep -e "-window-log" | wc -l)
if [ $verify_log_result -ne 0 ]; then
  sub_name=$(echo $RANDOM | md5sum | head -c 20; echo;)
  verify_log_topic_result=$(kubectl exec -n ${PULSAR_NAMESPACE} ${PULSAR_RELEASE_NAME}-pulsar-broker-0 -- bin/pulsar-client consume -n 15 -s $sub_name --subscription-position Earliest "persistent://public/default/window-function-logs" | grep  -e "-window-log" | wc -l)
  if [ $verify_log_topic_result -ne 0 ]; then
    echo "e2e-test: ok" | yq eval -
  else
    echo "$verify_log_topic_result"
  fi
else
  echo "$verify_log_result"
fi

kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
