import { createSignupSchema, loginSchema } from '@/lib/validation';
import { useMemo } from 'react';
import { useFormWithValidation } from './useFormWithValidation';

export function useLoginForm() {
  return useFormWithValidation({
    schema: loginSchema,
    formKey: 'login',
    enablePersistence: true,
    enableRealtimeValidation: true,
    defaultValues: {
      email: '',
      password: '',
      rememberMe: false,
    },
  });
}

export function useSignupForm(inviteRequired = false) {
  const schema = useMemo(() => createSignupSchema(inviteRequired), [inviteRequired]);
  return useFormWithValidation({
    schema,
    formKey: 'signup',
    enablePersistence: true,
    enableRealtimeValidation: true,
    defaultValues: {
      name: '',
      dateOfBirth: '',
      email: '',
      username: '',
      companyName: '',
      inviteCode: '',
      password: '',
      confirmPassword: '',
      termsAccepted: false,
    },
  });
}
