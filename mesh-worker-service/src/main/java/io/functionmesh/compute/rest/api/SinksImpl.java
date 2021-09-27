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
import io.functionmesh.compute.models.MeshWorkerServiceCustomConfig;
import io.functionmesh.compute.sinks.models.V1alpha1Sink;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecJava;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecPod;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecPodVolumeMounts;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecPodVolumes;
import io.functionmesh.compute.util.CommonUtil;
import io.functionmesh.compute.util.KubernetesUtils;
import io.functionmesh.compute.util.SinksUtil;
import io.functionmesh.compute.MeshWorkerService;
import io.functionmesh.compute.sinks.models.V1alpha1SinkStatus;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.kubernetes.client.openapi.apis.CustomObjectsApi;
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
import org.apache.pulsar.common.functions.Resources;
import org.apache.pulsar.common.functions.UpdateOptionsImpl;
import org.apache.pulsar.common.io.ConfigFieldDefinition;
import org.apache.pulsar.common.io.ConnectorDefinition;
import org.apache.pulsar.common.io.SinkConfig;
import org.apache.pulsar.common.policies.data.SinkStatus;
import org.apache.pulsar.common.util.RestException;
import org.apache.pulsar.functions.proto.Function;
import org.apache.pulsar.functions.proto.InstanceCommunication;
import org.apache.pulsar.functions.proto.InstanceControlGrpc;
import org.apache.pulsar.functions.utils.ComponentTypeUtils;
import org.apache.pulsar.functions.worker.service.api.Sinks;
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
public class SinksImpl extends MeshComponentImpl
        implements Sinks<MeshWorkerService> {
    private String kind = "Sink";

    private String plural = "sinks";

    public SinksImpl(Supplier<MeshWorkerService> meshWorkerServiceSupplier) {
        super(meshWorkerServiceSupplier, Function.FunctionDetails.ComponentType.SINK);
        super.plural = this.plural;
        super.kind = this.kind;
    }

    private void validateSinkEnabled() {
        MeshWorkerServiceCustomConfig customConfig = worker().getMeshWorkerServiceCustomConfig();
        if (customConfig != null && !customConfig.isSinkEnabled()) {
            throw new RestException(Response.Status.BAD_REQUEST, "Sink API is disabled");
        }
    }

    private void validateRegisterSinkRequestParams(
            String tenant, String namespace, String sinkName, SinkConfig sinkConfig, boolean jarUploaded) {
        if (tenant == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Tenant is not provided");
        }
        if (namespace == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Namespace is not provided");
        }
        if (sinkName == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Sink name is not provided");
        }
        if (sinkConfig == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Sink config is not provided");
        }
        MeshWorkerServiceCustomConfig customConfig = worker().getMeshWorkerServiceCustomConfig();
        if (jarUploaded &&  customConfig != null && !customConfig.isUploadEnabled()) {
            throw new RestException(Response.Status.BAD_REQUEST, "Uploading Jar File is not enabled");
        }
        Resources sinkResources = sinkConfig.getResources();
        this.validateResources(sinkConfig.getResources(), worker().getWorkerConfig().getFunctionInstanceMinResources(),
                worker().getWorkerConfig().getFunctionInstanceMaxResources());
    }

    private void validateUpdateSinkRequestParams(
            String tenant, String namespace, String sinkName, SinkConfig sinkConfig, boolean uploadedJar) {
        this.validateRegisterSinkRequestParams(tenant, namespace, sinkName, sinkConfig, uploadedJar);
    }

    @Override
    public void registerSink(
            final String tenant,
            final String namespace,
            final String sinkName,
            final InputStream uploadedInputStream,
            final FormDataContentDisposition fileDetail,
            final String sinkPkgUrl,
            final SinkConfig sinkConfig,
            final String clientRole,
            AuthenticationDataHttps clientAuthenticationDataHttps) {
        validateSinkEnabled();
        validateRegisterSinkRequestParams(tenant, namespace, sinkName, sinkConfig, uploadedInputStream != null);
        this.validatePermission(tenant,
                namespace,
                clientRole,
                clientAuthenticationDataHttps,
                ComponentTypeUtils.toString(componentType));
        this.validateTenantIsExist(tenant, namespace, sinkName, clientRole);
        String cluster = worker().getWorkerConfig().getPulsarFunctionsCluster();
        V1alpha1Sink v1alpha1Sink =
                SinksUtil.createV1alpha1SkinFromSinkConfig(
                        kind,
                        group,
                        version,
                        sinkName,
                        sinkPkgUrl,
                        uploadedInputStream,
                        sinkConfig,
                        this.meshWorkerServiceSupplier.get().getConnectorsManager(),
                        cluster, worker());
        // override namesapce by configuration
        v1alpha1Sink.getMetadata().setNamespace(KubernetesUtils.getNamespace(worker().getFactoryConfig()));
        try {
            Map<String, String> customLabels = Maps.newHashMap();
            customLabels.put(CLUSTER_LABEL_CLAIM, v1alpha1Sink.getSpec().getClusterName());
            customLabels.put(TENANT_LABEL_CLAIM, tenant);
            customLabels.put(NAMESPACE_LABEL_CLAIM, namespace);
            customLabels.put(COMPONENT_LABEL_CLAIM, sinkName);
            V1alpha1SinkSpecPod pod = new V1alpha1SinkSpecPod();
            if (worker().getFactoryConfig() != null && worker().getFactoryConfig().getCustomLabels() != null) {
                customLabels.putAll(worker().getFactoryConfig().getCustomLabels());
            }
            pod.setLabels(customLabels);
            v1alpha1Sink.getMetadata().setLabels(customLabels);
            v1alpha1Sink.getSpec().setPod(pod);
            this.upsertSink(tenant, namespace, sinkName, sinkConfig, v1alpha1Sink, clientAuthenticationDataHttps);
            Call call =
                    worker().getCustomObjectsApi()
                            .createNamespacedCustomObjectCall(
                                    group,
                                    version,
                                    KubernetesUtils.getNamespace(worker().getFactoryConfig()),
                                    plural,
                                    v1alpha1Sink,
                                    null,
                                    null,
                                    null,
                                    null);
           executeCall(call, V1alpha1Sink.class);
        } catch (RestException restException){
            log.error(
                    "register {}/{}/{} sink failed, error message: {}",
                    tenant,
                    namespace,
                    sinkConfig,
                    restException.getMessage());
            throw restException;
        } catch (Exception e) {
            log.error(
                    "register {}/{}/{} sink failed, error message: {}",
                    tenant,
                    namespace,
                    sinkConfig,
                    e);
            throw new RestException(Response.Status.INTERNAL_SERVER_ERROR, e.getMessage());
        }
    }

    @Override
    public void updateSink(
            final String tenant,
            final String namespace,
            final String sinkName,
            final InputStream uploadedInputStream,
            final FormDataContentDisposition fileDetail,
            final String sinkPkgUrl,
            final SinkConfig sinkConfig,
            final String clientRole,
            AuthenticationDataHttps clientAuthenticationDataHttps,
            UpdateOptionsImpl updateOptions) {
        validateSinkEnabled();
        validateUpdateSinkRequestParams(tenant, namespace, sinkName, sinkConfig, uploadedInputStream != null);
        this.validatePermission(tenant,
                namespace,
                clientRole,
                clientAuthenticationDataHttps,
                ComponentTypeUtils.toString(componentType));
        String cluster = worker().getWorkerConfig().getPulsarFunctionsCluster();
        try {
            V1alpha1Sink v1alpha1Sink =
                    SinksUtil.createV1alpha1SkinFromSinkConfig(
                            kind,
                            group,
                            version,
                            sinkName,
                            sinkPkgUrl,
                            uploadedInputStream,
                            sinkConfig, this.meshWorkerServiceSupplier.get().getConnectorsManager(),
                            cluster, worker());
            CustomObjectsApi customObjectsApi = worker().getCustomObjectsApi();
            Call getCall =
                    customObjectsApi.getNamespacedCustomObjectCall(
                                    group, version,
                                    KubernetesUtils.getNamespace(worker().getFactoryConfig()), plural,
                                    v1alpha1Sink.getMetadata().getName(), null);
            V1alpha1Sink oldRes = executeCall(getCall, V1alpha1Sink.class);
            if (oldRes.getMetadata() == null || oldRes.getMetadata().getLabels() == null) {
                throw new RestException(Response.Status.NOT_FOUND, "This sink resource was not found");
            }
            Map<String, String> labels = oldRes.getMetadata().getLabels();
            V1alpha1SinkSpecPod pod = new V1alpha1SinkSpecPod();
            pod.setLabels(labels);
            v1alpha1Sink.getMetadata().setLabels(labels);
            v1alpha1Sink.getSpec().setPod(pod);
            this.upsertSink(tenant, namespace, sinkName, sinkConfig, v1alpha1Sink, clientAuthenticationDataHttps);
            v1alpha1Sink.getMetadata().setNamespace(KubernetesUtils.getNamespace(worker().getFactoryConfig()));
            v1alpha1Sink
                    .getMetadata()
                    .setResourceVersion(oldRes.getMetadata().getResourceVersion());
            Call replaceCall = customObjectsApi.replaceNamespacedCustomObjectCall(
                                    group,
                                    version,
                                    KubernetesUtils.getNamespace(worker().getFactoryConfig()),
                                    plural,
                                    v1alpha1Sink.getMetadata().getName(),
                                    v1alpha1Sink,
                                    null,
                                    null,
                                    null);
            executeCall(replaceCall, Object.class);
        } catch (Exception e) {
            log.error(
                    "update {}/{}/{} sink failed, error message: {}",
                    tenant,
                    namespace,
                    sinkConfig,
                    e);
            throw new RestException(Response.Status.INTERNAL_SERVER_ERROR, e.getMessage());
        }
    }

    @Override
    public SinkStatus.SinkInstanceStatus.SinkInstanceStatusData getSinkInstanceStatus(
            final String tenant,
            final String namespace,
            final String sinkName,
            final String instanceId,
            final URI uri,
            final String clientRole,
            final AuthenticationDataSource clientAuthenticationDataHttps) {
        validateSinkEnabled();
        SinkStatus.SinkInstanceStatus.SinkInstanceStatusData sinkInstanceStatusData =
                new SinkStatus.SinkInstanceStatus.SinkInstanceStatusData();
        return sinkInstanceStatusData;
    }

    @Override
    public SinkStatus getSinkStatus(
            final String tenant,
            final String namespace,
            final String componentName,
            final URI uri,
            final String clientRole,
            final AuthenticationDataSource clientAuthenticationDataHttps) {
        validateSinkEnabled();
        SinkStatus sinkStatus = new SinkStatus();
        this.validatePermission(tenant,
                namespace,
                clientRole,
                clientAuthenticationDataHttps,
                ComponentTypeUtils.toString(componentType));
        try {
            String hashName = CommonUtil.generateObjectName(worker(), tenant, namespace, componentName);
            String nameSpaceName = KubernetesUtils.getNamespace(worker().getFactoryConfig());
            Call call =
                    worker().getCustomObjectsApi()
                            .getNamespacedCustomObjectCall(
                                    group, version, nameSpaceName,
                                    plural, hashName, null);
            V1alpha1Sink v1alpha1Sink = executeCall(call, V1alpha1Sink.class);
            V1alpha1SinkStatus v1alpha1SinkStatus = v1alpha1Sink.getStatus();
            if (v1alpha1SinkStatus == null) {
                log.error(
                        "get status {}/{}/{} sink failed, no SinkStatus exists",
                        tenant,
                        namespace,
                        componentName);
                throw new RestException(Response.Status.NOT_FOUND, "no SinkStatus exists");
            }
            if (v1alpha1Sink.getMetadata() == null) {
                log.error(
                        "get status {}/{}/{} sink failed, no Metadata exists",
                        tenant,
                        namespace,
                        componentName);
                throw new RestException(Response.Status.NOT_FOUND, "no Metadata exists");
            }
            String sinkLabelSelector = v1alpha1SinkStatus.getSelector();
            String jobName = CommonUtil.makeJobName(v1alpha1Sink.getMetadata().getName(), CommonUtil.COMPONENT_SINK);
            V1StatefulSet v1StatefulSet = worker().getAppsV1Api().readNamespacedStatefulSet(jobName, nameSpaceName, null, null, null);
            String statefulSetName = "";
            String subdomain = "";
            if (v1StatefulSet == null) {
                log.error(
                        "get status {}/{}/{} sink failed, no StatefulSet exists",
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
                        "get status {}/{}/{} sink failed, no statefulSetName exists",
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
                        "get status {}/{}/{} sink failed, no ServiceName exists",
                        tenant,
                        namespace,
                        componentName);
                throw new RestException(Response.Status.NOT_FOUND, "no ServiceName exists");
            }
            if (v1StatefulSet.getStatus() != null) {
                Integer replicas = v1StatefulSet.getStatus().getReplicas();
                if (replicas != null) {
                    sinkStatus.setNumInstances(replicas);
                    for(int i=0; i<replicas; i++) {
                        SinkStatus.SinkInstanceStatus sinkInstanceStatus = new SinkStatus.SinkInstanceStatus();
                        SinkStatus.SinkInstanceStatus.SinkInstanceStatusData sinkInstanceStatusData = new SinkStatus.SinkInstanceStatus.SinkInstanceStatusData();
                        sinkInstanceStatus.setInstanceId(i);
                        sinkInstanceStatus.setStatus(sinkInstanceStatusData);
                        sinkStatus.addInstance(sinkInstanceStatus);
                    }
                    if (v1StatefulSet.getStatus().getReadyReplicas() != null) {
                        sinkStatus.setNumRunning(v1StatefulSet.getStatus().getReadyReplicas());
                    }
                }
            } else {
                log.error(
                        "no StatefulSet status exists when get status of sink {}/{}/{}",
                        tenant,
                        namespace,
                        componentName);
                throw new RestException(Response.Status.NOT_FOUND, "no StatefulSet status exists");
            }
            V1PodList podList = worker().getCoreV1Api().listNamespacedPod(
                    nameSpaceName, null, null, null, null,
                    sinkLabelSelector, null, null, null, null,
                    null);
            if (podList != null ) {
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
                        SinkStatus.SinkInstanceStatus sinkInstanceStatus = null;
                        for (SinkStatus.SinkInstanceStatus ins : sinkStatus.getInstances()) {
                            if (ins.getInstanceId() == shardId) {
                                sinkInstanceStatus = ins;
                                break;
                            }
                        }
                        if (sinkInstanceStatus != null) {
                            SinkStatus.SinkInstanceStatus.SinkInstanceStatusData sinkInstanceStatusData = sinkInstanceStatus.getStatus();
                            V1PodStatus podStatus = pod.getStatus();
                            if (v1alpha1Sink.getSpec() != null && StringUtils.isNotEmpty(v1alpha1Sink.getSpec().getClusterName())) {
                                sinkInstanceStatusData.setWorkerId(v1alpha1Sink.getSpec().getClusterName());
                            }
                            if (podStatus != null) {
                                sinkInstanceStatusData.setRunning(KubernetesUtils.isPodRunning(pod));
                                if (podStatus.getContainerStatuses() != null && !podStatus.getContainerStatuses().isEmpty()) {
                                    V1ContainerStatus containerStatus = podStatus.getContainerStatuses().get(0);
                                    sinkInstanceStatusData.setNumRestarts(containerStatus.getRestartCount());
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
                                if (channel[podIndex]!= null) {
                                    log.debug("closing channel {}", podIndex);
                                    channel[podIndex].shutdown();
                                }
                                if (e != null) {
                                    log.error("Get sink {}-{} status from grpc failed from namespace {}, error message: {}",
                                            finalStatefulSetName,
                                            shardId,
                                            nameSpaceName,
                                            e.getMessage());
                                    sinkInstanceStatusData.setError(e.getMessage());
                                } else if (fs != null) {
                                    SinksUtil.convertFunctionStatusToInstanceStatusData(fs, sinkInstanceStatusData);
                                }
                            });
                            completableFutureSet.add(future);
                        } else {
                            log.error("Get sink {}-{} status failed from namespace {}, cannot find status for shardId {}",
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
                        SinkStatus.SinkInstanceStatus sinkInstanceStatus = null;
                        for (SinkStatus.SinkInstanceStatus ins : sinkStatus.getInstances()) {
                            if (ins.getInstanceId() == shardId) {
                                sinkInstanceStatus = ins;
                                break;
                            }
                        }
                        if (sinkInstanceStatus != null) {
                            SinkStatus.SinkInstanceStatus.SinkInstanceStatusData sinkInstanceStatusData = sinkInstanceStatus.getStatus();
                            V1PodStatus podStatus = pod.getStatus();
                            if (podStatus != null) {
                                List<V1ContainerStatus> containerStatuses = podStatus.getContainerStatuses();
                                if (containerStatuses != null && !containerStatuses.isEmpty()) {
                                    V1ContainerStatus containerStatus = null;
                                    for (V1ContainerStatus s : containerStatuses){
                                        if (s.getImage().contains(v1alpha1Sink.getSpec().getImage())) {
                                            containerStatus = s;
                                            break;
                                        }
                                    }
                                    if (containerStatus != null) {
                                        V1ContainerState state = containerStatus.getState();
                                        if (state != null && state.getTerminated() != null) {
                                            sinkInstanceStatusData.setError(state.getTerminated().getMessage());
                                        } else if (state != null && state.getWaiting() != null) {
                                            sinkInstanceStatusData.setError(state.getWaiting().getMessage());
                                        } else {
                                            V1ContainerState lastState = containerStatus.getLastState();
                                            if (lastState != null && lastState.getTerminated() != null) {
                                                sinkInstanceStatusData.setError(lastState.getTerminated().getMessage());
                                            } else if (lastState != null && lastState.getWaiting() != null) {
                                                sinkInstanceStatusData.setError(lastState.getWaiting().getMessage());
                                            }
                                        }
                                        if (containerStatus.getRestartCount() != null) {
                                            sinkInstanceStatusData.setNumRestarts(containerStatus.getRestartCount());
                                        }
                                    } else {
                                        sinkInstanceStatusData.setError(podStatus.getPhase());
                                    }
                                }
                            }
                        }  else {
                            log.error("Get sink {}-{} status failed from namespace {}, cannot find status for shardId {}",
                                finalStatefulSetName,
                                shardId,
                                nameSpaceName,
                                shardId);
                        }
                    });
                }
            }
        } catch (Exception e) {
            log.error(
                    "Get sink {} status failed from namespace {}, error message: {}",
                    componentName,
                    namespace,
                    e.getMessage());
        }

        return sinkStatus;
    }

    @Override
    public SinkConfig getSinkInfo(
            final String tenant, final String namespace, final String componentName) {
        validateSinkEnabled();
        this.validateGetInfoRequestParams(tenant, namespace, componentName, kind);
        try {
            String hashName = CommonUtil.generateObjectName(worker(), tenant, namespace, componentName);
            Call call =
                    worker().getCustomObjectsApi()
                            .getNamespacedCustomObjectCall(
                                    group, version, KubernetesUtils.getNamespace(worker().getFactoryConfig()),
                                    plural, hashName, null);

            V1alpha1Sink v1alpha1Sink = executeCall(call, V1alpha1Sink.class);
            return SinksUtil.createSinkConfigFromV1alpha1Sink(
                    tenant, namespace, componentName, v1alpha1Sink);
        } catch (Exception e) {
            log.error(
                    "get {}/{}/{} function info failed, error message: {}",
                    tenant,
                    namespace,
                    componentName,
                    e);
            throw new RestException(Response.Status.INTERNAL_SERVER_ERROR, e.getMessage());
        }
    }

    @Override
    public List<String> listFunctions(final String tenant,
                                      final String namespace,
                                      final String clientRole,
                                      final AuthenticationDataSource clientAuthenticationDataHttps) {
        validateSinkEnabled();
        return super.listFunctions(tenant, namespace, clientRole, clientAuthenticationDataHttps);
    }

    @Override
    public List<ConnectorDefinition> getSinkList() {
        validateSinkEnabled();
        List<ConnectorDefinition> connectorDefinitions = getListOfConnectors();
        List<ConnectorDefinition> retval = new ArrayList<>();
        for (ConnectorDefinition connectorDefinition : connectorDefinitions) {
            if (!org.apache.commons.lang.StringUtils.isEmpty(connectorDefinition.getSinkClass())) {
                retval.add(connectorDefinition);
            }
        }
        return retval;
    }

    @Override
    public List<ConfigFieldDefinition> getSinkConfigDefinition(String name) {
        validateSinkEnabled();
        return new ArrayList<>();
    }

    private void upsertSink(final String tenant,
                            final String namespace,
                            final String sinkName,
                            final SinkConfig sinkConfig,
                            V1alpha1Sink v1alpha1Sink,
                            AuthenticationDataHttps clientAuthenticationDataHttps) {
        if (worker().getWorkerConfig().isAuthenticationEnabled()) {
            if (clientAuthenticationDataHttps != null) {
                try {
                    V1alpha1SinkSpecPod podPolicy = v1alpha1Sink.getSpec().getPod();
                    if (podPolicy == null) {
                        podPolicy = new V1alpha1SinkSpecPod();
                        v1alpha1Sink.getSpec().setPod(podPolicy);
                    }
                    MeshWorkerServiceCustomConfig customConfig = worker().getMeshWorkerServiceCustomConfig();
                    if (customConfig != null && StringUtils.isNotEmpty(customConfig.getExtraDependenciesDir())) {
                        V1alpha1SinkSpecJava v1alpha1SinkSpecJava = null;
                        if (v1alpha1Sink.getSpec() != null && v1alpha1Sink.getSpec().getJava() != null) {
                            v1alpha1SinkSpecJava = v1alpha1Sink.getSpec().getJava();
                        } else if (v1alpha1Sink.getSpec() != null && v1alpha1Sink.getSpec().getJava() == null &&
                                v1alpha1Sink.getSpec().getPython() == null &&
                                v1alpha1Sink.getSpec().getGolang() == null){
                            v1alpha1SinkSpecJava = new V1alpha1SinkSpecJava();
                        }
                        if (v1alpha1SinkSpecJava != null) {
                            v1alpha1SinkSpecJava.setExtraDependenciesDir(customConfig.getExtraDependenciesDir());
                            v1alpha1Sink.getSpec().setJava(v1alpha1SinkSpecJava);
                        }
                    }
                    List<V1alpha1SinkSpecPodVolumes> volumesList = customConfig.asV1alpha1SinkSpecPodVolumesList();
                    if (volumesList != null && !volumesList.isEmpty()) {
                        v1alpha1Sink.getSpec().getPod().setVolumes(volumesList);
                    }
                    List<V1alpha1SinkSpecPodVolumeMounts> volumeMountsList =
                            customConfig.asV1alpha1SinkSpecPodVolumeMountsList();
                    if (volumeMountsList != null && !volumeMountsList.isEmpty()) {
                        v1alpha1Sink.getSpec().setVolumeMounts(volumeMountsList);
                    }
                    if (!StringUtils.isEmpty(worker().getWorkerConfig().getBrokerClientAuthenticationPlugin())
                            && !StringUtils.isEmpty(worker().getWorkerConfig().getBrokerClientAuthenticationParameters())) {
                        String authSecretName = KubernetesUtils.upsertSecret(kind.toLowerCase(), "auth",
                                v1alpha1Sink.getSpec().getClusterName(), tenant, namespace, sinkName,
                                worker().getWorkerConfig(), worker().getCoreV1Api(), worker().getFactoryConfig());
                        v1alpha1Sink.getSpec().getPulsar().setAuthSecret(authSecretName);
                    }
                    if (worker().getWorkerConfig().getTlsEnabled()) {
                        String tlsSecretName = KubernetesUtils.upsertSecret(kind.toLowerCase(), "tls",
                                v1alpha1Sink.getSpec().getClusterName(), tenant, namespace, sinkName,
                                worker().getWorkerConfig(), worker().getCoreV1Api(), worker().getFactoryConfig());
                        v1alpha1Sink.getSpec().getPulsar().setTlsSecret(tlsSecretName);
                    }
                    if (!StringUtils.isEmpty(customConfig.getDefaultServiceAccountName())
                            && StringUtils.isEmpty(v1alpha1Sink.getSpec().getPod().getServiceAccountName())) {
                        v1alpha1Sink.getSpec().getPod().setServiceAccountName(
                                customConfig.getDefaultServiceAccountName());
                    }
                } catch (Exception e) {
                    log.error("Error create or update auth or tls secret data for {} {}/{}/{}",
                            ComponentTypeUtils.toString(componentType), tenant, namespace, sinkName, e);


                    throw new RestException(Response.Status.INTERNAL_SERVER_ERROR,
                            String.format("Error create or update auth or tls secret for %s %s:- %s",
                            ComponentTypeUtils.toString(componentType), sinkName, e.getMessage()));
                }
            }
        }
    }
}
