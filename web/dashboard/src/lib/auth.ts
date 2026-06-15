/**
 * FunctionFly API Authentication Client
 * 
 * DEPRECATED: This class is deprecated. Use `apiClient` from `@/api/client` and
 * `authStore` from `@/stores/authStore` instead.
 * 
 * This class handles JWT tokens, CSRF tokens, and HMAC signing for API requests.
 * 
 * Security: Tokens are stored encrypted via TokenVault to prevent XSS exfiltration.
 * 
 * @deprecated Use apiClient (Axios-based) for all API calls. This class will be removed
 * in a future version. HMAC signing is disabled in browser environments anyway.
 */

import { tokenVault } from '@/utils/token-vault';

interface AuthTokens {
  jwt: string;
  refreshToken?: string;
}

interface CSRFTokens {
  token: string;
  expiresAt: string;
}

interface APIRequest {
  method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  path: string;
  body?: any;
  headers?: Record<string, string>;
}

interface SignedRequest extends APIRequest {
  headers: Record<string, string>;
}

/**
 * @deprecated Use apiClient from @/api/client for all API requests.
 * This class is kept for backward compatibility and will be removed.
 */
export class FunctionFlyAuth {
  private baseURL: string;
  private jwtToken: string | null = null;
  private refreshTokenValue: string | null = null;
  private csrfToken: string | null = null;
  private csrfExpiresAt: Date | null = null;
  private sharedSecret: string;

  constructor(baseURL: string = '', sharedSecret: string = '') {
    this.baseURL = baseURL || (typeof window !== 'undefined' ? window.location.origin : '');
    // Get shared secret from environment or configuration
    this.sharedSecret = sharedSecret || this.getSharedSecret();
  }

  private getSharedSecret(): string {
    // In production, this should come from secure configuration
    // For now, return empty string - HMAC signing will be disabled
    if (typeof window !== 'undefined') {
      return (window as any).FFLY_SHARED_SECRET || '';
    }
    return '';
  }

  /**
   * Initialize the auth client and load tokens from secure storage
   */
  async initialize(): Promise<void> {
    await tokenVault.initialize();
    this.jwtToken = await tokenVault.getAccessToken();
    this.refreshTokenValue = await tokenVault.getRefreshToken();
  }

  /**
   * Authenticate user with email/password
   */
  async login(email: string, password: string): Promise<AuthTokens> {
    const response = await fetch(`${this.baseURL}/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ email, password }),
    });

    if (!response.ok) {
      throw new Error(`Login failed: ${response.statusText}`);
    }

    const data = await response.json();
    this.jwtToken = data.token;
    this.refreshTokenValue = data.refresh_token;

    // Store tokens in encrypted storage
    await tokenVault.setAccessToken(data.token);
    if (data.refresh_token) {
      await tokenVault.setRefreshToken(data.refresh_token);
    }

    return {
      jwt: data.token,
      refreshToken: data.refresh_token,
    };
  }

  /**
   * Request a magic link to be sent to the user's email
   */
  async requestMagicLink(email: string, redirectPath?: string): Promise<{ message: string; email_sent: boolean }> {
    const response = await fetch(`${this.baseURL}/v1/auth/magic-link`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ email, redirect_path: redirectPath }),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.message || `Magic link request failed: ${response.statusText}`);
    }

    return response.json();
  }

  /**
   * Verify a magic link token and authenticate
   */
  async verifyMagicLink(token: string): Promise<AuthTokens & { user: any; new_user?: boolean }> {
    const response = await fetch(`${this.baseURL}/v1/auth/magic-link/verify`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ token }),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.message || `Magic link verification failed: ${response.statusText}`);
    }

    const data = await response.json();
    this.jwtToken = data.token;
    this.refreshTokenValue = data.refresh_token;

    // Store tokens in encrypted storage
    await tokenVault.setAccessToken(data.token);
    if (data.refresh_token) {
      await tokenVault.setRefreshToken(data.refresh_token);
    }

    return {
      jwt: data.token,
      refreshToken: data.refresh_token,
      user: data.user,
      new_user: data.new_user,
    };
  }

  /**
   * Refresh JWT token (public method)
   */
  async refreshAuthToken(): Promise<string> {
    if (!this.refreshTokenValue) {
      throw new Error('No refresh token available');
    }

    const response = await fetch(`${this.baseURL}/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ refresh_token: this.refreshTokenValue }),
    });

    if (!response.ok) {
      throw new Error(`Token refresh failed: ${response.statusText}`);
    }

    const data = await response.json();
    this.jwtToken = data.token;
    
    // Store in encrypted storage
    await tokenVault.setAccessToken(data.token);
    if (data.refresh_token) {
      await tokenVault.setRefreshToken(data.refresh_token);
    }

    return data.token;
  }

  /**
   * Get fresh CSRF token
   */
  async getCSRFTokens(): Promise<CSRFTokens> {
    const response = await fetch(`${this.baseURL}/v1/admin/csrf`, {
      headers: {
        'Authorization': `Bearer ${this.jwtToken}`,
      },
    });

    if (!response.ok) {
      throw new Error(`CSRF token fetch failed: ${response.statusText}`);
    }

    const data = await response.json();
    this.csrfToken = data.token;
    this.csrfExpiresAt = new Date(data.expires_at);

    return data;
  }

  /**
   * Check if CSRF token is still valid (with 5 minute buffer)
   */
  private isCSRFTokenValid(): boolean {
    if (!this.csrfToken || !this.csrfExpiresAt) {
      return false;
    }

    const now = new Date();
    const bufferTime = 5 * 60 * 1000; // 5 minutes buffer
    return this.csrfExpiresAt.getTime() > (now.getTime() + bufferTime);
  }

  /**
   * Generate HMAC signature for request (disabled in browser by default)
   */
  private async generateHMAC(method: string, path: string, body: string): Promise<{ signature: string; timestamp: number }> {
    // Skip HMAC signing in browser environments
    // Only enable in Node.js server environments
    if (typeof window !== 'undefined' || !this.sharedSecret) {
      return { signature: '', timestamp: Math.floor(Date.now() / 1000) };
    }

    const timestamp = Math.floor(Date.now() / 1000);

    try {
      // Node.js implementation - dynamic import to avoid browser parse errors
      const crypto = await import('crypto');

      // Calculate SHA256 hash of body
      const bodyHash = body ? crypto.createHash('sha256').update(body).digest('hex') : '';

      // Create signature string: timestamp + method + path + bodyHash
      const signatureString = `${timestamp}${method}${path}${bodyHash}`;

      // Calculate HMAC-SHA256
      const signature = crypto.createHmac('sha256', this.sharedSecret).update(signatureString).digest('hex');

      return { signature, timestamp };
    } catch (error) {
      // Fallback: skip HMAC signing if crypto not available
      return { signature: '', timestamp };
    }
  }

  /**
   * Prepare authenticated API request with proper headers
   */
  async prepareRequest(request: APIRequest): Promise<SignedRequest> {
    const headers: Record<string, string> = {
      ...request.headers,
    };

    // Add JWT token
    if (this.jwtToken) {
      headers['Authorization'] = `Bearer ${this.jwtToken}`;
    }

    // Add CSRF token for mutations
    if (['POST', 'PUT', 'PATCH', 'DELETE'].includes(request.method)) {
      if (!this.isCSRFTokenValid()) {
        await this.getCSRFTokens();
      }
      if (this.csrfToken) {
        headers['X-CSRF-Token'] = this.csrfToken;
      }
    }

    // Add HMAC signature for admin endpoints that require it
    if (this.sharedSecret && request.path.startsWith('/v1/admin/') &&
        ['POST', 'PUT', 'PATCH', 'DELETE'].includes(request.method)) {
      const bodyString = request.body ? JSON.stringify(request.body) : '';
      const { signature, timestamp } = await this.generateHMAC(request.method, request.path, bodyString);

      if (signature) {
        headers['X-FFLY-Timestamp'] = timestamp.toString();
        headers['X-FFLY-Signature'] = signature;
      }
    }

    return {
      ...request,
      headers,
    };
  }

  /**
   * Make authenticated API request
   */
  async request(request: APIRequest): Promise<Response> {
    const signedRequest = await this.prepareRequest(request);

    const response = await fetch(`${this.baseURL}${signedRequest.path}`, {
      method: signedRequest.method,
      headers: signedRequest.headers,
      body: signedRequest.body ? JSON.stringify(signedRequest.body) : undefined,
    });

    // Handle token refresh on 401
    if (response.status === 401 && this.refreshTokenValue) {
      try {
        await this.refreshAuthToken();
        // Retry request with new token
        return this.request(request);
      } catch (error) {
        // Refresh failed, user needs to login again
        await this.clearTokens();
        throw new Error('Session expired, please login again');
      }
    }

    return response;
  }

  /**
   * Convenience methods for HTTP verbs
   */
  async get(path: string, headers?: Record<string, string>): Promise<Response> {
    return this.request({ method: 'GET', path, headers });
  }

  async post(path: string, body?: any, headers?: Record<string, string>): Promise<Response> {
    return this.request({ method: 'POST', path, body, headers });
  }

  async put(path: string, body?: any, headers?: Record<string, string>): Promise<Response> {
    return this.request({ method: 'PUT', path, body, headers });
  }

  async patch(path: string, body?: any, headers?: Record<string, string>): Promise<Response> {
    return this.request({ method: 'PATCH', path, body, headers });
  }

  async delete(path: string, headers?: Record<string, string>): Promise<Response> {
    return this.request({ method: 'DELETE', path, body: undefined, headers });
  }

  /**
   * Clear tokens from secure storage
   */
  async clearTokens(): Promise<void> {
    this.jwtToken = null;
    this.refreshTokenValue = null;
    this.csrfToken = null;
    this.csrfExpiresAt = null;

    await tokenVault.clearTokens();
    await tokenVault.clearSessionKey();
  }

  /**
   * Logout
   */
  async logout(): Promise<void> {
    try {
      await fetch(`${this.baseURL}/auth/logout`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.jwtToken}`,
        },
      });
    } catch (error) {
      // Ignore logout errors
    }

    await this.clearTokens();
  }

  /**
   * Check if user is authenticated
   */
  isAuthenticated(): boolean {
    return !!this.jwtToken;
  }
}

// Default instance
export const auth = new FunctionFlyAuth();

// Initialize on module load (browser)
if (typeof window !== 'undefined') {
  auth.initialize();
}
