import { useParams, Link } from 'react-router-dom'
import ToolGuide from '../components/ToolGuide'
import { tools } from '../data/tools'

export default function ToolPage() {
  const { id } = useParams<{ id: string }>()
  const tool = id ? tools[id] : undefined

  if (!tool) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center text-text-dim font-jetbrains">
        <h1 className="text-magenta text-xl mb-4">404</h1>
        <p className="mb-4">Tool not found</p>
        <Link to="/docs" className="text-magenta hover:underline text-base">← Back to Documentation</Link>
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
