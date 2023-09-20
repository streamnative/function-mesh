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

if [ $# -eq 0 ]; then
    echo "Required argument with new version"
    exit 1
fi

NEW_APP_VERSION=$1
NEW_CAHRT_VERSION=$2

### version self-increment
## $1: number of part: 0 – major, 1 – minor, 2 – patch
generate_increment_version() {
    pattern='version:.*([0-9]+\.[0-9]+\.[0-9]+)'
    content=$(grep -i "^version:" charts/function-mesh-operator/Chart.yaml)
    [[ $content =~ $pattern ]]
    current_version=${BASH_REMATCH[1]}

    if [ ! "$current_version" ]; then
        echo "failed to get current chart version"
        exit 1
    fi

    local delimiter=.
    local array=($(echo "$current_version" | tr $delimiter '\n'))
    array[$1]=$((array[$1]+1))
    NEW_CAHRT_VERSION=$(local IFS=$delimiter ; echo "${array[*]}")
}

if [ ! "$NEW_CAHRT_VERSION" ]; then
    # increment the patch value in version semantics by default
    generate_increment_version 2
fi

echo "Bumping function mesh to $NEW_APP_VERSION"
echo "Bumping function mesh helm charts to $NEW_CAHRT_VERSION"

# change Makefile version
sed -i.bak -E "s/(VERSION \?= )(.+)/\1$NEW_APP_VERSION/" Makefile

# change chart appVersion
sed -i.bak -E "s/(appVersion\: )(.+)/\1$NEW_APP_VERSION/" charts/function-mesh-operator/Chart.yaml
sed -i.bak -E "s/(appVersion\: )(.+)/\1$NEW_APP_VERSION/" charts/function-mesh-operator/charts/admission-webhook/Chart.yaml

# change chart value version
sed -i.bak -E "s/(operatorImage\: streamnative\/function\-mesh\:v)(.+)/\1$NEW_APP_VERSION/" charts/function-mesh-operator/values.yaml

# change install.sh
sed -i.bak -E "s/(local fm_version\=)(.+)/\1\"v$NEW_APP_VERSION\"/" install.sh

# change README.md
sed -i.bak -E "s/(.+)v(.+)(\/install.sh)/\1v$NEW_APP_VERSION\3/" README.md

# change chart version
sed -i.bak -E "s/(version\: )(.+)/\1$NEW_CAHRT_VERSION/" charts/function-mesh-operator/Chart.yaml
sed -i.bak -E "s/(version\: )(.+)/\1$NEW_CAHRT_VERSION/" charts/function-mesh-operator/charts/admission-webhook/Chart.yaml

TEST_IMAGE=kind-registry:5000/streamnativeio/function-mesh-catalog:v${NEW_APP_VERSION} yq eval -i '.spec.image = strenv(TEST_IMAGE)' .ci/olm-tests/catalog.yml
TEST_CSV=function-mesh.v${NEW_APP_VERSION} yq eval -i '.spec.startingCSV = strenv(TEST_CSV)' .ci/olm-tests/subs.yml

pushd charts
helm-docs
popd
