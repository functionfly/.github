/**
 * A2AExplorerPage — A2A playground: send a task, watch lifecycle,
 * view task, view receipt. Uses TaskTimeline for state visualization.
 */

import { useState } from 'react';
import { Play, Loader2, Clock, CheckCircle2, XCircle, Ban } from 'lucide-react';
import { ProtocolBadge, TaskTimeline } from '@/components/common/protocol';
import type { TaskState } from '@/components/common/protocol';

interface TaskResult {
  id: string;
  status: {
    state: TaskState;
    message?: string;
    timestamp: string;
  };
  artifacts?: {
    name: string;
    parts: { type: string; text?: string; data?: unknown }[];
  }[];
}

export function A2AExplorerPage() {
  const [agentId, setAgentId] = useState('');
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<TaskResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [history, setHistory] = useState<TaskResult[]>([]);

  async function handleSendTask(e: React.FormEvent) {
    e.preventDefault();
    if (!agentId.trim() || !message.trim()) return;

    setLoading(true);
    setError(null);

    try {
      const res = await fetch(`/api/v1/a2a/${encodeURIComponent(agentId)}/tasks/send`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          message: {
            role: 'user',
            parts: [{ type: 'text', text: message }],
          },
        }),
      });

      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data?.error?.message || `HTTP ${res.status}`);
      }

      const data: TaskResult = await res.json();
      setResult(data);
      setHistory((prev) => [data, ...prev].slice(0, 20));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  }

  async function pollTask(taskId: string) {
    try {
      const res = await fetch(`/api/v1/a2a/tasks/${taskId}`);
      if (!res.ok) return;
      const data: TaskResult = await res.json();
      setResult(data);
    } catch {
      // ignore polling errors
    }
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-2xl font-bold text-text-primary tracking-tight flex items-center gap-2">
          <Play className="h-6 w-6 text-emerald-500" />
          A2A Explorer
        </h1>
        <p className="text-text-secondary mt-1">
          Send tasks to A2A agents, watch task lifecycle, and inspect results
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Send Task Form */}
        <div className="lg:col-span-2">
          <div className="p-6 bg-card border border-border-subtle rounded-xl">
            <h2 className="text-lg font-semibold text-text-primary mb-4 flex items-center gap-2">
              <ProtocolBadge protocol="a2a" size="md" />
              Send Task
            </h2>

            <form onSubmit={handleSendTask} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-text-primary mb-1.5">
                  Agent ID
                </label>
                <input
                  type="text"
                  value={agentId}
                  onChange={(e) => setAgentId(e.target.value)}
                  placeholder="my-org/my-agent"
                  className="w-full px-3 py-2 bg-bg-secondary border border-border-subtle rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:ring-2 focus:ring-brand-500/30"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-text-primary mb-1.5">
                  Message
                </label>
                <textarea
                  value={message}
                  onChange={(e) => setMessage(e.target.value)}
                  placeholder="What should the agent do?"
                  rows={4}
                  className="w-full px-3 py-2 bg-bg-secondary border border-border-subtle rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:ring-2 focus:ring-brand-500/30"
                />
              </div>

              <button
                type="submit"
                disabled={loading || !agentId.trim() || !message.trim()}
                className="flex items-center gap-2 px-6 py-2.5 bg-emerald-500 text-white rounded-lg font-medium hover:bg-emerald-600 disabled:opacity-50 transition-colors"
              >
                {loading ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  <Play className="w-4 h-4" />
                )}
                Send Task
              </button>
            </form>

            {error && (
              <div className="mt-4 p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400 text-sm">
                {error}
              </div>
            )}
          </div>

          {/* Result */}
          {result && (
            <div className="mt-4 p-6 bg-card border border-border-subtle rounded-xl">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-semibold text-text-primary">Task Result</h2>
                <button
                  onClick={() => pollTask(result.id)}
                  className="px-3 py-1.5 text-xs bg-zinc-700 text-white rounded-lg hover:bg-zinc-600 transition-colors"
                >
                  Refresh
                </button>
              </div>

              <div className="space-y-4">
                <div className="flex items-center gap-3">
                  <span className="text-sm text-text-secondary">Task ID:</span>
                  <code className="text-sm font-mono text-brand-400">{result.id}</code>
                </div>

                <TaskTimeline
                  currentState={result.status.state}
                  className="py-2"
                />

                <div className="flex items-center gap-2">
                  <StateIcon state={result.status.state} />
                  <span className="text-sm font-medium text-text-primary capitalize">
                    {result.status.state}
                  </span>
                  {result.status.message && (
                    <span className="text-sm text-text-secondary">
                      — {result.status.message}
                    </span>
                  )}
                </div>

                {result.artifacts && result.artifacts.length > 0 && (
                  <div>
                    <h3 className="text-sm font-medium text-text-primary mb-2">Artifacts</h3>
                    {result.artifacts.map((artifact, idx) => (
                      <div
                        key={idx}
                        className="p-3 bg-bg-secondary rounded-lg border border-border-subtle"
                      >
                        <p className="text-xs text-text-secondary mb-1">{artifact.name}</p>
                        {artifact.parts.map((part, pidx) => (
                          <pre
                            key={pidx}
                            className="text-sm text-text-primary font-mono whitespace-pre-wrap break-all"
                          >
                            {part.text || JSON.stringify(part.data, null, 2)}
                          </pre>
                        ))}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          )}
        </div>

        {/* History sidebar */}
        <div>
          <div className="p-4 bg-card border border-border-subtle rounded-xl">
            <h3 className="text-sm font-semibold text-text-primary mb-3 flex items-center gap-2">
              <Clock className="w-4 h-4 text-text-secondary" />
              Recent Tasks
            </h3>
            {history.length === 0 ? (
              <p className="text-sm text-text-secondary">No tasks sent yet</p>
            ) : (
              <div className="space-y-2">
                {history.map((task) => (
                  <button
                    key={task.id}
                    onClick={() => setResult(task)}
                    className="w-full text-left p-2.5 bg-bg-secondary rounded-lg hover:bg-zinc-700/50 transition-colors"
                  >
                    <div className="flex items-center justify-between mb-1">
                      <code className="text-xs font-mono text-text-secondary truncate">
                        {task.id.slice(0, 12)}...
                      </code>
                      <StateIcon state={task.status.state} className="w-3.5 h-3.5" />
                    </div>
                    <span className="text-xs text-text-secondary capitalize">
                      {task.status.state}
                    </span>
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function StateIcon({ state, className = 'w-4 h-4' }: { state: TaskState; className?: string }) {
  switch (state) {
    case 'completed':
      return <CheckCircle2 className={`${className} text-emerald-400`} />;
    case 'failed':
      return <XCircle className={`${className} text-red-400`} />;
    case 'canceled':
      return <Ban className={`${className} text-zinc-400`} />;
    case 'working':
      return <Loader2 className={`${className} text-amber-400 animate-spin`} />;
    default:
      return <Clock className={`${className} text-blue-400`} />;
  }
}

export default A2AExplorerPage;
