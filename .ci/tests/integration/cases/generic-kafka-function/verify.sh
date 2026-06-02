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
BASE_DIR=$(cd "${E2E_DIR}"/../../../../..; pwd)
PULSAR_NAMESPACE=${PULSAR_NAMESPACE:-"default"}
PULSAR_RELEASE_NAME=${PULSAR_RELEASE_NAME:-"sn-platform"}
E2E_KUBECONFIG=${E2E_KUBECONFIG:-"/tmp/e2e-k8s.config"}

source "${BASE_DIR}"/.ci/helm.sh

if [ ! "$KUBECONFIG" ]; then
  export KUBECONFIG=${E2E_KUBECONFIG}
fi

manifests_file="${BASE_DIR}"/.ci/tests/integration/cases/generic-kafka-function/manifests.yaml
kafka_client_file="${BASE_DIR}"/.ci/tests/integration/cases/generic-kafka-function/kafka-client.yaml
kafka_bootstrap_server=sn-platform-pulsar-broker-0.sn-platform-pulsar-broker.default.svc.cluster.local:9092
kafka_properties_file=kafka.properties
input_topic=input-kafka-topic
output_topic=output-kafka-topic
consumer_group=generic-kafka-function-sub
input_message="test-message-${RANDOM}-$(date +%s)"
output_message="${input_message}!"

kubectl apply -f "${kafka_client_file}" > /dev/null 2>&1
kubectl wait pod kafka-client --for=condition=Ready --timeout=2m > /dev/null || {
  kubectl get pod kafka-client -o yaml >&2 || true
  kubectl delete -f "${kafka_client_file}" > /dev/null 2>&1 || true
  exit 1
}

ci::ensure_kafka_topic "${input_topic}" "${kafka_bootstrap_server}" "${kafka_properties_file}" > /dev/null 2>&1
ci::ensure_kafka_topic "${output_topic}" "${kafka_bootstrap_server}" "${kafka_properties_file}" > /dev/null 2>&1
ci::wait_kafka_topic_ready "${input_topic}" "${kafka_bootstrap_server}" "${kafka_properties_file}" > /dev/null 2>&1
ci::wait_kafka_topic_ready "${output_topic}" "${kafka_bootstrap_server}" "${kafka_properties_file}" > /dev/null 2>&1
ci::wait_kafka_group_coordinator_ready "${consumer_group}" "${kafka_bootstrap_server}" "${kafka_properties_file}" > /dev/null 2>&1
kubectl apply -f "${manifests_file}" > /dev/null 2>&1

kubectl wait -l compute.functionmesh.io/name=generic-kafka-function --for=condition=Ready pod --timeout=2m > /dev/null || {
  kubectl get pods -l compute.functionmesh.io/name=generic-kafka-function >&2
  kubectl logs -l compute.functionmesh.io/name=generic-kafka-function --tail=100 >&2 || true
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  kubectl delete -f "${kafka_client_file}" > /dev/null 2>&1 || true
  exit 1
}

set +e
verify_result=$(NAMESPACE=${PULSAR_NAMESPACE} CLUSTER=${PULSAR_RELEASE_NAME} \
  ci::verify_kafka_exclamation_function "${input_topic}" "${output_topic}" \
  "${input_message}" "${output_message}" "${kafka_bootstrap_server}" "${kafka_properties_file}" 2>&1)
verify_status=$?
set -e

if [ ${verify_status} -eq 0 ]; then
  echo "e2e-test: ok" | yq eval -
else
  echo "failed to verify: $verify_result" >&2
  kubectl logs -l compute.functionmesh.io/name=generic-kafka-function --tail=100 >&2 || true
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  kubectl delete -f "${kafka_client_file}" > /dev/null 2>&1 || true
  exit 1
fi

kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
kubectl delete -f "${kafka_client_file}" > /dev/null 2>&1 || true
