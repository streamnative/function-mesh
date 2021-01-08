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
package io.streamnative.function.mesh.proxy.rest.api;

import io.streamnative.cloud.models.function.V1alpha1Function;
import io.streamnative.function.mesh.proxy.FunctionMeshProxyService;
import lombok.extern.slf4j.Slf4j;
import okhttp3.Call;
import okhttp3.Response;
import org.apache.pulsar.broker.authentication.AuthenticationDataHttps;
import org.apache.pulsar.broker.authentication.AuthenticationDataSource;
import org.apache.pulsar.common.functions.FunctionConfig;
import org.apache.pulsar.common.functions.UpdateOptions;
import org.apache.pulsar.common.policies.data.FunctionStatus;
import org.apache.pulsar.functions.proto.Function;
import org.apache.pulsar.functions.worker.service.api.Functions;
import org.glassfish.jersey.media.multipart.FormDataContentDisposition;

import java.io.IOException;
import java.io.InputStream;
import java.net.URI;
import java.util.function.Supplier;

@Slf4j
public class FunctionsImpl extends FunctionMeshComponentImpl implements Functions<FunctionMeshProxyService> {

    final String plural = "functions";

    public FunctionsImpl(Supplier<FunctionMeshProxyService> functionMeshProxyServiceSupplier) {
        super(functionMeshProxyServiceSupplier, Function.FunctionDetails.ComponentType.FUNCTION);
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
            Response response = call.execute();
            if (response.isSuccessful()) {
                if (response.body() != null) {
                    String data = response.body().string();
                    V1alpha1Function v1alpha1Function = worker().getApiClient()
                                                        .getJSON().getGson().fromJson(data, V1alpha1Function.class);
                    FunctionStatus.FunctionInstanceStatus functionInstanceStatus = new FunctionStatus.FunctionInstanceStatus();
                    FunctionStatus.FunctionInstanceStatus.FunctionInstanceStatusData functionInstanceStatusData =
                            new FunctionStatus.FunctionInstanceStatus.FunctionInstanceStatusData();
                    functionInstanceStatusData.setRunning(true);
                    if ( v1alpha1Function.getStatus() != null) {
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
                } else {
                    log.error("Parse json failed for {}", response.body());
                }
            }
        } catch (io.kubernetes.client.openapi.ApiException e) {
            log.error("Get function {} status failed from namespace {}, error message: {}",
                    componentName, namespace, e.getMessage());
        } catch (IOException ioException) {
            log.error("Parse response failed, error message: {}", ioException.getMessage());
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
