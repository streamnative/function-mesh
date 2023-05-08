ARG PULSAR_IMAGE
ARG PULSAR_IMAGE_TAG
FROM ${PULSAR_IMAGE}:${PULSAR_IMAGE_TAG} as pulsar
FROM pulsar-functions-pulsarctl-runner-base:latest

COPY --from=pulsar --chown=$UID:$GID /pulsar/instances/python-instance /pulsar/instances/python-instance

RUN apt-get update \
     && DEBIAN_FRONTEND=noninteractive apt-get install -y python3 python3-dev python3-setuptools python3-yaml python3-kazoo \
                 libreadline-gplv2-dev libncursesw5-dev libssl-dev libsqlite3-dev tk-dev libgdbm-dev libc6-dev libbz2-dev \
                 curl ca-certificates\
     && apt-get clean autoclean && apt-get autoremove --yes && rm -rf /var/lib/apt/lists/* \
     && mkdir -p /etc/pki/tls/certs && cp /etc/ssl/certs/ca-certificates.crt /etc/pki/tls/certs/ca-bundle.crt \
     && update-alternatives --install /usr/bin/python python /usr/bin/python3 10 \
     && curl https://bootstrap.pypa.io/get-pip.py -o get-pip.py \
     && python3 get-pip.py && pip3 --upgrade install pip \
     && pip3 install protobuf==3.20.1 grpcio>=1.8.2 apache-bookkeeper-client>=4.16.1 prometheus_client ratelimit fastavro==1.7.3 pulsar-client

WORKDIR /pulsar

USER $USER
