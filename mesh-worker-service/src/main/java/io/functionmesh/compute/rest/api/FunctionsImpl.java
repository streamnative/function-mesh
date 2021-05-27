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
package io.functionmesh.compute.rest.api;

import com.google.common.collect.Maps;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecJava;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecPod;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecPodVolumeMounts;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecPodVolumes;
import io.functionmesh.compute.util.FunctionsUtil;
import io.functionmesh.compute.functions.models.V1alpha1Function;
import io.functionmesh.compute.MeshWorkerService;
import io.functionmesh.compute.util.KubernetesUtils;
import lombok.extern.slf4j.Slf4j;
import okhttp3.Call;
import org.apache.commons.lang.StringUtils;
import org.apache.pulsar.broker.authentication.AuthenticationDataHttps;
import org.apache.pulsar.broker.authentication.AuthenticationDataSource;
import org.apache.pulsar.common.functions.FunctionConfig;
import org.apache.pulsar.common.functions.Resources;
import org.apache.pulsar.common.functions.UpdateOptions;
import org.apache.pulsar.common.policies.data.FunctionStatus;
import org.apache.pulsar.common.util.RestException;
import org.apache.pulsar.functions.proto.Function;
import org.apache.pulsar.functions.utils.ComponentTypeUtils;
import org.apache.pulsar.functions.worker.service.api.Functions;
import org.glassfish.jersey.media.multipart.FormDataContentDisposition;
import javax.ws.rs.core.Response;
import java.io.InputStream;
import java.net.URI;
import java.util.List;
import java.util.Map;
import java.util.function.Supplier;

@Slf4j
public class FunctionsImpl extends MeshComponentImpl implements Functions<MeshWorkerService> {

    public FunctionsImpl(Supplier<MeshWorkerService> meshWorkerServiceSupplier) {
        super(meshWorkerServiceSupplier, Function.FunctionDetails.ComponentType.FUNCTION);
    }


    private void validateRegisterFunctionRequestParams(String tenant, String namespace, String functionName,
                                                       FunctionConfig functionConfig, boolean jarUploaded) {
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
        Map<String, Object> customConfig = worker().getWorkerConfig().getFunctionsWorkerServiceCustomConfigs();
        if (jarUploaded &&  customConfig != null && customConfig.get("uploadEnabled") != null &&
                ! (Boolean) customConfig.get("uploadEnabled") ) {
            throw new RestException(Response.Status.BAD_REQUEST, "Uploading Jar File is not enabled");
        }
        this.validateResources(functionConfig.getResources(), worker().getWorkerConfig().getFunctionInstanceMinResources(),
                worker().getWorkerConfig().getFunctionInstanceMaxResources());
    }

    private void validateUpdateFunctionRequestParams(String tenant, String namespace, String functionName,
                                                     FunctionConfig functionConfig, boolean uploadedJar) {
        validateRegisterFunctionRequestParams(tenant, namespace, functionName, functionConfig, uploadedJar);
    }

    private void validateGetFunctionInfoRequestParams(String tenant, String namespace, String functionName) {
        this.validateGetInfoRequestParams(tenant, namespace, functionName, kind);
    }

    private void validateFunctionEnabled() {
        Map<String, Object> customConfig = worker().getWorkerConfig().getFunctionsWorkerServiceCustomConfigs();
        if (customConfig != null) {
            Boolean functionEnabled = (Boolean) customConfig.get("functionEnabled");
            if (functionEnabled != null && !functionEnabled) {
                throw new RestException(Response.Status.BAD_REQUEST, "Function API is disabled");
            }
        }
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
        validateFunctionEnabled();

        validateRegisterFunctionRequestParams(tenant, namespace, functionName, functionConfig, uploadedInputStream != null);
        this.validatePermission(tenant,
                namespace,
                clientRole,
                clientAuthenticationDataHttps,
                ComponentTypeUtils.toString(componentType));
        this.validateTenantIsExist(tenant, namespace, functionName, clientRole);

        V1alpha1Function v1alpha1Function = FunctionsUtil.createV1alpha1FunctionFromFunctionConfig(
                kind,
                group,
                version,
                functionName,
                functionPkgUrl,
                functionConfig,
                worker().getWorkerConfig().getFunctionsWorkerServiceCustomConfigs()
        );
        // override namespace by configuration file
        v1alpha1Function.getMetadata().setNamespace(KubernetesUtils.getNamespace(worker().getFactoryConfig()));
        Map<String, String> customLabels = Maps.newHashMap();
        customLabels.put(TENANT_LABEL_CLAIM, tenant);
        customLabels.put(NAMESPACE_LABEL_CLAIM, namespace);
        V1alpha1FunctionSpecPod pod = new V1alpha1FunctionSpecPod();
        if (worker().getFactoryConfig() != null && worker().getFactoryConfig().getCustomLabels() != null) {
            customLabels.putAll(worker().getFactoryConfig().getCustomLabels());
        }
        pod.setLabels(customLabels);
        v1alpha1Function.getSpec().setPod(pod);
        try {
            this.upsertFunction(tenant, namespace, functionName, functionConfig, v1alpha1Function, clientAuthenticationDataHttps);
            Call call = worker().getCustomObjectsApi().createNamespacedCustomObjectCall(
                    group,
                    version,
                    KubernetesUtils.getNamespace(worker().getFactoryConfig()),
                    plural,
                    v1alpha1Function,
                    null,
                    null,
                    null,
                    null
            );
            executeCall(call, V1alpha1Function.class);
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
        validateFunctionEnabled();

        validateUpdateFunctionRequestParams(tenant, namespace, functionName, functionConfig, uploadedInputStream != null);

        try {
            Call getCall = worker().getCustomObjectsApi().getNamespacedCustomObjectCall(
                    group,
                    version,
                    KubernetesUtils.getNamespace(worker().getFactoryConfig()),
                    plural,
                    functionName,
                    null
            );
            V1alpha1Function oldFn = executeCall(getCall, V1alpha1Function.class);
            V1alpha1Function v1alpha1Function = FunctionsUtil.createV1alpha1FunctionFromFunctionConfig(
                    kind,
                    group,
                    version,
                    functionName,
                    functionPkgUrl,
                    functionConfig,
                    worker().getWorkerConfig().getFunctionsWorkerServiceCustomConfigs()
            );
            v1alpha1Function.getMetadata().setResourceVersion(oldFn.getMetadata().getResourceVersion());
            this.upsertFunction(tenant, namespace, functionName, functionConfig, v1alpha1Function, clientAuthenticationDataHttps);
            Call replaceCall = worker().getCustomObjectsApi().replaceNamespacedCustomObjectCall(
                    group,
                    version,
                    KubernetesUtils.getNamespace(worker().getFactoryConfig()),
                    plural,
                    functionName,
                    v1alpha1Function,
                    null,
                    null,
                    null
            );
            executeCall(replaceCall, V1alpha1Function.class);
        } catch (Exception e) {
            log.error("update {}/{}/{} function failed, error message: {}", tenant, namespace, functionName, e);
            throw new RestException(Response.Status.INTERNAL_SERVER_ERROR, e.getMessage());
        }
    }

    @Override
    public FunctionConfig getFunctionInfo(final String tenant,
                                          final String namespace,
                                          final String componentName,
                                          final String clientRole,
                                          final AuthenticationDataSource clientAuthenticationDataHttps) {
        validateFunctionEnabled();
        validateGetFunctionInfoRequestParams(tenant, namespace, componentName);

        this.validatePermission(tenant,
                namespace,
                clientRole,
                clientAuthenticationDataHttps,
                ComponentTypeUtils.toString(componentType));

        try {
            Call call = worker().getCustomObjectsApi().getNamespacedCustomObjectCall(
                    group,
                    version,
                    KubernetesUtils.getNamespace(worker().getFactoryConfig()),
                    plural,
                    componentName,
                    null
            );
            V1alpha1Function v1alpha1Function = executeCall(call, V1alpha1Function.class);
            return FunctionsUtil.createFunctionConfigFromV1alpha1Function(tenant, namespace, componentName,
                    v1alpha1Function);
        } catch (Exception e) {
            log.error("get {}/{}/{} function failed, error message: {}", tenant, namespace, componentName, e);
            throw new RestException(Response.Status.INTERNAL_SERVER_ERROR, e.getMessage());
        }
    }

    @Override
    public FunctionStatus.FunctionInstanceStatus.FunctionInstanceStatusData getFunctionInstanceStatus(final String tenant,
                                                                                                      final String namespace,
                                                                                                      final String componentName,
                                                                                                      final String instanceId,
                                                                                                      final URI uri,
                                                                                                      final String clientRole,
                                                                                                      final AuthenticationDataSource clientAuthenticationDataHttps) {

        throw new RestException(Response.Status.BAD_REQUEST, "Unsupported Operation");
    }

    @Override
    public FunctionStatus getFunctionStatus(final String tenant,
                                            final String namespace,
                                            final String componentName,
                                            final URI uri,
                                            final String clientRole,
                                            final AuthenticationDataSource clientAuthenticationDataHttps) {
        this.validatePermission(tenant,
                namespace,
                clientRole,
                clientAuthenticationDataHttps,
                ComponentTypeUtils.toString(componentType));
        FunctionStatus functionStatus = new FunctionStatus();
        try {
            Call call = worker().getCustomObjectsApi().getNamespacedCustomObjectCall(
                    group, version, KubernetesUtils.getNamespace(worker().getFactoryConfig()),
                    plural, componentName, null);
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
                                             final String clientRole,
                                             final AuthenticationDataSource clientAuthenticationDataHttps) {
        throw new RestException(Response.Status.BAD_REQUEST, "Unsupported Operation");
    }

    private void upsertFunction(final String tenant,
                                final String namespace,
                                final String functionName,
                                final FunctionConfig functionConfig,
                                V1alpha1Function v1alpha1Function,
                                AuthenticationDataHttps clientAuthenticationDataHttps) {
        if (worker().getWorkerConfig().isAuthenticationEnabled()) {
            if (clientAuthenticationDataHttps != null) {
                try {

                    Map<String, Object>  functionsWorkerServiceCustomConfigs = worker()
                            .getWorkerConfig().getFunctionsWorkerServiceCustomConfigs();
                    Object volumes = functionsWorkerServiceCustomConfigs.get("volumes");
                    if (volumes != null) {
                        List<V1alpha1FunctionSpecPodVolumes> volumesList = (List<V1alpha1FunctionSpecPodVolumes>) volumes;
                        v1alpha1Function.getSpec().getPod().setVolumes(volumesList);
                    }
                    Object volumeMounts = functionsWorkerServiceCustomConfigs.get("volumeMounts");
                    if (volumeMounts != null) {
                        List<V1alpha1FunctionSpecPodVolumeMounts> volumeMountsList = (List<V1alpha1FunctionSpecPodVolumeMounts>) volumeMounts;
                        v1alpha1Function.getSpec().setVolumeMounts(volumeMountsList);
                    }
                    if (functionsWorkerServiceCustomConfigs.get("extraDependenciesDir") != null) {
                        V1alpha1FunctionSpecJava v1alpha1FunctionSpecJava;
                        if (v1alpha1Function.getSpec() != null && v1alpha1Function.getSpec().getJava() != null) {
                            v1alpha1FunctionSpecJava = v1alpha1Function.getSpec().getJava();
                        } else {
                            v1alpha1FunctionSpecJava = new V1alpha1FunctionSpecJava();
                        }
                        v1alpha1FunctionSpecJava.setExtraDependenciesDir(
                                (String)functionsWorkerServiceCustomConfigs.get("extraDependenciesDir"));
                        v1alpha1Function.getSpec().setJava(v1alpha1FunctionSpecJava);
                    }
                    if (!StringUtils.isEmpty(worker().getWorkerConfig().getBrokerClientAuthenticationPlugin())
                            && !StringUtils.isEmpty(worker().getWorkerConfig().getBrokerClientAuthenticationParameters())) {
                        String authSecretName = KubernetesUtils.upsertSecret(kind.toLowerCase(), "auth",
                                v1alpha1Function.getSpec().getClusterName(), tenant, namespace, functionName,
                                worker().getWorkerConfig(), worker().getCoreV1Api(), worker().getFactoryConfig());
                        v1alpha1Function.getSpec().getPulsar().setAuthSecret(authSecretName);
                    }
                    if (worker().getWorkerConfig().getTlsEnabled()) {
                        String tlsSecretName = KubernetesUtils.upsertSecret(kind.toLowerCase(), "tls",
                                v1alpha1Function.getSpec().getClusterName(), tenant, namespace, functionName,
                                worker().getWorkerConfig(), worker().getCoreV1Api(), worker().getFactoryConfig());
                        v1alpha1Function.getSpec().getPulsar().setTlsSecret(tlsSecretName);
                    }
                } catch (Exception e) {
                    log.error("Error create or update auth or tls secret for {} {}/{}/{}",
                            ComponentTypeUtils.toString(componentType), tenant, namespace, functionName, e);


                    throw new RestException(Response.Status.INTERNAL_SERVER_ERROR,
                            String.format("Error create or update auth or tls secret for %s %s:- %s",
                            ComponentTypeUtils.toString(componentType), functionName, e.getMessage()));
                }
            }
        }
    }
}
