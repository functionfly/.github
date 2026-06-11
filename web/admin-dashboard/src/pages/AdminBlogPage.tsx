/**
 * Admin Blog Page
 * Tabbed: Settings, Analytics, Posts (CRUD), Categories (CRUD)
 */

import { RichTextEditor } from '@/components/ui/RichTextEditor';
import { adminApiClient } from '@/lib/api/adminClient';
import { logger } from '@/lib/monitoring/logger';
import { blogSettingsStore, type BlogSettings } from '@/stores/blogSettingsStore';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { BarChart3, FileText, FolderTree, Pencil, Plus, Settings, Trash2, Users, X } from 'lucide-react';
import { useEffect, useState } from 'react';

export default function AdminBlogPage() {
  const [activeTab, setActiveTab] = useState<TabId>('settings');

  const tabs = [
    { id: 'settings' as const, label: 'Settings', icon: Settings },
    { id: 'analytics' as const, label: 'Analytics', icon: BarChart3 },
    { id: 'posts' as const, label: 'Posts', icon: FileText },
    { id: 'categories' as const, label: 'Categories', icon: FolderTree },
    { id: 'authors' as const, label: 'Authors', icon: Users },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">Blog</h1>
        <p className="mt-2 text-gray-600 dark:text-gray-400">
          Manage blog settings, analytics, posts, and categories.
        </p>
      </div>

      <div className="border-b border-gray-200 dark:border-gray-700">
        <nav className="flex gap-6 flex-wrap">
          {tabs.map(({ id, label, icon: Icon }) => (
            <button
              key={id}
              type="button"
              onClick={() => setActiveTab(id)}
              className={`flex items-center gap-2 pb-3 px-1 border-b-2 font-medium text-sm ${
                activeTab === id
                  ? 'border-blue-600 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-200 dark:hover:border-gray-600'
              }`}
            >
              <Icon className="w-4 h-4" />
              {label}
            </button>
          ))}
        </nav>
      </div>

      {activeTab === 'settings' && <BlogSettingsTab />}
      {activeTab === 'analytics' && <BlogAnalyticsTab />}
      {activeTab === 'posts' && <BlogPostsTab />}
      {activeTab === 'categories' && <BlogCategoriesTab />}
      {activeTab === 'authors' && <BlogAuthorsTab />}
    </div>
  );
}

function BlogSettingsTab() {
  const queryClient = useQueryClient();

  const { data: remoteSettings } = useQuery({
    queryKey: ['admin-blog-settings'],
    queryFn: async () => {
      const res = await adminApiClient.get<{
        id: string;
        blogTitle: string;
        postsPerPage: number;
        metaDescription: string;
      }>('/blog/settings');
      return {
        blogTitle: res.data?.blogTitle ?? 'FunctionFly Blog',
        postsPerPage: res.data?.postsPerPage ?? 10,
        metaDescription: res.data?.metaDescription ?? '',
      } as BlogSettings;
    },
  });

  const localDefaults = blogSettingsStore.load();
  const initial: BlogSettings = {
    blogTitle: remoteSettings?.blogTitle ?? localDefaults.blogTitle,
    postsPerPage: remoteSettings?.postsPerPage ?? localDefaults.postsPerPage,
    metaDescription: remoteSettings?.metaDescription ?? localDefaults.metaDescription,
  };

  const [blogTitle, setBlogTitle] = useState(initial.blogTitle);
  const [postsPerPage, setPostsPerPage] = useState(initial.postsPerPage);
  const [metaDescription, setMetaDescription] = useState(initial.metaDescription);
  const [saving, setSaving] = useState(false);
  const [saveMessage, setSaveMessage] = useState<'success' | 'error' | null>(null);

  useEffect(() => {
    if (remoteSettings) {
      setBlogTitle(remoteSettings.blogTitle);
      setPostsPerPage(remoteSettings.postsPerPage);
      setMetaDescription(remoteSettings.metaDescription);
    }
  }, [remoteSettings]);

  const saveMutation = useMutation({
    mutationFn: async (settings: BlogSettings) => {
      await adminApiClient.put('/blog/settings', {
        blogTitle: settings.blogTitle,
        postsPerPage: settings.postsPerPage,
        metaDescription: settings.metaDescription,
      });
    },
    onSuccess: (_data, variables) => {
      blogSettingsStore.save(variables);
      queryClient.invalidateQueries({ queryKey: ['admin-blog-settings'] });
      setSaveMessage('success');
      setTimeout(() => setSaveMessage(null), 3000);
    },
    onError: () => {
      setSaveMessage('error');
    },
    onSettled: () => {
      setSaving(false);
    },
  });

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault();
    if (saving) return;
    setSaving(true);
    setSaveMessage(null);
    saveMutation.mutate({ blogTitle, postsPerPage, metaDescription });
  };

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6 max-w-2xl">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Blog settings</h2>
        {saveMessage === 'success' && (
          <span className="text-sm text-emerald-600 font-medium">Saved</span>
        )}
        {saveMessage === 'error' && (
          <span className="text-sm text-red-600 font-medium">Failed to save</span>
        )}
      </div>
      <p className="text-gray-600 dark:text-gray-400 mb-6">Configure how your blog appears and behaves.</p>
      <form className="space-y-5" onSubmit={handleSave}>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Blog title</label>
          <input
            type="text"
            value={blogTitle}
            onChange={(e) => setBlogTitle(e.target.value)}
            className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Posts per page</label>
          <input
            type="number"
            min={1}
            max={50}
            value={postsPerPage}
            onChange={(e) => setPostsPerPage(Number(e.target.value) || 10)}
            className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Default meta description
          </label>
          <textarea
            rows={2}
            value={metaDescription}
            onChange={(e) => setMetaDescription(e.target.value)}
            className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
            placeholder="Short description for SEO"
          />
        </div>
        <button
          type="submit"
          disabled={saving}
          className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
        >
          {saving ? 'Saving…' : 'Save settings'}
        </button>
      </form>
    </div>
  );
}

function BlogAnalyticsTab() {
  return (
    <div className="space-y-6">
      <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2">Blog analytics</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Track views, engagement, and top posts. Integrate with your analytics provider (e.g.
          Google Analytics) or add server-side event tracking to see data here.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="rounded-lg bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 p-4">
            <p className="text-sm text-gray-600 dark:text-gray-400">Total views</p>
            <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">—</p>
          </div>
          <div className="rounded-lg bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 p-4">
            <p className="text-sm text-gray-600 dark:text-gray-400">Posts published</p>
            <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">—</p>
          </div>
          <div className="rounded-lg bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 p-4">
            <p className="text-sm text-gray-600 dark:text-gray-400">Top post</p>
            <p className="text-lg font-medium text-gray-900 dark:text-gray-100">—</p>
          </div>
        </div>
        <p className="mt-4 text-sm text-gray-500 dark:text-gray-400">
          Platform-wide analytics are available under Admin → Analytics. Blog-specific events can be
          added in a future release.
        </p>
      </div>
    </div>
  );
}

function BlogPostsTab() {
  const queryClient = useQueryClient();
  const [showForm, setShowForm] = useState(false);
  const [editingPost, setEditingPost] = useState<BlogPost | null>(null);
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null);

  const {
    data: listData,
    isLoading,
    error: queryError,
  } = useQuery({
    queryKey: ['admin-blog-posts'],
    queryFn: async () => {
      const res = await adminApiClient.get<{ data: BlogPost[]; meta: { total: number; page: number; limit: number; totalPages: number } }>(
        '/blog/posts?limit=50'
      );
      const posts: BlogPost[] = (res.data?.data ?? []).map((p: any) => {
        return {
          id: p.id,
          title: p.title,
          slug: p.slug,
          content: typeof p.body === 'object' && p.body !== null ? JSON.stringify(p.body) : (typeof p.content === 'string' ? p.content : ''),
          body: typeof p.body === 'object' && p.body !== null ? p.body : null,
          excerpt: p.description ?? '',
          author: { name: typeof p.author === 'string' ? p.author : (p.author?.name ?? 'Unknown') },
          tags: Array.isArray(p.tags) ? p.tags : [],
          featured_image: typeof p.heroImage === 'object' && p.heroImage !== null ? (p.heroImage as { url?: string }).url : null,
          is_published: p.status === 'published',
          published_at: p.published_at,
          created_at: p.created_at,
          updated_at: p.updated_at ?? p.created_at,
          description: p.description ?? '',
          heroImage: p.heroImage,
          status: p.status,
          publishedAt: p.published_at,
          createdAt: p.created_at,
          updatedAt: p.updated_at,
        };
      });
      return { posts, limit: 50, offset: 0 };
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => adminApiClient.delete(`/blog/posts/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-blog-posts'] });
      setDeleteConfirm(null);
    },
  });

  const posts = listData?.posts ?? [];

  // Handle loading state after all hooks are called (React Rules of Hooks)
  if (isLoading) {
    return <div className="text-gray-500 dark:text-gray-400">Loading posts…</div>;
  }

  // Handle error state gracefully without crashing
  if (queryError) {
    return (
      <div className="space-y-6">
        <div className="flex justify-between items-center">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Blog posts</h2>
          <button
            type="button"
            onClick={() => {
              setEditingPost(null);
              setShowForm(true);
            }}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            <Plus className="w-4 h-4" />
            New post
          </button>
        </div>
        <div className="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-lg p-4 text-amber-800 dark:text-amber-200">
          <p className="font-medium">Unable to load posts</p>
          <p className="text-sm mt-1">Please try refreshing the page or check your connection.</p>
        </div>
        {showForm && (
          <PostForm
            post={editingPost ?? undefined}
            onClose={() => {
              setShowForm(false);
              setEditingPost(null);
            }}
            onSaved={() => {
              queryClient.invalidateQueries({ queryKey: ['admin-blog-posts'] });
              setShowForm(false);
              setEditingPost(null);
            }}
          />
        )}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Blog posts</h2>
        <button
          type="button"
          onClick={() => {
            setEditingPost(null);
            setShowForm(true);
          }}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          <Plus className="w-4 h-4" />
          New post
        </button>
      </div>

      {showForm && (
        <PostForm
          post={editingPost ?? undefined}
          onClose={() => {
            setShowForm(false);
            setEditingPost(null);
          }}
          onSaved={() => {
            queryClient.invalidateQueries({ queryKey: ['admin-blog-posts'] });
            setShowForm(false);
            setEditingPost(null);
          }}
        />
      )}

      <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">Title</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">Slug</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">Author</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">Status</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">Updated</th>
              <th className="px-6 py-3 text-right text-sm font-semibold text-gray-700 dark:text-gray-300">Actions</th>
            </tr>
          </thead>
          <tbody>
            {posts.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-6 py-8 text-center text-gray-500 dark:text-gray-400">
                  No posts yet. Create one with "New post".
                </td>
              </tr>
            ) : (
              posts.map((post) => (
                <tr key={post.id} className="border-b border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800">
                  <td className="px-6 py-4 text-sm font-medium text-gray-900 dark:text-gray-100">{post.title}</td>
                  <td className="px-6 py-4 text-sm text-gray-600 dark:text-gray-400">{post.slug}</td>
                   <td className="px-6 py-4 text-sm text-gray-600 dark:text-gray-400">
                     {typeof post.author === 'string' ? post.author : post.author?.name ?? 'Unknown'}
                   </td>
                  <td className="px-6 py-4 text-sm">
                    <span className={post.is_published ? 'text-green-600 dark:text-green-400' : 'text-amber-600 dark:text-amber-400'}>
                      {post.is_published ? 'Published' : 'Draft'}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                    {new Date(post.updated_at).toLocaleDateString()}
                  </td>
                  <td className="px-6 py-4 text-sm text-right">
                    <button
                      type="button"
                      onClick={() => {
                        setEditingPost(post);
                        setShowForm(true);
                      }}
                      className="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 mr-3"
                    >
                      <Pencil className="w-4 h-4 inline" />
                    </button>
                    {deleteConfirm === post.id ? (
                      <span className="flex items-center justify-end gap-2">
                        <button
                          type="button"
                          onClick={() => deleteMutation.mutate(post.id)}
                          className="text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300 font-medium"
                        >
                          Confirm
                        </button>
                        <button
                          type="button"
                          onClick={() => setDeleteConfirm(null)}
                          className="text-gray-600 dark:text-gray-400"
                        >
                          Cancel
                        </button>
                      </span>
                    ) : (
                      <button
                        type="button"
                        onClick={() => setDeleteConfirm(post.id)}
                        className="text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300"
                      >
                        <Trash2 className="w-4 h-4 inline" />
                      </button>
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function PostForm({
  post,
  onClose,
  onSaved,
}: {
  post?: BlogPost;
  onClose: () => void;
  onSaved: () => void;
}) {
  const [title, setTitle] = useState(post?.title ?? '');
  const [slug, setSlug] = useState(post?.slug ?? '');
  const [content, setContent] = useState(post?.content ?? '');
  const [excerpt, setExcerpt] = useState(post?.excerpt ?? '');
  const [author, setAuthor] = useState(typeof post?.author === 'string' ? post.author : (post?.author?.name ?? ''));
  const [tagsStr, setTagsStr] = useState(Array.isArray(post?.tags) ? post.tags.join(', ') : '');
  const [isPublished, setIsPublished] = useState(post?.is_published ?? false);
  const [featuredImage, setFeaturedImage] = useState(post?.featured_image ?? '');
  const [seoTitle, setSeoTitle] = useState(post?.seoTitle ?? '');
  const [seoDescription, setSeoDescription] = useState(post?.seoDescription ?? '');
  const [keywordsStr, setKeywordsStr] = useState(Array.isArray(post?.keywords) ? post.keywords.join(', ') : '');
  const [canonicalUrl, setCanonicalUrl] = useState(post?.canonicalUrl ?? '');
  const [ogImageUrl, setOgImageUrl] = useState(post?.ogImage?.url ?? '');
  const [ogImageAlt, setOgImageAlt] = useState(post?.ogImage?.alt ?? '');

  useEffect(() => {
    if (post) {
      setTitle(post.title);
      setSlug(post.slug);
      // Prefer body (TipTap JSON) over content (plain text)
      if (post.body && typeof post.body === 'object') {
        setContent(JSON.stringify(post.body));
      } else if (post.content) {
        // Convert legacy content to TipTap format
        const paragraphs = post.content
          .split('\n\n')
          .filter((p) => p.trim())
          .map((p) => ({
            type: 'paragraph',
            content: [{ type: 'text', text: p.trim() }],
          }));
        setContent(JSON.stringify({ type: 'doc', content: paragraphs }));
      } else {
        setContent('{"type":"doc","content":[]}');
      }
      setExcerpt(post.excerpt ?? '');
      setAuthor(typeof post.author === 'string' ? post.author : (post.author?.name ?? 'Unknown'));
      setTagsStr(Array.isArray(post.tags) ? post.tags.join(', ') : '');
      setIsPublished(post.is_published ?? false);
      setFeaturedImage(post.featured_image ?? '');
      setSeoTitle(post.seoTitle ?? '');
      setSeoDescription(post.seoDescription ?? '');
      setKeywordsStr(Array.isArray(post.keywords) ? post.keywords.join(', ') : '');
      setCanonicalUrl(post.canonicalUrl ?? '');
      setOgImageUrl(post.ogImage?.url ?? '');
      setOgImageAlt(post.ogImage?.alt ?? '');
    }
  }, [post]);

  const createMutation = useMutation({
    mutationFn: (payload: Record<string, unknown>) => adminApiClient.post('/blog/posts', payload),
    onSuccess: () => {
      onSaved();
    },
    onError: (error) => {
      logger.error('Create post error', { error });
      alert(`Failed to create post: ${error.message}`);
    },
  });
  const updateMutation = useMutation({
    mutationFn: (payload: Record<string, unknown>) =>
      adminApiClient.put(`/blog/posts/${post!.id}`, payload),
    onSuccess: () => {
      onSaved();
    },
    onError: (error) => {
      logger.error('Update post error', { error });
      alert(`Failed to update post: ${error.message}`);
    },
  });

  const tags = tagsStr
    .split(',')
    .map((t) => t.trim())
    .filter(Boolean);

  const keywords = keywordsStr
    .split(',')
    .map((k) => k.trim())
    .filter(Boolean);

  // Parse content as TipTap JSON or fallback to plain text in content field
  let bodyContent: unknown;
  try {
    bodyContent = JSON.parse(content);
  } catch {
    // If not valid JSON, wrap as plain text paragraphs
    bodyContent = {
      type: 'doc',
      content: content
        .split('\n\n')
        .filter((p) => p.trim())
        .map((p) => ({
          type: 'paragraph',
          content: [{ type: 'text', text: p.trim() }],
        })),
    };
  }

  // Helper to extract plain text from TipTap JSON
  const extractPlainText = (body: unknown): string => {
    if (typeof body !== 'object' || body === null) return String(body || '');
    const doc = body as { content?: Array<{ content?: Array<{ text?: string }> }> };
    if (!doc.content) return '';
    return doc.content.map((p) => p.content?.map((c) => c.text).join(' ') || '').join('\n\n');
  };

  const payload = {
    title,
    slug:
      slug ||
      title
        .toLowerCase()
        .replace(/\s+/g, '-')
        .replace(/[^a-z0-9-]/g, ''),
    content:
      typeof bodyContent === 'object' && bodyContent !== null
        ? extractPlainText(bodyContent)
        : content,
    body: bodyContent,
    excerpt,
    author,
    tags,
    is_published: isPublished,
    featured_image: featuredImage || undefined,
    seo_title: seoTitle || undefined,
    seo_description: seoDescription || undefined,
    keywords: keywords.length > 0 ? keywords : undefined,
    canonical_url: canonicalUrl || undefined,
    og_image: ogImageUrl ? { url: ogImageUrl, alt: ogImageAlt } : undefined,
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (post) {
      updateMutation.mutate(payload, {
        onError: (err: { message?: string } & { response?: { data?: unknown } }) => {
          logger.error('Update post error', { error: err, response: err.response?.data });
          alert(`Failed to update post: ${err.message ?? 'Unknown error'}`);
        }
      });
    } else {
      createMutation.mutate(payload);
    }
  };

  const saving = createMutation.isPending || updateMutation.isPending;

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
      <div className="flex justify-between items-center mb-4">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">{post ? 'Edit post' : 'New post'}</h3>
        <button type="button" onClick={onClose} className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200">
          <X className="w-5 h-5" />
        </button>
      </div>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Title *</label>
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            required
            className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Slug *</label>
          <input
            type="text"
            value={slug}
            onChange={(e) => setSlug(e.target.value)}
            placeholder={title ? title.toLowerCase().replace(/\s+/g, '-') : ''}
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-gray-900 focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Author *</label>
          <input
            type="text"
            value={author}
            onChange={(e) => setAuthor(e.target.value)}
            required
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-gray-900 focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Excerpt</label>
          <textarea
            rows={2}
            value={excerpt}
            onChange={(e) => setExcerpt(e.target.value)}
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-gray-900 focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">Content *</label>
          <RichTextEditor
            content={
              content ||
              '{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Start writing your post..."}]}]}'
            }
            onChange={setContent}
            placeholder="Start writing your blog post..."
            minHeight="400px"
          />
          <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
            Use the toolbar to format text, add headings, lists, links, and code blocks.
          </p>
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Tags (comma-separated)
          </label>
          <input
            type="text"
            value={tagsStr}
            onChange={(e) => setTagsStr(e.target.value)}
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-gray-900 focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Featured image URL</label>
          <input
            type="url"
            value={featuredImage}
            onChange={(e) => setFeaturedImage(e.target.value)}
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-gray-900 focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
          />
          {featuredImage && (
            <div className="mt-2">
              <p className="text-xs text-gray-500 mb-1">Preview</p>
              <img
                src={featuredImage}
                alt="Featured preview"
                className="max-w-xs rounded border border-gray-200 dark:border-gray-600"
                onError={(e) => {
                  (e.target as HTMLImageElement).style.display = 'none';
                }}
              />
            </div>
          )}
        </div>

        <details className="group border border-gray-200 dark:border-gray-700 rounded-lg p-4">
          <summary className="text-sm font-medium text-gray-700 dark:text-gray-300 cursor-pointer list-none flex items-center justify-between">
            <span>SEO Settings</span>
            <span className="text-gray-400 group-open:rotate-180 transition-transform">▼</span>
          </summary>
          <div className="mt-4 space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">SEO Title</label>
              <input
                type="text"
                value={seoTitle}
                onChange={(e) => setSeoTitle(e.target.value)}
                className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
                placeholder="Override title for search engines"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">SEO Description</label>
              <textarea
                rows={2}
                value={seoDescription}
                onChange={(e) => setSeoDescription(e.target.value)}
                className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
                placeholder="Meta description for search results"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Keywords (comma-separated)</label>
              <input
                type="text"
                value={keywordsStr}
                onChange={(e) => setKeywordsStr(e.target.value)}
                className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
                placeholder="keyword1, keyword2, keyword3"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Canonical URL</label>
              <input
                type="url"
                value={canonicalUrl}
                onChange={(e) => setCanonicalUrl(e.target.value)}
                className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
                placeholder="https://..."
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">OG Image URL</label>
                <input
                  type="url"
                  value={ogImageUrl}
                  onChange={(e) => setOgImageUrl(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
                  placeholder="https://.../og-image.png"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">OG Image Alt</label>
                <input
                  type="text"
                  value={ogImageAlt}
                  onChange={(e) => setOgImageAlt(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
                  placeholder="Alt text for OG image"
                />
              </div>
            </div>
            {ogImageUrl && (
              <div>
                <p className="text-xs text-gray-500 dark:text-gray-400 mb-1">OG Image Preview</p>
                <img src={ogImageUrl} alt={ogImageAlt || 'OG preview'} className="max-w-xs rounded border border-gray-200 dark:border-gray-600" />
              </div>
            )}
          </div>
        </details>

        <div className="flex items-center gap-3">
          <input
            type="checkbox"
            id="is_published"
            checked={isPublished}
            onChange={(e) => setIsPublished(e.target.checked)}
            className="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
          />
          <label htmlFor="is_published" className="text-sm font-medium text-gray-700 dark:text-gray-300">
            Published
          </label>
        </div>
        <div className="flex gap-3 pt-2">
          <button
            type="submit"
            disabled={saving}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
          >
            {saving ? 'Saving…' : post ? 'Update post' : 'Create post'}
          </button>
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  );
}

function BlogCategoriesTab() {
  const queryClient = useQueryClient();
  const [showAdd, setShowAdd] = useState(false);
  const [editing, setEditing] = useState<BlogCategory | null>(null);
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null);

  const {
    data: categories = [],
    isLoading,
    error: queryError,
  } = useQuery({
    queryKey: ['admin-blog-categories'],
    queryFn: async () => {
      const res = await adminApiClient.get<BlogCategory[]>('/content/categories');
      return res?.data ?? [];
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => adminApiClient.delete(`/content/categories/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-blog-categories'] });
      setDeleteConfirm(null);
    },
  });

  // Handle loading state after all hooks are called (React Rules of Hooks)
  if (isLoading) {
    return <div className="text-gray-500 dark:text-gray-400">Loading categories…</div>;
  }

  // Handle error state gracefully without crashing
  if (queryError) {
    return (
      <div className="space-y-6">
        <div className="flex justify-between items-center">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Categories</h2>
          <button
            type="button"
            onClick={() => {
              setEditing(null);
              setShowAdd(true);
            }}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            <Plus className="w-4 h-4" />
            Add category
          </button>
        </div>
        <div className="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-lg p-4 text-amber-800 dark:text-amber-200">
          <p className="font-medium">Unable to load categories</p>
          <p className="text-sm mt-1">Please try refreshing the page or check your connection.</p>
        </div>
        {(showAdd || editing) && (
          <CategoryForm
            category={editing ?? undefined}
            onClose={() => {
              setShowAdd(false);
              setEditing(null);
            }}
            onSaved={() => {
              queryClient.invalidateQueries({ queryKey: ['admin-blog-categories'] });
              setShowAdd(false);
              setEditing(null);
            }}
          />
        )}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Categories</h2>
        <button
          type="button"
          onClick={() => {
            setEditing(null);
            setShowAdd(true);
          }}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          <Plus className="w-4 h-4" />
          Add category
        </button>
      </div>

      {(showAdd || editing) && (
        <CategoryForm
          category={editing ?? undefined}
          onClose={() => {
            setShowAdd(false);
            setEditing(null);
          }}
          onSaved={() => {
            queryClient.invalidateQueries({ queryKey: ['admin-blog-categories'] });
            setShowAdd(false);
            setEditing(null);
          }}
        />
      )}

      <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="bg-gray-50 border-b border-gray-200">
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Title</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Slug</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">
                Description
              </th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Order</th>
              <th className="px-6 py-3 text-right text-sm font-semibold text-gray-700">Actions</th>
            </tr>
          </thead>
          <tbody>
            {categories.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-6 py-8 text-center text-gray-500">
                  No categories yet. Add one with “Add category”.
                </td>
              </tr>
            ) : (
              categories.map((cat) => (
                <tr key={cat.id} className="border-b border-gray-100 hover:bg-gray-50">
                  <td className="px-6 py-4 text-sm font-medium text-gray-900">{cat.title}</td>
                  <td className="px-6 py-4 text-sm text-gray-600">{cat.slug}</td>
                  <td className="px-6 py-4 text-sm text-gray-600 max-w-xs truncate">
                    {cat.description || '—'}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-600">{cat.order}</td>
                  <td className="px-6 py-4 text-sm text-right">
                    <button
                      type="button"
                      onClick={() => {
                        setEditing(cat);
                        setShowAdd(false);
                      }}
                      className="text-blue-600 hover:text-blue-800 mr-3"
                    >
                      <Pencil className="w-4 h-4 inline" />
                    </button>
                    {deleteConfirm === cat.id ? (
                      <span className="flex items-center justify-end gap-2">
                        <button
                          type="button"
                          onClick={() => deleteMutation.mutate(cat.id)}
                          className="text-red-600 hover:text-red-800 font-medium"
                        >
                          Confirm
                        </button>
                        <button
                          type="button"
                          onClick={() => setDeleteConfirm(null)}
                          className="text-gray-600"
                        >
                          Cancel
                        </button>
                      </span>
                    ) : (
                      <button
                        type="button"
                        onClick={() => setDeleteConfirm(cat.id)}
                        className="text-red-600 hover:text-red-800"
                      >
                        <Trash2 className="w-4 h-4 inline" />
                      </button>
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function CategoryForm({
  category,
  onClose,
  onSaved,
}: {
  category?: BlogCategory;
  onClose: () => void;
  onSaved: () => void;
}) {
  const [title, setTitle] = useState(category?.title ?? '');
  const [slug, setSlug] = useState(category?.slug ?? '');
  const [description, setDescription] = useState(category?.description ?? '');
  const [color, setColor] = useState(category?.color ?? '');
  const [icon, setIcon] = useState(category?.icon ?? '');
  const [order, setOrder] = useState(category?.order ?? 0);

  useEffect(() => {
    if (category) {
      setTitle(category.title);
      setSlug(category.slug);
      setDescription(category.description ?? '');
      setColor(category.color ?? '');
      setIcon(category.icon ?? '');
      setOrder(category.order ?? 0);
    }
  }, [category]);

  const createMutation = useMutation({
    mutationFn: (payload: Record<string, unknown>) =>
      adminApiClient.post('/content/categories', payload),
    onSuccess: () => onSaved(),
  });
  const updateMutation = useMutation({
    mutationFn: (payload: Record<string, unknown>) =>
      adminApiClient.put(`/content/categories/${category!.id}`, payload),
    onSuccess: () => onSaved(),
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const payload = {
      title: title.trim(),
      slug: slug.trim() || undefined,
      description: description.trim(),
      color,
      icon,
      order,
    };
    if (category) {
      updateMutation.mutate(payload);
    } else {
      createMutation.mutate(payload);
    }
  };

  const saving = createMutation.isPending || updateMutation.isPending;

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
      <div className="flex justify-between items-center mb-4">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
          {category ? 'Edit category' : 'Add category'}
        </h3>
        <button type="button" onClick={onClose} className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200">
          <X className="w-5 h-5" />
        </button>
      </div>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Title *</label>
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            required
            className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Slug</label>
          <input
            type="text"
            value={slug}
            onChange={(e) => setSlug(e.target.value)}
            placeholder={title ? title.toLowerCase().replace(/\s+/g, '-') : ''}
            className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Description</label>
          <textarea
            rows={2}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
          />
        </div>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Color</label>
            <input
              type="text"
              value={color}
              onChange={(e) => setColor(e.target.value)}
              placeholder="e.g. blue"
              className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Icon</label>
            <input
              type="text"
              value={icon}
              onChange={(e) => setIcon(e.target.value)}
              placeholder="e.g. folder"
              className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
            />
          </div>
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Order</label>
          <input
            type="number"
            min={0}
            value={order}
            onChange={(e) => setOrder(Number(e.target.value) || 0)}
            className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
          />
        </div>
        <div className="flex gap-3 pt-2">
          <button
            type="submit"
            disabled={saving}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
          >
            {saving ? 'Saving…' : category ? 'Update category' : 'Add category'}
          </button>
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  );
}

function BlogAuthorsTab() {
  const queryClient = useQueryClient();
  const [showAdd, setShowAdd] = useState(false);
  const [editing, setEditing] = useState<BlogAuthor | null>(null);
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null);

  const {
    data: authors = [],
    isLoading,
    error: queryError,
  } = useQuery({
    queryKey: ['admin-blog-authors'],
    queryFn: async () => {
      const res = await adminApiClient.get<BlogAuthor[]>('/content/authors');
      return res?.data ?? [];
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => adminApiClient.delete(`/content/authors/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-blog-authors'] });
      setDeleteConfirm(null);
    },
  });

  if (isLoading) {
    return <div className="text-gray-500 dark:text-gray-400">Loading authors…</div>;
  }

  if (queryError) {
    return (
      <div className="space-y-6">
        <div className="flex justify-between items-center">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Authors</h2>
          <button
            type="button"
            onClick={() => {
              setEditing(null);
              setShowAdd(true);
            }}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            <Plus className="w-4 h-4" />
            Add author
          </button>
        </div>
        <div className="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-lg p-4 text-amber-800 dark:text-amber-200">
          <p className="font-medium">Unable to load authors</p>
          <p className="text-sm mt-1">Please try refreshing the page or check your connection.</p>
        </div>
        {(showAdd || editing) && (
          <AuthorForm
            author={editing ?? undefined}
            onClose={() => {
              setShowAdd(false);
              setEditing(null);
            }}
            onSaved={() => {
              queryClient.invalidateQueries({ queryKey: ['admin-blog-authors'] });
              setShowAdd(false);
              setEditing(null);
            }}
          />
        )}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Authors</h2>
        <button
          type="button"
          onClick={() => {
            setEditing(null);
            setShowAdd(true);
          }}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          <Plus className="w-4 h-4" />
          Add author
        </button>
      </div>

      {(showAdd || editing) && (
        <AuthorForm
          author={editing ?? undefined}
          onClose={() => {
            setShowAdd(false);
            setEditing(null);
          }}
          onSaved={() => {
            queryClient.invalidateQueries({ queryKey: ['admin-blog-authors'] });
            setShowAdd(false);
            setEditing(null);
          }}
        />
      )}

      <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="bg-gray-50 border-b border-gray-200">
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Name</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Slug</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Email</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Role</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Active</th>
              <th className="px-6 py-3 text-right text-sm font-semibold text-gray-700">Actions</th>
            </tr>
          </thead>
          <tbody>
            {authors.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-6 py-8 text-center text-gray-500">
                  No authors yet. Add one with "Add author".
                </td>
              </tr>
            ) : (
              authors.map((author) => (
                <tr key={author.id} className="border-b border-gray-100 hover:bg-gray-50">
                  <td className="px-6 py-4 text-sm font-medium text-gray-900">{author.name}</td>
                  <td className="px-6 py-4 text-sm text-gray-600">{author.slug}</td>
                  <td className="px-6 py-4 text-sm text-gray-600">{author.email ?? '—'}</td>
                  <td className="px-6 py-4 text-sm text-gray-600">{author.role ?? '—'}</td>
                  <td className="px-6 py-4 text-sm">
                    <span className={author.active ? 'text-green-600 dark:text-green-400' : 'text-gray-400'}>
                      {author.active ? 'Yes' : 'No'}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-sm text-right">
                    <button
                      type="button"
                      onClick={() => {
                        setEditing(author);
                        setShowAdd(false);
                      }}
                      className="text-blue-600 hover:text-blue-800 mr-3"
                    >
                      <Pencil className="w-4 h-4 inline" />
                    </button>
                    {deleteConfirm === author.id ? (
                      <span className="flex items-center justify-end gap-2">
                        <button
                          type="button"
                          onClick={() => deleteMutation.mutate(author.id)}
                          className="text-red-600 hover:text-red-800 font-medium"
                        >
                          Confirm
                        </button>
                        <button
                          type="button"
                          onClick={() => setDeleteConfirm(null)}
                          className="text-gray-600"
                        >
                          Cancel
                        </button>
                      </span>
                    ) : (
                      <button
                        type="button"
                        onClick={() => setDeleteConfirm(author.id)}
                        className="text-red-600 hover:text-red-800"
                      >
                        <Trash2 className="w-4 h-4 inline" />
                      </button>
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function AuthorForm({
  author,
  onClose,
  onSaved,
}: {
  author?: BlogAuthor;
  onClose: () => void;
  onSaved: () => void;
}) {
  const [name, setName] = useState(author?.name ?? '');
  const [slug, setSlug] = useState(author?.slug ?? '');
  const [bio, setBio] = useState(author?.bio ?? '');
  const [email, setEmail] = useState(author?.email ?? '');
  const [website, setWebsite] = useState(author?.website ?? '');
  const [role, setRole] = useState(author?.role ?? '');
  const [photoUrl, setPhotoUrl] = useState(author?.photo?.url ?? '');
  const [active, setActive] = useState(author?.active ?? true);

  useEffect(() => {
    if (author) {
      setName(author.name);
      setSlug(author.slug);
      setBio(author.bio ?? '');
      setEmail(author.email ?? '');
      setWebsite(author.website ?? '');
      setRole(author.role ?? '');
      setPhotoUrl(author.photo?.url ?? '');
      setActive(author.active);
    }
  }, [author]);

  const createMutation = useMutation({
    mutationFn: (payload: Record<string, unknown>) =>
      adminApiClient.post('/content/authors', payload),
    onSuccess: () => onSaved(),
  });
  const updateMutation = useMutation({
    mutationFn: (payload: Record<string, unknown>) =>
      adminApiClient.put(`/content/authors/${author!.id}`, payload),
    onSuccess: () => onSaved(),
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const payload = {
      name: name.trim(),
      slug: slug.trim() || undefined,
      bio: bio.trim() || undefined,
      email: email.trim() || undefined,
      website: website.trim() || undefined,
      role: role.trim() || undefined,
      photo: photoUrl ? { url: photoUrl } : undefined,
      active,
    };
    if (author) {
      updateMutation.mutate(payload);
    } else {
      createMutation.mutate(payload);
    }
  };

  const saving = createMutation.isPending || updateMutation.isPending;

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
      <div className="flex justify-between items-center mb-4">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
          {author ? 'Edit author' : 'Add author'}
        </h3>
        <button type="button" onClick={onClose} className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200">
          <X className="w-5 h-5" />
        </button>
      </div>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Name *</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Slug</label>
          <input
            type="text"
            value={slug}
            onChange={(e) => setSlug(e.target.value)}
            placeholder={name ? name.toLowerCase().replace(/\s+/g, '-') : ''}
            className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Bio</label>
          <textarea
            rows={2}
            value={bio}
            onChange={(e) => setBio(e.target.value)}
            className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
          />
        </div>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Email</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Website</label>
            <input
              type="url"
              value={website}
              onChange={(e) => setWebsite(e.target.value)}
              placeholder="https://..."
              className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
            />
          </div>
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Role</label>
          <input
            type="text"
            value={role}
            onChange={(e) => setRole(e.target.value)}
            placeholder="e.g. Editor, Contributor"
            className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Photo URL</label>
          <input
            type="url"
            value={photoUrl}
            onChange={(e) => setPhotoUrl(e.target.value)}
            placeholder="https://.../photo.jpg"
            className="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-900 dark:text-gray-100 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:bg-gray-700"
          />
        </div>
        <div className="flex items-center gap-3">
          <input
            type="checkbox"
            id="author_active"
            checked={active}
            onChange={(e) => setActive(e.target.checked)}
            className="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
          />
          <label htmlFor="author_active" className="text-sm font-medium text-gray-700 dark:text-gray-300">
            Active
          </label>
        </div>
        <div className="flex gap-3 pt-2">
          <button
            type="submit"
            disabled={saving}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
          >
            {saving ? 'Saving…' : author ? 'Update author' : 'Add author'}
          </button>
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  );
}
