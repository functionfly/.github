/**
 * FunctionFly API Authentication Client
 * Handles JWT tokens, CSRF tokens, and HMAC signing for API requests
 */

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

    // Store tokens (in production, use secure storage)
    this.storeTokens();

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

    // Store tokens (in production, use secure storage)
    this.storeTokens();

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
    this.storeTokens();

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
  private generateHMAC(method: string, path: string, body: string): { signature: string; timestamp: number } {
    // Skip HMAC signing in browser environments (requires crypto-js)
    // Only enable in Node.js server environments
    if (typeof window !== 'undefined' || !this.sharedSecret) {
      return { signature: '', timestamp: Math.floor(Date.now() / 1000) };
    }

    const timestamp = Math.floor(Date.now() / 1000);

    try {
      // Node.js implementation
      const crypto = require('crypto');

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
      const { signature, timestamp } = this.generateHMAC(request.method, request.path, bodyString);

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
        this.clearTokens();
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
   * Token storage (use secure storage in production)
   */
  private storeTokens(): void {
    if (typeof window !== 'undefined') {
      if (this.jwtToken) {
        localStorage.setItem('ffly_jwt', this.jwtToken);
      }
      if (this.refreshTokenValue) {
        localStorage.setItem('ffly_refresh', this.refreshTokenValue);
      }
    }
  }

  private loadTokens(): void {
    if (typeof window !== 'undefined') {
      this.jwtToken = localStorage.getItem('ffly_jwt');
      this.refreshTokenValue = localStorage.getItem('ffly_refresh');
    }
  }

  private clearTokens(): void {
    this.jwtToken = null;
    this.refreshTokenValue = null;
    this.csrfToken = null;
    this.csrfExpiresAt = null;

    if (typeof window !== 'undefined') {
      localStorage.removeItem('ffly_jwt');
      localStorage.removeItem('ffly_refresh');
    }
  }

  /**
   * Initialize from stored tokens
   */
  initialize(): void {
    this.loadTokens();
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

    this.clearTokens();
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