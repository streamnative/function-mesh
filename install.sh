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

# This is a script to quickly install function-mesh.
# This script will check if docker and kubernetes are installed. If local mode is set and kubernetes is not installed,
# it will use KinD to install the kubernetes cluster according to the configuration.
# Finally, when all dependencies are installed, function-mesh will be installed.

usage() {
  cat <<EOF
This script is used to install function-mesh.
Before running this script, please ensure that:
* have installed docker if you run function-mesh in local.
* have installed Kubernetes if you run function-mesh in normal Kubernetes cluster
USAGE:
    install.sh [FLAGS] [OPTIONS]
FLAGS:
    -h, --help               Prints help information
    -d, --dependency-only    Install dependencies only, including kind, kubectl, local-kube.
        --force              Force reinstall all components if they are already installed, include: kind, local-kube, function-mesh
        --force-function-mesh   Force reinstall function-mesh if it is already installed
        --force-local-kube   Force reinstall local Kubernetes cluster if it is already installed
        --force-kubectl      Force reinstall kubectl client if it is already installed
        --force-kind         Force reinstall KinD if it is already installed
OPTIONS:
    -v, --version            Version of function-mesh, default value: latest
    -l, --local [kind]       Choose a way to run a local kubernetes cluster, supported value: kind,
                             If this value is not set and the Kubernetes is not installed, this script will exit with 1.
    -n, --name               Name of Kubernetes cluster, default value: kind
    -c  --crd                The path of the crd files.
        --kind-version       Version of the Kind tool, default value: v0.7.0
        --node-num           The count of the cluster nodes,default value: 3
        --k8s-version        Version of the Kubernetes cluster,default value: v1.17.2
        --volume-num         The volumes number of each kubernetes node,default value: 5
        --release-name       Release name of function-mesh, default value: function-mesh
        --namespace          Namespace of function-mesh, default value: default
EOF
}

main() {
  local local_kube="kind"
  local cm_version="latest"
  local kind_name="kind"
  local kind_version="v0.7.0"
  local node_num=2
  local k8s_version="v1.17.2"
  local volume_num=2
  local release_name="function-mesh"
  local namespace="default"
  local force_function_mesh=false
  local force_local_kube=false
  local force_kubectl=false
  local force_kind=false
  local crd=""
  local runtime="docker"
  local install_dependency_only=false
  local docker_registry="docker.io"

  while [[ $# -gt 0 ]]; do
    key="$1"
    case "$key" in
    -h | --help)
      usage
      exit 0
      ;;
    -l | --local)
      local_kube="$2"
      shift
      shift
      ;;
    -v | --version)
      cm_version="$2"
      shift
      shift
      ;;
    -n | --name)
      kind_name="$2"
      shift
      shift
      ;;
    -c | --crd)
      crd="$2"
      shift
      shift
      ;;
    -d | --dependency-only)
      install_dependency_only=true
      shift
      ;;
    --force)
      force_function_mesh=true
      force_local_kube=true
      force_kubectl=true
      force_kind=true
      shift
      ;;
    --force-local-kube)
      force_local_kube=true
      shift
      ;;
    --force-kubectl)
      force_kubectl=true
      shift
      ;;
    --force-kind)
      force_kind=true
      shift
      ;;
    --force-function-mesh)
      force_function_mesh=true
      shift
      ;;
    --kind-version)
      kind_version="$2"
      shift
      shift
      ;;
    --node-num)
      node_num="$2"
      shift
      shift
      ;;
    --k8s-version)
      k8s_version="$2"
      shift
      shift
      ;;
    --volume-num)
      volume_num="$2"
      shift
      shift
      ;;
    --release-name)
      release_name="$2"
      shift
      shift
      ;;
    --namespace)
      namespace="$2"
      shift
      shift
      ;;
    --docker-registry)
      docker_registry="$2"
      shift
      shift
      ;;
    *)
      echo "unknown flag or option $key"
      usage
      exit 1
      ;;
    esac
  done

  if [ "${local_kube}" != "kind" ]; then
    printf "local Kubernetes by %s is not supported\n" "${local_kube}"
    exit 1
  fi

  #    if [ "${local_kube}" == "kind" ]; then
  #        runtime="containerd"
  #    fi

  if [ "${crd}" == "" ]; then
    crd="https://raw.githubusercontent.com/streamnative/function-mesh/${cm_version}/manifests/crd.yaml"
  fi

  need_cmd "sed"
  need_cmd "tr"

  if [ "${local_kube}" == "kind" ]; then
    prepare_env
    install_kubectl "${k8s_version}" ${force_kubectl}

    check_docker
    install_kind "${kind_version}" ${force_kind}
    install_kubernetes_by_kind "${kind_name}" "${k8s_version}" "${node_num}" "${volume_num}" ${force_local_kube}
  fi

  if [ "${install_dependency_only}" = true ]; then
    exit 0
  fi

  check_kubernetes
  install_function_mesh "${release_name}" "${namespace}" "${crd}" "${runtime}" "${cm_version}" "${docker_registry}"
  ensure_pods_ready "${namespace}" "app.kubernetes.io/component=controller-manager" 100
  printf "Function Mesh %s is installed successfully\n" "${release_name}"
}

prepare_env() {
  mkdir -p "$HOME/local/bin"
  local set_path="export PATH=$HOME/local/bin:\$PATH"
  local env_file="$HOME/.bash_profile"
  if [[ ! -e "${env_file}" ]]; then
    ensure touch "${env_file}"
  fi
  grep -qF -- "${set_path}" "${env_file}" || echo "${set_path}" >>"${env_file}"
  ensure source "${env_file}"
}

check_kubernetes() {
  need_cmd "kubectl"
  kubectl_err_msg=$(kubectl version 2>&1 1>/dev/null)
  if [ "$kubectl_err_msg" != "" ]; then
    printf "check Kubernetes failed, error: %s\n" "${kubectl_err_msg}"
    exit 1
  fi

  check_kubernetes_version
}

check_kubernetes_version() {
  version_info=$(kubectl version | sed 's/.*GitVersion:\"v\([0-9.]*\).*/\1/g')

  for v in $version_info; do
    if version_lt "$v" "1.12.0"; then
      printf "Function Mesh requires Kubernetes cluster running 1.12 or later\n"
      exit 1
    fi
  done
}

install_kubectl() {
  local kubectl_version=$1
  local force_install=$2

  printf "Install kubectl client\n"

  err_msg=$(kubectl version --client=true 2>&1 1>/dev/null)
  if [ "$err_msg" == "" ]; then
    v=$(kubectl version --client=true | sed 's/.*GitVersion:\"v\([0-9.]*\).*/\1/g')
    target_version=$(echo "${kubectl_version}" | sed s/v//g)
    if version_lt "$v" "${target_version}"; then
      printf "Function Mesh requires kubectl version %s or later\n" "${target_version}"
    else
      printf "kubectl Version %s has been installed\n" "$v"
      if [ "$force_install" != "true" ]; then
        return
      fi
    fi
  fi

  need_cmd "curl"
  local KUBECTL_BIN="${HOME}/local/bin/kubectl"
  local target_os=$(lowercase $(uname))

  ensure curl -Lo /tmp/kubectl https://storage.googleapis.com/kubernetes-release/release/${kubectl_version}/bin/${target_os}/amd64/kubectl
  ensure chmod +x /tmp/kubectl
  ensure mv /tmp/kubectl "${KUBECTL_BIN}"
}

install_kubernetes_by_kind() {
  local cluster_name=$1
  local cluster_version=$2
  local node_num=$3
  local volume_num=$4
  local force_install=$5

  printf "Install local Kubernetes %s\n" "${cluster_name}"

  need_cmd "kind"

  work_dir=${HOME}/kind/${cluster_name}
  kubeconfig_path=${work_dir}/config
  data_dir=${work_dir}/data
  clusters=$(kind get clusters)
  cluster_exist=false
  for c in $clusters; do
    if [ "$c" == "$cluster_name" ]; then
      printf "Kind cluster %s has been installed\n" "${cluster_name}"
      cluster_exist=true
      break
    fi
  done

  if [ "$cluster_exist" == "true" ]; then
    if [ "$force_install" == "true" ]; then
      printf "Delete Kind Kubernetes cluster %s\n" "${cluster_name}"
      kind delete cluster --name="${cluster_name}"
      status=$?
      if [ $status -ne 0 ]; then
        printf "Delete Kind Kubernetes cluster %s failed\n" "${cluster_name}"
        exit 1
      fi
    else
      ensure kind get kubeconfig --name="${cluster_name}" >"${kubeconfig_path}"
      return
    fi
  fi

  ensure mkdir -p "${work_dir}"

  printf "Clean data dir: %s\n" "${data_dir}"
  if [ -d "${data_dir}" ]; then
    ensure rm -rf "${data_dir}"
  fi

  config_file=${work_dir}/kind-config.yaml
  cat <<EOF >"${config_file}"
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
kubeadmConfigPatches:
- |
  apiVersion: kind.x-k8s.io/v1alpha4
  kind: ClusterConfiguration
  metadata:
    name: config
  apiServerExtraArgs:
    enable-admission-plugins: NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 5000
    hostPort: 5000
    listenAddress: 127.0.0.1
    protocol: TCP
EOF

  for ((i = 0; i < "${node_num}"; i++)); do
    ensure mkdir -p "${data_dir}/worker${i}"
    cat <<EOF >>"${config_file}"
- role: worker
  extraMounts:
EOF
    for ((k = 1; k <= "${volume_num}"; k++)); do
      ensure mkdir -p "${data_dir}/worker${i}/vol${k}"
      cat <<EOF >>"${config_file}"
  - containerPath: /mnt/disks/vol${k}
    hostPath: ${data_dir}/worker${i}/vol${k}
EOF
    done
  done

  local kind_image="kindest/node:${cluster_version}"

  printf "start to create kubernetes cluster %s" "${cluster_name}"
  ensure kind create cluster --config "${config_file}" --image="${kind_image}" --name="${cluster_name}" --retain -v 1
  ensure kind get kubeconfig --name="${cluster_name}" >"${kubeconfig_path}"
  ensure export KUBECONFIG="${kubeconfig_path}"
}

install_kind() {
  local kind_version=$1
  local force_install=$2

  printf "Install Kind tool\n"

  err_msg=$(kind version 2>&1 1>/dev/null)
  if [ "$err_msg" == "" ]; then
    v=$(kind version | awk '{print $2}' | sed s/v//g)
    target_version=$(echo "${kind_version}" | sed s/v//g)
    if version_lt "$v" "${target_version}"; then
      printf "Function Mesh requires Kind version %s or later\n" "${target_version}"
    else
      printf "Kind Version %s has been installed\n" "$v"
      if [ "$force_install" != "true" ]; then
        return
      fi
    fi
  fi

  local KIND_BIN="${HOME}/local/bin/kind"
  local target_os=$(lowercase $(uname))
  ensure curl -Lo /tmp/kind https://github.com/kubernetes-sigs/kind/releases/download/"$1"/kind-"${target_os}"-amd64
  ensure chmod +x /tmp/kind
  ensure mv /tmp/kind "$KIND_BIN"
}

install_function_mesh() {
  local release_name=$1
  local namespace=$2
  local crd=$3
  local runtime=$4
  local version=$5
  local docker_registry=$6
  printf "Install Function Mesh %s\n" "${release_name}"

  local function_mesh_image="${docker_registry}/streamnative/function-mesh:${version}"

  gen_crd_manifests "${crd}" | kubectl apply --validate=false -f - || exit 1
  gen_function_mesh_manifests "${release_name}" "${namespace}" "${runtime}" "${version}" "${host_network}" "${docker_registry}" | kubectl apply -f - || exit 1
}

version_lt() {
  vercomp $1 $2
  if [ $? == 2 ]; then
    return 0
  fi

  return 1
}

vercomp() {
  if [[ $1 == $2 ]]; then
    return 0
  fi
  local IFS=.
  local i ver1=($1) ver2=($2)
  # fill empty fields in ver1 with zeros
  for ((i = ${#ver1[@]}; i < ${#ver2[@]}; i++)); do
    ver1[i]=0
  done
  for ((i = 0; i < ${#ver1[@]}; i++)); do
    if [[ -z ${ver2[i]} ]]; then
      # fill empty fields in ver2 with zeros
      ver2[i]=0
    fi
    if ((10#${ver1[i]} > 10#${ver2[i]})); then
      return 1
    fi
    if ((10#${ver1[i]} < 10#${ver2[i]})); then
      return 2
    fi
  done
  return 0
}

check_docker() {
  need_cmd "docker"
  docker_err_msg=$(docker version 2>&1 1>/dev/null)
  if [ "$docker_err_msg" != "" ]; then
    printf "check docker failed:\n"
    echo "$docker_err_msg"
    exit 1
  fi
}

say() {
  printf 'install function-mesh: %s\n' "$1"
}

err() {
  say "$1" >&2
  exit 1
}

need_cmd() {
  if ! check_cmd "$1"; then
    err "need '$1' (command not found)"
  fi
}

check_cmd() {
  command -v "$1" >/dev/null 2>&1
}

lowercase() {
  echo "$@" | tr "[A-Z]" "[a-z]"
}

# Run a command that should never fail. If the command fails execution
# will immediately terminate with an error showing the failing
# command.
ensure() {
  if ! "$@"; then err "command failed: $*"; fi
}

ensure_pods_ready() {
  local namespace=$1
  local labels=""
  local limit=$3

  if [ "$2" != "" ]; then
    labels="-l $2"
  fi

  count=0
  while [ -n "$(kubectl get pods -n "${namespace}" ${labels} --no-headers | grep -v Running)" ]; do
    echo "Waiting for pod running" && sleep 10

    kubectl get pods -n "${namespace}" ${labels} --no-headers | grep >&2 -v Running || true

    ((count = count + 1))
    if [ $count -gt $limit ]; then
      printf "Waiting for pod status running timeout\n"
      exit 1
    fi
  done
}

gen_crd_manifests() {
  local crd=$1

  if check_url "$crd"; then
    need_cmd curl
    ensure curl -sSL "$crd"
    return
  fi

  ensure cat "$crd"
}

check_url() {
  local url=$1
  local regex='^(https?|ftp|file)://[-A-Za-z0-9\+&@#/%?=~_|!:,.;]*[-A-Za-z0-9\+&@#/%=~_|]\.[-A-Za-z0-9\+&@#/%?=~_|!:,.;]*[-A-Za-z0-9\+&@#/%=~_|]$'
  if [[ $url =~ $regex ]]; then
    return 0
  else
    return 1
  fi
}

gen_function_mesh_manifests() {
  local release_name=$1
  local namespace=$2
  local runtime=$3
  local version=$4
  local host_network=$5
  local docker_registry=$6

  VERSION_TAG="${version}"

  DOCKER_REGISTRY_PREFIX="${docker_registry}"
  # function-mesh.yaml start
  # below content is generated by "helm template function-mesh ./charts/function-mesh-operator"
  cat <<EOF

  # function-mesh.yaml end
---
# Source: function-mesh-operator/templates/controller-manager-rbac.yaml
kind: ServiceAccount
apiVersion: v1
metadata:
  name: function-mesh-controller-manager
  labels:
    app.kubernetes.io/name: function-mesh-operator
    app.kubernetes.io/instance: ${release_name}
    app.kubernetes.io/component: controller-manager
---
# Source: function-mesh-operator/templates/controller-manager-rbac.yaml
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: function-mesh-function-mesh-controller-manager
  labels:
    app.kubernetes.io/name: function-mesh-operator
    app.kubernetes.io/instance: ${release_name}
    app.kubernetes.io/component: controller-manager
rules:
- apiGroups:
    - ""
  resources:
    - services
    - events
  verbs:
    - "*"
- apiGroups:
    - ""
  resources:
    - configmaps
    - endpoints
  verbs:
    - get
    - list
    - watch
    - create
    - update
    - patch
    - delete
- apiGroups:
    - ""
  resources:
    - configmaps/status
  verbs:
    - get
    - update
    - patch
- apiGroups:
    - cloud.streamnative.io
  resources:
    - functionmeshes
  verbs:
    - create
    - delete
    - get
    - list
    - patch
    - update
    - watch
- apiGroups:
    - cloud.streamnative.io
  resources:
    - functionmeshes/status
  verbs:
    - get
    - patch
    - update
- apiGroups:
    - cloud.streamnative.io
  resources:
    - functions
  verbs:
    - create
    - delete
    - get
    - list
    - patch
    - update
    - watch
- apiGroups:
    - cloud.streamnative.io
  resources:
    - functions/status
  verbs:
    - get
    - patch
    - update
- apiGroups:
    - cloud.streamnative.io
  resources:
    - sinks
  verbs:
    - create
    - delete
    - get
    - list
    - patch
    - update
    - watch
- apiGroups:
    - cloud.streamnative.io
  resources:
    - sinks/status
  verbs:
    - get
    - patch
    - update
- apiGroups:
    - cloud.streamnative.io
  resources:
    - sources
  verbs:
    - create
    - delete
    - get
    - list
    - patch
    - update
    - watch
- apiGroups:
    - cloud.streamnative.io
  resources:
    - sources/status
  verbs:
    - get
    - patch
    - update
- apiGroups:
    - authentication.k8s.io
  resources:
    - tokenreviews
  verbs:
    - create
- apiGroups:
    - authorization.k8s.io
  resources:
    - subjectaccessreviews
  verbs:
    - create
- nonResourceURLs:
    - /metrics
  verbs:
    - get
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["create", "update", "get", "list", "watch","delete"]
- apiGroups: [""]
  resources: ["persistentvolumeclaims"]
  verbs: ["get", "list", "watch", "create", "update", "delete", "patch"]
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "watch","update", "delete"]
- apiGroups: ["apps"]
  resources: ["statefulsets","deployments", "controllerrevisions"]
  verbs: ["*"]
- apiGroups: ["extensions"]
  resources: ["ingresses"]
  verbs: ["*"]
- apiGroups: [""]
  resources: ["serviceaccounts"]
  verbs: ["create","get","update","delete"]
- apiGroups: [""]
  resources: ["nodes"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["persistentvolumes"]
  verbs: ["get", "list", "watch", "patch","update"]
- apiGroups: ["storage.k8s.io"]
  resources: ["storageclasses"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["rbac.authorization.k8s.io"]
  resources: [clusterroles,roles]
  verbs: ["escalate","create","get","update", "delete"]
- apiGroups: ["rbac.authorization.k8s.io"]
  resources: ["rolebindings","clusterrolebindings"]
  verbs: ["create","get","update", "delete"]
- apiGroups: ["autoscaling"]
  resources: ["horizontalpodautoscalers"]
  verbs: ["create", "delete", "patch", "update", "get", "watch", "list"]
---
# Source: function-mesh-operator/templates/controller-manager-rbac.yaml
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: function-mesh-function-mesh-controller-manager
  labels:
    app.kubernetes.io/name: function-mesh-operator
    app.kubernetes.io/instance: ${release_name}
    app.kubernetes.io/component: controller-manager
subjects:
  - kind: ServiceAccount
    name: function-mesh-controller-manager
    namespace: default
roleRef:
  kind: ClusterRole
  name: function-mesh-function-mesh-controller-manager
  apiGroup: rbac.authorization.k8s.io
---
# Source: function-mesh-operator/templates/controller-manager-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: function-mesh-controller-manager
  labels:
    app.kubernetes.io/name: function-mesh-operator
    app.kubernetes.io/instance: function-mesh
    app.kubernetes.io/component: controller-manager
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: function-mesh-operator
      app.kubernetes.io/instance: ${release_name}
      app.kubernetes.io/component: controller-manager
  template:
    metadata:
      labels:
        app.kubernetes.io/name: function-mesh-operator
        app.kubernetes.io/instance: ${release_name}
        app.kubernetes.io/component: controller-manager
    spec:
      serviceAccount: function-mesh-controller-manager
      containers:
      - name: function-mesh-operator
        image: streamnative/function-mesh:${VERSION_TAG}
        imagePullPolicy: IfNotPresent
        resources:
            requests:
              cpu: 80m
              memory: 50Mi
        command:
          - /usr/local/bin/function-mesh-controller-manager
        env:
          - name: NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
EOF

}

main "$@" || exit 1
