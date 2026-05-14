import { useEffect } from 'react'
import { useLocation } from 'react-router-dom'
import Layout from '../components/Layout'
import DocCard from '../components/DocCard'
import { categories, tools } from '../data/tools'

export default function DocsHub() {
  const location = useLocation()

  useEffect(() => {
    if (location.hash) {
      const el = document.getElementById(location.hash.slice(1))
      if (el) {
        const top = el.getBoundingClientRect().top + window.scrollY - 64
        window.scrollTo({ top, behavior: 'smooth' })
      }
    }
  }, [location])
  return (
    <Layout variant="docs" title="Documentation">
      <div className="doc-content py-8">
        <h1 className="text-3xl font-bold text-magenta font-share-tech mb-2">Documentation</h1>
        <p className="text-base text-text-dim mb-8">
          Everything you need to know about the tools in your bootstrapped and AI-powered development environment.
        </p>

        {categories.map((cat) => (
          <section key={cat.title} id={cat.title.toLowerCase().replace(/[\s&]+/g, '-')} className="mb-8 scroll-mt-16">
            <div className="flex items-baseline gap-3 mb-4">
              <h2 className="text-lg font-bold text-text-primary">{cat.title}</h2>
              <span className="text-xs text-text-dim">{cat.desc}</span>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
              {cat.ids.map((id) => {
                const tool = tools[id]
                if (!tool) return null
                return (
                  <DocCard
                    key={id}
                    to={`/docs/${id}`}
                    cmd={tool.title}
                    desc={tool.subtitle}
                  />
                )
              })}
            </div>
          </section>
        ))}

        <div className="disclaimer mt-12 pt-6 border-t border-fuchsia-500/20">
          <p className="text-sm text-text-dim">
            Datakraften orchestrates existing tools -- it does not replace them.
            Each tool retains its own license, documentation, and update mechanism.
            See the respective official sites for detailed documentation.
          </p>
        </div>
      </div>
    </Layout>
  )
}
