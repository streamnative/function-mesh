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

set -ex
BINDIR=`dirname "$0"`
PULSAR_HOME=`cd ${BINDIR}/..;pwd`

NAMESPACE=default
CLUSTER=sn-platform

case ${1} in
  java)
    kubectl exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin packages upload function://public/default/test-java-function --path /pulsar/examples/api-examples.jar --description "test java function"
    ;;
  py)
    kubectl exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin packages upload function://public/default/test-py-function --path /pulsar/examples/python-examples/exclamation_function.py --description "test python function"
    ;;
  pyzip)
    kubectl cp "${PULSAR_HOME}/.ci/examples/py-examples" "${NAMESPACE}/${CLUSTER}-pulsar-broker-0:/pulsar/"
    kubectl exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin packages upload function://public/default/test-py-zip-function --path /pulsar/py-examples/exclamation.zip --description "test python zip function"
    ;;
  pypip)
    kubectl cp "${PULSAR_HOME}/.ci/examples/py-examples" "${NAMESPACE}/${CLUSTER}-pulsar-broker-0:/pulsar/"
    kubectl exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin packages upload function://public/default/test-py-pip-function --path /pulsar/py-examples/pulsarfunction-1.0.0-py3-none-any.whl --description "test python pip function"
    ;;
  go)
    kubectl cp "${PULSAR_HOME}/.ci/examples/go-examples" "${NAMESPACE}/${CLUSTER}-pulsar-broker-0:/pulsar/"
    kubectl exec -n ${NAMESPACE} ${CLUSTER}-pulsar-broker-0 -- bin/pulsar-admin packages upload function://public/default/test-go-function --path /pulsar/go-examples/exclamationFunc --description "test golang function"
    ;;
esac
