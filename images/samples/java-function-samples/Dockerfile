ARG PULSAR_IMAGE_TAG
FROM streamnative/sn-platform:${PULSAR_IMAGE_TAG} as pulsar-all
FROM streamnative/pulsar-functions-java-runner:${PULSAR_IMAGE_TAG}
COPY --from=pulsar-all --chown=$UID:$GID /pulsar/examples /pulsar/examples
