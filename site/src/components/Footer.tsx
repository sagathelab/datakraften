import { Link } from 'react-router-dom'

export default function Footer() {
  return (
    <footer className="footer border-t border-fuchsia-500/20 px-6 py-8 mt-16">
      <div className="max-w-4xl mx-auto flex flex-col items-center gap-4 text-sm text-text-dim">
        <a
          href="https://github.com/sagathelab/datakraften"
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-center gap-2 hover:text-magenta transition-colors"
        >
          <svg width="20" height="20" viewBox="0 0 16 16" fill="currentColor">
            <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27s1.36.09 2 .27c1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0 0 16 8c0-4.42-3.58-8-8-8"/>
          </svg>
          sagathelab/datakraften
        </a>
        <div className="flex gap-4 flex-wrap justify-center">
          <a href="https://github.com/sagathelab/datakraften/blob/main/LICENSE" target="_blank" rel="noopener noreferrer" className="hover:text-magenta transition-colors">License</a>
          <Link to="/docs" className="hover:text-magenta transition-colors">Documentation</Link>
          <a href="https://github.com/sagathelab/datakraften/blob/main/CONTRIBUTING.md" target="_blank" rel="noopener noreferrer" className="hover:text-magenta transition-colors">Contributing</a>
          <a href="https://github.com/sagathelab/datakraften/blob/main/SECURITY.md" target="_blank" rel="noopener noreferrer" className="hover:text-magenta transition-colors">Security</a>
          <Link to="/privacy" className="hover:text-magenta transition-colors">Privacy</Link>
          <Link to="/terms" className="hover:text-magenta transition-colors">Terms</Link>
        </div>
        <p className="text-xs text-text-dim">&copy; 2026 Datakraften &mdash; open source (Apache 2.0)</p>
      </div>
    </footer>
  )
}
