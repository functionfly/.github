/**
 * Neural Constellation - Premium 3D Graph Visualization
 * 
 * A distinctive, production-ready visualization of an AI neural network / function graph.
 * Features geometric node shapes, energy pulses along connections, and a deep space aesthetic
 * that feels like viewing a real-time AI processing network.
 */

import { useRef, useMemo, useEffect, useState, useCallback } from 'react';
import { Canvas, useFrame, useThree, extend } from '@react-three/fiber';
import { 
  OrbitControls, 
  Stars,
  Trail,
  Float,
  Sphere,
  Line,
  Text,
  Billboard,
  Html,
  MeshDistortMaterial,
  Shape,
} from '@react-three/drei';
import * as THREE from 'three';
import { motion } from 'framer-motion';

// Neural network brand palette - deep space with cyan/magenta energy
const NEURAL_BRAND = {
  void: new THREE.Color('#030014'),        // Deep space background
  cosmic: new THREE.Color('#0a0a1a'),     // Subtle gradient layer
  nodeCore: new THREE.Color('#1a1a2e'),   // Node center
  nodeOuter: new THREE.Color('#16213e'),   // Node glow
  energyCyan: new THREE.Color('#00f5ff'),  // Primary energy - cyan
  energyMagenta: new THREE.Color('#ff006e'), // Secondary energy - magenta
  energyGold: new THREE.Color('#ffd60a'),  // Accent energy - gold
  starWhite: new THREE.Color('#ffffff'),
  starBlue: new THREE.Color('#60a5fa'),
  connectionLine: new THREE.Color('#1e3a5f'), // Subtle connection
};

// Geometric node shape - hexagonal prism for visual distinction
function NeuralNodeGeometry({ 
  sides = 6, 
  radius = 1, 
  depth = 0.4 
}: { sides?: number; radius?: number; depth?: number }) {
  const shape = useMemo(() => {
    const s = new THREE.Shape();
    for (let i = 0; i < sides; i++) {
      const angle = (i / sides) * Math.PI * 2 - Math.PI / 2;
      const x = Math.cos(angle) * radius;
      const y = Math.sin(angle) * radius;
      if (i === 0) s.moveTo(x, y);
      else s.lineTo(x, y);
    }
    s.closePath();
    return s;
  }, [sides, radius]);

  return (
    <extrudeGeometry 
      args={[
        shape, 
        { depth, bevelEnabled: true, bevelThickness: 0.05, bevelSize: 0.05, bevelSegments: 2 }
      ]} 
    />
  );
}

// Premium node with multiple layers and effects
interface NeuralNodeProps {
  position: [number, number, number];
  color: THREE.Color;
  size?: number;
  label?: string;
  sublabel?: string;
  isHub?: boolean;
  pulseSpeed?: number;
  rotationOffset?: number;
}

function NeuralNode({ 
  position, 
  color, 
  size = 1, 
  label, 
  sublabel,
  isHub = false,
  pulseSpeed = 1,
  rotationOffset = 0
}: NeuralNodeProps) {
  const groupRef = useRef<THREE.Group>(null);
  const coreRef = useRef<THREE.Mesh>(null);
  const ringRef = useRef<THREE.Mesh>(null);
  const glowRef = useRef<THREE.Mesh>(null);
  const [hovered, setHovered] = useState(false);
  const [clicked, setClicked] = useState(false);

  useFrame(({ clock }) => {
    if (!groupRef.current) return;
    const t = clock.getElapsedTime() * pulseSpeed + rotationOffset;
    
    // Gentle floating animation
    groupRef.current.position.y = position[1] + Math.sin(t * 0.5) * 0.15;
    groupRef.current.rotation.y = t * 0.15;
    
    // Pulsing glow effect
    if (glowRef.current) {
      const pulse = (Math.sin(t * 2) + 1) * 0.5;
      const mat = glowRef.current.material as THREE.MeshBasicMaterial;
      mat.opacity = hovered ? 0.4 + pulse * 0.2 : 0.15 + pulse * 0.1;
    }
    
    // Ring rotation
    if (ringRef.current) {
      ringRef.current.rotation.z = t * 0.3;
      ringRef.current.rotation.x = t * 0.2;
    }
    
    // Scale response to hover
    if (coreRef.current) {
      const targetScale = hovered ? size * 1.15 : size;
      coreRef.current.scale.lerp(new THREE.Vector3(targetScale, targetScale, targetScale), 0.1);
    }
  });

  const nodeColor = color;

  return (
    <group 
      ref={groupRef} 
      position={position}
      onPointerOver={() => setHovered(true)}
      onPointerOut={() => setHovered(false)}
      onClick={() => setClicked(!clicked)}
    >
      {/* Outer glow halo */}
      <mesh ref={glowRef} scale={isHub ? 2.5 : 2}>
        <sphereGeometry args={[1, 16, 16]} />
        <meshBasicMaterial
          color={nodeColor}
          transparent
          opacity={0.15}
          side={THREE.BackSide}
        />
      </mesh>
      
      {/* Rotating ring */}
      <mesh ref={ringRef} scale={isHub ? 1.8 : 1.4}>
        <torusGeometry args={[1, 0.03, 8, 32]} />
        <meshBasicMaterial
          color={nodeColor}
          transparent
          opacity={0.4}
        />
      </mesh>
      
      {/* Main hexagonal core */}
      <mesh ref={coreRef}>
        <NeuralNodeGeometry sides={6} radius={0.8} depth={0.3} />
        <meshStandardMaterial
          color={NEURAL_BRAND.nodeCore}
          metalness={0.9}
          roughness={0.15}
          emissive={nodeColor}
          emissiveIntensity={hovered ? 0.6 : 0.25}
        />
      </mesh>
      
      {/* Inner glowing point */}
      <mesh scale={0.2}>
        <sphereGeometry args={[1, 16, 16]} />
        <meshBasicMaterial color={nodeColor} />
      </mesh>
      
      {/* Floating label - only show on hover or hub */}
      {(hovered || isHub) && label && (
        <Billboard position={[0, -1.8, 0]} follow lockZ={false}>
          <group>
            <Text
              fontSize={isHub ? 0.45 : 0.35}
              color="#ffffff"
              anchorX="center"
              anchorY="middle"
              font="/fonts/SpaceGrotesk-Bold.woff"
              outlineWidth={0.02}
              outlineColor={nodeColor}
            >
              {label}
            </Text>
            {sublabel && (
              <Text
                position={[0, -0.35, 0]}
                fontSize={0.2}
                color={nodeColor.getHexString()}
                anchorX="center"
                anchorY="middle"
                outlineWidth={0.01}
                outlineColor="#000000"
              >
                {sublabel}
              </Text>
            )}
          </group>
        </Billboard>
      )}
    </group>
  );
}

// Energy pulse traveling along a connection
interface EnergyPulseProps {
  start: [number, number, number];
  end: [number, number, number];
  color: THREE.Color;
  delay?: number;
  speed?: number;
}

function EnergyPulse({ start, end, color, delay = 0, speed = 1 }: EnergyPulseProps) {
  const meshRef = useRef<THREE.Mesh>(null);
  const progress = useRef(delay);

  useFrame((state, delta) => {
    progress.current += delta * speed;
    if (progress.current > 1) progress.current = 0;

    if (meshRef.current) {
      const t = progress.current;
      const x = THREE.MathUtils.lerp(start[0], end[0], t);
      const y = THREE.MathUtils.lerp(start[1], end[1], t);
      const z = THREE.MathUtils.lerp(start[2], end[2], t);
      meshRef.current.position.set(x, y, z);
      
      // Scale pulse as it travels
      const scale = Math.sin(t * Math.PI) * 0.3 + 0.1;
      meshRef.current.scale.setScalar(scale);
    }
  });

  return (
    <mesh ref={meshRef}>
      <sphereGeometry args={[1, 8, 8]} />
      <meshBasicMaterial
        color={color}
        transparent
        opacity={0.9}
      />
    </mesh>
  );
}

// Connection between nodes with energy pulses
interface NeuralConnectionProps {
  start: [number, number, number];
  end: [number, number, number];
  color: THREE.Color;
  intensity?: number;
}

function NeuralConnection({ start, end, color, intensity = 1 }: NeuralConnectionProps) {
  const points = useMemo(() => [
    new THREE.Vector3(...start),
    new THREE.Vector3(...end)
  ], [start, end]);

  const midpoint = useMemo(() => {
    return [
      (start[0] + end[0]) / 2,
      (start[1] + end[1]) / 2,
      (start[2] + end[2]) / 2,
    ] as [number, number, number];
  }, [start, end]);

  return (
    <group>
      {/* Subtle base line */}
      <Line
        points={points}
        color={NEURAL_BRAND.connectionLine}
        lineWidth={1}
        transparent
        opacity={0.3}
      />
      
      {/* Energy pulses traveling along connection */}
      <EnergyPulse start={start} end={end} color={color} delay={0} speed={0.8} />
      <EnergyPulse start={start} end={end} color={color} delay={0.5} speed={0.8} />
      <EnergyPulse start={start} end={end} color={color} delay={0.25} speed={1.2} />
    </group>
  );
}

// Ambient particle field - cosmic dust
function CosmicDust({ count = 300 }: { count?: number }) {
  const points = useRef<THREE.Points>(null);
  
  const { positions, colors, sizes } = useMemo(() => {
    const positions = new Float32Array(count * 3);
    const colors = new Float32Array(count * 3);
    const sizes = new Float32Array(count);
    
    for (let i = 0; i < count; i++) {
      // Spherical distribution
      const theta = Math.random() * Math.PI * 2;
      const phi = Math.acos(2 * Math.random() - 1);
      const r = 15 + Math.random() * 35;
      
      positions[i * 3] = r * Math.sin(phi) * Math.cos(theta);
      positions[i * 3 + 1] = r * Math.sin(phi) * Math.sin(theta);
      positions[i * 3 + 2] = r * Math.cos(phi);
      
      // Random color from brand palette
      const choice = Math.random();
      if (choice < 0.4) {
        colors[i * 3] = NEURAL_BRAND.energyCyan.r;
        colors[i * 3 + 1] = NEURAL_BRAND.energyCyan.g;
        colors[i * 3 + 2] = NEURAL_BRAND.energyCyan.b;
      } else if (choice < 0.7) {
        colors[i * 3] = NEURAL_BRAND.energyMagenta.r;
        colors[i * 3 + 1] = NEURAL_BRAND.energyMagenta.g;
        colors[i * 3 + 2] = NEURAL_BRAND.energyMagenta.b;
      } else {
        colors[i * 3] = NEURAL_BRAND.starWhite.r;
        colors[i * 3 + 1] = NEURAL_BRAND.starWhite.g;
        colors[i * 3 + 2] = NEURAL_BRAND.starWhite.b;
      }
      
      sizes[i] = 0.05 + Math.random() * 0.15;
    }
    
    return { positions, colors, sizes };
  }, [count]);

  useFrame(({ clock }) => {
    if (points.current) {
      points.current.rotation.y = clock.getElapsedTime() * 0.01;
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
        <bufferAttribute
          attach="attributes-color"
          count={count}
          array={colors}
          itemSize={3}
        />
      </bufferGeometry>
      <pointsMaterial
        size={0.1}
        vertexColors
        transparent
        opacity={0.6}
        sizeAttenuation
        blending={THREE.AdditiveBlending}
      />
    </points>
  );
}

// Constellation grid - subtle background structure
function ConstellationGrid() {
  const linesRef = useRef<THREE.Group>(null);
  
  const gridLines = useMemo(() => {
    const lines: [THREE.Vector3, THREE.Vector3][] = [];
    const size = 40;
    const divisions = 8;
    const step = size / divisions;
    
    // Horizontal rings
    for (let i = 0; i <= divisions; i++) {
      const y = -size / 2 + i * step;
      for (let j = 0; j < 8; j++) {
        const angle = (j / 8) * Math.PI * 2;
        lines.push([
          new THREE.Vector3(Math.cos(angle) * size, y, Math.sin(angle) * size),
          new THREE.Vector3(Math.cos(angle + Math.PI / 4) * size, y, Math.sin(angle + Math.PI / 4) * size),
        ]);
      }
    }
    
    return lines;
  }, []);

  useFrame(({ clock }) => {
    if (linesRef.current) {
      linesRef.current.rotation.y = clock.getElapsedTime() * 0.02;
    }
  });

  return (
    <group ref={linesRef}>
      {gridLines.map((line, i) => (
        <Line
          key={i}
          points={line}
          color={NEURAL_BRAND.connectionLine}
          lineWidth={0.5}
          transparent
          opacity={0.15}
        />
      ))}
    </group>
  );
}

// Main scene graph
function ConstellationGraph() {
  const { viewport } = useThree();
  
  // Define nodes with meaningful labels
  const nodes = useMemo(() => [
    // Hub node - center
    { id: 'hub', position: [0, 0, 0] as [number, number, number], color: NEURAL_BRAND.energyCyan, label: 'NEURAL HUB', sublabel: '12.4M params', isHub: true, size: 1.4, rotationOffset: 0 },
    
    // Layer 1 - Input nodes
    { id: 'input1', position: [-7, 2, -4] as [number, number, number], color: NEURAL_BRAND.energyCyan, label: 'INPUT A', sublabel: 'webhook', size: 0.9, rotationOffset: 1 },
    { id: 'input2', position: [-6, -2, 3] as [number, number, number], color: NEURAL_BRAND.energyCyan, label: 'INPUT B', sublabel: 'stream', size: 0.9, rotationOffset: 2 },
    { id: 'input3', position: [-8, 0, 0] as [number, number, number], color: NEURAL_BRAND.energyCyan, label: 'INPUT C', sublabel: 'api', size: 0.9, rotationOffset: 3 },
    
    // Layer 2 - Processing nodes
    { id: 'process1', position: [-2, 4, -2] as [number, number, number], color: NEURAL_BRAND.energyMagenta, label: 'TRANSFORM', sublabel: 'gpt-4o-mini', size: 1.1, rotationOffset: 4 },
    { id: 'process2', position: [-1, -3, 4] as [number, number, number], color: NEURAL_BRAND.energyMagenta, label: 'ENRICH', sublabel: 'context', size: 1.0, rotationOffset: 5 },
    { id: 'process3', position: [0, 2, -5] as [number, number, number], color: NEURAL_BRAND.energyMagenta, label: 'FILTER', sublabel: 'validation', size: 0.95, rotationOffset: 6 },
    
    // Layer 3 - Output nodes
    { id: 'output1', position: [6, 3, 2] as [number, number, number], color: NEURAL_BRAND.energyGold, label: 'OUTPUT A', sublabel: 'response', size: 1.0, rotationOffset: 7 },
    { id: 'output2', position: [7, -1, -3] as [number, number, number], color: NEURAL_BRAND.energyGold, label: 'OUTPUT B', sublabel: 'store', size: 0.9, rotationOffset: 8 },
    { id: 'output3', position: [5, 0, 5] as [number, number, number], color: NEURAL_BRAND.energyGold, label: 'OUTPUT C', sublabel: 'trigger', size: 0.9, rotationOffset: 9 },
  ], []);

  // Define connections (from -> to)
  const connections = useMemo(() => [
    // Input to hub
    { from: 'input1', to: 'hub' },
    { from: 'input2', to: 'hub' },
    { from: 'input3', to: 'hub' },
    
    // Hub to process
    { from: 'hub', to: 'process1' },
    { from: 'hub', to: 'process2' },
    { from: 'hub', to: 'process3' },
    
    // Process to output
    { from: 'process1', to: 'output1' },
    { from: 'process2', to: 'output2' },
    { from: 'process3', to: 'output3' },
    
    // Cross connections for visual interest
    { from: 'process1', to: 'process2' },
    { from: 'process2', to: 'process3' },
    { from: 'output1', to: 'output2' },
  ], []);

  // Build node map for connection lookup
  const nodeMap = useMemo(() => {
    const map = new Map<string, typeof nodes[0]>();
    nodes.forEach(n => map.set(n.id, n));
    return map;
  }, [nodes]);

  return (
    <>
      {/* Ambient lighting */}
      <ambientLight intensity={0.2} />
      <directionalLight position={[10, 10, 5]} intensity={0.8} color="#ffffff" />
      <pointLight position={[0, 0, 0]} intensity={2} color={NEURAL_BRAND.energyCyan} distance={25} />
      <pointLight position={[-10, 5, -5]} intensity={0.8} color={NEURAL_BRAND.energyMagenta} distance={20} />
      <pointLight position={[10, 5, 5]} intensity={0.8} color={NEURAL_BRAND.energyGold} distance={20} />
      
      {/* Background stars */}
      <Stars 
        radius={80} 
        depth={50} 
        count={2500} 
        factor={3} 
        saturation={0.3} 
        fade 
        speed={0.3}
      />
      
      {/* Background grid */}
      <ConstellationGrid />
      
      {/* Cosmic dust particles */}
      <CosmicDust count={200} />
      
      {/* Neural connections */}
      {connections.map((conn, i) => {
        const fromNode = nodeMap.get(conn.from);
        const toNode = nodeMap.get(conn.to);
        if (!fromNode || !toNode) return null;
        
        return (
          <NeuralConnection
            key={i}
            start={fromNode.position}
            end={toNode.position}
            color={toNode.color}
            intensity={1}
          />
        );
      })}
      
      {/* Neural nodes */}
      {nodes.map((node) => (
        <NeuralNode
          key={node.id}
          position={node.position}
          color={node.color}
          label={node.label}
          sublabel={node.sublabel}
          isHub={node.isHub}
          size={node.size}
          pulseSpeed={0.8}
          rotationOffset={node.rotationOffset}
        />
      ))}
      
      {/* Camera controls */}
      <OrbitControls
        enablePan={true}
        enableZoom={true}
        enableRotate={true}
        autoRotate
        autoRotateSpeed={0.3}
        minDistance={8}
        maxDistance={50}
        target={[0, 0, 0]}
      />
    </>
  );
}

// Main export component
export function GraphNetwork3D() {
  return (
    <div className="w-full h-full min-h-[600px] relative">
      <Canvas
        camera={{ position: [12, 8, 12], fov: 55 }}
        gl={{ 
          antialias: true, 
          alpha: false,
          powerPreference: 'high-performance',
        }}
        dpr={[1, 1.5]}
        style={{ background: '#030014' }}
      >
        <color attach="background" args={['#030014']} />
        <fog attach="fog" args={['#030014', 30, 80]} />
        <ConstellationGraph />
      </Canvas>
      
      {/* Overlay UI - subtle branding */}
      <div className="absolute bottom-6 left-6 pointer-events-none">
        <div className="flex flex-col gap-1">
          <div className="text-xs font-mono text-white/40 tracking-widest uppercase">
            Neural Constellation
          </div>
          <div className="text-xs text-cyan-400/60">
            12 Active Nodes • 847M connections/sec
          </div>
        </div>
      </div>
      
      {/* Interaction hint */}
      <div className="absolute top-4 right-4 pointer-events-none">
        <div className="text-xs text-white/30 font-mono">
          Drag to rotate • Scroll to zoom
        </div>
      </div>
    </div>
  );
}

export default GraphNetwork3D;