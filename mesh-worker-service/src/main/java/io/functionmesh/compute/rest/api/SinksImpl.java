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
import io.kubernetes.client.openapi.apis.CustomObjectsApi;
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
import org.apache.pulsar.functions.utils.ComponentTypeUtils;
import org.apache.pulsar.functions.worker.service.api.Sinks;
import org.glassfish.jersey.media.multipart.FormDataContentDisposition;
import javax.ws.rs.core.Response;
import java.io.InputStream;
import java.net.URI;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.function.Supplier;

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
        Map<String, Object> customConfig = worker().getWorkerConfig().getFunctionsWorkerServiceCustomConfigs();
        if (customConfig != null) {
            Boolean sinkEnabled = (Boolean) customConfig.get("sinkEnabled");
            if (sinkEnabled != null && !sinkEnabled) {
                throw new RestException(Response.Status.BAD_REQUEST, "Sink API is disabled");
            }
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
        Map<String, Object> customConfig = worker().getWorkerConfig().getFunctionsWorkerServiceCustomConfigs();
        if (jarUploaded &&  customConfig != null && customConfig.get("uploadEnabled") != null &&
                ! (Boolean) customConfig.get("uploadEnabled") ) {
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
                        worker().getWorkerConfig().getFunctionsWorkerServiceCustomConfigs(), cluster);
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
            e.printStackTrace();
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
                            worker().getWorkerConfig().getFunctionsWorkerServiceCustomConfigs(), cluster);
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
            Call call =
                    worker().getCustomObjectsApi()
                            .getNamespacedCustomObjectCall(
                                    group, version, KubernetesUtils.getNamespace(worker().getFactoryConfig()),
                                    plural, hashName, null);
            V1alpha1Sink v1alpha1Sink = executeCall(call, V1alpha1Sink.class);
            SinkStatus.SinkInstanceStatus sinkInstanceStatus = new SinkStatus.SinkInstanceStatus();
            SinkStatus.SinkInstanceStatus.SinkInstanceStatusData sinkInstanceStatusData =
                    new SinkStatus.SinkInstanceStatus.SinkInstanceStatusData();
            sinkInstanceStatusData.setRunning(false);
            V1alpha1SinkStatus v1alpha1SinkStatus = v1alpha1Sink.getStatus();
            if (v1alpha1SinkStatus != null) {
                boolean running =
                        v1alpha1SinkStatus.getConditions().values().stream()
                                .allMatch(
                                        n ->
                                                StringUtils.isNotEmpty(n.getStatus())
                                                        && n.getStatus().equals("True"));
                sinkInstanceStatusData.setRunning(running);
                sinkInstanceStatusData.setWorkerId(v1alpha1Sink.getSpec().getClusterName());
                sinkStatus.addInstance(sinkInstanceStatus);
                sinkInstanceStatus.setStatus(sinkInstanceStatusData);
            } else {
                sinkInstanceStatusData.setRunning(false);
            }

            sinkStatus.setNumInstances(sinkStatus.getInstances().size());
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
                    Map<String, Object>  functionsWorkerServiceCustomConfigs = worker()
                            .getWorkerConfig().getFunctionsWorkerServiceCustomConfigs();
                    Object volumes = functionsWorkerServiceCustomConfigs.get("volumes");
                    if (functionsWorkerServiceCustomConfigs.get("extraDependenciesDir") != null) {
                        V1alpha1SinkSpecJava v1alpha1SinkSpecJava;
                        if (v1alpha1Sink.getSpec() != null && v1alpha1Sink.getSpec().getJava() != null) {
                            v1alpha1SinkSpecJava = v1alpha1Sink.getSpec().getJava();
                        } else {
                            v1alpha1SinkSpecJava = new V1alpha1SinkSpecJava();
                        }
                        v1alpha1SinkSpecJava.setExtraDependenciesDir(
                                (String)functionsWorkerServiceCustomConfigs.get("extraDependenciesDir"));
                        v1alpha1Sink.getSpec().setJava(v1alpha1SinkSpecJava);
                    }
                    if (volumes != null) {
                        List<V1alpha1SinkSpecPodVolumes> volumesList = (List<V1alpha1SinkSpecPodVolumes>) volumes;
                        v1alpha1Sink.getSpec().getPod().setVolumes(volumesList);
                    }
                    Object volumeMounts = functionsWorkerServiceCustomConfigs.get("volumeMounts");
                    if (volumeMounts != null) {
                        List<V1alpha1SinkSpecPodVolumeMounts> volumeMountsList = (List<V1alpha1SinkSpecPodVolumeMounts>) volumeMounts;
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
