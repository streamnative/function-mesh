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
package io.functionmesh.compute;

import io.netty.util.concurrent.DefaultThreadFactory;
import lombok.extern.slf4j.Slf4j;
import org.apache.bookkeeper.common.util.OrderedExecutor;
import org.apache.pulsar.broker.PulsarServerException;
import org.apache.pulsar.broker.ServiceConfiguration;
import org.apache.pulsar.broker.authentication.AuthenticationService;
import org.apache.pulsar.broker.authorization.AuthorizationService;
import org.apache.pulsar.broker.cache.ConfigurationCacheService;
import org.apache.pulsar.common.configuration.PulsarConfigurationLoader;
import org.apache.pulsar.functions.worker.ErrorNotifier;
import org.apache.pulsar.functions.worker.WorkerConfig;
import org.apache.pulsar.functions.worker.WorkerService;
import org.apache.pulsar.functions.worker.rest.WorkerServer;
import org.apache.pulsar.zookeeper.GlobalZooKeeperCache;
import org.apache.pulsar.zookeeper.ZooKeeperClientFactory;
import org.apache.pulsar.zookeeper.ZookeeperBkClientFactoryImpl;

import java.io.IOException;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;

/**
 * This class for test.
 */
@Slf4j
public class MeshWorker {

    private ZooKeeperClientFactory zkClientFactory = null;
    private final OrderedExecutor orderedExecutor = OrderedExecutor.newBuilder().numThreads(8).name("zk-cache-ordered").build();
    private final ScheduledExecutorService cacheExecutor = Executors.newScheduledThreadPool(10,
            new DefaultThreadFactory("zk-cache-callback"));
    private GlobalZooKeeperCache globalZkCache;
    private ConfigurationCacheService configurationCacheService;
    private final WorkerConfig workerConfig;
    private final WorkerService workerService;
    private final ErrorNotifier errorNotifier;
    private WorkerServer server;


    public MeshWorker(WorkerConfig workerConfig) {
        this.workerConfig = workerConfig;
        this.workerService = new MeshWorkerService();
        this.errorNotifier = ErrorNotifier.getDefaultImpl();
    }

    protected void start() throws Exception {
        workerService.initAsStandalone(workerConfig);
        workerService.start(getAuthenticationService(), getAuthorizationService(), errorNotifier);
        server = new WorkerServer(workerService, getAuthenticationService());
        server.start();
        log.info("/** Started worker server on port={} **/", this.workerConfig.getWorkerPort());

        try {
            errorNotifier.waitForError();
        } catch (Throwable th) {
            log.error("!-- Fatal error encountered. Worker will exit now. --!", th);
            throw th;
        }
    }

    public ZooKeeperClientFactory getZooKeeperClientFactory() {
        if (zkClientFactory == null) {
            zkClientFactory = new ZookeeperBkClientFactoryImpl(orderedExecutor);
        }
        // Return default factory
        return zkClientFactory;
    }

    private AuthorizationService getAuthorizationService() throws PulsarServerException {
        if (this.workerConfig.isAuthorizationEnabled()) {
            log.info("starting configuration cache service");

            this.globalZkCache = new GlobalZooKeeperCache(getZooKeeperClientFactory(),
                    (int) workerConfig.getZooKeeperSessionTimeoutMillis(),
                    workerConfig.getZooKeeperOperationTimeoutSeconds(),
                    workerConfig.getConfigurationStoreServers(),
                    orderedExecutor, cacheExecutor,
                    workerConfig.getZooKeeperOperationTimeoutSeconds());
            try {
                this.globalZkCache.start();
            } catch (IOException e) {
                throw new PulsarServerException(e);
            }

            this.configurationCacheService = new ConfigurationCacheService(
                    this.globalZkCache, this.workerConfig.getPulsarFunctionsCluster());
            return new AuthorizationService(getServiceConfiguration(), this.configurationCacheService);
        }
        return null;
    }

    private AuthenticationService getAuthenticationService() throws PulsarServerException {
        return new AuthenticationService(getServiceConfiguration());
    }

    private ServiceConfiguration getServiceConfiguration() {
        ServiceConfiguration serviceConfiguration = PulsarConfigurationLoader.convertFrom(workerConfig);
        serviceConfiguration.setClusterName(workerConfig.getPulsarFunctionsCluster());
        return serviceConfiguration;
    }

    protected void stop() {
        if (null != this.server) {
            this.server.stop();
        }
        workerService.stop();
        if (this.globalZkCache != null) {
            try {
                this.globalZkCache.close();
            } catch (IOException e) {
                log.warn("Failed to close global zk cache ", e);
            }
        }
    }
}
