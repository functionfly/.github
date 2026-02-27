import { useForm, UseFormProps } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useEffect, useCallback, useRef } from 'react';
import {
  saveFormData,
  loadFormData,
  clearFormData,
  FORM_STORAGE_KEYS
} from '@/lib/validation';

interface UseFormWithValidationOptions extends Omit<UseFormProps, 'resolver'> {
  schema: z.ZodSchema<any>;
  formKey?: keyof typeof FORM_STORAGE_KEYS;
  enablePersistence?: boolean;
  enableRealtimeValidation?: boolean;
  validationDelay?: number;
}

export function useFormWithValidation({
  schema,
  formKey,
  enablePersistence = false,
  enableRealtimeValidation = true,
  validationDelay = 300,
  defaultValues,
  ...formOptions
}: UseFormWithValidationOptions): ReturnType<typeof useForm> & {
  saveForm: () => void;
  loadForm: () => void;
  clearForm: () => void;
  isValidating: boolean;
} {
  const validationTimeoutRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const isValidatingRef = useRef(false);

  // Load persisted data if enabled
  const loadPersistedData = useCallback(() => {
    if (enablePersistence && formKey) {
      const persistedData = loadFormData(formKey);
      return persistedData || defaultValues;
    }
    return defaultValues;
  }, [enablePersistence, formKey, defaultValues]);

  const form = useForm({
    ...formOptions,
    resolver: zodResolver(schema as any),
    defaultValues: loadPersistedData(),
    mode: enableRealtimeValidation ? 'onChange' : 'onSubmit',
  });

  // Auto-save form data on changes
  useEffect(() => {
    if (!enablePersistence || !formKey) return;

    const subscription = form.watch((data) => {
      // Debounce saving to avoid too frequent localStorage writes
      if (validationTimeoutRef.current) {
        clearTimeout(validationTimeoutRef.current);
      }

      validationTimeoutRef.current = setTimeout(() => {
        saveFormData(formKey, data as Record<string, any>);
      }, validationDelay);
    });

    return () => subscription.unsubscribe();
  }, [form, formKey, enablePersistence, validationDelay]);

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
    };
  }, []);

  const saveForm = useCallback(() => {
    if (enablePersistence && formKey) {
      const data = form.getValues();
      saveFormData(formKey, data as Record<string, any>);
    }
  }, [form, formKey, enablePersistence]);

  const loadForm = useCallback(() => {
    if (enablePersistence && formKey) {
      const persistedData = loadFormData(formKey);
      if (persistedData) {
        form.reset(persistedData as any);
      }
    }
  }, [form, formKey, enablePersistence]);

  const clearForm = useCallback(() => {
    if (enablePersistence && formKey) {
      clearFormData(formKey);
    }
    form.reset(defaultValues as any);
  }, [form, formKey, enablePersistence, defaultValues]);

  return {
    ...form,
    saveForm,
    loadForm,
    clearForm,
    isValidating: isValidatingRef.current,
  };
}