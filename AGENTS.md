# Development Context for AI Agents

## Core Concept
Datakraften (`dk`) is a CLI for bootstrapping developer workstations. It orchestrates existing tools (apt, brew, fnm, uv, Devbox, Nix) — it does NOT replace them.

## Platform Strategy
- **WSL-first** — primary target, widest gap in tooling
- **Not WSL-only** — architecture abstracts platform detection so native Linux/macOS can be added later
- Platform-specific logic is isolated behind interfaces, not scattered through modules

## Key Directives
- Idempotent by design — `dk apply` is safe to run repeatedly
- Opinionated defaults with escape hatches — profiles for different developer types
- AI-native — install and configure AI dev tools, generate agent-friendly project files
- `dk doctor` should be exceptional — diagnostics are core to the product, not an afterthought
- Orchestrate, do not replace — delegate to apt, brew, fnm, uv, Devbox, Nix

## Current Phase — Fase 1 (Core MVP)
Achieved:
- Go CLI skeleton with Cobra (6 commands: init, apply, doctor, status, profile, update)
- System detection (WSL, distro, systemd, shell)
- YAML config loading with Viper
- Profile definitions (minimal, default, ai, dotnet, frontend, platform)
- install bootstrap script
- Makefile build system
- CI/CD with GitHub Actions (cross-compile release workflow)
- State management (`~/.local/state/datakraften/state.json`)
- `dk init` generates real config
- `dk apply` runs APT installer, Homebrew installer, runtime setup (Node via fnm, Python via uv, .NET via brew)
- `dk doctor` with category checks (system, tools, runtimes, editors, docker, shell)
- `dk status` with actual tool detection
- `dk update` with self-update from GitHub releases + SHA256 verification
- `dk profile list/use` — list and switch profiles
- Fish/Starship/Atuin/fzf shell setup with managed config blocks
- Editor detection (VS Code, Zed, Cursor) including Windows-side detection
- Docker Desktop WSL integration check
- Multi-package-manager support (APT, DNF, YUM, PACMAN, BREW)
- `--dry-run` and `--json` global flags

In progress / known gaps:
- AI tool installation not yet wired into `dk apply` (Codex, OpenCode, Claude Code, Gemini CLI)
- `doctor.Check`/`Report` framework exists but is not yet integrated — doctor uses ad-hoc print functions
- Profile YAML files missing for `dotnet`, `frontend`, `platform` (3 of 6 exist)
- `dk apply` does not install .NET SDK (EnsureDotnet exists but is not called)
- No test files yet
- `--json` flag is registered but no command outputs JSON yet
- `--yes` flag is registered but ignored

## Next Phase — Fase 2 (Profiles & AI)
- Complete all 6 profile YAML files
- Wire AI tool installation into `dk apply`
- Wire .NET SDK installation into `dk apply`
- Integrate `doctor.Check`/`Report` framework
- Implement `dk doctor --fix`
- Implement `dk ai doctor` and `dk ai init-project`
- Add JSON output for `dk status` and `dk doctor`
- Add test coverage
- Implement `dk upgrade --tools` (planned, separate from `dk update --self`)

## Repository Structure
```
cmd/dk/main.go              # Entry point
internal/
  app/                       # CLI commands + apply/state logic
  config/                    # YAML config loader (Viper)
  system/                    # Platform/WSL/distro detection
  exec/                      # Command execution wrapper
  doctor/                    # Check/result types for diagnostics (framework exists, not yet integrated)
  profiles/                  # Profile listing
  installers/                # APT, Homebrew, multi-PM support
  runtimes/                  # fnm (Node), uv (Python), .NET
  shell/                     # Fish config with managed blocks
  editors/                   # Editor detection (VS Code, Zed, Cursor)
  docker/                    # Docker Desktop WSL integration check
profiles/                    # YAML profile definitions (minimal, default, ai; dotnet/frontend/platform missing)
install                      # Bootstrap installer
```

## Build & Test
```bash
make build          # Compile dk CLI
make install        # Install to ~/.local/bin/dk
make test           # Run tests (none yet — TODO)
make lint           # go vet ./...
go vet ./...        # Lint check
```

## CLI Commands Overview
```bash
dk init             # Generate config
dk apply            # Install everything
dk doctor           # Diagnostics
dk status           # Tool overview
dk profile list     # List profiles
dk profile use      # Switch profile
dk update           # Self-update CLI
```
