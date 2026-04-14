#!/usr/bin/env bash

set -eo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[1;34m'
NC='\033[0m'

if [ -z "${GRAZHDA_DIR:-}" ]; then
    GRAZHDA_DIR="$HOME/.grazhda"
    GRAZHDA_DIR_DISPLAY="\$HOME/.grazhda"
else
    GRAZHDA_DIR_DISPLAY="$GRAZHDA_DIR"
fi
export GRAZHDA_DIR

LOG_FILE="$GRAZHDA_DIR/logs/install.log"

if [[ -f "${HOME}/.bashrc.user" ]]; then
    bashrc_path="${HOME}/.bashrc.user"
else
    bashrc_path="${HOME}/.bashrc"
fi

grazhda_init_snippet=$( cat << EOF
export GRAZHDA_DIR="$GRAZHDA_DIR_DISPLAY"
[[ -s "${GRAZHDA_DIR_DISPLAY}/bin/grazhda-init.sh" ]] && source "${GRAZHDA_DIR_DISPLAY}/bin/grazhda-init.sh"
EOF
)

# ---------------------------------------------------------------------------
# Logging
# ---------------------------------------------------------------------------

setup_logging() {
    mkdir -p "$GRAZHDA_DIR/logs"
    : > "$LOG_FILE"
    _log "=== Grazhda Installation Started: $(date) ==="
    _log "GRAZHDA_DIR=$GRAZHDA_DIR"
    echo -e "${BLUE}Install log: $LOG_FILE${NC}"
}

_log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" >> "$LOG_FILE"
}

# Run a command, sending all stdout/stderr to the log file.
# Prints a one-line status to stdout; on failure shows last 20 log lines.
run_logged() {
    local desc="$1"
    shift
    _log ">>> $desc"
    _log "    Command: $*"
    if ! "$@" >> "$LOG_FILE" 2>&1; then
        echo -e "${RED}✗ $desc failed. Last 20 log lines:${NC}" >&2
        tail -20 "$LOG_FILE" >&2
        echo -e "${YELLOW}Full log: $LOG_FILE${NC}" >&2
        exit 1
    fi
    _log "<<< $desc: OK"
}

on_error() {
    local lineno="${1:-?}"
    echo -e "${RED}✗ Unexpected error at line $lineno.${NC}" >&2
    if [[ -f "$LOG_FILE" ]]; then
        echo -e "${YELLOW}Last log entries:${NC}" >&2
        tail -20 "$LOG_FILE" >&2
        echo -e "${YELLOW}Full log: $LOG_FILE${NC}" >&2
    fi
}
trap 'on_error $LINENO' ERR

# ---------------------------------------------------------------------------

check_binary() {
    local binary=$1
    local install_hint=$2

    if ! command -v "$binary" &> /dev/null; then
        echo -e "${RED}  ✗ '$binary' is not installed.${NC}"
        if [ -n "$install_hint" ]; then
            echo -e "${YELLOW}    $install_hint${NC}"
        fi
        _log "MISSING: $binary"
        return 1
    fi
    _log "Found: $binary ($(command -v "$binary"))"
}

verify_requirements() {
    echo -e "${BLUE}Checking system requirements...${NC}"
    _log "=== Verifying requirements ==="
    local failed=0

    check_binary "git"    "Install from: https://git-scm.com/"                               || failed=1
    check_binary "go"     "Install from: https://golang.org/dl/"                             || failed=1
    check_binary "just"   "Install from: https://github.com/casey/just/releases"             || failed=1
    check_binary "protoc" "Install from: https://github.com/protocolbuffers/protobuf/releases" || failed=1

    if [ $failed -eq 1 ]; then
        echo -e "${RED}Please install all missing dependencies and try again.${NC}"
        exit 1
    fi

    echo -e "${GREEN}✓ All required binaries found${NC}"
    echo ""
}

install_from_sources() {
    # Ensure GOPATH/bin is on PATH so protoc plugins installed by just generate are reachable.
    export PATH="$(go env GOPATH)/bin:$PATH"

    echo -e "${BLUE}Building from sources (this may take a minute)...${NC}"
    _log "=== Building from sources ==="
    # just build: installs protoc plugins, generates proto, compiles zgard + dukh, copies scripts.
    run_logged "Build (just build)" just build

    mkdir -p "$GRAZHDA_DIR/bin"
    run_logged "Copy binaries" cp bin/* "$GRAZHDA_DIR/bin/"
    run_logged "Set permissions" chmod +x "$GRAZHDA_DIR/bin/"*

    _log "Installed binaries:"
    ls -lah "$GRAZHDA_DIR/bin/" >> "$LOG_FILE" 2>&1

    echo -e "${GREEN}✓ Grazhda built and installed to $GRAZHDA_DIR/bin/${NC}"
}

install_from_github() {
    local repo_url="https://github.com/vhula/grazhda.git"
    local source_dir="$GRAZHDA_DIR/sources"

    mkdir -p "$GRAZHDA_DIR"
    echo -e "${BLUE}Cloning Grazhda repository...${NC}"
    _log "=== Cloning from $repo_url ==="
    run_logged "Clone repository" git clone "$repo_url" "$source_dir"

    cd "$source_dir"
    install_from_sources
}

install_from_local() {
    local source_dir="${1:-.}"
    local target_dir="$GRAZHDA_DIR/sources"

    mkdir -p "$GRAZHDA_DIR"
    echo -e "${BLUE}Copying from local directory: $source_dir${NC}"
    _log "=== Copying from local directory: $source_dir ==="
    mkdir -p "$target_dir"
    run_logged "Copy sources" cp -R "$source_dir/." "$target_dir/"

    cd "$target_dir"
    install_from_sources
}

create_config() {
    local config_file="$GRAZHDA_DIR/config.yaml"

    if [ -f "$config_file" ]; then
        echo -e "${YELLOW}⚠ Configuration already exists: $config_file${NC}"
        _log "Config already exists, skipping: $config_file"
        return
    fi

    echo -e "${BLUE}Creating configuration...${NC}"
    _log "=== Creating config: $config_file ==="

    cp "$GRAZHDA_DIR/sources/config.template.yaml" "$config_file"

    if [[ "$(uname)" == "Darwin" ]]; then
        sed -i '' "s|\${GRAZHDA_DIR}|$GRAZHDA_DIR|g" "$config_file"
    else
        sed -i "s|\${GRAZHDA_DIR}|$GRAZHDA_DIR|g" "$config_file"
    fi

    _log "Config created: $config_file"
    echo -e "${GREEN}✓ Configuration file created: $config_file${NC}"
}

copy_pkgs_registry() {
    local pkgs_file="$GRAZHDA_DIR/.grazhda.pkgs.yaml"

    echo -e "${BLUE}Copying package registry...${NC}"
    _log "=== Copying package registry: $pkgs_file ==="

    cp "$GRAZHDA_DIR/sources/.grazhda.pkgs.yaml" "$pkgs_file"

    _log "Package registry copied: $pkgs_file"
    echo -e "${GREEN}✓ Package registry copied: $pkgs_file${NC}"
}

main() {
    echo ""
    echo -e "${BLUE}╔═══════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║          Grazhda Installer            ║${NC}"
    echo -e "${BLUE}╚═══════════════════════════════════════╝${NC}"
    echo ""

    setup_logging
    echo ""

    verify_requirements

    echo -e "${BLUE}Installation directory: $GRAZHDA_DIR_DISPLAY${NC}"
    _log "Installation directory: $GRAZHDA_DIR"
    echo ""

    # Check for an existing install (bin/ or sources/); the logs/ dir is created by setup_logging
    # so we cannot use [ -d "$GRAZHDA_DIR" ] as the indicator anymore.
    if [ -d "$GRAZHDA_DIR/bin" ] || [ -d "$GRAZHDA_DIR/sources" ]; then
        echo -e "${YELLOW}Existing Grazhda installation found at: $GRAZHDA_DIR${NC}"
        echo ""
        read -p "Do you want to reinstall? (y/n) " -r </dev/tty
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo -e "${GREEN}Installation cancelled.${NC}"
            _log "Installation cancelled by user."
            exit 0
        fi
        echo "Removing existing installation..."
        _log "Removing: $GRAZHDA_DIR/sources and $GRAZHDA_DIR/bin"
        rm -rf "$GRAZHDA_DIR/sources" "${GRAZHDA_DIR:?}/bin"
    fi

    if [ -n "${LOCAL_GRAZHDA_REPO_DIR:-}" ]; then
        install_from_local "$LOCAL_GRAZHDA_REPO_DIR"
    else
        install_from_github
    fi

    create_config

    copy_pkgs_registry

    if ! grep -q 'grazhda-init.sh' "$bashrc_path"; then
        printf '\n%s\n' "$grazhda_init_snippet" >> "$bashrc_path"
        echo "Added Grazhda init snippet to $bashrc_path"
        _log "Appended init snippet to $bashrc_path"
    else
        _log "Init snippet already present in $bashrc_path"
    fi

    _log "=== Installation Completed: $(date) ==="

    echo ""
    echo -e "${GREEN}╔═══════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║      Installation Successful!         ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${BLUE}Grazhda installed to: $GRAZHDA_DIR_DISPLAY${NC}"
    echo -e "${BLUE}Sources:              $GRAZHDA_DIR_DISPLAY/sources${NC}"
    echo -e "${BLUE}Binaries:             $GRAZHDA_DIR_DISPLAY/bin${NC}"
    echo -e "${BLUE}Configuration:        $GRAZHDA_DIR_DISPLAY/config.yaml${NC}"
    echo -e "${BLUE}Package registry:     $GRAZHDA_DIR_DISPLAY/.grazhda.pkgs.yaml${NC}"
    echo -e "${BLUE}Install log:          $GRAZHDA_DIR_DISPLAY/logs/install.log${NC}"
    echo ""
    echo "To get started:"
    echo -e "  ${YELLOW}source \"$GRAZHDA_DIR/bin/grazhda-init.sh\"${NC}"
    echo ""
    echo "For more information, visit:"
    echo "  https://github.com/vhula/grazhda"
    echo ""
}

main
