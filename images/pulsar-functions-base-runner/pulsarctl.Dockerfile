FROM alpine:3.21 as functions-runner

ARG TARGETARCH

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
    && case "${TARGETARCH}" in \
         amd64|arm64) PULSARCTL_ARCH="${TARGETARCH}" ;; \
         *) echo "unsupported TARGETARCH: ${TARGETARCH}" && exit 1 ;; \
       esac \
    && PULSARCTL_DIST="pulsarctl-${PULSARCTL_ARCH}-linux.tar.gz" \
    && wget "https://github.com/streamnative/pulsarctl/releases/latest/download/${PULSARCTL_DIST}" -P /pulsar/bin/ \
    && tar -xzf "/pulsar/bin/${PULSARCTL_DIST}" -C /pulsar/bin/ \
    && rm -rf "/pulsar/bin/${PULSARCTL_DIST}" \
    && chmod +x "/pulsar/bin/pulsarctl-${PULSARCTL_ARCH}-linux/pulsarctl" \
    && ln -s "/pulsar/bin/pulsarctl-${PULSARCTL_ARCH}-linux/pulsarctl" /usr/local/bin/pulsarctl

WORKDIR /pulsar
