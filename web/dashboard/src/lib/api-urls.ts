import { API_BASE_URL } from './constants';

/**
 * API Version
 */
const API_VERSION = 'v1';

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
  },

  // ========================================================================
  // Function Registry Endpoints
  // ========================================================================
  functions: {
    list: (page = 1, limit = 20) =>
      `${API}/registry/functions?page=${page}&limit=${limit}`,
    get: (author: string, name: string) =>
      `${API}/registry/functions/${author}/${name}`,
    create: `${API}/registry/functions`,
    update: (author: string, name: string) =>
      `${API}/registry/functions/${author}/${name}`,
    delete: (author: string, name: string) =>
      `${API}/registry/functions/${author}/${name}`,
    search: (query: string, page = 1, limit = 20) =>
      `${API}/registry/search?q=${encodeURIComponent(query)}&page=${page}&limit=${limit}`,
    byAuthor: (author: string, page = 1, limit = 20) =>
      `${API}/registry/authors/${author}/functions?page=${page}&limit=${limit}`,
    versions: (author: string, name: string) =>
      `${API}/registry/functions/${author}/${name}/versions`,
    latestVersion: (author: string, name: string) =>
      `${API}/registry/functions/${author}/${name}/latest`,
  },

  // ========================================================================
  // Function Execution Endpoints
  // ========================================================================
  execution: {
    execute: (author: string, name: string) =>
      `${API}/fx/${author}/${name}`,
    executeWithVersion: (author: string, name: string, version: string) =>
      `${API}/fx/${author}/${name}@${version}`,
    executeLatest: (author: string, name: string) =>
      `${API}/fx/${author}/${name}/latest`,
    batchExecute: `${API}/fx/batch`,
  },

  // ========================================================================
  // Execution Replay/History Endpoints
  // ========================================================================
  replay: {
    list: (page = 1, limit = 20) =>
      `${API}/replay?page=${page}&limit=${limit}`,
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
    functions: (author: string, name: string) =>
      `${API}/analytics/functions/${author}/${name}`,
    usage: `${API}/analytics/usage`,
    errors: (page = 1, limit = 20) =>
      `${API}/analytics/errors?page=${page}&limit=${limit}`,
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
  // Admin Endpoints
  // ========================================================================
  admin: {
    tenants: {
      list: (page = 1, limit = 20) =>
        `/admin/v1/tenants?page=${page}&limit=${limit}`,
      get: (tenantId: string) => `/admin/v1/tenants/${tenantId}`,
      create: `/admin/v1/tenants`,
      update: (tenantId: string) => `/admin/v1/tenants/${tenantId}`,
      delete: (tenantId: string) => `/admin/v1/tenants/${tenantId}`,
    },
    users: {
      list: (page = 1, limit = 20) =>
        `/admin/v1/users?page=${page}&limit=${limit}`,
      get: (userId: string) => `/admin/v1/users/${userId}`,
      update: (userId: string) => `/admin/v1/users/${userId}`,
      delete: (userId: string) => `/admin/v1/users/${userId}`,
    },
    functions: {
      list: (page = 1, limit = 20) =>
        `/admin/v1/functions?page=${page}&limit=${limit}`,
      get: (functionId: string) => `/admin/v1/functions/${functionId}`,
      update: (functionId: string) => `/admin/v1/functions/${functionId}`,
      delete: (functionId: string) => `/admin/v1/functions/${functionId}`,
    },
    system: {
      health: `/admin/v1/system/health`,
      metrics: `/admin/v1/system/metrics`,
      config: `/admin/v1/system/config`,
    },
  },

  // ========================================================================
  // Blog/Content Endpoints
  // ========================================================================
  blog: {
    posts: {
      list: (page = 1, limit = 10) =>
        `${API}/blog/posts?page=${page}&limit=${limit}`,
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
    list: (page = 1, limit = 20) =>
      `${API}/state-fabric?page=${page}&limit=${limit}`,
    get: (id: string) => `${API}/state-fabric/${id}`,
    create: `${API}/state-fabric`,
    update: (id: string) => `${API}/state-fabric/${id}`,
    delete: (id: string) => `${API}/state-fabric/${id}`,
    entries: (id: string) => `${API}/state-fabric/${id}/entries`,
  },
} as const;

// ============================================================================
// Helper Types
// ============================================================================

export type APIUrls = typeof API_URLS;
