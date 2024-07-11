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

manifests_file="${BASE_DIR}"/.ci/tests/integration/cases/global-and-namespaced-config/manifests.yaml
mesh_config_file="${BASE_DIR}"/.ci/tests/integration/cases/global-and-namespaced-config/mesh-config.yaml
mesh_config_file_in_kube_system="${BASE_DIR}"/.ci/tests/integration/cases/global-and-namespaced-config/mesh-config-kube-system.yaml
global_mesh_config_file="${BASE_DIR}"/.ci/clusters/global_backend_config.yaml


kubectl apply -f "${mesh_config_file}" > /dev/null 2>&1
kubectl apply -f "${manifests_file}" > /dev/null 2>&1

verify_fm_result=$(ci::verify_function_mesh function-sample-env 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_fm_result"
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

# if the function defines the same env, it will use the value from the function instead of the backend config
verify_env_result=$(ci::verify_env "function-sample-env-function-0" podenv podenv=podvalue 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_env_result"
  kubectl delete -f "${mesh_config_file}" > /dev/null 2>&1
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

verify_env_result=$(ci::verify_env "function-sample-env-function-0" global1 global1=globalvalue1 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_env_result"
  kubectl delete -f "${mesh_config_file}" > /dev/null 2>&1
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

verify_env_result=$(ci::verify_env "function-sample-env-function-0" namespaced1 namespaced1=namespacedvalue1 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_env_result"
  kubectl delete -f "${mesh_config_file}" > /dev/null 2>&1
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

# if global and namespaced config has same key, the value from namespace should be used
verify_env_result=$(ci::verify_env "function-sample-env-function-0" shared1 shared1=fromnamespace 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_env_result"
  kubectl delete -f "${mesh_config_file}" > /dev/null 2>&1
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

# verify liveness config
verify_liveness_result=$(ci::verify_liveness_probe function-sample-env-function-0 '{"failureThreshold":3,"httpGet":{"path":"/","port":9094,"scheme":"HTTP"},"initialDelaySeconds":30,"periodSeconds":10,"successThreshold":1,"timeoutSeconds":10}' 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_liveness_result"
  kubectl delete -f "${mesh_config_file}" > /dev/null 2>&1
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

# update the namespaced config, it should trigger the reconcile since the autoUpdate is true
kubectl patch BackendConfig backend-config --type='json' -p='[{"op": "replace", "path": "/spec/env/shared1", "value": "newvalue"}]'
sleep 30

verify_env_result=$(ci::verify_env "function-sample-env-function-0" shared1 shared1=newvalue 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_env_result"
  kubectl delete -f "${mesh_config_file}" > /dev/null 2>&1
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

# the liveness probe should also be updated
kubectl patch BackendConfig backend-config --type='json' -p='[{"op": "replace", "path": "/spec/pod/liveness/initialDelaySeconds", "value": 20}]'
sleep 30

verify_liveness_result=$(ci::verify_liveness_probe function-sample-env-function-0 '{"failureThreshold":3,"httpGet":{"path":"/","port":9094,"scheme":"HTTP"},"initialDelaySeconds":20,"periodSeconds":10,"successThreshold":1,"timeoutSeconds":10}' 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_liveness_result"
  kubectl delete -f "${mesh_config_file}" > /dev/null 2>&1
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

# delete the namespaced config, the function should be reconciled without namespaced env injected
kubectl delete -f "${mesh_config_file}" > /dev/null 2>&1
sleep 30

verify_fm_result=$(ci::verify_function_mesh function-sample-env 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_fm_result"
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

verify_env_result=$(ci::verify_env "function-sample-env-function-0" global1 global1=globalvalue1 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_env_result"
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

verify_env_result=$(ci::verify_env "function-sample-env-function-0" shared1 shared1=fromglobal 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_env_result"
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

verify_env_result=$(ci::verify_env "function-sample-env-function-0" namespaced1 "" 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_env_result"
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

# it should use liveness config from global config
verify_liveness_result=$(ci::verify_liveness_probe function-sample-env-function-0 '{"failureThreshold":3,"httpGet":{"path":"/","port":9094,"scheme":"HTTP"},"initialDelaySeconds":10,"periodSeconds":30,"successThreshold":1,"timeoutSeconds":30}' 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_liveness_result"
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

# delete the global config, the function should be reconciled without global env injected
kubectl delete -f "${global_mesh_config_file}" -n $FUNCTION_MESH_NAMESPACE > /dev/null 2>&1 || true
sleep 30

verify_fm_result=$(ci::verify_function_mesh function-sample-env 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_fm_result"
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

verify_env_result=$(ci::verify_env "function-sample-env-function-0" global1 "" 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_env_result"
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

# it should has no liveness config
verify_liveness_result=$(ci::verify_liveness_probe function-sample-env-function-0 "" 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_liveness_result"
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

# config created in an another namespace should not affect functions in other namespaces
kubectl apply -f "${mesh_config_file_in_kube_system}" > /dev/null 2>&1
sleep 30

verify_fm_result=$(ci::verify_function_mesh function-sample-env 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_fm_result"
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

# it should has no liveness config
verify_liveness_result=$(ci::verify_liveness_probe function-sample-env-function-0 "" 2>&1)
if [ $? -ne 0 ]; then
  echo "$verify_liveness_result"
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

verify_env_result=$(ci::verify_env "function-sample-env-function-0" namespaced1 "" 2>&1)
if [ $? -eq 0 ]; then
  echo "e2e-test: ok" | yq eval -
else
  echo "$verify_env_result"
  kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
  exit 1
fi

kubectl delete -f "${manifests_file}" > /dev/null 2>&1 || true
