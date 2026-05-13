#!/usr/bin/env bash
#
# DATAKRAFTEN WSL Developer Bootstrapper
# =======================================
# Usage: curl -fsSL https://datakraften.no/bootstrap/wsl | bash
#
set -Eeuo pipefail

export DEBIAN_FRONTEND=noninteractive

# ── Colors ──────────────────────────────────────────────────
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
BOLD='\033[1m'
NC='\033[0m'

# ── Configuration ───────────────────────────────────────────
BREW_PACKAGES=(
    azure-cli
    dotnet
    gh
    fnm
    uv
    atuin
    fzf
    broot
    fd
    bottom
    starship
    powershell
    docker
    docker-compose
)


# ── Helpers ──────────────────────────────────────────────────
log()      { printf "${MAGENTA}==>${NC} %s\n" "$*"; }
log_ok()   { printf "${GREEN}  ✓${NC} %s\n" "$*"; }
log_skip() { printf "${YELLOW}  –${NC} %s\n" "$*"; }
log_info() { printf "${CYAN}  ℹ${NC} %s\n" "$*"; }
log_err()  { printf "${RED}  ✗${NC} %s\n" "$*" >&2; }

run_step() {
    local name="$1" desc="$2" rc=0
    shift 2
    log "${desc}"
    "$@" || rc=$?
    if [[ $rc -eq 0 ]]; then
        log_ok "${name}"
    else
        log_err "${name} failed"
    fi
    return $rc
}

check_not_root() {
    if [[ $EUID -eq 0 ]]; then
        log_err "This script should not be run as root. Run as a regular user."
        exit 1
    fi
}

check_wsl() {
    if ! grep -qi microsoft /proc/version 2>/dev/null; then
        log_info "This script is optimized for WSL, but works fine on native Linux too."
    fi
}

refresh_shell_command_cache() {
    hash -r 2>/dev/null || true
}

command_is_windows_backed() {
    local resolved
    resolved="$(command -v "$1" 2>/dev/null)" || return 1
    [[ "$resolved" == /mnt/[a-zA-Z]/* ]]
}

is_linux_command() {
    local resolved
    resolved="$(command -v "$1" 2>/dev/null)" || return 1
    [[ "$resolved" != /mnt/[a-zA-Z]/* ]]
}

verify_linux_command() {
    if ! is_linux_command "$1"; then
        log_err "$1 points to Windows path: $(command -v "$1")"
        return 1
    fi
}

# ── Logo ─────────────────────────────────────────────────────
show_logo() {
    printf "${MAGENTA}${BOLD}"
    cat <<'LOGO'
██████╗  █████╗ ████████╗ █████╗ ██╗  ██╗██████╗  █████╗ ███████╗████████╗███████╗███╗   ██╗
██╔══██╗██╔══██╗╚══██╔══╝██╔══██╗██║ ██╔╝██╔══██╗██╔══██╗██╔════╝╚══██╔══╝██╔════╝████╗  ██║
██║  ██║███████║   ██║   ███████║█████╔╝ ██████╔╝███████║█████╗     ██║   █████╗  ██╔██╗ ██║
██║  ██║██╔══██║   ██║   ██╔══██║██╔═██╗ ██╔══██╗██╔══██║██╔══╝     ██║   ██╔══╝  ██║╚██╗██║
██████╔╝██║  ██║   ██║   ██║  ██║██║  ██╗██║  ██║██║  ██║██║        ██║   ███████╗██║ ╚████║
╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝        ╚═╝   ╚══════╝╚═╝  ╚═══╝
LOGO
    printf "${NC}"
    printf "${MAGENTA}${BOLD}             🚀  WSL Developer Bootstrapper  —  v1.0.0${NC}\n\n"
}

# ── System ───────────────────────────────────────────────────
system_update() {
    sudo apt-get update -qq && sudo apt-get upgrade -y -qq || return 1
    refresh_shell_command_cache
    log_ok "System packages updated"
}

install_system_deps() {
    local deps=(build-essential curl wget git ca-certificates gnupg lsb-release)
    sudo apt-get install -y -qq "${deps[@]}" || return 1
    refresh_shell_command_cache
    log_ok "System dependencies installed"
}

# ── Homebrew ─────────────────────────────────────────────────
install_homebrew() {
    if is_linux_command brew; then
        log_skip "Homebrew $(brew --version 2>/dev/null | head -1) — already installed"
        return 0
    fi

    if command -v brew &>/dev/null && command_is_windows_backed brew; then
        log_info "brew points to Windows path — will install Linux version"
    fi

    log "Installing Homebrew..."
    NONINTERACTIVE=1 /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)" || return 1

    eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"
    refresh_shell_command_cache
    log_ok "Homebrew $(brew --version | head -1) installed"
}

setup_homebrew_path() {
    if [[ -f /home/linuxbrew/.linuxbrew/bin/brew ]]; then
        eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"
    elif [[ -f ~/.linuxbrew/bin/brew ]]; then
        eval "$(~/.linuxbrew/bin/brew shellenv)"
    fi
}

# ── Brew Packages ────────────────────────────────────────────
install_brew_packages() {
    setup_homebrew_path
    refresh_shell_command_cache

    local to_install=()
    for pkg in "${BREW_PACKAGES[@]}"; do
        if brew list "$pkg" &>/dev/null 2>&1; then
            log_skip "$pkg — already installed"
        else
            to_install+=("$pkg")
        fi
    done

    if [[ ${#to_install[@]} -gt 0 ]]; then
        brew install "${to_install[@]}" || return 1
        refresh_shell_command_cache
        log_ok "${#to_install[@]} Homebrew packages installed"
    fi
}

# ── Node.js / fnm ────────────────────────────────────────────
setup_node() {
    eval "$(fnm env --use-on-cd --shell bash 2>/dev/null)" || true
    refresh_shell_command_cache

    if is_linux_command node; then
        log_skip "Node.js $(node --version) — already installed"
        return 0
    fi

    if command -v node &>/dev/null && command_is_windows_backed node; then
        log_info "node points to Windows path — installing Linux version"
    fi

    log "Installing Node.js LTS..."
    fnm install --lts || return 1
    fnm default lts-latest || return 1
    eval "$(fnm env --use-on-cd --shell bash)"
    refresh_shell_command_cache
    log_ok "Node.js $(node --version) / npm $(npm --version) installed"
}

# ── Python / uv ──────────────────────────────────────────────
setup_python() {
    if uv python list 2>/dev/null | grep -q .; then
        local py_ver
        py_ver=$(uv python list 2>/dev/null | head -1 | awk '{print $1}')
        log_skip "Python ${py_ver} via uv — already installed"
        return 0
    fi

    log "Installing Python via uv..."
    uv python install || return 1
    local py_ver
    py_ver=$(uv python list | head -1 | awk '{print $1}')
    log_ok "Python ${py_ver} installed via uv"
}

# ── OpenCode ─────────────────────────────────────────────────
install_opencode() {
    if is_linux_command opencode; then
        log_skip "OpenCode — already installed"
        return 0
    fi

    curl -fsSL https://opencode.ai/install | bash || return 1
    refresh_shell_command_cache
    log_ok "OpenCode installed"
}

# ── Zed ──────────────────────────────────────────────────────
install_zed() {
    if is_linux_command zed; then
        log_skip "Zed — already installed"
        return 0
    fi

    curl -f https://zed.dev/install.sh | sh || return 1
    export PATH="$HOME/.local/bin:$PATH"
    refresh_shell_command_cache
    log_ok "Zed installed"
}

# ── OpenAI Codex CLI ─────────────────────────────────────────
install_codex() {
    if is_linux_command codex; then
        log_skip "OpenAI Codex — already installed"
        return 0
    fi

    npm install -g @openai/codex || return 1
    refresh_shell_command_cache
    log_ok "OpenAI Codex CLI installed"
}

# ── GitHub Copilot CLI ───────────────────────────────────────
setup_gh_copilot() {
    if ! is_linux_command gh; then
        log_err "gh is not a Linux command — cannot install Copilot extension"
        return 1
    fi

    if gh extension list 2>/dev/null | grep -q "gh-copilot"; then
        log_skip "GitHub Copilot CLI — already installed"
        return 0
    fi

    gh extension install github/gh-copilot
    log_ok "GitHub Copilot CLI installed"
}

# ── VS Code ──────────────────────────────────────────────────
check_vscode() {
    if command -v code &>/dev/null; then
        log_ok "VS Code CLI available (from Windows + WSL extension)"
    else
        log_info "VS Code CLI not found in WSL"
        log_info "Install VS Code on Windows and add the WSL extension"
        log_info "  https://code.visualstudio.com/"
    fi
}

# ── Docker ───────────────────────────────────────────────────
setup_docker() {
    if ! is_linux_command docker; then
        log_info "Docker CLI missing — should be installed via Homebrew"
        return 1
    fi

    if [[ -S /var/run/docker.sock ]]; then
        log_info "Docker Desktop WSL integration detected"

        if groups "$USER" 2>/dev/null | grep -q docker; then
            log_skip "docker group — already a member"
        else
            sudo usermod -aG docker "$USER"
            log_info "User added to docker group (restart shell to activate)"
        fi

        log_ok "Docker ready"
    else
        log_info "Docker Desktop WSL integration not detected"
        log_info "Install Docker Desktop on Windows and enable WSL2 integration:"
        log_info "  https://docs.docker.com/desktop/wsl/"
        log_skip "Docker (waiting for Docker Desktop)"
    fi
}

# ── Fish ─────────────────────────────────────────────────────
setup_fish() {
    if ! is_linux_command fish; then
        log_err "Fish shell was not installed"
        return 1
    fi

    local fish_path
    fish_path="$(command -v fish)"

    if [[ "$SHELL" == "$fish_path" ]]; then
        log_skip "Fish as default shell — already set"
    else
        if chsh -s "$fish_path" 2>/dev/null; then
            log_ok "Fish set as default shell (requires re-login)"
        else
            log_info "Could not change shell — run manually: chsh -s ${fish_path}"
        fi
    fi

    local config_dir="$HOME/.config/fish"
    local config_file="$config_dir/config.fish"

    mkdir -p "$config_dir"

    if [[ -f "$config_file" ]]; then
        log_skip "Fish config.fish — already exists"
    else
        cat >"$config_file" <<'FISH_CONFIG'
# ── DATAKRAFTEN WSL Bootstrap ────────────────────────────────

# Homebrew
if test -f /home/linuxbrew/.linuxbrew/bin/brew
    eval (/home/linuxbrew/.linuxbrew/bin/brew shellenv)
end

# fnm — Node.js version manager
if command -qv fnm
    fnm env --use-on-cd | source
end

# uv — Python package manager
if command -qv uv
    uv generate-shell-completion fish | source
end

# Atuin — shell history
if command -qv atuin
    atuin init fish | source
end

# FZF — fuzzy finder
if command -qv fzf
    fzf --fish | source
end

# Starship — prompt
if command -qv starship
    starship init fish | source
end

# broot — directory tree
if command -qv broot
    broot --print-shell-function fish | source
end

# Local binaries (Zed, etc.)
if test -d $HOME/.local/bin
    fish_add_path -g $HOME/.local/bin
end

# .NET SDK
if command -qv dotnet
    set -x DOTNET_ROOT (brew --prefix dotnet)/libexec
end

# Aliases
alias g   git
alias ga  'git add'
alias gc  'git commit'
alias gp  'git push'
alias gl  'git log --oneline --graph'
alias gs  'git status'
alias gd  'git diff'

# Environment
set -x EDITOR "zed --wait"
set -x VISUAL "zed --wait"
FISH_CONFIG
        log_ok "Fish config written to ${config_file}"
    fi
}

# ── Summary ──────────────────────────────────────────────────
print_summary() {
    local shell_cmd
    if is_linux_command fish; then
        shell_cmd="exec fish"
    else
        shell_cmd="exec bash -l"
    fi

    echo
    printf "${MAGENTA}${BOLD}╔══════════════════════════════════════════════════════════╗${NC}\n"
    printf "${MAGENTA}${BOLD}║        ✅  DATAKRAFTEN WSL Bootstrap complete!           ║${NC}\n"
    printf "${MAGENTA}${BOLD}╚══════════════════════════════════════════════════════════╝${NC}\n"
    echo
    log_info "Start a new shell to activate all changes:"
    log_info "  ${shell_cmd}"
    echo
    log_info "Installation overview:"
    for cmd in node python3 az dotnet gh opencode zed; do
        if is_linux_command "$cmd"; then
            printf "  ${GREEN}✓${NC} %s\n" "$cmd"
        elif command -v "$cmd" &>/dev/null; then
            printf "  ${YELLOW}–${NC} %s (Windows path — use Linux version)\n" "$cmd"
        else
            printf "  ${YELLOW}–${NC} %s (not installed)\n" "$cmd"
        fi
    done
    echo
    log_info "Next steps:"
    log_info "  1. Install VS Code on Windows with the WSL extension"
    log_info "  2. Run 'gh auth login' to authenticate GitHub CLI"
    log_info "  3. Run 'az login' to authenticate Azure CLI"
    log_info "  4. Run 'opencode' to start OpenCode"
    echo
}

# ── Main ─────────────────────────────────────────────────────
main() {
    show_logo
    check_not_root
    check_wsl

    refresh_shell_command_cache

    run_step "System update"       "Updating system packages..."       system_update
    run_step "System dependencies" "Installing system dependencies..."  install_system_deps
    run_step "Homebrew"            "Installing Homebrew..."            install_homebrew

    setup_homebrew_path

    run_step "Homebrew packages"   "Installing Homebrew packages..."    install_brew_packages
    run_step "Node.js"             "Setting up Node.js LTS via fnm..."  setup_node
    run_step "Python"              "Setting up Python via uv..."       setup_python
    run_step "OpenCode"            "Installing OpenCode..."             install_opencode
    run_step "Zed Editor"          "Installing Zed Editor..."           install_zed
    run_step "OpenAI Codex"        "Installing OpenAI Codex CLI..."     install_codex
    run_step "GitHub Copilot"      "Installing GitHub Copilot CLI..."   setup_gh_copilot
    run_step "VS Code"             "Checking VS Code WSL setup..."      check_vscode
    run_step "Docker"              "Setting up Docker WSL..."           setup_docker
    run_step "Fish shell"          "Setting up Fish shell..."           setup_fish

    print_summary
}

main "$@"
