import { useEffect, useState } from 'react'
import { useLocation } from 'react-router-dom'
import Layout from '../components/Layout'
import DocCard from '../components/DocCard'
import { categories, tools } from '../data/tools'

export default function DocsHub() {
  const location = useLocation()
  const [query, setQuery] = useState('')

  useEffect(() => {
    if (location.hash) {
      const el = document.getElementById(location.hash.slice(1))
      if (el) {
        const top = el.getBoundingClientRect().top + window.scrollY - 64
        window.scrollTo({ top, behavior: 'smooth' })
      }
    }
  }, [location])

  const filtered = query.trim()
    ? categories
        .map((cat) => ({
          ...cat,
          ids: cat.ids.filter((id) => {
            const t = tools[id]
            return (
              t &&
              (t.title.toLowerCase().includes(query.toLowerCase()) ||
                t.subtitle.toLowerCase().includes(query.toLowerCase()))
            )
          }),
        }))
        .filter((cat) => cat.ids.length > 0)
    : categories

  return (
    <Layout variant="docs" title="Documentation">
      <div className="doc-content py-8">
        <h1 className="text-3xl font-bold text-magenta font-share-tech mb-2">Documentation</h1>
        <p className="text-base text-text-dim mb-6">
          Everything you need to know about the tools in your bootstrapped and AI-powered
          development environment.
        </p>

        <div className="relative mb-8">
          <span className="absolute left-3 top-1/2 -translate-y-1/2 text-text-dim text-sm pointer-events-none">
            ~
          </span>
          <input
            type="text"
            placeholder="Search documentation..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="w-full bg-bg-card border border-fuchsia-500/20 rounded-lg py-2.5 pl-8 pr-4 text-sm text-text-primary placeholder:text-text-dim/50 focus:outline-none focus:border-magenta transition-colors font-jetbrains"
          />
        </div>

        {filtered.map((cat) => (
          <section
            key={cat.title}
            id={cat.title.toLowerCase().replace(/[\s&]+/g, '-')}
            className="mb-8 scroll-mt-16"
          >
            <div className="flex items-baseline gap-3 mb-4">
              <h2 className="text-xl font-bold text-text-primary">{cat.title}</h2>
              <span className="text-xs text-text-dim">{cat.desc}</span>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
              {cat.ids.map((id) => {
                const tool = tools[id]
                if (!tool) return null
                return <DocCard key={id} to={`/docs/${id}`} cmd={tool.title} desc={tool.subtitle} />
              })}
            </div>
          </section>
        ))}

        {filtered.length === 0 && (
          <p className="text-text-dim text-sm text-center py-8">No results for "{query}"</p>
        )}

        <div className="disclaimer mt-12 pt-6 border-t border-fuchsia-500/20">
          <p className="text-sm text-text-dim">
            Datakraften orchestrates existing tools -- it does not replace them. Each tool retains
            its own license, documentation, and update mechanism. See the respective official sites
            for detailed documentation.
          </p>
        </div>
      </div>
    </Layout>
  )
}
