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
import io.functionmesh.compute.functions.models.V1alpha1Function;
import io.functionmesh.compute.util.FunctionsUtil;
import lombok.extern.slf4j.Slf4j;
import okhttp3.Call;
import org.apache.pulsar.broker.authentication.AuthenticationDataHttps;
import org.apache.pulsar.broker.authentication.AuthenticationDataSource;
import org.apache.pulsar.common.functions.FunctionConfig;
import org.apache.pulsar.common.functions.UpdateOptions;
import org.apache.pulsar.common.policies.data.FunctionStatus;
import org.apache.pulsar.common.util.RestException;
import org.apache.pulsar.functions.proto.Function;
import org.apache.pulsar.functions.worker.service.api.Functions;
import org.glassfish.jersey.media.multipart.FormDataContentDisposition;

import javax.ws.rs.core.Response;
import java.io.InputStream;
import java.net.URI;
import java.util.function.Supplier;

@Slf4j
public class FunctionsImpl extends MeshComponentImpl implements Functions<MeshWorkerService> {

    public FunctionsImpl(Supplier<MeshWorkerService> meshWorkerServiceSupplier) {
        super(meshWorkerServiceSupplier, Function.FunctionDetails.ComponentType.FUNCTION);
    }


    private void validateRegisterFunctionRequestParams(String tenant, String namespace, String functionName,
                                                       FunctionConfig functionConfig) {
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
    }

    private void validateUpdateFunctionRequestParams(String tenant, String namespace, String functionName,
                                                     FunctionConfig functionConfig) {
        validateRegisterFunctionRequestParams(tenant, namespace, functionName, functionConfig);
    }

    private void validateDeregisterFunctionRequestParams(String tenant, String namespace, String functionName) {
        if (tenant == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Tenant is not provided");
        }
        if (namespace == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Namespace is not provided");
        }
        if (functionName == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Function name is not provided");
        }
    }

    private void validateGetFunctionInfoRequestParams(String tenant, String namespace, String functionName) {
        validateDeregisterFunctionRequestParams(tenant, namespace, functionName);
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
        validateRegisterFunctionRequestParams(tenant, namespace, functionName, functionConfig);

        V1alpha1Function v1alpha1Function = FunctionsUtil.createV1alpha1FunctionFromFunctionConfig(
                kind,
                group,
                version,
                functionName,
                functionPkgUrl,
                functionConfig
        );
        try {
            Call call = worker().getCustomObjectsApi().createNamespacedCustomObjectCall(
                    group,
                    version,
                    namespace,
                    plural,
                    v1alpha1Function,
                    null,
                    null,
                    null,
                    null
            );
            executeCall(call, V1alpha1Function.class);
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
                               UpdateOptions updateOptions) {
        validateUpdateFunctionRequestParams(tenant, namespace, functionName, functionConfig);

        try {
            Call getCall = worker().getCustomObjectsApi().getNamespacedCustomObjectCall(
                    group,
                    version,
                    namespace,
                    plural,
                    functionName,
                    null
            );
            V1alpha1Function oldFn = executeCall(getCall, V1alpha1Function.class);
            V1alpha1Function v1alpha1Function = FunctionsUtil.createV1alpha1FunctionFromFunctionConfig(
                    kind,
                    group,
                    version,
                    functionName,
                    functionPkgUrl,
                    functionConfig
            );
            v1alpha1Function.getMetadata().setResourceVersion(oldFn.getMetadata().getResourceVersion());
            Call replaceCall = worker().getCustomObjectsApi().replaceNamespacedCustomObjectCall(
                    group,
                    version,
                    namespace,
                    plural,
                    functionName,
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
        validateGetFunctionInfoRequestParams(tenant, namespace, componentName);

        try {
            Call call = worker().getCustomObjectsApi().getNamespacedCustomObjectCall(
                    group,
                    version,
                    namespace,
                    plural,
                    componentName,
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

        return null;
    }

    @Override
    public FunctionStatus getFunctionStatus(final String tenant,
                                            final String namespace,
                                            final String componentName,
                                            final URI uri,
                                            final String clientRole,
                                            final AuthenticationDataSource clientAuthenticationDataHttps) {
        FunctionStatus functionStatus = new FunctionStatus();
        try {
            Call call = worker().getCustomObjectsApi().getNamespacedCustomObjectCall(
                    group, version, namespace, plural, componentName, null);
            V1alpha1Function v1alpha1Function = executeCall(call, V1alpha1Function.class);
            FunctionStatus.FunctionInstanceStatus functionInstanceStatus = new FunctionStatus.FunctionInstanceStatus();
            FunctionStatus.FunctionInstanceStatus.FunctionInstanceStatusData functionInstanceStatusData =
                    new FunctionStatus.FunctionInstanceStatus.FunctionInstanceStatusData();
            functionInstanceStatusData.setRunning(true);
            if (v1alpha1Function.getStatus() != null) {
                v1alpha1Function.getStatus().getConditions().forEach((s, v1alpha1FunctionStatusConditions) -> {
                    if (v1alpha1FunctionStatusConditions.getStatus() != null
                            && v1alpha1FunctionStatusConditions.getStatus().equals("False")) {
                        functionInstanceStatusData.setRunning(false);
                    }
                });
                functionInstanceStatusData.setWorkerId(v1alpha1Function.getSpec().getClusterName());
                functionInstanceStatus.setStatus(functionInstanceStatusData);
                functionStatus.addInstance(functionInstanceStatus);
            } else {
                functionInstanceStatusData.setRunning(false);
            }
            functionStatus.setNumInstances(functionStatus.getInstances().size());
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
                                             final String clientRole) {

    }
}
