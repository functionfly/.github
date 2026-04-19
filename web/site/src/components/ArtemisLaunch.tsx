/**
 * ArtemisLaunchScene - Cinematic Rocket Launch Experience
 *
 * A dramatic Artemis-style rocket launch into outer space with:
 * - Detailed rocket model with realistic materials
 * - Multi-stage plume effects with particle systems
 * - Launch pad atmosphere with smoke/dust
 * - Space transition with Earth view
 * - Cinematic camera movements
 */

import { Html, Stars } from "@react-three/drei";
import { Canvas, useFrame, useThree } from "@react-three/fiber";
import { Suspense, useEffect, useMemo, useRef, useState } from "react";
import * as THREE from "three";

// Brand Palette
const BRAND = {
  flame: 0xff6b35,
  cyan: 0x00d4ff,
  stratosphere: 0x5b7cf5,
  space: 0x0a0a1a,
  white: 0xffffff,
  orange: 0xff4500,
  yellow: 0xffd700,
};

// ============================================
// ROCKET MODEL - Artemis-inspired Design
// ============================================
function ArtemisRocket({
  stage,
  progress,
}: {
  stage: "launch" | "ascend" | "space";
  progress: number;
}) {
  const rocketRef = useRef<THREE.Group>(null);

  const materials = useMemo(
    () => ({
      body: new THREE.MeshStandardMaterial({
        color: 0x1a1a2e,
        metalness: 0.7,
        roughness: 0.3,
      }),
      white: new THREE.MeshStandardMaterial({
        color: 0xffffff,
        metalness: 0.2,
        roughness: 0.5,
      }),
      orange: new THREE.MeshStandardMaterial({
        color: 0xff6b35,
        metalness: 0.3,
        roughness: 0.4,
      }),
      engine: new THREE.MeshStandardMaterial({
        color: 0x2a2a3a,
        metalness: 0.9,
        roughness: 0.2,
      }),
      glowOrange: new THREE.MeshBasicMaterial({
        color: 0xff4500,
        transparent: true,
        opacity: 0.8,
      }),
      grid: new THREE.MeshStandardMaterial({
        color: 0x333344,
        metalness: 0.8,
        roughness: 0.4,
        wireframe: true,
      }),
    }),
    [],
  );

  const engineGlowRef = useRef<THREE.PointLight>(null);

  useFrame(({ clock }) => {
    if (!rocketRef.current) return;
    const t = clock.getElapsedTime();

    if (stage === "launch") {
      rocketRef.current.position.y = 0;
      rocketRef.current.rotation.x = Math.sin(t * 30) * 0.002;
      rocketRef.current.rotation.z = Math.cos(t * 25) * 0.002;
    }

    if (stage === "ascend") {
      const yPos = progress * 50;
      rocketRef.current.position.y = yPos;
      rocketRef.current.rotation.x = Math.sin(t * 10) * 0.005;
    }

    if (stage === "space") {
      rocketRef.current.position.y = 50 + progress * 20;
      rocketRef.current.rotation.y = t * 0.1;
    }

    if (engineGlowRef.current) {
      const intensity =
        stage === "launch" || stage === "ascend" ? 5 + Math.sin(t * 15) * 2 : 0;
      engineGlowRef.current.intensity = intensity;
    }
  });

  const visibleEngine = stage !== "space";

  return (
    <group ref={rocketRef} scale={0.6}>
      {/* Payload Fairing */}
      <mesh position={[0, 9, 0]} material={materials.white}>
        <cylinderGeometry args={[1.2, 1.5, 3, 32]} />
      </mesh>

      {/* Upper Stage */}
      <mesh position={[0, 6.5, 0]} material={materials.white}>
        <cylinderGeometry args={[1.5, 1.2, 3, 32]} />
      </mesh>

      {/* Interstage Ring */}
      <mesh position={[0, 4.5, 0]} material={materials.orange}>
        <cylinderGeometry args={[1.6, 1.5, 1, 32]} />
      </mesh>

      {/* Main Stage */}
      <mesh position={[0, 0, 0]} material={materials.white}>
        <cylinderGeometry args={[1.6, 1.6, 8, 32]} />
      </mesh>

      {/* Engine Section */}
      <mesh position={[0, -5, 0]} material={materials.engine}>
        <cylinderGeometry args={[1.6, 1.3, 2, 32]} />
      </mesh>

      {/* Nose Cone */}
      <mesh position={[0, 11.5, 0]} material={materials.white}>
        <coneGeometry args={[1.2, 3, 32]} />
      </mesh>

      {/* Orange Stripe */}
      <mesh position={[0, 2, 0]} material={materials.orange}>
        <cylinderGeometry args={[1.61, 1.61, 0.3, 32]} />
      </mesh>

      {/* Grid Fins */}
      <mesh position={[0, 3, 1.7]} material={materials.grid}>
        <boxGeometry args={[0.1, 1.5, 1.2]} />
      </mesh>
      <mesh position={[0, 3, -1.7]} material={materials.grid}>
        <boxGeometry args={[0.1, 1.5, 1.2]} />
      </mesh>

      {/* Engines */}
      {visibleEngine && (
        <group position={[0, -6, 0]}>
          {/* RS-25 Engines */}
          {[
            [0.6, 0, 0.6],
            [-0.6, 0, 0.6],
            [0.6, 0, -0.6],
            [-0.6, 0, -0.6],
          ].map(([x, y, z], i) => (
            <group key={i} position={[x as number, y as number, z as number]}>
              <mesh rotation={[Math.PI, 0, 0]}>
                <coneGeometry args={[0.4, 1.2, 16]} />
                <meshStandardMaterial
                  color={0x2a2a3a}
                  metalness={0.9}
                  roughness={0.2}
                />
              </mesh>
            </group>
          ))}

          {/* Engine Glow */}
          <mesh position={[0, -1.5, 0]} material={materials.glowOrange}>
            <sphereGeometry args={[1.2, 32, 32]} />
          </mesh>
          <pointLight
            ref={engineGlowRef}
            position={[0, -2, 0]}
            color={BRAND.orange}
            intensity={5}
            distance={30}
            decay={2}
          />
        </group>
      )}

      {/* Side Boosters */}
      <mesh position={[1.8, 0, 0]} material={materials.orange}>
        <cylinderGeometry args={[0.1, 0.1, 8, 8]} />
      </mesh>
      <mesh position={[-1.8, 0, 0]} material={materials.orange}>
        <cylinderGeometry args={[0.1, 0.1, 8, 8]} />
      </mesh>
    </group>
  );
}

// ============================================
// ENGINE PLUME
// ============================================
function EnginePlume({
  intensity,
  visible,
}: {
  intensity: number;
  visible: boolean;
}) {
  const plumeRef = useRef<THREE.Group>(null);
  const particlesRef = useRef<THREE.Points>(null);
  const particleCount = 500;

  const [positions, velocities] = useMemo(() => {
    const pos = new Float32Array(particleCount * 3);
    const vel = new Float32Array(particleCount);
    for (let i = 0; i < particleCount; i++) {
      const angle = Math.random() * Math.PI * 2;
      const radius = Math.random() * 0.8;
      pos[i * 3] = Math.cos(angle) * radius;
      pos[i * 3 + 1] = -Math.random() * 8;
      pos[i * 3 + 2] = Math.sin(angle) * radius;
      vel[i] = 0.1 + Math.random() * 0.3;
    }
    return [pos, vel];
  }, []);

  useFrame((_, delta) => {
    if (!plumeRef.current || !particlesRef.current || !visible) return;
    const pos = particlesRef.current.geometry.attributes.position
      .array as Float32Array;
    for (let i = 0; i < particleCount; i++) {
      pos[i * 3 + 1] -= velocities[i] * delta * 20;
      pos[i * 3] += (Math.random() - 0.5) * 0.1;
      pos[i * 3 + 2] += (Math.random() - 0.5) * 0.1;
      if (pos[i * 3 + 1] < -15) {
        const angle = Math.random() * Math.PI * 2;
        const radius = Math.random() * 0.5;
        pos[i * 3] = Math.cos(angle) * radius;
        pos[i * 3 + 1] = 0;
        pos[i * 3 + 2] = Math.sin(angle) * radius;
      }
    }
    particlesRef.current.geometry.attributes.position.needsUpdate = true;
  });

  if (!visible) return null;

  return (
    <group ref={plumeRef} position={[0, -8, 0]} scale={0.6}>
      {/* Core Flame */}
      <mesh>
        <coneGeometry args={[1.5, 10, 32, 1, true]} />
        <meshBasicMaterial
          color={BRAND.orange}
          transparent
          opacity={0.8 * intensity}
          side={THREE.DoubleSide}
          blending={THREE.AdditiveBlending}
        />
      </mesh>

      {/* White Hot Core */}
      <mesh position={[0, -2, 0]}>
        <coneGeometry args={[0.6, 6, 32, 1, true]} />
        <meshBasicMaterial
          color={0xffffff}
          transparent
          opacity={0.9 * intensity}
          blending={THREE.AdditiveBlending}
        />
      </mesh>

      {/* Shock Diamonds */}
      {[1.5, 3, 4.5].map((y, i) => (
        <mesh key={i} position={[0, -y, 0]}>
          <sphereGeometry args={[0.3 - i * 0.08, 16, 16]} />
          <meshBasicMaterial
            color={0xffffff}
            transparent
            opacity={0.6 * intensity}
            blending={THREE.AdditiveBlending}
          />
        </mesh>
      ))}

      {/* Particles */}
      <points ref={particlesRef}>
        <bufferGeometry>
          <bufferAttribute
            attach="attributes-position"
            count={particleCount}
            array={positions}
            itemSize={3}
          />
        </bufferGeometry>
        <pointsMaterial
          size={0.15}
          color={BRAND.yellow}
          transparent
          opacity={0.8 * intensity}
          blending={THREE.AdditiveBlending}
          sizeAttenuation
        />
      </points>
    </group>
  );
}

// ============================================
// LAUNCH PAD
// ============================================
function LaunchPad({ visible }: { visible: boolean }) {
  const smokeRef = useRef<THREE.Points>(null);
  const smokeParticles = 300;

  const [positions, velocities] = useMemo(() => {
    const pos = new Float32Array(smokeParticles * 3);
    const vel = new Float32Array(smokeParticles * 3);
    for (let i = 0; i < smokeParticles; i++) {
      const angle = Math.random() * Math.PI * 2;
      const radius = 2 + Math.random() * 5;
      pos[i * 3] = Math.cos(angle) * radius;
      pos[i * 3 + 1] = Math.random() * 3;
      pos[i * 3 + 2] = Math.sin(angle) * radius;
      vel[i * 3] = (Math.random() - 0.5) * 0.5;
      vel[i * 3 + 1] = 0.5 + Math.random() * 1;
      vel[i * 3 + 2] = (Math.random() - 0.5) * 0.5;
    }
    return [pos, vel];
  }, []);

  useFrame((_, delta) => {
    if (!smokeRef.current || !visible) return;
    const pos = smokeRef.current.geometry.attributes.position
      .array as Float32Array;
    for (let i = 0; i < smokeParticles; i++) {
      pos[i * 3] += velocities[i * 3] * delta * 2;
      pos[i * 3 + 1] += velocities[i * 3 + 1] * delta * 3;
      pos[i * 3 + 2] += velocities[i * 3 + 2] * delta * 2;
      if (pos[i * 3 + 1] > 10) {
        const angle = Math.random() * Math.PI * 2;
        const radius = 2 + Math.random() * 3;
        pos[i * 3] = Math.cos(angle) * radius;
        pos[i * 3 + 1] = 0;
        pos[i * 3 + 2] = Math.sin(angle) * radius;
      }
    }
    smokeRef.current.geometry.attributes.position.needsUpdate = true;
  });

  if (!visible) return null;

  return (
    <group position={[0, -15, 0]}>
      {/* Launch Tower */}
      <mesh position={[5, 5, 0]}>
        <boxGeometry args={[0.5, 20, 0.5]} />
        <meshStandardMaterial
          color={0x444444}
          metalness={0.8}
          roughness={0.4}
        />
      </mesh>

      {/* Support Structure */}
      <mesh position={[0, -2, 0]}>
        <cylinderGeometry args={[4, 5, 2, 32]} />
        <meshStandardMaterial
          color={0x333333}
          metalness={0.6}
          roughness={0.5}
        />
      </mesh>

      {/* Ground */}
      <mesh position={[0, -3.5, 0]} rotation={[-Math.PI / 2, 0, 0]}>
        <circleGeometry args={[20, 64]} />
        <meshStandardMaterial
          color={0x1a1a1a}
          metalness={0.2}
          roughness={0.9}
        />
      </mesh>

      {/* Smoke */}
      <points ref={smokeRef}>
        <bufferGeometry>
          <bufferAttribute
            attach="attributes-position"
            count={smokeParticles}
            array={positions}
            itemSize={3}
          />
        </bufferGeometry>
        <pointsMaterial
          size={0.5}
          color={0x888888}
          transparent
          opacity={0.6}
          sizeAttenuation
        />
      </points>
    </group>
  );
}

// ============================================
// EARTH
// ============================================
function Earth({ visible }: { visible: boolean }) {
  const earthRef = useRef<THREE.Mesh>(null);

  useFrame(({ clock }) => {
    if (!earthRef.current || !visible) return;
    earthRef.current.rotation.y = clock.getElapsedTime() * 0.05;
  });

  if (!visible) return null;

  return (
    <group position={[30, -40, -80]} scale={3}>
      <mesh ref={earthRef}>
        <sphereGeometry args={[10, 64, 64]} />
        <meshStandardMaterial
          color={0x1a4f7a}
          metalness={0.1}
          roughness={0.8}
        />
      </mesh>
      <mesh>
        <sphereGeometry args={[10.3, 64, 64]} />
        <meshBasicMaterial
          color={0x4a9fff}
          transparent
          opacity={0.2}
          side={THREE.BackSide}
          blending={THREE.AdditiveBlending}
        />
      </mesh>
      <mesh>
        <sphereGeometry args={[10.15, 64, 64]} />
        <meshStandardMaterial color={0xffffff} transparent opacity={0.3} />
      </mesh>
    </group>
  );
}

// ============================================
// CINEMATIC CAMERA
// ============================================
function CinematicCamera({
  stage,
  progress,
}: {
  stage: "launch" | "ascend" | "space";
  progress: number;
}) {
  const { camera } = useThree();

  useFrame(({ clock }) => {
    const t = clock.getElapsedTime();

    if (stage === "launch") {
      camera.position.x = Math.sin(t * 0.2) * 2;
      camera.position.y = 5 + Math.sin(t * 0.3) * 0.5;
      camera.position.z = 15 + Math.sin(t * 0.1) * 2;
      camera.lookAt(0, 5, 0);
    }

    if (stage === "ascend") {
      const dist = 15 + progress * 30;
      camera.position.x = 20 - progress * 15;
      camera.position.y = 10 + progress * 40;
      camera.position.z = dist;
      camera.lookAt(0, progress * 50, 0);
    }

    if (stage === "space") {
      camera.position.x = 30;
      camera.position.y = 70;
      camera.position.z = 40;
      camera.lookAt(0, 70, 0);
    }
  });

  return null;
}

// ============================================
// MAIN SCENE
// ============================================
function Scene() {
  const [stage, setStage] = useState<"launch" | "ascend" | "space">("launch");
  const [progress, setProgress] = useState(0);
  const [engineVisible, setEngineVisible] = useState(true);
  const [launchPadVisible, setLaunchPadVisible] = useState(true);
  const [earthVisible, setEarthVisible] = useState(false);

  useEffect(() => {
    const launchTimer = setTimeout(() => {
      setStage("ascend");
      let p = 0;
      const progressInterval = setInterval(() => {
        p += 0.01;
        setProgress(p);
        if (p > 0.5) setLaunchPadVisible(false);
        if (p > 0.7) setEngineVisible(false);
        if (p > 0.85) {
          setStage("space");
          setEarthVisible(true);
        }
        if (p >= 1) clearInterval(progressInterval);
      }, 50);
    }, 3000);

    return () => clearTimeout(launchTimer);
  }, []);

  return (
    <>
      <color
        attach="background"
        args={[stage === "space" ? BRAND.space : 0x0a0a0a]}
      />
      <fog attach="fog" args={["#111122", 50, stage === "space" ? 200 : 100]} />
      <ambientLight intensity={0.3} />
      <directionalLight
        position={[20, 30, 10]}
        intensity={1.2}
        color="#fff5e6"
      />
      <pointLight
        position={[0, -10, 0]}
        intensity={2}
        color={BRAND.orange}
        distance={50}
      />

      <Stars
        radius={200}
        depth={100}
        count={stage === "space" ? 6000 : 3000}
        factor={stage === "space" ? 8 : 5}
        saturation={0.5}
        fade={stage === "space"}
        speed={0.5}
      />

      <CinematicCamera stage={stage} progress={progress} />
      <ArtemisRocket stage={stage} progress={progress} />
      <EnginePlume
        intensity={stage === "space" ? 0 : 1}
        visible={engineVisible}
      />
      <LaunchPad visible={launchPadVisible} />
      <Earth visible={earthVisible} />
    </>
  );
}

// ============================================
// LOADING SCREEN
// ============================================
function Loader() {
  return (
    <Html center>
      <div
        style={{
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          gap: "20px",
          color: "#fff",
          fontFamily: "JetBrains Mono, monospace",
        }}
      >
        <div
          style={{
            width: "60px",
            height: "60px",
            border: "3px solid rgba(255, 107, 53, 0.3)",
            borderTop: `3px solid ${BRAND.flame}`,
            borderRadius: "50%",
            animation: "spin 1s linear infinite",
          }}
        />
        <p
          style={{
            fontSize: "14px",
            letterSpacing: "0.2em",
            color: "rgba(255,255,255,0.7)",
          }}
        >
          INITIALIZING LAUNCH SEQUENCE
        </p>
        <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
      </div>
    </Html>
  );
}

// ============================================
// MAIN EXPORT
// ============================================
export default function ArtemisLaunch() {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) {
    return (
      <div
        style={{
          width: "100%",
          height: "600px",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          background: "linear-gradient(180deg, #0a0a1a 0%, #111122 100%)",
        }}
      >
        <div
          style={{
            width: "48px",
            height: "48px",
            border: "3px solid rgba(255, 107, 53, 0.3)",
            borderTop: `3px solid ${BRAND.flame}`,
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
          alpha: false,
          powerPreference: "high-performance",
        }}
        dpr={[1, 2]}
      >
        <Suspense fallback={<Loader />}>
          <Scene />
        </Suspense>
      </Canvas>

      {/* Mission Status */}
      <div
        style={{
          position: "absolute",
          bottom: "24px",
          left: "24px",
          pointerEvents: "none",
          fontFamily: "JetBrains Mono, monospace",
        }}
      >
        <div
          style={{
            fontSize: "11px",
            letterSpacing: "0.15em",
            color: "rgba(255,255,255,0.5)",
            marginBottom: "4px",
          }}
        >
          MISSION STATUS
        </div>
        <div
          style={{
            fontSize: "24px",
            fontWeight: 700,
            color: BRAND.flame,
            textShadow: `0 0 20px ${BRAND.flame}`,
          }}
        >
          ARTEMIS I
        </div>
        <div
          style={{
            fontSize: "12px",
            color: "rgba(255,255,255,0.6)",
            marginTop: "8px",
          }}
        >
          LUNAR MISSION
        </div>
      </div>

      {/* Telemetry */}
      <div
        style={{
          position: "absolute",
          top: "24px",
          right: "24px",
          pointerEvents: "none",
          fontFamily: "JetBrains Mono, monospace",
          fontSize: "11px",
          textAlign: "right",
        }}
      >
        <div style={{ color: BRAND.cyan, marginBottom: "4px" }}>
          ALT: <span id="altitude">0</span> KM
        </div>
        <div style={{ color: BRAND.flame, marginBottom: "4px" }}>
          VEL: MACH 0
        </div>
        <div style={{ color: "rgba(255,255,255,0.5)" }}>STAGE: IGNITION</div>
      </div>

      <script
        dangerouslySetInnerHTML={{
          __html: `
        let alt = 0;
        setInterval(() => {
          alt += Math.random() * 5;
          const el = document.getElementById('altitude');
          if (el) el.textContent = Math.floor(alt);
        }, 100);
      `,
        }}
      />
    </div>
  );
}
