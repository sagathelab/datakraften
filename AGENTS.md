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

## Current Phase — Fase 0 (prosjektoppsett)
- Go CLI skeleton with Cobra
- System detection (WSL, distro, systemd, shell)
- YAML config loading with Viper
- Profile definitions (minimal, default, ai, dotnet, frontend, platform)
- install bootstrap script
- Makefile build system

## Next Phase — Fase 1 (Core MVP)
- `dk init` generates real config
- `dk apply` runs APT installer, Homebrew installer, runtime setup
- `dk doctor` with full category checks
- `dk status` with actual tool detection
- Fish/Starship/Atuin/fzf shell setup
- Editor detection (VS Code, Zed)
- Docker Desktop WSL integration check

## Repository Structure
```
cmd/dk/main.go              # Entry point
internal/
  app/                       # CLI commands (init, apply, doctor, status, profile)
  config/                    # YAML config loader
  system/                    # Platform/WSL/distro detection
  exec/                      # Command execution wrapper
  doctor/                    # Check/result types for diagnostics
  profiles/                  # Profile listing
  installers/                # APT, Homebrew, etc. (to be implemented)
  runtimes/                  # fnm, uv, .NET (to be implemented)
  shell/                     # Fish config (to be implemented)
  editors/                   # Editor detection (to be implemented)
  docker/                    # Docker checks (to be implemented)
profiles/                    # YAML profile definitions
install                      # Bootstrap installer
```

## Build & Test
```bash
make build          # Compile dk CLI
make install        # Install to ~/.local/bin/dk
go vet ./...        # Lint check
```
