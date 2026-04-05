#!/usr/bin/env just --justfile

default: help

help:
    echo "Grazhda Monorepo Build System"
    echo ""
    echo "Usage: just [task]"
    echo ""
    echo "Tasks:"
    echo "  build          - Build all binaries (zgard + scripts)"
    echo "  build-zgard    - Build only Zgard CLI"
    echo "  copy-scripts   - Copy bash scripts to bin directory"
    echo "  clean          - Remove all built binaries"
    echo "  test           - Run tests for all modules"
    echo "  fmt            - Format Go source across all modules"
    echo "  tidy           - Sync workspace and tidy all modules"
    echo "  help           - Show this help message"

build: build-zgard copy-scripts
    echo "✓ All modules built successfully"

build-zgard:
    echo "Building Zgard..."
    mkdir -p bin
    cd zgard && go build -o ../bin/zgard .
    echo "✓ bin/zgard built"

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
    cd cmd && go test ./...
    cd internal && go test ./...
    cd zgard && go test ./...
    echo "✓ Tests passed"

tidy:
    echo "Syncing workspace and tidying modules..."
    go work sync
    cd cmd && go mod tidy
    cd internal && go mod tidy
    echo "✓ Go modules tidied"

fmt:
    echo "Formatting Go modules..."
    cd cmd && go fmt ./...
    cd internal && go fmt ./...
    cd zgard && go fmt ./...
    echo "✓ Go modules formatted"