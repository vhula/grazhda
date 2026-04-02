#!/usr/bin/env just --justfile

default: help

help:
    echo "Grazhda Monorepo Build System"
    echo ""
    echo "Usage: just [task]"
    echo ""
    echo "Tasks:"
    echo "  build          - Build all Go modules (dukh and zgard)"
    echo "  build-dukh     - Build only Dukh CLI"
    echo "  build-zgard    - Build only Zgard CLI"
    echo "  copy-scripts   - Copy bash scripts to bin directory"
    echo "  clean          - Remove all built binaries"
    echo "  test           - Run tests for all modules"
    echo "  zip            - Create a zip file with the built project"
    echo "  help           - Show this help message"

build: build-dukh build-zgard
    echo "✓ All modules built successfully"

build-dukh:
    echo "Building Dukh..."
    mkdir -p bin
    cd dukh && go build -o ../bin/dukh

build-zgard:
    echo "Building Zgard..."
    mkdir -p bin
    cd zgard && go build -o ../bin/zgard

copy-scripts:
    echo "Copying bash scripts..."
    mkdir -p bin
    cp install.sh bin/
    cp grazhda-init.sh bin/
    echo "✓ Scripts copied to bin/"

clean:
    echo "Cleaning up..."
    rm -rf bin
    echo "✓ Clean complete"

test:
    echo "Running tests..."
    cd dukh && go test ./...
    cd zgard && go test ./...
    echo "✓ Tests passed"

zip: build copy-scripts
    echo "Creating zip archive..."
    mkdir -p bin
    cd bin && zip -r grazhda.zip * && cd ..
    echo "✓ Zip created: bin/grazhda.zip"
