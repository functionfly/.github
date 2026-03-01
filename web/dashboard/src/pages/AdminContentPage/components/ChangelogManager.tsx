'use client';

import { useState, useEffect, forwardRef, useImperativeHandle } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Label } from '@/components/ui/label';
import { Calendar, Plus, Edit, Trash2, Eye, EyeOff, Loader2, Sparkles } from 'lucide-react';
import { toast } from 'sonner';
import { contentAdminApi, ChangelogEntry } from '@/api/content';

const ChangelogManager = forwardRef<{ openCreateDialog: () => void }>((props, ref) => {
  const [entries, setEntries] = useState<ChangelogEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingEntry, setEditingEntry] = useState<ChangelogEntry | null>(null);
  const [generatingAi, setGeneratingAi] = useState(false);
  const [formData, setFormData] = useState({
    version: '',
    type: 'minor' as 'major' | 'minor' | 'patch',
    title: '',
    description: '',
    date: '',
    is_published: false,
  });

  useImperativeHandle(ref, () => ({
    openCreateDialog: () => {
      resetForm();
      setDialogOpen(true);
    }
  }));

  useEffect(() => {
    fetchEntries();
  }, []);

  const fetchEntries = async () => {
    try {
      setLoading(true);
      const result = await contentAdminApi.listChangelogEntries({ limit: 50 });
      setEntries(Array.isArray(result.entries) ? result.entries : []);
    } catch (error) {
      console.error('Failed to fetch changelog entries:', error);
      setEntries([]);
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async () => {
    try {
      await contentAdminApi.createChangelogEntry({
        ...formData,
        date: formData.date || new Date().toISOString(),
        changes: [],
      });
      setDialogOpen(false);
      resetForm();
      fetchEntries();
    } catch (error) {
      console.error('Failed to create changelog entry:', error);
    }
  };

  const handleUpdate = async () => {
    if (!editingEntry) return;

    try {
      await contentAdminApi.updateChangelogEntry(editingEntry.id, formData);
      setDialogOpen(false);
      setEditingEntry(null);
      resetForm();
      fetchEntries();
    } catch (error) {
      console.error('Failed to update changelog entry:', error);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this changelog entry?')) return;

    try {
      await contentAdminApi.deleteChangelogEntry(id);
      fetchEntries();
    } catch (error) {
      console.error('Failed to delete changelog entry:', error);
    }
  };

  const handleEdit = (entry: ChangelogEntry) => {
    setEditingEntry(entry);
    setFormData({
      version: entry.version,
      type: entry.type,
      title: entry.title,
      description: entry.description,
      date: entry.date,
      is_published: entry.is_published,
    });
    setDialogOpen(true);
  };

  const resetForm = () => {
    setFormData({
      version: '',
      type: 'minor',
      title: '',
      description: '',
      date: '',
      is_published: false,
    });
    setEditingEntry(null);
  };

  const handleGenerateWithAi = async () => {
    setGeneratingAi(true);
    try {
      const res = await contentAdminApi.generateChangelogContent({
        version: formData.version || undefined,
        type: formData.type,
        topic: formData.title || formData.description || undefined,
      });
      setFormData(prev => ({
        ...prev,
        title: res.title || prev.title,
        description: res.description || prev.description,
      }));
      if (res.title || res.description) toast.success('Title and description generated');
      else toast.info('No content generated');
    } catch (e: any) {
      const msg = e?.response?.status === 503 ? 'Open Router not configured (OPENROUTER_API_KEY)' : 'Failed to generate';
      toast.error(msg);
    } finally {
      setGeneratingAi(false);
    }
  };

  const getBadgeVariant = (type: string) => {
    switch (type) {
      case 'major':
        return 'default';
      case 'minor':
        return 'secondary';
      case 'patch':
        return 'outline';
      default:
        return 'outline';
    }
  };

  if (loading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-8">
          <Loader2 className="h-6 w-6 animate-spin mr-2" />
          Loading changelog entries...
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>Changelog Entries</CardTitle>
          <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
            <DialogTrigger asChild>
              <Button onClick={resetForm}>
                <Plus className="h-4 w-4 mr-2" />
                Add Entry
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-2xl">
              <DialogHeader>
                <DialogTitle>
                  {editingEntry ? 'Edit Changelog Entry' : 'Create Changelog Entry'}
                </DialogTitle>
              </DialogHeader>
              <div className="grid gap-4 py-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label htmlFor="version">Version</Label>
                    <Input
                      id="version"
                      value={formData.version}
                      onChange={(e) => setFormData(prev => ({ ...prev, version: e.target.value }))}
                      placeholder="v1.0.0"
                    />
                  </div>
                  <div>
                    <Label htmlFor="type">Type</Label>
                    <Select value={formData.type} onValueChange={(value: 'major' | 'minor' | 'patch') =>
                      setFormData(prev => ({ ...prev, type: value }))
                    }>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="major">Major</SelectItem>
                        <SelectItem value="minor">Minor</SelectItem>
                        <SelectItem value="patch">Patch</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <div>
                  <Label htmlFor="title">Title</Label>
                  <Input
                    id="title"
                    value={formData.title}
                    onChange={(e) => setFormData(prev => ({ ...prev, title: e.target.value }))}
                    placeholder="Release title"
                  />
                </div>
                <div>
                  <div className="flex items-center justify-between gap-2">
                    <Label htmlFor="description">Description</Label>
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
                    id="description"
                    value={formData.description}
                    onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                    placeholder="Release description"
                    rows={3}
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label htmlFor="date">Date</Label>
                    <Input
                      id="date"
                      type="date"
                      value={formData.date ? new Date(formData.date).toISOString().split('T')[0] : ''}
                      onChange={(e) => setFormData(prev => ({
                        ...prev,
                        date: e.target.value ? new Date(e.target.value).toISOString() : ''
                      }))}
                    />
                  </div>
                  <div className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      id="is_published"
                      checked={formData.is_published}
                      onChange={(e) => setFormData(prev => ({ ...prev, is_published: e.target.checked }))}
                    />
                    <Label htmlFor="is_published">Published</Label>
                  </div>
                </div>
              </div>
              <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={() => setDialogOpen(false)}>
                  Cancel
                </Button>
                <Button onClick={editingEntry ? handleUpdate : handleCreate}>
                  {editingEntry ? 'Update' : 'Create'}
                </Button>
              </div>
            </DialogContent>
          </Dialog>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Version</TableHead>
                <TableHead>Title</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Date</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {(entries ?? []).map((entry) => (
                <TableRow key={entry.id}>
                  <TableCell className="font-medium">{entry.version}</TableCell>
                  <TableCell>{entry.title}</TableCell>
                  <TableCell>
                    <Badge variant={getBadgeVariant(entry.type)}>
                      {entry.type.toUpperCase()}
                    </Badge>
                  </TableCell>
                  <TableCell>{new Date(entry.date).toLocaleDateString()}</TableCell>
                  <TableCell>
                    {entry.is_published ? (
                      <Badge variant="default">
                        <Eye className="h-3 w-3 mr-1" />
                        Published
                      </Badge>
                    ) : (
                      <Badge variant="secondary">
                        <EyeOff className="h-3 w-3 mr-1" />
                        Draft
                      </Badge>
                    )}
                  </TableCell>
                  <TableCell>
                    <div className="flex gap-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleEdit(entry)}
                      >
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleDelete(entry.id)}
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

ChangelogManager.displayName = 'ChangelogManager';

export default ChangelogManager;
