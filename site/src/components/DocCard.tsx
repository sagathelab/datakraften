import { Link } from 'react-router-dom'

interface DocCardProps {
  to: string
  cmd: string
  desc: string
  badge?: string
}

export default function DocCard({ to, cmd, desc, badge }: DocCardProps) {
  return (
    <Link
      to={to}
      className="doc-card block bg-bg-card border border-lime-400/20 rounded-lg p-4 hover:border-magenta hover:bg-lime-400/5 transition-all"
    >
      <div className="flex items-center justify-between gap-2">
        <span className="font-jetbrains text-base text-magenta">{cmd}</span>
        {badge && (
          <span className="text-xs px-1.5 py-0.5 rounded bg-lime-400/10 text-text-dim uppercase tracking-wider">
            {badge}
          </span>
        )}
      </div>
      <p className="mt-2 text-sm text-text-dim leading-relaxed">{desc}</p>
    </Link>
  )
}
