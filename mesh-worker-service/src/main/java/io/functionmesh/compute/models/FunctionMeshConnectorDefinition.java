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

import lombok.Data;
import lombok.NoArgsConstructor;
import org.apache.commons.lang3.StringUtils;
import org.apache.pulsar.common.io.ConnectorDefinition;

@Data
@NoArgsConstructor
public class FunctionMeshConnectorDefinition extends ConnectorDefinition {

    private static String DEFAULT_REGISTRY = "docker.io/";

    /**
     * The id of the IO connector.
     */
    private String id;

    /**
     * The version of the connector.
     */
    private String version;

    /**
     * The imageRegistry where host the connector image
     * By default the imageRegistry is empty, which refer to Docker Hub
     */
    private String imageRegistry;

    /**
     * The imageRepository to the connector
     * Usually it in format of NAMESPACE/REPOSITORY
     */
    private String imageRepository;

    /**
     * The imageTag to the connector image
     * By default it will align with Pulsar's version
     * TODO: set imageTag to version by default
     */
    private String imageTag;

    /**
     * Type class name of the connector or function
     * If not set, the default value '[B' will be used
     */
    private String typeClassName;

    /**
     * Type class name of the source connector
     * If not set, will inherit the value from typeClassName
     */
    private String sourceTypeClassName;

    /**
     * Type class name of the sink connector
     * If not set, will inherit the value from typeClassName
     */
    private String sinkTypeClassName;

    /**
     * Default schema type of the connector's topic, optional.
     */
    private String defaultSchemaType;

    /**
     * Default serde class name of the connector's topic, optional.
     */
    private String defaultSerdeClassName;

    public String toFullImageURL() {
        return String.format("%s%s:%s", imageRegistry != null ? imageRegistry : DEFAULT_REGISTRY,
                imageRepository, imageTag != null ? imageTag : version);
    }

    public String getJar() {
        return String.format("connectors/%s-%s.nar", id, version);
    }

    public String getSourceTypeClassName() {
        if (StringUtils.isNotBlank(sourceTypeClassName)) {
            return sourceTypeClassName;
        } else if (StringUtils.isNotBlank(typeClassName)) {
            return typeClassName;
        } else {
            return "[B";
        }
    }

    public String getSinkTypeClassName() {
        if (StringUtils.isNotBlank(sinkTypeClassName)) {
            return sinkTypeClassName;
        } else if (StringUtils.isNotBlank(typeClassName)) {
            return typeClassName;
        } else {
            return "[B";
        }
    }
}
