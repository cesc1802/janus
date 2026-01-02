#!/bin/sh
# Install script for migrate-tool
# Downloads and installs migrate-tool from GitHub releases
# Usage: ./install.sh [--version VERSION]

set -e

# Configuration
REPO="cesc1802/migration-tool"
BINARY_NAME="migrate-tool"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
GITHUB_API="https://api.github.com"
GITHUB_DOWNLOAD="https://github.com"

# Colors (respects NO_COLOR)
if [ -z "${NO_COLOR:-}" ] && [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[0;33m'
    BLUE='\033[0;34m'
    NC='\033[0m'
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    NC=''
fi

# Logging functions
info() { printf "${BLUE}[INFO]${NC} %s\n" "$1" >&2; }
success() { printf "${GREEN}[OK]${NC} %s\n" "$1" >&2; }
warn() { printf "${YELLOW}[WARN]${NC} %s\n" "$1" >&2; }
error() { printf "${RED}[ERROR]${NC} %s\n" "$1" >&2; }
die() { error "$1"; exit 1; }

# Cleanup on exit
TEMP_DIR=""
cleanup() {
    if [ -n "$TEMP_DIR" ] && [ -d "$TEMP_DIR" ]; then
        rm -rf "$TEMP_DIR"
    fi
}
trap cleanup EXIT INT TERM

# Detect operating system
detect_os() {
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$os" in
        linux) echo "linux" ;;
        darwin) echo "darwin" ;;
        *) die "Unsupported OS: $os (supported: linux, darwin)" ;;
    esac
}

# Detect CPU architecture
detect_arch() {
    arch=$(uname -m)
    case "$arch" in
        x86_64|amd64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *) die "Unsupported architecture: $arch (supported: amd64, arm64)" ;;
    esac
}

# Check for required commands
check_dependencies() {
    if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
        die "Either curl or wget is required. Please install one and try again."
    fi
    if ! command -v tar >/dev/null 2>&1; then
        die "tar is required. Please install it and try again."
    fi
}

# HTTP GET request (curl/wget fallback)
http_get() {
    url="$1"
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$url"
    else
        wget -qO- "$url"
    fi
}

# Download file (curl/wget fallback)
download_file() {
    url="$1"
    output="$2"
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL -o "$output" "$url"
    else
        wget -q -O "$output" "$url"
    fi
}

# Get latest release version from GitHub API
get_latest_version() {
    info "Fetching latest version..."
    response=$(http_get "${GITHUB_API}/repos/${REPO}/releases/latest" 2>/dev/null) || {
        die "Failed to fetch latest release. Check network or try: --version vX.Y.Z"
    }
    # Parse tag_name from JSON without jq
    version=$(echo "$response" | grep -o '"tag_name"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')
    if [ -z "$version" ]; then
        die "Failed to parse version from GitHub API response"
    fi
    # Validate version format (vX.Y.Z or vX.Y.Z-suffix)
    case "$version" in
        v[0-9]*) ;;
        *) die "Invalid version format from API: $version" ;;
    esac
    echo "$version"
}

# Verify checksum using sha256sum or shasum
verify_checksum() {
    archive="$1"
    checksums_file="$2"
    archive_name="${archive##*/}"

    # Extract expected checksum for this file
    expected=$(grep "$archive_name" "$checksums_file" | awk '{print $1}')
    if [ -z "$expected" ]; then
        die "Checksum not found for $archive_name in checksums.txt"
    fi

    # Calculate actual checksum
    if command -v sha256sum >/dev/null 2>&1; then
        actual=$(sha256sum "$archive" | awk '{print $1}')
    elif command -v shasum >/dev/null 2>&1; then
        actual=$(shasum -a 256 "$archive" | awk '{print $1}')
    else
        die "Neither sha256sum nor shasum available for checksum verification"
    fi

    if [ "$expected" != "$actual" ]; then
        error "Checksum verification failed!"
        error "Expected: $expected"
        error "Actual:   $actual"
        die "Downloaded file may be corrupted or tampered with"
    fi
    success "Checksum verified"
}

# Print usage
usage() {
    cat <<EOF
Install migrate-tool from GitHub releases

Usage: $0 [OPTIONS]

Options:
    -v, --version VERSION   Install specific version (e.g., v0.0.3)
    -d, --dir DIRECTORY     Install directory (default: /usr/local/bin)
    -h, --help              Show this help message

Environment Variables:
    INSTALL_DIR             Install directory (default: /usr/local/bin)
    NO_COLOR                Disable colored output

Examples:
    $0                      Install latest version
    $0 --version v0.0.3     Install specific version
    INSTALL_DIR=~/.local/bin $0   Install to user directory
EOF
    exit 0
}

# Parse command line arguments
parse_args() {
    VERSION=""
    while [ $# -gt 0 ]; do
        case "$1" in
            -v|--version)
                [ -z "${2:-}" ] && die "Version argument required"
                VERSION="$2"
                shift 2
                ;;
            -d|--dir)
                [ -z "${2:-}" ] && die "Directory argument required"
                INSTALL_DIR="$2"
                shift 2
                ;;
            -h|--help)
                usage
                ;;
            *)
                die "Unknown option: $1. Use --help for usage."
                ;;
        esac
    done
}

# Main installation function
main() {
    parse_args "$@"

    info "Installing migrate-tool"
    echo ""

    # Check dependencies
    check_dependencies

    # Detect platform
    OS=$(detect_os)
    ARCH=$(detect_arch)
    info "Detected platform: ${OS}/${ARCH}"

    # Get version
    if [ -z "$VERSION" ]; then
        VERSION=$(get_latest_version)
    fi
    # Strip 'v' prefix if present for archive naming
    VERSION_NUM="${VERSION#v}"
    info "Version: ${VERSION}"

    # Create temp directory
    TEMP_DIR=$(mktemp -d)

    # Build download URLs
    ARCHIVE_NAME="${BINARY_NAME}_${VERSION_NUM}_${OS}_${ARCH}.tar.gz"
    ARCHIVE_URL="${GITHUB_DOWNLOAD}/${REPO}/releases/download/${VERSION}/${ARCHIVE_NAME}"
    CHECKSUMS_URL="${GITHUB_DOWNLOAD}/${REPO}/releases/download/${VERSION}/checksums.txt"

    # Download archive and checksums
    info "Downloading ${ARCHIVE_NAME}..."
    download_file "$ARCHIVE_URL" "${TEMP_DIR}/${ARCHIVE_NAME}" || die "Failed to download archive from ${ARCHIVE_URL}"
    success "Downloaded archive"

    info "Downloading checksums..."
    download_file "$CHECKSUMS_URL" "${TEMP_DIR}/checksums.txt" || die "Failed to download checksums"
    success "Downloaded checksums"

    # Verify checksum
    info "Verifying checksum..."
    verify_checksum "${TEMP_DIR}/${ARCHIVE_NAME}" "${TEMP_DIR}/checksums.txt"

    # Extract archive
    info "Extracting archive..."
    tar -xzf "${TEMP_DIR}/${ARCHIVE_NAME}" -C "${TEMP_DIR}" || die "Failed to extract archive"
    success "Extracted archive"

    # Install binary
    info "Installing to ${INSTALL_DIR}..."
    if [ ! -d "$INSTALL_DIR" ]; then
        mkdir -p "$INSTALL_DIR" 2>/dev/null || {
            error "Cannot create ${INSTALL_DIR}. Try:"
            error "  sudo $0 $*"
            error "  INSTALL_DIR=~/.local/bin $0 $*"
            die "Installation failed"
        }
    fi

    if [ ! -w "$INSTALL_DIR" ]; then
        error "Cannot write to ${INSTALL_DIR}. Try:"
        error "  sudo $0 $*"
        error "  INSTALL_DIR=~/.local/bin $0 $*"
        die "Installation failed"
    fi

    mv "${TEMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}" || die "Failed to move binary"
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}" || die "Failed to set executable permission"
    success "Installed ${BINARY_NAME} to ${INSTALL_DIR}"

    echo ""
    success "Installation complete!"
    echo ""

    # Verify installation
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        info "Version: $($BINARY_NAME version 2>/dev/null || echo 'unknown')"
    else
        warn "${INSTALL_DIR} may not be in your PATH"
        info "Add to PATH: export PATH=\"${INSTALL_DIR}:\$PATH\""
    fi
}

main "$@"
