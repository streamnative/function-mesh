ARG PULSAR_IMAGE
ARG PULSAR_IMAGE_TAG
FROM ${PULSAR_IMAGE}:${PULSAR_IMAGE_TAG} as pulsar
FROM pulsar-functions-runner-base:latest

COPY --from=pulsar --chown=pulsar:0 /pulsar/instances/python-instance /pulsar/instances/python-instance
COPY --from=pulsar --chown=pulsar:0 /pulsar/instances/deps /pulsar/instances/deps
COPY --from=pulsar --chown=pulsar:0 /pulsar/pulsar-client /pulsar/pulsar-client
# Pulsar 2.8.0 removes /pulsar/cpp-client from docker image
# But it required with Pulsar 2.7.X and below
# to make this Dockerfile compalicate with different Pulsar versions
# Below is a hacky way to copy /pulsar/cpp-client if exist in pulsar image
COPY --from=pulsar --chown=pulsar:0 /pulsar/README /pulsar/cpp-clien* /tmp/pulsar/
RUN mv /tmp/pulsar/cpp-client /pulsar/cpp-client || true

# Install some utilities
RUN apt-get update \
     && DEBIAN_FRONTEND=noninteractive apt-get install -y python3.7 python3.7-dev python3-setuptools python3-yaml python3-kazoo \
                 libreadline-gplv2-dev libncursesw5-dev libssl-dev libsqlite3-dev tk-dev libgdbm-dev libc6-dev libbz2-dev \
                 curl \
     && apt-get clean autoclean && apt-get autoremove --yes && rm -rf /var/lib/apt/lists/*

RUN curl https://bootstrap.pypa.io/get-pip.py -o get-pip.py
RUN python3.7 get-pip.py

RUN update-alternatives --install /usr/bin/python python /usr/bin/python3.7 10

RUN if [ -d "/pulsar/cpp-client" ]; then apt-get update \
     && apt install -y /pulsar/cpp-client/*.deb \
     && apt-get clean autoclean && apt-get autoremove --yes && rm -rf /var/lib/apt/lists/* ; fi

WORKDIR /pulsar

# This script is intentionally run as the root user to make the dependencies
# available to the root user and the pulsar user
RUN /pulsar/bin/install-pulsar-client-37.sh

# Switch to the pulsar user to ensure container defaults to run as a non root user
USER pulsar