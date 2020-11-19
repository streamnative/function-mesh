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

set -ex

BINDIR=`dirname "$0"`
PULSAR_HOME=`cd ${BINDIR}/..;pwd`
FUNCTION_MESH_HOME=${PULSAR_HOME}
OUTPUT_BIN=${FUNCTION_MESH_HOME}/output/bin
KIND_BIN=$OUTPUT_BIN/kind
HELM=${OUTPUT_BIN}/helm
KUBECTL=${OUTPUT_BIN}/kubectl
NAMESPACE=pulsar
CLUSTER=pulsar-ci
CLUSTER_ID=$(uuidgen)

function ci::create_cluster() {
    echo "Creating a kind cluster ..."
    ${FUNCTION_MESH_HOME}/hack/kind-cluster-build.sh --name pulsar-ci-${CLUSTER_ID} -c 1 -v 10
    echo "Successfully created a kind cluster."
}

function ci::delete_cluster() {
    echo "Deleting a kind cluster ..."
    kind delete cluster --name=pulsar-ci-${CLUSTER_ID}
    echo "Successfully delete a kind cluster."
}

function ci::install_storage_provisioner() {
    echo "Installing the local storage provisioner ..."
    ${HELM} repo add streamnative https://charts.streamnative.io
    ${HELM} repo update
    ${HELM} install local-storage-provisioner streamnative/local-storage-provisioner
    WC=$(${KUBECTL} get pods --field-selector=status.phase=Running | grep local-storage-provisioner | wc -l)
    while [[ ${WC} -lt 1 ]]; do
      echo ${WC};
      sleep 15
      ${KUBECTL} get pods --field-selector=status.phase=Running
      WC=$(${KUBECTL} get pods --field-selector=status.phase=Running | grep local-storage-provisioner | wc -l)
    done
    echo "Successfully installed the local storage provisioner."
}

function ci::install_pulsar_charts() {
    echo "Installing the pulsar charts ..."
    git clone https://github.com/streamnative/charts.git
    cd charts
    ./scripts/pulsar/prepare_helm_release.sh -n ${NAMESPACE} -k function-mesh -c
    cp ../.ci/clusters/values.yaml charts/pulsar/mini_values.yaml
    cd charts
    helm repo add loki https://grafana.github.io/loki/charts
    helm dependency update pulsar
    ${HELM} install sn-platform --values ./pulsar/mini_values.yaml ./pulsar

    echo "wait until broker is alive"
    WC=$(${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep ${CLUSTER}-broker | wc -l)
    while [[ ${WC} -lt 1 ]]; do
      echo ${WC};
      sleep 15
      ${KUBECTL} get pods -n ${NAMESPACE}
      WC=$(${KUBECTL} get pods -n ${NAMESPACE} | grep ${CLUSTER}-broker | wc -l)
      if [[ ${WC} -gt 1 ]]; then
        ${KUBECTL} describe pod -n ${NAMESPACE} pulsar-ci-broker-0
        ${KUBECTL} logs -n ${NAMESPACE} pulsar-ci-broker-0
      fi
      WC=$(${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep ${CLUSTER}-broker | wc -l)
    done

    ${KUBECTL} get service -n ${NAMESPACE}
}

function ci::test_pulsar_producer() {
    sleep 120
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-toolset-0 -- bash -c 'until nslookup pulsar-ci-broker; do sleep 3; done'
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-toolset-0 -- bash -c 'until nslookup pulsar-ci-proxy; do sleep 3; done'
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-bookie-0 -- df -h
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-bookie-0 -- cat conf/bookkeeper.conf
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-toolset-0 -- bin/bookkeeper shell listbookies -rw
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-toolset-0 -- bin/bookkeeper shell listbookies -ro
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-toolset-0 -- bin/pulsar-admin tenants create pulsar-ci
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-toolset-0 -- bin/pulsar-admin namespaces create pulsar-ci/test
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-toolset-0 -- bin/pulsar-client produce -m "test-message" pulsar-ci/test/test-topic
}

function ci::wait_function_running() {
    num_running=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-toolset-0 -- bash -c 'bin/pulsar-admin functions status --tenant pulsar-ci --namespace test --name test-function | bin/jq .numRunning') 
    while [[ ${num_running} -lt 1 ]]; do
      echo ${num_running}
      sleep 15
      ${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running
      num_running=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-toolset-0 -- bash -c 'bin/pulsar-admin functions status --tenant pulsar-ci --namespace test --name test-function | bin/jq .numRunning') 
    done
}

function ci::wait_message_processed() {
    num_processed=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-toolset-0 -- bash -c 'bin/pulsar-admin functions stats --tenant pulsar-ci --namespace test --name test-function | bin/jq .processedSuccessfullyTotal') 
    while [[ ${num_processed} -lt 1 ]]; do
      echo ${num_processed}
      sleep 15
      ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-toolset-0 -- bin/pulsar-admin functions stats --tenant pulsar-ci --namespace test --name test-function
      num_processed=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-toolset-0 -- bash -c 'bin/pulsar-admin functions stats --tenant pulsar-ci --namespace test --name test-function | bin/jq .processedSuccessfullyTotal') 
    done
}

function ci::test_pulsar_function() {
    sleep 120
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-toolset-0 -- bash -c 'until nslookup pulsar-ci-broker; do sleep 3; done'
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-toolset-0 -- bash -c 'until nslookup pulsar-ci-proxy; do sleep 3; done'
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-bookie-0 -- df -h
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-toolset-0 -- bin/bookkeeper shell listbookies -rw
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-toolset-0 -- bin/bookkeeper shell listbookies -ro
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-toolset-0 -- curl --retry 10 -L -o bin/jq https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-toolset-0 -- chmod +x bin/jq
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-toolset-0 -- bin/pulsar-admin functions create --tenant pulsar-ci --namespace test --name test-function --inputs "pulsar-ci/test/test_input" --output "pulsar-ci/test/test_output" --parallelism 1 --classname org.apache.pulsar.functions.api.examples.ExclamationFunction --jar /pulsar/examples/api-examples.jar

    # wait until the function is running
    # TODO: re-enable function test
    # ci::wait_function_running
    # ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-toolset-0 -- bin/pulsar-client produce -m "hello pulsar function!" pulsar-ci/test/test_input
    # ci::wait_message_processed
}