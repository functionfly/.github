import { getApiBaseUrl } from '@/lib/constants';

export interface SignupConfigResponse {
  inviteRequired: boolean;
}

/** Captcha provider configuration for signup. */
export interface SignupCaptchaPublic {
  provider: 'turnstile' | 'recaptcha_v3';
  siteKey?: string;
}

/** Public endpoint: reflects SIGNUP_REQUIRE_INVITE_CODE on the API. */
export async function fetchSignupConfig(): Promise<SignupConfigResponse> {
  const base = getApiBaseUrl().replace(/\/$/, '');
  const url = `${base}/auth/signup-config`;
  const res = await fetch(url, { method: 'GET', credentials: 'include' });
  if (!res.ok) {
    throw new Error(`signup-config failed: ${res.status}`);
  }
  return res.json() as Promise<SignupConfigResponse>;
}
