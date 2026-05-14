#!/usr/bin/env bash
set -euo pipefail

# Datakraften Installer
# Installs the 'dk' CLI.
#
# Usage:
#   curl -fsSL https://datakraften.no/install | bash
#   curl -fsSL https://datakraften.no/install | bash -s -- --version

DATARKAFTEN_REPO="https://github.com/sagathelab/datakraften.git"
DATARKAFTEN_DIR="${HOME}/.local/share/datakraften/source"
BIN_DIR="${HOME}/.local/bin"
BINARY="${BIN_DIR}/dk"
VERSION="dev"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

info()  { echo -e "${CYAN}  ${1}${NC}"; }
ok()    { echo -e "${GREEN}  ✓ ${1}${NC}"; }
warn()  { echo -e "${YELLOW}  ⚠ ${1}${NC}"; }
err()   { echo -e "${RED}  ✗ ${1}${NC}"; }

echo ""
echo -e "${CYAN}Datakraften Installer${NC}"
echo ""

# --- Pre-flight checks ---

# Check for --version flag
if [ "${1:-}" = "--version" ]; then
	echo "Datakraften installer ${VERSION}"
	exit 0
fi

# Must not be root
if [ "$(id -u)" = "0" ]; then
	err "Do not run this installer as root."
	err "Run it as a regular user with sudo privileges."
	exit 1
fi

# Platform detection (informational only)
if grep -qi microsoft /proc/version 2>/dev/null || grep -qi wsl /proc/version 2>/dev/null; then
	info "WSL environment detected"
else
	info "Native Linux environment detected"
fi

# --- Dependencies ---

info "Checking dependencies..."

HAVE_GO=false
if command -v go &>/dev/null; then
	HAVE_GO=true
	ok "Go found: $(go version)"
fi

# Detect package manager
install_pkg() {
	if command -v apt-get &>/dev/null; then
		sudo apt-get update -qq
		sudo apt-get install -y -qq "$@"
	elif command -v dnf &>/dev/null; then
		sudo dnf install -y "$@"
	elif command -v yum &>/dev/null; then
		sudo yum install -y "$@"
	elif command -v pacman &>/dev/null; then
		sudo pacman -S --noconfirm "$@"
	elif command -v brew &>/dev/null; then
		brew install "$@"
	else
		warn "No known package manager found. Please install manually: $*"
		return 1
	fi
}

# Map dk dependency names to distro-specific packages
git_pkg()   { echo "git"; }
curl_pkg()  { echo "curl"; }
make_pkg()  {
	if command -v apt-get &>/dev/null; then echo "build-essential make"
	elif command -v dnf &>/dev/null || command -v yum &>/dev/null; then echo "make"
	else echo "make"; fi
}

if ! command -v git &>/dev/null; then
	info "Installing git..."
	install_pkg "$(git_pkg)" && ok "git installed"
else
	ok "git found"
fi

if ! command -v curl &>/dev/null; then
	info "Installing curl..."
	install_pkg "$(curl_pkg)" && ok "curl installed"
else
	ok "curl found"
fi

if ! command -v make &>/dev/null; then
	info "Installing build tools..."
	install_pkg "$(make_pkg)" && ok "build tools installed"
else
	ok "make found"
fi

# --- Go installation ---

if [ "${HAVE_GO}" = false ]; then
	info "Installing Go..."
	GO_VERSION="1.22.5"
	GO_TARBALL="go${GO_VERSION}.linux-amd64.tar.gz"
	GO_URL="https://go.dev/dl/${GO_TARBALL}"

	curl -fsSL "${GO_URL}" -o "/tmp/${GO_TARBALL}"
	sudo tar -C /usr/local -xzf "/tmp/${GO_TARBALL}"
	rm "/tmp/${GO_TARBALL}"

	export PATH="/usr/local/go/bin:${PATH}"
	ok "Go $(go version) installed"
fi

# Ensure go is in PATH for the rest of the script
if ! command -v go &>/dev/null; then
	export PATH="/usr/local/go/bin:${PATH}"
fi

# --- Clone or update source ---

info "Setting up Datakraften source..."

mkdir -p "${DATARKAFTEN_DIR}"

if [ -d "${DATARKAFTEN_DIR}/.git" ]; then
	info "Updating existing source..."
	cd "${DATARKAFTEN_DIR}"
	git pull --ff-only
	ok "Source updated"
else
	info "Cloning Datakraften repository..."
	git clone --depth 1 "${DATARKAFTEN_REPO}" "${DATARKAFTEN_DIR}"
	ok "Source cloned"
fi

# --- Build ---

info "Building dk CLI..."
cd "${DATARKAFTEN_DIR}"
if command -v make &>/dev/null; then
	make build
else
	go build -o bin/dk ./cmd/dk/
fi
ok "dk built"

# --- Install ---

info "Installing dk to ${BIN_DIR}..."
mkdir -p "${BIN_DIR}"
cp "${DATARKAFTEN_DIR}/bin/dk" "${BINARY}"
ok "dk installed to ${BINARY}"

# --- Add to PATH ---

if ! echo "${PATH}" | tr ':' '\n' | grep -q "${BIN_DIR}"; then
	SHELL_CONFIG="${HOME}/.profile"
	if [ -f "${HOME}/.bashrc" ]; then
		SHELL_CONFIG="${HOME}/.bashrc"
	fi

	if ! grep -q "datakraften" "${SHELL_CONFIG}" 2>/dev/null; then
		cat >> "${SHELL_CONFIG}" <<- 'EOF'

# >>> datakraften >>>
export PATH="${HOME}/.local/bin:${PATH}"
# <<< datakraften <<<
EOF
		ok "Added ${BIN_DIR} to PATH in ${SHELL_CONFIG}"
	fi
fi

echo ""
ok "Datakraften installed successfully!"
echo ""
info "Next steps:"
info "  1. Restart your shell or run: source ${SHELL_CONFIG}"
info "  2. Run: dk init"
info "  3. Run: dk doctor"
echo ""
