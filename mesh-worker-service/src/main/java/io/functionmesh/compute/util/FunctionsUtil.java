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

import io.functionmesh.compute.functions.models.*;
import io.functionmesh.proxy.models.CustomRuntimeOptions;
import io.kubernetes.client.openapi.models.V1ObjectMeta;
import org.apache.commons.lang3.StringUtils;
import org.apache.pulsar.common.functions.ConsumerConfig;
import org.apache.pulsar.common.functions.FunctionConfig;
import org.apache.pulsar.common.functions.Resources;
import org.apache.pulsar.common.util.RestException;

import javax.ws.rs.core.Response;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.Map;

public class FunctionsUtil {
    public final static String cpuKey = "cpu";
    public final static String memoryKey = "memory";
    public final static String sourceKey = "source";

    public static V1alpha1Function createV1alpha1FunctionFromFunctionConfig(String kind, String group, String version
            , String functionName, String functionPkgUrl, FunctionConfig functionConfig) {
        V1alpha1Function v1alpha1Function = new V1alpha1Function();
        v1alpha1Function.setKind(kind);
        v1alpha1Function.setApiVersion(String.format("%s/%s", group, version));

        V1ObjectMeta v1ObjectMeta = new V1ObjectMeta();
        v1ObjectMeta.setName(functionConfig.getName());
        v1ObjectMeta.setNamespace(functionConfig.getNamespace());
        v1alpha1Function.setMetadata(v1ObjectMeta);

        V1alpha1FunctionSpec v1alpha1FunctionSpec = new V1alpha1FunctionSpec();
        v1alpha1FunctionSpec.setClassName(functionConfig.getClassName());

        V1alpha1FunctionSpecInput v1alpha1FunctionSpecInput = new V1alpha1FunctionSpecInput();

        if (functionConfig.getInputSpecs() != null) {
            ConsumerConfig consumerConfig = functionConfig.getInputSpecs().get(sourceKey);
            if (consumerConfig != null && !StringUtils.isBlank(consumerConfig.getSerdeClassName())) {
                v1alpha1FunctionSpecInput.setTypeClassName(consumerConfig.getSerdeClassName());
            }
        }
        if (!StringUtils.isBlank(functionConfig.getOutputSerdeClassName())) {
            v1alpha1FunctionSpecInput.setTypeClassName(functionConfig.getOutputSerdeClassName());
        }

        v1alpha1FunctionSpec.setForwardSourceMessageProperty(functionConfig.getForwardSourceMessageProperty());
        v1alpha1FunctionSpec.setMaxPendingAsyncRequests(functionConfig.getMaxPendingAsyncRequests());

        Integer parallelism = functionConfig.getParallelism() == null ? 1 : functionConfig.getParallelism();
        v1alpha1FunctionSpec.setReplicas(parallelism);
        v1alpha1FunctionSpec.setMaxReplicas(parallelism);

        v1alpha1FunctionSpec.setLogTopic(functionConfig.getLogTopic());


        v1alpha1FunctionSpecInput.setTopics(new ArrayList<>(functionConfig.getInputs()));
        v1alpha1FunctionSpec.setInput(v1alpha1FunctionSpecInput);

        V1alpha1FunctionSpecOutput v1alpha1FunctionSpecOutput = new V1alpha1FunctionSpecOutput();
        v1alpha1FunctionSpecOutput.setTopic(functionConfig.getOutput());
        v1alpha1FunctionSpec.setOutput(v1alpha1FunctionSpecOutput);

        V1alpha1FunctionSpecResources v1alpha1FunctionSpecResources = new V1alpha1FunctionSpecResources();
        if (functionConfig.getResources() == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "resources is not provided");
        }
        Double cpu = functionConfig.getResources().getCpu();
        if (functionConfig.getResources().getCpu() == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "resources.cpu is not provided");
        }
        Long memory = functionConfig.getResources().getRam();
        if (functionConfig.getResources().getRam() == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "resources.ram is not provided");
        }
        String cpuValue = cpu.toString();
        String memoryValue = memory.toString();
        Map<String, String> limits = new HashMap<>();
        limits.put(cpuKey, cpuValue);
        limits.put(memoryKey, memoryValue);

        v1alpha1FunctionSpecResources.setLimits(limits);
        v1alpha1FunctionSpecResources.setRequests(limits);
        v1alpha1FunctionSpec.setResources(v1alpha1FunctionSpecResources);

        V1alpha1FunctionSpecPulsar v1alpha1FunctionSpecPulsar = new V1alpha1FunctionSpecPulsar();
        v1alpha1FunctionSpecPulsar.setPulsarConfig(functionName);
        v1alpha1FunctionSpec.setPulsar(v1alpha1FunctionSpecPulsar);

        String location = String.format("%s/%s/%s", functionConfig.getTenant(), functionConfig.getNamespace(),
                functionName);
        if (StringUtils.isNotEmpty(functionPkgUrl)) {
            location = functionPkgUrl;
        }
        if (StringUtils.isNotEmpty(functionConfig.getJar())) {
            V1alpha1FunctionSpecJava v1alpha1FunctionSpecJava = new V1alpha1FunctionSpecJava();
            Path path = Paths.get(functionConfig.getJar());
            v1alpha1FunctionSpecJava.setJar(path.getFileName().toString());
            v1alpha1FunctionSpecJava.setJarLocation(location);
            v1alpha1FunctionSpec.setJava(v1alpha1FunctionSpecJava);
        } else if (StringUtils.isNotEmpty(functionConfig.getPy())) {
            V1alpha1FunctionSpecPython v1alpha1FunctionSpecPython = new V1alpha1FunctionSpecPython();
            Path path = Paths.get(functionConfig.getPy());
            v1alpha1FunctionSpecPython.setPy(path.getFileName().toString());
            v1alpha1FunctionSpecPython.setPyLocation(location);
            v1alpha1FunctionSpec.setPython(v1alpha1FunctionSpecPython);
        } else if (StringUtils.isNotEmpty(functionConfig.getGo())) {
            V1alpha1FunctionSpecGolang v1alpha1FunctionSpecGolang = new V1alpha1FunctionSpecGolang();
            Path path = Paths.get(functionConfig.getGo());
            v1alpha1FunctionSpecGolang.setGo(path.getFileName().toString());
            v1alpha1FunctionSpecGolang.setGoLocation(location);
            v1alpha1FunctionSpec.setGolang(v1alpha1FunctionSpecGolang);
        }

        if (functionConfig.getUserConfig() == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "userConfig is not provided");
        }
        Object clusterName = functionConfig.getUserConfig().get(CustomRuntimeOptions.clusterNameKey);
        if (clusterName == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "userConfig.clusterName is not provided");
        }

        Object typeClassNameObj = functionConfig.getUserConfig().get(CustomRuntimeOptions.typeClassNameKey);
        String typeClassName = typeClassNameObj != null ? typeClassNameObj.toString() : "";

        v1alpha1FunctionSpecInput.setTypeClassName(typeClassName);
        v1alpha1FunctionSpecOutput.setTypeClassName(typeClassName);

        v1alpha1FunctionSpec.setClusterName(clusterName.toString());
        v1alpha1FunctionSpec.setAutoAck(functionConfig.getAutoAck());

        v1alpha1Function.setSpec(v1alpha1FunctionSpec);

        return v1alpha1Function;
    }

    public static FunctionConfig createFunctionConfigFromV1alpha1Function(String tenant, String namespace,
                                                                          String functionName,
                                                                          V1alpha1Function v1alpha1Function) {
        FunctionConfig functionConfig = new FunctionConfig();

        functionConfig.setName(functionName);
        functionConfig.setNamespace(namespace);
        functionConfig.setTenant(tenant);

        V1alpha1FunctionSpec v1alpha1FunctionSpec = v1alpha1Function.getSpec();

        functionConfig.setClassName(v1alpha1FunctionSpec.getClassName());

        Map<String, ConsumerConfig> inputSpecs = new HashMap<>();
        ConsumerConfig sourceVaule = new ConsumerConfig();
        sourceVaule.setSerdeClassName(v1alpha1FunctionSpec.getInput().getTypeClassName());
        inputSpecs.put(sourceKey, sourceVaule);
        functionConfig.setInputSpecs(inputSpecs);
        functionConfig.setOutputSerdeClassName(v1alpha1FunctionSpec.getInput().getTypeClassName());

        functionConfig.setInputs(v1alpha1FunctionSpec.getInput().getTopics());
        functionConfig.setOutput(v1alpha1FunctionSpec.getOutput().getTopic());

        functionConfig.setParallelism(v1alpha1FunctionSpec.getReplicas());

        Resources resources = new Resources();
        Map<String, String> functionResource = v1alpha1FunctionSpec.getResources().getLimits();
        resources.setCpu(Double.parseDouble(functionResource.get(cpuKey)));
        resources.setRam(Long.parseLong(functionResource.get(memoryKey)));
        functionConfig.setResources(resources);

        Map<String, Object> userConfig = new HashMap<>();
        userConfig.put(CustomRuntimeOptions.clusterNameKey, v1alpha1FunctionSpec.getClusterName());
        functionConfig.setUserConfig(userConfig);

        if (v1alpha1FunctionSpec.getJava() != null) {
            functionConfig.setJar(v1alpha1FunctionSpec.getJava().getJar());
        }
        if (v1alpha1FunctionSpec.getGolang() != null) {
            functionConfig.setGo(v1alpha1FunctionSpec.getGolang().getGo());
        }
        if (v1alpha1FunctionSpec.getPython() != null) {
            functionConfig.setPy(v1alpha1FunctionSpec.getPython().getPy());
        }

        return functionConfig;
    }

}
