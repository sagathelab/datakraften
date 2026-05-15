import { mkdirSync, readFileSync, writeFileSync } from 'fs'
import { dirname, join } from 'path'
import { renderToString } from 'react-dom/server'
import { MemoryRouter } from 'react-router-dom'
import App from '../src/App'
import { getPageMeta, getStaticRoutes, renderHeadMarkup } from '../src/seo'

const dist = new URL('../dist', import.meta.url).pathname
const template = readFileSync(join(dist, 'index.html'), 'utf-8')
const routes = getStaticRoutes()

function renderRoute(route: string) {
  return renderToString(
    <MemoryRouter initialEntries={[route]}>
      <App />
    </MemoryRouter>,
  )
}

function outputPath(route: string) {
  if (route === '/') {
    return join(dist, 'index.html')
  }

  return join(dist, route.slice(1), 'index.html')
}

function writeRoute(route: string) {
  const meta = getPageMeta(route)
  const html = template
    .replace('<!--app-head-->', renderHeadMarkup(meta))
    .replace('<div id="root"></div>', `<div id="root">${renderRoute(route)}</div>`)

  const outPath = outputPath(route)
  mkdirSync(dirname(outPath), { recursive: true })
  writeFileSync(outPath, html)
}

for (const route of routes) {
  writeRoute(route)
}

writeFileSync(join(dist, '404.html'), outputPath('/') === join(dist, 'index.html') ? readFileSync(join(dist, 'index.html'), 'utf-8') : template)

const sitemap = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${routes
  .map((route) => {
    const path = route === '/' ? '' : route
    const priority = route === '/' ? '1.0' : route === '/docs' ? '0.9' : route.startsWith('/docs/') ? '0.8' : '0.6'
    return `  <url><loc>https://datakraften.no${path}</loc><priority>${priority}</priority></url>`
  })
  .join('\n')}
</urlset>`

writeFileSync(join(dist, 'sitemap.xml'), sitemap)

console.log(`✓ Prerendered ${routes.length} routes`)
console.log('✓ Generated sitemap.xml')
