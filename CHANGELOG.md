# Function Mesh v0.5.0 Release Notes

## v0.5.0 What's New

- [Test] Add tests for packages download ([#429](https://github.com/streamnative/function-mesh/pull/429))
- [Test] refine test cases of the Sink/Source/Crypto function ([#435](https://github.com/streamnative/function-mesh/pull/435))
- [Test] Use v1.8.2 for cert-manager tests ([#433](https://github.com/streamnative/function-mesh/pull/433))
- [Helm Charts] Add handle logic for FunctionMesh to clean up orphaned subcomponents ([#431](https://github.com/streamnative/function-mesh/pull/431))
- [Controller] Add the ability to set logging levels ([#445](https://github.com/streamnative/function-mesh/pull/445))
- [Controller] Update CSV for openshift ([#417](https://github.com/streamnative/function-mesh/pull/417))
- [Controller] Add sky-walking workflow for self-signed tls pulsar cluster ([#439](https://github.com/streamnative/function-mesh/pull/439))
- [Controller] Add a limit to the resource name ([#437](https://github.com/streamnative/function-mesh/pull/437/files))
- [Function Mesh Worker service] Support subscriptionPosition for function&sink ([#322](https://github.com/streamnative/function-mesh-worker-service/pull/322))

For the full changes in this release, see the [Pull Requests](https://github.com/streamnative/function-mesh/pulls?q=is%3Apr+label%3Am%2F2022-08+is%3Aclosed).

# Function Mesh v0.4.0 Release Notes

## v0.4.0 What's New

- [Makefile] Add `kustomize` before `crd` in the `release` command in Makefile ([#395](https://github.com/streamnative/function-mesh/pull/395))
- [Makefile] Remove `helm-crds` from the `release` command in Makefile ([#396](https://github.com/streamnative/function-mesh/pull/396))
- [Makefile] Fix some small typos in README and Makefile ([#423](https://github.com/streamnative/function-mesh/pull/423))
- [Test] Fix failed actions to wait for the controller POD ready in Github Action ([#406](https://github.com/streamnative/function-mesh/pull/406))
- [Test] Add the `paths-ignore` configuration to allow changing some files without triggering the CI test workflow. ([#410](https://github.com/streamnative/function-mesh/pull/410))
- [Test] Update the CI test workflow ([#420](https://github.com/streamnative/function-mesh/pull/420))
- [Test] Add the `make e2e` command for E2E tests based on [SkyWalking Infra E2E](https://github.com/apache/skywalking-infra-e2e) ([#422](https://github.com/streamnative/function-mesh/pull/422))
- [Helm Charts] Remove invalid caBundle in webhook patches ([#398](https://github.com/streamnative/function-mesh/pull/398))
- [Helm Charts] Improve tools for generating helm chart templates ([#409](https://github.com/streamnative/function-mesh/pull/409))
- [Controller] Update the codeowners ([#407](https://github.com/streamnative/function-mesh/pull/407))
- [Controller] Fix comments on the Dockerfile ([#405](https://github.com/streamnative/function-mesh/pull/405))
- [Controller] Remove useless omitempty ([#402](https://github.com/streamnative/function-mesh/pull/402))
- [Controller] Support mounting secrets to provide sensitive information to Function Mesh ([#414](https://github.com/streamnative/function-mesh/pull/414))
- [Controller] Allow key-value environment variable passthrough ([#424](https://github.com/streamnative/function-mesh/pull/424))
- [Controller] Use `ctrl.CreateOrUpdate` to determine whether a resource should be updated or created ([#427](https://github.com/streamnative/function-mesh/pull/427))

For the full changes in this release, see the [Pull Requests](https://github.com/streamnative/function-mesh/pulls?q=is%3Apr+label%3Am%2F2022-07+is%3Aclosed).

# Function Mesh v0.3.0 Release Notes

## v0.3.0 What's New

- [Helm Charts] Support deploying the of webhook-enabled operator through Helm Charts ([#319](https://github.com/streamnative/function-mesh/issues/319))
- [Controller] Support converting memory values to a decimal format ([#375](https://github.com/streamnative/function-mesh/issues/375))
- [Controller] Support specifying an object as an unmanaged object ([#376](https://github.com/streamnative/function-mesh/issues/376), [#377](https://github.com/streamnative/function-mesh/pull/377))
- [Controller] Fix the Python runner Protobuf version ([#379](https://github.com/streamnative/function-mesh/issues/379))
- [Controller] Pass `LeaderElectionNamespace` when running the Function Mesh operator outside of a Kubernetes cluster ([#382](https://github.com/streamnative/function-mesh/issues/382))

For the full changes in this release, see the [Pull Requests](https://github.com/streamnative/function-mesh/pulls?q=is%3Apr+label%3Am%2F2022-06+is%3Aclosed).

# Function Mesh v0.2.0 Release Notes

## v0.2.0 What's New

- [Helm Charts] Release the FunctionMesh operator to helm charts repository ([#356](https://github.com/streamnative/function-mesh/issues/356))
- [Controller] Support stateful functions for Pulsar Functions ([#325](https://github.com/streamnative/function-mesh/issues/325))
- [Controller] Migrate to golang 1.18 ([#366](https://github.com/streamnative/function-mesh/issues/366))

For the full changes in this release, see the [Pull Requests](https://github.com/streamnative/function-mesh/pulls?q=is%3Apr+label%3Am%2F2022-05+is%3Aclosed).

# Function Mesh v0.1.11 Release Notes

## v0.1.11 What's New

- [Helm Charts] Support more controller manager parameters for Helm charts ([#339](https://github.com/streamnative/function-mesh/pull/339))
- [Helm Charts] Fix Helm charts installation issues ([#347](https://github.com/streamnative/function-mesh/pull/347)), ([#346](https://github.com/streamnative/function-mesh/pull/346)), ([#345](https://github.com/streamnative/function-mesh/pull/345)), ([#344](https://github.com/streamnative/function-mesh/pull/344))
- [Helm Charts] Support namespace configurations for Helm charts ([#327](https://github.com/streamnative/function-mesh/pull/327))
- [Controller] Apply the Istio service port name convention to services ([#343](https://github.com/streamnative/function-mesh/pull/343))
- [Controller] Fix the Webhook service ([#341](https://github.com/streamnative/function-mesh/pull/341))
- [Controller] Move `apiextensions.k8s.io/v1beta1` to `apiextensions.k8s.io/v1` ([#328](https://github.com/streamnative/function-mesh/pull/328))
- [Controller] fix service `controller-manager-metrics-service` wrong selector ([#333](https://github.com/streamnative/function-mesh/pull/333))
- [Controller] expose more parameters for manager ([#329](https://github.com/streamnative/function-mesh/pull/329))
- [Controller] add controller config file ([#322](https://github.com/streamnative/function-mesh/pull/322))
- [Controller] Add annotation to exclude the webhook port from Istio proxying ([#321](https://github.com/streamnative/function-mesh/pull/321))
- [OLM] Increase resource limit for manager ([#324](https://github.com/streamnative/function-mesh/pull/324))
- [Runner Images] Verify Pulsar 2.10 ([#336](https://github.com/streamnative/function-mesh/pull/336))
- [Runner Images] Bump sample dependencies ([#317](https://github.com/streamnative/function-mesh/pull/317)), ([#318](https://github.com/streamnative/function-mesh/pull/318))
- [Test] Bump versions for the test script toolsets ([#334](https://github.com/streamnative/function-mesh/pull/334))
- [Test] Remove unused steps from the test scripts ([#331](https://github.com/streamnative/function-mesh/pull/331))

# Function Mesh v0.1.10 Release Notes

## v0.1.10 What's New

- [Controller] Remove unused `CAP_FOWNER` Capability from SecurityContext ([#313](https://github.com/streamnative/function-mesh/pull/313))
- [Controller] Fix Json marshal nil to `"null"` and break FunctionDetails ([#312](https://github.com/streamnative/function-mesh/pull/312))
- [Controller] Sync function, source, or sink configurations to the Pod when they are updated. ([#262](https://github.com/streamnative/function-mesh/pull/262))

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
