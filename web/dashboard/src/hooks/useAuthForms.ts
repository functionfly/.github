import { loginSchema, signupSchema } from '@/lib/validation';
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

export function useSignupForm() {
  return useFormWithValidation({
    schema: signupSchema,
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
