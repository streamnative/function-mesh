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

import io.functionmesh.compute.MeshWorkerService;
import io.functionmesh.compute.models.FunctionMeshConnectorDefinition;
import io.functionmesh.compute.models.MeshWorkerServiceCustomConfig;
import io.functionmesh.compute.sources.models.V1alpha1Source;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpec;
import io.functionmesh.compute.testdata.Generate;
import io.functionmesh.compute.worker.MeshConnectorsManager;
import java.util.Collections;
import org.apache.commons.io.FileUtils;
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
        String typeClassName = "org.apache.pulsar.common.schema.KeyValue";
        String archive = "connectors/pulsar-io-debezium-mongodb-2.7.0.nar";
        int parallelism = 1;
        String clusterName = "test-pulsar";
        Map<String, Object> configs = new HashMap<>();
        String configsName = "test-sourceConfig";
        configs.put("name", configsName);
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

        SourceConfig sourceConfig = Generate.CreateSourceConfig(tenant, namespace, componentName);

        MeshWorkerService meshWorkerService =
                PowerMockito.mock(MeshWorkerService.class);
        PowerMockito.when(meshWorkerService.getMeshWorkerServiceCustomConfig()).thenReturn(new MeshWorkerServiceCustomConfig());
        V1alpha1Source v1alpha1Source = SourcesUtil.createV1alpha1SourceFromSourceConfig(kind, group, version,
                componentName, null, uploadedInputStream, sourceConfig, null,
                null, meshWorkerService);

        Assert.assertEquals(v1alpha1Source.getKind(), kind);
        V1alpha1SourceSpec v1alpha1SourceSpec = v1alpha1Source.getSpec();
        Assert.assertEquals(v1alpha1SourceSpec.getClassName(), className);
        Assert.assertEquals(v1alpha1SourceSpec.getReplicas().intValue(), parallelism);
        Assert.assertEquals(v1alpha1SourceSpec.getOutput().getTopic(), topicName);
        Assert.assertEquals(v1alpha1SourceSpec.getPulsar().getPulsarConfig(),
                CommonUtil.getPulsarClusterConfigMapName(clusterName));
        Assert.assertEquals(v1alpha1SourceSpec.getOutput().getTypeClassName(), typeClassName);
        Assert.assertEquals(v1alpha1SourceSpec.getJava().getJar(), archive);
        Assert.assertEquals(v1alpha1SourceSpec.getSourceConfig(), configs);
    }

    @Test
    public void testCreateSourceConfigFromV1alpha1Source() throws ClassNotFoundException, IOException, URISyntaxException {
        String tenant = "public";
        String namespace = "default";
        String componentName = "source-mongodb-sample";
        String className = "org.apache.pulsar.io.debezium.mongodb.DebeziumMongoDbSource";

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

        SourceConfig sourceConfig = Generate.CreateSourceConfig(tenant, namespace, componentName);

        MeshWorkerService meshWorkerService =
                PowerMockito.mock(MeshWorkerService.class);
        PowerMockito.when(meshWorkerService.getMeshWorkerServiceCustomConfig()).thenReturn(new MeshWorkerServiceCustomConfig());

        V1alpha1Source v1alpha1Source = SourcesUtil.createV1alpha1SourceFromSourceConfig(kind, group, version,
                componentName, null, uploadedInputStream, sourceConfig, null,
                null, meshWorkerService);

        SourceConfig newSourceConfig = SourcesUtil.createSourceConfigFromV1alpha1Source(tenant, namespace,
                componentName, v1alpha1Source);

        Assert.assertEquals(sourceConfig.getName(), newSourceConfig.getName());
        Assert.assertEquals(sourceConfig.getNamespace(), newSourceConfig.getNamespace());
        Assert.assertEquals(sourceConfig.getTenant(), newSourceConfig.getTenant());
        Assert.assertEquals(sourceConfig.getConfigs(), newSourceConfig.getConfigs());
        Assert.assertEquals(sourceConfig.getArchive(), newSourceConfig.getArchive());
        Assert.assertEquals(sourceConfig.getResources(), newSourceConfig.getResources());
        Assert.assertEquals(sourceConfig.getClassName(), newSourceConfig.getClassName());
        Assert.assertEquals(sourceConfig.getCustomRuntimeOptions(), newSourceConfig.getCustomRuntimeOptions());
        Assert.assertEquals(sourceConfig.getTopicName(), newSourceConfig.getTopicName());
        Assert.assertEquals(sourceConfig.getParallelism(), newSourceConfig.getParallelism());
        Assert.assertEquals(sourceConfig.getRuntimeFlags(), newSourceConfig.getRuntimeFlags());
    }

    @Test
    public void testCreateV1alpha1SourceFromSinkConfigWithBuiltin() throws IOException, URISyntaxException, ClassNotFoundException {
        String tenant = "public";
        String namespace = "default";
        String componentName = "source-mongodb-sample";
        String className = "org.apache.pulsar.io.debezium.mongodb.DebeziumMongoDbSource";
        String topicName = "persistent://public/default/destination";
        String typeClassName = "org.apache.pulsar.common.schema.KeyValue";
        String archive = "connectors/pulsar-io-debezium-mongodb-2.7.0.nar";
        int parallelism = 1;
        String clusterName = "test-pulsar";
        Map<String, Object> configs = new HashMap<>();
        String configsName = "test-sourceConfig";
        configs.put("name", configsName);

        MeshConnectorsManager connectorsManager = PowerMockito.mock(MeshConnectorsManager.class);
        FunctionMeshConnectorDefinition connectorDefinition = PowerMockito.mock(FunctionMeshConnectorDefinition.class);
        PowerMockito.when(connectorDefinition.getId()).thenReturn("debezium-mongodb");
        PowerMockito.when(connectorDefinition.getVersion()).thenReturn("2.7.0");
        PowerMockito.when(connectorDefinition.getImageTag()).thenReturn("2.7.0");
        PowerMockito.when(connectorDefinition.getImageRepository()).thenReturn("streamnative/pulsar-io-debezium-mongodb");
        PowerMockito.when(connectorDefinition.getJar()).thenReturn("connectors/pulsar-io-debezium-mongodb-2.7.0.nar");
        PowerMockito.when(connectorsManager.getConnectorDefinition("debezium-mongodb"))
                .thenReturn(connectorDefinition);

        SourceConfig sourceConfig = Generate.CreateSourceConfigBuiltin(tenant, namespace, componentName);

        MeshWorkerService meshWorkerService =
                PowerMockito.mock(MeshWorkerService.class);
        PowerMockito.when(meshWorkerService.getMeshWorkerServiceCustomConfig()).thenReturn(new MeshWorkerServiceCustomConfig());

        V1alpha1Source v1alpha1Source =
                SourcesUtil.createV1alpha1SourceFromSourceConfig(
                        kind, group, version, componentName, null, null, sourceConfig, connectorsManager,
                        null, meshWorkerService);

        Assert.assertEquals(v1alpha1Source.getKind(), kind);
        V1alpha1SourceSpec v1alpha1SourceSpec = v1alpha1Source.getSpec();
        Assert.assertEquals(v1alpha1SourceSpec.getClassName(), className);
        Assert.assertEquals(v1alpha1SourceSpec.getReplicas().intValue(), parallelism);
        Assert.assertEquals(v1alpha1SourceSpec.getOutput().getTopic(), topicName);
        Assert.assertEquals(v1alpha1SourceSpec.getPulsar().getPulsarConfig(),
                CommonUtil.getPulsarClusterConfigMapName(clusterName));
        Assert.assertEquals(v1alpha1SourceSpec.getOutput().getTypeClassName(), typeClassName);
        Assert.assertEquals(v1alpha1SourceSpec.getJava().getJar(), archive);
        Assert.assertEquals(v1alpha1SourceSpec.getSourceConfig(), configs);
    }
}
