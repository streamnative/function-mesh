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
import io.functionmesh.compute.sinks.models.V1alpha1Sink;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpec;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecInput;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecInputCryptoConfig;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecInputSourceSpecs;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecJava;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecPulsar;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecPodResources;
import io.functionmesh.compute.worker.MeshConnectorsManager;
import io.kubernetes.client.custom.Quantity;
import lombok.extern.slf4j.Slf4j;
import io.kubernetes.client.openapi.models.V1ObjectMeta;
import org.apache.commons.io.FileUtils;
import org.apache.commons.lang.StringUtils;
import org.apache.logging.log4j.util.Strings;
import org.apache.pulsar.common.functions.ConsumerConfig;
import org.apache.pulsar.common.functions.Resources;
import org.apache.pulsar.common.io.SinkConfig;
import org.apache.pulsar.common.nar.NarClassLoader;
import org.apache.pulsar.common.util.RestException;
import org.apache.pulsar.functions.proto.Function;
import org.apache.pulsar.functions.utils.FunctionCommon;
import org.apache.pulsar.functions.utils.SinkConfigUtils;
import org.apache.pulsar.functions.utils.io.ConnectorUtils;

import javax.ws.rs.core.Response;
import java.io.File;
import java.io.InputStream;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.Map;

import static org.apache.pulsar.common.functions.Utils.BUILTIN;

@Slf4j
public class SinksUtil {
    public static final String cpuKey = "cpu";
    public static final String memoryKey = "memory";

    public static V1alpha1Sink createV1alpha1SkinFromSinkConfig(String kind, String group, String version
            , String sinkName, String sinkPkgUrl, InputStream uploadedInputStream, SinkConfig sinkConfig,
                                                                MeshConnectorsManager connectorsManager) {
        V1alpha1Sink v1alpha1Sink = new V1alpha1Sink();
        v1alpha1Sink.setKind(kind);
        v1alpha1Sink.setApiVersion(String.format("%s/%s", group, version));
        v1alpha1Sink.setMetadata(CommonUtil.makeV1ObjectMeta(sinkConfig.getName(), sinkConfig.getNamespace()));

        V1alpha1SinkSpec v1alpha1SinkSpec = new V1alpha1SinkSpec();
        v1alpha1SinkSpec.setClassName(sinkConfig.getClassName());

        Integer parallelism = sinkConfig.getParallelism() == null ? 1 : sinkConfig.getParallelism();
        v1alpha1SinkSpec.setReplicas(parallelism);
        v1alpha1SinkSpec.setMaxReplicas(parallelism);

        String customRuntimeOptionsJSON = sinkConfig.getCustomRuntimeOptions();
        CustomRuntimeOptions customRuntimeOptions = null;
        if (Strings.isEmpty(customRuntimeOptionsJSON)) {
            throw new RestException(
                    Response.Status.BAD_REQUEST, "customRuntimeOptions is not provided");
        }

        String clusterName = CommonUtil.getCurrentClusterName();
        try {
            customRuntimeOptions =
                    new Gson().fromJson(customRuntimeOptionsJSON, CustomRuntimeOptions.class);
        } catch (Exception ignored) {
            throw new RestException(
                    Response.Status.BAD_REQUEST, "customRuntimeOptions cannot be deserialized.");
        }

        if (Strings.isNotEmpty(customRuntimeOptions.getClusterName())) {
            clusterName = customRuntimeOptions.getClusterName();
        }

        if (Strings.isEmpty(clusterName)) {
            throw new RestException(Response.Status.BAD_REQUEST, "clusterName is not provided.");
        }

        String location =
                String.format(
                        "%s/%s/%s",
                        sinkConfig.getTenant(), sinkConfig.getNamespace(), sinkConfig.getName());
        if (StringUtils.isNotEmpty(sinkPkgUrl)) {
            location = sinkPkgUrl;
        }
        String archive = sinkConfig.getArchive();
        SinkConfigUtils.ExtractedSinkDetails extractedSinkDetails =
                new SinkConfigUtils.ExtractedSinkDetails("", customRuntimeOptions.getInputTypeClassName());
        V1alpha1SinkSpecJava v1alpha1SinkSpecJava = new V1alpha1SinkSpecJava();
        if (connectorsManager != null && archive.startsWith(BUILTIN)) {
            String connectorType = archive.replaceFirst("^builtin://", "");
            FunctionMeshConnectorDefinition definition = connectorsManager.getConnectorDefinition(connectorType);
            if (definition != null) {
                v1alpha1SinkSpec.setImage(definition.toFullImageURL());
                if (definition.getSinkClass() != null && v1alpha1SinkSpec.getClassName() == null) {
                    v1alpha1SinkSpec.setClassName(definition.getSinkClass());
                    extractedSinkDetails.setSinkClassName(definition.getSinkClass());
                }
                v1alpha1SinkSpecJava.setJar(definition.getJar());
                v1alpha1SinkSpecJava.setJarLocation("");
                v1alpha1SinkSpec.setJava(v1alpha1SinkSpecJava);
            } else {
                log.warn("cannot find built-in connector {}", connectorType);
                throw new RestException(Response.Status.BAD_REQUEST, String.format("connectorType %s is not supported yet", connectorType));
            }
        } else {
            v1alpha1SinkSpecJava.setJar(sinkConfig.getArchive());
            v1alpha1SinkSpecJava.setJarLocation(location);
            v1alpha1SinkSpec.setJava(v1alpha1SinkSpecJava);
            extractedSinkDetails.setSinkClassName(sinkConfig.getClassName());
        }

        Function.FunctionDetails functionDetails = null;
        try {
            functionDetails = SinkConfigUtils.convert(sinkConfig, extractedSinkDetails);
        } catch (Exception ex) {
            log.error("cannot convert SinkConfig to FunctionDetails", ex);
            throw new RestException(Response.Status.BAD_REQUEST, "functionConfig cannot be parsed into functionDetails");
        }

        V1alpha1SinkSpecInput v1alpha1SinkSpecInput = new V1alpha1SinkSpecInput();

        for (Map.Entry<String, Function.ConsumerSpec> inputSpecs : functionDetails.getSource().getInputSpecsMap().entrySet()) {
            V1alpha1SinkSpecInputSourceSpecs inputSourceSpecsItem = new V1alpha1SinkSpecInputSourceSpecs();
            if (Strings.isNotEmpty(inputSpecs.getValue().getSerdeClassName())) {
                inputSourceSpecsItem.setSerdeClassname(inputSpecs.getValue().getSerdeClassName());
            }
            if (Strings.isNotEmpty(inputSpecs.getValue().getSchemaType())) {
                inputSourceSpecsItem.setSchemaType(inputSpecs.getValue().getSchemaType());
            }
            if (inputSpecs.getValue().hasReceiverQueueSize()) {
                inputSourceSpecsItem.setReceiverQueueSize(inputSpecs.getValue().getReceiverQueueSize().getValue());
            }
            if (inputSpecs.getValue().hasCryptoSpec()) {
                inputSourceSpecsItem.setCryptoConfig(convertFromCryptoSpec(inputSpecs.getValue().getCryptoSpec()));
            }
            inputSourceSpecsItem.setIsRegexPattern(inputSpecs.getValue().getIsRegexPattern());
            inputSourceSpecsItem.setSchemaProperties(inputSpecs.getValue().getSchemaPropertiesMap());
            v1alpha1SinkSpecInput.putSourceSpecsItem(inputSpecs.getKey(), inputSourceSpecsItem);
        }

        if (Strings.isNotEmpty(customRuntimeOptions.getInputTypeClassName())) {
            v1alpha1SinkSpecInput.setTypeClassName(customRuntimeOptions.getInputTypeClassName());
        }

        v1alpha1SinkSpecInput.setTopics(new ArrayList<>(sinkConfig.getInputs()));
        v1alpha1SinkSpec.setInput(v1alpha1SinkSpecInput);

        if (Strings.isNotEmpty(functionDetails.getSource().getSubscriptionName())) {
            v1alpha1SinkSpec.setSubscriptionName(functionDetails.getSource().getSubscriptionName());
        }
        v1alpha1SinkSpec.setRetainOrdering(functionDetails.getRetainOrdering());

        v1alpha1SinkSpec.setCleanupSubscription(functionDetails.getSource().getCleanupSubscription());
        v1alpha1SinkSpec.setAutoAck(functionDetails.getAutoAck());

        if (functionDetails.getSource().getTimeoutMs() != 0) {
            v1alpha1SinkSpec.setTimeout((int) functionDetails.getSource().getTimeoutMs());
        }

        if (Strings.isNotEmpty(functionDetails.getLogTopic())) {
            v1alpha1SinkSpec.setLogTopic(functionDetails.getLogTopic());
        }

        if (functionDetails.hasRetryDetails()) {
            v1alpha1SinkSpec.setMaxMessageRetry(functionDetails.getRetryDetails().getMaxMessageRetries());
            if (Strings.isNotEmpty(functionDetails.getRetryDetails().getDeadLetterTopic())) {
                v1alpha1SinkSpec.setDeadLetterTopic(functionDetails.getRetryDetails().getDeadLetterTopic());
            }
        }

        v1alpha1SinkSpec.setReplicas(functionDetails.getParallelism());
        v1alpha1SinkSpec.setMaxReplicas(functionDetails.getParallelism());

        double cpu = sinkConfig.getResources() != null &&
                sinkConfig.getResources().getCpu() != 0 ? sinkConfig.getResources().getCpu() : 1;
        long ramRequest = sinkConfig.getResources() != null &&
                sinkConfig.getResources().getRam() != 0 ? sinkConfig.getResources().getRam() : 1073741824;

        Map<String, String> limits = new HashMap<>();
        Map<String, String> requests = new HashMap<>();

        long padding = Math.round(ramRequest * (10.0 / 100.0)); // percentMemoryPadding is 0.1
        long ramWithPadding = ramRequest + padding;

        limits.put(cpuKey, Quantity.fromString(Double.toString(cpu)).toSuffixedString());
        limits.put(memoryKey, Quantity.fromString(Long.toString(ramWithPadding)).toSuffixedString());

        requests.put(cpuKey, Quantity.fromString(Double.toString(cpu)).toSuffixedString());
        requests.put(memoryKey, Quantity.fromString(Long.toString(ramRequest)).toSuffixedString());

        V1alpha1SinkSpecPodResources v1alpha1SinkSpecResources = new V1alpha1SinkSpecPodResources();
        v1alpha1SinkSpecResources.setLimits(limits);
        v1alpha1SinkSpecResources.setRequests(requests);
        v1alpha1SinkSpec.setResources(v1alpha1SinkSpecResources);

        V1alpha1SinkSpecPulsar v1alpha1SinkSpecPulsar = new V1alpha1SinkSpecPulsar();
        v1alpha1SinkSpecPulsar.setPulsarConfig(CommonUtil.getPulsarClusterConfigMapName(clusterName));
        // TODO: auth
        // v1alpha1SinkSpecPulsar.setAuthConfig(CommonUtil.getPulsarClusterAuthConfigMapName(clusterName));
        v1alpha1SinkSpec.setPulsar(v1alpha1SinkSpecPulsar);

        v1alpha1SinkSpec.setClusterName(clusterName);
        v1alpha1SinkSpec.setAutoAck(sinkConfig.getAutoAck());

        v1alpha1SinkSpec.setSinkConfig(CommonUtil.transformedMapValueToString(sinkConfig.getConfigs()));

        v1alpha1Sink.setSpec(v1alpha1SinkSpec);

        return v1alpha1Sink;
    }

    public static SinkConfig createSinkConfigFromV1alpha1Source(
            String tenant, String namespace, String sinkName, V1alpha1Sink v1alpha1Sink) {
        SinkConfig sinkConfig = new SinkConfig();

        sinkConfig.setName(sinkName);
        sinkConfig.setNamespace(namespace);
        sinkConfig.setTenant(tenant);

        V1alpha1SinkSpec v1alpha1SinkSpec = v1alpha1Sink.getSpec();

        if (v1alpha1SinkSpec == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Sink CRD without Spec defined.");
        }
        sinkConfig.setParallelism(v1alpha1SinkSpec.getReplicas());
        if (v1alpha1SinkSpec.getProcessingGuarantee() != null) {
            sinkConfig.setProcessingGuarantees(CommonUtil.convertProcessingGuarantee(v1alpha1SinkSpec.getProcessingGuarantee()));
        }

        CustomRuntimeOptions customRuntimeOptions = new CustomRuntimeOptions();

        Map<String, ConsumerConfig> consumerConfigMap = new HashMap<>();
        V1alpha1SinkSpecInput v1alpha1SinkSpecInput = v1alpha1SinkSpec.getInput();
        if (v1alpha1SinkSpecInput == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "SinkSpec CRD without Input defined.");
        }
        if (Strings.isNotEmpty(v1alpha1SinkSpecInput.getTypeClassName())) {
            customRuntimeOptions.setInputTypeClassName(v1alpha1SinkSpecInput.getTypeClassName());
        }
        if (Strings.isNotEmpty(v1alpha1SinkSpec.getClusterName())) {
            customRuntimeOptions.setClusterName(v1alpha1SinkSpec.getClusterName());
        }

        if (v1alpha1SinkSpecInput.getTopics() != null) {
            for (String topic : v1alpha1SinkSpecInput.getTopics()) {
                ConsumerConfig consumerConfig = new ConsumerConfig();
                consumerConfig.setRegexPattern(false);
                consumerConfigMap.put(topic, consumerConfig);
            }
        }

        if (Strings.isNotEmpty(v1alpha1SinkSpecInput.getTopicPattern())) {
            String patternTopic = v1alpha1SinkSpecInput.getTopicPattern();
            ConsumerConfig consumerConfig = consumerConfigMap.getOrDefault(patternTopic, new ConsumerConfig());
            consumerConfig.setRegexPattern(true);
            consumerConfigMap.put(patternTopic, consumerConfig);
        }

        if (v1alpha1SinkSpecInput.getCustomSerdeSources() != null) {
            for (Map.Entry<String, String> source : v1alpha1SinkSpecInput.getCustomSerdeSources().entrySet()) {
                String topic = source.getKey();
                String serdeClassName = source.getValue();
                ConsumerConfig consumerConfig = consumerConfigMap.getOrDefault(topic, new ConsumerConfig());
                consumerConfig.setRegexPattern(false);
                consumerConfig.setSerdeClassName(serdeClassName);
                consumerConfigMap.put(topic, consumerConfig);
            }
        }

        if (v1alpha1SinkSpecInput.getSourceSpecs() != null) {
            for (Map.Entry<String, V1alpha1SinkSpecInputSourceSpecs> source : v1alpha1SinkSpecInput.getSourceSpecs().entrySet()) {
                String topic = source.getKey();
                V1alpha1SinkSpecInputSourceSpecs sourceSpecs = source.getValue();
                ConsumerConfig consumerConfig = consumerConfigMap.getOrDefault(topic, new ConsumerConfig());
                if (sourceSpecs.getIsRegexPattern() != null) {
                    consumerConfig.setRegexPattern(sourceSpecs.getIsRegexPattern());
                }
                consumerConfig.setSchemaType(sourceSpecs.getSchemaType());
                consumerConfig.setSerdeClassName(sourceSpecs.getSerdeClassname());
                consumerConfig.setReceiverQueueSize(sourceSpecs.getReceiverQueueSize());
                consumerConfig.setSchemaProperties(sourceSpecs.getSchemaProperties());
                consumerConfig.setConsumerProperties(sourceSpecs.getConsumerProperties());
                if (sourceSpecs.getCryptoConfig() != null) {
                    // TODO: convert CryptoConfig to function config
                }
                consumerConfigMap.put(topic, consumerConfig);
            }
        }

        sinkConfig.setInputSpecs(consumerConfigMap);
        sinkConfig.setInputs(consumerConfigMap.keySet());

        if (Strings.isNotEmpty(v1alpha1SinkSpec.getSubscriptionName())) {
            sinkConfig.setSourceSubscriptionName(v1alpha1SinkSpec.getSubscriptionName());
        }
        if (v1alpha1SinkSpec.getRetainOrdering() != null) {
            sinkConfig.setRetainOrdering(v1alpha1SinkSpec.getRetainOrdering());
        }
        if (v1alpha1SinkSpec.getCleanupSubscription() != null) {
            sinkConfig.setCleanupSubscription(v1alpha1SinkSpec.getCleanupSubscription());
        }
        if (v1alpha1SinkSpec.getAutoAck() != null) {
            sinkConfig.setAutoAck(v1alpha1SinkSpec.getAutoAck());
        } else {
            sinkConfig.setAutoAck(true);
        }
        if (v1alpha1SinkSpec.getTimeout() != null && v1alpha1SinkSpec.getTimeout() != 0) {
            sinkConfig.setTimeoutMs(v1alpha1SinkSpec.getTimeout().longValue());
        }
        if (v1alpha1SinkSpec.getMaxMessageRetry() != null) {
            sinkConfig.setMaxMessageRetries(v1alpha1SinkSpec.getMaxMessageRetry());
            if (Strings.isNotEmpty(v1alpha1SinkSpec.getDeadLetterTopic())) {
                sinkConfig.setDeadLetterTopic(v1alpha1SinkSpec.getDeadLetterTopic());
            }
        }
        sinkConfig.setClassName(v1alpha1SinkSpec.getClassName());
        if (v1alpha1SinkSpec.getSinkConfig() != null) {
            sinkConfig.setConfigs(new HashMap<>(v1alpha1SinkSpec.getSinkConfig()));
        }
        // TODO: secretsMap

        Resources resources = new Resources();
        Map<String, String> sinkResources = v1alpha1SinkSpec.getResources().getRequests();
        Quantity cpuQuantity = Quantity.fromString(sinkResources.get(cpuKey));
        Quantity memoryQuantity = Quantity.fromString(sinkResources.get(memoryKey));
        resources.setCpu(cpuQuantity.getNumber().doubleValue());
        resources.setRam(memoryQuantity.getNumber().longValue());
        sinkConfig.setResources(resources);

        String customRuntimeOptionsJSON = new Gson().toJson(customRuntimeOptions, CustomRuntimeOptions.class);
        sinkConfig.setCustomRuntimeOptions(customRuntimeOptionsJSON);

        if (Strings.isNotEmpty(v1alpha1SinkSpec.getRuntimeFlags())) {
            sinkConfig.setRuntimeFlags(v1alpha1SinkSpec.getRuntimeFlags());
        }

        if (v1alpha1SinkSpec.getJava() != null && Strings.isNotEmpty(v1alpha1SinkSpec.getJava().getJar())) {
            sinkConfig.setArchive(v1alpha1SinkSpec.getJava().getJar());
        }

        return sinkConfig;
    }

    private static V1alpha1SinkSpecInputCryptoConfig convertFromCryptoSpec(Function.CryptoSpec cryptoSpec) {
        // TODO: convertFromCryptoSpec
        return null;
    }

}
