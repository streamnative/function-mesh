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
import io.functionmesh.compute.testdata.Generate;
import java.util.Collections;
import org.apache.pulsar.common.functions.FunctionConfig;
import org.junit.Assert;
import org.junit.Test;

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
        String jar = "word-count.jar";

        FunctionConfig functionConfig = Generate.CreateJavaFunctionConfig(tenant, namespace, functionName);

        V1alpha1Function v1alpha1Function = FunctionsUtil.createV1alpha1FunctionFromFunctionConfig(kind, group, version,
                functionName, null, functionConfig, Collections.emptyMap());

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

        FunctionConfig functionConfig = Generate.CreateJavaFunctionConfig(tenant, namespace, functionName);

        V1alpha1Function v1alpha1Function = FunctionsUtil.createV1alpha1FunctionFromFunctionConfig(kind, group, version,
                functionName, null, functionConfig, Collections.emptyMap());

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
