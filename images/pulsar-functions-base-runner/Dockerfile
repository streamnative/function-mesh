ARG PULSAR_IMAGE
ARG PULSAR_IMAGE_TAG
FROM ${PULSAR_IMAGE}:${PULSAR_IMAGE_TAG} as pulsar
FROM ubuntu:20.04 as functions-runner

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

RUN apt-get update \
     && apt-get -y dist-upgrade \
     && apt-get -y install openjdk-11-jdk-headless netcat dnsutils less procps iputils-ping \
     && apt-get -y --purge autoremove \
     && apt-get autoclean \
     && apt-get clean \
     && rm -rf /var/lib/apt/lists/*

ENV PULSAR_ROOT_LOGGER=INFO,CONSOLE

WORKDIR /pulsar