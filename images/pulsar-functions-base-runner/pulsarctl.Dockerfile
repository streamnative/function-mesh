FROM alpine:3.21 as functions-runner

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
    && chmod -R g=u /pulsar \
    && apk update && apk add --no-cache wget bash \
         && wget https://github.com/streamnative/pulsarctl/releases/latest/download/pulsarctl-amd64-linux.tar.gz -P /pulsar/bin/ \
         && tar -xzf /pulsar/bin/pulsarctl-amd64-linux.tar.gz -C /pulsar/bin/ \
         && rm -rf /pulsar/bin/pulsarctl-amd64-linux.tar.gz \
         && chmod +x /pulsar/bin/pulsarctl-amd64-linux/pulsarctl \
         && ln -s /pulsar/bin/pulsarctl-amd64-linux/pulsarctl /usr/local/bin/pulsarctl

WORKDIR /pulsar
