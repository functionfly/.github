import './TrustIndicator.css'

interface TrustIndicatorProps {
  score: number
  successRate: number
  latency: number
}

export default function TrustIndicator({ score, successRate, latency }: TrustIndicatorProps) {
  const getScoreColor = () => {
    if (score >= 90) return 'var(--color-success)'
    if (score >= 70) return 'var(--color-warning)'
    return 'var(--color-error)'
  }

  const getScoreLabel = () => {
    if (score >= 90) return 'Highly Reliable'
    if (score >= 70) return 'Moderate'
    return 'Low Reliability'
  }

  return (
    <div className="trust-indicator">
      <div className="trust-score" style={{ backgroundColor: getScoreColor() }}>
        <span className="score-value">{score.toFixed(0)}</span>
        <span className="score-label">{getScoreLabel()}</span>
      </div>
      <div className="trust-metrics">
        <div className="metric">
          <span className="metric-label">Success Rate</span>
          <span className="metric-value">{(successRate * 100).toFixed(1)}%</span>
        </div>
        <div className="metric">
          <span className="metric-label">Avg. Latency</span>
          <span className="metric-value">{latency}ms</span>
        </div>
      </div>
    </div>
  )
}
