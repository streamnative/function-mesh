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
import io.functionmesh.compute.MeshWorkerService;
import io.functionmesh.compute.functions.models.V1alpha1Function;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpec;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecGolang;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecInput;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecInputCryptoConfig;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecInputSourceSpecs;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecJava;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecOutput;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecOutputProducerConf;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecPod;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecPodResources;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecPulsar;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecPython;
import io.functionmesh.compute.models.CustomRuntimeOptions;
import io.functionmesh.compute.models.MeshWorkerServiceCustomConfig;
import io.kubernetes.client.custom.Quantity;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.lang3.RandomStringUtils;
import org.apache.commons.lang3.StringUtils;
import org.apache.logging.log4j.util.Strings;
import org.apache.pulsar.client.admin.PulsarAdminException;
import org.apache.pulsar.common.functions.ConsumerConfig;
import org.apache.pulsar.common.functions.FunctionConfig;
import org.apache.pulsar.common.functions.ProducerConfig;
import org.apache.pulsar.common.functions.Resources;
import org.apache.pulsar.common.functions.Utils;
import org.apache.pulsar.common.policies.data.ExceptionInformation;
import org.apache.pulsar.common.policies.data.FunctionStatus;
import org.apache.pulsar.common.util.RestException;
import org.apache.pulsar.functions.proto.Function;
import org.apache.pulsar.functions.proto.InstanceCommunication;
import org.apache.pulsar.functions.utils.FunctionCommon;
import org.apache.pulsar.functions.utils.FunctionConfigUtils;

import javax.ws.rs.core.Response;
import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.LinkedList;
import java.util.List;
import java.util.Map;

import static io.functionmesh.compute.util.CommonUtil.getExceptionInformation;

@Slf4j
public class FunctionsUtil {
    public final static String cpuKey = "cpu";
    public final static String memoryKey = "memory";
    public final static String sourceKey = "source";

    public static V1alpha1Function createV1alpha1FunctionFromFunctionConfig(String kind, String group, String version
            , String functionName, String functionPkgUrl, FunctionConfig functionConfig
            , String cluster, MeshWorkerService worker) {
        MeshWorkerServiceCustomConfig customConfig = worker.getMeshWorkerServiceCustomConfig();
        CustomRuntimeOptions customRuntimeOptions = CommonUtil.getCustomRuntimeOptions(functionConfig.getCustomRuntimeOptions());
        String clusterName = CommonUtil.getClusterName(cluster, customRuntimeOptions);
        String serviceAccountName = customRuntimeOptions.getServiceAccountName();

        Function.FunctionDetails functionDetails;
        try {
            functionDetails = FunctionConfigUtils.convert(functionConfig, null);
        } catch (IllegalArgumentException ex) {
            log.error("cannot convert FunctionConfig to FunctionDetails", ex);
            throw new RestException(Response.Status.BAD_REQUEST, "functionConfig cannot be parsed into functionDetails");
        }

        V1alpha1Function v1alpha1Function = new V1alpha1Function();
        v1alpha1Function.setKind(kind);
        v1alpha1Function.setApiVersion(String.format("%s/%s", group, version));
        v1alpha1Function.setMetadata(CommonUtil.makeV1ObjectMeta(functionConfig.getName(),
                functionConfig.getNamespace(),
                functionDetails.getNamespace(),
                functionDetails.getTenant(),
                clusterName,
                CommonUtil.getOwnerReferenceFromCustomConfigs(customConfig)));

        V1alpha1FunctionSpec v1alpha1FunctionSpec = new V1alpha1FunctionSpec();
        v1alpha1FunctionSpec.setClassName(functionConfig.getClassName());

        V1alpha1FunctionSpecInput v1alpha1FunctionSpecInput = new V1alpha1FunctionSpecInput();

        for (Map.Entry<String, Function.ConsumerSpec> inputSpecs : functionDetails.getSource().getInputSpecsMap().entrySet()) {
            V1alpha1FunctionSpecInputSourceSpecs inputSourceSpecsItem = new V1alpha1FunctionSpecInputSourceSpecs();
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
            v1alpha1FunctionSpecInput.putSourceSpecsItem(inputSpecs.getKey(), inputSourceSpecsItem);
        }

        if (Strings.isNotEmpty(customRuntimeOptions.getInputTypeClassName())) {
            v1alpha1FunctionSpecInput.setTypeClassName(customRuntimeOptions.getInputTypeClassName());
        }

        if (functionConfig.getInputs() != null) {
            v1alpha1FunctionSpecInput.setTopics(new ArrayList<>(functionConfig.getInputs()));
        }

        if ((v1alpha1FunctionSpecInput.getTopics() == null || v1alpha1FunctionSpecInput.getTopics().size() == 0) &&
                (v1alpha1FunctionSpecInput.getSourceSpecs() == null || v1alpha1FunctionSpecInput.getSourceSpecs().size() == 0)
        ) {
            log.warn("invalid FunctionSpecInput {}", v1alpha1FunctionSpecInput);
            throw new RestException(Response.Status.BAD_REQUEST, "invalid FunctionSpecInput");
        }
        v1alpha1FunctionSpec.setInput(v1alpha1FunctionSpecInput);

        if (!StringUtils.isEmpty(functionDetails.getSource().getSubscriptionName())) {
            v1alpha1FunctionSpec.setSubscriptionName(functionDetails.getSource().getSubscriptionName());
        }
        v1alpha1FunctionSpec.setRetainOrdering(functionDetails.getRetainOrdering());
        v1alpha1FunctionSpec.setRetainKeyOrdering(functionDetails.getRetainKeyOrdering());

        v1alpha1FunctionSpec.setCleanupSubscription(functionDetails.getSource().getCleanupSubscription());
        v1alpha1FunctionSpec.setAutoAck(functionDetails.getAutoAck());

        if (functionDetails.getSource().getTimeoutMs() != 0) {
            v1alpha1FunctionSpec.setTimeout((int) functionDetails.getSource().getTimeoutMs());
        }

        V1alpha1FunctionSpecOutput v1alpha1FunctionSpecOutput = new V1alpha1FunctionSpecOutput();
        if (!StringUtils.isEmpty(functionDetails.getSink().getTopic())) {
            v1alpha1FunctionSpecOutput.setTopic(functionDetails.getSink().getTopic());
            // process CustomSchemaSinks
            if (functionDetails.getSink().getSchemaPropertiesCount() > 0
                    && functionDetails.getSink().getConsumerPropertiesCount() > 0) {
                Map<String, String> customSchemaSinks = new HashMap<>();
                if (functionConfig.getCustomSchemaOutputs() != null
                        && functionConfig.getCustomSchemaOutputs().containsKey(functionConfig.getOutput())) {
                    String conf = functionConfig.getCustomSchemaOutputs().get(functionConfig.getOutput());
                    customSchemaSinks.put(functionDetails.getSink().getTopic(), conf);
                }
                v1alpha1FunctionSpecOutput.customSchemaSinks(customSchemaSinks);
            }
        }
        if (!StringUtils.isEmpty(functionDetails.getSink().getSerDeClassName())) {
            v1alpha1FunctionSpecOutput.setSinkSerdeClassName(functionDetails.getSink().getSerDeClassName());
        }
        if (!StringUtils.isEmpty(functionDetails.getSink().getSchemaType())) {
            v1alpha1FunctionSpecOutput.setSinkSchemaType(functionDetails.getSink().getSchemaType());
        }
        // process ProducerConf
        V1alpha1FunctionSpecOutputProducerConf v1alpha1FunctionSpecOutputProducerConf
                = new V1alpha1FunctionSpecOutputProducerConf();
        Function.ProducerSpec producerSpec = functionDetails.getSink().getProducerSpec();
        if (Strings.isNotEmpty(producerSpec.getBatchBuilder())) {
            v1alpha1FunctionSpecOutputProducerConf.setBatchBuilder(producerSpec.getBatchBuilder());
        }
        v1alpha1FunctionSpecOutputProducerConf.setMaxPendingMessages(producerSpec.getMaxPendingMessages());
        v1alpha1FunctionSpecOutputProducerConf.setMaxPendingMessagesAcrossPartitions(
                producerSpec.getMaxPendingMessagesAcrossPartitions());
        v1alpha1FunctionSpecOutputProducerConf.useThreadLocalProducers(producerSpec.getUseThreadLocalProducers());
        if (producerSpec.hasCryptoSpec()) {
            v1alpha1FunctionSpecOutputProducerConf.setCryptoConfig(
                    convertFromCryptoSpec(producerSpec.getCryptoSpec()));
        }

        v1alpha1FunctionSpecOutput.setProducerConf(v1alpha1FunctionSpecOutputProducerConf);

        if (Strings.isNotEmpty(customRuntimeOptions.getOutputTypeClassName())) {
            v1alpha1FunctionSpecOutput.setTypeClassName(customRuntimeOptions.getOutputTypeClassName());
        }

        v1alpha1FunctionSpec.setOutput(v1alpha1FunctionSpecOutput);

        if (!StringUtils.isEmpty(functionDetails.getLogTopic())) {
            v1alpha1FunctionSpec.setLogTopic(functionDetails.getLogTopic());
        }
        v1alpha1FunctionSpec.setForwardSourceMessageProperty(functionDetails.getSink().getForwardSourceMessageProperty());

        if (functionDetails.hasRetryDetails()) {
            v1alpha1FunctionSpec.setMaxMessageRetry(functionDetails.getRetryDetails().getMaxMessageRetries());
            if (!StringUtils.isEmpty(functionDetails.getRetryDetails().getDeadLetterTopic())) {
                v1alpha1FunctionSpec.setDeadLetterTopic(functionDetails.getRetryDetails().getDeadLetterTopic());
            }
        }

        v1alpha1FunctionSpec.setMaxPendingAsyncRequests(functionConfig.getMaxPendingAsyncRequests());

        v1alpha1FunctionSpec.setReplicas(functionDetails.getParallelism());
        if (customRuntimeOptions.getMaxReplicas() > functionDetails.getParallelism()) {
            v1alpha1FunctionSpec.setMaxReplicas(customRuntimeOptions.getMaxReplicas());
        }

        v1alpha1FunctionSpec.setLogTopic(functionConfig.getLogTopic());

        V1alpha1FunctionSpecPodResources v1alpha1FunctionSpecResources = new V1alpha1FunctionSpecPodResources();

        double cpu = functionConfig.getResources() != null &&
                functionConfig.getResources().getCpu() != 0 ? functionConfig.getResources().getCpu() : 1;
        long ramRequest = functionConfig.getResources() != null &&
                functionConfig.getResources().getRam() != 0 ? functionConfig.getResources().getRam() : 1073741824;

        Map<String, String> limits = new HashMap<>();
        Map<String, String> requests = new HashMap<>();

        long padding = Math.round(ramRequest * (10.0 / 100.0)); // percentMemoryPadding is 0.1
        long ramWithPadding = ramRequest + padding;

        limits.put(cpuKey, Quantity.fromString(Double.toString(cpu)).toSuffixedString());
        limits.put(memoryKey, Quantity.fromString(Long.toString(ramWithPadding)).toSuffixedString());

        requests.put(cpuKey, Quantity.fromString(Double.toString(cpu)).toSuffixedString());
        requests.put(memoryKey, Quantity.fromString(Long.toString(ramRequest)).toSuffixedString());

        v1alpha1FunctionSpecResources.setLimits(limits);
        v1alpha1FunctionSpecResources.setRequests(requests);
        v1alpha1FunctionSpec.setResources(v1alpha1FunctionSpecResources);

        V1alpha1FunctionSpecPulsar v1alpha1FunctionSpecPulsar = new V1alpha1FunctionSpecPulsar();
        v1alpha1FunctionSpecPulsar.setPulsarConfig(CommonUtil.getPulsarClusterConfigMapName(clusterName));
        // TODO: auth
        // v1alpha1FunctionSpecPulsar.setAuthConfig(CommonUtil.getPulsarClusterAuthConfigMapName(clusterName));
        v1alpha1FunctionSpec.setPulsar(v1alpha1FunctionSpecPulsar);

        // TODO: dynamic file name to function CRD
        String fileName = "/pulsar/function-executable";
        boolean isPkgUrlProvided = StringUtils.isNotEmpty(functionPkgUrl);
        File componentPackageFile = null;
        try {
            if (isPkgUrlProvided) {
                if (Utils.hasPackageTypePrefix(functionPkgUrl)) {
                    componentPackageFile = downloadPackageFile(worker, functionPkgUrl);
                } else {
                    log.warn("get unsupported function package url {}", functionPkgUrl);
                    throw new IllegalArgumentException("Function Package url is not valid. supported url (function/sink/source)");
                }
            } else {
                // TODO: support upload JAR to bk
                throw new IllegalArgumentException("uploading package to mesh worker service is not supported yet.");
            }
        } catch (Exception e) {
            log.error("Invalid register function request {}: {}", functionName, e);
            throw new RestException(Response.Status.BAD_REQUEST, e.getMessage());
        }
        Class<?>[] typeArgs = null;
        if (componentPackageFile != null) {
            typeArgs = extractTypeArgs(functionConfig, componentPackageFile, worker.getWorkerConfig().isForwardSourceMessageProperty());
        }
        if (StringUtils.isNotEmpty(functionConfig.getJar())) {
            V1alpha1FunctionSpecJava v1alpha1FunctionSpecJava = new V1alpha1FunctionSpecJava();
            v1alpha1FunctionSpecJava.setJar(fileName);
            if (isPkgUrlProvided) {
                v1alpha1FunctionSpecJava.setJarLocation(functionPkgUrl);
            }
            String extraDependenciesDir = "";
            if (StringUtils.isNotEmpty(worker.getFactoryConfig().getExtraFunctionDependenciesDir())) {
                if (Paths.get(worker.getFactoryConfig().getExtraFunctionDependenciesDir()).isAbsolute()) {
                    extraDependenciesDir = worker.getFactoryConfig().getExtraFunctionDependenciesDir();
                } else {
                    extraDependenciesDir = "/pulsar/" + worker.getFactoryConfig().getExtraFunctionDependenciesDir();
                }
            } else {
                extraDependenciesDir = "/pulsar/instances/deps";
            }
            v1alpha1FunctionSpecJava.setExtraDependenciesDir(extraDependenciesDir);
            v1alpha1FunctionSpec.setJava(v1alpha1FunctionSpecJava);
            if (typeArgs != null) {
                if (typeArgs.length == 2 && typeArgs[0] != null) {
                    v1alpha1FunctionSpecInput.setTypeClassName(typeArgs[0].getName());
                }
                if (typeArgs.length == 2 && typeArgs[1] != null) {
                    v1alpha1FunctionSpecOutput.setTypeClassName(typeArgs[1].getName());
                }
            }
        } else if (StringUtils.isNotEmpty(functionConfig.getPy())) {
            V1alpha1FunctionSpecPython v1alpha1FunctionSpecPython = new V1alpha1FunctionSpecPython();
            v1alpha1FunctionSpecPython.setPy(fileName);
            if (isPkgUrlProvided) {
                v1alpha1FunctionSpecPython.setPyLocation(functionPkgUrl);
            }
            v1alpha1FunctionSpec.setPython(v1alpha1FunctionSpecPython);
        } else if (StringUtils.isNotEmpty(functionConfig.getGo())) {
            V1alpha1FunctionSpecGolang v1alpha1FunctionSpecGolang = new V1alpha1FunctionSpecGolang();
            v1alpha1FunctionSpecGolang.setGo(fileName);
            if (isPkgUrlProvided) {
                v1alpha1FunctionSpecGolang.setGoLocation(functionPkgUrl);
            }
            v1alpha1FunctionSpec.setGolang(v1alpha1FunctionSpecGolang);
        }

        v1alpha1FunctionSpec.setClusterName(clusterName);
        v1alpha1FunctionSpec.setAutoAck(functionConfig.getAutoAck());

        V1alpha1FunctionSpecPod specPod = new V1alpha1FunctionSpecPod();
        if (worker.getMeshWorkerServiceCustomConfig().isAllowUserDefinedServiceAccountName() &&
                StringUtils.isNotEmpty(serviceAccountName)) {
            specPod.setServiceAccountName(serviceAccountName);
            v1alpha1FunctionSpec.setPod(specPod);
        }

        v1alpha1Function.setSpec(v1alpha1FunctionSpec);

        return v1alpha1Function;
    }

    private static V1alpha1FunctionSpecInputCryptoConfig convertFromCryptoSpec(Function.CryptoSpec cryptoSpec) {
        // TODO: convertFromCryptoSpec
        return null;
    }

    public static FunctionConfig createFunctionConfigFromV1alpha1Function(String tenant, String namespace,
                                                                          String functionName,
                                                                          V1alpha1Function v1alpha1Function) {
        FunctionConfig functionConfig = new FunctionConfig();

        functionConfig.setName(functionName);
        functionConfig.setNamespace(namespace);
        functionConfig.setTenant(tenant);

        V1alpha1FunctionSpec v1alpha1FunctionSpec = v1alpha1Function.getSpec();

        if (v1alpha1FunctionSpec == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Function CRD without Spec defined.");
        }
        functionConfig.setParallelism(v1alpha1FunctionSpec.getReplicas());
        if (v1alpha1FunctionSpec.getProcessingGuarantee() != null) {
            functionConfig.setProcessingGuarantees(
                    CommonUtil.convertProcessingGuarantee(v1alpha1FunctionSpec.getProcessingGuarantee()));
        }

        CustomRuntimeOptions customRuntimeOptions = new CustomRuntimeOptions();

        Map<String, ConsumerConfig> consumerConfigMap = new HashMap<>();
        V1alpha1FunctionSpecInput v1alpha1FunctionSpecInput = v1alpha1FunctionSpec.getInput();
        if (v1alpha1FunctionSpecInput == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "FunctionSpec CRD without Input defined.");
        }
        if (Strings.isNotEmpty(v1alpha1FunctionSpecInput.getTypeClassName())) {
            customRuntimeOptions.setInputTypeClassName(v1alpha1FunctionSpecInput.getTypeClassName());
        }

        if (Strings.isNotEmpty(v1alpha1FunctionSpec.getClusterName())) {
            customRuntimeOptions.setClusterName(v1alpha1FunctionSpec.getClusterName());
        }

        if (v1alpha1FunctionSpec.getMaxReplicas() != null && v1alpha1FunctionSpec.getMaxReplicas() > 0) {
            customRuntimeOptions.setMaxReplicas(v1alpha1FunctionSpec.getMaxReplicas());
        }

        if (v1alpha1FunctionSpec.getPod() != null &&
                Strings.isNotEmpty(v1alpha1FunctionSpec.getPod().getServiceAccountName())) {
            customRuntimeOptions.setServiceAccountName(v1alpha1FunctionSpec.getPod().getServiceAccountName());
        }

        if (v1alpha1FunctionSpecInput.getTopics() != null) {
            for (String topic : v1alpha1FunctionSpecInput.getTopics()) {
                ConsumerConfig consumerConfig = new ConsumerConfig();
                consumerConfig.setRegexPattern(false);
                consumerConfigMap.put(topic, consumerConfig);
            }
        }

        if (Strings.isNotEmpty(v1alpha1FunctionSpecInput.getTopicPattern())) {
            String patternTopic = v1alpha1FunctionSpecInput.getTopicPattern();
            ConsumerConfig consumerConfig = consumerConfigMap.getOrDefault(patternTopic, new ConsumerConfig());
            consumerConfig.setRegexPattern(true);
            consumerConfigMap.put(patternTopic, consumerConfig);
        }

        if (v1alpha1FunctionSpecInput.getCustomSerdeSources() != null) {
            for (Map.Entry<String, String> source : v1alpha1FunctionSpecInput.getCustomSerdeSources().entrySet()) {
                String topic = source.getKey();
                String serdeClassName = source.getValue();
                ConsumerConfig consumerConfig = consumerConfigMap.getOrDefault(topic, new ConsumerConfig());
                consumerConfig.setRegexPattern(false);
                consumerConfig.setSerdeClassName(serdeClassName);
                consumerConfigMap.put(topic, consumerConfig);
            }
        }

        if (v1alpha1FunctionSpecInput.getSourceSpecs() != null) {
            for (Map.Entry<String, V1alpha1FunctionSpecInputSourceSpecs> source : v1alpha1FunctionSpecInput.getSourceSpecs().entrySet()) {
                String topic = source.getKey();
                V1alpha1FunctionSpecInputSourceSpecs sourceSpecs = source.getValue();
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

        functionConfig.setInputSpecs(consumerConfigMap);
        functionConfig.setInputs(consumerConfigMap.keySet());

        if (Strings.isNotEmpty(v1alpha1FunctionSpec.getSubscriptionName())) {
            functionConfig.setSubName(v1alpha1FunctionSpec.getSubscriptionName());
        }
        if (v1alpha1FunctionSpec.getRetainOrdering() != null) {
            functionConfig.setRetainOrdering(v1alpha1FunctionSpec.getRetainOrdering());
        }
        if (v1alpha1FunctionSpec.getRetainKeyOrdering() != null) {
            functionConfig.setRetainKeyOrdering(v1alpha1FunctionSpec.getRetainKeyOrdering());
        }
        if (v1alpha1FunctionSpec.getCleanupSubscription() != null) {
            functionConfig.setCleanupSubscription(v1alpha1FunctionSpec.getCleanupSubscription());
        }
        if (v1alpha1FunctionSpec.getAutoAck() != null) {
            functionConfig.setAutoAck(v1alpha1FunctionSpec.getAutoAck());
        } else {
            functionConfig.setAutoAck(true);
        }
        if (v1alpha1FunctionSpec.getTimeout() != null && v1alpha1FunctionSpec.getTimeout() != 0) {
            functionConfig.setTimeoutMs(v1alpha1FunctionSpec.getTimeout().longValue());
        }
        if (v1alpha1FunctionSpec.getOutput() != null) {
            if (Strings.isNotEmpty(v1alpha1FunctionSpec.getOutput().getTopic())) {
                functionConfig.setOutput(v1alpha1FunctionSpec.getOutput().getTopic());
            }
            if (Strings.isNotEmpty(v1alpha1FunctionSpec.getOutput().getSinkSerdeClassName())) {
                functionConfig.setOutputSerdeClassName(v1alpha1FunctionSpec.getOutput().getSinkSerdeClassName());
            }
            if (Strings.isNotEmpty(v1alpha1FunctionSpec.getOutput().getSinkSchemaType())) {
                functionConfig.setOutputSchemaType(v1alpha1FunctionSpec.getOutput().getSinkSchemaType());
            }
            if (v1alpha1FunctionSpec.getOutput().getProducerConf() != null) {
                ProducerConfig producerConfig = new ProducerConfig();
                Integer maxPendingMessages = v1alpha1FunctionSpec.getOutput().getProducerConf().getMaxPendingMessages();
                if (maxPendingMessages != null && maxPendingMessages != 0) {
                    producerConfig.setMaxPendingMessages(maxPendingMessages);
                }
                Integer maxPendingMessagesAcrossPartitions = v1alpha1FunctionSpec.getOutput()
                        .getProducerConf().getMaxPendingMessagesAcrossPartitions();
                if (maxPendingMessagesAcrossPartitions != null && maxPendingMessagesAcrossPartitions != 0) {
                    producerConfig.setMaxPendingMessagesAcrossPartitions(maxPendingMessagesAcrossPartitions);
                }
                if (Strings.isNotEmpty(v1alpha1FunctionSpec.getOutput().getProducerConf().getBatchBuilder())) {
                    producerConfig.setBatchBuilder(v1alpha1FunctionSpec.getOutput()
                            .getProducerConf().getBatchBuilder());
                }
                producerConfig.setUseThreadLocalProducers(v1alpha1FunctionSpec.getOutput()
                        .getProducerConf().getUseThreadLocalProducers());
                functionConfig.setProducerConfig(producerConfig);
            }
            customRuntimeOptions.setOutputTypeClassName(v1alpha1FunctionSpec.getOutput().getTypeClassName());
        }
        if (Strings.isNotEmpty(v1alpha1FunctionSpec.getLogTopic())) {
            functionConfig.setLogTopic(v1alpha1FunctionSpec.getLogTopic());
        }
        if (v1alpha1FunctionSpec.getForwardSourceMessageProperty() != null) {
            functionConfig.setForwardSourceMessageProperty(v1alpha1FunctionSpec.getForwardSourceMessageProperty());
        }
        if (v1alpha1FunctionSpec.getJava() != null) {
            functionConfig.setRuntime(FunctionConfig.Runtime.JAVA);
            functionConfig.setJar(v1alpha1FunctionSpec.getJava().getJar());
            if (Strings.isNotEmpty(v1alpha1FunctionSpec.getJava().getJarLocation())) {
                functionConfig.setJar(v1alpha1FunctionSpec.getJava().getJarLocation());
            }
        } else if (v1alpha1FunctionSpec.getPython() != null) {
            functionConfig.setRuntime(FunctionConfig.Runtime.PYTHON);
            functionConfig.setPy(v1alpha1FunctionSpec.getPython().getPy());
            if (Strings.isNotEmpty(v1alpha1FunctionSpec.getPython().getPyLocation())) {
                functionConfig.setJar(v1alpha1FunctionSpec.getPython().getPyLocation());
            }
        } else if (v1alpha1FunctionSpec.getGolang() != null) {
            functionConfig.setRuntime(FunctionConfig.Runtime.GO);
            functionConfig.setGo(v1alpha1FunctionSpec.getGolang().getGo());
            if (Strings.isNotEmpty(v1alpha1FunctionSpec.getGolang().getGoLocation())) {
                functionConfig.setJar(v1alpha1FunctionSpec.getGolang().getGoLocation());
            }
        }
        if (v1alpha1FunctionSpec.getMaxMessageRetry() != null) {
            functionConfig.setMaxMessageRetries(v1alpha1FunctionSpec.getMaxMessageRetry());
            if (Strings.isNotEmpty(v1alpha1FunctionSpec.getDeadLetterTopic())) {
                functionConfig.setDeadLetterTopic(v1alpha1FunctionSpec.getDeadLetterTopic());
            }
        }
        functionConfig.setClassName(v1alpha1FunctionSpec.getClassName());

        // TODO: secretsMap
        // TODO: externalPulsarConfig

        Resources resources = new Resources();
        Map<String, String> functionResource = v1alpha1FunctionSpec.getResources().getLimits();
        resources.setCpu(Double.parseDouble(functionResource.get(cpuKey)));
        resources.setRam(Long.parseLong(functionResource.get(memoryKey)));
        functionConfig.setResources(resources);

        String customRuntimeOptionsJSON = new Gson().toJson(customRuntimeOptions, CustomRuntimeOptions.class);
        functionConfig.setCustomRuntimeOptions(customRuntimeOptionsJSON);

        if (Strings.isNotEmpty(v1alpha1FunctionSpec.getRuntimeFlags())) {
            functionConfig.setRuntimeFlags(v1alpha1FunctionSpec.getRuntimeFlags());
        }

        return functionConfig;
    }

    public static void convertFunctionStatusToInstanceStatusData(InstanceCommunication.FunctionStatus functionStatus,
                                                                 FunctionStatus.FunctionInstanceStatus.FunctionInstanceStatusData functionInstanceStatusData) {
        if (functionStatus == null || functionInstanceStatusData == null) {
            return;
        }
        functionInstanceStatusData.setRunning(functionStatus.getRunning());
        functionInstanceStatusData.setError(functionStatus.getFailureException());
        functionInstanceStatusData.setNumRestarts(functionStatus.getNumRestarts());
        functionInstanceStatusData.setNumReceived(functionStatus.getNumReceived());
        functionInstanceStatusData.setNumSuccessfullyProcessed(functionStatus.getNumSuccessfullyProcessed());
        functionInstanceStatusData.setNumUserExceptions(functionStatus.getNumUserExceptions());

        List<ExceptionInformation> userExceptionInformationList = new LinkedList<>();
        for (InstanceCommunication.FunctionStatus.ExceptionInformation exceptionEntry : functionStatus.getLatestUserExceptionsList()) {
            ExceptionInformation exceptionInformation = getExceptionInformation(exceptionEntry);
            userExceptionInformationList.add(exceptionInformation);
        }
        functionInstanceStatusData.setLatestUserExceptions(userExceptionInformationList);

        // For regular functions source/sink errors are system exceptions
        functionInstanceStatusData.setNumSystemExceptions(functionStatus.getNumSystemExceptions()
                + functionStatus.getNumSourceExceptions() + functionStatus.getNumSinkExceptions());
        List<ExceptionInformation> systemExceptionInformationList = new LinkedList<>();
        for (InstanceCommunication.FunctionStatus.ExceptionInformation exceptionEntry : functionStatus.getLatestSystemExceptionsList()) {
            ExceptionInformation exceptionInformation = getExceptionInformation(exceptionEntry);
            systemExceptionInformationList.add(exceptionInformation);
        }
        for (InstanceCommunication.FunctionStatus.ExceptionInformation exceptionEntry : functionStatus.getLatestSourceExceptionsList()) {
            ExceptionInformation exceptionInformation = getExceptionInformation(exceptionEntry);
            systemExceptionInformationList.add(exceptionInformation);
        }
        for (InstanceCommunication.FunctionStatus.ExceptionInformation exceptionEntry : functionStatus.getLatestSinkExceptionsList()) {
            ExceptionInformation exceptionInformation = getExceptionInformation(exceptionEntry);
            systemExceptionInformationList.add(exceptionInformation);
        }
        functionInstanceStatusData.setLatestSystemExceptions(systemExceptionInformationList);

        functionInstanceStatusData.setAverageLatency(functionStatus.getAverageLatency());
        functionInstanceStatusData.setLastInvocationTime(functionStatus.getLastInvocationTime());
    }

    private static File downloadPackageFile(MeshWorkerService worker, String packageName) throws IOException, PulsarAdminException {
        Path tempDirectory;
        if (worker.getWorkerConfig().getDownloadDirectory() != null) {
            tempDirectory = Paths.get(worker.getWorkerConfig().getDownloadDirectory());
        } else {
            // use the Nar extraction directory as a temporary directory for downloaded files
            tempDirectory = Paths.get(worker.getWorkerConfig().getNarExtractionDirectory());
        }
        if (Files.notExists(tempDirectory)) {
            Files.createDirectories(tempDirectory);
        }
        String fileName = String.format("function-%s.tmp", RandomStringUtils.random(5, true, true).toLowerCase());
        Path filePath = Paths.get(tempDirectory.toString(), fileName);
        Files.deleteIfExists(filePath);
        worker.getBrokerAdmin().packages().download(packageName, filePath.toString());
        return filePath.toFile();
    }

    private static Class<?>[] extractTypeArgs(final FunctionConfig functionConfig,
                                             final File componentPackageFile,
                                              final boolean isForwardSourceMessageProperty) {
        Class<?>[] typeArgs = null;
        FunctionConfigUtils.inferMissingArguments(
                functionConfig, isForwardSourceMessageProperty);
        if (componentPackageFile == null) {
            return null;
        }
        ClassLoader clsLoader = FunctionConfigUtils.validate(functionConfig, componentPackageFile);
        if (functionConfig.getRuntime() == FunctionConfig.Runtime.JAVA && clsLoader != null) {
            try {
                typeArgs = FunctionCommon.getFunctionTypes(functionConfig, clsLoader);
            } catch (ClassNotFoundException | NoClassDefFoundError e) {
                throw new IllegalArgumentException(
                        String.format("Function class %s must be in class path", functionConfig.getClassName()), e);
            }
        }
        return typeArgs;
    }

}
