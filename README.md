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
dk update          # Update managed tools
```

## What is Datakraften?

Datakraften (`dk`) is a CLI that orchestrates existing tools — including `apt`, `brew`, `fnm` and `uv` — to set up a development workstation and run project tasks with a project-specific Node.js version.

It uses existing package and version managers, with workstation configuration and diagnostics.

## Configuration

`dk init` creates a local workstation configuration. `dk init --from <https-url>` imports a partial YAML template once, merges it with the built-in defaults and saves the result locally; later edits are yours and the URL is not fetched again. `dk init --migrate-team` explicitly converts an existing legacy team configuration and preserves a backup. Review the import warnings: some existing workstation settings are not yet used by `dk apply`.

## Commands

| Command | Description |
|---------|-------------|
| `dk init` | Generate configuration file |
| `dk apply` | Apply the workstation configuration |
| `dk doctor` | Run diagnostics |
| `dk status` | Show installed tools overview |
| `dk project status` | Inspect the current project's requirements without installing or running anything |
| `dk project trust` | Review and approve the current project configuration |
| `dk project setup` | Install the missing project Node version after confirmation |
| `dk run <task>` | Run a trusted project task in its project environment |
| `dk update` | Update managed tools; npm updates only enabled, installed AI tools |
| `dk upgrade` | Upgrade dk to the latest release |
| `dk npm setup` | Configure a user-owned npm prefix and shell PATH |

`dk npm setup --dry-run` previews the prefix and shell changes. Running the setup changes
your npm user prefix; existing global packages are not moved and may need reinstalling.
A custom prefix may conflict with Node version managers such as fnm or nvm. Use a
version manager's own writable prefix instead when appropriate.

Use `--help` on a command to see the flags it supports. `dk project status --json` returns a machine-readable project status.

## Quick start

```bash
# Install dk
curl -fsSL https://datakraften.no/install | bash

# Initialize a workstation configuration
dk init

# Apply the configuration
dk apply

# Verify everything
dk doctor
```

## Project environments

Put a `datakraften.yaml` in the project directory:

```yaml
version: 1
project:
  tools:
    node: "22"
  env:
    NODE_ENV: development
  tasks:
    dev: npm run dev
    test: npm test
```

Run `dk project status`, review and trust the configuration with `dk project trust`, then run `dk project setup` and `dk run dev`. Project tasks run through `/bin/sh` from the directory containing the configuration. `dk project trust --yes` explicitly approves the current file without an interactive prompt; any change requires renewed trust. Project variables are passed to the task; do not put secrets in a shared YAML file or print them in tasks.

The first project backend supports Node.js through `fnm` in `PATH`, with numeric major, major.minor or exact versions. `dk project setup` installs a missing version into a separate Datakraften-managed fnm store without changing your fnm default. Only Node is supported as a project tool in this release; other tools are rejected. Existing `.nvmrc` and `mise.toml` files are not used for version selection yet. Project tasks inherit the process environment; this is not a security sandbox.

## AI tooling

The configuration supports tools including:

- [OpenAI Codex CLI](https://github.com/openai/codex)
- [OpenCode](https://opencode.ai)
- [GitHub Copilot CLI](https://github.com/github/gh-copilot)
- [Claude Code](https://docs.anthropic.com/en/docs/claude-code)
- [Gemini CLI](https://github.com/google-gemini/gemini-cli)

Enable the desired tools in the local configuration and run `dk apply`.

## Development

```bash
make build      # Build dk CLI
make test       # Run tests
make lint       # golangci-lint run ./...
make install    # Install to ~/.local/bin/dk
```

## Project plan

See the [Datakraften 3.0 roadmap](plans/datakraften_3.0_roadmap.md) for the product direction and planned milestones. The earlier [platform plan](plans/datakraften-platform-plan.md) remains available for historical context.

## Platform targets

1. **WSL (Ubuntu/Debian)** — primary target
2. **Native Linux** — future (Fedora, Arch)
3. **macOS** — future

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

[Apache 2.0](LICENSE) © 2026 Datakraften
