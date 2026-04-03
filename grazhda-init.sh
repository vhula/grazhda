#!/usr/bin/env bash
set -e

echo 'Add Grazhda to PATH...'

if [ -z "$GRAZHDA_DIR" ]; then
    export GRAZHDA_DIR="$HOME/.grazhda"
fi

if [[ ":$PATH:" != *":$GRAZHDA_DIR/bin:"* ]]; then
    export PATH="$GRAZHDA_DIR/bin:$PATH"
    echo "Grazhda added to PATH: $GRAZHDA_DIR/bin"
else
    echo "Grazhda is already in PATH."
fi