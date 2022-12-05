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

PULSAR_IMAGE_TAG=${PULSAR_IMAGE_TAG:-"2.7.1"}
PULSAR_IMAGE=${PULSAR_IMAGE:-"streamnative/pulsar-all"}
JAVA_RUNNER_IMAGE=${RUNNER_IMAGE:-"streamnative/pulsar-functions-java-runner"}
DOCKER_REPO=${DOCKER_REPO:-"streamnative"}
JAVA_SAMPLE="pulsar-functions-java-sample"
PYTHON_SAMPLE="pulsar-functions-python-sample"
GO_SAMPLE="pulsar-functions-go-sample"
CONNECTOR_ES_SAMPLE="pulsar-io-elasticsearch"
KIND_PUSH=${KIND_PUSH:-false}
CI_TEST=${CI_TEST:-false}

echo "build java sample"
docker build --platform linux/amd64 -t ${JAVA_SAMPLE} images/samples/java-function-samples --build-arg PULSAR_IMAGE_TAG="$PULSAR_IMAGE_TAG"
docker tag ${JAVA_SAMPLE} "${DOCKER_REPO}"/${JAVA_SAMPLE}:"${PULSAR_IMAGE_TAG}"

echo "build python sample"
docker build --platform linux/amd64 -t ${PYTHON_SAMPLE} images/samples/python-function-samples --build-arg PULSAR_IMAGE_TAG="$PULSAR_IMAGE_TAG"
docker tag ${PYTHON_SAMPLE} "${DOCKER_REPO}"/${PYTHON_SAMPLE}:"${PULSAR_IMAGE_TAG}"

echo "build go sample"
docker build --platform linux/amd64 -t ${GO_SAMPLE} images/samples/go-function-samples --build-arg PULSAR_IMAGE_TAG="$PULSAR_IMAGE_TAG"
docker tag ${GO_SAMPLE} "${DOCKER_REPO}"/${GO_SAMPLE}:"${PULSAR_IMAGE_TAG}"

echo "build connector sample"
docker build --platform linux/amd64 -t ${CONNECTOR_ES_SAMPLE} images/samples/pulsar-io-connector/pulsar-io-elasticsearch --build-arg PULSAR_IMAGE_TAG="$PULSAR_IMAGE_TAG" --build-arg PULSAR_IMAGE="$PULSAR_IMAGE" --build-arg RUNNER_IMAGE="$JAVA_RUNNER_IMAGE"
docker tag ${CONNECTOR_ES_SAMPLE} "${DOCKER_REPO}"/${CONNECTOR_ES_SAMPLE}:"${PULSAR_IMAGE_TAG}"

if [ "$KIND_PUSH" = true ] ; then
  echo "push images to kind"
  clusters=$(kind get clusters)
  echo $clusters
  for cluster in $clusters
  do
    kind load docker-image "${DOCKER_REPO}"/${JAVA_SAMPLE}:"${PULSAR_IMAGE_TAG}" --name $cluster
    kind load docker-image "${DOCKER_REPO}"/${PYTHON_SAMPLE}:"${PULSAR_IMAGE_TAG}" --name $cluster
    kind load docker-image "${DOCKER_REPO}"/${GO_SAMPLE}:"${PULSAR_IMAGE_TAG}" --name $cluster
  done
fi
