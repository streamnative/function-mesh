package io.functionmesh.compute.worker;

import io.functionmesh.compute.models.FunctionMeshConnectorDefinition;
import lombok.Getter;
import lombok.extern.slf4j.Slf4j;
import org.apache.pulsar.functions.utils.io.Connector;
import org.apache.pulsar.functions.utils.io.ConnectorUtils;
import org.apache.pulsar.functions.worker.WorkerConfig;

import java.io.IOException;
import java.util.TreeMap;

@Slf4j
public class FunctionMeshConnectorsManager {
    @Getter
    private volatile TreeMap<String, FunctionMeshConnectorDefinition> connectors;

    public FunctionMeshConnectorDefinition getConnectorDefinition(String connectorType) {
        return connectors.get(connectorType);
    }

    public void reloadConnectors(WorkerConfig workerConfig) throws IOException {
        connectors = searchForConnectors(workerConfig.getConnectorsDirectory(), workerConfig.getNarExtractionDirectory());
    }
}
