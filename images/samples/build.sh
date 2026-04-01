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
set -euo pipefail

PULSAR_IMAGE_TAG=${PULSAR_IMAGE_TAG:-"2.7.1"}
PULSAR_IMAGE=${PULSAR_IMAGE:-"streamnative/sn-platform"}
JAVA_RUNNER_IMAGE=${RUNNER_IMAGE:-"streamnative/pulsar-functions-java-runner"}
DOCKER_REPO=${DOCKER_REPO:-"streamnative"}
JAVA_SAMPLE="pulsar-functions-java-sample"
PYTHON_SAMPLE="pulsar-functions-python-sample"
GO_SAMPLE="pulsar-functions-go-sample"
CONNECTOR_ES_SAMPLE="pulsar-io-elasticsearch"
KIND_PUSH=${KIND_PUSH:-false}
CI_TEST=${CI_TEST:-false}
PLATFORMS=${PLATFORMS:-"linux/amd64"}
PRIMARY_PLATFORM=${PLATFORMS%%,*}

JAVA_SAMPLE_IMAGE="${DOCKER_REPO}/${JAVA_SAMPLE}:${PULSAR_IMAGE_TAG}"
PYTHON_SAMPLE_IMAGE="${DOCKER_REPO}/${PYTHON_SAMPLE}:${PULSAR_IMAGE_TAG}"
GO_SAMPLE_IMAGE="${DOCKER_REPO}/${GO_SAMPLE}:${PULSAR_IMAGE_TAG}"
CONNECTOR_ES_SAMPLE_IMAGE="${DOCKER_REPO}/${CONNECTOR_ES_SAMPLE}:${PULSAR_IMAGE_TAG}"

MULTI_PLATFORM=false
if [[ "${PLATFORMS}" == *,* ]]; then
  MULTI_PLATFORM=true
fi

PUSH_DEFAULT=false
if [[ "${DOCKER_REPO}" == "localhost:5000" || "${MULTI_PLATFORM}" == "true" ]]; then
  PUSH_DEFAULT=true
fi
PUSH=${PUSH:-$PUSH_DEFAULT}

if [[ "${MULTI_PLATFORM}" == "true" && "${PUSH}" != "true" ]]; then
  echo "multi-platform sample builds require PUSH=true" >&2
  exit 1
fi

if [[ "${MULTI_PLATFORM}" == "true" && "${KIND_PUSH}" == "true" ]]; then
  echo "KIND_PUSH=true is only supported for single-platform sample builds" >&2
  exit 1
fi

build_image() {
  local image=$1
  local context=$2
  shift 2

  if [[ "${MULTI_PLATFORM}" == "true" ]]; then
    docker buildx build --platform "${PLATFORMS}" --push -t "${image}" "$@" "${context}"
  else
    docker build --platform "${PRIMARY_PLATFORM}" -t "${image}" "$@" "${context}"
  fi
}

echo "build java sample"
build_image "${JAVA_SAMPLE_IMAGE}" images/samples/java-function-samples --build-arg PULSAR_IMAGE_TAG="${PULSAR_IMAGE_TAG}"

echo "build python sample"
build_image "${PYTHON_SAMPLE_IMAGE}" images/samples/python-function-samples --build-arg PULSAR_IMAGE_TAG="${PULSAR_IMAGE_TAG}"

echo "build go sample"
build_image "${GO_SAMPLE_IMAGE}" images/samples/go-function-samples --build-arg PULSAR_IMAGE_TAG="${PULSAR_IMAGE_TAG}"

echo "build connector sample"
build_image "${CONNECTOR_ES_SAMPLE_IMAGE}" images/samples/pulsar-io-connector/pulsar-io-elasticsearch \
  --build-arg PULSAR_IMAGE_TAG="${PULSAR_IMAGE_TAG}" \
  --build-arg PULSAR_IMAGE="${PULSAR_IMAGE}" \
  --build-arg RUNNER_IMAGE="${JAVA_RUNNER_IMAGE}"

if [ "$KIND_PUSH" = true ] ; then
  echo "push images to kind"
  clusters=$(kind get clusters)
  echo $clusters
  for cluster in $clusters
  do
    kind load docker-image "${JAVA_SAMPLE_IMAGE}" --name $cluster
    kind load docker-image "${PYTHON_SAMPLE_IMAGE}" --name $cluster
    kind load docker-image "${GO_SAMPLE_IMAGE}" --name $cluster
  done
fi

if [[ "${PUSH}" == "true" && "${MULTI_PLATFORM}" != "true" ]]; then
  docker push "${JAVA_SAMPLE_IMAGE}"
  docker push "${PYTHON_SAMPLE_IMAGE}"
  docker push "${GO_SAMPLE_IMAGE}"
  docker push "${CONNECTOR_ES_SAMPLE_IMAGE}"
fi
