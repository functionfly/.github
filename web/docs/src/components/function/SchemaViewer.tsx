import type { IOType } from '../../types/function'
import './SchemaViewer.css'

interface SchemaViewerProps {
  ioType: IOType
}

export default function SchemaViewer({ ioType }: SchemaViewerProps) {
  const exampleJson = ioType.example
    ? JSON.stringify(ioType.example, null, 2)
    : '{}'

  const schemaJson = ioType.schema
    ? JSON.stringify(ioType.schema, null, 2)
    : JSON.stringify({ type: ioType.type }, null, 2)

  return (
    <div className="schema-viewer">
      <div className="schema-section">
        <h4>Type</h4>
        <code className="type-badge">{ioType.type}</code>
        {ioType.required && <span className="required-badge">Required</span>}
      </div>

      <div className="schema-section">
        <h4>Schema</h4>
        <pre className="code-block">{schemaJson}</pre>
      </div>

      {ioType.example !== undefined && (
        <div className="schema-section">
          <h4>Example</h4>
          <pre className="code-block example">{exampleJson}</pre>
        </div>
      )}
    </div>
  )
}
