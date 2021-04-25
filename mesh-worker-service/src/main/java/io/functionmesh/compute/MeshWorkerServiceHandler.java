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

import javax.net.ssl.SSLContext;
import javax.servlet.ServletConfig;
import javax.servlet.ServletException;
import javax.servlet.http.HttpServletRequest;
import java.io.File;
import java.net.URI;
import java.nio.charset.StandardCharsets;
import java.security.cert.X509Certificate;
import java.util.concurrent.Executor;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.io.FileUtils;
import org.apache.pulsar.common.util.SecurityUtility;
import org.eclipse.jetty.client.HttpClient;
import org.eclipse.jetty.client.ProtocolHandlers;
import org.eclipse.jetty.client.RedirectProtocolHandler;
import org.eclipse.jetty.client.api.Request;
import org.eclipse.jetty.proxy.ProxyServlet;
import org.eclipse.jetty.util.HttpCookieStore;
import org.eclipse.jetty.util.ssl.SslContextFactory;
import org.eclipse.jetty.util.thread.QueuedThreadPool;

/**
 * Function mesh proxy.
 */
@Slf4j
public class MeshWorkerServiceHandler extends ProxyServlet {

    private static final String FUNCTION_MESH_PATH_PREFIX = "/apis/compute.functionmesh.io/v1alpha1/namespaces";

    private static final String FUNCTION_MESH_KEY = "functionmeshes";

    private static final String KUBERNETES_SERVICE_HOST = "KUBERNETES_SERVICE_HOST";

    private static final String KUBERNETES_SERVICE_PORT = "443";

    private static final String KUBERNETES_CA_CRT_PATH = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt";

    private static final String KUBERNETES_TOKEN_PATH = "/var/run/secrets/kubernetes.io/serviceaccount/token";

    @Override
    protected HttpClient createHttpClient() throws ServletException {
        ServletConfig config = getServletConfig();

        HttpClient httpClient = newHttpClient();
        httpClient.setFollowRedirects(true);
        httpClient.setCookieStore(new HttpCookieStore.Empty());

        Executor executor;
        String value = config.getInitParameter("maxThreads");
        if (value == null || "-".equals(value)) {
            executor = (Executor) getServletContext().getAttribute("org.eclipse.jetty.server.Executor");
            if (executor == null)
                throw new IllegalStateException("No server executor for proxy");
        } else {
            QueuedThreadPool qtp = new QueuedThreadPool(Integer.parseInt(value));
            String servletName = config.getServletName();
            int dot = servletName.lastIndexOf('.');
            if (dot >= 0)
                servletName = servletName.substring(dot + 1);
            qtp.setName(servletName);
            executor = qtp;
        }

        httpClient.setExecutor(executor);

        value = config.getInitParameter("maxConnections");
        if (value == null)
            value = "256";
        httpClient.setMaxConnectionsPerDestination(Integer.parseInt(value));

        value = config.getInitParameter("idleTimeout");
        if (value == null)
            value = "30000";
        httpClient.setIdleTimeout(Long.parseLong(value));

        value = config.getInitParameter("requestBufferSize");
        if (value != null)
            httpClient.setRequestBufferSize(Integer.parseInt(value));

        value = config.getInitParameter("responseBufferSize");
        if (value != null)
            httpClient.setResponseBufferSize(Integer.parseInt(value));

        try {
            httpClient.start();

            // Content must not be decoded, otherwise the client gets confused.
            httpClient.getContentDecoderFactories().clear();

            // Pass traffic to the client, only intercept what's necessary.
            ProtocolHandlers protocolHandlers = httpClient.getProtocolHandlers();
            protocolHandlers.clear();
            protocolHandlers.put(new RedirectProtocolHandler(httpClient));

            return httpClient;
        } catch (Exception x) {
            throw new ServletException(x);
        }

    }

    @Override
    protected HttpClient newHttpClient() {

        try {
            X509Certificate[] trustCertificates = SecurityUtility
                    .loadCertificatesFromPemFile(KUBERNETES_CA_CRT_PATH);

            SSLContext sslCtx = SecurityUtility.createSslContext(
                    false,
                    trustCertificates
            );


            SslContextFactory contextFactory = new SslContextFactory.Client(true);
            contextFactory.setSslContext(sslCtx);

            return new HttpClient(contextFactory);
        } catch (Exception e) {
            log.error("Init http client failed for proxy" + e.getMessage());
        }

        // return an unauthenticated client, every request will fail.
        return new HttpClient();
    }

    @Override
    protected String rewriteTarget(HttpServletRequest request) {
        StringBuilder url = new StringBuilder();
        boolean isFunctionMeshRestRequest = false;
        String requestUri = request.getRequestURI();
        if (requestUri.startsWith(FUNCTION_MESH_PATH_PREFIX)) {
            String [] requestUriPath = requestUri.split("/");
            if (requestUriPath.length >= 7 && requestUriPath[6].equals(FUNCTION_MESH_KEY)) {
                isFunctionMeshRestRequest = true;
            }
        }
        if (isFunctionMeshRestRequest) {
            String controllerHost = this.getEnvironment(KUBERNETES_SERVICE_HOST);
            url.append("https://").append(controllerHost).append(":").append(KUBERNETES_SERVICE_PORT).append(requestUri);
            String query = request.getQueryString();
            if (query != null) {
                url.append("?").append(query);
            }

            URI rewrittenUrl = URI.create(url.toString()).normalize();

            if (!validateDestination(rewrittenUrl.getHost(), rewrittenUrl.getPort())) {
                return null;
            }
            return rewrittenUrl.toString();
        }
        return null;
    }

    protected String getEnvironment(String key) {
        return System.getenv(key);
    }

    @Override
    protected void addProxyHeaders(HttpServletRequest clientRequest, Request proxyRequest) {
        super.addProxyHeaders(clientRequest, proxyRequest);
        try {
            File file = new File(KUBERNETES_TOKEN_PATH);
            String cloudControllerAuthToken = FileUtils.readFileToString(file, StandardCharsets.UTF_8);
            proxyRequest.header("Authorization", "Bearer " + cloudControllerAuthToken);
        } catch (java.io.IOException e) {
            log.error("Init cloud controller ca cert failed, message: {}", e.getMessage());
        }
    }
}
