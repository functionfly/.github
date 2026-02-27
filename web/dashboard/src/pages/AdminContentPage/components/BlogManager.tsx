'use client';

import { useState, useEffect, forwardRef, useImperativeHandle } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { BookOpen, Plus, Edit, Trash2, Eye, EyeOff, Loader2, Calendar } from 'lucide-react';
import { blogApi, BlogPost, ContentStatus, Author, Category } from '@/api/blog';

const BlogManager = forwardRef<{ openCreateDialog: () => void }>((props, ref) => {
  const [posts, setPosts] = useState<BlogPost[]>([]);
  const [authors, setAuthors] = useState<Author[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingPost, setEditingPost] = useState<BlogPost | null>(null);
  const [formData, setFormData] = useState({
    title: '',
    slug: '',
    description: '',
    body: '',
    authorId: '',
    categoryId: '',
    tags: [] as string[],
    heroImage: { url: '', alt: '', caption: '' },
    status: ContentStatus.DRAFT,
    publishedAt: '',
    scheduledAt: '',
    seoTitle: '',
    seoDescription: '',
    keywords: [] as string[],
    canonicalUrl: '',
    ogImage: { url: '', alt: '' },
    campaign: '',
  });

  useImperativeHandle(ref, () => ({
    openCreateDialog: () => {
      resetForm();
      setDialogOpen(true);
    }
  }));

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const [postsResult, authorsResult, categoriesResult] = await Promise.all([
        blogApi.getPosts({ limit: 50 }),
        blogApi.getAuthors(),
        blogApi.getCategories()
      ]);
      setPosts(Array.isArray(postsResult.data) ? postsResult.data : []);
      setAuthors(Array.isArray(authorsResult) ? authorsResult : []);
      setCategories(Array.isArray(categoriesResult) ? categoriesResult : []);
    } catch (error) {
      console.error('Failed to fetch data:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchPosts = async () => {
    try {
      const result = await blogApi.getPosts({ limit: 50 });
      setPosts(result.data);
    } catch (error) {
      console.error('Failed to fetch blog posts:', error);
    }
  };

  const handleCreate = async () => {
    try {
      const postData = {
        ...formData,
        tags: formData.tags.filter(tag => tag.trim() !== ''),
        heroImage: formData.heroImage.url ? formData.heroImage : undefined,
        ogImage: formData.ogImage.url ? formData.ogImage : undefined,
        keywords: formData.keywords.filter(k => k.trim() !== ''),
        publishedAt: formData.publishedAt || (formData.status === ContentStatus.PUBLISHED ? new Date().toISOString() : undefined),
        authorId: formData.authorId || undefined,
        categoryId: formData.categoryId || undefined,
      };

      await blogApi.createPost(postData);
      setDialogOpen(false);
      resetForm();
      fetchPosts();
    } catch (error) {
      console.error('Failed to create blog post:', error);
    }
  };

  const handleUpdate = async () => {
    if (!editingPost) return;

    try {
      const postData = {
        ...formData,
        tags: formData.tags.filter(tag => tag.trim() !== ''),
        heroImage: formData.heroImage.url ? formData.heroImage : undefined,
        ogImage: formData.ogImage.url ? formData.ogImage : undefined,
        keywords: formData.keywords.filter(k => k.trim() !== ''),
        authorId: formData.authorId || undefined,
        categoryId: formData.categoryId || undefined,
      };

      await blogApi.updatePost(editingPost.id, postData);
      setDialogOpen(false);
      setEditingPost(null);
      resetForm();
      fetchPosts();
    } catch (error) {
      console.error('Failed to update blog post:', error);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this blog post?')) return;

    try {
      await blogApi.deletePost(id);
      fetchPosts();
    } catch (error) {
      console.error('Failed to delete blog post:', error);
    }
  };

  const handleEdit = (post: BlogPost) => {
    setEditingPost(post);
    setFormData({
      title: post.title,
      slug: post.slug,
      description: post.description,
      body: typeof post.body === 'string' ? post.body : JSON.stringify(post.body, null, 2),
      authorId: post.authorId || '',
      categoryId: post.categoryId || '',
      tags: [...(post.tags || [])],
      heroImage: post.heroImage || { url: '', alt: '', caption: '' },
      status: post.status,
      publishedAt: post.publishedAt || '',
      scheduledAt: post.scheduledAt || '',
      seoTitle: post.seoTitle || '',
      seoDescription: post.seoDescription || '',
      keywords: [...(post.keywords || [])],
      canonicalUrl: post.canonicalUrl || '',
      ogImage: post.ogImage || { url: '', alt: '' },
      campaign: post.campaign || '',
    });
    setDialogOpen(true);
  };

  const resetForm = () => {
    setFormData({
      title: '',
      slug: '',
      description: '',
      body: '',
      authorId: '',
      categoryId: '',
      tags: [],
      heroImage: { url: '', alt: '', caption: '' },
      status: ContentStatus.DRAFT,
      publishedAt: '',
      scheduledAt: '',
      seoTitle: '',
      seoDescription: '',
      keywords: [],
      canonicalUrl: '',
      ogImage: { url: '', alt: '' },
      campaign: '',
    });
    setEditingPost(null);
  };

  const addTag = () => {
    setFormData(prev => ({
      ...prev,
      tags: [...prev.tags, '']
    }));
  };

  const updateTag = (index: number, value: string) => {
    setFormData(prev => ({
      ...prev,
      tags: prev.tags.map((tag, i) => i === index ? value : tag)
    }));
  };

  const removeTag = (index: number) => {
    setFormData(prev => ({
      ...prev,
      tags: prev.tags.filter((_, i) => i !== index)
    }));
  };

  const addKeyword = () => {
    setFormData(prev => ({
      ...prev,
      keywords: [...prev.keywords, '']
    }));
  };

  const updateKeyword = (index: number, value: string) => {
    setFormData(prev => ({
      ...prev,
      keywords: prev.keywords.map((keyword, i) => i === index ? value : keyword)
    }));
  };

  const removeKeyword = (index: number) => {
    setFormData(prev => ({
      ...prev,
      keywords: prev.keywords.filter((_, i) => i !== index)
    }));
  };

  if (loading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-8">
          <Loader2 className="h-6 w-6 animate-spin mr-2" />
          Loading blog posts...
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>Blog Posts</CardTitle>
          <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
            <DialogTrigger asChild>
              <Button onClick={resetForm}>
                <Plus className="h-4 w-4 mr-2" />
                Add Post
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
              <DialogHeader>
                <DialogTitle>
                  {editingPost ? 'Edit Blog Post' : 'Create Blog Post'}
                </DialogTitle>
              </DialogHeader>
              <div className="grid gap-4 py-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label htmlFor="title">Title</Label>
                    <Input
                      id="title"
                      value={formData.title}
                      onChange={(e) => setFormData(prev => ({ ...prev, title: e.target.value }))}
                      placeholder="Post title"
                    />
                  </div>
                  <div>
                    <Label htmlFor="slug">Slug</Label>
                    <Input
                      id="slug"
                      value={formData.slug}
                      onChange={(e) => setFormData(prev => ({ ...prev, slug: e.target.value }))}
                      placeholder="post-slug"
                    />
                  </div>
                </div>
                <div>
                  <Label htmlFor="description">Description</Label>
                  <Textarea
                    id="description"
                    value={formData.description}
                    onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                    placeholder="Brief description (for SEO and previews)"
                    rows={2}
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label htmlFor="authorId">Author</Label>
                    <Select value={formData.authorId} onValueChange={(value) => setFormData(prev => ({ ...prev, authorId: value }))}>
                      <SelectTrigger>
                        <SelectValue placeholder="Select author" />
                      </SelectTrigger>
                      <SelectContent>
                        {authors.map((author) => (
                          <SelectItem key={author.id} value={author.id}>
                            {author.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div>
                    <Label htmlFor="categoryId">Category</Label>
                    <Select value={formData.categoryId} onValueChange={(value) => setFormData(prev => ({ ...prev, categoryId: value }))}>
                      <SelectTrigger>
                        <SelectValue placeholder="Select category" />
                      </SelectTrigger>
                      <SelectContent>
                        {categories.map((category) => (
                          <SelectItem key={category.id} value={category.id}>
                            {category.title}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <div>
                  <Label htmlFor="status">Status</Label>
                  <Select value={formData.status} onValueChange={(value: ContentStatus) => setFormData(prev => ({ ...prev, status: value }))}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value={ContentStatus.DRAFT}>Draft</SelectItem>
                      <SelectItem value={ContentStatus.IN_REVIEW}>In Review</SelectItem>
                      <SelectItem value={ContentStatus.APPROVED}>Approved</SelectItem>
                      <SelectItem value={ContentStatus.SCHEDULED}>Scheduled</SelectItem>
                      <SelectItem value={ContentStatus.PUBLISHED}>Published</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div>
                  <Label htmlFor="body">Content (JSON/Markdown)</Label>
                  <Textarea
                    id="body"
                    value={formData.body}
                    onChange={(e) => setFormData(prev => ({ ...prev, body: e.target.value }))}
                    placeholder="Post content (JSON for structured content or Markdown)"
                    rows={8}
                  />
                </div>
                <div>
                  <Label>Tags</Label>
                  <div className="space-y-2">
                    {formData.tags.map((tag, index) => (
                      <div key={index} className="flex gap-2">
                        <Input
                          value={tag}
                          onChange={(e) => updateTag(index, e.target.value)}
                          placeholder="Tag name"
                        />
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={() => removeTag(index)}
                        >
                          Remove
                        </Button>
                      </div>
                    ))}
                    <Button type="button" variant="outline" size="sm" onClick={addTag}>
                      Add Tag
                    </Button>
                  </div>
                </div>

                {/* Hero Image */}
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label htmlFor="heroImageUrl">Hero Image URL</Label>
                    <Input
                      id="heroImageUrl"
                      value={formData.heroImage.url}
                      onChange={(e) => setFormData(prev => ({
                        ...prev,
                        heroImage: { ...prev.heroImage, url: e.target.value }
                      }))}
                      placeholder="https://example.com/image.jpg"
                    />
                  </div>
                  <div>
                    <Label htmlFor="heroImageAlt">Hero Image Alt Text</Label>
                    <Input
                      id="heroImageAlt"
                      value={formData.heroImage.alt}
                      onChange={(e) => setFormData(prev => ({
                        ...prev,
                        heroImage: { ...prev.heroImage, alt: e.target.value }
                      }))}
                      placeholder="Alt text for accessibility"
                    />
                  </div>
                </div>

                {/* Publishing */}
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label htmlFor="publishedAt">Published At</Label>
                    <Input
                      id="publishedAt"
                      type="datetime-local"
                      value={formData.publishedAt ? new Date(formData.publishedAt).toISOString().slice(0, 16) : ''}
                      onChange={(e) => setFormData(prev => ({
                        ...prev,
                        publishedAt: e.target.value ? new Date(e.target.value).toISOString() : ''
                      }))}
                    />
                  </div>
                  <div>
                    <Label htmlFor="scheduledAt">Scheduled At</Label>
                    <Input
                      id="scheduledAt"
                      type="datetime-local"
                      value={formData.scheduledAt ? new Date(formData.scheduledAt).toISOString().slice(0, 16) : ''}
                      onChange={(e) => setFormData(prev => ({
                        ...prev,
                        scheduledAt: e.target.value ? new Date(e.target.value).toISOString() : ''
                      }))}
                    />
                  </div>
                </div>

                {/* SEO Fields */}
                <div className="space-y-4 border-t pt-4">
                  <h4 className="font-medium">SEO Settings</h4>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label htmlFor="seoTitle">SEO Title</Label>
                      <Input
                        id="seoTitle"
                        value={formData.seoTitle}
                        onChange={(e) => setFormData(prev => ({ ...prev, seoTitle: e.target.value }))}
                        placeholder="Custom SEO title"
                      />
                    </div>
                    <div>
                      <Label htmlFor="canonicalUrl">Canonical URL</Label>
                      <Input
                        id="canonicalUrl"
                        value={formData.canonicalUrl}
                        onChange={(e) => setFormData(prev => ({ ...prev, canonicalUrl: e.target.value }))}
                        placeholder="https://functionfly.com/blog/post"
                      />
                    </div>
                  </div>
                  <div>
                    <Label htmlFor="seoDescription">SEO Description</Label>
                    <Textarea
                      id="seoDescription"
                      value={formData.seoDescription}
                      onChange={(e) => setFormData(prev => ({ ...prev, seoDescription: e.target.value }))}
                      placeholder="SEO meta description"
                      rows={2}
                    />
                  </div>
                  <div>
                    <Label>Keywords</Label>
                    <div className="space-y-2">
                      {formData.keywords.map((keyword, index) => (
                        <div key={index} className="flex gap-2">
                          <Input
                            value={keyword}
                            onChange={(e) => updateKeyword(index, e.target.value)}
                            placeholder="SEO keyword"
                          />
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={() => removeKeyword(index)}
                          >
                            Remove
                          </Button>
                        </div>
                      ))}
                      <Button type="button" variant="outline" size="sm" onClick={addKeyword}>
                        Add Keyword
                      </Button>
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label htmlFor="ogImageUrl">OG Image URL</Label>
                      <Input
                        id="ogImageUrl"
                        value={formData.ogImage.url}
                        onChange={(e) => setFormData(prev => ({
                          ...prev,
                          ogImage: { ...prev.ogImage, url: e.target.value }
                        }))}
                        placeholder="https://example.com/og-image.jpg"
                      />
                    </div>
                    <div>
                      <Label htmlFor="campaign">Campaign</Label>
                      <Input
                        id="campaign"
                        value={formData.campaign}
                        onChange={(e) => setFormData(prev => ({ ...prev, campaign: e.target.value }))}
                        placeholder="campaign-name"
                      />
                    </div>
                  </div>
                </div>
              </div>
              <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={() => setDialogOpen(false)}>
                  Cancel
                </Button>
                <Button onClick={editingPost ? handleUpdate : handleCreate}>
                  {editingPost ? 'Update' : 'Create'}
                </Button>
              </div>
            </DialogContent>
          </Dialog>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Title</TableHead>
                <TableHead>Author</TableHead>
                <TableHead>Category</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Published</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {posts.map((post) => (
                <TableRow key={post.id}>
                  <TableCell className="font-medium max-w-xs truncate" title={post.title}>
                    {post.title}
                  </TableCell>
                  <TableCell>{post.author?.name || 'Unknown'}</TableCell>
                  <TableCell>{post.category?.title || 'Uncategorized'}</TableCell>
                  <TableCell>
                    <Badge variant={
                      post.status === ContentStatus.PUBLISHED ? "default" :
                      post.status === ContentStatus.DRAFT ? "secondary" :
                      post.status === ContentStatus.IN_REVIEW ? "outline" :
                      post.status === ContentStatus.APPROVED ? "default" :
                      "secondary"
                    }>
                      {post.status === ContentStatus.PUBLISHED && <Eye className="h-3 w-3 mr-1" />}
                      {post.status === ContentStatus.DRAFT && <EyeOff className="h-3 w-3 mr-1" />}
                      {post.status.replace('_', ' ').toUpperCase()}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    {post.publishedAt ? (
                      new Date(post.publishedAt).toLocaleDateString()
                    ) : (
                      <span className="text-muted-foreground">Not published</span>
                    )}
                  </TableCell>
                  <TableCell>
                    <div className="flex gap-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleEdit(post)}
                      >
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleDelete(post.id)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
});

BlogManager.displayName = 'BlogManager';

export default BlogManager;