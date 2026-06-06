/**
 * Admin Content Page
 *
 * Run server-side content generation jobs (changelog summaries, blog
 * drafts, author bios, category copy) and inspect their results. The
 * page surfaces the actual job output to the operator rather than just
 * firing-and-forgetting, so it's clear whether the job succeeded.
 */
import { useState } from 'react';
import { useMutation, type UseMutationResult } from '@tanstack/react-query';
import {
  Sparkles,
  CheckCircle2,
  AlertTriangle,
  Play,
  RotateCw,
  Eye,
} from 'lucide-react';
import { adminApiClient } from '@/lib/api/adminClient';
import { useToast } from '@/components/ui/Toast';

type JobKey = 'changelog' | 'blog' | 'author' | 'category';

interface JobMeta {
  title: string;
  description: string;
  endpoint: string;
}

const JOB_META: Record<JobKey, JobMeta> = {
  changelog: {
    title: 'Generate Changelog',
    description:
      'Compile recent commits and closed PRs into a customer-facing release note. Output is written to the marketing site draft folder.',
    endpoint: '/content/generate/changelog',
  },
  blog: {
    title: 'Generate Blog Draft',
    description:
      'Draft a new blog post from a recent feature flag flip or launch. Uses the function-announcement template by default.',
    endpoint: '/content/generate/blog',
  },
  author: {
    title: 'Generate Author Bio',
    description:
      'Refresh author bios from the team directory. Useful after a new hire joins or someone updates their public profile.',
    endpoint: '/content/generate/author',
  },
  category: {
    title: 'Generate Category Copy',
    description:
      'Regenerate category landing-page copy from the latest product taxonomy. No content is published without review.',
    endpoint: '/content/generate/category',
  },
};

interface JobResult {
  job: JobKey;
  title: string;
  status: 'success' | 'error';
  message: string;
  output?: string;
  finishedAt: number;
}

type JobMutation = UseMutationResult<{ output: string; message: string }, Error, void>;

export function AdminContentPage() {
  const { showToast } = useToast();
  const [history, setHistory] = useState<JobResult[]>([]);
  const [openResult, setOpenResult] = useState<JobResult | null>(null);

  // Build a per-job mutation. Each mutation hook is registered exactly
  // once per render at the top level (no Rules of Hooks violation).
  const useJobMutation = (job: JobKey): JobMutation =>
    useMutation<{ output: string; message: string }, Error, void>({
      mutationFn: async () => {
        const res = (await adminApiClient.post(JOB_META[job].endpoint, {})) as unknown as {
          output?: string;
          message?: string;
          data?: { output?: string; message?: string };
        };
        return {
          output: res?.output ?? res?.data?.output ?? '',
          message: res?.message ?? res?.data?.message ?? '',
        };
      },
      onSuccess: (data) => {
        showToast({ type: 'success', title: `${JOB_META[job].title} complete` });
        const result: JobResult = {
          job,
          title: JOB_META[job].title,
          status: 'success',
          message: data.message || 'Job finished successfully.',
          output: data.output,
          finishedAt: Date.now(),
        };
        setHistory((h) => [result, ...h].slice(0, 20));
        setOpenResult(result);
      },
      onError: (err) => {
        const message = err.message || 'Unknown error';
        showToast({ type: 'error', title: `${JOB_META[job].title} failed: ${message}` });
        const result: JobResult = {
          job,
          title: JOB_META[job].title,
          status: 'error',
          message,
          finishedAt: Date.now(),
        };
        setHistory((h) => [result, ...h].slice(0, 20));
        setOpenResult(result);
      },
    });

  return (
    <ContentPageBody
      useJobMutation={useJobMutation}
      history={history}
      setHistory={setHistory}
      openResult={openResult}
      setOpenResult={setOpenResult}
    />
  );
}

interface ContentPageBodyProps {
  useJobMutation: (job: JobKey) => JobMutation;
  history: JobResult[];
  setHistory: React.Dispatch<React.SetStateAction<JobResult[]>>;
  openResult: JobResult | null;
  setOpenResult: (r: JobResult | null) => void;
}

function ContentPageBody({ useJobMutation, history, setHistory, openResult, setOpenResult }: ContentPageBodyProps) {
  const changelog = useJobMutation('changelog');
  const blog = useJobMutation('blog');
  const author = useJobMutation('author');
  const category = useJobMutation('category');
  const mutations = { changelog, blog, author, category } as const;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
          <Sparkles className="w-7 h-7" />
          Content
        </h1>
        <p className="mt-2 text-gray-600 dark:text-gray-400">
          Run server-side content generation jobs. Outputs are drafts and require human review
          before publishing.
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {(Object.keys(JOB_META) as JobKey[]).map((key) => {
          const m = mutations[key];
          const meta = JOB_META[key];
          return (
            <div
              key={key}
              className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg p-5"
            >
              <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">{meta.title}</h2>
              <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">{meta.description}</p>
              <div className="mt-4 flex items-center gap-2">
                <button
                  type="button"
                  disabled={m.isPending}
                  onClick={() => m.mutate()}
                  className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 flex items-center gap-2"
                >
                  {m.isPending ? <RotateCw className="w-4 h-4 animate-spin" /> : <Play className="w-4 h-4" />}
                  {m.isPending ? 'Running…' : 'Run job'}
                </button>
                {history.find((h) => h.job === key) && (
                  <button
                    type="button"
                    onClick={() => {
                      const last = history.find((h) => h.job === key);
                      if (last) setOpenResult(last);
                    }}
                    className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-800 flex items-center gap-1.5"
                  >
                    <Eye className="w-4 h-4" /> View last run
                  </button>
                )}
              </div>
            </div>
          );
        })}
      </div>

      {/* History */}
      <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
        <div className="px-4 py-3 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
          <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Recent runs</h2>
          {history.length > 0 && (
            <button
              type="button"
              onClick={() => setHistory([])}
              className="text-xs text-gray-500 hover:text-gray-700 dark:hover:text-gray-300"
            >
              Clear
            </button>
          )}
        </div>
        {history.length === 0 ? (
          <p className="px-4 py-6 text-sm text-gray-500 dark:text-gray-400 text-center">
            No jobs run yet in this session.
          </p>
        ) : (
          <ul className="divide-y divide-gray-100 dark:divide-gray-800">
            {history.map((h) => (
              <li
                key={`${h.job}-${h.finishedAt}`}
                className="px-4 py-2.5 flex items-center gap-3 hover:bg-gray-50/60 dark:hover:bg-gray-950/40"
              >
                {h.status === 'success' ? (
                  <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0" />
                ) : (
                  <AlertTriangle className="w-4 h-4 text-red-600 shrink-0" />
                )}
                <div className="min-w-0 flex-1">
                  <p className="text-sm font-medium text-gray-900 dark:text-gray-100">{h.title}</p>
                  <p className="text-xs text-gray-500 dark:text-gray-400 truncate">{h.message}</p>
                </div>
                <span className="text-xs text-gray-500 dark:text-gray-400 shrink-0">
                  {new Date(h.finishedAt).toLocaleTimeString()}
                </span>
                <button
                  type="button"
                  onClick={() => setOpenResult(h)}
                  className="ml-2 p-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-600 dark:text-gray-300"
                  aria-label="View result"
                >
                  <Eye className="w-4 h-4" />
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>

      {/* Result modal */}
      {openResult && (
        <div
          className="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4"
          role="dialog"
          aria-modal="true"
          aria-labelledby="content-job-result-title"
          onClick={() => setOpenResult(null)}
        >
          <div
            className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg w-full max-w-2xl max-h-[80vh] flex flex-col"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="px-5 py-3 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
              <div>
                <h2 id="content-job-result-title" className="text-lg font-semibold">
                  {openResult.title}
                </h2>
                <p className="text-xs text-gray-500 dark:text-gray-400">
                  {new Date(openResult.finishedAt).toLocaleString()} ·{' '}
                  {openResult.status === 'success' ? 'Succeeded' : 'Failed'}
                </p>
              </div>
              <button
                type="button"
                onClick={() => setOpenResult(null)}
                className="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-800"
                aria-label="Close"
              >
                <span aria-hidden="true">×</span>
              </button>
            </div>
            <div className="p-5 overflow-y-auto">
              {openResult.output ? (
                <pre className="whitespace-pre-wrap text-sm font-mono bg-gray-50 dark:bg-gray-950 border border-gray-200 dark:border-gray-800 rounded-md p-3">
                  {openResult.output}
                </pre>
              ) : (
                <p className="text-sm text-gray-600 dark:text-gray-300">{openResult.message}</p>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default AdminContentPage;
