# Datakraften 1.0 Roadmap

## Vision

Datakraften aims to become the preferred orchestration platform for developer environments — a YAML-driven bootstrapper delivering reproducible, AI-ready developer workstations across WSL, Linux, and macOS.

---

## Phase 1 — Core MVP (current state ✅)

- [x] 6 CLI commands: `init`, `apply`, `doctor`, `status`, `profile`, `update`
- [x] System detection (WSL, distro, systemd, shell)
- [x] Multi-PM (APT, DNF, YUM, PACMAN, BREW)
- [x] Runtimes: Node (fnm), Python (uv), .NET (brew)
- [x] Fish shell with managed blocks
- [x] Editor detection (VS Code, Zed, Cursor)
- [x] Docker WSL integration check
- [x] AI tool installation (5 CLI + 3 desktop)
- [x] State management (`~/.local/state/datakraften/state.json`)
- [x] Self-update with SHA256 verification
- [x] Bootstrap install script
- [x] Team profile (remote YAML fetch-on-apply)
- [x] Linting: golangci-lint, Prettier, stylelint, TypeScript strict

---

## Phase 2 — Diagnostics & Maintenance (near term)

### 2.1 `dk doctor --fix`

- [ ] Wire the existing `--fix` flag to actually fix issues
- [ ] Populate the `doctor.Check.Fix` field (struct field exists, never used)
- [ ] Iterate checks with non-empty `Fix` values and execute as shell commands
- [ ] Fix categories:
  - `system` — `apt-get update` if stale, install missing deps
  - `tools` — install missing tools via `dk apply` path
  - `shell` — run `FishEnsureSetup()` / `BashEnsureSetup()` / `ZshEnsureSetup()`
  - `docker` — print actionable guidance
  - `editors` — print install guidance
- [ ] `--fix --category <cat>` for targeted fixing
- [ ] `--fix --dry-run` to preview fixes without executing

### 2.2 `dk upgrade` + new `dk update`

- [ ] **Rename**: `cmd_update.go` → `cmd_upgrade.go`, `dk update` → `dk upgrade` (self-update)
- [ ] **New `dk update`**: update all managed tools
  - `brew update && brew upgrade`
  - `fnm install lts` (re-install LTS to get latest)
  - `uv python install` (re-install latest)
  - `npm update -g` for any npm-installed AI tools
- [ ] `dk update <tool>` — update a specific tool only (e.g. `dk update codex`)
- [ ] `dk update --list` — show which tools have available updates
- [ ] `dk update --dry-run` — preview updates without running
- [ ] State tracking: populate the `InstalledTools` field in `state.json`

### 2.3 `ai_tools` / `ai_apps` YAML consistency

- [ ] Fix `profiles/default.yaml` — replace `ai:` with `ai_tools:` / `ai_apps:`
- [ ] Fix `profiles/minimal.yaml` — same change
- [ ] Create `profiles/ai.yaml` with `ai_tools`/`ai_apps` structure
- [ ] Register `ai` profile in `internal/profiles/profiles.go`
- [ ] Verify `cmd_init.go` generates correct YAML with `ai_tools`/`ai_apps`
- [ ] Remove unused `AIToolPackages` map from `brew.go`
- [ ] Remove unused `CustomConfig` empty struct (or give it purpose)

### 2.4 Shell expansion

- [ ] **Bash**: `internal/shell/bash.go` — managed block pattern (same as Fish)
  - Markers: `# >>> datakraften >>>` / `# <<< datakraften <<<`
  - Content: brew shellenv, fnm, uv, starship, atuin, fzf, EDITOR
  - Config file: `~/.bashrc` (or `~/.bash_profile` on macOS)
- [ ] **Zsh**: `internal/shell/zsh.go` — same pattern
  - Config file: `~/.zshrc`
- [ ] Update `shell` config section to support multiple shells (bash/zsh/fish)
- [ ] `dk doctor shell` checks correct config file based on selected shell
- [ ] Shell-agnostic helper: `WriteManagedBlock(path, content)` to avoid duplication

---

## Phase 3 — Profiles & Config (medium term)

### 3.1 Profile library

- [ ] Create YAML files: `ai.yaml`, `dotnet.yaml`, `frontend.yaml`, `platform.yaml`, `python.yaml`
- [ ] Register all profiles in `internal/profiles/profiles.go`
- [ ] `dk init --profile` gets updated listing
- [ ] `profiles/minimal.yaml` updated to correct config structure

### 3.2 Config consolidation

- [ ] Wire `Config.Tools` map to actual installation logic (currently ignored)
- [ ] Wire `Config.Editors` map to actual installation logic
- [ ] Validate config on load (Viper post-unmarshal validation)
- [ ] Ensure profile `ai:` key migration doesn't break existing user configs (backward compat)

---

## Phase 4 — Testing & Quality (ongoing)

### 4.1 Test coverage

- [ ] Unit tests for `config`, `system`, `exec`, `doctor` packages
- [ ] Integration tests for `apply` (dry-run, all profiles, team flow)
- [ ] Mock installers for CI-friendly testing
- [ ] Test coverage goal: > 60% before Phase 5

### 4.2 CLI polish

- [ ] `--yes` flag on ALL commands (init, doctor, status, profile, upgrade, update)
- [ ] `--dry-run` on more commands (init, upgrade, update)
- [ ] `dk logs` — view logs from `~/.local/state/datakraften/logs/`
- [ ] `datakraften` binary alias (symlink to `dk`)
- [ ] Consistent exit codes across all commands

---

## Phase 5 — Cross-platform: WSL, Linux, macOS (long term)

### 5.1 macOS as a first-class citizen

- [ ] Platform abstraction layer (interface in `internal/system/`)
- [ ] macOS detection → use `brew` as native package manager
- [ ] macOS-specific paths (`~/Library/Application Support/`, `~/.bash_profile`)
- [ ] macOS editor detection (not just WSL detection)
- [ ] Docker Desktop for Mac check
- [ ] Darwin-specific base packages
- [ ] CI/CD: cross-compile for darwin/amd64 + darwin/arm64

### 5.2 Full Linux distro support

- [ ] Fedora: DNF-specific installer with distro-specific packages
- [ ] Arch: PACMAN-specific installer with distro-specific packages
- [ ] Platform-specific `BasePackage` lists for each distro
- [ ] `NativeUpdate()` called before `EnsurePackages()` on first install
- [ ] Test in CI (Docker-based matrix workflows)

### 5.3 WSL improvements

- [ ] WSL2-specific checks (kernel version, vmmem, .wslconfig)
- [ ] Auto-detect: Docker Desktop vs Rancher Desktop
- [ ] Windows-side editor CLI detection with proper PATH handling
- [ ] .wslconfig generator (`dk wsl configure`)

---

## Phase 6 — Reproducibility & DevEx (very long term)

### 6.1 Devbox integration

- [ ] `internal/devbox/` — Devbox installation module
- [ ] `dk devbox init` — generate `devbox.json` from Datakraften config
- [ ] `dk devbox doctor` — validate Devbox setup
- [ ] Templates for Node, Python, .NET, AI, Platform
- [ ] `templates/devbox/` directory with starter configs

### 6.2 Nix integration

- [ ] `internal/nix/` — Nix installation (single-user, multi-user, determinate systems)
- [ ] `dk nix init` — generate `flake.nix` from Datakraften config
- [ ] `dk nix doctor` — validate Nix setup (channels, flakes enabled)
- [ ] NixOS-WSL: `dk nixos-wsl init` — generate starter configuration
- [ ] `templates/nix/` directory with starter configs
- [ ] Keep Nix as optional, power-user mode only

### 6.3 Team onboarding

- [ ] `dk team join <name>` — join a team with a well-known URL or registry lookup
- [ ] Repository bootstrap: clone repos, set up pre-commit hooks
- [ ] Auth checks: GitHub CLI, Azure CLI, AI tools
- [ ] Team dashboard: `dk team report` — produces onboarding summary
- [ ] `.ai/` project files generated for team conventions

---

## Milestones

| Milestone | Deliverables | Horizon |
|-----------|--------------|---------|
| **M1** Doctor Fix + Upgrade/Update | `--fix` implemented, `dk upgrade` + `dk update`, YAML consistency | Near |
| **M2** Shell + Profiles | Bash/Zsh managed blocks, new profile YAMLs | Near–Medium |
| **M3** Testing | Unit/integration tests, CLI polish, `--yes` everywhere | Medium |
| **M4** macOS Support | Platform abstraction, darwin builds, brew-native on macOS | Medium–Long |
| **M5** Full Linux Support | Distro-specific installers, CI test matrix | Medium–Long |
| **M6** Devbox/Nix | Reproducible environments, templates, project-level configs | Long |
| **M7** Team Onboarding | Team join flow, repo bootstrap, auth checks, reports | Long |

---

## Success Criteria

Datakraften 1.0 is shipping when:

1. A developer can run `curl -fsSL https://datakraften.no/install | bash` on a fresh machine (WSL, Linux, or macOS) and have a productive, AI-ready environment within minutes.
2. `dk doctor --fix` repairs common issues without manual intervention.
3. `dk upgrade` keeps Datakraften itself up to date; `dk update` keeps tools up to date.
4. Profiles (`minimal`, `default`, `ai`, `dotnet`, `frontend`, `platform`, `custom`, `team`) are fully functional and YAML-driven.
5. The codebase has meaningful test coverage and passes all lint/format/type checks in CI.
6. Team onboarding is a documented, reliable flow — not an afterthought.
7. The platform abstraction is proven: adding a new OS target requires implementing an interface, not rewriting modules.
