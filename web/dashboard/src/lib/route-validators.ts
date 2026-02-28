import { z } from 'zod';

// ============================================================================
// Common Route Parameter Schemas
// ============================================================================

/**
 * Valid author name: alphanumeric, underscores, hyphens, 1-50 chars
 */
export const authorSchema = z
  .string()
  .min(1, 'Author name is required')
  .max(50, 'Author name must be 50 characters or less')
  .regex(/^[a-zA-Z0-9_-]+$/, 'Invalid author name format');

/**
 * Valid function name: alphanumeric, underscores, hyphens, 1-100 chars
 */
export const functionNameSchema = z
  .string()
  .min(1, 'Function name is required')
  .max(100, 'Function name must be 100 characters or less')
  .regex(/^[a-zA-Z0-9_-]+$/, 'Invalid function name format');

/**
 * Valid UUID schema
 */
export const uuidSchema = z.string().uuid('Invalid UUID format');

/**
 * Valid slug: lowercase alphanumeric with hyphens, 1-200 chars
 */
export const slugSchema = z
  .string()
  .min(1, 'Slug is required')
  .max(200, 'Slug must be 200 characters or less')
  .regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/, 'Invalid slug format');

/**
 * Valid username: alphanumeric, underscores, 3-30 chars
 */
export const usernameSchema = z
  .string()
  .min(3, 'Username must be at least 3 characters')
  .max(30, 'Username must be 30 characters or less')
  .regex(/^[a-zA-Z0-9_]+$/, 'Invalid username format');

/**
 * Valid page number: positive integer
 */
export const pageSchema = z.coerce.number().int().positive().default(1);

/**
 * Valid limit/offset pagination
 */
export const paginationSchema = z.object({
  page: pageSchema.default(1),
  limit: z.coerce.number().int().positive().max(100).default(20),
});

/**
 * Valid email schema
 */
export const emailSchema = z.string().email('Invalid email format');

/**
 * Valid version string (semver-like)
 */
export const versionSchema = z
  .string()
  .regex(/^\d+\.\d+\.\d+(?:-[a-zA-Z0-9.-]+)?$/, 'Invalid version format');

// ============================================================================
// Route Parameter Validation Functions
// ============================================================================

/**
 * Validate route parameters against a Zod schema
 * @param params - Route parameters to validate
 * @param schema - Zod schema to validate against
 * @returns Parsed and validated parameters
 * @throws ZodError if validation fails
 */
export function validateRouteParams<T>(
  params: unknown,
  schema: z.ZodSchema<T>
): T {
  return schema.parse(params);
}

/**
 * Safely validate route parameters (returns result object instead of throwing)
 * @param params - Route parameters to validate
 * @param schema - Zod schema to validate against
 * @returns Result object with success status and data or error
 */
export function safeValidateRouteParams<T>(
  params: unknown,
  schema: z.ZodSchema<T>
): { success: true; data: T } | { success: false; error: z.ZodError } {
  const result = schema.safeParse(params);
  if (result.success) {
    return { success: true, data: result.data };
  }
  return { success: false, error: result.error };
}

/**
 * Validate author and function name together
 */
export const functionParamsSchema = z.object({
  author: authorSchema,
  name: functionNameSchema,
});

/**
 * Validate blog post parameters
 */
export const blogPostParamsSchema = z.object({
  slug: slugSchema,
});

/**
 * Validate user profile parameters
 */
export const userProfileParamsSchema = z.object({
  username: usernameSchema,
});

/**
 * Validate execution replay parameters
 */
export const replayParamsSchema = z.object({
  execId: uuidSchema,
});

/**
 * Validate function versioned parameters
 */
export const versionedFunctionParamsSchema = z.object({
  author: authorSchema,
  name: functionNameSchema,
  version: versionSchema.optional(),
});

/**
 * Validate search query parameters
 */
export const searchQuerySchema = z.object({
  q: z.string().min(1, 'Search query is required').max(200),
  page: pageSchema.default(1),
  limit: z.coerce.number().int().positive().max(100).default(20),
});

// ============================================================================
// Type Exports
// ============================================================================

export type AuthorParams = z.infer<typeof authorSchema>;
export type FunctionNameParams = z.infer<typeof functionNameSchema>;
export type UuidParams = z.infer<typeof uuidSchema>;
export type SlugParams = z.infer<typeof slugSchema>;
export type UsernameParams = z.infer<typeof usernameSchema>;
export type PageParams = z.infer<typeof pageSchema>;
export type PaginationParams = z.infer<typeof paginationSchema>;
export type EmailParams = z.infer<typeof emailSchema>;
export type VersionParams = z.infer<typeof versionSchema>;
export type FunctionParams = z.infer<typeof functionParamsSchema>;
export type BlogPostParams = z.infer<typeof blogPostParamsSchema>;
export type UserProfileParams = z.infer<typeof userProfileParamsSchema>;
export type ReplayParams = z.infer<typeof replayParamsSchema>;
export type VersionedFunctionParams = z.infer<typeof versionedFunctionParamsSchema>;
export type SearchQuery = z.infer<typeof searchQuerySchema>;
