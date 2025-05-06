# Function-Mesh
A Kubernetes-Native way to run pulsar functions, connectors and composed function meshes.

## Install

```bash
curl -sSL https://github.com/streamnative/function-mesh/releases/download/v0.24.1-rc.1/install.sh | bash
```

The above command installs all the CRDs, required service account configuration, and all function-mesh operator components. Before you start running a function-mesh example, verify if Function Mesh is installed correctly.

Note:

> install.sh is suitable for trying Function Mesh out. If you want to use Function Mesh in production or other serious scenarios, Helm is the recommended deployment method.

## Prerequisite
- Git
- [operator-sdk](https://sdk.operatorframework.io/)
- [kubernetes cluster](https://kubernetes.io/)
- [pulsar-cluster](https://pulsar.apache.org/docs/en/pulsar-2.0/)
- [pulsar-functions](https://pulsar.apache.org/docs/en/functions-overview/)

## Compatibility

### Kubernetes compatibility matrix

This table outlines the supported Kubernetes versions. We have tested these versions in their respective branches. But note that other versions might work as well.

| Function Mesh operator                                                          | Kubernetes 1.19 | Kubernetes 1.20 | Kubernetes 1.21 | Kubernetes 1.22 | Kubernetes 1.23 | Kubernetes 1.24 | Kubernetes 1.25 |
|---------------------------------------------------------------------------------|---------------|--------------|---------------|----------------|-----------------| --------------- | --------------- |
| [`v0.20.0`](https://github.com/streamnative/function-mesh/releases/tag/v0.20.0) | ✔             | ✔            | ✔             | ✔               | ✔               | ✔               | ✔               |
| [`v0.19.0`](https://github.com/streamnative/function-mesh/releases/tag/v0.19.0) | ✔             | ✔            | ✔             | ✔               | ✔               | ✔               | ✔               |
| [`v0.18.0`](https://github.com/streamnative/function-mesh/releases/tag/v0.18.0) | ✔             | ✔            | ✔             | ✔               | ✔               | ✔               | ✔               |
| [`v0.17.0`](https://github.com/streamnative/function-mesh/releases/tag/v0.17.0) | ✔             | ✔            | ✔             | ✔               | ✔               | ✔               | ✔               |
| [`v0.16.0`](https://github.com/streamnative/function-mesh/releases/tag/v0.16.0) | ✔             | ✔            | ✔             | ✔               | ✔               | ✔               | ✔               |
| [`v0.15.0`](https://github.com/streamnative/function-mesh/releases/tag/v0.15.0) | ✔             | ✔            | ✔             | ✔               | ✔               | ✔               | ✔               |
| [`v0.14.0`](https://github.com/streamnative/function-mesh/releases/tag/v0.14.0) | ✔             | ✔            | ✔             | ✔              | ✔               | ✔               | ✔               |
| [`v0.13.0`](https://github.com/streamnative/function-mesh/releases/tag/v0.13.0) | ✔             | ✔            | ✔             | ✔              | ✔               | ✔               | ✔               |
| [`v0.12.0`](https://github.com/streamnative/function-mesh/releases/tag/v0.12.0) | ✔             | ✔            | ✔             | ✔              | ✔               | ✔               | ✔               |
| [`v0.11.2`](https://github.com/streamnative/function-mesh/releases/tag/v0.11.2) | ✔             | ✔            | ✔             | ✔              | ✔               | ✔               | ✔               |
| [`Master`](https://github.com/streamnative/function-mesh/tree/master)           | ✔             | ✔            | ✔             | ✔               | ✔               | ✔               | ✔               |

## Development

- install Git and download the repo

```bash
git clone https://github.com/streamnative/function-mesh.git
```

- install operator-sdk and use it to add CRD, controller or webhooks

> **Note**
>
> The following command will generate the scaffolding files in the `api/<version>/` path, in this case `api/v1alpha1`, and then you need to move the files contained in it to the `api/<group>/<version>` directory manually, in this case `api/compute/v1alpha1`.

```bash
operator-sdk create api --group compute --version v1alpha1 --kind Function --resource=true --controller=true
```

```bash
operator-sdk create webhook --group compute.functionmesh.io --version v1alpha1 --kind Function --defaulting --programmatic-validation
```

## Deployment

1. make sure connected to a kubernetes cluster(gke, mini-kube etc.)
    ```bash
    gcloud container clusters get-credentials cluster-1 --region $CLUSTER_REGION --project $PROJECT_ID
    ```
2. compile the repo to generate related resources in the root dir of the repo.
    ```bash
    make generate
    ```
3. install the CRD into your k8s cluster.
    ```bash
    make install
    ```
4. start the controller locally. Only the controller itself is running in your local terminal, all the resources will be running inside the connected kubernetes cluster
    ```bash
    make run
    ```
5. submit a sample CRD to the cluster. You can also submit other CRDs under the `./config/samples` directory
    ```bash
    kubectl apply -f config/samples/compute_v1alpha1_function.yaml
    ```
6. verify your submission with `kubectl`, and you will see the function pod is running
    ```bash
    $ kubectl get all
    NAME                                READY   STATUS      RESTARTS   AGE
    pod/function-sample-0               1/1     Running     0          77s
    ```
7. in order for function actually work, you need to have a pulsar cluster available for visiting. you can use the [helm-chart](https://pulsar.apache.org/docs/en/helm-overview/) to deploy one
