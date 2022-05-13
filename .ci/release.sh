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

BINDIR=`dirname "$0"`
CHARTS_HOME=`cd ${BINDIR}/..;pwd`
CHARTS_PKGS=${CHARTS_HOME}/.chart-packages
CHARTS_INDEX=${CHARTS_HOME}/.chart-index
CHARTS_REPO=${CHARTS_REPO:-"https://charts.functionmesh.io"}
OWNER=${OWNER:-streamnative}
REPO=${REPO:-function-mesh}
GITHUB_TOKEN=${GITHUB_TOKEN:-"UNSET"}
PUBLISH_CHARTS=${PUBLISH_CHARTS:-"false"}
GITUSER=${GITUSER:-"UNSET"}
GITEMAIL=${GITEMAIL:-"UNSET"}
CHART="$1"
RELEASE_BRANCH="$2"

# hack/common.sh need this variable to be set
PULSAR_CHART_HOME=${CHARTS_HOME}

source ${CHARTS_HOME}/hack/common.sh
source ${CHARTS_HOME}/.ci/git.sh

# allow overwriting cr binary
CR="docker run -v ${CHARTS_HOME}:/cr quay.io/helmpack/chart-releaser:v${CR_VERSION} cr"

function release::ensure_dir() {
    local dir=$1
    if [[ -d ${dir} ]]; then
        rm -rf ${dir}
    fi
    mkdir -p ${dir}
}


function release::package_chart() {
    echo "Packaging chart '${CHART}'..."
    helm dependency update ${CHARTS_HOME}/charts/${CHART}
    helm package ${CHARTS_HOME}/charts/${CHART} --destination ${CHARTS_PKGS}
}

function release::upload_packages() {
    echo "Uploading charts..."
    ${CR} upload --owner ${OWNER} --git-repo ${REPO} -t ${GITHUB_TOKEN} --commit ${RELEASE_BRANCH} --package-path /cr/.chart-packages --release-name-template "{{ .Name }}-{{ .Version }}"
}

function release::update_chart_index() {
    echo "Updating chart index..."
    ${CR} index -o ${OWNER} -r ${REPO} -t "${GITHUB_TOKEN}" -c ${CHARTS_REPO} --index-path /cr/.chart-index --package-path /cr/.chart-packages --release-name-template "{{ .Name }}-{{ .Version }}"
}

function release::publish_charts() {
    git config user.email "${GITEMAIL}"
    git config user.name "${GITUSER}"
    git pull
    git checkout gh-pages
    cp --force ${CHARTS_INDEX}/index.yaml index.yaml
    git add index.yaml
    git commit --message="Publish new charts to ${CHARTS_REPO}" --signoff
    git remote -v
    git remote add sn https://${SNBOT_USER}:${GITHUB_TOKEN}@github.com/${OWNER}/${REPO} 
    git push sn gh-pages 
}

# install cr
# hack::ensure_cr
docker pull quay.io/helmpack/chart-releaser:v${CR_VERSION}

release::ensure_dir ${CHARTS_PKGS}
release::ensure_dir ${CHARTS_INDEX}

release::package_chart

release::upload_packages
release::update_chart_index

release::publish_charts
