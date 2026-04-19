/**
 * 3D Font preloading utility for @react-three/drei Text components
 * Prevents infinite loading by ensuring fonts are available
 */

import { useLoader } from '@react-three/fiber';
import { FontLoader } from 'three/examples/jsm/loaders/FontLoader.js';

// Use a reliable TTF-based font URL that works with drei
// Using Helvetiker font from three.js examples (TTF-based, not woff2)
const DEFAULT_FONT_URL = 'https://unpkg.com/three@0.160.0/examples/fonts/helvetiker_bold.typeface.json';

// Local fallback path
const LOCAL_FONT_PATH = '/fonts/helvetiker_bold.typeface.json';

/**
 * Preload font for use with drei Text component
 * Call this in your scene component to ensure font is loaded before Text renders
 */
export function usePreloadFont() {
  try {
    // Try to preload the font - this will be cached
    const font = useLoader(FontLoader, DEFAULT_FONT_URL);
    return font;
  } catch (e) {
    // Fallback to local if CDN fails
    try {
      return useLoader(FontLoader, LOCAL_FONT_PATH);
    } catch (e2) {
      console.warn('Failed to load font:', e2);
      return null;
    }
  }
}

/**
 * Font URL to use with drei Text component font prop
 */
export const FONT_URL = DEFAULT_FONT_URL;
