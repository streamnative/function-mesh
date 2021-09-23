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

import com.google.common.collect.Maps;
import io.functionmesh.compute.MeshWorkerService;
import io.functionmesh.compute.models.MeshWorkerServiceCustomConfig;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecPod;
import io.functionmesh.compute.util.CommonUtil;
import io.functionmesh.compute.util.KubernetesUtils;
import io.functionmesh.compute.util.SinksUtil;
import io.functionmesh.compute.sinks.models.V1alpha1Sink;
import io.kubernetes.client.openapi.ApiClient;
import io.kubernetes.client.openapi.ApiException;
import io.kubernetes.client.openapi.JSON;
import io.kubernetes.client.openapi.apis.AppsV1Api;
import io.kubernetes.client.openapi.apis.CoreV1Api;
import io.kubernetes.client.openapi.apis.CustomObjectsApi;
import io.kubernetes.client.openapi.models.V1ObjectMeta;
import io.kubernetes.client.openapi.models.V1Pod;
import io.kubernetes.client.openapi.models.V1PodList;
import io.kubernetes.client.openapi.models.V1PodStatus;
import io.kubernetes.client.openapi.models.V1StatefulSet;
import io.kubernetes.client.openapi.models.V1StatefulSetSpec;
import io.kubernetes.client.openapi.models.V1StatefulSetStatus;
import okhttp3.Call;
import okhttp3.Response;
import okhttp3.ResponseBody;
import okhttp3.internal.http.RealResponseBody;
import org.apache.commons.io.FileUtils;
import org.apache.pulsar.client.admin.PulsarAdmin;
import org.apache.pulsar.client.admin.PulsarAdminException;
import org.apache.pulsar.client.admin.Tenants;
import org.apache.pulsar.common.functions.Resources;
import org.apache.pulsar.common.io.SinkConfig;
import org.apache.pulsar.common.nar.NarClassLoader;
import org.apache.pulsar.common.policies.data.SinkStatus;
import org.apache.pulsar.functions.proto.InstanceCommunication;
import org.apache.pulsar.functions.proto.InstanceControlGrpc;
import org.apache.pulsar.functions.utils.FunctionCommon;
import org.apache.pulsar.functions.utils.io.ConnectorUtils;
import org.apache.pulsar.functions.worker.WorkerConfig;
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
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.CompletableFuture;
import java.util.function.Supplier;

import static org.mockito.Matchers.any;
import static org.mockito.Matchers.anyObject;
import static org.mockito.Matchers.anyString;
import static org.powermock.api.mockito.PowerMockito.spy;

@RunWith(PowerMockRunner.class)
@PrepareForTest({
        Response.class,
        RealResponseBody.class,
        FunctionCommon.class,
        ConnectorUtils.class,
        FileUtils.class,
        CommonUtil.class,
        InstanceControlGrpc.InstanceControlFutureStub.class
})
@PowerMockIgnore({"javax.management.*"})
public class SinksImpTest {
    private final String kind = "Sink";
    private final String plural = "sinks";
    private final String group = "compute.functionmesh.io";
    private final String version = "v1alpha1";

    @Test
    public void testRegisterSink()
            throws ApiException, IOException, ClassNotFoundException, PulsarAdminException {
        // testBody is used to return a V1alpha1Sink JSON and does not care about the content.
        String testBody =
                "{\n"
                        + "  \"apiVersion\": \"compute.functionmesh.io/v1alpha1\",\n"
                        + "  \"kind\": \"Sink\",\n"
                        + "  \"metadata\": {\n"
                        + "    \"annotations\": {\n"
                        + "      \"kubectl.kubernetes.io/last-applied-configuration\": \"{\\\"apiVersion\\\":\\\"compute.functionmesh.io/v1alpha1\\\",\\\"kind\\\":\\\"Sink\\\",\\\"metadata\\\":{\\\"annotations\\\":{},\\\"name\\\":\\\"sink-sample\\\",\\\"namespace\\\":\\\"default\\\"},\\\"spec\\\":{\\\"autoAck\\\":true,\\\"className\\\":\\\"org.apache.pulsar.io.elasticsearch.ElasticSearchSink\\\",\\\"clusterName\\\":\\\"test-pulsar\\\",\\\"input\\\":{\\\"topics\\\":[\\\"persistent://public/default/input\\\"]},\\\"java\\\":{\\\"jar\\\":\\\"connectors/pulsar-io-elastic-search-2.7.0-rc-pm-3.nar\\\",\\\"jarLocation\\\":\\\"\\\"},\\\"maxReplicas\\\":1,\\\"pulsar\\\":{\\\"pulsarConfig\\\":\\\"test-sink\\\"},\\\"replicas\\\":1,\\\"resources\\\":{\\\"limits\\\":{\\\"cpu\\\":\\\"0.2\\\",\\\"memory\\\":\\\"1.1G\\\"},\\\"requests\\\":{\\\"cpu\\\":\\\"0.1\\\",\\\"memory\\\":\\\"1G\\\"}},\\\"sinkConfig\\\":{\\\"elasticSearchUrl\\\":\\\"http://quickstart-es-http.default.svc.cluster.local:9200\\\",\\\"indexName\\\":\\\"my_index\\\",\\\"password\\\":\\\"wJ757TmoXEd941kXm07Z2GW3\\\",\\\"typeName\\\":\\\"doc\\\",\\\"username\\\":\\\"elastic\\\"},\\\"sinkType\\\":\\\"[B\\\",\\\"sourceType\\\":\\\"[B\\\"}}\\n\"\n"
                        + "    },\n"
                        + "    \"creationTimestamp\": \"2020-11-27T07:32:51Z\",\n"
                        + "    \"generation\": 1,\n"
                        + "    \"name\": \"sink-sample\",\n"
                        + "    \"namespace\": \"default\",\n"
                        + "    \"resourceVersion\": \"888546\",\n"
                        + "    \"selfLink\": \"/apis/compute.functionmesh.io/v1alpha1/namespaces/default/sinks/sink-sample\",\n"
                        + "    \"uid\": \"4cd65795-18d4-46ee-a514-abee5048f1a1\"\n"
                        + "  },\n"
                        + "  \"spec\": {\n"
                        + "    \"autoAck\": true,\n"
                        + "    \"className\": \"org.apache.pulsar.io.elasticsearch.ElasticSearchSink\",\n"
                        + "    \"clusterName\": \"test-pulsar\",\n"
                        + "    \"input\": {\n"
                        + "      \"topics\": [\n"
                        + "        \"persistent://public/default/input\"\n"
                        + "      ]\n"
                        + "    },\n"
                        + "    \"java\": {\n"
                        + "      \"jar\": \"connectors/pulsar-io-elastic-search-2.7.0-rc-pm-3.nar\",\n"
                        + "      \"jarLocation\": \"\"\n"
                        + "    },\n"
                        + "    \"maxReplicas\": 1,\n"
                        + "    \"pulsar\": {\n"
                        + "      \"pulsarConfig\": \"test-sink\"\n"
                        + "    },\n"
                        + "    \"replicas\": 1,\n"
                        + "    \"resources\": {\n"
                        + "      \"limits\": {\n"
                        + "        \"cpu\": \"0.2\",\n"
                        + "        \"memory\": \"1.1G\"\n"
                        + "      },\n"
                        + "      \"requests\": {\n"
                        + "        \"cpu\": \"0.1\",\n"
                        + "        \"memory\": \"1G\"\n"
                        + "      }\n"
                        + "    },\n"
                        + "    \"sinkConfig\": {\n"
                        + "      \"elasticSearchUrl\": \"http://quickstart-es-http.default.svc.cluster.local:9200\",\n"
                        + "      \"indexName\": \"my_index\",\n"
                        + "      \"password\": \"wJ757TmoXEd941kXm07Z2GW3\",\n"
                        + "      \"typeName\": \"doc\",\n"
                        + "      \"username\": \"elastic\"\n"
                        + "    },\n"
                        + "    \"sinkType\": \"[B\",\n"
                        + "    \"sourceType\": \"[B\"\n"
                        + "  },\n"
                        + "  \"status\": {\n"
                        + "    \"conditions\": {\n"
                        + "      \"HorizontalPodAutoscaler\": {\n"
                        + "        \"action\": \"NoAction\",\n"
                        + "        \"condition\": \"HPAReady\",\n"
                        + "        \"status\": \"True\"\n"
                        + "      },\n"
                        + "      \"Service\": {\n"
                        + "        \"action\": \"NoAction\",\n"
                        + "        \"condition\": \"ServiceReady\",\n"
                        + "        \"status\": \"True\"\n"
                        + "      },\n"
                        + "      \"StatefulSet\": {\n"
                        + "        \"action\": \"NoAction\",\n"
                        + "        \"condition\": \"StatefulSetReady\",\n"
                        + "        \"status\": \"True\"\n"
                        + "      }\n"
                        + "    },\n"
                        + "    \"replicas\": 1,\n"
                        + "    \"selector\": \"component=sink,name=sink-sample,namespace=default\"\n"
                        + "  }\n"
                        + "}";
        MeshWorkerService meshWorkerService =
                PowerMockito.mock(MeshWorkerService.class);
        Supplier<MeshWorkerService> meshWorkerServiceSupplier =
                () -> meshWorkerService;
        CustomObjectsApi customObjectsApi = PowerMockito.mock(CustomObjectsApi.class);
        PowerMockito.when(meshWorkerService.getCustomObjectsApi())
                .thenReturn(customObjectsApi);
        WorkerConfig workerConfig = PowerMockito.mock(WorkerConfig.class);
        PowerMockito.when(meshWorkerService.getWorkerConfig()).thenReturn(workerConfig);
        PowerMockito.when(workerConfig.isAuthorizationEnabled()).thenReturn(false);
        PowerMockito.when(workerConfig.isAuthenticationEnabled()).thenReturn(false);
        PulsarAdmin pulsarAdmin = PowerMockito.mock(PulsarAdmin.class);
        PowerMockito.when(meshWorkerService.getBrokerAdmin()).thenReturn(pulsarAdmin);
        PowerMockito.when(meshWorkerService.getMeshWorkerServiceCustomConfig()).thenReturn(new MeshWorkerServiceCustomConfig());
        Tenants tenants = PowerMockito.mock(Tenants.class);
        PowerMockito.when(pulsarAdmin.tenants()).thenReturn(tenants);
        Call call = PowerMockito.mock(Call.class);
        Response response = PowerMockito.mock(Response.class);
        ResponseBody responseBody = PowerMockito.mock(RealResponseBody.class);
        ApiClient apiClient = PowerMockito.mock(ApiClient.class);

        String tenant = "public";
        String namespace = "default";
        String componentName = "sink-es-sample";
        String className = "org.apache.pulsar.io.elasticsearch.ElasticSearchSink";
        String inputTopic = "persistent://public/default/input";
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
        PowerMockito.when(ConnectorUtils.getIOSinkClass(narClassLoader)).thenReturn(className);
        PowerMockito.<Class<?>>when(FunctionCommon.getSinkType(null)).thenReturn(getClass());

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

        PowerMockito.when(tenants.getTenantInfo(tenant)).thenReturn(null);

        MeshWorkerServiceCustomConfig meshWorkerServiceCustomConfig = PowerMockito.mock(MeshWorkerServiceCustomConfig.class);
        PowerMockito.when(meshWorkerServiceCustomConfig.isUploadEnabled()).thenReturn(true);
        PowerMockito.when(meshWorkerServiceCustomConfig.isSinkEnabled()).thenReturn(true);
        PowerMockito.when(meshWorkerService.getMeshWorkerServiceCustomConfig()).thenReturn(meshWorkerServiceCustomConfig);

        V1alpha1Sink v1alpha1Sink =
                SinksUtil.createV1alpha1SkinFromSinkConfig(
                        kind, group, version, componentName, null, uploadedInputStream, sinkConfig, null,
                        null, meshWorkerService);

        Map<String, String> customLabels = Maps.newHashMap();
        customLabels.put("pulsar-cluster", clusterName);
        customLabels.put("pulsar-tenant", tenant);
        customLabels.put("pulsar-namespace", namespace);
        customLabels.put("pulsar-component", componentName);
        V1alpha1SinkSpecPod pod = new V1alpha1SinkSpecPod();
        pod.setLabels(customLabels);
        v1alpha1Sink.getSpec().pod(pod);
        v1alpha1Sink.getMetadata().setLabels(customLabels);
        PowerMockito.when(
                meshWorkerService
                                .getCustomObjectsApi()
                                .createNamespacedCustomObjectCall(
                                        group,
                                        version,
                                        KubernetesUtils.getNamespace(),
                                        plural,
                                        v1alpha1Sink,
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

        SinksImpl sinks = spy(new SinksImpl(meshWorkerServiceSupplier));
        System.out.println(KubernetesUtils.getNamespace());
        try {
            sinks.registerSink(
                    tenant,
                    namespace,
                    componentName,
                    uploadedInputStream,
                    null,
                    null,
                    sinkConfig,
                    null,
                    null);
        } catch (Exception exception) {
            Assert.fail("Got exception: " + exception);
        }
    }

    @Test
    public void testUpdateSink()
            throws ApiException, IOException, ClassNotFoundException, URISyntaxException {
        String getBody =
                "{\n"
                        + "    \"apiVersion\": \"compute.functionmesh.io/v1alpha1\",\n"
                        + "    \"kind\": \"Sink\",\n"
                        + "    \"metadata\": {\n"
                        + "          \"labels\": {\n"
                        + "                     \"pulsar-namespace\": \"default\",\n"
                        + "                     \"pulsar-tenant\": \"public\",\n"
                        + "                     \"pulsar-component\": \"sink-es-sample\",\n"
                        + "                     \"pulsar-cluster\": \"test-sink\""
                        + "            },\n"
                        + "        \"resourceVersion\": \"881033\""
                        + "    }\n"
                        + "}";

        String replaceBody =
                "{\n"
                        + "  \"apiVersion\": \"compute.functionmesh.io/v1alpha1\",\n"
                        + "  \"kind\": \"Sink\",\n"
                        + "  \"metadata\": {\n"
                        + "    \"creationTimestamp\": \"2020-11-27T07:32:51Z\",\n"
                        + "    \"generation\": 1,\n"
                        + "    \"name\": \"sink-es-sample\",\n"
                        + "    \"namespace\": \"default\",\n"
                        + "    \"resourceVersion\": \"888546\",\n"
                        + "    \"selfLink\": \"/apis/compute.functionmesh.io/v1alpha1/namespaces/default/sinks/sink-sample\",\n"
                        + "    \"uid\": \"4cd65795-18d4-46ee-a514-abee5048f1a1\"\n"
                        + "  },\n"
                        + "  \"spec\": {\n"
                        + "    \"autoAck\": true,\n"
                        + "    \"className\": \"org.apache.pulsar.io.elasticsearch.ElasticSearchSink\",\n"
                        + "    \"clusterName\": \"test-pulsar\",\n"
                        + "    \"input\": {\n"
                        + "      \"topics\": [\n"
                        + "        \"persistent://public/default/input\"\n"
                        + "      ]\n"
                        + "    },\n"
                        + "    \"java\": {\n"
                        + "      \"jar\": \"connectors/pulsar-io-elastic-search-2.7.0-rc-pm-3.nar\",\n"
                        + "      \"jarLocation\": \"\"\n"
                        + "    },\n"
                        + "    \"maxReplicas\": 1,\n"
                        + "    \"pulsar\": {\n"
                        + "      \"pulsarConfig\": \"test-sink\"\n"
                        + "    },\n"
                        + "    \"replicas\": 1,\n"
                        + "    \"resources\": {\n"
                        + "      \"limits\": {\n"
                        + "        \"cpu\": \"0.1\",\n"
                        + "        \"memory\": \"1\"\n"
                        + "      },\n"
                        + "      \"requests\": {\n"
                        + "        \"cpu\": \"0.1\",\n"
                        + "        \"memory\": \"1\"\n"
                        + "      }\n"
                        + "    },\n"
                        + "    \"sinkConfig\": {\n"
                        + "      \"elasticSearchUrl\": \"https://testing-es.app\",\n"
                        + "    },\n"
                        + "    \"sinkType\": \"[B\",\n"
                        + "    \"sourceType\": \"[B\"\n"
                        + "  },\n"
                        + "  \"status\": {\n"
                        + "    \"conditions\": {\n"
                        + "      \"HorizontalPodAutoscaler\": {\n"
                        + "        \"action\": \"NoAction\",\n"
                        + "        \"condition\": \"HPAReady\",\n"
                        + "        \"status\": \"True\"\n"
                        + "      },\n"
                        + "      \"Service\": {\n"
                        + "        \"action\": \"NoAction\",\n"
                        + "        \"condition\": \"ServiceReady\",\n"
                        + "        \"status\": \"True\"\n"
                        + "      },\n"
                        + "      \"StatefulSet\": {\n"
                        + "        \"action\": \"NoAction\",\n"
                        + "        \"condition\": \"StatefulSetReady\",\n"
                        + "        \"status\": \"True\"\n"
                        + "      }\n"
                        + "    },\n"
                        + "    \"replicas\": 1,\n"
                        + "    \"selector\": \"component=sink,name=sink-sample,namespace=default\"\n"
                        + "  }\n"
                        + "}";

        String tenant = "public";
        String namespace = "default";
        String componentName = "sink-es-sample";
        String className = "org.apache.pulsar.io.elasticsearch.ElasticSearchSink";
        String inputTopic = "persistent://public/default/input";
        String archive = "connectors/pulsar-io-elastic-search-2.7.0-rc-pm-3.nar";
        boolean autoAck = true;
        int parallelism = 1;
        Double cpu = 0.1;
        Long ram = 1L;
        String customRuntimeOptions = "{\"clusterName\": \"test-pulsar\"}";
        Map<String, Object> configs = new HashMap<>();
        configs.put("elasticSearchUrl", "https://testing-es.app");
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
        PowerMockito.when(ConnectorUtils.getIOSinkClass(narClassLoader)).thenReturn(className);
        PowerMockito.<Class<?>>when(FunctionCommon.getSinkType(null)).thenReturn(getClass());

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

        MeshWorkerService meshWorkerService =
                PowerMockito.mock(MeshWorkerService.class);
        Supplier<MeshWorkerService> meshWorkerServiceSupplier =
                () -> meshWorkerService;
        CustomObjectsApi customObjectsApi = PowerMockito.mock(CustomObjectsApi.class);
        PowerMockito.when(meshWorkerService.getCustomObjectsApi())
                .thenReturn(customObjectsApi);
        WorkerConfig workerConfig = PowerMockito.mock(WorkerConfig.class);
        PowerMockito.when(meshWorkerService.getWorkerConfig()).thenReturn(workerConfig);
        PowerMockito.when(workerConfig.isAuthorizationEnabled()).thenReturn(false);
        PowerMockito.when(workerConfig.isAuthenticationEnabled()).thenReturn(false);
        MeshWorkerServiceCustomConfig meshWorkerServiceCustomConfig = PowerMockito.mock(MeshWorkerServiceCustomConfig.class);
        PowerMockito.when(meshWorkerServiceCustomConfig.isUploadEnabled()).thenReturn(true);
        PowerMockito.when(meshWorkerServiceCustomConfig.isSinkEnabled()).thenReturn(true);
        PowerMockito.when(meshWorkerService.getMeshWorkerServiceCustomConfig()).thenReturn(meshWorkerServiceCustomConfig);

        Call getCall = PowerMockito.mock(Call.class);
        Response getResponse = PowerMockito.mock(Response.class);
        ResponseBody getResponseBody = PowerMockito.mock(RealResponseBody.class);
        PowerMockito.when(getCall.execute()).thenReturn(getResponse);
        PowerMockito.when(getResponse.isSuccessful()).thenReturn(true);
        PowerMockito.when(getResponse.body()).thenReturn(getResponseBody);
        PowerMockito.when(getResponseBody.string()).thenReturn(getBody);


        ApiClient apiClient = PowerMockito.mock(ApiClient.class);
        PowerMockito.when(meshWorkerService.getApiClient()).thenReturn(apiClient);
        JSON json = new JSON();
        PowerMockito.when(apiClient.getJSON()).thenReturn(json);

        PowerMockito.when(
                customObjectsApi
                        .getNamespacedCustomObjectCall(
                                group, version, namespace, plural, "sink-es-sample-e84d93a8", null))
                .thenReturn(getCall);

        Call replaceCall = PowerMockito.mock(Call.class);
        PowerMockito.when(
                customObjectsApi
                                .replaceNamespacedCustomObjectCall(
                                        anyString(),
                                        anyString(),
                                        anyString(),
                                        anyString(),
                                        anyString(),
                                        any(V1alpha1Sink.class),
                                        anyObject(),
                                        anyObject(),
                                        anyObject()))
                .thenReturn(replaceCall);
        Response replaceResponse = PowerMockito.mock(Response.class);
        ResponseBody replaceResponseBody = PowerMockito.mock(RealResponseBody.class);
        PowerMockito.when(replaceCall.execute()).thenReturn(replaceResponse);
        PowerMockito.when(replaceResponse.isSuccessful()).thenReturn(true);
        PowerMockito.when(replaceResponse.body()).thenReturn(getResponseBody);
        PowerMockito.when(replaceResponseBody.string()).thenReturn(replaceBody);

        SinksImpl sinks = spy(new SinksImpl(meshWorkerServiceSupplier));

        try {
            sinks.updateSink(
                    tenant,
                    namespace,
                    componentName,
                    uploadedInputStream,
                    null,
                    null,
                    sinkConfig,
                    null,
                    null,
                    null);
        } catch (Exception exception) {
            Assert.fail("Expected no exception to be thrown but got exception: " + exception);
        }
    }

    @Test
    public void testGetSinkStatus() throws Exception {
        String testBody =
                "{\n"
                        + "  \"apiVersion\": \"compute.functionmesh.io/v1alpha1\",\n"
                        + "  \"kind\": \"Sink\",\n"
                        + "  \"metadata\": {\n"
                        + "    \"creationTimestamp\": \"2020-11-27T07:32:51Z\",\n"
                        + "    \"generation\": 1,\n"
                        + "    \"name\": \"sink-sample\",\n"
                        + "    \"namespace\": \"default\",\n"
                        + "    \"resourceVersion\": \"888546\",\n"
                        + "    \"selfLink\": \"/apis/compute.functionmesh.io/v1alpha1/namespaces/default/sinks/sink-sample\",\n"
                        + "    \"uid\": \"4cd65795-18d4-46ee-a514-abee5048f1a1\"\n"
                        + "  },\n"
                        + "  \"spec\": {\n"
                        + "    \"autoAck\": true,\n"
                        + "    \"className\": \"org.apache.pulsar.io.elasticsearch.ElasticSearchSink\",\n"
                        + "    \"clusterName\": \"test-pulsar\",\n"
                        + "    \"input\": {\n"
                        + "      \"topics\": [\n"
                        + "        \"persistent://public/default/input\"\n"
                        + "      ]\n"
                        + "    },\n"
                        + "    \"java\": {\n"
                        + "      \"jar\": \"connectors/pulsar-io-elastic-search-2.7.0-rc-pm-3.nar\",\n"
                        + "      \"jarLocation\": \"\"\n"
                        + "    },\n"
                        + "    \"maxReplicas\": 1,\n"
                        + "    \"pulsar\": {\n"
                        + "      \"pulsarConfig\": \"test-sink\"\n"
                        + "    },\n"
                        + "    \"replicas\": 1,\n"
                        + "    \"resources\": {\n"
                        + "      \"limits\": {\n"
                        + "        \"cpu\": \"0.1\",\n"
                        + "        \"memory\": \"1\"\n"
                        + "      },\n"
                        + "      \"requests\": {\n"
                        + "        \"cpu\": \"0.1\",\n"
                        + "        \"memory\": \"1\"\n"
                        + "      }\n"
                        + "    },\n"
                        + "    \"sinkConfig\": {\n"
                        + "      \"elasticSearchUrl\": \"https://testing-es.app\"\n"
                        + "    },\n"
                        + "    \"sinkType\": \"[B\",\n"
                        + "    \"sourceType\": \"[B\"\n"
                        + "  },\n"
                        + "  \"status\": {\n"
                        + "    \"conditions\": {\n"
                        + "      \"HorizontalPodAutoscaler\": {\n"
                        + "        \"action\": \"NoAction\",\n"
                        + "        \"condition\": \"HPAReady\",\n"
                        + "        \"status\": \"True\"\n"
                        + "      },\n"
                        + "      \"Service\": {\n"
                        + "        \"action\": \"NoAction\",\n"
                        + "        \"condition\": \"ServiceReady\",\n"
                        + "        \"status\": \"True\"\n"
                        + "      },\n"
                        + "      \"StatefulSet\": {\n"
                        + "        \"action\": \"NoAction\",\n"
                        + "        \"condition\": \"StatefulSetReady\",\n"
                        + "        \"status\": \"True\"\n"
                        + "      }\n"
                        + "    },\n"
                        + "    \"replicas\": 1,\n"
                        + "    \"selector\": \"component=sink,name=sink-sample,namespace=default\"\n"
                        + "  }\n"
                        + "}";
        MeshWorkerService meshWorkerService =
                PowerMockito.mock(MeshWorkerService.class);
        Supplier<MeshWorkerService> meshWorkerServiceSupplier =
                () -> meshWorkerService;
        CustomObjectsApi customObjectsApi = PowerMockito.mock(CustomObjectsApi.class);
        CoreV1Api coreV1Api = PowerMockito.mock(CoreV1Api.class);
        AppsV1Api appsV1Api = PowerMockito.mock(AppsV1Api.class);
        PowerMockito.when(meshWorkerService.getCustomObjectsApi())
                .thenReturn(customObjectsApi);
        PowerMockito.when(meshWorkerService.getCoreV1Api())
                .thenReturn(coreV1Api);
        PowerMockito.when(meshWorkerService.getAppsV1Api())
                .thenReturn(appsV1Api);
        WorkerConfig workerConfig = PowerMockito.mock(WorkerConfig.class);
        PowerMockito.when(meshWorkerService.getWorkerConfig()).thenReturn(workerConfig);
        PowerMockito.when(workerConfig.isAuthorizationEnabled()).thenReturn(false);
        PowerMockito.when(workerConfig.isAuthenticationEnabled()).thenReturn(false);

        PowerMockito.when(workerConfig.getPulsarFunctionsCluster()).thenReturn("test-pulsar");

        Call call = PowerMockito.mock(Call.class);
        Response response = PowerMockito.mock(Response.class);
        ResponseBody responseBody = PowerMockito.mock(RealResponseBody.class);
        ApiClient apiClient = PowerMockito.mock(ApiClient.class);

        String group = "compute.functionmesh.io";
        String tenant = "public";
        String namespace = "default";
        String componentName = "sink-sample";
        String hashName = CommonUtil.generateObjectName(meshWorkerService, tenant, namespace, componentName);
        String jobName = CommonUtil.makeJobName(componentName, CommonUtil.COMPONENT_SINK);

        PowerMockito.when(
                meshWorkerService
                        .getCustomObjectsApi()
                        .getNamespacedCustomObjectCall(
                                group, version, namespace, plural,
                                CommonUtil.createObjectName("test-pulsar", tenant, namespace, componentName), null))
                .thenReturn(call);
        PowerMockito.when(call.execute()).thenReturn(response);
        PowerMockito.when(response.isSuccessful()).thenReturn(true);
        PowerMockito.when(response.body()).thenReturn(responseBody);
        PowerMockito.when(responseBody.string()).thenReturn(testBody);
        PowerMockito.when(meshWorkerService.getApiClient()).thenReturn(apiClient);
        JSON json = new JSON();
        PowerMockito.when(apiClient.getJSON()).thenReturn(json);

        V1StatefulSet v1StatefulSet = PowerMockito.mock(V1StatefulSet.class);
        PowerMockito.when(appsV1Api.readNamespacedStatefulSet(jobName, namespace, null, null, null)).thenReturn(v1StatefulSet);

        V1ObjectMeta v1StatefulSetV1ObjectMeta = PowerMockito.mock(V1ObjectMeta.class);
        PowerMockito.when(v1StatefulSet.getMetadata()).thenReturn(v1StatefulSetV1ObjectMeta);
        V1StatefulSetSpec v1StatefulSetSpec = PowerMockito.mock(V1StatefulSetSpec.class);
        PowerMockito.when(v1StatefulSet.getSpec()).thenReturn(v1StatefulSetSpec);
        PowerMockito.when(v1StatefulSetV1ObjectMeta.getName()).thenReturn(jobName);
        PowerMockito.when(v1StatefulSetSpec.getServiceName()).thenReturn(jobName);
        V1StatefulSetStatus v1StatefulSetStatus = PowerMockito.mock(V1StatefulSetStatus.class);
        PowerMockito.when(v1StatefulSet.getStatus()).thenReturn(v1StatefulSetStatus);
        PowerMockito.when(v1StatefulSetStatus.getReplicas()).thenReturn(1);
        PowerMockito.when(v1StatefulSetStatus.getReadyReplicas()).thenReturn(1);
        V1PodList list = PowerMockito.mock(V1PodList.class);
        List<V1Pod> podList = new ArrayList<>();
        V1Pod pod = PowerMockito.mock(V1Pod.class);
        podList.add(pod);
        PowerMockito.when(coreV1Api.listNamespacedPod(any(), any(), any(), any(), any(), any(),
                any(), any(), any(), any(), any())).thenReturn(list);
        PowerMockito.when(list.getItems()).thenReturn(podList);
        V1ObjectMeta podV1ObjectMeta = PowerMockito.mock(V1ObjectMeta.class);
        PowerMockito.when(pod.getMetadata()).thenReturn(podV1ObjectMeta);
        PowerMockito.when(podV1ObjectMeta.getName()).thenReturn(hashName+"-sink-0");
        V1PodStatus podStatus = PowerMockito.mock(V1PodStatus.class);
        PowerMockito.when(pod.getStatus()).thenReturn(podStatus);
        PowerMockito.when(podStatus.getPhase()).thenReturn("Running");
        PowerMockito.when(podStatus.getContainerStatuses()).thenReturn(new ArrayList<>());
        InstanceCommunication.FunctionStatus.Builder builder = InstanceCommunication.FunctionStatus.newBuilder();
        builder.setRunning(true);
        PowerMockito.mockStatic(InstanceControlGrpc.InstanceControlFutureStub.class);
        PowerMockito.stub(PowerMockito.method(CommonUtil.class, "getFunctionStatusAsync")).toReturn(CompletableFuture.completedFuture(builder.build()));

        SinksImpl sinks = spy(new SinksImpl(meshWorkerServiceSupplier));
        SinkStatus sinkStatus =
                sinks.getSinkStatus(tenant, namespace, componentName, null, null, null);

        SinkStatus expectedSinkStatus = new SinkStatus();
        SinkStatus.SinkInstanceStatus expectedSinkInstanceStatus =
                new SinkStatus.SinkInstanceStatus();
        SinkStatus.SinkInstanceStatus.SinkInstanceStatusData expectedSinkInstanceStatusData =
                new SinkStatus.SinkInstanceStatus.SinkInstanceStatusData();
        expectedSinkInstanceStatusData.setRunning(true);
        expectedSinkInstanceStatusData.setWorkerId("test-pulsar");
        expectedSinkInstanceStatusData.setError("");
        expectedSinkInstanceStatusData.setLatestSinkExceptions(Collections.emptyList());
        expectedSinkInstanceStatusData.setLatestSystemExceptions(Collections.emptyList());
        expectedSinkInstanceStatus.setStatus(expectedSinkInstanceStatusData);
        expectedSinkStatus.addInstance(expectedSinkInstanceStatus);
        expectedSinkStatus.setNumInstances(expectedSinkStatus.getInstances().size());
        expectedSinkStatus.setNumRunning(expectedSinkStatus.getInstances().size());

        Assert.assertEquals(expectedSinkStatus, sinkStatus);
    }

    @Test
    public void testGetSinkInfo() throws ApiException, IOException {
        String testBody =
                "{\n"
                        + "  \"apiVersion\": \"compute.functionmesh.io/v1alpha1\",\n"
                        + "  \"kind\": \"Sink\",\n"
                        + "  \"metadata\": {\n"
                        + "    \"creationTimestamp\": \"2020-11-27T07:32:51Z\",\n"
                        + "    \"generation\": 1,\n"
                        + "    \"name\": \"sink-sample\",\n"
                        + "    \"namespace\": \"default\",\n"
                        + "    \"resourceVersion\": \"888546\",\n"
                        + "    \"selfLink\": \"/apis/compute.functionmesh.io/v1alpha1/namespaces/default/sinks/sink-sample\",\n"
                        + "    \"uid\": \"4cd65795-18d4-46ee-a514-abee5048f1a1\"\n"
                        + "  },\n"
                        + "  \"spec\": {\n"
                        + "    \"autoAck\": true,\n"
                        + "    \"className\": \"org.apache.pulsar.io.elasticsearch.ElasticSearchSink\",\n"
                        + "    \"clusterName\": \"test-pulsar\",\n"
                        + "    \"input\": {\n"
                        + "      \"topics\": [\n"
                        + "        \"persistent://public/default/input\"\n"
                        + "      ]\n"
                        + "    },\n"
                        + "    \"java\": {\n"
                        + "      \"jar\": \"pulsar-io-elastic-search-2.7.0-rc-pm-3.nar\",\n"
                        + "      \"jarLocation\": \"\"\n"
                        + "    },\n"
                        + "    \"maxReplicas\": 1,\n"
                        + "    \"pulsar\": {\n"
                        + "      \"pulsarConfig\": \"test-sink\"\n"
                        + "    },\n"
                        + "    \"replicas\": 1,\n"
                        + "    \"resources\": {\n"
                        + "      \"limits\": {\n"
                        + "        \"cpu\": \"0.1\",\n"
                        + "        \"memory\": \"1\"\n"
                        + "      },\n"
                        + "      \"requests\": {\n"
                        + "        \"cpu\": \"0.1\",\n"
                        + "        \"memory\": \"1\"\n"
                        + "      }\n"
                        + "    },\n"
                        + "    \"sinkConfig\": {\n"
                        + "      \"elasticSearchUrl\": \"https://testing-es.app\"\n"
                        + "    },\n"
                        + "    \"sinkType\": \"testing.SinkType\",\n"
                        + "    \"sourceType\": \"testing.SourceType\"\n"
                        + "  },\n"
                        + "  \"status\": {\n"
                        + "    \"conditions\": {\n"
                        + "      \"HorizontalPodAutoscaler\": {\n"
                        + "        \"action\": \"NoAction\",\n"
                        + "        \"condition\": \"HPAReady\",\n"
                        + "        \"status\": \"True\"\n"
                        + "      },\n"
                        + "      \"Service\": {\n"
                        + "        \"action\": \"NoAction\",\n"
                        + "        \"condition\": \"ServiceReady\",\n"
                        + "        \"status\": \"True\"\n"
                        + "      },\n"
                        + "      \"StatefulSet\": {\n"
                        + "        \"action\": \"NoAction\",\n"
                        + "        \"condition\": \"StatefulSetReady\",\n"
                        + "        \"status\": \"True\"\n"
                        + "      }\n"
                        + "    },\n"
                        + "    \"replicas\": 1,\n"
                        + "    \"selector\": \"component=sink,name=sink-sample,namespace=default\"\n"
                        + "  }\n"
                        + "}";
        MeshWorkerService meshWorkerService =
                PowerMockito.mock(MeshWorkerService.class);
        Supplier<MeshWorkerService> meshWorkerServiceSupplier =
                () -> meshWorkerService;
        CustomObjectsApi customObjectsApi = PowerMockito.mock(CustomObjectsApi.class);
        PowerMockito.when(meshWorkerService.getCustomObjectsApi())
                .thenReturn(customObjectsApi);
        WorkerConfig workerConfig = PowerMockito.mock(WorkerConfig.class);
        PowerMockito.when(meshWorkerService.getWorkerConfig()).thenReturn(workerConfig);
        PowerMockito.when(workerConfig.isAuthorizationEnabled()).thenReturn(false);
        PowerMockito.when(workerConfig.isAuthenticationEnabled()).thenReturn(false);

        PowerMockito.when(workerConfig.getPulsarFunctionsCluster()).thenReturn("test-pulsar");

        Call call = PowerMockito.mock(Call.class);
        Response response = PowerMockito.mock(Response.class);
        ResponseBody responseBody = PowerMockito.mock(RealResponseBody.class);
        ApiClient apiClient = PowerMockito.mock(ApiClient.class);

        String tenant = "public";
        String namespace = "default";
        String componentName = "sink-sample";

        PowerMockito.when(
                meshWorkerService
                        .getCustomObjectsApi()
                        .getNamespacedCustomObjectCall(
                                group, version, namespace, plural,
                                CommonUtil.createObjectName("test-pulsar", tenant, namespace, componentName), null))
                .thenReturn(call);
        PowerMockito.when(call.execute()).thenReturn(response);
        PowerMockito.when(response.isSuccessful()).thenReturn(true);
        PowerMockito.when(response.body()).thenReturn(responseBody);
        PowerMockito.when(responseBody.string()).thenReturn(testBody);
        PowerMockito.when(meshWorkerService.getApiClient()).thenReturn(apiClient);
        JSON json = new JSON();
        PowerMockito.when(apiClient.getJSON()).thenReturn(json);

        String className = "org.apache.pulsar.io.elasticsearch.ElasticSearchSink";
        String inputTopic = "persistent://public/default/input";
        String archive = "pulsar-io-elastic-search-2.7.0-rc-pm-3.nar";
        boolean autoAck = true;
        int parallelism = 1;
        Double cpu = 0.1;
        Long ram = 1L;
        String clusterName = "test-pulsar";
        String customRuntimeOptions = "{\"clusterName\":\"" + clusterName + "\",\"maxReplicas\":1}";
        Map<String, Object> configs = new HashMap<>();
        configs.put("elasticSearchUrl", "https://testing-es.app");

        SinkConfig expectedSinkConfig = new SinkConfig();
        expectedSinkConfig.setTenant(tenant);
        expectedSinkConfig.setNamespace(namespace);
        expectedSinkConfig.setName(componentName);
        expectedSinkConfig.setClassName(className);
        expectedSinkConfig.setInputs(Collections.singletonList(inputTopic));
        expectedSinkConfig.setConfigs(configs);
        expectedSinkConfig.setArchive(archive);
        expectedSinkConfig.setParallelism(parallelism);
        Resources resources = new Resources();
        resources.setRam(ram);
        resources.setCpu(cpu);
        expectedSinkConfig.setResources(resources);
        expectedSinkConfig.setCustomRuntimeOptions(customRuntimeOptions);
        expectedSinkConfig.setAutoAck(autoAck);

        SinksImpl sinks = spy(new SinksImpl(meshWorkerServiceSupplier));
        SinkConfig actualSinkConfig = sinks.getSinkInfo(tenant, namespace, componentName);
        Assert.assertEquals(expectedSinkConfig.getName(), actualSinkConfig.getName());
        Assert.assertEquals(expectedSinkConfig.getNamespace(), actualSinkConfig.getNamespace());
        Assert.assertEquals(expectedSinkConfig.getTenant(), actualSinkConfig.getTenant());
        Assert.assertEquals(expectedSinkConfig.getConfigs(), actualSinkConfig.getConfigs());
        Assert.assertEquals(expectedSinkConfig.getArchive(), actualSinkConfig.getArchive());
        Assert.assertEquals(expectedSinkConfig.getResources(), actualSinkConfig.getResources());
        Assert.assertEquals(expectedSinkConfig.getClassName(), actualSinkConfig.getClassName());
        Assert.assertEquals(expectedSinkConfig.getAutoAck(), actualSinkConfig.getAutoAck());
        Assert.assertEquals(expectedSinkConfig.getCustomRuntimeOptions(), actualSinkConfig.getCustomRuntimeOptions());
        Assert.assertArrayEquals(expectedSinkConfig.getInputs().toArray(), actualSinkConfig.getInputs().toArray());
        Assert.assertEquals(expectedSinkConfig.getMaxMessageRetries(), actualSinkConfig.getMaxMessageRetries());
        Assert.assertEquals(expectedSinkConfig.getCleanupSubscription(), actualSinkConfig.getCleanupSubscription());
        Assert.assertEquals(expectedSinkConfig.getParallelism(), actualSinkConfig.getParallelism());
        Assert.assertEquals(expectedSinkConfig.getRuntimeFlags(), actualSinkConfig.getRuntimeFlags());
    }
}
