import { useFormWithValidation } from './useFormWithValidation';
import { loginSchema, signupSchema, LoginFormData, SignupFormData } from '@/lib/validation';

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
