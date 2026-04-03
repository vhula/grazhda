#!/usr/bin/env bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

if [ -z "$GRAZHDA_DIR" ]; then
    GRAZHDA_DIR="$HOME/.grazhda"
    GRAZHDA_DIR_DISPLAY="\$HOME/.grazhda"
else
    GRAZHDA_DIR_DISPLAY="$GRAZHDA_DIR"
fi
export GRAZHDA_DIR

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

check_binary() {
    local binary=$1
    local install_hint=$2

    if ! command -v "$binary" &> /dev/null; then
        echo -e "${RED}Error: '$binary' is not installed.${NC}"
        if [ -n "$install_hint" ]; then
            echo -e "${YELLOW}$install_hint${NC}"
        fi
        return 1
    fi
}

verify_requirements() {
    echo -e "${BLUE}Checking system requirements...${NC}"
    local failed=0

    check_binary "git" "Install from: https://git-scm.com/" || failed=1
    check_binary "go" "Install from: https://golang.org/dl/" || failed=1
    check_binary "just" "Install from: https://github.com/casey/just/releases" || failed=1
    check_binary "protoc" "Install from: https://github.com/protocolbuffers/protobuf/releases" || failed=1

    if [ $failed -eq 1 ]; then
        echo -e "${RED}Please install all missing dependencies and try again.${NC}"
        exit 1
    fi

    echo -e "${GREEN}✓ All required binaries found${NC}"
    echo ""
}

install_from_sources() {
    just build

    cp bin/* "$GRAZHDA_DIR/bin/"

    echo -e "${GREEN}✓ Grazhda built successfully${NC}"
    echo "Binaries are located in: $GRAZHDA_DIR/bin/"
    ls -lah "$GRAZHDA_DIR/bin/"
}

install_from_github() {
    local repo_url="https://github.com/vhula/grazhda.git"
    local source_dir="$GRAZHDA_DIR/sources"

    mkdir -p "$GRAZHDA_DIR"
    echo -e "${BLUE}Cloning Grazhda repository from $repo_url...${NC}"
    git clone "$repo_url" "$source_dir"

    echo -e "${BLUE}Building from sources...${NC}"
    cd "$source_dir"

    install_from_sources
}

install_from_local() {
    local source_dir="${1:-.}"
    local target_dir="$GRAZHDA_DIR/sources"

    mkdir -p "$GRAZHDA_DIR"
    echo -e "${BLUE}Copying Grazhda from local directory...${NC}"
    mkdir -p "$target_dir"
    cp -R "$source_dir"/* "$target_dir"

    echo -e "${BLUE}Building from sources...${NC}"
    cd "$target_dir"

    install_from_sources
}

create_config() {
    local config_file="$GRAZHDA_DIR/config.yaml"

    echo -e "${BLUE}Creating configuration file...${NC}"

    cp "$GRAZHDA_DIR"/sources/config.template.yaml "$config_file"

    sed -i "s|\${GRAZHDA_DIR}|$GRAZHDA_DIR|g" "$config_file"

    echo -e "${GREEN}✓ Configuration file created: $config_file${NC}"
}

main() {
    echo ""
    echo -e "${BLUE}╔═══════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║          Grazhda Installer            ║${NC}"
    echo -e "${BLUE}╚═══════════════════════════════════════╝${NC}"
    echo ""

    verify_requirements

    echo -e "${BLUE}Installation Directory: $GRAZHDA_DIR_DISPLAY${NC}"
    echo ""

    if [ -d "$GRAZHDA_DIR/bin" ]; then
        echo -e "${YELLOW}Existing Grazhda installation found at: $GRAZHDA_DIR${NC}"
        echo ""
        read -p "Do you want to reinstall? (y/n) " -r </dev/tty
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo -e "${GREEN}Installation cancelled.${NC}"
            exit 0
        fi
        echo "Removing existing installation..."
        rm -rf "$GRAZHDA_DIR/sources" "${GRAZHDA_DIR:?}/bin"
    fi

    if [ -n "$LOCAL_REPO_PATH" ]; then
        install_from_local "$LOCAL_REPO_PATH"
    else
        install_from_github
    fi

    create_config

    if [[ -z $(grep 'grazhda-init.sh' "$bashrc_path") ]]; then
      echo -e "\n$grazhda_init_snippet" >> "$bashrc_path"
      echo "Added grazhda init snippet to $bashrc_path"
    fi

    echo ""
    echo -e "${GREEN}╔═══════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║      Installation Successful!         ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${BLUE}Grazhda installed to: $GRAZHDA_DIR_DISPLAY${NC}"
    echo -e "${BLUE}Sources directory: $GRAZHDA_DIR_DISPLAY/sources${NC}"
    echo -e "${BLUE}Binaries: $GRAZHDA_DIR_DISPLAY/bin${NC}"
    echo -e "${BLUE}Configuration: $GRAZHDA_DIR_DISPLAY/config.yaml${NC}"
    echo ""
    echo "To get started:"
    echo -e "  ${YELLOW}source \"$GRAZHDA_DIR/bin/grazhda-init.sh\"${NC}"
    echo ""
    echo "For more information, visit:"
    echo "  https://github.com/vhula/grazhda"
    echo ""
}

main
