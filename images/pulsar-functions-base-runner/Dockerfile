ARG PULSAR_IMAGE
ARG PULSAR_IMAGE_TAG
FROM ${PULSAR_IMAGE}:${PULSAR_IMAGE_TAG} as pulsar
FROM alpine:3.20 as functions-runner

ENV GID=10001
ENV UID=10000
ENV USER=pulsar
RUN addgroup -g $GID pulsar
RUN adduser -u $UID -G pulsar -D -g '' $USER

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

ARG PULSAR_IMAGE_TAG
ENV VERSION_TAG=${PULSAR_IMAGE_TAG}
RUN echo "VERSION_TAG=${VERSION_TAG}" && \
    VERSION_MAJOR=$(echo $VERSION_TAG | cut -d. -f1) && \
    VERSION_MINOR=$(echo $VERSION_TAG | cut -d. -f2) && \
    VERSION_PATCH=$(echo $VERSION_TAG | cut -d. -f3) && \
    if [ $VERSION_MAJOR -eq 2 ] && [ $VERSION_MINOR -eq 7 ]; then \
        echo "Pulsar version is 2.7, use java 1.8" && \
        export JRE_PACKAGE_NAME=openjdk8; \
    elif [ $VERSION_MAJOR -eq 2 ] && [ $VERSION_MINOR -eq 8 ]; then \
        echo "Pulsar version is 2.8, use java 1.8" && \
        export JRE_PACKAGE_NAME=openjdk8; \
    elif [ $VERSION_MAJOR -eq 2 ] && [ $VERSION_MINOR -eq 9 ]; then \
        echo "Pulsar version is 2.9, use java 11" && \
        export JRE_PACKAGE_NAME=openjdk11; \
    elif [ $VERSION_MAJOR -eq 2 ] && [ $VERSION_MINOR -eq 10 ]; then \
        echo "Pulsar version is 2.10, use java 11" && \
        export JRE_PACKAGE_NAME='openjdk11 gcompat'; \
    elif [ $VERSION_MAJOR -eq 2 ] && [ $VERSION_MINOR -eq 11 ]; then \
        echo "Pulsar version is 2.11, use java 17" && \
        export JRE_PACKAGE_NAME='openjdk17 gcompat'; \
    elif [ $VERSION_MAJOR -eq 3 ]; then \
        echo "Pulsar version is 3.x, use java 17 instead" && \
        export JRE_PACKAGE_NAME='openjdk17 gcompat'; \
    else \
        echo "Pulsar version is not in the list, use java 21 instead" && \
        export JRE_PACKAGE_NAME='openjdk17 gcompat'; \
    fi && \
    apk update && apk add --no-cache $JRE_PACKAGE_NAME bash

ENV LD_PRELOAD=/lib/libgcompat.so.0

COPY --from=pulsar --chown=$UID:$GID /pulsar/conf /pulsar/conf
COPY --from=pulsar --chown=$UID:$GID /pulsar/bin /pulsar/bin
COPY --from=pulsar --chown=$UID:$GID /pulsar/lib /pulsar/lib

# remove the vertx jar since it's not need ans has a cve
RUN rm -rf /pulsar/lib/io.vertx-vertx-core-*.jar || true

# remove presto dependencies because they are not needed
RUN rm -rf /pulsar/lib/presto || true
RUN rm -rf /pulsar/conf/presto || true
RUN rm -rf /pulsar/bin/pulsar-detector || true

ENV PULSAR_ROOT_LOGGER=INFO,CONSOLE
ENV java.io.tmpdir=/pulsar/tmp/

# Copy the kinesis_producer native executable compiled for Alpine musl to the pulsar-all image
# This is required to support the Pulsar IO Kinesis sink connector
COPY --from=pulsar /opt/amazon-kinesis-producer/bin/kinesis_producer /opt/amazon-kinesis-producer/bin/.os_info /opt/amazon-kinesis-producer/bin/.build_time /opt/amazon-kinesis-producer/bin/.revision /opt/amazon-kinesis-producer/bin/.system_info /opt/amazon-kinesis-producer/bin/.version /opt/amazon-kinesis-producer/bin/
# Set the environment variable to point to the kinesis_producer native executable
ENV PULSAR_IO_KINESIS_KPL_PATH=/opt/amazon-kinesis-producer/bin/kinesis_producer
# Install the required dependencies for the kinesis_producer native executable
USER 0
RUN apk update && apk add --no-cache \
    brotli-libs \
    c-ares \
    libcrypto3 \
    libcurl \
    libgcc \
    libidn2 \
    libpsl \
    libssl3 \
    libunistring \
    nghttp2-libs \
    zlib \
    zstd-libs \
    libuuid

WORKDIR /pulsar
