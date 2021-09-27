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
package io.functionmesh.compute.util;

import static org.junit.Assert.assertNotNull;
import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertNull;

import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecPodVolumes;
import io.functionmesh.compute.models.MeshWorkerServiceCustomConfig;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecPodVolumes;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecPodVolumes;
import io.kubernetes.client.openapi.models.V1OwnerReference;

import java.util.List;
import java.util.Map;
import lombok.extern.slf4j.Slf4j;
import org.apache.pulsar.functions.runtime.RuntimeUtils;
import org.apache.pulsar.functions.worker.WorkerConfig;
import org.junit.Test;

@Slf4j
public class WorkerConfigTest {
    @Test
    public void testCustomConfigs() throws Exception {
        WorkerConfig workerConfig = WorkerConfig.load(getClass().getClassLoader().getResource("test_worker_config.yaml")
                .toURI().getPath());
        log.info("Got worker config [{}]", workerConfig);
        Map<String, Object> customConfigs = workerConfig.getFunctionsWorkerServiceCustomConfigs();
        log.info("Got custom configs [{}]", customConfigs);
        MeshWorkerServiceCustomConfig customConfig = RuntimeUtils.getRuntimeFunctionConfig(
                customConfigs, MeshWorkerServiceCustomConfig.class);

        assertEquals("service-account", customConfig.getDefaultServiceAccountName());

        V1OwnerReference ownerRef = CommonUtil.getOwnerReferenceFromCustomConfigs(customConfig);
        log.info("Got owner ref [{}]", ownerRef);
        assertEquals("pulsar.streamnative.io/v1alpha1", ownerRef.getApiVersion());
        assertEquals("PulsarBroker", ownerRef.getKind());
        assertEquals("test", ownerRef.getName());
        assertEquals("4627a402-35f2-40ac-b3fc-1bae5a2bd626", ownerRef.getUid());
        assertNull(ownerRef.getBlockOwnerDeletion());
        assertNull(ownerRef.getController());

        List<V1alpha1SourceSpecPodVolumes> v1alpha1SourceSpecPodVolumesList =
                customConfig.asV1alpha1SourceSpecPodVolumesList();
        assertEquals(1, v1alpha1SourceSpecPodVolumesList.size());
        assertEquals("secret-pulsarcluster-data", v1alpha1SourceSpecPodVolumesList.get(0).getName());
        assertNotNull("pulsarcluster-data", v1alpha1SourceSpecPodVolumesList.get(0).getSecret());
        assertEquals("pulsarcluster-data", v1alpha1SourceSpecPodVolumesList.get(0).getSecret().getSecretName());

        List<V1alpha1SinkSpecPodVolumes> v1alpha1SinkSpecPodVolumesList =
                customConfig.asV1alpha1SinkSpecPodVolumesList();
        assertEquals(1, v1alpha1SinkSpecPodVolumesList.size());
        assertEquals("secret-pulsarcluster-data", v1alpha1SinkSpecPodVolumesList.get(0).getName());
        assertNotNull("pulsarcluster-data", v1alpha1SinkSpecPodVolumesList.get(0).getSecret());
        assertEquals("pulsarcluster-data", v1alpha1SinkSpecPodVolumesList.get(0).getSecret().getSecretName());

        List<V1alpha1FunctionSpecPodVolumes> v1alpha1FunctionSpecPodVolumesList =
                customConfig.asV1alpha1FunctionSpecPodVolumesList();
        assertEquals(1, v1alpha1FunctionSpecPodVolumesList.size());
        assertEquals("secret-pulsarcluster-data", v1alpha1FunctionSpecPodVolumesList.get(0).getName());
        assertNotNull("pulsarcluster-data", v1alpha1FunctionSpecPodVolumesList.get(0).getSecret());
        assertEquals("pulsarcluster-data", v1alpha1FunctionSpecPodVolumesList.get(0).getSecret().getSecretName());
    }
}
