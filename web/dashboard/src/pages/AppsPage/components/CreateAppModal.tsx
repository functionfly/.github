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
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import type { App } from '@/types';
import { AlertCircle, CheckCircle2, Edit3, Info, Loader2, Plus } from 'lucide-react';
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
    <div className="flex items-center gap-2 px-3 py-2 rounded-[var(--radius-sm)]" style={{ background: 'rgba(143, 255, 208, 0.06)', border: '1px solid rgba(143, 255, 208, 0.2)' }}>
      <span className="text-xs" style={{ color: 'var(--text-faint)' }}>URL:</span>
      <code className="font-mono text-xs" style={{ color: 'var(--status-ok)' }}>/apps/{slug}</code>
    </div>
  );
}

function ValidationMessage({ message, type }: { message: string; type: 'error' | 'success' }) {
  return (
    <div
      className="flex items-center gap-2 text-xs"
      style={{ color: type === 'error' ? 'var(--status-revoked)' : 'var(--status-ok)' }}
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

function CharacterCounter({ current, max }: { current: number; max: number }) {
  const isNearLimit = current >= max * 0.8;
  return (
    <span
      className="text-xs tabular-nums"
      style={{ color: isNearLimit ? 'var(--status-pending)' : 'var(--text-faint)' }}
    >
      {current}/{max}
    </span>
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
  const [showSuccess, setShowSuccess] = useState(false);

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

  const handleEditSlug = () => {
    setSlugManuallyEdited(false);
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
      setShowSuccess(true);
      setTimeout(() => {
        toast.success(`App "${app.name}" created successfully`);
        setOpen(false);
        resetForm();
        onSuccess?.(app);
        navigate(`/apps/${encodeURIComponent(app.slug)}`);
      }, 600);
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
    setShowSuccess(false);
  };

  const handleOpenChange = (isOpen: boolean) => {
    if (!isOpen) resetForm();
    setOpen(isOpen);
  };

  const isSlugValid = slug.length >= 2 && /^[a-z0-9][a-z0-9-]*[a-z0-9]$/.test(slug);
  const isFormValid = name.trim().length >= 3 && isSlugValid;

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
      <DialogContent className="sm:max-w-md" style={{ background: 'var(--panel)', borderColor: 'var(--panel-edge)', borderRadius: 'var(--radius-lg)', boxShadow: 'var(--shadow-chamber)' }}>
        <TooltipProvider>
        <div>
          <DialogHeader>
            <DialogTitle className="text-xl" style={{ fontFamily: 'var(--font-display)' }}>Create New App</DialogTitle>
            <DialogDescription style={{ color: 'var(--text-dim)' }}>
              Apps organize your functions and manage multi-cloud deployments.
            </DialogDescription>
          </DialogHeader>

          <form onSubmit={handleSubmit} className="space-y-5 mt-2">
            {/* General error */}
            {generalError && (
              <div className="flex items-start gap-2.5 p-3.5 text-sm rounded-[var(--radius)]" style={{ color: 'var(--status-revoked)', background: 'rgba(255, 107, 107, 0.06)', border: '1px solid rgba(255, 107, 107, 0.2)' }}>
                <AlertCircle className="w-4 h-4 mt-0.5 flex-shrink-0" />
                <span>{generalError}</span>
              </div>
            )}

            {/* Fields card */}
            <div className="rounded-[var(--radius-lg)] p-4 space-y-5" style={{ border: '1px solid var(--panel-edge)', background: 'var(--panel-raised)' }}>
              {/* App Name */}
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor="app-name" className="text-sm font-medium" style={{ color: 'var(--text)' }}>
                    App Name <span style={{ color: 'var(--status-revoked)' }}>*</span>
                  </Label>
                  <CharacterCounter current={name.length} max={50} />
                </div>
                <Input
                  id="app-name"
                  placeholder="My Awesome App"
                  value={name}
                  onChange={(e) => handleNameChange(e.target.value)}
                  disabled={isSubmitting}
                  style={nameError ? { borderColor: 'var(--status-revoked)' } : undefined}
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
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-1.5">
                    <Label htmlFor="app-slug" className="text-sm font-medium" style={{ color: 'var(--text)' }}>
                      App Slug <span style={{ color: 'var(--status-revoked)' }}>*</span>
                    </Label>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Info className="w-3.5 h-3.5 cursor-help" style={{ color: 'var(--text-faint)' }} />
                      </TooltipTrigger>
                      <TooltipContent side="top" className="max-w-[200px]" style={{ background: 'var(--panel-raised)', borderColor: 'var(--steel)', borderRadius: 'var(--radius-sm)' }}>
                        {slugManuallyEdited
                          ? 'Slug is customized. Reset to auto-generate from name.'
                          : 'Slug is auto-generated from the app name.'}
                      </TooltipContent>
                    </Tooltip>
                  </div>
                  <CharacterCounter current={slug.length} max={63} />
                </div>
                <div className="relative">
                  <Input
                    id="app-slug"
                    placeholder="my-awesome-app"
                    value={slug}
                    onChange={(e) => handleSlugChange(e.target.value)}
                    disabled={isSubmitting}
                    className="font-mono pr-10"
                    style={{
                      ...(slugError ? { borderColor: 'var(--status-revoked)' } : {}),
                      ...(isSlugValid && !slugError ? { borderColor: 'rgba(143, 255, 208, 0.5)' } : {}),
                    }}
                    maxLength={63}
                    aria-describedby={slugError ? 'slug-error' : 'slug-hint'}
                  />
                  {isSlugValid && !slugError && (
                    <CheckCircle2 className="absolute right-10 top-1/2 -translate-y-1/2 w-4 h-4" style={{ color: 'var(--status-ok)' }} />
                  )}
                  {!slugManuallyEdited && slug ? (
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <button
                          type="button"
                          onClick={handleEditSlug}
                          className="absolute right-3 top-1/2 -translate-y-1/2 p-0.5 rounded-[var(--radius-sm)] transition-colors"
                          style={{ color: 'var(--text-faint)' }}
                          tabIndex={-1}
                        >
                          <Edit3 className="w-3.5 h-3.5" />
                        </button>
                      </TooltipTrigger>
                      <TooltipContent side="top" style={{ background: 'var(--panel-raised)', borderColor: 'var(--steel)', borderRadius: 'var(--radius-sm)' }}>Customize slug</TooltipContent>
                    </Tooltip>
                  ) : slugManuallyEdited && name ? (
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <button
                          type="button"
                          onClick={() => {
                            setSlugManuallyEdited(false);
                            setSlug(generateSlug(name));
                          }}
                          className="absolute right-3 top-1/2 -translate-y-1/2 p-0.5 rounded-[var(--radius-sm)] transition-colors"
                          style={{ color: 'var(--text-faint)' }}
                          tabIndex={-1}
                        >
                          <Edit3 className="w-3.5 h-3.5" />
                        </button>
                      </TooltipTrigger>
                      <TooltipContent side="top" style={{ background: 'var(--panel-raised)', borderColor: 'var(--steel)', borderRadius: 'var(--radius-sm)' }}>Reset to auto-generate</TooltipContent>
                    </Tooltip>
                  ) : null}
                </div>
                {slugError ? (
                  <ValidationMessage message={slugError} type="error" />
                ) : (
                  <SlugPreview slug={slug} />
                )}
              </div>
            </div>

            {/* Keyboard hint */}
            {isFormValid && !isSubmitting && (
              <p className="text-xs text-center" style={{ color: 'var(--text-faint)' }}>
                Press <kbd className="px-1.5 py-0.5 rounded-[var(--radius-sm)] text-xs font-mono" style={{ background: 'var(--panel-raised)', border: '1px solid var(--panel-edge)' }}>Enter</kbd> to submit
              </p>
            )}

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
                ) : showSuccess ? (
                  <>
                    <CheckCircle2 className="w-4 h-4 mr-2" />
                    Created!
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
        </div>
        </TooltipProvider>
      </DialogContent>
    </Dialog>
  );
}