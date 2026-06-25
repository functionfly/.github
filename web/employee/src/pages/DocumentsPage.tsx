import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { documentsApi, type Document } from '@/api/documents';
import { FileText, Plus, Share2, Eye, Search, Filter } from 'lucide-react';

const typeColors: Record<string, string> = {
  note: 'bg-gray-500/20 text-gray-400',
  policy: 'bg-red-500/20 text-red-400',
  template: 'bg-purple-500/20 text-purple-400',
  meeting_notes: 'bg-blue-500/20 text-blue-400',
  design_doc: 'bg-green-500/20 text-green-400',
};

export function DocumentsPage() {
  const queryClient = useQueryClient();
  const [tab, setTab] = useState<'all' | 'templates'>('all');
  const [typeFilter, setTypeFilter] = useState('');
  const [showCreate, setShowCreate] = useState(false);
  const [showShare, setShowShare] = useState<string | null>(null);
  const [shareEmployeeId, setShareEmployeeId] = useState('');
  const [form, setForm] = useState({ title: '', doc_type: 'note', category: '' });

  const { data, isLoading } = useQuery({
    queryKey: ['documents', tab, typeFilter],
    queryFn: () => tab === 'templates' ? documentsApi.listTemplates() : documentsApi.list(typeFilter ? { doc_type: typeFilter } : undefined),
  });

  const createMutation = useMutation({
    mutationFn: (data: Partial<Document>) => documentsApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['documents'] });
      setShowCreate(false);
      setForm({ title: '', doc_type: 'note', category: '' });
    },
  });

  const shareMutation = useMutation({
    mutationFn: ({ docId, employeeId }: { docId: string; employeeId: string }) => documentsApi.share(docId, employeeId),
    onSuccess: () => {
      setShowShare(null);
      setShareEmployeeId('');
    },
  });

  const documents = data?.data?.documents || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Documents</h1>
        <button
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          <Plus className="h-4 w-4" />
          New Document
        </button>
      </div>

      <div className="flex gap-1 rounded-lg bg-gray-900 p-1">
        <button
          onClick={() => setTab('all')}
          className={`flex-1 rounded-md px-4 py-2 text-sm font-medium ${tab === 'all' ? 'bg-blue-600 text-white' : 'text-gray-400 hover:text-gray-200'}`}
        >
          All Documents
        </button>
        <button
          onClick={() => setTab('templates')}
          className={`flex-1 rounded-md px-4 py-2 text-sm font-medium ${tab === 'templates' ? 'bg-blue-600 text-white' : 'text-gray-400 hover:text-gray-200'}`}
        >
          Templates
        </button>
      </div>

      {tab === 'all' && (
        <div className="flex items-center gap-3">
          <Filter className="h-4 w-4 text-gray-400" />
          <select
            value={typeFilter}
            onChange={(e) => setTypeFilter(e.target.value)}
            className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
          >
            <option value="">All Types</option>
            <option value="note">Note</option>
            <option value="policy">Policy</option>
            <option value="template">Template</option>
            <option value="meeting_notes">Meeting Notes</option>
            <option value="design_doc">Design Doc</option>
          </select>
        </div>
      )}

      {isLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : documents.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
          <FileText className="mb-4 h-12 w-12 text-gray-600" />
          <p className="text-gray-400">{tab === 'templates' ? 'No templates available' : 'No documents yet'}</p>
        </div>
      ) : (
        <div className="space-y-3">
          {documents.map((doc) => (
            <div key={doc.id} className="flex items-center justify-between rounded-xl border border-gray-800 bg-gray-900 p-4">
              <div className="flex items-center gap-4">
                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-gray-800">
                  <FileText className="h-5 w-5 text-gray-400" />
                </div>
                <div>
                  <h3 className="font-medium text-gray-100">{doc.title}</h3>
                  <div className="mt-1 flex items-center gap-3 text-xs text-gray-500">
                    <span className={`rounded-full px-2 py-0.5 ${typeColors[doc.doc_type] || ''}`}>
                      {doc.doc_type.replace('_', ' ')}
                    </span>
                    {doc.category && <span>{doc.category}</span>}
                    <span className="flex items-center gap-1">
                      <Eye className="h-3 w-3" />
                      {doc.view_count}
                    </span>
                    <span>{new Date(doc.updated_at).toLocaleDateString()}</span>
                  </div>
                </div>
              </div>
              <button
                onClick={() => setShowShare(doc.id)}
                className="rounded-lg p-2 text-gray-400 hover:bg-gray-800 hover:text-gray-200"
              >
                <Share2 className="h-4 w-4" />
              </button>
            </div>
          ))}
        </div>
      )}

      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">New Document</h2>
            <input
              type="text"
              placeholder="Title"
              value={form.title}
              onChange={(e) => setForm({ ...form, title: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              autoFocus
            />
            <select
              value={form.doc_type}
              onChange={(e) => setForm({ ...form, doc_type: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
            >
              <option value="note">Note</option>
              <option value="policy">Policy</option>
              <option value="template">Template</option>
              <option value="meeting_notes">Meeting Notes</option>
              <option value="design_doc">Design Doc</option>
            </select>
            <input
              type="text"
              placeholder="Category (optional)"
              value={form.category}
              onChange={(e) => setForm({ ...form, category: e.target.value })}
              className="mb-4 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
            />
            <div className="flex justify-end gap-3">
              <button
                onClick={() => setShowCreate(false)}
                className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200"
              >
                Cancel
              </button>
              <button
                onClick={() => createMutation.mutate({
                  title: form.title,
                  doc_type: form.doc_type,
                  category: form.category || undefined,
                })}
                disabled={!form.title.trim()}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                Create
              </button>
            </div>
          </div>
        </div>
      )}

      {showShare && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-sm rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Share Document</h2>
            <input
              type="text"
              placeholder="Employee ID"
              value={shareEmployeeId}
              onChange={(e) => setShareEmployeeId(e.target.value)}
              className="mb-4 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              autoFocus
            />
            <div className="flex justify-end gap-3">
              <button
                onClick={() => { setShowShare(null); setShareEmployeeId(''); }}
                className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200"
              >
                Cancel
              </button>
              <button
                onClick={() => shareMutation.mutate({ docId: showShare, employeeId: shareEmployeeId })}
                disabled={!shareEmployeeId.trim()}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                Share
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
