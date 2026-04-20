FROM --platform=$BUILDPLATFORM golang:1.25.9-trixie AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace/api
COPY api/ .
RUN go mod download

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY main.go main.go
COPY pkg/ pkg/
COPY controllers/ controllers/
COPY utils/ utils/

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} GO111MODULE=on go build -a -trimpath -o /workspace/manager main.go

FROM alpine:3.21

ENV GID=10001
ENV UID=10000
ENV USER=pulsar

RUN apk upgrade --no-cache \
    && apk add tzdata --no-cache \
    && addgroup -g $GID pulsar \
    && adduser -u $UID -G pulsar -D -g '' $USER

COPY --from=builder /workspace/manager /manager

USER $USER
