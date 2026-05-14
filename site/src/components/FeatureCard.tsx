import { Link } from 'react-router-dom'

interface FeatureCardProps {
  title: string
  packages: string
  tagline: string
  to: string
}

export default function FeatureCard({ title, packages, tagline, to }: FeatureCardProps) {
  return (
    <Link
      to={to}
      className="block bg-bg-card border border-fuchsia-500/20 rounded-lg p-4 hover:border-magenta hover:bg-fuchsia-500/5 transition-all"
    >
      <h3 className="font-jetbrains text-base text-magenta mb-2">{title}</h3>
      <p className="text-sm text-text-primary leading-relaxed mb-1">{packages}</p>
      <p className="text-sm text-text-dim leading-relaxed">{tagline}</p>
    </Link>
  )
}
