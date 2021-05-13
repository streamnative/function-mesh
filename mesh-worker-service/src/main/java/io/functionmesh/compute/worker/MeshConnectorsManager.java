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
package io.functionmesh.compute.worker;

import io.functionmesh.compute.models.FunctionMeshConnectorDefinition;
import lombok.Getter;
import lombok.extern.slf4j.Slf4j;
import org.apache.pulsar.common.io.ConnectorDefinition;
import org.apache.pulsar.common.util.ObjectMapperFactory;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.List;
import java.util.TreeMap;
import java.util.stream.Collectors;

@Slf4j
public class MeshConnectorsManager {
    private static final String PULSAR_IO_CONNECTORS_CONFIG = "conf/connectors.yaml";

    @Getter
    private volatile TreeMap<String, FunctionMeshConnectorDefinition> connectors;

    public MeshConnectorsManager() {
        this.connectors = searchForConnectors();
    }

    public static TreeMap<String, FunctionMeshConnectorDefinition> searchForConnectors() {
        Path path = Paths.get(PULSAR_IO_CONNECTORS_CONFIG).toAbsolutePath();
        log.info("Connectors configs in {}", path);

        TreeMap<String, FunctionMeshConnectorDefinition> results = new TreeMap<>();

        if (!path.toFile().exists()) {
            log.warn("Connectors configs not found");
            return results;
        }
        try {
            String configs = new String(Files.readAllBytes(path), StandardCharsets.UTF_8);
            FunctionMeshConnectorDefinition[] data = ObjectMapperFactory.getThreadLocalYaml().readValue(configs, FunctionMeshConnectorDefinition[].class);
            for (FunctionMeshConnectorDefinition d : data) {
                results.put(d.getName(), d);
            }
        } catch (IOException e) {
            log.error("Cannot parse connector definitions", e);
        }

        return results;
    }

    public FunctionMeshConnectorDefinition getConnectorDefinition(String connectorType) {
        return connectors.get(connectorType);
    }

    public void reloadConnectors() {
        connectors = searchForConnectors();
    }

    public List<ConnectorDefinition> getConnectorDefinitions() {
        return connectors.values().stream().map(v -> (ConnectorDefinition)v).collect(Collectors.toList());
    }
}
