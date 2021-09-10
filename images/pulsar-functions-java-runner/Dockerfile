ARG PULSAR_IMAGE
ARG PULSAR_IMAGE_TAG
FROM ${PULSAR_IMAGE}:${PULSAR_IMAGE_TAG} as pulsar
FROM pulsar-functions-runner-base:latest

COPY --from=pulsar --chown=$PULSAR_RUNNER_USER:$PULSAR_RUNNER_GROUP /pulsar/instances/java-instance.jar /pulsar/instances/java-instance.jar
COPY --from=pulsar --chown=$PULSAR_RUNNER_USER:$PULSAR_RUNNER_GROUP /pulsar/instances/deps /pulsar/instances/deps

WORKDIR /pulsar

USER $PULSAR_RUNNER_USER