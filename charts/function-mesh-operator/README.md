# function-mesh-operator

![Version: 0.2.7](https://img.shields.io/badge/Version-0.2.7-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.7.0](https://img.shields.io/badge/AppVersion-0.7.0-informational?style=flat-square)

The Function Mesh operator Helm chart for Kubernetes

**Homepage:** <https://github.com/streamnative/function-mesh>

## Maintainers

| Name | Email                                | URL |
| ---- |--------------------------------------| --- |
| Function Mesh | mailto:function-mesh@streamnative.io | https://github.com/streamnative/function-mesh |

## Source Code

* <https://github.com/streamnative/function-mesh>

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| (Built-in) | admission-webhook | 0.2.7 |

## Values

| Key | Type | Default |
|-----|------|---------|
| admissionWebhook.enabled | bool | `true` |
| controllerManager.affinity | object | `{}` |
| controllerManager.autoFailover | bool | `true` |
| controllerManager.configFile | string | `"/etc/config/config.yaml"` |
| controllerManager.create | bool | `true` |
| controllerManager.enableLeaderElection | bool | `true` |
| controllerManager.healthProbe.port | int | `8000` |
| controllerManager.metrics.port | int | `8080` |
| controllerManager.nodeSelector | object | `{}` |
| controllerManager.pprof.enable | bool | `false` |
| controllerManager.pprof.port | int | `8090` |
| controllerManager.replicas | int | `1` |
| controllerManager.resources.requests.cpu | string | `"80m"` |
| controllerManager.resources.requests.memory | string | `"50Mi"` |
| controllerManager.selector | list | `[]` |
| controllerManager.serviceAccount | string | `"function-mesh-controller-manager"` |
| controllerManager.tolerations | list | `[]` |
| imagePullPolicy | string | `"IfNotPresent"` |
| imagePullSecrets | list | `[]` |
| installation.namespace | string | `"function-mesh-system"` |
| operatorImage | string | `"streamnative/function-mesh:v0.7.0"` |
| rbac.create | bool | `true` |

----------------------------------------------
## Function Mesh Helm Charts Usage

### Install cert-manager

Function Mesh is enabled with the [admission control webhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#what-are-admission-webhooks) by default. Therefore, you need to prepare the relevant signed certificate. Secrets that contain signed certificates are named with the fixed name `function-mesh-admission-webhook-server-cert`, which is controlled by the [Certificate CRD](https://cert-manager.io/docs/concepts/certificate/).

It is recommended to use [cert-manager](https://cert-manager.io/) to manage these certificates and you can install the cert-manager as follows.

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

### Install Function Mesh

> **Note**
>
> - Before installation, ensure that Helm v3 is installed properly.
> - For the use of `kubectl` commands, see [kubectl command reference](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands).

1. Add the StreamNative Function Mesh repository.

    ```shell
    helm repo add function-mesh http://charts.functionmesh.io/
    helm repo update
    ```

2. Install the Function Mesh Operator.

    Let's set some variables for convenient use later.

    ```shell
    export FUNCTION_MESH_RELEASE_NAME=function-mesh  # change the release name according to your scenario
    export FUNCTION_MESH_RELEASE_NAMESPACE=function-mesh  # change the namespace to where you want to install Function Mesh
    ```

    Install the Function Mesh Operator via following command.

    > **Note**
    >
    > - If no Kubernetes namespace is specified, the `default` namespace is used.
    >
    > - If the namespace ${FUNCTION_MESH_RELEASE_NAMESPACE} doesn't exist yet, you can add the parameter `--create-namespace ` to create it automatically.

    ```shell
    helm install ${FUNCTION_MESH_RELEASE_NAME} function-mesh/function-mesh-operator -n ${FUNCTION_MESH_RELEASE_NAMESPACE}
    ```

    You can refer to the [Values](#values) table above to learn about the configurable parameters of the Function Mesh operator and their default values.

    For example, if you want to enable `pprof` for the Function Mesh Operator, set the `controllerManager.pprof.enable` to `true`.
    
    ```shell
    helm install ${FUNCTION_MESH_RELEASE_NAME} function-mesh/function-mesh-operator -n ${FUNCTION_MESH_RELEASE_NAMESPACE} \
      --set controllerManager.pprof.enable=true
    ```
    
3. Check whether Function Mesh is installed successfully.

    ```shell
    kubectl get pods --namespace ${FUNCTION_MESH_RELEASE_NAMESPACE} -l app.kubernetes.io/instance=function-mesh
    ```

    **Output**

    ```
    NAME                                               READY   STATUS    RESTARTS   AGE
    function-mesh-controller-manager-5f867557c-d6vf4   1/1     Running   0          8s
    ```

### Uninstall Function Mesh

1. Use the following command to uninstall Function Mesh through Helm.

   > **Note**
   >
   > `${NAMESPACE}` indicates the namespace where Function Mesh Operator is installed.

   ```bash
   helm delete function-mesh -n ${NAMESPACE}
   ```

2. Remove the Secrets that contain the signed certificate.

   > **Note**
   >
   > If the Secrets are not cleaned up, future installations in this environment might behave abnormally. For details about how to automatically clean up the corresponding Secrets when you delete a Certificate, see [Cleaning up Secrets when Certificates are deleted](https://cert-manager.io/docs/usage/certificate/#cleaning-up-secrets-when-certificates-are-deleted).

   ```shell
   kubectl delete secret function-mesh-admission-webhook-server-cert -n ${NAMESPACE}
   ```

*Autogenerated from chart metadata using [helm-docs v1.11.0](https://github.com/norwoodj/helm-docs/releases/v1.11.0)*
