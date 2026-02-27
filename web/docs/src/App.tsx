import { Routes, Route } from 'react-router-dom'
import Layout from './components/layout/Layout'
import Index from './pages/Index.tsx'
import Function from './pages/Function.tsx'

function App() {
  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<Index />} />
        <Route path="functions/:author/:name" element={<Function />} />
      </Route>
    </Routes>
  )
}

export default App
