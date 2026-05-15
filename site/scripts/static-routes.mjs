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
  'dk', 'config', 'teams', 'node', 'python', 'dotnet', 'fish', 'starship', 'atuin',
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

const sitemap = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>https://datakraften.no/</loc><priority>1.0</priority></url>
  <url><loc>https://datakraften.no/docs</loc><priority>0.9</priority></url>
${routes.slice(1).map(r => `  <url><loc>https://datakraften.no${r}</loc><priority>0.6</priority></url>`).join('\n')}
${toolIds.map(id => `  <url><loc>https://datakraften.no/docs/${id}</loc><priority>0.8</priority></url>`).join('\n')}
</urlset>`
writeFileSync(join(dist, 'sitemap.xml'), sitemap)

console.log(`✓ Generated static HTML for ${routes.length + toolIds.length} routes`)
console.log('✓ Generated sitemap.xml')
