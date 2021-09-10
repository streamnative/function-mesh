ARG PULSAR_IMAGE_TAG
FROM streamnative/pulsar-all:${PULSAR_IMAGE_TAG} as pulsar-all
FROM streamnative/pulsar-functions-python-runner:${PULSAR_IMAGE_TAG}
COPY --from=pulsar-all --chown=$PULSAR_RUNNER_USER:$PULSAR_RUNNER_GROUP /pulsar/examples /pulsar/examples
