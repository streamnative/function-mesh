# Function Mesh v0.25.1 Release Notes

## v0.25.1 What's New

* [Runtime] Use POD_NAME as the name for filebeat sidecar container ([#817](https://github.com/streamnative/function-mesh/pull/817))

# Function Mesh v0.25.0 Release Notes

## v0.25.0 What's New

* [Dependency] Upgrade go to 1.24.4 ([#813](https://github.com/streamnative/function-mesh/pull/813))
* [CRD] Add default anti affinity ([#810](https://github.com/streamnative/function-mesh/pull/810))
* [Runtime] Use java21 for pulsar 4.0+ ([#809](https://github.com/streamnative/function-mesh/pull/809))
* [Controller] Add more fields to BackendConfig ([#804](https://github.com/streamnative/function-mesh/pull/804))

# Function Mesh v0.24.0 Release Notes

## v0.24.0 What's New

* [CRD] Change field processingGuarantee to processingGuarantees for WindowConfig ([#794](https://github.com/streamnative/function-mesh/pull/794))
* [Controller] Making hpa utilization with factor ([#795](https://github.com/streamnative/function-mesh/pull/795))
* [Controller] Update MaxRAMPercentage to 70 ([#797](https://github.com/streamnative/function-mesh/pull/797)
* [Controller] make function path consistency when initContainer downloader enabled ([#798](https://github.com/streamnative/function-mesh/pull/798))
* [Image] Fix CVEs ([#800](https://github.com/streamnative/function-mesh/pull/800))

# Function Mesh v0.23.0 Release Notes

## v0.23.0 What's New

* [Controller] Remove `/pulsar/lib/*` from classPath ([#787](https://github.com/streamnative/function-mesh/pull/787))
* [CRD] Add ProcessGuarantee to WindowConfig ([#792](https://github.com/streamnative/function-mesh/pull/792))
* [Controller] added feature to provide imagePullSecrets and imagePullPolicy to runner images when pulling ([#791](https://github.com/streamnative/function-mesh/pull/791))
* [Image] Support set python version when build python runner image ([#784](https://github.com/streamnative/function-mesh/pull/784))

For the full changes in this release, see the full changelog: [v0.22.0...v0.23.0](https://github.com/streamnative/function-mesh/compare/v0.22.0...v0.23.0)

# Function Mesh v0.22.0 Release Notes

## v0.22.0 What's New

* [Image] Update runner image base to alpine ([#747](https://github.com/streamnative/function-mesh/pull/747))
* [Image] Install gcompat for Pulsar 2.10.x ([#756](https://github.com/streamnative/function-mesh/pull/756))
* [Controller] Add liveness in backend config ([#767](https://github.com/streamnative/function-mesh/pull/767)) 
* [Controller] Change type of TerminationGracePeriodSeconds to `*int64` ([#778](https://github.com/streamnative/function-mesh/pull/778))

For the full changes in this release, see the full changelog: [v0.21.6...v0.22.0](https://github.com/streamnative/function-mesh/compare/v0.21.6...v0.22.0)

# Function Mesh v0.21.0 Release Notes

## v0.21.0 What's New

* [Controller] Add BackendConfig crd to provider cluster wide and namespace wide configs ([#734](https://github.com/streamnative/function-mesh/pull/734))
* [Controller] Remove initial ram percentage ([#724](https://github.com/streamnative/function-mesh/pull/724))
* [Image] Update versions of golang and alpine ([#743](https://github.com/streamnative/function-mesh/pull/743))
* [Controller] Add exit_on_oom=tru ([#744](https://github.com/streamnative/function-mesh/pull/744))

For the full changes in this release, see the [Pull Requests](https://github.com/streamnative/function-mesh/pulls?q=is%3Apr+label%3Am%2F2024-04+is%3Aclosed+)

# Function Mesh v0.20.0 Release Notes

## v0.20.0 What's New

* [Controller] Delete HPA when it's disabled ([#726](https://github.com/streamnative/function-mesh/pull/726))
* [Helm Charts] Add annotation to exclude the webhook port from Istio proxying ([#728](https://github.com/streamnative/function-mesh/pull/728))
* [Controller] Use numeric uid:gid in Dockerfile to support Tanzu's PSP ([#720](https://github.com/streamnative/function-mesh/pull/720))
* [Release] Fix redhat api changes for bundle release ([#719](https://github.com/streamnative/function-mesh/pull/719))
* [CI] Use /dev/sdb1 to save docker's data in CI ([#723](https://github.com/streamnative/function-mesh/pull/723))

For the full changes in this release, see the [Pull Requests](https://github.com/streamnative/function-mesh/pulls?q=is%3Apr+label%3Am%2F2024-02+is%3Aclosed+)

# Function Mesh v0.19.0 Release Notes

## v0.19.0 What's New

* [Controller] Optimize jvm parameters ([#708](https://github.com/streamnative/function-mesh/pull/708))
* [Controller] Avoid NPE error when convert HPA V2 to V2beta2 ([#713](https://github.com/streamnative/function-mesh/pull/713))
* [Controller] Use `bash -c` instead of `sh -c` ([#714](https://github.com/streamnative/function-mesh/pull/714))

For the full changes in this release, see the [Pull Requests](https://github.com/streamnative/function-mesh/pulls?q=is%3Apr+label%3Am%2F2024-01+is%3Aclosed+)

# Function Mesh v0.18.0 Release Notes

## v0.18.0 What's New

* [Controller] Support generic runtime ([#688](https://github.com/streamnative/function-mesh/pull/688))
* [Controller] Set -Xms and -XX:UseG1GC for java runtime ([#698](https://github.com/streamnative/function-mesh/pull/698))
* [Controller] Allow memory requests to be equal as the memory limits ([#692](https://github.com/streamnative/function-mesh/pull/692))

For the full changes in this release, see the [Pull Requests](https://github.com/streamnative/function-mesh/pulls?q=is%3Apr+label%3Am%2F2023-10+is%3Aclosed+)

# Function Mesh v0.17.0 Release Notes

## v0.17.0 What's New

* [Controller] Provides the health and reconcile metrics ([#689](https://github.com/streamnative/function-mesh/pull/689))
* [Controller] Support very long resource name ([#654](https://github.com/streamnative/function-mesh/issues/654))

For the full changes in this release, see the [Pull Requests](https://github.com/streamnative/function-mesh/pulls?q=is%3Apr+label%3Am%2F2023-09+is%3Aclosed+)

# Function Mesh v0.16.0 Release Notes

## v0.16.0 What's New

* [Controller] Decide to whether use pulsarctl based on image name ([#669](https://github.com/streamnative/function-mesh/pull/669))
* [CRD] Remove connector catalog CRD ([#672](https://github.com/streamnative/function-mesh/pull/672))
* [Controller] Support using sidecar to routing pod logs to Pulsar topics ([#673](https://github.com/streamnative/function-mesh/pull/673))
* [Controller] Fix yaml format log config ([#681](https://github.com/streamnative/function-mesh/pull/681))
* [Controller] Use regex to check whether runner image has pulsarctl ([#682](https://github.com/streamnative/function-mesh/pull/682))
* [Controller] Do not create filebeat container when log topic is not set ([#683](https://github.com/streamnative/function-mesh/pull/683))

For the full changes in this release, see the [Pull Requests](https://github.com/streamnative/function-mesh/pulls?q=is%3Apr+label%3Am%2F2023-08+is%3Aclosed+)

# Function Mesh v0.15.0 Release Notes

## v0.15.0 What's New

* [Controller] Add support for autoscaling/v2 ([#658](https://github.com/streamnative/function-mesh/pull/658))
* [Runner Image] Fix python runner image with pulsar-client==3.2.0 ([#660](https://github.com/streamnative/function-mesh/pull/659))
* [CRD][Controller] Support YAML format for log4j ([#667](https://github.com/streamnative/function-mesh/pull/667))
* [CI] Upgrade to autoscaling/v2 for HPA ([#668](https://github.com/streamnative/function-mesh/pull/668))
* [Controller] back support autoscaling/v2beta2 ([#675](https://github.com/streamnative/function-mesh/pull/675))
* [Controller] Fix yaml log config filename ([#676](https://github.com/streamnative/function-mesh/pull/676))
* [Controller] Use correct value for resources in FunctionDetails ([#677](https://github.com/streamnative/function-mesh/pull/677))

For the full changes in this release, see the [Pull Requests](https://github.com/streamnative/function-mesh/pulls?q=is%3Apr+label%3Am%2F2023-07+is%3Aclosed+)

# Function Mesh v0.14.0 Release Notes

## v0.14.0 What's New

* [Runner Image] Support Pulsar 3.0 ([#625](https://github.com/streamnative/function-mesh/pull/625))
* [Helm Charts] Allow the `function-mesh-secrets-webhook` Helm chart to be deployed into non-default Kubernetes namespaces ([#628](https://github.com/streamnative/function-mesh/pull/628))
* [Controller] Support the `CleanupSubscription` configuration ([#622](https://github.com/streamnative/function-mesh/pull/622))
* [CI] Trigger the release action when the tag with the prefix `v` is pushed ([#631](https://github.com/streamnative/function-mesh/pull/631))
* [CRD] Support the `PersistentVolumeClaimRetentionPolicy` configuration ([#633](https://github.com/streamnative/function-mesh/pull/633))
* [Runner Image] Release runner images with `pulsarctl` ([#630](https://github.com/streamnative/function-mesh/pull/630))
* [Controller] Sync to the latest Pulsar Function Proto files ([#637](https://github.com/streamnative/function-mesh/pull/637))
* [Controller] Upgrade to Golang 1.20.4 ([#641](https://github.com/streamnative/function-mesh/pull/641))
* [Controller] Prevent cleaning up resources that are managed by the Function Mesh Worker service ([#640](https://github.com/streamnative/function-mesh/pull/640))
* [Controller] Prevent updating unmodifiable fields of StatefulSet when reconciling ([#639](https://github.com/streamnative/function-mesh/pull/639))
* [CI] Allow release the bundle to operatorhub and openshift catalog via Github Actions ([#642](https://github.com/streamnative/function-mesh/pull/642))

For the full changes in this release, see the [Pull Requests](https://github.com/streamnative/function-mesh/pulls?q=is%3Apr+label%3Am%2F2023-05+is%3Aclosed+).

# Function Mesh v0.13.0 Release Notes

## v0.13.0 What's New

* [Controller] Add GenericAuth to auth config ([#620](https://github.com/streamnative/function-mesh/pull/620))
* [Makefile] rename IMG_DIGEST variable ([#621](https://github.com/streamnative/function-mesh/pull/621))

For the full changes in this release, see the [Pull Requests](https://github.com/streamnative/function-mesh/pulls?q=is%3Apr+label%3Am%2F2023-04+is%3Aclosed+).

# Function Mesh v0.12.0 Release Notes

## v0.12.0 What's New

* [Controller] Ensure the Pod security standard follows restricted specifications ([#574](https://github.com/streamnative/function-mesh/pull/574))
* [Controller] Fix the issue that the `fsGroup` does not work when a volume mount uses an existing directory as a subpath ([#577](https://github.com/streamnative/function-mesh/issues/577))
* [Controller] Add the default value for `minReplicas` and `replicas` ([#582](https://github.com/streamnative/function-mesh/pull/582))
* [Controller] Introduce the `ConnectorCatalog` CRD ([#585](https://github.com/streamnative/function-mesh/pull/585), [#610](https://github.com/streamnative/function-mesh/pull/610))
* [Controller] Allow passing Function Mesh Operator configurations through the environment variables ([#587](https://github.com/streamnative/function-mesh/pull/587))
* [Controller] Allow scaling a function or connector to zero when HPA is disabled ([#590](https://github.com/streamnative/function-mesh/pull/590))
* [Controller] Fix the issue that `SubscriptionType` is not handled properly ([#592](https://github.com/streamnative/function-mesh/pull/592))
* [Controller] Fix the issue that `DeadLetterTopic` is set to an incorrect value in some specific cases ([#595](https://github.com/streamnative/function-mesh/pull/595))
* [Controller] Fix the issue that `MaxPendingAsyncRequests` is not working ([#599](https://github.com/streamnative/function-mesh/pull/599))
* [Controller] Fix the issue that the liveness probe is not working ([#604](https://github.com/streamnative/function-mesh/pull/604))
* [Controller] Support configuring multiple PVs to a CR ([#605](https://github.com/streamnative/function-mesh/pull/605))
* [Controller] Introduce `ShowPreciseParallelism` to CRDs ([#607](https://github.com/streamnative/function-mesh/pull/607))
* [Controller] Make `Limits.CPU` optional ([#608](https://github.com/streamnative/function-mesh/pull/608))
* [Controller] Move `webhook` into `pkg` ([#613](https://github.com/streamnative/function-mesh/pull/613))
* [Runner Image] Remove unused dependencies from runner images ([#576](https://github.com/streamnative/function-mesh/pull/576))
* [CVE] Resolve vulnerabilities and bump dependencies ([#568](https://github.com/streamnative/function-mesh/pull/568), [#583](https://github.com/streamnative/function-mesh/pull/583), [#600](https://github.com/streamnative/function-mesh/pull/600))

For the full changes in this release, see the [Pull Requests](https://github.com/streamnative/function-mesh/pulls?q=is%3Apr+label%3Am%2F2023-03+is%3Aclosed+).

# Function Mesh v0.11.0 Release Notes

## v0.11.0 What's New

* [Runner Images] Copy /pulsar/pulsar-client only when it exists ([#569](https://github.com/streamnative/function-mesh/pull/569))
* [CVE] Resolve vulnerabilities ([#571](https://github.com/streamnative/function-mesh/pull/571))

For the full changes in this release, see the [Pull Requests](https://github.com/streamnative/function-mesh/pulls?q=is%3Apr+label%3Am%2F2023-02+is%3Aclosed+).

# Function Mesh v0.10.0 Release Notes

## v0.10.0 What's New

* [Controller] Use wget for http/https package in downloader container ([#556](https://github.com/streamnative/function-mesh/pull/556))
* [Controller] Add exclusivity to the BuiltinHPARule ([#549](https://github.com/streamnative/function-mesh/pull/549))

For the full changes in this release, see the [Pull Requests](https://github.com/streamnative/function-mesh/pulls?q=is%3Apr+label%3Am%2F2023-01+is%3Aclosed+).

# Function Mesh v0.9.0 Release Notes

## v0.9.0 What's New

* [Controller] Support for Pulsar Function health check ([#538](https://github.com/streamnative/function-mesh/pull/538))
* [Controller] Fix operator vulnerability ([#550](https://github.com/streamnative/function-mesh/pull/550))
* [Controller] Add an overall condition field to Mesh resource ([#546](https://github.com/streamnative/function-mesh/pull/538))
* [Controller] Bump pulsarctl image to 2.10.2.3 ([#548](https://github.com/streamnative/function-mesh/pull/548))
* [Controller] Add a validation webhook for the `spec.pulsar` connection configs ([#534](https://github.com/streamnative/function-mesh/pull/534))
* [Bundle] Address openshift security vulnerabilities ([#536](https://github.com/streamnative/function-mesh/pull/536))
* [Sample] Update the FunctionMesh examples to use the package with the Docker image ([#544](https://github.com/streamnative/function-mesh/pull/544))
* [Sample] Change the example function to make the function works with the documentation ([#542](https://github.com/streamnative/function-mesh/pull/542))
* [Sample] Add a Pulsar IO Connector sample ([#541](https://github.com/streamnative/function-mesh/pull/541))
* [Makefile] change `truncate` to `>` ([#530](https://github.com/streamnative/function-mesh/pull/530))
* [Test] Validate the init container feature with legacy auth and tls config ([#532](https://github.com/streamnative/function-mesh/pull/532))
* [Test] Enable trivy scanner ([#527](https://github.com/streamnative/function-mesh/pull/527))

For the full changes in this release, see the [Pull Requests](https://github.com/streamnative/function-mesh/pulls?q=is%3Apr+label%3Am%2F2022-12+is%3Aclosed).

# Function Mesh v0.8.0 Release Notes

## v0.8.0 What's New

* [Controller] Upgrade the k8s dependency libraries to v1.22.\*. ([#494](https://github.com/streamnative/function-mesh/pull/494))
* [Controller] Add batch source config ([#496](https://github.com/streamnative/function-mesh/pull/496))
* [Controller] Fix python runner oauth2 ([#506](https://github.com/streamnative/function-mesh/pull/506))
* [Controller] Support vpa ([#479](https://github.com/streamnative/function-mesh/pull/479))
* [Controller] Enable stateful configs to connectors ([#510](https://github.com/streamnative/function-mesh/pull/510))
* [Controller] 503: temporarily fix the protobuf conflict in run command ([#512](https://github.com/streamnative/function-mesh/pull/512))
* [Controller] Resolve the namespace conflict of protocol buffers package ([#517](https://github.com/streamnative/function-mesh/pull/517))
* [Controller] Add a flag to enable/disable downloader ([#515](https://github.com/streamnative/function-mesh/pull/515))
* [Controller] Add java opts ([#484](https://github.com/streamnative/function-mesh/pull/484))
* [Controller] Fix legacy authSecret and tlsSecret error ([#523](https://github.com/streamnative/function-mesh/pull/523))
* [CRD] Add finalizers to RBAC ([#485](https://github.com/streamnative/function-mesh/pull/485))
* [CRD] Migrate ClusterRole from v1beta1 to v1 ([#492](https://github.com/streamnative/function-mesh/pull/492))
* [CRD] Upgrade OLM operator capability level to "Full Lifecycle" ([#518](https://github.com/streamnative/function-mesh/pull/518))
* [CRD] Reserve unknown fields for batchSourceConfig ([#513](https://github.com/streamnative/function-mesh/pull/513))
* [Doc] Add readme docs to helm chart ([#497](https://github.com/streamnative/function-mesh/pull/497))
* [Makefile] Adopt openshift olm annotation ([#488](https://github.com/streamnative/function-mesh/pull/488))
* [Test] Fix release bug ([#520](https://github.com/streamnative/function-mesh/pull/520))
* [Test] Change e2e to pull_request_target trigger type ([#519](https://github.com/streamnative/function-mesh/pull/519))

For the full changes in this release, see the [Pull Requests](https://github.com/streamnative/function-mesh/pulls?q=is%3Apr+label%3Am%2F2022-11+is%3Aclosed).

# Function Mesh v0.7.0 Release Notes

## v0.7.0 What's New

* [CVE] Resolve vulnerabilities ([#478](https://github.com/streamnative/function-mesh/pull/478))
* [Test] Remove secret data in CI workflows ([#477](https://github.com/streamnative/function-mesh/pull/477))
* [Test] Migrate e2e tests to skywalking-e2e-infra action ([#476](https://github.com/streamnative/function-mesh/pull/476))
* [Controller] Add client for FunctionMesh ([#473](https://github.com/streamnative/function-mesh/pull/473))
* [Controller] Use init container to download packages and functions ([#411](https://github.com/streamnative/function-mesh/pull/411))
* [Controller] Fix logging window function ([#481](https://github.com/streamnative/function-mesh/pull/481))

For the full changes in this release, see the [Pull Requests](https://github.com/streamnative/function-mesh/pulls?q=is%3Apr+label%3Am%2F2022-10+is%3Aclosed).

# Function Mesh v0.6.0 Release Notes

## v0.6.0 What's New

- [Test] Add integration tests for Pulsar using OAuth2 authentication ([#459](https://github.com/streamnative/function-mesh/pull/459))
- [Controller] Support OAuth2 authentication for Pulsar ([#463](https://github.com/streamnative/function-mesh/pull/463))
- [Controller] Support log rotation ([#417](https://github.com/streamnative/function-mesh/pull/417))
- [Controller] Rename the metrics port name to make Istio compatible ([#466](https://github.com/streamnative/function-mesh/pull/466)
- [Controller] Improve the stability of autoscaling when HPA is enabled ([#450](https://github.com/streamnative/function-mesh/pull/450))
- [Controller] Support Pulsar window function ([#460](https://github.com/streamnative/function-mesh/pull/460))
- [Controller] Improve Function Mesh labels ([#451](https://github.com/streamnative/function-mesh/pull/451))
- [Runner Images] Make runner images use different Java versions based on the Pulsar versions ([#469](https://github.com/streamnative/function-mesh/pull/469))

For the full changes in this release, see the [Pull Requests](https://github.com/streamnative/function-mesh/pulls?q=is%3Apr+label%3Am%2F2022-09+is%3Aclosed).

# Function Mesh v0.5.0 Release Notes

## v0.5.0 What's New

- [Test] Add tests for package download ([#429](https://github.com/streamnative/function-mesh/pull/429))
- [Test] Refine the test cases of the sinks, sources, and Crypto functions ([#435](https://github.com/streamnative/function-mesh/pull/435))
- [Test] Use cert-manager v1.8.2 for tests ([#433](https://github.com/streamnative/function-mesh/pull/433))
- [Helm Charts] Add the handle logic for Function Mesh to clean up orphaned subcomponents ([#431](https://github.com/streamnative/function-mesh/pull/431))
- [Controller] Support setting the logging levels ([#445](https://github.com/streamnative/function-mesh/pull/445))
- [Controller] Update the CSV for OpenShift ([#417](https://github.com/streamnative/function-mesh/pull/417))
- [Controller] Add the sky-walking workflow for the Pulsar cluster enabled with the self-signed TLS certificate ([#439](https://github.com/streamnative/function-mesh/pull/439))
- [Controller] Add a limit to the resource name ([#437](https://github.com/streamnative/function-mesh/pull/437/files))

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
