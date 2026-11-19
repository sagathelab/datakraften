export interface ToolSection {
  title?: string
  body: string
  id?: string
}

export interface ToolTask {
  title: string
  command?: string
  details?: string
}

export interface ToolFaq {
  question: string
  answer: string
}

export interface ToolDef {
  id: string
  title: string
  subtitle: string
  sections: ToolSection[]
  website: string
  websiteLabel?: string
  summary?: string
  supportedPlatforms?: string[]
  commonTasks?: ToolTask[]
  troubleshooting?: ToolFaq[]
  faqs?: ToolFaq[]
  keywords?: string[]
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
        body: 'Create a workstation configuration in `~/.config/datakraften/config.yaml`. This path does not use `XDG_CONFIG_HOME`. Initialization writes configuration; run `dk apply` separately to install and configure tools.\n\nGenerate the built-in defaults:\n$ dk init\n\nChoose a custom starting point interactively, or supply a local YAML file:\n$ dk init --custom\n$ dk init --custom ./workstation.yaml\n\nImport a template once from an HTTPS URL that serves raw YAML:\n$ dk init --from https://example.com/workstation.yaml\n\nDatakraften validates and merges the template into the built-in defaults, not into your existing file. It shows the template host and a summary of overrides, with text values hidden. The result is an editable local `source: custom` file. If a config already exists, dk asks before overwriting it unless you use `--yes`.',
      },
      {
        title: 'dk apply',
        id: 'apply',
        body: "Apply the workstation configuration from `~/.config/datakraften/config.yaml`:\n$ dk apply\n\nPreview without applying changes:\n$ dk apply --dry-run\n\nApply reads the local configuration. It does not fetch an imported template again, though installers can still download software. It checks the baseline system and Homebrew packages, then applies enabled runtimes, editors, AI tools, and the selected shell's managed block. It may request sudo for system operations or an explicitly approved npm retry.\n\nNOTE: A repository's `datakraften.yaml` is separate; `dk apply` does not execute its tasks.",
      },
      {
        title: 'dk doctor',
        id: 'doctor',
        body: 'Diagnose the system, tools, runtimes, editors, Docker, and shell. The report separates configured intent, what is detected, and the last apply. For shells, it distinguishes the configured target from the shell named by `SHELL` and reports the managed blocks it finds.\n\n$ dk doctor\n$ dk doctor --json',
      },
      {
        title: 'dk status',
        body: 'See your configured workstation settings, detected tools and versions, and the last apply as separate sections. A missing config does not prevent inspection; an invalid config is reported as an error.\n\n$ dk status\n$ dk status --json',
      },
      {
        title: 'Config sources',
        id: 'sources',
        body: 'Both workstation workflows produce a local file at `~/.config/datakraften/config.yaml`:\n\n- `source: default` — the built-in defaults written by `dk init`\n- `source: custom` — an editable local config, including files created with `dk init --custom` and `dk init --from`\n\nChanges to a hosted template do not update a local copy. To use a newer template, import it explicitly or edit your local file. A repository’s `datakraften.yaml` is separate from the workstation config.',
      },
      {
        title: 'Project environments',
        id: 'project',
        body: 'A repository can describe its Node.js requirement, non-secret environment values, and tasks in `datakraften.yaml`:\n\n| version: 1\n| project:\n|   tools:\n|     node:\n|       version: "22"\n|       manager: nvm\n|   env:\n|     NODE_ENV: development\n|   tasks:\n|     dev: npm run dev\n|     test: npm test\n\nFrom the repository or a subdirectory, inspect the project without running its tasks or installing anything:\n$ dk project status\n\nReview the file before trusting it. Then preview or install a missing Node version:\n$ dk project trust\n$ dk project setup --dry-run\n$ dk project setup\n\nRun named tasks from the trusted configuration. They execute in the directory containing the project file:\n$ dk run dev\n$ dk run test\n\nNode.js is the only project runtime currently supported. The short `node: "22"` form uses fnm; the mapping above selects nvm. Use numeric major, major.minor, or major.minor.patch versions. Install the selected manager first: fnm must be on `PATH`, or dk must be able to find `nvm.sh` in `NVM_DIR` or a supported location. Setup installs missing versions in a Datakraften-managed tools directory without changing your normal Node default.\n\nNOTE: Tasks execute as your user and may access files, the network, and inherited environment. Do not store secrets in shared YAML.',
      },
      {
        title: 'dk upgrade',
        id: 'upgrade',
        body: 'Upgrade the Datakraften CLI to the latest release. Downloads the correct binary for your platform from GitHub, verifies the SHA256 checksum, and performs an atomic replacement.\n\n$ dk upgrade\n\n> Current version: v0.1.0\n> Latest version:  v0.2.0\n> Downloading dk-linux-amd64...\n> Checksum verified ✓\n> Updated to v0.2.0\n\nHow it works:\n- Fetches the latest release from api.github.com\n- Detects your OS and architecture to download the matching binary\n- Verifies the binary against its SHA256 checksum\n- Replaces the current binary atomically\n\nTIP: Run `dk upgrade` periodically to get the latest version of Datakraften.',
      },
      {
        title: 'dk update',
        id: 'update',
        body: 'Update the default selection of developer tools:\n$ dk update\n\nThe Homebrew step only targets installed formulae on the Datakraften default list. This includes matching formulae you installed yourself. It uses local Homebrew metadata and disables automatic cleanup. Casks are not targeted, but Homebrew can still change dependencies or dependents.\n\nPreview or run just the scoped Homebrew update:\n$ dk update brew --dry-run\n$ dk update brew\n\nOpt in to a full Homebrew refresh and upgrade, including eligible casks:\n$ dk update brew-all --dry-run\n$ dk update brew-all\n\nOther targets are `fnm`, `uv`, and `npm`. The npm step only updates enabled AI tools installed in the active global npm prefix.\n$ dk update --list',
      },
    ],
    website: 'https://github.com/sagathelab/datakraften',
    websiteLabel: 'GitHub repository',
    summary:
      'Use dk to configure your workstation from local YAML, import an HTTPS template once, and run trusted project tasks with fnm or nvm.',
    supportedPlatforms: ['WSL', 'Linux', 'macOS (partial support)'],
    commonTasks: [
      {
        title: 'Initialize a default config',
        command: 'dk init',
        details:
          'Generate a full default config and detect package manager/platform automatically.',
      },
      {
        title: 'Apply the workstation setup',
        command: 'dk apply',
        details:
          'Install runtimes, shell tooling, editors, cloud tools, and AI tools from your config.',
      },
      {
        title: 'Diagnose your machine',
        command: 'dk doctor',
        details: 'Check system, tools, runtimes, editors, Docker, and shell state.',
      },
      {
        title: 'Refresh managed tools',
        command: 'dk update',
        details:
          'Update selected Homebrew formulae, fnm, uv, and enabled npm AI tools. A full Homebrew upgrade requires dk update brew-all.',
      },
    ],
    troubleshooting: [
      {
        question: 'Why did my workstation template import fail?',
        answer:
          'Use an accessible HTTPS URL that returns raw YAML, not an HTML file preview. Check that the response contains one YAML mapping with valid workstation keys and values, then retry `dk init --from`.',
      },
      {
        question: 'Why does `dk update` fail on npm permissions?',
        answer:
          'System-owned npm paths can cause EACCES. When an update fails for permissions, dk offers an interactive sudo retry for the selected packages and current prefix. Alternatively, use a Node version manager or preview `dk npm setup --dry-run` and run `dk npm setup` for a user-owned prefix. Existing packages are not moved; custom prefixes can conflict with version managers.',
      },
    ],
    faqs: [
      {
        question: 'Does Datakraften replace apt, brew, fnm, or uv?',
        answer:
          'No. Datakraften orchestrates existing package managers and runtime tools instead of replacing them.',
      },
      {
        question: 'Is `dk apply` safe to run repeatedly?',
        answer:
          'Yes. The apply flow is designed to be idempotent and should only install or configure what is missing.',
      },
      {
        question: 'Can I share workstation settings with my team?',
        answer:
          'Share a raw YAML template over HTTPS. Each developer imports it once with dk init --from, reviews the result, and edits their own local configuration. Later template changes are not fetched by dk apply.',
      },
    ],
    keywords: [
      'dk cli',
      'developer workstation bootstrap',
      'WSL setup',
      'Linux setup',
      'team onboarding',
      'workstation template import',
      'trusted project tasks',
    ],
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
        body: 'fnm manages multiple installed Node.js versions. Its shell integration can switch versions when you enter a project with a .nvmrc or .node-version file, provided fnm is activated in that shell. Datakraften configures this integration in its managed Fish setup; existing Bash and other shells may need their own setup.\n\nWorkstation setup uses fnm. For a Datakraften project, declare a numeric Node version in `datakraften.yaml` and use `dk project status`, `dk project trust`, `dk project setup`, and `dk run`. The `manager` field selects fnm or nvm; the short version form defaults to fnm. dk puts the selected Node on the task’s `PATH` without changing the global default.',
      },
      {
        title: 'Tips',
        body: 'TIP: For workstation Node versions managed by Datakraften, use fnm rather than installing Node directly via apt or brew.\n\nTIP: Use `fnm install --lts` to install the latest LTS release.',
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
        body: 'TIP: Run `gh auth login` first for commands that need GitHub access.\n\nTIP: `gh status` shows your latest notifications. GitHub Copilot CLI is a separate `copilot` command.',
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
        body: 'GitHub Copilot CLI is a standalone coding assistant that can explain code, propose edits, and run tools from your terminal. It replaces the discontinued `gh copilot` extension.',
      },
      {
        title: 'Install',
        body: 'Enable `ai_tools.copilot.enabled` in your workstation configuration and run `dk apply`. Datakraften uses the official Homebrew cask on supported systems. It leaves an existing `copilot` on `PATH` alone.\n\nWith Homebrew, you can also install it directly:\n$ brew install --cask copilot-cli\n\nNOTE: Routine `dk update` does not target casks. Follow the official Copilot CLI instructions for Copilot updates.',
      },
      {
        title: 'Basic Usage',
        body: 'Start an interactive session:\n$ copilot\n\nUse `/login` in Copilot CLI to sign in, then ask a question in your own words. Review proposed edits and commands before approving them.\n\nNOTE: You need a GitHub Copilot plan. Your organization may also need to enable Copilot CLI.',
      },
    ],
    website: 'https://docs.github.com/en/copilot/how-tos/copilot-cli/install-copilot-cli',
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
    subtitle: 'Workstation templates and project environments',
    sections: [
      {
        title: 'Overview',
        body: 'Datakraften uses two separate YAML files:\n\n- `~/.config/datakraften/config.yaml` describes your workstation. This path is fixed under your home directory; `XDG_CONFIG_HOME` does not change it.\n- `datakraften.yaml` in a repository describes project tools, environment values, and tasks. Project commands do not rewrite the workstation configuration.',
      },
      {
        title: 'Importing workstation defaults',
        body: 'Use an HTTPS URL that returns raw YAML, not an HTML file preview:\n\n$ dk init --from https://example.com/workstation.yaml\n\nDatakraften fetches the template during initialization and merges it recursively with the built-in defaults, not with your existing configuration. Supplied values override defaults, including `false`; omitted keys retain defaults. The template must be one YAML mapping; unknown settings, invalid types, and null values are rejected.\n\nThe CLI shows the template host and a summary of overrides with text values hidden. Older supported file formats can contain settings such as `runtimes.node.version` or `tools` that dk reads but does not apply; those settings produce a warning. Review the saved `source: custom` file before applying it.\n\nNOTE: If a config already exists, the import asks before replacing it unless you use `--yes`. Later `dk apply` runs read the local file without fetching the template again. Installing software may still use the network.',
      },
      {
        title: 'Creating a local workstation configuration',
        body: 'Run `dk init` to write the built-in defaults as `source: default`. Run `dk init --custom` to choose an empty skeleton or a starting point pre-filled with defaults. To supply settings from a local YAML file instead:\n\n$ dk init --custom ./workstation.yaml\n\nUnlike the HTTPS import, the local-file option merges supplied settings into the existing configuration, or starts from the supplied settings if there is no configuration yet. Nested maps merge recursively. If the file exists, dk asks before overwriting it unless you use `--yes`.',
      },
      {
        title: 'Workstation settings',
        body: 'For example, a partial workstation import can select Bash, disable the workstation Node setup, and enable Claude:\n\n| shell:\n|   default: bash\n| runtimes:\n|   node:\n|     enabled: false\n| ai_tools:\n|   claude:\n|     enabled: true\n\n`shell.default` accepts `fish`, `bash`, `zsh`, or `none`. `dk init` defaults to Fish. Omitting the shell from a standalone local config leaves shell setup off. Only Fish can be installed by `dk apply`; Bash and Zsh must already be on `PATH`. Switching or disabling shell setup leaves previous managed blocks in place and never changes the login shell.\n\nUse `enabled` for supported runtimes and AI entries, and `editors.zed` for Zed. A disabled entry does not uninstall anything. Baseline system and Homebrew packages still apply; project Node version selection uses the separate project file.',
      },
      {
        title: 'Project configuration',
        body: 'Create datakraften.yaml at your project root:\n\n| version: 1\n| project:\n|   tools:\n|     node: "22"\n|   env:\n|     NODE_ENV: development\n|   tasks:\n|     test: npm test\n\nIn a Git repository, project commands search the current directory and its parents up to the nearest repository root; the nearest file wins. Git worktrees and nested repositories have their own boundary. Outside a repository, only the current directory is checked. The YAML file itself must be a regular file, not a symlink.\n\nProject runtime support currently covers Node.js via fnm or nvm. To select nvm, write node as a mapping with version: "22" and manager: nvm. The short node: "22" form uses fnm. Use numeric major, major.minor, or major.minor.patch versions. Project values affect the task process and its children, not your parent shell or global Node default. `dk init` and `dk apply` continue to use the workstation configuration.',
      },
      {
        title: 'Trust and tasks',
        body: 'Use `dk project status` to see the discovered file, trust state, task names, and Node.js availability. It does not execute tasks, invoke the version manager, or install anything. `dk project trust` shows task commands and environment variable names, not values. Review the file, including its environment values, before approving it. Trust is tied to the file location and contents; moving or editing the file requires approval again.\n\n$ dk project status\n$ dk project trust\n$ dk project setup --dry-run\n$ dk project setup\n$ dk run test\n\nSetup checks for a usable Node version and, if one is missing, asks for trust and installation approval before installing it with the configured manager. `dk run` executes a named task only once the project is trusted and its declared Node version is available; it does not install Node for you. In a non-interactive session, review the file and use `dk project trust --yes` to grant trust explicitly. `dk project setup --yes` approves installation, not project trust.\n\nNOTE: Tasks run as your user, not in a sandbox. Do not store secrets in shared YAML. Tasks themselves can print inherited or project environment values.',
      },
    ],
    website: '',
    summary:
      'Configure dk with local workstation YAML, import a raw HTTPS template once, and use version 1 datakraften.yaml for trusted project tasks.',
    keywords: ['dk init --from', 'workstation YAML', 'datakraften.yaml', 'fnm', 'nvm'],
    commonTasks: [
      {
        title: 'Import a shared workstation template',
        command: 'dk init --from <https-url>',
        details:
          'Merge a partial HTTPS YAML template with defaults and save a local, editable config.',
      },
      {
        title: 'Start from a local YAML file',
        command: 'dk init --custom ./workstation.yaml',
        details: 'Use a local file instead of downloading a template.',
      },
      {
        title: 'Inspect a project',
        command: 'dk project status',
        details: 'Review the discovered project file, Node version, trust and tasks.',
      },
    ],
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
        body: 'Before installing, make sure your system meets these requirements:\n\n- **OS** — Windows 10+ with WSL2 enabled (recommended), or any Linux distro with APT/DNF/YUM/PACMAN\n- **Architecture** — x86_64 or ARM64\n- **Internet** — installing dk and workstation tools requires downloads\n- **sudo access** — system packages are installed with sudo\n- **curl or wget** — needed for the one-liner install\n\nMinimal disk usage: the dk CLI binary is ~15 MB. A workstation setup with tools can take 1–3 GB depending on your configuration.',
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
    title: 'Sharing Configurations',
    subtitle: 'Share a workstation starting point and configure each repository',
    sections: [
      {
        title: 'Share a workstation starting point',
        body: 'Host a partial workstation YAML file at an HTTPS URL that serves the raw file, not a web preview. Each developer can import it once:\n\n$ dk init --from https://example.com/workstation.yaml\n\nReview the source before importing. dk validates the YAML, merges it into the built-in defaults, and summarizes the overrides. The import does not install tools. Review the saved file, then preview and apply it:\n\n$ dk apply --dry-run\n$ dk apply',
      },
      {
        title: 'Example: partial workstation template',
        body: 'This partial file disables Python and enables Claude in the imported workstation configuration:\n\n| runtimes:\n|   python:\n|     enabled: false\n| ai_tools:\n|   claude:\n|     enabled: true\n\nOther settings keep their built-in defaults. Importing the template does not install tools. Run `dk apply` separately after reviewing the resulting config.\n\nNOTE: A shared template is a starting point, not a guarantee of identical machines.',
      },
      {
        title: 'Local ownership and updates',
        body: 'After import, `~/.config/datakraften/config.yaml` is an editable `source: custom` file. `dk apply` uses this local copy without contacting the template host again for the YAML. To pick up changes, edit the local file or explicitly re-import. A new import starts from built-in defaults and asks before replacing an existing config. Keep a copy of local changes you want to retain.\n\nUse a versioned raw HTTPS URL when you need a specific template revision. Do not put passwords, tokens, or secrets in the YAML or URL.',
      },
      {
        title: 'Project configuration for each repository',
        body: 'Commit a separate `datakraften.yaml` to each repository. It can declare a numeric Node version, tasks, and non-secret environment values:\n\n| version: 1\n| project:\n|   tools:\n|     node: "22"\n|   env:\n|     NODE_ENV: development\n|   tasks:\n|     test: npm test\n\n`dk project status` inspects the file without running tasks or installing tools. Review the file, then approve it with `dk project trust`. Use `dk project setup` to install the declared Node version if needed and `dk run` to execute a task from the trusted file:\n\n$ dk project status\n$ dk project trust\n$ dk project setup\n$ dk run test\n\nNode.js is the only project runtime currently supported. The short form uses fnm. To use nvm, set `node` to a mapping with `version: "22"` and `manager: nvm`. The selected manager must already be available. Tasks run as your user and are not sandboxed.',
      },
    ],
    website: '',
    summary:
      'Share a raw HTTPS workstation template for one-time import, and commit a separate datakraften.yaml for trusted project tasks.',
    keywords: [
      'workstation templates',
      'team onboarding',
      'dk init --from',
      'project configuration',
    ],
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
    desc: 'Local workstation setup, shared templates, and project workflows',
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
