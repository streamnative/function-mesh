ARG PULSAR_IMAGE_TAG
FROM golang:1.18 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY func/go.mod go.mod
COPY func/go.sum go.sum
RUN go mod download

# Copy the go source
COPY func/exclamationFunc.go exclamationFunc.go

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o exclamationFunc exclamationFunc.go

FROM streamnative/pulsar-functions-go-runner:${PULSAR_IMAGE_TAG}
COPY --from=builder --chown=$UID:$GID /workspace/exclamationFunc /pulsar/examples/go-exclamation-func
