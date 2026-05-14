import { Link } from 'react-router-dom'

interface FeatureCardProps {
  title: string
  desc: string
  to: string
}

export default function FeatureCard({ title, desc, to }: FeatureCardProps) {
  return (
    <Link
      to={to}
      className="block bg-bg-card border border-fuchsia-500/20 rounded-lg p-4 hover:border-magenta hover:bg-fuchsia-500/5 transition-all"
    >
      <h3 className="font-jetbrains text-sm text-magenta mb-1">{title}</h3>
      <p className="text-xs text-text-dim leading-relaxed">{desc}</p>
    </Link>
  )
}
