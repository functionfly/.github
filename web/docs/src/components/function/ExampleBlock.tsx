import type { ExecutionExample } from '../../types/function'
import './ExampleBlock.css'

interface ExampleBlockProps {
  example: ExecutionExample
}

export default function ExampleBlock({ example }: ExampleBlockProps) {
  const inputJson = JSON.stringify(example.input, null, 2)
  const outputJson = JSON.stringify(example.output, null, 2)

  return (
    <div className="example-block">
      <div className="example-header">
        <span className="example-label">Input</span>
        <span className="example-meta">
          {example.cached && <span className="cached-badge">Cached</span>}
          <span className="duration">{example.duration_ms}ms</span>
        </span>
      </div>
      <pre className="example-code">{inputJson}</pre>

      <div className="example-header">
        <span className="example-label">Output</span>
      </div>
      <pre className="example-code output">{outputJson}</pre>
    </div>
  )
}
