# Function Mesh Runner Images

Function Mesh uses runner images as functions / connectors' Pod images. Each runtime runner image only contains necessary tool-chains and libraries for specified runtime.

## Images
### Base Runner
The base Runner is located at `./pulsar-functions-base-runner`. The base runner contains basic tool-chains like `/pulsar/bin`, `/pulsar/conf` and `/pulsar/lib` to ensure that the `pulsar-admin` CLI tool works properly to support [Apache Pulsar Packages](http://pulsar.apache.org/docs/en/next/admin-api-packages/).

### Java Runner
[![Docker Image Version (tag latest semver)](https://img.shields.io/docker/v/streamnative/pulsar-functions-java-runner/2.7.1?style=for-the-badge)](https://hub.docker.com/r/streamnative/pulsar-functions-java-runner)

Java Runner is based on Base Runner, contains Java function instance to run Java functions/connectors.

`streamnative/pulsar-functions-java-runner` at Docker Hub will be automatically update align with Apache Pulsar Release.

### Python Runner
[![Docker Image Version (tag latest semver)](https://img.shields.io/docker/v/streamnative/pulsar-functions-python-runner/2.7.1?style=for-the-badge)](https://hub.docker.com/r/streamnative/pulsar-functions-python-runner)

Python Runner is based on Base Runner, contains Python function instance to run Python functions. User can build their own Python runner to customize python dependencies.

`streamnative/pulsar-functions-python-runner` at Docker Hub will be automatically update align with Apache Pulsar Release.

### Golang Runner
[![Docker Image Version (tag latest semver)](https://img.shields.io/docker/v/streamnative/pulsar-functions-go-runner/2.7.1?style=for-the-badge)](https://hub.docker.com/r/streamnative/pulsar-functions-go-runner)

Golang Runner provides all the toolchain and dependencies needed to run golang functions.

`streamnative/pulsar-functions-go-runner` at Docker Hub will be automatically update align with Apache Pulsar Release.

## How to use

### Use with [Apache Pulsar Packages](http://pulsar.apache.org/docs/en/next/admin-api-packages/)
#### Prerequisite

- Apache Pulsar 2.8.0
- Function Mesh v0.1.3 

#### Package your function locally
You can package Pulsar Functions in Java, Python and Go. Please refer to the [official document](http://pulsar.apache.org/docs/en/next/functions-package/) to packaging your Pulsar Functions.

#### Upload your function to Pulsar Cluster
With the packaged functions, user can upload the function package into [Apache Pulsar Packages](http://pulsar.apache.org/docs/en/next/admin-api-packages/).

You can upload a package to the package management service by `pulsar-admin`:

```bash
bin/pulsar-admin packages upload function://my-tenant/my-ns/my-function@0.1 --path package-file --description package-description
```

Then user can define Function Mesh CRDs to use the uploaded function package.

#### Submit to Function Mesh

Below is a sample CRD to create a Java function from package `function://my-tenant/my-ns/my-function@0.1`.
```yaml
apiVersion: compute.functionmesh.io/v1alpha1
kind: Function
metadata:
  name: java-function-sample
  namespace: default
spec:
  image: streamnative/pulsar-functions-java-runner:2.7.1 # using java function runner
  className: org.apache.pulsar.functions.api.examples.ExclamationFunction
  sourceType: java.lang.String
  sinkType: java.lang.String
  forwardSourceMessageProperty: true
  MaxPendingAsyncRequests: 1000
  replicas: 1
  maxReplicas: 5
  logTopic: persistent://public/default/logging-function-logs
  input:
    topics:
    - persistent://public/default/java-function-input-topic
  output:
    topic: persistent://public/default/java-function-output-topic
  pulsar:
    pulsarConfig: "test-pulsar"
  java:
    jar: my-function.jar # the package will download as this filename.
    jarLocation: function://my-tenant/my-ns/my-function@0.1 # function package URL
```

If you are not using self built runner image, you can skip `image`, Function Mesh manager will choose image according to the runtime defined in CRD. (for example, if user define `java` runtime settings, Function Mesh will use `streamnative/pulsar-functions-java-runner` by default)

### Use with self-built Docker image

#### Prerequisite

- Apache Pulsar 2.7.0
- Function Mesh v0.1.3 

#### Package your function locally
You can package Pulsar Functions in Java, Python and Go. Please refer to the [official document](http://pulsar.apache.org/docs/en/next/functions-package/) to packaging your Pulsar Functions.

#### Build your function image
You can write up a `Dockerfile` to build your function image packed with your function executable and all dependencies.
For example, with a packed Java function called `example-function.jar`, you can define a `Dockerfile` as below.

```dockerfile
# Use pulsar-functions-java-runner since we pack Java function
FROM streamnative/pulsar-functions-java-runner:2.7.1
# Copy function JAR package into /pulsar directory  
COPY example-function.jar /pulsar/
```

With the `Dockerfile`, you can build the function image and push into image registry (like Docker Hub, or any private registry)

#### Submit to Function Mesh

Assume we build the image from previous step as `streamnative/example-function-image:latest`, and pushed into Docker Hub already.

Below is a sample CRD to create a Java function from docker image.
```yaml
apiVersion: compute.functionmesh.io/v1alpha1
kind: Function
metadata:
  name: java-function-sample
  namespace: default
spec:
  image: streamnative/example-function-image:latest # using function image here
  className: org.apache.pulsar.functions.api.examples.ExclamationFunction
  sourceType: java.lang.String
  sinkType: java.lang.String
  forwardSourceMessageProperty: true
  MaxPendingAsyncRequests: 1000
  replicas: 1
  maxReplicas: 5
  logTopic: persistent://public/default/logging-function-logs
  input:
    topics:
    - persistent://public/default/java-function-input-topic
  output:
    topic: persistent://public/default/java-function-output-topic
  pulsar:
    pulsarConfig: "test-pulsar"
  java:
    jar: /pulsar/example-function.jar # the package location in image
    jarLocation: "" # leave empty since we will not download package from Pulsar Packages
```

When you use Function Mesh with self build images, you need to define the executable location within CRD yaml file, below are the settings for different function runtime.

```yaml
  java: # Java runtime
    jar: example-function.jar
  python: # Python runtime
    py: exclamation_function.py
  golang: # Golang runtime
    go: go_func_executable
```

## Scripts

`build.sh` is a bash script to build runner base, Java, Python, and Go runner locally.
