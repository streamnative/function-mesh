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

import io.functionmesh.compute.MeshWorkerService;
import io.kubernetes.client.openapi.models.V1ObjectMeta;
import io.kubernetes.client.openapi.models.V1OwnerReference;
import java.util.Collections;
import org.apache.commons.codec.digest.DigestUtils;
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
        return toValidResourceName(String.format("%s-function-mesh-config", cluster)); // Need to manage the configMap for each Pulsar Cluster
    }

    public static String getPulsarClusterAuthConfigMapName(String cluster) {
        return toValidResourceName(String.format("%s-auth-config-map", cluster)); // Need to manage the configMap for each Pulsar Cluster
    }

    private static String toValidResourceName(String ori) {
        return ori.toLowerCase().replaceAll("[^a-z0-9-\\.]", "-");
    }

    public static V1OwnerReference getOwnerReferenceFromCustomConfigs(Map<String, Object> customConfigs) {
        if (customConfigs == null) {
            return null;
        }
        Map<String, Object> ownerRef = (Map<String, Object>) customConfigs.get("ownerReference");
        if (ownerRef == null) {
            return null;
        }
        return new V1OwnerReference()
                .apiVersion(String.valueOf(ownerRef.get("apiVersion")))
                .kind(String.valueOf(ownerRef.get("kind")))
                .name(String.valueOf(ownerRef.get("name")))
                .uid(String.valueOf(ownerRef.get("uid")));
    }

    public static V1ObjectMeta makeV1ObjectMeta(String name, String k8sNamespace, String pulsarNamespace, String tenant,
                                                String cluster, V1OwnerReference ownerReference) {
        V1ObjectMeta v1ObjectMeta = new V1ObjectMeta();
        v1ObjectMeta.setName(createObjectName(cluster, tenant, pulsarNamespace, name));
        v1ObjectMeta.setNamespace(k8sNamespace);
        if (ownerReference != null) {
            v1ObjectMeta.setOwnerReferences(Collections.singletonList(ownerReference));
        }

        return v1ObjectMeta;
    }

    public static String createObjectName(String cluster, String tenant, String namespace, String functionName) {
        final String convertedJobName = toValidPodName(functionName);
        // use of functionName may cause naming collisions,
        // add a short hash here to avoid it
        final String hashName = String.format("%s-%s-%s-%s", cluster, tenant, namespace, functionName);
        final String shortHash = DigestUtils.sha1Hex(hashName).toLowerCase().substring(0, 8);
        return convertedJobName + "-" + shortHash;
    }

    public static String generateObjectName(MeshWorkerService meshWorkerService,
                                            String tenant,
                                            String namespace,
                                            String componentName) {
        String pulsarCluster = meshWorkerService.getWorkerConfig().getPulsarFunctionsCluster();
        return createObjectName(pulsarCluster, tenant, namespace, componentName);
    }

    private static String toValidPodName(String ori) {
        return ori.toLowerCase().replaceAll("[^a-z0-9-\\.]", "-");
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
