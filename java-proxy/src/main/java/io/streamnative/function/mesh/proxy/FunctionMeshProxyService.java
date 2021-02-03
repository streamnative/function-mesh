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
package io.streamnative.function.mesh.proxy;

import io.kubernetes.client.openapi.ApiClient;
import io.kubernetes.client.openapi.apis.CoreV1Api;
import io.kubernetes.client.openapi.apis.CustomObjectsApi;
import io.kubernetes.client.util.Config;
import io.streamnative.function.mesh.proxy.rest.api.FunctionsImpl;
import lombok.Getter;
import lombok.extern.slf4j.Slf4j;
import org.apache.pulsar.broker.ServiceConfiguration;
import org.apache.pulsar.broker.authentication.AuthenticationService;
import org.apache.pulsar.broker.authorization.AuthorizationService;
import org.apache.pulsar.broker.cache.ConfigurationCacheService;
import org.apache.pulsar.common.conf.InternalConfigurationData;
import org.apache.pulsar.common.util.SimpleTextOutputStream;
import org.apache.pulsar.functions.worker.ErrorNotifier;
import org.apache.pulsar.functions.worker.WorkerConfig;
import org.apache.pulsar.functions.worker.WorkerService;
import org.apache.pulsar.functions.worker.service.api.Functions;
import org.apache.pulsar.functions.worker.service.api.FunctionsV2;
import org.apache.pulsar.functions.worker.service.api.Sinks;
import org.apache.pulsar.functions.worker.service.api.Sources;
import org.apache.pulsar.functions.worker.service.api.Workers;
import org.apache.pulsar.zookeeper.ZooKeeperCache;

import java.io.File;
import java.io.FileReader;
import java.io.IOException;
import java.io.Reader;

/**
 * Function mesh proxy implement.
 */
@Slf4j
@Getter
public class FunctionMeshProxyService implements WorkerService {

    private volatile boolean isInitialized = false;

    private WorkerConfig workerConfig;
    private Functions<FunctionMeshProxyService> functions;
    private FunctionsV2<FunctionMeshProxyService> functionsV2;
    private Sinks<FunctionMeshProxyService> sinks;
    private Sources<FunctionMeshProxyService> sources;
    private CoreV1Api coreV1Api;
    private CustomObjectsApi customObjectsApi;
    private ApiClient apiClient;

    private AuthenticationService authenticationService;
    private AuthorizationService authorizationService;

    public FunctionMeshProxyService() {

    }

    @Override
    public void initAsStandalone(WorkerConfig workerConfig) throws Exception {
        this.init(workerConfig);
    }

    @Override
    public void initInBroker(ServiceConfiguration brokerConfig,
                           WorkerConfig workerConfig,
                           ZooKeeperCache zooKeeperCache,
                           ConfigurationCacheService configurationCacheService,
                           InternalConfigurationData internalConfigurationData) {
        // to do https://github.com/streamnative/function-mesh/issues/57
    }

    public void init(WorkerConfig workerConfig) throws Exception {
        this.workerConfig = workerConfig;
        this.initKubernetesClient();
        this.functions = new FunctionsImpl(() -> FunctionMeshProxyService.this);
    }

    private void initKubernetesClient() throws IOException {
        try {
            apiClient = Config.defaultClient();
            coreV1Api = new CoreV1Api(Config.defaultClient());
            customObjectsApi = new CustomObjectsApi(Config.defaultClient());
        } catch (java.io.IOException e) {
            log.error("Initialization kubernetes client failed, exception: {}", e.getMessage());
            throw e;
        }
    }

    public void start(AuthenticationService authenticationService,
                      AuthorizationService authorizationService,
                      ErrorNotifier errorNotifier) {
        this.authenticationService = authenticationService;
        this.authorizationService = authorizationService;
    }

    public void stop() {

    }

    public boolean isInitialized() {
        return isInitialized;
    }

    public Workers<? extends WorkerService> getWorkers() {
        // to do https://github.com/streamnative/function-mesh/issues/55
        return null;
    }

    public void generateFunctionsStats(SimpleTextOutputStream out) {
        // to do https://github.com/streamnative/function-mesh/issues/56
    }

}
