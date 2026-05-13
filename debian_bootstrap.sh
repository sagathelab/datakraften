#!/usr/bin/env bash
set -Eeuo pipefail

export DEBIAN_FRONTEND=noninteractive
export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$PATH"

log() {
  printf "\n[%s] %s\n" "$(date +%H:%M:%S)" "$*"
}

warn() {
  printf "\n[WARN] %s\n" "$*"
}

die() {
  printf "\n[ERROR] %s\n" "$*" >&2
  exit 1
}

have_cmd() {
  command -v "$1" >/dev/null 2>&1
}

apt_pkg_installed() {
  dpkg -s "$1" >/dev/null 2>&1
}

ensure_line_in_file() {
  local line="$1"
  local file="$2"

  mkdir -p "$(dirname "$file")"
  touch "$file"

  if ! grep -Fqx "$line" "$file"; then
    printf '%s\n' "$line" >> "$file"
  fi
}

is_wsl() {
  grep -qi microsoft /proc/version 2>/dev/null
}

refresh_shell_command_cache() {
  hash -r 2>/dev/null || true
}

command_resolved_path() {
  command -v "$1" 2>/dev/null || true
}

command_is_windows_backed() {
  local resolved
  resolved="$(command_resolved_path "$1")"
  [[ -n "$resolved" && "$resolved" == /mnt/[a-zA-Z]/* ]]
}

verify_linux_command() {
  local cmd="$1"
  local resolved

  resolved="$(command_resolved_path "$cmd")"
  [[ -n "$resolved" ]] || die "$cmd ble ikke funnet i PATH"

  if [[ "$resolved" == /mnt/[a-zA-Z]/* ]]; then
    die "$cmd peker til Windows-installasjonen: $resolved"
  fi
}

require_debian() {
  [[ -f /etc/os-release ]] || die "/etc/os-release mangler"
  . /etc/os-release
  [[ "${ID:-}" == "debian" ]] || die "Dette scriptet er laget for Debian"
}

require_sudo() {
  have_cmd sudo || die "sudo mangler"
}

cleanup_temp() {
  rm -f /tmp/go.tar.gz
  rm -f /tmp/nodesource_setup.sh
  rm -f /tmp/packages-microsoft-prod.deb
  rm -f /tmp/powershell.deb
}
trap cleanup_temp EXIT

APT_UPDATED=0
apt_update_once() {
  if [[ "$APT_UPDATED" -eq 0 ]]; then
    log "Kjører apt-get update"
    sudo apt-get update
    APT_UPDATED=1
  fi
}

install_apt_packages() {
  local pkgs=("$@")
  local missing=()

  for p in "${pkgs[@]}"; do
    if ! apt_pkg_installed "$p"; then
      missing+=("$p")
    fi
  done

  if [[ ${#missing[@]} -eq 0 ]]; then
    log "APT-pakker allerede installert: ${pkgs[*]}"
    return
  fi

  apt_update_once
  log "Installerer APT-pakker: ${missing[*]}"
  sudo apt-get install -y "${missing[@]}"
}

warn_if_windows_path_conflicts() {
  if ! is_wsl; then
    return
  fi

  local tools=(az git node npm python3 pwsh dotnet devcontainer)
  local t resolved

  for t in "${tools[@]}"; do
    if have_cmd "$t"; then
      resolved="$(command_resolved_path "$t")"
      if [[ "$resolved" == /mnt/[a-zA-Z]/* ]]; then
        warn "$t peker til Windows-path: $resolved"
      fi
    fi
  done
}

install_base() {
  install_apt_packages \
    git \
    curl \
    wget \
    ca-certificates \
    gpg \
    jq \
    python3 \
    python3-pip \
    lsb-release \
    apt-transport-https \
    make \
    build-essential \
    unzip \
    zip
}

install_gh() {
  refresh_shell_command_cache

  if have_cmd gh && ! command_is_windows_backed gh; then
    log "gh OK ($(gh --version | head -n1))"
    return
  fi

  if have_cmd gh && command_is_windows_backed gh; then
    warn "Fant Windows-gh i PATH: $(command_resolved_path gh)"
  fi

  log "Installerer GitHub CLI"
  sudo mkdir -p -m 755 /etc/apt/keyrings

  curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
    | sudo tee /etc/apt/keyrings/githubcli-archive-keyring.gpg >/dev/null

  sudo chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg

  echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
    | sudo tee /etc/apt/sources.list.d/github-cli.list >/dev/null

  APT_UPDATED=0
  install_apt_packages gh

  refresh_shell_command_cache
  verify_linux_command gh
}

install_az() {
  refresh_shell_command_cache

  if have_cmd az && ! command_is_windows_backed az; then
    log "az OK ($(az version 2>/dev/null | jq -r '."azure-cli"' 2>/dev/null || echo 'installert'))"
    return
  fi

  if have_cmd az && command_is_windows_backed az; then
    warn "Fant Windows-az i PATH: $(command_resolved_path az)"
    warn "Installerer Linux-versjonen av Azure CLI og forventer at /usr/bin/az brukes"
  else
    log "Installerer Azure CLI via offisielt install-script"
  fi

  curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash

  refresh_shell_command_cache

  if [[ -x /usr/bin/az ]]; then
    log "Linux az installert: /usr/bin/az"
  fi

  verify_linux_command az
}

configure_microsoft_repo() {
  . /etc/os-release

  if [[ -f /etc/apt/sources.list.d/microsoft-prod.list ]] || [[ -f /etc/apt/sources.list.d/microsoft-prod.sources ]]; then
    log "Microsoft repository finnes allerede"
    APT_UPDATED=0
    apt_update_once
    return
  fi

  log "Legger til Microsoft repository for Debian ${VERSION_ID}"
  wget -q "https://packages.microsoft.com/config/debian/${VERSION_ID}/packages-microsoft-prod.deb" -O /tmp/packages-microsoft-prod.deb
  sudo dpkg -i /tmp/packages-microsoft-prod.deb
  rm -f /tmp/packages-microsoft-prod.deb

  APT_UPDATED=0
  apt_update_once
}

install_dotnet() {
  local need_install=0

  refresh_shell_command_cache

  if have_cmd dotnet && ! command_is_windows_backed dotnet; then
    if dotnet --list-sdks 2>/dev/null | grep -q '^10\.'; then
      log ".NET SDK 10 er installert"
    else
      warn ".NET SDK 10 mangler"
      need_install=1
    fi

    if dotnet --list-sdks 2>/dev/null | grep -q '^8\.'; then
      log ".NET SDK 8 er installert"
    else
      warn ".NET SDK 8 mangler"
      need_install=1
    fi

    if [[ "$need_install" -eq 0 ]]; then
      log "dotnet SDK OK (8 og 10 finnes allerede)"
      return
    fi
  else
    if have_cmd dotnet && command_is_windows_backed dotnet; then
      warn "Fant Windows-dotnet i PATH: $(command_resolved_path dotnet)"
    fi
    log "Installerer .NET SDK 10 og 8"
  fi

  configure_microsoft_repo
  install_apt_packages dotnet-sdk-10.0 dotnet-sdk-8.0

  refresh_shell_command_cache
  verify_linux_command dotnet
}

install_pwsh_from_apt() {
  configure_microsoft_repo

  if apt-cache policy powershell 2>/dev/null | grep -q 'Candidate:'; then
    local candidate
    candidate="$(apt-cache policy powershell | awk '/Candidate:/ {print $2}')"
    if [[ -n "$candidate" && "$candidate" != "(none)" ]]; then
      log "Fant powershell i apt-repo ($candidate)"
      sudo apt-get install -y powershell
      return 0
    fi
  fi

  return 1
}

install_pwsh_from_deb() {
  log "Installerer PowerShell via direkte .deb"

  local arch ps_arch version url
  arch="$(dpkg --print-architecture)"

  case "$arch" in
    amd64) ps_arch="amd64" ;;
    arm64) ps_arch="arm64" ;;
    *) die "PowerShell støttes ikke for arkitektur: $arch" ;;
  esac

  version="$(curl -fsSL https://api.github.com/repos/PowerShell/PowerShell/releases/latest | jq -r '.tag_name')"
  [[ -n "$version" && "$version" != "null" ]] || die "Kunne ikke hente siste PowerShell-versjon"

  version="${version#v}"
  url="https://github.com/PowerShell/PowerShell/releases/download/v${version}/powershell_${version}-1.deb_${ps_arch}.deb"

  log "Laster ned PowerShell ${version} for ${ps_arch}"
  curl -fsSL "$url" -o /tmp/powershell.deb
  sudo apt-get install -y /tmp/powershell.deb
}

install_pwsh() {
  refresh_shell_command_cache

  if have_cmd pwsh && ! command_is_windows_backed pwsh; then
    log "pwsh OK ($(pwsh --version))"
    return
  fi

  if have_cmd pwsh && command_is_windows_backed pwsh; then
    warn "Fant Windows-pwsh i PATH: $(command_resolved_path pwsh)"
  fi

  log "Installerer PowerShell"

  if install_pwsh_from_apt; then
    :
  else
    warn "powershell ble ikke funnet i apt-repo, bruker direkte .deb i stedet"
    install_pwsh_from_deb
  fi

  refresh_shell_command_cache
  verify_linux_command pwsh
}

remove_old_nodesource() {
  if [[ -f /etc/apt/sources.list.d/nodesource.list ]]; then
    warn "Fjerner gammel NodeSource-liste"
    sudo rm -f /etc/apt/sources.list.d/nodesource.list
  fi

  if [[ -f /etc/apt/sources.list.d/nodesource.sources ]]; then
    warn "Fjerner gammel NodeSource-sources"
    sudo rm -f /etc/apt/sources.list.d/nodesource.sources
  fi
}

install_node() {
  refresh_shell_command_cache

  if have_cmd node && ! command_is_windows_backed node; then
    local major
    major="$(node -p 'process.versions.node.split(".")[0]')"
    if [[ "$major" -ge 22 ]]; then
      log "node OK ($(node --version))"
      return
    fi
    warn "Eksisterende Node er for gammel: $(node --version). Oppgraderer til 22.x"
  else
    if have_cmd node && command_is_windows_backed node; then
      warn "Fant Windows-node i PATH: $(command_resolved_path node)"
    fi
    log "Installerer Node.js 22.x"
  fi

  remove_old_nodesource

  curl -fsSL https://deb.nodesource.com/setup_22.x -o /tmp/nodesource_setup.sh
  sudo -E bash /tmp/nodesource_setup.sh

  APT_UPDATED=0
  install_apt_packages nodejs

  refresh_shell_command_cache
  verify_linux_command node
  verify_linux_command npm
}

install_uv() {
  refresh_shell_command_cache

  if have_cmd uv && ! command_is_windows_backed uv; then
    log "uv OK ($(uv --version))"
    return
  fi

  if have_cmd uv && command_is_windows_backed uv; then
    warn "Fant Windows-uv i PATH: $(command_resolved_path uv)"
  fi

  log "Installerer uv"
  curl -LsSf https://astral.sh/uv/install.sh | sh

  export PATH="$HOME/.local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

  refresh_shell_command_cache
  verify_linux_command uv
}

install_go() {
  local latest current arch
  latest="$(curl -fsSL https://go.dev/VERSION?m=text | head -n1)"

  [[ -n "$latest" ]] || die "Kunne ikke hente siste Go-versjon"

  refresh_shell_command_cache

  if have_cmd go && ! command_is_windows_backed go; then
    current="$(go version | awk '{print $3}')"
    if [[ "$current" == "$latest" ]]; then
      log "go OK ($current)"
      return
    fi
    warn "Oppgraderer Go fra $current til $latest"
  else
    if have_cmd go && command_is_windows_backed go; then
      warn "Fant Windows-go i PATH: $(command_resolved_path go)"
    fi
    log "Installerer Go $latest"
  fi

  arch="$(dpkg --print-architecture)"
  case "$arch" in
    amd64) arch="amd64" ;;
    arm64) arch="arm64" ;;
    *) die "Unsupported arch: $arch" ;;
  esac

  curl -fsSL "https://go.dev/dl/${latest}.linux-${arch}.tar.gz" -o /tmp/go.tar.gz
  sudo rm -rf /usr/local/go
  sudo tar -C /usr/local -xzf /tmp/go.tar.gz
  rm -f /tmp/go.tar.gz

  export PATH="/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$HOME/.local/bin"

  refresh_shell_command_cache
  verify_linux_command go
}

install_npm_tools() {
  export PATH="/usr/local/go/bin:$HOME/.local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

  verify_linux_command npm

  if ! npm list -g @github/copilot >/dev/null 2>&1; then
    log "Installerer Copilot CLI"
    sudo npm install -g @github/copilot
  else
    log "Copilot CLI OK"
  fi

  if ! npm list -g @openai/codex >/dev/null 2>&1; then
    log "Installerer Codex CLI"
    sudo npm install -g @openai/codex
  else
    log "Codex CLI OK"
  fi
}

install_devcontainer_cli() {
  export PATH="/usr/local/go/bin:$HOME/.local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

  verify_linux_command npm

  if ! npm list -g @devcontainers/cli >/dev/null 2>&1; then
    log "Installerer devcontainer CLI"
    sudo npm install -g @devcontainers/cli
  else
    log "devcontainer CLI OK"
  fi
}

setup_shell() {
  log "Oppdaterer ~/.bashrc"

  ensure_line_in_file 'export PATH="$HOME/.local/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$PATH"' "$HOME/.bashrc"
  ensure_line_in_file 'alias openai="codex"' "$HOME/.bashrc"

  if is_wsl; then
    ensure_line_in_file '# WSL: prioriter Linux-binærer foran eventuelle Windows PATH-innslag' "$HOME/.bashrc"
  fi
}

verify_tools() {
  log "Verifiserer installasjoner"

  have_cmd git || die "git mangler"
  have_cmd gh || die "gh mangler"
  have_cmd az || die "az mangler"
  have_cmd pwsh || die "pwsh mangler"
  have_cmd dotnet || die "dotnet mangler"
  have_cmd node || die "node mangler"
  have_cmd npm || die "npm mangler"
  have_cmd devcontainer || die "devcontainer mangler"
  have_cmd python3 || die "python3 mangler"
  have_cmd pip3 || die "pip3 mangler"
  have_cmd wget || die "wget mangler"
  have_cmd make || die "make mangler"
  have_cmd uv || die "uv mangler"
  have_cmd go || die "go mangler"

  if is_wsl; then
    verify_linux_command git
    verify_linux_command gh
    verify_linux_command az
    verify_linux_command pwsh
    verify_linux_command dotnet
    verify_linux_command node
    verify_linux_command npm
    verify_linux_command devcontainer
    verify_linux_command python3
    verify_linux_command pip3
    verify_linux_command wget
    verify_linux_command make
    verify_linux_command uv
    verify_linux_command go
  fi
}

print_versions() {
  log "Installerte versjoner"
  echo "git:          $(git --version)"
  echo "gh:           $(gh --version | head -n1)"
  echo "az:           $(az version 2>/dev/null | jq -r '."azure-cli"' 2>/dev/null || echo 'ukjent')"
  echo "pwsh:         $(pwsh --version)"
  echo "dotnet:       $(dotnet --version)"
  echo "sdks:"
  dotnet --list-sdks
  echo "node:         $(node --version)"
  echo "npm:          $(npm --version)"
  echo "devcontainer: $(devcontainer --version)"
  echo "python3:      $(python3 --version)"
  echo "pip3:         $(pip3 --version)"
  echo "wget:         $(wget --version | head -n1)"
  echo "make:         $(make --version | head -n1)"
  echo "uv:           $(uv --version)"
  echo "go:           $(go version)"
  echo "az-path:      $(command_resolved_path az)"
}

print_wsl_recommendations() {
  if ! is_wsl; then
    return
  fi

  log "WSL-notat"
  echo "Hvis Windows PATH fortsatt skaper konflikter, vurder å sette dette i /etc/wsl.conf:"
  echo
  echo "[interop]"
  echo "appendWindowsPath=false"
  echo
  echo "Deretter kjør 'wsl --shutdown' fra Windows PowerShell og start Debian på nytt."
}

main() {
  require_debian
  require_sudo

  warn_if_windows_path_conflicts

  install_base
  install_gh
  install_az
  install_dotnet
  install_pwsh
  install_node
  install_uv
  install_go
  setup_shell

  export PATH="/usr/local/go/bin:$HOME/.local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$PATH"
  refresh_shell_command_cache

  install_npm_tools
  install_devcontainer_cli
  verify_tools
  print_versions
  print_wsl_recommendations

  log "Ferdig 🎉"
  echo "Kjør:"
  echo "  source ~/.bashrc"
  echo "  hash -r"
  echo "  gh auth login"
  echo "  az login"
  echo "  pwsh"
  echo "  dotnet --list-sdks"
  echo "  dotnet new console -o hello-dotnet"
  echo "  cd hello-dotnet && dotnet run"
  echo "  devcontainer --version"
  echo "  copilot auth login"
  echo "  codex"
}

main "$@"