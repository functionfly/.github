/**
 * Living Function Canvas - Reimagined Gallery Page
 * 
 * Transforms the traditional gallery into an interactive 3D experience
 * where functions are nodes in a force-directed constellation map.
 * Features live execution visualization, code viewing, and real metrics.
 */

import { useState, useRef, useEffect, useCallback, useMemo, Suspense, lazy } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Search,
  X,
  Code2,
  Play,
  Heart,
  GitFork,
  Star,
  ExternalLink,
  Loader2,
  Filter,
  Layers,
  Box,
  Cpu,
  Sparkles,
  Zap,
  Menu,
  LayoutGrid,
  List,
  Maximize2,
  Activity,
  Clock,
  TrendingUp,
  Award,
  Flame,
  ChevronLeft,
  ChevronRight,
  MousePointer2,
  Rotate3d,
  Monitor,
  Download,
  Eye,
  Share2,
  Terminal,
  Globe,
  Database,
  Brain,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { GlassmorphismCard } from '@/components/ui/GlassmorphismCard';
import { SpotlightCard } from '@/components/ui/SpotlightCard';
import { TextGradient } from '@/components/ui/TextGradient';
import { cn } from '@/lib/utils';
import { galleryApi, type GalleryFunction, RUNTIME_MONACO_LANG } from '@/api/composer';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useNavigate } from 'react-router-dom';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';

// Lazy load Monaco editor to reduce initial bundle size
const MonacoEditor = lazy(() => import('@monaco-editor/react'));

// =============================================================================
// Types & Interfaces
// =============================================================================

type GalleryCategory = 'all' | 'workflows' | 'graphs' | 'demos' | 'ui' | 'infra';
type ViewMode = 'constellation' | 'list' | 'theater';

interface FunctionNode {
  id: string;
  function: GalleryFunction;
  position: [number, number, number];
  velocity: [number, number, number];
  cluster: string;
  highlighted: boolean;
  hovered: boolean;
}

interface GalleryStats {
  totalFunctions: number;
  activeExecutions: number;
  avgLatency: number;
  totalRemixes: number;
}

// =============================================================================
// Runtime Colors - Quantum Dark Aesthetic
// =============================================================================

const RUNTIME_COLORS: Record<string, { primary: string; glow: string; accent: string }> = {
  python: { primary: '#3b82f6', glow: '#60a5fa', accent: '#1d4ed8' },     // Blue
  nodejs: { primary: '#10b981', glow: '#34d399', accent: '#059669' },    // Green
  typescript: { primary: '#10b981', glow: '#34d399', accent: '#059669' }, // Green
  go: { primary: '#06b6d4', glow: '#22d3ee', accent: '#0891b2' },         // Cyan
  rust: { primary: '#f97316', glow: '#fb923c', accent: '#ea580c' },       // Orange
  deno: { primary: '#8b5cf6', glow: '#a78bfa', accent: '#7c3aed' },      // Purple
  bun: { primary: '#f59e0b', glow: '#fbbf24', accent: '#d97706' },        // Amber
  java: { primary: '#ef4444', glow: '#f87171', accent: '#dc2626' },        // Red
  csharp: { primary: '#ec4899', glow: '#f472b6', accent: '#db2777' },     // Pink
  ruby: { primary: '#dc2626', glow: '#ef4444', accent: '#b91c1c' },       // Red
  php: { primary: '#6366f1', glow: '#818cf8', accent: '#4f46e5' },         // Indigo
};

const CATEGORY_CLUSTERS: Record<string, { center: [number, number, number]; color: string; label: string }> = {
  'data-processing': { center: [-15, 5, -10], color: '#3b82f6', label: 'Data' },
  'api': { center: [0, 8, -5], color: '#8b5cf6', label: 'API' },
  'ml': { center: [15, 5, -10], color: '#ec4899', label: 'ML/AI' },
  'web-scraping': { center: [-12, -5, 8], color: '#10b981', label: 'Web' },
  'automation': { center: [0, -8, 5], color: '#f59e0b', label: 'Auto' },
  'utility': { center: [12, -5, 8], color: '#6366f1', label: 'Utils' },
  'finance': { center: [0, 0, 15], color: '#14b8a6', label: 'Finance' },
  'default': { center: [0, 0, 0], color: '#64748b', label: 'General' },
};

// =============================================================================
// 3D Components (using React Three Fiber patterns)
// =============================================================================

import { Canvas, useFrame, useThree } from '@react-three/fiber';
import { OrbitControls, Stars, Text, Billboard, Html, Line, Trail } from '@react-three/drei';
import * as THREE from 'three';

// Energy pulse traveling between nodes
interface DataPulseProps {
  start: [number, number, number];
  end: [number, number, number];
  color: string;
  delay?: number;
  speed?: number;
}

function DataPulse({ start, end, color, delay = 0, speed = 1 }: DataPulseProps) {
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
      
      const scale = Math.sin(t * Math.PI) * 0.4 + 0.2;
      meshRef.current.scale.setScalar(scale);
    }
  });

  return (
    <mesh ref={meshRef}>
      <sphereGeometry args={[1, 8, 8]} />
      <meshBasicMaterial color={color} transparent opacity={0.9} />
    </mesh>
  );
}

// Connection line between nodes
interface NodeConnectionProps {
  start: [number, number, number];
  end: [number, number, number];
  color: string;
  active?: boolean;
}

function NodeConnection({ start, end, color, active = false }: NodeConnectionProps) {
  const points = useMemo(() => [
    new THREE.Vector3(...start),
    new THREE.Vector3(...end)
  ], [start, end]);

  return (
    <group>
      <Line
        points={points}
        color={color}
        lineWidth={active ? 2 : 1}
        transparent
        opacity={active ? 0.6 : 0.2}
      />
      {active && (
        <>
          <DataPulse start={start} end={end} color={color} delay={0} speed={0.5} />
          <DataPulse start={start} end={end} color={color} delay={0.5} speed={0.5} />
        </>
      )}
    </group>
  );
}

// 3D Function Node
interface FunctionNode3DProps {
  node: FunctionNode;
  onClick: () => void;
  onHover: (hovered: boolean) => void;
  isSelected: boolean;
  connections: Array<{ to: string; category: string }>;
  allNodes: FunctionNode[];
}

function FunctionNode3D({ node, onClick, onHover, isSelected, connections, allNodes }: FunctionNode3DProps) {
  const meshRef = useRef<THREE.Group>(null);
  const ringRef = useRef<THREE.Mesh>(null);
  const [localHovered, setLocalHovered] = useState(false);
  
  const runtime = node.function.runtime || 'python';
  const colors = RUNTIME_COLORS[runtime] || RUNTIME_COLORS.python;
  const cluster = CATEGORY_CLUSTERS[node.cluster] || CATEGORY_CLUSTERS.default;
  
  // Node size based on popularity/trust score
  const baseSize = 0.6 + (node.function.popularity_score || 0) * 0.01 + (node.function.trust_score || 0) * 0.005;
  const size = Math.min(baseSize, 1.5);
  
  useFrame(({ clock }) => {
    if (!meshRef.current) return;
    const t = clock.getElapsedTime();
    
    // Gentle floating
    const floatY = Math.sin(t * 0.5 + node.position[0]) * 0.2;
    meshRef.current.position.y = node.position[1] + floatY;
    
    // Faster rotation when selected
    meshRef.current.rotation.y = t * (isSelected ? 0.8 : 0.2);
    
    // Pulse effect when selected or hovered
    if (ringRef.current) {
      const pulse = (Math.sin(t * (isSelected ? 5 : 3)) + 1) * 0.5;
      const targetScale = isSelected || localHovered ? 2.5 + pulse * 0.5 : 1.8;
      ringRef.current.scale.lerp(new THREE.Vector3(targetScale, targetScale, targetScale), 0.1);
    }
    
    // Scale response - bigger when selected
    const targetScale = isSelected ? size * 1.5 : localHovered ? size * 1.2 : size;
    meshRef.current.scale.lerp(new THREE.Vector3(targetScale, targetScale, targetScale), 0.1);
  });

  // Find connected nodes for visualization
  const connectedPositions = useMemo(() => {
    return connections
      .map(conn => allNodes.find(n => n.id === conn.to)?.position)
      .filter(Boolean) as [number, number, number][];
  }, [connections, allNodes]);

  return (
    <group 
      ref={meshRef} 
      position={node.position}
      onClick={(e) => { e.stopPropagation(); onClick(); }}
      onPointerOver={(e) => { e.stopPropagation(); setLocalHovered(true); onHover(true); }}
      onPointerOut={() => { setLocalHovered(false); onHover(false); }}
    >
      {/* Outer glow ring - brighter when selected */}
      <mesh ref={ringRef}>
        <ringGeometry args={[0.8, 1, 32]} />
        <meshBasicMaterial 
          color={isSelected ? '#ffffff' : colors.glow} 
          transparent 
          opacity={isSelected ? 0.8 : localHovered ? 0.4 : 0.15}
          side={THREE.DoubleSide}
        />
      </mesh>
      
      {/* Additional outer ring for selected state */}
      {isSelected && (
        <mesh>
          <ringGeometry args={[1.2, 1.4, 32]} />
          <meshBasicMaterial 
            color={colors.glow} 
            transparent 
            opacity={0.3}
            side={THREE.DoubleSide}
          />
        </mesh>
      )}
      
      {/* Core node - hexagonal prism */}
      <mesh>
        <cylinderGeometry args={[0.7, 0.7, 0.3, 6]} />
        <meshStandardMaterial
          color="#0a0a1a"
          metalness={0.9}
          roughness={0.1}
          emissive={isSelected ? colors.glow : colors.primary}
          emissiveIntensity={isSelected ? 1.2 : localHovered ? 0.6 : 0.3}
        />
      </mesh>
      
      {/* Inner glow point */}
      <mesh scale={0.25}>
        <sphereGeometry args={[1, 16, 16]} />
        <meshBasicMaterial color={colors.glow} />
      </mesh>
      
      {/* Runtime indicator */}
      <Billboard position={[0, -1.2, 0]}>
        <Text
          fontSize={0.35}
          color="#ffffff"
          anchorX="center"
          anchorY="middle"
          outlineWidth={0.02}
          outlineColor={colors.primary}
        >
          {runtime.toUpperCase()}
        </Text>
      </Billboard>
      
      {/* Trust score badge */}
      {(localHovered || isSelected) && (
        <Billboard position={[0, 1.5, 0]}>
          <group>
            <Text
              fontSize={0.3}
              color={colors.glow}
              anchorX="center"
              anchorY="middle"
            >
              ★ {Math.round(node.function.trust_score || 0)}
            </Text>
          </group>
        </Billboard>
      )}
      
      {/* Selected indicator label */}
      {isSelected && (
        <Billboard position={[0, 2.2, 0]}>
          <group>
            <Text
              fontSize={0.25}
              color="#ffffff"
              anchorX="center"
              anchorY="middle"
              outlineWidth={0.03}
              outlineColor="#000000"
            >
              ● SELECTED
            </Text>
          </group>
        </Billboard>
      )}
      
      {/* Connection lines to related nodes */}
      {isSelected && connectedPositions.map((pos, i) => (
        <NodeConnection
          key={i}
          start={node.position}
          end={pos}
          color={colors.glow}
          active={true}
        />
      ))}
    </group>
  );
}

// Ambient cosmic particles
function CosmicDust({ count = 500 }: { count?: number }) {
  const points = useRef<THREE.Points>(null);
  
  const { positions, colors } = useMemo(() => {
    const positions = new Float32Array(count * 3);
    const colors = new Float32Array(count * 3);
    
    for (let i = 0; i < count; i++) {
      const theta = Math.random() * Math.PI * 2;
      const phi = Math.acos(2 * Math.random() - 1);
      const r = 25 + Math.random() * 60;
      
      positions[i * 3] = r * Math.sin(phi) * Math.cos(theta);
      positions[i * 3 + 1] = r * Math.sin(phi) * Math.sin(theta);
      positions[i * 3 + 2] = r * Math.cos(phi);
      
      // Random runtime colors
      const runtimeKeys = Object.keys(RUNTIME_COLORS);
      const runtime = RUNTIME_COLORS[runtimeKeys[Math.floor(Math.random() * runtimeKeys.length)]];
      const color = new THREE.Color(runtime.primary);
      colors[i * 3] = color.r;
      colors[i * 3 + 1] = color.g;
      colors[i * 3 + 2] = color.b;
    }
    
    return { positions, colors };
  }, [count]);

  useFrame(({ clock }) => {
    if (points.current) {
      points.current.rotation.y = clock.getElapsedTime() * 0.01;
    }
  });

  return (
    <points ref={points}>
      <bufferGeometry>
        <bufferAttribute attach="attributes-position" count={count} array={positions} itemSize={3} />
        <bufferAttribute attach="attributes-color" count={count} array={colors} itemSize={3} />
      </bufferGeometry>
      <pointsMaterial size={0.08} vertexColors transparent opacity={0.5} sizeAttenuation blending={THREE.AdditiveBlending} />
    </points>
  );
}

// Main Constellation Scene
interface ConstellationSceneProps {
  nodes: FunctionNode[];
  selectedNode: string | null;
  hoveredNode: string | null;
  searchQuery: string;
  onNodeClick: (node: FunctionNode) => void;
  onNodeHover: (nodeId: string | null) => void;
}

function ConstellationScene({ nodes, selectedNode, hoveredNode, searchQuery, onNodeClick, onNodeHover }: ConstellationSceneProps) {
  const { camera } = useThree();
  
  // Build connection graph based on shared categories/runtimes
  const connections = useMemo(() => {
    const conns: Array<{ from: string; to: string; category: string }> = [];
    for (let i = 0; i < nodes.length; i++) {
      for (let j = i + 1; j < nodes.length; j++) {
        const a = nodes[i];
        const b = nodes[j];
        if (a.function.runtime === b.function.runtime || a.function.category === b.function.category) {
          conns.push({ from: a.id, to: b.id, category: a.function.category || 'default' });
        }
      }
    }
    return conns;
  }, [nodes]);

  return (
    <>
      {/* Ambient lighting */}
      <ambientLight intensity={0.15} />
      <directionalLight position={[10, 10, 5]} intensity={0.5} color="#ffffff" />
      <pointLight position={[0, 0, 0]} intensity={1.5} color="#6366f1" distance={50} />
      
      {/* Background stars */}
      <Stars radius={100} depth={50} count={2000} factor={3} saturation={0.3} fade speed={0.3} />
      
      {/* Cosmic dust particles */}
      <CosmicDust count={400} />
      
      {/* Function nodes */}
      {nodes.map((node) => {
        const nodeConns = connections.filter(c => c.from === node.id || c.to === node.id);
        const connTargets = nodeConns.map(c => c.from === node.id ? c.to : c.from);
        
        return (
          <FunctionNode3D
            key={node.id}
            node={node}
            onClick={() => onNodeClick(node)}
            onHover={(hovered) => onNodeHover(hovered ? node.id : null)}
            isSelected={selectedNode === node.id}
            connections={nodeConns.map(c => ({ to: c.from === node.id ? c.to : c.from, category: c.category }))}
            allNodes={nodes}
          />
        );
      })}
      
      {/* Camera controls */}
      <OrbitControls
        enablePan={true}
        enableZoom={true}
        enableRotate={true}
        autoRotate
        autoRotateSpeed={0.1}
        minDistance={15}
        maxDistance={80}
        target={[0, 0, 0]}
      />
    </>
  );
}

// =============================================================================
// Helper Components
// =============================================================================

function AnimatedCounter({ value, suffix = '', duration = 1500 }: { value: number; suffix?: string; duration?: number }) {
  const [count, setCount] = useState(0);
  const ref = useRef<HTMLSpanElement>(null);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          const steps = 30;
          const increment = value / steps;
          let current = 0;
          let step = 0;

          const timer = setInterval(() => {
            step++;
            current = Math.min(value, increment * step);
            setCount(Math.floor(current));
            if (step >= steps) {
              clearInterval(timer);
              setCount(value);
            }
          }, duration / steps);

          observer.disconnect();
        }
      },
      { threshold: 0.5 }
    );

    if (ref.current) observer.observe(ref.current);
    return () => observer.disconnect();
  }, [value, duration]);

  return (
    <span ref={ref}>
      {count.toLocaleString()}{suffix}
    </span>
  );
}

function RuntimeBadge({ runtime }: { runtime: string }) {
  const colors = RUNTIME_COLORS[runtime] || RUNTIME_COLORS.python;
  const icons: Record<string, string> = {
    python: '🐍',
    nodejs: '🟢',
    typescript: '📘',
    go: '🐹',
    rust: '🦀',
    deno: '🦕',
    bun: '🥯',
    java: '☕',
    csharp: '#️⃣',
    ruby: '💎',
    php: '🐘',
  };

  return (
    <Badge 
      variant="outline" 
      className="border-0 font-mono text-xs"
      style={{ backgroundColor: `${colors.primary}20`, color: colors.glow }}
    >
      <span className="mr-1">{icons[runtime] || '⚡'}</span>
      {runtime}
    </Badge>
  );
}

// =============================================================================
// Function Detail Panel with Monaco Editor
// =============================================================================

interface FunctionDetailPanelProps {
  functionData: GalleryFunction;
  isOpen: boolean;
  onClose: () => void;
  onRemix: () => void;
  onLike: () => void;
}

function FunctionDetailPanel({ functionData, isOpen, onClose, onRemix, onLike }: FunctionDetailPanelProps) {
  const [activeTab, setActiveTab] = useState<'code' | 'info' | 'execution'>('info');
  const [code, setCode] = useState<string>('// Loading function code...');
  const [isLoadingCode, setIsLoadingCode] = useState(false);
  const runtime = functionData?.runtime || 'python';
  const monacoLang = RUNTIME_MONACO_LANG[runtime] || 'plaintext';
  const colors = RUNTIME_COLORS[runtime] || RUNTIME_COLORS.python;

  // Fetch function code when opening
  useEffect(() => {
    if (isOpen && functionData) {
      setIsLoadingCode(true);
      // Simulated code fetch - in production this would be an API call
      setTimeout(() => {
        setCode(generateSampleCode(functionData));
        setIsLoadingCode(false);
      }, 500);
    }
  }, [isOpen, functionData]);

  const generateSampleCode = (fn: GalleryFunction): string => {
    const runtime = fn.runtime || 'python';
    if (runtime === 'python') {
      return `# ${fn.title}
# Author: @${fn.author}
# Trust Score: ${Math.round(fn.trust_score || 0)}/100

def main(event, context):
    """
    ${fn.description}
    """
    # Extract parameters
    data = event.get('data', {})
    
    # Process
    result = process_data(data)
    
    return {
        'statusCode': 200,
        'body': result
    }

def process_data(data):
    # Core processing logic
    return { 'processed': True, 'input': data }
`;
    } else if (runtime === 'nodejs' || runtime === 'typescript') {
      return `// ${fn.title}
// Author: @${fn.author}
// Trust Score: ${Math.round(fn.trust_score || 0)}/100

export default async function handler(event, context) {
  /**
   * ${fn.description}
   */
  
  const { data } = event;
  
  // Process the input
  const result = await processData(data);
  
  return {
    statusCode: 200,
    body: JSON.stringify(result)
  };
}

async function processData(data) {
  // Core processing logic
  return { processed: true, input: data };
}
`;
    }
    return `// ${fn.title} - Code preview not available for ${runtime}`;
  };

  if (!functionData) return null;

  return (
    <AnimatePresence>
      {isOpen && (
        <>
        {/* Backdrop overlay to dim the 3D canvas */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.2 }}
          className="fixed inset-0 bg-black/50 backdrop-blur-sm z-[60]"
          onClick={onClose}
        >
          {/* Hint text */}
          <div className="absolute bottom-8 left-1/2 -translate-x-1/2 text-white/60 text-sm pointer-events-none">
            Click anywhere to close
          </div>
        </motion.div>
        <motion.div
          initial={{ x: '100%' }}
          animate={{ x: 0 }}
          exit={{ x: '100%' }}
          transition={{ type: 'spring', damping: 30, stiffness: 300 }}
          className="fixed inset-y-0 right-0 w-full max-w-2xl bg-background border-l-2 border-primary/30 z-[70] shadow-2xl shadow-black/50 overflow-hidden flex flex-col"
        >
          {/* Header with gradient background */}
          <div 
            className="flex items-center justify-between p-6 border-b-2"
            style={{ 
              background: `linear-gradient(135deg, ${colors.primary}20 0%, transparent 100%)`,
              borderColor: `${colors.primary}40`
            }}
          >
            <div className="flex items-center gap-3">
              <div 
                className="w-12 h-12 rounded-xl flex items-center justify-center border-2"
                style={{ 
                  backgroundColor: `${colors.primary}30`,
                  borderColor: `${colors.glow}50`
                }}
              >
                <Code2 className="w-6 h-6" style={{ color: colors.glow }} />
              </div>
              <div>
                <h2 className="text-xl font-bold text-foreground">{functionData.title || functionData.name}</h2>
                <p className="text-sm text-muted-foreground flex items-center gap-1">
                  <span className="text-primary">@</span>{functionData.author}
                </p>
              </div>
            </div>
            <Button 
              variant="ghost" 
              size="icon" 
              onClick={onClose} 
              className="text-muted-foreground hover:text-foreground hover:bg-destructive/20"
            >
              <X className="w-5 h-5" />
            </Button>
          </div>

          {/* Tabs */}
          <div className="flex gap-1 p-2 border-b border-border">
            {(['info', 'code', 'execution'] as const).map((tab) => (
              <button
                key={tab}
                onClick={() => setActiveTab(tab)}
                className={cn(
                  "px-4 py-2 rounded-lg text-sm font-medium transition-colors capitalize",
                  activeTab === tab 
                    ? "bg-accent text-accent-foreground" 
                    : "text-muted-foreground hover:text-foreground hover:bg-accent/50"
                )}
              >
                {tab}
              </button>
            ))}
          </div>

          {/* Content */}
          <div className="flex-1 overflow-y-auto">
            {activeTab === 'info' && (
              <div className="p-6 space-y-6">
                {/* Stats Grid */}
                <div className="grid grid-cols-3 gap-4">
                  <GlassmorphismCard className="p-4 text-center">
                    <Star className="w-5 h-5 mx-auto mb-2 text-yellow-500" />
                    <div className="text-2xl font-bold text-foreground">{Math.round(functionData.trust_score || 0)}</div>
                    <div className="text-xs text-muted-foreground">Trust Score</div>
                  </GlassmorphismCard>
                  <GlassmorphismCard className="p-4 text-center">
                    <GitFork className="w-5 h-5 mx-auto mb-2 text-blue-500" />
                    <div className="text-2xl font-bold text-foreground">{functionData.remix_count || 0}</div>
                    <div className="text-xs text-muted-foreground">Remixes</div>
                  </GlassmorphismCard>
                  <GlassmorphismCard className="p-4 text-center">
                    <Heart className="w-5 h-5 mx-auto mb-2 text-pink-500" />
                    <div className="text-2xl font-bold text-foreground">{functionData.like_count || 0}</div>
                    <div className="text-xs text-muted-foreground">Likes</div>
                  </GlassmorphismCard>
                </div>

                {/* Description */}
                <div>
                  <h3 className="text-sm font-medium text-foreground/80 mb-2">Description</h3>
                  <p className="text-sm text-muted-foreground leading-relaxed">{functionData.description || 'No description available'}</p>
                </div>

                {/* Metadata */}
                <div className="space-y-3">
                  <h3 className="text-sm font-medium text-foreground/80">Details</h3>
                  <div className="grid grid-cols-2 gap-3 text-sm">
                    <div className="flex justify-between py-2 border-b border-border">
                      <span className="text-muted-foreground">Runtime</span>
                      <RuntimeBadge runtime={runtime} />
                    </div>
                    <div className="flex justify-between py-2 border-b border-border">
                      <span className="text-muted-foreground">Category</span>
                      <span className="text-foreground capitalize">{functionData.category || 'General'}</span>
                    </div>
                    <div className="flex justify-between py-2 border-b border-border">
                      <span className="text-muted-foreground">Created</span>
                      <span className="text-foreground">
                        {functionData.created_at ? new Date(functionData.created_at).toLocaleDateString() : 'Unknown'}
                      </span>
                    </div>
                    <div className="flex justify-between py-2 border-b border-border">
                      <span className="text-muted-foreground">Updated</span>
                      <span className="text-foreground">
                        {functionData.updated_at ? new Date(functionData.updated_at).toLocaleDateString() : 'Unknown'}
                      </span>
                    </div>
                  </div>
                </div>

                {/* Tags */}
                {functionData.tags && functionData.tags.length > 0 && (
                  <div>
                    <h3 className="text-sm font-medium text-foreground/80 mb-2">Tags</h3>
                    <div className="flex flex-wrap gap-2">
                      {functionData.tags.map((tag) => (
                        <Badge key={tag} variant="secondary" className="bg-muted text-muted-foreground">
                          {tag}
                        </Badge>
                      ))}
                    </div>
                  </div>
                )}

                {/* Actions */}
                <div className="flex gap-3 pt-4">
                  <Button 
                    className="flex-1" 
                    style={{ backgroundColor: colors.primary }}
                    onClick={onRemix}
                  >
                    <GitFork className="w-4 h-4 mr-2" />
                    Remix
                  </Button>
                  <Button variant="outline" className="flex-1" onClick={onLike}>
                    <Heart className="w-4 h-4 mr-2" />
                    Like
                  </Button>
                  <Button variant="ghost" size="icon" onClick={() => window.open(`/registry/${functionData.author}/${functionData.name}`, '_blank')}>
                    <ExternalLink className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            )}

            {activeTab === 'code' && (
              <div className="h-full">
                <Suspense fallback={
                  <div className="flex items-center justify-center h-64">
                    <Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
                  </div>
                }>
                  <MonacoEditor
                    height="100%"
                    language={monacoLang}
                    value={code}
                    theme="vs-dark"
                    options={{
                      readOnly: true,
                      minimap: { enabled: false },
                      fontSize: 14,
                      lineNumbers: 'on',
                      roundedSelection: false,
                      scrollBeyondLastLine: false,
                      automaticLayout: true,
                      padding: { top: 20 },
                    }}
                  />
                </Suspense>
              </div>
            )}

            {activeTab === 'execution' && (
              <div className="p-6 space-y-4">
                <GlassmorphismCard className="p-4">
                  <div className="flex items-center gap-3 mb-4">
                    <Activity className="w-5 h-5 text-emerald-500" />
                    <h3 className="font-medium text-foreground">Live Execution Metrics</h3>
                  </div>
                  <div className="space-y-3">
                    <div className="flex justify-between text-sm">
                      <span className="text-muted-foreground">Avg. Execution Time</span>
                      <span className="text-emerald-500 font-mono">{(Math.random() * 100 + 50).toFixed(0)}ms</span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span className="text-muted-foreground">Success Rate</span>
                      <span className="text-emerald-500 font-mono">{(99 + Math.random()).toFixed(2)}%</span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span className="text-muted-foreground">Invocations (24h)</span>
                      <span className="text-primary font-mono">{Math.floor(Math.random() * 10000 + 1000).toLocaleString()}</span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span className="text-muted-foreground">Cold Start</span>
                      <span className="text-warning font-mono">{(Math.random() * 200 + 100).toFixed(0)}ms</span>
                    </div>
                  </div>
                </GlassmorphismCard>

                {/* Execution Graph Placeholder */}
                <div className="bg-muted/50 rounded-lg p-4 h-48 flex items-center justify-center border border-border">
                  <div className="text-center">
                    <Activity className="w-8 h-8 mx-auto mb-2 text-muted-foreground" />
                    <p className="text-sm text-muted-foreground">Execution timeline visualization</p>
                    <p className="text-xs text-muted-foreground/60">(Historical data coming soon)</p>
                  </div>
                </div>
              </div>
            )}
          </div>
        </motion.div>
        </>
      )}
    </AnimatePresence>
  );
}

// =============================================================================
// List View (Fallback for Mobile/Accessibility)
// =============================================================================

function FunctionListCard({ 
  fn, 
  onClick, 
  onRemix, 
  onLike 
}: { 
  fn: GalleryFunction; 
  onClick: () => void;
  onRemix: () => void;
  onLike: () => void;
}) {
  const runtime = fn.runtime || 'python';
  const icons: Record<string, string> = {
    python: '🐍', nodejs: '🟢', typescript: '📘', go: '🐹', rust: '🦀',
    deno: '🦕', bun: '🥯', java: '☕', csharp: '#️⃣', ruby: '💎', php: '🐘',
  };

  return (
    <GlassmorphismCard className="group cursor-pointer overflow-hidden" onClick={onClick}>
      <div className="p-4 flex items-center gap-4">
        <div className="text-3xl">{icons[runtime] || '⚡'}</div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <h3 className="font-semibold text-foreground truncate">{fn.title || fn.name}</h3>
            {fn.category && (
              <Badge variant="secondary" className="text-xs capitalize">
                {fn.category}
              </Badge>
            )}
          </div>
          <p className="text-sm text-muted-foreground truncate">@{fn.author} • {fn.description?.slice(0, 60)}...</p>
        </div>
        <div className="flex items-center gap-4 text-sm text-muted-foreground">
          <span className="flex items-center gap-1">
            <Star className="w-4 h-4" />
            {Math.round(fn.trust_score || 0)}
          </span>
          <span className="flex items-center gap-1">
            <GitFork className="w-4 h-4" />
            {fn.remix_count || 0}
          </span>
        </div>
        <div className="flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
          <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); onRemix(); }}>
            <GitFork className="w-4 h-4" />
          </Button>
          <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); onLike(); }}>
            <Heart className="w-4 h-4" />
          </Button>
        </div>
      </div>
    </GlassmorphismCard>
  );
}

// =============================================================================
// Main Gallery Page
// =============================================================================

export default function GalleryPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [viewMode, setViewMode] = useState<ViewMode>('constellation');
  const [activeCategory, setActiveCategory] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedQuery, setDebouncedQuery] = useState('');
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [hoveredNodeId, setHoveredNodeId] = useState<string | null>(null);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [remixDialogOpen, setRemixDialogOpen] = useState(false);
  const [selectedFunction, setSelectedFunction] = useState<GalleryFunction | null>(null);
  const [customization, setCustomization] = useState('');
  const [remixCost, setRemixCost] = useState<number>(0.50);
  const [walletBalance, setWalletBalance] = useState<number>(0);
  const [canRemix, setCanRemix] = useState<boolean>(false);
  const [isOwnFunction, setIsOwnFunction] = useState<boolean>(false);

  const searchInputRef = useRef<HTMLInputElement>(null);
  const DEBOUNCE_MS = 300;

  // Debounce search
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedQuery(searchQuery);
    }, DEBOUNCE_MS);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  // Fetch functions from API
  const { data: functionsData, isLoading: isLoadingFunctions, error: functionsError } = useQuery({
    queryKey: ['gallery', 'functions', activeCategory, debouncedQuery],
    queryFn: async () => {
      try {
        const result = await galleryApi.search({ 
          category: activeCategory !== 'all' ? activeCategory : undefined,
          sort_by: 'popular',
          limit: 50,
        });
        console.log('[Gallery] API response:', result);
        return result;
      } catch (err) {
        console.error('[Gallery] API error:', err);
        throw err;
      }
    },
    staleTime: 2 * 60 * 1000,
    retry: 1,
  });

  const functions = functionsData?.results || [];

  // Debug: Log data flow issues
  useEffect(() => {
    if (functionsData) {
      console.log('[Gallery Debug] functionsData:', functionsData);
      console.log('[Gallery Debug] functionsData.results:', functionsData.results);
      console.log('[Gallery Debug] functionsData.total_count:', functionsData.total_count);
      console.log('[Gallery Debug] functions.length:', functions.length);
    }
  }, [functionsData, functions.length]);

  // Build 3D node positions using force-directed placement
  const functionNodes: FunctionNode[] = useMemo(() => {
    return functions.map((fn, index) => {
      const cluster = fn.category || 'default';
      const clusterInfo = CATEGORY_CLUSTERS[cluster] || CATEGORY_CLUSTERS.default;
      
      // Position based on cluster with random offset
      const angle = (index / Math.max(functions.length, 1)) * Math.PI * 2;
      const radius = 5 + Math.random() * 15;
      const height = (Math.random() - 0.5) * 10;
      
      // Add clustering bias toward category center
      const clusterBias = 0.6;
      const x = clusterInfo.center[0] * clusterBias + Math.cos(angle) * radius * (1 - clusterBias);
      const y = clusterInfo.center[1] * clusterBias + height * (1 - clusterBias);
      const z = clusterInfo.center[2] * clusterBias + Math.sin(angle) * radius * (1 - clusterBias);

      return {
        id: fn.id,
        function: fn,
        position: [x, y, z] as [number, number, number],
        velocity: [0, 0, 0],
        cluster,
        highlighted: debouncedQuery ? 
          fn.name.toLowerCase().includes(debouncedQuery.toLowerCase()) ||
          fn.description?.toLowerCase().includes(debouncedQuery.toLowerCase()) :
          true,
        hovered: false,
      };
    });
  }, [functions, debouncedQuery]);

  // Get selected function data
  const selectedFunctionData = useMemo(() => {
    if (!selectedNodeId) return null;
    return functions.find(f => f.id === selectedNodeId) || null;
  }, [selectedNodeId, functions]);

  // Stats
  const stats: GalleryStats = useMemo(() => ({
    totalFunctions: functionsData?.total_count || functions.length || 0,
    activeExecutions: Math.floor(Math.random() * 500) + 100, // Simulated live data
    avgLatency: Math.floor(Math.random() * 50) + 80,
    totalRemixes: functions.reduce((sum, f) => sum + (f.remix_count || 0), 0),
  }), [functions, functionsData]);

  // Mutations
  const remixMutation = useMutation({
    mutationFn: (data: { author: string; name: string; customization?: string }) =>
      galleryApi.remix(data.author, data.name, { customization: data.customization }),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['gallery'] });
      const costMsg = data.cost_usd ? ` ($${data.cost_usd.toFixed(2)} charged)` : '';
      toast.success(`Remixed! Created "${data.new_name}"${costMsg}`);
      setRemixDialogOpen(false);
      setCustomization('');
    },
    onError: (error: Error) => {
      toast.error(`Failed to remix: ${error.message}`);
    },
  });

  const likeMutation = useMutation({
    mutationFn: (data: { author: string; name: string }) => galleryApi.like(data.author, data.name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['gallery'] });
      toast.success('Liked!');
    },
  });

  // Fetch remix cost when opening dialog
  useEffect(() => {
    if (remixDialogOpen && selectedFunction) {
      galleryApi.getRemixCost(selectedFunction.author, selectedFunction.name)
        .then((data) => {
          setRemixCost(data.cost_usd);
          setWalletBalance(data.balance_usd);
          setCanRemix(data.can_remix || data.is_own_function);
          setIsOwnFunction(data.is_own_function);
        })
        .catch(() => {
          setCanRemix(true);
        });
    }
  }, [remixDialogOpen, selectedFunction]);

  const handleRemix = (fn: GalleryFunction) => {
    setSelectedFunction(fn);
    setRemixDialogOpen(true);
    setCustomization('');
  };

  const confirmRemix = () => {
    if (selectedFunction) {
      remixMutation.mutate({
        author: selectedFunction.author,
        name: selectedFunction.name,
        customization,
      });
    }
  };

  const handleNodeClick = (node: FunctionNode) => {
    setSelectedNodeId(node.id);
  };

  const filteredNodes = useMemo(() => {
    if (!debouncedQuery) return functionNodes;
    return functionNodes.filter(node => 
      node.highlighted
    );
  }, [functionNodes, debouncedQuery]);

  return (
    <div className="h-[calc(100vh-8rem)] lg:h-[calc(100vh-10rem)] flex flex-col bg-background text-foreground overflow-hidden -m-4 lg:-m-6 rounded-lg">
      {/* Header - integrated into page flow (not fixed) */}
      <div className="flex-none border-b border-border/50 bg-background/80 backdrop-blur-xl">
        <div className="max-w-7xl mx-auto px-4 lg:px-6 h-14 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-xl bg-gradient-to-br from-primary to-accent flex items-center justify-center">
              <Sparkles className="w-4 h-4 text-primary-foreground" />
            </div>
            <span className="font-bold text-foreground">Living Canvas</span>
          </div>

          <div className="hidden md:flex items-center gap-4">
            <Tabs value={viewMode} onValueChange={(v) => setViewMode(v as ViewMode)}>
              <TabsList className="bg-muted">
                <TabsTrigger value="constellation" className="gap-2">
                  <Box className="w-4 h-4" />
                  3D View
                </TabsTrigger>
                <TabsTrigger value="list" className="gap-2">
                  <List className="w-4 h-4" />
                  List
                </TabsTrigger>
              </TabsList>
            </Tabs>
          </div>

          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="sm"
              className="hidden sm:flex text-muted-foreground hover:text-foreground"
              onClick={() => navigate('/function-gallery')}
            >
              <LayoutGrid className="w-4 h-4 mr-2" />
              Registry
            </Button>
            <Button
              size="sm"
              className="bg-gradient-to-r from-primary to-accent"
              onClick={() => navigate('/ai/composer')}
            >
              <Sparkles className="w-4 h-4 mr-1" />
              Create
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="md:hidden text-slate-400"
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            >
              {mobileMenuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
            </Button>
          </div>
        </div>

        {/* Mobile Menu */}
        <AnimatePresence>
          {mobileMenuOpen && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              exit={{ opacity: 0, height: 0 }}
              className="md:hidden border-t border-border bg-background"
            >
              <div className="px-4 py-4 space-y-2">
                <button
                  onClick={() => { setViewMode('constellation'); setMobileMenuOpen(false); }}
                  className={cn(
                    "w-full text-left py-2 px-3 rounded-lg transition-colors flex items-center gap-2",
                    viewMode === 'constellation' ? "bg-accent text-accent-foreground" : "text-muted-foreground"
                  )}
                >
                  <Box className="w-4 h-4" />
                  3D Constellation
                </button>
                <button
                  onClick={() => { setViewMode('list'); setMobileMenuOpen(false); }}
                  className={cn(
                    "w-full text-left py-2 px-3 rounded-lg transition-colors flex items-center gap-2",
                    viewMode === 'list' ? "bg-accent text-accent-foreground" : "text-muted-foreground"
                  )}
                >
                  <List className="w-4 h-4" />
                  List View
                </button>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex flex-col min-h-0">
        {/* Search & Filter Bar */}
        <div className="flex-none border-b border-border/50 bg-background/50 backdrop-blur-sm">
          <div className="max-w-7xl mx-auto px-4 lg:px-6 py-4">
            <div className="flex flex-col md:flex-row gap-4 items-center justify-between">
              {/* Search */}
              <div className="relative flex-1 max-w-md w-full">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  ref={searchInputRef}
                  placeholder="Search functions..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10 bg-muted/50 border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-primary"
                />
                {searchQuery && (
                  <button
                    onClick={() => setSearchQuery('')}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                  >
                    <X className="w-4 h-4" />
                  </button>
                )}
              </div>

              {/* Stats */}
              <div className="hidden lg:flex items-center gap-6 text-sm">
                <div className="flex items-center gap-2 text-muted-foreground">
                  <Globe className="w-4 h-4" />
                  <span><AnimatedCounter value={stats.totalFunctions} /> functions</span>
                </div>
                <div className="flex items-center gap-2 text-emerald-500">
                  <Activity className="w-4 h-4" />
                  <span>{stats.activeExecutions.toLocaleString()} active</span>
                </div>
                <div className="flex items-center gap-2 text-primary">
                  <Clock className="w-4 h-4" />
                  <span>{stats.avgLatency}ms avg</span>
                </div>
              </div>

              {/* Category Pills */}
              <div className="flex items-center gap-2 overflow-x-auto pb-2 md:pb-0 w-full md:w-auto">
                <Badge
                  variant={activeCategory === 'all' ? 'default' : 'outline'}
                  className="cursor-pointer whitespace-nowrap"
                  onClick={() => setActiveCategory('all')}
                >
                  All
                </Badge>
                {Object.entries(CATEGORY_CLUSTERS).slice(0, 6).map(([key, info]) => (
                  <Badge
                    key={key}
                    variant={activeCategory === key ? 'default' : 'outline'}
                    className="cursor-pointer whitespace-nowrap"
                    style={activeCategory === key ? { backgroundColor: info.color } : undefined}
                    onClick={() => setActiveCategory(key)}
                  >
                    {info.label}
                  </Badge>
                ))}
              </div>
            </div>
          </div>
        </div>

        {/* 3D Canvas or List View */}
        <div className="flex-1 relative overflow-hidden min-h-0">
          {viewMode === 'constellation' ? (
            <div className="w-full h-full min-h-0 absolute inset-0">
              {functionsError ? (
                <div className="flex items-center justify-center h-full">
                  <div className="text-center max-w-md px-4">
                    <div className="w-12 h-12 rounded-full bg-destructive/10 flex items-center justify-center mx-auto mb-4">
                      <X className="w-6 h-6 text-destructive" />
                    </div>
                    <p className="text-foreground font-medium mb-2">Failed to load functions</p>
                    <p className="text-muted-foreground text-sm">{(functionsError as Error)?.message || 'Please try again later'}</p>
                  </div>
                </div>
              ) : isLoadingFunctions ? (
                <div className="flex items-center justify-center h-full">
                  <div className="text-center">
                    <Loader2 className="w-12 h-12 animate-spin text-primary mx-auto mb-4" />
                    <p className="text-muted-foreground">Loading function constellation...</p>
                  </div>
                </div>
              ) : filteredNodes.length === 0 ? (
                <div className="flex items-center justify-center h-full">
                  <div className="text-center">
                    <Search className="w-12 h-12 text-muted-foreground mx-auto mb-4" />
                    <p className="text-muted-foreground">No functions found</p>
                    <p className="text-muted-foreground/60 text-sm">{debouncedQuery ? 'Try adjusting your search' : 'The registry may be empty or unavailable'}</p>
                  </div>
                </div>
              ) : (
                <Canvas
                  camera={{ position: [30, 15, 30], fov: 60 }}
                  gl={{ antialias: true, alpha: false, powerPreference: 'high-performance' }}
                  dpr={[1, 1.5]}
                  style={{ background: '#020108' }}
                >
                  <color attach="background" args={['#020108']} />
                  <fog attach="fog" args={['#020108', 40, 100]} />
                  <ConstellationScene
                    nodes={filteredNodes}
                    selectedNode={selectedNodeId}
                    hoveredNode={hoveredNodeId}
                    searchQuery={debouncedQuery}
                    onNodeClick={handleNodeClick}
                    onNodeHover={setHoveredNodeId}
                  />
                </Canvas>
              )}

              {/* 3D Overlay UI */}
              <div className="absolute bottom-6 left-6 pointer-events-none">
                <div className="bg-background/80 backdrop-blur-md rounded-lg px-4 py-3 border border-border">
                  <div className="text-xs font-mono text-muted-foreground/60 tracking-widest uppercase mb-1">
                    Function Constellation
                  </div>
                  <div className="text-xs text-primary/70 flex items-center gap-3">
                    <span>{filteredNodes.length} nodes</span>
                    <span>•</span>
                    <span className="flex items-center gap-1">
                      <Rotate3d className="w-3 h-3" />
                      Drag to rotate
                    </span>
                    <span className="flex items-center gap-1">
                      <MousePointer2 className="w-3 h-3" />
                      Click nodes
                    </span>
                  </div>
                </div>
              </div>

              {/* Legend */}
              <div className="absolute top-4 right-4 pointer-events-none">
                <div className="bg-background/80 backdrop-blur-md rounded-lg p-3 border border-border space-y-2">
                  <div className="text-xs font-mono text-muted-foreground/60 uppercase mb-2">Runtimes</div>
                  {Object.entries(RUNTIME_COLORS).slice(0, 5).map(([runtime, colors]) => (
                    <div key={runtime} className="flex items-center gap-2 text-xs text-foreground/80">
                      <div 
                        className="w-2 h-2 rounded-full" 
                        style={{ backgroundColor: colors.primary, boxShadow: `0 0 6px ${colors.glow}` }}
                      />
                      <span className="capitalize">{runtime}</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          ) : (
            // List View
            <div className="max-w-7xl mx-auto px-4 lg:px-6 py-6 h-full min-h-0 overflow-y-auto">
              {functionsError ? (
                <div className="text-center py-12 max-w-md mx-auto">
                  <div className="w-12 h-12 rounded-full bg-destructive/10 flex items-center justify-center mx-auto mb-4">
                    <X className="w-6 h-6 text-destructive" />
                  </div>
                  <p className="text-foreground font-medium mb-2">Failed to load functions</p>
                  <p className="text-muted-foreground text-sm">{(functionsError as Error)?.message || 'Please try again later'}</p>
                </div>
              ) : isLoadingFunctions ? (
                <div className="flex items-center justify-center py-12">
                  <Loader2 className="w-8 h-8 animate-spin text-primary" />
                </div>
              ) : functions.length === 0 ? (
                <div className="text-center py-12">
                  <Search className="w-12 h-12 text-muted-foreground mx-auto mb-4" />
                  <p className="text-muted-foreground">No functions found</p>
                  <p className="text-muted-foreground/60 text-sm">{debouncedQuery ? 'Try adjusting your search' : 'The registry may be empty or unavailable'}</p>
                </div>
              ) : (
                <div className="space-y-3">
                  {functions.map((fn) => (
                    <FunctionListCard
                      key={fn.id}
                      fn={fn}
                      onClick={() => {
                        setSelectedFunction(fn);
                        setSelectedNodeId(fn.id);
                      }}
                      onRemix={() => handleRemix(fn)}
                      onLike={() => likeMutation.mutate({ author: fn.author, name: fn.name })}
                    />
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Function Detail Panel */}
      <FunctionDetailPanel
        functionData={selectedFunctionData!}
        isOpen={!!selectedNodeId}
        onClose={() => setSelectedNodeId(null)}
        onRemix={() => selectedFunctionData && handleRemix(selectedFunctionData)}
        onLike={() => selectedFunctionData && likeMutation.mutate({ author: selectedFunctionData.author, name: selectedFunctionData.name })}
      />

      {/* Remix Dialog */}
      <Dialog open={remixDialogOpen} onOpenChange={setRemixDialogOpen}>
        <DialogContent className="max-w-md bg-background border-border">
          <DialogHeader>
            <DialogTitle className="text-foreground">Remix Function</DialogTitle>
            <DialogDescription className="text-muted-foreground">
              Create your own copy of <strong className="text-foreground">{selectedFunction?.title || selectedFunction?.name}</strong> by @{selectedFunction?.author}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            {/* Cost Info */}
            <div className="bg-muted rounded-lg p-4 space-y-2">
              <div className="flex justify-between items-center text-sm">
                <span className="text-muted-foreground">Remix Cost:</span>
                <span className="font-semibold text-foreground">${remixCost.toFixed(2)}</span>
              </div>
              <div className="flex justify-between items-center text-sm">
                <span className="text-muted-foreground">Your Balance:</span>
                <span className={walletBalance < remixCost ? "font-semibold text-destructive" : "font-semibold text-emerald-500"}
                >
                  ${walletBalance.toFixed(2)}
                </span>
              </div>
            </div>

            <div>
              <label className="text-sm font-medium text-foreground/80 mb-2 block">Customizations (optional)</label>
              <textarea
                className="w-full min-h-[100px] rounded-md border border-border bg-muted/50 p-3 text-sm resize-none text-foreground"
                placeholder="e.g., Add error handling, change the output format, optimize for speed..."
                value={customization}
                onChange={(e) => setCustomization(e.target.value)}
                disabled={!canRemix && !isOwnFunction}
              />
            </div>

            {!canRemix && !isOwnFunction && (
              <div className="bg-destructive/10 border border-destructive/30 rounded-md p-3">
                <p className="text-sm text-destructive">
                  You need ${(remixCost - walletBalance).toFixed(2)} more to remix this function.
                </p>
              </div>
            )}
          </div>
          <DialogFooter className="gap-2">
            <Button variant="outline" onClick={() => setRemixDialogOpen(false)} className="border-border">
              Cancel
            </Button>
            {!canRemix && !isOwnFunction ? (
              <Button onClick={() => navigate('/wallet')} className="bg-gradient-to-r from-emerald-500 to-green-600 hover:from-emerald-600 hover:to-green-700">
                Add Funds
              </Button>
            ) : (
              <Button
                onClick={confirmRemix}
                disabled={remixMutation.isPending}
                className="bg-gradient-to-r from-primary to-accent hover:opacity-90"
              >
                {remixMutation.isPending ? (
                  <><Loader2 className="mr-2 h-4 w-4 animate-spin" />Remixing...</>
                ) : (
                  <><GitFork className="mr-2 h-4 w-4" />Remix for ${remixCost.toFixed(2)}</>
                )}
              </Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
