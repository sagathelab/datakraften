# Development Context for AI Agents

## Core Concept

Datakraften (`dk`) is a CLI for bootstrapping developer workstations. It orchestrates existing tools (apt, brew, fnm, uv) — it does NOT replace them.

The vision is to become the preferred orchestration platform for developer environments — a YAML-driven bootstrapper delivering reproducible, AI-ready developer workstations across WSL, Linux, and macOS.

## Platform Strategy

- **WSL-first** — primary target, widest gap in tooling
- **Not WSL-only** — architecture abstracts platform detection so native Linux/macOS can be added later
- Platform-specific logic is isolated behind interfaces, not scattered through modules
- **Full cross-platform** (WSL, Linux, macOS) is the long-term goal

## Key Directives

- Idempotent by design — `dk apply` is safe to run repeatedly
- Opinionated defaults with escape hatches — config sources for different setups (default, custom, team)
- AI-native — install and configure AI dev tools
- `dk doctor` should be exceptional — diagnostics are core to the product, not an afterthought
- Orchestrate, do not replace — delegate to apt, brew, fnm, uv
- YAML-driven configuration — all tooling and runtime choices are expressed in YAML

## Current Phase — Phase 1 (Core MVP ✅)

Achieved:
- Go CLI skeleton with Cobra (6 commands: init, apply, doctor, status, upgrade, update)
- System detection (WSL, distro, systemd, shell)
- YAML config loading with Viper
- Config sources: `source: default | custom | team` (embedded default, local custom file, remote team URL)
- Bootstrap install script with GitHub Releases download + SHA256 + source fallback
- Makefile build system
- CI/CD: Go CI, cross-compile release workflow, site deploy workflow
- Linting: golangci-lint, Prettier, stylelint, TypeScript strict
- State management (`~/.local/state/datakraften/state.json`)
- `dk init` — single entry point: no args = default, `--custom [file]` = custom, `--team [url]` = team
- `dk apply` — system packages (APT/DNF/YUM/PACMAN/BREW), Homebrew, runtimes (Node/fnm, Python/uv, .NET/brew), shell, editors, AI tools
- `dk doctor` with category checks (system, tools, runtimes, editors, docker, shell) and `--json`/`--category`/`--fix`
- `dk status` with actual tool detection and version output
- `dk upgrade` with self-update from GitHub releases + SHA256 verification
- Fish/Starship/Atuin/fzf shell setup with managed config blocks
- Editor detection (VS Code, Zed, Cursor) including Windows-side detection
- Docker Desktop WSL integration check
- Multi-package-manager support (APT, DNF, YUM, PACMAN, BREW) with method-based dispatch
- AI tool installation (5 CLI tools via brew/npm, 3 desktop apps via brew cask/VS Code extensions)
- Team source: thin local config + remote YAML fetch-on-apply
- `--dry-run` and `--json` global flags
- Website with docs, landing page, install page, config guide, teams guide (in `site/`)
- opencode project config with AGENTS.md + SITE.md instructions

## Roadmap — Phase 2 and Beyond

See `plans/datakraften_1.0_roadmap.md` for full details. Near-term priorities:

1. **Testing** — add unit and integration tests
2. **Config consolidation** — wire `Config.Tools` and `Config.Editors` maps to actual logic, add validation
3. **Cross-platform** — macOS as first-class citizen, full Linux distro support, WSL improvements
4. **Reproducibility** — Devbox and Nix integration
5. **Team onboarding** — team join flow, repo bootstrap, auth checks

## Repository Structure

```
cmd/dk/main.go                 # Entry point
internal/
  app/                         # CLI commands + apply/state logic
  config/                      # YAML config loader (Viper)
  system/                      # Platform/WSL/distro detection
  exec/                        # Command execution wrapper
  doctor/                      # Check/result types for diagnostics
  installers/                  # APT, Homebrew, multi-PM support (apt.go, brew.go, packages.go)
  runtimes/                    # fnm (Node), uv (Python), .NET
  shell/                       # Fish/Bash/Zsh config with managed blocks
  ai/                          # AI tool installation (CLI + desktop apps)
  editors/                     # Editor detection (VS Code, Zed, Cursor)
  docker/                      # Docker Desktop WSL integration check
install                        # Bootstrap installer script
plans/                         # Roadmap and planning documents
  datakraften_1.0_roadmap.md
site/                          # Website and documentation (see SITE.md)
opencode.json                  # opencode project configuration
AGENTS.md                      # This file — agent development context
SITE.md                        # Website maintenance instructions
```

## Build & Test

```bash
make build          # Compile dk CLI
make install        # Install to ~/.local/bin/dk
make test           # Run Go tests
make lint           # golangci-lint run ./...
make clean          # Remove bin/
```

## CLI Commands Overview

```bash
dk init             # Generate config: no args = default, --custom [file], --team [url]
dk apply            # Install everything (supports --dry-run, --yes)
dk doctor           # Diagnostics (supports --json, --category, --fix)
dk status           # Tool overview (supports --json)
dk upgrade          # Self-update CLI
dk update           # Update managed tools (supports --list, --dry-run)
```

## Code Conventions

- **Go**: Standard library + Cobra + Viper + survey. Follow existing patterns in `internal/`.
- **No comments** in code unless the logic is genuinely non-obvious.
- **Idempotency first** — every install/configure function must be safe to re-run.
- **Config** uses `source: default | custom | team` (not `profile:`). Old `profile:` key is transparently mapped.
- **Config keys** use `ai_tools` and `ai_apps` (not the old `ai:` key).
- **Managed shell blocks** use `# >>> datakraften >>>` / `# <<< datakraften <<<` markers.
- **Website**: Bun + Vite + React + TypeScript + Tailwind v4. All doc content in `site/src/data/tools.ts`. See `SITE.md` for full details.
