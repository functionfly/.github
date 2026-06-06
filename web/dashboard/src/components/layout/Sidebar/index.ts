/**
 * Sidebar module — re-exports the main Sidebar component.
 *
 * File layout:
 *   Sidebar/
 *   ├── index.ts         ← re-exports (import from here)
 *   ├── navigation.ts    ← nav config, types, icons, constants
 *   ├── helpers.ts       ← pure utilities and shared hooks
 *   └── Sidebar.tsx      ← main component
 */

export {
  createRateLimitedHandler, escapeHtml, isItemActive,
  translateLabel,
  useDebounce
} from './helpers';
export {
  NAV_LABEL_KEYS, navigationSections, SECTION_VARIANTS
} from './navigation';
export { Sidebar, type SidebarProps } from './Sidebar';
