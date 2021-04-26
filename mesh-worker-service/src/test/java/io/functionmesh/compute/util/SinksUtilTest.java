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

import io.functionmesh.compute.sinks.models.V1alpha1Sink;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpec;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecInput;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecJava;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecPulsar;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecResources;
import io.kubernetes.client.openapi.models.V1ObjectMeta;
import org.apache.commons.io.FileUtils;
import org.apache.pulsar.common.functions.Resources;
import org.apache.pulsar.common.io.SinkConfig;
import org.apache.pulsar.common.nar.NarClassLoader;
import org.apache.pulsar.functions.utils.FunctionCommon;
import org.apache.pulsar.functions.utils.io.ConnectorUtils;
import org.junit.Assert;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.powermock.api.mockito.PowerMockito;
import org.powermock.core.classloader.annotations.PowerMockIgnore;
import org.powermock.core.classloader.annotations.PrepareForTest;
import org.powermock.modules.junit4.PowerMockRunner;

import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.net.URISyntaxException;
import java.util.Collections;
import java.util.HashMap;
import java.util.Map;

@RunWith(PowerMockRunner.class)
@PrepareForTest({FunctionCommon.class, ConnectorUtils.class, FileUtils.class})
@PowerMockIgnore({"javax.management.*"})
public class SinksUtilTest {
    private final String kind = "Sink";
    private final String plural = "functions";
    private final String group = "compute.functionmesh.io";
    private final String version = "v1alpha1";

    @Test
    public void testCreateV1alpha1SinkFromSinkConfig()
            throws ClassNotFoundException, IOException, URISyntaxException {
        String tenant = "public";
        String namespace = "default";
        String componentName = "sink-es-sample";
        String className = "org.apache.pulsar.io.elasticsearch.ElasticSearchSink";
        String inputTopic = "persistent://public/default/input";
        String sourceType = getClass().getName();
        String sinkType = sourceType;
        String archive = "connectors/pulsar-io-elastic-search-2.7.0-rc-pm-3.nar";
        boolean autoAck = true;
        int parallelism = 1;
        Double cpu = 0.1;
        Long ram = 1L;
        String clusterName = "test-pulsar";
        String customRuntimeOptions = "{\"clusterName\": \"" + clusterName + "\"}";
        Map<String, Object> configs = new HashMap<>();
        File narFile = PowerMockito.mock(File.class);
        PowerMockito.when(narFile.getPath()).thenReturn("");
        FileInputStream uploadedInputStream = PowerMockito.mock(FileInputStream.class);

        NarClassLoader narClassLoader = PowerMockito.mock(NarClassLoader.class);
        PowerMockito.when(narClassLoader.loadClass(className)).thenReturn(null);
        PowerMockito.mockStatic(FunctionCommon.class);
        PowerMockito.mockStatic(ConnectorUtils.class);
        PowerMockito.mockStatic(FileUtils.class);
        PowerMockito.when(FunctionCommon.extractNarClassLoader(narFile, null))
                .thenReturn(narClassLoader);
        PowerMockito.when(FunctionCommon.createPkgTempFile()).thenReturn(narFile);
        PowerMockito.when(ConnectorUtils.getIOSourceClass(narClassLoader)).thenReturn(className);
        PowerMockito.<Class<?>>when(FunctionCommon.getSourceType(null)).thenReturn(getClass());

        SinkConfig sinkConfig = new SinkConfig();
        sinkConfig.setTenant(tenant);
        sinkConfig.setNamespace(namespace);
        sinkConfig.setName(componentName);
        sinkConfig.setInputs(Collections.singletonList(inputTopic));
        sinkConfig.setConfigs(configs);
        sinkConfig.setArchive(archive);
        sinkConfig.setParallelism(parallelism);
        Resources resources = new Resources();
        resources.setRam(ram);
        resources.setCpu(cpu);
        sinkConfig.setResources(resources);
        sinkConfig.setCustomRuntimeOptions(customRuntimeOptions);
        sinkConfig.setAutoAck(autoAck);

        V1alpha1Sink actualV1alpha1Sink =
                SinksUtil.createV1alpha1SkinFromSinkConfig(
                        kind, group, version, componentName, null, uploadedInputStream, sinkConfig);

        V1alpha1Sink expectedV1alpha1Sink = new V1alpha1Sink();

        expectedV1alpha1Sink.setKind(kind);
        expectedV1alpha1Sink.setApiVersion(String.format("%s/%s", group, version));

        V1ObjectMeta v1ObjectMeta = new V1ObjectMeta();
        v1ObjectMeta.setName(componentName);
        v1ObjectMeta.setNamespace(namespace);
        expectedV1alpha1Sink.setMetadata(v1ObjectMeta);

        V1alpha1SinkSpec v1alpha1SinkSpec = new V1alpha1SinkSpec();
        v1alpha1SinkSpec.setClassName(className);

        v1alpha1SinkSpec.setReplicas(parallelism);
        v1alpha1SinkSpec.setMaxReplicas(parallelism);

        V1alpha1SinkSpecInput v1alpha1SinkSpecInput = new V1alpha1SinkSpecInput();
        v1alpha1SinkSpecInput.setTopics(Collections.singletonList(inputTopic));
        v1alpha1SinkSpecInput.setTypeClassName(sourceType);
        v1alpha1SinkSpec.setInput(v1alpha1SinkSpecInput);

        v1alpha1SinkSpec.setSinkConfig(SinksUtil.transformedMapValueToString(configs));

        V1alpha1SinkSpecPulsar v1alpha1SinkSpecPulsar = new V1alpha1SinkSpecPulsar();
        v1alpha1SinkSpecPulsar.setPulsarConfig(componentName);
        v1alpha1SinkSpec.setPulsar(v1alpha1SinkSpecPulsar);

        V1alpha1SinkSpecResources v1alpha1SinkSpecResources = new V1alpha1SinkSpecResources();
        Map<String, String> requestRes = new HashMap<>();
        String cpuValue = cpu.toString();
        String memoryValue = ram.toString();
        requestRes.put(SourcesUtil.cpuKey, cpuValue);
        requestRes.put(SourcesUtil.memoryKey, memoryValue);
        v1alpha1SinkSpecResources.setLimits(requestRes);
        v1alpha1SinkSpecResources.setRequests(requestRes);
        v1alpha1SinkSpec.setResources(v1alpha1SinkSpecResources);

        String location = String.format("%s/%s/%s", tenant, namespace, componentName);

        V1alpha1SinkSpecJava v1alpha1SinkSpecJava = new V1alpha1SinkSpecJava();
        v1alpha1SinkSpecJava.setJar(archive.split("/")[1]);
        v1alpha1SinkSpecJava.setJarLocation(location);
        v1alpha1SinkSpec.setJava(v1alpha1SinkSpecJava);

        v1alpha1SinkSpec.setClusterName(clusterName);

        v1alpha1SinkSpec.setAutoAck(autoAck);

        expectedV1alpha1Sink.setSpec(v1alpha1SinkSpec);

        Assert.assertEquals(expectedV1alpha1Sink, actualV1alpha1Sink);
    }

    @Test
    public void testCreateSinkConfigFromV1alpha1Sink() {
        String tenant = "public";
        String namespace = "default";
        String componentName = "sink-es-sample";
        String className = "org.apache.pulsar.io.elasticsearch.ElasticSearchSink";
        String inputTopic = "persistent://public/default/input";
        String sourceType = getClass().getName();
        String sinkType = sourceType;
        String archive = "connectors/pulsar-io-elastic-search-2.7.0-rc-pm-3.nar";
        boolean autoAck = true;
        int parallelism = 1;
        Double cpu = 0.1;
        Long ram = 1L;
        String clusterName = "test-pulsar";
        String customRuntimeOptions = "{\"clusterName\":\"" + clusterName + "\"}";
        Map<String, Object> configs = new HashMap<>();
        configs.put("name", "test-config");

        V1alpha1Sink v1alpha1Sink = new V1alpha1Sink();

        v1alpha1Sink.setKind(kind);
        v1alpha1Sink.setApiVersion(String.format("%s/%s", group, version));

        V1ObjectMeta v1ObjectMeta = new V1ObjectMeta();
        v1ObjectMeta.setName(componentName);
        v1ObjectMeta.setNamespace(namespace);
        v1alpha1Sink.setMetadata(v1ObjectMeta);

        V1alpha1SinkSpec v1alpha1SinkSpec = new V1alpha1SinkSpec();
        v1alpha1SinkSpec.setClassName(className);

        v1alpha1SinkSpec.setReplicas(parallelism);
        v1alpha1SinkSpec.setMaxReplicas(parallelism);

        V1alpha1SinkSpecInput v1alpha1SinkSpecInput = new V1alpha1SinkSpecInput();
        v1alpha1SinkSpecInput.setTopics(Collections.singletonList(inputTopic));
        v1alpha1SinkSpecInput.setTypeClassName(sourceType);
        v1alpha1SinkSpec.setInput(v1alpha1SinkSpecInput);

        v1alpha1SinkSpec.setSinkConfig(SinksUtil.transformedMapValueToString(configs));

        V1alpha1SinkSpecPulsar v1alpha1SinkSpecPulsar = new V1alpha1SinkSpecPulsar();
        v1alpha1SinkSpecPulsar.setPulsarConfig(componentName);
        v1alpha1SinkSpec.setPulsar(v1alpha1SinkSpecPulsar);

        V1alpha1SinkSpecResources v1alpha1SinkSpecResources = new V1alpha1SinkSpecResources();
        Map<String, String> requestRes = new HashMap<>();
        String cpuValue = cpu.toString();
        String memoryValue = ram.toString();
        requestRes.put(SourcesUtil.cpuKey, cpuValue);
        requestRes.put(SourcesUtil.memoryKey, memoryValue);
        v1alpha1SinkSpecResources.setLimits(requestRes);
        v1alpha1SinkSpecResources.setRequests(requestRes);
        v1alpha1SinkSpec.setResources(v1alpha1SinkSpecResources);

        String location = String.format("%s/%s/%s", tenant, namespace, componentName);

        V1alpha1SinkSpecJava v1alpha1SinkSpecJava = new V1alpha1SinkSpecJava();
        v1alpha1SinkSpecJava.setJar(archive.split("/")[1]);
        v1alpha1SinkSpecJava.setJarLocation(location);
        v1alpha1SinkSpec.setJava(v1alpha1SinkSpecJava);

        v1alpha1SinkSpec.setClusterName(clusterName);

        v1alpha1SinkSpec.setAutoAck(autoAck);

        v1alpha1Sink.setSpec(v1alpha1SinkSpec);

        SinkConfig actualSinkConfig =
                SinksUtil.createSinkConfigFromV1alpha1Source(
                        tenant, namespace, componentName, v1alpha1Sink);

        SinkConfig expectedSinkConfig = new SinkConfig();
        expectedSinkConfig.setTenant(tenant);
        expectedSinkConfig.setNamespace(namespace);
        expectedSinkConfig.setName(componentName);
        expectedSinkConfig.setClassName(className);
        expectedSinkConfig.setInputs(Collections.singletonList(inputTopic));
        expectedSinkConfig.setConfigs(configs);
        expectedSinkConfig.setArchive(archive.split("/")[1]);
        expectedSinkConfig.setParallelism(parallelism);
        Resources resources = new Resources();
        resources.setRam(ram);
        resources.setCpu(cpu);
        expectedSinkConfig.setResources(resources);
        expectedSinkConfig.setCustomRuntimeOptions(customRuntimeOptions);
        expectedSinkConfig.setAutoAck(autoAck);

        Assert.assertEquals(expectedSinkConfig, actualSinkConfig);
    }
}
