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
package io.functionmesh.compute.auth;

import io.kubernetes.client.openapi.ApiException;
import io.kubernetes.client.openapi.apis.CoreV1Api;
import io.kubernetes.client.openapi.models.V1ConfigMap;
import org.apache.commons.lang.StringUtils;
import org.apache.pulsar.broker.authentication.AuthenticationDataSource;
import org.apache.pulsar.functions.auth.FunctionAuthData;
import org.apache.pulsar.functions.proto.Function;
import org.junit.Assert;
import org.junit.Test;

import java.util.Optional;

import static org.mockito.Matchers.any;
import static org.mockito.Matchers.anyString;
import static org.powermock.api.mockito.PowerMockito.doReturn;
import static org.powermock.api.mockito.PowerMockito.mock;

public class KubernetesFunctionMeshAuthTokenProviderTests {

	@Test
	public void testCacheAuthData() throws ApiException {
		CoreV1Api coreV1Api = mock(CoreV1Api.class);
		doReturn(new V1ConfigMap()).when(coreV1Api).createNamespacedConfigMap(anyString(), any(), anyString(), anyString(), anyString());
		KubernetesFunctionMeshAuthTokenProvider kubernetesFunctionMeshAuthTokenProvider = new KubernetesFunctionMeshAuthTokenProvider();
		kubernetesFunctionMeshAuthTokenProvider.initialize(coreV1Api,  null, (fd) -> "default");
		Function.FunctionDetails funcDetails = Function.FunctionDetails.newBuilder().setTenant("test-tenant").setNamespace("test-ns").setName("test-func").build();
		Optional<FunctionAuthData> functionAuthData = kubernetesFunctionMeshAuthTokenProvider.cacheAuthData(funcDetails, new AuthenticationDataSource() {
			@Override
			public boolean hasDataFromCommand() {
				return true;
			}

			@Override
			public String getCommandData() {
				return "test-token";
			}
		});
		System.out.println(new String(functionAuthData.get().getData()));
		Assert.assertTrue(StringUtils.isNotBlank(new String(functionAuthData.get().getData())));
	}

	@Test
	public void testUpdateAuthData() throws Exception {
		CoreV1Api coreV1Api = mock(CoreV1Api.class);
		KubernetesFunctionMeshAuthTokenProvider kubernetesFunctionMeshAuthTokenProvider = new KubernetesFunctionMeshAuthTokenProvider();
		kubernetesFunctionMeshAuthTokenProvider.initialize(coreV1Api, null, (fd) -> "default");
		// test when existingFunctionAuthData is empty
		Optional<FunctionAuthData> existingFunctionAuthData = Optional.empty();
		Function.FunctionDetails funcDetails = Function.FunctionDetails.newBuilder().setTenant("test-tenant").setNamespace("test-ns").setName("test-func").build();
		Optional<FunctionAuthData> functionAuthData = kubernetesFunctionMeshAuthTokenProvider.updateAuthData(funcDetails, existingFunctionAuthData, new AuthenticationDataSource() {
			@Override
			public boolean hasDataFromCommand() {
				return true;
			}

			@Override
			public String getCommandData() {
				return "test-token";
			}
		});


		Assert.assertTrue(functionAuthData.isPresent());
		Assert.assertTrue(StringUtils.isNotBlank(new String(functionAuthData.get().getData())));

		// test when existingFunctionAuthData is NOT empty
		existingFunctionAuthData = Optional.of(FunctionAuthData.builder().data("pf-configmap-z7mxx".getBytes()).provider(null).build());
		functionAuthData = kubernetesFunctionMeshAuthTokenProvider.updateAuthData(funcDetails, existingFunctionAuthData, new AuthenticationDataSource() {
			@Override
			public boolean hasDataFromCommand() {
				return true;
			}

			@Override
			public String getCommandData() {
				return "test-token";
			}
		});


		Assert.assertTrue(functionAuthData.isPresent());
		Assert.assertEquals(new String(functionAuthData.get().getData()), "pf-configmap-z7mxx");
	}
}
