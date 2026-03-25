import { appsApi } from '@/api/apps';
import { publicAppUrl } from '@/lib/constants';
import type { App, CreateAppRequest } from '@/types';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useCallback, useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';

export type Environment = 'development' | 'staging' | 'production';
export type Visibility = 'public' | 'private';

export interface AppFormData {
  name: string;
  slug: string;
  description: string;
  iconColor: string;
  iconEmoji: string;
  visibility: Visibility;
  environment: Environment;
  tags: string[];
}

export interface FormErrors {
  name?: string;
  slug?: string;
  description?: string;
  general?: string;
}

export interface UseCreateAppReturn {
  // Form data
  formData: AppFormData;
  errors: FormErrors;
  isDirty: boolean;
  isSubmitting: boolean;
  isValid: boolean;
  lastSaved: Date | null;

  // Field handlers
  setName: (name: string) => void;
  setSlug: (slug: string) => void;
  setDescription: (description: string) => void;
  setIconColor: (color: string) => void;
  setIconEmoji: (emoji: string) => void;
  setVisibility: (visibility: Visibility) => void;
  setEnvironment: (environment: Environment) => void;
  addTag: (tag: string) => void;
  removeTag: (tag: string) => void;

  // Actions
  handleSubmit: () => Promise<void>;
  handleCancel: () => void;
  resetForm: () => void;

  // Derived state
  slugPreview: string;
  nameValidation: { valid: boolean; message?: string };
  slugValidation: { valid: boolean; message?: string };
  descriptionValidation: { valid: boolean; message?: string };
}

// Icon color presets with hex values for inline styles
export const ICON_COLORS = [
  { name: 'indigo', value: 'indigo', from: '#6366f1', to: '#8b5cf6' },
  { name: 'blue', value: 'blue', from: '#3b82f6', to: '#06b6d4' },
  { name: 'emerald', value: 'emerald', from: '#10b981', to: '#14b8a6' },
  { name: 'amber', value: 'amber', from: '#f59e0b', to: '#f97316' },
  { name: 'rose', value: 'rose', from: '#f43f5e', to: '#ec4899' },
  { name: 'violet', value: 'violet', from: '#8b5cf6', to: '#a855f7' },
  { name: 'cyan', value: 'cyan', from: '#06b6d4', to: '#0ea5e9' },
  { name: 'slate', value: 'slate', from: '#64748b', to: '#6b7280' },
] as const;

// Icon options with both emoji and Iconify support
export const ICON_OPTIONS = [
  { emoji: '🚀', icon: 'fluent:rocket-24-filled', label: 'Rocket' },
  { emoji: '⚡', icon: 'fluent:flash-24-filled', label: 'Flash' },
  { emoji: '🔥', icon: 'fluent:fire-24-filled', label: 'Fire' },
  { emoji: '💎', icon: 'fluent:gem-24-filled', label: 'Gem' },
  { emoji: '🎯', icon: 'fluent:target-24-filled', label: 'Target' },
  { emoji: '🔧', icon: 'fluent:wrench-24-filled', label: 'Wrench' },
  { emoji: '📦', icon: 'fluent:box-24-filled', label: 'Box' },
  { emoji: '🌟', icon: 'fluent:star-24-filled', label: 'Star' },
  { emoji: '🎨', icon: 'fluent:paint-brush-24-filled', label: 'Paint' },
  { emoji: '🔐', icon: 'fluent:lock-shield-24-filled', label: 'Lock' },
  { emoji: '🌐', icon: 'fluent:globe-24-filled', label: 'Globe' },
  { emoji: '⚙️', icon: 'fluent:settings-24-filled', label: 'Settings' },
  { emoji: '📊', icon: 'fluent:chart-24-filled', label: 'Chart' },
  { emoji: '🔔', icon: 'fluent:alert-24-filled', label: 'Alert' },
  { emoji: '☁️', icon: 'fluent:cloud-24-filled', label: 'Cloud' },
  { emoji: '🛡️', icon: 'fluent:shield-24-filled', label: 'Shield' },
] as const;

// Legacy export for backward compatibility
export const ICON_EMOJIS = ICON_OPTIONS.map((opt) => opt.emoji);

const MAX_NAME_LENGTH = 50;
const MAX_SLUG_LENGTH = 63;
const MAX_DESCRIPTION_LENGTH = 500;
const MAX_TAGS = 10;

const generateSlug = (name: string): string => {
  return name
    .toLowerCase()
    .trim()
    .replace(/[^\w\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
    .slice(0, MAX_SLUG_LENGTH);
};

const validateSlug = (slug: string): { valid: boolean; message?: string } => {
  if (!slug) {
    return { valid: false, message: 'App slug is required' };
  }
  if (slug.length < 2) {
    return { valid: false, message: 'Slug must be at least 2 characters' };
  }
  if (slug.length > MAX_SLUG_LENGTH) {
    return { valid: false, message: `Slug must be ${MAX_SLUG_LENGTH} characters or less` };
  }
  if (!/^[a-z0-9]/.test(slug)) {
    return { valid: false, message: 'Slug must start with a letter or number' };
  }
  if (!/[a-z0-9]$/.test(slug)) {
    return { valid: false, message: 'Slug must end with a letter or number' };
  }
  if (!/^[a-z0-9-]+$/.test(slug)) {
    return { valid: false, message: 'Only lowercase letters, numbers, and hyphens allowed' };
  }
  if (/--/.test(slug)) {
    return { valid: false, message: 'Consecutive hyphens are not allowed' };
  }
  return { valid: true };
};

const validateName = (name: string): { valid: boolean; message?: string } => {
  if (!name.trim()) {
    return { valid: false, message: 'App name is required' };
  }
  if (name.trim().length < 3) {
    return { valid: false, message: 'App name must be at least 3 characters' };
  }
  if (name.trim().length > MAX_NAME_LENGTH) {
    return { valid: false, message: `App name must be ${MAX_NAME_LENGTH} characters or less` };
  }
  return { valid: true };
};

const validateDescription = (description: string): { valid: boolean; message?: string } => {
  if (description.length > MAX_DESCRIPTION_LENGTH) {
    return {
      valid: false,
      message: `Description must be ${MAX_DESCRIPTION_LENGTH} characters or less`,
    };
  }
  return { valid: true };
};

const initialFormData: AppFormData = {
  name: '',
  slug: '',
  description: '',
  iconColor: ICON_COLORS[0].value,
  iconEmoji: ICON_EMOJIS[0],
  visibility: 'private',
  environment: 'development',
  tags: [],
};

export function useCreateApp(): UseCreateAppReturn {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [formData, setFormData] = useState<AppFormData>(initialFormData);
  const [errors, setErrors] = useState<FormErrors>({});
  const [isDirty, setIsDirty] = useState(false);
  const [lastSaved, setLastSaved] = useState<Date | null>(null);
  const [slugManuallyEdited, setSlugManuallyEdited] = useState(false);
  const autoSaveTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Auto-save to localStorage
  useEffect(() => {
    if (isDirty && formData.name) {
      if (autoSaveTimeoutRef.current) {
        clearTimeout(autoSaveTimeoutRef.current);
      }
      autoSaveTimeoutRef.current = setTimeout(() => {
        localStorage.setItem('app_create_draft', JSON.stringify(formData));
        setLastSaved(new Date());
      }, 2000);
    }
    return () => {
      if (autoSaveTimeoutRef.current) {
        clearTimeout(autoSaveTimeoutRef.current);
      }
    };
  }, [formData, isDirty]);

  // Load draft on mount
  useEffect(() => {
    const draft = localStorage.getItem('app_create_draft');
    if (draft) {
      try {
        const parsed = JSON.parse(draft);
        setFormData((prev) => ({ ...prev, ...parsed }));
        setIsDirty(true);
      } catch {
        // Ignore invalid draft
      }
    }
  }, []);

  const createMutation = useMutation<App, Error, CreateAppRequest>({
    mutationFn: (data) => appsApi.create(data),
    onSuccess: (app) => {
      queryClient.invalidateQueries({ queryKey: ['apps'] });
      localStorage.removeItem('app_create_draft');
      toast.success(`App "${app.name}" created successfully!`, {
        description: 'Redirecting to your new app...',
      });
      navigate(`/apps/${encodeURIComponent(app.slug)}`);
    },
    onError: (error) => {
      setErrors((prev) => ({ ...prev, general: error.message }));
      toast.error('Failed to create app', {
        description: error.message,
      });
    },
  });

  const validateForm = useCallback((): boolean => {
    const nameValidation = validateName(formData.name);
    const slugValidation = validateSlug(formData.slug);
    const descriptionValidation = validateDescription(formData.description);

    const newErrors: FormErrors = {};

    if (!nameValidation.valid) {
      newErrors.name = nameValidation.message;
    }
    if (!slugValidation.valid) {
      newErrors.slug = slugValidation.message;
    }
    if (!descriptionValidation.valid) {
      newErrors.description = descriptionValidation.message;
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  }, [formData.name, formData.slug, formData.description]);

  const setName = useCallback(
    (name: string) => {
      setFormData((prev) => {
        const newSlug = slugManuallyEdited ? prev.slug : generateSlug(name);
        return { ...prev, name, slug: newSlug };
      });
      setIsDirty(true);
      setErrors((prev) => ({ ...prev, name: undefined, general: undefined }));
    },
    [slugManuallyEdited]
  );

  const setSlug = useCallback((slug: string) => {
    const cleaned = slug.toLowerCase().replace(/[^a-z0-9-]/g, '');
    setFormData((prev) => ({ ...prev, slug: cleaned }));
    setSlugManuallyEdited(true);
    setIsDirty(true);
    setErrors((prev) => ({ ...prev, slug: undefined, general: undefined }));
  }, []);

  const setDescription = useCallback((description: string) => {
    if (description.length <= MAX_DESCRIPTION_LENGTH) {
      setFormData((prev) => ({ ...prev, description }));
      setIsDirty(true);
      setErrors((prev) => ({ ...prev, description: undefined }));
    }
  }, []);

  const setIconColor = useCallback((iconColor: string) => {
    setFormData((prev) => ({ ...prev, iconColor }));
    setIsDirty(true);
  }, []);

  const setIconEmoji = useCallback((iconEmoji: string) => {
    setFormData((prev) => ({ ...prev, iconEmoji }));
    setIsDirty(true);
  }, []);

  const setVisibility = useCallback((visibility: Visibility) => {
    setFormData((prev) => ({ ...prev, visibility }));
    setIsDirty(true);
  }, []);

  const setEnvironment = useCallback((environment: Environment) => {
    setFormData((prev) => ({ ...prev, environment }));
    setIsDirty(true);
  }, []);

  const addTag = useCallback((tag: string) => {
    const trimmed = tag.trim().toLowerCase();
    if (!trimmed) return;
    if (trimmed.length > 20) {
      toast.error('Tag must be 20 characters or less');
      return;
    }
    setFormData((prev) => {
      if (prev.tags.includes(trimmed)) {
        return prev;
      }
      if (prev.tags.length >= MAX_TAGS) {
        toast.error(`Maximum ${MAX_TAGS} tags allowed`);
        return prev;
      }
      return { ...prev, tags: [...prev.tags, trimmed] };
    });
    setIsDirty(true);
  }, []);

  const removeTag = useCallback((tag: string) => {
    setFormData((prev) => ({
      ...prev,
      tags: prev.tags.filter((t) => t !== tag),
    }));
    setIsDirty(true);
  }, []);

  const handleSubmit = useCallback(async () => {
    if (!validateForm()) {
      toast.error('Please fix the errors before submitting');
      return;
    }

    await createMutation.mutateAsync({
      name: formData.name.trim(),
      slug: formData.slug,
    });
  }, [validateForm, createMutation, formData.name, formData.slug]);

  const handleCancel = useCallback(() => {
    if (isDirty) {
      const confirmed = window.confirm('You have unsaved changes. Are you sure you want to leave?');
      if (!confirmed) return;
    }
    navigate('/apps');
  }, [isDirty, navigate]);

  const resetForm = useCallback(() => {
    setFormData(initialFormData);
    setErrors({});
    setIsDirty(false);
    setSlugManuallyEdited(false);
    localStorage.removeItem('app_create_draft');
  }, []);

  const nameValidation = validateName(formData.name);
  const slugValidation = validateSlug(formData.slug);
  const descriptionValidation = validateDescription(formData.description);

  const isValid = nameValidation.valid && slugValidation.valid && descriptionValidation.valid;

  return {
    formData,
    errors,
    isDirty,
    isSubmitting: createMutation.isPending,
    isValid,
    lastSaved,

    setName,
    setSlug,
    setDescription,
    setIconColor,
    setIconEmoji,
    setVisibility,
    setEnvironment,
    addTag,
    removeTag,

    handleSubmit,
    handleCancel,
    resetForm,

    slugPreview: formData.slug ? publicAppUrl(formData.slug) : '',
    nameValidation,
    slugValidation,
    descriptionValidation,
  };
}
