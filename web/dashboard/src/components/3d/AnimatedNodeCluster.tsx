/**
 * Particle Swarm - Premium Organic Node Network
 * 
 * An organic, living network visualization featuring nodes that orbit
 * and pulse with energy, connected by flowing streams of particles.
 */

import { useRef, useMemo, useState } from 'react';
import { Canvas, useFrame, useThree } from '@react-three/fiber';
import { 
  OrbitControls, 
  Stars,
  Trail,
  Float,
  Text,
  Billboard,
  Line,
} from '@react-three/drei';
import * as THREE from 'three';

// Organic particle palette - bioluminescent ocean theme
const SWARM = {
  void: new THREE.Color('#020814'),
  deep: new THREE.Color('#030a1f'),
  azure: new THREE.Color('#0ea5e9'),
  cyan: new THREE.Color('#22d3ee'),
  violet: new THREE.Color('#8b5cf6'),
  magenta: new THREE.Color('#d946ef'),
  emerald: new THREE.Color('#10b981'),
  amber: new THREE.Color('#f59e0b'),
  white: new THREE.Color('#f8fafc'),
};

// Organic node with trail effect
interface SwarmNodeProps {
  initialPosition: [number, number, number];
  color: THREE.Color;
  label?: string;
  size?: number;
  orbitRadius?: number;
  orbitSpeed?: number;
  verticalAmplitude?: number;
  trailColor?: THREE.Color;
}

function SwarmNode({ 
  initialPosition, 
  color,
  size = 1, 
  label,
  orbitRadius = 3,
  orbitSpeed = 0.4,
  verticalAmplitude = 1,
  trailColor
}: SwarmNodeProps) {
  const meshRef = useRef<THREE.Mesh>(null);
  const [hovered, setHovered] = useState(false);
  
  // Random offsets for organic variety
  const orbitOffset = useMemo(() => Math.random() * Math.PI * 2, []);
  const speedVariation = useMemo(() => 0.7 + Math.random() * 0.6, []);
  const verticalOffset = useMemo(() => Math.random() * Math.PI, []);
  
  useFrame(({ clock }) => {
    if (!meshRef.current) return;
    
    const t = clock.getElapsedTime() * orbitSpeed * speedVariation + orbitOffset;
    
    // Orbital motion with vertical wave
    const x = initialPosition[0] + Math.cos(t) * orbitRadius;
    const y = initialPosition[1] + Math.sin(t * 0.5) * orbitRadius * verticalAmplitude + Math.sin(t + verticalOffset) * 0.5;
    const z = initialPosition[2] + Math.sin(t) * orbitRadius;
    
    meshRef.current.position.set(x, y, z);
    
    // Self rotation
    meshRef.current.rotation.x += 0.008;
    meshRef.current.rotation.y += 0.015;
    
    // Scale response
    if (hovered) {
      meshRef.current.scale.lerp(new THREE.Vector3(size * 1.25, size * 1.25, size * 1.25), 0.1);
    } else {
      meshRef.current.scale.lerp(new THREE.Vector3(size, size, size), 0.1);
    }
  });

  const mesh = (
    <mesh 
      ref={meshRef}
      onPointerOver={() => setHovered(true)}
      onPointerOut={() => setHovered(false)}
    >
      <icosahedronGeometry args={[1, 1]} />
      <meshStandardMaterial
        color={SWARM.deep}
        metalness={0.8}
        roughness={0.2}
        emissive={color}
        emissiveIntensity={hovered ? 0.9 : 0.4}
        flatShading
      />
    </mesh>
  );

  return (
    <>
      {trailColor ? (
        <Trail
          width={2}
          color={trailColor || color}
          length={6}
          decay={2}
          attenuation={(t) => t * t}
        >
          {mesh}
        </Trail>
      ) : mesh}
      
      {/* Glowing point at center */}
      <mesh position={initialPosition}>
        <sphereGeometry args={[0.1, 8, 8]} />
        <meshBasicMaterial color={color} />
      </mesh>
      
      {/* Label */}
      {label && (
        <Billboard position={[initialPosition[0] + orbitRadius * 0.8, initialPosition[1] - 2, initialPosition[2]]}>
          <Text
            fontSize={0.4}
            color="#ffffff"
            anchorX="center"
            anchorY="middle"
            outlineWidth={0.02}
            outlineColor={color}
          >
            {label}
          </Text>
        </Billboard>
      )}
    </>
  );
}

// Pulsing connection line between nodes
interface SwarmConnectionProps {
  start: [number, number, number];
  end: [number, number, number];
  color: THREE.Color;
  intensity?: number;
}

function SwarmConnection({ start, end, color, intensity = 0.6 }: SwarmConnectionProps) {
  const points = useMemo(() => [
    new THREE.Vector3(...start),
    new THREE.Vector3(...end)
  ], [start, end]);

  return (
    <Line
      points={points}
      color={color}
      lineWidth={1.5}
      transparent
      opacity={intensity}
    />
  );
}

// Bioluminescent particle swarm
function BioSwarm({ count = 250, center }: { count?: number; center: [number, number, number] }) {
  const points = useRef<THREE.Points>(null);
  
  const { positions, velocities, colors } = useMemo(() => {
    const positions: number[] = [];
    const velocities: number[] = [];
    const colors: number[] = [];
    
    for (let i = 0; i < count; i++) {
      // Spiral distribution around center
      const angle = Math.random() * Math.PI * 2;
      const radius = Math.random() * 18 + 4;
      const height = (Math.random() - 0.5) * 25;
      
      positions.push(
        center[0] + Math.cos(angle) * radius,
        center[1] + height,
        center[2] + Math.sin(angle) * radius
      );
      
      // Random drift velocities
      velocities.push(
        (Math.random() - 0.5) * 0.015,
        (Math.random() - 0.5) * 0.015,
        (Math.random() - 0.5) * 0.015
      );
      
      // Color variety - bioluminescent palette
      const choice = Math.random();
      let c;
      if (choice < 0.35) c = SWARM.cyan;
      else if (choice < 0.55) c = SWARM.violet;
      else if (choice < 0.75) c = SWARM.azure;
      else if (choice < 0.9) c = SWARM.emerald;
      else c = SWARM.magenta;
      
      colors.push(c.r, c.g, c.b);
    }
    
    return { positions, velocities, colors };
  }, [count, center]);

  useFrame((state, delta) => {
    if (!points.current) return;
    
    const posArray = points.current.geometry.attributes.position.array as Float32Array;
    
    for (let i = 0; i < count; i++) {
      const px = posArray[i * 3];
      const py = posArray[i * 3 + 1];
      const pz = posArray[i * 3 + 2];
      
      // Apply velocity
      posArray[i * 3] += velocities[i * 3];
      posArray[i * 3 + 1] += velocities[i * 3 + 1];
      posArray[i * 3 + 2] += velocities[i * 3 + 2];
      
      // Boundary bounce - keep particles in sphere
      const dist = Math.sqrt(
        (px - center[0]) ** 2 + 
        (py - center[1]) ** 2 + 
        (pz - center[2]) ** 2
      );
      
      if (dist > 20) {
        velocities[i * 3] *= -1;
        velocities[i * 3 + 1] *= -1;
        velocities[i * 3 + 2] *= -1;
      }
      
      // Add gentle turbulence
      velocities[i * 3] += (Math.random() - 0.5) * 0.0005;
      velocities[i * 3 + 1] += (Math.random() - 0.5) * 0.0005;
      velocities[i * 3 + 2] += (Math.random() - 0.5) * 0.0005;
    }
    
    points.current.geometry.attributes.position.needsUpdate = true;
    points.current.rotation.y += delta * 0.02;
  });

  return (
    <points ref={points}>
      <bufferGeometry>
        <bufferAttribute
          attach="attributes-position"
          count={count}
          array={new Float32Array(positions)}
          itemSize={3}
        />
        <bufferAttribute
          attach="attributes-color"
          count={count}
          array={new Float32Array(colors)}
          itemSize={3}
        />
      </bufferGeometry>
      <pointsMaterial
        size={0.1}
        vertexColors
        transparent
        opacity={0.75}
        sizeAttenuation
        blending={THREE.AdditiveBlending}
      />
    </points>
  );
}

// Ambient floating particles - distant
function AmbientDust({ count = 100 }: { count?: number }) {
  const points = useRef<THREE.Points>(null);
  
  const positions = useMemo(() => {
    const pos = new Float32Array(count * 3);
    for (let i = 0; i < count; i++) {
      pos[i * 3] = (Math.random() - 0.5) * 100;
      pos[i * 3 + 1] = (Math.random() - 0.5) * 60;
      pos[i * 3 + 2] = (Math.random() - 0.5) * 100;
    }
    return pos;
  }, [count]);

  useFrame(({ clock }) => {
    if (points.current) {
      points.current.rotation.y = clock.getElapsedTime() * 0.005;
    }
  });

  return (
    <points ref={points}>
      <bufferGeometry>
        <bufferAttribute
          attach="attributes-position"
          count={count}
          array={positions}
          itemSize={3}
        />
      </bufferGeometry>
      <pointsMaterial
        size={0.15}
        color={SWARM.azure}
        transparent
        opacity={0.3}
        sizeAttenuation
      />
    </points>
  );
}

// Main node cluster scene
function NodeClusterScene() {
  const center: [number, number, number] = [0, 0, 0];
  
  const mainNodes = useMemo(() => [
    // Central hub - largest
    { id: 'hub', position: [0, 0, 0] as [number, number, number], color: SWARM.cyan, label: 'HUB', size: 2, orbitRadius: 0, orbitSpeed: 0, trailColor: SWARM.cyan },
    
    // Primary orbit nodes
    { id: 'worker1', position: [6, 2, 3] as [number, number, number], color: SWARM.violet, label: 'WORKER', size: 1.2, orbitRadius: 4 + Math.random() * 2, orbitSpeed: 0.35, trailColor: SWARM.violet },
    { id: 'worker2', position: [-5, 1, -4] as [number, number, number], color: SWARM.azure, label: 'WORKER', size: 1.1, orbitRadius: 3 + Math.random() * 2, orbitSpeed: 0.4, trailColor: SWARM.azure },
    { id: 'worker3', position: [3, -3, 5] as [number, number, number], color: SWARM.emerald, label: 'WORKER', size: 1.1, orbitRadius: 4 + Math.random() * 2, orbitSpeed: 0.3, trailColor: SWARM.emerald },
    { id: 'worker4', position: [-4, -2, -5] as [number, number, number], color: SWARM.magenta, label: 'WORKER', size: 1.0, orbitRadius: 3.5 + Math.random() * 2, orbitSpeed: 0.45, trailColor: SWARM.magenta },
    
    // Secondary orbit nodes
    { id: 'gateway', position: [8, 0, 0] as [number, number, number], color: SWARM.amber, label: 'GATEWAY', size: 1.3, orbitRadius: 5 + Math.random() * 3, orbitSpeed: 0.25, verticalAmplitude: 0.5 },
    { id: 'storage', position: [-8, 0, 0] as [number, number, number], color: SWARM.emerald, label: 'STORAGE', size: 1.2, orbitRadius: 4.5 + Math.random() * 3, orbitSpeed: 0.28, verticalAmplitude: 0.6 },
    { id: 'cache', position: [0, 7, 0] as [number, number, number], color: SWARM.violet, label: 'CACHE', size: 1.0, orbitRadius: 3 + Math.random() * 2, orbitSpeed: 0.5, verticalAmplitude: 0.4 },
    { id: 'queue', position: [0, -7, 0] as [number, number, number], color: SWARM.cyan, label: 'QUEUE', size: 1.0, orbitRadius: 3 + Math.random() * 2, orbitSpeed: 0.55, verticalAmplitude: 0.4 },
  ], []);

  // Define connections
  const connections = useMemo(() => [
    { from: [6, 2, 3], to: [0, 0, 0], color: SWARM.violet },
    { from: [-5, 1, -4], to: [0, 0, 0], color: SWARM.azure },
    { from: [3, -3, 5], to: [0, 0, 0], color: SWARM.emerald },
    { from: [-4, -2, -5], to: [0, 0, 0], color: SWARM.magenta },
    { from: [8, 0, 0], to: [0, 0, 0], color: SWARM.amber },
    { from: [-8, 0, 0], to: [0, 0, 0], color: SWARM.emerald },
  ], []);

  return (
    <>
      {/* Lighting */}
      <ambientLight intensity={0.15} />
      <directionalLight position={[10, 10, 5]} intensity={0.6} color="#ffffff" />
      <pointLight position={[0, 0, 0]} intensity={2.5} color={SWARM.cyan} distance={35} />
      <pointLight position={[-10, 5, -10]} intensity={1} color={SWARM.violet} distance={30} />
      <pointLight position={[10, 5, 10]} intensity={1} color={SWARM.azure} distance={30} />
      
      {/* Background */}
      <Stars radius={100} depth={50} count={2000} factor={3} saturation={0.4} fade speed={0.4} />
      
      {/* Ambient dust */}
      <AmbientDust count={80} />
      
      {/* Dynamic connections */}
      {connections.map((conn, i) => (
        <SwarmConnection
          key={i}
          start={conn.from as [number, number, number]}
          end={conn.to as [number, number, number]}
          color={conn.color}
          intensity={0.4}
        />
      ))}
      
      {/* Orbital nodes with trails */}
      {mainNodes.map((node) => (
        <SwarmNode
          key={node.id}
          initialPosition={node.position}
          color={node.color}
          label={node.label}
          size={node.size}
          orbitRadius={node.orbitRadius}
          orbitSpeed={node.orbitSpeed || 0.3}
          verticalAmplitude={node.verticalAmplitude || 1}
          trailColor={node.trailColor}
        />
      ))}
      
      {/* Bioluminescent swarm */}
      <BioSwarm count={300} center={center} />
      
      {/* Controls */}
      <OrbitControls
        enablePan={true}
        enableZoom={true}
        enableRotate={true}
        autoRotate
        autoRotateSpeed={0.15}
        minDistance={10}
        maxDistance={45}
        target={[0, 0, 0]}
      />
    </>
  );
}

export function AnimatedNodeCluster() {
  return (
    <div className="w-full h-full min-h-[600px] relative">
      <Canvas
        camera={{ position: [18, 10, 18], fov: 55 }}
        gl={{ antialias: true, alpha: false, powerPreference: 'high-performance' }}
        dpr={[1, 1.5]}
        style={{ background: '#020814' }}
      >
        <color attach="background" args={['#020814']} />
        <fog attach="fog" args={['#020814', 35, 75]} />
        <NodeClusterScene />
      </Canvas>
      
      {/* Overlay */}
      <div className="absolute bottom-6 left-6 pointer-events-none">
        <div className="flex flex-col gap-1">
          <div className="text-xs font-mono text-white/40 tracking-widest uppercase">
            Particle Swarm
          </div>
          <div className="text-xs text-cyan-400/60">
            9 Nodes • 300 Particles • Active
          </div>
        </div>
      </div>
      
      {/* Interaction hint */}
      <div className="absolute top-4 right-4 pointer-events-none">
        <div className="text-xs text-white/30 font-mono">
          Hover nodes to highlight
        </div>
      </div>
    </div>
  );
}

export default AnimatedNodeCluster;