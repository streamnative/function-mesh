ARG PULSAR_IMAGE_TAG
FROM --platform=$BUILDPLATFORM golang:1.25.9-trixie as builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY func/go.mod go.mod
COPY func/go.sum go.sum
RUN go mod download

# Copy the go source
COPY func/exclamationFunc.go exclamationFunc.go

# Build
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} GO111MODULE=on go build -a -o exclamationFunc exclamationFunc.go

FROM streamnative/pulsar-functions-go-runner:${PULSAR_IMAGE_TAG}
COPY --from=builder --chown=$UID:$GID /workspace/exclamationFunc /pulsar/examples/go-exclamation-func
