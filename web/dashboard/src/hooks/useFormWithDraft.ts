import { useForm, UseFormProps, FieldValues, UseFormReturn } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useEffect, useCallback, useRef, useState } from 'react';
import { toast } from 'sonner';

// Storage key type - can use existing keys or custom string keys
export type DraftStorageKey = string;

export interface DraftMetadata {
  hasDraft: boolean;
  lastSavedAt: Date | null;
  draftAge: number | null; // in milliseconds
  isDirty: boolean;
  saveStatus: 'idle' | 'saving' | 'saved' | 'error';
}

export interface UseFormWithDraftOptions<T extends FieldValues> extends Omit<UseFormProps<T>, 'resolver'> {
  schema: z.ZodSchema<T>;
  draftKey?: DraftStorageKey;
  enableDraftSave?: boolean;
  draftMaxAgeMs?: number;
  enableRealtimeValidation?: boolean;
  validationDelay?: number;
  onDraftRestored?: (data: T) => void;
  onDraftSaved?: (metadata: DraftMetadata) => void;
  showDraftRestoredToast?: boolean;
  onBeforeSave?: (data: T) => boolean | Promise<boolean>;
}

export interface UseFormWithDraftReturn<T extends FieldValues> extends UseFormReturn<T> {
  draft: {
    metadata: DraftMetadata;
    restoreDraft: () => boolean;
    clearDraft: () => void;
    saveDraft: () => void;
    discardDraft: () => void;
  };
}

interface DraftStoragePayload {
  data: Record<string, unknown>;
  timestamp: number;
  version: number;
}

interface LoadDraftResult {
  data: Record<string, unknown> | null;
  metadata: { age: number; timestamp: number } | null;
}

// Save to localStorage with metadata
const saveDraftToStorage = (key: string, data: Record<string, unknown>): boolean => {
  try {
    const payload: DraftStoragePayload = {
      data,
      timestamp: Date.now(),
      version: 1,
    };
    localStorage.setItem(`draft:${key}`, JSON.stringify(payload));
    return true;
  } catch (error) {
    console.warn('Failed to save draft:', error);
    return false;
  }
};

// Load from localStorage with metadata
const loadDraftFromStorage = (key: string, maxAgeMs: number): LoadDraftResult => {
  try {
    const stored = localStorage.getItem(`draft:${key}`);
    if (!stored) return { data: null, metadata: null };

    const payload = JSON.parse(stored);
    const age = Date.now() - (payload.timestamp || 0);

    if (age > maxAgeMs) {
      localStorage.removeItem(`draft:${key}`);
      return { data: null, metadata: null };
    }

    return {
      data: payload.data || null,
      metadata: { age, timestamp: payload.timestamp }
    };
  } catch (error) {
    console.warn('Failed to load draft:', error);
    return { data: null, metadata: null };
  }
};

// Clear draft from localStorage
const clearDraftFromStorage = (key: string) => {
  try {
    localStorage.removeItem(`draft:${key}`);
  } catch (error) {
    console.warn('Failed to clear draft:', error);
  }
};

export function useFormWithDraft<T extends FieldValues>({
  schema,
  draftKey,
  enableDraftSave = true,
  draftMaxAgeMs = 7 * 24 * 60 * 60 * 1000, // 7 days default
  enableRealtimeValidation = true,
  validationDelay = 300,
  defaultValues,
  onDraftRestored,
  onDraftSaved,
  showDraftRestoredToast = true,
  onBeforeSave,
  ...formOptions
}: UseFormWithDraftOptions<T>): UseFormWithDraftReturn<T> {
  const [draftMetadata, setDraftMetadata] = useState<DraftMetadata>({
    hasDraft: false,
    lastSavedAt: null,
    draftAge: null,
    isDirty: false,
    saveStatus: 'idle',
  });

  const validationTimeoutRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const saveTimeoutRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const isValidatingRef = useRef(false);
  const lastSavedDataRef = useRef<string>('');
  const originalDataRef = useRef<string>('');

  // Check for existing draft on mount
  const getInitialValues = useCallback(() => {
    if (!enableDraftSave || !draftKey) {
      originalDataRef.current = JSON.stringify(defaultValues || {});
      return defaultValues;
    }

    const { data, metadata } = loadDraftFromStorage(draftKey, draftMaxAgeMs);

    if (data && metadata) {
      // Draft found - prepare it but don't auto-restore
      // Let the component decide when to restore (e.g., via a modal)
      setDraftMetadata({
        hasDraft: true,
        lastSavedAt: new Date(metadata.timestamp),
        draftAge: metadata.age,
        isDirty: false,
        saveStatus: 'saved',
      });
      originalDataRef.current = JSON.stringify(defaultValues || {});
      lastSavedDataRef.current = JSON.stringify(data);
      return defaultValues;
    }

    originalDataRef.current = JSON.stringify(defaultValues || {});
    return defaultValues;
  }, [draftKey, draftMaxAgeMs, defaultValues, enableDraftSave]);

  const form = useForm<T>({
    ...formOptions,
    resolver: zodResolver(schema as any) as any,
    defaultValues: getInitialValues(),
    mode: enableRealtimeValidation ? 'onChange' : 'onSubmit',
  });

  // Track form dirty state
  const watchValues = form.watch();

  useEffect(() => {
    const currentData = JSON.stringify(watchValues);
    const isDirty = currentData !== originalDataRef.current && currentData !== lastSavedDataRef.current;

    setDraftMetadata(prev => ({
      ...prev,
      isDirty,
    }));
  }, [watchValues]);

  // Auto-save draft with debouncing
  useEffect(() => {
    if (!enableDraftSave || !draftKey) return;

    const subscription = form.watch((data) => {
      // Clear existing timeout
      if (saveTimeoutRef.current) {
        clearTimeout(saveTimeoutRef.current);
      }

      setDraftMetadata(prev => ({ ...prev, saveStatus: 'saving' }));

      saveTimeoutRef.current = setTimeout(async () => {
        const shouldSave = onBeforeSave ? await onBeforeSave(data as T) : true;

        if (!shouldSave) {
          setDraftMetadata(prev => ({ ...prev, saveStatus: 'idle' }));
          return;
        }

        const success = saveDraftToStorage(draftKey, data as Record<string, unknown>);

        if (success) {
          lastSavedDataRef.current = JSON.stringify(data);
          const timestamp = Date.now();
          setDraftMetadata(prev => ({
            ...prev,
            hasDraft: true,
            lastSavedAt: new Date(timestamp),
            draftAge: 0,
            isDirty: false,
            saveStatus: 'saved',
          }));
          onDraftSaved?.({
            hasDraft: true,
            lastSavedAt: new Date(timestamp),
            draftAge: 0,
            isDirty: false,
            saveStatus: 'saved',
          });
        } else {
          setDraftMetadata(prev => ({ ...prev, saveStatus: 'error' }));
        }
      }, validationDelay);
    });

    return () => {
      subscription.unsubscribe();
      if (saveTimeoutRef.current) {
        clearTimeout(saveTimeoutRef.current);
      }
    };
  }, [form, draftKey, enableDraftSave, validationDelay, onBeforeSave, onDraftSaved]);

  // Real-time validation with debouncing
  useEffect(() => {
    if (!enableRealtimeValidation) return;

    const subscription = form.watch(() => {
      if (validationTimeoutRef.current) {
        clearTimeout(validationTimeoutRef.current);
      }

      isValidatingRef.current = true;

      validationTimeoutRef.current = setTimeout(() => {
        isValidatingRef.current = false;
      }, validationDelay);
    });

    return () => {
      subscription.unsubscribe();
      if (validationTimeoutRef.current) {
        clearTimeout(validationTimeoutRef.current);
      }
    };
  }, [form, enableRealtimeValidation, validationDelay]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (validationTimeoutRef.current) {
        clearTimeout(validationTimeoutRef.current);
      }
      if (saveTimeoutRef.current) {
        clearTimeout(saveTimeoutRef.current);
      }
    };
  }, []);

  const restoreDraft = useCallback(() => {
    if (!draftKey) return false;

    const { data, metadata } = loadDraftFromStorage(draftKey, draftMaxAgeMs);

    if (data && metadata) {
      form.reset(data as T);
      lastSavedDataRef.current = JSON.stringify(data);

      setDraftMetadata({
        hasDraft: false,
        lastSavedAt: new Date(metadata.timestamp),
        draftAge: metadata.age,
        isDirty: false,
        saveStatus: 'saved',
      });

      onDraftRestored?.(data as T);

      if (showDraftRestoredToast) {
        const ageMinutes = Math.floor(metadata.age / 60000);
        const ageText = ageMinutes < 1 ? 'just now' :
                       ageMinutes < 60 ? `${ageMinutes}m ago` :
                       `${Math.floor(ageMinutes / 60)}h ago`;

        toast.success('Draft restored', {
          description: `Your draft from ${ageText} has been loaded`,
        });
      }

      return true;
    }

    return false;
  }, [draftKey, draftMaxAgeMs, form, onDraftRestored, showDraftRestoredToast]);

  const clearDraft = useCallback(() => {
    if (!draftKey) return;

    clearDraftFromStorage(draftKey);
    setDraftMetadata({
      hasDraft: false,
      lastSavedAt: null,
      draftAge: null,
      isDirty: false,
      saveStatus: 'idle',
    });
  }, [draftKey]);

  const saveDraft = useCallback(() => {
    if (!draftKey) return;

    const data = form.getValues();
    const success = saveDraftToStorage(draftKey, data as Record<string, unknown>);

    if (success) {
      const timestamp = Date.now();
      setDraftMetadata({
        hasDraft: true,
        lastSavedAt: new Date(timestamp),
        draftAge: 0,
        isDirty: false,
        saveStatus: 'saved',
      });

      toast.success('Draft saved manually');
    } else {
      toast.error('Failed to save draft');
    }
  }, [draftKey, form]);

  const discardDraft = useCallback(() => {
    clearDraft();
    form.reset();
    toast.info('Draft discarded');
  }, [clearDraft, form]);

  return {
    ...form,
    draft: {
      metadata: draftMetadata,
      restoreDraft,
      clearDraft,
      saveDraft,
      discardDraft,
    },
  };
}

interface DraftSummary {
  key: string;
  timestamp: number;
  age: number;
}

// Hook for multiple drafts management
export function useDraftManager() {
  const listDrafts = useCallback((): DraftSummary[] => {
    const drafts: DraftSummary[] = [];

    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (key?.startsWith('draft:')) {
        try {
          const data = JSON.parse(localStorage.getItem(key) || '{}');
          const age = Date.now() - (data.timestamp || 0);
          drafts.push({
            key: key.replace('draft:', ''),
            timestamp: data.timestamp,
            age,
          });
        } catch {
          // Skip invalid entries
        }
      }
    }

    return drafts.sort((a, b) => b.timestamp - a.timestamp);
  }, []);

  const clearAllDrafts = useCallback(() => {
    const draftKeys: string[] = [];

    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (key?.startsWith('draft:')) {
        draftKeys.push(key);
      }
    }

    draftKeys.forEach(key => localStorage.removeItem(key));
    toast.success('All drafts cleared');
  }, []);

  return {
    listDrafts,
    clearAllDrafts,
  };
}
