Function Mesh provides a tool, which is used to migrate Pulsar Functions from the function worker of a Pulsar cluster to Function Mesh. The tool encapsulates the `pulsarctl` configurations, which is used to transfer Pulsar configurations.

## Prerequisites

- Go 1.15 or higher

## Steps

1. Build the tool.

   1. Define a `pulsarctl.properties` configuration file for the tool.

        ```json
        webServiceUrl=http://localhost:8080
        tlsAllowInsecureConnection=false
        tlsTrustCertsFilePath=
        brokerServiceUrl=
        authParams=
        authPlugin=
        tlsEnableHostnameVerification=false
        ```

   2. Declare environment variables `PULSAR_CLIENT_CONF`.

       ```bash
       export PULSAR_CLIENT_CONF=/PATH/pulsarctl.properties
       ```

       Replace the `PATH` variable with the absolute path to the `pulsarctl.properties` configuration file.

   3. Build the tool from the source code.

       ```bash
       git clone https://github.com/streamnative/function-mesh
       cd function-mesh/tools
       go build
       ```

2. Generate the YAML file which is used to create the function in Function Mesh.

    ```bash
    ./tools
    ```

    The generated YAML file adopts a structure as below, where the `public`, `default`, and `function-sample` represent the tenant name, namespace name, and function name respectively.

    ```
    functions
    └── public
        └── default
            └── function-sample.yaml
    ```

    This is an example of the generated YAML file.

        ```yaml
        apiVersion: compute.functionmesh.io/v1alpha1
        kind: Function
        metadata:
          creationTimestamp: null
          name: function-sample
          namespace: default
        spec:
          autoAck: true
          className: org.apache.pulsar.functions.api.examples.ExclamationFunction
          cleanupSubscription: true
          clusterName: standalone
          funcConfig:
            PublishTopic: test_result
          input:
            sourceSpecs:
            test_src: {}
            topics:
            - test_src
          # other function config
        ```

After the YAML file is generated, you can use the `kubectl apply -f` command to create the function.

```shell
kubectl apply -f /path/to/function-sample.yaml
```