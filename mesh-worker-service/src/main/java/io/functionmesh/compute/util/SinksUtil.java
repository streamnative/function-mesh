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
import io.functionmesh.compute.models.FunctionMeshConnectorDefinition;
import io.functionmesh.compute.worker.FunctionMeshConnectorsManager;
import io.functionmesh.compute.sinks.models.V1alpha1Sink;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpec;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecInput;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecJava;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecPulsar;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecResources;
import io.functionmesh.proxy.models.CustomRuntimeOptions;
import io.kubernetes.client.custom.Quantity;
import io.kubernetes.client.openapi.models.V1ObjectMeta;
import lombok.Data;
import lombok.extern.slf4j.Slf4j;
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

import static org.apache.pulsar.common.functions.Utils.BUILTIN;
import static org.apache.pulsar.functions.utils.FunctionCommon.roundDecimal;

@Slf4j
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
            SinkConfig sinkConfig, FunctionMeshConnectorsManager connectorsManager)
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

        File file = null;
        if (Strings.isNotEmpty(sinkPkgUrl)) {
            file = FunctionCommon.extractFileFromPkgURL(sinkPkgUrl);
        } else if (uploadedInputStream != null) {
            file = FunctionCommon.createPkgTempFile();
            FileUtils.copyInputStreamToFile(uploadedInputStream, file);
        }
        if (file != null) {
            NarClassLoader narClassLoader = FunctionCommon.extractNarClassLoader(file, null);
            String sinkClassName = ConnectorUtils.getIOSinkClass(narClassLoader);
            Class<?> sinkClass = narClassLoader.loadClass(sinkClassName);
            Class<?> sinkType = FunctionCommon.getSourceType(sinkClass);

            v1alpha1SinkSpec.setClassName(sinkClassName);
            v1alpha1SinkSpec.setTypeClassName(sinkType.getName());
        }

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

        double cpu = sinkConfig.getResources() == null && sinkConfig.getResources().getCpu() != 0 ? sinkConfig.getResources().getCpu() : 1;
        long ramRequest = sinkConfig.getResources() == null && sinkConfig.getResources().getRam() != 0 ? sinkConfig.getResources().getRam() : 1073741824;

        Map<String, String> limits = new HashMap<>();
        Map<String, String> requests = new HashMap<>();

        long padding = Math.round(ramRequest * (10.0 / 100.0)); // percentMemoryPadding is 0.1
        long ramWithPadding = ramRequest + padding;

        limits.put(cpuKey, Quantity.fromString(Double.toString(roundDecimal(cpu, 3))).toSuffixedString());
        limits.put(memoryKey, Quantity.fromString(Long.toString(ramWithPadding)).toSuffixedString());

        requests.put(cpuKey, Quantity.fromString(Double.toString(roundDecimal(cpu, 3))).toSuffixedString());
        requests.put(memoryKey, Quantity.fromString(Long.toString(ramRequest)).toSuffixedString());

        V1alpha1SinkSpecResources v1alpha1SinkSpecResources = new V1alpha1SinkSpecResources();
        v1alpha1SinkSpecResources.setLimits(limits);
        v1alpha1SinkSpecResources.setRequests(requests);
        v1alpha1SinkSpec.setResources(v1alpha1SinkSpecResources);

        String location =
                String.format(
                        "%s/%s/%s",
                        sinkConfig.getTenant(), sinkConfig.getNamespace(), sinkConfig.getName());
        if (StringUtils.isNotEmpty(sinkPkgUrl)) {
            location = sinkPkgUrl;
        }
        String archive = sinkConfig.getArchive();
        V1alpha1SinkSpecJava v1alpha1SinkSpecJava = new V1alpha1SinkSpecJava();
        if (connectorsManager != null && archive.startsWith(BUILTIN)) {
            String connectorType = archive.replaceFirst("^builtin://", "");
            FunctionMeshConnectorDefinition definition = connectorsManager.getConnectorDefinition(connectorType);
            if (definition != null) {
                v1alpha1SinkSpec.setImage(definition.toFullImageURL());
                if (definition.getSinkClass() != null && v1alpha1SinkSpec.getClassName() == null) {
                    v1alpha1SinkSpec.setClassName(definition.getSinkClass());
                }
                v1alpha1SinkSpecJava.setJar(definition.getJar());
                v1alpha1SinkSpecJava.setJarLocation("");
                v1alpha1SinkSpec.setJava(v1alpha1SinkSpecJava);
            } else {
                log.warn("cannot find built-in connector {}", connectorType);
                throw new RestException(Response.Status.BAD_REQUEST, String.format("connectorType %s is not supported yet", connectorType));
            }
        } else {
            Path path = Paths.get(sinkConfig.getArchive());
            String jar = path.getFileName().toString();

            v1alpha1SinkSpecJava.setJar(jar);
            v1alpha1SinkSpecJava.setJarLocation(location);
            v1alpha1SinkSpec.setJava(v1alpha1SinkSpecJava);
        }

        String customRuntimeOptionsJSON = sinkConfig.getCustomRuntimeOptions();
        if (Strings.isEmpty(customRuntimeOptionsJSON)) {
            throw new RestException(
                    Response.Status.BAD_REQUEST, "customRuntimeOptions is not provided");
        }

        String clusterName = CommonUtil.getCurrentClusterName();
        try {
            CustomRuntimeOptions customRuntimeOptions =
                    new Gson().fromJson(customRuntimeOptionsJSON, CustomRuntimeOptions.class);

            if (Strings.isNotEmpty(customRuntimeOptions.getClusterName())) {
                clusterName = customRuntimeOptions.getClusterName();
            }

            if (Strings.isNotEmpty(customRuntimeOptions.getTypeClassName())) {
                v1alpha1SinkSpec.setTypeClassName(customRuntimeOptions.getTypeClassName());
            }
        } catch (Exception ignored) {
        }

        if (Strings.isEmpty(clusterName)) {
            throw new RestException(Response.Status.BAD_REQUEST, "clusterName is not provided in customRuntimeOptions");
        }

        v1alpha1SinkSpec.setClusterName(clusterName);

        boolean autoAck = sinkConfig.getAutoAck() != null ? sinkConfig.getAutoAck() : true;
        v1alpha1SinkSpec.setAutoAck(autoAck);

        V1alpha1SinkSpecPulsar pulsar = new V1alpha1SinkSpecPulsar();
        pulsar.setPulsarConfig(CommonUtil.getPulsarClusterConfigMapName(clusterName)); // try to reuse the configMap
        v1alpha1SinkSpec.setPulsar(pulsar);

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

    @Data
    static class CustomRuntimeOptions {
        private String clusterName;
        private String typeClassName;
    }
}
