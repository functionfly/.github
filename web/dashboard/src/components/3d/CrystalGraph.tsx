/**
 * Crystal Graph 3D Visualization
 * Geometric, crystalline structure representing interconnected functions
 */

import { useRef, useMemo, useState } from 'react';
import { Canvas, useFrame } from '@react-three/fiber';
import { 
  OrbitControls, 
  Stars,
  Float,
  Text,
  Html,
  Billboard,
  MeshDistortMaterial,
  Line,
} from '@react-three/drei';
import * as THREE from 'three';
import { motion } from 'framer-motion';

// Crystal node with geometric shape
interface CrystalNodeProps {
  position: [number, number, number];
  color: string;
  size?: number;
  label?: string;
  description?: string;
  onClick?: () => void;
}

function CrystalNode({ position, color, size = 1.5, label, description, onClick }: CrystalNodeProps) {
  const meshRef = useRef<THREE.Mesh>(null);
  const glowRef = useRef<THREE.Mesh>(null);
  const [hovered, setHovered] = useState(false);
  
  useFrame(({ clock }) => {
    if (meshRef.current) {
      meshRef.current.rotation.x = Math.sin(clock.getElapsedTime() * 0.3) * 0.1;
      meshRef.current.rotation.y = clock.getElapsedTime() * 0.2;
      meshRef.current.scale.setScalar(hovered ? size * 1.2 : size);
    }
    if (glowRef.current) {
      glowRef.current.rotation.x = meshRef.current?.rotation.x || 0;
      glowRef.current.rotation.y = meshRef.current?.rotation.y || 0;
      const pulse = Math.sin(clock.getElapsedTime() * 2) * 0.1 + 0.9;
      (glowRef.current.material as THREE.Material).opacity = hovered ? pulse * 0.5 : pulse * 0.2;
    }
  });

  // Icosahedron geometry for crystal look
  const geometry = useMemo(() => new THREE.IcosahedronGeometry(1, 0), []);

  return (
    <Float rotationIntensity={0.5} floatIntensity={0.5} speed={2}>
      <group 
        position={position}
        onPointerOver={() => setHovered(true)}
        onPointerOut={() => setHovered(false)}
        onClick={onClick}
      >
        {/* Crystal core */}
        <mesh ref={meshRef} geometry={geometry}>
          <MeshDistortMaterial
            color={color}
            roughness={0.1}
            metalness={0.8}
            emissive={color}
            emissiveIntensity={hovered ? 0.8 : 0.4}
            distort={0.3}
            speed={2}
          />
        </mesh>
        
        {/* Outer glow shell */}
        <mesh ref={glowRef} geometry={geometry} scale={1.8}>
          <meshBasicMaterial
            color={color}
            transparent
            opacity={0.2}
            wireframe
          />
        </mesh>
        
        {/* Floating label */}
        <Billboard>
          <Text
            position={[0, -2.5, 0]}
            fontSize={0.6}
            color="white"
            anchorX="center"
            outlineWidth={0.02}
            outlineColor="black"
          >
            {label}
          </Text>
        </Billboard>
        
        {/* Connection points */}
        {[0, 1, 2, 3].map((i) => (
          <mesh 
            key={i}
            position={[
              Math.cos((i / 4) * Math.PI * 2) * 1.3,
              Math.sin((i / 4) * Math.PI * 2) * 0.3,
              Math.sin((i / 4) * Math.PI * 2) * 1.3
            ]}
          >
            <sphereGeometry args={[0.1, 8, 8]} />
            <meshStandardMaterial 
              color="white" 
              emissive="white" 
              emissiveIntensity={hovered ? 3 : 1}
            />
          </mesh>
        ))}
        
        {/* Info tooltip on hover */}
        {hovered && description && (
          <Html distanceFactor={10}>
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              className="bg-black/80 backdrop-blur-sm text-white px-4 py-2 rounded-lg text-sm whitespace-nowrap pointer-events-none"
            >
              {description}
            </motion.div>
          </Html>
        )}
      </group>
    </Float>
  );
}

// Energy beam connecting crystals - using Line from drei
interface EnergyBeamProps {
  start: [number, number, number];
  end: [number, number, number];
  color: string;
  intensity?: number;
}

function EnergyBeam({ start, end, color, intensity = 1 }: EnergyBeamProps) {
  const points = useMemo(() => [new THREE.Vector3(...start), new THREE.Vector3(...end)], [start, end]);

  return (
    <>
      <Line
        points={points}
        color={color}
        lineWidth={2}
        transparent
        opacity={intensity}
      />
      <Line
        points={points}
        color={color}
        lineWidth={4}
        transparent
        opacity={intensity * 0.3}
      />
    </>
  );
}

// Floating data particles
function FloatingParticles({ count = 50 }: { count?: number }) {
  const points = useRef<THREE.Points>(null);
  
  const positions = useMemo(() => {
    const pos = new Float32Array(count * 3);
    const colors = new Float32Array(count * 3);
    
    for (let i = 0; i < count; i++) {
      pos[i * 3] = (Math.random() - 0.5) * 40;
      pos[i * 3 + 1] = (Math.random() - 0.5) * 40;
      pos[i * 3 + 2] = (Math.random() - 0.5) * 40;
      
      // Random colors: blue, purple, green
      const colorChoice = Math.random();
      if (colorChoice < 0.33) {
        colors[i * 3] = 0.23; colors[i * 3 + 1] = 0.51; colors[i * 3 + 2] = 0.95; // Blue
      } else if (colorChoice < 0.66) {
        colors[i * 3] = 0.55; colors[i * 3 + 1] = 0.36; colors[i * 3 + 2] = 0.96; // Purple
      } else {
        colors[i * 3] = 0.06; colors[i * 3 + 1] = 0.65; colors[i * 3 + 2] = 0.51; // Green
      }
    }
    return { pos, colors };
  }, [count]);

  useFrame(({ clock }) => {
    if (points.current) {
      points.current.rotation.y = clock.getElapsedTime() * 0.05;
      points.current.rotation.x = Math.sin(clock.getElapsedTime() * 0.1) * 0.1;
    }
  });

  return (
    <points ref={points}>
      <bufferGeometry>
        <bufferAttribute
          attach="attributes-position"
          count={count}
          array={positions.pos}
          itemSize={3}
        />
        <bufferAttribute
          attach="attributes-color"
          count={count}
          array={positions.colors}
          itemSize={3}
        />
      </bufferGeometry>
      <pointsMaterial
        size={0.15}
        vertexColors
        transparent
        opacity={0.8}
        sizeAttenuation
      />
    </points>
  );
}

// Main crystal graph scene
function CrystalScene() {
  const [selectedNode, setSelectedNode] = useState<string | null>(null);
  
  const nodes = [
    { 
      id: 'orchestrate', 
      position: [0, 0, 0] as [number, number, number], 
      color: '#8b5cf6', 
      label: 'Orchestrate',
      description: 'Central workflow orchestration hub'
    },
    { 
      id: 'transform', 
      position: [-6, 3, 2] as [number, number, number], 
      color: '#3b82f6', 
      label: 'Transform',
      description: 'Data transformation & mapping'
    },
    { 
      id: 'compute', 
      position: [6, 3, -2] as [number, number, number], 
      color: '#10b981', 
      label: 'Compute',
      description: 'Heavy computation & processing'
    },
    { 
      id: 'analyze', 
      position: [-6, -3, -2] as [number, number, number], 
      color: '#f59e0b', 
      label: 'Analyze',
      description: 'AI-powered analysis & insights'
    },
    { 
      id: 'stream', 
      position: [6, -3, 2] as [number, number, number], 
      color: '#ec4899', 
      label: 'Stream',
      description: 'Real-time streaming & events'
    },
    { 
      id: 'persist', 
      position: [0, 5, 0] as [number, number, number], 
      color: '#6366f1', 
      label: 'Persist',
      description: 'State persistence & storage'
    },
    { 
      id: 'route', 
      position: [0, -5, 0] as [number, number, number], 
      color: '#06b6d4', 
      label: 'Route',
      description: 'Intelligent routing & switching'
    },
  ];

  const connections = [
    { from: 'orchestrate', to: 'transform', color: '#8b5cf6' },
    { from: 'orchestrate', to: 'compute', color: '#8b5cf6' },
    { from: 'orchestrate', to: 'analyze', color: '#8b5cf6' },
    { from: 'orchestrate', to: 'stream', color: '#8b5cf6' },
    { from: 'persist', to: 'transform', color: '#6366f1' },
    { from: 'persist', to: 'compute', color: '#6366f1' },
    { from: 'stream', to: 'route', color: '#ec4899' },
    { from: 'analyze', to: 'route', color: '#f59e0b' },
  ];

  const nodeMap = new Map(nodes.map(n => [n.id, n.position]));

  return (
    <>
      <ambientLight intensity={0.4} />
      <directionalLight position={[10, 10, 5]} intensity={1.5} />
      <pointLight position={[0, 0, 0]} intensity={2} color="#8b5cf6" />
      
      <Stars radius={80} depth={30} count={3000} factor={4} saturation={0.5} fade speed={0.5} />
      
      {/* Crystal nodes */}
      {nodes.map((node) => (
        <CrystalNode
          key={node.id}
          position={node.position}
          color={node.color}
          size={node.id === 'orchestrate' ? 2 : 1.5}
          label={node.label}
          description={node.description}
          onClick={() => setSelectedNode(node.id)}
        />
      ))}
      
      {/* Energy beams */}
      {connections.map((conn, i) => (
        <EnergyBeam
          key={i}
          start={nodeMap.get(conn.from) as [number, number, number]}
          end={nodeMap.get(conn.to) as [number, number, number]}
          color={conn.color}
          intensity={selectedNode && (conn.from === selectedNode || conn.to === selectedNode) ? 1 : 0.5}
        />
      ))}
      
      <FloatingParticles count={100} />
      
      <OrbitControls
        enablePan={true}
        enableZoom={true}
        enableRotate={true}
        autoRotate
        autoRotateSpeed={0.4}
        minDistance={10}
        maxDistance={50}
      />
    </>
  );
}

export function CrystalGraph() {
  return (
    <div className="w-full h-full min-h-[700px]">
      <Canvas
        camera={{ position: [0, 0, 25], fov: 50 }}
        gl={{ antialias: true, alpha: true }}
        dpr={[1, 2]}
      >
        <CrystalScene />
      </Canvas>
    </div>
  );
}

export default CrystalGraph;
