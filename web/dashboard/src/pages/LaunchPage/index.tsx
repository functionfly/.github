import { useState, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { getApiBaseUrl } from '@/lib/constants';
import { MetaTags } from '@/components/seo/MetaTags';
import './LaunchPage.css';

const INTERESTS = [
  { id: 'serverless', label: 'Serverless functions' },
  { id: 'ai-agents', label: 'AI agents & tooling' },
  { id: 'edge', label: 'Edge deployment' },
  { id: 'early-access', label: 'Early access / beta' },
  { id: 'curious', label: 'Just curious' },
] as const;

export function LaunchPage() {
  const [email, setEmail] = useState('');
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [status, setStatus] = useState<'idle' | 'submitting' | 'success' | 'error'>('idle');
  const [errorMessage, setErrorMessage] = useState('');

  const toggleInterest = useCallback((id: string) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }, []);

  const handleSubmit = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();
      const trimmed = email.trim();
      if (!trimmed) {
        setErrorMessage('Please enter your email.');
        setStatus('error');
        return;
      }
      if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(trimmed)) {
        setErrorMessage('Please enter a valid email address.');
        setStatus('error');
        return;
      }

      setStatus('submitting');
      setErrorMessage('');

      const base = getApiBaseUrl() || (typeof window !== 'undefined' ? window.location.origin : '');
      const formData = new FormData();
      formData.append('feedbackType', 'launch_waitlist');
      formData.append('subject', 'Launch signup');
      formData.append('message', selected.size ? Array.from(selected).join(', ') : 'No interests selected');
      formData.append('email', trimmed);

      try {
        const res = await fetch(`${base}/v1/feedback`, {
          method: 'POST',
          body: formData,
        });
        if (!res.ok) {
          const err = await res.json().catch(() => ({}));
          throw new Error((err as { error?: string }).error || 'Something went wrong.');
        }
        setStatus('success');
      } catch (err) {
        setStatus('error');
        setErrorMessage(err instanceof Error ? err.message : 'Could not sign up. Try again.');
      }
    },
    [email, selected]
  );

  return (
    <div className="launch-page">
      <MetaTags
        title="FunctionFly — Coming Soon | Serverless & AI Agent Infrastructure"
        description="FunctionFly is coming. Join the waitlist for serverless functions and AI agent infrastructure at the edge."
        keywords={['functionfly', 'coming soon', 'serverless', 'AI agents', 'waitlist']}
      />
      <div className="launch-bg" aria-hidden />
      <div className="launch-noise" aria-hidden />

      <div className="relative z-10 flex min-h-screen flex-col items-center justify-center px-4 py-16 sm:px-6">
        <div className="w-full max-w-xl mx-auto text-center">
          {status === 'success' ? (
            <>
              <div className="launch-reveal mb-6">
                <div className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-[var(--launch-accent-dim)] text-[var(--launch-accent)]">
                  <svg className="w-8 h-8" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                </div>
              </div>
              <h1 className="launch-display-font launch-reveal text-3xl sm:text-4xl font-bold tracking-tight text-[var(--launch-text)]">
                You're on the list
              </h1>
              <p className="launch-reveal mt-3 text-lg text-[var(--launch-muted)]">
                We'll notify you as soon as FunctionFly is ready. Until then, build something fun.
              </p>
            </>
          ) : (
            <>
              <div className="launch-reveal launch-float mb-2">
                <span className="inline-block rounded-full border border-[var(--launch-border)] bg-[var(--launch-surface)] px-4 py-1.5 text-sm text-[var(--launch-muted)]">
                  Coming soon
                </span>
              </div>
              <h1 className="launch-display-font launch-reveal text-4xl sm:text-5xl md:text-6xl font-bold tracking-tight text-[var(--launch-text)]">
                FunctionFly
              </h1>
              <p className="launch-reveal mt-4 text-xl sm:text-2xl text-[var(--launch-muted)]">
                Serverless functions & AI agent infrastructure
              </p>
              <p className="launch-reveal mt-6 text-[var(--launch-muted)] max-w-md mx-auto">
                We're putting the finishing touches on the platform. Drop your email and we'll tell you the moment you can ship.
              </p>

              <form onSubmit={handleSubmit} className="launch-reveal mt-10 text-left">
                <div className="rounded-2xl border border-[var(--launch-border)] bg-[var(--launch-surface)] p-6 sm:p-8 backdrop-blur-sm">
                  <label htmlFor="launch-email" className="block text-sm font-medium text-[var(--launch-text)] mb-2">
                    Email
                  </label>
                  <input
                    id="launch-email"
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="you@company.com"
                    disabled={status === 'submitting'}
                    className="launch-input w-full rounded-xl border border-[var(--launch-border)] bg-black/30 px-4 py-3 text-[var(--launch-text)] placeholder-[var(--launch-muted)] transition-colors"
                    autoComplete="email"
                  />

                  <p className="mt-4 text-sm font-medium text-[var(--launch-text)] mb-3">
                    What are you here for? <span className="text-[var(--launch-muted)]">(optional)</span>
                  </p>
                  <div className="space-y-2">
                    {INTERESTS.map(({ id, label }) => (
                      <label
                        key={id}
                        className="launch-checkbox-wrap flex cursor-pointer items-center gap-3 rounded-lg border border-[var(--launch-border)] bg-transparent px-4 py-3"
                      >
                        <input
                          type="checkbox"
                          checked={selected.has(id)}
                          onChange={() => toggleInterest(id)}
                          disabled={status === 'submitting'}
                          className="h-4 w-4 rounded border-[var(--launch-border)] bg-transparent text-[var(--launch-accent)] focus:ring-[var(--launch-accent)]"
                        />
                        <span className="launch-checkbox-label text-sm text-[var(--launch-text)]">{label}</span>
                      </label>
                    ))}
                  </div>

                  {errorMessage && (
                    <p className="mt-4 text-sm text-red-400" role="alert">
                      {errorMessage}
                    </p>
                  )}

                  <button
                    type="submit"
                    disabled={status === 'submitting'}
                    className="launch-cta mt-6 w-full rounded-xl bg-[var(--launch-accent)] py-3.5 font-semibold text-black disabled:opacity-60 disabled:cursor-not-allowed"
                  >
                    {status === 'submitting' ? 'Signing up…' : 'Notify me'}
                  </button>
                </div>
              </form>

              {/* What we're building — visual interest */}
              <div className="mt-16 grid grid-cols-1 sm:grid-cols-3 gap-4 max-w-3xl mx-auto">
                {[
                  { title: 'Ship to the edge', desc: 'Deploy once, run everywhere.' },
                  { title: 'AI-native', desc: 'Agents, tools, and cost controls.' },
                  { title: 'No lock-in', desc: 'Your code, your infra.' },
                ].map((item, i) => (
                  <div
                    key={item.title}
                    className="launch-reveal rounded-xl border border-[var(--launch-border)] bg-[var(--launch-surface)] p-4 text-left backdrop-blur-sm"
                    style={{ animationDelay: `${1.4 + i * 0.15}s` }}
                  >
                    <p className="launch-display-font font-semibold text-[var(--launch-text)]">{item.title}</p>
                    <p className="mt-1 text-sm text-[var(--launch-muted)]">{item.desc}</p>
                  </div>
                ))}
              </div>
            </>
          )}
        </div>

        <footer className="absolute bottom-6 left-0 right-0 text-center">
          <p className="text-sm text-[var(--launch-muted)]">
            © {new Date().getFullYear()} FunctionFly
            <Link to="/privacy" className="ml-4 hover:text-[var(--launch-text)] transition-colors">Privacy</Link>
            <Link to="/terms" className="ml-3 hover:text-[var(--launch-text)] transition-colors">Terms</Link>
          </p>
        </footer>
      </div>
    </div>
  );
}
