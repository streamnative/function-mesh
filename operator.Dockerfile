FROM alpine:3.20.3

ENV GID=10001
ENV UID=10000
ENV USER=pulsar

RUN apk upgrade --no-cache \
    && apk add tzdata --no-cache \
    && addgroup -g $GID pulsar \
    && adduser -u $UID -G pulsar -D -g '' $USER

ADD bin/function-mesh-controller-manager /manager

USER $USER
