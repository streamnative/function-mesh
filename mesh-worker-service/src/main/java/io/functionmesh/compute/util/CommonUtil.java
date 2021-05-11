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

import io.kubernetes.client.openapi.models.V1ObjectMeta;
import org.apache.pulsar.common.functions.FunctionConfig;

import java.util.Map;
import java.util.stream.Collectors;

public class CommonUtil {
    private static final String CLUSTER_NAME_ENV = "clusterName";

    public static String getCurrentClusterName() {
        return System.getenv(CLUSTER_NAME_ENV);
    }

    public static String getDefaultPulsarConfig() {
        return toValidResourceName(String.format("%s-pulsar-config-map", System.getenv(CLUSTER_NAME_ENV)));
    }

    public static String getPulsarClusterConfigMapName(String cluster) {
        return toValidResourceName(String.format("%s-pulsar-config-map", cluster)); // Need to manage the configMap for each Pulsar Cluster
    }

    public static String getPulsarClusterAuthConfigMapName(String cluster) {
        return toValidResourceName(String.format("%s-auth-config-map", cluster)); // Need to manage the configMap for each Pulsar Cluster
    }

    private static String toValidResourceName(String ori) {
        return ori.toLowerCase().replaceAll("[^a-z0-9-\\.]", "-");
    }

    public static V1ObjectMeta makeV1ObjectMeta(String name, String namespace) {
        V1ObjectMeta v1ObjectMeta = new V1ObjectMeta();
        v1ObjectMeta.setName(name);
        v1ObjectMeta.setNamespace(namespace);

        return v1ObjectMeta;
    }

    public static FunctionConfig.ProcessingGuarantees convertProcessingGuarantee(String processingGuarantees) {
        switch (processingGuarantees) {
            case "atleast_once":
                return FunctionConfig.ProcessingGuarantees.ATLEAST_ONCE;
            case "atmost_once":
                return FunctionConfig.ProcessingGuarantees.ATMOST_ONCE;
            case "effectively_once":
                return FunctionConfig.ProcessingGuarantees.EFFECTIVELY_ONCE;
        }
        return null;
    }

    public static Map<String, String> transformedMapValueToString(Map<String, Object> map) {
        if (map == null) {
            return null;
        }
        return map.entrySet().stream().collect(Collectors.toMap(Map.Entry::getKey, e -> String.valueOf(e.getValue())));
    }

    public static Map<String, Object> transformedMapValueToObject(Map<String, String> map) {
        if (map == null) {
            return null;
        }
        return map.entrySet().stream().collect(Collectors.toMap(Map.Entry::getKey, Map.Entry::getValue));
    }
}
