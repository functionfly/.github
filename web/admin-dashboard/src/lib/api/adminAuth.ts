import { getApiBaseUrl } from '@/lib/constants';
import type { AdminAuthLoginResponse, AdminSessionBootstrapResponse } from '@/types';

async function parseErrorMessage(response: Response): Promise<string> {
  const text = await response.text();
  if (!text) {
    return 'Request failed';
  }

  try {
    const json = JSON.parse(text) as { message?: string; error?: string };
    return json.message || json.error || 'Request failed';
  } catch {
    return text;
  }
}

export async function loginAdmin(email: string, password: string): Promise<AdminAuthLoginResponse> {
  const apiBaseUrl = getApiBaseUrl();
  const response = await fetch(`${apiBaseUrl}/auth/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ email, password }),
  });

  if (!response.ok) {
    throw new Error(await parseErrorMessage(response));
  }

  return response.json() as Promise<AdminAuthLoginResponse>;
}

export async function bootstrapAdminSession(token: string): Promise<AdminSessionBootstrapResponse> {
  const apiBaseUrl = getApiBaseUrl();

  const sessionResponse = await fetch(`${apiBaseUrl}/v1/admin/auth/session`, {
    method: 'GET',
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (sessionResponse.ok) {
    return sessionResponse.json() as Promise<AdminSessionBootstrapResponse>;
  }

  // Fallback for older backends where /v1/admin/auth/session is not available.
  const validateResponse = await fetch(`${apiBaseUrl}/v1/auth/validate`, {
    method: 'GET',
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (!validateResponse.ok) {
    throw new Error(await parseErrorMessage(validateResponse));
  }

  const payload = await validateResponse.json() as { token: string; user: AdminSessionBootstrapResponse['user'] };

  return {
    session: {
      id: `jwt-${Date.now()}`,
      user_id: payload.user.id,
      session_token_hash: 'jwt',
      access_token: token,
      ip_address: 'unknown',
      user_agent: navigator.userAgent,
      created_at: new Date().toISOString(),
      last_activity_at: new Date().toISOString(),
      expires_at: new Date(Date.now() + 30 * 60 * 1000).toISOString(),
    },
    user: payload.user,
  };
}
