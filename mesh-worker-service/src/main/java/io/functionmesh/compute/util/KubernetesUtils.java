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

import io.kubernetes.client.openapi.ApiException;
import io.kubernetes.client.openapi.apis.CoreV1Api;
import io.kubernetes.client.openapi.models.V1ConfigMap;
import io.kubernetes.client.openapi.models.V1ObjectMeta;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.io.FileUtils;
import org.apache.commons.lang3.RandomStringUtils;
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

	public static String getConfigMapName(String type, String tenant, String namespace, String name) {
		return "function-mesh-configmap-" + type + "-" + tenant + "-" + namespace + "-" + name ;
	}

	private static Map<String, String> buildConfigMap(WorkerConfig workerConfig) {
		Map<String, String> valueMap = new HashMap<>();
		valueMap.put(CLIENT_AUTHENTICATION_PLUGIN_CLAIM, workerConfig.getBrokerClientAuthenticationPlugin());
		valueMap.put(CLIENT_AUTHENTICATION_PARAMETERS_CLAIM, workerConfig.getBrokerClientAuthenticationParameters());
		valueMap.put(TLS_TRUST_CERTS_FILE_PATH_CLAIM, workerConfig.getTlsCertificateFilePath());
		valueMap.put(USE_TLS_CLAIM, String.valueOf(workerConfig.getTlsEnabled()));
		valueMap.put(TLS_ALLOW_INSECURE_CONNECTION_CLAIM, String.valueOf(workerConfig.isTlsAllowInsecureConnection()));
		valueMap.put(TLS_HOSTNAME_VERIFICATION_ENABLE_CLAIM, String.valueOf(workerConfig.isTlsEnableHostnameVerification()));
		return valueMap;
	}

	public static String createConfigMap(
			String type,
			String tenant,
			String namespace,
			String name,
			WorkerConfig workerConfig,
			CoreV1Api coreV1Api,
			KubernetesRuntimeFactoryConfig factoryConfig) throws ApiException, InterruptedException {

		String configMapName = getConfigMapName(type, tenant, namespace, name);
		StringBuilder sb = new StringBuilder();
		Actions.Action createAuthConfigMap = Actions.Action.builder()
				.actionName(String.format("Creating authentication config map for function %s/%s/%s", tenant, namespace, name))
				.numRetries(NUM_RETRIES)
				.sleepBetweenInvocationsMs(SLEEP_BETWEEN_RETRIES_MS)
				.supplier(() -> {
					String id =  RandomStringUtils.random(5, true, true).toLowerCase();
					V1ConfigMap v1ConfigMap = new V1ConfigMap()
							.metadata(new V1ObjectMeta().name(configMapName))
							.data(buildConfigMap(workerConfig));
					try {
						coreV1Api.createNamespacedConfigMap(KubernetesUtils.getNamespace(factoryConfig), v1ConfigMap, null, null, null);
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

}
