# Datakraften 2.0

> **The missing bootstrap layer for AI-native developer workstations.**

Datakraften is the opinionated, declarative workstation bootstrapper that turns a fresh WSL/Linux environment into a complete, reproducible, AI-ready developer workstation in minutes.

It orchestrates existing tools (apt, brew, fnm, uv, dotnet) rather than replacing them, and treats AI coding tools as first-class citizens alongside runtimes, shell, editors, and Git.

---

## 0. State of Datakraften Today (1.0)

Datakraften 1.0 proved the concept: a Go CLI that detects your platform, reads a declarative YAML config, and idempotently installs a full developer workstation.

### Shipped in 1.0

| Area | Status |
|------|--------|
| 6 CLI commands: `init`, `apply`, `doctor`, `status`, `upgrade`, `update` | ✅ |
| System detection (WSL, distro, systemd, shell) | ✅ |
| Multi-PM (APT, DNF, YUM, PACMAN, BREW) | ✅ |
| Runtimes: Node (fnm), Python (uv), Go (brew), .NET (brew) | ✅ |
| Fish shell with managed config blocks (`# >>> datakraften >>>`) | ✅ |
| Editor detection (VS Code, Zed, Cursor) | ✅ |
| Docker Desktop WSL integration check | ✅ |
| AI tools: 5 CLI + 3 desktop apps | ✅ |
| Self-upgrade with SHA256 verification | ✅ |
| Tool updates (`dk update` with `--list`, `--dry-run`) | ✅ |
| Bootstrap install script (`curl ... | bash`) | ✅ |
| Config sources: `source: default \| custom \| team` | ✅ |
| `--dry-run` and `--json` global flags | ✅ |
| Website with docs (React + TypeScript + Tailwind v4) | ✅ |
| CI/CD: Go CI, cross-compile release, site deploy | ✅ |
| Linting: golangci-lint, ESLint, stylelint, Prettier, TypeScript strict | ✅ |

### Gaps that 2.0 must close

| Gap | Priority |
|-----|----------|
| `dk doctor --fix` flag exists but is never wired | High |
| No Bash/Zsh shell support (only Fish) | High |
| Config validation is minimal (no schema checks on load) | High |
| `internal/profiles/` and `internal/tools/` are empty stubs | Medium |
| No unit or integration tests in most packages | Medium |
| `.NET` install path fragile (brew only) | Medium |
| `Config.Tools` and `Config.Editors` maps exist but are not wired to logic | Medium |
| No `dk plan` command (only `--dry-run` on apply) | Low |
| macOS support is detected but not fully wired | Low |
| No `dk remove` for clean uninstall | Low |

---

## 1. Core Values

These values guide every decision in Datakraften 2.0.

### AI-native by default
AI coding tools are not an optional plugin or afterthought. They belong in the same config as runtimes, editors, and shell. Datakraften should create environments that are ready for AI-assisted development from the first shell.

### Opinionated, not rigid
Strong defaults that work immediately. Escape hatches for those who need them. The default path should feel obvious; customization should be possible without fighting the tool.

### Safe by default
Every operation must be idempotent. Never destroy existing config. Never run without transparency. `dk apply` should converge toward the desired state without destructive surprises.

### Workstation-level
Datakraften focuses on the whole machine, not one layer. Project-specific tools (dev containers, Devbox, npm scripts) complement Datakraften but do not replace it.

### Declarative > imperative
Describe the desired workstation state in YAML. The tool figures out what to install, configure, and verify. No shell scripts, no tribal knowledge, no README steps to follow manually.

### Trust through transparency
Open source, signed releases, SHA256 checksums, readable install script, clear output, no hidden telemetry. Users should understand what Datakraften does and why.

---

## 2. Target Users

### Individual developers
Developers who frequently set up new machines, WSL environments, cloud workstations, or test environments. They want to go from fresh install to productive in minutes, not hours.

### Consultants
Consultants who move between customers, projects, and environments. They need to become productive quickly on each new machine without re-doing setup from scratch.

### Platform teams
Teams that want to standardize the developer workstation experience across an organization. They define the shared baseline once; every developer gets the same environment.

### Fullstack developers
Developers working across frontend, backend, APIs, databases, Docker, cloud, and CI/CD. They need a broad set of tools installed and configured coherently.

### AI-native developers
Developers who use AI coding tools as a normal part of their workflow. They want Codex, OpenCode, Copilot, Claude Code, and Gemini installed, authenticated, and ready to use — without reading five different install guides.

---

## 3. Positioning

### The gap Datakraften fills

| Category | Examples | What they solve | Datakraften's role |
|----------|----------|-----------------|-------------------|
| Dotfiles | chezmoi, yadm, GNU Stow | Personal config files | Can manage or integrate workstation-level config |
| Runtime managers | mise, asdf, nvm, pyenv | Tool and language versions | Can configure and orchestrate runtime installation |
| Dev environments | Devbox, Flox, devenv | Project or shell environments | Can complement them at workstation level |
| Dev containers | Dev Containers, Codespaces | Containerized project environments | Can install tooling and prepare the host |
| Package managers | apt, Homebrew, npm, pip | Package installation | Can declaratively drive them |
| Provisioning | Ansible, shell scripts | Automation and machine setup | Provides a developer-first profile model |

Datakraften sits above these layers and connects them into one coherent developer experience.

### Core differentiator

The unique focus is not package installation alone — it is:

> **A complete, opinionated, AI-native developer workstation baseline.**

Datakraften combines workstation bootstrap, developer experience, AI tooling, WSL/Linux focus, declarative configuration, and safe idempotent execution — into one product.

---

## 4. Core Promise

```text
Fresh WSL/Linux environment
        ↓
Datakraften preset
        ↓
Complete AI-native developer workstation
```

A developer should be able to run:

```bash
curl -fsSL https://datakraften.no/install | bash
dk init --preset ai-native
dk apply
dk doctor
```

and end up with a ready-to-use workstation that feels consistent, modern, fast, and ready for real work.

---

## 5. First-Class Target Environment

The primary target is:

```text
WSL on Windows (Ubuntu, Fedora)
```

with strong support for:

```text
Ubuntu/Debian (native Linux)
Fedora (native Linux)
macOS (experimental)
```

WSL is the starting point because many developers work on Windows machines but want a Linux-native development experience, and no existing tool owns this gap well.

---

## 6. Config Model 2.0

### Two orthogonal concepts

**`source:`** — where the configuration comes from (already shipped in 1.0)

| Source | Meaning |
|--------|---------|
| `default` | Built-in config embedded in the `dk` binary |
| `custom` | Local file owned and edited by the user |
| `team` | Thin pointer to a remote YAML URL (fetched fresh on every `dk apply`) |

**`preset:`** — which built-in template was used as a starting point (new in 2.0)

A preset is a named built-in template that generates a `source: custom` config. Once generated, the user owns the file and can edit it freely.

```bash
dk init --preset ai-native     # Creates custom config from ai-native template
dk init --team <url>           # Creates team config (unchanged from 1.0)
dk init                        # Creates default config (unchanged from 1.0)
```

### Config schema

```yaml
version: 2
source: custom
preset: ai-native

system_packages:
  - build-essential
  - curl
  - git
  - unzip

brew_packages:
  - fish
  - starship
  - atuin
  - fzf
  - gh
  - docker

runtimes:
  node:
    enabled: true
    version: lts
  python:
    enabled: true
    version: latest
  go:
    enabled: true
    version: latest
  dotnet:
    enabled: false

shell:
  fish:
    enabled: true
    default: true
    managed_config: true

editors:
  vscode:
    enabled: true
  zed:
    enabled: true
  cursor:
    enabled: false

ai_tools:
  codex:
    enabled: true
  opencode:
    enabled: true
  copilot:
    enabled: true
  claude:
    enabled: false
  gemini:
    enabled: false

ai_apps:
  codex:
    enabled: true
  copilot:
    enabled: true
  claude:
    enabled: false

git:
  user_name: null
  user_email: null
  default_branch: main
  pull_rebase: false
  signing:
    enabled: false

github:
  cli:
    enabled: true
  auth:
    check: true

containers:
  docker:
    enabled: true
    compose: true
  devcontainers:
    enabled: false

databases:
  postgresql:
    client: false
    server: false
  redis:
    client: false
    server: false

cloud:
  azure:
    cli: false
  aws:
    cli: false
  gcloud:
    cli: false
```

### Schema principles

- Maps/objects for configurable features (not flat booleans)
- Explicit `enabled:` — no magical implicit behavior
- New sections (`git`, `github`, `containers`, `databases`, `cloud`) are optional and disabled by default
- Existing keys (`system_packages`, `brew_packages`, etc.) remain unchanged
- Validate on load — reject unknown keys and invalid values before applying

---

## 7. Presets / Built-in Templates

Presets are opinionated starting points. They are not locked — once generated, the user owns the config.

| Preset | Audience | Includes |
|--------|----------|----------|
| `minimal` | Minimalists | Shell, Git, curl, build tools |
| `default` | General developers | Runtimes, shell tools, Docker, editors, AI tools |
| `ai-native` | AI-first developers | All AI CLI + desktop tools, runtimes, shell, editors |
| `dotnet-fullstack` | .NET teams | .NET SDK, Node.js, PostgreSQL client, Docker, VS Code C# tools |
| `frontend` | Frontend developers | Node.js, pnpm, VS Code/Zed, browser tooling, frontend extensions |
| `platform-engineer` | Platform teams | Docker, Kubernetes tools, cloud CLIs, GitHub CLI |
| `python-ai` | Python AI/ML | uv, Python, ruff, AI tools, Jupyter (optional) |

```bash
dk init                           # Default preset (source: default)
dk init --preset ai-native        # AI preset (source: custom, you own it)
dk init --preset dotnet-fullstack # .NET preset (source: custom)
```

Each preset is defined in code as a `ToolDef`-like structure — not as YAML files that drift from the implementation.

---

## 8. Command Model 2.0

### `dk init`

Create a configuration file.

```bash
dk init                           # Default preset → source: default
dk init --preset ai-native        # Generate custom config from a preset
dk init --team <url>              # Team config (existing behavior)
dk init --custom                  # Interactive custom config creation
```

### `dk plan`

Show what Datakraften will do before changing the machine. Builds on the existing `--dry-run` infrastructure but produces a standalone, readable diff.

```bash
dk plan
> Will install:
>   apt: build-essential, curl, git
>   brew: fish, starship, gh
>   runtimes: node lts, python latest
> 
> Will configure:
>   shell: fish (default)
>   prompt: starship
```

### `dk apply`

Apply the configuration. Same semantics as 1.0 but with broader coverage.

```bash
dk apply                          # Apply everything
dk apply --dry-run                # Preview only
dk apply --preset ai-native       # Apply without init — use a preset directly
```

### `dk doctor`

Diagnose the workstation. The `--fix` flag finally works.

```bash
dk doctor                         # Full diagnostic
dk doctor --json                  # Machine-readable output
dk doctor --fix                   # Auto-fix detected issues
dk doctor --fix --dry-run         # Preview fixes
dk doctor --category shell        # Check only shell category
```

### `dk status`

Show current state compared to config.

```bash
dk status
> System packages:
>   ✓ git
>   ✓ curl
>   ✗ redis-tools
> 
> Runtimes:
>   ✓ node 22.11.0
>   ✗ dotnet 9.0 — not installed
```

### `dk upgrade`

Upgrade Datakraften itself. Existing behavior from 1.0, kept stable.

```bash
dk upgrade
```

### `dk update`

Update managed tools. Existing behavior from 1.0, kept stable.

```bash
dk update                         # Update all tools
dk update brew                    # Update specific tool category
dk update --list                  # List available updates
dk update --dry-run               # Preview updates
```

### `dk remove`

Remove Datakraften-managed components where safe. **New in 2.0.**

```bash
dk remove dotnet                  # Remove a managed runtime
dk remove codex                   # Remove an AI tool
dk remove --dry-run               # Preview removals
```

---

## 9. Feature Areas

Each area describes expected capabilities for 2.0. Sections marked **new** did not exist in 1.0.

### 9.1 System Packages (existing, extend)

- Install missing packages via native PM (apt, dnf, yum, pacman)
- Detect already-installed packages (skip)
- Multi-distro support: Debian/Ubuntu, Fedora, Arch
- Dry-run support
- `NativeUpdate()` called before `EnsurePackages()` on first install

### 9.2 Homebrew Packages (existing, extend)

- Install Homebrew if missing
- Install/uninstall packages
- Detect existing packages
- Upgrade packages via `dk update`

### 9.3 Language Runtimes (existing, extend)

- Node.js via fnm (`version: lts | 20 | 22`)
- Python via uv (`version: latest | 3.12 | 3.13`)
- Go via brew
- .NET SDK via brew or dotnet-install script
- Runtime version detection in `dk doctor`
- Version aliases: `lts`, `latest`, `stable`

### 9.4 Shell Experience (existing, extend)

- **Fish** — existing support, improve managed config blocks
- **Bash** — **new**: managed block support (`~/.bashrc`)
- **Zsh** — **new**: managed block support (`~/.zshrc`)
- Starship prompt integration for all shells
- Atuin shell history for all shells
- fzf fuzzy finder for all shells
- `EDITOR` variable set based on config

### 9.5 Editor Setup (existing, extend)

- VS Code detection (Windows-side under WSL, native Linux)
- Zed detection and auto-install
- Cursor detection (Windows-side under WSL)
- Editor extension installation **(new)**
- Editor settings configuration **(new)**

### 9.6 AI Coding Tools (existing, extend)

- Codex CLI — install via npm
- OpenCode — install via Homebrew
- GitHub Copilot CLI — install via Homebrew
- Claude Code — install via npm
- Gemini CLI — install via npm
- Authentication state detection
- Login guidance for unauthenticated tools

### 9.7 AI Applications (existing, extend)

- Codex desktop app
- GitHub Copilot VS Code extension
- Claude desktop app (macOS)
- Desktop app installation via brew cask or VS Code

### 9.8 Git and GitHub **(new section)**

- Install Git if missing
- Configure global Git defaults (user.name, user.email, default branch)
- Install GitHub CLI
- Check GitHub authentication
- Configure useful aliases
- Optional commit signing

### 9.9 Docker and Containers **(new section)**

- Install Docker-related CLI tooling
- Detect Docker availability (daemon, socket)
- Detect WSL/Docker Desktop integration
- Docker Compose support
- Devcontainer CLI (optional)

### 9.10 Databases and Local Services **(new section)**

- PostgreSQL client installation
- Redis client installation
- Server installation (optional, opt-in)
- SQLite support

### 9.11 Cloud Tooling **(new section)**

- Azure CLI installation and auth check
- AWS CLI installation and auth check (optional)
- Google Cloud CLI installation and auth check (optional)

---

## 10. Managed Config Strategy

Datakraften must clearly distinguish between:

```text
owned by Datakraften
owned by the user
generated by Datakraften
detected but unmanaged
```

### Directory layout

```
~/.config/datakraften/config.yaml       # User's config file
~/.config/datakraften/generated/        # Datakraften-generated config files
~/.config/datakraften/state/            # Runtime state
~/.config/datakraften/backups/          # Backup of files before modification
```

### Shell config pattern

Managed blocks with clear markers:

```fish
# >>> datakraften >>>
source ~/.config/datakraften/generated/fish/config.fish
# <<< datakraften <<<
```

This pattern is applied to Fish (`~/.config/fish/config.fish`), Bash (`~/.bashrc`), and Zsh (`~/.zshrc`). Datakraften only touches content between its markers — everything else is left untouched.

### Invariant

Datakraften never takes over the entire user config. It only manages clearly marked sections.

---

## 11. State & Locking

### State file

```text
~/.local/state/datakraften/state.json
```

Tracks:
- Applied config hash
- Installed components and versions
- Managed files
- Generated files
- Last run timestamp
- Datakraften version
- Warnings from last run

### Lock file (future)

```yaml
# dk.lock.yaml (generated after each apply)
lock:
  node:
    requested: lts
    resolved: 22.11.0
  dotnet:
    requested: "9.0"
    resolved: "9.0.100"
```

A lock file makes runs reproducible by recording exactly what was resolved. Useful for team environments where deterministic versions matter.

### Logs

```text
~/.local/state/datakraften/logs/
```

```bash
dk logs           # View recent logs
dk logs last      # View last run log
```

---

## 12. Team Usage 2.0

### Existing model (1.0, unchanged)

```bash
dk init --team https://example.com/team.yaml
```

A thin local config stores only `source: team` and `url:`. Every `dk apply` fetches the remote YAML fresh. The remote file is the single source of truth.

### Extensions for 2.0

- **Preset-based team config**: `dk init --team <url> --preset dotnet-fullstack` — validate remote YAML against a preset schema before applying
- **Team dashboard**: `dk team status` — compare local state against team config
- **Auth verification**: `dk doctor` checks team-required auth (GitHub, Azure, AI tools)
- **Version pinning**: Team configs can specify exact versions for reproducibility

### Team config best practices

- Host YAML in a versioned path (e.g., `https://raw.githubusercontent.com/org/team-config/v2/dk.yaml`)
- Keep the remote config minimal — install only what every developer needs
- Run `dk doctor` after `dk apply` to verify everything is correct
- Document exceptions for things that cannot be automated

---

## 13. Security Model

### Principles

- Never store secrets directly in profile files
- Warn before running remote scripts or fetching remote configs
- Show installation sources for every package
- Prefer official package sources (apt, brew, npm, uv)
- Support `--dry-run` for all mutating operations
- Keep authentication interactive and explicit
- Never automatically grant permissions or store credentials

### Install script transparency

```bash
curl -fsSL https://datakraften.no/install | bash
```

The install script must be:
- Small and readable (under 100 lines)
- Print what it is doing at each step
- Avoid hidden telemetry
- Avoid destructive operations
- Support `--dry-run`

### Trust signals

- Open source on GitHub
- Signed GitHub releases
- SHA256 checksums for every release binary
- Clear docs on what gets installed and why

---

## 14. Output & UX

### Output style

The CLI should feel modern, fast, and readable:

```
Datakraften
The missing bootstrap layer for AI-native developer workstations.

Profile: ai-native

✓ git already installed
✓ curl already installed
→ installing fish
→ installing starship
✓ node lts installed
! GitHub CLI installed but not authenticated

Next steps:
  gh auth login
  dk doctor
```

### Conventions

- Clear section headers
- Checkmarks for installed, arrows for in-progress, warnings for issues
- Next steps printed at the end
- Minimal noise by default, verbose mode available
- Consistent exit codes across all commands
- `--json` for machine-readable output everywhere

---

## 15. Error Handling

Errors should be actionable.

```text
Bad:
  Command failed.

Good:
  Docker is installed, but the daemon is not available.

  Possible causes:
    - Docker Desktop is not running
    - WSL integration is disabled
    - Your user does not have permission

  Try:
    1. Start Docker Desktop
    2. Enable WSL integration for this distribution
    3. Run: docker version
```

Every error message should help the developer move forward.

---

## 16. Extensibility

### Component model (future)

Each managed tool should implement a component interface:

```text
component:
  detect    → is this tool installed?
  plan      → what would change?
  apply     → install/configure
  verify    → is it working?
  remove    → uninstall safely
```

This makes it straightforward to add new tools over time without modifying the core logic.

### Plugin candidates for 2.0

- AWS CLI
- Google Cloud CLI
- kubectl + helm + k9s
- Rust via rustup
- JetBrains Toolbox
- Neovim
- Devbox integration
- Nix integration

These are deferred unless explicitly prioritized in the delivery plan.

---

## 17. What Datakraften Should Not Try To Be

- A full configuration management system (Ansible, Puppet)
- A replacement for Docker or dev containers
- A replacement for Nix
- A secrets manager
- An enterprise MDM or device manager
- A CI/CD system
- A project framework
- A package manager

The strongest position is **workstation bootstrap and developer experience** — not replacing the tools below that layer.

---

## 18. Leveranseplan

### M1 — Config 2.0 + Presets

**Goal**: Versioned schema, preset system, and `dk plan` command.

| Task | Description |
|------|-------------|
| Schema validation on config load | Reject unknown keys, validate types, report line-accurate errors |
| `version:` field in config | Wire `version: 2` schema; provide migration path from v1 (implicit) to v2 |
| Preset system | Define built-in presets in code; `dk init --preset <name>` generates `source: custom` config |
| `dk plan` command | Show diff between desired and current state. Reuses `--dry-run` infrastructure |
| New config sections | Parse and validated but not yet wired: `git`, `github`, `containers`, `databases`, `cloud` |

**Acceptance criteria**: `dk init --preset ai-native` generates a valid `source: custom` config with `version: 2`. `dk plan` shows structured output without modifying anything.

---

### M2 — Doctor ++

**Goal**: `--fix` actually works, shell support expands beyond Fish.

| Task | Description |
|------|-------------|
| Wire `--fix` flag | Populate `doctor.Check.Fix` fields; execute fixes for all categories |
| `--fix --dry-run` | Preview fixes without executing |
| `--fix --category <cat>` | Targeted fixing |
| Bash managed blocks | `internal/shell/bash.go` — same markers as Fish |
| Zsh managed blocks | `internal/shell/zsh.go` — same markers as Fish |
| Shell-agnostic helper | `WriteManagedBlock(path, content)` to avoid duplication |

**Acceptance criteria**: `dk doctor --fix` installs missing tools, repairs shell config, and reports what it did. Fish, Bash, and Zsh all support managed blocks.

---

### M3 — Wire New Config Sections

**Goal**: All `version: 2` config keys are backed by real logic.

| Task | Description |
|------|-------------|
| `git:` section | Configure user.name, user.email, default branch, signing |
| `github:` section | Install gh, check auth status |
| `containers.docker:` section | Docker CLI, compose, WSL integration check |
| `containers.devcontainers:` section | Install devcontainer CLI (optional) |
| `databases:` section | Install postgresql-client, redis-tools, sqlite3 |
| `cloud.azure:` section | Install az CLI, check auth |

**Acceptance criteria**: Adding any new section to config.yaml changes what `dk apply` installs. Disabled sections are safely skipped.

---

### M4 — Wire Editor Extensions + Settings

**Goal**: Datakraften can install VS Code extensions and apply editor settings.

| Task | Description |
|------|-------------|
| VS Code extension install | Install from list in `editors.vscode.extensions` |
| VS Code settings | Apply `editors.vscode.settings` to VS Code's settings.json |
| Editor config wiring | Wire existing `Config.Editors` map to actual logic |

**Acceptance criteria**: Adding `github.copilot` to `editors.vscode.extensions` installs it on next apply. Editor settings are applied without overwriting unrelated settings.

---

### M5 — Remove + State

**Goal**: `dk remove` for safe uninstall; improved state tracking.

| Task | Description |
|------|-------------|
| `dk remove <component>` | Uninstall a managed component where safe |
| `dk remove --dry-run` | Preview removals |
| State tracking | Populate `InstalledTools` in state.json |
| Lock file generation | Generate `dk.lock.yaml` on apply with resolved versions |
| `dk logs` command | View logs from `~/.local/state/datakraften/logs/` |

**Acceptance criteria**: `dk remove codex` uninstalls Codex CLI and reports success. State file accurately reflects installed tools and versions.

---

### M6 — Testing & Quality

**Goal**: Meaningful test coverage, CI reliability.

| Task | Description |
|------|-------------|
| Unit tests for `config` | Schema validation, preset loading, source resolution |
| Unit tests for `system` | Platform detection, distro detection |
| Unit tests for `doctor` | Check types, fix execution, category filtering |
| Unit tests for `shell` | Managed block writing, idempotency |
| Integration tests for `apply` | Dry-run mode, all presets, team config flow |
| Mock installers in CI | Test apply logic without sudo or real packages |
| Test coverage goal | > 60 % across all packages |

**Acceptance criteria**: `go test ./...` passes in CI with > 60 % coverage. Lint, build, and test all run in CI on every PR.

---

### M7 — Polish & Documentation

**Goal**: CLI feels polished. Website docs cover 2.0 features.

| Task | Description |
|------|-------------|
| `--yes` flag on all commands | Bypass confirmation prompts |
| Consistent exit codes | Documented exit code conventions |
| `datakraften` binary alias | Symlink or copy of `dk` |
| Website: Config 2.0 docs | Update `tools.ts` with new config sections |
| Website: Presets page | Document each preset, what it includes |
| Website: Changelog | Public changelog for 2.0 releases |
| Updated README | Match 2.0 command model and features |

**Acceptance criteria**: `bun run build` passes. Website covers all 2.0 features. README is up to date.

---

## 19. Suksesskriterier

Datakraften 2.0 is shipping when:

1. **Config 2.0** — versioned schema, preset system, validation, and all new config sections (git, github, containers, databases, cloud) are defined, parsed, and validated.

2. **Doctor ++** — `dk doctor --fix` repairs common issues without manual intervention.

3. **Shell breadth** — Fish, Bash, and Zsh all support managed config blocks.

4. **Editor setup** — VS Code extensions and settings can be declared in config and applied automatically.

5. **Safe removal** — `dk remove` provides clean uninstall for managed components.

6. **Testing** — the codebase has > 60 % test coverage and passes all checks in CI.

7. **Documentation** — website docs, README, and changelog cover all 2.0 features.

8. **Upgrade path** — existing 1.0 configs continue to work. `dk upgrade` migrates from 1.x to 2.x.

### Success in the words of users

> "I installed a fresh WSL environment, ran Datakraften, and had everything I needed to work within minutes."

> "We use Datakraften to give every developer the same high-quality baseline without maintaining fragile setup documentation."

> "My AI tools, shell, editor, runtimes, and local tooling are part of the same reproducible workstation profile."
