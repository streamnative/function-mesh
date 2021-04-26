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
import io.functionmesh.compute.util.SourcesUtil;
import io.functionmesh.compute.sources.models.V1alpha1Source;
import io.kubernetes.client.openapi.ApiClient;
import io.kubernetes.client.openapi.ApiException;
import io.kubernetes.client.openapi.JSON;
import io.kubernetes.client.openapi.apis.CustomObjectsApi;
import okhttp3.Call;
import okhttp3.Response;
import okhttp3.ResponseBody;
import okhttp3.internal.http.RealResponseBody;
import org.apache.commons.io.FileUtils;
import org.apache.pulsar.common.functions.ProducerConfig;
import org.apache.pulsar.common.functions.Resources;
import org.apache.pulsar.common.io.ConnectorDefinition;
import org.apache.pulsar.common.io.SourceConfig;
import org.apache.pulsar.common.nar.NarClassLoader;
import org.apache.pulsar.common.policies.data.SourceStatus;
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
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.function.Supplier;

import static org.powermock.api.mockito.PowerMockito.spy;

@RunWith(PowerMockRunner.class)
@PrepareForTest({
    Response.class,
    RealResponseBody.class,
    FunctionCommon.class,
    ConnectorUtils.class,
    FileUtils.class
})
@PowerMockIgnore({"javax.management.*"})
public class SourcesImpTest {
    @Test
    public void testRegisterSource()
            throws ApiException, IOException, ClassNotFoundException, URISyntaxException {
        String testBody =
                "{\n"
                        + "    \"apiVersion\": \"compute.functionmesh.io/v1alpha1\",\n"
                        + "    \"kind\": \"Source\",\n"
                        + "    \"metadata\": {\n"
                        + "        \"annotations\": {\n"
                        + "            \"kubectl.kubernetes.io/last-applied-configuration\": \"{\\\"apiVersion\\\":\\\"compute.functionmesh.io/v1alpha1\\\",\\\"kind\\\":\\\"Source\\\",\\\"metadata\\\":{\\\"annotations\\\":{},\\\"name\\\":\\\"source-sample\\\",\\\"namespace\\\":\\\"default\\\"},\\\"spec\\\":{\\\"className\\\":\\\"org.apache.pulsar.io.debezium.mongodb.DebeziumMongoDbSource\\\",\\\"clusterName\\\":\\\"test-pulsar\\\",\\\"java\\\":{\\\"jar\\\":\\\"connectors/pulsar-io-debezium-mongodb-2.7.0-rc-pm-3.nar\\\",\\\"jarLocation\\\":\\\"\\\"},\\\"maxReplicas\\\":1,\\\"output\\\":{\\\"producerConf\\\":{\\\"maxPendingMessages\\\":1000,\\\"maxPendingMessagesAcrossPartitions\\\":50000,\\\"useThreadLocalProducers\\\":true},\\\"topic\\\":\\\"persistent://public/default/destination\\\"},\\\"pulsar\\\":{\\\"pulsarConfig\\\":\\\"test-source\\\"},\\\"replicas\\\":1,\\\"resources\\\":{\\\"limits\\\":{\\\"cpu\\\":\\\"0.2\\\",\\\"memory\\\":\\\"1.1G\\\"},\\\"requests\\\":{\\\"cpu\\\":\\\"0.1\\\",\\\"memory\\\":\\\"1G\\\"}},\\\"sinkType\\\":\\\"org.apache.pulsar.common.schema.KeyValue\\\",\\\"sourceConfig\\\":{\\\"database.whitelist\\\":\\\"inventory\\\",\\\"mongodb.hosts\\\":\\\"rs0/mongo-dbz-0.mongo.default.svc.cluster.local:27017,rs0/mongo-dbz-1.mongo.default.svc.cluster.local:27017,rs0/mongo-dbz-2.mongo.default.svc.cluster.local:27017\\\",\\\"mongodb.name\\\":\\\"dbserver1\\\",\\\"mongodb.password\\\":\\\"dbz\\\",\\\"mongodb.task.id\\\":\\\"1\\\",\\\"mongodb.user\\\":\\\"debezium\\\",\\\"pulsar.service.url\\\":\\\"pulsar://test-pulsar-broker.default.svc.cluster.local:6650\\\"},\\\"sourceType\\\":\\\"org.apache.pulsar.common.schema.KeyValue\\\"}}\\n\"\n"
                        + "        },\n"
                        + "        \"creationTimestamp\": \"2020-11-27T07:07:57Z\",\n"
                        + "        \"generation\": 1,\n"
                        + "        \"name\": \"source-sample\",\n"
                        + "        \"namespace\": \"default\",\n"
                        + "        \"resourceVersion\": \"881034\",\n"
                        + "        \"selfLink\": \"/apis/compute.functionmesh.io/v1alpha1/namespaces/default/sources/source-sample\",\n"
                        + "        \"uid\": \"8aed505e-38e4-4a8b-93f6-6f753dbf7ebc\"\n"
                        + "    },\n"
                        + "    \"spec\": {\n"
                        + "        \"className\": \"org.apache.pulsar.io.debezium.mongodb.DebeziumMongoDbSource\",\n"
                        + "        \"clusterName\": \"test-pulsar\",\n"
                        + "        \"java\": {\n"
                        + "            \"jar\": \"pulsar-io-debezium-mongodb-2.7.0-rc-pm-3.nar\",\n"
                        + "            \"jarLocation\": \"public/default/source-sample\"\n"
                        + "        },\n"
                        + "        \"maxReplicas\": 1,\n"
                        + "        \"output\": {\n"
                        + "            \"producerConf\": {\n"
                        + "                \"maxPendingMessages\": 1000,\n"
                        + "                \"maxPendingMessagesAcrossPartitions\": 50000,\n"
                        + "                \"useThreadLocalProducers\": true\n"
                        + "            },\n"
                        + "            \"topic\": \"persistent://public/default/destination\"\n"
                        + "        },\n"
                        + "        \"pulsar\": {\n"
                        + "            \"pulsarConfig\": \"test-source\"\n"
                        + "        },\n"
                        + "        \"replicas\": 1,\n"
                        + "        \"resources\": {\n"
                        + "            \"limits\": {\n"
                        + "                \"cpu\": \"0.1\",\n"
                        + "                \"memory\": \"1\"\n"
                        + "            },\n"
                        + "            \"requests\": {\n"
                        + "                \"cpu\": \"0.1\",\n"
                        + "                \"memory\": \"1\"\n"
                        + "            }\n"
                        + "        },\n"
                        + "        \"sinkType\": \"org.apache.pulsar.common.schema.KeyValue\",\n"
                        + "        \"sourceConfig\": {\n"
                        + "            \"name\": \"test-sourceConfig\""
                        + "        },\n"
                        + "        \"sourceType\": \"org.apache.pulsar.common.schema.KeyValue\"\n"
                        + "    },\n"
                        + "    \"status\": {\n"
                        + "        \"conditions\": {\n"
                        + "            \"HorizontalPodAutoscaler\": {\n"
                        + "                \"action\": \"NoAction\",\n"
                        + "                \"condition\": \"HPAReady\",\n"
                        + "                \"status\": \"True\"\n"
                        + "            },\n"
                        + "            \"Service\": {\n"
                        + "                \"action\": \"NoAction\",\n"
                        + "                \"condition\": \"ServiceReady\",\n"
                        + "                \"status\": \"True\"\n"
                        + "            },\n"
                        + "            \"StatefulSet\": {\n"
                        + "                \"action\": \"NoAction\",\n"
                        + "                \"condition\": \"StatefulSetReady\",\n"
                        + "                \"status\": \"True\"\n"
                        + "            }\n"
                        + "        },\n"
                        + "        \"replicas\": 1,\n"
                        + "        \"selector\": \"component=source,name=source-sample,namespace=default\"\n"
                        + "    }\n"
                        + "}";
        MeshWorkerService meshWorkerService =
                PowerMockito.mock(MeshWorkerService.class);
        Supplier<MeshWorkerService> meshWorkerServiceSupplier =
                () -> meshWorkerService;
        CustomObjectsApi customObjectsApi = PowerMockito.mock(CustomObjectsApi.class);
        PowerMockito.when(meshWorkerService.getCustomObjectsApi())
                .thenReturn(customObjectsApi);
        Call call = PowerMockito.mock(Call.class);
        Response response = PowerMockito.mock(Response.class);
        ResponseBody responseBody = PowerMockito.mock(RealResponseBody.class);
        ApiClient apiClient = PowerMockito.mock(ApiClient.class);

        String group = "compute.functionmesh.io";
        String plural = "sources";
        String version = "v1alpha1";
        String kind = "Source";
        String tenant = "public";
        String namespace = "default";
        String componentName = "source-mongodb-sample";
        String className = "org.apache.pulsar.io.debezium.mongodb.DebeziumMongoDbSource";
        String topicName = "persistent://public/default/destination";
        String archive = "connectors/pulsar-io-debezium-mongodb-2.7.0.nar";
        int parallelism = 1;
        Double cpu = 0.1;
        Long ram = 1L;
        String clusterName = "test-pulsar";
        String customRuntimeOptions = "{\"clusterName\": \"" + clusterName + "\"}";
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
        PowerMockito.when(FunctionCommon.extractNarClassLoader(narFile, null))
                .thenReturn(narClassLoader);
        PowerMockito.when(FunctionCommon.createPkgTempFile()).thenReturn(narFile);
        PowerMockito.when(ConnectorUtils.getIOSourceClass(narClassLoader)).thenReturn(className);
        PowerMockito.<Class<?>>when(FunctionCommon.getSourceType(null)).thenReturn(getClass());

        SourceConfig sourceConfig = new SourceConfig();
        sourceConfig.setTenant(tenant);
        sourceConfig.setNamespace(namespace);
        sourceConfig.setName(componentName);
        sourceConfig.setTopicName(topicName);
        sourceConfig.setArchive(archive);
        sourceConfig.setParallelism(parallelism);
        sourceConfig.setConfigs(configs);
        Resources resources = new Resources();
        resources.setRam(ram);
        resources.setCpu(cpu);
        sourceConfig.setResources(resources);
        sourceConfig.setCustomRuntimeOptions(customRuntimeOptions);

        V1alpha1Source v1alpha1Source =
                SourcesUtil.createV1alpha1SourceFromSourceConfig(
                        kind,
                        group,
                        version,
                        componentName,
                        null,
                        uploadedInputStream,
                        sourceConfig);

        PowerMockito.when(
                meshWorkerService
                                .getCustomObjectsApi()
                                .createNamespacedCustomObjectCall(
                                        group,
                                        version,
                                        namespace,
                                        plural,
                                        v1alpha1Source,
                                        null,
                                        null,
                                        null,
                                        null))
                .thenReturn(call);
        PowerMockito.when(call.execute()).thenReturn(response);
        PowerMockito.when(response.isSuccessful()).thenReturn(true);
        PowerMockito.when(response.body()).thenReturn(responseBody);
        PowerMockito.when(responseBody.string()).thenReturn(testBody);
        PowerMockito.when(meshWorkerService.getApiClient()).thenReturn(apiClient);
        JSON json = new JSON();
        PowerMockito.when(apiClient.getJSON()).thenReturn(json);

        SourcesImpl sources = spy(new SourcesImpl(meshWorkerServiceSupplier));
        try {
            sources.registerSource(
                    tenant,
                    namespace,
                    componentName,
                    uploadedInputStream,
                    null,
                    null,
                    sourceConfig,
                    null,
                    null);
        } catch (Exception exception) {
            Assert.fail("No exception, but got error message:" + exception.getMessage());
        }
    }

    @Test
    public void testUpdateSource()
            throws ApiException, IOException, ClassNotFoundException, URISyntaxException {
        String getBody =
                "{\n"
                        + "    \"apiVersion\": \"compute.functionmesh.io/v1alpha1\",\n"
                        + "    \"kind\": \"Source\",\n"
                        + "    \"metadata\": {\n"
                        + "        \"resourceVersion\": \"881033\""
                        + "    }\n"
                        + "}";

        String replaceBody =
                "{\n"
                        + "    \"apiVersion\": \"compute.functionmesh.io/v1alpha1\",\n"
                        + "    \"kind\": \"Source\",\n"
                        + "    \"metadata\": {\n"
                        + "        \"annotations\": {\n"
                        + "            \"kubectl.kubernetes.io/last-applied-configuration\": \"{\\\"apiVersion\\\":\\\"compute.functionmesh.io/v1alpha1\\\",\\\"kind\\\":\\\"Source\\\",\\\"metadata\\\":{\\\"annotations\\\":{},\\\"name\\\":\\\"source-sample\\\",\\\"namespace\\\":\\\"default\\\"},\\\"spec\\\":{\\\"className\\\":\\\"org.apache.pulsar.io.debezium.mongodb.DebeziumMongoDbSource\\\",\\\"clusterName\\\":\\\"test-pulsar\\\",\\\"java\\\":{\\\"jar\\\":\\\"connectors/pulsar-io-debezium-mongodb-2.7.0-rc-pm-3.nar\\\",\\\"jarLocation\\\":\\\"\\\"},\\\"maxReplicas\\\":1,\\\"output\\\":{\\\"producerConf\\\":{\\\"maxPendingMessages\\\":1000,\\\"maxPendingMessagesAcrossPartitions\\\":50000,\\\"useThreadLocalProducers\\\":true},\\\"topic\\\":\\\"persistent://public/default/destination\\\"},\\\"pulsar\\\":{\\\"pulsarConfig\\\":\\\"test-source\\\"},\\\"replicas\\\":1,\\\"resources\\\":{\\\"limits\\\":{\\\"cpu\\\":\\\"0.2\\\",\\\"memory\\\":\\\"1.1G\\\"},\\\"requests\\\":{\\\"cpu\\\":\\\"0.1\\\",\\\"memory\\\":\\\"1G\\\"}},\\\"sinkType\\\":\\\"org.apache.pulsar.common.schema.KeyValue\\\",\\\"sourceConfig\\\":{\\\"database.whitelist\\\":\\\"inventory\\\",\\\"mongodb.hosts\\\":\\\"rs0/mongo-dbz-0.mongo.default.svc.cluster.local:27017,rs0/mongo-dbz-1.mongo.default.svc.cluster.local:27017,rs0/mongo-dbz-2.mongo.default.svc.cluster.local:27017\\\",\\\"mongodb.name\\\":\\\"dbserver1\\\",\\\"mongodb.password\\\":\\\"dbz\\\",\\\"mongodb.task.id\\\":\\\"1\\\",\\\"mongodb.user\\\":\\\"debezium\\\",\\\"pulsar.service.url\\\":\\\"pulsar://test-pulsar-broker.default.svc.cluster.local:6650\\\"},\\\"sourceType\\\":\\\"org.apache.pulsar.common.schema.KeyValue\\\"}}\\n\"\n"
                        + "        },\n"
                        + "        \"creationTimestamp\": \"2020-11-27T07:07:57Z\",\n"
                        + "        \"generation\": 1,\n"
                        + "        \"name\": \"source-sample\",\n"
                        + "        \"namespace\": \"default\",\n"
                        + "        \"resourceVersion\": \"881034\",\n"
                        + "        \"selfLink\": \"/apis/compute.functionmesh.io/v1alpha1/namespaces/default/sources/source-sample\",\n"
                        + "        \"uid\": \"8aed505e-38e4-4a8b-93f6-6f753dbf7ebc\"\n"
                        + "    },\n"
                        + "    \"spec\": {\n"
                        + "        \"className\": \"org.apache.pulsar.io.debezium.mongodb.DebeziumMongoDbSource\",\n"
                        + "        \"clusterName\": \"test-pulsar\",\n"
                        + "        \"java\": {\n"
                        + "            \"jar\": \"pulsar-io-debezium-mongodb-2.7.0-rc-pm-3.nar\",\n"
                        + "            \"jarLocation\": \"public/default/source-sample\"\n"
                        + "        },\n"
                        + "        \"maxReplicas\": 1,\n"
                        + "        \"output\": {\n"
                        + "            \"producerConf\": {\n"
                        + "                \"maxPendingMessages\": 1000,\n"
                        + "                \"maxPendingMessagesAcrossPartitions\": 50000,\n"
                        + "                \"useThreadLocalProducers\": true\n"
                        + "            },\n"
                        + "            \"topic\": \"persistent://public/default/destination\"\n"
                        + "        },\n"
                        + "        \"pulsar\": {\n"
                        + "            \"pulsarConfig\": \"test-source\"\n"
                        + "        },\n"
                        + "        \"replicas\": 1,\n"
                        + "        \"resources\": {\n"
                        + "            \"limits\": {\n"
                        + "                \"cpu\": \"0.1\",\n"
                        + "                \"memory\": \"1\"\n"
                        + "            },\n"
                        + "            \"requests\": {\n"
                        + "                \"cpu\": \"0.1\",\n"
                        + "                \"memory\": \"1\"\n"
                        + "            }\n"
                        + "        },\n"
                        + "        \"sinkType\": \"org.apache.pulsar.common.schema.KeyValue\",\n"
                        + "        \"sourceConfig\": {\n"
                        + "            \"name\": \"test-sourceConfig\""
                        + "        },\n"
                        + "        \"sourceType\": \"org.apache.pulsar.common.schema.KeyValue\"\n"
                        + "    },\n"
                        + "    \"status\": {\n"
                        + "        \"conditions\": {\n"
                        + "            \"HorizontalPodAutoscaler\": {\n"
                        + "                \"action\": \"NoAction\",\n"
                        + "                \"condition\": \"HPAReady\",\n"
                        + "                \"status\": \"True\"\n"
                        + "            },\n"
                        + "            \"Service\": {\n"
                        + "                \"action\": \"NoAction\",\n"
                        + "                \"condition\": \"ServiceReady\",\n"
                        + "                \"status\": \"True\"\n"
                        + "            },\n"
                        + "            \"StatefulSet\": {\n"
                        + "                \"action\": \"NoAction\",\n"
                        + "                \"condition\": \"StatefulSetReady\",\n"
                        + "                \"status\": \"True\"\n"
                        + "            }\n"
                        + "        },\n"
                        + "        \"replicas\": 1,\n"
                        + "        \"selector\": \"component=source,name=source-sample,namespace=default\"\n"
                        + "    }\n"
                        + "}";

        String group = "compute.functionmesh.io";
        String plural = "sources";
        String version = "v1alpha1";
        String kind = "Source";
        String tenant = "public";
        String namespace = "default";
        String componentName = "source-mongodb-sample";
        String className = "org.apache.pulsar.io.debezium.mongodb.DebeziumMongoDbSource";
        String topicName = "persistent://public/default/destination";
        String archive = "connectors/pulsar-io-debezium-mongodb-2.7.0.nar";
        int parallelism = 1;
        Double cpu = 0.1;
        Long ram = 1L;
        String clusterName = "test-pulsar";
        String customRuntimeOptions = "{\"clusterName\": \"" + clusterName + "\"}";
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
        PowerMockito.when(FunctionCommon.extractNarClassLoader(narFile, null))
                .thenReturn(narClassLoader);
        PowerMockito.when(FunctionCommon.createPkgTempFile()).thenReturn(narFile);
        PowerMockito.when(ConnectorUtils.getIOSourceClass(narClassLoader)).thenReturn(className);
        PowerMockito.<Class<?>>when(FunctionCommon.getSourceType(null)).thenReturn(getClass());

        SourceConfig sourceConfig = new SourceConfig();
        sourceConfig.setTenant(tenant);
        sourceConfig.setNamespace(namespace);
        sourceConfig.setName(componentName);
        sourceConfig.setTopicName(topicName);
        sourceConfig.setArchive(archive);
        sourceConfig.setParallelism(parallelism);
        sourceConfig.setConfigs(configs);
        Resources resources = new Resources();
        resources.setRam(ram);
        resources.setCpu(cpu);
        sourceConfig.setResources(resources);
        sourceConfig.setCustomRuntimeOptions(customRuntimeOptions);

        MeshWorkerService meshWorkerService =
                PowerMockito.mock(MeshWorkerService.class);
        Supplier<MeshWorkerService> meshWorkerServiceSupplier =
                () -> meshWorkerService;
        CustomObjectsApi customObjectsApi = PowerMockito.mock(CustomObjectsApi.class);
        PowerMockito.when(meshWorkerService.getCustomObjectsApi())
                .thenReturn(customObjectsApi);

        Call getCall = PowerMockito.mock(Call.class);
        Response getResponse = PowerMockito.mock(Response.class);
        ResponseBody getResponseBody = PowerMockito.mock(RealResponseBody.class);
        PowerMockito.when(getCall.execute()).thenReturn(getResponse);
        PowerMockito.when(getResponse.isSuccessful()).thenReturn(true);
        PowerMockito.when(getResponse.body()).thenReturn(getResponseBody);
        PowerMockito.when(getResponseBody.string()).thenReturn(getBody);

        Call replaceCall = PowerMockito.mock(Call.class);
        Response replaceResponse = PowerMockito.mock(Response.class);
        ResponseBody replaceResponseBody = PowerMockito.mock(RealResponseBody.class);

        PowerMockito.when(replaceCall.execute()).thenReturn(replaceResponse);
        PowerMockito.when(replaceResponse.isSuccessful()).thenReturn(true);
        PowerMockito.when(replaceResponse.body()).thenReturn(getResponseBody);
        PowerMockito.when(replaceResponseBody.string()).thenReturn(replaceBody);

        ApiClient apiClient = PowerMockito.mock(ApiClient.class);
        PowerMockito.when(meshWorkerService.getApiClient()).thenReturn(apiClient);
        JSON json = new JSON();
        PowerMockito.when(apiClient.getJSON()).thenReturn(json);

        PowerMockito.when(
                meshWorkerService
                                .getCustomObjectsApi()
                                .getNamespacedCustomObjectCall(
                                        group, version, namespace, plural, componentName, null))
                .thenReturn(getCall);

        V1alpha1Source v1alpha1Source =
                SourcesUtil.createV1alpha1SourceFromSourceConfig(
                        kind,
                        group,
                        version,
                        componentName,
                        null,
                        uploadedInputStream,
                        sourceConfig);
        v1alpha1Source.getMetadata().setResourceVersion("881033");

        PowerMockito.when(
                meshWorkerService
                                .getCustomObjectsApi()
                                .replaceNamespacedCustomObjectCall(
                                        group,
                                        version,
                                        namespace,
                                        plural,
                                        componentName,
                                        v1alpha1Source,
                                        null,
                                        null,
                                        null))
                .thenReturn(getCall);

        SourcesImpl sources = spy(new SourcesImpl(meshWorkerServiceSupplier));

        try {
            sources.updateSource(
                    tenant,
                    namespace,
                    componentName,
                    uploadedInputStream,
                    null,
                    null,
                    sourceConfig,
                    null,
                    null,
                    null);
        } catch (Exception exception) {
            Assert.fail("Expected no exception to be thrown but got" + exception.getMessage());
        }
    }

    @Test
    public void testGetSourceStatus()
            throws ClassNotFoundException, IOException, URISyntaxException, ApiException {
        String testBody =
                "{\n"
                        + "    \"apiVersion\": \"compute.functionmesh.io/v1alpha1\",\n"
                        + "    \"kind\": \"Source\",\n"
                        + "    \"metadata\": {\n"
                        + "        \"annotations\": {\n"
                        + "            \"kubectl.kubernetes.io/last-applied-configuration\": \"{\\\"apiVersion\\\":\\\"compute.functionmesh.io/v1alpha1\\\",\\\"kind\\\":\\\"Source\\\",\\\"metadata\\\":{\\\"annotations\\\":{},\\\"name\\\":\\\"source-sample\\\",\\\"namespace\\\":\\\"default\\\"},\\\"spec\\\":{\\\"className\\\":\\\"org.apache.pulsar.io.debezium.mongodb.DebeziumMongoDbSource\\\",\\\"clusterName\\\":\\\"test-pulsar\\\",\\\"java\\\":{\\\"jar\\\":\\\"connectors/pulsar-io-debezium-mongodb-2.7.0-rc-pm-3.nar\\\",\\\"jarLocation\\\":\\\"\\\"},\\\"maxReplicas\\\":1,\\\"output\\\":{\\\"producerConf\\\":{\\\"maxPendingMessages\\\":1000,\\\"maxPendingMessagesAcrossPartitions\\\":50000,\\\"useThreadLocalProducers\\\":true},\\\"topic\\\":\\\"persistent://public/default/destination\\\"},\\\"pulsar\\\":{\\\"pulsarConfig\\\":\\\"test-source\\\"},\\\"replicas\\\":1,\\\"resources\\\":{\\\"limits\\\":{\\\"cpu\\\":\\\"0.2\\\",\\\"memory\\\":\\\"1.1G\\\"},\\\"requests\\\":{\\\"cpu\\\":\\\"0.1\\\",\\\"memory\\\":\\\"1G\\\"}},\\\"sinkType\\\":\\\"org.apache.pulsar.common.schema.KeyValue\\\",\\\"sourceConfig\\\":{\\\"database.whitelist\\\":\\\"inventory\\\",\\\"mongodb.hosts\\\":\\\"rs0/mongo-dbz-0.mongo.default.svc.cluster.local:27017,rs0/mongo-dbz-1.mongo.default.svc.cluster.local:27017,rs0/mongo-dbz-2.mongo.default.svc.cluster.local:27017\\\",\\\"mongodb.name\\\":\\\"dbserver1\\\",\\\"mongodb.password\\\":\\\"dbz\\\",\\\"mongodb.task.id\\\":\\\"1\\\",\\\"mongodb.user\\\":\\\"debezium\\\",\\\"pulsar.service.url\\\":\\\"pulsar://test-pulsar-broker.default.svc.cluster.local:6650\\\"},\\\"sourceType\\\":\\\"org.apache.pulsar.common.schema.KeyValue\\\"}}\\n\"\n"
                        + "        },\n"
                        + "        \"creationTimestamp\": \"2020-11-27T07:07:57Z\",\n"
                        + "        \"generation\": 1,\n"
                        + "        \"name\": \"source-mongodb-sample\",\n"
                        + "        \"namespace\": \"default\",\n"
                        + "        \"resourceVersion\": \"881034\",\n"
                        + "        \"selfLink\": \"/apis/compute.functionmesh.io/v1alpha1/namespaces/default/sources/source-sample\",\n"
                        + "        \"uid\": \"8aed505e-38e4-4a8b-93f6-6f753dbf7ebc\"\n"
                        + "    },\n"
                        + "    \"spec\": {\n"
                        + "        \"className\": \"org.apache.pulsar.io.debezium.mongodb.DebeziumMongoDbSource\",\n"
                        + "        \"clusterName\": \"test-pulsar\",\n"
                        + "        \"java\": {\n"
                        + "            \"jar\": \"pulsar-io-debezium-mongodb-2.7.0-rc-pm-3.nar\",\n"
                        + "            \"jarLocation\": \"public/default/source-sample\"\n"
                        + "        },\n"
                        + "        \"maxReplicas\": 1,\n"
                        + "        \"output\": {\n"
                        + "            \"producerConf\": {\n"
                        + "                \"maxPendingMessages\": 1000,\n"
                        + "                \"maxPendingMessagesAcrossPartitions\": 50000,\n"
                        + "                \"useThreadLocalProducers\": true\n"
                        + "            },\n"
                        + "            \"topic\": \"persistent://public/default/destination\"\n"
                        + "        },\n"
                        + "        \"pulsar\": {\n"
                        + "            \"pulsarConfig\": \"test-source\"\n"
                        + "        },\n"
                        + "        \"replicas\": 1,\n"
                        + "        \"resources\": {\n"
                        + "            \"limits\": {\n"
                        + "                \"cpu\": \"0.1\",\n"
                        + "                \"memory\": \"1\"\n"
                        + "            },\n"
                        + "            \"requests\": {\n"
                        + "                \"cpu\": \"0.1\",\n"
                        + "                \"memory\": \"1\"\n"
                        + "            }\n"
                        + "        },\n"
                        + "        \"sinkType\": \"org.apache.pulsar.common.schema.KeyValue\",\n"
                        + "        \"sourceConfig\": {\n"
                        + "            \"name\": \"test-sourceConfig\""
                        + "        },\n"
                        + "        \"sourceType\": \"org.apache.pulsar.common.schema.KeyValue\"\n"
                        + "    },\n"
                        + "    \"status\": {\n"
                        + "        \"conditions\": {\n"
                        + "            \"HorizontalPodAutoscaler\": {\n"
                        + "                \"action\": \"NoAction\",\n"
                        + "                \"condition\": \"HPAReady\",\n"
                        + "                \"status\": \"True\"\n"
                        + "            },\n"
                        + "            \"Service\": {\n"
                        + "                \"action\": \"NoAction\",\n"
                        + "                \"condition\": \"ServiceReady\",\n"
                        + "                \"status\": \"True\"\n"
                        + "            },\n"
                        + "            \"StatefulSet\": {\n"
                        + "                \"action\": \"NoAction\",\n"
                        + "                \"condition\": \"StatefulSetReady\",\n"
                        + "                \"status\": \"True\"\n"
                        + "            }\n"
                        + "        },\n"
                        + "        \"replicas\": 1,\n"
                        + "        \"selector\": \"component=source,name=source-sample,namespace=default\"\n"
                        + "    }\n"
                        + "}";
        MeshWorkerService meshWorkerService =
                PowerMockito.mock(MeshWorkerService.class);
        Supplier<MeshWorkerService> meshWorkerServiceSupplier =
                () -> meshWorkerService;
        CustomObjectsApi customObjectsApi = PowerMockito.mock(CustomObjectsApi.class);
        PowerMockito.when(meshWorkerService.getCustomObjectsApi())
                .thenReturn(customObjectsApi);
        Call call = PowerMockito.mock(Call.class);
        Response response = PowerMockito.mock(Response.class);
        ResponseBody responseBody = PowerMockito.mock(RealResponseBody.class);
        ApiClient apiClient = PowerMockito.mock(ApiClient.class);

        String group = "compute.functionmesh.io";
        String plural = "sources";
        String version = "v1alpha1";
        String tenant = "public";
        String namespace = "default";
        String componentName = "source-mongodb-sample";

        PowerMockito.when(
                meshWorkerService
                                .getCustomObjectsApi()
                                .getNamespacedCustomObjectCall(
                                        group, version, namespace, plural, componentName, null))
                .thenReturn(call);
        PowerMockito.when(call.execute()).thenReturn(response);
        PowerMockito.when(response.isSuccessful()).thenReturn(true);
        PowerMockito.when(response.body()).thenReturn(responseBody);
        PowerMockito.when(responseBody.string()).thenReturn(testBody);
        PowerMockito.when(meshWorkerService.getApiClient()).thenReturn(apiClient);
        JSON json = new JSON();
        PowerMockito.when(apiClient.getJSON()).thenReturn(json);

        SourcesImpl sources = spy(new SourcesImpl(meshWorkerServiceSupplier));
        SourceStatus sourceStatus =
                sources.getSourceStatus(tenant, namespace, componentName, null, null, null);

        SourceStatus expectedSourceStatus = new SourceStatus();
        SourceStatus.SourceInstanceStatus expectedSourceInstanceStatus =
                new SourceStatus.SourceInstanceStatus();
        SourceStatus.SourceInstanceStatus.SourceInstanceStatusData
                expectedSourceInstanceStatusData =
                        new SourceStatus.SourceInstanceStatus.SourceInstanceStatusData();
        expectedSourceInstanceStatusData.setRunning(true);
        expectedSourceInstanceStatusData.setWorkerId("test-pulsar");
        expectedSourceInstanceStatus.setStatus(expectedSourceInstanceStatusData);
        expectedSourceStatus.addInstance(expectedSourceInstanceStatus);
        expectedSourceStatus.setNumInstances(expectedSourceStatus.getInstances().size());

        Assert.assertEquals(expectedSourceStatus, sourceStatus);
    }

    @Test
    public void testGetSourceInfo()
            throws ApiException, IOException, ClassNotFoundException, URISyntaxException {
        String testBody =
                "{\n"
                        + "    \"apiVersion\": \"compute.functionmesh.io/v1alpha1\",\n"
                        + "    \"kind\": \"Source\",\n"
                        + "    \"metadata\": {\n"
                        + "        \"annotations\": {\n"
                        + "            \"kubectl.kubernetes.io/last-applied-configuration\": \"{\\\"apiVersion\\\":\\\"compute.functionmesh.io/v1alpha1\\\",\\\"kind\\\":\\\"Source\\\",\\\"metadata\\\":{\\\"annotations\\\":{},\\\"name\\\":\\\"source-sample\\\",\\\"namespace\\\":\\\"default\\\"},\\\"spec\\\":{\\\"className\\\":\\\"org.apache.pulsar.io.debezium.mongodb.DebeziumMongoDbSource\\\",\\\"clusterName\\\":\\\"test-pulsar\\\",\\\"java\\\":{\\\"jar\\\":\\\"connectors/pulsar-io-debezium-mongodb-2.7.0-rc-pm-3.nar\\\",\\\"jarLocation\\\":\\\"\\\"},\\\"maxReplicas\\\":1,\\\"output\\\":{\\\"producerConf\\\":{\\\"maxPendingMessages\\\":1000,\\\"maxPendingMessagesAcrossPartitions\\\":50000,\\\"useThreadLocalProducers\\\":true},\\\"topic\\\":\\\"persistent://public/default/destination\\\"},\\\"pulsar\\\":{\\\"pulsarConfig\\\":\\\"test-source\\\"},\\\"replicas\\\":1,\\\"resources\\\":{\\\"limits\\\":{\\\"cpu\\\":\\\"0.2\\\",\\\"memory\\\":\\\"1.1G\\\"},\\\"requests\\\":{\\\"cpu\\\":\\\"0.1\\\",\\\"memory\\\":\\\"1G\\\"}},\\\"sinkType\\\":\\\"org.apache.pulsar.common.schema.KeyValue\\\",\\\"sourceConfig\\\":{\\\"database.whitelist\\\":\\\"inventory\\\",\\\"mongodb.hosts\\\":\\\"rs0/mongo-dbz-0.mongo.default.svc.cluster.local:27017,rs0/mongo-dbz-1.mongo.default.svc.cluster.local:27017,rs0/mongo-dbz-2.mongo.default.svc.cluster.local:27017\\\",\\\"mongodb.name\\\":\\\"dbserver1\\\",\\\"mongodb.password\\\":\\\"dbz\\\",\\\"mongodb.task.id\\\":\\\"1\\\",\\\"mongodb.user\\\":\\\"debezium\\\",\\\"pulsar.service.url\\\":\\\"pulsar://test-pulsar-broker.default.svc.cluster.local:6650\\\"},\\\"sourceType\\\":\\\"org.apache.pulsar.common.schema.KeyValue\\\"}}\\n\"\n"
                        + "        },\n"
                        + "        \"creationTimestamp\": \"2020-11-27T07:07:57Z\",\n"
                        + "        \"generation\": 1,\n"
                        + "        \"name\": \"source-sample\",\n"
                        + "        \"namespace\": \"default\",\n"
                        + "        \"resourceVersion\": \"881034\",\n"
                        + "        \"selfLink\": \"/apis/compute.functionmesh.io/v1alpha1/namespaces/default/sources/source-sample\",\n"
                        + "        \"uid\": \"8aed505e-38e4-4a8b-93f6-6f753dbf7ebc\"\n"
                        + "    },\n"
                        + "    \"spec\": {\n"
                        + "        \"className\": \"org.apache.pulsar.io.debezium.mongodb.DebeziumMongoDbSource\",\n"
                        + "        \"clusterName\": \"test-pulsar\",\n"
                        + "        \"java\": {\n"
                        + "            \"jar\": \"pulsar-io-debezium-mongodb-2.7.0-rc-pm-3.nar\",\n"
                        + "            \"jarLocation\": \"public/default/source-sample\"\n"
                        + "        },\n"
                        + "        \"maxReplicas\": 1,\n"
                        + "        \"output\": {\n"
                        + "            \"producerConf\": {\n"
                        + "                \"maxPendingMessages\": 1000,\n"
                        + "                \"maxPendingMessagesAcrossPartitions\": 50000,\n"
                        + "                \"useThreadLocalProducers\": true\n"
                        + "            },\n"
                        + "            \"topic\": \"persistent://public/default/destination\"\n"
                        + "        },\n"
                        + "        \"pulsar\": {\n"
                        + "            \"pulsarConfig\": \"test-source\"\n"
                        + "        },\n"
                        + "        \"replicas\": 1,\n"
                        + "        \"resources\": {\n"
                        + "            \"limits\": {\n"
                        + "                \"cpu\": \"0.1\",\n"
                        + "                \"memory\": \"1\"\n"
                        + "            },\n"
                        + "            \"requests\": {\n"
                        + "                \"cpu\": \"0.1\",\n"
                        + "                \"memory\": \"1\"\n"
                        + "            }\n"
                        + "        },\n"
                        + "        \"sinkType\": \"org.apache.pulsar.common.schema.KeyValue\",\n"
                        + "        \"sourceConfig\": {\n"
                        + "            \"name\": \"test-sourceConfig\""
                        + "        },\n"
                        + "        \"sourceType\": \"org.apache.pulsar.common.schema.KeyValue\"\n"
                        + "    },\n"
                        + "    \"status\": {\n"
                        + "        \"conditions\": {\n"
                        + "            \"HorizontalPodAutoscaler\": {\n"
                        + "                \"action\": \"NoAction\",\n"
                        + "                \"condition\": \"HPAReady\",\n"
                        + "                \"status\": \"True\"\n"
                        + "            },\n"
                        + "            \"Service\": {\n"
                        + "                \"action\": \"NoAction\",\n"
                        + "                \"condition\": \"ServiceReady\",\n"
                        + "                \"status\": \"True\"\n"
                        + "            },\n"
                        + "            \"StatefulSet\": {\n"
                        + "                \"action\": \"NoAction\",\n"
                        + "                \"condition\": \"StatefulSetReady\",\n"
                        + "                \"status\": \"True\"\n"
                        + "            }\n"
                        + "        },\n"
                        + "        \"replicas\": 1,\n"
                        + "        \"selector\": \"component=source,name=source-sample,namespace=default\"\n"
                        + "    }\n"
                        + "}";
        MeshWorkerService meshWorkerService =
                PowerMockito.mock(MeshWorkerService.class);
        Supplier<MeshWorkerService> meshWorkerServiceSupplier =
                () -> meshWorkerService;
        CustomObjectsApi customObjectsApi = PowerMockito.mock(CustomObjectsApi.class);
        PowerMockito.when(meshWorkerService.getCustomObjectsApi())
                .thenReturn(customObjectsApi);
        Call call = PowerMockito.mock(Call.class);
        Response response = PowerMockito.mock(Response.class);
        ResponseBody responseBody = PowerMockito.mock(RealResponseBody.class);
        ApiClient apiClient = PowerMockito.mock(ApiClient.class);

        String group = "compute.functionmesh.io";
        String plural = "sources";
        String version = "v1alpha1";
        String tenant = "public";
        String namespace = "default";
        String componentName = "source-mongodb-sample";

        PowerMockito.when(
                meshWorkerService
                                .getCustomObjectsApi()
                                .getNamespacedCustomObjectCall(
                                        group, version, namespace, plural, componentName, null))
                .thenReturn(call);
        PowerMockito.when(call.execute()).thenReturn(response);
        PowerMockito.when(response.isSuccessful()).thenReturn(true);
        PowerMockito.when(response.body()).thenReturn(responseBody);
        PowerMockito.when(responseBody.string()).thenReturn(testBody);
        PowerMockito.when(meshWorkerService.getApiClient()).thenReturn(apiClient);
        JSON json = new JSON();
        PowerMockito.when(apiClient.getJSON()).thenReturn(json);

        String className = "org.apache.pulsar.io.debezium.mongodb.DebeziumMongoDbSource";
        String topicName = "persistent://public/default/destination";
        String archive = "pulsar-io-debezium-mongodb-2.7.0-rc-pm-3.nar";
        int parallelism = 1;
        Double cpu = 0.1;
        Long ram = 1L;
        String clusterName = "test-pulsar";
        String customRuntimeOptions = "{\"clusterName\":\"" + clusterName + "\"}";
        Map<String, Object> configs = new HashMap<>();
        String configsName = "test-sourceConfig";
        configs.put("name", configsName);
        SourceConfig expectedSourceConfig = new SourceConfig();
        expectedSourceConfig.setTenant(tenant);
        expectedSourceConfig.setNamespace(namespace);
        expectedSourceConfig.setName(componentName);
        expectedSourceConfig.setClassName(className);
        expectedSourceConfig.setTopicName(topicName);
        expectedSourceConfig.setArchive(archive);
        expectedSourceConfig.setParallelism(parallelism);
        expectedSourceConfig.setConfigs(configs);

        ProducerConfig producerConfig = new ProducerConfig();
        producerConfig.setMaxPendingMessages(1000);
        producerConfig.setMaxPendingMessagesAcrossPartitions(50000);
        producerConfig.setUseThreadLocalProducers(true);
        expectedSourceConfig.setProducerConfig(producerConfig);

        Resources resources = new Resources();
        resources.setRam(ram);
        resources.setCpu(cpu);
        expectedSourceConfig.setResources(resources);
        expectedSourceConfig.setCustomRuntimeOptions(customRuntimeOptions);

        SourcesImpl sources = spy(new SourcesImpl(meshWorkerServiceSupplier));
        SourceConfig actualSourceConfig = sources.getSourceInfo(tenant, namespace, componentName);
        Assert.assertEquals(expectedSourceConfig, actualSourceConfig);
    }

    @Test
    public void testGetSourceList()
            throws ApiException, IOException, ClassNotFoundException, URISyntaxException {
        String testBody =
                "{\n"
                        + "    \"apiVersion\": \"v1\",\n"
                        + "    \"items\": [\n"
                        + "        {\n"
                        + "            \"apiVersion\": \"compute.functionmesh.io/v1alpha1\",\n"
                        + "            \"kind\": \"Source\",\n"
                        + "            \"metadata\": {\n"
                        + "                \"annotations\": {\n"
                        + "                    \"kubectl.kubernetes.io/last-applied-configuration\": \"{\\\"apiVersion\\\":\\\"compute.functionmesh.io/v1alpha1\\\",\\\"kind\\\":\\\"Source\\\",\\\"metadata\\\":{\\\"annotations\\\":{},\\\"name\\\":\\\"source-sample\\\",\\\"namespace\\\":\\\"default\\\"},\\\"spec\\\":{\\\"className\\\":\\\"org.apache.pulsar.io.debezium.mongodb.DebeziumMongoDbSource\\\",\\\"clusterName\\\":\\\"test-pulsar\\\",\\\"java\\\":{\\\"jar\\\":\\\"connectors/pulsar-io-debezium-mongodb-2.7.0-rc-pm-3.nar\\\",\\\"jarLocation\\\":\\\"\\\"},\\\"maxReplicas\\\":1,\\\"output\\\":{\\\"producerConf\\\":{\\\"maxPendingMessages\\\":1000,\\\"maxPendingMessagesAcrossPartitions\\\":50000,\\\"useThreadLocalProducers\\\":true},\\\"topic\\\":\\\"persistent://public/default/destination\\\"},\\\"pulsar\\\":{\\\"pulsarConfig\\\":\\\"test-source\\\"},\\\"replicas\\\":1,\\\"resources\\\":{\\\"limits\\\":{\\\"cpu\\\":\\\"0.2\\\",\\\"memory\\\":\\\"1.1G\\\"},\\\"requests\\\":{\\\"cpu\\\":\\\"0.1\\\",\\\"memory\\\":\\\"1G\\\"}},\\\"sinkType\\\":\\\"org.apache.pulsar.common.schema.KeyValue\\\",\\\"sourceConfig\\\":{\\\"database.whitelist\\\":\\\"inventory\\\",\\\"mongodb.hosts\\\":\\\"rs0/mongo-dbz-0.mongo.default.svc.cluster.local:27017,rs0/mongo-dbz-1.mongo.default.svc.cluster.local:27017,rs0/mongo-dbz-2.mongo.default.svc.cluster.local:27017\\\",\\\"mongodb.name\\\":\\\"dbserver1\\\",\\\"mongodb.password\\\":\\\"dbz\\\",\\\"mongodb.task.id\\\":\\\"1\\\",\\\"mongodb.user\\\":\\\"debezium\\\",\\\"pulsar.service.url\\\":\\\"pulsar://test-pulsar-broker.default.svc.cluster.local:6650\\\"},\\\"sourceType\\\":\\\"org.apache.pulsar.common.schema.KeyValue\\\"}}\\n\"\n"
                        + "                },\n"
                        + "                \"creationTimestamp\": \"2020-11-27T07:07:57Z\",\n"
                        + "                \"generation\": 1,\n"
                        + "                \"name\": \"source-sample\",\n"
                        + "                \"namespace\": \"default\",\n"
                        + "                \"resourceVersion\": \"881034\",\n"
                        + "                \"selfLink\": \"/apis/compute.functionmesh.io/v1alpha1/namespaces/default/sources/source-sample\",\n"
                        + "                \"uid\": \"8aed505e-38e4-4a8b-93f6-6f753dbf7ebc\"\n"
                        + "            },\n"
                        + "            \"spec\": {\n"
                        + "                \"className\": \"org.apache.pulsar.io.debezium.mongodb.DebeziumMongoDbSource\",\n"
                        + "                \"clusterName\": \"test-pulsar\",\n"
                        + "                \"java\": {\n"
                        + "                    \"jar\": \"connectors/pulsar-io-debezium-mongodb-2.7.0-rc-pm-3.nar\",\n"
                        + "                    \"jarLocation\": \"\"\n"
                        + "                },\n"
                        + "                \"maxReplicas\": 1,\n"
                        + "                \"output\": {\n"
                        + "                    \"producerConf\": {\n"
                        + "                        \"maxPendingMessages\": 1000,\n"
                        + "                        \"maxPendingMessagesAcrossPartitions\": 50000,\n"
                        + "                        \"useThreadLocalProducers\": true\n"
                        + "                    },\n"
                        + "                    \"topic\": \"persistent://public/default/destination\"\n"
                        + "                },\n"
                        + "                \"pulsar\": {\n"
                        + "                    \"pulsarConfig\": \"test-source\"\n"
                        + "                },\n"
                        + "                \"replicas\": 1,\n"
                        + "                \"resources\": {\n"
                        + "                    \"limits\": {\n"
                        + "                        \"cpu\": \"0.2\",\n"
                        + "                        \"memory\": \"1.1G\"\n"
                        + "                    },\n"
                        + "                    \"requests\": {\n"
                        + "                        \"cpu\": \"0.1\",\n"
                        + "                        \"memory\": \"1G\"\n"
                        + "                    }\n"
                        + "                },\n"
                        + "                \"sinkType\": \"org.apache.pulsar.common.schema.KeyValue\",\n"
                        + "                \"sourceConfig\": {\n"
                        + "                    \"database.whitelist\": \"inventory\",\n"
                        + "                    \"mongodb.hosts\": \"rs0/mongo-dbz-0.mongo.default.svc.cluster.local:27017,rs0/mongo-dbz-1.mongo.default.svc.cluster.local:27017,rs0/mongo-dbz-2.mongo.default.svc.cluster.local:27017\",\n"
                        + "                    \"mongodb.name\": \"dbserver1\",\n"
                        + "                    \"mongodb.password\": \"dbz\",\n"
                        + "                    \"mongodb.task.id\": \"1\",\n"
                        + "                    \"mongodb.user\": \"debezium\",\n"
                        + "                    \"pulsar.service.url\": \"pulsar://test-pulsar-broker.default.svc.cluster.local:6650\"\n"
                        + "                },\n"
                        + "                \"sourceType\": \"org.apache.pulsar.common.schema.KeyValue\"\n"
                        + "            },\n"
                        + "            \"status\": {\n"
                        + "                \"conditions\": {\n"
                        + "                    \"HorizontalPodAutoscaler\": {\n"
                        + "                        \"action\": \"NoAction\",\n"
                        + "                        \"condition\": \"HPAReady\",\n"
                        + "                        \"status\": \"True\"\n"
                        + "                    },\n"
                        + "                    \"Service\": {\n"
                        + "                        \"action\": \"NoAction\",\n"
                        + "                        \"condition\": \"ServiceReady\",\n"
                        + "                        \"status\": \"True\"\n"
                        + "                    },\n"
                        + "                    \"StatefulSet\": {\n"
                        + "                        \"action\": \"NoAction\",\n"
                        + "                        \"condition\": \"StatefulSetReady\",\n"
                        + "                        \"status\": \"True\"\n"
                        + "                    }\n"
                        + "                },\n"
                        + "                \"replicas\": 1,\n"
                        + "                \"selector\": \"component=source,name=source-sample,namespace=default\"\n"
                        + "            }\n"
                        + "        }\n"
                        + "    ],\n"
                        + "    \"kind\": \"List\",\n"
                        + "    \"metadata\": {\n"
                        + "        \"resourceVersion\": \"\",\n"
                        + "        \"selfLink\": \"\"\n"
                        + "    }\n"
                        + "}";
        MeshWorkerService meshWorkerService =
                PowerMockito.mock(MeshWorkerService.class);
        Supplier<MeshWorkerService> meshWorkerServiceSupplier =
                () -> meshWorkerService;
        CustomObjectsApi customObjectsApi = PowerMockito.mock(CustomObjectsApi.class);
        PowerMockito.when(meshWorkerService.getCustomObjectsApi())
                .thenReturn(customObjectsApi);
        Call call = PowerMockito.mock(Call.class);
        Response response = PowerMockito.mock(Response.class);
        ResponseBody responseBody = PowerMockito.mock(RealResponseBody.class);
        ApiClient apiClient = PowerMockito.mock(ApiClient.class);

        String group = "compute.functionmesh.io";
        String plural = "sources";
        String version = "v1alpha1";

        PowerMockito.when(
                meshWorkerService
                                .getCustomObjectsApi()
                                .listClusterCustomObjectCall(
                                        group, version, plural, null, null, null, null, null, null,
                                        null, null, null))
                .thenReturn(call);
        PowerMockito.when(call.execute()).thenReturn(response);
        PowerMockito.when(response.isSuccessful()).thenReturn(true);
        PowerMockito.when(response.body()).thenReturn(responseBody);
        PowerMockito.when(responseBody.string()).thenReturn(testBody);
        PowerMockito.when(meshWorkerService.getApiClient()).thenReturn(apiClient);
        JSON json = new JSON();
        PowerMockito.when(apiClient.getJSON()).thenReturn(json);

        String className = "org.apache.pulsar.io.debezium.mongodb.DebeziumMongoDbSource";
        String sourceType = "org.apache.pulsar.common.schema.KeyValue";

        List<ConnectorDefinition> expectedList = new ArrayList<>();
        ConnectorDefinition connectorDefinition = new ConnectorDefinition();
        connectorDefinition.setSourceClass(sourceType);
        connectorDefinition.setName(className);
        expectedList.add(connectorDefinition);

        SourcesImpl sources = spy(new SourcesImpl(meshWorkerServiceSupplier));
        List<ConnectorDefinition> actualList = sources.getSourceList();

        Assert.assertEquals(expectedList, actualList);
    }
}
