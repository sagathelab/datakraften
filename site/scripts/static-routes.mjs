import { readFileSync, mkdirSync, writeFileSync } from 'fs'
import { join, dirname } from 'path'

const dist = new URL('../dist', import.meta.url).pathname
const src = readFileSync(join(dist, 'index.html'), 'utf-8')

const routes = [
  '/docs',
  '/privacy',
  '/terms',
]

const toolIds = [
  'dk', 'node', 'python', 'dotnet', 'fish', 'starship', 'atuin',
  'fzf', 'fd', 'broot', 'btm', 'brew', 'gh', 'gh-copilot',
  'az', 'docker', 'codex', 'opencode', 'vscode', 'zed', 'pwsh',
]

for (const route of routes) {
  const outPath = join(dist, route.slice(1), 'index.html')
  mkdirSync(dirname(outPath), { recursive: true })
  writeFileSync(outPath, src)
}

for (const id of toolIds) {
  const outPath = join(dist, 'docs', id, 'index.html')
  mkdirSync(dirname(outPath), { recursive: true })
  writeFileSync(outPath, src)
}

console.log(`✓ Generated static HTML for ${routes.length + toolIds.length} routes`)
