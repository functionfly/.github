import { useState, useCallback } from 'react'
import Editor from '@monaco-editor/react'
import { useTheme } from '../../components/common/ThemeProvider'
import { Button } from '../../components/ui/button'
import { Badge } from '../../components/ui/badge'
import { LoadingSpinner } from '../../components/ui/loading-spinner'
import { playgroundAPI } from '../../api/playground'

const DEFAULT_CODE = `def handler(event):
    """
    FunctionFly Playground — Edit and click ▶ Run!
    """
    name = event.get("name", "World")
    items = event.get("items", [1, 2, 3])
    return {
        "message": f"Hello, {name}!",
        "doubled": [x * 2 for x in items],
        "count": len(items)
    }
`

const DEFAULT_INPUT = `{
  "name": "FunctionFly",
  "items": [1, 2, 3, 4, 5]
}`

type Status = 'idle' | 'running' | 'success' | 'error'

interface Result {
  output?: unknown
  error?: string
  durationMs?: number
  logs?: string[]
}

const RUNTIMES = [
  { value: 'python3.12', label: 'Python 3.12' },
  { value: 'python3.11', label: 'Python 3.11' },
  { value: 'node20', label: 'Node.js 20' },
  { value: 'deno', label: 'Deno' },
]

export function StandalonePlaygroundPage() {
  const { theme } = useTheme()
  const [code, setCode] = useState(DEFAULT_CODE)
  const [input, setInput] = useState(DEFAULT_INPUT)
  const [runtime, setRuntime] = useState('python3.12')
  const [status, setStatus] = useState<Status>('idle')
  const [result, setResult] = useState<Result | null>(null)
  const [tab, setTab] = useState<'output' | 'logs'>('output')

  const handleRun = useCallback(async () => {
    setStatus('running')
    setResult(null)
    try {
      let parsedInput: unknown
      try { parsedInput = JSON.parse(input) } catch {
        setStatus('error')
        setResult({ error: 'Invalid JSON in input panel' })
        return
      }
      const resp = await playgroundAPI.execute({ code, runtime, input: parsedInput })
      setStatus('success')
      setResult({ output: resp.output, durationMs: resp.duration_ms, logs: resp.logs })
    } catch (err: unknown) {
      setStatus('error')
      setResult({ error: err instanceof Error ? err.message : 'Execution failed' })
    }
  }, [code, input, runtime])

  const monacoTheme = theme === 'dark' ? 'vs-dark' : 'light'

  return (
    <div className="flex flex-col h-screen bg-background">
      <div className="flex items-center justify-between px-4 py-3 border-b border-border bg-card">
        <div className="flex items-center gap-3">
          <h1 className="text-lg font-semibold text-foreground">Playground</h1>
          <Badge variant="secondary" className="text-xs">Live Execution</Badge>
        </div>
        <div className="flex items-center gap-2">
          <select value={runtime} onChange={e => setRuntime(e.target.value)}
            className="text-sm border border-border rounded-md px-2 py-1 bg-background text-foreground">
            {RUNTIMES.map(r => <option key={r.value} value={r.value}>{r.label}</option>)}
          </select>
          <Button variant="outline" size="sm" onClick={() => { setResult(null); setStatus('idle') }}
            disabled={status === 'running'} className="text-xs">Clear</Button>
          <Button size="sm" onClick={handleRun} disabled={status === 'running'} className="min-w-[80px]">
            {status === 'running'
              ? <span className="flex items-center gap-1"><LoadingSpinner className="w-3 h-3" />Running</span>
              : '▶ Run'}
          </Button>
        </div>
      </div>

      <div className="flex flex-1 overflow-hidden">
        <div className="flex flex-col flex-1 border-r border-border min-w-0">
          <div className="px-3 py-2 border-b border-border bg-muted/30">
            <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">Code</span>
          </div>
          <div className="flex-1 overflow-hidden">
            <Editor height="100%" language={runtime.startsWith('python') ? 'python' : 'javascript'}
              value={code} onChange={v => setCode(v ?? '')} theme={monacoTheme}
              options={{ minimap: { enabled: false }, fontSize: 13, lineNumbers: 'on',
                scrollBeyondLastLine: false, wordWrap: 'on', tabSize: 4, automaticLayout: true,
                padding: { top: 8, bottom: 8 } }} />
          </div>
        </div>

        <div className="flex flex-col w-80 xl:w-96 min-w-0">
          <div className="flex flex-col flex-1 border-b border-border">
            <div className="px-3 py-2 border-b border-border bg-muted/30">
              <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">Input (JSON)</span>
            </div>
            <div className="flex-1 overflow-hidden">
              <Editor height="100%" language="json" value={input} onChange={v => setInput(v ?? '')}
                theme={monacoTheme}
                options={{ minimap: { enabled: false }, fontSize: 12, lineNumbers: 'off',
                  scrollBeyondLastLine: false, wordWrap: 'on', automaticLayout: true,
                  padding: { top: 8, bottom: 8 } }} />
            </div>
          </div>

          <div className="flex flex-col flex-1">
            <div className="flex items-center justify-between px-3 py-2 border-b border-border bg-muted/30">
              <div className="flex items-center gap-2">
                {(['output', 'logs'] as const).map(t => (
                  <button key={t} onClick={() => setTab(t)}
                    className={`text-xs font-medium uppercase tracking-wide transition-colors ${tab === t ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'}`}>
                    {t}
                  </button>
                ))}
              </div>
              <div className="flex items-center gap-2">
                {result?.durationMs !== undefined && (
                  <span className="text-xs text-muted-foreground">{result.durationMs}ms</span>
                )}
                {status === 'success' && <Badge variant="default" className="text-xs bg-green-500/10 text-green-600 border-green-500/20">OK</Badge>}
                {status === 'error' && <Badge variant="destructive" className="text-xs">Error</Badge>}
              </div>
            </div>
            <div className="flex-1 overflow-auto p-3 font-mono text-xs">
              {status === 'idle' && <p className="text-muted-foreground text-center mt-8">Click ▶ Run to execute</p>}
              {status === 'running' && (
                <div className="flex items-center justify-center mt-8 gap-2 text-muted-foreground">
                  <LoadingSpinner className="w-4 h-4" /><span>Executing...</span>
                </div>
              )}
              {status !== 'idle' && status !== 'running' && result && (
                tab === 'output' ? (
                  result.error
                    ? <pre className="text-red-500 whitespace-pre-wrap break-words">{result.error}</pre>
                    : <pre className="text-foreground whitespace-pre-wrap break-words">{JSON.stringify(result.output, null, 2)}</pre>
                ) : (
                  result.logs?.length
                    ? result.logs.map((l, i) => <div key={i} className="text-muted-foreground mb-1">{l}</div>)
                    : <p className="text-muted-foreground">No logs</p>
                )
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default StandalonePlaygroundPage
