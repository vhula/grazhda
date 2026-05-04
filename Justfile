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
    echo "  test-bash      - Run bash script tests"
    echo "  fmt            - Format Go source across all modules"
    echo "  tidy           - Tidy all modules"
    echo "  man            - Generate man pages into man/man1/"
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

build: generate build-zgard build-dukh copy-scripts copy-configs
    echo "✓ All modules built successfully"

build-zgard:
    echo "Building Zgard..."
    mkdir -p bin
    cd zgard && go build -ldflags "-X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" -o ../bin/zgard .
    echo "✓ bin/zgard built"

build-dukh:
    echo "Building Dukh..."
    mkdir -p bin
    cd dukh && go build -ldflags "-X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" -o ../bin/dukh ./cmd
    echo "✓ bin/dukh built"

copy-scripts:
    echo "Copying bash scripts..."
    cp grazhda ./bin/
    cp grazhda-init.sh ./bin/
    chmod +x ./bin/grazhda ./bin/grazhda-init.sh
    echo "✓ Scripts copied to bin/"

copy-configs:
    echo "Copying config files..."
    cp .grazhda.env ./bin/
    cp .grazhda.pkgs.yaml ./bin/
    echo "✓ Config files copied to bin/"

clean:
    echo "Cleaning up..."
    rm -rf bin
    echo "✓ Clean complete"

test:
    echo "Running tests..."
    cd internal && go test ./...
    cd zgard && go test ./...
    cd dukh && go test ./...
    bash tests/bash/test_scripts.sh
    echo "✓ Tests passed"

test-bash:
    echo "Running bash script tests..."
    bash tests/bash/test_scripts.sh
    echo "✓ Bash script tests passed"

tidy:
    echo "Tidying modules..."
    cd internal && go mod tidy
    cd zgard && go mod tidy
    cd dukh && go mod tidy
    echo "✓ Go modules tidied"

fmt:
    echo "Formatting Go modules..."
    cd internal && go fmt ./...
    cd zgard && go fmt ./...
    cd dukh && go fmt ./...
    echo "✓ Go modules formatted"

man:
    echo "Generating man pages..."
    mkdir -p man/man1
    cd zgard && go run ../tools/gen-manpages/main.go
    echo "✓ Man pages written to man/man1/"
