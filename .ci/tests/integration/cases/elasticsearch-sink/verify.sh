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

manifests_file="${BASE_DIR}"/.ci/tests/integration/cases/elasticsearch-sink/manifests.yaml
es_file="${BASE_DIR}"/.ci/tests/integration/cases/elasticsearch-sink/elasticsearch.yaml

function install_elasticsearch_cluster() {
    # install eck operator
    kubectl create -f https://download.elastic.co/downloads/eck/2.3.0/crds.yaml
    kubectl apply -f https://download.elastic.co/downloads/eck/2.3.0/operator.yaml
    num=0
    while [[ ${num} -lt 1 ]]; do
        sleep 5s
        kubectl get pods -n elastic-system
        num=$(kubectl get pods -n elastic-system -l control-plane=elastic-operator | wc -l)
    done
    kubectl wait -n elastic-system -l control-plane=elastic-operator --for=condition=Ready pod --timeout=5m && true

    # install es cluster
    kubectl apply -f "${es_file}"
    num=0
    while [[ ${num} -lt 1 ]]; do
        sleep 5s
        kubectl get pods
        num=$(kubectl get pods -l elasticsearch.k8s.elastic.co/cluster-name=quickstart | wc -l)
    done
    kubectl wait -l elasticsearch.k8s.elastic.co/cluster-name=quickstart --for=condition=Ready pod --timeout=5m && true
}

function uninstall_elasticsearch_cluster() {
    kubectl delete elasticsearches.elasticsearch.k8s.elastic.co quickstart
    while true; do
        kubectl get elasticsearches.elasticsearch.k8s.elastic.co quickstart > /dev/null 2>&1
        if [ $? -eq 1 ]; then
            break
        fi
        sleep 5s
    done

    kubectl delete -f https://download.elastic.co/downloads/eck/2.3.0/operator.yaml
    kubectl delete -f https://download.elastic.co/downloads/eck/2.3.0/crds.yaml
}

# setup es cluster
setup_es_result=$(install_elasticsearch_cluster 2>&1)
if [ $? -ne 0 ]; then
  echo "$setup_es_result"
  uninstall_elasticsearch_cluster > /dev/null 2>&1
fi

password=$(kubectl get secret quickstart-es-elastic-user -o go-template='{{.data.elastic | base64decode}}')
sed -i.bak "s/password: \(.*\)/password: ${password}/g" "${manifests_file}"
kubectl apply -f "${manifests_file}" > /dev/null 2>&1

verify_fm_result=$(ci::verify_function_mesh sink-sample 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_fm_result"
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

verify_sink_result=$(NAMESPACE=${PULSAR_NAMESPACE} CLUSTER=${PULSAR_RELEASE_NAME} ci::verify_sink 2>&1)
if [ $? -eq 0 ]; then
  echo "e2e-test: ok" | yq eval -
else
  echo "$verify_sink_result"
fi
kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
uninstall_elasticsearch_cluster > /dev/null 2>&1 || true
