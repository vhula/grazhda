#!/usr/bin/env just --justfile

default: help

help:
    echo "Grazhda Monorepo Build System"
    echo ""
    echo "Usage: just [task]"
    echo ""
    echo "Tasks:"
    echo "  generate       - Generate protobuf Go code from proto/dukh.proto"
    echo "  build          - Build all binaries (zgard + dukh + scripts)"
    echo "  build-zgard    - Build only Zgard CLI"
    echo "  build-dukh     - Build only Dukh server"
    echo "  copy-scripts   - Copy bash scripts to bin directory"
    echo "  clean          - Remove all built binaries"
    echo "  test           - Run tests for all modules"
    echo "  fmt            - Format Go source across all modules"
    echo "  tidy           - Tidy all modules"
    echo "  help           - Show this help message"

generate:
    echo "Installing protobuf Go plugins..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    echo "Generating protobuf code..."
    protoc \
      --go_out=dukh/proto --go_opt=paths=source_relative \
      --go-grpc_out=dukh/proto --go-grpc_opt=paths=source_relative \
      --proto_path=proto \
      proto/dukh.proto
    echo "✓ dukh/proto/ generated"

build: generate build-zgard build-dukh copy-scripts
    echo "✓ All modules built successfully"

build-zgard:
    echo "Building Zgard..."
    mkdir -p bin
    cd zgard && go build -o ../bin/zgard .
    echo "✓ bin/zgard built"

build-dukh:
    echo "Building Dukh..."
    mkdir -p bin
    cd dukh && go build -o ../bin/dukh ./cmd
    echo "✓ bin/dukh built"

copy-scripts:
    echo "Copying bash scripts..."
    cp grazhda ./bin/
    cp grazhda-init.sh ./bin/
    echo "✓ Scripts copied to bin/"

clean:
    echo "Cleaning up..."
    rm -rf bin
    echo "✓ Clean complete"

test:
    echo "Running tests..."
    cd internal && go test ./...
    cd zgard && go test ./...
    echo "✓ Tests passed"

tidy:
    echo "Tidying modules..."
    cd internal && go mod tidy
    cd zgard && go mod tidy -e
    cd dukh && go mod tidy -e
    echo "✓ Go modules tidied"

fmt:
    echo "Formatting Go modules..."
    cd internal && go fmt ./...
    cd zgard && go fmt ./...
    cd dukh && go fmt ./...
    echo "✓ Go modules formatted"