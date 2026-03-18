import type { LoginRequest, LoginResponse, SignupRequest, SignupResponse } from '@/types';
import { apiClient } from './client';

/** Server-provided WebAuthn assertion options (re-auth / reveal gate) */
export interface WebAuthnAssertionBeginResponse {
  options: PublicKeyCredentialRequestOptionsJSON;
  sessionID: string;
}

/** JSON-serializable form of PublicKeyCredentialRequestOptions (challenge is base64url) */
export interface PublicKeyCredentialRequestOptionsJSON {
  challenge: string;
  timeout?: number;
  rpId?: string;
  allowCredentials?: Array<{ type: string; id: string; transports?: string[] }>;
  userVerification?: UserVerificationRequirement;
}

/** Decode base64url to bytes. Returns ArrayBuffer-backed Uint8Array for WebAuthn BufferSource. */
function decodeBase64url(b64: string): Uint8Array<ArrayBuffer> {
  const base64 = b64.replace(/-/g, '+').replace(/_/g, '/');
  const binary = atob(base64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
  return bytes as Uint8Array<ArrayBuffer>;
}

/** Convert server options to browser PublicKeyCredentialRequestOptions */
export function toPublicKeyRequestOptions(
  json: PublicKeyCredentialRequestOptionsJSON
): PublicKeyCredentialRequestOptions {
  const allowCredentials = json.allowCredentials?.map((c) => ({
    type: 'public-key' as const,
    id: decodeBase64url(c.id),
    transports: c.transports as AuthenticatorTransport[] | undefined,
  }));
  return {
    challenge: decodeBase64url(json.challenge),
    timeout: json.timeout,
    rpId: json.rpId,
    allowCredentials,
    userVerification: json.userVerification ?? 'required',
  };
}

/** Start WebAuthn assertion (re-authentication). Requires existing auth. Returns options + sessionID. */
export async function webauthnAssertionBegin(): Promise<WebAuthnAssertionBeginResponse> {
  const data = await apiClient.post<WebAuthnAssertionBeginResponse>(
    '/v1/auth/webauthn/login/begin',
    {}
  );
  return data;
}

/** Complete WebAuthn assertion with the credential response. Returns success. */
export async function webauthnAssertionComplete(
  sessionId: string,
  response: unknown
): Promise<void> {
  await apiClient.post('/v1/auth/webauthn/login/complete', {
    sessionId,
    response,
  });
}

export const authApi = {
  login: (data: LoginRequest) => apiClient.post<LoginResponse>('/v1/auth/login', data),

  signup: (data: SignupRequest) => apiClient.post<SignupResponse>('/v1/auth/signup', data),

  checkUsernameAvailability: (username: string) =>
    apiClient.get<{ available: boolean; username: string }>(
      `/v1/auth/check-username?username=${encodeURIComponent(username)}`
    ),

  resendVerification: (email: string) =>
    apiClient.post<{ message: string }>('/v1/auth/resend-verification', { email }),

  logout: () => {
    apiClient.clearToken();
  },

  getCurrentUser: () => {
    // Decode JWT to get user info
    const token = apiClient.getToken();
    if (!token) return null;

    try {
      const base64Url = token.split('.')[1];
      const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
      const jsonPayload = decodeURIComponent(
        atob(base64)
          .split('')
          .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
          .join('')
      );
      return JSON.parse(jsonPayload);
    } catch {
      return null;
    }
  },

  /** Verify current user's password (re-auth for sensitive actions). Throws on 401/400. */
  verifyPassword: (password: string) =>
    apiClient.post<{ message: string }>('/v1/auth/verify-password', { password }),
};
