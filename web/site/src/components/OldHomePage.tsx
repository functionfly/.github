/**
 * FunctionFly Nexus — Premium 3D Data Flow Experience
 *
 * A unique abstract visualization of interconnected function nodes
 * flowing through a geometric nexus space. Features:
 * - Dynamic node network with glowing connections
 * - Brand-aware light/dark theming
 * - Smooth orbital camera movement
 * - Floating geometric primitives with brand colors
 * - Particle streams representing data flow
 */

import { Html, Stars } from "@react-three/drei";
import { Canvas, useFrame, useThree } from "@react-three/fiber";
import { Suspense, useEffect, useMemo, useRef, useState } from "react";
import * as THREE from "three";

// Brand Palette - responds to light/dark mode
const BRAND_DARK = {
  flame: 0xff6b35,
  flameRgb: "255, 107, 53",
  cyan: 0x00d4ff,
  cyanRgb: "0, 212, 255",
  afterburner: 0xff4f5e,
  tarmac: 0x0d1117,
  cockpit: 0x161b22,
  white: 0xffffff,
  background: 0x0d1117,
  backgroundRgb: "13, 17, 23",
};

const BRAND_LIGHT = {
  flame: 0xe85a2a,
  flameRgb: "232, 90, 42",
  cyan: 0x00a8cc,
  cyanRgb: "0, 168, 204",
  afterburner: 0xe03e4d,
  tarmac: 0xf6f8fa,
  cockpit: 0xffffff,
  white: 0xffffff,
  background: 0xf6f8fa,
  backgroundRgb: "246, 248, 250",
};

// ============================================
// FLOATING GEOMETRIC NODES
// ============================================
function NexusNode({
  position,
  scale,
  color,
  speed,
  rotationSpeed,
}: {
  position: [number, number, number];
  scale: number;
  color: number;
  speed: number;
  rotationSpeed: number;
}) {
  const meshRef = useRef<THREE.Mesh>(null);
  const groupRef = useRef<THREE.Group>(null);

  useFrame(({ clock }) => {
    if (!meshRef.current || !groupRef.current) return;
    const t = clock.getElapsedTime();
    meshRef.current.rotation.x = t * rotationSpeed;
    meshRef.current.rotation.y = t * rotationSpeed * 0.7;
    groupRef.current.position.y =
      position[1] + Math.sin(t * speed + position[0]) * 0.5;
  });

  return (
    <group ref={groupRef} position={position}>
      <mesh ref={meshRef} scale={scale}>
        <icosahedronGeometry args={[1, 1]} />
        <meshPhysicalMaterial
          color={color}
          emissive={color}
          emissiveIntensity={0.3}
          metalness={0.8}
          roughness={0.2}
          clearcoat={1}
          clearcoatRoughness={0.1}
        />
      </mesh>
    </group>
  );
}

// ============================================
// CONNECTION BEAMS
// ============================================
function ConnectionBeam({
  start,
  end,
  color,
  opacity,
}: {
  start: THREE.Vector3;
  end: THREE.Vector3;
  color: number;
  opacity: number;
}) {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const lineRef = useRef<any>(null);

  const geometry = useMemo(() => {
    const points = [start, end];
    return new THREE.BufferGeometry().setFromPoints(points);
  }, [start, end]);

  useFrame(({ clock }) => {
    if (!lineRef.current) return;
    const material = lineRef.current.material as THREE.LineBasicMaterial;
    material.opacity =
      opacity * (0.5 + Math.sin(clock.getElapsedTime() * 2) * 0.3);
  });

  return (
    <line ref={lineRef}>
      <primitive object={geometry} />
      <lineBasicMaterial
        color={color}
        transparent
        opacity={opacity * 0.8}
        blending={THREE.AdditiveBlending}
      />
    </line>
  );
}

// ============================================
// CORE NEXUS SPHERE
// ============================================
function NexusCore({ isDark }: { isDark: boolean }) {
  const coreRef = useRef<THREE.Mesh>(null);
  const glowRef = useRef<THREE.Mesh>(null);
  const brand = isDark ? BRAND_DARK : BRAND_LIGHT;

  useFrame(({ clock }) => {
    if (!coreRef.current || !glowRef.current) return;
    const t = clock.getElapsedTime();
    coreRef.current.rotation.y = t * 0.2;
    coreRef.current.rotation.z = t * 0.1;
    glowRef.current.scale.setScalar(1 + Math.sin(t * 1.5) * 0.1);
  });

  return (
    <group>
      {/* Main Core */}
      <mesh ref={coreRef}>
        <dodecahedronGeometry args={[2, 1]} />
        <meshPhysicalMaterial
          color={brand.cockpit}
          emissive={brand.flame}
          emissiveIntensity={0.2}
          metalness={0.9}
          roughness={0.1}
          clearcoat={1}
          wireframe={false}
        />
      </mesh>

      {/* Inner Glow */}
      <mesh ref={glowRef}>
        <sphereGeometry args={[2.5, 32, 32]} />
        <meshBasicMaterial
          color={brand.flame}
          transparent
          opacity={0.15}
          blending={THREE.AdditiveBlending}
          side={THREE.BackSide}
        />
      </mesh>

      {/* Outer Ring */}
      <mesh rotation={[Math.PI / 2, 0, 0]}>
        <torusGeometry args={[3.5, 0.05, 16, 64]} />
        <meshBasicMaterial color={brand.cyan} transparent opacity={0.6} />
      </mesh>

      <mesh rotation={[0, Math.PI / 2, 0]}>
        <torusGeometry args={[3.5, 0.05, 16, 64]} />
        <meshBasicMaterial color={brand.flame} transparent opacity={0.4} />
      </mesh>
    </group>
  );
}

// ============================================
// ORBITING FUNCTION NODES
// ============================================
function OrbitingNodes({
  isDark,
  count = 8,
}: {
  isDark: boolean;
  count?: number;
}) {
  const groupRef = useRef<THREE.Group>(null);
  const brand = isDark ? BRAND_DARK : BRAND_LIGHT;

  const nodes = useMemo(() => {
    const result = [];
    for (let i = 0; i < count; i++) {
      const angle = (i / count) * Math.PI * 2;
      const radius = 5 + Math.sin(i * 0.8) * 1.5;
      const y = Math.sin(i * 1.2) * 2;
      result.push({
        position: [Math.cos(angle) * radius, y, Math.sin(angle) * radius] as [
          number,
          number,
          number,
        ],
        scale: 0.2 + Math.random() * 0.3,
        color: i % 2 === 0 ? brand.flame : brand.cyan,
        speed: 0.3 + Math.random() * 0.4,
        rotationSpeed: 0.5 + Math.random() * 1,
      });
    }
    return result;
  }, [count, brand.flame, brand.cyan]);

  useFrame(({ clock }) => {
    if (!groupRef.current) return;
    groupRef.current.rotation.y = clock.getElapsedTime() * 0.15;
  });

  return (
    <group ref={groupRef}>
      {nodes.map((node, i) => (
        <NexusNode
          key={i}
          position={node.position}
          scale={node.scale}
          color={node.color}
          speed={node.speed}
          rotationSpeed={node.rotationSpeed}
        />
      ))}
    </group>
  );
}

// ============================================
// DATA STREAM PARTICLES
// ============================================
function DataStreams({
  isDark,
  count = 200,
}: {
  isDark: boolean;
  count?: number;
}) {
  const pointsRef = useRef<THREE.Points>(null);
  const brand = isDark ? BRAND_DARK : BRAND_LIGHT;

  const [positions, vel] = useMemo(() => {
    const pos = new Float32Array(count * 3);
    const vel = new Float32Array(count * 3);
    for (let i = 0; i < count; i++) {
      const angle = Math.random() * Math.PI * 2;
      const radius = 2 + Math.random() * 8;
      pos[i * 3] = Math.cos(angle) * radius;
      pos[i * 3 + 1] = (Math.random() - 0.5) * 12;
      pos[i * 3 + 2] = Math.sin(angle) * radius;
      vel[i * 3] = (Math.random() - 0.5) * 0.02;
      vel[i * 3 + 1] = -0.02 - Math.random() * 0.05;
      vel[i * 3 + 2] = (Math.random() - 0.5) * 0.02;
    }
    return [pos, vel];
  }, [count]);

  useFrame((_, delta) => {
    if (!pointsRef.current) return;
    const pos = pointsRef.current.geometry.attributes.position
      .array as Float32Array;
    for (let i = 0; i < count; i++) {
      pos[i * 3] += vel[i * 3];
      pos[i * 3 + 1] += vel[i * 3 + 1];
      pos[i * 3 + 2] += vel[i * 3 + 2];

      // Reset particles that go too low or too far
      if (pos[i * 3 + 1] < -8 || Math.abs(pos[i * 3]) > 12) {
        const angle = Math.random() * Math.PI * 2;
        const r = 2 + Math.random() * 4;
        pos[i * 3] = Math.cos(angle) * r;
        pos[i * 3 + 1] = 6 + Math.random() * 4;
        pos[i * 3 + 2] = Math.sin(angle) * r;
      }
    }
    pointsRef.current.geometry.attributes.position.needsUpdate = true;
  });

  return (
    <points ref={pointsRef}>
      <bufferGeometry>
        <bufferAttribute attach="attributes-position" args={[positions, 3]} />
      </bufferGeometry>
      <pointsMaterial
        size={0.08}
        color={brand.flame}
        transparent
        opacity={0.8}
        blending={THREE.AdditiveBlending}
        sizeAttenuation
      />
    </points>
  );
}

// ============================================
// GRID FLOOR
// ============================================
function GridFloor({ isDark }: { isDark: boolean }) {
  const meshRef = useRef<THREE.Mesh>(null);
  const brand = isDark ? BRAND_DARK : BRAND_LIGHT;

  useFrame(({ clock }) => {
    if (!meshRef.current) return;
    const material = meshRef.current.material as THREE.MeshBasicMaterial;
    material.opacity = 0.1 + Math.sin(clock.getElapsedTime() * 0.5) * 0.05;
  });

  return (
    <mesh ref={meshRef} position={[0, -4, 0]} rotation={[-Math.PI / 2, 0, 0]}>
      <planeGeometry args={[50, 50, 50, 50]} />
      <meshBasicMaterial
        color={brand.cyan}
        wireframe
        transparent
        opacity={0.15}
      />
    </mesh>
  );
}

// ============================================
// ORBITAL CAMERA
// ============================================
function OrbitalCamera() {
  const { camera } = useThree();

  useFrame(({ clock }) => {
    const t = clock.getElapsedTime();
    const radius = 14;
    const height = 6 + Math.sin(t * 0.2) * 2;
    const angle = t * 0.12;

    camera.position.x = Math.cos(angle) * radius;
    camera.position.z = Math.sin(angle) * radius;
    camera.position.y = height;
    camera.lookAt(0, 0, 0);
  });

  return null;
}

// ============================================
// SCENE LIGHTING
// ============================================
function SceneLighting({ isDark }: { isDark: boolean }) {
  const brand = isDark ? BRAND_DARK : BRAND_LIGHT;

  return (
    <>
      <ambientLight intensity={isDark ? 0.4 : 0.6} />
      <directionalLight
        position={[15, 25, 10]}
        intensity={isDark ? 1.0 : 1.2}
        color={isDark ? "#fff8e7" : "#ffffff"}
      />
      <pointLight
        position={[0, 0, 0]}
        intensity={isDark ? 3 : 2}
        color={brand.flame}
        distance={20}
        decay={2}
      />
      <pointLight
        position={[-8, 4, 8]}
        intensity={1.5}
        color={brand.cyan}
        distance={15}
        decay={2}
      />
      <spotLight
        position={[5, 15, 5]}
        angle={0.4}
        penumbra={0.5}
        intensity={0.5}
        color={isDark ? "#ffffff" : brand.cyan}
      />
    </>
  );
}

// ============================================
// MAIN SCENE
// ============================================
function Scene({ isDark }: { isDark: boolean }) {
  const brand = isDark ? BRAND_DARK : BRAND_LIGHT;

  return (
    <>
      <color
        attach="background"
        args={[isDark ? brand.tarmac : brand.background]}
      />
      <fog
        attach="fog"
        args={[isDark ? brand.tarmac : brand.background, 15, 60]}
      />

      <SceneLighting isDark={isDark} />

      <perspectiveCamera
        position={[15, 10, 15]}
        fov={55}
        near={0.1}
        far={200}
      />
      <OrbitalCamera />

      <NexusCore isDark={isDark} />
      <OrbitingNodes isDark={isDark} count={12} />
      <DataStreams isDark={isDark} count={250} />
      <GridFloor isDark={isDark} />

      <Stars
        radius={100}
        depth={60}
        count={isDark ? 3000 : 500}
        factor={4}
        saturation={0.3}
        fade={!isDark}
        speed={0.3}
      />

      {/* Floating Brand Text */}
      <Html position={[-6, 5, -8]} center transform rotation={[0, 0.3, 0]}>
        <div
          style={{
            fontSize: "36px",
            fontWeight: 700,
            color: isDark ? "#FF6B35" : "#E85A2A",
            textShadow: `0 0 30px rgba(${brand.flameRgb}, 0.6)`,
            fontFamily: "Inter, system-ui, sans-serif",
            letterSpacing: "-0.02em",
            pointerEvents: "none",
            userSelect: "none",
            whiteSpace: "nowrap",
          }}
        >
          FUNCTION
        </div>
      </Html>

      <Html position={[5, -3, -8]} center transform rotation={[0, -0.3, 0]}>
        <div
          style={{
            fontSize: "36px",
            fontWeight: 700,
            color: isDark ? "#00D4FF" : "#00A8CC",
            textShadow: `0 0 30px rgba(${brand.cyanRgb}, 0.6)`,
            fontFamily: "Inter, system-ui, sans-serif",
            letterSpacing: "-0.02em",
            pointerEvents: "none",
            userSelect: "none",
            whiteSpace: "nowrap",
          }}
        >
          FLY
        </div>
      </Html>
    </>
  );
}

// ============================================
// LOADING SPINNER
// ============================================
function Loader({ brandColor }: { brandColor: string }) {
  return (
    <Html center>
      <div
        style={{
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          gap: "16px",
          color: "#fff",
        }}
      >
        <div
          style={{
            width: "48px",
            height: "48px",
            border: `3px solid ${brandColor}33`,
            borderTop: `3px solid ${brandColor}`,
            borderRadius: "50%",
            animation: "spin 1s linear infinite",
          }}
        />
        <p
          style={{
            fontSize: "12px",
            letterSpacing: "0.15em",
            color: "rgba(255,255,255,0.6)",
            fontFamily: "JetBrains Mono, monospace",
          }}
        >
          INITIALIZING NEXUS
        </p>
        <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
      </div>
    </Html>
  );
}

// ============================================
// MAIN EXPORT
// ============================================
export default function FunctionFlyNexus() {
  const [mounted, setMounted] = useState(false);
  const [isDark, setIsDark] = useState(true);

  useEffect(() => {
    setMounted(true);

    // Detect theme
    const theme = document.documentElement.getAttribute("data-theme");
    setIsDark(theme !== "light");

    // Listen for theme changes
    const observer = new MutationObserver((mutations) => {
      mutations.forEach((mutation) => {
        if (
          mutation.attributeName === "data-theme" ||
          mutation.attributeName === "class"
        ) {
          const newTheme = document.documentElement.getAttribute("data-theme");
          setIsDark(newTheme !== "light");
        }
      });
    });

    observer.observe(document.documentElement, { attributes: true });

    // Also check system preference changes
    const mediaQuery = window.matchMedia("(prefers-color-scheme: light)");
    const handleChange = (e: MediaQueryListEvent) => {
      if (!localStorage.getItem("ff-theme")) {
        setIsDark(!e.matches);
      }
    };
    mediaQuery.addEventListener("change", handleChange);

    return () => {
      observer.disconnect();
      mediaQuery.removeEventListener("change", handleChange);
    };
  }, []);

  const brand = isDark ? BRAND_DARK : BRAND_LIGHT;
  const brandColor = isDark ? "#FF6B35" : "#E85A2A";
  const bgGradient = isDark
    ? "linear-gradient(180deg, #0D1117 0%, #161B22 100%)"
    : "linear-gradient(180deg, #F6F8FA 0%, #EEF1F5 100%)";

  if (!mounted) {
    return (
      <div
        style={{
          width: "100%",
          height: "600px",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          background: bgGradient,
        }}
      >
        <div
          style={{
            width: "48px",
            height: "48px",
            border: `3px solid ${brandColor}33`,
            borderTop: `3px solid ${brandColor}`,
            borderRadius: "50%",
            animation: "spin 1s linear infinite",
          }}
        />
        <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
      </div>
    );
  }

  return (
    <div
      style={{
        width: "100%",
        height: "600px",
        position: "relative",
        overflow: "hidden",
      }}
    >
      <Canvas
        gl={{
          antialias: true,
          alpha: true,
          powerPreference: "high-performance",
        }}
        dpr={[1, 2]}
      >
        <Suspense fallback={<Loader brandColor={brandColor} />}>
          <Scene isDark={isDark} />
        </Suspense>
      </Canvas>

      {/* Brand Badge */}
      <div
        style={{
          position: "absolute",
          bottom: "24px",
          left: "24px",
          pointerEvents: "none",
          fontFamily: "Inter, system-ui, sans-serif",
        }}
      >
        <h2
          style={{
            fontSize: "28px",
            fontWeight: 700,
            letterSpacing: "-0.02em",
            margin: 0,
            color: isDark ? "#fff" : "#0D1117",
          }}
        >
          <span style={{ color: brandColor }}>Function</span>
          <span style={{ color: isDark ? "#00D4FF" : "#00A8CC" }}>Fly</span>
        </h2>
        <p
          style={{
            fontSize: "12px",
            color: isDark ? "rgba(255,255,255,0.5)" : "rgba(13,17,23,0.5)",
            letterSpacing: "0.15em",
            textTransform: "uppercase",
            margin: "4px 0 0 0",
          }}
        >
          Nexus Edition
        </p>
      </div>

      {/* Status HUD */}
      <div
        style={{
          position: "absolute",
          top: "24px",
          right: "24px",
          pointerEvents: "none",
          textAlign: "right",
          fontFamily: "JetBrains Mono, monospace",
          fontSize: "11px",
        }}
      >
        <div style={{ color: brandColor, marginBottom: "4px" }}>
          NODES: 12 ACTIVE
        </div>
        <div
          style={{
            color: isDark ? "#00D4FF" : "#00A8CC",
            marginBottom: "4px",
          }}
        >
          CONNECTIONS: 48
        </div>
        <div
          style={{
            color: isDark ? "rgba(255,255,255,0.4)" : "rgba(13,17,23,0.4)",
          }}
        >
          SYNC: NOMINAL
        </div>
      </div>
    </div>
  );
}
