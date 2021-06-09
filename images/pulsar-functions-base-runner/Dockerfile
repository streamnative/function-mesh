ARG PULSAR_IMAGE
ARG PULSAR_IMAGE_TAG
FROM ${PULSAR_IMAGE}:${PULSAR_IMAGE_TAG} as pulsar
FROM openjdk:11-jre-slim as functions-runner

RUN mkdir -p /pulsar/bin/
RUN mkdir -p /pulsar/lib/
RUN mkdir -p /pulsar/conf/
RUN mkdir -p /pulsar/instances/
RUN mkdir -p /pulsar/connectors/
RUN mkdir -p /pulsar/logs/

RUN chmod -R g=u /pulsar

COPY --from=pulsar /pulsar/conf /pulsar/conf
COPY --from=pulsar /pulsar/bin /pulsar/bin
COPY --from=pulsar /pulsar/lib /pulsar/lib

RUN chmod -R g=u /pulsar

ENV PULSAR_ROOT_LOGGER=INFO,CONSOLE

WORKDIR /pulsar