import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import Editor from '@monaco-editor/react'
import { executeFunction, sharePlayground } from '../../api/playground'
import type { IOType, PlaygroundResponse } from '../../types/function'
import './Playground.css'

interface PlaygroundProps {
  author: string
  name: string
  inputSchema?: IOType
}

export default function Playground({ author, name, inputSchema }: PlaygroundProps) {
  const defaultInput = inputSchema?.example
    ? JSON.stringify(inputSchema.example, null, 2)
    : '{}'

  const [input, setInput] = useState(defaultInput)
  const [output, setOutput] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const executeMutation = useMutation({
    mutationFn: () => executeFunction(author, name, JSON.parse(input)),
    onSuccess: (data: PlaygroundResponse) => {
      if (data.ok && data.data) {
        setOutput(JSON.stringify(data.data, null, 2))
        setError(null)
      } else if (data.error) {
        setError(data.error.message)
        setOutput(null)
      }
    },
    onError: (err: Error) => {
      setError(err.message)
      setOutput(null)
    },
  })

  const shareMutation = useMutation({
    mutationFn: () => sharePlayground(author, name, JSON.parse(input)),
    onSuccess: (data) => {
      navigator.clipboard.writeText(data.share_url)
      alert('Share URL copied to clipboard!')
    },
    onError: () => {
      alert('Failed to create share URL')
    },
  })

  const handleRun = () => {
    try {
      JSON.parse(input) // Validate JSON
      executeMutation.mutate()
    } catch {
      setError('Invalid JSON input')
    }
  }

  const handleShare = () => {
    try {
      JSON.parse(input)
      shareMutation.mutate()
    } catch {
      setError('Invalid JSON input')
    }
  }

  return (
    <div className="playground">
      <div className="playground-header">
        <h3>Input</h3>
        <div className="playground-actions">
          <button
            className="btn"
            onClick={handleRun}
            disabled={executeMutation.isPending}
          >
            {executeMutation.isPending ? 'Running...' : 'Run'}
          </button>
          <button
            className="btn btn-secondary"
            onClick={handleShare}
            disabled={shareMutation.isPending}
          >
            {shareMutation.isPending ? 'Sharing...' : 'Share'}
          </button>
        </div>
      </div>

      <div className="editor-container">
        <Editor
          height="200px"
          defaultLanguage="json"
          value={input}
          onChange={(value) => setInput(value || '{}')}
          theme="vs-dark"
          options={{
            minimap: { enabled: false },
            lineNumbers: 'on',
            scrollBeyondLastLine: false,
            automaticLayout: true,
            tabSize: 2,
            fontSize: 14,
          }}
        />
      </div>

      {error && (
        <div className="playground-error">
          <h4>Error</h4>
          <pre>{error}</pre>
        </div>
      )}

      {output && (
        <div className="playground-output">
          <h4>Output</h4>
          <div className="output-editor">
            <Editor
              height="200px"
              defaultLanguage="json"
              value={output}
              theme="vs-dark"
              options={{
                readOnly: true,
                minimap: { enabled: false },
                lineNumbers: 'on',
                scrollBeyondLastLine: false,
                automaticLayout: true,
                fontSize: 14,
              }}
            />
          </div>
        </div>
      )}
    </div>
  )
}
