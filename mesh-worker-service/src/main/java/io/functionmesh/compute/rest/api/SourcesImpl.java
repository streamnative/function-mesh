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
import io.functionmesh.compute.sources.models.V1alpha1Source;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecJava;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecPod;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecPodVolumeMounts;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecPodVolumes;
import io.functionmesh.compute.sources.models.V1alpha1SourceStatus;
import io.functionmesh.compute.util.CommonUtil;
import io.functionmesh.compute.util.KubernetesUtils;
import io.functionmesh.compute.util.SourcesUtil;
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
import org.apache.commons.lang3.StringUtils;
import org.apache.pulsar.broker.authentication.AuthenticationDataHttps;
import org.apache.pulsar.broker.authentication.AuthenticationDataSource;
import org.apache.pulsar.common.functions.UpdateOptionsImpl;
import org.apache.pulsar.common.io.ConfigFieldDefinition;
import org.apache.pulsar.common.io.ConnectorDefinition;
import org.apache.pulsar.common.io.SourceConfig;
import org.apache.pulsar.common.policies.data.SourceStatus;
import org.apache.pulsar.common.util.RestException;
import org.apache.pulsar.functions.proto.Function;
import org.apache.pulsar.functions.proto.InstanceCommunication;
import org.apache.pulsar.functions.proto.InstanceControlGrpc;
import org.apache.pulsar.functions.utils.ComponentTypeUtils;
import org.apache.pulsar.functions.worker.service.api.Sources;
import org.glassfish.jersey.media.multipart.FormDataContentDisposition;

import javax.ws.rs.core.Response;
import java.io.InputStream;
import java.net.URI;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.concurrent.CompletableFuture;
import java.util.function.Supplier;
import java.util.stream.Collectors;

@Slf4j
public class SourcesImpl extends MeshComponentImpl implements Sources<MeshWorkerService> {
    private final String kind = "Source";

    private final String plural = "sources";

    public SourcesImpl(Supplier<MeshWorkerService> meshWorkerServiceSupplier) {
        super(meshWorkerServiceSupplier, Function.FunctionDetails.ComponentType.SOURCE);
        super.plural = this.plural;
        super.kind = this.kind;
    }

    private void validateSourceEnabled() {
        Map<String, Object> customConfig = worker().getWorkerConfig().getFunctionsWorkerServiceCustomConfigs();
        if (customConfig != null) {
            Boolean sourceEnabled = (Boolean) customConfig.get("sourceEnabled");
            if (sourceEnabled != null && !sourceEnabled) {
                throw new RestException(Response.Status.BAD_REQUEST, "Source API is disabled");
            }
        }
    }

    private void validateRegisterSourceRequestParams(String tenant, String namespace, String sourceName,
                                                     SourceConfig sourceConfig, boolean jarUploaded) {
        if (tenant == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Tenant is not provided");
        }
        if (namespace == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Namespace is not provided");
        }
        if (sourceName == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Source name is not provided");
        }
        if (sourceConfig == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Source config is not provided");
        }
        Map<String, Object> customConfig = worker().getWorkerConfig().getFunctionsWorkerServiceCustomConfigs();
        if (jarUploaded && customConfig != null && customConfig.get("uploadEnabled") != null &&
                !(Boolean) customConfig.get("uploadEnabled")) {
            throw new RestException(Response.Status.BAD_REQUEST, "Uploading Jar File is not enabled");
        }
        this.validateResources(sourceConfig.getResources(), worker().getWorkerConfig().getFunctionInstanceMinResources(),
                worker().getWorkerConfig().getFunctionInstanceMaxResources());
    }

    private void validateUpdateSourceRequestParams(String tenant, String namespace, String sourceName,
                                                   SourceConfig sourceConfig, boolean jarUploaded) {
        this.validateRegisterSourceRequestParams(tenant, namespace, sourceName, sourceConfig, jarUploaded);
    }

    private void validateGetSourceInfoRequestParams(String tenant, String namespace, String sourceName) {
        if (tenant == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Tenant is not provided");
        }
        if (namespace == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Namespace is not provided");
        }
        if (sourceName == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Source name is not provided");
        }
    }

    public void registerSource(final String tenant,
                               final String namespace,
                               final String sourceName,
                               final InputStream uploadedInputStream,
                               final FormDataContentDisposition fileDetail,
                               final String sourcePkgUrl,
                               final SourceConfig sourceConfig,
                               final String clientRole,
                               AuthenticationDataHttps clientAuthenticationDataHttps) {
        validateSourceEnabled();
        validateRegisterSourceRequestParams(tenant, namespace, sourceName, sourceConfig, uploadedInputStream != null);
        this.validatePermission(tenant,
                namespace,
                clientRole,
                clientAuthenticationDataHttps,
                ComponentTypeUtils.toString(componentType));
        this.validateTenantIsExist(tenant, namespace, sourceName, clientRole);
        String cluster = worker().getWorkerConfig().getPulsarFunctionsCluster();
        V1alpha1Source v1alpha1Source = SourcesUtil
                .createV1alpha1SourceFromSourceConfig(
                        kind,
                        group,
                        version,
                        sourceName,
                        sourcePkgUrl,
                        uploadedInputStream,
                        sourceConfig,
                        this.meshWorkerServiceSupplier.get().getConnectorsManager(),
                        worker().getWorkerConfig().getFunctionsWorkerServiceCustomConfigs(), cluster);
        try {

            // override namesapce by configuration
            v1alpha1Source.getMetadata().setNamespace(KubernetesUtils.getNamespace(worker().getFactoryConfig()));
            Map<String, String> customLabels = Maps.newHashMap();
            customLabels.put(CLUSTER_LABEL_CLAIM, v1alpha1Source.getSpec().getClusterName());
            customLabels.put(TENANT_LABEL_CLAIM, tenant);
            customLabels.put(NAMESPACE_LABEL_CLAIM, namespace);
            customLabels.put(COMPONENT_LABEL_CLAIM, sourceName);
            V1alpha1SourceSpecPod pod = new V1alpha1SourceSpecPod();
            if (worker().getFactoryConfig() != null && worker().getFactoryConfig().getCustomLabels() != null) {
                customLabels.putAll(worker().getFactoryConfig().getCustomLabels());
            }
            pod.setLabels(customLabels);
            v1alpha1Source.getMetadata().setLabels(customLabels);
            v1alpha1Source.getSpec().setPod(pod);
            this.upsertSource(tenant, namespace, sourceName, sourceConfig, v1alpha1Source, clientAuthenticationDataHttps);
            Call call = worker().getCustomObjectsApi().createNamespacedCustomObjectCall(
                    group, version, KubernetesUtils.getNamespace(worker().getFactoryConfig()),
                    plural,
                    v1alpha1Source,
                    null,
                    null,
                    null,
                    null);
            executeCall(call, V1alpha1Source.class);
        } catch (RestException restException) {
            log.error(
                    "register {}/{}/{} source failed, error message: {}",
                    tenant,
                    namespace,
                    sourceConfig,
                    restException.getMessage());
            throw restException;
        } catch (Exception e) {
            log.error("register {}/{}/{} source failed, error message: {}", tenant, namespace, sourceConfig, e);
            throw new RestException(Response.Status.INTERNAL_SERVER_ERROR, e.getMessage());
        }
    }

    public void updateSource(final String tenant,
                             final String namespace,
                             final String sourceName,
                             final InputStream uploadedInputStream,
                             final FormDataContentDisposition fileDetail,
                             final String sourcePkgUrl,
                             final SourceConfig sourceConfig,
                             final String clientRole,
                             AuthenticationDataHttps clientAuthenticationDataHttps,
                             UpdateOptionsImpl updateOptions) {
        validateSourceEnabled();
        validateUpdateSourceRequestParams(tenant, namespace, sourceName, sourceConfig, uploadedInputStream != null);
        this.validatePermission(tenant,
                namespace,
                clientRole,
                clientAuthenticationDataHttps,
                ComponentTypeUtils.toString(componentType));
        try {
            String cluster = worker().getWorkerConfig().getPulsarFunctionsCluster();
            V1alpha1Source v1alpha1Source = SourcesUtil
                    .createV1alpha1SourceFromSourceConfig(
                            kind,
                            group,
                            version,
                            sourceName,
                            sourcePkgUrl,
                            uploadedInputStream,
                            sourceConfig,
                            this.meshWorkerServiceSupplier.get().getConnectorsManager(),
                            worker().getWorkerConfig().getFunctionsWorkerServiceCustomConfigs(), cluster);
            Call getCall = worker().getCustomObjectsApi().getNamespacedCustomObjectCall(
                    group,
                    version,
                    KubernetesUtils.getNamespace(worker().getFactoryConfig()),
                    plural,
                    v1alpha1Source.getMetadata().getName(),
                    null
            );
            V1alpha1Source oldRes = executeCall(getCall, V1alpha1Source.class);

            V1alpha1SourceSpecPod pod = new V1alpha1SourceSpecPod();
            Map<String, String> labels = oldRes.getMetadata().getLabels();
            pod.setLabels(labels);
            v1alpha1Source.getMetadata().setLabels(labels);
            v1alpha1Source.getSpec().setPod(pod);
            v1alpha1Source.getMetadata().setNamespace(KubernetesUtils.getNamespace(worker().getFactoryConfig()));
            v1alpha1Source.getMetadata().setResourceVersion(oldRes.getMetadata().getResourceVersion());
            this.upsertSource(tenant, namespace, sourceName, sourceConfig, v1alpha1Source, clientAuthenticationDataHttps);
            Call replaceCall = worker().getCustomObjectsApi().replaceNamespacedCustomObjectCall(
                    group,
                    version,
                    KubernetesUtils.getNamespace(worker().getFactoryConfig()),
                    plural,
                    v1alpha1Source.getMetadata().getName(),
                    v1alpha1Source,
                    null,
                    null,
                    null
            );
            executeCall(replaceCall, Object.class);
        } catch (Exception e) {
            log.error("update {}/{}/{} source failed, error message: {}", tenant, namespace, sourceConfig, e);
            throw new RestException(Response.Status.INTERNAL_SERVER_ERROR, e.getMessage());
        }

    }


    public SourceStatus getSourceStatus(final String tenant,
                                        final String namespace,
                                        final String componentName,
                                        final URI uri,
                                        final String clientRole,
                                        final AuthenticationDataSource clientAuthenticationDataHttps) {
        validateSourceEnabled();
        SourceStatus sourceStatus = new SourceStatus();
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
            V1alpha1Source v1alpha1Source = executeCall(call, V1alpha1Source.class);
            V1alpha1SourceStatus v1alpha1SourceStatus = v1alpha1Source.getStatus();
            if (v1alpha1SourceStatus == null) {
                log.error(
                        "get status {}/{}/{} source failed, no SourceStatus exists",
                        tenant,
                        namespace,
                        componentName);
                throw new RestException(Response.Status.NOT_FOUND, "no SourceStatus exists");
            }
            if (v1alpha1Source.getMetadata() == null) {
                log.error(
                        "get status {}/{}/{} source failed, no Metadata exists",
                        tenant,
                        namespace,
                        componentName);
                throw new RestException(Response.Status.NOT_FOUND, "no Metadata exists");
            }
            String sourceLabelSelector = v1alpha1SourceStatus.getSelector();
            String jobName = CommonUtil.makeJobName(v1alpha1Source.getMetadata().getName(), CommonUtil.COMPONENT_SOURCE);
            V1StatefulSet v1StatefulSet = worker().getAppsV1Api().readNamespacedStatefulSet(jobName, nameSpaceName, null, null, null);
            String statefulSetName = "";
            String subdomain = "";
            if (v1StatefulSet == null) {
                log.error(
                        "get status {}/{}/{} source failed, no StatefulSet exists",
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
                        "get status {}/{}/{} source failed, no statefulSetName exists",
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
                        "get status {}/{}/{} source failed, no ServiceName exists",
                        tenant,
                        namespace,
                        componentName);
                throw new RestException(Response.Status.NOT_FOUND, "no ServiceName exists");
            }
            if (v1StatefulSet.getStatus() != null) {
                Integer replicas = v1StatefulSet.getStatus().getReplicas();
                if (replicas != null) {
                    sourceStatus.setNumInstances(replicas);
                    for (int i = 0; i < replicas; i++) {
                        SourceStatus.SourceInstanceStatus sourceInstanceStatus = new SourceStatus.SourceInstanceStatus();
                        SourceStatus.SourceInstanceStatus.SourceInstanceStatusData sourceInstanceStatusData = new SourceStatus.SourceInstanceStatus.SourceInstanceStatusData();
                        sourceInstanceStatus.setInstanceId(i);
                        sourceInstanceStatus.setStatus(sourceInstanceStatusData);
                        sourceStatus.addInstance(sourceInstanceStatus);
                    }
                    if (v1StatefulSet.getStatus().getReadyReplicas() != null) {
                        sourceStatus.setNumRunning(v1StatefulSet.getStatus().getReadyReplicas());
                    }
                }
            } else {
                log.error(
                        "no StatefulSet status exists when get status of source {}/{}/{}",
                        tenant,
                        namespace,
                        componentName);
                throw new RestException(Response.Status.NOT_FOUND, "no StatefulSet status exists");
            }
            V1PodList podList = worker().getCoreV1Api().listNamespacedPod(
                    nameSpaceName, null, null, null, null,
                    sourceLabelSelector, null, null, null, null,
                    null);
            if (podList != null) {
                List<V1Pod> runningPods = podList.getItems().stream().
                        filter(KubernetesUtils::isPodRunning).collect(Collectors.toList());
                List<V1Pod> pendingPods = podList.getItems().stream().
                        filter(pod -> !KubernetesUtils.isPodRunning(pod)).collect(Collectors.toList());
                String finalStatefulSetName = statefulSetName;
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
                        SourceStatus.SourceInstanceStatus sourceInstanceStatus = null;
                        for (SourceStatus.SourceInstanceStatus ins : sourceStatus.getInstances()) {
                            if (ins.getInstanceId() == shardId) {
                                sourceInstanceStatus = ins;
                                break;
                            }
                        }
                        if (sourceInstanceStatus != null) {
                            SourceStatus.SourceInstanceStatus.SourceInstanceStatusData sourceInstanceStatusData = sourceInstanceStatus.getStatus();
                            V1PodStatus podStatus = pod.getStatus();
                            if (v1alpha1Source.getSpec() != null && StringUtils.isNotEmpty(v1alpha1Source.getSpec().getClusterName())) {
                                sourceInstanceStatusData.setWorkerId(v1alpha1Source.getSpec().getClusterName());
                            }
                            if (podStatus != null) {
                                sourceInstanceStatusData.setRunning(KubernetesUtils.isPodRunning(pod));
                                if (podStatus.getContainerStatuses() != null && !podStatus.getContainerStatuses().isEmpty()) {
                                    V1ContainerStatus containerStatus = podStatus.getContainerStatuses().get(0);
                                    sourceInstanceStatusData.setNumRestarts(containerStatus.getRestartCount());
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
                                    log.error("Get source {}-{} status from grpc failed from namespace {}, error message: {}",
                                            finalStatefulSetName,
                                            shardId,
                                            nameSpaceName,
                                            e.getMessage());
                                    sourceInstanceStatusData.setError(e.getMessage());
                                } else if (fs != null) {
                                    SourcesUtil.convertFunctionStatusToInstanceStatusData(fs, sourceInstanceStatusData);
                                }
                            });
                            completableFutureSet.add(future);
                        } else {
                            log.error("Get source {}-{} status failed from namespace {}, cannot find status for shardId {}",
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
                        SourceStatus.SourceInstanceStatus sourceInstanceStatus = null;
                        for (SourceStatus.SourceInstanceStatus ins : sourceStatus.getInstances()) {
                            if (ins.getInstanceId() == shardId) {
                                sourceInstanceStatus = ins;
                                break;
                            }
                        }
                        if (sourceInstanceStatus != null) {
                            SourceStatus.SourceInstanceStatus.SourceInstanceStatusData sourceInstanceStatusData = sourceInstanceStatus.getStatus();
                            V1PodStatus podStatus = pod.getStatus();
                            if (podStatus != null) {
                                List<V1ContainerStatus> containerStatuses = podStatus.getContainerStatuses();
                                if (containerStatuses != null && !containerStatuses.isEmpty()) {
                                    V1ContainerStatus containerStatus = null;
                                    for (V1ContainerStatus s : containerStatuses){
                                        // TODO need more precise way to find the status
                                        if (s.getImage().contains(v1alpha1Source.getSpec().getImage())) {
                                            containerStatus = s;
                                            break;
                                        }
                                    }
                                    if (containerStatus != null) {
                                        V1ContainerState state = containerStatus.getState();
                                        if (state != null && state.getTerminated() != null) {
                                            sourceInstanceStatusData.setError(state.getTerminated().getMessage());
                                        } else if (state != null && state.getWaiting() != null) {
                                            sourceInstanceStatusData.setError(state.getWaiting().getMessage());
                                        } else {
                                            V1ContainerState lastState = containerStatus.getLastState();
                                            if (lastState != null && lastState.getTerminated() != null) {
                                                sourceInstanceStatusData.setError(lastState.getTerminated().getMessage());
                                            } else if (lastState != null && lastState.getWaiting() != null) {
                                                sourceInstanceStatusData.setError(lastState.getWaiting().getMessage());
                                            }
                                        }
                                        if (containerStatus.getRestartCount() != null) {
                                            sourceInstanceStatusData.setNumRestarts(containerStatus.getRestartCount());
                                        }
                                    } else {
                                        sourceInstanceStatusData.setError(podStatus.getPhase());
                                    }
                                }
                            }
                        } else {
                            log.error("Get source {}-{} status failed from namespace {}, cannot find status for shardId {}",
                                finalStatefulSetName,
                                shardId,
                                nameSpaceName,
                                shardId);
                        }
                    });
                }
            }
        } catch (Exception e) {
            log.error("Get source {} status failed from namespace {}, error message: {}",
                    componentName, namespace, e.getMessage());
        }

        return sourceStatus;
    }

    public SourceStatus.SourceInstanceStatus.SourceInstanceStatusData getSourceInstanceStatus(final String tenant,
                                                                                              final String namespace,
                                                                                              final String sourceName,
                                                                                              final String instanceId,
                                                                                              final URI uri,
                                                                                              final String clientRole,
                                                                                              final AuthenticationDataSource clientAuthenticationDataHttps) {
        validateSourceEnabled();
        SourceStatus.SourceInstanceStatus.SourceInstanceStatusData sourceInstanceStatusData =
                new SourceStatus.SourceInstanceStatus.SourceInstanceStatusData();
        return sourceInstanceStatusData;
    }

    public SourceConfig getSourceInfo(final String tenant,
                                      final String namespace,
                                      final String componentName) {
        validateSourceEnabled();
        this.validateGetInfoRequestParams(tenant, namespace, componentName, kind);
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

            V1alpha1Source v1alpha1Source = executeCall(call, V1alpha1Source.class);
            return SourcesUtil.createSourceConfigFromV1alpha1Source(tenant, namespace, componentName, v1alpha1Source);
        } catch (Exception e) {
            log.error("Get source info {}/{}/{} {} failed, error message: {}", tenant, namespace, componentName, plural, e);
            throw new RestException(Response.Status.INTERNAL_SERVER_ERROR, e.getMessage());
        }
    }

    @Override
    public List<ConnectorDefinition> getSourceList() {
        validateSourceEnabled();
        List<ConnectorDefinition> connectorDefinitions = getListOfConnectors();
        List<ConnectorDefinition> retval = new ArrayList<>();
        for (ConnectorDefinition connectorDefinition : connectorDefinitions) {
            if (!org.apache.commons.lang.StringUtils.isEmpty(connectorDefinition.getSourceClass())) {
                retval.add(connectorDefinition);
            }
        }
        return retval;
    }

    public List<ConfigFieldDefinition> getSourceConfigDefinition(String name) {
        validateSourceEnabled();
        return new ArrayList<>();
    }

    private void upsertSource(final String tenant,
                              final String namespace,
                              final String sourceName,
                              SourceConfig sourceConfig,
                              V1alpha1Source v1alpha1Source,
                              AuthenticationDataHttps clientAuthenticationDataHttps) {
        if (worker().getWorkerConfig().isAuthenticationEnabled()) {
            if (clientAuthenticationDataHttps != null) {
                try {
                    Map<String, Object> functionsWorkerServiceCustomConfigs = worker()
                            .getWorkerConfig().getFunctionsWorkerServiceCustomConfigs();
                    Object volumes = functionsWorkerServiceCustomConfigs.get("volumes");
                    if (volumes != null) {
                        List<V1alpha1SourceSpecPodVolumes> volumesList = (List<V1alpha1SourceSpecPodVolumes>) volumes;
                        v1alpha1Source.getSpec().getPod().setVolumes(volumesList);
                    }
                    if (functionsWorkerServiceCustomConfigs.get("extraDependenciesDir") != null) {
                        V1alpha1SourceSpecJava v1alpha1SourceSpecJava;
                        if (v1alpha1Source.getSpec() != null && v1alpha1Source.getSpec().getJava() != null) {
                            v1alpha1SourceSpecJava = v1alpha1Source.getSpec().getJava();
                        } else {
                            v1alpha1SourceSpecJava = new V1alpha1SourceSpecJava();
                        }
                        v1alpha1SourceSpecJava.setExtraDependenciesDir(
                                (String) functionsWorkerServiceCustomConfigs.get("extraDependenciesDir"));
                        v1alpha1Source.getSpec().setJava(v1alpha1SourceSpecJava);
                    }
                    Object volumeMounts = functionsWorkerServiceCustomConfigs.get("volumeMounts");
                    if (volumeMounts != null) {
                        List<V1alpha1SourceSpecPodVolumeMounts> volumeMountsList = (List<V1alpha1SourceSpecPodVolumeMounts>) volumeMounts;
                        v1alpha1Source.getSpec().setVolumeMounts(volumeMountsList);
                    }
                    if (!StringUtils.isEmpty(worker().getWorkerConfig().getBrokerClientAuthenticationPlugin())
                            && !StringUtils.isEmpty(worker().getWorkerConfig().getBrokerClientAuthenticationParameters())) {
                        String authSecretName = KubernetesUtils.upsertSecret(kind.toLowerCase(), "auth",
                                v1alpha1Source.getSpec().getClusterName(), tenant, namespace, sourceName,
                                worker().getWorkerConfig(), worker().getCoreV1Api(), worker().getFactoryConfig());
                        v1alpha1Source.getSpec().getPulsar().setAuthSecret(authSecretName);
                    }
                    if (worker().getWorkerConfig().getTlsEnabled()) {
                        String tlsSecretName = KubernetesUtils.upsertSecret(kind.toLowerCase(), "tls",
                                v1alpha1Source.getSpec().getClusterName(), tenant, namespace, sourceName,
                                worker().getWorkerConfig(), worker().getCoreV1Api(), worker().getFactoryConfig());
                        v1alpha1Source.getSpec().getPulsar().setTlsSecret(tlsSecretName);
                    }
                } catch (Exception e) {
                    log.error("Error create or update auth or tls secret for {} {}/{}/{}",
                            ComponentTypeUtils.toString(componentType), tenant, namespace, sourceName, e);


                    throw new RestException(Response.Status.INTERNAL_SERVER_ERROR,
                            String.format("Error create or update auth or tls secret %s %s:- %s",
                                    ComponentTypeUtils.toString(componentType), sourceName, e.getMessage()));
                }
            }
        }
    }
}
