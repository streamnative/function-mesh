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
LABEL com.redhat.openshift.versions="v4.8-v4.11"
LABEL com.redhat.delivery.operator.bundle=true
LABEL com.redhat.delivery.backport=true
LABEL operators.operatorframework.io.bundle.channel.default.v1="alpha"
EOF

# Add them to the bundle metadata also
yq eval -i '.annotations."com.redhat.openshift.versions" = "v4.8-v4.11"' bundle/metadata/annotations.yaml
yq eval -i '.annotations."com.redhat.delivery.operator.bundle" = true' bundle/metadata/annotations.yaml
yq eval -i '.annotations."com.redhat.delivery.backport" = true' bundle/metadata/annotations.yaml
yq eval -i '.annotations."operators.operatorframework.io.bundle.channel.default.v1" = "alpha"' bundle/metadata/annotations.yaml
yq eval -i '.annotations."com.redhat.openshift.versions" headComment = "Certified Openshift required labels"' bundle/metadata/annotations.yaml

# Add relatedImages
yq -i '.spec.relatedImages = []' bundle/manifests/function-mesh.clusterserviceversion.yaml
yq -i '.spec.relatedImages += {"name": "function-mesh", "image": ""}' bundle/manifests/function-mesh.clusterserviceversion.yaml
yq -i '.spec.relatedImages += {"name": "kube-rbac-proxy", "image": "docker.cloudsmith.io/streamnative/mirrors/gcr.io/kubebuilder/kube-rbac-proxy@sha256:67ecb332573384515406ebd71816781366b70adb0eb66345e5980e92603373e1"}' bundle/manifests/function-mesh.clusterserviceversion.yaml
yq -i '.spec.relatedImages[0].image += env(IMG_DIGEST)' bundle/manifests/function-mesh.clusterserviceversion.yaml
