/**
 * GPU tier detection and reduced-motion preference.
 * Used by ReducedMotionGate to decide what to render.
 */

export type GPULevel = 'low' | 'medium' | 'high';

export interface GPUDetectionResult {
  level: GPULevel;
  reason: string;
  cores: number;
  screenSize: { width: number; height: number } | null;
  prefersReducedMotion: boolean;
}

/**
 * Detect a rough GPU level based on hardware hints.
 * - hardwareConcurrency (CPU cores - proxy for system capability)
 * - screen size (small mobile = low)
 * - device memory (if available via navigator.deviceMemory)
 *
 * This is a heuristic, not a benchmark. It's good enough to decide
 * whether to ship expensive shaders or static fallbacks.
 */
export function detectGPULevel(): GPULevel {
  if (typeof window === 'undefined') return 'high'; // SSR default

  const cores = navigator.hardwareConcurrency || 4;
  const memory = (navigator as Navigator & { deviceMemory?: number }).deviceMemory;
  const screen = window.screen;

  // Tiny screens (older mobile) -> low
  const isSmallScreen =
    screen && (screen.width < 768 || (screen.height < 768 && screen.width < 500));

  if (isSmallScreen) return 'low';
  if (memory !== undefined && memory <= 2) return 'low';
  if (cores <= 2) return 'low';

  if (memory !== undefined && memory <= 4) return 'medium';
  if (cores <= 4) return 'medium';

  return 'high';
}

/**
 * Read prefers-reduced-motion media query. Returns false on SSR.
 */
export function getReducedMotionPreference(): boolean {
  if (typeof window === 'undefined') return false;
  if (!window.matchMedia) return false;
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

const ENHANCED_VISUALS_KEY = 'ff-enhanced-visuals';

/**
 * Read the user-facing "Enhanced visuals" preference.
 * Default: true (effects on). If the OS requests reduced motion,
 * default to false unless the user explicitly opted back in.
 */
export function getEnhancedVisualsPreference(): boolean {
  if (typeof window === 'undefined') return true; // SSR default

  const stored = window.localStorage.getItem(ENHANCED_VISUALS_KEY);
  if (stored === 'false') return false;
  if (stored === 'true') return true;

  // No explicit choice: follow OS reduced-motion preference
  return !getReducedMotionPreference();
}

/**
 * Persist the "Enhanced visuals" user preference.
 */
export function setEnhancedVisualsPreference(enabled: boolean): void {
  if (typeof window === 'undefined') return;
  window.localStorage.setItem(ENHANCED_VISUALS_KEY, String(enabled));
  window.dispatchEvent(new CustomEvent('ff-enhanced-visuals-change', { detail: enabled }));
}

/**
 * Full detection result. Useful for diagnostics.
 */
export function getFullGPUInfo(): GPUDetectionResult {
  return {
    level: detectGPULevel(),
    reason: 'heuristic',
    cores: navigator.hardwareConcurrency || 0,
    screenSize:
      typeof window !== 'undefined' && window.screen
        ? { width: window.screen.width, height: window.screen.height }
        : null,
    prefersReducedMotion: getReducedMotionPreference(),
  };
}
