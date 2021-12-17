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

NEW_VERSION=$1
echo "Bumping function mesh to $NEW_VERSION"

# change Makefile version
sed -i.bak -E "s/(VERSION \?= )(.+)/\1$NEW_VERSION/" Makefile

# change chart version
sed -i.bak -E "s/(version\: )(.+)/\1$NEW_VERSION/" charts/function-mesh-operator/Chart.yaml

# change chart value version
sed -i.bak -E "s/(operatorImage\: streamnative\/function\-mesh\:v)(.+)/\1$NEW_VERSION/" charts/function-mesh-operator/values.yaml

# change install.sh
sed -i.bak -E "s/(local fm_version\=)(.+)/\1\"v$NEW_VERSION\"/" install.sh

# change README.md
sed -i.bak -E "s/(.+)v(.+)(\/install.sh)/\1v$NEW_VERSION\3/" README.md
