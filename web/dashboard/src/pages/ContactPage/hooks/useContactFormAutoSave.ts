import { useState, useEffect, useCallback } from 'react';
import { useForm, UseFormReturn } from 'react-hook-form';
import { debounce } from 'lodash';
import useLocalStorageState from 'use-local-storage-state';
import { ContactFormData } from '../validation';

export function useContactFormAutoSave(form: UseFormReturn<ContactFormData>) {
  // Auto-save state with localStorage
  const [savedDraft, setSavedDraft] = useLocalStorageState<ContactFormData | null>('contact-form-draft', {
    defaultValue: null
  });
  const [lastSaved, setLastSaved] = useState<Date | null>(null);
  const [showDraftIndicator, setShowDraftIndicator] = useState(false);

  // Watch form values for auto-save
  const watchedValues = form.watch();

  // Debounced auto-save function
  const debouncedSave = useCallback(
    debounce((data: ContactFormData) => {
      setSavedDraft(data);
      setLastSaved(new Date());
      setShowDraftIndicator(true);
      setTimeout(() => setShowDraftIndicator(false), 2000);
    }, 1000),
    [setSavedDraft]
  );

  // Auto-save effect
  useEffect(() => {
    if (Object.values(watchedValues).some(value => value !== '')) {
      debouncedSave(watchedValues);
    }
  }, [watchedValues, debouncedSave]);

  // Load draft on mount
  useEffect(() => {
    if (savedDraft) {
      Object.entries(savedDraft).forEach(([key, value]) => {
        form.setValue(key as keyof ContactFormData, value as string);
      });
    }
  }, [savedDraft, form.setValue]);

  const clearDraft = () => {
    setSavedDraft(null);
    setLastSaved(null);
    form.reset({
      name: '',
      email: '',
      subject: '',
      message: '',
      category: 'general'
    });
  };

  return {
    lastSaved,
    showDraftIndicator,
    clearDraft
  };
}