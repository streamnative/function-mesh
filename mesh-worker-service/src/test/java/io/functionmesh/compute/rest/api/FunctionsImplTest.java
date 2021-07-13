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
import io.functionmesh.compute.functions.models.V1alpha1Function;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecPod;
import io.functionmesh.compute.testdata.Generate;
import io.functionmesh.compute.util.CommonUtil;
import io.functionmesh.compute.util.FunctionsUtil;
import io.functionmesh.compute.util.KubernetesUtils;
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
import org.apache.commons.codec.digest.DigestUtils;
import org.apache.pulsar.client.admin.PulsarAdmin;
import org.apache.pulsar.client.admin.PulsarAdminException;
import org.apache.pulsar.client.admin.Tenants;
import org.apache.pulsar.common.functions.FunctionConfig;
import org.apache.pulsar.common.policies.data.FunctionStatus;
import org.apache.pulsar.functions.proto.InstanceCommunication;
import org.apache.pulsar.functions.proto.InstanceControlGrpc;
import org.apache.pulsar.functions.worker.WorkerConfig;
import org.junit.Assert;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.powermock.api.mockito.PowerMockito;
import org.powermock.core.classloader.annotations.PowerMockIgnore;
import org.powermock.core.classloader.annotations.PrepareForTest;
import org.powermock.modules.junit4.PowerMockRunner;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;
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
        CommonUtil.class,
        InstanceControlGrpc.InstanceControlFutureStub.class})
@PowerMockIgnore({"javax.management.*"})
public class FunctionsImplTest {
    @Test
    public void getFunctionStatusTest() throws ApiException, IOException {
        String testBody = "{\n" +
                "  \"apiVersion\": \"compute.functionmesh.io/v1alpha1\",\n" +
                "  \"kind\": \"Function\",\n" +
                "  \"metadata\": {\n" +
                "    \"creationTimestamp\": \"2020-11-27T08:08:32Z\",\n" +
                "    \"generation\": 1,\n" +
                "    \"name\": \"functionmesh-sample-ex1\",\n" +
                "    \"namespace\": \"default\",\n" +
                "    \"ownerReferences\": [\n" +
                "      {\n" +
                "        \"apiVersion\": \"compute.functionmesh.io/v1alpha1\",\n" +
                "        \"blockOwnerDeletion\": true,\n" +
                "        \"controller\": true,\n" +
                "        \"kind\": \"FunctionMesh\",\n" +
                "        \"name\": \"functionmesh-sample\",\n" +
                "        \"uid\": \"5ef04d92-c8bf-40ac-ad9e-4a02bb75cb8b\"\n" +
                "      }\n" +
                "    ],\n" +
                "    \"resourceVersion\": \"899291\",\n" +
                "    \"selfLink\": \"/apis/compute.functionmesh.io/v1alpha1/namespaces/default/functions/functionmesh-sample-ex1\",\n" +
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

        MeshWorkerService meshWorkerService = PowerMockito.mock(MeshWorkerService.class);
        Supplier<MeshWorkerService> meshWorkerServiceSupplier = () -> meshWorkerService;
        CustomObjectsApi customObjectsApi = PowerMockito.mock(CustomObjectsApi.class);
        CoreV1Api coreV1Api = PowerMockito.mock(CoreV1Api.class);
        AppsV1Api appsV1Api = PowerMockito.mock(AppsV1Api.class);
        PowerMockito.when(meshWorkerService.getCustomObjectsApi()).thenReturn(customObjectsApi);
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

        String tenant = "public";
        String namespace = "default";
        String name = "test";
        String group = "compute.functionmesh.io";
        String plural = "functions";
        String version = "v1alpha1";
        String hashName = CommonUtil.generateObjectName(meshWorkerService, tenant, namespace, name);
        String jobName = CommonUtil.makeJobName(name, CommonUtil.COMPONENT_SINK);

        PowerMockito.when(meshWorkerService.getCustomObjectsApi()
                .getNamespacedCustomObjectCall(
                        group, version, namespace, plural,
                        CommonUtil.createObjectName("test-pulsar", tenant, namespace, name),
                        null)).thenReturn(call);
        PowerMockito.when(call.execute()).thenReturn(response);
        PowerMockito.when(response.isSuccessful()).thenReturn(true);
        PowerMockito.when(response.body()).thenReturn(responseBody);
        PowerMockito.when(responseBody.string()).thenReturn(testBody);
        PowerMockito.when(meshWorkerService.getApiClient()).thenReturn(apiClient);
        JSON json = new JSON();
        PowerMockito.when(apiClient.getJSON()).thenReturn(json);

        V1StatefulSet v1StatefulSet = PowerMockito.mock(V1StatefulSet.class);
        PowerMockito.when(appsV1Api.readNamespacedStatefulSet(any(), any(), any(), any(), any())).thenReturn(v1StatefulSet);

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
        PowerMockito.when(podV1ObjectMeta.getName()).thenReturn(hashName + "-function-0");
        V1PodStatus podStatus = PowerMockito.mock(V1PodStatus.class);
        PowerMockito.when(pod.getStatus()).thenReturn(podStatus);
        PowerMockito.when(podStatus.getPhase()).thenReturn("Running");
        PowerMockito.when(podStatus.getContainerStatuses()).thenReturn(new ArrayList<>());
        InstanceCommunication.FunctionStatus.Builder builder = InstanceCommunication.FunctionStatus.newBuilder();
        builder.setRunning(true);
        PowerMockito.mockStatic(InstanceControlGrpc.InstanceControlFutureStub.class);
        PowerMockito.stub(PowerMockito.method(CommonUtil.class, "getFunctionStatusAsync")).toReturn(CompletableFuture.completedFuture(builder.build()));

        FunctionsImpl functions = spy(new FunctionsImpl(meshWorkerServiceSupplier));
        FunctionStatus functionStatus = functions.getFunctionStatus(
                tenant, namespace, name, null, null, null);

        FunctionStatus expectedFunctionStatus = new FunctionStatus();
        FunctionStatus.FunctionInstanceStatus expectedFunctionInstanceStatus =
                new FunctionStatus.FunctionInstanceStatus();
        FunctionStatus.FunctionInstanceStatus.FunctionInstanceStatusData expectedFunctionInstanceStatusData =
                new FunctionStatus.FunctionInstanceStatus.FunctionInstanceStatusData();
        expectedFunctionInstanceStatusData.setWorkerId("test-pulsar");
        expectedFunctionInstanceStatusData.setRunning(true);
        expectedFunctionInstanceStatusData.setError("");
        expectedFunctionInstanceStatusData.setLatestUserExceptions(Collections.emptyList());
        expectedFunctionInstanceStatusData.setLatestSystemExceptions(Collections.emptyList());
        expectedFunctionInstanceStatus.setStatus(expectedFunctionInstanceStatusData);
        expectedFunctionStatus.addInstance(expectedFunctionInstanceStatus);
        expectedFunctionStatus.setNumInstances(1);
        expectedFunctionStatus.setNumRunning(1);

        Assert.assertEquals(expectedFunctionStatus, functionStatus);
    }

    @Test
    public void registerFunctionTest() throws ApiException, IOException, PulsarAdminException {
        String testBody = "{\n" +
                "  \"apiVersion\": \"compute.functionmesh.io/v1alpha1\",\n" +
                "  \"kind\": \"Function\",\n" +
                "  \"metadata\": {\n" +
                "    \"creationTimestamp\": \"2021-01-19T13:19:17Z\",\n" +
                "    \"generation\": 2,\n" +
                "    \"name\": \"word-count\",\n" +
                "    \"namespace\": \"default\",\n" +
                "    \"resourceVersion\": \"24794021\",\n" +
                "    \"selfLink\": \"/apis/compute.functionmesh.io/v1alpha1/namespaces/default/functions/word-count\",\n" +
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
        MeshWorkerService meshWorkerService = PowerMockito.mock(MeshWorkerService.class);
        Supplier<MeshWorkerService> meshWorkerServiceSupplier = () -> meshWorkerService;
        CustomObjectsApi customObjectsApi = PowerMockito.mock(CustomObjectsApi.class);
        PowerMockito.when(meshWorkerService.getCustomObjectsApi()).thenReturn(customObjectsApi);
        WorkerConfig workerConfig = PowerMockito.mock(WorkerConfig.class);
        PowerMockito.when(meshWorkerService.getWorkerConfig()).thenReturn(workerConfig);
        PowerMockito.when(workerConfig.isAuthorizationEnabled()).thenReturn(false);
        PowerMockito.when(workerConfig.isAuthenticationEnabled()).thenReturn(false);
        PulsarAdmin pulsarAdmin = PowerMockito.mock(PulsarAdmin.class);
        PowerMockito.when(meshWorkerService.getBrokerAdmin()).thenReturn(pulsarAdmin);
        Tenants tenants = PowerMockito.mock(Tenants.class);
        PowerMockito.when(pulsarAdmin.tenants()).thenReturn(tenants);
        Call call = PowerMockito.mock(Call.class);
        Response response = PowerMockito.mock(Response.class);
        ResponseBody responseBody = PowerMockito.mock(RealResponseBody.class);
        ApiClient apiClient = PowerMockito.mock(ApiClient.class);

        String tenant = "public";
        String namespace = "default";
        String functionName = "word-count";
        String group = "compute.functionmesh.io";
        String plural = "functions";
        String version = "v1alpha1";
        String kind = "Function";

        FunctionConfig functionConfig = Generate.CreateJavaFunctionConfig(tenant, namespace, functionName);

        PowerMockito.when(tenants.getTenantInfo(tenant)).thenReturn(null);

        V1alpha1Function v1alpha1Function = FunctionsUtil.createV1alpha1FunctionFromFunctionConfig(kind, group,
                version, functionName, null, functionConfig, Collections.emptyMap(), null);

        String clusterName = "test-pulsar";
        Map<String, String> customLabels = Maps.newHashMap();
        customLabels.put("pulsar-cluster", clusterName);
        customLabels.put("pulsar-tenant", tenant);
        customLabels.put("pulsar-namespace", namespace);
        customLabels.put("pulsar-component", functionName);
        V1alpha1FunctionSpecPod pod = new V1alpha1FunctionSpecPod();
        pod.setLabels(customLabels);
        v1alpha1Function.getSpec().pod(pod);
        v1alpha1Function.getMetadata().setLabels(customLabels);
        PowerMockito.when(meshWorkerService.getCustomObjectsApi()
                .createNamespacedCustomObjectCall(
                        group,
                        version,
                        KubernetesUtils.getNamespace(),
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
        PowerMockito.when(meshWorkerService.getApiClient()).thenReturn(apiClient);
        JSON json = new JSON();
        PowerMockito.when(apiClient.getJSON()).thenReturn(json);
        FunctionsImpl functions = spy(new FunctionsImpl(meshWorkerServiceSupplier));
        try {
            functions.registerFunction(
                    tenant,
                    namespace,
                    functionName,
                    null,
                    null,
                    null,
                    functionConfig,
                    null,
                    null);
        } catch (Exception exception) {
            Assert.fail("Expected no exception to be thrown but got exception: " + exception);
        }
    }

    @Test
    public void updateFunctionTest() throws ApiException, IOException {
        String getBody = "{\n" +
                "  \"apiVersion\": \"compute.functionmesh.io/v1alpha1\",\n" +
                "  \"kind\": \"Function\",\n" +
                "  \"metadata\": {\n" +
                "          \"labels\": {\n" +
                "                     \"pulsar-namespace\": \"default\",\n" +
                "                     \"pulsar-tenant\": \"public\",\n" +
                "                     \"pulsar-component\": \"word-count\",\n" +
                "                     \"pulsar-cluster\": \"test-pulsar\"" +
                "            },\n" +
                "    \"creationTimestamp\": \"2021-01-19T13:19:17Z\",\n" +
                "    \"generation\": 2,\n" +
                "    \"name\": \"word-count\",\n" +
                "    \"namespace\": \"default\",\n" +
                "    \"resourceVersion\": \"24794021\",\n" +
                "    \"selfLink\": \"/apis/compute.functionmesh.io/v1alpha1/namespaces/default/functions/word-count\",\n" +
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
        String replaceBody = "{\n" +
                "  \"apiVersion\": \"compute.functionmesh.io/v1alpha1\",\n" +
                "  \"kind\": \"Function\",\n" +
                "  \"metadata\": {\n" +
                "    \"creationTimestamp\": \"2021-01-19T13:19:17Z\",\n" +
                "    \"generation\": 2,\n" +
                "    \"name\": \"word-count-640ae9e6\",\n" +
                "    \"namespace\": \"default\",\n" +
                "    \"resourceVersion\": \"27794021\",\n" +
                "    \"selfLink\": \"/apis/compute.functionmesh.io/v1alpha1/namespaces/default/functions/word-count\",\n" +
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
        String tenant = "public";
        String namespace = "default";
        String functionName = "word-count";
        String group = "compute.functionmesh.io";
        String plural = "functions";
        String version = "v1alpha1";
        String kind = "Function";

        MeshWorkerService meshWorkerService = PowerMockito.mock(MeshWorkerService.class);
        Supplier<MeshWorkerService> meshWorkerServiceSupplier = () -> meshWorkerService;
        CustomObjectsApi customObjectsApi = PowerMockito.mock(CustomObjectsApi.class);
        PowerMockito.when(meshWorkerService.getCustomObjectsApi()).thenReturn(customObjectsApi);
        WorkerConfig workerConfig = PowerMockito.mock(WorkerConfig.class);
        PowerMockito.when(meshWorkerService.getWorkerConfig()).thenReturn(workerConfig);
        PowerMockito.when(workerConfig.isAuthorizationEnabled()).thenReturn(false);
        PowerMockito.when(workerConfig.isAuthenticationEnabled()).thenReturn(false);

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

        PowerMockito.when(meshWorkerService.getCustomObjectsApi()
                .getNamespacedCustomObjectCall(
                        group,
                        version,
                        namespace,
                        plural,
                        "word-count-640ae9e6",
                        null
                )).thenReturn(getCall);

        FunctionConfig functionConfig = Generate.CreateJavaFunctionConfig(tenant, namespace, functionName);

        PowerMockito.when(meshWorkerService.getCustomObjectsApi()
                .replaceNamespacedCustomObjectCall(
                        anyString(),
                        anyString(),
                        anyString(),
                        anyString(),
                        anyString(),
                        anyObject(),
                        anyString(),
                        anyString(),
                        anyObject()
                )).thenReturn(getCall);

        FunctionsImpl functions = spy(new FunctionsImpl(meshWorkerServiceSupplier));

        try {
            functions.updateFunction(
                    tenant,
                    namespace,
                    functionName,
                    null,
                    null,
                    null,
                    functionConfig,
                    null,
                    null,
                    null);
        } catch (Exception exception) {
            Assert.fail("Expected no exception to be thrown but got exception: " + exception);
        }
    }

    @Test
    public void deregisterFunctionTest() throws ApiException, IOException {

        String testBody = "{\n" +
                "  \"apiVersion\": \"compute.functionmesh.io/v1alpha1\",\n" +
                "  \"kind\": \"Function\",\n" +
                "  \"metadata\": {\n" +
                "    \"creationTimestamp\": \"2021-01-19T13:19:17Z\",\n" +
                "    \"generation\": 2,\n" +
                "    \"name\": \"word-count\",\n" +
                "    \"namespace\": \"default\",\n" +
                "    \"resourceVersion\": \"24794021\",\n" +
                "    \"selfLink\": \"/apis/compute.functionmesh.io/v1alpha1/namespaces/default/functions/word-count\",\n" +
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
                "        \"memory\": \"1024\"\n" +
                "      },\n" +
                "      \"requests\": {\n" +
                "        \"cpu\": \"0.1\",\n" +
                "        \"memory\": \"1024\"\n" +
                "      }\n" +
                "    },\n" +
                "    \"sinkType\": \"java.lang.String\",\n" +
                "    \"sourceType\": \"java.lang.String\"\n" +
                "  }\n" +
                "}";

        MeshWorkerService meshWorkerService = PowerMockito.mock(MeshWorkerService.class);
        Supplier<MeshWorkerService> meshWorkerServiceSupplier = () -> meshWorkerService;
        CustomObjectsApi customObjectsApi = PowerMockito.mock(CustomObjectsApi.class);
        PowerMockito.when(meshWorkerService.getCustomObjectsApi()).thenReturn(customObjectsApi);
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
        String functionName = "word-count";
        String group = "compute.functionmesh.io";
        String plural = "functions";
        String version = "v1alpha1";

        Response functionInfoResponse = PowerMockito.mock(Response.class);
        Call functionInfoCall = PowerMockito.mock(Call.class);
        PowerMockito.when(meshWorkerService.getCustomObjectsApi()
                .getNamespacedCustomObjectCall(
                        group,
                        version,
                        namespace,
                        plural,
                        functionName,
                        null)).thenReturn(functionInfoCall);
        PowerMockito.when(functionInfoCall.execute()).thenReturn(functionInfoResponse);
        PowerMockito.when(functionInfoResponse.isSuccessful()).thenReturn(true);
        PowerMockito.when(functionInfoResponse.body()).thenReturn(responseBody);

        PowerMockito.when(meshWorkerService.getCustomObjectsApi()
                .deleteNamespacedCustomObjectCall(
                        group,
                        version,
                        namespace,
                        plural,
                        "word-count-640ae9e6",
                        null,
                        null,
                        null,
                        null,
                        null,
                        null
                )).thenReturn(call);
        PowerMockito.when(call.execute()).thenReturn(response);
        PowerMockito.when(response.isSuccessful()).thenReturn(true);
        PowerMockito.when(response.body()).thenReturn(responseBody);
        PowerMockito.when(responseBody.string()).thenReturn(testBody);
        PowerMockito.when(meshWorkerService.getApiClient()).thenReturn(apiClient);
        JSON json = new JSON();
        PowerMockito.when(apiClient.getJSON()).thenReturn(json);

        CoreV1Api coreV1Api = PowerMockito.mock(CoreV1Api.class);
        PowerMockito.when(meshWorkerService.getCoreV1Api()).thenReturn(coreV1Api);
        Call deleteAuthSecretCall = PowerMockito.mock(Call.class);
        PowerMockito.when(coreV1Api.deleteNamespacedSecretCall(
                KubernetesUtils.getUniqueSecretName(
                        "function",
                        "auth",
                        DigestUtils.sha256Hex(
                                KubernetesUtils.getSecretName(
                                        "test-pulsar",
                                        tenant, namespace, "word-count"))),
                namespace,
                null,
                null,
                30,
                false,
                null,
                null,
                null
        )).thenReturn(deleteAuthSecretCall);
        PowerMockito.when(deleteAuthSecretCall.execute()).thenReturn(response);
        Call deleteTlsSecretCall = PowerMockito.mock(Call.class);
        PowerMockito.when(coreV1Api.deleteNamespacedSecretCall(
                KubernetesUtils.getUniqueSecretName(
                        "function",
                        "tls",
                        DigestUtils.sha256Hex(
                                KubernetesUtils.getSecretName(
                                        "test-pulsar",
                                        tenant, namespace, "word-count"))),
                namespace,
                null,
                null,
                30,
                false,
                null,
                null,
                null
        )).thenReturn(deleteTlsSecretCall);
        PowerMockito.when(deleteTlsSecretCall.execute()).thenReturn(response);
        FunctionsImpl functions = spy(new FunctionsImpl(meshWorkerServiceSupplier));
        try {
            functions.deregisterFunction(tenant, namespace, functionName, null, null);
        } catch (Exception exception) {
            Assert.fail("Expected no exception to be thrown but got exception: " + exception);
        }
    }

    @Test
    public void getFunctionInfoTest() throws ApiException, IOException {
        String testBody = "{\n" +
                "  \"apiVersion\": \"compute.functionmesh.io/v1alpha1\",\n" +
                "  \"kind\": \"Function\",\n" +
                "  \"metadata\": {\n" +
                "    \"creationTimestamp\": \"2021-01-19T13:19:17Z\",\n" +
                "    \"generation\": 2,\n" +
                "    \"name\": \"word-count\",\n" +
                "    \"namespace\": \"default\",\n" +
                "    \"resourceVersion\": \"24794021\",\n" +
                "    \"selfLink\": \"/apis/compute.functionmesh.io/v1alpha1/namespaces/default/functions/word-count\",\n" +
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
                "        \"memory\": \"1024\"\n" +
                "      },\n" +
                "      \"requests\": {\n" +
                "        \"cpu\": \"0.1\",\n" +
                "        \"memory\": \"1024\"\n" +
                "      }\n" +
                "    },\n" +
                "    \"sinkType\": \"java.lang.String\",\n" +
                "    \"sourceType\": \"java.lang.String\"\n" +
                "  }\n" +
                "}";
        MeshWorkerService meshWorkerService = PowerMockito.mock(MeshWorkerService.class);
        Supplier<MeshWorkerService> meshWorkerServiceSupplier = () -> meshWorkerService;
        CustomObjectsApi customObjectsApi = PowerMockito.mock(CustomObjectsApi.class);
        PowerMockito.when(meshWorkerService.getCustomObjectsApi()).thenReturn(customObjectsApi);
        WorkerConfig workerConfig = PowerMockito.mock(WorkerConfig.class);
        PowerMockito.when(meshWorkerService.getWorkerConfig()).thenReturn(workerConfig);
        PowerMockito.when(workerConfig.isAuthorizationEnabled()).thenReturn(false);
        PowerMockito.when(workerConfig.isAuthenticationEnabled()).thenReturn(false);
        PowerMockito.when(workerConfig.getPulsarFunctionsCluster()).thenReturn("test-pulsar");
        Call call = PowerMockito.mock(Call.class);
        Response response = PowerMockito.mock(Response.class);
        ResponseBody responseBody = PowerMockito.mock(RealResponseBody.class);
        ApiClient apiClient = PowerMockito.mock(ApiClient.class);
        JSON json = new JSON();
        PowerMockito.when(apiClient.getJSON()).thenReturn(json);

        String tenant = "public";
        String namespace = "default";
        String functionName = "word-count";
        String group = "compute.functionmesh.io";
        String plural = "functions";
        String version = "v1alpha1";

        PowerMockito.when(meshWorkerService.getCustomObjectsApi()
                .getNamespacedCustomObjectCall(
                        group,
                        version,
                        namespace,
                        plural,
                        CommonUtil.createObjectName("test-pulsar", tenant, namespace, functionName),
                        null
                )).thenReturn(call);
        PowerMockito.when(call.execute()).thenReturn(response);
        PowerMockito.when(response.isSuccessful()).thenReturn(true);
        PowerMockito.when(response.body()).thenReturn(responseBody);
        PowerMockito.when(responseBody.string()).thenReturn(testBody);
        PowerMockito.when(meshWorkerService.getApiClient()).thenReturn(apiClient);
        FunctionsImpl functions = spy(new FunctionsImpl(meshWorkerServiceSupplier));
        FunctionConfig functionConfig = functions.getFunctionInfo(
                tenant, namespace, functionName, null, null);

        V1alpha1Function v1alpha1Function = json.getGson().fromJson(testBody, V1alpha1Function.class);
        FunctionConfig expectedFunctionConfig = FunctionsUtil.createFunctionConfigFromV1alpha1Function(tenant,
                namespace, functionName, v1alpha1Function);
        Assert.assertEquals(expectedFunctionConfig, functionConfig);
    }
}
