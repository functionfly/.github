/**
 * Re-export shim — keeps existing imports working while the real Sidebar
 * has been moved to the Sidebar/ sub-folder.
 *
 * Old path: @/components/layout/Sidebar (this file)
 * New path:  @/components/layout/Sidebar/index.ts (the actual module)
 *
 * Consumers should migrate to:
 *   import { Sidebar } from '@/components/layout/Sidebar';
 * which resolves to Sidebar/index.ts → Sidebar/Sidebar.tsx
 */
export { Sidebar, type SidebarProps } from './Sidebar/index';