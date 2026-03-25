import { appsApi } from '@/api/apps';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { cn } from '@/lib/utils';
import type { App } from '@/types';
import { AlertCircle, CheckCircle2, Info, Loader2, Plus } from 'lucide-react';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';

interface CreateAppModalProps {
  onSuccess?: (app: App) => void;
  trigger?: React.ReactNode;
}

function SlugPreview({ slug }: { slug: string }) {
  if (!slug) return null;
  return (
    <div className="flex items-center gap-1.5 text-xs text-muted-foreground mt-1.5">
      <Info className="w-3 h-3 flex-shrink-0" />
      <span>
        URL: <code className="font-mono text-foreground/80">/apps/{slug}</code>
      </span>
    </div>
  );
}

function ValidationMessage({ message, type }: { message: string; type: 'error' | 'success' }) {
  return (
    <div
      className={cn(
        'flex items-center gap-2 text-xs mt-1.5',
        type === 'error' ? 'text-destructive' : 'text-emerald-500'
      )}
    >
      {type === 'error' ? (
        <AlertCircle className="w-3 h-3 flex-shrink-0" />
      ) : (
        <CheckCircle2 className="w-3 h-3 flex-shrink-0" />
      )}
      {message}
    </div>
  );
}

export function CreateAppModal({ onSuccess, trigger }: CreateAppModalProps) {
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);
  const [name, setName] = useState('');
  const [slug, setSlug] = useState('');
  const [slugManuallyEdited, setSlugManuallyEdited] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [nameError, setNameError] = useState<string | null>(null);
  const [slugError, setSlugError] = useState<string | null>(null);
  const [generalError, setGeneralError] = useState<string | null>(null);

  const generateSlug = (value: string) =>
    value
      .toLowerCase()
      .trim()
      .replace(/\s+/g, '-')
      .replace(/[^a-z0-9-]/g, '')
      .replace(/-+/g, '-')
      .replace(/^-|-$/g, '');

  const handleNameChange = (value: string) => {
    setName(value);
    setNameError(null);
    setGeneralError(null);
    if (!slugManuallyEdited) {
      setSlug(generateSlug(value));
    }
  };

  const handleSlugChange = (value: string) => {
    const cleaned = value.toLowerCase().replace(/[^a-z0-9-]/g, '');
    setSlug(cleaned);
    setSlugManuallyEdited(true);
    setSlugError(null);
    setGeneralError(null);
  };

  const validateForm = (): boolean => {
    let valid = true;

    if (!name.trim()) {
      setNameError('App name is required');
      valid = false;
    } else if (name.trim().length < 3) {
      setNameError('App name must be at least 3 characters');
      valid = false;
    } else if (name.trim().length > 50) {
      setNameError('App name must be 50 characters or less');
      valid = false;
    }

    if (!slug.trim()) {
      setSlugError('App slug is required');
      valid = false;
    } else if (!/^[a-z0-9][a-z0-9-]*[a-z0-9]$/.test(slug) && slug.length > 1) {
      setSlugError('Slug must start and end with a letter or number');
      valid = false;
    } else if (slug.length < 2) {
      setSlugError('Slug must be at least 2 characters');
      valid = false;
    } else if (slug.length > 63) {
      setSlugError('Slug must be 63 characters or less');
      valid = false;
    }

    return valid;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setGeneralError(null);

    if (!validateForm()) return;

    setIsSubmitting(true);

    try {
      const app = await appsApi.create({ name: name.trim(), slug });
      toast.success(`App "${app.name}" created successfully`);
      setOpen(false);
      resetForm();
      onSuccess?.(app);
      navigate(`/apps/${encodeURIComponent(app.slug)}`);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to create app';
      setGeneralError(message);
    } finally {
      setIsSubmitting(false);
    }
  };

  const resetForm = () => {
    setName('');
    setSlug('');
    setSlugManuallyEdited(false);
    setNameError(null);
    setSlugError(null);
    setGeneralError(null);
  };

  const handleOpenChange = (isOpen: boolean) => {
    if (!isOpen) resetForm();
    setOpen(isOpen);
  };

  const isSlugValid = slug.length >= 2 && /^[a-z0-9][a-z0-9-]*[a-z0-9]$/.test(slug);

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        {trigger ?? (
          <Button className="gap-2">
            <Plus className="w-4 h-4" />
            Create App
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="text-xl">Create New App</DialogTitle>
          <DialogDescription>
            Apps organize your functions and manage multi-cloud deployments.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-5 mt-2">
          {/* General error */}
          {generalError && (
            <div className="flex items-start gap-2.5 p-3.5 text-sm text-destructive bg-destructive/10 border border-destructive/20 rounded-lg">
              <AlertCircle className="w-4 h-4 mt-0.5 flex-shrink-0" />
              <span>{generalError}</span>
            </div>
          )}

          {/* App Name */}
          <div className="space-y-1.5">
            <Label htmlFor="app-name" className="text-sm font-medium">
              App Name <span className="text-destructive">*</span>
            </Label>
            <Input
              id="app-name"
              placeholder="My Awesome App"
              value={name}
              onChange={(e) => handleNameChange(e.target.value)}
              disabled={isSubmitting}
              className={cn(
                'transition-colors',
                nameError && 'border-destructive focus-visible:ring-destructive/30'
              )}
              autoFocus
              maxLength={50}
              aria-describedby={nameError ? 'name-error' : undefined}
            />
            {nameError ? (
              <ValidationMessage message={nameError} type="error" />
            ) : name.trim().length >= 3 ? (
              <ValidationMessage message="Looks good!" type="success" />
            ) : null}
          </div>

          {/* App Slug */}
          <div className="space-y-1.5">
            <Label htmlFor="app-slug" className="text-sm font-medium">
              App Slug <span className="text-destructive">*</span>
            </Label>
            <div className="relative">
              <Input
                id="app-slug"
                placeholder="my-awesome-app"
                value={slug}
                onChange={(e) => handleSlugChange(e.target.value)}
                disabled={isSubmitting}
                className={cn(
                  'font-mono transition-colors',
                  slugError && 'border-destructive focus-visible:ring-destructive/30',
                  isSlugValid &&
                    !slugError &&
                    'border-emerald-500/50 focus-visible:ring-emerald-500/30'
                )}
                maxLength={63}
                aria-describedby={slugError ? 'slug-error' : 'slug-hint'}
              />
              {isSlugValid && !slugError && (
                <CheckCircle2 className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-emerald-500" />
              )}
            </div>
            {slugError ? (
              <ValidationMessage message={slugError} type="error" />
            ) : (
              <SlugPreview slug={slug} />
            )}
          </div>

          {/* Actions */}
          <div className="flex justify-end gap-3 pt-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => setOpen(false)}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={isSubmitting || !name.trim() || !slug.trim()}
              className="min-w-[120px]"
            >
              {isSubmitting ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Creating...
                </>
              ) : (
                <>
                  <Plus className="w-4 h-4 mr-2" />
                  Create App
                </>
              )}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
