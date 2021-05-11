# Function Mesh v0.1.4 Release Notes

## v0.1.4 What's New

Action required: Users to manage functions with `mesh worker service`, should notice that we changed custom parameters from `--user-config` to `--custom-runtime-options`, in order to support passing custom parameters from `pulsar-admin` APIs for both functions and connectors.

- Change K8S Api group/domain to `compute.functionmesh.io`. ([#118](https://github.com/streamnative/function-mesh/pull/118))
- Add custom Pod image and image pull policy. ([#104](https://github.com/streamnative/function-mesh/pull/104))
- Add runner base image, and Java, Python, Golang runner image based on runner base. ([#95](https://github.com/streamnative/function-mesh/pull/95))
- Add KeyBased Batcher support for Java Functions in Function Mesh. ([#112](https://github.com/streamnative/function-mesh/pull/112))
- Allow user to specify `typeClassName` for both `Input` and `Output` topics. ([#133](https://github.com/streamnative/function-mesh/pull/133))
- Change `Java-Proxy` to `Mesh Worker Service`. ([#138](https://github.com/streamnative/function-mesh/pull/138))
- Support OLM bundle. ([#119](https://github.com/streamnative/function-mesh/pull/119))
- Support manage connectors with `pulsar-admin` and `Mesh Worker Service`. ([#126](https://github.com/streamnative/function-mesh/pull/126))
- Add migration tools to let user migrate existing Pulsar Functions to Function Mesh. ([#80](https://github.com/streamnative/function-mesh/pull/80))
- Add example function image. ([#144](https://github.com/streamnative/function-mesh/pull/144))
- NPE fix when create Function CRD. ([#109](https://github.com/streamnative/function-mesh/pull/109))
- Bump k8s client dependency to 10.0.1. ([#121](https://github.com/streamnative/function-mesh/pull/121))
- Runner images on Apache Pulsar 2.7.1
    - [![](https://img.shields.io/badge/streamnative%2Fpulsar--functions--java--runner-2.7.1-green)](https://hub.docker.com/r/streamnative/pulsar-functions-java-runner)
    - [![](https://img.shields.io/badge/streamnative%2Fpulsar--functions--python--runner-2.7.1-green)](https://hub.docker.com/r/streamnative/pulsar-functions-python-runner)
    - [![](https://img.shields.io/badge/streamnative%2Fpulsar--functions--go--runner-2.7.1-green)](https://hub.docker.com/r/streamnative/pulsar-functions-go-runner)
