import queryString from 'query-string';
import slugify from 'slugify';
import validator from 'validator';

// ============================================================================
// Query Parameter Utilities
// ============================================================================

/**
 * Parse query string parameters from a URL search string
 * @param search - The search string (e.g., window.location.search)
 * @returns Record of query parameter key-value pairs
 */
export function parseQueryParams(search: string): Record<string, string> {
  return queryString.parse(search) as Record<string, string>;
}

/**
 * Build a query string from an object of parameters
 * @param params - Object of query parameters
 * @returns Encoded query string
 */
export function buildQueryString(params: Record<string, unknown>): string {
  return queryString.stringify(params);
}

/**
 * Parse URL and extract components
 * @param url - The URL to parse
 * @returns Parsed URL object
 */
export function parseUrl(url: string): queryString.ParsedUrl {
  return queryString.parseUrl(url);
}

// ============================================================================
// Slug Utilities
// ============================================================================

export interface CreateSlugOptions {
  /** Maximum length of the slug */
  maxLength?: number;
  /** Character separator (default: '-') */
  separator?: string;
  /** Whether to lowercase (default: true) */
  lower?: boolean;
}

/**
 * Create a URL-friendly slug from text
 * @param text - The text to convert to a slug
 * @param options - Configuration options
 * @returns URL-friendly slug
 */
export function createSlug(
  text: string,
  options?: CreateSlugOptions
): string {
  const separator = options?.separator || '-';
  const maxLength = options?.maxLength || 100;
  const lower = options?.lower ?? true;

  const slug = slugify(text, {
    lower,
    strict: true,
    // slugify's option is `replacement`, not `separator`
    replacement: separator,
    trim: true,
  });

  return slug.slice(0, maxLength);
}

// ============================================================================
// URL Validation & Sanitization
// ============================================================================

export interface IsValidUrlOptions {
  /** Require a protocol (default: true) */
  require_protocol?: boolean;
  /** Valid protocols (default: ['http', 'https']) */
  protocols?: string[];
  /** Allow hyphen in host (default: false) */
  allow_hyphen_in_host?: boolean;
}

/**
 * Validate if a string is a valid URL
 * @param url - The URL to validate
 * @param options - Validation options
 * @returns Whether the URL is valid
 */
export function isValidUrl(
  url: string,
  options?: IsValidUrlOptions
): boolean {
  return validator.isURL(url, {
    require_protocol: options?.require_protocol ?? true,
    require_valid_protocol: true,
    protocols: options?.protocols || ['http', 'https'],
    allow_hyphen_in_host: options?.allow_hyphen_in_host ?? false,
  });
}

/**
 * Sanitize a URL by escaping special characters
 * @param url - The URL to sanitize
 * @returns Sanitized URL string
 */
export function sanitizeUrl(url: string): string {
  return validator.trim(validator.escape(url));
}

/**
 * Validate and sanitize an email address
 * @param email - The email to validate
 * @returns Whether the email is valid
 */
export function isValidEmail(email: string): boolean {
  return validator.isEmail(email);
}

/**
 * Check if a string is a valid UUID
 * @param uuid - The string to check
 * @returns Whether the string is a valid UUID
 */
export function isValidUuid(uuid: string): boolean {
  return validator.isUUID(uuid);
}

// ============================================================================
// Path Building Utilities
// ============================================================================

/**
 * Build a clean path from base and parts
 * @param base - Base path
 * @param parts - Path parts to append
 * @returns Cleaned and joined path
 */
export function buildPath(base: string, ...parts: string[]): string {
  const cleanParts = parts
    .filter(Boolean)
    .map((p) => p.replace(/^\/|\/$/g, ''));
  return `/${[base, ...cleanParts].filter(Boolean).join('/')}`;
}

/**
 * Join multiple path segments safely
 * @param paths - Path segments to join
 * @returns Joined path
 */
export function joinPaths(...paths: string[]): string {
  return paths
    .filter(Boolean)
    .map((p) => p.replace(/^\/|\/$/g, ''))
    .filter(Boolean)
    .join('/');
}

/**
 * Get the current URL path without query parameters
 * @returns Current path
 */
export function getCurrentPath(): string {
  return window.location.pathname;
}

/**
 * Get the current URL search params
 * @returns Current search params
 */
export function getCurrentSearchParams(): URLSearchParams {
  return new URLSearchParams(window.location.search);
}

// ============================================================================
// Redirect & Navigation
// ============================================================================

/**
 * Navigate to a URL
 * @param url - Target URL
 * @param openInNewTab - Whether to open in a new tab
 */
export function navigateTo(url: string, openInNewTab = false): void {
  if (openInNewTab) {
    window.open(url, '_blank', 'noopener,noreferrer');
  } else {
    window.location.href = url;
  }
}

/**
 * Redirect to a relative path
 * @param path - Target path
 */
export function redirectTo(path: string): void {
  window.location.href = path;
}

// ============================================================================
// URL Encoding/Decoding
// ============================================================================

/**
 * Safely encode a URI component
 * @param value - Value to encode
 * @returns Encoded value
 */
export function encodeParam(value: string): string {
  return encodeURIComponent(value);
}

/**
 * Safely decode a URI component
 * @param value - Value to decode
 * @returns Decoded value
 */
export function decodeParam(value: string): string {
  return decodeURIComponent(value);
}
