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

import com.google.gson.Gson;
import io.functionmesh.compute.sinks.models.V1alpha1Sink;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpec;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecInput;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecJava;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecPulsar;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecResources;
import io.functionmesh.proxy.models.CustomRuntimeOptions;
import io.kubernetes.client.openapi.models.V1ObjectMeta;
import org.apache.commons.io.FileUtils;
import org.apache.commons.lang.StringUtils;
import org.apache.logging.log4j.util.Strings;
import org.apache.pulsar.common.functions.Resources;
import org.apache.pulsar.common.io.SinkConfig;
import org.apache.pulsar.common.nar.NarClassLoader;
import org.apache.pulsar.common.util.RestException;
import org.apache.pulsar.functions.utils.FunctionCommon;
import org.apache.pulsar.functions.utils.io.ConnectorUtils;

import javax.ws.rs.core.Response;
import java.io.File;
import java.io.IOException;
import java.io.InputStream;
import java.net.URISyntaxException;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.Map;
import java.util.stream.Collectors;

public class SinksUtil {
    public static final String cpuKey = "cpu";
    public static final String memoryKey = "memory";

    public static Map<String, String> transformedMapValueToString(Map<String, Object> map) {
        if (map == null) {
            return null;
        }
        return map.entrySet().stream()
                .collect(Collectors.toMap(Map.Entry::getKey, e -> String.valueOf(e.getValue())));
    }

    public static Map<String, Object> transformedMapValueToObject(Map<String, String> map) {
        if (map == null) {
            return null;
        }
        return map.entrySet().stream()
                .collect(Collectors.toMap(Map.Entry::getKey, Map.Entry::getValue));
    }

    public static V1alpha1Sink createV1alpha1SkinFromSinkConfig(
            String kind,
            String group,
            String version,
            String sinkName,
            String sinkPkgUrl,
            InputStream uploadedInputStream,
            SinkConfig sinkConfig)
            throws IOException, URISyntaxException, ClassNotFoundException {
        String typeClassName = "";
        V1alpha1Sink v1alpha1Sink = new V1alpha1Sink();
        v1alpha1Sink.setKind(kind);
        v1alpha1Sink.setApiVersion(String.format("%s/%s", group, version));

        V1ObjectMeta v1ObjectMeta = new V1ObjectMeta();
        v1ObjectMeta.setName(sinkConfig.getName());
        v1ObjectMeta.setNamespace(sinkConfig.getNamespace());
        v1alpha1Sink.setMetadata(v1ObjectMeta);

        V1alpha1SinkSpec v1alpha1SinkSpec = new V1alpha1SinkSpec();

        File file;
        if (Strings.isNotEmpty(sinkPkgUrl)) {
            file = FunctionCommon.extractFileFromPkgURL(sinkPkgUrl);
        } else {
            file = FunctionCommon.createPkgTempFile();
            FileUtils.copyInputStreamToFile(uploadedInputStream, file);
        }
        NarClassLoader narClassLoader = FunctionCommon.extractNarClassLoader(file, null);
        String sourceClassName = ConnectorUtils.getIOSourceClass(narClassLoader);
        Class<?> sourceClass = narClassLoader.loadClass(sourceClassName);
        Class<?> sourceType = FunctionCommon.getSourceType(sourceClass);

        v1alpha1SinkSpec.setClassName(sourceClassName);
        typeClassName = sourceType.getName();

        Integer parallelism = sinkConfig.getParallelism() == null ? 1 : sinkConfig.getParallelism();
        v1alpha1SinkSpec.setReplicas(parallelism);
        v1alpha1SinkSpec.setMaxReplicas(parallelism);

        V1alpha1SinkSpecInput v1alpha1SinkSpecInput = new V1alpha1SinkSpecInput();
        v1alpha1SinkSpecInput.setTopics(new ArrayList<>(sinkConfig.getInputs()));
        v1alpha1SinkSpecInput.setTypeClassName(typeClassName);
        v1alpha1SinkSpec.setInput(v1alpha1SinkSpecInput);

        v1alpha1SinkSpec.setSinkConfig(transformedMapValueToString(sinkConfig.getConfigs()));

        if (sinkConfig.getResources() == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "cpu and ram is not provided");
        }
        Double cpu = sinkConfig.getResources().getCpu();
        if (cpu == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "cpu is not provided");
        }
        Long memory = sinkConfig.getResources().getRam();
        if (memory == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "ram is not provided");
        }
        String cpuValue = cpu.toString();
        String memoryValue = memory.toString();
        Map<String, String> limits = new HashMap<>();
        limits.put(cpuKey, cpuValue);
        limits.put(memoryKey, memoryValue);

        V1alpha1SinkSpecResources v1alpha1SinkSpecResources = new V1alpha1SinkSpecResources();
        v1alpha1SinkSpecResources.setLimits(limits);
        v1alpha1SinkSpecResources.setRequests(limits);
        v1alpha1SinkSpec.setResources(v1alpha1SinkSpecResources);

        V1alpha1SinkSpecPulsar pulsar = new V1alpha1SinkSpecPulsar();
        pulsar.setPulsarConfig(sinkName);
        v1alpha1SinkSpec.setPulsar(pulsar);

        String location =
                String.format(
                        "%s/%s/%s",
                        sinkConfig.getTenant(), sinkConfig.getNamespace(), sinkConfig.getName());
        if (StringUtils.isNotEmpty(sinkPkgUrl)) {
            location = sinkPkgUrl;
        }
        Path path = Paths.get(sinkConfig.getArchive());
        String jar = path.getFileName().toString();
        V1alpha1SinkSpecJava v1alpha1SinkSpecJava = new V1alpha1SinkSpecJava();
        v1alpha1SinkSpecJava.setJar(jar);
        v1alpha1SinkSpecJava.setJarLocation(location);
        v1alpha1SinkSpec.setJava(v1alpha1SinkSpecJava);

        String customRuntimeOptionsJSON = sinkConfig.getCustomRuntimeOptions();
        if (Strings.isEmpty(customRuntimeOptionsJSON)) {
            throw new RestException(
                    Response.Status.BAD_REQUEST, "customRuntimeOptions is not provided");
        }

        CustomRuntimeOptions customRuntimeOptions =
                new Gson().fromJson(customRuntimeOptionsJSON, CustomRuntimeOptions.class);
        String clusterName = customRuntimeOptions.getClusterName();
        if (clusterName == null) {
            throw new RestException(
                    Response.Status.BAD_REQUEST,
                    "clusterName is not provided in customRuntimeOptions");
        }
        v1alpha1SinkSpec.setClusterName(clusterName);
        v1alpha1SinkSpec.setAutoAck(sinkConfig.getAutoAck());

        if (Strings.isNotEmpty(customRuntimeOptions.getTypeClassName()) && Strings.isEmpty(typeClassName)) {
            typeClassName = customRuntimeOptions.getTypeClassName();
        }

        v1alpha1SinkSpecInput.setTypeClassName(typeClassName);
        v1alpha1SinkSpec.setInput(v1alpha1SinkSpecInput);

        v1alpha1Sink.setSpec(v1alpha1SinkSpec);

        return v1alpha1Sink;
    }

    public static SinkConfig createSinkConfigFromV1alpha1Source(
            String tenant, String namespace, String sinkName, V1alpha1Sink v1alpha1Sink) {
        SinkConfig sinkConfig = new SinkConfig();
        sinkConfig.setTenant(tenant);
        sinkConfig.setNamespace(namespace);
        sinkConfig.setName(sinkName);

        V1alpha1SinkSpec v1alpha1SinkSpec = v1alpha1Sink.getSpec();

        sinkConfig.setClassName(v1alpha1SinkSpec.getClassName());

        sinkConfig.setParallelism(v1alpha1SinkSpec.getReplicas());

        V1alpha1SinkSpecInput v1alpha1SinkSpecInput = v1alpha1SinkSpec.getInput();
        sinkConfig.setInputs(v1alpha1SinkSpecInput.getTopics());

        sinkConfig.setConfigs(transformedMapValueToObject(v1alpha1SinkSpec.getSinkConfig()));

        V1alpha1SinkSpecResources v1alpha1SinkSpecResources = v1alpha1SinkSpec.getResources();
        if (v1alpha1SinkSpecResources != null) {
            Resources resources = new Resources();
            Map<String, String> limits = v1alpha1SinkSpecResources.getLimits();
            resources.setCpu(Double.parseDouble(limits.get(cpuKey)));
            resources.setRam(Long.parseLong(limits.get(memoryKey)));
            sinkConfig.setResources(resources);
        }

        V1alpha1SinkSpecJava v1alpha1SinkSpecJava = v1alpha1SinkSpec.getJava();
        if (v1alpha1SinkSpecJava != null) {
            sinkConfig.setArchive(v1alpha1SinkSpecJava.getJar());
        }

        CustomRuntimeOptions customRuntimeOptions = new CustomRuntimeOptions();
        customRuntimeOptions.setClusterName(v1alpha1SinkSpec.getClusterName());
        sinkConfig.setCustomRuntimeOptions(new Gson().toJson(customRuntimeOptions));

        sinkConfig.setAutoAck(v1alpha1SinkSpec.getAutoAck());

        return sinkConfig;
    }
}
