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
package io.functionmesh.compute.testdata;

import io.functionmesh.compute.util.FunctionsUtil;
import io.functionmesh.compute.models.CustomRuntimeOptions;
import org.apache.pulsar.common.functions.ConsumerConfig;
import org.apache.pulsar.common.functions.FunctionConfig;
import org.apache.pulsar.common.functions.Resources;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class Generate {
    public static FunctionConfig CreateJavaFunctionConfig(String tenant, String namespace, String functionName) {
        FunctionConfig functionConfig = new FunctionConfig();
        functionConfig.setName(functionName);
        functionConfig.setTenant(tenant);
        functionConfig.setNamespace(namespace);
        functionConfig.setClassName(String.format("org.example.functions.%s", functionName.toUpperCase()));
        Map<String, ConsumerConfig> inputSpecs = new HashMap<>();
        ConsumerConfig sourceValue = new ConsumerConfig();
        sourceValue.setSerdeClassName("java.lang.String");
        inputSpecs.put(FunctionsUtil.sourceKey, sourceValue);
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
        userConfig.put(CustomRuntimeOptions.clusterNameKey, String.format("%s-%s", tenant, namespace));
        userConfig.put(CustomRuntimeOptions.typeClassNameKey, "java.lang.String");
        functionConfig.setUserConfig(userConfig);
        functionConfig.setJar(String.format("%s.jar", functionName));
        return functionConfig;
    }

}
