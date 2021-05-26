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
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecJava;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecPod;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecPodVolumeMounts;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecPodVolumes;
import io.functionmesh.compute.util.KubernetesUtils;
import io.functionmesh.compute.util.SourcesUtil;
import io.functionmesh.compute.MeshWorkerService;
import io.functionmesh.compute.sources.models.V1alpha1Source;
import io.functionmesh.compute.sources.models.V1alpha1SourceStatus;
import lombok.extern.slf4j.Slf4j;
import okhttp3.Call;
import org.apache.commons.lang3.StringUtils;
import org.apache.pulsar.broker.authentication.AuthenticationDataHttps;
import org.apache.pulsar.broker.authentication.AuthenticationDataSource;
import org.apache.pulsar.common.functions.Resources;
import org.apache.pulsar.common.functions.UpdateOptions;
import org.apache.pulsar.common.io.ConfigFieldDefinition;
import org.apache.pulsar.common.io.ConnectorDefinition;
import org.apache.pulsar.common.io.SourceConfig;
import org.apache.pulsar.common.policies.data.SourceStatus;
import org.apache.pulsar.common.util.RestException;
import org.apache.pulsar.functions.proto.Function;
import org.apache.pulsar.functions.utils.ComponentTypeUtils;
import org.apache.pulsar.functions.worker.service.api.Sources;
import org.glassfish.jersey.media.multipart.FormDataContentDisposition;
import javax.ws.rs.core.Response;
import java.io.InputStream;
import java.net.URI;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.function.Supplier;

@Slf4j
public class SourcesImpl extends MeshComponentImpl implements Sources<MeshWorkerService> {
    private String kind = "Source";

    private String plural = "sources";

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
        if (jarUploaded &&  customConfig != null && customConfig.get("uploadEnabled") != null &&
                ! (Boolean) customConfig.get("uploadEnabled") ) {
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
        try {
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
                            worker().getWorkerConfig().getFunctionsWorkerServiceCustomConfigs());
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
                             UpdateOptions updateOptions) {
        validateSourceEnabled();
        validateUpdateSourceRequestParams(tenant, namespace, sourceName, sourceConfig, uploadedInputStream != null);
        this.validatePermission(tenant,
                namespace,
                clientRole,
                clientAuthenticationDataHttps,
                ComponentTypeUtils.toString(componentType));
        try {
            Call getCall = worker().getCustomObjectsApi().getNamespacedCustomObjectCall(
                    group,
                    version,
                    namespace,
                    plural,
                    sourceName,
                    null
            );
            V1alpha1Source oldRes = executeCall(getCall, V1alpha1Source.class);

            V1alpha1Source v1alpha1Source = SourcesUtil.createV1alpha1SourceFromSourceConfig(
                    kind,
                    group,
                    version,
                    sourceName,
                    sourcePkgUrl,
                    uploadedInputStream,
                    sourceConfig,
                    this.meshWorkerServiceSupplier.get().getConnectorsManager(),
                    worker().getWorkerConfig().getFunctionsWorkerServiceCustomConfigs()
            );
            v1alpha1Source.getMetadata().setResourceVersion(oldRes.getMetadata().getResourceVersion());
            this.upsertSource(tenant, namespace, sourceName, sourceConfig, v1alpha1Source, clientAuthenticationDataHttps);
            Call replaceCall = worker().getCustomObjectsApi().replaceNamespacedCustomObjectCall(
                    group,
                    version,
                    KubernetesUtils.getNamespace(worker().getFactoryConfig()),
                    plural,
                    sourceName,
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
        this.validatePermission(tenant,
                namespace,
                clientRole,
                clientAuthenticationDataHttps,
                ComponentTypeUtils.toString(componentType));
        SourceStatus sourceStatus = new SourceStatus();
        try {
            Call call = worker().getCustomObjectsApi().getNamespacedCustomObjectCall(
                    group, version, KubernetesUtils.getNamespace(worker().getFactoryConfig()),
                    plural, componentName, null);
            V1alpha1Source v1alpha1Source = executeCall(call, V1alpha1Source.class);
            SourceStatus.SourceInstanceStatus sourceInstanceStatus = new SourceStatus.SourceInstanceStatus();
            SourceStatus.SourceInstanceStatus.SourceInstanceStatusData sourceInstanceStatusData =
                    new SourceStatus.SourceInstanceStatus.SourceInstanceStatusData();
            sourceInstanceStatusData.setRunning(false);
            V1alpha1SourceStatus v1alpha1SourceStatus = v1alpha1Source.getStatus();
            if (v1alpha1SourceStatus != null) {
                boolean running = v1alpha1SourceStatus.getConditions().values()
                        .stream().allMatch(n -> StringUtils.isNotEmpty(n.getStatus()) && n.getStatus().equals("True"));
                sourceInstanceStatusData.setRunning(running);
                sourceInstanceStatusData.setWorkerId(v1alpha1Source.getSpec().getClusterName());
                sourceInstanceStatus.setStatus(sourceInstanceStatusData);
                sourceStatus.addInstance(sourceInstanceStatus);
            } else {
                sourceInstanceStatusData.setRunning(false);
            }
            sourceStatus.setNumInstances(sourceStatus.getInstances().size());
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

        try {
            Call call = worker().getCustomObjectsApi().getNamespacedCustomObjectCall(
                    group,
                    version,
                    KubernetesUtils.getNamespace(worker().getFactoryConfig()),
                    plural,
                    componentName,
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
                    Map<String, Object>  functionsWorkerServiceCustomConfigs = worker()
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
                                (String)functionsWorkerServiceCustomConfigs.get("extraDependenciesDir"));
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
