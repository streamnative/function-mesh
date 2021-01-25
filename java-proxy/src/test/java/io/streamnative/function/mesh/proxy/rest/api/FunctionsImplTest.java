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

import io.kubernetes.client.openapi.ApiClient;
import io.kubernetes.client.openapi.ApiException;
import io.kubernetes.client.openapi.JSON;
import io.kubernetes.client.openapi.apis.CustomObjectsApi;
import io.kubernetes.client.openapi.models.V1ObjectMeta;
import io.streamnative.cloud.models.function.*;
import io.streamnative.function.mesh.proxy.FunctionMeshProxyService;
import okhttp3.Call;
import okhttp3.Response;
import okhttp3.ResponseBody;
import okhttp3.internal.http.RealResponseBody;
import org.apache.pulsar.common.functions.ConsumerConfig;
import org.apache.pulsar.common.functions.FunctionConfig;
import org.apache.pulsar.common.functions.Resources;
import org.apache.pulsar.common.policies.data.FunctionStatus;
import org.junit.Assert;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.powermock.api.mockito.PowerMockito;
import org.powermock.core.classloader.annotations.PowerMockIgnore;
import org.powermock.core.classloader.annotations.PrepareForTest;
import org.powermock.modules.junit4.PowerMockRunner;

import java.io.IOException;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.function.Supplier;

import static org.powermock.api.mockito.PowerMockito.spy;

@RunWith(PowerMockRunner.class)
@PrepareForTest({Response.class, RealResponseBody.class})
@PowerMockIgnore({"javax.management.*"})
public class FunctionsImplTest {
    @Test
    public void getFunctionStatusTest() throws ApiException, IOException {
        String testBody = "{\n" +
                "  \"apiVersion\": \"cloud.streamnative.io/v1alpha1\",\n" +
                "  \"kind\": \"Function\",\n" +
                "  \"metadata\": {\n" +
                "    \"creationTimestamp\": \"2020-11-27T08:08:32Z\",\n" +
                "    \"generation\": 1,\n" +
                "    \"name\": \"functionmesh-sample-ex1\",\n" +
                "    \"namespace\": \"default\",\n" +
                "    \"ownerReferences\": [\n" +
                "      {\n" +
                "        \"apiVersion\": \"cloud.streamnative.io/v1alpha1\",\n" +
                "        \"blockOwnerDeletion\": true,\n" +
                "        \"controller\": true,\n" +
                "        \"kind\": \"FunctionMesh\",\n" +
                "        \"name\": \"functionmesh-sample\",\n" +
                "        \"uid\": \"5ef04d92-c8bf-40ac-ad9e-4a02bb75cb8b\"\n" +
                "      }\n" +
                "    ],\n" +
                "    \"resourceVersion\": \"899291\",\n" +
                "    \"selfLink\": \"/apis/cloud.streamnative.io/v1alpha1/namespaces/default/functions/functionmesh-sample-ex1\",\n" +
                "    \"uid\": \"9e4509ba-c8bd-4c76-8905-ea0cf7251552\"\n" +
                "  },\n" +
                "  \"spec\": {\n" +
                "    \"autoAck\": true,\n" +
                "    \"className\": \"org.apache.pulsar.functions.api.examples.ExclamationFunction\",\n" +
                "    \"clusterName\": \"test-pulsar\",\n" +
                "    \"forwardSourceMessageProperty\": true,\n" +
                "    \"input\": {\n" +
                "      \"topics\": [\n" +
                "        \"persistent://public/default/in-topic\"\n" +
                "      ]\n" +
                "    },\n" +
                "    \"java\": {\n" +
                "      \"jar\": \"pulsar-functions-api-examples.jar\",\n" +
                "      \"jarLocation\": \"public/default/nlu-test-java\"\n" +
                "    },\n" +
                "    \"logTopic\": \"persistent://public/default/logging-function-log\",\n" +
                "    \"maxReplicas\": 1,\n" +
                "    \"name\": \"ex1\",\n" +
                "    \"output\": {\n" +
                "      \"topic\": \"persistent://public/default/mid-topic\"\n" +
                "    },\n" +
                "    \"pulsar\": {\n" +
                "      \"pulsarConfig\": \"mesh-test-pulsar\"\n" +
                "    },\n" +
                "    \"replicas\": 1,\n" +
                "    \"resources\": {\n" +
                "      \"limits\": {\n" +
                "        \"cpu\": \"200m\",\n" +
                "        \"memory\": \"1100M\"\n" +
                "      },\n" +
                "      \"requests\": {\n" +
                "        \"cpu\": \"100m\",\n" +
                "        \"memory\": \"1G\"\n" +
                "      }\n" +
                "    },\n" +
                "    \"sinkType\": \"java.lang.String\",\n" +
                "    \"sourceType\": \"java.lang.String\"\n" +
                "  },\n" +
                "  \"status\": {\n" +
                "    \"conditions\": {\n" +
                "      \"HorizontalPodAutoscaler\": {\n" +
                "        \"action\": \"NoAction\",\n" +
                "        \"condition\": \"HPAReady\",\n" +
                "        \"status\": \"True\"\n" +
                "      },\n" +
                "      \"Service\": {\n" +
                "        \"action\": \"NoAction\",\n" +
                "        \"condition\": \"ServiceReady\",\n" +
                "        \"status\": \"True\"\n" +
                "      },\n" +
                "      \"StatefulSet\": {\n" +
                "        \"action\": \"NoAction\",\n" +
                "        \"condition\": \"StatefulSetReady\",\n" +
                "        \"status\": \"True\"\n" +
                "      }\n" +
                "    },\n" +
                "    \"replicas\": 1,\n" +
                "    \"selector\": \"component=function,name=functionmesh-sample-ex1,namespace=default\"\n" +
                "  }\n" +
                "}";

        FunctionMeshProxyService functionMeshProxyService = PowerMockito.mock(FunctionMeshProxyService.class);
        Supplier<FunctionMeshProxyService> functionMeshProxyServiceSupplier = new Supplier<FunctionMeshProxyService>() {
            @Override
            public FunctionMeshProxyService get() {
                return functionMeshProxyService;
            }
        };
        CustomObjectsApi customObjectsApi = PowerMockito.mock(CustomObjectsApi.class);
        PowerMockito.when(functionMeshProxyService.getCustomObjectsApi()).thenReturn(customObjectsApi);
        String tenant = "public";
        String namespace = "default";
        String name = "test";
        String group = "cloud.streamnative.io";
        String plural = "functions";
        String version = "v1alpha1";
        Call call = PowerMockito.mock(Call.class);
        Response response = PowerMockito.mock(Response.class);
        ResponseBody responseBody = PowerMockito.mock(RealResponseBody.class);
        ApiClient apiClient = PowerMockito.mock(ApiClient.class);
        PowerMockito.when(functionMeshProxyService.getCustomObjectsApi()
                .getNamespacedCustomObjectCall(
                        group, version, namespace, plural, name, null)).thenReturn(call);
        PowerMockito.when(call.execute()).thenReturn(response);
        PowerMockito.when(response.isSuccessful()).thenReturn(true);
        PowerMockito.when(response.body()).thenReturn(responseBody);
        PowerMockito.when(responseBody.string()).thenReturn(testBody);
        PowerMockito.when(functionMeshProxyService.getApiClient()).thenReturn(apiClient);
        JSON json = new JSON();
        PowerMockito.when(apiClient.getJSON()).thenReturn(json);
        FunctionStatus expectedFunctionStatus = new FunctionStatus();
        expectedFunctionStatus.setNumInstances(0);
        expectedFunctionStatus.setNumInstances(1);
        List<FunctionStatus.FunctionInstanceStatus> expectedFunctionInstanceStatusList = new ArrayList<>();
        FunctionStatus.FunctionInstanceStatus expectedFunctionInstanceStatus =
                new FunctionStatus.FunctionInstanceStatus();
        FunctionStatus.FunctionInstanceStatus.FunctionInstanceStatusData expectedFunctionInstanceStatusData =
                new FunctionStatus.FunctionInstanceStatus.FunctionInstanceStatusData();
        expectedFunctionInstanceStatusData.setWorkerId("test-pulsar");
        expectedFunctionInstanceStatusData.setRunning(true);
        expectedFunctionInstanceStatus.setStatus(expectedFunctionInstanceStatusData);
        expectedFunctionInstanceStatusList.add(expectedFunctionInstanceStatus);
        expectedFunctionStatus.setInstances(expectedFunctionInstanceStatusList);
        FunctionsImpl functions = spy(new FunctionsImpl(functionMeshProxyServiceSupplier));
        FunctionStatus functionStatus = functions.getFunctionStatus(
                tenant, namespace, name, null, null, null);
        String jsonData = json.getGson().toJson(functionStatus);
        String expectedData = json.getGson().toJson(expectedFunctionStatus);
        Assert.assertEquals(jsonData, expectedData);
    }

    @Test
    public void registerFunctionTest() throws ApiException, IOException {
        String testBody = "{\n" +
                "  \"apiVersion\": \"cloud.streamnative.io/v1alpha1\",\n" +
                "  \"kind\": \"Function\",\n" +
                "  \"metadata\": {\n" +
                "    \"creationTimestamp\": \"2021-01-19T13:19:17Z\",\n" +
                "    \"generation\": 2,\n" +
                "    \"name\": \"word-count\",\n" +
                "    \"namespace\": \"default\",\n" +
                "    \"resourceVersion\": \"24794021\",\n" +
                "    \"selfLink\": \"/apis/cloud.streamnative.io/v1alpha1/namespaces/default/functions/word-count\",\n" +
                "    \"uid\": \"b9e3ada1-b945-4d70-901c-00d7c7a7b0af\"\n" +
                "  },\n" +
                "  \"spec\": {\n" +
                "    \"className\": \"org.example.functions.WordCountFunction\",\n" +
                "    \"clusterName\": \"test-pulsar\",\n" +
                "    \"forwardSourceMessageProperty\": true,\n" +
                "    \"input\": {\n" +
                "      \"topics\": [\n" +
                "        \"persistent://public/default/sentences\"\n" +
                "      ]\n" +
                "    },\n" +
                "    \"java\": {\n" +
                "      \"jar\": \"word-count.jar\",\n" +
                "      \"jarLocation\": \"public/default/word-count\"\n" +
                "    },\n" +
                "    \"maxReplicas\": 1,\n" +
                "    \"output\": {\n" +
                "      \"topic\": \"persistent://public/default/count\"\n" +
                "    },\n" +
                "    \"pulsar\": {\n" +
                "      \"pulsarConfig\": \"test-pulsar\"\n" +
                "    },\n" +
                "    \"replicas\": 1,\n" +
                "    \"resources\": {\n" +
                "      \"limits\": {\n" +
                "        \"cpu\": \"0.1\",\n" +
                "        \"memory\": \"1G\"\n" +
                "      },\n" +
                "      \"requests\": {\n" +
                "        \"cpu\": \"0.1\",\n" +
                "        \"memory\": \"1G\"\n" +
                "      }\n" +
                "    },\n" +
                "    \"sinkType\": \"java.lang.String\",\n" +
                "    \"sourceType\": \"java.lang.String\"\n" +
                "  }\n" +
                "}";
        FunctionMeshProxyService functionMeshProxyService = PowerMockito.mock(FunctionMeshProxyService.class);
        Supplier<FunctionMeshProxyService> functionMeshProxyServiceSupplier = () -> functionMeshProxyService;
        CustomObjectsApi customObjectsApi = PowerMockito.mock(CustomObjectsApi.class);
        PowerMockito.when(functionMeshProxyService.getCustomObjectsApi()).thenReturn(customObjectsApi);
        Call call = PowerMockito.mock(Call.class);
        Response response = PowerMockito.mock(Response.class);
        ResponseBody responseBody = PowerMockito.mock(RealResponseBody.class);
        ApiClient apiClient = PowerMockito.mock(ApiClient.class);

        String tenant = "public";
        String namespace = "default";
        String functionName = "word-count";
        String group = "cloud.streamnative.io";
        String plural = "functions";
        String version = "v1alpha1";
        String kind = "Function";

        FunctionConfig functionConfig = new FunctionConfig();
        functionConfig.setName(functionName);
        functionConfig.setTenant(tenant);
        functionConfig.setNamespace(tenant);
        functionConfig.setClassName("org.example.functions.WordCountFunction");
        Map<String, ConsumerConfig> inputSpecs = new HashMap<>();
        ConsumerConfig sourceVaule = new ConsumerConfig();
        sourceVaule.setSerdeClassName("java.lang.String");
        inputSpecs.put("source", sourceVaule);
        functionConfig.setInputSpecs(inputSpecs);
        functionConfig.setOutputSerdeClassName("java.lang.String");
        functionConfig.setParallelism(1);
        List<String> inputs = new ArrayList<>();
        inputs.add("persistent://public/default/sentences");
        functionConfig.setInputs(inputs);
        functionConfig.setOutput("persistent://public/default/count");
        Resources resources = new Resources();
        resources.setCpu(0.1);
        resources.setRam(1L);
        functionConfig.setResources(resources);
        Map<String, Object> userConfig = new HashMap<>();
        userConfig.put("clusterName", "test-pulsar");
        functionConfig.setUserConfig(userConfig);
        functionConfig.setJar("word-count.jar");

        V1alpha1Function v1alpha1Function = new V1alpha1Function();
        v1alpha1Function.setKind(kind);
        v1alpha1Function.setApiVersion(String.format("%s/%s", group, version));

        V1ObjectMeta v1ObjectMeta = new V1ObjectMeta();
        v1ObjectMeta.setName(functionConfig.getName());
        v1ObjectMeta.setNamespace(functionConfig.getNamespace());
        v1alpha1Function.setMetadata(v1ObjectMeta);

        V1alpha1FunctionSpec v1alpha1FunctionSpec = new V1alpha1FunctionSpec();
        v1alpha1FunctionSpec.setClassName(functionConfig.getClassName());

        ConsumerConfig consumerConfig = functionConfig.getInputSpecs().get("source");
        v1alpha1FunctionSpec.setSourceType(consumerConfig.getSerdeClassName());
        v1alpha1FunctionSpec.setSinkType(functionConfig.getOutputSerdeClassName());

        v1alpha1FunctionSpec.setForwardSourceMessageProperty(functionConfig.getForwardSourceMessageProperty());
        v1alpha1FunctionSpec.setMaxPendingAsyncRequests(functionConfig.getMaxPendingAsyncRequests());

        Integer parallelism = functionConfig.getParallelism() == null ? 1 : functionConfig.getParallelism();
        v1alpha1FunctionSpec.setReplicas(parallelism);
        v1alpha1FunctionSpec.setMaxReplicas(parallelism);

        v1alpha1FunctionSpec.setLogTopic(functionConfig.getLogTopic());

        V1alpha1FunctionSpecInput v1alpha1FunctionSpecInput = new V1alpha1FunctionSpecInput();
        v1alpha1FunctionSpecInput.setTopics(new ArrayList<>(functionConfig.getInputs()));
        v1alpha1FunctionSpec.setInput(v1alpha1FunctionSpecInput);

        V1alpha1FunctionSpecOutput v1alpha1FunctionSpecOutput = new V1alpha1FunctionSpecOutput();
        v1alpha1FunctionSpecOutput.setTopic(functionConfig.getOutput());
        v1alpha1FunctionSpec.setOutput(v1alpha1FunctionSpecOutput);

        V1alpha1FunctionSpecResources v1alpha1FunctionSpecResources = new V1alpha1FunctionSpecResources();
        Map<String, Object> limits = new HashMap<>();
        Double cpu = functionConfig.getResources().getCpu();
        Long memory = functionConfig.getResources().getRam();
        String cpuValue = cpu.toString();
        String memoryValue = memory.toString() + "G";
        limits.put("cpu", cpuValue);
        limits.put("memory", memoryValue);
        Map<String, Object> requests = new HashMap<>();
        limits.put("cpu", cpuValue);
        limits.put("memory", memoryValue);
        v1alpha1FunctionSpecResources.setLimits(limits);
        v1alpha1FunctionSpecResources.setRequests(requests);
        v1alpha1FunctionSpec.setResources(v1alpha1FunctionSpecResources);

        V1alpha1FunctionSpecPulsar v1alpha1FunctionSpecPulsar = new V1alpha1FunctionSpecPulsar();
        v1alpha1FunctionSpecPulsar.setPulsarConfig(functionName);
        v1alpha1FunctionSpec.setPulsar(v1alpha1FunctionSpecPulsar);

        String location = String.format("%s/%s/%s", tenant, namespace, functionName);
        V1alpha1FunctionSpecJava v1alpha1FunctionSpecJava = new V1alpha1FunctionSpecJava();
        Path path = Paths.get(functionConfig.getJar());
        v1alpha1FunctionSpecJava.setJar(path.getFileName().toString());
        v1alpha1FunctionSpecJava.setJarLocation(location);
        v1alpha1FunctionSpec.setJava(v1alpha1FunctionSpecJava);

        Object clusterName = functionConfig.getUserConfig().get("clusterName");
        v1alpha1FunctionSpec.setClusterName(clusterName.toString());

        v1alpha1FunctionSpec.setAutoAck(functionConfig.getAutoAck());

        v1alpha1Function.setSpec(v1alpha1FunctionSpec);

        PowerMockito.when(functionMeshProxyService.getCustomObjectsApi()
                .createNamespacedCustomObjectCall(
                        group,
                        version,
                        namespace,
                        plural,
                        v1alpha1Function,
                        null,
                        null,
                        null,
                        null
                )).thenReturn(call);
        PowerMockito.when(call.execute()).thenReturn(response);
        PowerMockito.when(response.isSuccessful()).thenReturn(true);
        PowerMockito.when(response.body()).thenReturn(responseBody);
        PowerMockito.when(responseBody.string()).thenReturn(testBody);
        PowerMockito.when(functionMeshProxyService.getApiClient()).thenReturn(apiClient);
        JSON json = new JSON();
        PowerMockito.when(apiClient.getJSON()).thenReturn(json);
        FunctionsImpl functions = spy(new FunctionsImpl(functionMeshProxyServiceSupplier));
        try {
            functions.registerFunction(tenant, namespace, functionName, null, null, null, functionConfig, null, null);
        } catch (Exception exception) {
            Assert.fail("No exception, but got error message:" + exception.getMessage());
        }
    }
}
