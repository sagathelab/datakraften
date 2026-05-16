# AGENTS.md — Datakraften

Two independent sub-projects in this repo: a **Go CLI** (`cmd/dk/`) and a **website** (`site/`).

## Go CLI (Datakraften / `dk`)

- **Go 1.23**, module `github.com/sagathelab/datakraften`
- Entrypoint: `cmd/dk/main.go`, all logic in `internal/`
- Dependencies: Cobra, Viper, survey (no other major frameworks)
- `make build` → `go build -o bin/dk ./cmd/dk/`
- `make test` → `go test ./...`
- `make lint` → `golangci-lint run ./...`
- `make install` → copies to `~/.local/bin/dk`
- CI order: golangci-lint → `go build` → `go test ./...`
- Release: cross-compiles 4 targets (linux-amd64, linux-arm64, darwin-amd64, darwin-arm64), attached to GitHub Releases with SHA256 + `install` script
- **No comments** in code unless logic is genuinely non-obvious
- Every install/configure function must be idempotent (safe to re-run)
- `internal/profiles/` and `internal/tools/` are **empty** (planned, not yet wired)
- Config uses `source: default | custom | team`, not `profile:` (old `profile:` key is transparently mapped)
- Config keys: `ai_tools`, `ai_apps` (not the old `ai:` key)
- Managed shell blocks use `# >>> datakraften >>>` / `# <<< datakraften <<<` markers
- `--dry-run` and `--json` are global flags on all commands

## Website (`site/`)

- **Bun** runtime (not npm). Run all commands from `site/`.
- Stack: Vite + React 19 + TypeScript 6 + Tailwind v4 (CSS-based, no PostCSS config)
- `bun run dev` — Vite HMR dev server
- `bun run build` — `tsc -b && vite build && bun scripts/prerender.tsx` (3 steps in sequence, **must pass before committing**)
- `bun run lint` — ESLint + stylelint
- `bun run format` — Prettier (singleQuote, trailingComma all, printWidth 100, no semi)
- `bun run format:check` — Prettier check (runs in CI)
- All doc content is a single source of truth: `src/data/tools.ts` (no markdown files)
- Doc pages use a custom mini-syntax in section bodies (backtick inline code, `~tilde~` highlights, `$`/`>` code blocks, `| ` YAML blocks, `TIP:`/`NOTE:` boxes)
- Stale `site/README.md` is Vite boilerplate — **do not use as reference**; use `SITE.md` instead
- Deployed via GitHub Actions on push to `main` touching `site/**`
