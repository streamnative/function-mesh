ARG PULSAR_IMAGE
ARG PULSAR_IMAGE_TAG
ARG DISTROLESS_IMAGE
ARG DISTROLESS_IMAGE_TAG
FROM ${PULSAR_IMAGE}:${PULSAR_IMAGE_TAG} as pulsar

FROM ubuntu:20.04 as ubuntu

ENV GID=10001
ENV UID=10000
ENV USER=pulsar
RUN groupadd -g $GID pulsar
RUN adduser -u $UID --gid $GID --disabled-login --disabled-password --gecos '' $USER

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

FROM ${DISTROLESS_IMAGE}:${DISTROLESS_IMAGE_TAG} as functions-runner

ENV GID=10001
ENV UID=10000
ENV USER=pulsar

# copy user and group files
COPY --from=ubuntu /etc/passwd /etc/passwd
COPY --from=ubuntu /etc/group /etc/group
COPY --from=ubuntu --chown=$UID:$GID --chmod=g=u /pulsar /pulsar/
COPY --from=pulsar --chown=$UID:$GID /pulsar/conf /pulsar/conf
COPY --from=pulsar --chown=$UID:$GID /pulsar/bin /pulsar/bin
COPY --from=pulsar --chown=$UID:$GID /pulsar/lib /pulsar/lib

ENV PULSAR_ROOT_LOGGER=INFO,CONSOLE
ENV java.io.tmpdir=/pulsar/tmp/

WORKDIR /pulsar
