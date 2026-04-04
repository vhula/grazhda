#!/usr/bin/env bash

if [ -z "$GRAZHDA_DIR" ]; then
    export GRAZHDA_DIR="$HOME/.grazhda"
fi

if [[ ":$PATH:" != *":$GRAZHDA_DIR/bin:"* ]]; then
    export PATH="$GRAZHDA_DIR/bin:$PATH"
fi