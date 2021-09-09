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
package io.functionmesh.compute.models;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecPodVolumeMounts;
import io.functionmesh.compute.functions.models.V1alpha1FunctionSpecPodVolumes;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecPodVolumeMounts;
import io.functionmesh.compute.sinks.models.V1alpha1SinkSpecPodVolumes;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecPodVolumeMounts;
import io.functionmesh.compute.sources.models.V1alpha1SourceSpecPodVolumes;
import io.kubernetes.client.openapi.models.V1Volume;
import io.kubernetes.client.openapi.models.V1VolumeMount;
import lombok.Data;
import lombok.experimental.Accessors;
import org.apache.pulsar.common.configuration.FieldContext;
import org.apache.pulsar.common.util.ObjectMapperFactory;

import java.util.List;
import java.util.Map;

@Data
@Accessors(chain = true)
public class MeshWorkerServiceCustomConfig {
    @FieldContext(
            doc = "Enable user to upload custom function/source/sink jar/nar"
    )
    protected boolean uploadEnabled = false;

    @FieldContext(
            doc = "Enable the function api endpoint"
    )
    protected boolean functionEnabled = true;

    @FieldContext(
            doc = "Enable the sink api endpoint"
    )
    protected boolean sinkEnabled = true;

    @FieldContext(
            doc = "Enable the source api endpoint"
    )
    protected boolean sourceEnabled = true;

    @FieldContext(
            doc = "the directory for dropping extra function dependencies. "
    )
    protected String extraDependenciesDir = "/pulsar/lib/";

    @FieldContext(
            doc = "VolumeMount describes a mounting of a Volume within a container."
    )
    protected List<V1VolumeMount> volumeMounts;

    @FieldContext(
            doc = "List of volumes that can be mounted by containers belonging to the function/connector pod."
    )
    protected List<V1Volume> volumes;

    @FieldContext(
            doc = "ownerReference"
    )
    protected Map<String, Object> ownerReference;

    @FieldContext(
            doc = "if allow user to change the service account name with custom-runtime-options"
    )
    protected boolean allowUserDefinedServiceAccountName = false;

    @FieldContext(
            doc = "ServiceAccountName is the name of the ServiceAccount to use to run function/connector pod."
    )
    protected String defaultServiceAccountName;

    public List<V1alpha1SinkSpecPodVolumes> asV1alpha1SinkSpecPodVolumesList() throws JsonProcessingException {
        ObjectMapper objectMapper = ObjectMapperFactory.getThreadLocal();
        TypeReference<List<V1alpha1SinkSpecPodVolumes>> typeRef
                = new TypeReference<List<V1alpha1SinkSpecPodVolumes>>() {};
        String j = objectMapper.writeValueAsString(volumes);
        return objectMapper.readValue(j, typeRef);
    }

    public List<V1alpha1SourceSpecPodVolumes> asV1alpha1SourceSpecPodVolumesList() throws JsonProcessingException {
        ObjectMapper objectMapper = ObjectMapperFactory.getThreadLocal();
        TypeReference<List<V1alpha1SourceSpecPodVolumes>> typeRef
                = new TypeReference<List<V1alpha1SourceSpecPodVolumes>>() {};
        String j = objectMapper.writeValueAsString(volumes);
        return objectMapper.readValue(j, typeRef);
    }

    public List<V1alpha1FunctionSpecPodVolumes> asV1alpha1FunctionSpecPodVolumesList() throws JsonProcessingException {
        ObjectMapper objectMapper = ObjectMapperFactory.getThreadLocal();
        TypeReference<List<V1alpha1FunctionSpecPodVolumes>> typeRef
                = new TypeReference<List<V1alpha1FunctionSpecPodVolumes>>() {};
        String j = objectMapper.writeValueAsString(volumes);
        return objectMapper.readValue(j, typeRef);
    }

    public List<V1alpha1SinkSpecPodVolumeMounts> asV1alpha1SinkSpecPodVolumeMountsList() throws JsonProcessingException {
        ObjectMapper objectMapper = ObjectMapperFactory.getThreadLocal();
        TypeReference<List<V1alpha1SinkSpecPodVolumeMounts>> typeRef
                = new TypeReference<List<V1alpha1SinkSpecPodVolumeMounts>>() {};
        String j = objectMapper.writeValueAsString(volumeMounts);
        return objectMapper.readValue(j, typeRef);
    }

    public List<V1alpha1SourceSpecPodVolumeMounts> asV1alpha1SourceSpecPodVolumeMountsList() throws JsonProcessingException {
        ObjectMapper objectMapper = ObjectMapperFactory.getThreadLocal();
        TypeReference<List<V1alpha1SourceSpecPodVolumeMounts>> typeRef
                = new TypeReference<List<V1alpha1SourceSpecPodVolumeMounts>>() {};
        String j = objectMapper.writeValueAsString(volumeMounts);
        return objectMapper.readValue(j, typeRef);
    }

    public List<V1alpha1FunctionSpecPodVolumeMounts> asV1alpha1FunctionSpecPodVolumeMounts() throws JsonProcessingException {
        ObjectMapper objectMapper = ObjectMapperFactory.getThreadLocal();
        TypeReference<List<V1alpha1FunctionSpecPodVolumeMounts>> typeRef
                = new TypeReference<List<V1alpha1FunctionSpecPodVolumeMounts>>() {};
        String j = objectMapper.writeValueAsString(volumeMounts);
        return objectMapper.readValue(j, typeRef);
    }
}
