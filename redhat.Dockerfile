# Build the manager binary
FROM golang:1.18 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM registry.access.redhat.com/ubi8/ubi-micro:latest

ARG VERSION

LABEL name="function-mesh-operator" \
      vendor="StreamNative, Inc." \
      maintainer="StreamNative, Inc." \
      version="${VERSION}" \
      release="${VERSION}" \
	  summary="Function Mesh Operator is a Kubernetes operator that enables users to run Pulsar Functions and Pulsar connectors natively on Kubernetes." \
	  description="By providing a serverless framework that enables users to organize a collection of Pulsar Functions and connectors, Function Mesh simplifies the process of creating complex streaming jobs. Function Mesh is a valuable tool for users who are seeking cloud-native serverless streaming solutions"

WORKDIR /
COPY --from=builder /workspace/manager .
COPY LICENSE /licenses/LICENSE
USER 1001

ENTRYPOINT ["/manager"]
