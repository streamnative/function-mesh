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

import io.functionmesh.compute.functions.models.V1alpha1FunctionList;
import io.functionmesh.compute.MeshWorkerService;
import lombok.extern.slf4j.Slf4j;
import okhttp3.Call;
import okhttp3.Response;
import org.apache.pulsar.broker.authentication.AuthenticationDataHttps;
import org.apache.pulsar.broker.authentication.AuthenticationDataSource;
import org.apache.pulsar.common.functions.FunctionConfig;
import org.apache.pulsar.common.functions.FunctionState;
import org.apache.pulsar.common.io.ConnectorDefinition;
import org.apache.pulsar.common.policies.data.FunctionStats;
import org.apache.pulsar.functions.proto.Function;
import org.apache.pulsar.functions.worker.service.api.Component;

import javax.ws.rs.core.StreamingOutput;
import java.io.InputStream;
import java.net.URI;
import java.util.LinkedList;
import java.util.List;
import java.util.function.Supplier;

import static com.google.common.base.Preconditions.checkNotNull;

@Slf4j
public abstract class MeshComponentImpl implements Component<MeshWorkerService> {

    protected final Supplier<MeshWorkerService> meshWorkerServiceSupplier;
    protected final Function.FunctionDetails.ComponentType componentType;

    final String plural = "functions";

    final String group = "compute.functionmesh.io";

    final String version = "v1alpha1";

    final String kind = "Function";

    public MeshComponentImpl(Supplier<MeshWorkerService> meshWorkerServiceSupplier,
                             Function.FunctionDetails.ComponentType componentType) {
        this.meshWorkerServiceSupplier = meshWorkerServiceSupplier;
        // If you want to support function-mesh, this type needs to be changed
        this.componentType = componentType;
    }

    @Override
    public FunctionConfig getFunctionInfo(final String tenant,
                                          final String namespace,
                                          final String componentName,
                                          final String clientRole,
                                          final AuthenticationDataSource clientAuthenticationDataHttps) {

        FunctionConfig functionConfig = new FunctionConfig();
        return functionConfig;
    }

    @Override
    public void deregisterFunction(final String tenant,
                                   final String namespace,
                                   final String componentName,
                                   final String clientRole,
                                   AuthenticationDataHttps clientAuthenticationDataHttps) {

    }

    public <T> T executeCall(Call call, Class<T> c) throws Exception {
        Response response;
        response = call.execute();
        if (response.isSuccessful() && response.body() != null) {
            String data = response.body().string();
            if (c == null) {
                return null;
            }
            return worker().getApiClient().getJSON().getGson().fromJson(data, c);
        } else {
            String body = response.body() != null ? response.body().string() : "";
            String err = String.format(
                    "failed to perform the request: responseCode: %s, responseMessage: %s, responseBody: %s",
                    response.code(), response.message(), body);
            throw new Exception(err);
        }
    }

    @Override
    public MeshWorkerService worker() {
        try {
            return checkNotNull(meshWorkerServiceSupplier.get());
        } catch (Throwable t) {
            log.info("Failed to get worker service", t);
            throw t;
        }
    }

    @Override
    public void stopFunctionInstance(final String tenant,
                                     final String namespace,
                                     final String componentName,
                                     final String instanceId,
                                     final URI uri,
                                     final String clientRole,
                                     final AuthenticationDataSource clientAuthenticationDataHttps) {

    }

    @Override
    public void startFunctionInstance(final String tenant,
                                      final String namespace,
                                      final String componentName,
                                      final String instanceId,
                                      final URI uri,
                                      final String clientRole,
                                      final AuthenticationDataSource clientAuthenticationDataHttps) {

    }

    @Override
    public void restartFunctionInstance(final String tenant,
                                        final String namespace,
                                        final String componentName,
                                        final String instanceId,
                                        final URI uri,
                                        final String clientRole,
                                        final AuthenticationDataSource clientAuthenticationDataHttps) {

    }

    @Override
    public void startFunctionInstances(final String tenant,
                                       final String namespace,
                                       final String componentName,
                                       final String clientRole,
                                       final AuthenticationDataSource clientAuthenticationDataHttps) {

    }

    @Override
    public void stopFunctionInstances(final String tenant,
                                      final String namespace,
                                      final String componentName,
                                      final String clientRole,
                                      final AuthenticationDataSource clientAuthenticationDataHttps) {

    }

    @Override
    public void restartFunctionInstances(final String tenant,
                                         final String namespace,
                                         final String componentName,
                                         final String clientRole,
                                         final AuthenticationDataSource clientAuthenticationDataHttps) {

    }

    @Override
    public FunctionStats getFunctionStats(final String tenant,
                                          final String namespace,
                                          final String componentName,
                                          final URI uri,
                                          final String clientRole,
                                          final AuthenticationDataSource clientAuthenticationDataHttps) {
        FunctionStats functionStats = new FunctionStats();

        return functionStats;
    }

    @Override
    public FunctionStats.FunctionInstanceStats.FunctionInstanceStatsData getFunctionsInstanceStats(final String tenant,
                                                                                                   final String namespace,
                                                                                                   final String componentName,
                                                                                                   final String instanceId,
                                                                                                   final URI uri,
                                                                                                   final String clientRole,
                                                                                                   final AuthenticationDataSource clientAuthenticationDataHttps) {
        return new FunctionStats.FunctionInstanceStats.FunctionInstanceStatsData();
    }

    @Override
    public String triggerFunction(final String tenant,
                                  final String namespace,
                                  final String functionName,
                                  final String input,
                                  final InputStream uploadedInputStream,
                                  final String topic,
                                  final String clientRole,
                                  final AuthenticationDataSource clientAuthenticationDataHttps) {
        return "";
    }

    @Override
    public List<String> listFunctions(final String tenant,
                                      final String namespace,
                                      final String clientRole,
                                      final AuthenticationDataSource clientAuthenticationDataHttps) {
        List<String> result = new LinkedList<>();
        try {
            Call call = worker().getCustomObjectsApi().listNamespacedCustomObjectCall(
                    group,
                    version,
                    namespace, plural,
                    "false",
                    null,
                    null,
                    null,
                    null,
                    null,
                    null,
                    false,
                    null);

            V1alpha1FunctionList list = executeCall(call, V1alpha1FunctionList.class);
            list.getItems().forEach(n -> result.add(n.getMetadata().getName()));
        } catch (Exception e) {
            log.error("failed to fetch functions list from namespace {}, error message: {}", namespace, e.getMessage());
        }

        return result;
    }

    @Override
    public FunctionState getFunctionState(final String tenant,
                                          final String namespace,
                                          final String functionName,
                                          final String key,
                                          final String clientRole,
                                          final AuthenticationDataSource clientAuthenticationDataHttps) {
        // To do
        return new FunctionState();
    }

    @Override
    public void putFunctionState(final String tenant,
                                 final String namespace,
                                 final String functionName,
                                 final String key,
                                 final FunctionState state,
                                 final String clientRole,
                                 final AuthenticationDataSource clientAuthenticationDataHttps) {

    }

    @Override
    public void uploadFunction(final InputStream uploadedInputStream,
                               final String path,
                               String clientRole) {

    }

    @Override
    public StreamingOutput downloadFunction(String path,
                                            String clientRole,
                                            AuthenticationDataHttps clientAuthenticationDataHttps) {
        // To do
        return null;
    }

    @Override
    public StreamingOutput downloadFunction(String tenant,
                                            String namespace,
                                            String componentName,
                                            String clientRole,
                                            AuthenticationDataHttps clientAuthenticationDataHttps) {
        // To do
        return null;
    }

    @Override
    public List<ConnectorDefinition> getListOfConnectors() {
        // To do
        return null;
    }

    @Override
    public void reloadConnectors(String clientRole) {

    }
}

