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

usage() {
    cat <<EOF
This script use kind to create Kubernetes cluster, about kind please refer: https://kind.sigs.k8s.io/
Before run this script, please ensure that:
* have installed docker
* have installed kind and kind's version == ${KIND_VERSION}
Options:
       -h,--help              prints the usage message
       -n,--name              name of the Kubernetes cluster, default value: kind
       -c,--node              the count of the cluster nodes, default value: 3
       -k,--kubernetes        version of the Kubernetes cluster, default value: v1.21.12
       -v,--volume            the volumes number of each kubernetes node, default value: 10
       -o,--output            the output path of the cluster configuration file
Usage:
    $0 --name testCluster --node 4 --kubernetes v1.21
EOF
}

while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    -n|--name)
    cluster_name="$2"
    shift
    shift
    ;;
    -c|--node)
    node_num="$2"
    shift
    shift
    ;;
    -k|--kubernetes)
    k8s_version="$2"
    shift
    shift
    ;;
    -v|--volume)
    volume_num="$2"
    shift
    shift
    ;;
    -o|--output)
    output_file="$2"
    shift
    shift
    ;;
    -h|--help)
    usage
    exit 0
    ;;
    *)
    echo "unknown option: $key"
    usage
    exit 1
    ;;
esac
done

cluster_name=${cluster_name:-kind}
node_num=${node_num:-3}
k8s_version=${k8s_version:-v1.21.12}
volume_num=${volume_num:-10}

echo "== generate kind cluster configuration with specs as follows:"
echo "== cluster name: ${cluster_name}"
echo "== kubernetes version: ${k8s_version}"
echo "== worker node number: ${node_num}"
echo "== extra volume number: ${volume_num}"
echo "== output file path: ${output_file}"

echo "############# start generate cluster configuration:[${cluster_name}] #############"
work_dir=${HOME}/kind/"${cluster_name}"
mkdir -p "${work_dir}"

data_dir="${work_dir}"/data

echo "clean data dir: ${data_dir}"
if [ -d "${data_dir}" ]; then
    rm -rf "${data_dir}"
fi

image="kindest/node:${k8s_version}"
echo "select node image: ${image}"

cat <<EOF > "${output_file}"
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

kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
kubeadmConfigPatches:
  - |
    apiVersion: kubelet.config.k8s.io/v1beta1
    kind: KubeletConfiguration
    evictionHard:
      nodefs.available: "0%"
kubeadmConfigPatchesJSON6902:
  - group: kubeadm.k8s.io
    version: v1beta2
    kind: ClusterConfiguration
    patch: |
      - op: add
        path: /apiServer/certSANs/-
        value: my-hostname
nodes:
  - role: control-plane
    image: ${image}
    extraPortMappings:
      - containerPort: 31234
        hostPort: 80
        listenAddress: "127.0.0.1"
        protocol: TCP
EOF

for ((i=0;i<"${node_num}";i++))
do
    mkdir -p "${data_dir}"/worker"${i}"
    cat <<EOF >>  "${output_file}"
  - role: worker
    extraMounts:
EOF
    for ((k=1;k<="${volume_num}";k++))
    do
        mkdir -p "${data_dir}"/worker"${i}"/vol"${k}"
        cat <<EOF >> "${output_file}"
    - containerPath: /mnt/disks/vol${k}
      hostPath: ${data_dir}/worker${i}/vol${k}
EOF
    done
done

echo "configuration path: ${output_file}"
echo "############# Complete cluster configuration generation:[${cluster_name}] #############"