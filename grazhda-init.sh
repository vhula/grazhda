#!/usr/bin/env bash

if [ -z "$GRAZHDA_DIR" ]; then
    export GRAZHDA_DIR="$HOME/.grazhda"
fi

case ":$PATH:" in
    *":$GRAZHDA_DIR/bin:"*) ;;
    *) export PATH="$GRAZHDA_DIR/bin:$PATH" ;;
esac