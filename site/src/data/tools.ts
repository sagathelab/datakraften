export interface ToolSection {
  title?: string
  body: string
  id?: string
}

export interface ToolDef {
  id: string
  title: string
  subtitle: string
  sections: ToolSection[]
  website: string
  websiteLabel?: string
}

export const tools: Record<string, ToolDef> = {
  dk: {
    id: 'dk',
    title: 'dk CLI',
    subtitle: 'Bootstrap, configure, and diagnose your developer workstation',
    sections: [
      {
        title: 'dk init',
        id: 'init',
        body: "Generate your Datakraften configuration file. Detects your operating system, package manager, and writes ~/.config/datakraften/config.yaml.\n\n$ dk init\n\nNo arguments creates a default config (source: default) with full tooling defaults.\n\n$ dk init --custom\n\nCreate a custom config (source: custom) — you own this file. You'll be asked whether to start from an empty skeleton or a pre-filled template. You must edit the file to customize.\n\n$ dk init --custom ./my-config.yaml\n\nCreate from an existing YAML file. Validates the file, then uses it as your configuration (with source: custom).\n\n$ dk init --team https://example.com/team.yaml\n\nCreate a thin team config (source: team) pointing to a remote YAML. The remote URL is required.\n\nWhat it does:\n- Detects platform (WSL, Linux, macOS)\n- Identifies native package manager (apt, dnf, yum, pacman, brew)\n- Generates a YAML config based on the chosen source\n- source: default can overwrite existing; source: custom never overwrites without confirmation",
      },
      {
        title: 'dk apply',
        id: 'apply',
        body: 'Install everything defined in your configuration. Idempotent -- safe to run repeatedly.\n\n$ dk apply\n$ dk apply --dry-run\n\nNOTE: Preview without installing\n\nWhat it installs:\n- System packages -- via apt, dnf, yum, pacman, or brew\n- Homebrew -- installs brew if missing, then brew packages\n- Runtimes -- Node.js via fnm, Python via uv, Go, .NET SDK\n- Shell -- Fish shell config with managed blocks\n- AI tools -- CLI tools + desktop apps\n- Skips already-installed tools. Uses sudo only for system packages.\n\nTeam source:\nIf your config has source: team, `dk apply` fetches the remote YAML fresh before installing. The remote file is the single source of truth — your local config only stores ~source: team~ and ~url~. If the remote YAML is invalid or unreachable, `dk apply` aborts.\n\nDry-run mode:\nUse --dry-run to see what would be installed without making changes:\n\n$ dk apply --dry-run\n> [dry-run] would install system packages: git curl build-essential ...\n> [dry-run] would install brew packages: fish starship atuin fzf ...\n> [dry-run] would install Node.js via fnm (LTS)',
      },
      {
        title: 'dk doctor',
        id: 'doctor',
        body: 'Run comprehensive diagnostics on your system. Checks every category that dk apply configures.\n\n$ dk doctor\n$ dk doctor --json\n\nNOTE: Machine-readable output\n\nCheck categories:\n- System -- distribution, kernel, WSL version, systemd\n- Tools -- git, curl, build tools, Homebrew\n- Runtimes -- Node.js, Python, Go, .NET SDK\n- Editors -- VS Code, Zed, Cursor detection\n- Docker -- daemon status, socket access, WSL integration\n- Shell -- Fish config, Starship prompt, Atuin, FZF',
      },
      {
        title: 'dk status',
        body: 'Quick overview of installed tools and their versions.\n\n$ dk status\n> System:    Fedora 40 (WSL2)\n> Shell:     fish 3.7.1\n> Node:      22.x\n> Python:    3.12.x\n> Go:        1.24.x\n> .NET:      8.0.x\n> Git:       2.45.x\n> Brew:      4.x\n> Docker:    running',
      },
      {
        title: 'Config sources',
        id: 'sources',
        body: 'Datakraften supports three config sources. The source determines where your configuration comes from.\n\n**default** — The built-in config embedded in the dk binary. Full tooling defaults for a developer workstation. Run `dk init` (no args) to recreate it.\n\n**custom** — A local config file you own. Run `dk init --custom` to create one (empty skeleton or pre-filled template). Edit it freely — `dk init --custom` never overwrites without asking.\n\n**team** — A thin config pointing to a remote YAML URL. The remote file defines everything. Run `dk init --team <url>` to set it up. Every `dk apply` fetches the remote YAML fresh.\n\n$ dk init\n> source: default — full default config\n\n$ dk init --custom\n> ? Create config as: Empty skeleton / Pre-filled with defaults\n> source: custom — you must edit the file to customize\n\n$ dk init --team https://example.com/team.yaml\n> Fetching remote config from https://example.com/team.yaml...\n> ✓ Remote config validated\n> source: team — thin config, remote is single source of truth\n\nNOTE: The remote URL is required for team config — there is no fallback to a local config. If the remote YAML is invalid or unreachable, `dk apply` aborts with an error.',
      },
      {
        title: 'dk upgrade',
        id: 'upgrade',
        body: 'Upgrade the Datakraften CLI to the latest release. Downloads the correct binary for your platform from GitHub, verifies the SHA256 checksum, and performs an atomic replacement.\n\n$ dk upgrade\n\n> Current version: v0.1.0\n> Latest version:  v0.2.0\n> Downloading dk-linux-amd64...\n> Checksum verified ✓\n> Updated to v0.2.0\n\nHow it works:\n- Fetches the latest release from api.github.com\n- Detects your OS and architecture to download the matching binary\n- Verifies the binary against its SHA256 checksum\n- Replaces the current binary atomically\n\nTIP: Run `dk upgrade` periodically to get the latest version of Datakraften.',
      },
      {
        title: 'dk update',
        id: 'update',
        body: 'Update managed developer tools to their latest versions. Supports updating all tools at once or targeting a specific tool.\n\n$ dk update\n\nUpdate all tools — Homebrew packages, Node.js LTS via fnm, Python via uv, and global npm packages.\n\nUpdate a specific tool:\n\n$ dk update brew\n$ dk update fnm\n$ dk update uv\n$ dk update npm\n\nList available updatable tools:\n\n$ dk update --list\n\nDry-run to preview updates:\n\n$ dk update --dry-run',
      },
    ],
    website: 'https://github.com/sagathelab/datakraften',
    websiteLabel: 'GitHub repository',
  },

  node: {
    id: 'node',
    title: 'Node.js + fnm',
    subtitle: 'Fast Node Manager -- seamlessly switch between Node.js versions',
    sections: [
      {
        title: 'What is it?',
        body: 'fnm (Fast Node Manager) installs and manages multiple Node.js versions. It is written in Rust and is significantly faster than nvm. The bootstrapper installs the latest LTS version of Node.js and sets it as the default.',
      },
      {
        title: 'Basic Usage',
        body: 'Check current version:\n$ node --version\n$ npm --version\n\nInstall a specific Node version:\n$ fnm install 22\n$ fnm install 20\n\nSwitch between versions:\n$ fnm use 22\n$ fnm use 20\n\nSet a default version:\n$ fnm default 22\n\nList installed versions:\n$ fnm list\n\nInstall npm packages globally:\n$ npm install -g typescript\n$ npm install -g yarn',
      },
      {
        title: 'How it works',
        body: 'fnm adds a shim directory to your PATH and switches Node.js versions by updating symlinks. The shell integration (fnm env) ensures the right version is active in every new terminal. When you cd into a project with a .nvmrc file, fnm automatically reads the version and switches to it.',
      },
      {
        title: 'Tips',
        body: 'TIP: Node.js is managed by fnm -- never install it via apt or brew directly. Only fnm should manage Node versions.\n\nTIP: Use fnm install --lts to always get the latest LTS release.',
      },
    ],
    website: 'https://nodejs.org/',
    websiteLabel: 'Node.js',
  },

  python: {
    id: 'python',
    title: 'Python + uv',
    subtitle: 'Fast Python package manager in Rust -- replaces pip, venv, and more',
    sections: [
      {
        title: 'What is it?',
        body: 'uv is an extremely fast Python package and project manager written in Rust. It can manage Python versions, virtual environments, dependencies, and run scripts -- all in one tool. The bootstrapper installs uv and uses it to set up the default Python runtime.',
      },
      {
        title: 'Basic Usage',
        body: 'Manage Python versions:\n$ uv python install 3.12\n$ uv python list\n\nCreate a virtual environment:\n$ uv venv\n$ source .venv/bin/activate\n\nInstall packages:\n$ uv pip install requests\n$ uv pip install -r requirements.txt\n\nRun a script with its dependencies:\n$ uv run script.py\n$ uv run --with requests python script.py\n\nRun tools without installing (uvx):\n$ uvx ruff check .\n$ uvx black file.py\n$ uvx pyright\n\nManage project dependencies:\n$ uv add requests\n$ uv add --dev pytest\n$ uv sync',
      },
      {
        title: 'Tips',
        body: "TIP: uv replaces pip, pip-tools, pipx, and venv. Use uv pip install instead of pip install -- it's 10-100x faster.\n\nTIP: uvx runs Python tools in isolated environments -- no need to install them first. Great for linters, formatters, and type checkers.\n\nTIP: Use uv add to add dependencies to pyproject.toml and uv sync to install them.\n\nTIP: Run uv help to see all available commands and options.",
      },
    ],
    website: 'https://docs.astral.sh/uv/',
    websiteLabel: 'uv',
  },

  go: {
    id: 'go',
    title: 'Go',
    subtitle: 'Compiled, batteries-included language for fast CLIs, services, and tooling',
    sections: [
      {
        title: 'What is it?',
        body: 'Go is a statically typed, compiled language designed for simplicity, fast builds, and great tooling. Datakraften installs Go as a first-class runtime so you can build CLIs, backend services, and developer tools out of the box.',
      },
      {
        title: 'Basic Usage',
        body: 'Check the installed version:\n$ go version\n\nCreate a new module:\n$ mkdir myapp && cd myapp\n$ go mod init example.com/myapp\n\nRun a program:\n$ go run .\n\nBuild a binary:\n$ go build\n\nRun tests:\n$ go test ./...\n\nAdd a dependency:\n$ go get github.com/spf13/cobra\n\nFormat code:\n$ go fmt ./...',
      },
      {
        title: 'Tips',
        body: 'TIP: Go modules are the default dependency system. Run go mod init once per project.\n\nTIP: Use go test ./... to run all tests in the current module.\n\nTIP: Use gofmt or go fmt consistently -- formatting is part of the standard Go workflow.',
      },
    ],
    website: 'https://go.dev/',
    websiteLabel: 'Go',
  },

  dotnet: {
    id: 'dotnet',
    title: '.NET SDK',
    subtitle: 'Cross-platform SDK for building C#, F#, and VB applications',
    sections: [
      {
        title: 'What is it?',
        body: 'The .NET SDK is a cross-platform development framework for building a wide range of applications -- web, mobile, desktop, cloud, and IoT -- using C#, F#, or Visual Basic. The bootstrapper installs the latest .NET SDK via Homebrew.',
      },
      {
        title: 'Basic Usage',
        body: 'Create a new project:\n$ dotnet new console -n MyApp\n$ dotnet new webapi -n MyApi\n$ dotnet new mvc -n MyWebApp\n\nList available templates:\n$ dotnet new list\n\nBuild and run:\n$ dotnet build\n$ dotnet run\n\nRun tests:\n$ dotnet test\n\nPublish for deployment:\n$ dotnet publish -c Release -o ./publish\n\nCheck installed SDKs:\n$ dotnet --list-sdks\n$ dotnet --list-runtimes',
      },
      {
        title: 'Tips',
        body: 'TIP: DOTNET_ROOT is set in your fish config. The SDK is installed to /usr/share/dotnet.\n\nTIP: Use dotnet new list to browse all available project templates.\n\nTIP: Add --help to any dotnet command for detailed usage info.',
      },
    ],
    website: 'https://dotnet.microsoft.com/',
  },

  fish: {
    id: 'fish',
    title: 'Fish Shell',
    subtitle: 'Friendly interactive shell -- your default shell after bootstrapping',
    sections: [
      {
        title: 'What is it?',
        body: 'Fish (Friendly Interactive SHell) is a smart, user-friendly command-line shell with features like syntax highlighting, autosuggestions, and tab completions out of the box -- no configuration required. After bootstrapping, Fish is set as your default shell (effective after re-login).',
      },
      {
        title: 'Basic Usage',
        body: 'Start Fish:\n$ fish\n\nMake Fish default:\n$ chsh -s /usr/bin/fish\n\nAutosuggestions:\nFish shows dimmed suggestions based on your history as you type. Press → or Ctrl+F to accept the suggestion.\n\nTab Completions:\nPress Tab to complete commands, options, paths, and even git branches.',
      },
      {
        title: 'Configuration',
        body: 'The bootstrapper creates ~/.config/fish/config.fish with:\n- Homebrew PATH setup\n- fnm (Node.js version manager) integration\n- uv (Python package manager) shell completions\n- Atuin shell history\n- FZF fuzzy finder key bindings\n- Starship prompt\n- Useful aliases (g, ga, gc, gp, gl, gs, gd)\n- Editor set to zed --wait\n\nEdit config manually:\n$ nvim ~/.config/fish/config.fish\n\nReload config:\n$ source ~/.config/fish/config.fish',
      },
      {
        title: 'Tips',
        body: 'TIP: Fish syntax is different from Bash. Use if ... end instead of if ... fi, and set -x VAR value instead of export VAR=value.\n\nTIP: Run help to open the Fish web-based documentation.\n\nTIP: Use funced to edit a function interactively, and funcsave to persist it.',
      },
    ],
    website: 'https://fishshell.com/',
  },

  starship: {
    id: 'starship',
    title: 'Starship',
    subtitle: 'Minimal, blazing-fast prompt for any shell',
    sections: [
      {
        title: 'What is it?',
        body: 'Starship is a cross-shell prompt that shows you relevant context: current directory, git branch and status, Node.js/Python/.NET versions, command duration, and more -- all with zero configuration to get started. Starship works in Fish, Bash, Zsh, and PowerShell. The bootstrapper enables it automatically in your Fish config.',
      },
      {
        title: 'Usage',
        body: "Starship works automatically -- you don't need to run any commands. The prompt will show:\n- Current directory\n- Git branch and status (dirty/staged/ahead/behind)\n- Runtime versions when in a project directory\n- Command execution time if > 2s\n- Exit code of the last command if non-zero",
      },
      {
        title: 'Customization',
        body: 'Create ~/.config/starship.toml to customize every part of the prompt:\n\n# Disable the package module\n[package]\ndisabled = true\n\n# Change the Node.js symbol\n[nodejs]\nsymbol = "⬢ "\nformat = "via [$symbol($version )]($style)"\n\nChanges take effect immediately -- no reload needed.',
      },
      {
        title: 'Tips',
        body: 'TIP: Run starship preset nerd-font-symbols -o ~/.config/starship.toml to use Nerd Font icons.\n\nTIP: See all configuration options at starship.rs/config.',
      },
    ],
    website: 'https://starship.rs/',
  },

  atuin: {
    id: 'atuin',
    title: 'Atuin',
    subtitle: 'Magical shell history with search, sync, and encryption',
    sections: [
      {
        title: 'What is it?',
        body: "Atuin replaces your shell's built-in history with a powerful, encrypted database. It provides blazing-fast search across all your commands, syncs history between machines, and supports regex filtering. The bootstrapper enables Atuin for Fish shell.",
      },
      {
        title: 'Basic Usage',
        body: "Interactive search:\nNOTE: Press Ctrl+R and start typing\n\nSearch from the command line:\n$ atuin search docker\n$ atuin search --regex 'git.*push'\n$ atuin search --cwd /projects/myapp\n\nSync history across machines:\n$ atuin login\n$ atuin sync\n\nBrowse all history:\n$ atuin history list\n$ atuin history list --session-only",
      },
      {
        title: 'Tips',
        body: 'TIP: Atuin works automatically once installed -- it hooks into your shell and replaces the default history.\n\nTIP: atuin search supports regex patterns, filtering by host, directory, and session.\n\nTIP: Run atuin login and atuin sync to enable encrypted cloud sync.',
      },
    ],
    website: 'https://atuin.sh/',
  },

  fzf: {
    id: 'fzf',
    title: 'fzf',
    subtitle: 'General-purpose command-line fuzzy finder',
    sections: [
      {
        title: 'What is it?',
        body: 'fzf is a general-purpose command-line fuzzy finder. It interactively filters lines from stdin or files. Integrated into Fish shell via key bindings (Ctrl+T for files, Ctrl+R for history, Alt+C for cd).',
      },
      {
        title: 'Basic Usage',
        body: "Search files (Ctrl+T):\nNOTE: Press Ctrl+T, type to filter, select with Tab/Enter\n\nReverse history search (Ctrl+R):\nNOTE: Press Ctrl+R, type any part of a past command to find it\n\ncd into subdirectory (Alt+C):\nNOTE: Press Alt+C, select a directory to cd into it\n\nPipe to fzf:\n$ find . -type f | fzf\n\nWith preview:\n$ fzf --preview 'cat {}'",
      },
      {
        title: 'Tips',
        body: "TIP: Use fzf --preview 'cat {}' to preview file contents as you filter.\n\nTIP: Set FZF_DEFAULT_COMMAND to use fd for faster search.\n\nTIP: fzf is integrated into Fish shell by default via the bootstrapper.",
      },
    ],
    website: 'https://github.com/junegunn/fzf',
  },

  fd: {
    id: 'fd',
    title: 'fd',
    subtitle: 'Fast and user-friendly alternative to find',
    sections: [
      {
        title: 'What is it?',
        body: 'fd is a fast and user-friendly alternative to find. It uses simple, intuitive syntax, respects .gitignore by default, produces colorized output, and is significantly faster than find.',
      },
      {
        title: 'Basic Usage',
        body: 'Search by pattern:\n$ fd pattern\n\nSearch by extension:\n$ fd -e md\n\nSearch in a specific path:\n$ fd pattern /path/to/search\n\nFilter by file type:\n$ fd --type f\n$ fd --type d\n\nInclude hidden files:\n$ fd --hidden\n\nExecute a command on results:\n$ fd --exec vim',
      },
      {
        title: 'Tips',
        body: "TIP: If installed, fzf automatically uses fd as its default file search backend.\n\nTIP: Run fd '' to list all files recursively in the current directory.\n\nTIP: Use --type d to search for directories only.",
      },
    ],
    website: 'https://github.com/sharkdp/fd',
  },

  broot: {
    id: 'broot',
    title: 'broot',
    subtitle: 'A new way to see and navigate directory trees',
    sections: [
      {
        title: 'What is it?',
        body: 'broot provides a new way to see and navigate directory trees. It features an interactive tree view with fuzzy filtering, file preview, and file management actions -- all from the terminal.',
      },
      {
        title: 'Basic Usage',
        body: 'Open interactive tree:\n$ broot\n\nFilter files:\nNOTE: Start typing any pattern to fuzzy-filter the tree\n\nSelect and quit:\nNOTE: Press Alt+Enter on a directory to output its path and quit\n\nFile operations:\n- :cp — copy\n- :mv — move/rename\n- :rm — delete\n\nShow help:\n$ broot -h',
      },
      {
        title: 'Tips',
        body: 'TIP: Navigate with arrow keys or vim keys (j/k) for fast movement.\n\nTIP: Press ? to see all available verbs and keyboard shortcuts.\n\nTIP: Use :open to open a file, or press Alt+Enter on a directory to cd into it.',
      },
    ],
    website: 'https://dystroy.org/broot/',
  },

  btm: {
    id: 'btm',
    title: 'btm',
    subtitle: 'Cross-platform graphical system monitor',
    sections: [
      {
        title: 'What is it?',
        body: 'bottom (btm) is a cross-platform graphical system monitor for the terminal. It displays CPU, memory, disk, network, processes, and temperatures in a rich TUI interface.',
      },
      {
        title: 'Basic Usage',
        body: 'Start bottom:\n$ btm\n\nNavigate:\nArrow keys  -- move between widgets and processes\n?           -- show help\n1-9         -- toggle individual widgets on/off\nCtrl+C      -- quit\n\nBasic mode:\n$ btm -c',
      },
      {
        title: 'Tips',
        body: 'TIP: Use btm --mem_as_value to display memory as numeric values instead of a percentage.\n\nTIP: Configuration lives in ~/.config/bottom/bottom.toml.\n\nTIP: Run btm -t for the default widget arrangement.',
      },
    ],
    website: 'https://github.com/ClementTsang/bottom',
  },

  brew: {
    id: 'brew',
    title: 'Homebrew',
    subtitle: 'The missing package manager for Linux -- most tools are installed via brew',
    sections: [
      {
        title: 'What is it?',
        body: 'Homebrew is a package manager that started on macOS and now fully supports Linux (Linuxbrew). It installs packages to /home/linuxbrew/.linuxbrew and keeps them isolated from system packages. The bootstrapper uses Homebrew as the primary package manager for all developer tools.',
      },
      {
        title: 'Basic Usage',
        body: 'Install a package:\n$ brew install gh\n\nSearch for packages:\n$ brew search gh\n\nUpdate all packages:\n$ brew update && brew upgrade\n\nList installed packages:\n$ brew list\n\nCheck for outdated packages:\n$ brew outdated\n\nGet info about a package:\n$ brew info gh',
      },
      {
        title: 'Tips',
        body: "TIP: brew formulas are updated frequently. Run brew update once a week to stay current.\n\nTIP: Use brew doctor if something isn't working.\n\nTIP: The bootstrapper adds brew to ~/.profile and ~/.bashrc.",
      },
    ],
    website: 'https://brew.sh/',
  },

  gh: {
    id: 'gh',
    title: 'GitHub CLI',
    subtitle: 'GitHub from the command line -- PRs, issues, repos, and more',
    sections: [
      {
        title: 'What is it?',
        body: 'The GitHub CLI (gh) brings GitHub to your terminal. It lets you manage repositories, pull requests, issues, actions, and more without leaving the command line. Installed via Homebrew.',
      },
      {
        title: 'Basic Usage',
        body: 'Authenticate with GitHub:\n$ gh auth login\n\nCreate a repository:\n$ gh repo create my-project --public --clone\n\nWork with pull requests:\n$ gh pr create --title "My PR" --body "Description"\n$ gh pr view\n$ gh pr checkout 123\n\nList and manage issues:\n$ gh issue list\n$ gh issue create --title "Bug" --body "Details"\n$ gh issue view 42\n\nView GitHub Actions runs:\n$ gh run list\n$ gh run view\n$ gh run watch\n\nOpen in browser:\n$ gh browse',
      },
      {
        title: 'Tips',
        body: 'TIP: Always run gh auth login first.\n\nTIP: gh status shows your latest notifications.\n\nTIP: If you have the GitHub Copilot CLI extension installed, try gh copilot.',
      },
    ],
    website: 'https://cli.github.com/',
  },

  'gh-copilot': {
    id: 'gh-copilot',
    title: 'Copilot',
    subtitle: 'GitHub Copilot CLI — AI assistance right in your terminal',
    sections: [
      {
        title: 'What is it?',
        body: 'GitHub Copilot CLI brings Copilot to your terminal -- get command suggestions, explanations, and translations from natural language without leaving the command line.',
      },
      {
        title: 'Basic Usage',
        body: 'Suggest a command:\n$ gh copilot suggest "list all modified files"\n\nExplain a command:\n$ gh copilot explain "git rebase -i HEAD~3"\n\nWhat the shell?:\n$ gh copilot what-the-shell\n\nInteractive mode:\n$ gh copilot',
      },
      {
        title: 'Tips',
        body: 'TIP: Run gh copilot without arguments to enter interactive mode.\n\nTIP: Uses GPT models under the hood -- great for learning git and shell commands.\n\nTIP: You must be logged in with gh auth login and have a Copilot subscription.',
      },
    ],
    website: 'https://github.com/github/gh-copilot',
  },

  az: {
    id: 'az',
    title: 'Azure CLI',
    subtitle: 'Command-line tools for managing Azure resources',
    sections: [
      {
        title: 'What is it?',
        body: 'The Azure CLI (az) is a set of commands used to create, manage, and delete Azure resources. It is installed via Homebrew and provides full access to the Azure platform from the terminal.',
      },
      {
        title: 'Basic Usage',
        body: 'Authenticate with Azure:\n$ az login\n\nView account info:\n$ az account show\n$ az account list\n\nCreate a resource group:\n$ az group create --name MyGroup --location northeurope\n\nList virtual machines:\n$ az vm list\n$ az vm list --resource-group MyGroup\n\nCreate a web app:\n$ az webapp create --name MyApp --resource-group MyGroup --plan MyPlan',
      },
      {
        title: 'Tips',
        body: 'TIP: Run az configure to set default values like location and resource group.\n\nTIP: Use az --help to see all top-level command groups.\n\nTIP: Use az account set --subscription "My Sub" to switch between subscriptions.',
      },
    ],
    website: 'https://learn.microsoft.com/en-us/cli/azure/',
  },

  docker: {
    id: 'docker',
    title: 'Docker + Docker Compose',
    subtitle: 'Container platform -- Docker CLI and Compose installed via Homebrew',
    sections: [
      {
        title: 'What is it?',
        body: 'Docker is a platform for developing, shipping, and running applications in containers. Docker Compose lets you define and run multi-container applications with a single docker compose command. The Docker CLI is installed via Homebrew and connects to Docker Desktop through WSL2 integration.',
      },
      {
        title: 'Basic Usage',
        body: 'List running containers:\n$ docker ps\n$ docker ps -a\n\nManage images:\n$ docker images\n$ docker pull ubuntu:latest\n$ docker rmi ubuntu:latest\n\nStart and stop Compose services:\n$ docker compose up -d\n$ docker compose down\n$ docker compose logs -f\n\nExecute commands in a running container:\n$ docker exec -it my-container bash\n$ docker exec my-container ls -la\n\nView container logs:\n$ docker logs my-container\n$ docker logs -f my-container',
      },
      {
        title: 'Tips',
        body: 'TIP: Enable Docker Desktop WSL2 integration in Docker Desktop Settings -> Resources -> WSL Integration.\n\nTIP: The bootstrapper script adds your user to the docker group.\n\nTIP: The Docker socket at /var/run/docker.sock must exist for the CLI to connect.',
      },
    ],
    website: 'https://www.docker.com/',
  },

  codex: {
    id: 'codex',
    title: 'Codex CLI',
    subtitle: "OpenAI's command-line tool for code generation",
    sections: [
      {
        title: 'What is it?',
        body: 'OpenAI Codex CLI is a command-line tool for code generation. It uses AI models to generate, explain, and modify code directly in your terminal.',
      },
      {
        title: 'Basic Usage',
        body: 'Generate code from a prompt:\n$ codex "write a python function that sorts a list"\n\nInstall git hooks:\n$ codex --install-hooks\n\nHelp:\n$ codex --help',
      },
      {
        title: 'Tips',
        body: 'TIP: Run codex --install-hooks to set up git hooks for AI-assisted commit messages.\n\nTIP: Requires an OpenAI API key set via the OPENAI_API_KEY environment variable.',
      },
    ],
    website: 'https://github.com/openai/codex',
  },

  opencode: {
    id: 'opencode',
    title: 'OpenCode',
    subtitle: 'AI-powered coding tool that works in your terminal',
    sections: [
      {
        title: 'What is it?',
        body: 'OpenCode is an AI-powered coding tool that runs directly in your terminal. It helps with code generation, explanation, review, and refactoring -- understanding full project context to provide relevant assistance.',
      },
      {
        title: 'Basic Usage',
        body: 'Start interactive session:\n$ opencode\n\nAsk about code:\n$ opencode "describe this code"\n$ opencode "explain this function"\n\nHelp:\n$ opencode --help',
      },
      {
        title: 'Tips',
        body: 'TIP: Works best when run from the project root -- it can understand the full project context.\n\nTIP: Try opencode "explain this function" on selected code for quick understanding of unfamiliar codebases.',
      },
    ],
    website: 'https://opencode.ai/',
  },

  vscode: {
    id: 'vscode',
    title: 'VS Code WSL',
    subtitle: 'Visual Studio Code integration with WSL',
    sections: [
      {
        title: 'What is it?',
        body: 'The bootstrapper does NOT install VS Code in Linux. Instead, install VS Code on Windows with the "Remote -- WSL" extension to get the code command available in your WSL terminal. VS Code then connects to WSL seamlessly for editing Linux files.',
      },
      {
        title: 'Basic Usage',
        body: 'Open current folder:\n$ code .\n\nOpen a file:\n$ code file.rs\n\nManage extensions:\n$ code --list-extensions\n$ code --install-extension rust-lang.rust-analyzer',
      },
      {
        title: 'Tips',
        body: 'TIP: Install extensions from the WSL terminal with code --install-extension.\n\nTIP: Keyboard shortcuts and settings sync automatically between Windows and WSL when you sign in with your GitHub account.',
      },
    ],
    website: 'https://code.visualstudio.com/',
  },

  zed: {
    id: 'zed',
    title: 'Zed',
    subtitle: 'High-performance code editor written in Rust',
    sections: [
      {
        title: 'What is it?',
        body: 'Zed is a high-performance code editor written in Rust by the Atom team. It is multi-threaded, GPU-accelerated, and comes with built-in AI features, LSP support, and collaborative editing capabilities.',
      },
      {
        title: 'Basic Usage',
        body: 'Open current directory:\n$ zed .\n\nOpen a file:\n$ zed file.rs\n\nBlock until editor closes:\n$ zed --wait',
      },
      {
        title: 'Tips',
        body: 'TIP: Set as git editor: git config --global core.editor "zed --wait". Already configured as EDITOR in the fish config.\n\nTIP: Zed relies on WSLg for GUI support -- ensure WSLg is running for the editor window to display.',
      },
    ],
    website: 'https://zed.dev/',
  },

  pwsh: {
    id: 'pwsh',
    title: 'PowerShell',
    subtitle: 'Cross-platform PowerShell for automation and scripting',
    sections: [
      {
        title: 'What is it?',
        body: 'PowerShell Core is the cross-platform version of PowerShell -- an automation and scripting framework that runs alongside bash and fish on Linux. It brings powerful .NET-based scripting to any platform.',
      },
      {
        title: 'Basic Usage',
        body: 'Start PowerShell:\n$ pwsh\n\nRun a command:\n$ pwsh -c "Get-Process | Where-Object CPU -gt 10"\n\nList processes:\nPS> Get-Process\n\nList files:\nPS> Get-ChildItem\n\nOutput text:\nPS> Write-Host "Hello from PowerShell"',
      },
      {
        title: 'Tips',
        body: 'TIP: PowerShell cmdlets follow a Verb-Noun naming convention.\n\nTIP: Run Get-Command to list all available cmdlets.\n\nTIP: PowerShell uses .NET under the hood -- you can call any .NET class directly.',
      },
    ],
    website: 'https://learn.microsoft.com/en-us/powershell/',
  },
  config: {
    id: 'config',
    title: 'Configuration',
    subtitle: 'Understanding the Datakraften YAML config structure in depth',
    sections: [
      {
        title: 'Overview',
        body: 'Datakraften uses a single YAML file at ~/.config/datakraften/config.yaml to define your development environment. For source: default and source: custom, every setting lives in this one file. For source: team, it is a thin pointer — just ~source: team~ and ~url~ — and the actual tool settings are fetched fresh from the remote URL on each `dk apply`.\n\nThe config is split into top-level sections, each controlling a different aspect of your workstation. Here is the full structure:',
      },
      {
        title: 'source',
        body: 'The source field selects where your configuration comes from. Three options:\n\n- default — built-in config embedded in dk, full developer workstation defaults\n- custom — local-only, edit everything yourself. Created with `dk init --custom`\n- team — thin config pointing to a remote YAML URL (single source of truth)\n\nWhen source: team is active, a ~url~ field must be set pointing to a YAML file hosted anywhere accessible via HTTPS. The remote YAML defines every tool setting — ~system_packages~, ~brew_packages~, ~runtimes~, ~shell~, ~editors~, ~ai_tools~, and ~ai_apps~. The local config stores only ~source: team~ and ~url~; it is never a full config.\n\nSet up team config:\n\n$ dk init --team https://example.com/team-config.yaml\n> Fetching remote config from https://example.com/team-config.yaml...\n> ✓ Remote config validated\n\nAfter the URL is set, every `dk apply` fetches the remote YAML fresh and applies it:\n\n$ dk apply\n> Fetching remote config from https://example.com/team-config.yaml...\n> ✓ Remote config loaded\n\nNOTE: Because the remote YAML is fetched fresh each time, updates to the shared config take effect immediately on the next `dk apply`.',
      },
      {
        title: 'system_packages',
        body: "Packages installed via your native package manager (apt, dnf, yum, pacman, or brew). The bootstrapper auto-detects which manager to use.\n\n| system_packages:\n|   - build-essential\n|   - curl\n|   - git\n|   - unzip\n\nThese are installed with sudo (where applicable) before any other step. `dk apply` skips already-installed packages automatically.\n\nTIP: Use your package manager's native name for each package — dk passes the name directly to apt install or dnf install.",
      },
      {
        title: 'brew_packages',
        body: "Packages installed via Homebrew. Datakraften installs Homebrew if it's not already present, then uses it for the tools listed here. Most developer tooling comes through brew.\n\n| brew_packages:\n|   - fish\n|   - starship\n|   - atuin\n|   - fzf\n|   - fd\n|   - broot\n|   - bottom\n|   - gh\n|   - docker\n|   - docker-compose\n|   - powershell\n\nTIP: Homebrew keeps packages in /home/linuxbrew/.linuxbrew — isolated from system packages, which means cleaner upgrades and fewer conflicts.",
      },
      {
        title: 'runtimes',
        body: "Programming language runtimes and their version managers. Each runtime has its own enabled flag and optional version.\n\n| runtimes:\n|   node:\n|     enabled: true\n|     manager: fnm\n|     version: lts\n|   python:\n|     enabled: true\n|     manager: uv\n|     version: latest\n|   go:\n|     enabled: true\n|     manager: brew\n|     version: latest\n|   dotnet:\n|     enabled: false\n|     manager: brew\n|     version: latest\n\n- Node.js is managed by fnm (Fast Node Manager). Set version to lts for the latest LTS, or pin a specific version like 20 or 22.\n- Python is managed by uv. It installs the latest stable Python version and sets up uv as the default package manager.\n- Go is installed via Homebrew. Enabled by default.\n- .NET SDK is installed via Homebrew. Disabled by default.\n\nTIP: Set enabled: false for runtimes you don't need — `dk apply` will skip them entirely.",
      },
      {
        title: 'shell',
        body: 'Shell configuration managed by Datakraften. Currently supports Fish shell with managed config blocks.\n\n| shell:\n|   fish:\n|     enabled: true\n|     managed_config: true\n\nWhen enabled, `dk apply`:\n- Installs Fish shell via Homebrew\n- Sets Fish as your default shell (effective after re-login)\n- Writes a managed config to ~/.config/fish/config.fish with Homebrew PATH, fnm integration, uv completions, Atuin, FZF, Starship, and aliases\n- Sets managed blocks so Datakraften can safely re-apply without duplicating entries\n\nTIP: You can edit ~/.config/fish/config.fish freely — Datakraften only touches sections between managed block markers.',
      },
      {
        title: 'editors',
        body: 'Code editor detection and auto-install. Datakraften detects which editors are already installed and optionally installs Zed.\n\n| editors:\n|   vscode:\n|     enabled: true\n|   zed:\n|     enabled: true\n|   cursor:\n|     enabled: true\n\n- VS Code — detected on the Windows side when running under WSL (via Windows Registry), or as a native Linux install. Not installed by Datakraften; install manually from code.visualstudio.com.\n- Zed — detected on Linux/WSL. Can be auto-installed by `dk apply` if missing.\n- Cursor — detected on Windows side under WSL.\n\nTIP: On WSL, editors installed on Windows are detected automatically and the code (or cursor) command is made available inside your Linux environment.',
      },
      {
        title: 'ai_tools & ai_apps',
        body: "AI-powered developer tools — the differentiating feature of Datakraften. Two sections control what gets installed:\n\n**~ai_tools~** — CLI tools installed via npm or Homebrew:\n\n| ai_tools:\n|   codex:\n|     enabled: true\n|     manager: npm\n|     version: latest\n|   opencode:\n|     enabled: true\n|     manager: brew\n|     version: latest\n|   copilot:\n|     enabled: true\n|     manager: brew\n|     version: latest\n|   claude:\n|     enabled: false\n|     manager: npm\n|     version: latest\n|   gemini:\n|     enabled: false\n|     manager: npm\n|     version: latest\n\n**~ai_apps~** — desktop apps installed via Homebrew Cask or VS Code extension:\n\n| ai_apps:\n|   codex:\n|     enabled: true\n|     manager: brew\n|     version: latest\n|   claude:\n|     enabled: false\n|     manager: brew\n|     version: latest\n|   copilot:\n|     enabled: true\n|     manager: vscode\n|     version: latest\n\n`dk apply` handles both sections. CLI tools are installed first, then desktop apps. Apps that require a specific platform (e.g., macOS-only brew casks) are skipped gracefully.\n\nTIP: Some AI tools require API keys or subscriptions (GitHub Copilot subscription, OpenAI API key). Check each tool's documentation for authentication requirements.",
      },
    ],
    website: '',
  },

  install: {
    id: 'install',
    title: 'Installation',
    subtitle: 'Supported platforms, requirements, and how to get Datakraften on your machine',
    sections: [
      {
        title: 'Supported platforms',
        body: 'Datakraften is designed for developer workstations. The primary target is WSL2 (Ubuntu or Fedora), but it also works on native Linux and macOS.\n\n- **WSL2** (primary) — Ubuntu or Fedora distro on Windows. Full support for Windows-side editor detection (VS Code, Cursor) and Docker Desktop integration.\n- **Linux** (native) — works on any Debian/Ubuntu or Fedora-based distro. APT, DNF, YUM, and PACMAN package managers are supported.\n- **macOS** (experimental) — works with Homebrew as the native package manager. Some features (WSL-specific editor detection, Docker Desktop WSL integration) are skipped.\n\nNOTE: Datakraften does not support Windows itself (no native PowerShell or cmd support). Use WSL2 for the full experience.',
      },
      {
        title: 'System requirements',
        body: 'Before installing, make sure your system meets these requirements:\n\n- **OS** — Windows 10+ with WSL2 enabled (recommended), or any Linux distro with APT/DNF/YUM/PACMAN\n- **Architecture** — x86_64 or ARM64\n- **Internet** — the bootstrap script downloads tools from GitHub, npm, and Homebrew\n- **sudo access** — system packages are installed with sudo\n- **curl or wget** — needed for the one-liner install\n\nMinimal disk usage: the dk CLI binary is ~15 MB. Total install (with all tools) can be 1-3 GB depending on the profile.',
      },
      {
        title: 'Quick install (recommended)',
        body: 'One command to download and run the bootstrap script:\n\n$ curl -fsSL https://datakraften.no/install | bash\n\nWhat the script does:\n1. Detects your OS and architecture\n2. Downloads the latest dk binary from GitHub Releases\n3. Verifies the SHA256 checksum\n4. Installs to ~/.local/bin/dk\n5. Adds ~/.local/bin to PATH if not already present\n\nAfter the script completes, close and reopen your terminal, or run `source ~/.profile`.\n\nTIP: The script is intentionally minimal — it only installs the dk CLI. Run `dk init` and `dk apply` next to bootstrap your full environment.',
      },
      {
        title: 'Manual install',
        body: 'Download the latest release from GitHub, verify the checksum, and install manually:\n\n$ curl -fsSL -o dk https://github.com/sagathelab/datakraften/releases/latest/download/dk-linux-amd64\n$ curl -fsSL -o dk.sha256 https://github.com/sagathelab/datakraften/releases/latest/download/dk-linux-amd64.sha256\n$ sha256sum --check dk.sha256\n$ chmod +x dk\n$ mkdir -p ~/.local/bin\n$ mv dk ~/.local/bin/\n\nReplace `linux-amd64` with `linux-arm64`, `darwin-amd64`, or `darwin-arm64` as needed.\n\nOr use the package manager of your choice:\n\n$ brew install sagathelab/tap/datakraften\n\nNOTE: The Homebrew tap is community-maintained. For the latest version, use the direct download or the bootstrap script.',
      },
    ],
    website: 'https://github.com/sagathelab/datakraften',
  },

  teams: {
    id: 'teams',
    title: 'Team Configs',
    subtitle: 'Standardized dev environments across your whole team with shared YAML configs',
    sections: [
      {
        title: 'Why team profiles?',
        body: "Every team has a stack — specific tools, runtimes, linters, and conventions. Without automation, each new hire spends hours (or days) setting up their machine, and inconsistencies creep in across the team. Datakraften's team config solves this with a single shared YAML file.\n\nThe idea is simple: define your team's ideal workstation once, host the YAML file somewhere your team can reach it, and each developer's local config becomes a thin pointer — just ~source: team~ and ~url~. On every `dk apply`, the remote YAML is fetched fresh, validated, and applied. No local copy, no drift.\n\n$ curl -fsSL https://datakraften.no/install | bash\n$ dk init --team https://example.com/team.yaml\n$ dk apply\n\nNOTE: Onboarding goes from hours to minutes. Every machine is identical. No tribal knowledge needed. Updates to the shared config take effect on the next `dk apply`.",
      },
      {
        title: 'Setting up a team config',
        body: "The remote YAML defines everything — it IS the config. Create a file with the exact toolset your team needs and host it somewhere accessible. Each developer's local config only needs:\n\n| source: team\n| url: https://raw.githubusercontent.com/your-org/team-config/main/datakraften.yaml\n\nHere is an example of what the remote YAML file itself might look like:\n\n| system_packages:\n|   - build-essential\n|   - curl\n|   - git\n|   - unzip\n|   - postgresql-client\n|   - redis-tools\n|\n| brew_packages:\n|   - fish\n|   - starship\n|   - atuin\n|   - fzf\n|   - gh\n|   - docker\n|\n| runtimes:\n|   node:\n|     enabled: true\n|     version: lts\n|   python:\n|     enabled: true\n|   dotnet:\n|     enabled: true\n|\n| shell:\n|   fish:\n|     enabled: true\n|     managed_config: true\n|\n| editors:\n|   vscode:\n|     enabled: true\n|   zed:\n|     enabled: true\n|\n| ai_tools:\n|   codex:\n|     enabled: true\n|   opencode:\n|     enabled: true\n|   copilot:\n|     enabled: true\n|   claude:\n|     enabled: false\n|   gemini:\n|     enabled: false\n| ai_apps:\n|   codex:\n|     enabled: true\n|   claude:\n|     enabled: false\n|   copilot:\n|     enabled: true\n\nNOTE: Commit this file to your team's infrastructure repo or a shared gist, then share the raw URL. Every developer who runs `dk init --team <url>` gets exactly this setup — and every `dk apply` re-fetches it fresh, so updates are instant.",
      },
      {
        title: 'Hosting your YAML',
        body: "Any HTTPS-accessible URL works. Popular options:\n\n- GitHub repository — commit the YAML to your team's repo and use the raw.githubusercontent.com URL\n- GitHub Gist — create a secret or public gist and use the raw URL\n- Internal wiki — host it behind your company's SSO or VPN\n- S3 / Cloud Storage — with a signed URL or public bucket\n- Your own server — any static file server works\n\nThe URL is set during `dk init --team <url>`. Datakraften fetches the remote YAML to validate it, then stores only the URL locally. On every `dk apply`, the remote YAML is re-fetched fresh — the remote file is always the single source of truth. This means the environment is always up to date with the latest version of the shared config.\n\nTIP: Use a versioned path (e.g. /v1/datakraften.yaml or a git tag) to control when team members pick up changes. When you update the shared config, tell your team to run `dk apply` to pull the latest version immediately.",
      },
      {
        title: 'Best practices',
        body: "Start with the default profile, test it, then iterate:\n\n- Begin with the default profile on your own machine — run `dk init` then `dk apply`\n\n- Customize your remote YAML for your team's stack. Add specific packages, remove what you don't need.\n\n- Host the YAML and share the raw URL. Ask a teammate to run `dk init` with it on a fresh machine.\n\n- Keep the remote YAML minimal — install only what every developer needs. Let individuals add their own tools on top.\n\n- Version your config. Use a git tag or a versioned path so you can roll back if something breaks.\n\n- Document exceptions. Not everything fits in YAML. Add a comment or a README alongside your remote config for manual steps.\n\n- Run `dk doctor` after `dk apply` to verify everything is set up correctly.\n\n- Update periodically. Point your team to run `dk apply` when you push a new version of the remote YAML — changes take effect immediately.",
      },
      {
        title: 'Example: Node.js team config',
        body: 'A minimal remote YAML for a Node.js backend team. Developers point their local config to this URL and get the exact same environment:\n\n| system_packages:\n|   - build-essential\n|   - curl\n|   - git\n|\n| brew_packages:\n|   - fish\n|   - starship\n|   - gh\n|   - docker\n|\n| runtimes:\n|   node:\n|     enabled: true\n|     version: lts\n|   python:\n|     enabled: false\n|   dotnet:\n|     enabled: false\n|\n| shell:\n|   fish:\n|     enabled: true\n|\n| editors:\n|   vscode:\n|     enabled: true\n|\n| ai_tools:\n|   codex:\n|     enabled: true\n|   opencode:\n|     enabled: true\n|   copilot:\n|     enabled: true\n|   claude:\n|     enabled: false\n|   gemini:\n|     enabled: false\n| ai_apps:\n|   codex:\n|     enabled: true\n|   claude:\n|     enabled: false\n|   copilot:\n|     enabled: true\n\nNOTE: Every developer gets Node.js LTS, fish shell with Starship, GitHub CLI, Docker, VS Code detection, and AI tools — nothing more, nothing less.',
      },
      {
        title: 'Example: Full-stack team config',
        body: 'A comprehensive remote YAML for a full-stack team working with Node.js, Python, Go, and .NET:\n\n| system_packages:\n|   - build-essential\n|   - curl\n|   - git\n|   - unzip\n|   - postgresql-client\n|   - redis-tools\n|   - jq\n|\n| brew_packages:\n|   - fish\n|   - starship\n|   - atuin\n|   - fzf\n|   - fd\n|   - broot\n|   - bottom\n|   - gh\n|   - docker\n|   - powershell\n|\n| runtimes:\n|   node:\n|     enabled: true\n|     version: lts\n|   python:\n|     enabled: true\n|   go:\n|     enabled: true\n|   dotnet:\n|     enabled: true\n|\n| shell:\n|   fish:\n|     enabled: true\n|\n| editors:\n|   vscode:\n|     enabled: true\n|   zed:\n|     enabled: true\n|\n| ai_tools:\n|   codex:\n|     enabled: true\n|   opencode:\n|     enabled: true\n|   copilot:\n|     enabled: true\n|   claude:\n|     enabled: false\n|   gemini:\n|     enabled: false\n| ai_apps:\n|   codex:\n|     enabled: true\n|   claude:\n|     enabled: false\n|   copilot:\n|     enabled: true\n\nNOTE: Extends the minimal setup with additional system tools (postgresql-client, redis-tools, jq) and PowerShell — all from a single shared URL.',
      },
    ],
    website: '',
  },
}

export const toolsList = Object.values(tools)

export const categories = [
  {
    title: 'dk CLI',
    desc: 'Core Datakraften platform',
    ids: ['dk', 'install'],
  },
  {
    title: 'Guides',
    desc: 'Configuration and team workflow guides',
    ids: ['config', 'teams'],
  },
  {
    title: 'Runtimes',
    desc: 'Programming language runtimes and version managers',
    ids: ['node', 'python', 'go', 'dotnet'],
  },
  {
    title: 'Shell & CLI',
    desc: 'Shells, prompts, and terminal productivity tools',
    ids: ['fish', 'starship', 'atuin', 'fzf', 'fd', 'broot', 'btm'],
  },
  {
    title: 'Cloud & Dev',
    desc: 'Cloud platforms, containers, and package managers',
    ids: ['brew', 'gh', 'az', 'docker'],
  },
  {
    title: 'Editors & Tools',
    desc: 'Code editors and AI-powered development tools',
    ids: ['vscode', 'zed', 'codex', 'opencode', 'gh-copilot', 'pwsh'],
  },
]
