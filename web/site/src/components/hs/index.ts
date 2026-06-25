/**
 * Holographic Skeuomorphism Component Library
 * FunctionFly Marketing Site
 *
 * Public API for the hs/ design system. Every surface in the
 * marketing app should compose from these - nothing should
 * hand-roll its own glass or shadow.
 */

// Tokens
export { HS_TOKENS, computeShadowOffset, computeNeumorphicShadow } from './tokens';
export type { HS_TOKENS as HSTokens } from './tokens';

// Light source (context + hook)
export { LightSourceProvider, useLightSource } from './LightSourceProvider';
export type { LightSource, LightSourceProviderProps } from './LightSourceProvider';

// Glass + Neumorphic surfaces
export { GlassPanel } from './GlassPanel';
export type { GlassPanelProps, GlassDepth } from './GlassPanel';

export { NeumorphicSurface } from './NeumorphicSurface';
export type { NeumorphicSurfaceProps } from './NeumorphicSurface';

// Tactile controls
export { TactileButton } from './TactileButton';
export type { TactileButtonProps, TactileButtonVariant, TactileButtonSize } from './TactileButton';

export { TactileToggle } from './TactileToggle';
export type { TactileToggleProps } from './TactileToggle';

export { TactileSlider } from './TactileSlider';
export type { TactileSliderProps } from './TactileSlider';

export { TactileInput } from './TactileInput';
export type { TactileInputProps } from './TactileInput';

// Visual effects
export { ChromaticEdge } from './ChromaticEdge';
export type { ChromaticEdgeProps } from './ChromaticEdge';

// Accessibility
export { ReducedMotionGate } from './ReducedMotionGate';
export type { ReducedMotionGateProps } from './ReducedMotionGate';

// GPU detection and motion preferences
export {
  detectGPULevel,
  getReducedMotionPreference,
  getEnhancedVisualsPreference,
  setEnhancedVisualsPreference,
} from './utils/gpuDetect';
export type { GPULevel } from './utils/gpuDetect';

// Theme detection
export { useTheme } from './utils/useTheme';
export type { Theme } from './utils/useTheme';

// 3D effects (lazy-loaded internally)
export { RefractiveObject } from './RefractiveObject';
export type { RefractiveObjectProps, RefractiveGeometry } from './RefractiveObject';

export { HolographicBackdrop } from './HolographicBackdrop';
export type { HolographicBackdropProps } from './HolographicBackdrop';

export { ScrollParallaxCamera } from './ScrollParallaxCamera';
export type { ScrollParallaxCameraProps } from './ScrollParallaxCamera';

// Static fallbacks (always available, no lazy loading)
export { StaticGlassPanel } from './fallbacks/StaticGlassPanel';
export { StaticHolographicBackdrop } from './fallbacks/StaticHolographicBackdrop';
export { StaticRefractiveObject } from './fallbacks/StaticRefractiveObject';

// Composite hero components
export { HSHero, HSFeatureCard, HSStat } from './HSHero';
export type { HSHeroProps, HSFeatureCardProps, HSStatProps } from './HSHero';
