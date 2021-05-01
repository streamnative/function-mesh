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
DOCKER_REPO=${DOCKER_REPO:-"streamnative"}
JAVA_SAMPLE="pulsar-functions-java-sample"
PYTHON_SAMPLE="pulsar-functions-python-sample"
GO_SAMPLE="pulsar-functions-go-sample"
KIND_PUSH=${KIND_PUSH:-false}
CI_TEST=${CI_TEST:-false}


echo "build java sample"
docker build -t ${JAVA_SAMPLE} images/samples/java-function-samples --build-arg PULSAR_IMAGE_TAG="$PULSAR_IMAGE_TAG"
docker tag ${JAVA_SAMPLE} "${DOCKER_REPO}"/${JAVA_SAMPLE}:"${PULSAR_IMAGE_TAG}"

echo "build python sample"
docker build -t ${PYTHON_SAMPLE} images/samples/python-function-samples --build-arg PULSAR_IMAGE_TAG="$PULSAR_IMAGE_TAG"
docker tag ${PYTHON_SAMPLE} "${DOCKER_REPO}"/${PYTHON_SAMPLE}:"${PULSAR_IMAGE_TAG}"

echo "build go sample"
docker build -t ${GO_SAMPLE} images/samples/go-function-samples --build-arg PULSAR_IMAGE_TAG="$PULSAR_IMAGE_TAG"
docker tag ${GO_SAMPLE} "${DOCKER_REPO}"/${GO_SAMPLE}:"${PULSAR_IMAGE_TAG}"

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
