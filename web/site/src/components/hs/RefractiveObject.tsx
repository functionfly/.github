import React, { Suspense, useMemo, useRef, useState, useEffect } from 'react';
import { Canvas, useFrame, useThree } from '@react-three/fiber';
import { MeshTransmissionMaterial, Float, Environment } from '@react-three/drei';
import * as THREE from 'three';
import { useLightSource } from './LightSourceProvider';
import { useTheme } from './utils/useTheme';
import { StaticRefractiveObject } from './fallbacks/StaticRefractiveObject';

// NOTE: WebGPU support in @react-three/fiber is still experimental and
// inconsistent across browsers. This component deliberately uses WebGL
// (with high-performance power preference) for broad compatibility.
// The visual result of MeshTransmissionMaterial is excellent in WebGL.
// If WebGPU becomes stable and beneficial, gate it behind a feature flag
// and test on all target browsers before enabling by default.

export type RefractiveGeometry = 'box' | 'sphere' | 'torus' | 'icosahedron';

export interface RefractiveObjectProps {
  /** Geometry preset */
  geometry?: RefractiveGeometry;
  /** Material thickness (drei default: 0.2) */
  thickness?: number;
  /** Surface roughness (0 = mirror, 1 = fully rough) */
  roughness?: number;
  /** Chromatic aberration intensity */
  chromaticAberration?: number;
  /** Index of refraction (glass ~1.5) */
  ior?: number;
  /** Base color */
  color?: string;
  /** Scale factor (default 1) */
  scale?: number;
  /** Enable float animation (default true) */
  float?: boolean;
  /** Enable slow rotation (default true) */
  rotate?: boolean;
  /** Container className */
  className?: string;
  /** Container style */
  style?: React.CSSProperties;
}

const geometryComponents: Record<RefractiveGeometry, () => React.ReactElement> = {
  box: () => <boxGeometry args={[1.6, 1.6, 1.6]} />,
  sphere: () => <sphereGeometry args={[1, 64, 64]} />,
  torus: () => <torusGeometry args={[0.8, 0.3, 32, 96]} />,
  icosahedron: () => <icosahedronGeometry args={[1, 1]} />,
};

interface ObjectMeshProps {
  geometry: RefractiveGeometry;
  thickness: number;
  roughness: number;
  chromaticAberration: number;
  ior: number;
  color: string;
  rotate: boolean;
  light: { color: [number, number, number]; azimuth: number; elevation: number };
  theme: 'light' | 'dark';
}

function ObjectMesh({
  geometry,
  thickness,
  roughness,
  chromaticAberration,
  ior,
  color,
  rotate,
  light,
  theme,
}: ObjectMeshProps) {
  const meshRef = useRef<THREE.Mesh>(null);
  useFrame((_, delta) => {
    if (meshRef.current && rotate) {
      meshRef.current.rotation.x += delta * 0.15;
      meshRef.current.rotation.y += delta * 0.2;
    }
  });

  const lightColor = useMemo(
    () => new THREE.Color(light.color[0] / 255, light.color[1] / 255, light.color[2] / 255),
    [light.color]
  );

  // In light mode, increase roughness slightly so the object reads
  // better against a bright background (otherwise the glass is invisible).
  const adjustedRoughness = theme === 'light' ? Math.max(roughness, 0.2) : roughness;
  const adjustedSamples = theme === 'light' ? 4 : 6;

  return (
    <mesh ref={meshRef}>
      {geometryComponents[geometry]()}
      <MeshTransmissionMaterial
        samples={adjustedSamples}
        resolution={256}
        transmission={1}
        thickness={thickness}
        roughness={adjustedRoughness}
        ior={ior}
        chromaticAberration={chromaticAberration}
        distortion={0.2}
        distortionScale={0.3}
        temporalDistortion={0.1}
        backside
        backsideThickness={thickness * 0.5}
        color={color}
        attenuationDistance={2}
        attenuationColor={lightColor}
      />
    </mesh>
  );
}

interface CameraRigProps {
  azimuth: number;
  elevation: number;
}

function CameraRig({ azimuth, elevation }: CameraRigProps) {
  const { camera } = useThree();
  const targetPos = useRef(new THREE.Vector3(0, 0, 4));

  useEffect(() => {
    // Position camera according to light direction for consistent lighting
    const rad = (azimuth * Math.PI) / 180;
    const dist = 4 + elevation * 2;
    targetPos.current.set(Math.cos(rad) * 0.3, Math.sin(rad) * 0.3 + 1, dist);
    camera.position.copy(targetPos.current);
    camera.lookAt(0, 0, 0);
  }, [azimuth, elevation, camera]);

  return null;
}

/**
 * RefractiveObject - wraps an R3F Canvas + MeshTransmissionMaterial
 * (from drei) around a piece of geometry. This is the actual
 * real-time refraction layer - use it for a hero centerpiece,
 * loading indicator, or signature brand moment. DO NOT apply
 * everywhere - it's GPU-expensive and should be a signature element.
 *
 * Renders a static fallback during SSR and while the R3F bundle is
 * loading. Wrap in <ReducedMotionGate> to respect reduced motion,
 * the user-facing enhanced-visuals toggle, and low-GPU devices.
 */
export function RefractiveObject({
  geometry = 'icosahedron',
  thickness = 0.4,
  roughness = 0.1,
  chromaticAberration = 0.06,
  ior = 1.5,
  color = '#5fd0ff',
  scale = 1,
  float = true,
  rotate = true,
  className,
  style,
}: RefractiveObjectProps): React.ReactElement {
  const light = useLightSource();
  const theme = useTheme();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  // Avoid SSR issues: render fallback on server
  if (!mounted) {
    return (
      <StaticRefractiveObject
        color={color}
        className={className}
        style={{ transform: `scale(${scale})`, ...style }}
      />
    );
  }

  const content = (
    <div
      className={className}
      style={{
        position: 'relative',
        width: '100%',
        height: '100%',
        minHeight: 280,
        ...style,
      }}
    >
      <Canvas
        key={theme}
        camera={{ position: [0, 0, 4], fov: 45 }}
        gl={{ antialias: true, alpha: true, powerPreference: 'high-performance' }}
        dpr={[1, 1.5]}
        style={{ background: 'transparent' }}
      >
        <CameraRig azimuth={light.azimuth} elevation={light.elevation} />
        <ambientLight intensity={theme === 'light' ? 0.6 : 0.3} />
        <directionalLight position={[3, 3, 2]} intensity={theme === 'light' ? 1.5 : 1.2} color={'#ffffff'} />
        <directionalLight position={[-2, -1, 1]} intensity={0.4} color={'#5fd0ff'} />
        {float ? (
          <Float speed={1.2} rotationIntensity={0.3} floatIntensity={0.6}>
            <ObjectMesh
              geometry={geometry}
              thickness={thickness}
              roughness={roughness}
              chromaticAberration={chromaticAberration}
              ior={ior}
              color={color}
              rotate={rotate}
              light={light}
              theme={theme}
            />
          </Float>
        ) : (
          <ObjectMesh
            geometry={geometry}
            thickness={thickness}
            roughness={roughness}
            chromaticAberration={chromaticAberration}
            ior={ior}
            color={color}
            rotate={rotate}
            light={light}
            theme={theme}
          />
        )}
        <Environment preset={theme === 'light' ? 'apartment' : 'city'} />
      </Canvas>
    </div>
  );

  return (
    <Suspense
      fallback={
        <StaticRefractiveObject
          color={color}
          className={className}
          style={{ transform: `scale(${scale})`, ...style }}
        />
      }
    >
      {content}
    </Suspense>
  );
}
