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
import io.functionmesh.compute.functions.models.V1alpha1Function;
import io.functionmesh.compute.functions.models.V1alpha1FunctionList;
import io.functionmesh.compute.MeshWorkerService;
import io.functionmesh.compute.sinks.models.V1alpha1Sink;
import io.functionmesh.compute.sources.models.V1alpha1Source;
import io.functionmesh.compute.util.CommonUtil;
import io.functionmesh.compute.util.KubernetesUtils;
import lombok.extern.slf4j.Slf4j;
import okhttp3.Call;
import okhttp3.Response;
import org.apache.commons.codec.digest.DigestUtils;
import org.apache.commons.lang3.StringUtils;
import org.apache.pulsar.broker.authentication.AuthenticationDataHttps;
import org.apache.pulsar.broker.authentication.AuthenticationDataSource;
import org.apache.pulsar.client.admin.PulsarAdminException;
import org.apache.pulsar.common.functions.FunctionConfig;
import org.apache.pulsar.common.functions.FunctionState;
import org.apache.pulsar.common.functions.Resources;
import org.apache.pulsar.common.io.ConnectorDefinition;
import org.apache.pulsar.common.naming.NamespaceName;
import org.apache.pulsar.common.policies.data.FunctionStats;
import org.apache.pulsar.common.policies.data.TenantInfo;
import org.apache.pulsar.common.util.RestException;
import org.apache.pulsar.functions.proto.Function;
import org.apache.pulsar.functions.utils.ComponentTypeUtils;
import org.apache.pulsar.functions.worker.service.api.Component;

import javax.ws.rs.core.StreamingOutput;
import java.io.InputStream;
import java.net.URI;
import java.util.LinkedList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutionException;
import java.util.function.Supplier;

import static com.google.common.base.Preconditions.checkNotNull;
import static java.util.concurrent.TimeUnit.SECONDS;

@Slf4j
public abstract class MeshComponentImpl implements Component<MeshWorkerService> {

    protected final Supplier<MeshWorkerService> meshWorkerServiceSupplier;
    protected final Function.FunctionDetails.ComponentType componentType;

    String plural = "functions";

    final String group = "compute.functionmesh.io";

    final String version = "v1alpha1";

    String kind = "Function";

    final String CLUSTER_LABEL_CLAIM = "pulsar-cluster";

    final String TENANT_LABEL_CLAIM = "pulsar-tenant";

    final String NAMESPACE_LABEL_CLAIM = "pulsar-namespace";

    final String COMPONENT_LABEL_CLAIM = "pulsar-component";

    MeshComponentImpl(Supplier<MeshWorkerService> meshWorkerServiceSupplier,
                      Function.FunctionDetails.ComponentType componentType) {
        this.meshWorkerServiceSupplier = meshWorkerServiceSupplier;
        // If you want to support function-mesh, this type needs to be changed
        this.componentType = componentType;
    }

    @Override
    public FunctionConfig getFunctionInfo(final String tenant,
                                          final String namespace,
                                          final String componentName,
                                          final String clientRole,
                                          final AuthenticationDataSource clientAuthenticationDataHttps) {

        FunctionConfig functionConfig = new FunctionConfig();
        return functionConfig;
    }

    @Override
    public void deregisterFunction(final String tenant,
                                   final String namespace,
                                   final String componentName,
                                   final String clientRole,
                                   AuthenticationDataHttps clientAuthenticationDataHttps) {
        this.validateGetInfoRequestParams(tenant, namespace, componentName, kind);

        this.validatePermission(tenant,
                namespace,
                clientRole,
                clientAuthenticationDataHttps,
                ComponentTypeUtils.toString(componentType));
        try {
            String clusterName = worker().getWorkerConfig().getPulsarFunctionsCluster();
            String hashName = CommonUtil.createObjectName(clusterName, tenant, namespace, componentName);
            Call deleteObjectCall = worker().getCustomObjectsApi().deleteNamespacedCustomObjectCall(
                    group,
                    version,
                    KubernetesUtils.getNamespace(worker().getFactoryConfig()),
                    plural,
                    hashName,
                    null,
                    null,
                    null,
                    null,
                    null,
                    null
            );
            executeCall(deleteObjectCall, null);

            if (!StringUtils.isEmpty(worker().getWorkerConfig().getBrokerClientAuthenticationPlugin())
                    && !StringUtils.isEmpty(worker().getWorkerConfig().getBrokerClientAuthenticationParameters())) {
                Call deleteAuthSecretCall = worker().getCoreV1Api()
                        .deleteNamespacedSecretCall(
                                KubernetesUtils.getUniqueSecretName(
                                        kind.toLowerCase(),
                                        "auth",
                                        DigestUtils.sha256Hex(
                                                KubernetesUtils.getSecretName(
                                                        clusterName, tenant, namespace, componentName))),
                                KubernetesUtils.getNamespace(worker().getFactoryConfig()),
                                null,
                                null,
                                30,
                                false,
                                null,
                                null,
                                null
                        );
                executeCall(deleteAuthSecretCall, null);
            }
            if (worker().getWorkerConfig().getTlsEnabled()) {
                Call deleteTlsSecretCall = worker().getCoreV1Api()
                        .deleteNamespacedSecretCall(
                                KubernetesUtils.getUniqueSecretName(
                                        kind.toLowerCase(),
                                        "tls",
                                        DigestUtils.sha256Hex(
                                                KubernetesUtils.getSecretName(
                                                        clusterName, tenant, namespace, componentName))),
                                KubernetesUtils.getNamespace(worker().getFactoryConfig()),
                                null,
                                null,
                                30,
                                false,
                                null,
                                null,
                                null
                        );
                executeCall(deleteTlsSecretCall, null);
            }
        } catch (Exception e) {
            log.error("deregister {}/{}/{} {} failed", tenant, namespace, componentName, plural, e);
            throw new RestException(javax.ws.rs.core.Response.Status.INTERNAL_SERVER_ERROR, e.getMessage());
        }
    }

    public <T> T executeCall(Call call, Class<T> c) throws Exception {
        Response response;
        response = call.execute();
        if (response.isSuccessful() && response.body() != null) {
            String data = response.body().string();
            if (c == null) {
                return null;
            }
            return worker().getApiClient().getJSON().getGson().fromJson(data, c);
        } else {
            String body = response.body() != null ? response.body().string() : "";
            String err = String.format(
                    "failed to perform the request: responseCode: %s, responseMessage: %s, responseBody: %s",
                    response.code(), response.message(), body);
            throw new Exception(err);
        }
    }

    @Override
    public MeshWorkerService worker() {
        try {
            return checkNotNull(meshWorkerServiceSupplier.get());
        } catch (Throwable t) {
            log.info("Failed to get worker service", t);
            throw t;
        }
    }

    @Override
    public void stopFunctionInstance(final String tenant,
                                     final String namespace,
                                     final String componentName,
                                     final String instanceId,
                                     final URI uri,
                                     final String clientRole,
                                     final AuthenticationDataSource clientAuthenticationDataHttps) {

    }

    @Override
    public void startFunctionInstance(final String tenant,
                                      final String namespace,
                                      final String componentName,
                                      final String instanceId,
                                      final URI uri,
                                      final String clientRole,
                                      final AuthenticationDataSource clientAuthenticationDataHttps) {

    }

    @Override
    public void restartFunctionInstance(final String tenant,
                                        final String namespace,
                                        final String componentName,
                                        final String instanceId,
                                        final URI uri,
                                        final String clientRole,
                                        final AuthenticationDataSource clientAuthenticationDataHttps) {

    }

    @Override
    public void startFunctionInstances(final String tenant,
                                       final String namespace,
                                       final String componentName,
                                       final String clientRole,
                                       final AuthenticationDataSource clientAuthenticationDataHttps) {

    }

    @Override
    public void stopFunctionInstances(final String tenant,
                                      final String namespace,
                                      final String componentName,
                                      final String clientRole,
                                      final AuthenticationDataSource clientAuthenticationDataHttps) {

    }

    @Override
    public void restartFunctionInstances(final String tenant,
                                         final String namespace,
                                         final String componentName,
                                         final String clientRole,
                                         final AuthenticationDataSource clientAuthenticationDataHttps) {

    }

    @Override
    public FunctionStats getFunctionStats(final String tenant,
                                          final String namespace,
                                          final String componentName,
                                          final URI uri,
                                          final String clientRole,
                                          final AuthenticationDataSource clientAuthenticationDataHttps) {
        FunctionStats functionStats = new FunctionStats();

        return functionStats;
    }

    @Override
    public FunctionStats.FunctionInstanceStats.FunctionInstanceStatsData getFunctionsInstanceStats(final String tenant,
                                                                                                   final String namespace,
                                                                                                   final String componentName,
                                                                                                   final String instanceId,
                                                                                                   final URI uri,
                                                                                                   final String clientRole,
                                                                                                   final AuthenticationDataSource clientAuthenticationDataHttps) {
        return new FunctionStats.FunctionInstanceStats.FunctionInstanceStatsData();
    }

    @Override
    public String triggerFunction(final String tenant,
                                  final String namespace,
                                  final String functionName,
                                  final String input,
                                  final InputStream uploadedInputStream,
                                  final String topic,
                                  final String clientRole,
                                  final AuthenticationDataSource clientAuthenticationDataHttps) {
        return "";
    }

    @Override
    public List<String> listFunctions(final String tenant,
                                      final String namespace,
                                      final String clientRole,
                                      final AuthenticationDataSource clientAuthenticationDataHttps) {
        List<String> result = new LinkedList<>();
        try {
            String labelSelector;
            String cluster = worker().getWorkerConfig().getPulsarFunctionsCluster();
            labelSelector = String.format(
                    "%s=%s,%s=%s,%s=%s",
                    CLUSTER_LABEL_CLAIM, cluster,
                    TENANT_LABEL_CLAIM, tenant,
                    NAMESPACE_LABEL_CLAIM, namespace);
            Call call = worker().getCustomObjectsApi().listNamespacedCustomObjectCall(
                    group,
                    version,
                    KubernetesUtils.getNamespace(worker().getFactoryConfig()), plural,
                    "false",
                    null,
                    null,
                    labelSelector,
                    null,
                    null,
                    null,
                    false,
                    null);

            V1alpha1FunctionList list = executeCall(call, V1alpha1FunctionList.class);
            list.getItems().forEach(n -> result.add(n.getMetadata().getLabels().get(COMPONENT_LABEL_CLAIM)));
        } catch (Exception e) {
            log.error("failed to fetch functions list from namespace {}, error message: {}", namespace, e.getMessage());
        }

        return result;
    }

    @Override
    public FunctionState getFunctionState(final String tenant,
                                          final String namespace,
                                          final String functionName,
                                          final String key,
                                          final String clientRole,
                                          final AuthenticationDataSource clientAuthenticationDataHttps) {
        // To do
        return new FunctionState();
    }

    @Override
    public void putFunctionState(final String tenant,
                                 final String namespace,
                                 final String functionName,
                                 final String key,
                                 final FunctionState state,
                                 final String clientRole,
                                 final AuthenticationDataSource clientAuthenticationDataHttps) {

    }

    @Override
    public void uploadFunction(final InputStream uploadedInputStream,
                               final String path,
                               String clientRole,
                               final AuthenticationDataSource clientAuthenticationDataHttps) {

    }

    @Override
    public StreamingOutput downloadFunction(String path,
                                            String clientRole,
                                            AuthenticationDataHttps clientAuthenticationDataHttps) {
        // To do
        return null;
    }

    @Override
    public StreamingOutput downloadFunction(String tenant,
                                            String namespace,
                                            String componentName,
                                            String clientRole,
                                            AuthenticationDataHttps clientAuthenticationDataHttps) {
        // To do
        return null;
    }

    @Override
    public List<ConnectorDefinition> getListOfConnectors() {
        return meshWorkerServiceSupplier.get().getConnectorsManager().getConnectorDefinitions();
    }

    @Override
    public void reloadConnectors(String clientRole, final AuthenticationDataSource clientAuthenticationDataHttps) {
        meshWorkerServiceSupplier.get().getConnectorsManager().reloadConnectors();
    }

    public boolean isSuperUser(String clientRole, AuthenticationDataSource authenticationDataSource) {
        if (clientRole != null) {
            try {
                if ((worker().getWorkerConfig().getSuperUserRoles() != null
                        && worker().getWorkerConfig().getSuperUserRoles().contains(clientRole))) {
                    return true;
                }
                return worker().getAuthorizationService().isSuperUser(clientRole, authenticationDataSource)
                        .get(worker().getWorkerConfig().getZooKeeperOperationTimeoutSeconds(), SECONDS);
            } catch (InterruptedException e) {
                log.warn("Time-out {} sec while checking the role {} is a super user role ",
                        worker().getWorkerConfig().getZooKeeperOperationTimeoutSeconds(), clientRole);
                throw new RestException(javax.ws.rs.core.Response.Status.INTERNAL_SERVER_ERROR, e.getMessage());
            } catch (Exception e) {
                log.warn("Admin-client with Role - failed to check the role {} is a super user role {} ", clientRole,
                        e.getMessage(), e);
                throw new RestException(javax.ws.rs.core.Response.Status.INTERNAL_SERVER_ERROR, e.getMessage());
            }
        }
        return false;
    }

    public boolean isAuthorizedRole(String tenant, String namespace, String clientRole,
                                    AuthenticationDataSource authenticationData) throws PulsarAdminException {
        if (worker().getWorkerConfig().isAuthorizationEnabled()) {
            // skip authorization if client role is super-user
            if (isSuperUser(clientRole, authenticationData)) {
                return true;
            }

            if (clientRole != null) {
                try {
                    TenantInfo tenantInfo = worker().getBrokerAdmin().tenants().getTenantInfo(tenant);
                    if (tenantInfo != null && worker().getAuthorizationService().isTenantAdmin(tenant, clientRole, tenantInfo, authenticationData).get()) {
                        return true;
                    }
                } catch (PulsarAdminException.NotFoundException | InterruptedException | ExecutionException e) {

                }
            }

            // check if role has permissions granted
            if (clientRole != null && authenticationData != null) {
                return allowFunctionOps(NamespaceName.get(tenant, namespace), clientRole, authenticationData);
            } else {
                return false;
            }
        }
        return true;
    }

    public boolean allowFunctionOps(NamespaceName namespaceName, String role,
                                    AuthenticationDataSource authenticationData) {
        try {
            switch (componentType) {
                case SINK:
                    return worker().getAuthorizationService().allowSinkOpsAsync(
                            namespaceName, role, authenticationData).get(worker().getWorkerConfig().getZooKeeperOperationTimeoutSeconds(), SECONDS);
                case SOURCE:
                    return worker().getAuthorizationService().allowSourceOpsAsync(
                            namespaceName, role, authenticationData).get(worker().getWorkerConfig().getZooKeeperOperationTimeoutSeconds(), SECONDS);
                case FUNCTION:
                default:
                    return worker().getAuthorizationService().allowFunctionOpsAsync(
                            namespaceName, role, authenticationData).get(worker().getWorkerConfig().getZooKeeperOperationTimeoutSeconds(), SECONDS);
            }
        } catch (InterruptedException e) {
            log.warn("Time-out {} sec while checking function authorization on {} ", worker().getWorkerConfig().getZooKeeperOperationTimeoutSeconds(), namespaceName);
            throw new RestException(javax.ws.rs.core.Response.Status.INTERNAL_SERVER_ERROR, e.getMessage());
        } catch (Exception e) {
            log.warn("Admin-client with Role - {} failed to get function permissions for namespace - {}. {}", role, namespaceName,
                    e.getMessage(), e);
            throw new RestException(javax.ws.rs.core.Response.Status.INTERNAL_SERVER_ERROR, e.getMessage());
        }
    }

    void validatePermission(String tenant,
                            String namespace,
                            String clientRole,
                            AuthenticationDataSource clientAuthenticationDataHttps,
                            String componentName) {
        try {
            if (!isAuthorizedRole(tenant, namespace, clientRole, clientAuthenticationDataHttps)) {
                log.warn("{}/{}/{} Client [{}] is not authorized to get {}", tenant, namespace,
                        componentName, clientRole, ComponentTypeUtils.toString(componentType));
                throw new RestException(javax.ws.rs.core.Response.Status.UNAUTHORIZED, "client is not authorize to perform operation");
            }
        } catch (PulsarAdminException e) {
            log.error("{}/{}/{} Failed to authorize [{}]", tenant, namespace, componentName, e);
            throw new RestException(javax.ws.rs.core.Response.Status.INTERNAL_SERVER_ERROR, e.getMessage());
        }
    }

    void validateGetInfoRequestParams(
            String tenant, String namespace, String name, String type) {
        if (tenant == null) {
            throw new RestException(javax.ws.rs.core.Response.Status.BAD_REQUEST, "Tenant is not provided");
        }
        if (namespace == null) {
            throw new RestException(javax.ws.rs.core.Response.Status.BAD_REQUEST, "Namespace is not provided");
        }
        if (name == null) {
            throw new RestException(javax.ws.rs.core.Response.Status.BAD_REQUEST, type + " name is not provided");
        }
    }

    void validateTenantIsExist(String tenant, String namespace, String name, String clientRole) {
        try {
            // Check tenant exists
            worker().getBrokerAdmin().tenants().getTenantInfo(tenant);

        } catch (PulsarAdminException.NotAuthorizedException e) {
            log.error("{}/{}/{} Client [{}] is not authorized to operate {} on tenant", tenant, namespace,
                    name, clientRole, ComponentTypeUtils.toString(componentType));
            throw new RestException(javax.ws.rs.core.Response.Status.UNAUTHORIZED, "client is not authorize to perform operation");
        } catch (PulsarAdminException.NotFoundException e) {
            log.error("{}/{}/{} Tenant {} does not exist", tenant, namespace, name, tenant);
            throw new RestException(javax.ws.rs.core.Response.Status.BAD_REQUEST, "Tenant does not exist");
        } catch (PulsarAdminException e) {
            log.error("{}/{}/{} Issues getting tenant data", tenant, namespace, name, e);
            throw new RestException(javax.ws.rs.core.Response.Status.INTERNAL_SERVER_ERROR, e.getMessage());
        }
    }

    void validateResources(Resources componentResources, Resources minResource, Resources maxResource) {
        if (componentResources != null) {
            if (minResource != null && (componentResources.getCpu() < minResource.getCpu()
                    || componentResources.getRam() < minResource.getRam())) {
                throw new RestException(javax.ws.rs.core.Response.Status.BAD_REQUEST, "Resource is less than minimum requirement");
            }
            if (maxResource != null && (componentResources.getCpu() > maxResource.getCpu()
                    || componentResources.getRam() > maxResource.getRam())) {
                throw new RestException(javax.ws.rs.core.Response.Status.BAD_REQUEST, "Resource is larger than max requirement");
            }
        }
    }
}

