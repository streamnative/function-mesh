ARG PULSAR_IMAGE
ARG PULSAR_IMAGE_TAG
FROM ${PULSAR_IMAGE}:${PULSAR_IMAGE_TAG} as pulsar
FROM pulsar-functions-runner-base:latest

COPY --from=pulsar --chown=$UID:$GID /pulsar/instances/java-instance.jar /pulsar/instances/java-instance.jar
COPY --from=pulsar --chown=$UID:$GID /pulsar/instances/deps /pulsar/instances/deps

RUN cp /pulsar/lib/com.fasterxml.jackson.dataformat-jackson-dataformat-yaml-*.jar /pulsar/instances/deps/dataformat-jackson-dataformat-yaml.jar \
  && cp /pulsar/lib/org.yaml-snakeyaml-*.jar /pulsar/instances/deps/org.yaml-snakeyaml.jar

WORKDIR /pulsar

USER $USER
