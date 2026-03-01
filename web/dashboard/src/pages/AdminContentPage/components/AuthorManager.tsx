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
import { Users, Plus, Edit, Trash2, Loader2, Sparkles } from 'lucide-react';
import { toast } from 'sonner';
import { blogApi, Author } from '@/api/blog';
import { contentAdminApi } from '@/api/content';
import slugify from 'slugify';

const AuthorManager = forwardRef<{ openCreateDialog: () => void }>((props, ref) => {
  const [authors, setAuthors] = useState<Author[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingAuthor, setEditingAuthor] = useState<Author | null>(null);
  const [generatingAi, setGeneratingAi] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    slug: '',
    bio: '',
    photo: { url: '', alt: '' },
    email: '',
    website: '',
    socialLinks: {},
    role: '',
    active: true,
  });

  useImperativeHandle(ref, () => ({
    openCreateDialog: () => {
      resetForm();
      setDialogOpen(true);
    }
  }));

  useEffect(() => {
    fetchAuthors();
  }, []);

  const fetchAuthors = async () => {
    try {
      setLoading(true);
      const result = await blogApi.getAuthors();
      setAuthors(Array.isArray(result) ? result : []);
    } catch (error) {
      console.error('Failed to fetch authors:', error);
      setAuthors([]);
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async () => {
    try {
      const authorData = {
        ...formData,
        photo: formData.photo.url ? formData.photo : undefined,
        socialLinks: Object.keys(formData.socialLinks).length > 0 ? formData.socialLinks : undefined,
      };

      await blogApi.createAuthor(authorData);
      setDialogOpen(false);
      resetForm();
      fetchAuthors();
    } catch (error) {
      console.error('Failed to create author:', error);
    }
  };

  const handleUpdate = async () => {
    if (!editingAuthor) return;

    try {
      const authorData = {
        ...formData,
        photo: formData.photo.url ? formData.photo : undefined,
        socialLinks: Object.keys(formData.socialLinks).length > 0 ? formData.socialLinks : undefined,
      };

      await blogApi.updateAuthor(editingAuthor.id, authorData);
      setDialogOpen(false);
      setEditingAuthor(null);
      resetForm();
      fetchAuthors();
    } catch (error) {
      console.error('Failed to update author:', error);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this author?')) return;

    try {
      await blogApi.deleteAuthor(id);
      fetchAuthors();
    } catch (error) {
      console.error('Failed to delete author:', error);
    }
  };

  const handleEdit = (author: Author) => {
    setEditingAuthor(author);
    setFormData({
      name: author.name,
      slug: author.slug,
      bio: author.bio || '',
      photo: author.photo || { url: '', alt: '' },
      email: author.email || '',
      website: author.website || '',
      socialLinks: author.socialLinks || {},
      role: author.role || '',
      active: author.active,
    });
    setDialogOpen(true);
  };

  const handleGenerateWithAi = async () => {
    if (!formData.name.trim()) {
      toast.error('Enter author name to generate bio');
      return;
    }
    setGeneratingAi(true);
    try {
      const res = await contentAdminApi.generateAuthorContent({
        name: formData.name,
        role: formData.role || undefined,
      });
      setFormData(prev => ({ ...prev, bio: res.bio || prev.bio }));
      if (res.bio) toast.success('Bio generated');
      else toast.info('No bio generated');
    } catch (e: any) {
      const msg = e?.response?.status === 503 ? 'Open Router not configured (OPENROUTER_API_KEY)' : 'Failed to generate';
      toast.error(msg);
    } finally {
      setGeneratingAi(false);
    }
  };

  const resetForm = () => {
    setFormData({
      name: '',
      slug: '',
      bio: '',
      photo: { url: '', alt: '' },
      email: '',
      website: '',
      socialLinks: {},
      role: '',
      active: true,
    });
    setEditingAuthor(null);
  };

  if (loading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-8">
          <Loader2 className="h-6 w-6 animate-spin mr-2" />
          Loading authors...
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>Authors</CardTitle>
          <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
            <DialogTrigger asChild>
              <Button onClick={resetForm}>
                <Plus className="h-4 w-4 mr-2" />
                Add Author
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
              <DialogHeader>
                <DialogTitle>
                  {editingAuthor ? 'Edit Author' : 'Create Author'}
                </DialogTitle>
              </DialogHeader>
              <div className="grid gap-4 py-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label htmlFor="name">Name *</Label>
                    <Input
                      id="name"
                      value={formData.name}
                      onChange={(e) => {
                        const name = e.target.value;
                        setFormData(prev => ({ ...prev, name, slug: slugify(name, { lower: true, strict: true, trim: true }) }));
                      }}
                      placeholder="Author name"
                    />
                  </div>
                  <div>
                    <Label htmlFor="slug">Slug *</Label>
                    <Input
                      id="slug"
                      value={formData.slug}
                      onChange={(e) => setFormData(prev => ({ ...prev, slug: e.target.value }))}
                      placeholder="author-slug"
                    />
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label htmlFor="email">Email</Label>
                    <Input
                      id="email"
                      type="email"
                      value={formData.email}
                      onChange={(e) => setFormData(prev => ({ ...prev, email: e.target.value }))}
                      placeholder="author@example.com"
                    />
                  </div>
                  <div>
                    <Label htmlFor="role">Role</Label>
                    <Input
                      id="role"
                      value={formData.role}
                      onChange={(e) => setFormData(prev => ({ ...prev, role: e.target.value }))}
                      placeholder="Senior Developer, Content Writer, etc."
                    />
                  </div>
                </div>
                <div>
                  <Label htmlFor="website">Website</Label>
                  <Input
                    id="website"
                    value={formData.website}
                    onChange={(e) => setFormData(prev => ({ ...prev, website: e.target.value }))}
                    placeholder="https://author-website.com"
                  />
                </div>
                <div>
                  <div className="flex items-center justify-between gap-2">
                    <Label htmlFor="bio">Bio</Label>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={handleGenerateWithAi}
                      disabled={generatingAi}
                      className="shrink-0"
                    >
                      {generatingAi ? (
                        <>
                          <Loader2 className="h-3.5 w-3.5 mr-1.5 animate-spin" />
                          Generating...
                        </>
                      ) : (
                        <>
                          <Sparkles className="h-3.5 w-3.5 mr-1.5" />
                          Generate with AI (Open Router)
                        </>
                      )}
                    </Button>
                  </div>
                  <Textarea
                    id="bio"
                    value={formData.bio}
                    onChange={(e) => setFormData(prev => ({ ...prev, bio: e.target.value }))}
                    placeholder="Author biography"
                    rows={3}
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label htmlFor="photoUrl">Photo URL</Label>
                    <Input
                      id="photoUrl"
                      value={formData.photo.url}
                      onChange={(e) => setFormData(prev => ({
                        ...prev,
                        photo: { ...prev.photo, url: e.target.value }
                      }))}
                      placeholder="https://example.com/photo.jpg"
                    />
                  </div>
                  <div>
                    <Label htmlFor="photoAlt">Photo Alt Text</Label>
                    <Input
                      id="photoAlt"
                      value={formData.photo.alt}
                      onChange={(e) => setFormData(prev => ({
                        ...prev,
                        photo: { ...prev.photo, alt: e.target.value }
                      }))}
                      placeholder="Alt text for photo"
                    />
                  </div>
                </div>
                <div className="flex items-center space-x-2">
                  <input
                    type="checkbox"
                    id="active"
                    checked={formData.active}
                    onChange={(e) => setFormData(prev => ({ ...prev, active: e.target.checked }))}
                  />
                  <Label htmlFor="active">Active</Label>
                </div>
              </div>
              <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={() => setDialogOpen(false)}>
                  Cancel
                </Button>
                <Button onClick={editingAuthor ? handleUpdate : handleCreate}>
                  {editingAuthor ? 'Update' : 'Create'}
                </Button>
              </div>
            </DialogContent>
          </Dialog>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Email</TableHead>
                <TableHead>Role</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {(authors ?? []).map((author) => (
                <TableRow key={author.id}>
                  <TableCell className="font-medium">
                    {author.name}
                  </TableCell>
                  <TableCell>{author.email || '-'}</TableCell>
                  <TableCell>{author.role || '-'}</TableCell>
                  <TableCell>
                    <Badge variant={author.active ? "default" : "secondary"}>
                      {author.active ? 'Active' : 'Inactive'}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex gap-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleEdit(author)}
                      >
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleDelete(author.id)}
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

AuthorManager.displayName = 'AuthorManager';

export default AuthorManager;
