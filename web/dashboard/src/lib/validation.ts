import { z } from 'zod';

// Password strength validation
const passwordSchema = z
  .string()
  .min(8, 'Password must be at least 8 characters')
  .max(128, 'Password must be less than 128 characters')
  .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
  .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
  .regex(/[0-9]/, 'Password must contain at least one number')
  .regex(/[^A-Za-z0-9]/, 'Password must contain at least one special character')
  .refine((password) => !/\s/.test(password), 'Password cannot contain spaces');

// Login form schema
export const loginSchema = z.object({
  email: z.string().min(1, 'Email is required').email('Please enter a valid email address'),
  password: z.string().min(1, 'Password is required'),
  rememberMe: z.boolean().optional(),
});

/** Age in years from a calendar date (YYYY-MM-DD), local timezone. */
function ageFromISODate(isoDate: string): number {
  const parts = isoDate.split('-').map(Number);
  if (parts.length !== 3 || parts.some((n) => Number.isNaN(n))) return -1;
  const [y, m, d] = parts;
  const birth = new Date(y, m - 1, d);
  const today = new Date();
  let age = today.getFullYear() - birth.getFullYear();
  const md = today.getMonth() - birth.getMonth();
  if (md < 0 || (md === 0 && today.getDate() < birth.getDate())) {
    age--;
  }
  return age;
}

// Signup form schema
export const signupSchema = z
  .object({
    name: z
      .string()
      .min(1, 'Name is required')
      .min(2, 'Name must be at least 2 characters')
      .max(100, 'Name must be less than 100 characters'),
    dateOfBirth: z
      .string()
      .min(1, 'Date of birth is required')
      .regex(/^\d{4}-\d{2}-\d{2}$/, 'Enter a valid date')
      .refine((s) => !Number.isNaN(Date.parse(`${s}T12:00:00`)), 'Invalid date')
      .refine((s) => {
        const t = new Date(`${s}T12:00:00`);
        const startOfToday = new Date();
        startOfToday.setHours(0, 0, 0, 0);
        return t <= startOfToday;
      }, 'Date of birth cannot be in the future')
      .refine((s) => ageFromISODate(s) >= 13, 'You must be at least 13 years old'),
    email: z.string().min(1, 'Email is required').email('Please enter a valid email address'),
    username: z
      .string()
      .min(1, 'Username is required')
      .max(50, 'Username must be less than 50 characters')
      .regex(
        /^[a-zA-Z0-9_-]*$/,
        'Username can only contain letters, numbers, underscores and hyphens'
      ),
    companyName: z
      .string()
      .max(255, 'Company name must be less than 255 characters')
      .optional()
      .or(z.literal('')),
    inviteCode: z.string().optional(),
    password: passwordSchema,
    confirmPassword: z.string(),
    termsAccepted: z
      .boolean()
      .refine((val) => val === true, 'You must accept the terms and conditions'),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Passwords don't match",
    path: ['confirmPassword'],
  });

// Redirect form schema
export const redirectSchema = z.object({
  source: z
    .string()
    .min(1, 'Source path is required')
    .max(2048, 'Source path is too long')
    .regex(/^\/.*$/, 'Source path must start with "/"')
    .regex(
      /^[a-zA-Z0-9\-_\/.]+$/,
      'Source path can only contain letters, numbers, hyphens, underscores, slashes, and dots'
    )
    .refine((path) => !/\.\./.test(path), 'Source path cannot contain ".."')
    .refine((path) => !path.includes('//'), 'Source path cannot contain consecutive slashes'),
  destination: z
    .string()
    .min(1, 'Destination URL is required')
    .max(2048, 'Destination URL is too long')
    .url('Please enter a valid URL')
    .refine((url) => !url.includes(' '), 'URL cannot contain spaces')
    .refine((url) => url.length <= 2048, 'URL is too long'),
  statusCode: z
    .number()
    .int('Status code must be an integer')
    .refine(
      (code) => [301, 302, 307, 308].includes(code),
      'Status code must be 301, 302, 307, or 308'
    ),
  matchType: z.enum(['exact', 'prefix', 'regex'], {
    errorMap: () => ({ message: 'Match type must be exact, prefix, or regex' }),
  }),
  enabled: z.boolean(),
  notes: z.string().max(500, 'Notes must be less than 500 characters').optional(),
});

// Password strength evaluation function
export const evaluatePasswordStrength = (password: string) => {
  let score = 0;
  const checks = {
    length: password.length >= 8,
    uppercase: /[A-Z]/.test(password),
    lowercase: /[a-z]/.test(password),
    number: /[0-9]/.test(password),
    special: /[^A-Za-z0-9]/.test(password),
  };

  score = Object.values(checks).filter(Boolean).length;

  let strength: 'weak' | 'medium' | 'strong' | 'very-strong' = 'weak';
  let label = 'Very Weak';
  let color = 'bg-red-500';

  if (score >= 5) {
    strength = 'very-strong';
    label = 'Very Strong';
    color = 'bg-green-600';
  } else if (score >= 4) {
    strength = 'strong';
    label = 'Strong';
    color = 'bg-green-500';
  } else if (score >= 3) {
    strength = 'medium';
    label = 'Medium';
    color = 'bg-yellow-500';
  } else if (score >= 2) {
    strength = 'weak';
    label = 'Weak';
    color = 'bg-orange-500';
  }

  return {
    strength,
    label,
    color,
    score,
    maxScore: 5,
    checks,
  };
};

// Form persistence utilities
export const FORM_STORAGE_KEYS = {
  signup: 'functionfly_signup_form',
  login: 'functionfly_login_form',
  redirect: 'functionfly_redirect_form',
} as const;

export const saveFormData = (
  formKey: keyof typeof FORM_STORAGE_KEYS,
  data: Record<string, any>
) => {
  try {
    const timestamp = Date.now();
    const formData = { ...data, _timestamp: timestamp };
    localStorage.setItem(FORM_STORAGE_KEYS[formKey], JSON.stringify(formData));
  } catch (error) {
    console.warn('Failed to save form data:', error);
  }
};

export const loadFormData = (
  formKey: keyof typeof FORM_STORAGE_KEYS,
  maxAge = 24 * 60 * 60 * 1000
) => {
  try {
    const stored = localStorage.getItem(FORM_STORAGE_KEYS[formKey]);
    if (!stored) return null;

    const formData = JSON.parse(stored);
    const age = Date.now() - (formData._timestamp || 0);

    if (age > maxAge) {
      localStorage.removeItem(FORM_STORAGE_KEYS[formKey]);
      return null;
    }

    // Remove timestamp from returned data
    const { _timestamp, ...data } = formData;
    return data;
  } catch (error) {
    console.warn('Failed to load form data:', error);
    return null;
  }
};

export const clearFormData = (formKey: keyof typeof FORM_STORAGE_KEYS) => {
  try {
    localStorage.removeItem(FORM_STORAGE_KEYS[formKey]);
  } catch (error) {
    console.warn('Failed to clear form data:', error);
  }
};

// Additional utility schemas
export const uuidSchema = z.string().uuid('Invalid UUID format');

export const slugSchema = z
  .string()
  .min(1, 'Slug is required')
  .max(100, 'Slug is too long')
  .regex(/^[a-z0-9-]+$/, 'Slug can only contain lowercase letters, numbers, and hyphens')
  .refine(
    (slug) => !slug.startsWith('-') && !slug.endsWith('-'),
    'Slug cannot start or end with a hyphen'
  );

export const functionNameSchema = z
  .string()
  .min(1, 'Function name is required')
  .max(100, 'Function name is too long')
  .regex(
    /^[a-zA-Z_][a-zA-Z0-9_-]*$/,
    'Function name must start with a letter or underscore and contain only letters, numbers, underscores, and hyphens'
  );

export const environmentVariableKeySchema = z
  .string()
  .regex(/^[A-Z_][A-Z0-9_]*$/, 'Environment variable key must be uppercase with underscores only')
  .max(100, 'Environment variable key is too long');

export const environmentVariableValueSchema = z
  .string()
  .max(10000, 'Environment variable value is too long');

// Enhanced contact form schema with better validation
export const contactFormSchema = z.object({
  name: z
    .string()
    .min(2, 'Name must be at least 2 characters')
    .max(50, 'Name must be less than 50 characters')
    .regex(/^[a-zA-Z\s'-]+$/, 'Name can only contain letters, spaces, hyphens, and apostrophes'),
  email: z
    .string()
    .email('Please enter a valid email address')
    .max(254, 'Email address is too long'),
  category: z.string().min(1, 'Please select a category'),
  subject: z
    .string()
    .min(5, 'Subject must be at least 5 characters')
    .max(100, 'Subject must be less than 100 characters')
    .refine((subject) => subject.trim().length > 0, 'Subject cannot be empty or just whitespace'),
  message: z
    .string()
    .min(10, 'Message must be at least 10 characters')
    .max(2000, 'Message must be less than 2000 characters')
    .refine((message) => message.trim().length > 0, 'Message cannot be empty or just whitespace'),
});

// Types inferred from schemas
export type LoginFormData = z.infer<typeof loginSchema>;
export type SignupFormData = z.infer<typeof signupSchema>;
export type RedirectFormData = z.infer<typeof redirectSchema>;
export type ContactFormData = z.infer<typeof contactFormSchema>;
