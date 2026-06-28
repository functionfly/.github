import { API_BASE_URL } from './constants';

/**
 * API Version
 */
const API_VERSION = 'v1';

/** API root (no /v1). Well-known discovery is served at root. */
const apiRoot = API_BASE_URL.replace(/\/v1\/?$/, '') || API_BASE_URL;

/**
 * AI/LLM discovery manifest URL (GET). Public, no auth. Returns OpenAI-compatible tool schemas for all public functions.
 */
export const WELL_KNOWN_DISCOVERY_URL = `${apiRoot.replace(/\/$/, '')}/.well-known/functionfly.json`;

/**
 * Full API base URL with version
 */
export const API = `${API_BASE_URL}/${API_VERSION}`;

// ============================================================================
// API URL Builders
// ============================================================================

/**
 * API URL builders for consistent backend communication
 */
export const API_URLS = {
  // ========================================================================
  // Auth Endpoints
  // ========================================================================
  auth: {
    login: `${API}/auth/login`,
    signup: `${API}/auth/signup`,
    logout: `${API}/auth/logout`,
    refreshToken: `${API}/auth/refresh`,
    verifyEmail: (token: string) => `${API}/auth/verify/${token}`,
    forgotPassword: `${API}/auth/forgot-password`,
    resetPassword: (token: string) => `${API}/auth/reset-password/${token}`,
    changePassword: `${API}/auth/password`,
  },

  // ========================================================================
  // User Endpoints
  // ========================================================================
  user: {
    me: `${API}/user/me`,
    profile: (userId: string) => `${API}/users/${userId}`,
    updateProfile: `${API}/user/profile`,
    avatar: (userId: string) => `${API}/users/${userId}/avatar`,
    favorites: {
      list: `${API}/users/me/favorites`,
      add: `${API}/users/me/favorites`,
      remove: (functionId: string) => `${API}/users/me/favorites/${functionId}`,
      toggle: (functionId: string) => `${API}/users/me/favorites/${functionId}/toggle`,
      check: (functionId: string) => `${API}/users/me/favorites/${functionId}`,
      updatePosition: (functionId: string) => `${API}/users/me/favorites/${functionId}/position`,
    },
  },

  // ========================================================================
  // Function Registry Endpoints
  // ========================================================================
  functions: {
    list: (page = 1, limit = 20) => `${API}/functions?page=${page}&limit=${limit}`,
    get: (author: string, name: string) => `${API}/functions/${author}/${name}`,
    create: `${API}/functions`,
    update: (author: string, name: string) => `${API}/functions/${author}/${name}`,
    delete: (author: string, name: string) => `${API}/functions/${author}/${name}`,
    search: (query: string, page = 1, limit = 20) =>
      `${API}/functions/search?q=${encodeURIComponent(query)}&page=${page}&limit=${limit}`,
    byAuthor: (author: string, page = 1, limit = 20) =>
      `${API}/functions?author=${author}&page=${page}&limit=${limit}`,
    versions: (author: string, name: string) => `${API}/functions/${author}/${name}/versions`,
    latestVersion: (author: string, name: string) =>
      `${API}/functions/${author}/${name}/versions/latest`,
    settings: (author: string, name: string) => `${API}/functions/${author}/${name}/settings`,
  },

  // ========================================================================
  // Function Execution Endpoints
  // ========================================================================
  execution: {
    execute: (author: string, name: string) => `${API}/fx/${author}/${name}`,
    executeWithVersion: (author: string, name: string, version: string) =>
      `${API}/fx/${author}/${name}@${version}`,
    executeLatest: (author: string, name: string) => `${API}/fx/${author}/${name}/latest`,
    batchExecute: `${API}/fx/batch`,
  },

  // ========================================================================
  // Execution Replay/History Endpoints
  // ========================================================================
  replay: {
    list: (page = 1, limit = 20) => `${API}/replay?page=${page}&limit=${limit}`,
    get: (execId: string) => `${API}/replay/${execId}`,
    delete: (execId: string) => `${API}/replay/${execId}`,
    byFunction: (author: string, name: string, page = 1, limit = 20) =>
      `${API}/replay/function/${author}/${name}?page=${page}&limit=${limit}`,
  },

  // ========================================================================
  // Provider Endpoints
  // ========================================================================
  providers: {
    list: `${API}/providers`,
    get: (providerId: string) => `${API}/providers/${providerId}`,
    regions: (providerId: string) => `${API}/providers/${providerId}/regions`,
    status: `${API}/providers/status`,
  },

  // ========================================================================
  // Analytics Endpoints
  // ========================================================================
  analytics: {
    overview: `${API}/analytics/overview`,
    functions: (author: string, name: string) => `${API}/analytics/functions/${author}/${name}`,
    usage: `${API}/analytics/usage`,
    errors: (page = 1, limit = 20) => `${API}/analytics/errors?page=${page}&limit=${limit}`,
    latency: `${API}/analytics/latency`,
  },

  // ========================================================================
  // Settings Endpoints
  // ========================================================================
  settings: {
    account: `${API}/settings/account`,
    billing: `${API}/settings/billing`,
    apiKeys: `${API}/settings/api-keys`,
    apiKey: (keyId: string) => `${API}/settings/api-keys/${keyId}`,
    notifications: `${API}/settings/notifications`,
  },

  // ========================================================================
  // Blog/Content Endpoints
  // ========================================================================
  blog: {
    posts: {
      list: (page = 1, limit = 10) => `${API}/blog/posts?page=${page}&limit=${limit}`,
      get: (slug: string) => `${API}/blog/posts/${slug}`,
      create: `${API}/blog/posts`,
      update: (slug: string) => `${API}/blog/posts/${slug}`,
      delete: (slug: string) => `${API}/blog/posts/${slug}`,
    },
    categories: {
      list: `${API}/blog/categories`,
      get: (categoryId: string) => `${API}/blog/categories/${categoryId}`,
    },
    authors: {
      list: `${API}/blog/authors`,
      get: (authorId: string) => `${API}/blog/authors/${authorId}`,
    },
  },

  // ========================================================================
  // State Fabric Endpoints
  // ========================================================================
  stateFabric: {
    list: (page = 1, limit = 20) => `${API}/state-fabric?page=${page}&limit=${limit}`,
    get: (id: string) => `${API}/state-fabric/${id}`,
    create: `${API}/state-fabric`,
    update: (id: string) => `${API}/state-fabric/${id}`,
    delete: (id: string) => `${API}/state-fabric/${id}`,
    entries: (id: string) => `${API}/state-fabric/${id}/entries`,
  },

  // ========================================================================
  // Newsletter Endpoints (Public - no auth required)
  // ========================================================================
  newsletter: {
    subscribe: `${apiRoot.replace(/\/$/, '')}/newsletter/subscribe`,
    unsubscribe: `${apiRoot.replace(/\/$/, '')}/newsletter/unsubscribe`,
  },

  // ========================================================================
  // Founders Endpoints
  // ========================================================================
  founders: {
    status: `${API}/founders/status`,
    votes: `${API}/founders/votes`,
    earlyAccess: `${API}/founders/early-access`,
  },

  // ========================================================================
  // Execution Receipt Endpoints (Public - no auth required for read paths)
  // ========================================================================
  receipts: {
    get: (id: string) => `${API}/receipts/${id}`,
    trending: `${API}/receipts/trending`,
    byFunction: (author: string, name: string) => `${API}/receipts/function/${author}/${name}`,
    run: (id: string) => `${API}/receipts/${id}/run`,
    forkPayload: (id: string) => `${API}/receipts/${id}/fork-payload`,
    view: (id: string) => `${API}/receipts/${id}/view`,
    revoke: (id: string) => `${API}/receipts/${id}/revoke`,
  },

  // ========================================================================
  // Atlas Memory Engine Trace Endpoints
  // ========================================================================
  atlas: {
    traces: {
      list: () => `${API}/atlas/traces`,
      get: (runId: string) => `${API}/atlas/traces/${runId}`,
      graph: (runId: string) => `${API}/atlas/traces/${runId}/graph`,
      search: () => `${API}/atlas/traces/search`,
      health: () => `${API}/atlas/traces/health`,
    },
  },
} as const;

// ============================================================================
// Helper Types
// ============================================================================

export type APIUrls = typeof API_URLS;
