ARG PULSAR_IMAGE
ARG PULSAR_IMAGE_TAG
FROM ${PULSAR_IMAGE}:${PULSAR_IMAGE_TAG} as pulsar
FROM pulsar-functions-pulsarctl-runner-base:latest

COPY --from=pulsar --chown=$UID:$GID /pulsar/bin /pulsar/bin
COPY --from=pulsar --chown=$UID:$GID /pulsar/instances/python-instance /pulsar/instances/python-instance
# Pulsar 2.11.0 removes /pulsar/pulsar-client from docker image
# But it required with Pulsar 2.10.X and below
# to make this Dockerfile compalicate with different Pulsar versions
# Below is a hacky way to copy /pulsar/pulsar-client if exist in pulsar image
COPY --from=pulsar --chown=$UID:$GID /pulsar/README /pulsar/pulsar-clien* /pulsar/pulsar-client/

# Pulsar 2.8.0 removes /pulsar/cpp-client from docker image
# But it required with Pulsar 2.7.X and below
# to make this Dockerfile compalicate with different Pulsar versions
# Below is a hacky way to copy /pulsar/cpp-client if exist in pulsar image
COPY --from=pulsar --chown=$UID:$GID /pulsar/README /pulsar/cpp-clien* /pulsar/cpp-client/

RUN apt-get update \
     && DEBIAN_FRONTEND=noninteractive apt-get install -y python3 python3-dev python3-setuptools python3-yaml python3-kazoo \
                 libreadline-gplv2-dev libncursesw5-dev libssl-dev libsqlite3-dev tk-dev libgdbm-dev libc6-dev libbz2-dev \
                 curl ca-certificates\
     && apt-get clean autoclean && apt-get autoremove --yes && rm -rf /var/lib/apt/lists/* \
     && mkdir -p /etc/pki/tls/certs && cp /etc/ssl/certs/ca-certificates.crt /etc/pki/tls/certs/ca-bundle.crt \
     && update-alternatives --install /usr/bin/python python /usr/bin/python3 10 \
     && curl https://bootstrap.pypa.io/get-pip.py -o get-pip.py \
     && python3 get-pip.py && pip3 install --upgrade pip

RUN if [ -d "/pulsar/cpp-client" ]; then apt-get update \
     && apt install -y /pulsar/cpp-client/*.deb || true \
     && apt-get clean autoclean && apt-get autoremove --yes && rm -rf /var/lib/apt/lists/* ; fi

RUN if [ -f "/pulsar/bin/install-pulsar-client-37.sh" ]; then /pulsar/bin/install-pulsar-client-37.sh || pip3 install 'pulsar-client[all]==3.1.0' ; fi
RUN if [ -f "/pulsar/bin/install-pulsar-client.sh" ]; then /pulsar/bin/install-pulsar-client.sh || pip3 install 'pulsar-client[all]==3.1.0' ; fi

WORKDIR /pulsar

USER $USER
# a temp solution from https://github.com/apache/pulsar/pull/15846 to fix python protobuf version error
RUN pip3 install protobuf==3.20.1 --user
