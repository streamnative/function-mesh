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

import io.streamnative.function.mesh.proxy.FunctionMeshProxyService;
import lombok.extern.slf4j.Slf4j;
import okhttp3.Request;
import org.apache.pulsar.common.functions.WorkerInfo;
import org.apache.pulsar.common.io.ConnectorDefinition;
import org.apache.pulsar.common.policies.data.WorkerFunctionInstanceStats;
import org.apache.pulsar.functions.worker.service.api.Workers;

import javax.servlet.http.HttpServletRequest;
import javax.ws.rs.core.Context;
import java.io.IOException;
import java.net.URI;
import java.util.ArrayList;
import java.util.Collection;
import java.util.List;
import java.util.Map;
import java.util.function.Supplier;

import static com.google.common.base.Preconditions.checkNotNull;

@Slf4j
public class WorkerImpl implements Workers<FunctionMeshProxyService> {

    @Context
    protected HttpServletRequest httpRequest;

    private final Supplier<FunctionMeshProxyService> workerServiceSupplier;

    private final static String NAMESPACE_CLAIM = "x-api-namespace";

    private final static String TYPE_CLAIM = "x-api-type";

    private final static String NAME_CLAIM = "x-api-name";

    public WorkerImpl(Supplier<FunctionMeshProxyService> workerServiceSupplier) {
        this.workerServiceSupplier = workerServiceSupplier;
    }

    private FunctionMeshProxyService worker() {
        try {
            return checkNotNull(workerServiceSupplier.get());
        } catch (Throwable t) {
            log.info("Failed to get worker service", t);
            throw t;
        }
    }

    @Override
    public List<WorkerInfo> getCluster(String clientRole) {
        return null;
    }

    @Override
    public WorkerInfo getClusterLeader(String clientRole) {
        return null;
    }

    @Override
    public List<ConnectorDefinition> getListOfConnectors(String clientRole) {
        List<ConnectorDefinition> connectorDefinitions = new ArrayList<>();
        return connectorDefinitions;
    }

    @Override
    public void rebalance(final URI uri, final String clientRole) {

    }

    @Override
    public Map<String, Collection<String>> getAssignments(String clientRole) {
        return null;
    }

    @Override
    public List<org.apache.pulsar.common.stats.Metrics> getWorkerMetrics(final String clientRole) {
        String namespace = httpRequest.getHeader(NAMESPACE_CLAIM);
        String type = httpRequest.getHeader(TYPE_CLAIM);
        String name = httpRequest.getHeader(NAME_CLAIM);
        String serviceUrl = name;
        if (!name.contains("http://")) {
            serviceUrl = "http://" + name;
        }
//        Request request = new Request.Builder()
//                .url(serviceUrl)
//                .get()
//                .headers(new Headers.Builder().add(
//                        "Authorization",
//                        "Bearer " + this.cloudControllerEntity.getCloudControllerToken()).build())
//                .addHeader("Accept", "application/json")
//                .build();
//        worker().getApiClient().getHttpClient()
        return null;
    }

    @Override
    public List<WorkerFunctionInstanceStats> getFunctionsMetrics(String clientRole) throws IOException {
        return null;
    }

    @Override
    public Boolean isLeaderReady(final String clientRole) {
        return null;
    }
}
