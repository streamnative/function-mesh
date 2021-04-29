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

import io.functionmesh.compute.MeshWorkerService;
import io.functionmesh.compute.sinks.models.V1alpha1Sink;
import io.functionmesh.compute.sinks.models.V1alpha1SinkStatus;
import io.functionmesh.compute.util.SinksUtil;
import lombok.extern.slf4j.Slf4j;
import okhttp3.Call;
import org.apache.commons.lang3.StringUtils;
import org.apache.pulsar.broker.authentication.AuthenticationDataHttps;
import org.apache.pulsar.broker.authentication.AuthenticationDataSource;
import org.apache.pulsar.common.functions.UpdateOptions;
import org.apache.pulsar.common.io.ConfigFieldDefinition;
import org.apache.pulsar.common.io.ConnectorDefinition;
import org.apache.pulsar.common.io.SinkConfig;
import org.apache.pulsar.common.policies.data.SinkStatus;
import org.apache.pulsar.common.util.RestException;
import org.apache.pulsar.functions.proto.Function;
import org.apache.pulsar.functions.worker.service.api.Sinks;
import org.glassfish.jersey.media.multipart.FormDataContentDisposition;

import javax.ws.rs.core.Response;
import java.io.InputStream;
import java.net.URI;
import java.util.ArrayList;
import java.util.List;
import java.util.function.Supplier;

@Slf4j
public class SinksImpl extends MeshComponentImpl
        implements Sinks<MeshWorkerService> {
    private final String kind = "Sink";

    private final String plural = "sinks";

    public SinksImpl(Supplier<MeshWorkerService> meshWorkerServiceSupplier) {
        super(meshWorkerServiceSupplier, Function.FunctionDetails.ComponentType.SINK);
    }

    private void validateRegisterSinkRequestParams(
            String tenant, String namespace, String sinkName, SinkConfig sinkConfig) {
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
    }

    private void validateUpdateSinkRequestParams(
            String tenant, String namespace, String sinkName, SinkConfig sinkConfig) {
        this.validateRegisterSinkRequestParams(tenant, namespace, sinkName, sinkConfig);
    }

    private void validateGetSinkInfoRequestParams(
            String tenant, String namespace, String sinkName) {
        if (tenant == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Tenant is not provided");
        }
        if (namespace == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Namespace is not provided");
        }
        if (sinkName == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Sink name is not provided");
        }
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
        validateRegisterSinkRequestParams(tenant, namespace, sinkName, sinkConfig);
        try {
            V1alpha1Sink v1alpha1Sink =
                    SinksUtil.createV1alpha1SkinFromSinkConfig(
                            kind,
                            group,
                            version,
                            sinkName,
                            sinkPkgUrl,
                            uploadedInputStream,
                            sinkConfig,
                            this.meshWorkerServiceSupplier.get().getConnectorsManager());
            Call call =
                    worker().getCustomObjectsApi()
                            .createNamespacedCustomObjectCall(
                                    group,
                                    version,
                                    namespace,
                                    plural,
                                    v1alpha1Sink,
                                    null,
                                    null,
                                    null,
                                    null);
            executeCall(call, V1alpha1Sink.class);
        } catch (Exception e) {
            log.error(
                    "register {}/{}/{} sink failed, error message: {}",
                    tenant,
                    namespace,
                    sinkConfig,
                    e);
            e.printStackTrace();
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
            UpdateOptions updateOptions) {
        validateUpdateSinkRequestParams(tenant, namespace, sinkName, sinkConfig);

        try {
            Call getCall =
                    worker().getCustomObjectsApi()
                            .getNamespacedCustomObjectCall(
                                    group, version, namespace, plural, sinkName, null);
            V1alpha1Sink oldRes = executeCall(getCall, V1alpha1Sink.class);

            V1alpha1Sink v1alpha1Sink =
                    SinksUtil.createV1alpha1SkinFromSinkConfig(
                            kind,
                            group,
                            version,
                            sinkName,
                            sinkPkgUrl,
                            uploadedInputStream,
                            sinkConfig, this.meshWorkerServiceSupplier.get().getConnectorsManager());
            v1alpha1Sink
                    .getMetadata()
                    .setResourceVersion(oldRes.getMetadata().getResourceVersion());
            Call replaceCall =
                    worker().getCustomObjectsApi()
                            .replaceNamespacedCustomObjectCall(
                                    group,
                                    version,
                                    namespace,
                                    plural,
                                    sinkName,
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
        SinkStatus sinkStatus = new SinkStatus();
        try {
            Call call =
                    worker().getCustomObjectsApi()
                            .getNamespacedCustomObjectCall(
                                    group, version, namespace, plural, componentName, null);
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
        validateGetSinkInfoRequestParams(tenant, namespace, componentName);

        try {
            Call call =
                    worker().getCustomObjectsApi()
                            .getNamespacedCustomObjectCall(
                                    group, version, namespace, plural, componentName, null);

            V1alpha1Sink v1alpha1Sink = executeCall(call, V1alpha1Sink.class);
            return SinksUtil.createSinkConfigFromV1alpha1Source(
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
    public List<ConnectorDefinition> getSinkList() {
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
        return new ArrayList<>();
    }
}
