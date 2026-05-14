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
      <div className="scanlines" />
      <Navbar variant={variant} title={title} />
      <main className="max-w-4xl mx-auto px-4">
        {children}
      </main>
      <Footer />
    </div>
  )
}
