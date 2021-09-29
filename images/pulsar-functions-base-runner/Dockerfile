ARG PULSAR_IMAGE
ARG PULSAR_IMAGE_TAG
FROM ${PULSAR_IMAGE}:${PULSAR_IMAGE_TAG} as pulsar
FROM ubuntu:20.04 as functions-runner

ENV PULSAR_RUNNER_USER=pulsar
ENV PULSAR_RUNNER_GROUP=pulsar
RUN groupadd -g 10001 $GID
RUN adduser -u 10000 --gid 10001 --disabled-login --disabled-password --gecos '' $UID

RUN mkdir -p /pulsar/bin/ \
    && mkdir -p /pulsar/lib/ \
    && mkdir -p /pulsar/conf/ \
    && mkdir -p /pulsar/instances/ \
    && mkdir -p /pulsar/connectors/ \
    && mkdir -p /pulsar/logs/ \
    && mkdir -p /pulsar/tmp/ \
    && mkdir -p /pulsar/examples/ \
    && chown -R $UID:$GID /pulsar \
    && chmod -R g=u /pulsar

COPY --from=pulsar --chown=$UID:$GID /pulsar/conf /pulsar/conf
COPY --from=pulsar --chown=$UID:$GID /pulsar/bin /pulsar/bin
COPY --from=pulsar --chown=$UID:$GID /pulsar/lib /pulsar/lib

RUN chmod -R g=u /pulsar

RUN apt-get update \
     && apt-get -y dist-upgrade \
     && apt-get -y install openjdk-11-jdk-headless netcat dnsutils less procps iputils-ping \
     && apt-get -y --purge autoremove \
     && apt-get autoclean \
     && apt-get clean \
     && rm -rf /var/lib/apt/lists/*

ENV PULSAR_ROOT_LOGGER=INFO,CONSOLE
ENV java.io.tmpdir=/pulsar/tmp/

WORKDIR /pulsar
