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

set -x
CRD_FUNCTIONS_FILE=cloud.streamnative.io_functions.yaml # Target functions CRD file
CRD_SOURCES_FILE=cloud.streamnative.io_sources.yaml # Target sources CRD file
CRD_SINKS_FILE=cloud.streamnative.io_sinks.yaml # Target sinks CRD file

DEST_DIR=$PWD
GEN_DIR=/tmp/functions-mesh/crd
mkdir -p $GEN_DIR
cp ../config/crd/bases/* $GEN_DIR
cd $GEN_DIR

LOCAL_MANIFEST_FUNCTIONS_FILE=$GEN_DIR/$CRD_FUNCTIONS_FILE
LOCAL_MANIFEST_SOURCES_FILE=$GEN_DIR/$CRD_SOURCES_FILE
LOCAL_MANIFEST_SINKS_FILE=$GEN_DIR/$CRD_SINKS_FILE

# yq site: https://mikefarah.gitbook.io/yq/
yq eval ".spec.preserveUnknownFields = false" -i $CRD_FUNCTIONS_FILE
yq eval ".spec.preserveUnknownFields = false" -i $CRD_SOURCES_FILE
yq eval ".spec.preserveUnknownFields = false" -i $CRD_SINKS_FILE

docker pull docker.pkg.github.com/kubernetes-client/java/crd-model-gen:v1.0.3
docker pull kindest/node:v1.15.12
docker build --tag crd-model-gen:latest "${DEST_DIR}/tool/crd-model-gen"
#docker rm -f kind-control-plane
# Generate functions crd
docker run \
  --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v "$(pwd)":"$(pwd)" \
  --network host \
  crd-model-gen:latest \
  /generate.sh \
  -u $LOCAL_MANIFEST_FUNCTIONS_FILE \
  -n io.streamnative.cloud \
  -p io.streamnative.cloud.function \
  -o "$(pwd)"

# Generate sources crd
docker run \
  --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v "$(pwd)":"$(pwd)" \
  --network host \
  crd-model-gen:latest \
  /generate.sh \
  -u $LOCAL_MANIFEST_SOURCES_FILE \
  -n io.streamnative.cloud \
  -p io.streamnative.cloud.source \
  -o "$(pwd)"

# Generate sinks crd
docker run \
  --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v "$(pwd)":"$(pwd)" \
  --network host \
  crd-model-gen:latest \
  /generate.sh \
  -u $LOCAL_MANIFEST_SINKS_FILE \
  -n io.streamnative.cloud \
  -p io.streamnative.cloud.sink \
  -o "$(pwd)"
#open $GEN_DIR

mkdir -p $DEST_DIR/src/main/java/io/streamnative/cloud/
cp -r $GEN_DIR/src/main/java/io/streamnative/cloud/* $DEST_DIR/src/main/java/io/streamnative/cloud/
