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

PULSAR_IMAGE=${PULSAR_IMAGE:-"streamnative/sn-platform"}
PULSAR_IMAGE_TAG=${PULSAR_IMAGE_TAG:-"2.7.1"}
PYTHON_VERSION=${PYTHON_VERSION:-"3.12"}
DOCKER_REPO=${DOCKER_REPO:-"streamnative"}
RUNNER_BASE="pulsar-functions-runner-base"
PULSARCTL_RUNNER_BASE="pulsar-functions-pulsarctl-runner-base"
JAVA_RUNNER="pulsar-functions-java-runner"
PULSARCTL_JAVA_RUNNER="pulsar-functions-pulsarctl-java-runner"
GO_RUNNER="pulsar-functions-go-runner"
PULSARCTL_GO_RUNNER="pulsar-functions-pulsarctl-go-runner"
PYTHON_RUNNER="pulsar-functions-python-runner"
PULSARCTL_PYTHON_RUNNER="pulsar-functions-pulsarctl-python-runner"
RUNNER_TAG=${RUNNER_TAG:-$PULSAR_IMAGE_TAG}
KIND_PUSH=${KIND_PUSH:-false}
CI_TEST=${CI_TEST:-false}
PLATFORMS=${PLATFORMS:-"linux/amd64"}
PRIMARY_PLATFORM=${PLATFORMS%%,*}
RUNNER_BASE_IMAGE="${DOCKER_REPO}/${RUNNER_BASE}:${RUNNER_TAG}"
PULSARCTL_RUNNER_BASE_IMAGE="${DOCKER_REPO}/${PULSARCTL_RUNNER_BASE}:${RUNNER_TAG}"
JAVA_RUNNER_IMAGE="${DOCKER_REPO}/${JAVA_RUNNER}:${RUNNER_TAG}"
PULSARCTL_JAVA_RUNNER_IMAGE="${DOCKER_REPO}/${PULSARCTL_JAVA_RUNNER}:${RUNNER_TAG}"
GO_RUNNER_IMAGE="${DOCKER_REPO}/${GO_RUNNER}:${RUNNER_TAG}"
PULSARCTL_GO_RUNNER_IMAGE="${DOCKER_REPO}/${PULSARCTL_GO_RUNNER}:${RUNNER_TAG}"
PYTHON_RUNNER_IMAGE="${DOCKER_REPO}/${PYTHON_RUNNER}:${RUNNER_TAG}"
PULSARCTL_PYTHON_RUNNER_IMAGE="${DOCKER_REPO}/${PULSARCTL_PYTHON_RUNNER}:${RUNNER_TAG}"

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
  echo "multi-platform builds require PUSH=true so dependent images can resolve their base images" >&2
  exit 1
fi

if [[ "${MULTI_PLATFORM}" == "true" && "${KIND_PUSH}" == "true" ]]; then
  echo "KIND_PUSH=true is only supported for single-platform builds" >&2
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

echo "build runner base"
build_image "${RUNNER_BASE_IMAGE}" images/pulsar-functions-base-runner \
  --build-arg PULSAR_IMAGE="${PULSAR_IMAGE}" \
  --build-arg PULSAR_IMAGE_TAG="${PULSAR_IMAGE_TAG}" \
  --progress=plain
build_image "${PULSARCTL_RUNNER_BASE_IMAGE}" images/pulsar-functions-base-runner \
  -f images/pulsar-functions-base-runner/pulsarctl.Dockerfile \
  --build-arg PULSAR_IMAGE="${PULSAR_IMAGE}" \
  --build-arg PULSAR_IMAGE_TAG="${PULSAR_IMAGE_TAG}" \
  --progress=plain

echo "build java runner"
build_image "${JAVA_RUNNER_IMAGE}" images/pulsar-functions-java-runner \
  --build-arg BASE_IMAGE="${RUNNER_BASE_IMAGE}" \
  --build-arg PULSAR_IMAGE="${PULSAR_IMAGE}" \
  --build-arg PULSAR_IMAGE_TAG="${PULSAR_IMAGE_TAG}" \
  --progress=plain
build_image "${PULSARCTL_JAVA_RUNNER_IMAGE}" images/pulsar-functions-java-runner \
  -f images/pulsar-functions-java-runner/pulsarctl.Dockerfile \
  --build-arg BASE_IMAGE="${PULSARCTL_RUNNER_BASE_IMAGE}" \
  --build-arg PULSAR_IMAGE="${PULSAR_IMAGE}" \
  --build-arg PULSAR_IMAGE_TAG="${PULSAR_IMAGE_TAG}" \
  --progress=plain

echo "build python runner"
build_image "${PYTHON_RUNNER_IMAGE}" images/pulsar-functions-python-runner \
  --build-arg BASE_IMAGE="${RUNNER_BASE_IMAGE}" \
  --build-arg PULSAR_IMAGE="${PULSAR_IMAGE}" \
  --build-arg PULSAR_IMAGE_TAG="${PULSAR_IMAGE_TAG}" \
  --progress=plain
build_image "${PULSARCTL_PYTHON_RUNNER_IMAGE}" images/pulsar-functions-python-runner \
  -f images/pulsar-functions-python-runner/pulsarctl.Dockerfile \
  --build-arg PULSAR_IMAGE="${PULSAR_IMAGE}" \
  --build-arg PULSAR_IMAGE_TAG="${PULSAR_IMAGE_TAG}" \
  --build-arg PYTHON_VERSION="${PYTHON_VERSION}" \
  --progress=plain

echo "build go runner"
build_image "${GO_RUNNER_IMAGE}" images/pulsar-functions-go-runner \
  --build-arg BASE_IMAGE="${RUNNER_BASE_IMAGE}" \
  --progress=plain
build_image "${PULSARCTL_GO_RUNNER_IMAGE}" images/pulsar-functions-go-runner \
  -f images/pulsar-functions-go-runner/pulsarctl.Dockerfile \
  --build-arg BASE_IMAGE="${PULSARCTL_RUNNER_BASE_IMAGE}" \
  --progress=plain

if [ "$KIND_PUSH" = true ] ; then
  echo "push images to kind"
  clusters=$(kind get clusters)
  echo $clusters
  for cluster in $clusters
  do
    kind load docker-image "${JAVA_RUNNER_IMAGE}" --name $cluster
    kind load docker-image "${PULSARCTL_JAVA_RUNNER_IMAGE}" --name $cluster
    kind load docker-image "${PYTHON_RUNNER_IMAGE}" --name $cluster
    kind load docker-image "${PULSARCTL_PYTHON_RUNNER_IMAGE}" --name $cluster
    kind load docker-image "${GO_RUNNER_IMAGE}" --name $cluster
    kind load docker-image "${PULSARCTL_GO_RUNNER_IMAGE}" --name $cluster
  done
fi

if [[ "${PUSH}" == "true" && "${MULTI_PLATFORM}" != "true" ]]; then
  docker push "${RUNNER_BASE_IMAGE}"
  docker push "${PULSARCTL_RUNNER_BASE_IMAGE}"
  docker push "${JAVA_RUNNER_IMAGE}"
  docker push "${PULSARCTL_JAVA_RUNNER_IMAGE}"
  docker push "${PYTHON_RUNNER_IMAGE}"
  docker push "${PULSARCTL_PYTHON_RUNNER_IMAGE}"
  docker push "${GO_RUNNER_IMAGE}"
  docker push "${PULSARCTL_GO_RUNNER_IMAGE}"
fi
#
#if [ "$CI_TEST" = true ] ; then
#  echo "apply images to function mesh ci yaml"
#  clusters=$(kind get clusters)
#  echo $clusters
#  for cluster in $clusters
#  do
#    kind load docker-image "${DOCKER_REPO}"/${JAVA_RUNNER}:"${RUNNER_TAG}" --name $cluster
#    kind load docker-image "${DOCKER_REPO}"/${PYTHON_RUNNER}:"${RUNNER_TAG}" --name $cluster
#    kind load docker-image "${DOCKER_REPO}"/${GO_RUNNER}:"${RUNNER_TAG}" --name $cluster
#  done
#fi
