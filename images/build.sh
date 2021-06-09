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

PULSAR_IMAGE=${PULSAR_IMAGE:-"streamnative/pulsar-all"}
PULSAR_IMAGE_TAG=${PULSAR_IMAGE_TAG:-"2.7.1"}
DOCKER_REPO=${DOCKER_REPO:-"streamnative"}
RUNNER_BASE="pulsar-functions-runner-base"
JAVA_RUNNER="pulsar-functions-java-runner"
GO_RUNNER="pulsar-functions-go-runner"
PYTHON_RUNNER="pulsar-functions-python-runner"
RUNNER_TAG=${RUNNER_TAG:-$PULSAR_IMAGE_TAG}
KIND_PUSH=${KIND_PUSH:-false}
CI_TEST=${CI_TEST:-false}

echo "build runner base"
docker build -t ${RUNNER_BASE} images/pulsar-functions-base-runner --build-arg PULSAR_IMAGE="$PULSAR_IMAGE" --build-arg PULSAR_IMAGE_TAG="$PULSAR_IMAGE_TAG"
docker tag ${RUNNER_BASE} "${DOCKER_REPO}"/${RUNNER_BASE}:"${RUNNER_TAG}"

echo "build java runner"
docker build -t ${JAVA_RUNNER} images/pulsar-functions-java-runner --build-arg PULSAR_IMAGE="$PULSAR_IMAGE" --build-arg PULSAR_IMAGE_TAG="$PULSAR_IMAGE_TAG"
docker tag ${JAVA_RUNNER} "${DOCKER_REPO}"/${JAVA_RUNNER}:"${RUNNER_TAG}"

echo "build python runner"
docker build -t ${PYTHON_RUNNER} images/pulsar-functions-python-runner --build-arg PULSAR_IMAGE="$PULSAR_IMAGE" --build-arg PULSAR_IMAGE_TAG="$PULSAR_IMAGE_TAG"
docker tag ${PYTHON_RUNNER} "${DOCKER_REPO}"/${PYTHON_RUNNER}:"${RUNNER_TAG}"

echo "build go runner"
docker build -t ${GO_RUNNER} images/pulsar-functions-go-runner # go runner is almost the same as runner base, so we no need to given build args for go runner
docker tag ${GO_RUNNER} "${DOCKER_REPO}"/${GO_RUNNER}:"${RUNNER_TAG}"

if [ "$KIND_PUSH" = true ] ; then
  echo "push images to kind"
  clusters=$(kind get clusters)
  echo $clusters
  for cluster in $clusters
  do
    kind load docker-image "${DOCKER_REPO}"/${JAVA_RUNNER}:"${RUNNER_TAG}" --name $cluster
    kind load docker-image "${DOCKER_REPO}"/${PYTHON_RUNNER}:"${RUNNER_TAG}" --name $cluster
    kind load docker-image "${DOCKER_REPO}"/${GO_RUNNER}:"${RUNNER_TAG}" --name $cluster
  done
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