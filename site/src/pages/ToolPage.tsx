import { useParams, Link, useLocation } from 'react-router-dom'
import { useEffect } from 'react'
import ToolGuide from '../components/ToolGuide'
import Layout from '../components/Layout'
import { tools } from '../data/tools'

export default function ToolPage() {
  const { id } = useParams<{ id: string }>()
  const location = useLocation()
  const tool = id ? tools[id] : undefined

  useEffect(() => {
    if (location.hash) {
      const el = document.getElementById(location.hash.slice(1))
      if (el) {
        const top = el.getBoundingClientRect().top + window.scrollY - 64
        window.scrollTo({ top, behavior: 'smooth' })
      }
    }
  }, [location])

  if (!tool) {
    return (
      <Layout variant="docs" title="Not found">
        <div className="legal-page py-16 text-text-dim">
          <p className="product-eyebrow">404 / NOT FOUND</p>
          <h1 className="text-magenta text-3xl font-share-tech my-4">Tool not found</h1>
          <p className="mb-6">There is no documentation for this tool.</p>
          <Link to="/docs" className="text-magenta underline underline-offset-4">
            ← Back to documentation
          </Link>
        </div>
      </Layout>
    )
  }

  return (
    <ToolGuide
      title={tool.title}
      subtitle={tool.subtitle}
      sections={tool.sections}
      website={tool.website}
      summary={tool.summary}
      supportedPlatforms={tool.supportedPlatforms}
      commonTasks={tool.commonTasks}
      troubleshooting={tool.troubleshooting}
      faqs={tool.faqs}
    />
  )
}
