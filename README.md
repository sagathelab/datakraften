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

Datakraften (`dk`) uses existing tools like `apt`, `brew`, `fnm`, `nvm`, and `uv` to set up a development workstation and run project tasks with a project-specific Node.js version.

It uses existing package and version managers, with workstation configuration and diagnostics.

## Configuration

`dk init` creates a local workstation configuration in `~/.config/datakraften/config.yaml`. To share a starting point, run `dk init --from <https-url>` with a link to a raw YAML template. The template's settings override the built-in defaults where specified. The CLI shows which settings change and asks before replacing an existing configuration unless you pass `--yes`. Review and edit the local result before running `dk apply`. Datakraften does not save the URL or fetch the template again during `dk apply` or `dk update`. There is no live team configuration or migration command.

Workstation configuration and the repository's `datakraften.yaml` are separate. Do not put credentials in either file.

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

Run `dk project status`, review and trust the configuration with `dk project trust`, then run `dk project setup` and `dk run dev`. Project tasks run through `/bin/sh` from the directory containing the configuration. For non-interactive use, `dk project trust --yes` explicitly approves the reviewed file; `dk project setup --yes` approves only the installation. Editing or moving the file requires renewed trust. Project variables are passed to the task; do not put secrets in shared YAML.

For Node.js, choose `fnm` or `nvm` in the project YAML:

```yaml
version: 1
project:
  tools:
    node:
      version: "22"
      manager: nvm
```

The short `node: "22"` form uses `fnm`. Use a numeric major, major.minor or exact version. `fnm` must be available in `PATH`, or `nvm.sh` must exist in `NVM_DIR` or a supported location. `dk project setup` reuses installed versions or installs a missing one in a separate Datakraften-managed store without changing the user's default. Only Node is supported as a project tool. `dk` does not use `.nvmrc`, `.node-version`, `package.json` engines, or `mise.toml` for project version selection. Project tasks inherit the process environment; this is not a security sandbox.

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

Ubuntu/Debian on WSL is the primary target. CI builds and runs tests on Linux and macOS, but that does not verify every tool or installation workflow on those platforms. See the [3.0 platform matrix](plans/datakraften_3.0_roadmap.md#10-plattform-og-kompatibilitet) for targets and how support will be verified.

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

[Apache 2.0](LICENSE) © 2026 Datakraften
