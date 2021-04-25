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

import io.functionmesh.compute.sources.models.V1alpha1Source;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpec;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecJava;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecOutput;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecPulsar;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecResources;
import io.kubernetes.client.openapi.models.V1ObjectMeta;
import org.apache.commons.io.FileUtils;
import org.apache.pulsar.common.functions.Resources;
import org.apache.pulsar.common.io.SourceConfig;
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
import java.util.HashMap;
import java.util.Map;

@RunWith(PowerMockRunner.class)
@PrepareForTest({FunctionCommon.class, ConnectorUtils.class, FileUtils.class})
@PowerMockIgnore({"javax.management.*"})
public class SourcesUtilTest {
    private final String kind = "Source";
    private final String plural = "functions";
    private final String group = "compute.functionmesh.io";
    private final String version = "v1alpha1";

    @Test
    public void testCreateV1alpha1SourceFromSourceConfig() throws ClassNotFoundException, IOException, URISyntaxException {
        String tenant = "public";
        String namespace = "default";
        String componentName = "source-mongodb-sample";
        String className = "org.apache.pulsar.io.debezium.mongodb.DebeziumMongoDbSource";
        String topicName = "persistent://public/default/destination";
        String sourceType = getClass().getName();
        String sinkType = sourceType;
        String archive = "connectors/pulsar-io-debezium-mongodb-2.7.0.nar";
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
        PowerMockito.when(FunctionCommon.extractNarClassLoader(narFile, null)).thenReturn(narClassLoader);
        PowerMockito.when(FunctionCommon.createPkgTempFile()).thenReturn(narFile);
        PowerMockito.when(ConnectorUtils.getIOSourceClass(narClassLoader)).thenReturn(className);
        PowerMockito.<Class<?>>when(FunctionCommon.getSourceType(null)).thenReturn(getClass());

        SourceConfig sourceConfig = new SourceConfig();
        sourceConfig.setTenant(tenant);
        sourceConfig.setNamespace(namespace);
        sourceConfig.setName(componentName);
        sourceConfig.setClassName(className);
        sourceConfig.setTopicName(topicName);
        sourceConfig.setSchemaType(className);
        sourceConfig.setArchive(archive);
        sourceConfig.setParallelism(parallelism);
        sourceConfig.setConfigs(configs);
        Resources resources = new Resources();
        resources.setRam(ram);
        resources.setCpu(cpu);
        sourceConfig.setResources(resources);
        sourceConfig.setCustomRuntimeOptions(customRuntimeOptions);

        V1alpha1Source v1alpha1Source = SourcesUtil.createV1alpha1SourceFromSourceConfig(kind, group, version,
                componentName, null, uploadedInputStream, sourceConfig);

        V1alpha1Source expectedV1alpha1Source = new V1alpha1Source();

        expectedV1alpha1Source.setKind(kind);
        expectedV1alpha1Source.setApiVersion(String.format("%s/%s", group, version));

        V1ObjectMeta v1ObjectMeta = new V1ObjectMeta();
        v1ObjectMeta.setName(componentName);
        v1ObjectMeta.setNamespace(namespace);
        expectedV1alpha1Source.setMetadata(v1ObjectMeta);

        V1alpha1SourceSpec v1alpha1SourceSpec = new V1alpha1SourceSpec();
        v1alpha1SourceSpec.setClassName(className);
        v1alpha1SourceSpec.setSourceConfig(new HashMap<>());

        V1alpha1SourceSpecOutput v1alpha1SourceSpecOutput = new V1alpha1SourceSpecOutput();
        v1alpha1SourceSpecOutput.setTopic(topicName);
        v1alpha1SourceSpecOutput.setTypeClassName(sourceType);
        v1alpha1SourceSpec.setOutput(v1alpha1SourceSpecOutput);

        V1alpha1SourceSpecPulsar v1alpha1SourceSpecPulsar = new V1alpha1SourceSpecPulsar();
        v1alpha1SourceSpecPulsar.setPulsarConfig(componentName);
        v1alpha1SourceSpec.setPulsar(v1alpha1SourceSpecPulsar);

        V1alpha1SourceSpecResources v1alpha1SourceSpecResources = new V1alpha1SourceSpecResources();
        Map<String, String> requestRes = new HashMap<>();
        String cpuValue = cpu.toString();
        String memoryValue = ram.toString();
        requestRes.put(SourcesUtil.cpuKey, cpuValue);
        requestRes.put(SourcesUtil.memoryKey, memoryValue);

        v1alpha1SourceSpecResources.setLimits(requestRes);
        v1alpha1SourceSpecResources.setRequests(requestRes);
        v1alpha1SourceSpec.setResources(v1alpha1SourceSpecResources);

        v1alpha1SourceSpec.setReplicas(parallelism);
        v1alpha1SourceSpec.setMaxReplicas(parallelism);

        String location = String.format("%s/%s/%s", tenant, namespace, componentName);

        V1alpha1SourceSpecJava v1alpha1SourceSpecJava = new V1alpha1SourceSpecJava();
        v1alpha1SourceSpecJava.setJar(archive.split("/")[1]);
        v1alpha1SourceSpecJava.setJarLocation(location);
        v1alpha1SourceSpec.setJava(v1alpha1SourceSpecJava);

        v1alpha1SourceSpec.setClusterName(clusterName);

        expectedV1alpha1Source.setSpec(v1alpha1SourceSpec);

        Assert.assertEquals(expectedV1alpha1Source, v1alpha1Source);
    }

    @Test
    public void testCreateSourceConfigFromV1alpha1Source() {
        String tenant = "public";
        String namespace = "default";
        String componentName = "source-mongodb-sample";
        String className = "org.apache.pulsar.io.debezium.mongodb.DebeziumMongoDbSource";
        String sourceType = getClass().getName();
        String topicName = "persistent://public/default/destination";
        String archive = "pulsar-io-debezium-mongodb-2.7.0.nar";
        int parallelism = 1;
        Double cpu = 0.1;
        Long ram = 1L;
        String clusterName = "test-pulsar";
        String customRuntimeOptions = "{\"clusterName\":\"" + clusterName + "\"}";
        Map<String, Object> configs = new HashMap<>();
        configs.put("name","test-config");

        V1alpha1Source v1alpha1Source = new V1alpha1Source();

        v1alpha1Source.setKind(kind);
        v1alpha1Source.setApiVersion(String.format("%s/%s", group, version));

        V1ObjectMeta v1ObjectMeta = new V1ObjectMeta();
        v1ObjectMeta.setName(componentName);
        v1ObjectMeta.setNamespace(namespace);
        v1alpha1Source.setMetadata(v1ObjectMeta);

        V1alpha1SourceSpec v1alpha1SourceSpec = new V1alpha1SourceSpec();
        v1alpha1SourceSpec.setClassName(className);
        v1alpha1SourceSpec.setSourceConfig(new HashMap<>());

        V1alpha1SourceSpecOutput v1alpha1SourceSpecOutput = new V1alpha1SourceSpecOutput();
        v1alpha1SourceSpecOutput.setTopic(topicName);
        v1alpha1SourceSpecOutput.setTypeClassName(sourceType);
        v1alpha1SourceSpec.setOutput(v1alpha1SourceSpecOutput);

        V1alpha1SourceSpecPulsar v1alpha1SourceSpecPulsar = new V1alpha1SourceSpecPulsar();
        v1alpha1SourceSpecPulsar.setPulsarConfig(componentName);
        v1alpha1SourceSpec.setPulsar(v1alpha1SourceSpecPulsar);

        V1alpha1SourceSpecResources v1alpha1SourceSpecResources = new V1alpha1SourceSpecResources();
        Map<String, String> requestRes = new HashMap<>();
        String cpuValue = cpu.toString();
        String memoryValue = ram.toString();
        requestRes.put(SourcesUtil.cpuKey, cpuValue);
        requestRes.put(SourcesUtil.memoryKey, memoryValue);

        v1alpha1SourceSpecResources.setLimits(requestRes);
        v1alpha1SourceSpecResources.setRequests(requestRes);
        v1alpha1SourceSpec.setResources(v1alpha1SourceSpecResources);

        v1alpha1SourceSpec.setReplicas(parallelism);
        v1alpha1SourceSpec.setMaxReplicas(parallelism);

        String location = String.format("%s/%s/%s", tenant, namespace, componentName);

        V1alpha1SourceSpecJava v1alpha1SourceSpecJava = new V1alpha1SourceSpecJava();
        v1alpha1SourceSpecJava.setJar(archive);
        v1alpha1SourceSpecJava.setJarLocation(location);
        v1alpha1SourceSpec.setJava(v1alpha1SourceSpecJava);

        v1alpha1SourceSpec.setSourceConfig(SourcesUtil.transformedMapValueToString(configs));

        v1alpha1SourceSpec.setClusterName(clusterName);

        v1alpha1Source.setSpec(v1alpha1SourceSpec);

        SourceConfig sourceConfig = SourcesUtil.createSourceConfigFromV1alpha1Source(tenant, namespace, componentName
                , v1alpha1Source);

        SourceConfig expectedSourceConfig = new SourceConfig();
        expectedSourceConfig.setTenant(tenant);
        expectedSourceConfig.setNamespace(namespace);
        expectedSourceConfig.setName(componentName);
        expectedSourceConfig.setClassName(className);
        expectedSourceConfig.setTopicName(topicName);
        expectedSourceConfig.setArchive(archive);
        expectedSourceConfig.setParallelism(parallelism);
        expectedSourceConfig.setConfigs(configs);
        Resources resources = new Resources();
        resources.setRam(ram);
        resources.setCpu(cpu);
        expectedSourceConfig.setResources(resources);
        expectedSourceConfig.setCustomRuntimeOptions(customRuntimeOptions);

        Assert.assertEquals(expectedSourceConfig, sourceConfig);
    }
}
