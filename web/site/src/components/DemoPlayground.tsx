'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import '../styles/sc-main.css';

type DemoFunction = {
  name: string;
  label: string;
  description: string;
  default_input?: Record<string, string>;
};

type RateLimitInfo = {
  remaining: number;
  limit: number;
  reset_unix: number;
};

type ExecuteResponse = {
  success: boolean;
  output?: unknown;
  error?: string;
  function: string;
  latency_ms: number;
  timestamp: string;
  rate_limit?: RateLimitInfo;
};

const EXAMPLES: Record<string, string> = {
  'web-scraper': JSON.stringify(
    { url: 'https://example.com' },
    null,
    2,
  ),
  'text-summarizer': JSON.stringify(
    {
      text: 'FunctionFly is the platform for AI-accessible functions. You can publish in Python, Go, Rust, JavaScript, or WebAssembly, and let agents discover, call, and trust your tools with full observability and sandboxed execution. The trust layer verifies every execution so agents know which tools are safe to call.',
    },
    null,
    2,
  ),
  'currency-converter': JSON.stringify(
    { amount: 100, from_currency: 'USD', to_currency: 'EUR' },
    null,
    2,
  ),
};

export default function DemoPlayground({ authOrigin }: { authOrigin: string }) {
  const [functions, setFunctions] = useState<DemoFunction[]>([]);
  const [selected, setSelected] = useState<string>('');
  const [input, setInput] = useState<string>('');
  const [response, setResponse] = useState<ExecuteResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [ranOnce, setRanOnce] = useState(false);

  useEffect(() => {
    let cancelled = false;
    fetch('/v1/demo')
      .then((res) => res.json())
      .then((data: { functions: DemoFunction[] }) => {
        if (cancelled) return;
        const list = data.functions || [];
        setFunctions(list);
        if (list.length > 0) {
          const first = list[0];
          setSelected(first.name);
          setInput(EXAMPLES[first.name] || '{}');
        }
      })
      .catch(() => {
        if (!cancelled) {
          setFunctions([
            { name: 'web-scraper', label: 'Web Scraper', description: 'Fetches a URL and returns page title and meta description.', default_input: { url: 'https://example.com' } },
            { name: 'text-summarizer', label: 'Text Summarizer', description: 'Summarizes long text into a short abstract.' },
            { name: 'currency-converter', label: 'Currency Converter', description: 'Converts an amount between two ISO-4217 currencies.' },
          ]);
          setSelected('web-scraper');
          setInput(EXAMPLES['web-scraper']);
        }
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const handleSelect = useCallback((name: string) => {
    setSelected(name);
    setInput(EXAMPLES[name] || '{}');
    setResponse(null);
  }, []);

  const handleExecute = useCallback(async () => {
    setLoading(true);
    setResponse(null);
    try {
      const res = await fetch('/v1/demo/execute', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ function: selected, input: tryParse(input) }),
      });
      const data: ExecuteResponse = await res.json();
      setResponse(data);
      if (data.success) setRanOnce(true);
    } catch (err) {
      setResponse({
        success: false,
        error: err instanceof Error ? err.message : 'Request failed',
        function: selected,
        latency_ms: 0,
        timestamp: new Date().toISOString(),
      });
    } finally {
      setLoading(false);
    }
  }, [selected, input]);

  const remaining = useMemo(() => response?.rate_limit?.remaining ?? null, [response]);
  const limit = useMemo(() => response?.rate_limit?.limit ?? 10, [response]);

  const apiKeyHint = useMemo(() => {
    if (!response?.rate_limit) return null;
    const reset = new Date(response.rate_limit.reset_unix * 1000);
    return (
      <span className="demo-rate-hint">
        Resets in {reset.toLocaleTimeString()} · {remaining}/{limit} left today
      </span>
    );
  }, [response, remaining, limit]);

  return (
    <section className="demo-playground" data-animate>
      <div className="ff-container">
        <div className="demo-header ff-animate-slide-up">
          <h2 className="ff-text-4xl ff-weight-bold ff-tracking-tight ff-font-display">
            Try it right now — no account needed
          </h2>
          <p className="ff-text-lg ff-text-subtle ff-max-w-2xl ff-mx-auto ff-mt-4">
            Pick a demo function, hit Execute, and see a real FunctionFly execution.
            No signup. No credit card. Just results.
          </p>
        </div>

        <div className="demo-panel ff-animate-slide-up ff-delay-100">
          <div className="demo-tabs" role="tablist" aria-label="Demo functions">
            {functions.map((fn) => (
              <button
                key={fn.name}
                type="button"
                role="tab"
                aria-selected={selected === fn.name}
                className={`demo-tab ${selected === fn.name ? 'active' : ''}`}
                onClick={() => handleSelect(fn.name)}
              >
                <span className="demo-tab-label">{fn.label}</span>
                <span className="demo-tab-desc">{fn.description}</span>
              </button>
            ))}
          </div>

          <div className="demo-editor">
            <div className="demo-editor-header">
              <span className="demo-editor-title">input.json</span>
              <span className="demo-editor-hint">Edit the JSON, then run.</span>
            </div>
            <textarea
              className="demo-textarea"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              spellCheck={false}
              aria-label="Function input JSON"
            />
          </div>

          <div className="demo-actions">
            <button
              type="button"
              className="ff-btn ff-btn-primary ff-btn-lg ff-glow"
              onClick={handleExecute}
              disabled={loading}
            >
              {loading ? (
                <>
                  <span className="demo-spinner" aria-hidden="true" />
                  Running…
                </>
              ) : (
                <>Execute Function →</>
              )}
            </button>
            {apiKeyHint}
          </div>

          {response && (
            <div className="demo-result">
              <div className="demo-result-header">
                <span className={`demo-status ${response.success ? 'ok' : 'err'}`}>
                  {response.success ? 'Success' : 'Error'}
                </span>
                <span className="demo-meta">{response.latency_ms} ms</span>
                <span className="demo-meta">{response.function}</span>
              </div>
              <pre className="demo-result-body">
                <code>
                  {response.success
                    ? JSON.stringify(response.output, null, 2)
                    : response.error}
                </code>
              </pre>
            </div>
          )}

          {ranOnce && (
            <div className="demo-cta ff-animate-slide-up">
              <p className="demo-cta-text">
                You just ran a FunctionFly function.
                Build your own in 2 minutes →
              </p>
              <a
                href={`${authOrigin}/signup`}
                className="ff-btn ff-btn-primary ff-btn-lg"
              >
                Deploy your own
              </a>
            </div>
          )}
        </div>
      </div>
    </section>
  );
}

function tryParse(raw: string): Record<string, unknown> {
  try {
    return JSON.parse(raw) as Record<string, unknown>;
  } catch {
    return { data: raw };
  }
}
