/**
 * VelocityJetScene - Premium 3D Hero Experience
 * 
 * A cinematic private jet soaring through the FunctionFly universe
 * using @react-three/fiber and @react-three/drei.
 */

import { useRef, useMemo, Suspense, useState, useEffect } from 'react';
import { Canvas, useFrame } from '@react-three/fiber';
import { 
  Clouds,
  Cloud,
  Stars,
  Float,
  Html,
  Trail,
  Environment,
} from '@react-three/drei';
import * as THREE from 'three';

// Brand Palette
const BRAND = {
  flame: 0xFF6B35,
  cyan: 0x00D4FF,
  stratosphere: 0x5B7CF5,
  tarmac: 0x0D1117,
  cockpit: 0x161B22,
  white: 0xffffff,
};

// Jet Model - Sleek Private Aircraft
function VelocityJet() {
  const jetRef = useRef<THREE.Group>(null);
  const engineGlowRef = useRef<THREE.PointLight>(null);
  
  const materials = useMemo(() => ({
    fuselage: new THREE.MeshPhysicalMaterial({
      color: BRAND.tarmac,
      metalness: 0.9,
      roughness: 0.2,
      clearcoat: 1.0,
      clearcoatRoughness: 0.1,
      envMapIntensity: 1.5,
    }),
    wing: new THREE.MeshPhysicalMaterial({
      color: BRAND.cockpit,
      metalness: 0.85,
      roughness: 0.3,
      clearcoat: 0.8,
    }),
    accent: new THREE.MeshPhysicalMaterial({
      color: BRAND.flame,
      emissive: BRAND.flame,
      emissiveIntensity: 0.4,
      metalness: 0.6,
      roughness: 0.4,
    }),
    cockpit: new THREE.MeshPhysicalMaterial({
      color: BRAND.cyan,
      metalness: 0.1,
      roughness: 0.1,
      transmission: 0.9,
      thickness: 0.5,
      transparent: true,
      opacity: 0.8,
    }),
    engine: new THREE.MeshStandardMaterial({
      color: 0x1a1a2e,
      metalness: 0.9,
      roughness: 0.3,
    }),
    engineGlow: new THREE.MeshBasicMaterial({
      color: BRAND.flame,
      transparent: true,
      opacity: 0.7,
    }),
    lightRed: new THREE.MeshBasicMaterial({ color: 0xff0000 }),
    lightGreen: new THREE.MeshBasicMaterial({ color: 0x00ff00 }),
  }), []);

  useFrame(({ clock }) => {
    if (!jetRef.current) return;
    
    const time = clock.getElapsedTime();
    
    // Smooth banking and floating animation
    jetRef.current.rotation.z = Math.sin(time * 0.4) * 0.08;
    jetRef.current.rotation.x = Math.cos(time * 0.3) * 0.03;
    jetRef.current.rotation.y = Math.sin(time * 0.2) * 0.02;
    jetRef.current.position.y = Math.sin(time * 0.5) * 0.5;
    
    // Pulsing engine glow
    if (engineGlowRef.current) {
      engineGlowRef.current.intensity = 2 + Math.sin(time * 8) * 0.5;
    }
  });

  return (
    <group ref={jetRef} scale={0.8}>
      {/* Main Fuselage - Capsule shape */}
      <mesh material={materials.fuselage} castShadow>
        <capsuleGeometry args={[1, 7, 4, 16]} rotation={[0, 0, Math.PI / 2]} />
      </mesh>
      
      {/* Nose cone - smoother transition */}
      <mesh position={[4, 0, 0]} material={materials.fuselage}>
        <coneGeometry args={[1, 2, 32]} rotation={[0, 0, Math.PI / 2]} />
      </mesh>
      
      {/* Cockpit Windows */}
      <mesh position={[2, 0.7, 0]} material={materials.cockpit}>
        <sphereGeometry args={[0.8, 32, 32, 0, Math.PI * 2, 0, Math.PI * 0.3]} />
        <mesh rotation={[0, 0, -Math.PI / 4]} />
      </mesh>
      
      {/* Left Wing - Swept */}
      <group position={[0, -0.3, 2.2]}>
        <mesh material={materials.wing} castShadow>
          <boxGeometry args={[4, 0.15, 2.5]} />
        </mesh>
        {/* Winglet */}
        <mesh position={[1.5, 0.3, 1]} material={materials.accent}>
          <boxGeometry args={[0.3, 0.8, 0.1]} rotation={[0, 0, -Math.PI / 8]} />
        </mesh>
        {/* Engine nacelle */}
        <mesh position={[-0.5, -0.8, 0]} material={materials.engine}>
          <cylinderGeometry args={[0.4, 0.45, 1.8, 24]} rotation={[0, 0, Math.PI / 2]} />
        </mesh>
        {/* Engine intake */}
        <mesh position={[0.4, -0.8, 0]} material={materials.engineGlow}>
          <circleGeometry args={[0.35, 24]} rotation={[0, Math.PI / 2, 0]} />
        </mesh>
        {/* Engine exhaust trail */}
        <mesh position={[-1.5, -0.8, 0]} material={materials.engineGlow}>
          <cylinderGeometry args={[0.2, 0.35, 2.5, 16]} rotation={[0, 0, Math.PI / 2]} />
        </mesh>
      </group>
      
      {/* Right Wing */}
      <group position={[0, -0.3, -2.2]}>
        <mesh material={materials.wing} castShadow>
          <boxGeometry args={[4, 0.15, 2.5]} />
        </mesh>
        {/* Winglet */}
        <mesh position={[1.5, 0.3, -1]} material={materials.accent}>
          <boxGeometry args={[0.3, 0.8, 0.1]} rotation={[0, 0, -Math.PI / 8]} />
        </mesh>
        {/* Engine */}
        <mesh position={[-0.5, -0.8, 0]} material={materials.engine}>
          <cylinderGeometry args={[0.4, 0.45, 1.8, 24]} rotation={[0, 0, Math.PI / 2]} />
        </mesh>
        {/* Engine intake glow */}
        <mesh position={[0.4, -0.8, 0]} material={materials.engineGlow}>
          <circleGeometry args={[0.35, 24]} rotation={[0, Math.PI / 2, 0]} />
        </mesh>
        {/* Engine exhaust */}
        <mesh position={[-1.5, -0.8, 0]} material={materials.engineGlow}>
          <cylinderGeometry args={[0.2, 0.35, 2.5, 16]} rotation={[0, 0, Math.PI / 2]} />
        </mesh>
      </group>
      
      {/* Vertical Stabilizer (Tail) */}
      <group position={[-3, 1.2, 0]}>
        <mesh material={materials.wing} castShadow>
          <boxGeometry args={[1.5, 2, 0.15]} />
        </mesh>
        {/* Brand logo mark */}
        <mesh position={[-0.2, 0.8, 0.08]} material={materials.accent}>
          <circleGeometry args={[0.5, 32]} />
        </mesh>
      </group>
      
      {/* Horizontal stabilizers */}
      <mesh position={[-2.5, 0.6, 0.8]} material={materials.wing}>
        <boxGeometry args={[1, 0.1, 1.2]} />
      </mesh>
      <mesh position={[-2.5, 0.6, -0.8]} material={materials.wing}>
        <boxGeometry args={[1, 0.1, 1.2]} />
      </mesh>
      
      {/* Navigation Lights */}
      {/* Left wing tip - Red */}
      <mesh position={[2, -0.3, 2.2]} material={materials.lightRed}>
        <sphereGeometry args={[0.1, 16, 16]} />
      </mesh>
      <pointLight 
        ref={engineGlowRef}
        position={[2, -0.3, 2.2]} 
        color={0xff0000} 
        intensity={2} 
        distance={10}
        decay={2}
      />
      
      {/* Right wing tip - Green */}
      <mesh position={[2, -0.3, -2.2]} material={materials.lightGreen}>
        <sphereGeometry args={[0.1, 16, 16]} />
      </mesh>
      <pointLight position={[2, -0.3, -2.2]} color={0x00ff00} intensity={2} distance={10} decay={2} />
      
      {/* Beacon light on tail */}
      <mesh position={[-3.8, 2.2, 0]} material={materials.accent}>
        <sphereGeometry args={[0.12, 16, 16]} />
      </mesh>
      <pointLight position={[-3.8, 2.2, 0]} color={BRAND.flame} intensity={3} distance={15} decay={1.5} />
      
      {/* Accent stripes on fuselage */}
      <mesh position={[1, -0.1, 0]} material={materials.accent}>
        <boxGeometry args={[3, 0.05, 0.8]} />
      </mesh>
    </group>
  );
}

// Floating Clouds
function DriftingClouds() {
  const cloudsRef = useRef<THREE.Group>(null);
  
  const cloudPositions = useMemo(() => [
    { pos: [-15, -5, -25], scale: 2, color: BRAND.white, opacity: 0.3 },
    { pos: [10, 8, -20], scale: 1.8, color: BRAND.cyan, opacity: 0.25 },
    { pos: [-8, 5, -35], scale: 2.5, color: BRAND.white, opacity: 0.35 },
    { pos: [12, -3, -30], scale: 1.5, color: BRAND.cyan, opacity: 0.2 },
    { pos: [-20, 2, -40], scale: 2.2, color: BRAND.white, opacity: 0.3 },
    { pos: [5, 10, -15], scale: 1.6, color: BRAND.cyan, opacity: 0.25 },
    { pos: [-5, -8, -18], scale: 2, color: BRAND.white, opacity: 0.28 },
    { pos: [18, 0, -45], scale: 1.9, color: BRAND.cyan, opacity: 0.22 },
  ], []);

  useFrame(({ clock }) => {
    if (!cloudsRef.current) return;
    const time = clock.getElapsedTime();
    
    cloudsRef.current.children.forEach((cloud, i) => {
      const speed = 0.5 + (i % 3) * 0.3;
      cloud.position.z += 0.008 * speed;
      cloud.rotation.y += 0.001 * (i % 2 === 0 ? 1 : -1);
      
      // Wrap around
      if (cloud.position.z > 20) {
        cloud.position.z = -50;
        cloud.position.x = (Math.random() - 0.5) * 40;
      }
    });
  });

  return (
    <group ref={cloudsRef}>
      <Clouds>
        {cloudPositions.map((cloud, i) => (
          <Cloud
            key={i}
            position={cloud.pos}
            scale={cloud.scale}
            color={new THREE.Color(cloud.color)}
            opacity={cloud.opacity}
            speed={0}
            bounds={[6, 3, 4]}
          />
        ))}
      </Clouds>
    </group>
  );
}

// Speed Particles
function VelocityStreaks({ count = 60 }: { count?: number }) {
  const pointsRef = useRef<THREE.Points>(null);
  
  const [positions, colors] = useMemo(() => {
    const pos = new Float32Array(count * 3);
    const cols = new Float32Array(count * 3);
    
    for (let i = 0; i < count; i++) {
      const angle = Math.random() * Math.PI * 2;
      const radius = 6 + Math.random() * 10;
      
      pos[i * 3] = (Math.random() - 0.5) * 60;
      pos[i * 3 + 1] = Math.sin(angle) * radius * 0.5 + (Math.random() - 0.5) * 15;
      pos[i * 3 + 2] = (Math.random() - 0.5) * 80 - 10;
      
      // Mix brand colors
      if (Math.random() > 0.5) {
        cols[i * 3] = ((BRAND.flame >> 16) & 255) / 255;
        cols[i * 3 + 1] = ((BRAND.flame >> 8) & 255) / 255;
        cols[i * 3 + 2] = (BRAND.flame & 255) / 255;
      } else {
        cols[i * 3] = ((BRAND.cyan >> 16) & 255) / 255;
        cols[i * 3 + 1] = ((BRAND.cyan >> 8) & 255) / 255;
        cols[i * 3 + 2] = (BRAND.cyan & 255) / 255;
      }
    }
    return [pos, cols];
  }, [count]);

  useFrame((state, delta) => {
    if (!pointsRef.current) return;
    const pos = pointsRef.current.geometry.attributes.position.array as Float32Array;
    
    for (let i = 0; i < count; i++) {
      // Move particles past camera to simulate forward motion
      pos[i * 3 + 2] += delta * (15 + Math.random() * 25);
      
      // Reset when they pass
      if (pos[i * 3 + 2] > 40) {
        pos[i * 3 + 2] = -60;
        pos[i * 3] = (Math.random() - 0.5) * 50;
        pos[i * 3 + 1] = (Math.random() - 0.5) * 20;
      }
    }
    
    pointsRef.current.geometry.attributes.position.needsUpdate = true;
  });

  return (
    <points ref={pointsRef}>
      <bufferGeometry>
        <bufferAttribute
          attach="attributes-position"
          count={count}
          array={positions}
          itemSize={3}
        />
        <bufferAttribute
          attach="attributes-color"
          count={count}
          array={colors}
          itemSize={3}
        />
      </bufferGeometry>
      <pointsMaterial
        size={0.2}
        vertexColors
        transparent
        opacity={0.8}
        blending={THREE.AdditiveBlending}
        sizeAttenuation
      />
    </points>
  );
}

// Scene lighting
function SceneLighting() {
  return (
    <>
      <ambientLight intensity={0.4} />
      <directionalLight 
        position={[15, 25, 10]} 
        intensity={1.2}
        color="#fff8e7"
        castShadow
        shadow-mapSize={[1024, 1024]}
      />
      <pointLight position={[0, 10, 0]} intensity={0.8} color={BRAND.flame} distance={30} decay={1} />
      <pointLight position={[-10, 5, 10]} intensity={0.6} color={BRAND.cyan} distance={25} decay={1} />
      <spotLight
        position={[5, 15, 5]}
        angle={0.4}
        penumbra={0.5}
        intensity={0.5}
        castShadow
      />
    </>
  );
}

// Camera controller
function OrbitalCamera() {
  useFrame(({ clock, camera }) => {
    const time = clock.getElapsedTime();
    const radius = 18;
    const height = 8 + Math.sin(time * 0.3) * 2;
    const angle = time * 0.15;
    
    camera.position.x = Math.cos(angle) * radius;
    camera.position.z = Math.sin(angle) * radius;
    camera.position.y = height;
    camera.lookAt(0, 0, 0);
  });
  return null;
}

// Main Scene
function Scene() {
  return (
    <>
      {/* Background gradient */}
      <color attach="background" args={[new THREE.Color(BRAND.tarmac)]} />
      
      {/* Fog for depth */}
      <fog attach="fog" args={[BRAND.tarmac, 20, 100]} />
      
      {/* Lighting */}
      <SceneLighting />
      
      {/* Environment map for reflections */}
      <Environment preset="sunset" />
      
      {/* Camera */}
      <perspectiveCamera makeDefault position={[15, 10, 15]} fov={55} near={0.1} far={200} />
      <OrbitalCamera />
      
      {/* The Aircraft */}
      <Float 
        speed={0.8}
        rotationIntensity={0.15}
        floatIntensity={0.3}
      >
        <VelocityJet />
      </Float>
      
      {/* Atmosphere */}
      <DriftingClouds />
      <VelocityStreaks count={50} />
      
      {/* Stars background */}
      <Stars 
        radius={150}
        depth={80}
        count={3000}
        factor={5}
        saturation={0.4}
        fade
        speed={0.4}
      />
      
      {/* Brand text floating - using HTML overlay instead of drei Text to avoid font loading issues */}
      <Html position={[-8, 6, -10]} center transform>
        <div style={{
          fontSize: '48px',
          fontWeight: 700,
          color: '#FF6B35',
          textShadow: '0 0 20px rgba(255, 107, 53, 0.5)',
          fontFamily: 'Inter, system-ui, sans-serif',
          letterSpacing: '-0.02em',
          pointerEvents: 'none',
          userSelect: 'none',
        }}>
          FUNCTION
        </div>
      </Html>
      
      <Html position={[8, -6, -10]} center transform>
        <div style={{
          fontSize: '48px',
          fontWeight: 700,
          color: '#00D4FF',
          textShadow: '0 0 20px rgba(0, 212, 255, 0.5)',
          fontFamily: 'Inter, system-ui, sans-serif',
          letterSpacing: '-0.02em',
          pointerEvents: 'none',
          userSelect: 'none',
        }}>
          FLY
        </div>
      </Html>
    </>
  );
}

// Loading spinner
function Loader() {
  return (
    <Html center>
      <div style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: '16px',
        color: '#fff'
      }}>
        <div style={{
          width: '48px',
          height: '48px',
          border: '3px solid rgba(255, 107, 53, 0.3)',
          borderTop: '3px solid #FF6B35',
          borderRadius: '50%',
          animation: 'spin 1s linear infinite'
        }} />
        <p style={{ fontSize: '14px', letterSpacing: '0.1em' }}>INITIALIZING FLIGHT...</p>
        <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
      </div>
    </Html>
  );
}

// Main export
export default function VelocityJetScene() {
  const [mounted, setMounted] = useState(false);
  
  useEffect(() => {
    setMounted(true);
  }, []);
  
  if (!mounted) {
    return (
      <div style={{
        width: '100%',
        height: '600px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(180deg, #0D1117 0%, #161B22 100%)'
      }}>
        <div style={{
          width: '48px',
          height: '48px',
          border: '3px solid rgba(255, 107, 53, 0.3)',
          borderTop: '3px solid #FF6B35',
          borderRadius: '50%',
          animation: 'spin 1s linear infinite'
        }} />
        <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
      </div>
    );
  }

  return (
    <div style={{ width: '100%', height: '600px', position: 'relative' }}>
      <Canvas
        gl={{
          antialias: true,
          alpha: true,
          powerPreference: 'high-performance',
        }}
        dpr={[1, 1.5]}
        shadows
      >
        <Suspense fallback={<Loader />}>
          <Scene />
        </Suspense>
      </Canvas>
      
      {/* UI Overlay - Brand */}
      <div style={{
        position: 'absolute',
        bottom: '24px',
        left: '24px',
        pointerEvents: 'none',
        fontFamily: 'Inter, system-ui, sans-serif',
      }}>
        <h2 style={{
          fontSize: '28px',
          fontWeight: 700,
          letterSpacing: '-0.02em',
          margin: 0,
          color: '#fff'
        }}>
          <span style={{ color: '#FF6B35' }}>Function</span>
          <span style={{ color: '#00D4FF' }}>Fly</span>
        </h2>
        <p style={{
          fontSize: '12px',
          color: 'rgba(255,255,255,0.6)',
          letterSpacing: '0.15em',
          textTransform: 'uppercase',
          margin: '4px 0 0 0'
        }}>Velocity Edition</p>
      </div>
      
      {/* Flight Data HUD */}
      <div style={{
        position: 'absolute',
        top: '24px',
        right: '24px',
        pointerEvents: 'none',
        textAlign: 'right',
        fontFamily: 'JetBrains Mono, monospace',
        fontSize: '12px'
      }}>
        <div style={{ color: '#FF6B35', marginBottom: '4px' }}>ALT: 35,000 FT</div>
        <div style={{ color: '#00D4FF', marginBottom: '4px' }}>SPD: M 0.85</div>
        <div style={{ color: 'rgba(255,255,255,0.5)' }}>HDG: 247°</div>
      </div>
    </div>
  );
}
