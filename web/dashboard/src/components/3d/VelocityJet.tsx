/**
 * FunctionFly Velocity Jet - Cinematic 3D Experience
 * 
 * A breathtaking scene featuring a sleek private jet soaring through
 * volumetric clouds with the FunctionFly brand colors illuminating
 * the atmosphere. Particle trails, dynamic lighting, and atmospheric
 * effects create a premium aviation experience.
 */

import { useRef, useMemo, useEffect, Suspense } from 'react';
import { Canvas, useFrame, useThree } from '@react-three/fiber';
import { 
  Cloud,
  Clouds,
  Sky,
  Stars,
  Float,
  Trail,
  MeshDistortMaterial,
  useTexture,
  Html,
  PerspectiveCamera,
  CameraShake,
  Bvh,
  Lightformer,
  Environment,
  ContactShadows,
} from '@react-three/drei';
import * as THREE from 'three';
import { EffectComposer, Bloom, Noise, Vignette, DepthOfField, ChromaticAberration } from '@react-three/postprocessing';

// Brand Colors - Velocity Palette
const BRAND = {
  flame: new THREE.Color('#FF6B35'),
  flameDark: new THREE.Color('#E85A2A'),
  cyan: new THREE.Color('#00D4FF'),
  stratosphere: new THREE.Color('#5B7CF5'),
  taxiway: new THREE.Color('#00FF9D'),
  beacon: new THREE.Color('#FFB800'),
  tarmac: new THREE.Color('#0D1117'),
  cockpit: new THREE.Color('#161B22'),
};

// Custom shader material for the jet exhaust trail
const exhaustShader = {
  uniforms: {
    time: { value: 0 },
    colorCore: { value: BRAND.flame },
    colorOuter: { value: BRAND.cyan },
    intensity: { value: 2.0 },
  },
  vertexShader: `
    uniform float time;
    varying vec2 vUv;
    varying vec3 vPosition;
    varying float vAlpha;
    
    void main() {
      vUv = uv;
      vPosition = position;
      
      // Turbulence effect
      float turbulence = sin(position.x * 10.0 + time * 3.0) * 0.1;
      turbulence += cos(position.y * 8.0 + time * 2.0) * 0.05;
      turbulence += sin(position.z * 6.0 + time * 4.0) * 0.08;
      
      vec3 newPos = position + normal * turbulence * (1.0 - uv.x);
      
      // Fade alpha along the length
      vAlpha = 1.0 - smoothstep(0.0, 1.0, uv.x);
      
      gl_Position = projectionMatrix * modelViewMatrix * vec4(newPos, 1.0);
    }
  `,
  fragmentShader: `
    uniform float time;
    uniform vec3 colorCore;
    uniform vec3 colorOuter;
    uniform float intensity;
    varying vec2 vUv;
    varying float vAlpha;
    
    void main() {
      // Radial gradient
      float radial = 1.0 - length(vUv.y - 0.5) * 2.0;
      radial = max(0.0, radial);
      
      // Flicker effect
      float flicker = sin(time * 20.0) * 0.1 + 0.9;
      float noise = fract(sin(vUv.x * 100.0 + time) * 1000.0) * 0.2;
      
      // Core hot center, cool outer
      vec3 color = mix(colorCore, colorOuter, radial * radial);
      
      // Add flame intensity
      float fireIntensity = pow(radial, 2.0) * intensity * flicker * (1.0 + noise);
      
      // Alpha with fade
      float alpha = vAlpha * radial * 0.8;
      
      gl_FragColor = vec4(color * fireIntensity, alpha);
    }
  `,
  transparent: true,
  depthWrite: false,
  blending: THREE.AdditiveBlending,
  side: THREE.DoubleSide,
};

// Jet Aircraft Component - Sleek Private Jet Design
function VelocityJet() {
  const jetRef = useRef<THREE.Group>(null);
  const fuselageRef = useRef<THREE.Mesh>(null);
  const leftWingRef = useRef<THREE.Mesh>(null);
  const rightWingRef = useRef<THREE.Mesh>(null);
  const tailRef = useRef<THREE.Mesh>(null);
  const engineLeftRef = useRef<THREE.Mesh>(null);
  const engineRightRef = useRef<THREE.Mesh>(null);
  const exhaustLeftRef = useRef<THREE.Mesh>(null);
  const exhaustRightRef = useRef<THREE.Mesh>(null);
  const lightsRef = useRef<THREE.Group>(null);

  // Jet material - sleek metallic with brand accents
  const fuselageMaterial = useMemo(() => {
    return new THREE.MeshPhysicalMaterial({
      color: BRAND.tarmac,
      metalness: 0.9,
      roughness: 0.1,
      clearcoat: 1.0,
      clearcoatRoughness: 0.05,
      sheen: 0.5,
      sheenColor: BRAND.cyan,
      envMapIntensity: 1.5,
    });
  }, []);

  const wingMaterial = useMemo(() => {
    return new THREE.MeshPhysicalMaterial({
      color: BRAND.cockpit,
      metalness: 0.85,
      roughness: 0.15,
      clearcoat: 0.8,
      envMapIntensity: 1.2,
    });
  }, []);

  const accentMaterial = useMemo(() => {
    return new THREE.MeshPhysicalMaterial({
      color: BRAND.flame,
      emissive: BRAND.flame,
      emissiveIntensity: 0.2,
      metalness: 0.7,
      roughness: 0.3,
    });
  }, []);

  const engineMaterial = useMemo(() => {
    return new THREE.MeshPhysicalMaterial({
      color: new THREE.Color('#1a1a2e'),
      metalness: 0.95,
      roughness: 0.2,
      clearcoat: 1.0,
    });
  }, []);

  const exhaustMaterial = useMemo(() => {
    return new THREE.ShaderMaterial(exhaustShader);
  }, []);

  const glassMaterial = useMemo(() => {
    return new THREE.MeshPhysicalMaterial({
      color: BRAND.cyan,
      metalness: 0,
      roughness: 0,
      transmission: 0.95,
      thickness: 0.5,
      envMapIntensity: 1.5,
      clearcoat: 1.0,
    });
  }, []);

  const lightGlowMaterial = useMemo(() => {
    return new THREE.MeshBasicMaterial({
      color: BRAND.beacon,
      transparent: true,
      opacity: 0.9,
    });
  }, []);

  useFrame(({ clock }) => {
    if (!jetRef.current) return;
    
    const time = clock.getElapsedTime();
    
    // Smooth banking motion
    const bankAngle = Math.sin(time * 0.3) * 0.15;
    const pitchAngle = Math.cos(time * 0.25) * 0.05;
    const yawAngle = Math.sin(time * 0.2) * 0.03;
    
    jetRef.current.rotation.z = bankAngle;
    jetRef.current.rotation.x = pitchAngle;
    jetRef.current.rotation.y = yawAngle;
    
    // Subtle bobbing
    jetRef.current.position.y = Math.sin(time * 0.4) * 0.3;
    
    // Update exhaust shader
    if (exhaustLeftRef.current && exhaustRightRef.current) {
      const material = exhaustLeftRef.current.material as THREE.ShaderMaterial;
      material.uniforms.time.value = time;
      (exhaustRightRef.current.material as THREE.ShaderMaterial).uniforms.time.value = time;
    }
    
    // Pulse the running lights
    if (lightsRef.current) {
      lightsRef.current.children.forEach((light, i) => {
        const mesh = light as THREE.Mesh;
        const material = mesh.material as THREE.MeshBasicMaterial;
        material.opacity = 0.7 + Math.sin(time * 3 + i * Math.PI) * 0.3;
      });
    }
  });

  return (
    <group ref={jetRef} scale={0.5}>
      {/* Main Fuselage - Sleek aerodynamic shape */}
      <mesh ref={fuselageRef} material={fuselageMaterial} castShadow receiveShadow>
        <capsuleGeometry args={[1.2, 8, 4, 16]} />
      </mesh>
      
      {/* Fuselage taper - nose cone */}
      <mesh position={[4.2, 0, 0]} material={fuselageMaterial}>
        <coneGeometry args={[1.2, 2, 32]} rotation={[0, 0, -Math.PI / 2]} />
      </mesh>
      
      {/* Cockpit windshield */}
      <mesh position={[3, 0.6, 0]} material={glassMaterial}>
        <sphereGeometry args={[0.8, 32, 32, 0, Math.PI * 2, 0, Math.PI * 0.3]} />
        <mesh rotation={[0, 0, -Math.PI / 3]} />
      </mesh>
      
      {/* Left Wing - Swept back */}
      <group position={[-1, -0.3, 2.5]}>
        <mesh ref={leftWingRef} material={wingMaterial} castShadow>
          <extrudeGeometry 
            args={[
              new THREE.Shape([
                new THREE.Vector2(0, 0),
                new THREE.Vector2(4, 0.5),
                new THREE.Vector2(5, 0),
                new THREE.Vector2(4, -2),
                new THREE.Vector2(0, -0.5),
              ]),
              { depth: 0.2, bevelEnabled: true, bevelThickness: 0.05, bevelSize: 0.05, bevelSegments: 3 }
            ]}
          />
        </mesh>
        {/* Winglet at tip */}
        <mesh position={[4.8, -1, 0]} material={accentMaterial}>
          <boxGeometry args={[0.3, 1.5, 0.1]} rotation={[0, 0, -Math.PI / 6]} />
        </mesh>
        {/* Brand accent stripe */}
        <mesh position={[2, -0.5, 0.15]} material={accentMaterial}>
          <boxGeometry args={[3, 0.1, 0.02]} />
        </mesh>
      </group>
      
      {/* Right Wing - Mirror of left */}
      <group position={[-1, -0.3, -2.5]} rotation={[0, Math.PI, 0]}>
        <mesh ref={rightWingRef} material={wingMaterial} castShadow>
          <extrudeGeometry 
            args={[
              new THREE.Shape([
                new THREE.Vector2(0, 0),
                new THREE.Vector2(4, 0.5),
                new THREE.Vector2(5, 0),
                new THREE.Vector2(4, -2),
                new THREE.Vector2(0, -0.5),
              ]),
              { depth: 0.2, bevelEnabled: true, bevelThickness: 0.05, bevelSize: 0.05, bevelSegments: 3 }
            ]}
          />
        </mesh>
        {/* Winglet */}
        <mesh position={[4.8, -1, 0]} material={accentMaterial}>
          <boxGeometry args={[0.3, 1.5, 0.1]} rotation={[0, 0, -Math.PI / 6]} />
        </mesh>
        {/* Brand stripe */}
        <mesh position={[2, -0.5, 0.15]} material={accentMaterial}>
          <boxGeometry args={[3, 0.1, 0.02]} />
        </mesh>
      </group>
      
      {/* Vertical Stabilizer (Tail) */}
      <group position={[-3.5, 1.5, 0]}>
        <mesh ref={tailRef} material={wingMaterial} castShadow>
          <extrudeGeometry 
            args={[
              new THREE.Shape([
                new THREE.Vector2(0, 0),
                new THREE.Vector2(-2, 2.5),
                new THREE.Vector2(-3, 2.5),
                new THREE.Vector2(-2.5, 0),
              ]),
              { depth: 0.15, bevelEnabled: true, bevelThickness: 0.03, bevelSize: 0.03, bevelSegments: 2 }
            ]}
          />
        </mesh>
        {/* Brand logo area */}
        <mesh position={[-1.5, 1.2, 0.08]} material={accentMaterial}>
          <circleGeometry args={[0.6, 32]} />
        </mesh>
      </group>
      
      {/* Horizontal Stabilizers */}
      <mesh position={[-3.2, 0.5, 1.2]} material={wingMaterial} castShadow>
        <boxGeometry args={[1.5, 0.1, 1.5]} rotation={[0, 0, -0.1]} />
      </mesh>
      <mesh position={[-3.2, 0.5, -1.2]} material={wingMaterial} castShadow>
        <boxGeometry args={[1.5, 0.1, 1.5]} rotation={[0, 0, -0.1]} />
      </mesh>
      
      {/* Left Engine */}
      <group position={[-1, -0.8, 3]}>
        <mesh ref={engineLeftRef} material={engineMaterial}>
          <cylinderGeometry args={[0.5, 0.6, 2.5, 32]} rotation={[0, 0, Math.PI / 2]} />
        </mesh>
        {/* Engine intake lip */}
        <mesh position={[1.25, 0, 0]} material={accentMaterial}>
          <torusGeometry args={[0.55, 0.08, 16, 32]} rotation={[0, Math.PI / 2, 0]} />
        </mesh>
        {/* Exhaust nozzle */}
        <mesh position={[-1.25, 0, 0]} material={accentMaterial}>
          <cylinderGeometry args={[0.35, 0.5, 0.5, 32]} rotation={[0, 0, Math.PI / 2]} />
        </mesh>
        {/* Exhaust trail plane */}
        <mesh 
          ref={exhaustLeftRef}
          position={[-2.5, 0, 0]} 
          material={exhaustMaterial}
          rotation={[0, 0, -Math.PI / 2]}
        >
          <planeGeometry args={[8, 1.5, 32, 16]} />
        </mesh>
      </group>
      
      {/* Right Engine */}
      <group position={[-1, -0.8, -3]}>
        <mesh ref={engineRightRef} material={engineMaterial}>
          <cylinderGeometry args={[0.5, 0.6, 2.5, 32]} rotation={[0, 0, Math.PI / 2]} />
        </mesh>
        {/* Engine intake lip */}
        <mesh position={[1.25, 0, 0]} material={accentMaterial}>
          <torusGeometry args={[0.55, 0.08, 16, 32]} rotation={[0, Math.PI / 2, 0]} />
        </mesh>
        {/* Exhaust nozzle */}
        <mesh position={[-1.25, 0, 0]} material={accentMaterial}>
          <cylinderGeometry args={[0.35, 0.5, 0.5, 32]} rotation={[0, 0, Math.PI / 2]} />
        </mesh>
        {/* Exhaust trail plane */}
        <mesh 
          ref={exhaustRightRef}
          position={[-2.5, 0, 0]} 
          material={exhaustMaterial}
          rotation={[0, 0, -Math.PI / 2]}
        >
          <planeGeometry args={[8, 1.5, 32, 16]} />
        </mesh>
      </group>
      
      {/* Navigation Lights */}
      <group ref={lightsRef}>
        {/* Left wing tip - red */}
        <mesh position={[4.2, -1.8, 2.5]}>
          <sphereGeometry args={[0.12, 16, 16]} />
          <meshBasicMaterial color="#ff0000" transparent opacity={0.9} />
        </mesh>
        <pointLight position={[4.2, -1.8, 2.5]} color="#ff0000" intensity={2} distance={5} />
        
        {/* Right wing tip - green */}
        <mesh position={[4.2, -1.8, -2.5]}>
          <sphereGeometry args={[0.12, 16, 16]} />
          <meshBasicMaterial color="#00ff00" transparent opacity={0.9} />
        </mesh>
        <pointLight position={[4.2, -1.8, -2.5]} color="#00ff00" intensity={2} distance={5} />
        
        {/* Tail - white strobe */}
        <mesh position={[-4, 2.8, 0]} material={lightGlowMaterial}>
          <sphereGeometry args={[0.1, 16, 16]} />
        </mesh>
        <pointLight position={[-4, 2.8, 0]} color={BRAND.beacon} intensity={4} distance={8} />
        
        {/* Fuselage - beacon */}
        <mesh position={[-1, 1.5, 0]} material={lightGlowMaterial}>
          <sphereGeometry args={[0.08, 16, 16]} />
        </mesh>
        <pointLight position={[-1, 1.5, 0]} color={BRAND.beacon} intensity={3} distance={5} />
      </group>
      
      {/* Contrail particles will be added separately */}
    </group>
  );
}

// Volumetric Cloud Layer
function CloudLayer({ count = 8 }: { count?: number }) {
  const cloudsRef = useRef<THREE.Group>(null);
  
  const cloudConfigs = useMemo(() => {
    return Array.from({ length: count }, (_, i) => ({
      id: i,
      position: [
        (Math.random() - 0.5) * 60,
        (Math.random() - 0.5) * 20 - 5,
        (Math.random() - 0.5) * 60,
      ] as [number, number, number],
      scale: 2 + Math.random() * 4,
      opacity: 0.4 + Math.random() * 0.4,
      speed: 0.5 + Math.random() * 1,
      color: Math.random() > 0.7 
        ? BRAND.cyan.clone().multiplyScalar(0.3) // Occasional cyan-tinted clouds
        : new THREE.Color(0xffffff),
    }));
  }, [count]);

  useFrame(({ clock }) => {
    if (!cloudsRef.current) return;
    const time = clock.getElapsedTime();
    
    cloudsRef.current.children.forEach((cloud, i) => {
      const config = cloudConfigs[i];
      // Drift clouds
      cloud.position.x = config.position[0] + Math.sin(time * 0.1 * config.speed) * 5;
      cloud.position.z = config.position[2] + time * config.speed * 2;
      
      // Wrap around
      if (cloud.position.z > 30) {
        cloud.position.z = -30;
      }
      
      // Subtle rotation
      cloud.rotation.y = time * 0.05 * config.speed;
    });
  });

  return (
    <group ref={cloudsRef}>
      {cloudConfigs.map((config) => (
        <Cloud 
          key={config.id}
          position={config.position}
          scale={config.scale}
          color={config.color}
          opacity={config.opacity}
        />
      ))}
    </group>
  );
}

// Particle Speed Lines - Visualizing velocity
function SpeedLines({ count = 200 }: { count?: number }) {
  const linesRef = useRef<THREE.Points>(null);
  const geometryRef = useRef<THREE.BufferGeometry>(null);
  
  const { positions, colors, sizes } = useMemo(() => {
    const positions = new Float32Array(count * 3);
    const colors = new Float32Array(count * 3);
    const sizes = new Float32Array(count);
    
    for (let i = 0; i < count; i++) {
      // Position in a tunnel around the flight path
      const angle = Math.random() * Math.PI * 2;
      const radius = 5 + Math.random() * 15;
      positions[i * 3] = (Math.random() - 0.5) * 100;
      positions[i * 3 + 1] = Math.sin(angle) * radius + (Math.random() - 0.5) * 20;
      positions[i * 3 + 2] = Math.cos(angle) * radius + (Math.random() - 0.5) * 60;
      
      // Brand colors
      const colorMix = Math.random();
      if (colorMix < 0.4) {
        colors[i * 3] = BRAND.flame.r;
        colors[i * 3 + 1] = BRAND.flame.g;
        colors[i * 3 + 2] = BRAND.flame.b;
      } else if (colorMix < 0.7) {
        colors[i * 3] = BRAND.cyan.r;
        colors[i * 3 + 1] = BRAND.cyan.g;
        colors[i * 2 + 2] = BRAND.cyan.b;
      } else {
        colors[i * 3] = BRAND.stratosphere.r;
        colors[i * 3 + 1] = BRAND.stratosphere.g;
        colors[i * 3 + 2] = BRAND.stratosphere.b;
      }
      
      sizes[i] = 0.1 + Math.random() * 0.4;
    }
    
    return { positions, colors, sizes };
  }, [count]);

  useFrame((state, delta) => {
    if (!linesRef.current) return;
    
    const positionsArray = linesRef.current.geometry.attributes.position.array as Float32Array;
    
    for (let i = 0; i < count; i++) {
      // Move particles backward to simulate forward motion
      positionsArray[i * 3 + 2] += delta * (20 + Math.random() * 30);
      
      // Reset when they pass camera
      if (positionsArray[i * 3 + 2] > 50) {
        positionsArray[i * 3 + 2] = -50;
        positionsArray[i * 3] = (Math.random() - 0.5) * 100;
        positionsArray[i * 3 + 1] = (Math.random() - 0.5) * 30;
      }
    }
    
    linesRef.current.geometry.attributes.position.needsUpdate = true;
  });

  return (
    <points ref={linesRef}>
      <bufferGeometry ref={geometryRef}>
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
        <bufferAttribute
          attach="attributes-size"
          count={count}
          array={sizes}
          itemSize={1}
        />
      </bufferGeometry>
      <pointsMaterial
        size={0.3}
        vertexColors
        transparent
        opacity={0.8}
        blending={THREE.AdditiveBlending}
        sizeAttenuation
      />
    </points>
  );
}

// Atmospheric Dust Motes
function AtmosphericDust({ count = 500 }: { count?: number }) {
  const dustRef = useRef<THREE.Points>(null);
  
  const positions = useMemo(() => {
    const pos = new Float32Array(count * 3);
    for (let i = 0; i < count; i++) {
      pos[i * 3] = (Math.random() - 0.5) * 80;
      pos[i * 3 + 1] = (Math.random() - 0.5) * 40;
      pos[i * 3 + 2] = (Math.random() - 0.5) * 80;
    }
    return pos;
  }, [count]);

  useFrame(({ clock }) => {
    if (!dustRef.current) return;
    dustRef.current.rotation.y = clock.getElapsedTime() * 0.02;
    dustRef.current.rotation.x = Math.sin(clock.getElapsedTime() * 0.05) * 0.05;
  });

  return (
    <points ref={dustRef}>
      <bufferGeometry>
        <bufferAttribute
          attach="attributes-position"
          count={count}
          array={positions}
          itemSize={3}
        />
      </bufferGeometry>
      <pointsMaterial
        size={0.05}
        color={BRAND.cyan}
        transparent
        opacity={0.4}
        blending={THREE.AdditiveBlending}
        sizeAttenuation
      />
    </points>
  );
}

// Dynamic Lighting Setup
function SceneLighting() {
  const sunRef = useRef<THREE.DirectionalLight>(null);
  const ambientRef = useRef<THREE.AmbientLight>(null);
  
  useFrame(({ clock }) => {
    if (!sunRef.current) return;
    
    // Simulate sun position changing
    const time = clock.getElapsedTime();
    sunRef.current.position.x = Math.sin(time * 0.1) * 20 + 10;
    sunRef.current.position.y = Math.cos(time * 0.05) * 10 + 20;
    
    // Color temperature shift
    const temp = 0.5 + Math.sin(time * 0.2) * 0.3;
    sunRef.current.color.setHSL(0.1, 0.8, 0.5 + temp * 0.3);
  });

  return (
    <>
      {/* Main sun light */}
      <directionalLight
        ref={sunRef}
        position={[10, 20, 10]}
        intensity={2}
        color="#fff8e7"
        castShadow
        shadow-mapSize={[2048, 2048]}
        shadow-camera-far={100}
        shadow-camera-left={-30}
        shadow-camera-right={30}
        shadow-camera-top={30}
        shadow-camera-bottom={-30}
      />
      
      {/* Ambient fill */}
      <ambientLight ref={ambientRef} intensity={0.3} color={BRAND.stratosphere} />
      
      {/* Blue fill from below (sky reflection) */}
      <hemisphereLight
        groundColor={BRAND.cyan}
        color={BRAND.stratosphere}
        intensity={0.5}
      />
      
      {/* Brand accent point lights */}
      <pointLight position={[0, 10, 0]} color={BRAND.flame} intensity={1} distance={30} />
      <pointLight position={[-10, 5, -10]} color={BRAND.cyan} intensity={0.8} distance={25} />
    </>
  );
}

// Main scene composition
function VelocityScene() {
  const { camera } = useThree();
  const cameraRef = useRef<THREE.PerspectiveCamera>(null);
  
  useEffect(() => {
    if (cameraRef.current) {
      cameraRef.current.position.set(-15, 5, 15);
      cameraRef.current.lookAt(0, 0, 0);
    }
  }, []);

  useFrame(({ clock }) => {
    if (!cameraRef.current) return;
    
    const time = clock.getElapsedTime();
    
    // Cinematic camera orbit
    const radius = 20;
    const height = 8 + Math.sin(time * 0.2) * 3;
    const angle = time * 0.15;
    
    cameraRef.current.position.x = Math.cos(angle) * radius;
    cameraRef.current.position.z = Math.sin(angle) * radius;
    cameraRef.current.position.y = height;
    cameraRef.current.lookAt(2, 0, 0);
  });

  return (
    <>
      <PerspectiveCamera ref={cameraRef} makeDefault fov={50} near={0.1} far={1000} />
      
      <Bvh>
        <SceneLighting />
        
        {/* The Jet */}
        <VelocityJet />
        
        {/* Atmospheric effects */}
        <CloudLayer count={12} />
        <SpeedLines count={300} />
        <AtmosphericDust count={400} />
        
        {/* Environment */}
        <Stars 
          radius={200} 
          depth={50} 
          count={5000} 
          factor={4} 
          saturation={0.5}
          fade
          speed={0.5}
        />
        
        {/* Environment map for reflections */}
        <Environment preset="sunset" />
        
        {/* Ground plane (very distant) */}
        <mesh position={[0, -30, 0]} rotation={[-Math.PI / 2, 0, 0]}>
          <planeGeometry args={[200, 200]} />
          <meshBasicMaterial 
            color={BRAND.tarmac} 
            transparent 
            opacity={0.3}
          />
        </mesh>
      </Bvh>
    </>
  );
}

// Post-processing effects for cinematic look
function PostProcessing() {
  return (
    <EffectComposer>
      <DepthOfField focusDistance={0.02} focalLength={0.05} bokehScale={8} height={480} />
      <Bloom intensity={1.5} width={300} height={300} kernelSize={5} luminanceThreshold={0.4} luminanceSmoothing={0.8} />
      <ChromaticAberration offset={[0.001, 0.001]} />
      <Noise opacity={0.03} premultiply />
      <Vignette offset={0.3} darkness={0.4} />
    </EffectComposer>
  );
}

// Loading state
function Loader() {
  return (
    <Html center>
      <div className="flex flex-col items-center gap-4">
        <div className="w-16 h-16 border-4 border-brand-500 border-t-transparent rounded-full animate-spin" />
        <p className="text-white/80 text-lg font-medium tracking-wider">INITIALIZING VELOCITY...</p>
      </div>
    </Html>
  );
}

// Main export component
export function VelocityJetScene() {
  return (
    <div className="w-full h-full min-h-[800px] relative">
      <Canvas
        gl={{ 
          antialias: true, 
          alpha: true,
          powerPreference: 'high-performance',
          stencil: false,
          depth: true,
        }}
        dpr={[1, 2]}
        shadows
      >
        <Suspense fallback={<Loader />}>
          <VelocityScene />
          <PostProcessing />
        </Suspense>
      </Canvas>
      
      {/* UI Overlay */}
      <div className="absolute bottom-8 left-8 pointer-events-none">
        <div className="flex flex-col gap-2">
          <h2 className="text-4xl font-bold text-white tracking-tight">
            <span style={{ color: '#FF6B35' }}>Function</span>
            <span style={{ color: '#00D4FF' }}>Fly</span>
          </h2>
          <p className="text-white/60 text-sm tracking-widest uppercase">Velocity Edition</p>
        </div>
      </div>
      
      {/* Flight data overlay */}
      <div className="absolute top-8 right-8 pointer-events-none font-mono text-xs">
        <div className="flex flex-col gap-1 text-right">
          <div className="text-brand-500">ALT: 35,000 FT</div>
          <div className="text-brand-cyan">SPD: M 0.85</div>
          <div className="text-white/50">HDG: 247°</div>
        </div>
      </div>
    </div>
  );
}

export default VelocityJetScene;