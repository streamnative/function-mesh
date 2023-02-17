FROM streamnative/pulsar-all:2.9.2.23 as pulsar
FROM streamnative/pulsar-functions-runner-base:2.9.2.23

COPY --from=pulsar --chown=$UID:$GID /pulsar/README /pulsar/pulsar-client /tmp/pulsar/
RUN if [ -d "/tmp/pulsar/pulsar-client" ]; then mv /tmp/pulsar/pulsar-client /pulsar/pulsar-client || true ; fi

RUN ls -al /tmp/pulsar
