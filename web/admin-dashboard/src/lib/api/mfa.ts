import { getApiBaseUrl } from '@/lib/constants';
import { CACHE_KEYS } from '@/lib/constants';

async function parseErrorMessage(response: Response): Promise<string> {
  const text = await response.text();
  if (!text) return 'MFA verification failed';
  try {
    const json = JSON.parse(text) as { message?: string; error?: string };
    return json.message || json.error || 'MFA verification failed';
  } catch {
    return text;
  }
}

function getBearerToken(): string | null {
  try {
    return sessionStorage.getItem(CACHE_KEYS.ADMIN_ACCESS_TOKEN);
  } catch {
    return null;
  }
}

/** Verify a TOTP or backup code against the authenticated session. */
export async function verifyMfaCode(code: string, token?: string): Promise<boolean> {
  const bearer = token ?? getBearerToken();
  if (!bearer) {
    throw new Error('No active session for MFA verification');
  }

  const response = await fetch(`${getApiBaseUrl()}/auth/mfa/verify`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${bearer}`,
    },
    body: JSON.stringify({ code }),
  });

  if (!response.ok) {
    throw new Error(await parseErrorMessage(response));
  }

  const payload = (await response.json()) as { verified?: boolean };
  return payload.verified === true;
}
