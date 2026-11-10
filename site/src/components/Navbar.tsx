import { Link } from 'react-router-dom'

interface NavbarProps {
  variant: 'landing' | 'docs'
  title?: string
}

export default function Navbar({ variant, title }: NavbarProps) {
  if (variant === 'landing') {
    return (
      <nav
        aria-label="Main navigation"
        className="site-nav flex flex-wrap items-center justify-center gap-6 px-4 pt-6 pb-4 font-jetbrains text-base"
      >
        <Link to="/" className="site-wordmark">
          datakraften<span>.</span>
        </Link>
        <Link to="/docs" className="text-text-dim hover:text-magenta transition-colors">
          Docs
        </Link>
        <a
          href="https://github.com/sagathelab/datakraften"
          className="text-text-dim hover:text-magenta"
        >
          GitHub
        </a>
      </nav>
    )
  }

  return (
    <nav
      aria-label="Main navigation"
      className="doc-nav flex flex-wrap items-center justify-between gap-2 px-4 sm:px-6 py-3 font-jetbrains text-base border-b border-lime-400/20"
    >
      <div className="flex flex-wrap items-center gap-x-4 gap-y-2">
        <Link
          to="/"
          className="text-text-dim hover:text-magenta transition-colors inline-flex items-center gap-1"
        >
          <svg
            className="w-4 h-4"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M19 12H5M12 19l-7-7 7-7" />
          </svg>
          Home
        </Link>
        <Link to="/docs" className="text-text-dim hover:text-magenta transition-colors">
          Documentation
        </Link>
      </div>
      {title && title !== 'Documentation' && (
        <span className="text-text-dim text-sm break-words">{title}</span>
      )}
    </nav>
  )
}
