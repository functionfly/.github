/**
 * Animated Node Cluster 3D Visualization
 * Swarming, organic movement of interconnected nodes
 */

import { useRef, useMemo, useState } from 'react';
import { Canvas, useFrame } from '@react-three/fiber';
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

// Organic moving node
interface OrganicNodeProps {
  initialPosition: [number, number, number];
  color: string;
  label?: string;
  size?: number;
  orbitRadius?: number;
  orbitSpeed?: number;
}

function OrganicNode({ 
  initialPosition, 
  color, 
  size = 1, 
  label,
  orbitRadius = 2,
  orbitSpeed = 0.5
}: OrganicNodeProps) {
  const meshRef = useRef<THREE.Mesh>(null);
  const [hovered, setHovered] = useState(false);
  
  // Orbit parameters
  const orbitOffset = useMemo(() => Math.random() * Math.PI * 2, []);
  const yOffset = useMemo(() => Math.random() * 2 - 1, []);
  const speedVariation = useMemo(() => 0.8 + Math.random() * 0.4, []);
  
  useFrame(({ clock }) => {
    if (meshRef.current) {
      const t = clock.getElapsedTime() * orbitSpeed * speedVariation + orbitOffset;
      
      // Orbital motion
      const x = initialPosition[0] + Math.cos(t) * orbitRadius;
      const y = initialPosition[1] + Math.sin(t * 0.5) * orbitRadius * 0.5 + yOffset * Math.sin(t * 0.3);
      const z = initialPosition[2] + Math.sin(t) * orbitRadius;
      
      meshRef.current.position.set(x, y, z);
      meshRef.current.rotation.x += 0.01;
      meshRef.current.rotation.y += 0.02;
      
      // Scale on hover
      const targetScale = hovered ? size * 1.3 : size;
      meshRef.current.scale.lerp(new THREE.Vector3(targetScale, targetScale, targetScale), 0.1);
    }
  });

  return (
    <>
      {/* Trail effect */}
      <Trail
        width={2}
        color={color}
        length={8}
        decay={2}
        attenuation={(t) => t * t}
      >
        <mesh 
          ref={meshRef}
          onPointerOver={() => setHovered(true)}
          onPointerOut={() => setHovered(false)}
        >
          <sphereGeometry args={[1, 32, 32]} />
          <meshStandardMaterial
            color={color}
            emissive={color}
            emissiveIntensity={hovered ? 1 : 0.5}
            roughness={0.2}
            metalness={0.9}
          />
        </mesh>
      </Trail>
      
      {/* Label */}
      <Billboard>
        <Text
          position={[
            initialPosition[0] + orbitRadius,
            initialPosition[1] - 1.5,
            initialPosition[2]
          ]}
          fontSize={0.5}
          color="white"
          anchorX="center"
          outlineWidth={0.02}
          outlineColor="black"
        >
          {label}
        </Text>
      </Billboard>
    </>
  );
}

// Dynamic connector lines between nearest nodes - using Line from drei
function DynamicConnector({ start, end, intensity }: { start: [number, number, number]; end: [number, number, number]; intensity: number }) {
  const points = useMemo(() => [
    new THREE.Vector3(...start),
    new THREE.Vector3(...end)
  ], [start, end]);

  return (
    <Line
      points={points}
      color="#8b5cf6"
      lineWidth={1}
      transparent
      opacity={intensity * 0.6}
    />
  );
}

// Swarm of small particles
function ParticleSwarm({ count = 200, center }: { count?: number; center: [number, number, number] }) {
  const points = useRef<THREE.Points>(null);
  
  const { positions, velocities, colors } = useMemo(() => {
    const positions = [];
    const velocities = [];
    const colors = [];
    
    for (let i = 0; i < count; i++) {
      // Spiral distribution
      const angle = Math.random() * Math.PI * 2;
      const radius = Math.random() * 15 + 5;
      const height = (Math.random() - 0.5) * 20;
      
      positions.push(
        center[0] + Math.cos(angle) * radius,
        center[1] + height,
        center[2] + Math.sin(angle) * radius
      );
      
      velocities.push(
        (Math.random() - 0.5) * 0.02,
        (Math.random() - 0.5) * 0.02,
        (Math.random() - 0.5) * 0.02
      );
      
      // Purple/blue color variation
      colors.push(0.4 + Math.random() * 0.2);
      colors.push(0.2 + Math.random() * 0.3);
      colors.push(0.9 + Math.random() * 0.1);
    }
    
    return { positions, velocities, colors };
  }, [count, center]);

  useFrame((state, delta) => {
    if (!points.current) return;
    
    const positionArray = points.current.geometry.attributes.position.array as Float32Array;
    
    for (let i = 0; i < count; i++) {
      // Swarm behavior - move towards center while maintaining orbit
      const px = positionArray[i * 3];
      const py = positionArray[i * 3 + 1];
      const pz = positionArray[i * 3 + 2];
      
      // Apply velocity
      positionArray[i * 3] += velocities[i * 3];
      positionArray[i * 3 + 1] += velocities[i * 3 + 1];
      positionArray[i * 3 + 2] += velocities[i * 3 + 2];
      
      // Gentle pull towards center
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
      
      // Add some noise
      velocities[i * 3] += (Math.random() - 0.5) * 0.001;
      velocities[i * 3 + 1] += (Math.random() - 0.5) * 0.001;
      velocities[i * 3 + 2] += (Math.random() - 0.5) * 0.001;
    }
    
    points.current.geometry.attributes.position.needsUpdate = true;
    
    // Rotate entire swarm slowly
    points.current.rotation.y += delta * 0.05;
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
        size={0.08}
        vertexColors
        transparent
        opacity={0.7}
        sizeAttenuation
      />
    </points>
  );
}

// Main scene
function NodeClusterScene() {
  const center: [number, number, number] = [0, 0, 0];
  
  const mainNodes = useMemo(() => [
    { id: 'hub', position: [0, 0, 0] as [number, number, number], color: '#8b5cf6', label: 'HUB', size: 2 },
    { id: 'worker1', position: [5, 3, 2] as [number, number, number], color: '#3b82f6', label: 'W1', size: 1 },
    { id: 'worker2', position: [-4, 2, -3] as [number, number, number], color: '#10b981', label: 'W2', size: 1 },
    { id: 'worker3', position: [2, -4, 4] as [number, number, number], color: '#f59e0b', label: 'W3', size: 1 },
    { id: 'worker4', position: [-3, -2, -4] as [number, number, number], color: '#ec4899', label: 'W4', size: 1 },
    { id: 'gateway', position: [7, 0, 0] as [number, number, number], color: '#06b6d4', label: 'GW', size: 1.5 },
    { id: 'storage', position: [-7, 0, 0] as [number, number, number], color: '#6366f1', label: 'DB', size: 1.5 },
    { id: 'cache', position: [0, 6, 0] as [number, number, number], color: '#84cc16', label: 'CACHE', size: 1.2 },
  ], []);

  return (
    <>
      <ambientLight intensity={0.3} />
      <directionalLight position={[10, 10, 5]} intensity={1.5} />
      <pointLight position={[0, 0, 0]} intensity={3} color="#8b5cf6" distance={30} />
      
      <Stars radius={100} depth={50} count={2000} factor={4} saturation={0.5} fade speed={0.5} />
      
      {/* Main orbital nodes */}
      {mainNodes.map((node) => (
        <OrganicNode
          key={node.id}
          initialPosition={node.position}
          color={node.color}
          size={node.size}
          label={node.label}
          orbitRadius={node.id === 'hub' ? 0 : 2 + Math.random()}
          orbitSpeed={node.id === 'hub' ? 0 : 0.3 + Math.random() * 0.3}
        />
      ))}
      
      {/* Dynamic connections - simplified for demo */}
      <DynamicConnector start={[5, 3, 2]} end={[0, 0, 0]} intensity={0.5} />
      <DynamicConnector start={[-4, 2, -3]} end={[0, 0, 0]} intensity={0.5} />
      <DynamicConnector start={[2, -4, 4]} end={[0, 0, 0]} intensity={0.5} />
      
      {/* Particle swarm */}
      <ParticleSwarm count={300} center={center} />
      
      <OrbitControls
        enablePan={true}
        enableZoom={true}
        enableRotate={true}
        autoRotate
        autoRotateSpeed={0.2}
        minDistance={10}
        maxDistance={40}
      />
    </>
  );
}

export function AnimatedNodeCluster() {
  return (
    <div className="w-full h-full min-h-[600px]">
      <Canvas
        camera={{ position: [15, 10, 15], fov: 60 }}
        gl={{ antialias: true, alpha: true }}
        dpr={[1, 2]}
      >
        <NodeClusterScene />
      </Canvas>
    </div>
  );
}

export default AnimatedNodeCluster;
