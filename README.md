## Function Mesh Helm Charts Usage

[Helm](https://helm.sh) must be installed to use the charts.  Please refer to Helm's [documentation](https://helm.sh/docs) to get started.

Once Helm has been set up correctly, add the repo as follows:

```shell
helm repo add function-mesh http://charts.functionmesh.io/
```

If you had already added the this repo earlier, run `helm repo update` to retrieve the latest versions of the packages.

You can then run `helm search repo function-mesh` to see the charts.

Let's set some variables for convenient use later.

> Note
>
> - If no Kubernetes namespace is specified, the default namespace is used.
>
> - If the namespace ${FUNCTION_MESH_RELEASE_NAMESPACE} doesn't exist yet, you can add the parameter --create-namespace to create it automatically.

```shell
export FUNCTION_MESH_RELEASE_NAME=function-mesh  # change the release name according to your scenario
export FUNCTION_MESH_RELEASE_NAMESPACE=function-mesh  # change the namespace to where you want to install Function Mesh
```

## Prerequisites

Function Mesh has enabled admission control webhook by default, so we need to prepare the relevant certificate information.

### Option 1 (default)：Use Cert Manager

You can install Cert Manager as follows:

```shell
helm repo add jetstack https://charts.jetstack.io
helm repo update
helm install \
  cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.8.0 \
  --set installCRDs=true
```

### Option 2: Use custom secrets

If you don't want to introduce Cert Manager, you can use a custom certificate, which will come in the form of a Secrets resource, and you can install it as follows:

```shell
helm install ${FUNCTION_MESH_RELEASE_NAME}-secrets function-mesh/function-mesh-secrets-webhook -n ${FUNCTION_MESH_RELEASE_NAMESPACE}
```

## Install

Install the Function Mesh Operator via following command.

To install the `function-mesh-operator` chart:

```shell
helm install ${FUNCTION_MESH_RELEASE_NAME} function-mesh/function-mesh-operator -n ${FUNCTION_MESH_RELEASE_NAMESPACE}
```

To uninstall the chart:                   

```shell
helm delete ${FUNCTION_MESH_RELEASE_NAME}
```