import { useEffect } from 'react'
import { Routes, Route, useLocation } from 'react-router-dom'
import Landing from './pages/Landing'
import DocsHub from './pages/DocsHub'
import ToolPage from './pages/ToolPage'
import Privacy from './pages/Privacy'
import Terms from './pages/Terms'
import { applyPageMeta, getPageMeta } from './seo'

function ScrollToTop() {
  const { pathname, hash } = useLocation()

  useEffect(() => {
    if (!hash) {
      window.scrollTo({ top: 0, behavior: 'instant' })
    }
  }, [pathname, hash])

  return null
}

function RouteHead() {
  const { pathname } = useLocation()

  useEffect(() => {
    applyPageMeta(getPageMeta(pathname))
  }, [pathname])

  return null
}

export default function App() {
  return (
    <>
      <ScrollToTop />
      <RouteHead />
      <Routes>
        <Route path="/" element={<Landing />} />
        <Route path="/docs" element={<DocsHub />} />
        <Route path="/docs/:id" element={<ToolPage />} />
        <Route path="/privacy" element={<Privacy />} />
        <Route path="/terms" element={<Terms />} />
      </Routes>
    </>
  )
}
