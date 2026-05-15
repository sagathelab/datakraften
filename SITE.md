# Datakraften Website & Documentation

## Stack

- **Runtime**: Bun
- **Framework**: Vite + React 19 + TypeScript 6
- **Styling**: Tailwind CSS v4 (CSS-based, no PostCSS config)
- **Routing**: react-router-dom v7
- **Linting**: ESLint (with typescript-eslint, react-hooks, prettier) + stylelint
- **Formatting**: Prettier

## Directory Structure

```
site/
├── public/                  # Static files (CNAME, install, robots.txt, llms.txt)
│   ├── CNAME               # GitHub Pages custom domain
│   ├── install             # Bootstrap install script (served as /install)
│   ├── robots.txt
│   └── llms.txt
├── src/
│   ├── main.tsx            # Entry point (BrowserRouter wrapper)
│   ├── App.tsx             # Route definitions
│   ├── index.css           # Tailwind imports + all custom CSS
│   ├── data/
│   │   └── tools.ts        # ALL documentation content + ToolDef/ToolSection types
│   ├── components/
│   │   ├── Layout.tsx      # Base layout (navbar + footer + children)
│   │   ├── Navbar.tsx      # Top navigation (logo + docs link + page title)
│   │   ├── Footer.tsx      # Site footer
│   │   ├── ToolGuide.tsx   # Docs page renderer (CodeBlock, YamlBlock, TOC sidebar)
│   │   ├── Terminal.tsx    # Landing page terminal animation
│   │   ├── LogoSvg.tsx     # Landing page ASCII-art logo
│   │   ├── FeatureCard.tsx # Landing page feature cards
│   │   └── DocCard.tsx     # Docs hub navigation cards
│   └── pages/
│       ├── Landing.tsx     # Landing/homepage
│       ├── DocsHub.tsx     # /docs overview page
│       ├── ToolPage.tsx    # /docs/:id dynamic page
│       ├── Privacy.tsx     # /privacy page
│       └── Terms.tsx       # /terms page
├── scripts/
│   └── prerender.tsx      # Prerenders route HTML + metadata + sitemap.xml
├── eslint.config.js        # ESLint configuration
├── stylelint.config.js     # Stylelint configuration
├── tsconfig*.json          # TypeScript configuration
└── vite.config.ts          # Vite configuration
```

## Routes

| Path | Component | Description |
|------|-----------|-------------|
| `/` | Landing | Homepage with hero, terminal animation, features, CTA |
| `/docs` | DocsHub | Documentation hub with navigation cards |
| `/docs/:id` | ToolPage | Dynamic doc page (renders `tools.ts` by ID) |
| `/privacy` | Privacy | Privacy policy |
| `/terms` | Terms | Terms of service |

## Documentation Content

All docs content lives in `site/src/data/tools.ts`. Each tool is a `ToolDef`:

```typescript
interface ToolDef {
  id: string              // URL slug (e.g., "dk", "install", "config", "teams")
  title: string           // Page heading
  subtitle: string        // Description below heading
  sections: ToolSection[] // Content sections with title + body
  website: string         // External link URL
}
```

Adding a new doc page:
1. Add a new `ToolDef` entry to the `tools` record in `tools.ts`
2. Each section body uses a custom markdown-like syntax (see below)
3. Static routes are auto-generated from `tools.ts` keys in `scripts/prerender.tsx`

## Documentation Syntax

The `ToolGuide` component renders section bodies with a custom syntax:

### Inline code
- Backticks: `` `code` `` → `<code className="inline-code">` (magenta bordered)
- Tildes: `~key~` → `<mark className="inline-highlight">` (magenta background, for YAML key/value references)
- `dk <command>` inside backticks automatically becomes a clickable link to the command's section

### Code blocks
Lines starting with any of these prefixes are grouped into `<CodeBlock>`:
- `$ ` — command input (prompt shown as magenta `$`)
- `> ` — command output (shown in yellow)
- `// `, `# `, `<!-- ` — comments (shown in italic gray)
- Plain lines inside a code block also work

Empty lines inside code blocks are filtered out. A copy button (clipboard SVG icon) is shown in the top-right corner. When clicked, `$ `, `> `, `// `, `# ` prefixes are stripped for clean copy.

### YAML blocks
Lines starting with `| ` are grouped into `<YamlBlock>`:
- The `| ` prefix is stripped for display
- Content is shown in green (`#7ee787`)
- A copy button strips the `| ` prefix for clean copy

### Tip boxes
Lines starting with `TIP:` create a `<div className="tip-box">` with a magenta play icon.

### Note boxes
Lines starting with `NOTE:` create a `<div className="note-box">` with a dimmed `~` prefix.

### Lists
Lines starting with `- ` are grouped into bullet lists with magenta markers.

## Styling Conventions

- Custom CSS lives entirely in `src/index.css` (Tailwind v4 has no PostCSS config)
- Tailwind classes used for layout and spacing
- Custom classes for doc-specific styling: `.inline-code`, `.inline-highlight`, `.code-block`, `.yaml-block`, `.tip-box`, `.note-box`, `.code-prompt`, `.code-output`, `.code-comment`, `.yaml-copy`, `.code-copy`, `.yaml-line`, `.inline-code--link`
- Theme colors defined in `@theme` directive at the top of `index.css`
- Color palette: magenta (`#ff00ff`), dark bg (`#0a0a0b`, `#1a1a2e`), dim text (`#888`), body text (`#a0a0a0`)

## Build & Commands

All commands run from the `site/` directory:

```bash
bun run dev          # Start Vite dev server with HMR
bun run build        # TypeScript check + Vite build + prerendered routes + metadata + sitemap
bun run lint         # ESLint + stylelint
bun run format       # Prettier --write (all ts/tsx/css)
bun run format:check # Prettier --check (CI)
```

`bun run build` must succeed before committing any site changes. It generates:
- `dist/index.html` — main HTML
- `dist/404.html` — copy of index.html for SPA fallback
- `dist/assets/` — compiled JS + CSS bundles
- Prerendered HTML files for all public routes
- `dist/sitemap.xml`

## Deployment

Deploys automatically via GitHub Actions (`.github/workflows/deploy-site.yml`) on push to `main` when `site/**` files change. The workflow:
1. Checks out repo
2. Installs Bun
3. Runs `bun install`
4. Runs `bun run lint`
5. Runs `bun run format:check`
6. Runs `bun run build`
7. Uploads `site/dist` to GitHub Pages

## Key Rules

1. **Always build after site changes** — run `bun run build` from `site/` before committing
2. **Keep tools.ts as single source of truth** for all doc content — no separate markdown files
3. **Maintain backward compatibility** — doc IDs in `tools.ts` become URL slugs; renaming them breaks existing links
4. **Follow existing patterns** — when adding components, look at ToolGuide.tsx for rendering conventions
5. **Format before committing** — run `bun run format` if you edited any ts/tsx/css files
6. **Never create README.md or standalone documentation files** unless explicitly requested
