/**
 * Holographic Skeuomorphism Design Tokens (JS/TS)
 * FunctionFly Marketing Site
 *
 * MUST match hs-tokens.css exactly. Single source of truth for
 * values consumed by R3F shaders, Framer Motion, and inline JS.
 */

export const HS_TOKENS = {
  light: {
    azimuth: 135, // degrees
    elevation: 0.65, // 0-1
    color: [255, 245, 230] as [number, number, number],
  },
  glass: {
    blur: 18, // px
    opacity: 0.14,
    border: 'rgba(255, 255, 255, 0.18)',
    bg: 'rgba(22, 27, 34, 0.55)',
    bgDeep: 'rgba(22, 27, 34, 0.75)',
  },
  neumorphic: {
    distance: 6, // px
    blur: 16, // px
    intensity: 0.35,
    shadowLight: 'rgba(255, 245, 230, 0.04)',
    shadowDark: 'rgba(0, 0, 0, 0.5)',
  },
  radii: {
    sm: '10px',
    md: '18px',
    lg: '28px',
    pill: '9999px',
  },
  motion: {
    easeInOut: [0.4, 0, 0.2, 1] as [number, number, number, number],
    easeOut: [0, 0, 0.2, 1] as [number, number, number, number],
    easeSpring: [0.34, 1.56, 0.64, 1] as [number, number, number, number],
    fast: 0.15,
    normal: 0.25,
    slow: 0.4,
  },
  colors: {
    bg: '#0b0e14',
    surface: '#161b26',
    surfaceElevated: '#1f2533',
    accent: '#5fd0ff',
    accentWarm: '#ff9d6c',
    text: '#eef2f7',
    textDim: '#9aa6b8',
  },
} as const;

export type HS_TOKENS = typeof HS_TOKENS;

/**
 * Compute shadow offset from azimuth (degrees) and elevation (0-1).
 * - azimuth: 0deg = light from right, 90deg = from bottom, 135deg = from bottom-left
 * - elevation: 0 = grazing (long shadows), 1 = overhead (no shadow)
 */
export function computeShadowOffset(
  azimuth: number,
  elevation: number,
  distance: number = 6
): { x: number; y: number; blur: number } {
  const rad = (azimuth * Math.PI) / 180;
  const offsetMultiplier = 1 - elevation * 0.6; // higher elevation = shorter shadow
  return {
    x: Math.cos(rad) * distance * offsetMultiplier,
    y: Math.sin(rad) * distance * offsetMultiplier,
    blur: distance * 2.5,
  };
}

/**
 * Compute box-shadow string from light source values.
 * Returns the dual-shadow string for neumorphic surfaces
 * (light side highlight + dark side falloff).
 */
export function computeNeumorphicShadow(
  azimuth: number = 135,
  elevation: number = 0.65,
  intensity: number = 0.35,
  color: [number, number, number] = [255, 245, 230]
): string {
  const offset = computeShadowOffset(azimuth, elevation, 6);
  const lightSide = `${-offset.x}px ${-offset.y}px ${offset.blur}px rgba(${color.join(',')}, ${intensity * 0.15})`;
  const darkSide = `${offset.x}px ${offset.y}px ${offset.blur}px rgba(0, 0, 0, ${intensity})`;
  return `${lightSide}, ${darkSide}`;
}
