ARG PULSAR_IMAGE
ARG PULSAR_IMAGE_TAG
FROM ${PULSAR_IMAGE}:${PULSAR_IMAGE_TAG} as pulsar
FROM ubuntu:18.04 as functions-runner
RUN apt-get update && apt-get install -y openjdk-8-jre-headless && apt-get clean autoclean && apt-get autoremove --yes && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /pulsar/bin/
RUN mkdir -p /pulsar/lib/
RUN mkdir -p /pulsar/conf/
RUN mkdir -p /pulsar/instances/
RUN mkdir -p /pulsar/connectors/
RUN mkdir -p /pulsar/logs/

RUN chmod -R g=u /pulsar

# Create the pulsar group and user to make docker container run as a non root user by default
RUN groupadd -g 10001 pulsar
RUN adduser -u 10000 --gid 10001 --disabled-login --disabled-password --gecos '' pulsar

COPY --from=pulsar --chown=pulsar:0 /pulsar/conf /pulsar/conf
COPY --from=pulsar --chown=pulsar:0 /pulsar/bin /pulsar/bin
COPY --from=pulsar --chown=pulsar:0 /pulsar/lib /pulsar/lib

RUN chown -R pulsar:0 /pulsar && chmod -R g=u /pulsar

ENV PULSAR_ROOT_LOGGER=INFO,CONSOLE

WORKDIR /pulsar