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
package io.streamnative.function.mesh.proxy.rest.api;

import io.kubernetes.client.openapi.models.V1ObjectMeta;
import io.streamnative.cloud.models.function.*;
import io.streamnative.function.mesh.proxy.FunctionMeshProxyService;
import lombok.extern.slf4j.Slf4j;
import okhttp3.Call;
import org.apache.commons.lang3.StringUtils;
import org.apache.pulsar.broker.authentication.AuthenticationDataHttps;
import org.apache.pulsar.broker.authentication.AuthenticationDataSource;
import org.apache.pulsar.common.functions.ConsumerConfig;
import org.apache.pulsar.common.functions.FunctionConfig;
import org.apache.pulsar.common.functions.UpdateOptions;
import org.apache.pulsar.common.policies.data.FunctionStatus;
import org.apache.pulsar.common.util.RestException;
import org.apache.pulsar.functions.proto.Function;
import org.apache.pulsar.functions.worker.service.api.Functions;
import org.glassfish.jersey.media.multipart.FormDataContentDisposition;

import javax.ws.rs.core.Response;
import java.io.InputStream;
import java.net.URI;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.Map;
import java.util.function.Supplier;

@Slf4j
public class FunctionsImpl extends FunctionMeshComponentImpl implements Functions<FunctionMeshProxyService> {
    final String Kind = "Function";

    public FunctionsImpl(Supplier<FunctionMeshProxyService> functionMeshProxyServiceSupplier) {
        super(functionMeshProxyServiceSupplier, Function.FunctionDetails.ComponentType.FUNCTION);
    }

    @Override
    public void registerFunction(final String tenant,
                                 final String namespace,
                                 final String functionName,
                                 final InputStream uploadedInputStream,
                                 final FormDataContentDisposition fileDetail,
                                 final String functionPkgUrl,
                                 final FunctionConfig functionConfig,
                                 final String clientRole,
                                 AuthenticationDataHttps clientAuthenticationDataHttps) {
        if (tenant == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Tenant is not provided");
        }
        if (namespace == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Namespace is not provided");
        }
        if (functionName == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Function name is not provided");
        }
        if (functionConfig == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "Function config is not provided");
        }

        V1alpha1Function v1alpha1Function = new V1alpha1Function();
        v1alpha1Function.setKind(Kind);
        v1alpha1Function.setApiVersion(String.format("%s/%s", group, version));

        V1ObjectMeta v1ObjectMeta = new V1ObjectMeta();
        v1ObjectMeta.setName(functionConfig.getName());
        v1ObjectMeta.setNamespace(functionConfig.getNamespace());
        v1alpha1Function.setMetadata(v1ObjectMeta);

        V1alpha1FunctionSpec v1alpha1FunctionSpec = new V1alpha1FunctionSpec();
        v1alpha1FunctionSpec.setClassName(functionConfig.getClassName());

        ConsumerConfig consumerConfig = functionConfig.getInputSpecs().get("source");
        if (consumerConfig == null || StringUtils.isBlank(consumerConfig.getSerdeClassName())) {
            throw new RestException(Response.Status.BAD_REQUEST, "inputSpecs.source.serdeClassName is not provided");
        }
        if (StringUtils.isBlank(functionConfig.getOutputSerdeClassName())) {
            throw new RestException(Response.Status.BAD_REQUEST, "outputSerdeClassName is not provided");
        }
        v1alpha1FunctionSpec.setSourceType(consumerConfig.getSerdeClassName());
        v1alpha1FunctionSpec.setSinkType(functionConfig.getOutputSerdeClassName());

        v1alpha1FunctionSpec.setForwardSourceMessageProperty(functionConfig.getForwardSourceMessageProperty());
        v1alpha1FunctionSpec.setMaxPendingAsyncRequests(functionConfig.getMaxPendingAsyncRequests());

        Integer parallelism = functionConfig.getParallelism() == null ? 1 : functionConfig.getParallelism();
        v1alpha1FunctionSpec.setReplicas(parallelism);
        v1alpha1FunctionSpec.setMaxReplicas(parallelism);

        v1alpha1FunctionSpec.setLogTopic(functionConfig.getLogTopic());

        V1alpha1FunctionSpecInput v1alpha1FunctionSpecInput = new V1alpha1FunctionSpecInput();
        v1alpha1FunctionSpecInput.setTopics(new ArrayList<>(functionConfig.getInputs()));
        v1alpha1FunctionSpec.setInput(v1alpha1FunctionSpecInput);

        V1alpha1FunctionSpecOutput v1alpha1FunctionSpecOutput = new V1alpha1FunctionSpecOutput();
        v1alpha1FunctionSpecOutput.setTopic(functionConfig.getOutput());
        v1alpha1FunctionSpec.setOutput(v1alpha1FunctionSpecOutput);

        V1alpha1FunctionSpecResources v1alpha1FunctionSpecResources = new V1alpha1FunctionSpecResources();
        Map<String, Object> limits = new HashMap<>();
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
        String memoryValue = memory.toString() + "G";
        limits.put("cpu", cpuValue);
        limits.put("memory", memoryValue);
        Map<String, Object> requests = new HashMap<>();
        limits.put("cpu", cpuValue);
        limits.put("memory", memoryValue);
        v1alpha1FunctionSpecResources.setLimits(limits);
        v1alpha1FunctionSpecResources.setRequests(requests);
        v1alpha1FunctionSpec.setResources(v1alpha1FunctionSpecResources);

        V1alpha1FunctionSpecPulsar v1alpha1FunctionSpecPulsar = new V1alpha1FunctionSpecPulsar();
        v1alpha1FunctionSpecPulsar.setPulsarConfig(functionName);
        v1alpha1FunctionSpec.setPulsar(v1alpha1FunctionSpecPulsar);

        String location = String.format("%s/%s/%s",tenant,namespace,functionName);
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
        Object clusterName = functionConfig.getUserConfig().get("clusterName");
        if (clusterName == null) {
            throw new RestException(Response.Status.BAD_REQUEST, "userConfig.clusterName is not provided");
        }

        v1alpha1FunctionSpec.setClusterName(clusterName.toString());
        v1alpha1FunctionSpec.setAutoAck(functionConfig.getAutoAck());

        v1alpha1Function.setSpec(v1alpha1FunctionSpec);

        try {
            Call call = worker().getCustomObjectsApi().createNamespacedCustomObjectCall(group, version, namespace,
                    plural,
                    v1alpha1Function,
                    null,
                    null,
                    null,
                    null);
            V1alpha1Function res = executeCall(call, V1alpha1Function.class);
            if (res == null) {
                throw new RestException(Response.Status.INTERNAL_SERVER_ERROR, "failed to create this function: " +
                        "failed to create custom object");
            }
        } catch (Exception e) {
            log.error("register {}/{}/{} function failed, error message: {}", tenant, namespace, functionName, e);
            throw new RestException(Response.Status.INTERNAL_SERVER_ERROR, e.getMessage());
        }
    }

    @Override
    public void updateFunction(final String tenant,
                               final String namespace,
                               final String functionName,
                               final InputStream uploadedInputStream,
                               final FormDataContentDisposition fileDetail,
                               final String functionPkgUrl,
                               final FunctionConfig functionConfig,
                               final String clientRole,
                               AuthenticationDataHttps clientAuthenticationDataHttps,
                               UpdateOptions updateOptions) {

    }

    @Override
    public FunctionStatus.FunctionInstanceStatus.FunctionInstanceStatusData getFunctionInstanceStatus(final String tenant,
                                                                                                      final String namespace,
                                                                                                      final String componentName,
                                                                                                      final String instanceId,
                                                                                                      final URI uri,
                                                                                                      final String clientRole,
                                                                                                      final AuthenticationDataSource clientAuthenticationDataHttps) {

        return null;
    }

    @Override
    public FunctionStatus getFunctionStatus(final String tenant,
                                            final String namespace,
                                            final String componentName,
                                            final URI uri,
                                            final String clientRole,
                                            final AuthenticationDataSource clientAuthenticationDataHttps) {
        FunctionStatus functionStatus = new FunctionStatus();
        try {
            Call call = worker().getCustomObjectsApi().getNamespacedCustomObjectCall(
                    group, version, namespace, plural, componentName, null);
            V1alpha1Function v1alpha1Function = executeCall(call, V1alpha1Function.class);
            FunctionStatus.FunctionInstanceStatus functionInstanceStatus = new FunctionStatus.FunctionInstanceStatus();
            FunctionStatus.FunctionInstanceStatus.FunctionInstanceStatusData functionInstanceStatusData =
                    new FunctionStatus.FunctionInstanceStatus.FunctionInstanceStatusData();
            functionInstanceStatusData.setRunning(true);
            if (v1alpha1Function.getStatus() != null) {
                v1alpha1Function.getStatus().getConditions().forEach((s, v1alpha1FunctionStatusConditions) -> {
                    if (v1alpha1FunctionStatusConditions.getStatus() != null
                            && v1alpha1FunctionStatusConditions.getStatus().equals("False")) {
                        functionInstanceStatusData.setRunning(false);
                    }
                });
                functionInstanceStatusData.setWorkerId(v1alpha1Function.getSpec().getClusterName());
                functionInstanceStatus.setStatus(functionInstanceStatusData);
                functionStatus.addInstance(functionInstanceStatus);
            } else {
                functionInstanceStatusData.setRunning(false);
            }
            functionStatus.setNumInstances(functionStatus.getInstances().size());
        } catch (Exception e) {
            log.error("Get function {} status failed from namespace {}, error message: {}",
                    componentName, namespace, e.getMessage());
        }

        return functionStatus;
    }

    @Override
    public void updateFunctionOnWorkerLeader(final String tenant,
                                             final String namespace,
                                             final String functionName,
                                             final InputStream uploadedInputStream,
                                             final boolean delete,
                                             URI uri,
                                             final String clientRole) {

    }
}
