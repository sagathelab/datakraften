import type { ReactNode } from 'react'
import Navbar from './Navbar'
import Footer from './Footer'

interface LayoutProps {
  children: ReactNode
  variant?: 'landing' | 'docs'
  title?: string
}

export default function Layout({ children, variant = 'landing', title }: LayoutProps) {
  return (
    <div className="min-h-screen text-text-primary font-jetbrains">
      <a href="#main-content" className="skip-link">
        Skip to content
      </a>
      <div className="scanlines" />
      <Navbar variant={variant} title={title} />
      <main
        id="main-content"
        className={`mx-auto px-4 ${variant === 'docs' ? 'max-w-6xl' : 'max-w-4xl'}`}
      >
        {children}
      </main>
      <Footer />
    </div>
  )
}
