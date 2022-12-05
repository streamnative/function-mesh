ARG PULSAR_IMAGE
ARG PULSAR_IMAGE_TAG
ARG RUNNER_IMAGE
FROM ${PULSAR_IMAGE}:${PULSAR_IMAGE_TAG} as pulsar-all
FROM ${RUNNER_IMAGE}:${PULSAR_IMAGE_TAG}
COPY --from=pulsar-all --chown=$UID:$GID /pulsar/connectors/pulsar-io-elastic-search*.nar /pulsar/connectors/
