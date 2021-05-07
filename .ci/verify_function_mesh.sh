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
PULSAR_HOME=`cd ${BINDIR}/..;pwd`
TLS=${TLS:-"false"}
SYMMETRIC=${SYMMETRIC:-"false"}
FUNCTION=${FUNCTION:-"false"}

source ${PULSAR_HOME}/.ci/helm.sh

case ${1} in
  compute_v1alpha1_go_function)
    ci::verify_function_mesh go-function-sample
    sleep 60
    ci::print_function_log go-function-sample
    ci::verify_go_function go-function-sample
    ;;
  compute_v1alpha1_function)
    ci::verify_function_mesh function-sample
    sleep 60
    ci::print_function_log function-sample
    ci::verify_java_function function-sample
    ;;
  compute_v1alpha1_py_function)
    ci::verify_function_mesh py-function-sample
    sleep 60
    ci::print_function_log py-function-sample
    ci::verify_python_function py-function-sample
    ;;
  compute_v1alpha1_functionmesh)
    ci::verify_function_mesh functionmesh-sample-java-function
    ci::verify_function_mesh functionmesh-sample-golang-function
    ci::verify_function_mesh functionmesh-sample-python-function
    sleep 60
    ci::print_function_log functionmesh-sample-java-function
    ci::print_function_log functionmesh-sample-golang-function
    ci::print_function_log functionmesh-sample-python-function
    ci::verify_mesh_function
    ;;
esac