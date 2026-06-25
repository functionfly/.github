import { useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { knowledgeApi } from '@/api/knowledge';
import { ArrowLeft, Calendar, Eye, Tag } from 'lucide-react';
import { formatDate } from '@/lib/utils';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

export function KnowledgeArticlePage() {
  const { slug } = useParams<{ slug: string }>();
  const navigate = useNavigate();

  const { data, isLoading } = useQuery({
    queryKey: ['knowledge', slug],
    queryFn: () => knowledgeApi.get(slug!),
    enabled: !!slug,
  });

  const article = data?.data?.article;

  if (isLoading) {
    return (
      <div className="flex justify-center py-12">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
      </div>
    );
  }

  if (!article) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <p className="text-gray-400">Article not found</p>
        <button onClick={() => navigate('/knowledge')} className="mt-2 text-blue-400 hover:underline">
          Back to Knowledge Base
        </button>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <button onClick={() => navigate('/knowledge')} className="flex items-center gap-2 text-gray-400 hover:text-gray-200">
        <ArrowLeft className="h-4 w-4" />
        Back to Knowledge Base
      </button>

      <article className="rounded-xl border border-gray-800 bg-gray-900 p-6">
        <h1 className="mb-2 text-2xl font-bold text-gray-100">{article.title}</h1>

        <div className="mb-6 flex flex-wrap items-center gap-3 text-sm text-gray-500">
          {article.category && (
            <span className="rounded-full bg-gray-800 px-2 py-0.5 text-xs">{article.category}</span>
          )}
          {article.published_at && (
            <span className="flex items-center gap-1">
              <Calendar className="h-3 w-3" />
              {formatDate(article.published_at)}
            </span>
          )}
          <span className="flex items-center gap-1">
            <Eye className="h-3 w-3" />
            {article.view_count} views
          </span>
          {article.tags && article.tags.length > 0 && (
            <span className="flex items-center gap-1">
              <Tag className="h-3 w-3" />
              {article.tags.join(', ')}
            </span>
          )}
        </div>

        <div className="prose prose-invert max-w-none">
          <ReactMarkdown remarkPlugins={[remarkGfm]}>{article.body}</ReactMarkdown>
        </div>
      </article>
    </div>
  );
}
