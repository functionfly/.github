import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { getCategories } from '../api/functions'
import type { FunctionDocSummary } from '../types/function'
import './Index.css'

export default function Index() {
  const { data: categories, isLoading, error } = useQuery({
    queryKey: ['categories'],
    queryFn: getCategories,
  })

  if (isLoading) {
    return (
      <div className="container">
        <div className="loading">Loading functions...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="container">
        <div className="error">Failed to load functions. Please try again later.</div>
      </div>
    )
  }

  if (!categories || Object.keys(categories).length === 0) {
    return (
      <div className="container">
        <div className="empty-state">
          <h2>No functions yet</h2>
          <p>There are no public functions in the registry.</p>
        </div>
      </div>
    )
  }

  return (
    <div className="container">
      <div className="page-header">
        <h1>Function Registry</h1>
        <p>Discover and explore serverless functions</p>
      </div>

      <div className="categories">
        {Object.entries(categories).map(([category, functions]) => (
          <section key={category} className="category-section">
            <h2 className="category-title">{category}</h2>
            <div className="functions-grid">
              {functions.map((fn) => (
                <FunctionCard key={`${fn.author}/${fn.name}`} fn={fn} />
              ))}
            </div>
          </section>
        ))}
      </div>
    </div>
  )
}

function FunctionCard({ fn }: { fn: FunctionDocSummary }) {
  return (
    <Link to={`/functions/${fn.author}/${fn.name}`} className="function-card">
      <div className="function-card-header">
        <h3 className="function-title">{fn.title || fn.name}</h3>
        <TrustBadge score={fn.trust_score} />
      </div>
      <p className="function-description">{fn.description || 'No description'}</p>
      <div className="function-meta">
        <span className="function-author">@{fn.author}</span>
        <span className="function-version">v{fn.version}</span>
      </div>
    </Link>
  )
}

function TrustBadge({ score }: { score: number }) {
  const getColor = () => {
    if (score >= 90) return 'var(--color-success)'
    if (score >= 70) return 'var(--color-warning)'
    return 'var(--color-error)'
  }

  return (
    <span className="trust-badge" style={{ backgroundColor: getColor() }}>
      {score.toFixed(0)}%
    </span>
  )
}
