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

ENV PULSAR_CLIENT_PYTHON_VERSION=3.5.0

# Pulsar 2.8.0 removes /pulsar/cpp-client from docker image
# But it required with Pulsar 2.7.X and below
# to make this Dockerfile compalicate with different Pulsar versions
# Below is a hacky way to copy /pulsar/cpp-client if exist in pulsar image
COPY --from=pulsar --chown=$UID:$GID /pulsar/README /pulsar/cpp-clien* /pulsar/cpp-client/

RUN apk update \
     && apk add --no-cache python3 python3-dev tk-dev curl ca-certificates\
     && mkdir -p /etc/pki/tls/certs && cp /etc/ssl/certs/ca-certificates.crt /etc/pki/tls/certs/ca-bundle.crt \
     && curl https://bootstrap.pypa.io/get-pip.py -o get-pip.py \
     && mv /usr/lib/python3.*/EXTERNALLY-MANAGED  /tmp/EXTERNALLY-MANAGED.old \
     && python3 get-pip.py && pip3 install --upgrade pip

RUN if [ -f "/pulsar/bin/install-pulsar-client-37.sh" ]; then /pulsar/bin/install-pulsar-client-37.sh || pip3 install 'pulsar-client[all]==3.5.0' ; else pip3 install 'pulsar-client[all]==3.5.0' ; fi
RUN if [ -f "/pulsar/bin/install-pulsar-client.sh" ]; then /pulsar/bin/install-pulsar-client.sh || pip3 install 'pulsar-client[all]==3.5.0' ; else pip3 install 'pulsar-client[all]==3.5.0' ; fi

# this dir is duplicate with the installed pulsar-client pip package, and maybe not compatible with the `_pulsar`(the .so library package)
RUN rm -rf /pulsar/instances/python-instance/pulsar/ \
     && sed -i "s/serde.IdentitySerDe/pulsar.functions.serde.IdentitySerDe/g" /pulsar/instances/python-instance/python_instance.py \
     && sed -i "s/serde.IdentitySerDe/pulsar.functions.serde.IdentitySerDe/g" /pulsar/instances/python-instance/contextimpl.py

WORKDIR /pulsar

USER $USER
# a temp solution from https://github.com/apache/pulsar/pull/15846 to fix python protobuf version error
RUN pip3 install protobuf==3.20.2 --user
# to make the python runner could print json logs
RUN pip3 install python-json-logger --user
