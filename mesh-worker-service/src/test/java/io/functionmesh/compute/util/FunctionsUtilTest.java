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
import io.functionmesh.compute.functions.models.V1alpha1Function;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpec;
import io.functionmesh.compute.testdata.Generate;

import java.util.Collections;

import io.kubernetes.client.openapi.apis.CustomObjectsApi;
import okhttp3.Response;
import okhttp3.internal.http.RealResponseBody;
import org.apache.pulsar.client.admin.PulsarAdmin;
import org.apache.pulsar.common.functions.FunctionConfig;
import org.apache.pulsar.functions.proto.InstanceControlGrpc;
import org.apache.pulsar.functions.runtime.kubernetes.KubernetesRuntimeFactoryConfig;
import org.apache.pulsar.functions.worker.WorkerConfig;
import org.junit.Assert;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.powermock.api.mockito.PowerMockito;
import org.powermock.core.classloader.annotations.PowerMockIgnore;
import org.powermock.core.classloader.annotations.PrepareForTest;
import org.powermock.modules.junit4.PowerMockRunner;

@RunWith(PowerMockRunner.class)
@PrepareForTest({
        Response.class,
        RealResponseBody.class,
        CommonUtil.class,
        FunctionsUtil.class,
        InstanceControlGrpc.InstanceControlFutureStub.class})
@PowerMockIgnore({"javax.management.*"})
public class FunctionsUtilTest {
    @Test
    public void testCreateV1alpha1FunctionFromFunctionConfig() {
        String tenant = "public";
        String namespace = "default";
        String functionName = "word-count";
        String group = "compute.functionmesh.io";
        String version = "v1alpha1";
        String kind = "Function";
        String className = "org.example.functions.WordCountFunction";
        String typeClassName = "java.lang.String";
        int parallelism = 1;
        String input = "persistent://public/default/sentences";
        String output = "persistent://public/default/count";
        String clusterName = "test-pulsar";
        String jar = "/pulsar/function-executable";

        MeshWorkerService meshWorkerService = PowerMockito.mock(MeshWorkerService.class);
        CustomObjectsApi customObjectsApi = PowerMockito.mock(CustomObjectsApi.class);
        PowerMockito.when(meshWorkerService.getCustomObjectsApi()).thenReturn(customObjectsApi);
        WorkerConfig workerConfig = PowerMockito.mock(WorkerConfig.class);
        KubernetesRuntimeFactoryConfig factoryConfig = PowerMockito.mock(KubernetesRuntimeFactoryConfig.class);
        PowerMockito.when(meshWorkerService.getWorkerConfig()).thenReturn(workerConfig);
        PowerMockito.when(meshWorkerService.getFactoryConfig()).thenReturn(factoryConfig);
        PowerMockito.when(factoryConfig.getExtraFunctionDependenciesDir()).thenReturn("");
        PowerMockito.when(workerConfig.isAuthorizationEnabled()).thenReturn(false);
        PowerMockito.when(workerConfig.isAuthenticationEnabled()).thenReturn(false);
        PowerMockito.when(workerConfig.getFunctionsWorkerServiceCustomConfigs()).thenReturn(Collections.emptyMap());
        PulsarAdmin pulsarAdmin = PowerMockito.mock(PulsarAdmin.class);
        PowerMockito.when(meshWorkerService.getBrokerAdmin()).thenReturn(pulsarAdmin);
        PowerMockito.stub(PowerMockito.method(FunctionsUtil.class, "downloadPackageFile")).toReturn(null);


        FunctionConfig functionConfig = Generate.CreateJavaFunctionWithPackageURLConfig(tenant, namespace, functionName);

        V1alpha1Function v1alpha1Function = FunctionsUtil.createV1alpha1FunctionFromFunctionConfig(kind, group, version,
                functionName, functionConfig.getJar(), functionConfig, null, meshWorkerService);

        Assert.assertEquals(v1alpha1Function.getKind(), kind);

        V1alpha1FunctionSpec v1alpha1FunctionSpec = v1alpha1Function.getSpec();
        Assert.assertEquals(v1alpha1FunctionSpec.getClassName(), className);
        Assert.assertEquals(v1alpha1FunctionSpec.getCleanupSubscription(), true);
        Assert.assertEquals(v1alpha1FunctionSpec.getReplicas().intValue(), parallelism);
        Assert.assertEquals(v1alpha1FunctionSpec.getInput().getTopics().get(0), input);
        Assert.assertEquals(v1alpha1FunctionSpec.getOutput().getTopic(), output);
        Assert.assertEquals(v1alpha1FunctionSpec.getPulsar().getPulsarConfig(),
                CommonUtil.getPulsarClusterConfigMapName(clusterName));
        Assert.assertEquals(v1alpha1FunctionSpec.getInput().getTypeClassName(), typeClassName);
        Assert.assertEquals(v1alpha1FunctionSpec.getOutput().getTypeClassName(), typeClassName);
        Assert.assertEquals(v1alpha1FunctionSpec.getJava().getJar(), jar);

    }

    @Test
    public void testCreateFunctionConfigFromV1alpha1Function() {
        String tenant = "public";
        String namespace = "default";
        String functionName = "word-count";
        String group = "compute.functionmesh.io";
        String version = "v1alpha1";
        String kind = "Function";

        MeshWorkerService meshWorkerService = PowerMockito.mock(MeshWorkerService.class);
        CustomObjectsApi customObjectsApi = PowerMockito.mock(CustomObjectsApi.class);
        PowerMockito.when(meshWorkerService.getCustomObjectsApi()).thenReturn(customObjectsApi);
        WorkerConfig workerConfig = PowerMockito.mock(WorkerConfig.class);
        KubernetesRuntimeFactoryConfig factoryConfig = PowerMockito.mock(KubernetesRuntimeFactoryConfig.class);
        PowerMockito.when(meshWorkerService.getWorkerConfig()).thenReturn(workerConfig);
        PowerMockito.when(meshWorkerService.getFactoryConfig()).thenReturn(factoryConfig);
        PowerMockito.when(factoryConfig.getExtraFunctionDependenciesDir()).thenReturn("");
        PowerMockito.when(workerConfig.isAuthorizationEnabled()).thenReturn(false);
        PowerMockito.when(workerConfig.isAuthenticationEnabled()).thenReturn(false);
        PowerMockito.when(workerConfig.getFunctionsWorkerServiceCustomConfigs()).thenReturn(Collections.emptyMap());
        PulsarAdmin pulsarAdmin = PowerMockito.mock(PulsarAdmin.class);
        PowerMockito.when(meshWorkerService.getBrokerAdmin()).thenReturn(pulsarAdmin);
        PowerMockito.stub(PowerMockito.method(FunctionsUtil.class, "downloadPackageFile")).toReturn(null);

        FunctionConfig functionConfig = Generate.CreateJavaFunctionWithPackageURLConfig(tenant, namespace, functionName);

        V1alpha1Function v1alpha1Function = FunctionsUtil.createV1alpha1FunctionFromFunctionConfig(kind, group, version,
                functionName, functionConfig.getJar(), functionConfig, null, meshWorkerService);

        FunctionConfig newFunctionConfig = FunctionsUtil.createFunctionConfigFromV1alpha1Function(tenant, namespace,
                functionName, v1alpha1Function);

        Assert.assertEquals(functionConfig.getClassName(), newFunctionConfig.getClassName());
        Assert.assertEquals(functionConfig.getJar(), newFunctionConfig.getJar());
        Assert.assertEquals(functionConfig.getPy(), newFunctionConfig.getPy());
        Assert.assertEquals(functionConfig.getGo(), newFunctionConfig.getGo());
        Assert.assertArrayEquals(functionConfig.getInputs().toArray(), newFunctionConfig.getInputs().toArray());
        Assert.assertEquals(functionConfig.getOutput(), newFunctionConfig.getOutput());
        Assert.assertEquals(functionConfig.getMaxMessageRetries(), newFunctionConfig.getMaxMessageRetries());
        Assert.assertEquals(functionConfig.getAutoAck(), newFunctionConfig.getAutoAck());
        Assert.assertEquals(functionConfig.getBatchBuilder(), newFunctionConfig.getBatchBuilder());
        Assert.assertEquals(functionConfig.getCustomRuntimeOptions(), newFunctionConfig.getCustomRuntimeOptions());
        Assert.assertEquals(functionConfig.getSubName(), newFunctionConfig.getSubName());
        Assert.assertEquals(functionConfig.getCleanupSubscription(), newFunctionConfig.getCleanupSubscription());
        Assert.assertEquals(functionConfig.getCustomSchemaInputs(), newFunctionConfig.getCustomSchemaInputs());
        Assert.assertEquals(functionConfig.getCustomSerdeInputs(), newFunctionConfig.getCustomSerdeInputs());
        Assert.assertEquals(functionConfig.getCustomSchemaOutputs(), newFunctionConfig.getCustomSchemaOutputs());

    }
}
