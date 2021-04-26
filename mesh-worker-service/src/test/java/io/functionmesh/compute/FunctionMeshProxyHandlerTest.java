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
package io.functionmesh.compute;

import org.junit.Assert;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.powermock.api.mockito.PowerMockito;
import org.powermock.core.classloader.annotations.PowerMockIgnore;
import org.powermock.modules.junit4.PowerMockRunner;

import javax.servlet.http.HttpServletRequest;

import static org.powermock.api.mockito.PowerMockito.spy;

@RunWith(PowerMockRunner.class)
@PowerMockIgnore({"javax.management.*"})
public class FunctionMeshProxyHandlerTest {

    private MeshWorkerServiceHandler meshWorkerServiceHandler = spy(new MeshWorkerServiceHandler());

    @Test
    public void rewriteTargetTest() {

        String path = "/api/pods";
        HttpServletRequest httpServletRequest = PowerMockito.mock(HttpServletRequest.class);
        PowerMockito.when(httpServletRequest.getRequestURI()).thenReturn(path);
        Assert.assertNull(meshWorkerServiceHandler.rewriteTarget(httpServletRequest));

        PowerMockito.when(meshWorkerServiceHandler
                .getEnvironment("KUBERNETES_SERVICE_HOST")).thenReturn("localhost");
        path = "/apis/compute.functionmesh.io/v1alpha1/namespaces/default/functionmeshes";
        PowerMockito.when(httpServletRequest.getRequestURI()).thenReturn(path);
        PowerMockito.when(httpServletRequest.getQueryString()).thenReturn("limit=500");
        String rewriteTarget = meshWorkerServiceHandler.rewriteTarget(httpServletRequest);
        String expectedValue = "https://localhost:443" +
                "/apis/compute.functionmesh.io/v1alpha1/namespaces/default/functionmeshes?limit=500";
        Assert.assertEquals(rewriteTarget, expectedValue);
    }
}
