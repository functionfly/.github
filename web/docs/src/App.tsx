import { Routes, Route } from 'react-router-dom'
import Layout from './components/layout/Layout'
import Index from './legacy-vite/Index.tsx'
import Function from './legacy-vite/Function.tsx'
import ApiReference from './legacy-vite/ApiReference.tsx'
import Guides from './legacy-vite/Guides.tsx'

function App() {
  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<Index />} />
        <Route path="functions/:author/:name" element={<Function />} />
        <Route path="api-reference" element={<ApiReference />} />
        <Route path="guides" element={<Guides />} />
      </Route>
    </Routes>
  )
}

export default App
