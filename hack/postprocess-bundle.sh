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

# These labels are required to be in the bundle.Dockerfile, but can't be added by the operator-sdk automatically
cat <<EOF >> bundle.Dockerfile
# Certified Openshift required labels
LABEL com.redhat.openshift.versions="v4.6-v4.8"
LABEL com.redhat.delivery.operator.bundle=true
LABEL com.redhat.delivery.backport=true
LABEL operators.operatorframework.io.bundle.channel.default.v1="alpha"
EOF

# Add them to the bundle metadata also
yq eval -i '.annotations."com.redhat.openshift.versions" = "v4.6-v4.8"' bundle/metadata/annotations.yaml
yq eval -i '.annotations."com.redhat.delivery.operator.bundle" = true' bundle/metadata/annotations.yaml
yq eval -i '.annotations."com.redhat.delivery.backport" = true' bundle/metadata/annotations.yaml
yq eval -i '.annotations."operators.operatorframework.io.bundle.channel.default.v1" = "alpha"' bundle/metadata/annotations.yaml
yq eval -i '.annotations."com.redhat.openshift.versions" headComment = "Certified Openshift required labels"' bundle/metadata/annotations.yaml

