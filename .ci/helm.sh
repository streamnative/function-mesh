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
NAMESPACE=default
CLUSTER=sn-platform
CLUSTER_ID=$(uuidgen | tr "[:upper:]" "[:lower:]")

FUNCTION_NAME=$1

function ci::create_cluster() {
    echo "Creating a kind cluster ..."
    ${FUNCTION_MESH_HOME}/hack/kind-cluster-build.sh --name sn-platform-${CLUSTER_ID} -c 1 -v 10
    echo "Successfully created a kind cluster."
}

function ci::delete_cluster() {
    echo "Deleting a kind cluster ..."
    kind delete cluster --name=sn-platform-${CLUSTER_ID}
    echo "Successfully delete a kind cluster."
}

function ci::cleanup() {
    echo "Clean up kind clusters ..."
    clusters=( $(kind get clusters | grep sn-platform) )
    for cluster in "${clusters[@]}"
    do
       echo "Deleting a kind cluster ${cluster}"
       kind delete cluster --name=${cluster}
    done
    echo "Successfully clean up a kind cluster."
}

function ci::install_storage_provisioner() {
    echo "Installing the local storage provisioner ..."
    ${HELM} repo add streamnative https://charts.streamnative.io
    ${HELM} repo update
    ${HELM} install local-storage-provisioner streamnative/local-storage-provisioner --debug --wait --set namespace=default
    echo "Successfully installed the local storage provisioner."
}

function ci::install_pulsar_charts() {
    echo "Installing the pulsar charts ..."
    values=${1:-".ci/clusters/values.yaml"}
    echo $values
    if [ -d "pulsar-charts" ]; then
      rm -rf pulsar-charts
    fi
    git clone https://github.com/streamnative/charts.git pulsar-charts
    cp ${values} pulsar-charts/charts/pulsar/mini_values.yaml
    cd pulsar-charts
    cd charts
    helm repo add loki https://grafana.github.io/loki/charts
    helm dependency update pulsar
    ${HELM} install sn-platform --values ./pulsar/mini_values.yaml ./pulsar --debug

    echo "wait until broker is alive"
    WC=$(${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep ${CLUSTER}-pulsar-broker | wc -l)
    while [[ ${WC} -lt 1 ]]; do
      echo ${WC};
      sleep 20
      ${KUBECTL} get pods -n ${NAMESPACE}
      WC=$(${KUBECTL} get pods -n ${NAMESPACE} | grep ${CLUSTER}-pulsar-broker | wc -l)
      if [[ ${WC} -gt 1 ]]; then
        ${KUBECTL} describe pod -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0
        ${KUBECTL} describe pod -n ${NAMESPACE} ${CLUSTER}-pulsar-bookie-0
      fi
      WC=$(${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep ${CLUSTER}-pulsar-broker | wc -l)
    done

    ${KUBECTL} get service -n ${NAMESPACE}
}

function ci::test_pulsar_producer() {
    sleep 120
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-toolset-0 -- bash -c 'until nslookup sn-platform-pulsar-broker; do sleep 3; done'
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-bookie-0 -- df -h
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-bookie-0 -- cat conf/bookkeeper.conf
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin tenants create sn-platform
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin namespaces create sn-platform/test
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-client produce -m "test-message" sn-platform/test/test-topic
}

function ci::verify_function_mesh() {
    FUNCTION_NAME=$1
    WC=$(${KUBECTL} get pods -lname=${FUNCTION_NAME} --field-selector=status.phase=Running | wc -l)
    while [[ ${WC} -lt 1 ]]; do
      echo ${WC};
      sleep 15
      ${KUBECTL} get pods -A
      WC=$(${KUBECTL} get pods -lname=${FUNCTION_NAME} | wc -l)
      if [[ ${WC} -gt 1 ]]; then
        ${KUBECTL} describe pod -lname=${FUNCTION_NAME}
      fi
      WC=$(${KUBECTL} get pods -lname=${FUNCTION_NAME} --field-selector=status.phase=Running | wc -l)
    done
    ${KUBECTL} describe pod -lname=${FUNCTION_NAME}
}

function ci::test_function_runners() {
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin functions create --tenant public --namespace default --name test-java --className org.apache.pulsar.functions.api.examples.ExclamationFunction --inputs persistent://public/default/test-java-input --jar /pulsar/examples/api-examples.jar --cpu 0.1
    sleep 15
    ${KUBECTL} get pods -A
    sleep 5
    WC=$(${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep "test-java" | wc -l)
    while [[ ${WC} -lt 1 ]]; do
      echo ${WC};
      sleep 20
      ${KUBECTL} get pods -n ${NAMESPACE}
      ${KUBECTL} describe pod pf-public-default-test-java-0 
      WC=$(${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep "test-java" | wc -l)
    done
    echo "java runner test done"
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin functions delete --tenant public --namespace default --name test-java

    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin functions create --tenant public --namespace default --name test-python --classname exclamation_function.ExclamationFunction --inputs persistent://public/default/test-python-input --py /pulsar/examples/python-examples/exclamation_function.py --cpu 0.1
    sleep 15
    ${KUBECTL} get pods -A
    sleep 5
    WC=$(${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep "test-python" | wc -l)
    while [[ ${WC} -lt 1 ]]; do
      echo ${WC};
      sleep 20
      ${KUBECTL} get pods -n ${NAMESPACE}
      WC=$(${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep "test-python" | wc -l)
    done
    echo "python runner test done"
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin functions delete --tenant public --namespace default --name test-python

    ${KUBECTL} cp "${FUNCTION_MESH_HOME}/.ci/examples/go-examples" "${NAMESPACE}/${CLUSTER}-pulsar-broker-0:/pulsar/examples"
    sleep 1
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin functions create --tenant public --namespace default --name test-go --inputs persistent://public/default/test-go-input --go /pulsar/examples/go-examples/exclamationFunc --cpu 0.1
    sleep 15
    ${KUBECTL} get pods -A
    sleep 5
    WC=$(${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep "test-go" | wc -l)
    while [[ ${WC} -lt 1 ]]; do
      echo ${WC};
      sleep 20
      ${KUBECTL} get pods -n ${NAMESPACE}
      WC=$(${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep "test-go" | wc -l)
    done
    echo "golang runner test done"
    ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin functions delete --tenant public --namespace default --name test-go
}

function ci::verify_go_function() {
    FUNCTION_NAME=$1
    ${KUBECTL} describe pod -lname=${FUNCTION_NAME}
    ${KUBECTL} logs -lname=${FUNCTION_NAME}  --all-containers=true
    ci:verify_exclamation_function "persistent://public/default/input-go-topic" "persistent://public/default/output-go-topic" "test-message" "test-message!" 30
}

function ci::verify_java_function() {
    FUNCTION_NAME=$1
    ${KUBECTL} describe pod -lname=${FUNCTION_NAME}
    sleep 120
    ${KUBECTL} logs -lname=${FUNCTION_NAME}  --all-containers=true
    ci:verify_exclamation_function "persistent://public/default/input-java-topic" "persistent://public/default/output-java-topic" "test-message" "test-message!" 30
}

function ci::verify_python_function() {
    FUNCTION_NAME=$1
    ${KUBECTL} describe pod -lname=${FUNCTION_NAME}
    ${KUBECTL} logs -lname=${FUNCTION_NAME}  --all-containers=true
    ci:verify_exclamation_function "persistent://public/default/input-python-topic" "persistent://public/default/output-python-topic" "test-message" "test-message!" 30
}

function ci::verify_mesh_function() {
    ci:verify_exclamation_function "persistent://public/default/functionmesh-input-topic" "persistent://public/default/functionmesh-python-topic" "test-message" "test-message!!!" 120
}

function ci::print_function_log() {
    FUNCTION_NAME=$1
    ${KUBECTL} describe pod -lname=${FUNCTION_NAME}
    ${KUBECTL} logs -lname=${FUNCTION_NAME}  --all-containers=true
}

function ci:verify_exclamation_function() {
  inputtopic=$1
  outputtopic=$2
  inputmessage=$3
  outputmessage=$4
  timesleep=$5
  ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-client produce -m ${inputmessage} -n 1 ${inputtopic}
  sleep $timesleep
  MESSAGE=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-client consume -n 1 -s "sub" --subscription-position Earliest ${outputtopic})
  echo $MESSAGE
  if [[ "$MESSAGE" == *"$outputmessage"* ]]; then
    return 0
  fi
  return 1
}
