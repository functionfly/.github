import type { CommunityCategory } from '@/api/community';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { useState } from 'react';

interface CreatePostDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  categories: CommunityCategory[];
  defaultCategory?: string;
  onSubmit: (data: { category_slug: string; title: string; body: string }) => Promise<void>;
}

export function CreatePostDialog({
  open,
  onOpenChange,
  categories,
  defaultCategory,
  onSubmit,
}: CreatePostDialogProps) {
  const [categorySlug, setCategorySlug] = useState(defaultCategory || categories[0]?.slug || 'general');
  const [title, setTitle] = useState('');
  const [body, setBody] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      await onSubmit({ category_slug: categorySlug, title: title.trim(), body: body.trim() });
      setTitle('');
      setBody('');
      onOpenChange(false);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Ask the community</DialogTitle>
          <DialogDescription>
            Share your question with other FunctionFly users. Be specific so others can help quickly.
          </DialogDescription>
        </DialogHeader>
        <form className="community-create-form" onSubmit={handleSubmit}>
          <label>
            Category
            <select value={categorySlug} onChange={(e) => setCategorySlug(e.target.value)} required>
              {categories.map((cat) => (
                <option key={cat.id} value={cat.slug}>
                  {cat.name}
                </option>
              ))}
            </select>
          </label>
          <label>
            Title
            <input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="What's your question?"
              maxLength={300}
              required
            />
          </label>
          <label>
            Details
            <textarea
              value={body}
              onChange={(e) => setBody(e.target.value)}
              placeholder="Include what you've tried, error messages, and relevant context."
              maxLength={10000}
              required
            />
          </label>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={submitting || !title.trim() || !body.trim()}>
              {submitting ? 'Posting…' : 'Post thread'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
