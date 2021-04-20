package io.functionmesh.compute.models;

import lombok.Data;
import lombok.NoArgsConstructor;
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

    public String toFullImageURL() {
        return String.format("%s%s:%s", imageRegistry != null ? imageRegistry : DEFAULT_REGISTRY,
                imageRepository, imageTag != null ? imageTag : version);
    }

    public String getJar() {
        return String.format("connectors/%s-%s.nar", id, version);
    }
}
