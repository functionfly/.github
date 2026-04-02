import { API_ORIGIN } from "../config";

async function handleResponse<T>(res: Response): Promise<T> {
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    const msg = (data as { message?: string }).message || `Request failed (${res.status})`;
    throw new Error(msg);
  }
  return data as T;
}

export const api = {
  async post<T = unknown>(path: string, body: unknown): Promise<T> {
    const res = await fetch(`${API_ORIGIN}${path}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
      credentials: "include",
    });
    return handleResponse<T>(res);
  },

  async get<T = unknown>(path: string): Promise<T> {
    const res = await fetch(`${API_ORIGIN}${path}`, {
      method: "GET",
      credentials: "include",
    });
    return handleResponse<T>(res);
  },
};

export interface LoginResponse {
  token: string;
  refresh_token?: string;
  user: {
    id: string;
    email: string;
    username?: string;
  };
}

export interface SignupResponse {
  message: string;
  email_sent: boolean;
  requires_verification: boolean;
}

export interface OAuthProviders {
  providers: Record<string, { name: string; icon?: string }>;
}
