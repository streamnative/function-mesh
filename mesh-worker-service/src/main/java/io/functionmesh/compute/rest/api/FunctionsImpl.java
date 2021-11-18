/**
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */
package io.functionmesh.compute.rest.api;

import com.google.common.collect.Maps;
import io.functionmesh.compute.MeshWorkerService;
import io.functionmesh.compute.functions.models.V1alpha1Function;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecJava;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecPod;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecPodVolumeMounts;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecPodVolumes;
import io.functionmesh.compute.functions.models.V1alpha1FunctionStatus;
import io.functionmesh.compute.models.MeshWorkerServiceCustomConfig;
import io.functionmesh.compute.util.CommonUtil;
import io.functionmesh.compute.util.FunctionsUtil;
import io.functionmesh.compute.util.KubernetesUtils;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.kubernetes.client.openapi.models.V1ContainerState;
import io.kubernetes.client.openapi.models.V1ContainerStatus;
import io.kubernetes.client.openapi.models.V1Pod;
import io.kubernetes.client.openapi.models.V1PodList;
import io.kubernetes.client.openapi.models.V1PodStatus;
import io.kubernetes.client.openapi.models.V1StatefulSet;
import lombok.extern.slf4j.Slf4j;
import okhttp3.Call;
import org.apache.commons.lang.StringUtils;
import org.apache.pulsar.broker.authentication.AuthenticationDataHttps;
import org.apache.pulsar.broker.authentication.AuthenticationDataSource;
import org.apache.pulsar.common.functions.FunctionConfig;
import org.apache.pulsar.common.functions.UpdateOptionsImpl;
import org.apache.pulsar.common.policies.data.FunctionStatus;
import org.apache.pulsar.common.util.RestException;
import org.apache.pulsar.functions.proto.Function;
import org.apache.pulsar.functions.proto.InstanceCommunication;
import org.apache.pulsar.functions.proto.InstanceControlGrpc;
import org.apache.pulsar.functions.utils.ComponentTypeUtils;
import org.apache.pulsar.functions.worker.service.api.Functions;
import org.glassfish.jersey.media.multipart.FormDataContentDisposition;

import javax.ws.rs.core.Response;
import java.io.InputStream;
import java.net.URI;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.concurrent.CompletableFuture;
import java.util.function.Supplier;
import java.util.stream.Collectors;

@Slf4j
public class FunctionsImpl extends MeshComponentImpl implements Functions<MeshWorkerService> {

    public FunctionsImpl(Supplier<MeshWorkerService> meshWorkerServiceSupplier) {
        super(meshWorkerServiceSupplier, Function.FunctionDetails.ComponentType.FUNCTION);
    }


    private void validateRegisterFunctionRequestParams(String tenant, String namespace, String functionName,
                                                       FunctionConfig functionConfig, boolean jarUploaded) {
        if (tenant == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Tenant is not provided");
        }
        if (namespace == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Namespace is not provided");
        }
        if (functionName == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Function name is not provided");
        }
        if (functionConfig == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Function config is not provided");
        }
        MeshWorkerServiceCustomConfig customConfig = worker().getMeshWorkerServiceCustomConfig();
        if (jarUploaded && customConfig != null && !customConfig.isUploadEnabled()) {
            throw new RestException(Response.Status.BAD_REQUEST, "Uploading Jar File is not enabled");
        }
        this.validateResources(functionConfig.getResources(), worker().getWorkerConfig().getFunctionInstanceMinResources(),
                worker().getWorkerConfig().getFunctionInstanceMaxResources());
    }

    private void validateUpdateFunctionRequestParams(String tenant, String namespace, String functionName,
                                                     FunctionConfig functionConfig, boolean uploadedJar) {
        validateRegisterFunctionRequestParams(tenant, namespace, functionName, functionConfig, uploadedJar);
    }

    private void validateGetFunctionInfoRequestParams(String tenant, String namespace, String functionName) {
        this.validateGetInfoRequestParams(tenant, namespace, functionName, kind);
    }

    private void validateFunctionEnabled() {
        MeshWorkerServiceCustomConfig customConfig = worker().getMeshWorkerServiceCustomConfig();
        if (customConfig != null && !customConfig.isFunctionEnabled()) {
            throw new RestException(Response.Status.BAD_REQUEST, "Function API is disabled");
        }
    }

    @Override
    public void registerFunction(final String tenant,
                                 final String namespace,
                                 final String functionName,
                                 final InputStream uploadedInputStream,
                                 final FormDataContentDisposition fileDetail,
                                 final String functionPkgUrl,
                                 final FunctionConfig functionConfig,
                                 final String clientRole,
                                 AuthenticationDataHttps clientAuthenticationDataHttps) {
        validateFunctionEnabled();

        validateRegisterFunctionRequestParams(tenant, namespace, functionName, functionConfig, uploadedInputStream != null);
        this.validatePermission(tenant,
                namespace,
                clientRole,
                clientAuthenticationDataHttps,
                ComponentTypeUtils.toString(componentType));
        this.validateTenantIsExist(tenant, namespace, functionName, clientRole);
        String packageURL = functionPkgUrl;
        if (uploadedInputStream != null) {
            try {
                String tempDirectory = System.getProperty("java.io.tmpdir");
                packageURL = FunctionsUtil.uploadPackageToPackageService(worker().getBrokerAdmin(), tenant, namespace, functionName, uploadedInputStream, fileDetail, tempDirectory);
            } catch (Exception e) {
                log.error("register {}/{}/{} function failed, error message: {}", tenant, namespace, functionName, e);
                throw new RestException(Response.Status.INTERNAL_SERVER_ERROR, e.getMessage());
            }
        }

        String cluster = worker().getWorkerConfig().getPulsarFunctionsCluster();
        V1alpha1Function v1alpha1Function = FunctionsUtil.createV1alpha1FunctionFromFunctionConfig(
                kind,
                group,
                version,
                functionName,
                packageURL,
                functionConfig,
                cluster,
                worker()
        );
        // override namespace by configuration file
        v1alpha1Function.getMetadata().setNamespace(KubernetesUtils.getNamespace(worker().getFactoryConfig()));
        Map<String, String> customLabels = Maps.newHashMap();
        customLabels.put(CLUSTER_LABEL_CLAIM, v1alpha1Function.getSpec().getClusterName());
        customLabels.put(TENANT_LABEL_CLAIM, tenant);
        customLabels.put(NAMESPACE_LABEL_CLAIM, namespace);
        customLabels.put(COMPONENT_LABEL_CLAIM, functionName);
        V1alpha1FunctionSpecPod pod = new V1alpha1FunctionSpecPod();
        if (worker().getFactoryConfig() != null && worker().getFactoryConfig().getCustomLabels() != null) {
            customLabels.putAll(worker().getFactoryConfig().getCustomLabels());
        }
        pod.setLabels(customLabels);
        v1alpha1Function.getSpec().setPod(pod);
        v1alpha1Function.getMetadata().setLabels(customLabels);
        try {
            this.upsertFunction(tenant, namespace, functionName, functionConfig, v1alpha1Function, clientAuthenticationDataHttps);
            Call call = worker().getCustomObjectsApi().createNamespacedCustomObjectCall(
                    group,
                    version,
                    KubernetesUtils.getNamespace(worker().getFactoryConfig()),
                    plural,
                    v1alpha1Function,
                    null,
                    null,
                    null,
                    null
            );
            executeCall(call, V1alpha1Function.class);
        } catch (RestException restException) {
            log.error(
                    "register {}/{}/{} sink failed, error message: {}",
                    tenant,
                    namespace,
                    functionConfig,
                    restException.getMessage());
            throw restException;
        } catch (Exception e) {
            log.error("register {}/{}/{} function failed, error message: {}", tenant, namespace, functionName, e);
            throw new RestException(Response.Status.INTERNAL_SERVER_ERROR, e.getMessage());
        }
    }

    @Override
    public void updateFunction(final String tenant,
                               final String namespace,
                               final String functionName,
                               final InputStream uploadedInputStream,
                               final FormDataContentDisposition fileDetail,
                               final String functionPkgUrl,
                               final FunctionConfig functionConfig,
                               final String clientRole,
                               AuthenticationDataHttps clientAuthenticationDataHttps,
                               UpdateOptionsImpl updateOptions) {
        validateFunctionEnabled();

        validateUpdateFunctionRequestParams(tenant, namespace, functionName, functionConfig, uploadedInputStream != null);

        try {
            String cluster = worker().getWorkerConfig().getPulsarFunctionsCluster();
            V1alpha1Function v1alpha1Function = FunctionsUtil.createV1alpha1FunctionFromFunctionConfig(
                    kind,
                    group,
                    version,
                    functionName,
                    functionPkgUrl,
                    functionConfig,
                    cluster,
                    worker()
            );
            Call getCall = worker().getCustomObjectsApi().getNamespacedCustomObjectCall(
                    group,
                    version,
                    KubernetesUtils.getNamespace(worker().getFactoryConfig()),
                    plural,
                    v1alpha1Function.getMetadata().getName(),
                    null
            );
            V1alpha1Function oldFn = executeCall(getCall, V1alpha1Function.class);

            v1alpha1Function.getMetadata().setNamespace(KubernetesUtils.getNamespace(worker().getFactoryConfig()));
            Map<String, String> customLabels = oldFn.getMetadata().getLabels();
            V1alpha1FunctionSpecPod pod = new V1alpha1FunctionSpecPod();
            pod.setLabels(customLabels);
            v1alpha1Function.getSpec().setPod(pod);
            v1alpha1Function.getMetadata().setLabels(customLabels);
            v1alpha1Function.getMetadata().setResourceVersion(oldFn.getMetadata().getResourceVersion());
            this.upsertFunction(tenant, namespace, functionName, functionConfig, v1alpha1Function, clientAuthenticationDataHttps);
            Call replaceCall = worker().getCustomObjectsApi().replaceNamespacedCustomObjectCall(
                    group,
                    version,
                    KubernetesUtils.getNamespace(worker().getFactoryConfig()),
                    plural,
                    v1alpha1Function.getMetadata().getName(),
                    v1alpha1Function,
                    null,
                    null,
                    null
            );
            executeCall(replaceCall, V1alpha1Function.class);
        } catch (Exception e) {
            log.error("update {}/{}/{} function failed, error message: {}", tenant, namespace, functionName, e);
            throw new RestException(Response.Status.INTERNAL_SERVER_ERROR, e.getMessage());
        }
    }

    @Override
    public FunctionConfig getFunctionInfo(final String tenant,
                                          final String namespace,
                                          final String componentName,
                                          final String clientRole,
                                          final AuthenticationDataSource clientAuthenticationDataHttps) {
        validateFunctionEnabled();
        validateGetFunctionInfoRequestParams(tenant, namespace, componentName);

        this.validatePermission(tenant,
                namespace,
                clientRole,
                clientAuthenticationDataHttps,
                ComponentTypeUtils.toString(componentType));

        String hashName = CommonUtil.generateObjectName(worker(), tenant, namespace, componentName);
        try {
            Call call = worker().getCustomObjectsApi().getNamespacedCustomObjectCall(
                    group,
                    version,
                    KubernetesUtils.getNamespace(worker().getFactoryConfig()),
                    plural,
                    hashName,
                    null
            );
            V1alpha1Function v1alpha1Function = executeCall(call, V1alpha1Function.class);
            return FunctionsUtil.createFunctionConfigFromV1alpha1Function(tenant, namespace, componentName,
                    v1alpha1Function);
        } catch (Exception e) {
            log.error("get {}/{}/{} function failed, error message: {}", tenant, namespace, componentName, e);
            throw new RestException(Response.Status.INTERNAL_SERVER_ERROR, e.getMessage());
        }
    }

    @Override
    public FunctionStatus.FunctionInstanceStatus.FunctionInstanceStatusData getFunctionInstanceStatus(final String tenant,
                                                                                                      final String namespace,
                                                                                                      final String componentName,
                                                                                                      final String instanceId,
                                                                                                      final URI uri,
                                                                                                      final String clientRole,
                                                                                                      final AuthenticationDataSource clientAuthenticationDataHttps) {

        throw new RestException(Response.Status.BAD_REQUEST, "Unsupported Operation");
    }

    @Override
    public FunctionStatus getFunctionStatus(final String tenant,
                                            final String namespace,
                                            final String componentName,
                                            final URI uri,
                                            final String clientRole,
                                            final AuthenticationDataSource clientAuthenticationDataHttps) {
        validateFunctionEnabled();
        FunctionStatus functionStatus = new FunctionStatus();
        this.validatePermission(tenant,
                namespace,
                clientRole,
                clientAuthenticationDataHttps,
                ComponentTypeUtils.toString(componentType));
        try {
            String hashName = CommonUtil.generateObjectName(worker(), tenant, namespace, componentName);
            String nameSpaceName = KubernetesUtils.getNamespace(worker().getFactoryConfig());
            Call call = worker().getCustomObjectsApi().getNamespacedCustomObjectCall(
                    group, version, nameSpaceName,
                    plural, hashName, null);
            V1alpha1Function v1alpha1Function = executeCall(call, V1alpha1Function.class);
            V1alpha1FunctionStatus v1alpha1FunctionStatus = v1alpha1Function.getStatus();
            if (v1alpha1FunctionStatus == null) {
                log.error(
                        "get status {}/{}/{} function failed, no FunctionStatus exists",
                        tenant,
                        namespace,
                        componentName);
                throw new RestException(Response.Status.NOT_FOUND, "no FunctionStatus exists");
            }
            if (v1alpha1Function.getMetadata() == null) {
                log.error(
                        "get status {}/{}/{} function failed, no Metadata exists",
                        tenant,
                        namespace,
                        componentName);
                throw new RestException(Response.Status.NOT_FOUND, "no Metadata exists");
            }
            String functionLabelSelector = v1alpha1FunctionStatus.getSelector();
            String jobName = CommonUtil.makeJobName(v1alpha1Function.getMetadata().getName(), CommonUtil.COMPONENT_FUNCTION);
            V1StatefulSet v1StatefulSet = worker().getAppsV1Api().readNamespacedStatefulSet(jobName, nameSpaceName, null, null, null);
            String statefulSetName = "";
            String subdomain = "";
            if (v1StatefulSet == null) {
                log.error(
                        "get status {}/{}/{} function failed, no StatefulSet exists",
                        tenant,
                        namespace,
                        componentName);
                throw new RestException(Response.Status.NOT_FOUND, "no StatefulSet exists");
            }
            if (v1StatefulSet.getMetadata() != null &&
                    StringUtils.isNotEmpty(v1StatefulSet.getMetadata().getName())) {
                statefulSetName = v1StatefulSet.getMetadata().getName();
            } else {
                log.error(
                        "get status {}/{}/{} function failed, no statefulSetName exists",
                        tenant,
                        namespace,
                        componentName);
                throw new RestException(Response.Status.NOT_FOUND, "no statefulSetName exists");
            }
            if (v1StatefulSet.getSpec() != null &&
                    StringUtils.isNotEmpty(v1StatefulSet.getSpec().getServiceName())) {
                subdomain = v1StatefulSet.getSpec().getServiceName();
            } else {
                log.error(
                        "get status {}/{}/{} function failed, no ServiceName exists",
                        tenant,
                        namespace,
                        componentName);
                throw new RestException(Response.Status.NOT_FOUND, "no ServiceName exists");
            }
            if (v1StatefulSet.getStatus() != null) {
                Integer replicas = v1StatefulSet.getStatus().getReplicas();
                if (replicas != null) {
                    functionStatus.setNumInstances(replicas);
                    for (int i = 0; i < replicas; i++) {
                        FunctionStatus.FunctionInstanceStatus functionInstanceStatus = new FunctionStatus.FunctionInstanceStatus();
                        FunctionStatus.FunctionInstanceStatus.FunctionInstanceStatusData functionInstanceStatusData = new FunctionStatus.FunctionInstanceStatus.FunctionInstanceStatusData();
                        functionInstanceStatus.setInstanceId(i);
                        functionInstanceStatus.setStatus(functionInstanceStatusData);
                        functionStatus.addInstance(functionInstanceStatus);
                    }
                    if (v1StatefulSet.getStatus().getReadyReplicas() != null) {
                        functionStatus.setNumRunning(v1StatefulSet.getStatus().getReadyReplicas());
                    }
                }
            } else {
                log.error(
                        "no StatefulSet status exists when get status of function {}/{}/{}",
                        tenant,
                        namespace,
                        componentName);
                throw new RestException(Response.Status.NOT_FOUND, "no StatefulSet status exists");
            }
            V1PodList podList = worker().getCoreV1Api().listNamespacedPod(
                    nameSpaceName, null, null, null, null,
                    functionLabelSelector, null, null, null, null,
                    null);
            if (podList != null) {
                List<V1Pod> runningPods = podList.getItems().stream().
                        filter(KubernetesUtils::isPodRunning).collect(Collectors.toList());
                List<V1Pod> pendingPods = podList.getItems().stream().
                        filter(pod -> !KubernetesUtils.isPodRunning(pod)).collect(Collectors.toList());
                final String finalStatefulSetName = statefulSetName;
                if (!runningPods.isEmpty()) {
                    int podsCount = runningPods.size();
                    ManagedChannel[] channel = new ManagedChannel[podsCount];
                    InstanceControlGrpc.InstanceControlFutureStub[] stub =
                            new InstanceControlGrpc.InstanceControlFutureStub[podsCount];
                    final String finalSubdomain = subdomain;
                    Set<CompletableFuture<InstanceCommunication.FunctionStatus>> completableFutureSet = new HashSet<>();
                    runningPods.forEach(pod -> {
                        String podName = KubernetesUtils.getPodName(pod);
                        int shardId = CommonUtil.getShardIdFromPodName(podName);
                        int podIndex = runningPods.indexOf(pod);
                        String address = KubernetesUtils.getServiceUrl(podName, finalSubdomain, nameSpaceName);
                        if (shardId == -1) {
                            log.warn("shardId invalid {}", podName);
                            return;
                        }
                        FunctionStatus.FunctionInstanceStatus functionInstanceStatus = null;
                        for (FunctionStatus.FunctionInstanceStatus ins : functionStatus.getInstances()) {
                            if (ins.getInstanceId() == shardId) {
                                functionInstanceStatus = ins;
                                break;
                            }
                        }
                        if (functionInstanceStatus != null) {
                            FunctionStatus.FunctionInstanceStatus.FunctionInstanceStatusData functionInstanceStatusData = functionInstanceStatus.getStatus();
                            V1PodStatus podStatus = pod.getStatus();
                            if (v1alpha1Function.getSpec() != null && StringUtils.isNotEmpty(v1alpha1Function.getSpec().getClusterName())) {
                                functionInstanceStatusData.setWorkerId(v1alpha1Function.getSpec().getClusterName());
                            }
                            if (podStatus != null) {
                                functionInstanceStatusData.setRunning(KubernetesUtils.isPodRunning(pod));
                                if (podStatus.getContainerStatuses() != null && !podStatus.getContainerStatuses().isEmpty()) {
                                    V1ContainerStatus containerStatus = podStatus.getContainerStatuses().get(0);
                                    functionInstanceStatusData.setNumRestarts(containerStatus.getRestartCount());
                                }
                            }
                            // get status from grpc
                            if (channel[podIndex] == null && stub[podIndex] == null) {
                                channel[podIndex] = ManagedChannelBuilder.forAddress(address, 9093)
                                        .usePlaintext()
                                        .build();
                                stub[podIndex] = InstanceControlGrpc.newFutureStub(channel[podIndex]);
                            }
                            CompletableFuture<InstanceCommunication.FunctionStatus> future = CommonUtil.getFunctionStatusAsync(stub[podIndex]);
                            future.whenComplete((fs, e) -> {
                                if (channel[podIndex] != null) {
                                    log.debug("closing channel {}", podIndex);
                                    channel[podIndex].shutdown();
                                }
                                if (e != null) {
                                    log.error("Get function {}-{} status from grpc failed from namespace {}, error message: {}",
                                            finalStatefulSetName,
                                            shardId,
                                            nameSpaceName,
                                            e.getMessage());
                                    functionInstanceStatusData.setError(e.getMessage());
                                } else if (fs != null) {
                                    FunctionsUtil.convertFunctionStatusToInstanceStatusData(fs, functionInstanceStatusData);
                                }
                            });
                            completableFutureSet.add(future);
                        } else {
                            log.error("Get function {}-{} status failed from namespace {}, cannot find status for shardId {}",
                                    finalStatefulSetName,
                                    shardId,
                                    nameSpaceName,
                                    shardId);
                        }
                    });
                    completableFutureSet.forEach(CompletableFuture::join);
                }
                if (!pendingPods.isEmpty()) {
                    pendingPods.forEach(pod -> {
                        String podName = KubernetesUtils.getPodName(pod);
                        int shardId = CommonUtil.getShardIdFromPodName(podName);
                        if (shardId == -1) {
                            log.warn("shardId invalid {}", podName);
                            return;
                        }
                        FunctionStatus.FunctionInstanceStatus functionInstanceStatus = null;
                        for (FunctionStatus.FunctionInstanceStatus ins : functionStatus.getInstances()) {
                            if (ins.getInstanceId() == shardId) {
                                functionInstanceStatus = ins;
                                break;
                            }
                        }
                        if (functionInstanceStatus != null) {
                            FunctionStatus.FunctionInstanceStatus.FunctionInstanceStatusData functionInstanceStatusData = functionInstanceStatus.getStatus();
                            V1PodStatus podStatus = pod.getStatus();
                            if (podStatus != null) {
                                List<V1ContainerStatus> containerStatuses = podStatus.getContainerStatuses();
                                if (containerStatuses != null && !containerStatuses.isEmpty()) {
                                    V1ContainerStatus containerStatus = null;
                                    for (V1ContainerStatus s : containerStatuses){
                                        if (s.getImage().contains(v1alpha1Function.getSpec().getImage())) {
                                            containerStatus = s;
                                            break;
                                        }
                                    }
                                    if (containerStatus != null) {
                                        V1ContainerState state = containerStatus.getState();
                                        if (state != null && state.getTerminated() != null) {
                                            functionInstanceStatusData.setError(state.getTerminated().getMessage());
                                        } else if (state != null && state.getWaiting() != null) {
                                            functionInstanceStatusData.setError(state.getWaiting().getMessage());
                                        } else {
                                            V1ContainerState lastState = containerStatus.getLastState();
                                            if (lastState != null && lastState.getTerminated() != null) {
                                                functionInstanceStatusData.setError(lastState.getTerminated().getMessage());
                                            } else if (lastState != null && lastState.getWaiting() != null) {
                                                functionInstanceStatusData.setError(lastState.getWaiting().getMessage());
                                            }
                                        }
                                        if (containerStatus.getRestartCount() != null) {
                                            functionInstanceStatusData.setNumRestarts(containerStatus.getRestartCount());
                                        }
                                    } else {
                                        functionInstanceStatusData.setError(podStatus.getPhase());
                                    }
                                }
                            }
                        } else {
                            log.error("Get function {}-{} status failed from namespace {}, cannot find status for shardId {}",
                                finalStatefulSetName,
                                shardId,
                                nameSpaceName,
                                shardId);
                        }
                    });
                }
            }
        } catch (Exception e) {
            log.error("Get function {} status failed from namespace {}, error message: {}",
                    componentName, namespace, e.getMessage());
        }

        return functionStatus;
    }

    @Override
    public void updateFunctionOnWorkerLeader(final String tenant,
                                             final String namespace,
                                             final String functionName,
                                             final InputStream uploadedInputStream,
                                             final boolean delete,
                                             URI uri,
                                             final String clientRole,
                                             final AuthenticationDataSource clientAuthenticationDataHttps) {
        throw new RestException(Response.Status.BAD_REQUEST, "Unsupported Operation");
    }

    private void upsertFunction(final String tenant,
                                final String namespace,
                                final String functionName,
                                final FunctionConfig functionConfig,
                                V1alpha1Function v1alpha1Function,
                                AuthenticationDataHttps clientAuthenticationDataHttps) {
        if (worker().getWorkerConfig().isAuthenticationEnabled()) {
            if (clientAuthenticationDataHttps != null) {
                try {
                    V1alpha1FunctionSpecPod podPolicy = v1alpha1Function.getSpec().getPod();
                    if (podPolicy == null) {
                        podPolicy = new V1alpha1FunctionSpecPod();
                        v1alpha1Function.getSpec().setPod(podPolicy);
                    }
                    MeshWorkerServiceCustomConfig customConfig = worker().getMeshWorkerServiceCustomConfig();
                    List<V1alpha1FunctionSpecPodVolumes> volumesList = customConfig.asV1alpha1FunctionSpecPodVolumesList();
                    if (volumesList != null && !volumesList.isEmpty()) {
                        v1alpha1Function.getSpec().getPod().setVolumes(volumesList);
                    }
                    List<V1alpha1FunctionSpecPodVolumeMounts> volumeMountsList =
                            customConfig.asV1alpha1FunctionSpecPodVolumeMounts();
                    if (volumeMountsList != null && !volumeMountsList.isEmpty()) {
                        v1alpha1Function.getSpec().setVolumeMounts(volumeMountsList);
                    }
                    if (StringUtils.isNotEmpty(customConfig.getExtraDependenciesDir())) {
                        V1alpha1FunctionSpecJava v1alpha1FunctionSpecJava = null;
                        if (v1alpha1Function.getSpec() != null && v1alpha1Function.getSpec().getJava() != null) {
                            v1alpha1FunctionSpecJava = v1alpha1Function.getSpec().getJava();
                        } else if (v1alpha1Function.getSpec() != null && v1alpha1Function.getSpec().getJava() == null &&
                                v1alpha1Function.getSpec().getPython() == null &&
                                v1alpha1Function.getSpec().getGolang() == null){
                            v1alpha1FunctionSpecJava = new V1alpha1FunctionSpecJava();
                        }
                        if (v1alpha1FunctionSpecJava != null) {
                            v1alpha1FunctionSpecJava.setExtraDependenciesDir(customConfig.getExtraDependenciesDir());
                            v1alpha1Function.getSpec().setJava(v1alpha1FunctionSpecJava);
                        }
                    }
                    if (!StringUtils.isEmpty(worker().getWorkerConfig().getBrokerClientAuthenticationPlugin())
                            && !StringUtils.isEmpty(worker().getWorkerConfig().getBrokerClientAuthenticationParameters())) {
                        String authSecretName = KubernetesUtils.upsertSecret(kind.toLowerCase(), "auth",
                                v1alpha1Function.getSpec().getClusterName(), tenant, namespace, functionName,
                                worker().getWorkerConfig(), worker().getCoreV1Api(), worker().getFactoryConfig());
                        v1alpha1Function.getSpec().getPulsar().setAuthSecret(authSecretName);
                    }
                    if (worker().getWorkerConfig().getTlsEnabled()) {
                        String tlsSecretName = KubernetesUtils.upsertSecret(kind.toLowerCase(), "tls",
                                v1alpha1Function.getSpec().getClusterName(), tenant, namespace, functionName,
                                worker().getWorkerConfig(), worker().getCoreV1Api(), worker().getFactoryConfig());
                        v1alpha1Function.getSpec().getPulsar().setTlsSecret(tlsSecretName);
                    }
                    if (!StringUtils.isEmpty(customConfig.getDefaultServiceAccountName())
                            && StringUtils.isEmpty(v1alpha1Function.getSpec().getPod().getServiceAccountName())) {
                        v1alpha1Function.getSpec().getPod().setServiceAccountName(
                                customConfig.getDefaultServiceAccountName());
                    }
                } catch (Exception e) {
                    log.error("Error create or update auth or tls secret for {} {}/{}/{}",
                            ComponentTypeUtils.toString(componentType), tenant, namespace, functionName, e);


                    throw new RestException(Response.Status.INTERNAL_SERVER_ERROR,
                            String.format("Error create or update auth or tls secret for %s %s:- %s",
                                    ComponentTypeUtils.toString(componentType), functionName, e.getMessage()));
                }
            }
        }
    }

}
