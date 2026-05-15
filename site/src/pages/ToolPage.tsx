import { useParams, Link, useLocation } from 'react-router-dom'
import { useEffect } from 'react'
import ToolGuide from '../components/ToolGuide'
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
      <div className="min-h-screen flex flex-col items-center justify-center text-text-dim font-jetbrains">
        <h1 className="text-magenta text-xl mb-4">404</h1>
        <p className="mb-4">Tool not found</p>
        <Link to="/docs" className="text-magenta hover:underline text-base">
          ← Back to Documentation
        </Link>
      </div>
    )
  }

  return (
    <ToolGuide
      title={tool.title}
      subtitle={tool.subtitle}
      sections={tool.sections}
      website={tool.website}
    />
  )
}
