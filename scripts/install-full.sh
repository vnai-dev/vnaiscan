#!/bin/bash
# vnaiscan full installer - installs vnaiscan + all required tools
# Usage: curl -sSL https://scan.vnai.dev/install.sh | sh
#
# This installs:
#   - vnaiscan (the scanner CLI)
#   - trivy (CVE/secrets scanner)
#   - malcontent (binary capabilities analyzer)
#   - magika (AI file type detection)

set -e

REPO="vnai-dev/vnaiscan"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Tool versions (updated by nightly workflow)
TRIVY_VERSION="${TRIVY_VERSION:-0.69.0}"
MALCONTENT_VERSION="${MALCONTENT_VERSION:-1.20.4}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

info() { echo -e "${CYAN}$1${NC}"; }
success() { echo -e "${GREEN}✓ $1${NC}"; }
warn() { echo -e "${YELLOW}⚠ $1${NC}"; }
error() { echo -e "${RED}✗ $1${NC}" >&2; }
step() { echo -e "${BOLD}→ $1${NC}"; }

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$ARCH" in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) error "Unsupported architecture: $ARCH"; exit 1 ;;
    esac

    case "$OS" in
        linux) OS="linux" ;;
        darwin) OS="darwin" ;;
        *) error "Unsupported OS: $OS"; exit 1 ;;
    esac

    PLATFORM="${OS}_${ARCH}"
}

# Check if command exists
has_cmd() {
    command -v "$1" &>/dev/null
}

# Install with sudo if needed
install_bin() {
    local src="$1"
    local dst="$2"
    if [ -w "$(dirname "$dst")" ]; then
        mv "$src" "$dst"
        chmod +x "$dst"
    else
        sudo mv "$src" "$dst"
        sudo chmod +x "$dst"
    fi
}

# Get latest GitHub release version
get_latest_version() {
    local repo="$1"
    curl -sSL "https://api.github.com/repos/${repo}/releases/latest" 2>/dev/null | \
        grep '"tag_name":' | sed -E 's/.*"v?([^"]+)".*/\1/' | head -1
}

# Install vnaiscan
install_vnaiscan() {
    step "Installing vnaiscan..."
    
    local version=$(get_latest_version "$REPO")
    [ -z "$version" ] && version="0.1.0"
    
    local url="https://github.com/${REPO}/releases/download/v${version}/vnaiscan_${version}_${PLATFORM}.tar.gz"
    local tmp=$(mktemp -d)
    trap "rm -rf $tmp" RETURN
    
    if curl -sSL "$url" | tar -xz -C "$tmp" 2>/dev/null; then
        install_bin "$tmp/vnaiscan" "${INSTALL_DIR}/vnaiscan"
        success "vnaiscan v${version}"
    else
        warn "vnaiscan binary not available, use Docker instead"
        return 1
    fi
}

# Install Trivy
install_trivy() {
    if has_cmd trivy; then
        success "trivy already installed ($(trivy --version 2>&1 | head -1))"
        return 0
    fi
    
    step "Installing Trivy v${TRIVY_VERSION}..."
    curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | \
        sh -s -- -b "${INSTALL_DIR}" "v${TRIVY_VERSION}" 2>/dev/null
    success "trivy v${TRIVY_VERSION}"
}

# Install Malcontent
install_malcontent() {
    if has_cmd mal; then
        success "malcontent already installed ($(mal --version 2>&1 | head -1))"
        return 0
    fi
    
    step "Installing Malcontent..."
    
    # Malcontent doesn't provide pre-built binaries, need to use Docker or build
    if has_cmd docker; then
        # Create a wrapper script that uses Docker
        local wrapper="${INSTALL_DIR}/mal"
        cat > /tmp/mal-wrapper << 'EOF'
#!/bin/bash
# Malcontent wrapper - runs via Docker
docker run --rm -v "$(pwd):/workspace" -w /workspace cgr.dev/chainguard/malcontent:latest "$@"
EOF
        install_bin /tmp/mal-wrapper "$wrapper"
        success "malcontent (via Docker wrapper)"
    else
        warn "malcontent requires Docker or manual build from source"
        warn "  → https://github.com/chainguard-dev/malcontent"
        return 1
    fi
}

# Install Magika
install_magika() {
    if has_cmd magika; then
        success "magika already installed ($(magika --version 2>&1 | head -1))"
        return 0
    fi
    
    step "Installing Magika..."
    
    # Check if Rust/cargo available
    if has_cmd cargo; then
        cargo install --locked magika-cli 2>/dev/null
        success "magika (built from source)"
    elif has_cmd brew; then
        # Try homebrew on macOS
        warn "Building magika requires Rust. Install with:"
        echo "    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"
        echo "    cargo install --locked magika-cli"
        return 1
    else
        warn "magika requires Rust toolchain to build"
        warn "  → Install Rust: curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"
        warn "  → Then: cargo install --locked magika-cli"
        return 1
    fi
}

# Main
main() {
    echo ""
    info "╔═══════════════════════════════════════════════════════════════╗"
    info "║          vnaiscan - AI Agent Image Security Scanner           ║"
    info "║                     https://scan.vnai.dev                      ║"
    info "╚═══════════════════════════════════════════════════════════════╝"
    echo ""
    
    detect_platform
    info "Platform: ${PLATFORM}"
    info "Install directory: ${INSTALL_DIR}"
    echo ""
    
    local failed=0
    
    install_vnaiscan || ((failed++))
    install_trivy || ((failed++))
    install_malcontent || ((failed++))
    install_magika || ((failed++))
    
    echo ""
    if [ $failed -eq 0 ]; then
        success "All tools installed successfully!"
    else
        warn "Some tools could not be installed ($failed failed)"
        echo ""
        info "Recommendation: Use Docker for a complete setup:"
        echo "  docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \\"
        echo "    ghcr.io/vnai-dev/vnaiscan scan <image:tag>"
    fi
    
    echo ""
    info "Quick start:"
    echo "  vnaiscan scan alpine:latest"
    echo "  vnaiscan tools"
    echo ""
}

main "$@"
