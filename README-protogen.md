# Protobuf Generation Docker Environment

This Docker environment contains the tools and environment required to generate function-mesh protobuf Go definitions.

## Requirements

- protoc: v3.17.3
- protoc-gen-go: v1.25.0
- protoc-gen-go-grpc: v1.0.0
- Go: 1.24.4+

**Note**: This image uses a Debian base (not Alpine) to ensure compatibility with the protoc x86_64 binary, which is compiled for glibc rather than musl libc.

## Build and Usage

### 1. Build Docker Image

```bash
docker build -f Dockerfile.protogen -t function-mesh-protogen .
```

### 2. Run Protobuf Generation

```bash
# Run container and generate protobuf files
docker run --rm -v $(pwd):/workspace function-mesh-protogen

# Or run interactively for debugging
docker run --rm -it -v $(pwd):/workspace function-mesh-protogen bash
```

### 3. Manual Script Execution

If you need manual control over the generation process:

```bash
# Enter container
docker run --rm -it -v $(pwd):/workspace function-mesh-protogen bash

# Execute manually inside container
cd /workspace
./controllers/proto/generate.sh /workspace
```

## Generated Files

The script will update the following files:
- `controllers/proto/Function.pb.go`

## Environment Verification

You can verify tool versions inside the container:

```bash
protoc --version              # Should show libprotoc 3.17.3
go version                   # Should show go1.24.4+
```

## Troubleshooting

If you encounter permission issues, ensure the generate.sh script has execute permissions:

```bash
chmod +x controllers/proto/generate.sh
```

If you encounter protoc-gen-go not found issues, ensure $GOPATH/bin is in PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
``` 