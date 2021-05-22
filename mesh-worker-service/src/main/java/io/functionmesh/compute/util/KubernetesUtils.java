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

import com.google.common.collect.Maps;
import io.kubernetes.client.openapi.ApiException;
import io.kubernetes.client.openapi.apis.CoreV1Api;
import io.kubernetes.client.openapi.models.V1ObjectMeta;
import io.kubernetes.client.openapi.models.V1Secret;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.codec.digest.DigestUtils;
import org.apache.commons.io.FileUtils;
import org.apache.pulsar.functions.runtime.kubernetes.KubernetesRuntimeFactoryConfig;
import org.apache.pulsar.functions.utils.Actions;
import org.apache.pulsar.functions.worker.WorkerConfig;
import java.io.File;
import java.nio.charset.StandardCharsets;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.atomic.AtomicBoolean;
import static java.net.HttpURLConnection.HTTP_CONFLICT;

@Slf4j
public class KubernetesUtils {

	private static final String KUBERNETES_NAMESPACE_PATH = "/var/run/secrets/kubernetes.io/serviceaccount/namespace";

	private static final int NUM_RETRIES = 5;
	private static final long SLEEP_BETWEEN_RETRIES_MS = 500;
	private static final String CLIENT_AUTHENTICATION_PLUGIN_CLAIM = "clientAuthenticationPlugin";
	private static final String CLIENT_AUTHENTICATION_PLUGIN_NAME = "org.apache.pulsar.client.impl.auth.AuthenticationToken";
	private static final String CLIENT_AUTHENTICATION_PARAMETERS_CLAIM = "clientAuthenticationParameters";
	private static final String TLS_TRUST_CERTS_FILE_PATH_CLAIM = "tlsTrustCertsFilePath";
	private static final String USE_TLS_CLAIM = "useTls";
	private static final String TLS_ALLOW_INSECURE_CONNECTION_CLAIM = "tlsAllowInsecureConnection";
	private static final String TLS_HOSTNAME_VERIFICATION_ENABLE_CLAIM = "tlsHostnameVerificationEnable";

	public static String getNamespace() {
		String namespace = null;
		try {
			File file = new File(KUBERNETES_NAMESPACE_PATH);
			namespace = FileUtils.readFileToString(file, StandardCharsets.UTF_8);
		} catch (java.io.IOException e) {
			log.error("Get namespace from kubernetes path {}, message: {}", KUBERNETES_NAMESPACE_PATH, e.getMessage());
		}
		// Use the default namespace
		if (namespace == null) {
			return "default";
		}
		return namespace;
	}

	public static String getNamespace(KubernetesRuntimeFactoryConfig kubernetesRuntimeFactoryConfig) {
		if (kubernetesRuntimeFactoryConfig == null) {
			return KubernetesUtils.getNamespace();
		}
		String namespace = kubernetesRuntimeFactoryConfig.getJobNamespace();
		if (namespace == null) {
			return KubernetesUtils.getNamespace();
		}
		return namespace;
	}

	public static String getSecretName(String cluster, String tenant, String namespace, String name) {
		return cluster + "-" + tenant + "-" + namespace + "-" + name;
	}

	private static Map<String, byte[]> buildAuthConfigMap(WorkerConfig workerConfig) {
		Map<String, byte[]> valueMap = new HashMap<>();
		valueMap.put(CLIENT_AUTHENTICATION_PLUGIN_CLAIM, workerConfig.getBrokerClientAuthenticationPlugin().getBytes());
		valueMap.put(CLIENT_AUTHENTICATION_PARAMETERS_CLAIM, workerConfig.getBrokerClientAuthenticationParameters().getBytes());
		return valueMap;
	}

	private static Map<String, byte[]> buildTlsConfigMap(WorkerConfig workerConfig) {
		Map<String, byte[]> valueMap = new HashMap<>();
		valueMap.put(TLS_TRUST_CERTS_FILE_PATH_CLAIM, workerConfig.getTlsCertificateFilePath().getBytes());
		valueMap.put(USE_TLS_CLAIM, String.valueOf(workerConfig.getTlsEnabled()).getBytes());
		valueMap.put(TLS_ALLOW_INSECURE_CONNECTION_CLAIM, String.valueOf(workerConfig.isTlsAllowInsecureConnection()).getBytes());
		valueMap.put(TLS_HOSTNAME_VERIFICATION_ENABLE_CLAIM, String.valueOf(workerConfig.isTlsEnableHostnameVerification()).getBytes());
		return valueMap;
	}

	public static String getUniqueSecretName(String component, String type, String id) {
		return component + "-" + type + "-" + id;
	}

	public static String upsertSecret(
			String component,
			String type,
			String cluster,
			String tenant,
			String namespace,
			String name,
			WorkerConfig workerConfig,
			CoreV1Api coreV1Api,
			KubernetesRuntimeFactoryConfig factoryConfig) throws ApiException, InterruptedException {

		String combinationName = getSecretName(type, tenant, namespace, name);
		String hashcode = DigestUtils.sha256Hex(combinationName);
		String secretName = getUniqueSecretName(component, type, hashcode);
		Map<String, byte[]> data = Maps.newHashMap();
		if ("auth".equals(type)) {
			data = buildAuthConfigMap(workerConfig);
		} else if ("tls".equals(type)) {
			data = buildTlsConfigMap(workerConfig);
		} else {
			throw new RuntimeException(String.format("Failed to create secret type for %s %s/%s/%s",
					type, tenant, namespace, name));
		}
		Map<String, byte[]> finalData = data;
		Actions.Action createAuthSecret = Actions.Action.builder()
				.actionName(String.format("Creating secret for %s %s-%s/%s/%s",
						type, cluster, tenant, namespace, name))
				.numRetries(NUM_RETRIES)
				.sleepBetweenInvocationsMs(SLEEP_BETWEEN_RETRIES_MS)
				.supplier(() -> {
					V1Secret v1Secret = new V1Secret()
							.metadata(new V1ObjectMeta().name(secretName))
							.data(finalData);
					try {
						coreV1Api.createNamespacedSecret(
								KubernetesUtils.getNamespace(factoryConfig),
								v1Secret, null, null, null);
					} catch (ApiException e) {
						// already exists
						if (e.getCode() == HTTP_CONFLICT) {
							try {
								coreV1Api.replaceNamespacedSecret(
										secretName,
										KubernetesUtils.getNamespace(factoryConfig),
										v1Secret, null, null, null);
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
				.addAction(createAuthSecret.toBuilder()
						.onSuccess(ignore -> success.set(true))
						.build())
				.run();

		if (!success.get()) {
			throw new RuntimeException(String.format("Failed to create secret for %s %s-%s/%s/%s",
					type, cluster, tenant, namespace, name));
		}

		return secretName;
	}

}
