import { useFormWithValidation } from './useFormWithValidation';
import { redirectSchema, RedirectFormData } from '@/lib/validation';

export function useRedirectForm(defaultValues?: Partial<RedirectFormData>) {
  return useFormWithValidation<RedirectFormData>({
    schema: redirectSchema,
    formKey: 'redirect',
    enablePersistence: false, // Admin forms don't need persistence
    enableRealtimeValidation: true,
    defaultValues: {
      source: '',
      destination: '',
      statusCode: 301,
      matchType: 'exact',
      enabled: true,
      notes: '',
      ...defaultValues,
    },
  });
}