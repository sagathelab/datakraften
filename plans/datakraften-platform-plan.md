# Datakraften Platform Direction

## Working Title

**Datakraften — The AI-Native Developer Workstation Platform**

---

## Executive Summary

Datakraften should evolve from a WSL bootstrap script into a local developer environment platform — an orchestration layer that turns a fresh or unstructured machine into a consistent, productive, and AI-ready development environment.

The goal is to take any machine and quickly make it a powerful, reproducible developer workstation by combining:

- Workstation bootstrapping and automation
- Reproducible development environments
- AI-native developer tooling
- Modern command-line experience
- Team onboarding automation
- Opinionated, production-quality defaults
- Optional support for Devbox, Nix, and NixOS-WSL

WSL is the first and most important target, because no current tooling owns the Windows + WSL developer experience. But the ambition is broader — Datakraften should work wherever developers work.

Datakraften should not try to become another package manager.

Instead, Datakraften should become the orchestration layer above existing tools such as `apt`, `brew`, `fnm`, `uv`, `dotnet-install`, `Devbox`, `Nix`, Docker Desktop, VS Code, Zed, Cursor, Codex, Claude Code, OpenCode, and Gemini CLI.

The platform should answer one question:

> How do I turn this machine into a powerful, reproducible, AI-ready developer workstation — fast?

---

# 1. Product Vision

## 1.1 Vision Statement

**Datakraften is an opinionated AI-native developer workstation platform.**

It helps individuals and teams bootstrap, configure, validate, repair, and reproduce development environments across machines.

WSL is the primary target platform — the gap Datakraften fills is largest there — but the ambition spans Linux, macOS, and anywhere developers work.

Datakraften should become the default entry point for developers who want a professional development environment without manually installing and configuring dozens of tools.

---

## 1.2 Core Positioning

Datakraften is not primarily:

- A package manager
- A dotfiles repository
- A random install script
- A Linux distribution
- A cloud development environment
- A replacement for Nix, Devbox, Homebrew, or Docker

Datakraften is:

- A workstation orchestration layer
- A developer environment bootstrapper (WSL-first, multi-platform ambition)
- An onboarding automation system
- An AI developer environment bootstrapper
- A reproducible environment manager
- A golden-path tool for teams

---

## 1.3 Suggested Taglines

Possible positioning lines:

- **The developer workstation platform**
- **AI-native developer environments, bootstrapped**
- **From fresh machine to productive developer workstation**
- **Reproducible developer workstations, powered by profiles**
- **The golden path for modern developer onboarding**
- **One command to an AI-ready development environment**

---

# 2. Strategic Thesis

## 2.1 Why This Matters

A huge number of developers work on Windows machines but rely on Linux tooling, cloud platforms, containers, Node.js, Python, .NET, GitHub, Azure, and AI-assisted development. The same fragmentation exists on macOS and native Linux — just with different defaults and pain points.

The current setup experience is fragmented:

- Install WSL
- Choose a distro
- Configure Linux packages
- Install Git
- Install Node.js
- Install Python
- Install .NET
- Configure Docker Desktop
- Install VS Code / Cursor / Zed
- Configure GitHub CLI
- Configure Azure CLI
- Install AI tools
- Configure shell
- Configure PATH
- Fix Windows vs Linux command conflicts
- Configure SSH keys
- Authenticate services
- Clone repositories
- Install project dependencies
- Debug broken environment state

Each step is manageable, but together they become slow, inconsistent, and error-prone.

Datakraften should compress this entire experience into a coherent platform.

---

## 2.2 The Gap in the Market

There are strong tools in adjacent spaces:

- Homebrew handles package installation well.
- Nix handles reproducibility extremely well.
- Devbox makes Nix easier at the project level.
- Docker handles containerized services well.
- VS Code Remote WSL gives good editor integration.
- GitHub Codespaces and Gitpod solve cloud dev environments.
- Coder solves remote development infrastructure.
- Microsoft Dev Box solves enterprise cloud workstations.
- AI tools like Codex, Claude Code, Gemini CLI, OpenCode, and Copilot improve productivity.

But there is still a gap:

> There is no widely adopted, opinionated platform that turns a machine into a complete, reproducible, AI-ready developer workstation — regardless of OS.

Datakraften can own this space, starting with Windows + WSL where the gap is largest.

---

# 3. Product Principles

## 3.1 WSL-First, Not WSL-Only

Datakraften is WSL-first because that is where the pain is greatest and the tooling gap is widest. But the platform should be designed from day one to support multiple targets.

On WSL (primary target):

- Linux tools should run inside WSL.
- Windows-backed commands should be detected and avoided when inappropriate.
- Docker Desktop WSL integration should be supported.
- VS Code Remote WSL should be supported.
- Windows-specific setup guidance should be included when needed.
- The system should detect common WSL misconfigurations.

On native Linux and macOS (future targets):

- Detect the native package manager (apt, dnf, pacman, brew).
- Use platform-appropriate runtimes and paths.
- Skip WSL-specific checks gracefully.

The architecture should abstract platform detection so that most code works across targets. WSL-specific logic must be clearly isolated behind interfaces or feature flags, not scattered through every module.

---

## 3.2 Opinionated Defaults

Datakraften should make strong choices.

Developers should be able to install a productive default environment without answering dozens of prompts.

Recommended default stack:

- Shell: Fish
- Prompt: Starship
- Package layers: APT + Homebrew
- Node runtime: fnm
- Python runtime: uv
- .NET runtime: dotnet-install or curated install path
- GitHub: GitHub CLI
- Cloud: Azure CLI
- Containers: Docker Desktop WSL integration
- Editors: VS Code, Zed, Cursor support
- AI tools: Codex, Claude Code, OpenCode, Gemini CLI, GitHub Copilot CLI
- Project reproducibility: Devbox optional
- Full declarative reproducibility: Nix optional
- Advanced OS-level reproducibility: NixOS-WSL optional

Configuration should be possible, but the default path should feel obvious.

---

## 3.3 Orchestrate, Do Not Replace

Datakraften should not compete directly with:

- APT
- Homebrew
- Nix
- Devbox
- Docker
- fnm
- uv
- dotnet-install

Instead, Datakraften should coordinate them.

This gives Datakraften leverage:

- Use the best tool for each layer.
- Avoid maintaining package definitions.
- Focus on user experience.
- Focus on WSL-specific correctness.
- Focus on idempotency, diagnostics, and onboarding.

---

## 3.4 Idempotent by Design

Every Datakraften operation should be safe to run multiple times.

This includes:

- Installing packages
- Updating shell configuration
- Setting environment variables
- Adding PATH entries
- Installing runtimes
- Checking editor integrations
- Setting up Docker access
- Installing AI tools
- Applying profiles
- Running diagnostics

The user should trust:

```bash
dk apply
```

It should converge the system toward the desired state without destructive surprises.

---

## 3.5 Human-Friendly Diagnostics

Datakraften should become excellent at answering:

- What is installed?
- What is missing?
- What is broken?
- What points to Windows instead of Linux?
- What needs authentication?
- What needs a shell restart?
- What needs a Windows-side change?
- What should the user do next?

The `dk doctor` command is central to the product.

---

## 3.6 AI-Native, Not AI-Gimmicky

AI support should not just mean "install a few AI CLIs".

Datakraften should create environments that are ready for AI-assisted development.

That means:

- AI CLI installation
- Authentication guidance
- Standardized config locations
- Project context conventions
- Agent-friendly task files
- Repo bootstrap commands
- Safe shell execution patterns
- Optional local model tooling later
- Integration with project metadata
- Ability to generate environment manifests for AI agents

---

# 4. Platform Architecture

## 4.1 Multi-Platform Architecture

Datakraften should be designed as a layered platform with platform abstraction from the start.

```text
Datakraften CLI
├── Platform Abstraction Layer    ← detects OS, WSL, distro, package managers
├── Bootstrap Layer
├── System Layer
├── Tooling Layer
├── Runtime Layer
├── Shell Layer
├── Editor Layer
├── AI Tooling Layer
├── Reproducibility Layer
├── Team Onboarding Layer
└── Diagnostics Layer
```

The Platform Abstraction Layer is key to the multi-platform strategy. Every module should ask "what platform am I on?" rather than assuming WSL. This means:
- Installers detect the native package manager.
- Path logic adapts to the filesystem layout.
- WSL-specific checks (e.g. Windows-backed commands) are isolated behind platform interfaces.
- New platform targets can be added by implementing the platform interface, not by modifying every module.

---

## 4.2 Bootstrap Layer

Purpose:

Install the `dk` CLI and perform minimal prerequisites.

Example:

```bash
curl -fsSL https://datakraften.no/install | bash
```

Responsibilities:

- Detect WSL
- Detect distro
- Refuse unsafe root execution unless explicitly supported
- Install minimal dependencies
- Download or build the `dk` CLI
- Add `dk` to PATH
- Run `dk doctor` or `dk init`
- Print next steps

The bootstrap script should remain small and auditable.

It should not contain the whole product logic.

---

## 4.3 System Layer

Purpose:

Handle base operating-system dependencies.

Recommended backend:

- APT for Ubuntu/Debian WSL
- Later: DNF for Fedora WSL
- Later: Pacman for Arch WSL

Typical packages:

- `build-essential`
- `curl`
- `wget`
- `git`
- `ca-certificates`
- `gnupg`
- `lsb-release`
- `unzip`
- `tar`
- `jq`
- `procps`
- `locales`

This layer should be conservative and stable.

---

## 4.4 Tooling Layer

Purpose:

Install modern developer tools.

Recommended backend:

- Homebrew on Linux for global developer CLI tools

Typical packages:

- `gh`
- `azure-cli`
- `fnm`
- `uv`
- `atuin`
- `fzf`
- `broot`
- `fd`
- `bottom`
- `starship`
- `powershell`

Recommendation:

Use APT for base system dependencies. Use Homebrew for fast-moving developer tooling.

This should be documented clearly:

```text
APT = operating system foundation
Homebrew = global developer tools
Runtime managers = language runtimes
Devbox/Nix = reproducible project environments
```

---

## 4.5 Runtime Layer

Purpose:

Install and manage language runtimes.

Recommended tools:

- Node.js: `fnm`
- Python: `uv`
- .NET: `dotnet-install` or controlled package backend
- Go: optional
- Rust: optional via `rustup`
- Java: optional via SDKMAN or mise
- Additional runtimes later: Ruby, PHP, Bun, Deno

Avoid relying blindly on distro versions for fast-moving language runtimes.

---

## 4.6 Shell Layer

Purpose:

Provide a modern CLI experience.

Default:

- Fish shell
- Starship prompt
- Atuin shell history
- fzf fuzzy finder
- broot tree navigation
- Useful aliases
- Completion support

Alternative shells:

- Bash
- Zsh

Key requirements:

- Never destroy existing user config.
- Use managed blocks with clear markers.
- Make changes idempotent.
- Support rollback later.

Example managed block:

```bash
# >>> datakraften >>>
eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"
eval "$(fnm env --use-on-cd --shell bash)"
# <<< datakraften <<<
```

---

## 4.7 Editor Layer

Purpose:

Support common modern editors.

Initial support:

- VS Code
- Zed
- Cursor

Responsibilities:

- Detect installed editors
- Detect CLI availability inside WSL
- Provide install guidance where Windows-side install is required
- Configure default editor
- Validate `code`, `zed`, and `cursor` commands
- Support workspace bootstrap later

Example commands:

```bash
dk editor list
dk editor use zed
dk editor doctor
```

---

## 4.8 AI Tooling Layer

Purpose:

Install and validate AI developer tools.

Initial integrations:

- OpenAI Codex CLI
- Claude Code
- OpenCode
- Gemini CLI
- GitHub Copilot CLI

Responsibilities:

- Install tools
- Detect authentication state where possible
- Provide login guidance
- Validate commands
- Manage AI profile
- Create project conventions for AI agents
- Generate task files and environment summaries

Example commands:

```bash
dk ai install
dk ai doctor
dk ai status
dk ai init-project
```

Suggested AI project files:

```text
.ai/
├── agents/
├── tasks/
├── context/
├── plans/
└── runbooks/
```

Possible generated files:

```text
AGENTS.md
DEVELOPMENT.md
ENVIRONMENT.md
TASKS.md
```

---

## 4.9 Reproducibility Layer

Purpose:

Support project-level and machine-level reproducibility.

Three levels:

```text
Level 1: Datakraften profile
Level 2: Devbox project environment
Level 3: Nix / NixOS-WSL full reproducibility
```

### Level 1: Datakraften Profile

A simple YAML/TOML configuration that describes the desired workstation.

Example:

```yaml
version: 1

profile: ai-dotnet

shell:
  default: fish

tools:
  github_cli: true
  azure_cli: true
  docker: true
  starship: true
  atuin: true

runtimes:
  node:
    manager: fnm
    version: lts
  python:
    manager: uv
    version: latest
  dotnet:
    version: latest

editors:
  vscode: true
  zed: true
  cursor: true

ai:
  codex: true
  claude_code: true
  opencode: true
  gemini_cli: true
  github_copilot: true

reproducibility:
  devbox: optional
  nix: optional
```

### Level 2: Devbox

Devbox should be used for reproducible project environments.

Datakraften should be able to generate a `devbox.json`.

Example:

```bash
dk devbox init
dk devbox add node dotnet python azure-cli gh
dk devbox shell
```

### Level 3: Nix / NixOS-WSL

Nix should be advanced mode.

Datakraften should support:

- Installing Nix in WSL
- Generating flake templates
- Installing Devbox
- Creating NixOS-WSL starter configurations
- Validating Nix setup

Example:

```bash
dk nix init
dk nix doctor
dk nixos-wsl init
```

Nix should not be mandatory for default users.

---

## 4.10 Team Onboarding Layer

Purpose:

Enable organizations to define a golden path.

Example:

```bash
curl -fsSL https://company.example/bootstrap | bash
```

Or:

```bash
dk team join company-name
dk apply
```

Team config should describe:

- Required tools
- Required runtimes
- Required repositories
- Required authentication steps
- Required editor extensions
- Required shell defaults
- Required project environment files
- Internal documentation links
- Optional VPN guidance
- Secrets manager guidance

Example:

```yaml
version: 1

team:
  name: platform-team

repositories:
  - name: platform-api
    url: git@github.com:company/platform-api.git
  - name: platform-web
    url: git@github.com:company/platform-web.git

tools:
  - gh
  - azure-cli
  - docker
  - kubectl
  - helm

runtimes:
  node: 22
  dotnet: 9
  python: 3.12

auth:
  github: required
  azure: required

editors:
  vscode:
    extensions:
      - ms-vscode-remote.remote-wsl
      - github.copilot
  cursor: optional

ai:
  codex: enabled
  claude_code: optional
  opencode: enabled
```

---

# 5. Datakraften CLI Design

## 5.1 CLI Name

Primary command:

```bash
dk
```

Longer alias:

```bash
datakraften
```

Both can point to the same binary.

---

## 5.2 Top-Level Commands

Suggested command structure:

```bash
dk init
dk apply
dk doctor
dk status
dk upgrade
dk profile
dk tool
dk runtime
dk shell
dk editor
dk ai
dk devbox
dk nix
dk docker
dk team
dk repo
dk config
dk self
```

---

## 5.3 Core Commands

### `dk init`

Initializes Datakraften configuration.

```bash
dk init
dk init --profile ai
dk init --profile dotnet
dk init --profile frontend
dk init --profile platform
```

Responsibilities:

- Create config file
- Detect current system state
- Ask minimal questions if interactive
- Support non-interactive mode
- Suggest profile

---

### `dk apply`

Applies desired configuration.

```bash
dk apply
dk apply --dry-run
dk apply --profile ai
dk apply --yes
```

Responsibilities:

- Install missing tools
- Configure shell
- Configure runtimes
- Configure editor preferences
- Configure AI tooling
- Print summary
- Remain idempotent

---

### `dk doctor`

Diagnoses the environment.

```bash
dk doctor
dk doctor --json
dk doctor --fix
dk doctor --category docker
```

Checks:

- WSL version
- Distro
- Systemd support
- PATH correctness
- Windows-backed command conflicts
- APT health
- Homebrew health
- Node health
- Python health
- .NET health
- Docker Desktop WSL integration
- GitHub CLI authentication
- Azure CLI authentication
- AI tool availability
- Editor CLI availability
- Shell config health

This should be one of the best features in the product.

---

### `dk status`

Shows a friendly overview.

```bash
dk status
```

Example output:

```text
Datakraften status

WSL
  ✓ WSL2 detected
  ✓ Ubuntu 24.04
  ✓ systemd available

Tools
  ✓ git
  ✓ gh
  ✓ az
  ✓ docker
  ✓ fnm
  ✓ uv

Runtimes
  ✓ Node.js 22
  ✓ Python 3.12
  ✓ .NET 9

AI
  ✓ codex
  ✓ opencode
  – claude-code not installed
  – gemini not authenticated

Editors
  ✓ code
  ✓ zed
  – cursor not found
```

---

### `dk upgrade`

Upgrades Datakraften and optionally managed tools.

```bash
dk upgrade
dk upgrade --self
dk upgrade --tools
dk upgrade --all
```

---

# 6. Profiles

## 6.1 Why Profiles Matter

Profiles make Datakraften opinionated but flexible.

A profile represents a curated developer setup.

Profiles should be easy to understand:

```bash
dk profile list
dk profile use ai
dk apply
```

---

## 6.2 Suggested Built-In Profiles

### `minimal`

For users who want only the basics.

Includes:

- Git
- Curl
- Build tools
- Homebrew
- Fish optional
- Starship optional

---

### `default`

Recommended general developer setup.

Includes:

- System dependencies
- Homebrew
- GitHub CLI
- Azure CLI
- fnm
- uv
- .NET
- Fish
- Starship
- Atuin
- fzf
- Docker checks
- VS Code checks

---

### `frontend`

For frontend developers.

Includes:

- Node.js LTS
- fnm
- pnpm
- npm
- bun optional
- VS Code / Cursor / Zed
- AI tools
- Browser tooling guidance

---

### `dotnet`

For .NET developers.

Includes:

- .NET SDK
- PowerShell
- Azure CLI
- GitHub CLI
- Docker integration
- Node.js optional
- SQL tooling optional
- VS Code / Rider / Zed guidance

---

### `python`

For Python developers.

Includes:

- uv
- Python
- Ruff
- pytest
- ipython optional
- Jupyter optional

---

### `platform`

For platform engineers.

Includes:

- Docker
- kubectl
- helm
- k9s
- terraform or opentofu
- azure-cli
- gh
- jq
- yq
- direnv optional

---

### `ai`

For AI-native development.

Includes:

- Codex
- Claude Code
- OpenCode
- Gemini CLI
- GitHub Copilot CLI
- Node.js
- Python
- uv
- pnpm
- project AI conventions
- AGENTS.md generator

---

### `reproducible`

For Devbox/Nix-based environments.

Includes:

- Devbox
- Nix optional
- direnv
- flake templates
- devbox templates

---

### `nixos-wsl`

For advanced users who want NixOS in WSL.

Includes:

- NixOS-WSL guidance
- starter config
- flake template
- Home Manager pattern
- migration documentation

---

# 7. Configuration Format

## 7.1 Recommended Format

Use YAML for user-facing configuration.

Reason:

- Human-readable
- Common in DevOps
- Easy for AI agents to edit
- Works well for team configs
- Familiar from GitHub Actions, Kubernetes, Docker Compose, Azure DevOps

TOML is also good, but YAML is more natural for nested environment definitions.

Recommended file:

```text
datakraften.yaml
```

Alternative user config location:

```text
~/.config/datakraften/config.yaml
```

Project config:

```text
./datakraften.yaml
```

Team config:

```text
./datakraften.team.yaml
```

---

## 7.2 Example User Config

```yaml
version: 1

profile: default

system:
  package_manager: apt

tooling:
  package_manager: brew

shell:
  default: fish
  prompt: starship
  history: atuin
  fuzzy_finder: fzf

runtimes:
  node:
    enabled: true
    manager: fnm
    version: lts
  python:
    enabled: true
    manager: uv
    version: latest
  dotnet:
    enabled: true
    version: latest

tools:
  git: true
  github_cli: true
  azure_cli: true
  powershell: true
  docker: true

editors:
  vscode: true
  zed: true
  cursor: optional

ai:
  codex: true
  claude_code: optional
  opencode: true
  gemini_cli: optional
  github_copilot: true

reproducibility:
  devbox: optional
  nix: optional
```

---

# 8. Technical Stack Recommendation

## 8.1 Bootstrap Script

Use:

```text
Bash
```

Reason:

- Universal on Linux/WSL
- Easy to inspect
- Good for initial install
- Works before the main CLI exists

Responsibilities should be minimal.

The bootstrap script should install the real CLI and delegate everything else.

---

## 8.2 CLI Implementation Language

Recommended:

```text
Go
```

Alternative:

```text
Rust
```

### Why Go is probably the best first choice

Go is a strong fit because:

- Easy static binaries
- Fast compile times
- Simple deployment
- Good CLI ecosystem
- Easy cross-compilation
- Easier onboarding than Rust
- Good enough performance
- Strong standard library
- Excellent for system orchestration

Suggested Go libraries:

- Cobra for CLI commands
- Viper for config
- Bubble Tea for richer terminal UI later
- Lip Gloss for terminal styling
- Survey or Huh for interactive prompts
- Zerolog or slog for logging

### Why Rust is also attractive

Rust is strong because:

- Excellent safety
- Great performance
- Strong CLI ecosystem
- Feels premium for developer tooling

Suggested Rust libraries:

- clap
- serde
- tokio
- anyhow
- tracing
- inquire
- console
- indicatif

### Recommendation

Start with Go unless the team strongly prefers Rust.

The product value is orchestration, diagnostics, and UX — not low-level memory safety.

---

## 8.3 State Management

Datakraften should keep local state.

Suggested location:

```text
~/.local/state/datakraften/state.json
```

State can track:

- Datakraften version
- Last apply time
- Installed profile
- Managed tools
- Warnings
- Shell config modifications
- Last doctor result
- Installation logs

---

## 8.4 Logs

Suggested location:

```text
~/.local/state/datakraften/logs/
```

Commands:

```bash
dk logs
dk logs last
dk logs doctor
```

---

## 8.5 Cache

Suggested location:

```text
~/.cache/datakraften/
```

Used for:

- Downloaded installers
- Metadata
- Package indexes
- Templates
- Generated manifests

---

# 9. Desired CLI UX

## 9.1 Install

```bash
curl -fsSL https://datakraften.no/install | bash
```

Expected result:

```text
✓ Datakraften installed
✓ dk available
→ Run: dk init
```

---

## 9.2 First Run

```bash
dk init
```

Expected experience:

```text
Welcome to Datakraften

Detected:
  ✓ WSL2
  ✓ Ubuntu 24.04
  ✓ Windows host
  ✓ Docker Desktop not connected
  ✓ VS Code available
  – Cursor not found

Choose profile:
  1. Default
  2. AI Developer
  3. .NET Developer
  4. Frontend Developer
  5. Platform Engineer
  6. Reproducible / Devbox
```

---

## 9.3 Apply

```bash
dk apply
```

Expected output:

```text
Applying Datakraften profile: ai

System
  ✓ apt dependencies

Tooling
  ✓ Homebrew
  ✓ GitHub CLI
  ✓ Azure CLI

Runtimes
  ✓ Node.js LTS via fnm
  ✓ Python via uv
  ✓ .NET SDK

AI
  ✓ Codex CLI
  ✓ OpenCode
  ✓ GitHub Copilot CLI
  – Claude Code requires manual authentication

Editors
  ✓ VS Code WSL command
  ✓ Zed

Next steps:
  1. Run gh auth login
  2. Run az login
  3. Restart shell: exec fish
```

---

## 9.4 Doctor

```bash
dk doctor
```

Expected output:

```text
Datakraften Doctor

Critical
  ✓ No critical issues found

Warnings
  – docker command exists, but Docker Desktop WSL integration is not active
  – node points to Windows path before fnm initialization
  – cursor CLI not found

Suggested fixes
  dk doctor --fix shell
  Enable Docker Desktop WSL integration
```

---

# 10. Agent-Friendly Development Plan

This section is written specifically so AI coding agents can implement the platform incrementally.

---

## 10.1 Repository Structure

Suggested monorepo:

```text
datakraften/
├── README.md
├── LICENSE
├── install.sh
├── cmd/
│   └── dk/
│       └── main.go
├── internal/
│   ├── app/
│   ├── config/
│   ├── doctor/
│   ├── exec/
│   ├── installers/
│   ├── profiles/
│   ├── shell/
│   ├── system/
│   ├── tools/
│   ├── runtimes/
│   ├── ai/
│   ├── editors/
│   ├── docker/
│   ├── devbox/
│   ├── nix/
│   └── telemetry/
├── profiles/
│   ├── minimal.yaml
│   ├── default.yaml
│   ├── ai.yaml
│   ├── dotnet.yaml
│   ├── frontend.yaml
│   ├── platform.yaml
│   └── reproducible.yaml
├── templates/
│   ├── devbox/
│   ├── nix/
│   ├── agents/
│   └── shell/
├── docs/
│   ├── vision.md
│   ├── architecture.md
│   ├── cli.md
│   ├── profiles.md
│   ├── wsl.md
│   ├── ai-tooling.md
│   ├── devbox.md
│   └── nixos-wsl.md
└── tests/
```

---

## 10.2 MVP Goal

The MVP should prove that Datakraften can replace the current shell script with a structured CLI, while establishing the platform abstraction needed for future targets.

MVP commands:

```bash
dk init
dk apply
dk doctor
dk status
dk upgrade --self
```

MVP profile:

```bash
default
```

MVP platform target:

- **Primary**: Ubuntu/Debian WSL
- **Architecture**: Platform abstraction layer designed so native Linux and macOS can be added without rewriting modules

MVP support:

- Ubuntu/Debian WSL (primary target)
- Fedora/Arch Linux detection (structure ready, installation deferred)
- macOS detection (structure ready, installation deferred)
- APT base dependencies
- Homebrew install
- Brew packages
- fnm + Node LTS
- uv + Python
- GitHub CLI
- Azure CLI
- Fish shell
- Starship
- Atuin
- fzf
- Zed check/install
- VS Code WSL check
- Docker Desktop WSL check
- Codex/OpenCode installation
- GitHub Copilot CLI installation

---

## 10.3 Implementation Milestones

### Milestone 1 — Extract bootstrapper

Goal:

Move the current bash script into a cleaner install flow.

Tasks:

- Create `install.sh`
- Install or download `dk`
- Keep existing script as fallback
- Add version banner
- Add safe error handling
- Add WSL detection
- Add non-root check

Acceptance criteria:

- `curl -fsSL https://datakraften.no/install | bash` installs `dk`
- `dk --version` works
- `dk doctor` prints a basic report

---

### Milestone 2 — Build basic CLI

Tasks:

- Create Go CLI
- Add commands:
  - `dk init`
  - `dk apply`
  - `dk doctor`
  - `dk status`
- Add config parser
- Add profile loader
- Add command runner abstraction
- Add dry-run support

Acceptance criteria:

- CLI can load `datakraften.yaml`
- CLI can run dry-run
- CLI prints structured output
- CLI exits with meaningful status codes

---

### Milestone 3 — System and tooling installers

Tasks:

- Implement APT installer
- Implement Homebrew installer
- Implement package detection
- Implement Linux vs Windows command detection
- Implement idempotent package install
- Implement logs

Acceptance criteria:

- Re-running `dk apply` does not reinstall existing tools unnecessarily
- Windows-backed commands are detected
- Missing tools are installed
- Logs are stored

---

### Milestone 4 — Runtime management

Tasks:

- Implement fnm installer
- Implement Node LTS setup
- Implement uv installer
- Implement Python setup
- Implement .NET SDK setup
- Add runtime doctor checks

Acceptance criteria:

- `node`, `npm`, `python`, `uv`, and `dotnet` are available inside WSL
- Commands point to Linux paths
- Runtime versions are visible in `dk status`

---

### Milestone 5 — Shell experience

Tasks:

- Implement Fish setup
- Implement Starship setup
- Implement Atuin setup
- Implement fzf setup
- Add managed shell config blocks
- Avoid overwriting existing config
- Add rollback metadata

Acceptance criteria:

- Fish can be set as default shell
- Shell config is idempotent
- Managed blocks are clearly marked
- `dk doctor shell` detects broken config

---

### Milestone 6 — Editor and Docker integration

Tasks:

- Detect VS Code CLI
- Detect Zed CLI
- Detect Cursor CLI
- Set default editor
- Detect Docker CLI
- Detect Docker socket
- Detect Docker Desktop WSL integration
- Provide Windows-side instructions when needed

Acceptance criteria:

- `dk doctor editor` reports editor status
- `dk doctor docker` reports Docker Desktop integration status
- User receives actionable next steps

---

### Milestone 7 — AI tooling profile

Tasks:

- Add `ai` profile
- Install Codex CLI
- Install OpenCode
- Install Gemini CLI
- Install Claude Code if install method is available
- Install GitHub Copilot CLI
- Add authentication checks where possible
- Generate `AGENTS.md`
- Generate `.ai/` project structure

Acceptance criteria:

- `dk profile use ai && dk apply` installs AI tools
- `dk ai doctor` reports install/auth status
- `dk ai init-project` creates agent-friendly project files

---

### Milestone 8 — Devbox integration

Tasks:

- Add Devbox installation
- Add `dk devbox init`
- Generate `devbox.json`
- Add templates:
  - Node
  - Python
  - .NET
  - AI
  - Platform
- Add `dk devbox doctor`

Acceptance criteria:

- A project can get a generated `devbox.json`
- `devbox shell` works after Datakraften setup
- Devbox is positioned as project-level reproducibility

---

### Milestone 9 — Nix integration

Tasks:

- Add optional Nix installer
- Add flake templates
- Add `dk nix doctor`
- Add `dk nix init`
- Document Nix tradeoffs
- Avoid making Nix required for default users

Acceptance criteria:

- Advanced users can opt into Nix
- Datakraften can generate a starter flake
- Nix status is visible in `dk doctor`

---

### Milestone 10 — NixOS-WSL mode

Tasks:

- Create NixOS-WSL documentation
- Create starter flake/configuration templates
- Add `dk nixos-wsl init`
- Add migration guide from Ubuntu WSL to NixOS-WSL
- Add warnings about advanced complexity

Acceptance criteria:

- Users can follow a documented path to NixOS-WSL
- Datakraften can generate a starter NixOS-WSL config
- This remains an advanced mode, not the default

---

# 11. Current Script Evolution Plan

The existing script should not be thrown away immediately.

It should be evolved in stages.

## 11.1 Current Script Strengths

The current script already has several strong foundations:

- WSL awareness
- Non-root protection
- APT base dependencies
- Homebrew installation
- Modern CLI package list
- Node via fnm
- Python via uv
- AI tooling installation
- Zed installation
- GitHub Copilot CLI setup
- VS Code WSL check
- Docker Desktop WSL integration check
- Fish shell setup
- Final verification summary

This is a strong prototype.

---

## 11.2 Current Script Weaknesses

Areas to improve:

- Too much logic in one bash file
- No structured config
- No profile system
- No dry-run mode
- No rollback model
- No formal state tracking
- No command-level UX
- No modular installer abstraction
- No testability
- No clear separation between bootstrap and apply
- Homebrew is required rather than a chosen tooling layer
- Docker install strategy should be more careful
- AI tool install methods may change over time
- Shell config should use managed blocks
- There is no `doctor` command yet

---

## 11.3 Migration Path

### Step 1

Keep current script as:

```text
legacy/debian.sh
```

### Step 2

Create:

```text
install.sh
```

This installs the `dk` CLI.

### Step 3

Implement `dk apply` to do what the old script did.

### Step 4

Move each function from bash into a typed module.

Example mapping:

```text
install_homebrew       -> internal/installers/homebrew
setup_node             -> internal/runtimes/node
setup_python           -> internal/runtimes/python
setup_docker           -> internal/docker
setup_fish             -> internal/shell/fish
verify_all_tools       -> internal/doctor
```

### Step 5

Deprecate direct `debian.sh` install.

Final user flow:

```bash
curl -fsSL https://datakraften.no/install | bash
dk init
dk apply
dk doctor
```

---

# 12. Data Model

## 12.1 Tool Definition

Internal representation:

```yaml
name: gh
display_name: GitHub CLI
category: developer-tool
commands:
  - gh
install:
  apt: null
  brew: gh
  script: null
check:
  command: gh --version
auth:
  command: gh auth status
```

---

## 12.2 Runtime Definition

```yaml
name: node
manager: fnm
versions:
  default: lts
commands:
  - node
  - npm
install:
  manager: fnm
check:
  command: node --version
```

---

## 12.3 Doctor Check Definition

```yaml
id: docker.socket
title: Docker socket available
category: docker
severity: warning
check: test -S /var/run/docker.sock
fix: manual
message: Docker Desktop WSL integration is not active.
```

---

# 13. Safety and Trust

## 13.1 Install Script Transparency

Because users will run:

```bash
curl -fsSL https://datakraften.no/install | bash
```

Trust is crucial.

Required:

- The script must be readable.
- The script must be small.
- The script must print what it is doing.
- The script must avoid hidden telemetry.
- The script must avoid destructive operations.
- The script should support `--dry-run` where possible.
- The website should explain what gets installed.

---

## 13.2 Security Practices

Recommended:

- Publish source code on GitHub
- Use signed releases
- Provide checksums
- Pin installer URLs where practical
- Avoid executing remote scripts without clear reason
- Minimize `sudo`
- Explain every privileged operation
- Add `dk doctor security`
- Add supply-chain notes
- Add privacy policy if telemetry is added

---

## 13.3 Telemetry

Default recommendation:

No telemetry in early versions.

Later, optional telemetry may be useful, but it must be:

- Opt-in
- Clearly documented
- Easy to disable
- Anonymous where possible
- Never collect secrets, paths, repo names, usernames, tokens, or command history

Possible command:

```bash
dk telemetry enable
dk telemetry disable
dk telemetry status
```

---

# 14. Website Direction

## 14.1 Homepage Message

Suggested hero:

```text
Datakraften

The WSL-first developer workstation platform.

Bootstrap a modern, reproducible, AI-ready development environment on Windows + WSL with one command.
```

CTA:

```bash
curl -fsSL https://datakraften.no/install | bash
```

---

## 14.2 Website Sections

Recommended sections:

1. What is Datakraften?
2. Why WSL?
3. Install
4. Profiles
5. AI Tooling
6. Reproducible Environments
7. Devbox and Nix
8. Team Onboarding
9. Doctor / Diagnostics
10. Security and Transparency
11. Roadmap
12. GitHub

---

## 14.3 Documentation Pages

Recommended docs:

```text
/docs
/docs/install
/docs/getting-started
/docs/profiles
/docs/commands
/docs/wsl
/docs/docker
/docs/editors
/docs/ai
/docs/devbox
/docs/nix
/docs/nixos-wsl
/docs/team-onboarding
/docs/configuration
/docs/security
/docs/troubleshooting
```

---

# 15. Brand Direction

## 15.1 Brand Personality

Datakraften should feel:

- Technical
- Trustworthy
- Sharp
- Practical
- Nordic
- Developer-first
- Modern
- Slightly playful
- Not corporate-heavy

---

## 15.2 Tone of Voice

Use language like:

- "Your WSL workstation, ready for serious development."
- "Modern developer tooling without the setup pain."
- "AI-ready from the first shell."
- "Reproducible when you need it. Simple when you do not."
- "Opinionated defaults. Escape hatches included."

Avoid:

- Too much hype
- Overpromising
- Enterprise jargon too early
- Claiming to replace Nix, Devbox, or Docker

---

# 16. Roadmap

## Phase 0 — Current Prototype

Status:

- Bash bootstrapper
- WSL-aware
- Homebrew-based tooling
- AI tools
- Fish shell
- Docker and VS Code checks

Goal:

Stabilize and document.

---

## Phase 1 — CLI MVP

Deliver:

- `dk` CLI
- `dk init`
- `dk apply`
- `dk doctor`
- `dk status`
- `default` profile
- Config file support
- Legacy script parity

---

## Phase 2 — Profiles

Deliver:

- `minimal`
- `default`
- `ai`
- `dotnet`
- `frontend`
- `platform`
- Profile switching
- Profile documentation

---

## Phase 3 — AI Developer Environment

Deliver:

- AI profile
- `dk ai doctor`
- `dk ai init-project`
- AGENTS.md generator
- `.ai/` project structure
- Tool auth checks
- AI-ready project templates

---

## Phase 4 — Devbox Integration

Deliver:

- Devbox install
- `dk devbox init`
- Templates
- Project reproducibility docs
- Example repos

---

## Phase 5 — Nix Integration

Deliver:

- Optional Nix install
- Flake templates
- `dk nix doctor`
- Nix docs
- Power-user guide

---

## Phase 6 — NixOS-WSL Mode

Deliver:

- NixOS-WSL guide
- Starter configs
- Migration guide
- Advanced reproducible workstation mode

---

## Phase 7 — Team Onboarding

Deliver:

- Team config
- Repository bootstrap
- Auth checklist
- Organization profile
- Internal docs linking
- Team onboarding report

---

## Phase 8 — Commercial Possibilities

Potential future offerings:

- Hosted team profile registry
- Private organization templates
- Enterprise workstation compliance checks
- Internal developer platform onboarding
- Managed AI tooling setup
- Consulting around WSL developer platforms
- Paid support for teams
- Company-branded bootstrap endpoints

---

# 17. Commercial Positioning

## 17.1 Individual Developers

Value:

- Fast setup
- Better WSL experience
- Modern CLI tools
- AI-ready development
- Less manual configuration

---

## 17.2 Teams

Value:

- Faster onboarding
- Fewer environment issues
- Reproducible setup
- Standardized tooling
- Better AI development workflows
- Less senior developer time spent helping new team members

---

## 17.3 Companies

Value:

- Lower onboarding cost
- Consistent developer workstations
- More secure setup process
- Faster project ramp-up
- Better platform engineering practices
- A clearer internal developer experience

---

# 18. Key Differentiators

Datakraften should be differentiated by:

1. WSL-first focus
2. AI-native tooling
3. Opinionated developer workstation setup
4. Hybrid simplicity and reproducibility
5. Strong diagnostics
6. Team onboarding automation
7. Practical use of existing tools
8. Optional Devbox/Nix power mode
9. Developer-friendly brand
10. Clear upgrade path from simple bootstrap to serious platform

---

# 19. Risks

## 19.1 Too Much Scope

Risk:

Trying to support every tool and every workflow too early.

Mitigation:

Start with a small number of curated profiles.

---

## 19.2 Becoming a Package Manager

Risk:

Maintaining package metadata and installers for everything.

Mitigation:

Use APT, Homebrew, Devbox, Nix, and official installers as backends.

---

## 19.3 Nix Complexity

Risk:

Making the product feel too difficult.

Mitigation:

Keep Nix optional and advanced.

---

## 19.4 AI Tool Churn

Risk:

AI CLIs change quickly.

Mitigation:

Abstract AI tool installation and diagnostics behind modules.

---

## 19.5 Security Concerns

Risk:

Users hesitate to run remote install scripts.

Mitigation:

Small install script, open source, checksums, signed releases, clear docs.

---

# 20. Immediate Next Actions

## 20.1 Product

- Define the first three profiles:
  - `minimal`
  - `default`
  - `ai`
- Write the homepage positioning
- Write installation documentation
- Write security/transparency page
- Write roadmap page

---

## 20.2 Engineering

- Create `install.sh`
- Create `dk` CLI skeleton
- Implement:
  - `dk --version`
  - `dk doctor`
  - `dk status`
  - `dk init`
  - `dk apply --dry-run`
- Port current script functions into modules
- Add config loading
- Add profile loading
- Add logs
- Add state file

---

## 20.3 AI Agent Task Breakdown

Give AI agents these initial tasks:

### Agent Task 1 — CLI Skeleton

Build a Go CLI named `dk` with commands:

- `init`
- `apply`
- `doctor`
- `status`
- `profile`
- `version`

Include structured output, error handling, and dry-run plumbing.

---

### Agent Task 2 — Config and Profiles

Implement YAML config loading.

Create built-in profiles:

- `minimal`
- `default`
- `ai`

Support:

```bash
dk profile list
dk profile use ai
dk config show
```

---

### Agent Task 3 — System Detection

Implement detection for:

- WSL
- WSL version
- Linux distro
- systemd
- current shell
- Windows-backed commands
- PATH conflicts
- Docker socket
- editor CLIs

---

### Agent Task 4 — Installer Modules

Create installer modules for:

- APT dependencies
- Homebrew
- Brew packages
- fnm
- uv
- Node.js
- Python
- Fish
- Starship
- Atuin
- fzf

All installers must be idempotent.

---

### Agent Task 5 — Doctor Checks

Implement doctor checks with categories:

- system
- package-managers
- runtimes
- shell
- docker
- editors
- ai
- auth

Return results as both human text and JSON.

---

### Agent Task 6 — AI Profile

Implement AI tooling support:

- Codex
- OpenCode
- GitHub Copilot CLI
- Gemini CLI
- Claude Code where feasible

Add:

```bash
dk ai doctor
dk ai init-project
```

Generate:

```text
AGENTS.md
.ai/tasks/
.ai/context/
.ai/runbooks/
```

---

### Agent Task 7 — Devbox Integration

Implement:

```bash
dk devbox init
dk devbox doctor
```

Generate `devbox.json` from Datakraften config.

---

### Agent Task 8 — Documentation

Write documentation for:

- Install
- CLI commands
- Profiles
- WSL assumptions
- Docker Desktop WSL integration
- AI tooling
- Devbox
- Nix
- NixOS-WSL
- Security

---

# 21. Definition of Success

Datakraften is successful when a developer can start with a fresh or unstructured machine and quickly reach this state:

On WSL (primary target):

- WSL installed and healthy
- Linux developer tools installed
- Node.js working
- Python working
- .NET working
- Docker working
- GitHub CLI authenticated
- Azure CLI authenticated
- AI developer tools installed
- Modern shell configured
- Editor integration working
- Project environment reproducible
- Team onboarding documented
- Diagnostics show a clean system

On native Linux and macOS (future targets):

- A comparable set of tools installed using the native package manager
- Platform-appropriate paths and defaults
- No WSL-specific errors or confusion
- Same profile system, same `dk` commands, same AI tooling

The experience should feel like:

```bash
curl -fsSL https://datakraften.no/install | bash
dk init --profile ai
dk apply
dk doctor
```

And the result should be:

> A professional, modern, AI-ready developer workstation — on any machine.

---

# 22. Final Strategic Recommendation

Build Datakraften as a CLI-first platform.

Keep the install experience simple:

```bash
curl -fsSL https://datakraften.no/install | bash
```

Make `dk doctor` excellent.

Make profiles opinionated.

Keep Devbox and Nix optional, but first-class.

Position Datakraften as the bridge between:

- local developer workstations
- reproducible environments
- AI-native tooling
- team onboarding
- modern platform engineering

WSL is where Datakraften starts and dominates. Native Linux and macOS follow when the platform abstraction is proven.

This is a strong and differentiated direction.

Do not build a package manager.

Build the control plane for the modern developer workstation.
