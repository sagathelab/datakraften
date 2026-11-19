# Security

Datakraften is designed to be transparent and auditable.

The bootstrap script is intentionally small and only installs the `dk` CLI. The main logic lives in the open source repository at [github.com/sagathelab/datakraften](https://github.com/sagathelab/datakraften).

## Security Principles

- **Minimal use of sudo** — only used for system package installation (`apt`, `dnf`, etc.)
- **No hidden telemetry** — the CLI does not collect or send data by default
- **No secrets are collected** — we never ask for or store API keys, tokens, or credentials
- **No destructive operations without clear user intent** — `dk apply` never runs without confirmation; `--dry-run` shows everything beforehand
- **Third-party tools are installed from documented sources** — always from official channels (Homebrew, GitHub Releases, etc.)
- **Shared workstation templates are imported once** — `dk init --from` validates and previews an HTTPS YAML template before saving local configuration. `dk apply` and `dk update` do not fetch it again.
- **Project configuration is not trusted by default** — inspection does not run tasks or install tools. Approve a project's `datakraften.yaml` before using it to execute tasks. Approval is not sandboxing: tasks run with your user's permissions and can access inherited environment variables.
- **The install script should be readable before execution** — review it at [datakraften.no/install](https://datakraften.no/install) before running it

## Reporting a Vulnerability

If you discover a security issue, please do **NOT** open a public issue.

Instead, report it privately by opening a GitHub Security Advisory:

[https://github.com/sagathelab/datakraften/security/advisories/new](https://github.com/sagathelab/datakraften/security/advisories/new)

### What to include

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

### Response

We will acknowledge receipt within 48 hours and work on a fix. We will keep you informed throughout the process.

## Scope

**In scope:**
- The `dk` CLI binary and source code
- The `install` bootstrap script
- The website (datakraften.no)

**Out of scope:**
- Third-party tools installed or configured by Datakraften (Homebrew, fnm, uv, Docker, etc.)
- Browser extensions or unrelated tools
