import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { knowledgeApi, type KnowledgeArticle } from '@/api/knowledge';
import { Search, BookOpen, Plus, Eye, Calendar } from 'lucide-react';
import { formatDate } from '@/lib/utils';

export function KnowledgePage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [search, setSearch] = useState('');
  const [category, setCategory] = useState('');
  const [showCreate, setShowCreate] = useState(false);
  const [newTitle, setNewTitle] = useState('');
  const [newBody, setNewBody] = useState('');
  const [newCategory, setNewCategory] = useState('');

  const { data, isLoading } = useQuery({
    queryKey: ['knowledge', { search, category }],
    queryFn: () => search
      ? knowledgeApi.search(search).then((r) => ({ data: { articles: r.data.articles, total: r.data.articles.length } }))
      : knowledgeApi.list({ category: category || undefined }),
  });

  const createMutation = useMutation({
    mutationFn: (data: Partial<KnowledgeArticle>) => knowledgeApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['knowledge'] });
      setShowCreate(false);
      setNewTitle('');
      setNewBody('');
      setNewCategory('');
    },
  });

  const articles = data?.data?.articles || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Knowledge Base</h1>
        <button
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          <Plus className="h-4 w-4" />
          New Article
        </button>
      </div>

      <div className="flex gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
          <input
            type="text"
            placeholder="Search knowledge base..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full rounded-lg border border-gray-700 bg-gray-800 py-2 pl-10 pr-4 text-sm text-gray-100 placeholder-gray-500 focus:border-blue-500 focus:outline-none"
          />
        </div>
        <select
          value={category}
          onChange={(e) => setCategory(e.target.value)}
          className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
        >
          <option value="">All Categories</option>
          <option value="engineering">Engineering</option>
          <option value="product">Product</option>
          <option value="hr">HR</option>
          <option value="security">Security</option>
          <option value="ops">Operations</option>
        </select>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : articles.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
          <BookOpen className="mb-4 h-12 w-12 text-gray-600" />
          <p className="text-gray-400">No articles found</p>
        </div>
      ) : (
        <div className="space-y-3">
          {articles.map((article) => (
            <div
              key={article.id}
              onClick={() => navigate(`/knowledge/${article.slug}`)}
              className="cursor-pointer rounded-xl border border-gray-800 bg-gray-900 p-4 transition-colors hover:border-gray-700"
            >
              <div className="flex items-start justify-between">
                <div>
                  <h3 className="font-semibold text-gray-100">{article.title}</h3>
                  {article.category && (
                    <span className="mt-1 inline-block rounded-full bg-gray-800 px-2 py-0.5 text-xs text-gray-400">
                      {article.category}
                    </span>
                  )}
                </div>
                <div className="flex items-center gap-3 text-xs text-gray-500">
                  <span className="flex items-center gap-1">
                    <Eye className="h-3 w-3" />
                    {article.view_count}
                  </span>
                  {article.published_at && (
                    <span className="flex items-center gap-1">
                      <Calendar className="h-3 w-3" />
                      {formatDate(article.published_at)}
                    </span>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-lg rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">New Article</h2>
            <input
              type="text"
              placeholder="Title"
              value={newTitle}
              onChange={(e) => setNewTitle(e.target.value)}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
              autoFocus
            />
            <select
              value={newCategory}
              onChange={(e) => setNewCategory(e.target.value)}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
            >
              <option value="">Select category</option>
              <option value="engineering">Engineering</option>
              <option value="product">Product</option>
              <option value="hr">HR</option>
              <option value="security">Security</option>
              <option value="ops">Operations</option>
            </select>
            <textarea
              placeholder="Content (Markdown supported)"
              value={newBody}
              onChange={(e) => setNewBody(e.target.value)}
              className="mb-4 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
              rows={8}
            />
            <div className="flex justify-end gap-3">
              <button onClick={() => setShowCreate(false)} className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200">
                Cancel
              </button>
              <button
                onClick={() => createMutation.mutate({ title: newTitle, body: newBody, category: newCategory || undefined })}
                disabled={!newTitle.trim() || !newBody.trim()}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                Publish
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
