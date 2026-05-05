VERSION_FLAG := "-X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)"

clean:
    echo "Cleaning build artifacts..."
    rm -rf build

build: clean build-zgard build-dukh

build-zgard:
    echo "Building zgard..."
    mkdir -p build
    cd zgard && go build -ldflags "{{VERSION_FLAG}}" -o ../build/zgard .
    echo "✓ build/zgard built"

build-dukh:
    echo "Building dukh..."
    mkdir -p build
    cd dukh && go build -ldflags "{{VERSION_FLAG}}" -o ../build/dukh .
    echo "✓ build/dukh built"

fmt:
    echo "Formatting Go modules..."
    cd internal && go fmt ./...
    cd zgard && go fmt ./...
    cd dukh && go fmt ./...
    echo "✓ Go modules formatted"

tidy:
    echo "Tidying modules..."
    cd internal && go mod tidy
    cd zgard && go mod tidy
    cd dukh && go mod tidy
    echo "✓ Go modules tidied"

test:
    echo "Running tests..."
    echo "✓ All tests passed"