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
import io.functionmesh.compute.models.CustomRuntimeOptions;
import io.functionmesh.compute.models.FunctionMeshConnectorDefinition;
import io.functionmesh.compute.sources.models.V1alpha1Source;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpec;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecJava;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecOutput;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecOutputProducerConf;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecOutputProducerConfCryptoConfig;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecPulsar;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecPodResources;
import io.functionmesh.compute.worker.MeshConnectorsManager;
import io.kubernetes.client.custom.Quantity;
import lombok.extern.slf4j.Slf4j;
import io.functionmesh.compute.models.CustomRuntimeOptions;
import org.apache.commons.lang.StringUtils;
import org.apache.logging.log4j.util.Strings;
import org.apache.pulsar.common.functions.FunctionDefinition;
import org.apache.pulsar.common.functions.ProducerConfig;
import org.apache.pulsar.common.functions.Resources;
import org.apache.pulsar.common.io.SourceConfig;
import org.apache.pulsar.common.util.RestException;
import org.apache.pulsar.functions.proto.Function;
import org.apache.pulsar.functions.utils.SourceConfigUtils;

import javax.ws.rs.core.Response;
import java.io.IOException;
import java.io.InputStream;
import java.net.URISyntaxException;
import java.util.HashMap;
import java.util.Map;

import static org.apache.pulsar.common.functions.Utils.BUILTIN;

@Slf4j
public class SourcesUtil {
    public final static String cpuKey = "cpu";
    public final static String memoryKey = "memory";

    public static V1alpha1Source createV1alpha1SourceFromSourceConfig(String kind, String group, String version,
                                                                      String sourceName, String sourcePkgUrl,
                                                                      InputStream uploadedInputStream,
                                                                      SourceConfig sourceConfig,
                                                                      MeshConnectorsManager connectorsManager,
                                                                      Map<String, Object> customConfigs, String cluster) {
        CustomRuntimeOptions customRuntimeOptions = CommonUtil.getCustomRuntimeOptions(sourceConfig.getCustomRuntimeOptions());
        String clusterName = CommonUtil.getClusterName(cluster, customRuntimeOptions);

        String location = String.format("%s/%s/%s", sourceConfig.getTenant(), sourceConfig.getNamespace(),
                sourceConfig.getName());
        if (StringUtils.isNotEmpty(sourcePkgUrl)) {
            location = sourcePkgUrl;
        }
        String archive = sourceConfig.getArchive();
        SourceConfigUtils.ExtractedSourceDetails extractedSourceDetails =
                new SourceConfigUtils.ExtractedSourceDetails("", customRuntimeOptions.getInputTypeClassName());

        Function.FunctionDetails functionDetails = null;
        try {
            functionDetails = SourceConfigUtils.convert(sourceConfig, extractedSourceDetails);
        } catch (IllegalArgumentException ex) {
            log.error("cannot convert SourceConfig to FunctionDetails", ex);
            throw new RestException(Response.Status.BAD_REQUEST, "functionConfig cannot be parsed into functionDetails");
        }

        V1alpha1Source v1alpha1Source = new V1alpha1Source();
        v1alpha1Source.setKind(kind);
        v1alpha1Source.setApiVersion(String.format("%s/%s", group, version));
        v1alpha1Source.setMetadata(CommonUtil.makeV1ObjectMeta(sourceConfig.getName(),
                sourceConfig.getNamespace(),
                functionDetails.getNamespace(),
                functionDetails.getTenant(),
                clusterName,
                CommonUtil.getOwnerReferenceFromCustomConfigs(customConfigs)));

        V1alpha1SourceSpec v1alpha1SourceSpec = new V1alpha1SourceSpec();
        v1alpha1SourceSpec.setClassName(sourceConfig.getClassName());

        V1alpha1SourceSpecJava v1alpha1SourceSpecJava = new V1alpha1SourceSpecJava();
        if (connectorsManager != null && archive.startsWith(BUILTIN)) {
            String connectorType = archive.replaceFirst("^builtin://", "");
            FunctionMeshConnectorDefinition definition = connectorsManager.getConnectorDefinition(connectorType);
            if (definition != null) {
                v1alpha1SourceSpec.setImage(definition.toFullImageURL());
                if (definition.getSourceClass() != null && v1alpha1SourceSpec.getClassName() == null) {
                    v1alpha1SourceSpec.setClassName(definition.getSourceClass());
                    extractedSourceDetails.setSourceClassName(definition.getSourceClass());
                }
                v1alpha1SourceSpecJava.setJar(definition.getJar());
                v1alpha1SourceSpecJava.setJarLocation("");
                v1alpha1SourceSpec.setJava(v1alpha1SourceSpecJava);
            } else {
                log.warn("cannot find built-in connector {}", connectorType);
                throw new RestException(Response.Status.BAD_REQUEST, String.format("connectorType %s is not supported yet", connectorType));
            }
        } else {
            v1alpha1SourceSpecJava.setJar(sourceConfig.getArchive());
            v1alpha1SourceSpecJava.setJarLocation(location);
            v1alpha1SourceSpec.setJava(v1alpha1SourceSpecJava);
            extractedSourceDetails.setSourceClassName(sourceConfig.getClassName());
        }

        V1alpha1SourceSpecOutput v1alpha1SourceSpecOutput = new V1alpha1SourceSpecOutput();
        if (Strings.isNotEmpty(functionDetails.getSink().getTopic())) {
            v1alpha1SourceSpecOutput.setTopic(functionDetails.getSink().getTopic());
        }
        if (Strings.isNotEmpty(functionDetails.getSink().getSerDeClassName())) {
            v1alpha1SourceSpecOutput.setSinkSerdeClassName(functionDetails.getSink().getSerDeClassName());
        }
        if (Strings.isNotEmpty(functionDetails.getSink().getSchemaType())) {
            v1alpha1SourceSpecOutput.setSinkSchemaType(functionDetails.getSink().getSchemaType());
        }
        // process ProducerConf
        V1alpha1SourceSpecOutputProducerConf v1alpha1SourceSpecOutputProducerConf
                = new V1alpha1SourceSpecOutputProducerConf();
        Function.ProducerSpec producerSpec = functionDetails.getSink().getProducerSpec();
        if (Strings.isNotEmpty(producerSpec.getBatchBuilder())) {
            v1alpha1SourceSpecOutputProducerConf.setBatchBuilder(producerSpec.getBatchBuilder());
        }
        v1alpha1SourceSpecOutputProducerConf.setMaxPendingMessages(producerSpec.getMaxPendingMessages());
        v1alpha1SourceSpecOutputProducerConf.setMaxPendingMessagesAcrossPartitions(
                producerSpec.getMaxPendingMessagesAcrossPartitions());
        v1alpha1SourceSpecOutputProducerConf.useThreadLocalProducers(producerSpec.getUseThreadLocalProducers());
        if (producerSpec.hasCryptoSpec()) {
            v1alpha1SourceSpecOutputProducerConf.setCryptoConfig(
                    convertFromCryptoSpec(producerSpec.getCryptoSpec()));
        }

        v1alpha1SourceSpecOutput.setProducerConf(v1alpha1SourceSpecOutputProducerConf);

        if (Strings.isNotEmpty(customRuntimeOptions.getOutputTypeClassName())) {
            v1alpha1SourceSpecOutput.setTypeClassName(customRuntimeOptions.getOutputTypeClassName());
        } else {
            if (connectorsManager == null) {
                v1alpha1SourceSpecOutput.setTypeClassName("[B");
            } else {
                String connectorType = archive.replaceFirst("^builtin://", "");
                FunctionMeshConnectorDefinition functionMeshConnectorDefinition =
                        connectorsManager.getConnectorDefinition(connectorType);
                if (functionMeshConnectorDefinition == null) {
                    v1alpha1SourceSpecOutput.setTypeClassName("[B");
                } else {
                    if (functionMeshConnectorDefinition.getTypeClassName() == null) {
                        v1alpha1SourceSpecOutput.setTypeClassName("[B");
                    } else {
                        v1alpha1SourceSpecOutput.setTypeClassName(functionMeshConnectorDefinition.getTypeClassName());
                    }
                }
            }
        }

        v1alpha1SourceSpec.setOutput(v1alpha1SourceSpecOutput);

        if (!org.apache.commons.lang3.StringUtils.isEmpty(functionDetails.getLogTopic())) {
            v1alpha1SourceSpec.setLogTopic(functionDetails.getLogTopic());
        }

        v1alpha1SourceSpec.setReplicas(functionDetails.getParallelism());

        double cpu = sourceConfig.getResources() != null && sourceConfig.getResources().getCpu() != 0 ? sourceConfig.getResources().getCpu() : 1;
        long ramRequest = sourceConfig.getResources() != null && sourceConfig.getResources().getRam() != 0 ? sourceConfig.getResources().getRam() : 1073741824;

        Map<String, String> limits = new HashMap<>();
        Map<String, String> requests = new HashMap<>();

        long padding = Math.round(ramRequest * (10.0 / 100.0)); // percentMemoryPadding is 0.1
        long ramWithPadding = ramRequest + padding;

        limits.put(cpuKey, Quantity.fromString(Double.toString(cpu)).toSuffixedString());
        limits.put(memoryKey, Quantity.fromString(Long.toString(ramWithPadding)).toSuffixedString());

        requests.put(cpuKey, Quantity.fromString(Double.toString(cpu)).toSuffixedString());
        requests.put(memoryKey, Quantity.fromString(Long.toString(ramRequest)).toSuffixedString());

        V1alpha1SourceSpecPodResources v1alpha1SourceSpecResources = new V1alpha1SourceSpecPodResources();
        v1alpha1SourceSpecResources.setLimits(limits);
        v1alpha1SourceSpecResources.setRequests(requests);
        v1alpha1SourceSpec.setResources(v1alpha1SourceSpecResources);

        V1alpha1SourceSpecPulsar v1alpha1SourceSpecPulsar = new V1alpha1SourceSpecPulsar();
        v1alpha1SourceSpecPulsar.setPulsarConfig(CommonUtil.getPulsarClusterConfigMapName(clusterName));
        // TODO: auth
        // v1alpha1SourceSpecPulsar.setAuthConfig(CommonUtil.getPulsarClusterAuthConfigMapName(clusterName));
        v1alpha1SourceSpec.setPulsar(v1alpha1SourceSpecPulsar);

        v1alpha1SourceSpec.setClusterName(clusterName);

        v1alpha1SourceSpec.setSourceConfig(CommonUtil.transformedMapValueToString(sourceConfig.getConfigs()));

        v1alpha1Source.setSpec(v1alpha1SourceSpec);

        return v1alpha1Source;

    }

    public static SourceConfig createSourceConfigFromV1alpha1Source(String tenant, String namespace, String sourceName,
                                                                    V1alpha1Source v1alpha1Source) {
        SourceConfig sourceConfig = new SourceConfig();

        sourceConfig.setName(sourceName);
        sourceConfig.setNamespace(namespace);
        sourceConfig.setTenant(tenant);

        V1alpha1SourceSpec v1alpha1SourceSpec = v1alpha1Source.getSpec();

        if (v1alpha1SourceSpec == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Source CRD without Spec defined.");
        }
        sourceConfig.setParallelism(v1alpha1SourceSpec.getReplicas());
        if (v1alpha1SourceSpec.getProcessingGuarantee() != null) {
            sourceConfig.setProcessingGuarantees(
                    CommonUtil.convertProcessingGuarantee(v1alpha1SourceSpec.getProcessingGuarantee()));
        }
        sourceConfig.setClassName(v1alpha1SourceSpec.getClassName());

        CustomRuntimeOptions customRuntimeOptions = new CustomRuntimeOptions();

        if (v1alpha1SourceSpec.getOutput() != null) {
            if (Strings.isNotEmpty(v1alpha1SourceSpec.getOutput().getTopic())) {
                sourceConfig.setTopicName(v1alpha1SourceSpec.getOutput().getTopic());
            }
            if (Strings.isNotEmpty(v1alpha1SourceSpec.getOutput().getSinkSerdeClassName())) {
                sourceConfig.setSerdeClassName(v1alpha1SourceSpec.getOutput().getSinkSerdeClassName());
            }
            if (Strings.isNotEmpty(v1alpha1SourceSpec.getOutput().getSinkSchemaType())) {
                sourceConfig.setSchemaType(v1alpha1SourceSpec.getOutput().getSinkSchemaType());
            }
            if (v1alpha1SourceSpec.getOutput().getProducerConf() != null) {
                ProducerConfig producerConfig = new ProducerConfig();
                Integer maxPendingMessages = v1alpha1SourceSpec.getOutput().getProducerConf().getMaxPendingMessages();
                if (maxPendingMessages != null && maxPendingMessages != 0) {
                    producerConfig.setMaxPendingMessages(maxPendingMessages);
                }
                Integer maxPendingMessagesAcrossPartitions = v1alpha1SourceSpec.getOutput()
                        .getProducerConf().getMaxPendingMessagesAcrossPartitions();
                if (maxPendingMessagesAcrossPartitions != null && maxPendingMessagesAcrossPartitions != 0) {
                    producerConfig.setMaxPendingMessagesAcrossPartitions(maxPendingMessagesAcrossPartitions);
                }
                if (Strings.isNotEmpty(v1alpha1SourceSpec.getOutput().getProducerConf().getBatchBuilder())) {
                    producerConfig.setBatchBuilder(v1alpha1SourceSpec.getOutput()
                            .getProducerConf().getBatchBuilder());
                }
                producerConfig.setUseThreadLocalProducers(v1alpha1SourceSpec.getOutput()
                        .getProducerConf().getUseThreadLocalProducers());
                sourceConfig.setProducerConfig(producerConfig);
            }
            customRuntimeOptions.setOutputTypeClassName(v1alpha1SourceSpec.getOutput().getTypeClassName());
        }

        if (Strings.isNotEmpty(v1alpha1SourceSpec.getClusterName())) {
            customRuntimeOptions.setClusterName(v1alpha1SourceSpec.getClusterName());
        }

        if (v1alpha1SourceSpec.getSourceConfig() != null) {
            sourceConfig.setConfigs(new HashMap<>(v1alpha1SourceSpec.getSourceConfig()));
        }

        // TODO: secretsMap

        Resources resources = new Resources();
        Map<String, String> sourceResource = v1alpha1SourceSpec.getResources().getRequests();
        Quantity cpuQuantity = Quantity.fromString(sourceResource.get(cpuKey));
        Quantity memoryQuantity = Quantity.fromString(sourceResource.get(memoryKey));
        resources.setCpu(cpuQuantity.getNumber().doubleValue());
        resources.setRam(memoryQuantity.getNumber().longValue());
        sourceConfig.setResources(resources);

        String customRuntimeOptionsJSON = new Gson().toJson(customRuntimeOptions, CustomRuntimeOptions.class);
        sourceConfig.setCustomRuntimeOptions(customRuntimeOptionsJSON);

        if (Strings.isNotEmpty(v1alpha1SourceSpec.getRuntimeFlags())) {
            sourceConfig.setRuntimeFlags(v1alpha1SourceSpec.getRuntimeFlags());
        }

        if (v1alpha1SourceSpec.getJava() != null && Strings.isNotEmpty(v1alpha1SourceSpec.getJava().getJar())) {
            sourceConfig.setArchive(v1alpha1SourceSpec.getJava().getJar());
        }

        return sourceConfig;
    }

    private static V1alpha1SourceSpecOutputProducerConfCryptoConfig convertFromCryptoSpec(Function.CryptoSpec cryptoSpec) {
        // TODO: convertFromCryptoSpec
        return null;
    }
}
