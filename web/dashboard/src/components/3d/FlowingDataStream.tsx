/**
 * Flowing Data Stream 3D Visualization
 * Shows animated data flowing through tubes/pipes
 */

import { useRef, useMemo } from 'react';
import { Canvas, useFrame } from '@react-three/fiber';
import { 
  OrbitControls, 
  Trail,
  Tube,
  Line,
  useTexture,
  Float,
  Text,
  Html,
  Stars,
} from '@react-three/drei';
import * as THREE from 'three';

// Curved data flow path
interface DataStreamProps {
  curve: THREE.CatmullRomCurve3;
  color: string;
  speed?: number;
  thickness?: number;
}

function DataStream({ curve, color, speed = 1, thickness = 0.3 }: DataStreamProps) {
  const tubeRef = useRef<THREE.Mesh>(null);
  const flowRef = useRef<THREE.ShaderMaterial>(null);
  
  // Custom shader for flowing effect
  const shaderData = useMemo(() => ({
    uniforms: {
      time: { value: 0 },
      color: { value: new THREE.Color(color) },
    },
    vertexShader: `
      varying vec2 vUv;
      void main() {
        vUv = uv;
        gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);
      }
    `,
    fragmentShader: `
      uniform float time;
      uniform vec3 color;
      varying vec2 vUv;
      
      void main() {
        float flow = mod(vUv.x - time * 0.5, 1.0);
        float intensity = smoothstep(0.0, 0.3, flow) * smoothstep(1.0, 0.7, flow);
        gl_FragColor = vec4(color * (0.3 + intensity * 0.7), 0.8);
      }
    `,
  }), [color]);

  useFrame(({ clock }) => {
    if (flowRef.current) {
      flowRef.current.uniforms.time.value = clock.getElapsedTime() * speed;
    }
  });

  return (
    <Tube args={[curve, 100, thickness, 8, false]}>
      <shaderMaterial
        ref={flowRef}
        args={[shaderData]}
        transparent
        side={THREE.DoubleSide}
      />
    </Tube>
  );
}

// Glowing data packet moving along path
interface DataPacketProps {
  curve: THREE.CatmullRomCurve3;
  color: string;
  speed?: number;
}

function MovingPacket({ curve, color, speed = 0.5 }: DataPacketProps) {
  const meshRef = useRef<THREE.Mesh>(null);
  const progress = useRef(0);
  
  useFrame((state, delta) => {
    progress.current += delta * speed;
    if (progress.current > 1) progress.current = 0;
    
    if (meshRef.current) {
      const point = curve.getPoint(progress.current);
      const tangent = curve.getTangent(progress.current);
      meshRef.current.position.copy(point);
      meshRef.current.lookAt(point.clone().add(tangent));
    }
  });

  return (
    <mesh ref={meshRef}>
      <sphereGeometry args={[0.4, 32, 32]} />
      <meshStandardMaterial
        color={color}
        emissive={color}
        emissiveIntensity={3}
        toneMapped={false}
      />
      <pointLight color={color} intensity={5} distance={5} />
    </mesh>
  );
}

// Node in the data flow
interface FlowNodeProps {
  position: [number, number, number];
  color: string;
  label: string;
  icon?: string;
}

function FlowNode({ position, color, label }: FlowNodeProps) {
  const groupRef = useRef<THREE.Group>(null);
  
  useFrame(({ clock }) => {
    if (groupRef.current) {
      groupRef.current.rotation.y = Math.sin(clock.getElapsedTime() * 0.5) * 0.1;
    }
  });

  return (
    <group ref={groupRef} position={position}>
      {/* Main node body */}
      <mesh>
        <boxGeometry args={[2, 1.5, 0.5]} />
        <meshStandardMaterial
          color={color}
          emissive={color}
          emissiveIntensity={0.3}
          roughness={0.1}
          metalness={0.9}
        />
      </mesh>
      
      {/* Glowing border */}
      <mesh scale={1.05}>
        <boxGeometry args={[2, 1.5, 0.5]} />
        <meshBasicMaterial
          color={color}
          transparent
          opacity={0.3}
          wireframe
        />
      </mesh>
      
      {/* Label */}
      <Float>
        <Text
          position={[0, -1.5, 0]}
          fontSize={0.5}
          color="white"
          anchorX="center"
          outlineWidth={0.02}
          outlineColor="black"
        >
          {label}
        </Text>
      </Float>
      
      {/* Status indicator */}
      <mesh position={[0.8, 0.5, 0.3]}>
        <sphereGeometry args={[0.15, 16, 16]} />
        <meshStandardMaterial
          color="#22c55e"
          emissive="#22c55e"
          emissiveIntensity={2}
        />
      </mesh>
    </group>
  );
}

// Main data flow scene
function DataFlowScene() {
  // Create flowing curves
  const curves = useMemo(() => [
    new THREE.CatmullRomCurve3([
      new THREE.Vector3(-10, 0, 0),
      new THREE.Vector3(-5, 2, 3),
      new THREE.Vector3(0, 0, -2),
      new THREE.Vector3(5, -2, 1),
      new THREE.Vector3(10, 0, 0),
    ]),
    new THREE.CatmullRomCurve3([
      new THREE.Vector3(-10, 3, 2),
      new THREE.Vector3(-3, 5, 0),
      new THREE.Vector3(2, 4, -3),
      new THREE.Vector3(7, 3, 0),
      new THREE.Vector3(10, 3, -2),
    ]),
    new THREE.CatmullRomCurve3([
      new THREE.Vector3(-10, -3, -2),
      new THREE.Vector3(-4, -4, 1),
      new THREE.Vector3(1, -5, -1),
      new THREE.Vector3(6, -3, 2),
      new THREE.Vector3(10, -3, 0),
    ]),
  ], []);

  const nodes = [
    { position: [-12, 0, 0], color: '#3b82f6', label: 'API Input' },
    { position: [-12, 3, 2], color: '#8b5cf6', label: 'Webhook' },
    { position: [-12, -3, -2], color: '#f59e0b', label: 'Stream' },
    { position: [0, 0, 0], color: '#10b981', label: 'Process Hub' },
    { position: [12, 0, 0], color: '#ec4899', label: 'Output' },
    { position: [12, 3, -2], color: '#6366f1', label: 'Database' },
  ];

  return (
    <>
      <ambientLight intensity={0.3} />
      <directionalLight position={[10, 10, 5]} intensity={1.5} />
      <pointLight position={[-10, 0, 5]} intensity={1} color="#3b82f6" />
      <pointLight position={[0, 0, 5]} intensity={1} color="#10b981" />
      <pointLight position={[10, 0, 5]} intensity={1} color="#ec4899" />
      
      <Stars radius={100} depth={50} count={2000} factor={4} saturation={0} fade speed={1} />
      
      {/* Flowing streams */}
      {curves.map((curve, i) => (
        <group key={i}>
          <DataStream
            curve={curve}
            color={['#3b82f6', '#8b5cf6', '#10b981'][i]}
            speed={1 + i * 0.3}
            thickness={0.2 + i * 0.1}
          />
          {/* Multiple packets per stream */}
          {[0, 0.33, 0.66].map((offset, j) => (
            <MovingPacket
              key={j}
              curve={curve}
              color={['#60a5fa', '#a78bfa', '#34d399'][i]}
              speed={0.3}
            />
          ))}
        </group>
      ))}
      
      {/* Nodes */}
      {nodes.map((node, i) => (
        <FlowNode
          key={i}
          position={node.position as [number, number, number]}
          color={node.color}
          label={node.label}
        />
      ))}
      
      <OrbitControls
        enablePan={false}
        enableZoom={true}
        enableRotate={true}
        autoRotate
        autoRotateSpeed={0.3}
        minDistance={15}
        maxDistance={40}
      />
    </>
  );
}

export function FlowingDataStream() {
  return (
    <div className="w-full h-full min-h-[600px]">
      <Canvas
        camera={{ position: [20, 10, 20], fov: 60 }}
        gl={{ antialias: true, alpha: true }}
        dpr={[1, 2]}
      >
        <DataFlowScene />
      </Canvas>
    </div>
  );
}

export default FlowingDataStream;
