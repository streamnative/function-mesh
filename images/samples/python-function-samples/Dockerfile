ARG PULSAR_IMAGE_TAG
FROM streamnative/pulsar-all:${PULSAR_IMAGE_TAG} as pulsar-all
FROM streamnative/pulsar-functions-python-runner:${PULSAR_IMAGE_TAG}
COPY --from=pulsar-all --chown=pulsar:0 /pulsar/examples /pulsar/examples
