#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TMP_DIRS=()

cleanup() {
    local d
    for d in "${TMP_DIRS[@]:-}"; do
        rm -rf "$d"
    done
}
trap cleanup EXIT

fail() {
    echo "FAIL: $*" >&2
    exit 1
}

assert_eq() {
    local got="$1"
    local want="$2"
    local msg="${3:-values differ}"
    [[ "$got" == "$want" ]] || fail "$msg (got=$got want=$want)"
}

assert_file_contains() {
    local file="$1"
    local needle="$2"
    grep -Fq "$needle" "$file" || fail "expected $file to contain: $needle"
}

assert_exists() {
    local path="$1"
    [[ -e "$path" ]] || fail "expected path to exist: $path"
}

test_grazhda_init_path_and_pkgs_dir() {
    local tmp
    tmp="$(mktemp -d)"
    TMP_DIRS+=("$tmp")

    export HOME="$tmp/home"
    export GRAZHDA_DIR="$tmp/.grazhda"
    mkdir -p "$HOME"
    export PATH="/usr/bin:/bin"

    # shellcheck source=/dev/null
    source "$ROOT_DIR/grazhda-init.sh"
    # shellcheck source=/dev/null
    source "$ROOT_DIR/grazhda-init.sh"

    assert_exists "$GRAZHDA_DIR/pkgs"
    local count
    count="$(tr ':' '\n' <<<"$PATH" | grep -Fx "$GRAZHDA_DIR/bin" | wc -l | tr -d ' ')"
    assert_eq "$count" "1" "grazhda-init.sh should prepend bin path once"
}

test_installer_copy_pkgs_registry_overwrites() {
    local tmp
    tmp="$(mktemp -d)"
    TMP_DIRS+=("$tmp")

    export HOME="$tmp/home"
    export GRAZHDA_DIR="$tmp/.grazhda"
    mkdir -p "$HOME" "$GRAZHDA_DIR/sources"

    # shellcheck source=/dev/null
    source "$ROOT_DIR/grazhda-install.sh"
    LOG_FILE="$tmp/install.log"
    : > "$LOG_FILE"

    echo "global:new" > "$GRAZHDA_DIR/sources/.grazhda.pkgs.yaml"
    echo "old:value" > "$GRAZHDA_DIR/.grazhda.pkgs.yaml"

    copy_pkgs_registry
    assert_file_contains "$GRAZHDA_DIR/.grazhda.pkgs.yaml" "global:new"
}

test_installer_create_config_creates_and_keeps_existing() {
    local tmp
    tmp="$(mktemp -d)"
    TMP_DIRS+=("$tmp")

    export HOME="$tmp/home"
    export GRAZHDA_DIR="$tmp/.grazhda"
    mkdir -p "$HOME" "$GRAZHDA_DIR/sources"

    # shellcheck source=/dev/null
    source "$ROOT_DIR/grazhda-install.sh"
    LOG_FILE="$tmp/install.log"
    : > "$LOG_FILE"

    cat > "$GRAZHDA_DIR/sources/config.template.yaml" <<'EOF'
general:
  install_dir: ${GRAZHDA_DIR}
EOF

    create_config
    assert_file_contains "$GRAZHDA_DIR/config.yaml" "$GRAZHDA_DIR"

    echo "custom: true" > "$GRAZHDA_DIR/config.yaml"
    create_config
    assert_file_contains "$GRAZHDA_DIR/config.yaml" "custom: true"
}

test_upgrade_updates_registry_from_sources() {
    local tmp
    tmp="$(mktemp -d)"
    TMP_DIRS+=("$tmp")

    export HOME="$tmp/home"
    export GRAZHDA_DIR="$tmp/.grazhda"
    mkdir -p "$HOME" "$GRAZHDA_DIR/sources/.git" "$GRAZHDA_DIR/sources/bin" "$GRAZHDA_DIR/bin" "$tmp/fakebin"

    cat > "$tmp/fakebin/git" <<'EOF'
#!/usr/bin/env bash
if [[ "$1" == "pull" ]]; then
  echo "Already up to date."
  exit 0
fi
exit 0
EOF
    cat > "$tmp/fakebin/go" <<'EOF'
#!/usr/bin/env bash
if [[ "$1" == "env" && "$2" == "GOPATH" ]]; then
  echo "/tmp"
  exit 0
fi
exit 0
EOF
    cat > "$tmp/fakebin/just" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
    cat > "$tmp/fakebin/protoc" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
    chmod +x "$tmp/fakebin/"*
    export PATH="$tmp/fakebin:/usr/bin:/bin"

    echo "bin" > "$GRAZHDA_DIR/sources/bin/zgard"
    echo "bin" > "$GRAZHDA_DIR/sources/bin/dukh"
    echo "bin" > "$GRAZHDA_DIR/sources/bin/grazhda"
    echo "bin" > "$GRAZHDA_DIR/sources/bin/grazhda-init.sh"
    echo "registry:new" > "$GRAZHDA_DIR/sources/.grazhda.pkgs.yaml"
    echo "registry:old" > "$GRAZHDA_DIR/.grazhda.pkgs.yaml"

    # shellcheck source=/dev/null
    source "$ROOT_DIR/grazhda"
    stop_dukh_if_running() { :; }

    cmd_upgrade

    assert_file_contains "$GRAZHDA_DIR/.grazhda.pkgs.yaml" "registry:new"
    assert_exists "$GRAZHDA_DIR/bin/zgard"
}

run_all() {
    test_grazhda_init_path_and_pkgs_dir
    test_installer_copy_pkgs_registry_overwrites
    test_installer_create_config_creates_and_keeps_existing
    test_upgrade_updates_registry_from_sources
    echo "PASS: bash script tests"
}

run_all
