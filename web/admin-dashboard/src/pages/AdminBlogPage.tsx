/**
 * Admin Blog Page
 * Tabbed: Settings, Analytics, Posts (CRUD), Categories (CRUD)
 */

import { RichTextEditor } from '@/components/ui/RichTextEditor';
import { adminApiClient } from '@/lib/api/adminClient';
import { blogSettingsStore, type BlogSettings } from '@/stores/blogSettingsStore';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { BarChart3, FileText, FolderTree, Pencil, Plus, Settings, Trash2, X } from 'lucide-react';
import { useEffect, useState } from 'react';

type TabId = 'settings' | 'analytics' | 'posts' | 'categories';

interface BlogPost {
  id: string;
  title: string;
  slug: string;
  content: string;
  body?: unknown;
  excerpt: string;
  author: string;
  tags: string[];
  featured_image?: string | null;
  is_published: boolean;
  published_at?: string | null;
  created_at: string;
  updated_at: string;
}

interface BlogCategory {
  id: string;
  title: string;
  slug: string;
  description: string;
  color: string;
  icon: string;
  order: number;
  createdAt?: string;
  updatedAt?: string;
}

export function AdminBlogPage() {
  const [activeTab, setActiveTab] = useState<TabId>('posts');

  const tabs = [
    { id: 'settings' as const, label: 'Settings', icon: Settings },
    { id: 'analytics' as const, label: 'Analytics', icon: BarChart3 },
    { id: 'posts' as const, label: 'Posts', icon: FileText },
    { id: 'categories' as const, label: 'Categories', icon: FolderTree },
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
        blogTitle: res.blogTitle ?? 'FunctionFly Blog',
        postsPerPage: res.postsPerPage ?? 10,
        metaDescription: res.metaDescription ?? '',
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
      const posts: BlogPost[] = (res.data ?? []).map((p: BlogPost) => {
        return {
          id: p.id,
          title: p.title,
          slug: p.slug,
          content: typeof p.body === 'object' ? JSON.stringify(p.body) : p.body,
          body: typeof p.body === 'object' ? p.body : null,
          excerpt: p.description,
          author: p.author?.name ?? 'Unknown',
          tags: p.tags ?? [],
          featured_image: p.heroImage?.url,
          is_published: p.status === 'published',
          published_at: p.publishedAt,
          created_at: p.createdAt,
          updated_at: p.updatedAt ?? p.createdAt,
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
            console.log('onSaved called - invalidating queries and closing form');
            queryClient.invalidateQueries({ queryKey: ['admin-blog-posts'] });
            setShowForm(false);
            setEditingPost(null);
            console.log('Form should now be closed');
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
                  <td className="px-6 py-4 text-sm text-gray-600 dark:text-gray-400">{post.author}</td>
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
  const [author, setAuthor] = useState(post?.author ?? '');
  const [tagsStr, setTagsStr] = useState(Array.isArray(post?.tags) ? post.tags.join(', ') : '');
  const [isPublished, setIsPublished] = useState(post?.is_published ?? false);
  const [featuredImage, setFeaturedImage] = useState(post?.featured_image ?? '');

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
      setExcerpt(post.excerpt);
      setAuthor(post.author);
      setTagsStr(Array.isArray(post.tags) ? post.tags.join(', ') : '');
      setIsPublished(post.is_published);
      setFeaturedImage(post.featured_image ?? '');
    }
  }, [post]);

  const createMutation = useMutation({
    mutationFn: (payload: Record<string, unknown>) => adminApiClient.post('/blog/posts', payload),
    onSuccess: () => {
      onSaved();
    },
    onError: (error) => {
      console.error('Create post error:', error);
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
      console.error('Update post error:', error);
      alert(`Failed to update post: ${error.message}`);
    },
  });

  const tags = tagsStr
    .split(',')
    .map((t) => t.trim())
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
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    console.log('1. Form submitted, post:', post?.id);
    if (post) {
      console.log('2. Calling updateMutation.mutate');
      updateMutation.mutate(payload, {
        onSuccess: (data) => {
          console.log('3. Update success, data:', data);
          onSaved();
        },
        onError: (err, variables, context) => {
          console.error('3. Update error:', err, 'Response:', err.response?.data);
          alert(`Failed to update post: ${err.message}`);
        }
      });
    } else {
      console.log('2. Calling createMutation.mutate');
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
        </div>
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
      const res = await adminApiClient.get<BlogCategory[]>('/blog/categories');
      return Array.isArray(res) ? res : [];
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => adminApiClient.delete(`/blog/categories/${id}`),
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
      adminApiClient.post('/blog/categories', payload),
    onSuccess: () => onSaved(),
  });
  const updateMutation = useMutation({
    mutationFn: (payload: Record<string, unknown>) =>
      adminApiClient.put(`/blog/categories/${category!.id}`, payload),
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
