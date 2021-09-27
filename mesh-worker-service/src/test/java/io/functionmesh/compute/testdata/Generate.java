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

import com.google.gson.Gson;
import io.functionmesh.compute.models.CustomRuntimeOptions;
import io.functionmesh.compute.util.FunctionsUtil;
import org.apache.pulsar.common.functions.ConsumerConfig;
import org.apache.pulsar.common.functions.FunctionConfig;
import org.apache.pulsar.common.functions.Resources;
import org.apache.pulsar.common.io.SinkConfig;
import org.apache.pulsar.common.io.SourceConfig;

import java.util.Collections;
import java.util.HashMap;
import java.util.Map;

public class Generate {
    public static String TEST_CLUSTER_NAME = "test-pulsar";

    public static FunctionConfig CreateJavaFunctionConfig(String tenant, String namespace, String functionName) {
        FunctionConfig functionConfig = new FunctionConfig();
        functionConfig.setName(functionName);
        functionConfig.setTenant(tenant);
        functionConfig.setNamespace(namespace);
        functionConfig.setClassName("org.example.functions.WordCountFunction");
        functionConfig.setInputs(Collections.singletonList("persistent://public/default/sentences"));
        functionConfig.setParallelism(1);
        functionConfig.setCleanupSubscription(true);
        functionConfig.setOutput("persistent://public/default/count");
        Resources resources = new Resources();
        resources.setCpu(1.0);
        resources.setRam(102400L);
        functionConfig.setResources(resources);
        CustomRuntimeOptions customRuntimeOptions = new CustomRuntimeOptions();
        customRuntimeOptions.setClusterName(TEST_CLUSTER_NAME);
        customRuntimeOptions.setInputTypeClassName("java.lang.String");
        customRuntimeOptions.setOutputTypeClassName("java.lang.String");
        String customRuntimeOptionsJSON = new Gson().toJson(customRuntimeOptions, CustomRuntimeOptions.class);
        functionConfig.setCustomRuntimeOptions(customRuntimeOptionsJSON);
        functionConfig.setJar(String.format("%s.jar", functionName));
        functionConfig.setAutoAck(true);
        return functionConfig;
    }

    public static FunctionConfig CreateJavaFunctionWithPackageURLConfig(String tenant, String namespace, String functionName) {
        FunctionConfig functionConfig = new FunctionConfig();
        functionConfig.setName(functionName);
        functionConfig.setTenant(tenant);
        functionConfig.setNamespace(namespace);
        functionConfig.setClassName("org.example.functions.WordCountFunction");
        functionConfig.setInputs(Collections.singletonList("persistent://public/default/sentences"));
        functionConfig.setParallelism(1);
        functionConfig.setCleanupSubscription(true);
        functionConfig.setOutput("persistent://public/default/count");
        Resources resources = new Resources();
        resources.setCpu(1.0);
        resources.setRam(102400L);
        functionConfig.setResources(resources);
        CustomRuntimeOptions customRuntimeOptions = new CustomRuntimeOptions();
        customRuntimeOptions.setClusterName(TEST_CLUSTER_NAME);
        customRuntimeOptions.setInputTypeClassName("java.lang.String");
        customRuntimeOptions.setOutputTypeClassName("java.lang.String");
        String customRuntimeOptionsJSON = new Gson().toJson(customRuntimeOptions, CustomRuntimeOptions.class);
        functionConfig.setCustomRuntimeOptions(customRuntimeOptionsJSON);
        functionConfig.setJar(String.format("function://public/default/%s@1.0", functionName));
        functionConfig.setAutoAck(true);
        return functionConfig;
    }

    public static SinkConfig CreateSinkConfig(String tenant, String namespace, String functionName) {
        SinkConfig sinkConfig = new SinkConfig();
        sinkConfig.setName(functionName);
        sinkConfig.setTenant(tenant);
        sinkConfig.setNamespace(namespace);
        sinkConfig.setClassName("org.apache.pulsar.io.elasticsearch.ElasticSearchSink");
        sinkConfig.setInputs(Collections.singletonList("persistent://public/default/input"));
        sinkConfig.setParallelism(1);
        sinkConfig.setCleanupSubscription(true);
        Resources resources = new Resources();
        resources.setCpu(1.0);
        resources.setRam(102400L);
        sinkConfig.setResources(resources);
        CustomRuntimeOptions customRuntimeOptions = new CustomRuntimeOptions();
        customRuntimeOptions.setClusterName(TEST_CLUSTER_NAME);
        customRuntimeOptions.setInputTypeClassName("[B");
        String customRuntimeOptionsJSON = new Gson().toJson(customRuntimeOptions, CustomRuntimeOptions.class);
        sinkConfig.setCustomRuntimeOptions(customRuntimeOptionsJSON);
        sinkConfig.setArchive("connectors/pulsar-io-elastic-search-2.7.0-rc-pm-3.nar");
        sinkConfig.setAutoAck(true);
        Map<String, Object> configs = new HashMap<>();
        configs.put("elasticSearchUrl", "https://testing-es.app");
        sinkConfig.setConfigs(configs);
        return sinkConfig;
    }

    public static SinkConfig CreateSinkConfigBuiltin(String tenant, String namespace, String functionName) {
        SinkConfig sinkConfig = new SinkConfig();
        sinkConfig.setName(functionName);
        sinkConfig.setTenant(tenant);
        sinkConfig.setNamespace(namespace);
        sinkConfig.setClassName("org.apache.pulsar.io.elasticsearch.ElasticSearchSink");
        sinkConfig.setInputs(Collections.singletonList("persistent://public/default/input"));
        sinkConfig.setParallelism(1);
        sinkConfig.setCleanupSubscription(true);
        Resources resources = new Resources();
        resources.setCpu(1.0);
        resources.setRam(102400L);
        sinkConfig.setResources(resources);
        CustomRuntimeOptions customRuntimeOptions = new CustomRuntimeOptions();
        customRuntimeOptions.setClusterName(TEST_CLUSTER_NAME);
        customRuntimeOptions.setInputTypeClassName("[B");
        String customRuntimeOptionsJSON = new Gson().toJson(customRuntimeOptions, CustomRuntimeOptions.class);
        sinkConfig.setCustomRuntimeOptions(customRuntimeOptionsJSON);
        sinkConfig.setArchive("builtin://elastic-search");
        sinkConfig.setAutoAck(true);
        Map<String, Object> configs = new HashMap<>();
        configs.put("elasticSearchUrl", "https://testing-es.app");
        sinkConfig.setConfigs(configs);
        return sinkConfig;
    }

    public static SourceConfig CreateSourceConfig(String tenant, String namespace, String functionName) {
        SourceConfig sourceConfig = new SourceConfig();
        sourceConfig.setName(functionName);
        sourceConfig.setTenant(tenant);
        sourceConfig.setNamespace(namespace);
        sourceConfig.setClassName("org.apache.pulsar.io.debezium.mongodb.DebeziumMongoDbSource");
        sourceConfig.setTopicName("persistent://public/default/destination");
        sourceConfig.setParallelism(1);
        Resources resources = new Resources();
        resources.setCpu(1.0);
        resources.setRam(102400L);
        sourceConfig.setResources(resources);
        CustomRuntimeOptions customRuntimeOptions = new CustomRuntimeOptions();
        customRuntimeOptions.setClusterName(TEST_CLUSTER_NAME);
        customRuntimeOptions.setOutputTypeClassName("org.apache.pulsar.common.schema.KeyValue");
        String customRuntimeOptionsJSON = new Gson().toJson(customRuntimeOptions, CustomRuntimeOptions.class);
        sourceConfig.setCustomRuntimeOptions(customRuntimeOptionsJSON);
        sourceConfig.setArchive("connectors/pulsar-io-debezium-mongodb-2.7.0.nar");
        Map<String, Object> configs = new HashMap<>();
        String configsName = "test-sourceConfig";
        configs.put("name", configsName);
        sourceConfig.setConfigs(configs);
        return sourceConfig;
    }

    public static SourceConfig CreateSourceConfigBuiltin(String tenant, String namespace, String functionName) {
        SourceConfig sourceConfig = new SourceConfig();
        sourceConfig.setName(functionName);
        sourceConfig.setTenant(tenant);
        sourceConfig.setNamespace(namespace);
        sourceConfig.setClassName("org.apache.pulsar.io.debezium.mongodb.DebeziumMongoDbSource");
        sourceConfig.setTopicName("persistent://public/default/destination");
        sourceConfig.setParallelism(1);
        Resources resources = new Resources();
        resources.setCpu(1.0);
        resources.setRam(102400L);
        sourceConfig.setResources(resources);
        CustomRuntimeOptions customRuntimeOptions = new CustomRuntimeOptions();
        customRuntimeOptions.setClusterName(TEST_CLUSTER_NAME);
        customRuntimeOptions.setOutputTypeClassName("org.apache.pulsar.common.schema.KeyValue");
        String customRuntimeOptionsJSON = new Gson().toJson(customRuntimeOptions, CustomRuntimeOptions.class);
        sourceConfig.setCustomRuntimeOptions(customRuntimeOptionsJSON);
        sourceConfig.setArchive("builtin://debezium-mongodb");
        Map<String, Object> configs = new HashMap<>();
        String configsName = "test-sourceConfig";
        configs.put("name", configsName);
        sourceConfig.setConfigs(configs);
        return sourceConfig;
    }

}
