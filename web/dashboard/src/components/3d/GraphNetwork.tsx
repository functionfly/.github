/**
 * 3D Graph Network Visualization
 * Animated nodes connected by flowing data streams
 */

import { useRef, useMemo, useEffect, useState } from 'react';
import { Canvas, useFrame, useThree } from '@react-three/fiber';
import { 
  OrbitControls, 
  Stars, 
  Trail, 
  Float,
  Sphere,
  Line,
  Text,
  Billboard,
  useTexture,
  Html,
} from '@react-three/drei';
import * as THREE from 'three';
import { motion } from 'framer-motion';

// Animated node component
interface NodeProps {
  position: [number, number, number];
  color: string;
  size?: number;
  label?: string;
  isActive?: boolean;
  pulseSpeed?: number;
}

function GraphNode({ position, color, size = 1, label, isActive = true, pulseSpeed = 1 }: NodeProps) {
  const meshRef = useRef<THREE.Mesh>(null);
  const glowRef = useRef<THREE.Mesh>(null);
  
  useFrame(({ clock }) => {
    if (meshRef.current && isActive) {
      const pulse = Math.sin(clock.getElapsedTime() * pulseSpeed) * 0.1 + 1;
      meshRef.current.scale.setScalar(pulse * size);
    }
    if (glowRef.current && isActive) {
      const glow = Math.sin(clock.getElapsedTime() * pulseSpeed * 0.5) * 0.2 + 0.8;
      glowRef.current.material.opacity = glow * 0.3;
    }
  });

  return (
    <Float rotationIntensity={0.2} floatIntensity={0.3}>
      <group position={position}>
        {/* Main node sphere */}
        <mesh ref={meshRef}>
          <sphereGeometry args={[1, 32, 32]} />
          <meshStandardMaterial
            color={color}
            emissive={color}
            emissiveIntensity={0.5}
            roughness={0.2}
            metalness={0.8}
          />
        </mesh>
        
        {/* Glow effect */}
        <mesh ref={glowRef} scale={2}>
          <sphereGeometry args={[1, 32, 32]} />
          <meshBasicMaterial
            color={color}
            transparent
            opacity={0.3}
          />
        </mesh>
        
        {/* Label */}
        {label && (
          <Billboard>
            <Text
              position={[0, -1.8, 0]}
              fontSize={0.5}
              color="white"
              anchorX="center"
              anchorY="middle"
              outlineWidth={0.02}
              outlineColor="black"
            >
              {label}
            </Text>
          </Billboard>
        )}
        
        {/* Connection points */}
        <mesh position={[0, 0, 1.2]}>
          <sphereGeometry args={[0.2, 16, 16]} />
          <meshStandardMaterial color="white" emissive="white" emissiveIntensity={1} />
        </mesh>
      </group>
    </Float>
  );
}

// Animated data packet flowing along edge
interface DataPacketProps {
  start: [number, number, number];
  end: [number, number, number];
  speed?: number;
  color?: string;
}

function DataPacket({ start, end, speed = 1, color = "#8b5cf6" }: DataPacketProps) {
  const meshRef = useRef<THREE.Mesh>(null);
  const progress = useRef(0);
  
  useFrame((state, delta) => {
    progress.current += delta * speed;
    if (progress.current > 1) progress.current = 0;
    
    if (meshRef.current) {
      const x = THREE.MathUtils.lerp(start[0], end[0], progress.current);
      const y = THREE.MathUtils.lerp(start[1], end[1], progress.current);
      const z = THREE.MathUtils.lerp(start[2], end[2], progress.current);
      meshRef.current.position.set(x, y, z);
    }
  });

  return (
    <mesh ref={meshRef}>
      <sphereGeometry args={[0.15, 16, 16]} />
      <meshStandardMaterial
        color={color}
        emissive={color}
        emissiveIntensity={2}
      />
    </mesh>
  );
}

// Connection edge with animated flow
interface EdgeProps {
  start: [number, number, number];
  end: [number, number, number];
  color?: string;
  particleCount?: number;
}

function GraphEdge({ start, end, color = "#6366f1", particleCount = 3 }: EdgeProps) {
  const lineRef = useRef<THREE.BufferGeometry>(null);
  
  const points = useMemo(() => {
    return [new THREE.Vector3(...start), new THREE.Vector3(...end)];
  }, [start, end]);

  return (
    <group>
      {/* Main connection line */}
      <Line
        points={points}
        color={color}
        lineWidth={2}
        transparent
        opacity={0.6}
      />
      
      {/* Glowing core line */}
      <Line
        points={points}
        color={color}
        lineWidth={4}
        transparent
        opacity={0.2}
      />
      
      {/* Flowing particles */}
      {Array.from({ length: particleCount }).map((_, i) => (
        <DataPacket
          key={i}
          start={start}
          end={end}
          speed={0.5 + i * 0.2}
          color={color}
        />
      ))}
    </group>
  );
}

// Main graph scene
function GraphScene() {
  const { viewport } = useThree();
  
  // Node positions
  const nodes = useMemo(() => [
    { id: 'input', position: [-8, 0, 0], color: '#3b82f6', label: 'Input' },
    { id: 'transform', position: [-3, 2, 2], color: '#8b5cf6', label: 'Transform' },
    { id: 'process', position: [2, -1, -2], color: '#10b981', label: 'Process' },
    { id: 'analyze', position: [3, 3, 1], color: '#f59e0b', label: 'Analyze' },
    { id: 'output', position: [8, 0, 0], color: '#ec4899', label: 'Output' },
    { id: 'cache', position: [0, -4, 3], color: '#6366f1', label: 'Cache' },
  ], []);

  const edges = useMemo(() => [
    { from: 'input', to: 'transform' },
    { from: 'input', to: 'process' },
    { from: 'transform', to: 'analyze' },
    { from: 'process', to: 'analyze' },
    { from: 'process', to: 'cache' },
    { from: 'analyze', to: 'output' },
    { from: 'cache', to: 'output' },
  ], []);

  const nodeMap = useMemo(() => {
    const map = new Map();
    nodes.forEach(n => map.set(n.id, n.position));
    return map;
  }, [nodes]);

  return (
    <>
      {/* Environment */}
      <ambientLight intensity={0.5} />
      <directionalLight position={[10, 10, 5]} intensity={1} />
      <pointLight position={[0, 0, 0]} intensity={0.8} color="#8b5cf6" />
      
      {/* Stars background */}
      <Stars radius={100} depth={50} count={5000} factor={4} saturation={0} fade speed={1} />
      
      {/* Graph nodes */}
      {nodes.map((node) => (
        <GraphNode
          key={node.id}
          position={node.position as [number, number, number]}
          color={node.color}
          label={node.label}
          size={1.2}
          pulseSpeed={1.5 + Math.random()}
        />
      ))}
      
      {/* Graph edges */}
      {edges.map((edge, i) => (
        <GraphEdge
          key={i}
          start={nodeMap.get(edge.from) as [number, number, number]}
          end={nodeMap.get(edge.to) as [number, number, number]}
          color="#6366f1"
          particleCount={2}
        />
      ))}
      
      {/* Floating particles */}
      <ParticleField count={100} />
      
      {/* Orbit controls */}
      <OrbitControls
        enablePan={true}
        enableZoom={true}
        enableRotate={true}
        autoRotate
        autoRotateSpeed={0.5}
        minDistance={5}
        maxDistance={50}
      />
    </>
  );
}

// Floating particle field
function ParticleField({ count = 100 }) {
  const points = useMemo(() => {
    const positions = new Float32Array(count * 3);
    for (let i = 0; i < count; i++) {
      positions[i * 3] = (Math.random() - 0.5) * 50;
      positions[i * 3 + 1] = (Math.random() - 0.5) * 50;
      positions[i * 3 + 2] = (Math.random() - 0.5) * 50;
    }
    return positions;
  }, [count]);

  return (
    <points>
      <bufferGeometry>
        <bufferAttribute
          attach="attributes-position"
          count={count}
          array={points}
          itemSize={3}
        />
      </bufferGeometry>
      <pointsMaterial
        size={0.1}
        color="#8b5cf6"
        transparent
        opacity={0.8}
        sizeAttenuation
      />
    </points>
  );
}

// Main component export
export function GraphNetwork3D() {
  return (
    <div className="w-full h-full min-h-[500px]">
      <Canvas
        camera={{ position: [15, 10, 15], fov: 60 }}
        gl={{ antialias: true, alpha: true }}
        dpr={[1, 2]}
      >
        <GraphScene />
      </Canvas>
    </div>
  );
}

export default GraphNetwork3D;
