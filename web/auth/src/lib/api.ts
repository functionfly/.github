import { API_ORIGIN } from "../config";

const CSRF_HEADER = "X-CSRF-Token";
const CSRF_COOKIE = "csrf_token";

function getCsrfToken(): string | null {
  if (typeof document === "undefined") return null;
  const cookies = document.cookie.split(";");
  for (const cookie of cookies) {
    const [key, val] = cookie.trim().split("=");
    if (key === CSRF_COOKIE) return decodeURIComponent(val);
  }
  return null;
}

async function handleResponse<T>(res: Response): Promise<T> {
  let data: Record<string, unknown> = {};
  try {
    data = await res.json();
  } catch {
    if (!res.ok) {
      console.error(`[API] ${res.status} response parse failed`, res.url);
    }
  }
  if (!res.ok) {
    const msg = (data.message as string | undefined) || `Request failed (${res.status})`;
    throw new Error(msg);
  }
  return data as T;
}

export const api = {
  async post<T = unknown>(path: string, body: unknown): Promise<T> {
    const csrfToken = getCsrfToken();
    const headers: Record<string, string> = { "Content-Type": "application/json" };
    if (csrfToken) {
      headers[CSRF_HEADER] = csrfToken;
    }
    const res = await fetch(`${API_ORIGIN}${path}`, {
      method: "POST",
      headers,
      body: JSON.stringify(body),
      credentials: "include",
    });
    return handleResponse<T>(res);
  },

  async get<T = unknown>(path: string): Promise<T> {
    const csrfToken = getCsrfToken();
    const headers: Record<string, string> = {};
    if (csrfToken) {
      headers[CSRF_HEADER] = csrfToken;
    }
    const res = await fetch(`${API_ORIGIN}${path}`, {
      method: "GET",
      headers,
      credentials: "include",
    });
    return handleResponse<T>(res);
  },
};

export interface LoginResponse {
  token?: string;
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
