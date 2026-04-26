/**
 * Quantum Crystal - Premium Geometric Graph Visualization
 * 
 * A crystalline, gem-like visualization of interconnected functions.
 * Features faceted geometric nodes, energy beams, and a dark quantum aesthetic.
 */

import { useRef, useMemo, useState } from 'react';
import { Canvas, useFrame, useThree } from '@react-three/fiber';
import { 
  OrbitControls, 
  Stars,
  Float,
  Text,
  Html,
  Billboard,
  MeshDistortMaterial,
  Line,
  Shape,
} from '@react-three/drei';
import * as THREE from 'three';
import { motion } from 'framer-motion';

// Quantum crystal palette - dark with gem-like accents
const CRYSTAL = {
  void: new THREE.Color('#050510'),
  obsidian: new THREE.Color('#0a0a1f'),
  emerald: new THREE.Color('#10b981'),
  sapphire: new THREE.Color('#3b82f6'),
  ruby: new THREE.Color('#ef4444'),
  amethyst: new THREE.Color('#a855f7'),
  topaz: new THREE.Color('#f59e0b'),
  diamond: new THREE.Color('#e0e7ff'),
  pearl: new THREE.Color('#f8fafc'),
};

// Create a gem-like octahedron geometry
function GemGeometry({ size = 1 }: { size?: number }) {
  const geometry = useMemo(() => {
    const geo = new THREE.OctahedronGeometry(size, 0);
    return geo;
  }, [size]);
  return <primitive object={geometry} />;
}

// Premium crystal node with facets and inner glow
interface CrystalGemProps {
  position: [number, number, number];
  color: THREE.Color;
  size?: number;
  label?: string;
  description?: string;
  onClick?: () => void;
  pulseOffset?: number;
}

function CrystalGem({ 
  position, 
  color, 
  size = 1.2, 
  label, 
  description,
  onClick,
  pulseOffset = 0
}: CrystalGemProps) {
  const meshRef = useRef<THREE.Group>(null);
  const coreRef = useRef<THREE.Mesh>(null);
  const outerRef = useRef<THREE.Mesh>(null);
  const [hovered, setHovered] = useState(false);
  
  useFrame(({ clock }) => {
    if (!meshRef.current) return;
    const t = clock.getElapsedTime() + pulseOffset;
    
    // Gentle rotation and bobbing
    meshRef.current.rotation.y = t * 0.2;
    meshRef.current.rotation.x = Math.sin(t * 0.3) * 0.1;
    meshRef.current.position.y = position[1] + Math.sin(t * 0.5) * 0.1;
    
    // Pulsing glow
    if (outerRef.current) {
      const pulse = (Math.sin(t * 2) + 1) * 0.3 + 0.7;
      const mat = outerRef.current.material as THREE.MeshBasicMaterial;
      mat.opacity = hovered ? 0.35 * pulse : 0.15 * pulse;
    }
    
    // Scale on hover
    if (coreRef.current) {
      const targetScale = hovered ? size * 1.2 : size;
      coreRef.current.scale.lerp(new THREE.Vector3(targetScale, targetScale, targetScale), 0.1);
    }
  });

  const gemColor = color;

  return (
    <Float rotationIntensity={0.3} floatIntensity={0.4} speed={2}>
      <group 
        ref={meshRef}
        position={position}
        onPointerOver={() => setHovered(true)}
        onPointerOut={() => setHovered(false)}
        onClick={onClick}
      >
        {/* Outer transparent glow */}
        <mesh ref={outerRef} scale={1.8}>
          <sphereGeometry args={[1, 8, 8]} />
          <meshBasicMaterial
            color={gemColor}
            transparent
            opacity={0.15}
            side={THREE.BackSide}
          />
        </mesh>
        
        {/* Faceted gem core */}
        <mesh ref={coreRef}>
          <primitive object={new THREE.OctahedronGeometry(1, 0)} />
          <meshStandardMaterial
            color={CRYSTAL.obsidian}
            metalness={0.95}
            roughness={0.05}
            emissive={gemColor}
            emissiveIntensity={hovered ? 0.8 : 0.4}
            flatShading
          />
        </mesh>
        
        {/* Inner light point */}
        <mesh scale={0.25}>
          <sphereGeometry args={[1, 8, 8]} />
          <meshBasicMaterial color={gemColor} />
        </mesh>
        
        {/* Wireframe overlay */}
        <mesh scale={1.05}>
          <primitive object={new THREE.OctahedronGeometry(1, 0)} />
          <meshBasicMaterial
            color={gemColor}
            wireframe
            transparent
            opacity={hovered ? 0.4 : 0.15}
          />
        </mesh>
        
        {/* Floating label */}
        <Billboard position={[0, -2, 0]} follow lockZ={false}>
          <group>
            <Text
              fontSize={0.5}
              color="#ffffff"
              anchorX="center"
              anchorY="middle"
              outlineWidth={0.02}
              outlineColor={gemColor}
            >
              {label}
            </Text>
          </group>
        </Billboard>
        
        {/* Hover tooltip */}
        {hovered && description && (
          <Html distanceFactor={12} position={[0, 2.5, 0]} center>
            <motion.div
              initial={{ opacity: 0, y: 5 }}
              animate={{ opacity: 1, y: 0 }}
              className="bg-black/90 backdrop-blur-md border border-white/10 text-white px-4 py-2 rounded-lg text-sm whitespace-nowrap pointer-events-none"
            >
              <div className="font-medium mb-1" style={{ color: `#${gemColor.getHexString()}` }}>{label}</div>
              <div className="text-white/70">{description}</div>
            </motion.div>
          </Html>
        )}
      </group>
    </Float>
  );
}

// Energy beam connecting crystals with pulsing light
interface CrystalBeamProps {
  start: [number, number, number];
  end: [number, number, number];
  color: THREE.Color;
  intensity?: number;
}

function CrystalBeam({ start, end, color, intensity = 1 }: CrystalBeamProps) {
  const points = useMemo(() => [
    new THREE.Vector3(...start),
    new THREE.Vector3(...end)
  ], [start, end]);

  const midpoint = useMemo(() => (start[0] + end[0]) / 2, [start, end]);
  
  const progress = useRef(Math.random());
  
  useFrame((state, delta) => {
    progress.current += delta * 0.5;
    if (progress.current > 1) progress.current = 0;
  });

  return (
    <group>
      {/* Base line */}
      <Line
        points={points}
        color={color}
        lineWidth={1}
        transparent
        opacity={0.2 * intensity}
      />
      
      {/* Core beam */}
      <Line
        points={points}
        color={color}
        lineWidth={3}
        transparent
        opacity={0.1 * intensity}
      />
    </group>
  );
}

// Floating crystal shards/particles
function CrystalShards({ count = 60 }: { count?: number }) {
  const points = useRef<THREE.Points>(null);
  
  const { positions, colors } = useMemo(() => {
    const positions = new Float32Array(count * 3);
    const colors = new Float32Array(count * 3);
    
    for (let i = 0; i < count; i++) {
      // Random positions in sphere
      const theta = Math.random() * Math.PI * 2;
      const phi = Math.acos(2 * Math.random() - 1);
      const r = 8 + Math.random() * 20;
      
      positions[i * 3] = r * Math.sin(phi) * Math.cos(theta);
      positions[i * 3 + 1] = r * Math.sin(phi) * Math.sin(theta);
      positions[i * 3 + 2] = r * Math.cos(phi);
      
      // Random gem color
      const choice = Math.random();
      let c;
      if (choice < 0.25) c = CRYSTAL.emerald;
      else if (choice < 0.5) c = CRYSTAL.sapphire;
      else if (choice < 0.75) c = CRYSTAL.amethyst;
      else c = CRYSTAL.topaz;
      
      colors[i * 3] = c.r;
      colors[i * 3 + 1] = c.g;
      colors[i * 3 + 2] = c.b;
    }
    
    return { positions, colors };
  }, [count]);

  useFrame(({ clock }) => {
    if (points.current) {
      points.current.rotation.y = clock.getElapsedTime() * 0.02;
      points.current.rotation.x = Math.sin(clock.getElapsedTime() * 0.03) * 0.05;
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
        size={0.12}
        vertexColors
        transparent
        opacity={0.7}
        sizeAttenuation
        blending={THREE.AdditiveBlending}
      />
    </points>
  );
}

// Main crystal graph scene
function CrystalScene() {
  const { camera } = useThree();
  const [selectedNode, setSelectedNode] = useState<string | null>(null);
  
  // Define crystal nodes with meaningful labels
  const nodes = [
    { 
      id: 'orchestrate', 
      position: [0, 0, 0] as [number, number, number], 
      color: CRYSTAL.diamond, 
      label: 'ORCHESTRATOR',
      description: 'Central workflow coordination',
      size: 1.6,
      pulseOffset: 0
    },
    { 
      id: 'transform', 
      position: [-7, 4, -2] as [number, number, number], 
      color: CRYSTAL.sapphire, 
      label: 'TRANSFORM',
      description: 'Data transformation & mapping',
      size: 1.3,
      pulseOffset: 1
    },
    { 
      id: 'compute', 
      position: [7, 4, 2] as [number, number, number], 
      color: CRYSTAL.emerald, 
      label: 'COMPUTE',
      description: 'Heavy computation & processing',
      size: 1.3,
      pulseOffset: 2
    },
    { 
      id: 'analyze', 
      position: [-6, -3, 4] as [number, number, number], 
      color: CRYSTAL.amethyst, 
      label: 'ANALYZE',
      description: 'AI-powered analysis & insights',
      size: 1.2,
      pulseOffset: 3
    },
    { 
      id: 'stream', 
      position: [6, -3, -4] as [number, number, number], 
      color: CRYSTAL.ruby, 
      label: 'STREAM',
      description: 'Real-time event streaming',
      size: 1.2,
      pulseOffset: 4
    },
    { 
      id: 'persist', 
      position: [0, 6, 2] as [number, number, number], 
      color: CRYSTAL.topaz, 
      label: 'PERSIST',
      description: 'State persistence & storage',
      size: 1.1,
      pulseOffset: 5
    },
    { 
      id: 'route', 
      position: [0, -6, -2] as [number, number, number], 
      color: CRYSTAL.sapphire, 
      label: 'ROUTE',
      description: 'Intelligent routing & switching',
      size: 1.1,
      pulseOffset: 6
    },
    { 
      id: 'secure', 
      position: [-4, 0, -6] as [number, number, number], 
      color: CRYSTAL.ruby, 
      label: 'SECURE',
      description: 'Authentication & encryption',
      size: 1.0,
      pulseOffset: 7
    },
  ];

  // Define crystal connections
  const connections = [
    { from: 'orchestrate', to: 'transform', color: CRYSTAL.sapphire },
    { from: 'orchestrate', to: 'compute', color: CRYSTAL.emerald },
    { from: 'orchestrate', to: 'analyze', color: CRYSTAL.amethyst },
    { from: 'orchestrate', to: 'stream', color: CRYSTAL.ruby },
    { from: 'orchestrate', to: 'route', color: CRYSTAL.sapphire },
    { from: 'orchestrate', to: 'secure', color: CRYSTAL.ruby },
    { from: 'persist', to: 'transform', color: CRYSTAL.topaz },
    { from: 'persist', to: 'compute', color: CRYSTAL.topaz },
    { from: 'analyze', to: 'route', color: CRYSTAL.amethyst },
    { from: 'stream', to: 'route', color: CRYSTAL.ruby },
  ];

  const nodeMap = new Map(nodes.map(n => [n.id, n]));

  return (
    <>
      {/* Ambient lighting */}
      <ambientLight intensity={0.2} />
      <directionalLight position={[10, 10, 5]} intensity={0.8} color="#ffffff" />
      <pointLight position={[0, 0, 0]} intensity={2} color={CRYSTAL.diamond} distance={30} />
      <pointLight position={[-10, 5, -5]} intensity={1} color={CRYSTAL.sapphire} distance={25} />
      <pointLight position={[10, 5, 5]} intensity={1} color={CRYSTAL.emerald} distance={25} />
      
      {/* Background */}
      <Stars radius={80} depth={30} count={2500} factor={3} saturation={0.3} fade speed={0.4} />
      
      {/* Crystal shards */}
      <CrystalShards count={80} />
      
      {/* Connection beams */}
      {connections.map((conn, i) => {
        const fromNode = nodeMap.get(conn.from);
        const toNode = nodeMap.get(conn.to);
        if (!fromNode || !toNode) return null;
        
        const isActive = selectedNode === conn.from || selectedNode === conn.to;
        
        return (
          <CrystalBeam
            key={i}
            start={fromNode.position}
            end={toNode.position}
            color={conn.color}
            intensity={isActive ? 1.5 : 1}
          />
        );
      })}
      
      {/* Crystal nodes */}
      {nodes.map((node) => (
        <CrystalGem
          key={node.id}
          position={node.position}
          color={node.color}
          label={node.label}
          description={node.description}
          size={node.size}
          pulseOffset={node.pulseOffset}
          onClick={() => setSelectedNode(selectedNode === node.id ? null : node.id)}
        />
      ))}
      
      {/* Controls */}
      <OrbitControls
        enablePan={true}
        enableZoom={true}
        enableRotate={true}
        autoRotate
        autoRotateSpeed={0.35}
        minDistance={10}
        maxDistance={50}
        target={[0, 0, 0]}
      />
    </>
  );
}

export function CrystalGraph() {
  return (
    <div className="w-full h-full min-h-[700px] relative">
      <Canvas
        camera={{ position: [0, 2, 22], fov: 50 }}
        gl={{ antialias: true, alpha: false, powerPreference: 'high-performance' }}
        dpr={[1, 1.5]}
        style={{ background: '#050510' }}
      >
        <color attach="background" args={['#050510']} />
        <fog attach="fog" args={['#050510', 30, 70]} />
        <CrystalScene />
      </Canvas>
      
      {/* Overlay */}
      <div className="absolute bottom-6 left-6 pointer-events-none">
        <div className="flex flex-col gap-1">
          <div className="text-xs font-mono text-white/40 tracking-widest uppercase">
            Quantum Crystal
          </div>
          <div className="text-xs text-emerald-400/60">
            8 Facets • 10 Connections
          </div>
        </div>
      </div>
      
      {/* Interaction hint */}
      <div className="absolute top-4 right-4 pointer-events-none">
        <div className="text-xs text-white/30 font-mono">
          Click nodes to highlight
        </div>
      </div>
    </div>
  );
}

export default CrystalGraph;