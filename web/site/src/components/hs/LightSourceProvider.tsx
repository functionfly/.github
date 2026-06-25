import React, { createContext, useContext, useMemo, type ReactNode } from 'react';
import { computeShadowOffset, computeNeumorphicShadow, HS_TOKENS } from './tokens';

/**
 * The single virtual light source. Every glass surface, every shadow,
 * and every shader reads from this context. If you ever need to support
 * a "time of day" or dark/light mode shift, this is the only place that
 * changes.
 */
export interface LightSource {
  azimuth: number; // 0-360 degrees
  elevation: number; // 0-1
  color: [number, number, number];
  shadowOffset: { x: number; y: number; blur: number };
  highlightDirection: string; // CSS gradient angle string, e.g. "135deg"
  neumorphicShadow: string; // box-shadow string
}

export interface LightSourceProviderProps {
  children: ReactNode;
  /** Override the default light source values */
  overrides?: Partial<Omit<LightSource, 'shadowOffset' | 'highlightDirection' | 'neumorphicShadow'>>;
}

const defaultLightSource: LightSource = (() => {
  const { azimuth, elevation, color } = HS_TOKENS.light;
  return {
    azimuth,
    elevation,
    color,
    shadowOffset: computeShadowOffset(azimuth, elevation, 6),
    highlightDirection: `${azimuth}deg`,
    neumorphicShadow: computeNeumorphicShadow(azimuth, elevation),
  };
})();

const LightSourceContext = createContext<LightSource>(defaultLightSource);

/**
 * Provider that holds the single virtual light source. Wrap your app
 * (or a section) once with this. All hs/* components read from context.
 */
export function LightSourceProvider({
  children,
  overrides,
}: LightSourceProviderProps): React.ReactElement {
  const value = useMemo<LightSource>(() => {
    const azimuth = overrides?.azimuth ?? defaultLightSource.azimuth;
    const elevation = overrides?.elevation ?? defaultLightSource.elevation;
    const color = overrides?.color ?? defaultLightSource.color;
    return {
      azimuth,
      elevation,
      color,
      shadowOffset: computeShadowOffset(azimuth, elevation, 6),
      highlightDirection: `${azimuth}deg`,
      neumorphicShadow: computeNeumorphicShadow(azimuth, elevation),
    };
  }, [overrides?.azimuth, overrides?.elevation, overrides?.color]);

  return (
    <LightSourceContext.Provider value={value}>
      {children}
    </LightSourceContext.Provider>
  );
}

/**
 * Read the current light source. Use this in any hs/* component
 * that needs to align its shadow direction with the global light.
 */
export function useLightSource(): LightSource {
  return useContext(LightSourceContext);
}
