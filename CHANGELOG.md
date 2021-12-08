# Function Mesh v0.1.9 Release Notes

## v0.1.9 What's New

Action required: We have moved the Function Mesh Worker service into a separate repository [streamnative/function-mesh-worker-service](https://github.com/streamnative/function-mesh-worker-service). Therefore, the `function-mesh` repo will not include releases of the Function Mesh Worker service.

- [Runner Images] Make runner images rootless ([#278](https://github.com/streamnative/function-mesh/pull/278))
- [Runner Images] Simplify the runner image layers ([#292](https://github.com/streamnative/function-mesh/pull/292))
- [Controller] Support CRD validation with WebHook ([#238](https://github.com/streamnative/function-mesh/pull/238), [#299](https://github.com/streamnative/function-mesh/pull/299))
- [Controller] Support `EnvironmentBasedSecretsProvider` as the default secret provider ([#295](https://github.com/streamnative/function-mesh/pull/295))
- [Controller] Provide the default `SecurityContext` to run Pulsar Functions and connectors in the rootless mode ([#294](https://github.com/streamnative/function-mesh/pull/294))
- [Function Mesh Worker service] Support customizing runner images ([#291](https://github.com/streamnative/function-mesh/pull/291))
- [Function Mesh Worker service] Fix the default value for `ForwardSourceMessageProperty` ([#300](https://github.com/streamnative/function-mesh/pull/300))
- [Function Mesh Worker service] Use Pulsar Package Management Service as the backend and redirect uploading JAR requests to the backend ([#308](https://github.com/streamnative/function-mesh/pull/308))
- [Function Mesh Worker service] Move the Function Mesh Worker service into a separate repository ([#1](https://github.com/streamnative/function-mesh-worker-service/issues/1))
- [Function Mesh Worker service] Support `secretsMap` on branch-2.8 ([#21](https://github.com/streamnative/function-mesh-worker-service/pull/21))

> **Note**
>
> From this release, the release note will not contain any changes to the `function-mesh-worker-service` repository.

# Function Mesh v0.1.8 Release Notes

## v0.1.8 What's New

- [Controller] Update to Pulsar 2.8 Function's proto files ([#274](https://github.com/streamnative/function-mesh/pull/274))
- [Function Mesh Worker service] Support custom `ServiceAccountName` and `DefaultServiceAccountName` ([#276](https://github.com/streamnative/function-mesh/pull/276))
- [Controller] Fix the issue that HPA does work as expected on Sink and Source resources ([#281](https://github.com/streamnative/function-mesh/pull/281))
- [Function Mesh Worker service] Clean up temporarily downloaded files ([#282](https://github.com/streamnative/function-mesh/pull/282))
- [Function Mesh Worker service] Fix the resource convert exception with the Kubernetes Client ([#286](https://github.com/streamnative/function-mesh/pull/286))
- [Controller] Added function-mesh labels to Service and HPA ([#287](https://github.com/streamnative/function-mesh/pull/287))
- [Controller] Bump to Pulsar 2.8.1 ([#288](https://github.com/streamnative/function-mesh/pull/288))

# Function Mesh v0.1.7 Release Notes

## v0.1.7 What's New

- [Function Mesh Worker service] Add `typeClassName` in the source and sink connector definition ([#247](https://github.com/streamnative/function-mesh/pull/247))
- [Test] Add integration tests for Function Mesh Worker service ([#246](https://github.com/streamnative/function-mesh/pull/246))
- [Controller] Use untyped YAML configuration as function or connector configuration ([#233](https://github.com/streamnative/function-mesh/pull/233))
- [Controller] Enable source connector to set `ForwardMessageProperty` ([#250](https://github.com/streamnative/function-mesh/pull/250))
- [Runner image] Find the issue that the Python runner image cannot be built based on the latest Pulsar image ([#253](https://github.com/streamnative/function-mesh/pull/253)) ([#254](https://github.com/streamnative/function-mesh/pull/254))
- [controller] fix package download path not same as the executing command path ([#256](https://github.com/streamnative/function-mesh/pull/256))
- [controller] support metrics based HPA by upgrade apis to k8s.io/api/autoscaling/v2beta2 ([#245](https://github.com/streamnative/function-mesh/pull/245))
- [Controller] Fix misuse of Kubernetes namespace in Controller ([#259](https://github.com/streamnative/function-mesh/pull/259))
- [Controller] Support downloading the package with authentication or TLS ([#257](https://github.com/streamnative/function-mesh/pull/257))
- [Function Mesh Worker service] Support creating a function with the package URL ([#261](https://github.com/streamnative/function-mesh/pull/261))
- [Test] Add integration tests for Pulsar package management service ([#268](https://github.com/streamnative/function-mesh/pull/268))
- [Codebase] Remove the vendor folder in the codebase ([#267](https://github.com/streamnative/function-mesh/pull/267))

# Function Mesh v0.1.6 Release Notes

## v0.1.6 What's New

- Move the Function runner image to Java 11 to support Pulsar 2.8.0. ([#202](https://github.com/streamnative/function-mesh/pull/202))
- Fix the value of the `MaxReplicas` in Function Mesh Worker service. ([#212](https://github.com/streamnative/function-mesh/pull/212))
- Update the observation logics of the sinks and sources. ([#217](https://github.com/streamnative/function-mesh/pull/217))
- Support Pulsar Functions and connectors (sinks and sources) status API in Function Mesh Worker service. ([#220](https://github.com/streamnative/function-mesh/pull/220)) ([#224](https://github.com/streamnative/function-mesh/pull/224)) ([#225](https://github.com/streamnative/function-mesh/pull/225))
- Support setting the service account name for running Pods. ([#226](https://github.com/streamnative/function-mesh/pull/226))
- Fix the NPE that is generated when the `AutoAck` option is not set. ([#234](https://github.com/streamnative/function-mesh/pull/234))
- Fix the issue that the Function Mesh worker service gets status with a fault pod. ([#239](https://github.com/streamnative/function-mesh/pull/239))
- Remove the `logTopic` option from source and sink CRDs. ([#242](https://github.com/streamnative/function-mesh/pull/242))
- Fix the observation logic of the service name. ([#230](https://github.com/streamnative/function-mesh/pull/230))
- Bump Pulsar to version 2.8.0. ([#240](https://github.com/streamnative/function-mesh/pull/240))

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
