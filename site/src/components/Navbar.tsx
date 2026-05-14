import { Link } from 'react-router-dom'

interface NavbarProps {
  variant: 'landing' | 'docs'
  title?: string
}

export default function Navbar({ variant, title }: NavbarProps) {
  if (variant === 'landing') {
    return (
      <nav className="flex justify-center gap-8 pt-6 pb-4 font-jetbrains text-sm">
        <span className="text-magenta">Home</span>
        <Link to="/docs" className="text-text-dim hover:text-magenta transition-colors">Docs</Link>
      </nav>
    )
  }

  return (
    <nav className="doc-nav flex items-center justify-between px-6 py-3 font-jetbrains text-sm border-b border-fuchsia-500/20">
      <div className="flex items-center gap-4">
        <Link to="/" className="text-text-dim hover:text-magenta transition-colors">← Home</Link>
        <Link to="/docs" className="text-text-dim hover:text-magenta transition-colors">Documentation</Link>
      </div>
      {title && <span className="text-text-dim text-xs">{title}</span>}
    </nav>
  )
}
