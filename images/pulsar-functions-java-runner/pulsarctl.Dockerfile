ARG PULSAR_IMAGE
ARG PULSAR_IMAGE_TAG
FROM ${PULSAR_IMAGE}:${PULSAR_IMAGE_TAG} as pulsar
FROM pulsar-functions-pulsarctl-runner-base:latest

ARG PULSAR_IMAGE_TAG
ENV VERSION_TAG=${PULSAR_IMAGE_TAG}

RUN echo "VERSION_TAG=${VERSION_TAG}" && \
    VERSION_MAJOR=$(echo $VERSION_TAG | cut -d. -f1) && \
    VERSION_MINOR=$(echo $VERSION_TAG | cut -d. -f2) && \
    VERSION_PATCH=$(echo $VERSION_TAG | cut -d. -f3) && \
    if [ $VERSION_MAJOR -eq 2 ] && [ $VERSION_MINOR -eq 7 ]; then \
        echo "Pulsar version is 2.7, use java 1.8" && \
        export JRE_PACKAGE_NAME=openjdk-8-jre-headless && \
    elif [ $VERSION_MAJOR -eq 2 ] && [ $VERSION_MINOR -eq 8 ]; then \
        echo "Pulsar version is 2.8, use java 1.8" && \
        export JRE_PACKAGE_NAME=openjdk-8-jre-headless && \
    elif [ $VERSION_MAJOR -eq 2 ] && [ $VERSION_MINOR -eq 9 ]; then \
        echo "Pulsar version is 2.9, use java 11" && \
        export JRE_PACKAGE_NAME=openjdk-11-jre-headless && \
    elif [ $VERSION_MAJOR -eq 2 ] && [ $VERSION_MINOR -eq 10 ]; then \
        echo "Pulsar version is 2.10, use java 11" && \
        export JRE_PACKAGE_NAME=openjdk-11-jre-headless && \
    elif [ $VERSION_MAJOR -eq 2 ] && [ $VERSION_MINOR -eq 11 ]; then \
        echo "Pulsar version is 2.11, use java 17" && \
        export JRE_PACKAGE_NAME=openjdk-17-jre-headless && \
    else \
        echo "Pulsar version is not in the list, use java 17 instead" && \
        export JRE_PACKAGE_NAME=openjdk-17-jre-headless && \
    fi && \
    apt-get update \
         && apt-get -y dist-upgrade \
         && apt-get -y install $JRE_PACKAGE_NAME \
         && apt-get -y --purge autoremove \
         && apt-get autoclean \
         && apt-get clean \
         && rm -rf /var/lib/apt/lists/* \

COPY --from=pulsar --chown=$UID:$GID /pulsar/conf /pulsar/conf
COPY --from=pulsar --chown=$UID:$GID /pulsar/bin /pulsar/bin
COPY --from=pulsar --chown=$UID:$GID /pulsar/lib /pulsar/lib
COPY --from=pulsar --chown=$UID:$GID /pulsar/instances/java-instance.jar /pulsar/instances/java-instance.jar
COPY --from=pulsar --chown=$UID:$GID /pulsar/instances/deps /pulsar/instances/deps

# remove presto dependencies because they are not needed
RUN rm -rf /pulsar/lib/presto || true
RUN rm -rf /pulsar/conf/presto || true

ENV PULSAR_ROOT_LOGGER=INFO,CONSOLE
ENV java.io.tmpdir=/pulsar/tmp/

WORKDIR /pulsar

USER $USER
