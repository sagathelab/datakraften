# Datakraften

**The AI-native developer workstation platform.**

[![CI](https://github.com/sagathelab/datakraften/actions/workflows/ci.yml/badge.svg)](https://github.com/sagathelab/datakraften/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/sagathelab/datakraften)](https://go.dev/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

Bootstrap, configure, validate, and reproduce development environments across machines — WSL-first, multi-platform ambition.

```bash
curl -fsSL https://datakraften.no/install.sh | bash
dk init
dk apply
dk doctor
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

## Quick start

```bash
# Install dk
curl -fsSL https://datakraften.no/install.sh | bash

# Initialize with a profile
dk init --profile default

# Apply the configuration
dk apply

# Verify everything
dk doctor
```

## Development

```bash
make build      # Build dk CLI
make test       # Run tests
make lint       # go vet ./...
make install    # Install to ~/.local/bin/dk
```

## Platform targets

1. **WSL (Ubuntu/Debian)** — primary target
2. **Native Linux** — future (Fedora, Arch)
3. **macOS** — future

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

[MIT](LICENSE) © 2026 Datakraften
