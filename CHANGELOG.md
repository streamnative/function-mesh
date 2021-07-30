# Function Mesh v0.1.6 Release Notes

## v0.1.6 What's New

- Move the Function runner image to Java 11 to support Pulsar 2.8.0. ([#202](https://github.com/streamnative/function-mesh/pull/202))
- Fix the value of the `MaxReplicas` in Function Mesh Worker service. ([#212](https://github.com/streamnative/function-mesh/pull/212))
- Update the observation logics of the sinks and sources. ([#217](https://github.com/streamnative/function-mesh/pull/217))
- Support Pulsar Functions and connectors (sinks and sources) status API in Function Mesh Worker service. ([#220](https://github.com/streamnative/function-mesh/pull/220)) ([#224](https://github.com/streamnative/function-mesh/pull/224)) ([#225](https://github.com/streamnative/function-mesh/pull/225))
- Allow set service account name for running pod ([#226](https://github.com/streamnative/function-mesh/pull/226))
- Fixed NPE when not set AutoAck ([#234](https://github.com/streamnative/function-mesh/pull/234))
- Fixed mesh worker service get status with fault pod ([#239](https://github.com/streamnative/function-mesh/pull/239))
- Remove logTopic from source/sink CRD ([#242](https://github.com/streamnative/function-mesh/pull/242))
- Fixed service name observe ([#230](https://github.com/streamnative/function-mesh/pull/230))
- Bump Pulsar to 2.8 ([#240](https://github.com/streamnative/function-mesh/pull/240))

# Function Mesh v0.1.5 Release Notes

## v0.1.5 What's New

- Remove descriptions in CRDs. ([#150](https://github.com/streamnative/function-mesh/pull/150))
- Refactor CI integration tests. ([#148](https://github.com/streamnative/function-mesh/pull/148))
- Support custom label with worker service. ([#125](https://github.com/streamnative/function-mesh/pull/125))
- Fix misuse between pulsar's namespace and k8s namespace. ([#125](https://github.com/streamnative/function-mesh/pull/125))
- Support plugin authn & authz with worker service. ([#125](https://github.com/streamnative/function-mesh/pull/125))
- Split TLS and Auth parameters. ([#152](https://github.com/streamnative/function-mesh/pull/152))
- Better exception log with worker service. ([#154](https://github.com/streamnative/function-mesh/pull/154))
- Rename pulsar cluster's configmap name postfix to `function-mesh-config`. ([#157](https://github.com/streamnative/function-mesh/pull/157))
- Fix requested resources not equals to the limit resources in worker service. ([#156](https://github.com/streamnative/function-mesh/pull/156))
- Fix de-register functions/sinks/sources with worker service. ([#159](https://github.com/streamnative/function-mesh/pull/159)) ([#168](https://github.com/streamnative/function-mesh/pull/168))
- Add `extraDependenciesDir` to Java runtime. ([#155](https://github.com/streamnative/function-mesh/pull/155))
- Load auth & TLS from secret. ([#160](https://github.com/streamnative/function-mesh/pull/160))
- Support `volumeMounts` in worker service. ([#158](https://github.com/streamnative/function-mesh/pull/158))
- Dont assign default zero value to `maxReplicas`. ([#161](https://github.com/streamnative/function-mesh/pull/161))
- Support `extraDependenciesDir` with worker service. ([#162](https://github.com/streamnative/function-mesh/pull/162))
- Support config `typeClassName` with worker service. ([#162](https://github.com/streamnative/function-mesh/pull/162))
- Remove max replicas to avoiding create hpa with worker service. ([#162](https://github.com/streamnative/function-mesh/pull/162))
- Use project-specified service account for function mesh operator. ([#163](https://github.com/streamnative/function-mesh/pull/163))
- Replace builtin prefix with worker service. ([#165](https://github.com/streamnative/function-mesh/pull/165))
- Avoid override jar and location value if `extraDependenciesDir` specified. ([#166](https://github.com/streamnative/function-mesh/pull/166))
- Add `olm.skipRange` to allow OLM upgrades. ([#167](https://github.com/streamnative/function-mesh/pull/167))
- Add resource limit configs in worker service. ([#171](https://github.com/streamnative/function-mesh/pull/171))
- Add API access and Jar uploading configs in worker service. ([#169](https://github.com/streamnative/function-mesh/pull/169))
- Fix resource validation in worker service. ([#172](https://github.com/streamnative/function-mesh/pull/172))
- Use hash for resource object name. ([#173](https://github.com/streamnative/function-mesh/pull/173))
- Allow config `ownerReference` on K8S resources created by worker service. ([#174](https://github.com/streamnative/function-mesh/pull/174))
- Support filter tenant and namespace. ([#176](https://github.com/streamnative/function-mesh/pull/176))
- Do not generate retry details or `receiverQueueSize` in the command by default. ([#184](https://github.com/streamnative/function-mesh/pull/184))
- make `CustomRuntimeOptions` optional in worker service. ([#179](https://github.com/streamnative/function-mesh/pull/179))
- Fix python runner build failed with Pulsar 2.8.0. ([#185](https://github.com/streamnative/function-mesh/pull/185))
- Support CRD installation with helm charts. ([#186](https://github.com/streamnative/function-mesh/pull/186))

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
