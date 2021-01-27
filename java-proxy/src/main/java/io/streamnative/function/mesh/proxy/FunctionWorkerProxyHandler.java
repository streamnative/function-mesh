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
package io.streamnative.function.mesh.proxy;

import javax.servlet.ServletConfig;
import javax.servlet.ServletException;
import javax.servlet.http.HttpServletRequest;
import java.net.URI;
import java.util.Arrays;
import java.util.HashSet;
import java.util.Set;
import java.util.concurrent.Executor;
import lombok.extern.slf4j.Slf4j;
import org.eclipse.jetty.client.HttpClient;
import org.eclipse.jetty.client.ProtocolHandlers;
import org.eclipse.jetty.client.RedirectProtocolHandler;
import org.eclipse.jetty.proxy.ProxyServlet;
import org.eclipse.jetty.util.HttpCookieStore;
import org.eclipse.jetty.util.thread.QueuedThreadPool;

/**
 * Function worker proxy.
 */
@Slf4j
public class FunctionWorkerProxyHandler extends ProxyServlet {

    private static final String FUNCTION_PODS_SERVICE_CLAIM = "x-api-name";

    private static final String FUNCTION_SERVICE_METRICS_PORT = "9094";

    private static final Set<String> workerRoutes = new HashSet<>(Arrays.asList(
            "/admin/v2/worker",
            "/admin/v2/worker-stats",
            "/admin/worker-stats",
            "/admin/worker"
    ));

    @Override
    protected HttpClient createHttpClient() throws ServletException {
        ServletConfig config = getServletConfig();

        HttpClient httpClient = new HttpClient();
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
    protected String rewriteTarget(HttpServletRequest request) {
        StringBuilder url = new StringBuilder();
        boolean isWorkerRestRequest = false;
        String requestUri = request.getRequestURI();
        for (String routePrefix: workerRoutes) {
            if (requestUri.startsWith(routePrefix)) {
                isWorkerRestRequest = true;
                break;
            }
        }
        if (isWorkerRestRequest) {
            String serviceName = request.getHeader(FUNCTION_PODS_SERVICE_CLAIM);
            url.append("http://").append(serviceName).append(":").append(FUNCTION_SERVICE_METRICS_PORT);
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

}
