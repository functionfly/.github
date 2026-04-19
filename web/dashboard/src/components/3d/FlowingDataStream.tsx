/**
 * Data Plasma - Premium Real-Time Data Flow Visualization
 * 
 * Cinematic visualization of data streams flowing through a network,
 * featuring plasma-like energy tubes and glowing data packets.
 */

import { useRef, useMemo, useState } from 'react';
import { Canvas, useFrame, useThree } from '@react-three/fiber';
import { 
  OrbitControls, 
  Trail,
  Tube,
  Line,
  Float,
  Text,
  Billboard,
  Stars,
} from '@react-three/drei';
import * as THREE from 'three';

// Brand palette - electric plasma theme
const PLASMA = {
  void: new THREE.Color('#020108'),
  core: new THREE.Color('#0f0f1a'),
  cyan: new THREE.Color('#00f5ff'),
  magenta: new THREE.Color('#ff00ff'),
  gold: new THREE.Color('#ffd700'),
  violet: new THREE.Color('#8b5cf6'),
  white: new THREE.Color('#ffffff'),
};

// Custom flowing tube with animated energy
interface PlasmaTubeProps {
  curve: THREE.CatmullRomCurve3;
  color: THREE.Color;
  radius?: number;
  speed?: number;
}

function PlasmaTube({ curve, color, radius = 0.15, speed = 1 }: PlasmaTubeProps) {
  const outerRef = useRef<THREE.Mesh>(null);
  const innerRef = useRef<THREE.Mesh>(null);
  
  const shaderData = useMemo(() => ({
    uniforms: {
      time: { value: 0 },
      color: { value: color },
      speed: { value: speed },
    },
    vertexShader: `
      varying vec2 vUv;
      varying vec3 vPosition;
      
      void main() {
        vUv = uv;
        vPosition = position;
        gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);
      }
    `,
    fragmentShader: `
      uniform float time;
      uniform vec3 color;
      uniform float speed;
      varying vec2 vUv;
      varying vec3 vPosition;
      
      float plasma(vec2 uv, float time) {
        float v1 = sin(uv.x * 10.0 + time * 2.0);
        float v2 = sin(uv.y * 8.0 - time * 1.5);
        float v3 = sin(uv.x * 5.0 + uv.y * 5.0 + time);
        return (v1 + v2 + v3) / 3.0;
      }
      
      void main() {
        float p = plasma(vUv, time * speed);
        float intensity = 0.5 + p * 0.5;
        
        // Edge glow
        float edge = 1.0 - abs(vUv.y - 0.5) * 2.0;
        edge = pow(edge, 2.0);
        
        vec3 finalColor = color * intensity * (1.0 + edge * 0.5);
        float alpha = 0.6 + edge * 0.3;
        
        gl_FragColor = vec4(finalColor, alpha);
      }
    `,
  }), [color, speed]);

  const innerShader = useMemo(() => ({
    uniforms: {
      time: { value: 0 },
      color: { value: color },
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
        float flow = fract(vUv.x * 3.0 - time * 2.0);
        float pulse = smoothstep(0.0, 0.3, flow) * smoothstep(1.0, 0.7, flow);
        
        vec3 finalColor = color * (1.0 + pulse * 2.0);
        float alpha = 0.3 + pulse * 0.5;
        
        gl_FragColor = vec4(finalColor, alpha);
      }
    `,
  }), [color]);

  useFrame(({ clock }) => {
    const t = clock.getElapsedTime();
    if (outerRef.current) {
      const mat = outerRef.current.material as THREE.ShaderMaterial;
      mat.uniforms.time.value = t;
    }
    if (innerRef.current) {
      const mat = innerRef.current.material as THREE.ShaderMaterial;
      mat.uniforms.time.value = t;
    }
  });

  return (
    <group>
      {/* Outer glow tube */}
      <Tube args={[curve, 64, radius * 1.5, 12, false]}>
        <shaderMaterial
          ref={outerRef}
          args={[shaderData]}
          transparent
          side={THREE.DoubleSide}
          depthWrite={false}
        />
      </Tube>
      
      {/* Inner energy core */}
      <Tube args={[curve, 64, radius * 0.4, 8, false]}>
        <shaderMaterial
          ref={innerRef}
          args={[innerShader]}
          transparent
          side={THREE.DoubleSide}
          blending={THREE.AdditiveBlending}
          depthWrite={false}
        />
      </Tube>
    </group>
  );
}

// Glowing data packet traveling along curve
interface DataPlasmaProps {
  curve: THREE.CatmullRomCurve3;
  color: THREE.Color;
  progress?: number;
  size?: number;
}

function DataPlasma({ curve, color, progress = 0, size = 0.4 }: DataPlasmaProps) {
  const meshRef = useRef<THREE.Mesh>(null);
  const glowRef = useRef<THREE.Mesh>(null);
  const prog = useRef(progress);
  
  useFrame((state, delta) => {
    prog.current += delta * 0.15;
    if (prog.current > 1) prog.current = 0;
    
    if (meshRef.current) {
      const point = curve.getPoint(prog.current);
      const tangent = curve.getTangent(prog.current);
      meshRef.current.position.copy(point);
      meshRef.current.quaternion.setFromUnitVectors(
        new THREE.Vector3(0, 1, 0),
        tangent.normalize()
      );
    }
    if (glowRef.current && meshRef.current) {
      glowRef.current.position.copy(meshRef.current.position);
      const pulse = (Math.sin(state.clock.getElapsedTime() * 8) + 1) * 0.3 + 0.7;
      glowRef.current.scale.setScalar(pulse * 2);
    }
  });

  return (
    <>
      {/* Main packet */}
      <mesh ref={meshRef}>
        <capsuleGeometry args={[size, size * 2, 8, 16]} />
        <meshStandardMaterial
          color={color}
          emissive={color}
          emissiveIntensity={2}
          toneMapped={false}
        />
      </mesh>
      
      {/* Glow halo */}
      <mesh ref={glowRef}>
        <sphereGeometry args={[size * 1.5, 16, 16]} />
        <meshBasicMaterial
          color={color}
          transparent
          opacity={0.3}
          blending={THREE.AdditiveBlending}
          depthWrite={false}
        />
      </mesh>
      
      {/* Point light on packet */}
      <pointLight color={color} intensity={3} distance={8} />
    </>
  );
}

// Node in the data flow - compact hexagonal design
interface FlowNodeProps {
  position: [number, number, number];
  color: THREE.Color;
  label: string;
  status?: 'active' | 'processing' | 'idle';
  size?: number;
}

function FlowNode({ position, color, label, status = 'active', size = 1 }: FlowNodeProps) {
  const groupRef = useRef<THREE.Group>(null);
  const [hovered, setHovered] = useState(false);
  
  const statusColor = status === 'active' ? PLASMA.cyan : 
                      status === 'processing' ? PLASMA.magenta : 
                      PLASMA.violet;
  
  useFrame(({ clock }) => {
    if (groupRef.current) {
      groupRef.current.rotation.y = Math.sin(clock.getElapsedTime() * 0.3) * 0.05;
      
      // Subtle hover response
      if (hovered) {
        groupRef.current.scale.lerp(new THREE.Vector3(size * 1.1, size * 1.1, size * 1.1), 0.1);
      } else {
        groupRef.current.scale.lerp(new THREE.Vector3(size, size, size), 0.1);
      }
    }
  });

  return (
    <group 
      ref={groupRef} 
      position={position}
      onPointerOver={() => setHovered(true)}
      onPointerOut={() => setHovered(false)}
    >
      {/* Outer glow */}
      <mesh scale={1.6}>
        <sphereGeometry args={[1, 16, 16]} />
        <meshBasicMaterial
          color={color}
          transparent
          opacity={hovered ? 0.25 : 0.12}
          side={THREE.BackSide}
        />
      </mesh>
      
      {/* Main body */}
      <mesh>
        <cylinderGeometry args={[0.8, 0.8, 0.4, 6]} />
        <meshStandardMaterial
          color={PLASMA.core}
          metalness={0.85}
          roughness={0.15}
          emissive={color}
          emissiveIntensity={hovered ? 0.5 : 0.25}
        />
      </mesh>
      
      {/* Status indicator */}
      <mesh position={[0, 0.35, 0.5]}>
        <sphereGeometry args={[0.15, 16, 16]} />
        <meshBasicMaterial color={statusColor} />
      </mesh>
      
      {/* Label */}
      <Billboard position={[0, -1.2, 0]} follow lockZ={false}>
        <Text
          fontSize={0.35}
          color="#ffffff"
          anchorX="center"
          anchorY="middle"
          outlineWidth={0.02}
          outlineColor={color}
        >
          {label}
        </Text>
      </Billboard>
    </group>
  );
}

// Ambient particles
function PlasmaDust({ count = 150 }: { count?: number }) {
  const points = useRef<THREE.Points>(null);
  
  const positions = useMemo(() => {
    const pos = new Float32Array(count * 3);
    for (let i = 0; i < count; i++) {
      pos[i * 3] = (Math.random() - 0.5) * 50;
      pos[i * 3 + 1] = (Math.random() - 0.5) * 30;
      pos[i * 3 + 2] = (Math.random() - 0.5) * 50;
    }
    return pos;
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
      </bufferGeometry>
      <pointsMaterial
        size={0.08}
        color={PLASMA.cyan}
        transparent
        opacity={0.5}
        blending={THREE.AdditiveBlending}
        sizeAttenuation
      />
    </points>
  );
}

// Main data flow scene
function DataFlowScene() {
  const { camera } = useThree();
  
  // Curved paths for data flow
  const curves = useMemo(() => [
    // Primary flow path
    new THREE.CatmullRomCurve3([
      new THREE.Vector3(-15, 2, 0),
      new THREE.Vector3(-10, 3, 5),
      new THREE.Vector3(-5, 0, -3),
      new THREE.Vector3(0, -2, 2),
      new THREE.Vector3(5, 2, -2),
      new THREE.Vector3(10, 0, 3),
      new THREE.Vector3(15, -1, 0),
    ], false, 'catmullrom', 0.5),
    
    // Secondary path
    new THREE.CatmullRomCurve3([
      new THREE.Vector3(-15, -3, 3),
      new THREE.Vector3(-8, -2, 0),
      new THREE.Vector3(-2, -4, 4),
      new THREE.Vector3(4, -3, -2),
      new THREE.Vector3(12, -2, 2),
      new THREE.Vector3(15, -4, -1),
    ], false, 'catmullrom', 0.5),
    
    // Tertiary path  
    new THREE.CatmullRomCurve3([
      new THREE.Vector3(-15, 5, -3),
      new THREE.Vector3(-9, 4, 2),
      new THREE.Vector3(-3, 6, -1),
      new THREE.Vector3(3, 3, 3),
      new THREE.Vector3(9, 5, -3),
      new THREE.Vector3(15, 4, 1),
    ], false, 'catmullrom', 0.5),
  ], []);

  // Node positions
  const nodes = [
    { position: [-15, 0, 0] as [number, number, number], color: PLASMA.cyan, label: 'API INPUT', status: 'active' as const },
    { position: [0, 0, 0] as [number, number, number], color: PLASMA.magenta, label: 'PROCESS HUB', status: 'processing' as const },
    { position: [15, 0, 0] as [number, number, number], color: PLASMA.gold, label: 'OUTPUT', status: 'active' as const },
  ];

  return (
    <>
      {/* Lighting */}
      <ambientLight intensity={0.15} />
      <directionalLight position={[10, 10, 5]} intensity={0.5} color="#ffffff" />
      <pointLight position={[-10, 5, 0]} intensity={1.5} color={PLASMA.cyan} distance={30} />
      <pointLight position={[10, 5, 0]} intensity={1.5} color={PLASMA.gold} distance={30} />
      <pointLight position={[0, 0, 0]} intensity={2} color={PLASMA.magenta} distance={25} />
      
      {/* Background */}
      <Stars radius={100} depth={50} count={2000} factor={3} saturation={0.2} fade speed={0.5} />
      
      {/* Flowing tubes */}
      {curves.map((curve, i) => (
        <PlasmaTube
          key={i}
          curve={curve}
          color={i === 0 ? PLASMA.cyan : i === 1 ? PLASMA.magenta : PLASMA.violet}
          radius={0.12}
          speed={0.8 + i * 0.2}
        />
      ))}
      
      {/* Data packets */}
      {[0, 0.2, 0.4, 0.6, 0.8].map((offset, i) => (
        <DataPlasma key={i} curve={curves[0]} color={PLASMA.cyan} progress={offset} size={0.35} />
      ))}
      {[0.1, 0.4, 0.7].map((offset, i) => (
        <DataPlasma key={`b-${i}`} curve={curves[1]} color={PLASMA.magenta} progress={offset} size={0.3} />
      ))}
      {[0.3, 0.6, 0.9].map((offset, i) => (
        <DataPlasma key={`c-${i}`} curve={curves[2]} color={PLASMA.violet} progress={offset} size={0.3} />
      ))}
      
      {/* Nodes */}
      {nodes.map((node, i) => (
        <FlowNode
          key={i}
          position={node.position}
          color={node.color}
          label={node.label}
          status={node.status}
          size={1.2}
        />
      ))}
      
      {/* Ambient dust */}
      <PlasmaDust count={100} />
      
      {/* Controls */}
      <OrbitControls
        enablePan={false}
        enableZoom={true}
        enableRotate={true}
        autoRotate
        autoRotateSpeed={0.25}
        minDistance={15}
        maxDistance={50}
      />
    </>
  );
}

export function FlowingDataStream() {
  return (
    <div className="w-full h-full min-h-[600px] relative">
      <Canvas
        camera={{ position: [25, 12, 20], fov: 55 }}
        gl={{ antialias: true, alpha: false, powerPreference: 'high-performance' }}
        dpr={[1, 1.5]}
        style={{ background: '#020108' }}
      >
        <color attach="background" args={['#020108']} />
        <fog attach="fog" args={['#020108', 35, 80]} />
        <DataFlowScene />
      </Canvas>
      
      {/* Overlay */}
      <div className="absolute bottom-6 left-6 pointer-events-none">
        <div className="flex flex-col gap-1">
          <div className="text-xs font-mono text-white/40 tracking-widest uppercase">
            Data Plasma
          </div>
          <div className="text-xs text-magenta-400/60">
            3 Streams • 12.4K packets/sec
          </div>
        </div>
      </div>
    </div>
  );
}

export default FlowingDataStream;