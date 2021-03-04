ARG PULSAR_IMAGE
ARG PULSAR_IMAGE_TAG
FROM ${PULSAR_IMAGE}:${PULSAR_IMAGE_TAG} as pulsar
FROM pulsar-functions-runner-base:latest

COPY --from=pulsar --chown=pulsar:0 /pulsar/instances/python-instance /pulsar/instances/python-instance
COPY --from=pulsar --chown=pulsar:0 /pulsar/instances/deps /pulsar/instances/deps
COPY --from=pulsar --chown=pulsar:0 /pulsar/pulsar-client /pulsar/pulsar-client
COPY --from=pulsar --chown=pulsar:0 /pulsar/cpp-client /pulsar/cpp-client

# Install some utilities
RUN apt-get update \
     && DEBIAN_FRONTEND=noninteractive apt-get install -y python3.7 python3.7-dev python3-setuptools python3-yaml python3-kazoo \
                 libreadline-gplv2-dev libncursesw5-dev libssl-dev libsqlite3-dev tk-dev libgdbm-dev libc6-dev libbz2-dev \
                 curl \
     && apt-get clean autoclean && apt-get autoremove --yes && rm -rf /var/lib/apt/lists/*

RUN curl https://bootstrap.pypa.io/get-pip.py -o get-pip.py
RUN python3.7 get-pip.py

RUN update-alternatives --install /usr/bin/python python /usr/bin/python3.7 10

RUN apt-get update \
     && apt install -y /pulsar/cpp-client/*.deb \
     && apt-get clean autoclean && apt-get autoremove --yes && rm -rf /var/lib/apt/lists/*

WORKDIR /pulsar

# This script is intentionally run as the root user to make the dependencies
# available to the root user and the pulsar user
RUN /pulsar/bin/install-pulsar-client-37.sh

# Switch to the pulsar user to ensure container defaults to run as a non root user
USER pulsar