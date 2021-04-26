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

import io.functionmesh.compute.functions.models.V1alpha1Function;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpec;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecInput;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecJava;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecOutput;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecPulsar;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecResources;
import io.functionmesh.compute.models.CustomRuntimeOptions;
import io.kubernetes.client.openapi.models.V1ObjectMeta;
import org.apache.pulsar.common.functions.ConsumerConfig;
import org.apache.pulsar.common.functions.FunctionConfig;
import org.apache.pulsar.common.functions.Resources;
import org.junit.Assert;
import org.junit.Test;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

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
        String serdeClassName = "java.lang.String";
        int parallelism = 1;
        String input = "persistent://public/default/sentences";
        String output = "persistent://public/default/count";
        Double cpu = 0.1;
        Long ram = 1L;
        String clusterName = "test-pulsar";
        String jar = "word-count.jar";

        FunctionConfig functionConfig = new FunctionConfig();
        functionConfig.setName(functionName);
        functionConfig.setTenant(tenant);
        functionConfig.setNamespace(namespace);
        functionConfig.setClassName(className);
        Map<String, ConsumerConfig> inputSpecs = new HashMap<>();
        ConsumerConfig sourceValue = new ConsumerConfig();
        sourceValue.setSerdeClassName(serdeClassName);
        inputSpecs.put(FunctionsUtil.sourceKey, sourceValue);
        functionConfig.setInputSpecs(inputSpecs);
        functionConfig.setOutputSerdeClassName(serdeClassName);
        functionConfig.setParallelism(parallelism);
        List<String> inputs = new ArrayList<>();
        inputs.add(input);
        functionConfig.setInputs(inputs);
        functionConfig.setOutput(output);
        Resources resources = new Resources();
        resources.setCpu(cpu);
        resources.setRam(ram);
        functionConfig.setResources(resources);
        Map<String, Object> userConfig = new HashMap<>();
        userConfig.put(CustomRuntimeOptions.clusterNameKey, clusterName);
        userConfig.put(CustomRuntimeOptions.typeClassNameKey, serdeClassName);
        functionConfig.setUserConfig(userConfig);
        functionConfig.setJar(jar);

        V1alpha1Function v1alpha1Function = FunctionsUtil.createV1alpha1FunctionFromFunctionConfig(kind, group, version,
                functionName, null, functionConfig);

        V1alpha1Function expectedV1alpha1Function = new V1alpha1Function();
        expectedV1alpha1Function.setKind(kind);
        expectedV1alpha1Function.setApiVersion(String.format("%s/%s", group, version));

        V1ObjectMeta v1ObjectMeta = new V1ObjectMeta();
        v1ObjectMeta.setName(functionName);
        v1ObjectMeta.setNamespace(namespace);
        expectedV1alpha1Function.setMetadata(v1ObjectMeta);

        V1alpha1FunctionSpec v1alpha1FunctionSpec = new V1alpha1FunctionSpec();
        v1alpha1FunctionSpec.setClassName(className);
        v1alpha1FunctionSpec.setReplicas(parallelism);
        v1alpha1FunctionSpec.setMaxReplicas(parallelism);

        V1alpha1FunctionSpecInput v1alpha1FunctionSpecInput = new V1alpha1FunctionSpecInput();
        List<String> topics = new ArrayList<>();
        topics.add(input);
        v1alpha1FunctionSpecInput.setTopics(topics);
        v1alpha1FunctionSpecInput.setTypeClassName(serdeClassName);
        v1alpha1FunctionSpec.setInput(v1alpha1FunctionSpecInput);

        V1alpha1FunctionSpecOutput v1alpha1FunctionSpecOutput = new V1alpha1FunctionSpecOutput();
        v1alpha1FunctionSpecOutput.setTopic(output);
        v1alpha1FunctionSpecOutput.setTypeClassName(serdeClassName);
        v1alpha1FunctionSpec.setOutput(v1alpha1FunctionSpecOutput);

        V1alpha1FunctionSpecResources v1alpha1FunctionSpecResources = new V1alpha1FunctionSpecResources();
        Map<String, String> requestRes = new HashMap<>();
        String cpuValue = cpu.toString();
        String memoryValue = ram.toString();
        requestRes.put(FunctionsUtil.cpuKey, cpuValue);
        requestRes.put(FunctionsUtil.memoryKey, memoryValue);
        v1alpha1FunctionSpecResources.setLimits(requestRes);
        v1alpha1FunctionSpecResources.setRequests(requestRes);
        v1alpha1FunctionSpec.setResources(v1alpha1FunctionSpecResources);

        V1alpha1FunctionSpecPulsar v1alpha1FunctionSpecPulsar = new V1alpha1FunctionSpecPulsar();
        v1alpha1FunctionSpecPulsar.setPulsarConfig(functionName);
        v1alpha1FunctionSpec.setPulsar(v1alpha1FunctionSpecPulsar);

        String location = String.format("%s/%s/%s", tenant, namespace, functionName);
        V1alpha1FunctionSpecJava v1alpha1FunctionSpecJava = new V1alpha1FunctionSpecJava();
        v1alpha1FunctionSpecJava.setJar(jar);
        v1alpha1FunctionSpecJava.setJarLocation(location);
        v1alpha1FunctionSpec.setJava(v1alpha1FunctionSpecJava);

        v1alpha1FunctionSpec.setClusterName(clusterName);

        expectedV1alpha1Function.setSpec(v1alpha1FunctionSpec);

        Assert.assertEquals(expectedV1alpha1Function, v1alpha1Function);
    }

    @Test
    public void testCreateFunctionConfigFromV1alpha1Function() {
        String tenant = "public";
        String namespace = "default";
        String functionName = "word-count";
        String group = "compute.functionmesh.io";
        String version = "v1alpha1";
        String kind = "Function";
        String className = "org.example.functions.WordCountFunction";
        String serdeClassName = "java.lang.String";
        int parallelism = 1;
        String input = "persistent://public/default/sentences";
        String output = "persistent://public/default/count";
        Double cpu = 0.1;
        Long ram = 1L;
        String clusterName = "test-pulsar";
        String jar = "word-count.jar";

        V1alpha1Function v1alpha1Function = new V1alpha1Function();
        v1alpha1Function.setKind(kind);
        v1alpha1Function.setApiVersion(String.format("%s/%s", group, version));

        V1ObjectMeta v1ObjectMeta = new V1ObjectMeta();
        v1ObjectMeta.setName(functionName);
        v1ObjectMeta.setNamespace(namespace);
        v1alpha1Function.setMetadata(v1ObjectMeta);

        V1alpha1FunctionSpec v1alpha1FunctionSpec = new V1alpha1FunctionSpec();
        v1alpha1FunctionSpec.setClassName(className);
        v1alpha1FunctionSpec.setReplicas(parallelism);
        v1alpha1FunctionSpec.setMaxReplicas(parallelism);

        V1alpha1FunctionSpecInput v1alpha1FunctionSpecInput = new V1alpha1FunctionSpecInput();
        List<String> topics = new ArrayList<>();
        topics.add(input);
        v1alpha1FunctionSpecInput.setTopics(topics);
        v1alpha1FunctionSpecInput.setTypeClassName(serdeClassName);
        v1alpha1FunctionSpec.setInput(v1alpha1FunctionSpecInput);

        V1alpha1FunctionSpecOutput v1alpha1FunctionSpecOutput = new V1alpha1FunctionSpecOutput();
        v1alpha1FunctionSpecOutput.setTopic(output);
        v1alpha1FunctionSpecOutput.setTypeClassName(serdeClassName);
        v1alpha1FunctionSpec.setOutput(v1alpha1FunctionSpecOutput);

        V1alpha1FunctionSpecResources v1alpha1FunctionSpecResources = new V1alpha1FunctionSpecResources();
        Map<String, String> requestRes = new HashMap<>();
        String cpuValue = cpu.toString();
        String memoryValue = ram.toString();
        requestRes.put(FunctionsUtil.cpuKey, cpuValue);
        requestRes.put(FunctionsUtil.memoryKey, memoryValue);
        v1alpha1FunctionSpecResources.setLimits(requestRes);
        v1alpha1FunctionSpecResources.setRequests(requestRes);
        v1alpha1FunctionSpec.setResources(v1alpha1FunctionSpecResources);

        V1alpha1FunctionSpecPulsar v1alpha1FunctionSpecPulsar = new V1alpha1FunctionSpecPulsar();
        v1alpha1FunctionSpecPulsar.setPulsarConfig(functionName);
        v1alpha1FunctionSpec.setPulsar(v1alpha1FunctionSpecPulsar);

        String location = String.format("%s/%s/%s", tenant, namespace, functionName);
        V1alpha1FunctionSpecJava v1alpha1FunctionSpecJava = new V1alpha1FunctionSpecJava();
        v1alpha1FunctionSpecJava.setJar(jar);
        v1alpha1FunctionSpecJava.setJarLocation(location);
        v1alpha1FunctionSpec.setJava(v1alpha1FunctionSpecJava);

        v1alpha1FunctionSpec.setClusterName(clusterName);

        v1alpha1Function.setSpec(v1alpha1FunctionSpec);

        FunctionConfig expectedFunctionConfig = new FunctionConfig();
        expectedFunctionConfig.setName(functionName);
        expectedFunctionConfig.setTenant(tenant);
        expectedFunctionConfig.setNamespace(namespace);
        expectedFunctionConfig.setClassName(className);
        Map<String, ConsumerConfig> inputSpecs = new HashMap<>();
        ConsumerConfig sourceValue = new ConsumerConfig();
        sourceValue.setSerdeClassName(serdeClassName);
        inputSpecs.put(FunctionsUtil.sourceKey, sourceValue);
        expectedFunctionConfig.setInputSpecs(inputSpecs);
        expectedFunctionConfig.setOutputSerdeClassName(serdeClassName);
        expectedFunctionConfig.setParallelism(parallelism);
        List<String> inputs = new ArrayList<>();
        inputs.add(input);
        expectedFunctionConfig.setInputs(inputs);
        expectedFunctionConfig.setOutput(output);
        Resources resources = new Resources();
        resources.setCpu(cpu);
        resources.setRam(ram);
        expectedFunctionConfig.setResources(resources);
        Map<String, Object> userConfig = new HashMap<>();
        userConfig.put(CustomRuntimeOptions.clusterNameKey, clusterName);
        expectedFunctionConfig.setUserConfig(userConfig);
        expectedFunctionConfig.setJar(jar);

        FunctionConfig functionConfig = FunctionsUtil.createFunctionConfigFromV1alpha1Function(tenant, namespace,
                functionName, v1alpha1Function);

        Assert.assertEquals(expectedFunctionConfig, functionConfig);
    }
}
