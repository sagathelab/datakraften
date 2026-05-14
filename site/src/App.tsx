import { Routes, Route } from 'react-router-dom'
import Landing from './pages/Landing'
import DocsHub from './pages/DocsHub'
import ToolPage from './pages/ToolPage'
import Privacy from './pages/Privacy'
import Terms from './pages/Terms'

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<Landing />} />
      <Route path="/docs" element={<DocsHub />} />
      <Route path="/docs/:id" element={<ToolPage />} />
      <Route path="/privacy" element={<Privacy />} />
      <Route path="/terms" element={<Terms />} />
    </Routes>
  )
}
