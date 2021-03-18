ARG PULSAR_IMAGE
ARG PULSAR_IMAGE_TAG
FROM ${PULSAR_IMAGE}:${PULSAR_IMAGE_TAG} as pulsar
FROM pulsar-functions-runner-base:latest

COPY --from=pulsar --chown=pulsar:0 /pulsar/instances/java-instance.jar /pulsar/instances/java-instance.jar
COPY --from=pulsar --chown=pulsar:0 /pulsar/instances/deps /pulsar/instances/deps

WORKDIR /pulsar

# Switch to the pulsar user to ensure container defaults to run as a non root user
USER pulsar