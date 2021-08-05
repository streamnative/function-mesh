FROM streamnative/pulsar-all:2.8.0.7
COPY ./target/mesh-worker-service*.nar /pulsar/mesh-worker-service.nar
COPY ./integration-tests/docker/connectors.yaml /pulsar/conf/connectors.yaml
RUN mkdir -p /pulsar-nar
