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
    ${FUNCTION_MESH_HOME}/hack/kind-cluster-build.sh --name sn-platform-${CLUSTER_ID} -c 3 -v 10
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

function ci::install_metrics_server() {
    echo "install metrics-server"
    ${KUBECTL} apply -f https://github.com/kubernetes-sigs/metrics-server/releases/download/v0.3.7/components.yaml
    ${KUBECTL} patch deployment metrics-server -n kube-system -p '{"spec":{"template":{"spec":{"containers":[{"name":"metrics-server","args":["--cert-dir=/tmp", "--secure-port=4443", "--kubelet-insecure-tls","--kubelet-preferred-address-types=InternalIP"]}]}}}}'
    echo "Successfully installed the metrics-server."
    WC=$(${KUBECTL} get pods -n kube-system --field-selector=status.phase=Running | grep metrics-server | wc -l)
    while [[ ${WC} -lt 1 ]]; do
      echo ${WC};
      sleep 20
      ${KUBECTL} get pods -n kube-system
      WC=$(${KUBECTL} get pods -n kube-system --field-selector=status.phase=Running | grep metrics-server | wc -l)
    done
}

function ci::install_pulsar_charts() {
    echo "Installing the pulsar charts ..."
    values=${1:-".ci/clusters/values.yaml"}
    echo $values
    if [ "$values" = ".ci/clusters/values_mesh_worker_service.yaml" ]; then
      echo "load mesh-worker-service-integration-pulsar:latest ..."
      kind load docker-image mesh-worker-service-integration-pulsar:latest --name sn-platform-${CLUSTER_ID}
    fi
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
    ${KUBECTL} logs -lname=${FUNCTION_NAME}  --all-containers=true
}

function ci::verify_hpa() {
    FUNCTION_NAME=$1
    ${KUBECTL} get function
    ${KUBECTL} get function ${FUNCTION_NAME} -o yaml
    ${KUBECTL} get hpa.v2beta2.autoscaling
    ${KUBECTL} get hpa.v2beta2.autoscaling ${FUNCTION_NAME}-function -o yaml
    ${KUBECTL} describe hpa.v2beta2.autoscaling ${FUNCTION_NAME}-function
    WC=$(${KUBECTL} get hpa.v2beta2.autoscaling ${FUNCTION_NAME}-function -o jsonpath='{.status.conditions[?(@.type=="AbleToScale")].status}' | grep False | wc -l)
    while [[ ${WC} -lt 0 ]]; do
      echo ${WC};
      sleep 20
      ${KUBECTL} get hpa.v2beta2.autoscaling ${FUNCTION_NAME}-function -o yaml
      ${KUBECTL} describe hpa.v2beta2.autoscaling ${FUNCTION_NAME}-function
      ${KUBECTL} get hpa.v2beta2.autoscaling ${FUNCTION_NAME}-function -o jsonpath='{.status.conditions[?(@.type=="AbleToScale")].status}'
      WC=$(${KUBECTL} get hpa.v2beta2.autoscaling ${FUNCTION_NAME}-function -o jsonpath='{.status.conditions[?(@.type=="AbleToScale")].status}' | grep False | wc -l)
    done
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
    sleep 120
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

function ci::ensure_mesh_worker_service_role() {
  ${KUBECTL} apply -f ${FUNCTION_MESH_HOME}/config/rbac/role.yaml
  ${KUBECTL} create clusterrolebinding broker-acct-manager-role-binding --clusterrole=manager-role --serviceaccount=default:sn-platform-pulsar-broker-acct
}

function ci::ensure_function_mesh_config() {
  ${KUBECTL} apply -f ${FUNCTION_MESH_HOME}/.ci/clusters/mesh_worker_service_integration_test_pulsar_config.yaml
}

function ci::verify_mesh_worker_service_pulsar_admin() {
  ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin sinks available-sinks
  RET=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin sinks available-sinks)
  if [[ $RET != *"data-generator"* ]]; then
   return 1
  fi
  ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin sources available-sources
  RET=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin sources available-sources)
  if [[ $RET != *"data-generator"* ]]; then
   return 1
  fi
  RET=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin sources create --name data-generator-source --source-type data-generator --destination-topic-name persistent://public/default/random-data-topic --custom-runtime-options '{"outputTypeClassName": "org.apache.pulsar.io.datagenerator.Person"}' --source-config '{"sleepBetweenMessages": "1000"}')
  echo $RET
  if [[ $RET != *"successfully"* ]]; then
   return 1
  fi
  RET=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin sinks create --name data-generator-sink --sink-type data-generator --inputs persistent://public/default/random-data-topic --custom-runtime-options '{"inputTypeClassName": "org.apache.pulsar.io.datagenerator.Person"}')
  echo $RET
  if [[ $RET != *"successfully"* ]]; then
   return 1
  fi
  ${KUBECTL} get pods -n ${NAMESPACE}
  sleep 120
  ${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep "data-generator-source"
  WC=$(${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep "data-generator-sink" | wc -l)
  while [[ ${WC} -lt 1 ]]; do
    echo ${WC};
    sleep 20
    ${KUBECTL} get pods -n ${NAMESPACE}
    WC=$(${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep "data-generator-sink" | wc -l)
  done
  ${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep "data-generator-source"
  WC=$(${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep "data-generator-source" | wc -l)
  while [[ ${WC} -lt 1 ]]; do
    echo ${WC};
    sleep 20
    ${KUBECTL} get pods -n ${NAMESPACE}
    WC=$(${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep "data-generator-source" | wc -l)
  done
  sleep 120
  ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin sinks status --name data-generator-sink
  RET=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin sinks status --name data-generator-sink)
  if [[ $RET != *"true"* ]]; then
   return 1
  fi
  ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin sources status --name data-generator-source
  RET=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin sources status --name data-generator-source)
  if [[ $RET != *"true"* ]]; then
   return 1
  fi
  ${KUBECTL} get pods -n ${NAMESPACE}
  RET=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin sources delete --name data-generator-source)
  echo $RET
  RET=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin sinks delete --name data-generator-sink)
  echo $RET
  ${KUBECTL} get pods -n ${NAMESPACE}
  echo " === verify mesh worker service with empty connector config"
  RET=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin sources create --name data-generator-source --source-type data-generator --destination-topic-name persistent://public/default/random-data-topic --custom-runtime-options '{"outputTypeClassName": "org.apache.pulsar.io.datagenerator.Person"}')
  echo $RET
  if [[ $RET != *"successfully"* ]]; then
   return 1
  fi
  WC=$(${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep "data-generator-source" | wc -l)
  while [[ ${WC} -lt 1 ]]; do
    echo ${WC};
    sleep 20
    ${KUBECTL} get pods -n ${NAMESPACE}
    WC=$(${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep "data-generator-source" | wc -l)
  done
  ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin sources status --name data-generator-source
  RET=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin sources status --name data-generator-source)
  if [[ $RET != *"true"* ]]; then
    ${KUBECTL} logs -n ${NAMESPACE} data-generator-source-69865103-source-0
    ${KUBECTL} get pods data-generator-source-69865103-source-0 -o yaml
   return 1
  fi
  ${KUBECTL} get pods -n ${NAMESPACE}
  RET=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin sources delete --name data-generator-source)
  echo $RET
  ${KUBECTL} get pods -n ${NAMESPACE}

}

function ci::upload_java_package() {

  RET=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin packages upload function://public/default/java-function@1.0 --path /pulsar/examples/api-examples.jar --description java-function@1.0)
  if [[ $RET != *"successfully"* ]]; then
    echo "${RET}"
    ${KUBECTL} logs -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0
    sleep 60
    return 1
  fi
  RET=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin packages download function://public/default/java-function@1.0 --path /pulsar/api-examples.jar)
  if [[ $RET != *"successfully"* ]]; then
    echo "${RET}"
    ${KUBECTL} logs -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0
    sleep 60
    return 1
  fi
}

function ci::verify_java_package() {
  RET=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin functions create --jar function://public/default/java-function@1.0 --name package-java-fn --className org.apache.pulsar.functions.api.examples.ExclamationFunction --inputs persistent://public/default/package-java-fn-input --cpu 0.1)
  ${KUBECTL} logs -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0
  sleep 15
  echo $RET
  ${KUBECTL} logs -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0
  sleep 15
  ${KUBECTL} get pods -A
  sleep 5
  WC=$(${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep "package-java-fn" | wc -l)
  while [[ ${WC} -lt 1 ]]; do
    echo ${WC};
    sleep 20
    ${KUBECTL} get pods -n ${NAMESPACE}
    ${KUBECTL} describe pod package-java-fn-function-0
    WC=$(${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep "package-java-fn" | wc -l)
  done
  echo "java function test done"
  RET=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin functions delete --name package-java-fn)
  echo "${RET}"
}

function ci::upload_python_package() {
  RET=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin packages upload function://public/default/python-function@1.0 --path /pulsar/examples/python-examples/exclamation_function.py --description python-function@1.0)
  if [[ $RET != *"successfully"* ]]; then
    echo "${RET}"
    return 1
  fi
}

function ci::verify_python_package() {
  RET=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin functions create --py function://public/default/python-function@1.0 --name package-python-fn --classname exclamation_function.ExclamationFunction --inputs persistent://public/default/package-python-fn-input --cpu 0.1)
  if [[ $RET != *"successfully"* ]]; then
    echo "${RET}"
    return 1
  fi
  sleep 15
  ${KUBECTL} get pods -A
  sleep 5
  WC=$(${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep "package-python-fn" | wc -l)
  while [[ ${WC} -lt 1 ]]; do
    echo ${WC};
    sleep 20
    ${KUBECTL} get pods -n ${NAMESPACE}
    ${KUBECTL} describe pod package-python-fn-function-0
    WC=$(${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep "package-python-fn" | wc -l)
  done
  echo "python function test done"
  RET=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin functions delete --name package-python-fn)
  echo "${RET}"
}

function ci::upload_go_package() {
  ${KUBECTL} cp "${FUNCTION_MESH_HOME}/.ci/examples/go-examples" "${NAMESPACE}/${CLUSTER}-pulsar-broker-0:/pulsar/examples"
  sleep 1
  ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- ls -l /pulsar/examples/go-examples
  sleep 1
  RET=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin packages upload function://public/default/go-function@1.0 --path /pulsar/examples/go-examples/exclamationFunc --description go-function@1.0)
  if [[ $RET != *"successfully"* ]]; then
    echo "${RET}"
    return 1
  fi
}

function ci::verify_go_package() {
  RET=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin functions create --go function://public/default/go-function@1.0 --name package-go-fn --inputs persistent://public/default/package-go-fn-input --cpu 0.1)
  if [[ $RET != *"successfully"* ]]; then
    echo "${RET}"
    return 1
  fi
  sleep 15
  ${KUBECTL} get pods -A
  sleep 5
  WC=$(${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep "package-go-fn" | wc -l)
  while [[ ${WC} -lt 1 ]]; do
    echo ${WC};
    sleep 20
    ${KUBECTL} get pods -n ${NAMESPACE}
    ${KUBECTL} describe pod package-go-fn-function-0
    WC=$(${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep "package-go-fn" | wc -l)
  done
  echo "go function test done"
  RET=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin functions delete --name package-go-fn)
  echo "${RET}"
}

function ci::create_java_function_by_upload() {
  ${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- cat conf/functions_worker.yml
  RET=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin functions create --jar /pulsar/examples/api-examples.jar --name package-upload-java-fn --className org.apache.pulsar.functions.api.examples.ExclamationFunction --inputs persistent://public/default/package-upload-java-fn-input --cpu 0.1)
  ${KUBECTL} logs -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0
  sleep 15
  echo $RET
  ${KUBECTL} logs -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0
  sleep 15
  ${KUBECTL} get pods -A
  sleep 5
  WC=$(${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep "package-upload-java-fn" | wc -l)
  while [[ ${WC} -lt 1 ]]; do
    echo ${WC};
    sleep 20
    ${KUBECTL} get pods -n ${NAMESPACE}
    ${KUBECTL} describe pod package-upload-java-fn-function-0
    WC=$(${KUBECTL} get pods -n ${NAMESPACE} --field-selector=status.phase=Running | grep "package-upload-java-fn" | wc -l)
  done
  echo "java function test done"
  RET=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin functions delete --name package-upload-java-fn)
  echo "${RET}"
  RET=$(${KUBECTL} exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin packages get-metadata function://public/default/package-upload-java-fn@latest)
  echo "${RET}"
}