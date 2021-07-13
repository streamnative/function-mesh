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

import com.google.common.util.concurrent.FutureCallback;
import com.google.common.util.concurrent.Futures;
import com.google.common.util.concurrent.ListenableFuture;
import com.google.common.util.concurrent.MoreExecutors;
import com.google.gson.Gson;
import com.google.protobuf.Empty;
import io.functionmesh.compute.MeshWorkerService;
import io.functionmesh.compute.models.CustomRuntimeOptions;
import io.kubernetes.client.openapi.models.V1ObjectMeta;
import io.kubernetes.client.openapi.models.V1OwnerReference;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.codec.digest.DigestUtils;
import org.apache.logging.log4j.util.Strings;
import org.apache.pulsar.common.functions.FunctionConfig;
import org.apache.pulsar.common.policies.data.ExceptionInformation;
import org.apache.pulsar.common.util.RestException;
import org.apache.pulsar.functions.proto.InstanceCommunication;
import org.apache.pulsar.functions.proto.InstanceControlGrpc;

import javax.ws.rs.core.Response;
import java.util.Collections;
import java.util.Map;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;

import static io.functionmesh.compute.util.KubernetesUtils.GRPC_TIMEOUT_SECS;

@Slf4j
public class CommonUtil {
    public static final String COMPONENT_FUNCTION = "function";
    public static final String COMPONENT_SOURCE = "source";
    public static final String COMPONENT_SINK = "sink";
    public static final String COMPONENT_STATEFUL_SET = "StatefulSet";
    public static final String COMPONENT_SERVICE = "Service";
    public static final String COMPONENT_HPA = "HorizontalPodAutoscaler";
    private static final String CLUSTER_NAME_ENV = "clusterName";

    public static String getClusterNameEnv() {
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

    // Return a CustomRuntimeOption if a json string is provided, otherwise an empty object is returned
    public static CustomRuntimeOptions getCustomRuntimeOptions(String customRuntimeOptionsJSON) {
        CustomRuntimeOptions customRuntimeOptions;
        if (Strings.isNotEmpty(customRuntimeOptionsJSON)) {
            try {
                customRuntimeOptions =
                        new Gson().fromJson(customRuntimeOptionsJSON, CustomRuntimeOptions.class);
            } catch (Exception ignored) {
                throw new RestException(
                        Response.Status.BAD_REQUEST, "customRuntimeOptions cannot be deserialized.");
            }
        } else {
            customRuntimeOptions = new CustomRuntimeOptions();
        }

        return customRuntimeOptions;
    }

    public static String getClusterName(String cluster, CustomRuntimeOptions customRuntimeOptions) {
        if (cluster != null) {
            return cluster;
        } else if (Strings.isNotEmpty(customRuntimeOptions.getClusterName())){
            return customRuntimeOptions.getClusterName();
        } else if (Strings.isNotEmpty(CommonUtil.getClusterNameEnv())){
            return CommonUtil.getClusterNameEnv();
        } else {
            throw new RestException(Response.Status.BAD_REQUEST, "clusterName is not provided.");
        }
    }

    public static ExceptionInformation getExceptionInformation(InstanceCommunication.FunctionStatus.ExceptionInformation exceptionEntry) {
        ExceptionInformation exceptionInformation
                = new ExceptionInformation();
        exceptionInformation.setTimestampMs(exceptionEntry.getMsSinceEpoch());
        exceptionInformation.setExceptionString(exceptionEntry.getExceptionString());
        return exceptionInformation;
    }

    public static String makeJobName(String name, String suffix) {
        return String.format("%s-%s", name, suffix);
    }

    public static int getShardIdFromPodName(String podName) {
        int shardId = -1;
        try {
            shardId = new Integer(podName.substring(podName.lastIndexOf("-")+1));
        } catch (Exception ex) {
            log.error("getShardIdFromPodName failed with podName {}, exception: {}", podName, ex);
        }
        return shardId;
    }

    public static CompletableFuture<InstanceCommunication.FunctionStatus> getFunctionStatusAsync(InstanceControlGrpc.InstanceControlFutureStub stub) {
        CompletableFuture<InstanceCommunication.FunctionStatus> retval = new CompletableFuture<>();
        if (stub == null) {
            retval.completeExceptionally(new RuntimeException("Not alive"));
            return retval;
        }
        ListenableFuture<InstanceCommunication.FunctionStatus> response = stub.withDeadlineAfter(GRPC_TIMEOUT_SECS, TimeUnit.SECONDS).getFunctionStatus(Empty.newBuilder().build());
        Futures.addCallback(response, new FutureCallback<InstanceCommunication.FunctionStatus>() {
            @Override
            public void onFailure(Throwable throwable) {
                InstanceCommunication.FunctionStatus.Builder builder = InstanceCommunication.FunctionStatus.newBuilder();
                builder.setRunning(false);
                builder.setFailureException(throwable.getMessage());
                retval.complete(builder.build());
            }

            @Override
            public void onSuccess(InstanceCommunication.FunctionStatus t) {
                retval.complete(t);
            }
        }, MoreExecutors.directExecutor());
        return retval;
    }
}
