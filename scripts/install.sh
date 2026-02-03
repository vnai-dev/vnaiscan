#!/bin/bash
# vnaiscan installer - https://scan.vnai.dev
# Usage: curl -sSL https://scan.vnai.dev/install.sh | sh
#
# Downloads a complete bundle with all tools included:
#   - vnaiscan (scanner CLI)
#   - trivy (CVE/secrets)
#   - magika (AI file detection)
#   - malcontent (binary analysis) - Linux only

set -e

REPO="vnai-dev/vnaiscan"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
LIB_DIR="${LIB_DIR:-/usr/local/lib}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m'

info() { echo -e "${CYAN}$1${NC}"; }
success() { echo -e "${GREEN}✓ $1${NC}"; }
warn() { echo -e "${YELLOW}⚠ $1${NC}"; }
error() { echo -e "${RED}✗ $1${NC}" >&2; }

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

get_latest_version() {
    curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | \
        grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/' | head -1
}

install() {
    info "🔍 Detecting platform..."
    detect_platform
    info "   Platform: ${PLATFORM}"

    info "📦 Getting latest version..."
    VERSION=$(get_latest_version)
    if [ -z "$VERSION" ]; then
        VERSION="0.1.0"
        warn "   Could not detect latest version, using: v${VERSION}"
    else
        info "   Version: v${VERSION}"
    fi

    BUNDLE_NAME="vnaiscan_${VERSION}_${PLATFORM}"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${BUNDLE_NAME}.tar.gz"

    info "⬇️  Downloading vnaiscan bundle..."
    info "   URL: ${DOWNLOAD_URL}"

    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT

    if ! curl -sSL "$DOWNLOAD_URL" | tar -xz -C "$TMP_DIR" 2>/dev/null; then
        error "Failed to download vnaiscan bundle"
        echo ""
        warn "The release may not exist yet. Try Docker instead:"
        echo "  docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \\"
        echo "    ghcr.io/vnai-dev/vnaiscan scan <image:tag>"
        exit 1
    fi

    info "📁 Installing to ${INSTALL_DIR}..."
    
    # Install binaries
    for bin in vnaiscan trivy magika mal; do
        if [ -f "$TMP_DIR/${BUNDLE_NAME}/bin/$bin" ]; then
            if [ -w "$INSTALL_DIR" ]; then
                cp "$TMP_DIR/${BUNDLE_NAME}/bin/$bin" "$INSTALL_DIR/"
                chmod +x "$INSTALL_DIR/$bin"
            else
                sudo cp "$TMP_DIR/${BUNDLE_NAME}/bin/$bin" "$INSTALL_DIR/"
                sudo chmod +x "$INSTALL_DIR/$bin"
            fi
            success "$bin installed"
        fi
    done

    # Install libraries (for malcontent)
    if [ -d "$TMP_DIR/${BUNDLE_NAME}/lib" ]; then
        for lib in "$TMP_DIR/${BUNDLE_NAME}/lib/"*; do
            if [ -f "$lib" ]; then
                if [ -w "$LIB_DIR" ]; then
                    cp "$lib" "$LIB_DIR/"
                else
                    sudo cp "$lib" "$LIB_DIR/"
                fi
            fi
        done
        # Update library cache on Linux
        if [ "$OS" = "linux" ] && command -v ldconfig &>/dev/null; then
            sudo ldconfig 2>/dev/null || true
        fi
    fi

    echo ""
    success "vnaiscan installed successfully!"
    echo ""
    info "Quick start:"
    echo "  vnaiscan scan alpine:latest"
    echo "  vnaiscan tools"
    echo ""
}

main() {
    echo ""
    info "╔═══════════════════════════════════════════════════════════════╗"
    info "║          vnaiscan - AI Agent Image Security Scanner           ║"
    info "║                     https://scan.vnai.dev                      ║"
    info "╚═══════════════════════════════════════════════════════════════╝"
    echo ""

    # Check requirements
    for cmd in curl tar; do
        if ! command -v "$cmd" &>/dev/null; then
            error "Missing required tool: $cmd"
            exit 1
        fi
    done

    install
}

main
