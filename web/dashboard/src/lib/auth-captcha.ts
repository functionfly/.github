import type { SignupCaptchaPublic } from '@/api/signup-config';

/** Dashboard can render widgets for these providers (others need a future UI update). */
export function authCaptchaSupported(c: SignupCaptchaPublic | null | undefined): boolean {
  return !!c && (c.provider === 'turnstile' || c.provider === 'recaptcha_v3');
}

export function authCaptchaRequired(c: SignupCaptchaPublic | null | undefined): boolean {
  return authCaptchaSupported(c);
}

export function authCaptchaUnsupported(c: SignupCaptchaPublic | null | undefined): boolean {
  return !!c && !authCaptchaSupported(c);
}

export function authCaptchaBadgeLabel(c: SignupCaptchaPublic | null | undefined): string | null {
  if (!authCaptchaRequired(c)) return null;
  if (c!.provider === 'turnstile') return 'Protected by Cloudflare Turnstile';
  return 'Protected by Google reCAPTCHA';
}
