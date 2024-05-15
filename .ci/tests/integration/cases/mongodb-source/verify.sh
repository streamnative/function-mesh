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

manifests_file="${BASE_DIR}"/.ci/tests/integration/cases/mongodb-source/manifests.yaml
mongodb_file="${BASE_DIR}"/.ci/tests/integration/cases/mongodb-source/mongodb-dbz.yaml

function install_mongodb_server() {
    # change default storageclass
    default_sc=$(kubectl get storageclasses.storage.k8s.io | grep default | awk '{ print $1 }')
    sed -i.bak "s/storageClassName: \(.*\)/storageClassName: ${default_sc}/g" "${mongodb_file}"

    # install mongodb server
    kubectl apply -f "${mongodb_file}"
    num=0
    while [[ ${num} -lt 1 ]]; do
        sleep 5
        kubectl get pods
        num=$(kubectl get pods -l role=mongo | wc -l)
    done
    kubectl wait -l role=mongo --for=condition=Ready pod --timeout=5m && true

    # initialize the data
    kubectl exec mongo-dbz-0 -c mongo -- bash ./usr/local/bin/init-inventory.sh
}

function uninstall_mongodb_server() {
    # uninstall mongodb server
    kubectl delete -f "${mongodb_file}"
}

# setup mongodb server
setup_mongodb_result=$(install_mongodb_server 2>&1)
if [ $? -ne 0 ]; then
  echo "$setup_mongodb_result"
  uninstall_mongodb_server > /dev/null 2>&1
fi

kubectl apply -f "${manifests_file}" > /dev/null 2>&1

verify_fm_result=$(ci::verify_function_mesh source-sample 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_fm_result"
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

if [ $USE_TLS == "false" ]; then
  verify_log_topic=$(ci::verify_log_topic persistent://public/default/mongo-source-logs "it may be a NAR file" 10 2>&1)
  if [ $? -ne 0 ]; then
    echo "$verify_log_topic"
    kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
    exit 1
  fi
fi

verify_source_result=$(ci::verify_source 2>&1)
if [ $? -eq 0 ]; then
  echo "e2e-test: ok" | yq eval -
else
  echo "$verify_source_result"
fi
kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
uninstall_mongodb_server > /dev/null 2>&1 || true
