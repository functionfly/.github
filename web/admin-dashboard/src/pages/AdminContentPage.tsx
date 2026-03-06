import { useMutation } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';

export function AdminContentPage() {
  const changelogMutation = useMutation({
    mutationFn: () => adminApiClient.post('/content/generate/changelog'),
  });

  const blogMutation = useMutation({
    mutationFn: () => adminApiClient.post('/content/generate/blog'),
  });

  const authorMutation = useMutation({
    mutationFn: () => adminApiClient.post('/content/generate/author'),
  });

  const categoryMutation = useMutation({
    mutationFn: () => adminApiClient.post('/content/generate/category'),
  });

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Content</h1>
        <p className="mt-2 text-gray-600">Manage and generate admin content artifacts.</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <GenerateCard title="Generate Changelog" busy={changelogMutation.isPending} onRun={() => changelogMutation.mutate()} />
        <GenerateCard title="Generate Blog Draft" busy={blogMutation.isPending} onRun={() => blogMutation.mutate()} />
        <GenerateCard title="Generate Author Bio" busy={authorMutation.isPending} onRun={() => authorMutation.mutate()} />
        <GenerateCard title="Generate Category Copy" busy={categoryMutation.isPending} onRun={() => categoryMutation.mutate()} />
      </div>
    </div>
  );
}

function GenerateCard({
  title,
  busy,
  onRun,
}: {
  title: string;
  busy: boolean;
  onRun: () => void;
}) {
  return (
    <div className="bg-white rounded-lg border border-gray-200 p-5">
      <h2 className="text-lg font-semibold text-gray-900">{title}</h2>
      <p className="mt-1 text-sm text-gray-600">Triggers a server-side generation job.</p>
      <button
        type="button"
        disabled={busy}
        onClick={onRun}
        className="mt-4 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
      >
        {busy ? 'Running...' : 'Run'}
      </button>
    </div>
  );
}
