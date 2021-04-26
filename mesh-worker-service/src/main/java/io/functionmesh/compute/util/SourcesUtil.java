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
import io.functionmesh.compute.sources.models.V1alpha1Source;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpec;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecJava;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecOutput;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecOutputProducerConf;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecOutputProducerConfCryptoConfig;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecPulsar;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecResources;
import io.functionmesh.compute.models.CustomRuntimeOptions;
import io.kubernetes.client.openapi.models.V1ObjectMeta;
import org.apache.commons.io.FileUtils;
import org.apache.commons.lang.StringUtils;
import org.apache.logging.log4j.util.Strings;
import org.apache.pulsar.client.api.ConsumerCryptoFailureAction;
import org.apache.pulsar.client.api.ProducerCryptoFailureAction;
import org.apache.pulsar.common.functions.CryptoConfig;
import org.apache.pulsar.common.functions.ProducerConfig;
import org.apache.pulsar.common.functions.Resources;
import org.apache.pulsar.common.io.SourceConfig;
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
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

public class SourcesUtil {
    public final static String cpuKey = "cpu";
    public final static String memoryKey = "memory";

    public static Map<String, String> transformedMapValueToString(Map<String, Object> map) {
        if (map == null) {
            return null;
        }
        return map.entrySet().stream().collect(Collectors.toMap(Map.Entry::getKey, e -> String.valueOf(e.getValue())));
    }

    public static Map<String, Object> transformedMapValueToObject(Map<String, String> map) {
        if (map == null) {
            return null;
        }
        return map.entrySet().stream().collect(Collectors.toMap(Map.Entry::getKey, Map.Entry::getValue));
    }

    public static V1alpha1Source createV1alpha1SourceFromSourceConfig(String kind, String group, String version,
                                                                      String sourceName, String sourcePkgUrl,
                                                                      InputStream uploadedInputStream,
                                                                      SourceConfig sourceConfig) throws IOException,
            URISyntaxException, ClassNotFoundException {
        String typeClassName = "";
        V1alpha1Source v1alpha1Source = new V1alpha1Source();
        v1alpha1Source.setKind(kind);
        v1alpha1Source.setApiVersion(String.format("%s/%s", group, version));

        V1ObjectMeta v1ObjectMeta = new V1ObjectMeta();
        v1ObjectMeta.setName(sourceConfig.getName());
        v1ObjectMeta.setNamespace(sourceConfig.getNamespace());
        v1alpha1Source.setMetadata(v1ObjectMeta);

        V1alpha1SourceSpec v1alpha1SourceSpec = new V1alpha1SourceSpec();
        v1alpha1SourceSpec.setClassName(sourceConfig.getClassName());

        File file;
        if (Strings.isNotEmpty(sourcePkgUrl)) {
            file = FunctionCommon.extractFileFromPkgURL(sourcePkgUrl);
        } else {
            file = FunctionCommon.createPkgTempFile();
            FileUtils.copyInputStreamToFile(uploadedInputStream, file);
        }
        NarClassLoader narClassLoader = FunctionCommon.extractNarClassLoader(file, null);
        String sourceClassName = ConnectorUtils.getIOSourceClass(narClassLoader);
        Class<?> sourceClass = narClassLoader.loadClass(sourceClassName);
        Class<?> sourceType = FunctionCommon.getSourceType(sourceClass);

        typeClassName = sourceType.getName();

        Integer parallelism = sourceConfig.getParallelism() == null ? 1 : sourceConfig.getParallelism();
        v1alpha1SourceSpec.setReplicas(parallelism);
        v1alpha1SourceSpec.setMaxReplicas(parallelism);

        V1alpha1SourceSpecOutput v1alpha1SourceSpecOutput = new V1alpha1SourceSpecOutput();
        v1alpha1SourceSpecOutput.setTopic(sourceConfig.getTopicName());
        v1alpha1SourceSpecOutput.setTypeClassName(typeClassName);
        v1alpha1SourceSpec.setOutput(v1alpha1SourceSpecOutput);

        ProducerConfig producerConfig = sourceConfig.getProducerConfig();
        if (producerConfig != null) {
            V1alpha1SourceSpecOutputProducerConf v1alpha1SourceSpecOutputProducerConf =
                    new V1alpha1SourceSpecOutputProducerConf();
            CryptoConfig cryptoConfig = producerConfig.getCryptoConfig();
            if (cryptoConfig != null) {
                V1alpha1SourceSpecOutputProducerConfCryptoConfig v1alpha1SourceSpecOutputProducerConfCryptoConfig =
                        new V1alpha1SourceSpecOutputProducerConfCryptoConfig();
                v1alpha1SourceSpecOutputProducerConfCryptoConfig.setConsumerCryptoFailureAction(cryptoConfig.getConsumerCryptoFailureAction().name());
                v1alpha1SourceSpecOutputProducerConfCryptoConfig.setProducerCryptoFailureAction(cryptoConfig.getProducerCryptoFailureAction().name());
                v1alpha1SourceSpecOutputProducerConfCryptoConfig.setCryptoKeyReaderClassName(cryptoConfig.getCryptoKeyReaderClassName());
                v1alpha1SourceSpecOutputProducerConfCryptoConfig.setEncryptionKeys(Arrays.asList(cryptoConfig.getEncryptionKeys().clone()));
                Map<String, String> cryptoKeyReaderConfig = transformedMapValueToString(cryptoConfig.getCryptoKeyReaderConfig());
                v1alpha1SourceSpecOutputProducerConfCryptoConfig.setCryptoKeyReaderConfig(cryptoKeyReaderConfig);

                v1alpha1SourceSpecOutputProducerConf.setCryptoConfig(v1alpha1SourceSpecOutputProducerConfCryptoConfig);
            }
            v1alpha1SourceSpecOutputProducerConf.setMaxPendingMessages(producerConfig.getMaxPendingMessages());
            v1alpha1SourceSpecOutputProducerConf.setMaxPendingMessagesAcrossPartitions(producerConfig.getMaxPendingMessagesAcrossPartitions());
            v1alpha1SourceSpecOutputProducerConf.setUseThreadLocalProducers(producerConfig.getUseThreadLocalProducers());
        }

        if (sourceConfig.getResources() == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "cpu and ram is not provided");
        }
        Double cpu = sourceConfig.getResources().getCpu();
        if (sourceConfig.getResources().getCpu() == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "cpu is not provided");
        }
        Long memory = sourceConfig.getResources().getRam();
        if (sourceConfig.getResources().getRam() == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "ram is not provided");
        }
        String cpuValue = cpu.toString();
        String memoryValue = memory.toString();
        Map<String, String> limits = new HashMap<>();
        limits.put(cpuKey, cpuValue);
        limits.put(memoryKey, memoryValue);

        V1alpha1SourceSpecResources v1alpha1SourceSpecResources = new V1alpha1SourceSpecResources();
        v1alpha1SourceSpecResources.setLimits(limits);
        v1alpha1SourceSpecResources.setRequests(limits);
        v1alpha1SourceSpec.setResources(v1alpha1SourceSpecResources);

        v1alpha1SourceSpec.setSourceConfig(transformedMapValueToString(sourceConfig.getConfigs()));

        V1alpha1SourceSpecPulsar pulsar = new V1alpha1SourceSpecPulsar();
        pulsar.setPulsarConfig(sourceName);
        v1alpha1SourceSpec.setPulsar(pulsar);

        String location = String.format("%s/%s/%s", sourceConfig.getTenant(), sourceConfig.getNamespace(),
                sourceConfig.getName());
        if (StringUtils.isNotEmpty(sourcePkgUrl)) {
            location = sourcePkgUrl;
        }
        Path path = Paths.get(sourceConfig.getArchive());
        String jar = path.getFileName().toString();
        V1alpha1SourceSpecJava v1alpha1SourceSpecJava = new V1alpha1SourceSpecJava();
        v1alpha1SourceSpecJava.setJar(jar);
        v1alpha1SourceSpecJava.setJarLocation(location);
        v1alpha1SourceSpec.setJava(v1alpha1SourceSpecJava);

        String customRuntimeOptionsJSON = sourceConfig.getCustomRuntimeOptions();
        if (Strings.isEmpty(customRuntimeOptionsJSON)) {
            throw new RestException(Response.Status.BAD_REQUEST, "customRuntimeOptions is not provided");
        }

        CustomRuntimeOptions customRuntimeOptions = new Gson().fromJson(customRuntimeOptionsJSON,
                CustomRuntimeOptions.class);
        String clusterName = customRuntimeOptions.getClusterName();
        if (clusterName == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "clusterName is not provided in customRuntimeOptions");
        }
        v1alpha1SourceSpec.setClusterName(clusterName);

        if (Strings.isNotEmpty(customRuntimeOptions.getTypeClassName()) && Strings.isEmpty(typeClassName)) {
            typeClassName = customRuntimeOptions.getTypeClassName();
        }

        v1alpha1SourceSpecOutput.setTypeClassName(typeClassName);
        v1alpha1SourceSpec.setOutput(v1alpha1SourceSpecOutput);

        v1alpha1Source.setSpec(v1alpha1SourceSpec);

        return v1alpha1Source;
    }

    public static SourceConfig createSourceConfigFromV1alpha1Source(String tenant, String namespace, String sourceName,
                                                                    V1alpha1Source v1alpha1Source) {
        SourceConfig sourceConfig = new SourceConfig();
        sourceConfig.setTenant(tenant);
        sourceConfig.setNamespace(namespace);
        sourceConfig.setName(sourceName);

        V1alpha1SourceSpec sourceSpec = v1alpha1Source.getSpec();
        sourceConfig.setClassName(sourceSpec.getClassName());

        sourceConfig.setParallelism(sourceSpec.getReplicas());

        V1alpha1SourceSpecOutput sourceSpecOutput = sourceSpec.getOutput();
        if (sourceSpecOutput != null) {
            sourceConfig.setTopicName(sourceSpecOutput.getTopic());
            V1alpha1SourceSpecOutputProducerConf sourceSpecOutputProducerConf = sourceSpecOutput.getProducerConf();
            if (sourceSpecOutputProducerConf != null) {
                ProducerConfig producerConfig = new ProducerConfig();
                producerConfig.setMaxPendingMessages(sourceSpecOutputProducerConf.getMaxPendingMessages());
                producerConfig.setMaxPendingMessagesAcrossPartitions(sourceSpecOutputProducerConf.getMaxPendingMessagesAcrossPartitions());
                producerConfig.setUseThreadLocalProducers(sourceSpecOutputProducerConf.getUseThreadLocalProducers());

                V1alpha1SourceSpecOutputProducerConfCryptoConfig specOutputProducerConfCryptoConfig =
                        sourceSpecOutputProducerConf.getCryptoConfig();
                if (specOutputProducerConfCryptoConfig != null) {
                    CryptoConfig cryptoConfig = new CryptoConfig();
                    cryptoConfig.setConsumerCryptoFailureAction(ConsumerCryptoFailureAction.valueOf(specOutputProducerConfCryptoConfig.getConsumerCryptoFailureAction()));
                    cryptoConfig.setProducerCryptoFailureAction(ProducerCryptoFailureAction.valueOf(specOutputProducerConfCryptoConfig.getProducerCryptoFailureAction()));
                    cryptoConfig.setCryptoKeyReaderClassName(specOutputProducerConfCryptoConfig.getCryptoKeyReaderClassName());
                    cryptoConfig.setCryptoKeyReaderConfig(transformedMapValueToObject(specOutputProducerConfCryptoConfig.getCryptoKeyReaderConfig()));
                    List<String> encryptionKeys = specOutputProducerConfCryptoConfig.getEncryptionKeys();
                    if (encryptionKeys != null) {
                        cryptoConfig.setEncryptionKeys(encryptionKeys.toArray(new String[0]));
                    }
                    producerConfig.setCryptoConfig(cryptoConfig);
                }
                sourceConfig.setProducerConfig(producerConfig);
            }
        }

        V1alpha1SourceSpecResources sourceSpecResources = sourceSpec.getResources();
        if (sourceSpecResources != null) {
            Resources resources = new Resources();
            Map<String, String> limits = sourceSpecResources.getLimits();
            resources.setCpu(Double.parseDouble(limits.get(cpuKey)));
            resources.setRam(Long.parseLong(limits.get(memoryKey)));
            sourceConfig.setResources(resources);
        }

        V1alpha1SourceSpecJava sourceSpecJava = sourceSpec.getJava();
        if (sourceSpecJava != null) {
            sourceConfig.setArchive(sourceSpecJava.getJar());
        }

        sourceConfig.setConfigs(transformedMapValueToObject(sourceSpec.getSourceConfig()));

        CustomRuntimeOptions customRuntimeOptions = new CustomRuntimeOptions();
        customRuntimeOptions.setClusterName(sourceSpec.getClusterName());
        sourceConfig.setCustomRuntimeOptions(new Gson().toJson(customRuntimeOptions));

        return sourceConfig;
    }
}
