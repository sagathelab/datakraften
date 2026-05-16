# Datakraften

> **The missing bootstrap layer for AI-native developer workstations.**

Datakraften is an opinionated, declarative workstation bootstrapper for developers who want to turn a fresh WSL/Linux environment into a complete, productive, AI-ready developer workstation in minutes.

It is not just a package installer.  
It is not just a dotfiles manager.  
It is not just a runtime version manager.  
It is not just a project-specific dev container.

Datakraften sits above those layers and connects them into one coherent developer experience.

---

## 1. Purpose

Modern developers do not only need a compiler, a runtime, and an editor.

They need a full local environment that includes:

- system packages
- language runtimes
- shell configuration
- editor setup
- Git and GitHub tooling
- Docker and container support
- database clients
- cloud tooling
- AI coding assistants
- consistent conventions
- repeatable setup across machines

Today, developers often solve this with a mixture of shell scripts, README steps, dotfiles, package managers, copied commands, tribal knowledge, and manual setup.

Datakraften aims to replace that fragmented setup with a single declarative profile and a reliable bootstrap workflow.

---

## 2. Positioning

**Datakraften is the missing bootstrap layer for AI-native developer workstations.**

It focuses on the complete local developer workstation rather than only one isolated part of it.

Where other tools solve specific layers, Datakraften connects the layers.

| Category | Examples | What they solve | Datakraften's role |
|---|---|---|---|
| Dotfiles | chezmoi, yadm, GNU Stow | Personal config files | Can manage or integrate workstation-level config |
| Runtime managers | mise, asdf, nvm, pyenv | Tool and language versions | Can configure and orchestrate runtime installation |
| Dev environments | Devbox, Flox, devenv | Project or shell environments | Can complement them at workstation level |
| Dev containers | Dev Containers, Codespaces | Containerized project environments | Can install tooling and prepare the host |
| Package managers | apt, Homebrew, npm, pip | Package installation | Can declaratively drive them |
| Provisioning | Ansible, shell scripts | Automation and machine setup | Provides a developer-first profile model |

Datakraften should be opinionated enough to be useful immediately, but flexible enough to support different developer profiles.

---

## 3. Core Promise

A developer should be able to run something like:

```bash
dk init
dk apply
dk doctor
```

and end up with a ready-to-use workstation.

The goal is:

```text
Fresh WSL/Linux environment
        ↓
Datakraften profile
        ↓
Complete AI-native developer workstation
```

The workstation should feel consistent, modern, fast, and ready for real work.

---

## 4. Target Users

### Individual developers

Developers who frequently set up new machines, WSL environments, cloud workstations, or test environments.

### Consultants

Consultants who move between customers, projects, and environments and need to become productive quickly.

### Platform teams

Teams that want to standardize the developer workstation experience across an organization.

### Fullstack developers

Developers working across frontend, backend, APIs, databases, Docker, cloud, and CI/CD.

### AI-native developers

Developers who use AI coding tools as a normal part of their workflow and want those tools configured as first-class citizens.

---

## 5. First-Class Target Environment

The first major target should be:

```text
WSL on Windows
```

with strong support for:

```text
Ubuntu on WSL
Linux distributions
macOS later if relevant
```

WSL is a strong starting point because many developers work on Windows machines but want a Linux-native development experience.

Datakraften should make WSL feel less like an empty Linux box and more like a ready-made developer workstation.

---

## 6. What Datakraften Should Deliver

Datakraften should deliver a full workstation baseline.

### 6.1 System Packages

Install required system-level packages.

```yaml
system_packages:
  apt:
    - build-essential
    - curl
    - git
    - unzip
    - postgresql-client
    - redis-tools
```

Expected capabilities:

- install missing packages
- detect already installed packages
- support dry runs
- support package manager abstraction
- support different Linux distributions over time
- avoid reinstalling unchanged packages
- report clear errors when packages fail

---

### 6.2 Homebrew Packages

Support Homebrew on Linux as a developer-focused package layer.

```yaml
packages:
  brew:
    - fish
    - starship
    - atuin
    - fzf
    - gh
    - docker
```

Expected capabilities:

- install Homebrew if missing
- install packages
- upgrade packages optionally
- detect existing packages
- separate system packages from developer packages
- support package groups later

---

### 6.3 Language Runtimes

Install and configure common development runtimes.

```yaml
runtimes:
  node:
    enabled: true
    version: lts

  python:
    enabled: true
    version: "3.12"

  dotnet:
    enabled: true
    version: "9.0"
```

Expected capabilities:

- install Node.js
- install Python
- install .NET SDK
- support version selection
- support latest/lts/stable aliases
- expose runtime versions in `dk doctor`
- integrate with version managers where useful
- avoid conflicts with existing user-managed installations

Potential integrations:

- mise
- asdf
- nvm
- pyenv
- official Microsoft .NET packages
- Homebrew packages

---

### 6.4 Shell Experience

Configure a modern shell experience.

```yaml
shells:
  fish:
    enabled: true
    default: true
    managed_config: true
```

Expected capabilities:

- install shell
- optionally set default shell
- configure prompt
- configure shell history
- configure completions
- configure aliases/functions
- configure environment variables
- support managed and unmanaged modes
- allow users to inspect generated config

Recommended tools:

- fish
- starship
- atuin
- fzf

Important principle:

Datakraften should not silently destroy personal shell configuration. It should either manage clearly marked sections or generate separate managed files that are sourced by the user shell.

---

### 6.5 Editor Setup

Install and configure editors.

```yaml
editors:
  vscode:
    enabled: true
    extensions:
      - ms-dotnettools.csharp
      - esbenp.prettier-vscode
      - github.copilot
    settings:
      editor.formatOnSave: true
      terminal.integrated.defaultProfile.linux: fish

  zed:
    enabled: true
```

Expected capabilities:

- install editors
- configure extensions
- configure editor settings
- configure terminal profile
- configure AI extensions
- support project-independent defaults
- avoid overwriting user settings without clear ownership

Supported editors should begin with:

- Visual Studio Code
- Zed

Later possible support:

- JetBrains Toolbox
- Neovim
- Cursor
- Windsurf

---

### 6.6 AI Coding Tools

AI tools should be first-class citizens in Datakraften.

```yaml
ai_tools:
  codex:
    enabled: true
    install_cli: true

  opencode:
    enabled: true

  copilot:
    enabled: true
    install_cli: true
    vscode_extension: true

  claude:
    enabled: false

  gemini:
    enabled: false
```

Expected capabilities:

- install AI CLI tools
- install editor extensions
- configure shell integration where useful
- detect authentication state where possible
- report missing authentication
- avoid storing secrets directly in the profile
- provide clear next steps for login

Datakraften should treat AI tooling as part of the normal developer environment, not as an afterthought.

---

### 6.7 AI Applications

Some AI tools are not only CLI tools. They may also have desktop apps, editor plugins, background agents, or service integrations.

```yaml
ai_apps:
  codex:
    enabled: true

  copilot:
    enabled: true

  claude:
    enabled: false
```

Expected capabilities:

- distinguish between CLI tools and apps
- install supported apps where possible
- configure editor/app integration
- document manual steps where installation cannot be automated
- keep authentication outside the declarative config

---

### 6.8 Git and GitHub

Datakraften should make Git and GitHub ready for development.

```yaml
git:
  user_name: "Jan Helge Knutsen"
  user_email: "janhelgeknutsen@gmail.com"
  default_branch: main
  signing:
    enabled: false

github:
  cli:
    enabled: true
  auth:
    check: true
```

Expected capabilities:

- install Git
- configure global Git defaults
- install GitHub CLI
- check GitHub authentication
- configure useful aliases
- configure default branch
- optionally configure signing
- avoid forcing identity if not specified

---

### 6.9 Docker and Containers

Datakraften should prepare the machine for container-based development.

```yaml
containers:
  docker:
    enabled: true
    compose: true

  devcontainers:
    enabled: true
```

Expected capabilities:

- install Docker-related tooling
- detect Docker availability
- detect WSL/Docker Desktop integration
- install Docker Compose where needed
- install devcontainer CLI if enabled
- validate that containers can run
- report actionable diagnostics

---

### 6.10 Databases and Local Services

Datakraften should support clients and optionally local services.

```yaml
databases:
  postgresql:
    client: true
    server: false

  redis:
    client: true
    server: false
```

Expected capabilities:

- install database clients
- optionally install local services
- avoid starting heavyweight services by default
- provide clear distinction between client and server
- support Docker-based local services later

---

### 6.11 Cloud Tooling

Cloud tools should be optional but supported.

```yaml
cloud:
  azure:
    cli: true
    devops: true

  aws:
    cli: false

  gcloud:
    cli: false
```

Expected capabilities:

- install cloud CLIs
- check authentication status
- support Azure CLI
- support Azure DevOps extension
- support AWS CLI and Google Cloud CLI later
- avoid embedding credentials

---

### 6.12 Project Templates

Datakraften can later support project and workspace templates.

```yaml
templates:
  nextjs:
    enabled: true

  dotnet_api:
    enabled: true

  fullstack:
    enabled: true
```

Expected capabilities:

- scaffold new projects
- apply conventions
- generate recommended folder structures
- generate editor settings
- generate CI/CD starter files
- generate local development scripts

This should come after the workstation bootstrap layer is solid.

---

## 7. Profile Structure

Datakraften should use a declarative profile file.

Recommended naming options:

```text
datakraften.yaml
dk.yaml
.dk/profile.yaml
```

A profile should be human-readable, versionable, and easy to share.

Recommended style:

Use maps/objects for configurable features.

Good:

```yaml
editors:
  vscode:
    enabled: true
```

Avoid:

```yaml
editors:
  - vscode
```

The object format allows future properties without breaking the schema.

---

## 8. Example Full Profile

```yaml
version: 1

metadata:
  name: ai-native-fullstack
  description: AI-ready fullstack workstation for WSL/Linux developers

system_packages:
  apt:
    - build-essential
    - curl
    - git
    - unzip
    - postgresql-client
    - redis-tools

packages:
  brew:
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
    version: "3.12"

  dotnet:
    enabled: true
    version: "9.0"

shells:
  fish:
    enabled: true
    default: true
    managed_config: true

editors:
  vscode:
    enabled: true
    extensions:
      - ms-dotnettools.csharp
      - esbenp.prettier-vscode
      - github.copilot
    settings:
      editor.formatOnSave: true
      terminal.integrated.defaultProfile.linux: fish

  zed:
    enabled: true

ai_tools:
  codex:
    enabled: true
    install_cli: true

  opencode:
    enabled: true

  copilot:
    enabled: true
    install_cli: true
    vscode_extension: true

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

containers:
  docker:
    enabled: true
    compose: true

  devcontainers:
    enabled: true

git:
  default_branch: main
  pull_rebase: false

github:
  cli:
    enabled: true
  auth:
    check: true

databases:
  postgresql:
    client: true
    server: false

  redis:
    client: true
    server: false
```

---

## 9. Command Model

Datakraften should expose a small and predictable CLI.

### `dk init`

Create a starter profile.

```bash
dk init
```

Possible options:

```bash
dk init --profile ai-native
dk init --profile dotnet-fullstack
dk init --profile frontend
```

### `dk plan`

Show what Datakraften will do before changing the machine.

```bash
dk plan
```

Expected output:

```text
Will install:
  apt: build-essential, curl, git, unzip
  brew: fish, starship, atuin, fzf, gh
  runtimes: node lts, python 3.12, dotnet 9.0
  editors: vscode, zed

Will configure:
  shell: fish
  prompt: starship
  history: atuin
  github auth check
```

### `dk apply`

Apply the profile.

```bash
dk apply
```

Expected behavior:

- safe
- idempotent
- clear output
- fails with actionable messages
- does not overwrite unmanaged files silently

### `dk doctor`

Validate the workstation.

```bash
dk doctor
```

Expected checks:

- operating system
- WSL status
- package managers
- installed runtimes
- shell configuration
- Git config
- GitHub CLI auth
- Docker availability
- editor availability
- AI tool availability
- common PATH issues

### `dk status`

Show current state compared to the profile.

```bash
dk status
```

Expected output:

```text
Profile: ai-native-fullstack

System packages:
  ✓ git
  ✓ curl
  ✗ redis-tools

Runtimes:
  ✓ node 22.11.0
  ✓ python 3.12.3
  ✗ dotnet 9.0

AI tools:
  ✓ codex
  ✓ opencode
  ! copilot installed but not authenticated
```

### `dk update`

Update Datakraften-managed tools.

```bash
dk update
```

This should be intentionally conservative. Updating everything automatically can break developer environments.

### `dk remove`

Remove Datakraften-managed components where safe.

```bash
dk remove <component>
```

This should be careful and explicit.

---

## 10. Design Principles

### 10.1 Declarative

The profile describes the desired workstation state.

The user should not have to remember installation steps.

### 10.2 Idempotent

Running `dk apply` multiple times should be safe.

The second run should mostly say:

```text
Already installed
Already configured
No changes needed
```

### 10.3 Transparent

Datakraften should show what it is doing.

No mysterious hidden changes.

### 10.4 Safe by Default

Datakraften should avoid destructive actions.

It should not overwrite personal config without a clear strategy.

### 10.5 Opinionated, Not Rigid

Datakraften should provide useful defaults, but allow users and teams to customize.

### 10.6 AI-Native

AI tooling is not a plugin afterthought.

It belongs in the same profile as runtimes, editors, shell, and Git.

### 10.7 Workstation-Level

Datakraften focuses on the whole workstation.

Project-specific tools can still be handled by dev containers, Devbox, Flox, devenv, npm scripts, or other tools.

---

## 11. Managed Configuration Strategy

Datakraften should clearly distinguish between:

```text
owned by Datakraften
owned by the user
generated by Datakraften
detected but unmanaged
```

Recommended pattern:

```text
~/.config/datakraften/
~/.config/datakraften/generated/
~/.config/datakraften/state/
~/.config/datakraften/backups/
```

For shell config:

```fish
# Datakraften managed block - start
source ~/.config/datakraften/generated/fish/config.fish
# Datakraften managed block - end
```

This avoids taking over the entire user config.

---

## 12. State and Locking

Datakraften should maintain local state.

```text
~/.local/state/datakraften/state.json
~/.local/state/datakraften/lock.json
```

State can include:

- applied profile hash
- installed components
- managed files
- generated files
- last run timestamp
- Datakraften version
- warnings from last run

A lock file can make runs more reproducible.

Example:

```yaml
lock:
  node:
    requested: lts
    resolved: 22.11.0
  dotnet:
    requested: "9.0"
    resolved: "9.0.100"
```

---

## 13. Profiles and Presets

Datakraften should support built-in presets.

Examples:

```text
ai-native
dotnet-fullstack
frontend-nextjs
python-ai
platform-engineer
minimal
```

A user should be able to start with a preset and customize it.

```bash
dk init --preset dotnet-fullstack
```

Possible preset intent:

### `minimal`

Basic shell, Git, and package manager setup.

### `frontend-nextjs`

Node.js, pnpm, VS Code/Zed, browser tooling, frontend extensions.

### `dotnet-fullstack`

.NET SDK, Node.js, PostgreSQL client, Docker, VS Code C# tools.

### `platform-engineer`

Docker, Kubernetes tools, cloud CLI tools, GitHub CLI.

### `ai-native`

AI coding assistants, shell tooling, editors, runtimes, and developer productivity defaults.

---

## 14. Team Usage

Datakraften should support both personal and team-level profiles.

Example:

```text
company-dk.yaml
team-dk.yaml
personal-dk.yaml
```

A team may define:

- baseline packages
- required runtimes
- editor extensions
- Git conventions
- cloud tools
- AI tool preferences
- security requirements

A developer may override:

- editor preferences
- shell preferences
- optional tools
- local paths
- personal Git identity

Potential layering model:

```text
base profile
  + team profile
  + personal profile
  + machine-specific overrides
```

---

## 15. Security Model

Datakraften should be careful with trust.

Profiles can install software and change configuration, so they must be treated as executable intent.

Security principles:

- never store secrets directly in profile files
- warn before running remote scripts
- show installation sources
- prefer official package sources
- support dry run
- support signed or trusted profiles later
- clearly separate authentication from installation
- avoid automatically granting permissions

Authentication should be interactive and explicit.

Example:

```bash
gh auth login
codex login
```

Datakraften can detect and report auth state, but should not own credentials.

---

## 16. Error Handling

Errors should be actionable.

Bad:

```text
Command failed.
```

Good:

```text
Docker is installed, but the daemon is not available.

Possible causes:
  - Docker Desktop is not running
  - WSL integration is disabled
  - Your user does not have permission to access Docker

Try:
  1. Start Docker Desktop
  2. Enable WSL integration for this distribution
  3. Run: docker version
```

Datakraften should always optimize for helping the developer move forward.

---

## 17. Output Style

The CLI should feel modern and readable.

Example:

```text
Datakraften
The missing bootstrap layer for AI-native developer workstations.

Profile: ai-native-fullstack

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

Recommended output concepts:

- clear sections
- checkmarks
- warnings
- next steps
- minimal noise by default
- verbose mode available

---

## 18. Configuration Schema

Datakraften should have a versioned schema.

```yaml
version: 1
```

Schema principles:

- maps for configurable features
- arrays for simple lists
- explicit `enabled`
- avoid magical implicit behavior
- support comments in YAML examples
- validate before applying

Bad:

```yaml
ai_tools:
  codex: true
```

Better:

```yaml
ai_tools:
  codex:
    enabled: true
```

Best when more config is needed:

```yaml
ai_tools:
  codex:
    enabled: true
    install_cli: true
    auth_check: true
```

---

## 19. Extensibility

Datakraften should eventually support components/providers.

Conceptual model:

```text
component:
  detect
  plan
  apply
  verify
  remove
```

Each component should be able to say:

- current state
- desired state
- required changes
- apply changes
- verify result

This makes it easier to support new tools over time.

---

## 20. What Datakraften Should Not Try To Be

Datakraften should avoid becoming too broad too early.

It should not try to be:

- a full configuration management system like Ansible
- a replacement for Docker
- a replacement for dev containers
- a replacement for Nix
- a secrets manager
- a full MDM or enterprise device manager
- a CI/CD system
- a project framework

The strongest position is workstation bootstrap and developer experience.

---

## 21. MVP Scope

A strong MVP should focus on WSL/Linux and the most valuable workstation features.

Recommended MVP:

```text
dk init
dk plan
dk apply
dk doctor
```

Supported in MVP:

- apt packages
- Homebrew packages
- Git
- GitHub CLI
- Node.js LTS
- Python
- .NET SDK
- fish shell
- starship
- atuin
- fzf
- VS Code
- Zed
- Docker detection
- Codex CLI
- OpenCode
- GitHub Copilot CLI/extension where practical

Avoid in MVP:

- complex profile layering
- remote profile registry
- secrets management
- full GUI
- too many package managers
- too many operating systems
- advanced uninstall
- enterprise policy management

---

## 22. Future Scope

Possible future features:

### Profile registry

```bash
dk profile search dotnet
dk profile use datakraften/dotnet-fullstack
```

### Team profiles

```bash
dk apply --profile https://example.com/team-dk.yaml
```

### Lock files

```text
dk.lock.yaml
```

### Graphical dashboard

A local UI showing workstation health and missing tools.

### Dev container awareness

Detect `.devcontainer/devcontainer.json` and recommend host tooling.

### Project bootstrap

Generate new projects from Datakraften templates.

### Policy mode

For teams that want controlled workstation baselines.

### Remote environments

Support cloud workstations or remote Linux boxes.

---

## 23. Website Messaging

Suggested hero copy:

```text
The missing bootstrap layer for AI-native developer workstations.

Datakraften turns a fresh WSL/Linux environment into a complete, reproducible, and opinionated developer workstation — with runtimes, shell, editors, Docker, GitHub tooling, and AI coding assistants configured from one declarative profile.
```

Alternative shorter hero:

```text
From clean WSL to AI-ready developer workstation in minutes.
```

Supporting bullets:

```text
One declarative profile
Full workstation bootstrap
Built for WSL/Linux
AI tools as first-class citizens
Safe, transparent, and repeatable
```

---

## 24. README Intro

```markdown
# Datakraften

> The missing bootstrap layer for AI-native developer workstations.

Datakraften is a declarative workstation bootstrapper for developers who want a complete, reproducible, and opinionated local setup.

It helps you turn a fresh WSL/Linux environment into a modern developer workstation with system packages, runtimes, shell configuration, editors, Docker, GitHub tooling, and AI coding assistants.

```bash
dk init
dk plan
dk apply
dk doctor
```

Datakraften focuses on the whole workstation experience, not only dotfiles, not only package installation, and not only project-specific environments.
```

---

## 25. Naming Concepts

Useful product terms:

```text
Profile
Component
Preset
Provider
Plan
Apply
Doctor
Managed config
Workstation state
AI-native workstation
Bootstrap layer
```

Avoid unclear terms like:

```text
magic setup
machine brain
super config
```

The product should sound reliable, technical, and developer-focused.

---

## 26. Success Criteria

Datakraften is successful if a developer can say:

```text
I installed a fresh WSL environment, ran Datakraften, and had everything I needed to work within minutes.
```

A team lead should be able to say:

```text
We use Datakraften to give every developer the same high-quality baseline without maintaining fragile setup documentation.
```

An AI-native developer should be able to say:

```text
My AI tools, shell, editor, runtimes, and local tooling are part of the same reproducible workstation profile.
```

---

## 27. Core Differentiator

The unique focus is not package installation alone.

The unique focus is:

```text
A complete, opinionated, AI-native developer workstation baseline.
```

Datakraften's strength is combining:

```text
workstation bootstrap
developer experience
AI tooling
WSL/Linux focus
declarative configuration
safe idempotent execution
```

into one product.

---

## 28. Final Product Definition

Datakraften is a developer-first bootstrap system for modern local workstations.

It provides a declarative way to define, install, configure, validate, and maintain the tools developers need every day.

It starts with WSL/Linux and focuses on the real needs of fullstack, platform, and AI-native developers.

Its purpose is simple:

```text
Make a clean machine productive.
Make that setup repeatable.
Make AI-native development the default.
```

**Datakraften is the missing bootstrap layer for AI-native developer workstations.**
