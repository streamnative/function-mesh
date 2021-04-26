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

import io.functionmesh.compute.util.SourcesUtil;
import io.functionmesh.compute.MeshWorkerService;
import io.functionmesh.compute.sources.models.V1alpha1Source;
import io.functionmesh.compute.sources.models.V1alpha1SourceList;
import io.functionmesh.compute.sources.models.V1alpha1SourceStatus;
import lombok.extern.slf4j.Slf4j;
import okhttp3.Call;
import org.apache.commons.lang3.StringUtils;
import org.apache.pulsar.broker.authentication.AuthenticationDataHttps;
import org.apache.pulsar.broker.authentication.AuthenticationDataSource;
import org.apache.pulsar.common.functions.UpdateOptions;
import org.apache.pulsar.common.io.ConfigFieldDefinition;
import org.apache.pulsar.common.io.ConnectorDefinition;
import org.apache.pulsar.common.io.SourceConfig;
import org.apache.pulsar.common.policies.data.SourceStatus;
import org.apache.pulsar.common.util.RestException;
import org.apache.pulsar.functions.proto.Function;
import org.apache.pulsar.functions.worker.service.api.Sources;
import org.glassfish.jersey.media.multipart.FormDataContentDisposition;

import javax.ws.rs.core.Response;
import java.io.InputStream;
import java.net.URI;
import java.util.ArrayList;
import java.util.List;
import java.util.function.Supplier;

@Slf4j
public class SourcesImpl extends MeshComponentImpl implements Sources<MeshWorkerService> {
    private final String kind = "Source";

    private final String plural = "sources";

    public SourcesImpl(Supplier<MeshWorkerService> meshWorkerServiceSupplier) {
        super(meshWorkerServiceSupplier, Function.FunctionDetails.ComponentType.SOURCE);
    }

    private void validateRegisterSourceRequestParams(String tenant, String namespace, String sourceName,
                                                     SourceConfig sourceConfig) {
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
    }

    private void validateUpdateSourceRequestParams(String tenant, String namespace, String sourceName,
                                                   SourceConfig sourceConfig) {
        this.validateRegisterSourceRequestParams(tenant, namespace, sourceName, sourceConfig);
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
        validateRegisterSourceRequestParams(tenant, namespace, sourceName, sourceConfig);

        try {
            V1alpha1Source v1alpha1Source = SourcesUtil.createV1alpha1SourceFromSourceConfig(kind, group, version,
                    sourceName, sourcePkgUrl, uploadedInputStream, sourceConfig);
            Call call = worker().getCustomObjectsApi().createNamespacedCustomObjectCall(group, version, namespace,
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
        validateUpdateSourceRequestParams(tenant, namespace, sourceName, sourceConfig);

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
                    sourceConfig
            );
            v1alpha1Source.getMetadata().setResourceVersion(oldRes.getMetadata().getResourceVersion());
            Call replaceCall = worker().getCustomObjectsApi().replaceNamespacedCustomObjectCall(
                    group,
                    version,
                    namespace,
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
        SourceStatus sourceStatus = new SourceStatus();
        try {
            Call call = worker().getCustomObjectsApi().getNamespacedCustomObjectCall(
                    group, version, namespace, plural, componentName, null);
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
        SourceStatus.SourceInstanceStatus.SourceInstanceStatusData sourceInstanceStatusData =
                new SourceStatus.SourceInstanceStatus.SourceInstanceStatusData();
        return sourceInstanceStatusData;
    }

    public SourceConfig getSourceInfo(final String tenant,
                                      final String namespace,
                                      final String componentName) {
        validateGetSourceInfoRequestParams(tenant, namespace, componentName);

        try {
            Call call = worker().getCustomObjectsApi().getNamespacedCustomObjectCall(
                    group,
                    version,
                    namespace,
                    plural,
                    componentName,
                    null
            );

            V1alpha1Source v1alpha1Source = executeCall(call, V1alpha1Source.class);
            return SourcesUtil.createSourceConfigFromV1alpha1Source(tenant, namespace, componentName, v1alpha1Source);
        } catch (Exception e) {
            log.error("get {}/{}/{} function info failed, error message: {}", tenant, namespace, componentName, e);
            throw new RestException(Response.Status.INTERNAL_SERVER_ERROR, e.getMessage());
        }
    }

    public List<ConnectorDefinition> getSourceList() {
        List<ConnectorDefinition> connectorDefinitions = new ArrayList<>();

        try {
            Call call = worker().getCustomObjectsApi().listClusterCustomObjectCall(
                    group,
                    version,
                    plural,
                    null,
                    null,
                    null,
                    null,
                    null,
                    null,
                    null,
                    null,
                    null
            );

            V1alpha1SourceList v1alpha1SourceList = executeCall(call, V1alpha1SourceList.class);
            v1alpha1SourceList.getItems().stream().forEach((n) -> {
                ConnectorDefinition connectorDefinition = new ConnectorDefinition();
                connectorDefinition.setName(n.getSpec().getClassName());
                connectorDefinition.setSourceClass(n.getSpec().getSourceType());

                connectorDefinitions.add(connectorDefinition);
            });
        } catch (Exception e) {
            log.error("get source list failed, error message: {}", e.getMessage());
            throw new RestException(Response.Status.INTERNAL_SERVER_ERROR, e.getMessage());
        }

        return connectorDefinitions;
    }


    public List<ConfigFieldDefinition> getSourceConfigDefinition(String name) {
        return new ArrayList<>();
    }
}
