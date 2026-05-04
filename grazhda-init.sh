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
# Only source if readable to avoid shellcheck warnings and runtime errors.
# shellcheck disable=SC1091
[ -r "$GRAZHDA_DIR/grazhda-env.sh" ] && . "$GRAZHDA_DIR/grazhda-env.sh"
# shellcheck disable=SC1091
[ -r "$GRAZHDA_DIR/.grazhda.env" ] && . "$GRAZHDA_DIR/.grazhda.env"