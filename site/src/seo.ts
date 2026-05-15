import type { ToolDef, ToolFaq } from './data/tools'
import { tools } from './data/tools'

const siteUrl = 'https://datakraften.no'
const defaultImagePath = '/og-image.svg'
const generatedSelector = 'meta[data-dk-head], link[data-dk-head], script[data-dk-head]'

export interface PageMeta {
  title: string
  description: string
  path: string
  type: 'website' | 'article'
  keywords?: string[]
  imagePath?: string
  noindex?: boolean
  structuredData?: Record<string, unknown>[]
}

type BreadcrumbItem = {
  name: string
  path: string
}

function toAbsoluteUrl(path: string) {
  return new URL(path, siteUrl).toString()
}

function stripTrailingSlash(path: string) {
  if (path.length > 1 && path.endsWith('/')) {
    return path.slice(0, -1)
  }
  return path
}

export function normalizePath(path: string) {
  const [pathname] = path.split(/[?#]/)
  return stripTrailingSlash(pathname || '/')
}

function breadcrumbSchema(items: BreadcrumbItem[]) {
  return {
    '@context': 'https://schema.org',
    '@type': 'BreadcrumbList',
    itemListElement: items.map((item, index) => ({
      '@type': 'ListItem',
      position: index + 1,
      name: item.name,
      item: toAbsoluteUrl(item.path),
    })),
  }
}

function toolArticleSchema(tool: ToolDef) {
  return {
    '@context': 'https://schema.org',
    '@type': 'TechArticle',
    headline: tool.title,
    description: tool.summary ?? tool.subtitle,
    url: toAbsoluteUrl(`/docs/${tool.id}`),
    mainEntityOfPage: toAbsoluteUrl(`/docs/${tool.id}`),
    about: {
      '@type': 'SoftwareApplication',
      name: tool.title,
      operatingSystem: tool.supportedPlatforms?.join(', ') || 'WSL, Linux, macOS',
    },
    publisher: {
      '@type': 'Organization',
      name: 'Datakraften',
      url: siteUrl,
      logo: toAbsoluteUrl(defaultImagePath),
    },
  }
}

function faqSchema(faqs: ToolFaq[]) {
  return {
    '@context': 'https://schema.org',
    '@type': 'FAQPage',
    mainEntity: faqs.map((faq) => ({
      '@type': 'Question',
      name: faq.question,
      acceptedAnswer: {
        '@type': 'Answer',
        text: faq.answer,
      },
    })),
  }
}

function landingSchemas() {
  return [
    {
      '@context': 'https://schema.org',
      '@type': 'Organization',
      name: 'Datakraften',
      url: siteUrl,
      logo: toAbsoluteUrl(defaultImagePath),
      sameAs: ['https://github.com/sagathelab/datakraften'],
    },
    {
      '@context': 'https://schema.org',
      '@type': 'SoftwareApplication',
      name: 'Datakraften',
      applicationCategory: 'DeveloperApplication',
      operatingSystem: 'WSL, Linux, macOS',
      softwareVersion: 'current',
      description:
        'Datakraften bootstraps developer workstations for WSL, Linux, and macOS with a YAML-driven CLI.',
      url: siteUrl,
      downloadUrl: toAbsoluteUrl('/install'),
      publisher: {
        '@type': 'Organization',
        name: 'Datakraften',
      },
    },
  ]
}

function docsMeta() {
  return {
    title: 'Datakraften Documentation',
    description:
      'Documentation for Datakraften, dk CLI, supported tools, runtimes, shell setup, editors, and AI tooling.',
    path: '/docs',
    type: 'website' as const,
    keywords: ['Datakraften docs', 'dk CLI docs', 'developer workstation docs', 'WSL setup docs'],
    structuredData: [
      breadcrumbSchema([
        { name: 'Datakraften', path: '/' },
        { name: 'Documentation', path: '/docs' },
      ]),
    ],
  }
}

function legalMeta(path: '/privacy' | '/terms') {
  const isPrivacy = path === '/privacy'

  return {
    title: isPrivacy ? 'Datakraften Privacy Policy' : 'Datakraften Terms of Use',
    description: isPrivacy
      ? 'Read how Datakraften handles privacy, local configuration, and third-party tooling.'
      : 'Read the Datakraften terms of use, license position, and third-party tooling disclaimer.',
    path,
    type: 'website' as const,
    keywords: isPrivacy
      ? ['Datakraften privacy', 'privacy policy']
      : ['Datakraften terms', 'terms of use'],
    structuredData: [
      breadcrumbSchema([
        { name: 'Datakraften', path: '/' },
        { name: isPrivacy ? 'Privacy Policy' : 'Terms of Use', path },
      ]),
    ],
  }
}

function toolMeta(tool: ToolDef): PageMeta {
  const description = tool.summary ?? tool.subtitle
  const keywords = [tool.title, `${tool.title} docs`, 'Datakraften', ...(tool.keywords ?? [])]

  return {
    title: `${tool.title} Documentation — Datakraften`,
    description,
    path: `/docs/${tool.id}`,
    type: 'article',
    keywords,
    structuredData: [
      breadcrumbSchema([
        { name: 'Datakraften', path: '/' },
        { name: 'Documentation', path: '/docs' },
        { name: tool.title, path: `/docs/${tool.id}` },
      ]),
      toolArticleSchema(tool),
      ...(tool.faqs?.length ? [faqSchema(tool.faqs)] : []),
    ],
  }
}

export function getPageMeta(path: string): PageMeta {
  const pathname = normalizePath(path)

  if (pathname === '/') {
    return {
      title: 'Datakraften — Developer Workstation Bootstrap for WSL, Linux, and macOS',
      description:
        'Bootstrap a developer workstation for WSL, Linux, and macOS with Datakraften — a YAML-driven CLI for runtimes, shell setup, editors, cloud tools, and AI tooling.',
      path: '/',
      type: 'website',
      keywords: [
        'developer workstation bootstrap',
        'WSL setup tool',
        'Linux developer environment',
        'AI-ready developer workstation',
        'Datakraften',
      ],
      structuredData: landingSchemas(),
    }
  }

  if (pathname === '/docs') {
    return docsMeta()
  }

  if (pathname === '/privacy' || pathname === '/terms') {
    return legalMeta(pathname)
  }

  if (pathname.startsWith('/docs/')) {
    const tool = tools[pathname.replace('/docs/', '')]
    if (tool) {
      return toolMeta(tool)
    }
  }

  return {
    title: 'Datakraften',
    description: 'Datakraften bootstraps developer workstations with a YAML-driven CLI.',
    path: pathname,
    type: 'website',
    noindex: true,
  }
}

export function getStaticRoutes() {
  return ['/', '/docs', '/privacy', '/terms', ...Object.keys(tools).map((id) => `/docs/${id}`)]
}

export function renderHeadMarkup(meta: PageMeta) {
  const image = toAbsoluteUrl(meta.imagePath ?? defaultImagePath)
  const url = toAbsoluteUrl(meta.path)
  const keywords = meta.keywords?.join(', ')

  const tags = [
    `<title>${escapeHtml(meta.title)}</title>`,
    `<meta data-dk-head name="description" content="${escapeHtml(meta.description)}" />`,
    keywords ? `<meta data-dk-head name="keywords" content="${escapeHtml(keywords)}" />` : '',
    `<meta data-dk-head name="author" content="Datakraften" />`,
    `<meta data-dk-head name="robots" content="${meta.noindex ? 'noindex,nofollow' : 'index,follow'}" />`,
    `<link data-dk-head rel="canonical" href="${url}" />`,
    `<meta data-dk-head property="og:title" content="${escapeHtml(meta.title)}" />`,
    `<meta data-dk-head property="og:description" content="${escapeHtml(meta.description)}" />`,
    `<meta data-dk-head property="og:url" content="${url}" />`,
    `<meta data-dk-head property="og:type" content="${meta.type}" />`,
    `<meta data-dk-head property="og:image" content="${image}" />`,
    `<meta data-dk-head name="twitter:card" content="summary_large_image" />`,
    `<meta data-dk-head name="twitter:title" content="${escapeHtml(meta.title)}" />`,
    `<meta data-dk-head name="twitter:description" content="${escapeHtml(meta.description)}" />`,
    `<meta data-dk-head name="twitter:image" content="${image}" />`,
    ...(meta.structuredData ?? []).map(
      (item) => `<script data-dk-head type="application/ld+json">${JSON.stringify(item)}</script>`,
    ),
  ]

  return tags.filter(Boolean).join('\n    ')
}

function escapeHtml(value: string) {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
}

function setMeta(name: string, content: string, attribute: 'name' | 'property' = 'name') {
  if (typeof document === 'undefined') {
    return
  }

  const selector = `meta[${attribute}="${name}"]`
  let element = document.head.querySelector<HTMLMetaElement>(selector)

  if (!element) {
    element = document.createElement('meta')
    element.setAttribute(attribute, name)
    element.dataset.dkHead = 'true'
    document.head.appendChild(element)
  }

  element.content = content
}

function setCanonical(href: string) {
  if (typeof document === 'undefined') {
    return
  }

  let element = document.head.querySelector<HTMLLinkElement>('link[rel="canonical"]')

  if (!element) {
    element = document.createElement('link')
    element.rel = 'canonical'
    element.dataset.dkHead = 'true'
    document.head.appendChild(element)
  }

  element.href = href
}

export function applyPageMeta(meta: PageMeta) {
  if (typeof document === 'undefined') {
    return
  }

  document.title = meta.title

  const image = toAbsoluteUrl(meta.imagePath ?? defaultImagePath)
  const url = toAbsoluteUrl(meta.path)

  setMeta('description', meta.description)
  if (meta.keywords?.length) {
    setMeta('keywords', meta.keywords.join(', '))
  }
  setMeta('author', 'Datakraften')
  setMeta('robots', meta.noindex ? 'noindex,nofollow' : 'index,follow')
  setCanonical(url)

  setMeta('og:title', meta.title, 'property')
  setMeta('og:description', meta.description, 'property')
  setMeta('og:url', url, 'property')
  setMeta('og:type', meta.type, 'property')
  setMeta('og:image', image, 'property')
  setMeta('twitter:card', 'summary_large_image')
  setMeta('twitter:title', meta.title)
  setMeta('twitter:description', meta.description)
  setMeta('twitter:image', image)

  document.head.querySelectorAll(generatedSelector).forEach((node) => {
    if (node.tagName === 'SCRIPT') {
      node.remove()
    }
  })

  for (const item of meta.structuredData ?? []) {
    const script = document.createElement('script')
    script.type = 'application/ld+json'
    script.dataset.dkHead = 'true'
    script.textContent = JSON.stringify(item)
    document.head.appendChild(script)
  }
}
