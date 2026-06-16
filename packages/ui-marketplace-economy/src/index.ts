/**
 * @functionfly/ui-marketplace-economy
 * Marketplace Economy Components - Type definitions
 *
 * Types are defined inline here (rather than re-exported from a separate
 * file) so that consumers using `import type {...}` can resolve them
 * without depending on bundler resolution of re-exports, which can
 * silently fail in the bun workspace cache.
 *
 * To import a component, use the package subpath:
 *   import { CreatorEconomy } from '@functionfly/ui-marketplace-economy/components'
 */

// Re-export the original types.ts content so the file remains the
// source of truth for shared component prop interfaces.
export type * from './types';
