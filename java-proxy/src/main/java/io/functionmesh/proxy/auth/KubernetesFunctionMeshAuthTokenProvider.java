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
package io.functionmesh.proxy.auth;

import com.google.common.annotations.VisibleForTesting;
import io.functionmesh.proxy.util.KubernetesUtils;
import io.kubernetes.client.openapi.ApiException;
import io.kubernetes.client.openapi.apis.CoreV1Api;
import io.kubernetes.client.openapi.models.V1ConfigMap;
import io.kubernetes.client.openapi.models.V1ObjectMeta;
import io.kubernetes.client.openapi.models.V1StatefulSet;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.lang3.RandomStringUtils;
import org.apache.pulsar.broker.authentication.AuthenticationDataSource;
import org.apache.pulsar.common.functions.AuthenticationConfig;
import org.apache.pulsar.functions.auth.FunctionAuthData;
import org.apache.pulsar.functions.auth.KubernetesFunctionAuthProvider;
import org.apache.pulsar.functions.proto.Function;
import org.apache.pulsar.functions.utils.Actions;
import org.apache.pulsar.functions.utils.FunctionCommon;

import javax.naming.AuthenticationException;
import java.util.HashMap;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.atomic.AtomicBoolean;

import static java.net.HttpURLConnection.HTTP_CONFLICT;
import static java.net.HttpURLConnection.HTTP_NOT_FOUND;
import static org.apache.commons.lang3.StringUtils.isBlank;
import static org.apache.pulsar.broker.authentication.AuthenticationProviderToken.getToken;

@Slf4j
public class KubernetesFunctionMeshAuthTokenProvider implements KubernetesFunctionAuthProvider {

	private static final int NUM_RETRIES = 5;
	private static final long SLEEP_BETWEEN_RETRIES_MS = 500;
	private static final String CLIENT_AUTHENTICATION_PLUGIN_CLAIM = "clientAuthenticationPlugin";
	private static final String CLIENT_AUTHENTICATION_PLUGIN_NAME = "org.apache.pulsar.client.impl.auth.AuthenticationToken";
	private static final String CLIENT_AUTHENTICATION_PARAMETERS_CLAIM = "clientAuthenticationParameters";
	private static final String TLS_TRUST_CERTS_FILE_PATH_CLAIM = "tlsTrustCertsFilePath";
	private static final String USE_TLS_CLAIM = "useTls";
	private static final String TLS_ALLOW_INSECURE_CONNECTION_CLAIM = "tlsAllowInsecureConnection";
	private static final String TLS_HOSTNAME_VERIFICATION_ENABLE_CLAIM = "tlsHostnameVerificationEnable";

	private CoreV1Api coreClient;

	@Override
	public void initialize(CoreV1Api coreClient) {
		this.coreClient = coreClient;
	}

	@Override
	public void setCaBytes(byte[] caBytes) {
		// There is no need to implement this function in the function mesh
	}

	@Override
	public void setNamespaceProviderFunc(java.util.function.Function<Function.FunctionDetails, String> getNamespaceFromDetails) {
		// There is no need to implement this function in the function mesh
		// Dynamically get the k8s namespace
	}

	@Override
	public void configureAuthDataStatefulSet(V1StatefulSet statefulSet, Optional<FunctionAuthData> functionAuthData) {
		// There is no need to implement this function in the function mesh
	}

	@Override
	public void configureAuthenticationConfig(AuthenticationConfig authConfig, Optional<FunctionAuthData> functionAuthData) {
		// There is no need to implement this function in the function mesh
	}

	@Override
	public Optional<FunctionAuthData> cacheAuthData(Function.FunctionDetails funcDetails,
													AuthenticationDataSource authenticationDataSource) {
		String id = null;
		String tenant = funcDetails.getTenant();
		String namespace = funcDetails.getNamespace();
		String name = funcDetails.getName();
		try {
			String token = getToken(authenticationDataSource);
			if (token != null) {
				id = createConfigMap(token, funcDetails);
			}
		} catch (Exception e) {
			log.warn("Failed to get token for function {}", FunctionCommon.getFullyQualifiedName(tenant, namespace, name), e);
			// ignore exception and continue since anonymous user might to used
		}
		if (id != null) {
			return Optional.of(FunctionAuthData.builder().data(id.getBytes()).build());
		}
		return Optional.empty();
	}

	@Override
	public void cleanUpAuthData(Function.FunctionDetails funcDetails, Optional<FunctionAuthData> functionAuthData) throws Exception {
		if (!functionAuthData.isPresent()) {
			return;
		}

		String fqfn = FunctionCommon.getFullyQualifiedName(funcDetails.getTenant(), funcDetails.getNamespace(), funcDetails.getName());

		String configMapName = getConfigMapName(funcDetails.getTenant(), funcDetails.getNamespace(), funcDetails.getName());
		String kubeNamespace = KubernetesUtils.getNamespace();

		Actions.Action deleteConfigMap = Actions.Action.builder()
				.actionName(String.format("Deleting config map for function %s", fqfn))
				.numRetries(NUM_RETRIES)
				.sleepBetweenInvocationsMs(SLEEP_BETWEEN_RETRIES_MS)
				.supplier(() -> {
					try {
						// make sure configMapName is not null or empty string.
						// If deleteNamespacedConfigMap is called and configmap name is null or empty string
						// it will delete all the configmap in the namespace
						coreClient.deleteNamespacedConfigMap(configMapName, kubeNamespace, null, null,
								0, null, "Foreground", null);
					} catch (ApiException e) {
						// if already deleted
						if (e.getCode() == HTTP_NOT_FOUND) {
							log.warn("ConfigMap for function {} does not exist", fqfn);
							return Actions.ActionResult.builder().success(true).build();
						}

						String errorMsg = e.getResponseBody() != null ? e.getResponseBody() : e.getMessage();
						return Actions.ActionResult.builder()
								.success(false)
								.errorMsg(errorMsg)
								.build();
					}
					return Actions.ActionResult.builder().success(true).build();
				})
				.build();

		Actions.Action waitForConfigMapDeletion = Actions.Action.builder()
				.actionName(String.format("Waiting for configmap for function %s to complete deletion", fqfn))
				.actionName(String.format("Waiting for configmap for function %s to complete deletion", fqfn))
				.numRetries(NUM_RETRIES)
				.sleepBetweenInvocationsMs(SLEEP_BETWEEN_RETRIES_MS)
				.supplier(() -> {
					try {
						coreClient.readNamespacedConfigMap(configMapName, kubeNamespace,
								null, null, null);

					} catch (ApiException e) {
						// statefulset is gone
						if (e.getCode() == HTTP_NOT_FOUND) {
							return Actions.ActionResult.builder().success(true).build();
						}
						String errorMsg = e.getResponseBody() != null ? e.getResponseBody() : e.getMessage();
						return Actions.ActionResult.builder()
								.success(false)
								.errorMsg(errorMsg)
								.build();
					}
					return Actions.ActionResult.builder()
							.success(false)
							.build();
				})
				.build();

		AtomicBoolean success = new AtomicBoolean(false);
		Actions.newBuilder()
				.addAction(deleteConfigMap.toBuilder()
						.continueOn(true)
						.build())
				.addAction(waitForConfigMapDeletion.toBuilder()
						.continueOn(false)
						.onSuccess(ignore -> success.set(true))
						.build())
				.addAction(deleteConfigMap.toBuilder()
						.continueOn(true)
						.build())
				.addAction(waitForConfigMapDeletion.toBuilder()
						.onSuccess(ignore -> success.set(true))
						.build())
				.run();

		if (!success.get()) {
			throw new RuntimeException(String.format("Failed to delete configmap for function %s", fqfn));
		}
	}

	@Override
	public Optional<FunctionAuthData> updateAuthData(Function.FunctionDetails funcDetails,
													 Optional<FunctionAuthData> existingFunctionAuthData,
													 AuthenticationDataSource authenticationDataSource) throws Exception {
		String fqfn = FunctionCommon.getFullyQualifiedName(
				funcDetails.getTenant(), funcDetails.getNamespace(), funcDetails.getName());
		String token;
		try {
			token = getToken(authenticationDataSource);
			if (token != null) {
				String configMapName = getConfigMapName(funcDetails.getTenant(), funcDetails.getNamespace(), funcDetails.getName());
				updateConfigMap(token, funcDetails, configMapName);
				return Optional.of(FunctionAuthData.builder().data(configMapName.getBytes()).build());
			}
		} catch (AuthenticationException e) {
			log.error("Update auth {} data failed", fqfn);
		}

		return Optional.empty();

	}

	@VisibleForTesting
	private Map<String, String> buildConfigMap(String token) {
		Map<String, String> valueMap = new HashMap<>();
		valueMap.put(CLIENT_AUTHENTICATION_PLUGIN_CLAIM, CLIENT_AUTHENTICATION_PLUGIN_NAME);
		valueMap.put(CLIENT_AUTHENTICATION_PARAMETERS_CLAIM, token);
		valueMap.put(TLS_TRUST_CERTS_FILE_PATH_CLAIM, "");
		valueMap.put(USE_TLS_CLAIM, "false");
		valueMap.put(TLS_ALLOW_INSECURE_CONNECTION_CLAIM, "false");
		valueMap.put(TLS_HOSTNAME_VERIFICATION_ENABLE_CLAIM, "true");
		return valueMap;
	}

	private String createConfigMap(String token, Function.FunctionDetails functionDetails) throws ApiException, InterruptedException {
		String tenant = functionDetails.getTenant();
		String namespace = functionDetails.getNamespace();
		String name = functionDetails.getName();

		String configMapName = getConfigMapName(tenant, namespace, name);
		StringBuilder sb = new StringBuilder();
		Actions.Action createAuthConfigMap = Actions.Action.builder()
				.actionName(String.format("Creating authentication config map for function %s/%s/%s", tenant, namespace, name))
				.numRetries(NUM_RETRIES)
				.sleepBetweenInvocationsMs(SLEEP_BETWEEN_RETRIES_MS)
				.supplier(() -> {
					String id =  RandomStringUtils.random(5, true, true).toLowerCase();
					V1ConfigMap v1ConfigMap = new V1ConfigMap()
							.metadata(new V1ObjectMeta().name(configMapName))
							.data(buildConfigMap(token));
					try {
						coreClient.createNamespacedConfigMap(KubernetesUtils.getNamespace(), v1ConfigMap, null, null, null);
					} catch (ApiException e) {
						// already exists
						if (e.getCode() == HTTP_CONFLICT) {
							return Actions.ActionResult.builder()
									.errorMsg(String.format("ConfigMap %s already present", id))
									.success(false)
									.build();
						}

						String errorMsg = e.getResponseBody() != null ? e.getResponseBody() : e.getMessage();
						return Actions.ActionResult.builder()
								.success(false)
								.errorMsg(errorMsg)
								.build();
					}

					sb.append(id.toCharArray());
					return Actions.ActionResult.builder().success(true).build();
				})
				.build();

		AtomicBoolean success = new AtomicBoolean(false);
		Actions.newBuilder()
				.addAction(createAuthConfigMap.toBuilder()
						.onSuccess(ignore -> success.set(true))
						.build())
				.run();

		if (!success.get()) {
			throw new RuntimeException(String.format("Failed to create authentication configmap for function %s/%s/%s", tenant, namespace, name));
		}

		return sb.toString();
	}

	private void updateConfigMap(String token, Function.FunctionDetails functionDetails, String configMapName) throws ApiException, InterruptedException {
		String tenant = functionDetails.getTenant();
		String namespace = functionDetails.getNamespace();
		String name = functionDetails.getName();

		Actions.Action createAuthConfigMap = Actions.Action.builder()
				.actionName(String.format("Upsert authentication config map for function %s/%s/%s", tenant, namespace, name))
				.numRetries(NUM_RETRIES)
				.sleepBetweenInvocationsMs(SLEEP_BETWEEN_RETRIES_MS)
				.supplier(() -> {
					V1ConfigMap v1ConfigMap = new V1ConfigMap()
							.metadata(new V1ObjectMeta().name(configMapName))
							.data(buildConfigMap(token));
					try {
						coreClient.createNamespacedConfigMap(KubernetesUtils.getNamespace(), v1ConfigMap, null, null, null);
					} catch (ApiException e) {
						if (e.getCode() == HTTP_CONFLICT) {
							try {
								coreClient.replaceNamespacedConfigMap(configMapName, KubernetesUtils.getNamespace(), v1ConfigMap, null, null, null);
								return Actions.ActionResult.builder().success(true).build();

							} catch (ApiException e1) {
								String errorMsg = e.getResponseBody() != null ? e.getResponseBody() : e.getMessage();
								return Actions.ActionResult.builder()
										.success(false)
										.errorMsg(errorMsg)
										.build();
							}
						}

						String errorMsg = e.getResponseBody() != null ? e.getResponseBody() : e.getMessage();
						return Actions.ActionResult.builder()
								.success(false)
								.errorMsg(errorMsg)
								.build();
					}

					return Actions.ActionResult.builder().success(true).build();
				})
				.build();

		AtomicBoolean success = new AtomicBoolean(false);
		Actions.newBuilder()
				.addAction(createAuthConfigMap.toBuilder()
						.onSuccess(ignore -> success.set(true))
						.build())
				.run();

		if (!success.get()) {
			throw new RuntimeException(String.format("Failed to upsert authentication configmap for function %s/%s/%s", tenant, namespace, name));
		}
	}

	private String getConfigMapName(String tenant, String namespace, String name) {
		return "function-mesh-configmap-auth-" + tenant + "-" + namespace + "-" + name ;
	}
}
