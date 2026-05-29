/**
 * 3D Radar constellation view — functions as nodes in force-directed space
 */
import type { GalleryFunction } from '@/api/composer';
import { Billboard, OrbitControls, Stars, Text } from '@react-three/drei';
import { Canvas, useFrame } from '@react-three/fiber';
import { useMemo, useRef } from 'react';
import * as THREE from 'three';
import { CATEGORY_META, RUNTIME_COLORS } from '../constants';

interface FunctionNode {
  id: string;
  function: GalleryFunction;
  position: [number, number, number];
  cluster: string;
}

function buildNodes(functions: GalleryFunction[]): FunctionNode[] {
  return functions.map((fn, index) => {
    const cluster = fn.category || 'default';
    const clusterInfo = CATEGORY_META[cluster] || CATEGORY_META.default;
    const angle = (index / Math.max(functions.length, 1)) * Math.PI * 2;
    const radius = 5 + (index % 5) * 3;
    const height = ((index % 3) - 1) * 4;
    const bias = 0.55;

    return {
      id: fn.id,
      function: fn,
      position: [
        clusterInfo.center[0] * bias + Math.cos(angle) * radius * (1 - bias),
        clusterInfo.center[1] * bias + height * (1 - bias),
        clusterInfo.center[2] * bias + Math.sin(angle) * radius * (1 - bias),
      ] as [number, number, number],
      cluster,
    };
  });
}

function Node3D({
  node,
  isSelected,
  isHovered,
  onClick,
  onHover,
}: {
  node: FunctionNode;
  isSelected: boolean;
  isHovered: boolean;
  onClick: () => void;
  onHover: (hovered: boolean) => void;
}) {
  const groupRef = useRef<THREE.Group>(null);
  const runtime = node.function.runtime || 'python';
  const colors = RUNTIME_COLORS[runtime] || RUNTIME_COLORS.python;
  const size = 0.5 + Math.min((node.function.trust_score || 0) / 100, 0.8);

  useFrame(({ clock }) => {
    if (!groupRef.current) return;
    const t = clock.getElapsedTime();
    groupRef.current.position.y = node.position[1] + Math.sin(t * 0.6 + node.position[0]) * 0.15;
    groupRef.current.rotation.y = t * (isSelected ? 0.6 : 0.15);
  });

  const highlighted = isSelected || isHovered;

  return (
    <group ref={groupRef} position={node.position}>
      <mesh
        onClick={(e) => {
          e.stopPropagation();
          onClick();
        }}
        onPointerOver={(e) => {
          e.stopPropagation();
          onHover(true);
          document.body.style.cursor = 'pointer';
        }}
        onPointerOut={() => {
          onHover(false);
          document.body.style.cursor = 'auto';
        }}
      >
        <cylinderGeometry args={[0.65, 0.65, 0.25, 6]} />
        <meshStandardMaterial
          color="#0a0a14"
          metalness={0.85}
          roughness={0.15}
          emissive={highlighted ? colors.glow : colors.primary}
          emissiveIntensity={isSelected ? 1 : isHovered ? 0.6 : 0.35}
        />
      </mesh>
      <mesh scale={0.2}>
        <sphereGeometry args={[1, 12, 12]} />
        <meshBasicMaterial color={colors.glow} />
      </mesh>
      <Billboard position={[0, -1, 0]}>
        <Text
          fontSize={0.28}
          color="#fff"
          anchorX="center"
          outlineWidth={0.02}
          outlineColor={colors.primary}
        >
          {runtime.slice(0, 4).toUpperCase()}
        </Text>
      </Billboard>
      {isSelected && (
        <mesh scale={[size * 2.2, size * 2.2, 1]}>
          <ringGeometry args={[0.9, 1.1, 32]} />
          <meshBasicMaterial
            color={colors.glow}
            transparent
            opacity={0.5}
            side={THREE.DoubleSide}
          />
        </mesh>
      )}
    </group>
  );
}

function Scene({
  nodes,
  selectedId,
  hoveredId,
  onSelect,
  onHover,
}: {
  nodes: FunctionNode[];
  selectedId: string | null;
  hoveredId: string | null;
  onSelect: (id: string) => void;
  onHover: (nodeId: string | null) => void;
}) {
  return (
    <>
      <ambientLight intensity={0.2} />
      <pointLight position={[0, 0, 0]} intensity={1.2} color="#00d4ff" distance={60} />
      <directionalLight position={[10, 10, 5]} intensity={0.4} />
      <Stars radius={80} depth={40} count={1500} factor={2.5} saturation={0.2} fade speed={0.2} />
      {nodes.map((node) => (
        <Node3D
          key={node.id}
          node={node}
          isSelected={selectedId === node.id}
          isHovered={hoveredId === node.id}
          onClick={() => onSelect(node.id)}
          onHover={(hovered) => onHover(hovered ? node.id : null)}
        />
      ))}
      <OrbitControls
        autoRotate
        autoRotateSpeed={0.15}
        minDistance={12}
        maxDistance={70}
        enablePan
      />
    </>
  );
}

interface RadarViewProps {
  functions: GalleryFunction[];
  selectedNodeId: string | null;
  hoveredNodeId: string | null;
  onSelect: (fn: GalleryFunction) => void;
  onNodeHover: (nodeId: string | null) => void;
}

export function RadarView({
  functions,
  selectedNodeId,
  hoveredNodeId,
  onSelect,
  onNodeHover,
}: RadarViewProps) {
  const nodes = useMemo(() => buildNodes(functions), [functions]);

  const handleSelect = (id: string) => {
    const fn = functions.find((f) => f.id === id);
    if (fn) onSelect(fn);
  };

  return (
    <div className="flyway-radar-container">
      <Canvas
        camera={{ position: [25, 12, 25], fov: 55 }}
        dpr={[1, 1.5]}
        style={{ background: '#04040a' }}
      >
        <color attach="background" args={['#04040a']} />
        <fog attach="fog" args={['#04040a', 35, 90]} />
        <Scene
          nodes={nodes}
          selectedId={selectedNodeId}
          hoveredId={hoveredNodeId}
          onSelect={handleSelect}
          onHover={onNodeHover}
        />
      </Canvas>
      <div className="flyway-radar-hud">
        RADAR MODE · {functions.length} blips · drag to orbit · click to inspect
      </div>
    </div>
  );
}
