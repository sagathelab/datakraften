#!/usr/bin/env bash
#
# DATAKRAFTEN WSL Developer Bootstrapper
# =======================================
# Usage: curl -fsSL https://datakraften.no/bootstrap/wsl | bash
#
set -euo pipefail

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

VSCODE_EXTENSIONS=(
    ms-azuretools.vscode-azurecli
    ms-dotnettools.csharp
    GitHub.vscode-pull-request-github
)

# ── Helpers ──────────────────────────────────────────────────
log()      { printf "${MAGENTA}==>${NC} %s\n" "$*"; }
log_ok()   { printf "${GREEN}  ✓${NC} %s\n" "$*"; }
log_skip() { printf "${YELLOW}  –${NC} %s\n" "$*"; }
log_info() { printf "${CYAN}  ℹ${NC} %s\n" "$*"; }
log_err()  { printf "${RED}  ✗${NC} %s\n" "$*" >&2; }

run_step() {
    local name="$1" desc="$2"
    shift 2
    log "${desc}"
    if "$@"; then
        log_ok "${name}"
    else
        log_err "${name} feilet"
        return 1
    fi
}

check_not_root() {
    if [[ $EUID -eq 0 ]]; then
        log_err "Skriptet skal ikke kjøres som root. Kjør som vanlig bruker."
        exit 1
    fi
}

check_wsl() {
    if ! grep -qi microsoft /proc/version 2>/dev/null; then
        log_info "Dette skriptet er optimalisert for WSL, men kjører fint på native Linux også."
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
    sudo apt-get update -qq && sudo apt-get upgrade -y -qq
    log_ok "Systempakker oppdatert"
}

install_system_deps() {
    local deps=(build-essential curl wget git ca-certificates gnupg lsb-release software-properties-common)
    sudo apt-get install -y -qq "${deps[@]}"
    log_ok "Systemavhengigheter installert"
}

# ── Homebrew ─────────────────────────────────────────────────
install_homebrew() {
    if command -v brew &>/dev/null; then
        log_skip "Homebrew $(brew --version 2>/dev/null | head -1) — allerede installert"
        return 0
    fi

    log "Installerer Homebrew..."
    NONINTERACTIVE=1 /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

    eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"
    log_ok "Homebrew $(brew --version | head -1) installert"
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
    local to_install=()
    for pkg in "${BREW_PACKAGES[@]}"; do
        if brew list "$pkg" &>/dev/null 2>&1; then
            log_skip "$pkg — allerede installert"
        else
            to_install+=("$pkg")
        fi
    done

    if [[ ${#to_install[@]} -gt 0 ]]; then
        brew install "${to_install[@]}"
        log_ok "${#to_install[@]} Homebrew-pakker installert"
    fi
}

# ── Node.js / fnm ────────────────────────────────────────────
setup_node() {
    eval "$(fnm env --use-on-cd --shell bash 2>/dev/null)" || true

    if command -v node &>/dev/null; then
        log_skip "Node.js $(node --version) — allerede installert"
        return 0
    fi

    log "Installerer Node.js LTS..."
    fnm install --lts
    fnm default lts-latest
    eval "$(fnm env --use-on-cd --shell bash)"
    log_ok "Node.js $(node --version) / npm $(npm --version) installert"
}

# ── Python / uv ──────────────────────────────────────────────
setup_python() {
    if uv python list 2>/dev/null | grep -q .; then
        local py_ver
        py_ver=$(uv python list 2>/dev/null | head -1 | awk '{print $1}')
        log_skip "Python ${py_ver} via uv — allerede installert"
        return 0
    fi

    log "Installerer Python via uv..."
    uv python install
    local py_ver
    py_ver=$(uv python list | head -1 | awk '{print $1}')
    log_ok "Python ${py_ver} installert via uv"
}

# ── OpenCode ─────────────────────────────────────────────────
install_opencode() {
    if command -v opencode &>/dev/null; then
        log_skip "OpenCode — allerede installert"
        return 0
    fi

    curl -fsSL https://opencode.ai/install | bash
    log_ok "OpenCode installert"
}

# ── Zed ──────────────────────────────────────────────────────
install_zed() {
    if command -v zed &>/dev/null; then
        log_skip "Zed — allerede installert"
        return 0
    fi

    curl -f https://zed.dev/install.sh | sh
    export PATH="$HOME/.local/bin:$PATH"
    log_ok "Zed installert"
}

# ── OpenAI Codex CLI ─────────────────────────────────────────
install_codex() {
    if command -v codex &>/dev/null; then
        log_skip "OpenAI Codex — allerede installert"
        return 0
    fi

    npm install -g @openai/codex
    log_ok "OpenAI Codex CLI installert"
}

# ── GitHub Copilot CLI ───────────────────────────────────────
setup_gh_copilot() {
    if gh extension list 2>/dev/null | grep -q "gh-copilot"; then
        log_skip "GitHub Copilot CLI — allerede installert"
        return 0
    fi

    gh extension install github/gh-copilot
    log_ok "GitHub Copilot CLI installert"
}

# ── VS Code ──────────────────────────────────────────────────
install_vscode_cli() {
    if command -v code &>/dev/null; then
        log_skip "VS Code CLI — allerede installert"
        return 0
    fi

    local arch
    arch=$(dpkg --print-architecture 2>/dev/null || echo "amd64")

    sudo mkdir -p /etc/apt/keyrings
    curl -fsSL https://packages.microsoft.com/keys/microsoft.asc |
        sudo gpg --dearmor -o /etc/apt/keyrings/packages.microsoft.gpg

    sudo tee /etc/apt/sources.list.d/vscode.list >/dev/null <<<"deb [arch=${arch} signed-by=/etc/apt/keyrings/packages.microsoft.gpg] https://packages.microsoft.com/repos/code stable main"

    sudo apt-get update -qq
    sudo apt-get install -y -qq code
    log_ok "VS Code CLI installert"
}

install_vscode_extensions() {
    for ext in "${VSCODE_EXTENSIONS[@]}"; do
        if code --list-extensions 2>/dev/null | grep -qiFx "$ext"; then
            log_skip "${ext} — allerede installert"
        else
            code --install-extension "$ext" --force
            log_ok "${ext} installert"
        fi
    done
}

# ── Docker ───────────────────────────────────────────────────
setup_docker() {
    if ! command -v docker &>/dev/null; then
        log_info "Docker CLI mangler — skal være installert via Homebrew"
        return 1
    fi

    if [[ -S /var/run/docker.sock ]]; then
        log_info "Docker Desktop WSL-integrasjon oppdaget"

        if groups "$USER" 2>/dev/null | grep -q docker; then
            log_skip "docker-gruppe — allerede medlem"
        else
            sudo usermod -aG docker "$USER"
            log_info "Bruker lagt til docker-gruppen (start shell på nytt for å aktivere)"
        fi

        log_ok "Docker klar"
    else
        log_info "Docker Desktop WSL-integrasjon ikke oppdaget"
        log_info "Installer Docker Desktop på Windows og aktiver WSL2-integrasjon:"
        log_info "  https://docs.docker.com/desktop/wsl/"
        log_skip "Docker (venter på Docker Desktop)"
    fi
}

# ── Fish ─────────────────────────────────────────────────────
setup_fish() {
    if ! command -v fish &>/dev/null; then
        log_err "Fish shell ble ikke installert"
        return 1
    fi

    local fish_path
    fish_path="$(command -v fish)"

    if [[ "$SHELL" == "$fish_path" ]]; then
        log_skip "Fish som standard shell — allerede satt"
    else
        if chsh -s "$fish_path" 2>/dev/null; then
            log_ok "Fish satt som standard shell (krever ny pålogging)"
        else
            log_info "Kunne ikke endre shell — kjør manuelt: chsh -s ${fish_path}"
        fi
    fi

    local config_dir="$HOME/.config/fish"
    local config_file="$config_dir/config.fish"

    mkdir -p "$config_dir"

    if [[ -f "$config_file" ]]; then
        log_skip "Fish config.fish — finnes allerede"
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
        log_ok "Fish config skrevet til ${config_file}"
    fi
}

# ── Summary ──────────────────────────────────────────────────
print_summary() {
    local shell_cmd
    if command -v fish &>/dev/null; then
        shell_cmd="exec fish"
    else
        shell_cmd="exec bash -l"
    fi

    echo
    printf "${MAGENTA}${BOLD}╔══════════════════════════════════════════════════════════╗${NC}\n"
    printf "${MAGENTA}${BOLD}║        ✅  DATAKRAFTEN WSL Bootstrap fullført!          ║${NC}\n"
    printf "${MAGENTA}${BOLD}╚══════════════════════════════════════════════════════════╝${NC}\n"
    echo
    log_info "Start et nytt shell for å aktivere alle endringer:"
    log_info "  ${shell_cmd}"
    echo
    log_info "Verifiser installasjonen (oversikt):"
    for cmd in node python3 az dotnet gh opencode zed; do
        if command -v "$cmd" &>/dev/null; then
            printf "  ${GREEN}✓${NC} %s\n" "$cmd"
        else
            printf "  ${YELLOW}–${NC} %s (ikke installert)\n" "$cmd"
        fi
    done
    echo
    log_info "Neste steg:"
    log_info "  1. Åpne VS Code og logg inn på GitHub"
    log_info "  2. Kjør 'gh auth login' for å autentisere GitHub CLI"
    log_info "  3. Kjør 'az login' for å autentisere Azure CLI"
    log_info "  4. Kjør 'opencode' for å starte OpenCode"
    echo
}

# ── Main ─────────────────────────────────────────────────────
main() {
    show_logo
    check_not_root
    check_wsl

    run_step "Systemoppdatering"   "Oppdaterer systempakker..."        system_update
    run_step "Systemavhengigheter" "Installerer systemavhengigheter..." install_system_deps
    run_step "Homebrew"            "Installerer Homebrew..."            install_homebrew

    setup_homebrew_path

    run_step "Homebrew-pakker"     "Installerer Homebrew-pakker..."     install_brew_packages
    run_step "Node.js"             "Setter opp Node.js LTS via fnm..."  setup_node
    run_step "Python"              "Setter opp Python via uv..."       setup_python
    run_step "OpenCode"            "Installerer OpenCode..."            install_opencode
    run_step "Zed Editor"          "Installerer Zed Editor..."          install_zed
    run_step "OpenAI Codex"        "Installerer OpenAI Codex CLI..."    install_codex
    run_step "GitHub Copilot"      "Installerer GitHub Copilot CLI..."  setup_gh_copilot
    run_step "VS Code"             "Installerer VS Code CLI..."         install_vscode_cli
    run_step "VS Code Extensions"  "Installerer VS Code extensions..."  install_vscode_extensions
    run_step "Docker"              "Setter opp Docker WSL..."          setup_docker
    run_step "Fish shell"          "Setter opp Fish shell..."           setup_fish

    print_summary
}

main "$@"
