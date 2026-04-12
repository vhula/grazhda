#!/usr/bin/env bash

if [ -z "$GRAZHDA_DIR" ]; then
    export GRAZHDA_DIR="$HOME/.grazhda"
fi

case ":$PATH:" in
    *":$GRAZHDA_DIR/bin:"*) ;;
    *) export PATH="$GRAZHDA_DIR/bin:$PATH" ;;
esac
mkdir -p "$GRAZHDA_DIR/pkgs"
# Source legacy env file (backward compat) then the canonical pkg env file.
source "$GRAZHDA_DIR/grazhda-env.sh" 2>/dev/null || true
source "$GRAZHDA_DIR/.grazhda.env" 2>/dev/null || true