#!/usr/bin/env just --justfile

default: help

help:
    echo "Grazhda Monorepo Build System"
    echo ""
    echo "Usage: just [task]"
    echo ""
    echo "Tasks:"
    echo "  install        - Clone git repo and build from sources"
    echo "  build          - Build all Go modules (dukh and zgard)"
    echo "  build-dukh     - Build only Dukh CLI"
    echo "  build-zgard    - Build only Zgard CLI"
    echo "  copy-scripts   - Copy bash scripts to bin directory"
    echo "  generate       - Generate protobuf code for Dukh services"
    echo "  clean          - Remove all built binaries"
    echo "  test           - Run tests for all modules"
    echo "  zip            - Create a zip file with the built project"
    echo "  help           - Show this help message"

build: build-dukh build-zgard copy-scripts
    echo "✓ All modules built successfully"

build-dukh: generate
    echo "Building Dukh..."
    mkdir -p ../bin
    cd dukh/cmd && go build -o ../../bin/dukh

build-zgard:
    echo "Building Zgard..."
    mkdir -p ../bin
    cd zgard && go build -o ../bin/zgard

copy-scripts:
    echo "Copying bash scripts..."
    cp grazhda ./bin/
    cp grazhda-init.sh ./bin/
    echo "✓ Scripts copied to ../bin/"

generate:
    echo "Generating protobuf code..."
    cd proto && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    cd proto && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    cd proto && protoc --go_out=../internal/proto --go_opt=paths=source_relative --go-grpc_out=../internal/proto --go-grpc_opt=paths=source_relative workspace.proto
    cd proto && protoc --go_out=../internal/proto --go_opt=paths=source_relative --go-grpc_out=../internal/proto --go-grpc_opt=paths=source_relative dukh.proto
    echo "✓ Protobuf code generated"

clean:
    echo "Cleaning up..."
    rm -rf bin
    echo "✓ Clean complete"

test:
    echo "Running tests..."
    cd internal && go test ./...
    cd dukh && go test ./...
    cd zgard && go test ./...
    echo "✓ Tests passed"

tidy:
    echo "Tidying Go modules..."
    cd internal && go mod tidy
    cd dukh && go mod tidy
    cd zgard && go mod tidy
    echo "✓ Go modules tidied"

fmt:
    echo "Formatting Go modules..."
    cd internal && go fmt ./...
    cd dukh && go fmt ./...
    cd zgard && go fmt ./...
    echo "✓ Go modules formatted"