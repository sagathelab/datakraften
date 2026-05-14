# Datakraften

**The AI-native developer workstation platform.**

[![CI](https://github.com/sagathelab/datakraften/actions/workflows/ci.yml/badge.svg)](https://github.com/sagathelab/datakraften/actions/workflows/ci.yml)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/github/go-mod/go-version/sagathelab/datakraften)](https://go.dev/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)
[![Docs](https://img.shields.io/badge/docs-datakraften.no-blue)](https://datakraften.no/docs/)

Bootstrap, configure, validate, and reproduce development environments across machines — WSL-first, multi-platform ambition.

```bash
curl -fsSL https://datakraften.no/install | bash
dk init
dk apply
dk doctor
dk update          # Periodically update dk itself
```

## What is Datakraften?

Datakraften (`dk`) is a CLI that orchestrates existing tools — `apt`, `brew`, `fnm`, `uv`, `Devbox`, `Nix` and more — to turn a fresh or unstructured machine into a consistent, productive, and AI-ready development environment.

It is **not** a package manager. It is an orchestration layer with opinionated profiles and excellent diagnostics.

## Profiles

| Profile | Description |
|---------|-------------|
| `minimal` | Core system tools only |
| `default` | General developer setup |
| `ai` | AI-native development environment |
| `dotnet` | .NET developer setup |
| `frontend` | Frontend developer setup |
| `platform` | Platform engineer setup |

## Commands

| Command | Description |
|---------|-------------|
| `dk init` | Generate configuration file |
| `dk apply` | Install everything in your profile (idempotent) |
| `dk doctor` | Run diagnostics |
| `dk status` | Show installed tools overview |
| `dk profile list` | List available profiles |
| `dk profile use <name>` | Switch active profile |
| `dk update` | Self-update dk to latest release |

All commands support `--dry-run` (preview changes) and `--json` (structured output).

## Quick start

```bash
# Install dk
curl -fsSL https://datakraften.no/install | bash

# Initialize with a profile
dk init --profile default

# Apply the configuration
dk apply

# Verify everything
dk doctor
```

## AI tooling

The `ai` profile installs and configures:

- [OpenAI Codex CLI](https://github.com/openai/codex)
- [OpenCode](https://opencode.ai)
- [GitHub Copilot CLI](https://github.com/github/gh-copilot)
- [Claude Code](https://docs.anthropic.com/en/docs/claude-code)
- [Gemini CLI](https://github.com/google-gemini/gemini-cli)

Run `dk profile use ai && dk apply` to set up an AI-ready environment.

## Development

```bash
make build      # Build dk CLI
make test       # Run tests
make lint       # go vet ./...
make install    # Install to ~/.local/bin/dk
```

## Project plan

See [plans/datakraften-platform-plan.md](plans/datakraften-platform-plan.md) for the full product vision, architecture, and roadmap.

## Platform targets

1. **WSL (Ubuntu/Debian)** — primary target
2. **Native Linux** — future (Fedora, Arch)
3. **macOS** — future

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

[Apache 2.0](LICENSE) © 2026 Datakraften
