import { useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getFunctionDocs } from '../api/functions'
import TrustIndicator from '../components/function/TrustIndicator'
import SchemaViewer from '../components/function/SchemaViewer.tsx'
import ExampleBlock from '../components/function/ExampleBlock'
import Playground from '../components/playground/Playground'
import './Function.css'

export default function Function() {
  const { author, name } = useParams<{ author: string; name: string }>()

  const { data: docs, isLoading, error } = useQuery({
    queryKey: ['function', author, name],
    queryFn: () => getFunctionDocs(author!, name!),
    enabled: !!author && !!name,
  })

  if (isLoading) {
    return (
      <div className="container">
        <div className="loading">Loading function...</div>
      </div>
    )
  }

  if (error || !docs) {
    return (
      <div className="container">
        <div className="error">Failed to load function. Please try again later.</div>
      </div>
    )
  }

  const { function: fn, manifest, trust_score, success_rate, avg_latency_ms, examples, capabilities } = docs

  return (
    <div className="container">
      <div className="function-header">
        <div className="function-header-info">
          <h1>{fn.title || fn.name}</h1>
          <p className="function-author">@{fn.author}/{fn.name}</p>
          <p className="function-description">{fn.description || manifest.description}</p>
        </div>
        <TrustIndicator
          score={trust_score}
          successRate={success_rate}
          latency={avg_latency_ms}
        />
      </div>

      <div className="function-content">
        <section className="section">
          <h2>Input Schema</h2>
          {manifest.input ? (
            <SchemaViewer ioType={manifest.input} />
          ) : (
            <p className="no-schema">No input required</p>
          )}
        </section>

        <section className="section">
          <h2>Output Schema</h2>
          {manifest.output ? (
            <SchemaViewer ioType={manifest.output} />
          ) : (
            <p className="no-schema">No output defined</p>
          )}
        </section>

        {capabilities && capabilities.length > 0 && (
          <section className="section">
            <h2>Capabilities</h2>
            <div className="capabilities">
              {capabilities.map((cap) => (
                <span key={cap} className="capability-badge">{cap}</span>
              ))}
            </div>
          </section>
        )}

        {examples && examples.length > 0 && (
          <section className="section">
            <h2>Example Executions</h2>
            <div className="examples">
              {examples.map((example, idx) => (
                <ExampleBlock key={idx} example={example} />
              ))}
            </div>
          </section>
        )}

        <section className="section">
          <h2>Try It</h2>
          <Playground
            author={fn.author}
            name={fn.name}
            inputSchema={manifest.input}
          />
        </section>
      </div>
    </div>
  )
}
