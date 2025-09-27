#!/bin/bash

set -e

cd "$(dirname "$0")/.."

echo "generating protobuf code..."

# ensure protoc and plugins are available
if ! command -v protoc &> /dev/null; then
    echo "error: protoc not found"
    exit 1
fi

export PATH=$PATH:$(go env GOPATH)/bin

# generate go code
protoc \
  --go_out=server/internal/gen \
  --go-grpc_out=server/internal/gen \
  --go_opt=paths=source_relative \
  --go-grpc_opt=paths=source_relative \
  shared/proto/auth/v1/auth.proto \
  shared/proto/messaging/v1/messaging.proto \
  shared/proto/storage/v1/storage.proto

echo "protobuf code generation complete"
