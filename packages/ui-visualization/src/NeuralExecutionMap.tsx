/**
 * @functionfly/ui-visualization
 * Neural Execution Map - 3D cinematic visualization for FunctionFly Studio
 */

import * as React from "react";
import { Canvas, useFrame, useThree } from "@react-three/fiber";
import { OrbitControls, Text, useGLTF, Environment, Stars } from "@react-three/drei";
import { useSpring, animated, config } from "@react-spring/three";
import { Group, Vector3, Color, MathUtils } from "three";
import { cn } from "@functionfly/ui-core";

// --- Types ---
export interface NeuralNode {
  id: string;
  label: string;
  type: "agent" | "function" | "api" | "memory" | "database" | "input" | "output";
  position: [number, number, number];
  status?: "idle" | "active" | "error" | "processing";
  connections: string[];
  load?: number; // 0-1
  size?: number;
  color?: string;
  description?: string;
}

export interface NeuralConnection {
  from: string;
  to: string;
  type?: "data" | "control" | "signal";
  strength?: number; // 0-1
  active?: boolean;
}

export interface NeuralExecutionMapProps {
  nodes: NeuralNode[];
  connections: NeuralConnection[];
  className?: string;
  showLabels?: boolean;
  showConnections?: boolean;
  interactive?: boolean;
  autoRotate?: boolean;
  style?: "dark" | "light" | "cosmic";
}

export interface ParticleFlowProps {
  count?: number;
  speed?: number;
  color?: string;
  className?: string;
}

// --- Neural Node Component ---
function NeuralNode3D({ node, hovered, onHover }: { node: NeuralNode; hovered: boolean; onHover: (id: string | null) => void }) {
  const meshRef = React.useRef<Group>(null);
  const { nodes, connections, autoRotate } = useThree((state) => state.scene);

  const nodeSize = node.size || 1;
  const nodeColor = node.color || getNodeTypeColor(node.type);
  const glowIntensity = node.status === "active" ? 1.5 : node.status === "error" ? 2 : 0.5;

  const { scale, emissive } = useSpring({
    scale: hovered ? [nodeSize * 1.3, nodeSize * 1.3, nodeSize * 1.3] : [nodeSize, nodeSize, nodeSize],
    emissive: hovered ? nodeColor : nodeColor + "40",
    config: config.gentle,
  });

  React.useEffect(() => {
    if (autoRotate && meshRef.current) {
      meshRef.current.rotation.y += 0.002;
    }
  }, [autoRotate]);

  return (
    <group ref={meshRef} position={node.position} onPointerOver={() => onHover(node.id)} onPointerOut={() => onHover(null)}>
      {/* Node shape based on type */}
      {node.type === "agent" ? (
        <mesh>
          <octahedronGeometry args={[nodeSize * 0.8, 0]} />
          <meshStandardMaterial color={nodeColor} emissive={nodeColor + "60"} emissiveIntensity={glowIntensity} metalness={0.3} roughness={0.4} />
        </mesh>
      ) : node.type === "function" ? (
        <mesh>
          <boxGeometry args={[nodeSize, nodeSize * 0.6, nodeSize * 0.8]} />
          <meshStandardMaterial color={nodeColor} emissive={nodeColor + "40"} emissiveIntensity={glowIntensity} metalness={0.5} roughness={0.3} />
        </mesh>
      ) : node.type === "api" ? (
        <mesh>
          <cylinderGeometry args={[nodeSize * 0.5, nodeSize * 0.5, nodeSize * 1.2, 8]} />
          <meshStandardMaterial color={nodeColor} emissive={nodeColor + "50"} emissiveIntensity={glowIntensity} metalness={0.6} />
        </mesh>
      ) : node.type === "memory" ? (
        <mesh>
          <sphereGeometry args={[nodeSize * 0.7, 16, 16]} />
          <meshStandardMaterial color={nodeColor} emissive={nodeColor + "50"} emissiveIntensity={glowIntensity} wireframe metalness={0.2} />
        </mesh>
      ) : node.type === "database" ? (
        <mesh>
          <cylinderGeometry args={[nodeSize * 0.7, nodeSize * 0.7, nodeSize * 0.6, 32]} />
          <meshStandardMaterial color={nodeColor} emissive={nodeColor + "40"} emissiveIntensity={glowIntensity} metalness={0.4} />
        </mesh>
      ) : (
        <mesh>
          <sphereGeometry args={[nodeSize * 0.5, 12, 12]} />
          <meshStandardMaterial color={nodeColor} emissive={nodeColor + "30"} emissiveIntensity={glowIntensity} metalness={0.3} />
        </mesh>
      )}

      {/* Status indicator ring */}
      <ringGeometry args={[nodeSize * 0.9, nodeSize * 1.1, 24]} />
      <meshBasicMaterial color={node.status === "error" ? "#ef4444" : node.status === "active" ? "#10b981" : nodeColor} side={2} transparent opacity={0.3} />

      {/* Label */}
      {/* <Text position={[0, nodeSize + 0.5, 0]} fontSize={0.4} color="#ffffff" anchorX="center" anchorY="middle">
        {node.label}
      </Text> */}
    </group>
  );
}

// --- Connection Line Component ---
function NeuralConnectionLine({ from, to, active, strength = 0.5 }: { from: Vector3; to: Vector3; active?: boolean; strength?: number }) {
  const [progress, setProgress] = React.useState(0);

  useFrame(() => {
    if (active) {
      setProgress((p) => (p + 0.02) % 1);
    }
  });

  const midpoint = new Vector3().addVectors(from, to).multiplyScalar(0.5);
  const controlPoint = midpoint.clone().add(new Vector3(0, 2 + strength * 2, 0));

  const curvePoints = [from, controlPoint, to];
  const smoothPoints: Vector3[] = [];
  for (let i = 0; i <= 20; i++) {
    const t = i / 20;
    const mt = 1 - t;
    const x = mt * mt * curvePoints[0].x + 2 * mt * t * curvePoints[1].x + t * t * curvePoints[2].x;
    const y = mt * mt * curvePoints[0].y + 2 * mt * t * curvePoints[1].y + t * t * curvePoints[2].y;
    const z = mt * mt * curvePoints[0].z + 2 * mt * t * curvePoints[1].z + t * t * curvePoints[2].z;
    smoothPoints.push(new Vector3(x, y, z));
  }

  return (
    <group>
      {/* Main line */}
      <line>
        <bufferGeometry>
          <bufferAttribute
            attach="attributes-position"
            count={smoothPoints.length}
            array={Float32Array.from(flattenVectors(smoothPoints))}
            itemSize={3}
          />
        </bufferGeometry>
        <lineBasicMaterial color={active ? "#f97316" : "#3b82f6"} linewidth={1} opacity={strength * 0.5 + 0.1} transparent />
      </line>

      {/* Animated particle */}
      {active && (
        <mesh position={smoothPoints[Math.floor(progress * (smoothPoints.length - 1))] || from}>
          <sphereGeometry args={[0.08, 8, 8]} />
          <meshBasicMaterial color="#f97316" emissive="#f97316" emissiveIntensity={2} />
        </mesh>
      )}
    </group>
  );
}

function flattenVectors(vectors: Vector3[]): number[] {
  const result: number[] = [];
  for (const v of vectors) {
    result.push(v.x, v.y, v.z);
  }
  return result;
}

function getNodeTypeColor(type: string): string {
  return {
    agent: "#8b5cf6",
    function: "#f97316",
    api: "#3b82f6",
    memory: "#10b981",
    database: "#ef4444",
    input: "#06b6d4",
    output: "#6366f1",
  }[type] || "#6b7280";
}

// --- Particle Flow Background ---
export function ParticleFlow({ count = 200, speed = 0.5, color = "#f97316", className }: ParticleFlowProps) {
  const particlesRef = React.useRef<any>(null);
  const [positions] = React.useState(() => {
    const arr = new Float32Array(count * 3);
    for (let i = 0; i < count; i++) {
      arr[i * 3] = (Math.random() - 0.5) * 100;
      arr[i * 3 + 1] = (Math.random() - 0.5) * 100;
      arr[i * 3 + 2] = (Math.random() - 0.5) * 100;
    }
    return arr;
  });

  useFrame((state) => {
    if (particlesRef.current) {
      const positions = particlesRef.current.attributes.position.array as Float32Array;
      for (let i = 0; i < count; i++) {
        positions[i * 3 + 1] -= speed * 0.01;
        if (positions[i * 3 + 1] < -50) positions[i * 3 + 1] = 50;
      }
      particlesRef.current.attributes.position.needsUpdate = true;
    }
  });

  return (
    <points ref={particlesRef} className={className}>
      <bufferGeometry>
        <bufferAttribute attach="attributes-position" count={count} array={positions} itemSize={3} />
      </bufferGeometry>
      <pointsMaterial color={color} size={0.15} transparent opacity={0.3} sizeAttenuation />
    </points>
  );
}

// --- Main NeuralExecutionMap Component ---
export function NeuralExecutionMap({
  nodes,
  connections,
  className,
  showLabels = true,
  showConnections = true,
  interactive = true,
  autoRotate = true,
  style = "dark",
}: NeuralExecutionMapProps) {
  const [hoveredNode, setHoveredNode] = React.useState<string | null>(null);
  const nodeMap = React.useMemo(() => new Map(nodes.map((n) => [n.id, n])), [nodes]);
  const styleColors = {
    dark: { bg: "#050510", ambient: 0.3 },
    light: { bg: "#fafafa", ambient: 0.8 },
    cosmic: { bg: "#0a0015", ambient: 0.2 },
  };

  return (
    <div className={cn("w-full h-full rounded-xl overflow-hidden relative", className)}>
      {/* Status overlay */}
      <div className="absolute top-4 left-4 z-10 flex items-center gap-2 px-3 py-1.5 rounded-lg bg-black/50 backdrop-blur-sm text-[11px] text-text-secondary">
        <div className="size-2 rounded-full bg-green-500 animate-pulse" />
        Neural Map · {nodes.length} nodes · {connections.length} connections
      </div>

      <Canvas
        camera={{ position: [0, 15, 30], fov: 55 }}
        gl={{ antialias: true, alpha: true }}
        className="w-full h-full"
      >
        <color attach="background" args={[styleColors[style].bg]} />
        <ambientLight intensity={styleColors[style].ambient} />
        <directionalLight position={[10, 20, 15]} intensity={0.8} color="#f97316" />
        <directionalLight position={[-10, -10, -5]} intensity={0.3} />
        <pointLight position={[0, 10, 0]} intensity={0.5} color="#8b5cf6" />
        <Environment preset="night" />
        {style === "cosmic" && <Stars radius={100} depth={50} count={5000} factor={4} />}

        {/* Connections */}
        {showConnections &&
          connections.map((conn) => {
            const fromNode = nodeMap.get(conn.from);
            const toNode = nodeMap.get(conn.to);
            if (!fromNode || !toNode) return null;
            return (
              <NeuralConnectionLine
                key={`${conn.from}-${conn.to}`}
                from={new Vector3(...fromNode.position)}
                to={new Vector3(...toNode.position)}
                active={conn.active}
                strength={conn.strength}
              />
            );
          })}

        {/* Nodes */}
        {nodes.map((node) => (
          <NeuralNode3D
            key={node.id}
            node={node}
            hovered={hoveredNode === node.id}
            onHover={setHoveredNode}
          />
        ))}

        {/* Auto-rotation */}
        {interactive && <OrbitControls enableZoom enablePan autoRotate={autoRotate} autoRotateSpeed={0.5} />}
      </Canvas>

      {/* Tooltip */}
      {hoveredNode && (
        <div className="absolute bg-bg-primary border border-border-default rounded-lg px-3 py-2 text-sm shadow-lg pointer-events-none z-20"
          style={{ bottom: 16, left: 16 }}>
          <div className="font-medium text-text-primary">{nodeMap.get(hoveredNode)?.label}</div>
          <div className="text-[11px] text-text-muted">{nodeMap.get(hoveredNode)?.description}</div>
          <div className="text-[10px] text-brand-400 mt-1">{nodeMap.get(hoveredNode)?.type}</div>
        </div>
      )}
    </div>
  );
}

NeuralExecutionMap.displayName = "NeuralExecutionMap";
// ============================================================================
// TokenParticleSystem - Animated particle flow for token visualization
// ============================================================================

export interface TokenParticle {
  id: string
  type: "input" | "output" | "hidden"
  position: { x: number; y: number; z?: number }
  velocity?: { x: number; y: number; z?: number }
  size: number
  color: string
  alpha?: number
  lifespan?: number
  createdAt: number
}

export interface TokenParticleSystemProps {
  tokens: TokenParticle[]
  width?: number
  height?: number
  showTrails?: boolean
  colorScheme?: "brand" | "rainbow" | "monochrome"
  onParticleClick?: (particle: TokenParticle) => void
  className?: string
}

export function TokenParticleSystem({
  tokens,
  width = 600,
  height = 400,
  showTrails = true,
  colorScheme = "brand",
  onParticleClick,
  className,
}: TokenParticleSystemProps) {
  const canvasRef = React.useRef<HTMLCanvasElement>(null)
  const [particles, setParticles] = React.useState<TokenParticle[]>([])
  const animationRef = React.useRef<number>()

  React.useEffect(() => {
    setParticles(tokens)
  }, [tokens])

  React.useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext("2d")
    if (!ctx) return

    let frameCount = 0

    const animate = () => {
      frameCount++
      ctx.clearRect(0, 0, width, height)

      // Background gradient
      const gradient = ctx.createRadialGradient(width / 2, height / 2, 0, width / 2, height / 2, width / 2)
      gradient.addColorStop(0, "rgba(10, 10, 20, 1)")
      gradient.addColorStop(1, "rgba(5, 5, 15, 1)")
      ctx.fillStyle = gradient
      ctx.fillRect(0, 0, width, height)

      // Update and draw particles
      setParticles((prev) =>
        prev.map((p) => {
          // Move particle based on velocity
          const newPos = {
            x: p.position.x + (p.velocity?.x ?? 0),
            y: p.position.y + (p.velocity?.y ?? 0),
            z: (p.position.z ?? 0) + (p.velocity?.z ?? 0),
          }

          // Bounce off edges
          if (newPos.x < 0 || newPos.x > width) newPos.x = Math.max(0, Math.min(width, newPos.x))
          if (newPos.y < 0 || newPos.y > height) newPos.y = Math.max(0, Math.min(height, newPos.y))

          return { ...p, position: newPos }
        })
      )

      // Draw trails
      if (showTrails) {
        ctx.fillStyle = "rgba(0, 0, 0, 0.1)"
        ctx.fillRect(0, 0, width, height)
      }

      // Draw particles
      particles.forEach((p) => {
        const alpha = p.alpha ?? 1
        const size = p.size * (1 + Math.sin(frameCount * 0.1) * 0.2)

        ctx.beginPath()
        ctx.arc(p.position.x, p.position.y, size, 0, Math.PI * 2)

        const color = p.color || getParticleColor(p.type, colorScheme)
        ctx.fillStyle = color.replace("1)", `${alpha})`)
        ctx.fill()

        // Glow effect
        ctx.shadowColor = color
        ctx.shadowBlur = size * 3
        ctx.fill()
        ctx.shadowBlur = 0
      })

      animationRef.current = requestAnimationFrame(animate)
    }

    animate()

    return () => {
      if (animationRef.current) cancelAnimationFrame(animationRef.current)
    }
  }, [particles, width, height, showTrails, colorScheme])

  return (
    <div className={cn("relative rounded-lg overflow-hidden bg-[#050510]", className)}>
      <canvas
        ref={canvasRef}
        width={width}
        height={height}
        className="w-full h-full cursor-crosshair"
        onClick={(e) => {
          const rect = e.currentTarget.getBoundingClientRect()
          const x = e.clientX - rect.left
          const y = e.clientY - rect.top
          // Find nearest particle
          const nearest = particles.reduce((acc, p) => {
            const dist = Math.hypot(p.position.x - x, p.position.y - y)
            return dist < 30 ? p : acc
          }, null as TokenParticle | null)
          if (nearest) onParticleClick?.(nearest)
        }}
      />
      <div className="absolute bottom-2 right-2 text-[10px] text-text-muted">
        {particles.length} particles
      </div>
    </div>
  )
}

function getParticleColor(type: string, scheme: string): string {
  if (scheme === "rainbow") {
    const hue = type === "input" ? 200 : type === "output" ? 300 : 60
    return `hsla(${hue}, 80%, 60%, 1)`
  }
  if (scheme === "monochrome") {
    return "rgba(255, 255, 255, 1)"
  }
  // brand scheme
  return type === "input" ? "rgba(0, 212, 255, 1)" : type === "output" ? "rgba(255, 107, 53, 1)" : "rgba(91, 124, 245, 1)"
}

// ============================================================================
// InferenceFlowField - Flow field visualization for inference paths
// ============================================================================

export interface FlowVector {
  x: number
  y: number
  angle: number
  magnitude: number
}

export interface InferenceFlowFieldProps {
  width?: number
  height?: number
  density?: number
  autoStream?: boolean
  flowColor?: string
  showArrows?: boolean
  className?: string
}

export function InferenceFlowField({
  width = 600,
  height = 400,
  density = 20,
  autoStream = true,
  flowColor = "#ff6b35",
  showArrows = true,
  className,
}: InferenceFlowFieldProps) {
  const canvasRef = React.useRef<HTMLCanvasElement>(null)
  const [time, setTime] = React.useState(0)

  React.useEffect(() => {
    if (!autoStream) return
    const interval = setInterval(() => setTime((t) => t + 0.02), 16)
    return () => clearInterval(interval)
  }, [autoStream])

  React.useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext("2d")
    if (!ctx) return

    // Clear with dark background
    ctx.fillStyle = "#050510"
    ctx.fillRect(0, 0, width, height)

    const cellSize = Math.max(width, height) / density

    // Draw flow field
    for (let x = 0; x < density; x++) {
      for (let y = 0; y < density; y++) {
        const px = (x + 0.5) * cellSize
        const py = (y + 0.5) * cellSize

        // Create flowing wave pattern
        const wave1 = Math.sin(x * 0.3 + time) * Math.cos(y * 0.3 + time * 0.7)
        const wave2 = Math.sin(x * 0.2 - y * 0.2 + time * 1.3)
        const angle = (wave1 + wave2) * Math.PI
        const magnitude = 0.5 + (wave1 + wave2) * 0.5

        const endX = px + Math.cos(angle) * cellSize * magnitude * 0.4
        const endY = py + Math.sin(angle) * cellSize * magnitude * 0.4

        // Draw arrow
        ctx.beginPath()
        ctx.moveTo(px, py)
        ctx.lineTo(endX, endY)

        const alpha = 0.3 + magnitude * 0.4
        ctx.strokeStyle = flowColor.replace(")", `, ${alpha})`).replace("rgb", "rgba").replace("#", "")

        // Convert hex to rgba if needed
        if (flowColor.startsWith("#")) {
          const r = parseInt(flowColor.slice(1, 3), 16)
          const g = parseInt(flowColor.slice(3, 5), 16)
          const b = parseInt(flowColor.slice(5, 7), 16)
          ctx.strokeStyle = `rgba(${r}, ${g}, ${b}, ${alpha})`
        }

        ctx.lineWidth = 1
        ctx.stroke()

        // Arrow head
        if (showArrows) {
          const arrowSize = 4
          const arrowAngle = Math.atan2(endY - py, endX - px)
          ctx.beginPath()
          ctx.moveTo(endX, endY)
          ctx.lineTo(
            endX - arrowSize * Math.cos(arrowAngle - Math.PI / 6),
            endY - arrowSize * Math.sin(arrowAngle - Math.PI / 6)
          )
          ctx.moveTo(endX, endY)
          ctx.lineTo(
            endX - arrowSize * Math.cos(arrowAngle + Math.PI / 6),
            endY - arrowSize * Math.sin(arrowAngle + Math.PI / 6)
          )
          ctx.stroke()
        }
      }
    }
  }, [time, width, height, density, autoStream, flowColor, showArrows])

  return (
    <div className={cn("relative rounded-lg overflow-hidden bg-[#050510]", className)}>
      <canvas ref={canvasRef} width={width} height={height} className="w-full h-full" />
    </div>
  )
}

// ============================================================================
// CognitiveMesh - Interactive mesh network for cognitive processing
// ============================================================================

export interface MeshNode {
  id: string
  label: string
  position: { x: number; y: number }
  connections: string[]
  activation: number // 0-1
  type: "input" | "hidden" | "output" | "attention"
}

export interface CognitiveMeshProps {
  nodes: MeshNode[]
  width?: number
  height?: number
  onNodeClick?: (node: MeshNode) => void
  selectedNodeId?: string
  animated?: boolean
  className?: string
}

export function CognitiveMesh({
  nodes,
  width = 600,
  height = 400,
  onNodeClick,
  selectedNodeId,
  animated = true,
  className,
}: CognitiveMeshProps) {
  const canvasRef = React.useRef<HTMLCanvasElement>(null)
  const [frame, setFrame] = React.useState(0)

  React.useEffect(() => {
    if (!animated) return
    const interval = setInterval(() => setFrame((f) => f + 1), 50)
    return () => clearInterval(interval)
  }, [animated])

  React.useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext("2d")
    if (!ctx) return

    // Background with subtle grid
    ctx.fillStyle = "#08080c"
    ctx.fillRect(0, 0, width, height)

    // Draw grid
    ctx.strokeStyle = "rgba(255, 255, 255, 0.03)"
    ctx.lineWidth = 1
    for (let x = 0; x < width; x += 30) {
      ctx.beginPath()
      ctx.moveTo(x, 0)
      ctx.lineTo(x, height)
      ctx.stroke()
    }
    for (let y = 0; y < height; y += 30) {
      ctx.beginPath()
      ctx.moveTo(0, y)
      ctx.lineTo(width, y)
      ctx.stroke()
    }

    // Draw connections
    const nodeMap = new Map(nodes.map((n) => [n.id, n]))
    nodes.forEach((node) => {
      node.connections.forEach((connId) => {
        const conn = nodeMap.get(connId)
        if (!conn) return

        const isActive = node.activation > 0.5 && conn.activation > 0.5
        ctx.beginPath()
        ctx.moveTo(node.position.x, node.position.y)
        ctx.lineTo(conn.position.x, conn.position.y)
        ctx.strokeStyle = isActive ? "rgba(255, 107, 53, 0.6)" : "rgba(255, 255, 255, 0.1)"
        ctx.lineWidth = isActive ? 2 : 1
        ctx.stroke()

        // Animated pulse on active connections
        if (isActive && animated) {
          const pulsePos = (frame % 30) / 30
          const px = node.position.x + (conn.position.x - node.position.x) * pulsePos
          const py = node.position.y + (conn.position.y - node.position.y) * pulsePos
          ctx.beginPath()
          ctx.arc(px, py, 3, 0, Math.PI * 2)
          ctx.fillStyle = "rgba(255, 107, 53, 0.8)"
          ctx.fill()
        }
      })
    })

    // Draw nodes
    nodes.forEach((node) => {
      const isSelected = node.id === selectedNodeId
      const nodeColor = node.type === "input" ? "#00d4ff" : node.type === "output" ? "#ff6b35" : node.type === "attention" ? "#ffb800" : "#5b7cf5"
      const pulseScale = animated ? 1 + Math.sin(frame * 0.2 + nodes.indexOf(node) * 0.5) * 0.1 : 1
      const size = (8 + node.activation * 8) * pulseScale

      // Glow
      ctx.shadowColor = nodeColor
      ctx.shadowBlur = node.activation * 20

      ctx.beginPath()
      ctx.arc(node.position.x, node.position.y, size, 0, Math.PI * 2)
      ctx.fillStyle = nodeColor
      ctx.fill()

      if (isSelected) {
        ctx.strokeStyle = "#ffffff"
        ctx.lineWidth = 2
        ctx.stroke()
      }

      ctx.shadowBlur = 0

      // Label
      ctx.fillStyle = "rgba(255, 255, 255, 0.8)"
      ctx.font = "10px sans-serif"
      ctx.textAlign = "center"
      ctx.fillText(node.label, node.position.x, node.position.y - size - 5)
    })
  }, [nodes, frame, width, height, selectedNodeId, animated])

  return (
    <div className={cn("relative rounded-lg overflow-hidden bg-[#08080c]", className)}>
      <canvas
        ref={canvasRef}
        width={width}
        height={height}
        className="w-full h-full cursor-pointer"
        onClick={(e) => {
          const rect = e.currentTarget.getBoundingClientRect()
          const x = e.clientX - rect.left
          const y = e.clientY - rect.top
          const nearest = nodes.reduce((acc, n) => {
            const dist = Math.hypot(n.position.x - x, n.position.y - y)
            return dist < 30 && dist < (acc ? Math.hypot(acc.position.x - x, acc.position.y - y) : Infinity) ? n : acc
          }, null as MeshNode | null)
          if (nearest) onNodeClick?.(nearest)
        }}
      />
    </div>
  )
}

// ============================================================================
// RuntimeGalaxyView - 3D galaxy visualization for runtime clusters
// ============================================================================

export interface GalaxyCluster {
  id: string
  name: string
  position: [number, number, number]
  size: number
  nodes: Array<{
    id: string
    status: "healthy" | "degraded" | "down"
    connections: number
  }>
}

export interface RuntimeGalaxyViewProps {
  clusters: GalaxyCluster[]
  showLabels?: boolean
  autoRotate?: boolean
  onClusterClick?: (cluster: GalaxyCluster) => void
  className?: string
}

export function RuntimeGalaxyView({
  clusters,
  showLabels = true,
  autoRotate = true,
  onClusterClick,
  className,
}: RuntimeGalaxyViewProps) {
  const groupRef = React.useRef<any>(null)
  const [rotation, setRotation] = React.useState(0)

  React.useEffect(() => {
    if (!autoRotate) return
    const interval = setInterval(() => setRotation((r) => r + 0.002), 16)
    return () => clearInterval(interval)
  }, [autoRotate])

  React.useEffect(() => {
    if (groupRef.current && autoRotate) {
      groupRef.current.rotation.y = rotation
    }
  }, [rotation, autoRotate])

  return (
    <div className={cn("relative w-full h-96 rounded-lg overflow-hidden bg-[#050510]", className)}>
      {/* Simulated 3D canvas with CSS */}
      <div className="absolute inset-0 flex items-center justify-center">
        <div ref={groupRef} className="relative" style={{ transformStyle: "preserve-3d", perspective: "1000px" }}>
          {clusters.map((cluster, i) => {
            const [x, y, z] = cluster.position
            const scale = 1 - Math.abs(z) / 500
            const statusColors = {
              healthy: "#10b981",
              degraded: "#f59e0b",
              down: "#ef4444",
            }

            // Calculate cluster color based on node health
            const healthyCount = cluster.nodes.filter((n) => n.status === "healthy").length
            const clusterHealth = healthyCount / cluster.nodes.length

            return (
              <div
                key={cluster.id}
                className="absolute transition-all duration-300 cursor-pointer group"
                style={{
                  left: `50%`,
                  top: `50%`,
                  transform: `
                    translateX(${(x - 250) * scale}px)
                    translateY(${(y - 200) * scale}px)
                    translateZ(${z}px)
                    scale(${scale})
                  `,
                }}
                onClick={() => onClusterClick?.(cluster)}
              >
                {/* Cluster core */}
                <div
                  className="rounded-full border-2 border-white/20 transition-all group-hover:border-white/50"
                  style={{
                    width: `${cluster.size * scale}px`,
                    height: `${cluster.size * scale}px`,
                    background: `radial-gradient(circle, ${statusColors[cluster.nodes[0]?.status || "healthy"]}40 0%, transparent 70%)`,
                    boxShadow: `0 0 ${20 * scale}px ${statusColors[cluster.nodes[0]?.status || "healthy"]}40`,
                  }}
                >
                  {/* Nodes in cluster */}
                  <div className="absolute inset-0 flex items-center justify-center">
                    {cluster.nodes.slice(0, 8).map((node, j) => {
                      const angle = (j / cluster.nodes.length) * Math.PI * 2
                      const nodeSize = 4 + node.connections * 0.5
                      return (
                        <div
                          key={node.id}
                          className="absolute rounded-full transition-all"
                          style={{
                            width: `${nodeSize}px`,
                            height: `${nodeSize}px`,
                            background: statusColors[node.status],
                            boxShadow: `0 0 10px ${statusColors[node.status]}`,
                            transform: `translate(${Math.cos(angle) * 15}px, ${Math.sin(angle) * 15}px)`,
                          }}
                        />
                      )
                    })}
                  </div>
                </div>

                {/* Label */}
                {showLabels && (
                  <div className="absolute top-full left-1/2 -translate-x-1/2 mt-2 text-[10px] text-white/60 whitespace-nowrap opacity-0 group-hover:opacity-100 transition-opacity">
                    {cluster.name}
                  </div>
                )}
              </div>
            )
          })}
        </div>

        {/* Ambient stars */}
        <div className="absolute inset-0 overflow-hidden pointer-events-none">
          {Array.from({ length: 50 }, (_, i) => (
            <div
              key={i}
              className="absolute rounded-full bg-white"
              style={{
                width: `${1 + Math.random()}px`,
                height: `${1 + Math.random()}px`,
                left: `${Math.random() * 100}%`,
                top: `${Math.random() * 100}%`,
                opacity: 0.3 + Math.random() * 0.4,
              }}
            />
          ))}
        </div>
      </div>

      {/* Stats overlay */}
      <div className="absolute bottom-2 left-2 text-[10px] text-white/40">
        {clusters.length} clusters · {clusters.reduce((sum, c) => sum + c.nodes.length, 0)} nodes
      </div>
    </div>
  )
}

// ============================================================================
// AgentConstellation - Star map for agent relationships
// ============================================================================

export interface ConstellationAgent {
  id: string
  name: string
  x: number
  y: number
  size: number
  connections: string[]
  status: "active" | "idle" | "error"
  role?: string
}

export interface AgentConstellationProps {
  agents: ConstellationAgent[]
  onAgentClick?: (agent: ConstellationAgent) => void
  selectedAgentId?: string
  showRoles?: boolean
  className?: string
}

export function AgentConstellation({
  agents,
  onAgentClick,
  selectedAgentId,
  showRoles = false,
  className,
}: AgentConstellationProps) {
  const [hoveredId, setHoveredId] = React.useState<string | null>(null)
  const [time, setTime] = React.useState(0)

  React.useEffect(() => {
    const interval = setInterval(() => setTime((t) => t + 1), 50)
    return () => clearInterval(interval)
  }, [])

  const statusColors = {
    active: "#10b981",
    idle: "#6b7280",
    error: "#ef4444",
  }

  return (
    <div className={cn("relative w-full h-96 rounded-lg overflow-hidden bg-[#050510]", className)}>
      <svg className="w-full h-full">
        {/* Constellation lines */}
        {agents.map((agent) =>
          agent.connections.map((connId) => {
            const conn = agents.find((a) => a.id === connId)
            if (!conn) return null
            const isHighlighted = hoveredId === agent.id || hoveredId === connId
            return (
              <line
                key={`${agent.id}-${connId}`}
                x1={agent.x}
                y1={agent.y}
                x2={conn.x}
                y2={conn.y}
                stroke={isHighlighted ? "rgba(255,255,255,0.4)" : "rgba(255,255,255,0.1)"}
                strokeWidth={isHighlighted ? 1.5 : 0.5}
              />
            )
          })
        )}

        {/* Agent nodes */}
        {agents.map((agent) => {
          const isSelected = agent.id === selectedAgentId
          const isHovered = agent.id === hoveredId
          const color = statusColors[agent.status]
          const pulseSize = agent.status === "active" ? Math.sin(time * 0.1) * 2 : 0

          return (
            <g
              key={agent.id}
              className="cursor-pointer"
              onClick={() => onAgentClick?.(agent)}
              onMouseEnter={() => setHoveredId(agent.id)}
              onMouseLeave={() => setHoveredId(null)}
            >
              {/* Glow */}
              <circle cx={agent.x} cy={agent.y} r={agent.size + 10 + pulseSize} fill={color} opacity={0.2} />

              {/* Main circle */}
              <circle
                cx={agent.x}
                cy={agent.y}
                r={agent.size + pulseSize}
                fill={color}
                stroke={isSelected ? "#ffffff" : "transparent"}
                strokeWidth={2}
              />

              {/* Inner highlight */}
              <circle cx={agent.x - agent.size * 0.3} cy={agent.y - agent.size * 0.3} r={agent.size * 0.3} fill="rgba(255,255,255,0.3)" />

              {/* Label */}
              <text
                x={agent.x}
                y={agent.y + agent.size + 15}
                textAnchor="middle"
                fill="rgba(255,255,255,0.7)"
                fontSize="11"
              >
                {agent.name}
              </text>

              {/* Role badge */}
              {showRoles && agent.role && (
                <text
                  x={agent.x}
                  y={agent.y + agent.size + 27}
                  textAnchor="middle"
                  fill="rgba(255,255,255,0.4)"
                  fontSize="9"
                >
                  {agent.role}
                </text>
              )}
            </g>
          )
        })}
      </svg>

      {/* Legend */}
      <div className="absolute bottom-2 right-2 flex items-center gap-3 text-[10px]">
        <span className="flex items-center gap-1">
          <span className="size-2 rounded-full bg-[#10b981]" /> Active
        </span>
        <span className="flex items-center gap-1">
          <span className="size-2 rounded-full bg-[#6b7280]" /> Idle
        </span>
        <span className="flex items-center gap-1">
          <span className="size-2 rounded-full bg-[#ef4444]" /> Error
        </span>
      </div>
    </div>
  )
}

// ============================================================================
// ExecutionPulseMap - Pulsing visualization of execution flow
// ============================================================================

export interface PulseEvent {
  id: string
  timestamp: number
  source: string
  target: string
  type: "start" | "complete" | "error"
  duration?: number
}

export interface ExecutionPulseMapProps {
  events: PulseEvent[]
  width?: number
  height?: number
  onEventClick?: (event: PulseEvent) => void
  className?: string
}

export function ExecutionPulseMap({
  events,
  width = 600,
  height = 400,
  onEventClick,
  className,
}: ExecutionPulseMapProps) {
  const [time, setTime] = React.useState(Date.now())
  const canvasRef = React.useRef<HTMLCanvasElement>(null)

  React.useEffect(() => {
    const interval = setInterval(() => setTime(Date.now()), 50)
    return () => clearInterval(interval)
  }, [])

  React.useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext("2d")
    if (!ctx) return

    ctx.fillStyle = "#08080c"
    ctx.fillRect(0, 0, width, height)

    // Calculate node positions along a sine wave path
    const nodes = Array.from(new Set(events.flatMap((e) => [e.source, e.target])))

    // Draw sine wave path
    ctx.beginPath()
    ctx.moveTo(50, height / 2)
    for (let x = 50; x < width - 50; x += 5) {
      const y = height / 2 + Math.sin((x / width) * Math.PI * 4) * (height / 4)
      ctx.lineTo(x, y)
    }
    ctx.strokeStyle = "rgba(255, 255, 255, 0.1)"
    ctx.lineWidth = 2
    ctx.stroke()

    // Node positions
    const nodePositions = new Map<string, { x: number; y: number }>()
    nodes.forEach((node, i) => {
      const x = 50 + (i / Math.max(nodes.length - 1, 1)) * (width - 100)
      const y = height / 2 + Math.sin((x / width) * Math.PI * 4) * (height / 4)
      nodePositions.set(node, { x, y })
    })

    // Draw pulses
    events.forEach((event) => {
      const sourcePos = nodePositions.get(event.source)
      const targetPos = nodePositions.get(event.target)
      if (!sourcePos || !targetPos) return

      const elapsed = time - event.timestamp
      const speed = 0.003
      const progress = (elapsed * speed) % 1
      const pulseX = sourcePos.x + (targetPos.x - sourcePos.x) * progress
      const pulseY = sourcePos.y + (targetPos.y - sourcePos.y) * progress

      const color = event.type === "start" ? "#00d4ff" : event.type === "complete" ? "#10b981" : "#ef4444"

      // Draw pulse
      ctx.beginPath()
      ctx.arc(pulseX, pulseY, 6, 0, Math.PI * 2)
      ctx.fillStyle = color
      ctx.fill()

      // Trail
      ctx.beginPath()
      ctx.moveTo(pulseX, pulseY)
      ctx.lineTo(sourcePos.x, sourcePos.y)
      ctx.strokeStyle = color
      ctx.lineWidth = 2
      ctx.globalAlpha = 0.3
      ctx.stroke()
      ctx.globalAlpha = 1
    })

    // Draw nodes
    nodes.forEach((node) => {
      const pos = nodePositions.get(node)!
      ctx.beginPath()
      ctx.arc(pos.x, pos.y, 8, 0, Math.PI * 2)
      ctx.fillStyle = "#1a1a28"
      ctx.fill()
      ctx.strokeStyle = "rgba(255, 255, 255, 0.3)"
      ctx.lineWidth = 1
      ctx.stroke()
      ctx.fillStyle = "rgba(255, 255, 255, 0.6)"
      ctx.font = "9px sans-serif"
      ctx.textAlign = "center"
      ctx.fillText(node.length > 8 ? node.slice(0, 8) + "..." : node, pos.x, pos.y + 20)
    })
  }, [events, time, width, height])

  return (
    <div className={cn("relative rounded-lg overflow-hidden bg-[#08080c]", className)}>
      <canvas
        ref={canvasRef}
        width={width}
        height={height}
        className="w-full h-full cursor-pointer"
        onClick={(e) => {
          const rect = e.currentTarget.getBoundingClientRect()
          const x = e.clientX - rect.left
          const y = e.clientY - rect.top
          onEventClick?.(events[events.length - 1])
        }}
      />
    </div>
  )
}

// ============================================================================
// GlobalInfrastructureMap - World map showing global infrastructure
// ============================================================================

export interface InfrastructureRegion {
  id: string
  name: string
  position: "us-east" | "us-west" | "eu-west" | "asia-pacific" | "south-america"
  load: number // 0-1
  nodes: number
  status: "healthy" | "degraded" | "down"
}

export interface GlobalInfrastructureMapProps {
  regions: InfrastructureRegion[]
  onRegionClick?: (region: InfrastructureRegion) => void
  selectedRegionId?: string
  className?: string
}

export function GlobalInfrastructureMap({
  regions,
  onRegionClick,
  selectedRegionId,
  className,
}: GlobalInfrastructureMapProps) {
  const regionPositions: Record<string, { x: number; y: number }> = {
    "us-east": { x: 25, y: 35 },
    "us-west": { x: 15, y: 35 },
    "eu-west": { x: 48, y: 30 },
    "asia-pacific": { x: 75, y: 40 },
    "south-america": { x: 30, y: 60 },
  }

  const statusColors = {
    healthy: "#10b981",
    degraded: "#f59e0b",
    down: "#ef4444",
  }

  return (
    <div className={cn("relative w-full h-80 rounded-lg overflow-hidden bg-[#050510] border border-white/10", className)}>
      {/* Map background (simplified world outline) */}
      <svg className="w-full h-full" viewBox="0 0 100 100" preserveAspectRatio="xMidYMid meet">
        {/* Simplified continents */}
        <path
          d="M15,25 Q20,20 35,25 Q40,35 35,45 Q25,50 20,45 Q15,35 15,25"
          fill="rgba(255,255,255,0.05)"
          stroke="rgba(255,255,255,0.1)"
        />
        <path
          d="M45,20 Q60,15 75,25 Q80,35 70,40 Q55,45 45,35 Q40,25 45,20"
          fill="rgba(255,255,255,0.05)"
          stroke="rgba(255,255,255,0.1)"
        />
        <path
          d="M50,50 Q60,45 70,55 Q65,70 55,70 Q45,60 50,50"
          fill="rgba(255,255,255,0.05)"
          stroke="rgba(255,255,255,0.1)"
        />

        {/* Connection lines between regions */}
        {regions.map((region) => {
          const otherRegions = regions.filter((r) => r.id !== region.id)
          return otherRegions.slice(0, 2).map((other) => {
            const from = regionPositions[region.position]
            const to = regionPositions[other.position]
            if (!from || !to) return null
            return (
              <line
                key={`${region.id}-${other.id}`}
                x1={from.x}
                y1={from.y}
                x2={to.x}
                y2={to.y}
                stroke="rgba(255,255,255,0.1)"
                strokeWidth={0.5}
                strokeDasharray="2,2"
              />
            )
          })
        })}

        {/* Region nodes */}
        {regions.map((region) => {
          const pos = regionPositions[region.position]
          if (!pos) return null
          const isSelected = region.id === selectedRegionId
          const color = statusColors[region.status]
          const size = 3 + region.load * 5

          return (
            <g
              key={region.id}
              className="cursor-pointer"
              onClick={() => onRegionClick?.(region)}
            >
              {/* Glow */}
              <circle cx={pos.x} cy={pos.y} r={size * 3} fill={color} opacity={0.15} />

              {/* Main node */}
              <circle
                cx={pos.x}
                cy={pos.y}
                r={size}
                fill={color}
                stroke={isSelected ? "#ffffff" : "transparent"}
                strokeWidth={0.5}
              />

              {/* Inner pulse for active regions */}
              {region.status === "healthy" && (
                <circle cx={pos.x} cy={pos.y} r={size * 0.5} fill="rgba(255,255,255,0.5)">
                  <animate attributeName="r" values={`${size * 0.5};${size}`} dur="2s" repeatCount="indefinite" />
                  <animate attributeName="opacity" values={"0.5;0"} dur="2s" repeatCount="indefinite" />
                </circle>
              )}

              {/* Label */}
              <text x={pos.x} y={pos.y + size + 6} textAnchor="middle" fill="rgba(255,255,255,0.6)" fontSize="3">
                {region.name}
              </text>
              <text x={pos.x} y={pos.y + size + 10} textAnchor="middle" fill="rgba(255,255,255,0.4)" fontSize="2">
                {region.nodes} nodes
              </text>
            </g>
          )
        })}
      </svg>

      {/* Legend */}
      <div className="absolute bottom-2 right-2 flex items-center gap-2 text-[9px]">
        <span className="flex items-center gap-1">
          <span className="size-2 rounded-full bg-[#10b981]" /> Healthy
        </span>
        <span className="flex items-center gap-1">
          <span className="size-2 rounded-full bg-[#f59e0b]" /> Degraded
        </span>
        <span className="flex items-center gap-1">
          <span className="size-2 rounded-full bg-[#ef4444]" /> Down
        </span>
      </div>
    </div>
  )
}

// ============================================================================
// DataFlowRiver - Flowing river visualization for data streams
// ============================================================================

export interface DataStream {
  id: string
  label: string
  flow: number // positive = right, negative = left
  packets: number
  bandwidth: number
  type: "request" | "response" | "event"
}

export interface DataFlowRiverProps {
  streams: DataStream[]
  width?: number
  height?: number
  direction?: "horizontal" | "vertical"
  onStreamClick?: (stream: DataStream) => void
  className?: string
}

export function DataFlowRiver({
  streams,
  width = 600,
  height = 300,
  direction = "horizontal",
  onStreamClick,
  className,
}: DataFlowRiverProps) {
  const [time, setTime] = React.useState(0)

  React.useEffect(() => {
    const interval = setInterval(() => setTime((t) => t + 1), 30)
    return () => clearInterval(interval)
  }, [])

  const typeColors = {
    request: "#00d4ff",
    response: "#10b981",
    event: "#ff6b35",
  }

  return (
    <div className={cn("relative rounded-lg overflow-hidden bg-[#050510]", className)}>
      <svg width={width} height={height} className="w-full h-full">
        {/* River bed */}
        <rect x={0} y={0} width={width} height={height} fill="rgba(10, 10, 20, 0.5)" />

        {/* Flow lines */}
        {streams.map((stream, i) => {
          const isHorizontal = direction === "horizontal"
          const laneHeight = height / streams.length
          const y = i * laneHeight + laneHeight / 2
          const xOffset = isHorizontal ? 0 : y
          const yOffset = isHorizontal ? y : 0

          return (
            <g key={stream.id} className="cursor-pointer" onClick={() => onStreamClick?.(stream)}>
              {/* Lane background */}
              <rect
                x={xOffset}
                y={yOffset - laneHeight / 2 + 2}
                width={isHorizontal ? width : laneHeight - 4}
                height={laneHeight - 4}
                fill="rgba(255,255,255,0.02)"
              />

              {/* Flow direction indicator */}
              <text
                x={isHorizontal ? (stream.flow > 0 ? width - 10 : 10) : xOffset + (stream.flow > 0 ? width - 20 : 20)}
                y={y + 3}
                textAnchor="middle"
                fill="rgba(255,255,255,0.3)"
                fontSize="10"
              >
                {stream.flow > 0 ? "→" : "←"}
              </text>

              {/* Animated packets */}
              {Array.from({ length: Math.min(stream.packets, 20) }, (_, j) => {
                const packetProgress = ((time * stream.flow * 0.5 + j * 50) % (width || 100)) / (width || 100)
                const packetX = isHorizontal ? packetProgress * width : 50
                const packetY = isHorizontal ? y : packetProgress * height

                return (
                  <rect
                    key={j}
                    x={packetX}
                    y={packetY - 2}
                    width={stream.flow > 0 ? 8 : -8}
                    height={4}
                    rx={2}
                    fill={typeColors[stream.type]}
                    opacity={0.6 + Math.random() * 0.4}
                  >
                    <animate attributeName="x" from={stream.flow > 0 ? 0 : width} to={stream.flow > 0 ? width : 0} dur={`${2 + Math.random()}s`} repeatCount="indefinite" />
                  </rect>
                )
              })}

              {/* Label */}
              <text
                x={isHorizontal ? 60 : 10}
                y={y + 3}
                fill="rgba(255,255,255,0.6)"
                fontSize="11"
              >
                {stream.label}
              </text>

              {/* Stats */}
              <text
                x={isHorizontal ? width - 60 : width - 40}
                y={y + 3}
                textAnchor="end"
                fill="rgba(255,255,255,0.4)"
                fontSize="9"
              >
                {stream.bandwidth.toFixed(1)} MB/s
              </text>
            </g>
          )
        })}
      </svg>
    </div>
  )
}

// ============================================================================
// SemanticHeatmap - Heat map for semantic clustering
// ============================================================================

export interface SemanticCell {
  row: string
  col: string
  value: number
  label?: string
}

export interface SemanticHeatmapProps {
  cells: SemanticCell[]
  width?: number
  height?: number
  colorScheme?: "brand" | "cool" | "warm" | "viridis"
  onCellClick?: (cell: SemanticCell) => void
  className?: string
}

export function SemanticHeatmap({
  cells,
  width = 600,
  height = 400,
  colorScheme = "brand",
  onCellClick,
  className,
}: SemanticHeatmapProps) {
  const rows = Array.from(new Set(cells.map((c) => c.row)))
  const cols = Array.from(new Set(cells.map((c) => c.col)))
  const maxValue = Math.max(...cells.map((c) => c.value), 1)

  const getColor = (value: number) => {
    const t = value / maxValue
    switch (colorScheme) {
      case "cool":
        return `rgba(0, 212, 255, ${0.1 + t * 0.7})`
      case "warm":
        return `rgba(255, 107, 53, ${0.1 + t * 0.7})`
      case "viridis":
        const r = 0.13 + t * 0.73
        const g = 0.26 + t * 0.36
        const b = 0.33 + t * 0.57
        return `rgba(${Math.floor(r * 255)}, ${Math.floor(g * 255)}, ${Math.floor(b * 255)}, ${0.2 + t * 0.6})`
      default: // brand
        return `rgba(255, 107, 53, ${0.1 + t * 0.7})`
    }
  }

  const cellWidth = (width - 80) / cols.length
  const cellHeight = (height - 60) / rows.length

  return (
    <div className={cn("relative rounded-lg overflow-hidden bg-[#08080c]", className)}>
      <svg width={width} height={height} className="w-full h-full">
        {/* Cells */}
        {cells.map((cell, i) => {
          const rowIdx = rows.indexOf(cell.row)
          const colIdx = cols.indexOf(cell.col)
          const x = 80 + colIdx * cellWidth
          const y = 30 + rowIdx * cellHeight

          return (
            <g key={i} className="cursor-pointer" onClick={() => onCellClick?.(cell)}>
              <rect
                x={x}
                y={y}
                width={cellWidth - 1}
                height={cellHeight - 1}
                fill={getColor(cell.value)}
                stroke="rgba(255,255,255,0.05)"
                strokeWidth={0.5}
              >
                <title>
                  {cell.row} × {cell.col}: {cell.value.toFixed(2)}
                </title>
              </rect>
            </g>
          )
        })}

        {/* Row labels */}
        {rows.map((row, i) => (
          <text
            key={row}
            x={75}
            y={30 + i * cellHeight + cellHeight / 2 + 4}
            textAnchor="end"
            fill="rgba(255,255,255,0.6)"
            fontSize="10"
          >
            {row.length > 8 ? row.slice(0, 8) + "..." : row}
          </text>
        ))}

        {/* Column labels */}
        {cols.map((col, i) => (
          <text
            key={col}
            x={80 + i * cellWidth + cellWidth / 2}
            y={25}
            textAnchor="middle"
            fill="rgba(255,255,255,0.6)"
            fontSize="10"
            transform={`rotate(-45, ${80 + i * cellWidth + cellWidth / 2}, 20)`}
          >
            {col.length > 6 ? col.slice(0, 6) + "..." : col}
          </text>
        ))}
      </svg>

      {/* Color scale */}
      <div className="absolute bottom-2 right-2 flex items-center gap-1">
        <span className="text-[9px] text-text-muted">0</span>
        <div className="w-24 h-3 rounded" style={{
          background: `linear-gradient(to right, ${getColor(0)}, ${getColor(maxValue / 2)}, ${getColor(maxValue)})`
        }} />
        <span className="text-[9px] text-text-muted">{maxValue.toFixed(1)}</span>
      </div>
    </div>
  )
}

// ============================================================================
// AIReasoningTree - Tree visualization for AI reasoning steps
// ============================================================================

export interface ReasoningNode {
  id: string
  label: string
  type: "premise" | "inference" | "conclusion" | "assumption"
  children?: ReasoningNode[]
  confidence?: number
}

export interface AIReasoningTreeProps {
  root: ReasoningNode
  onNodeClick?: (node: ReasoningNode) => void
  selectedNodeId?: string
  direction?: "horizontal" | "vertical"
  className?: string
}

export function AIReasoningTree({
  root,
  onNodeClick,
  selectedNodeId,
  direction = "horizontal",
  className,
}: AIReasoningTreeProps) {
  const typeColors = {
    premise: "#00d4ff",
    inference: "#ff6b35",
    conclusion: "#10b981",
    assumption: "#f59e0b",
  }

  const renderNode = (node: ReasoningNode, depth = 0): JSX.Element => {
    const isSelected = node.id === selectedNodeId
    const color = typeColors[node.type]
    const hasChildren = node.children && node.children.length > 0

    const nodeWidth = direction === "horizontal" ? 120 : 150
    const nodeHeight = direction === "horizontal" ? 50 : 40
    const gap = direction === "horizontal" ? 60 : 40

    const x = direction === "horizontal" ? depth * (nodeWidth + gap) : 0
    const y = direction === "horizontal" ? 0 : depth * (nodeHeight + gap)

    return (
      <g key={node.id} onClick={() => onNodeClick?.(node)} className="cursor-pointer">
        {/* Connection line */}
        {depth > 0 && (
          <line
            x1={direction === "horizontal" ? x - gap : x + nodeWidth / 2}
            y1={direction === "horizontal" ? y + nodeHeight / 2 : y - gap}
            x2={direction === "horizontal" ? x : x + nodeWidth / 2}
            y2={direction === "horizontal" ? y + nodeHeight / 2 : y}
            stroke="rgba(255,255,255,0.2)"
            strokeWidth={1}
          />
        )}

        {/* Node background */}
        <rect
          x={x}
          y={y}
          width={nodeWidth}
          height={nodeHeight}
          rx={6}
          fill={color + "20"}
          stroke={isSelected ? "#ffffff" : color}
          strokeWidth={isSelected ? 2 : 1}
        />

        {/* Confidence indicator */}
        {node.confidence !== undefined && (
          <rect
            x={x + 2}
            y={y + nodeHeight - 6}
            width={(nodeWidth - 4) * node.confidence}
            height={3}
            rx={1.5}
            fill={color}
            opacity={0.6}
          />
        )}

        {/* Node label */}
        <text
          x={x + nodeWidth / 2}
          y={y + nodeHeight / 2 + 4}
          textAnchor="middle"
          fill="rgba(255,255,255,0.9)"
          fontSize="11"
        >
          {node.label.length > 15 ? node.label.slice(0, 15) + "..." : node.label}
        </text>

        {/* Children */}
        {hasChildren && direction === "horizontal" && (
          <g transform={`translate(${x + nodeWidth + gap / 2}, ${y})`}>
            {node.children!.map((child, i) => (
              <g key={child.id} transform={`translate(0, ${(i - (node.children!.length - 1) / 2) * 80})`}>
                {renderNode(child, 0)}
              </g>
            ))}
          </g>
        )}

        {hasChildren && direction === "vertical" && (
          <g transform={`translate(${x}, ${y + nodeHeight + gap / 2})`}>
            {node.children!.map((child, i) => (
              <g key={child.id}>{renderNode(child, 0)}</g>
            ))}
          </g>
        )}
      </g>
    )
  }

  return (
    <div className={cn("relative w-full h-96 rounded-lg overflow-auto bg-[#08080c]", className)}>
      <svg className="w-full h-full min-w-[500px]">
        <defs>
          <marker id="arrowhead" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
            <polygon points="0 0, 10 3.5, 0 7" fill="rgba(255,255,255,0.3)" />
          </marker>
        </defs>
        {renderNode(root)}
      </svg>

      {/* Legend */}
      <div className="absolute bottom-2 left-2 flex items-center gap-3 text-[9px]">
        {Object.entries(typeColors).map(([type, color]) => (
          <span key={type} className="flex items-center gap-1">
            <span className="size-2 rounded" style={{ background: color }} />
            <span className="capitalize text-text-muted">{type}</span>
          </span>
        ))}
      </div>
    </div>
  )
}

// ============================================================================
// LiveDependencyNebula - Nebula-style dependency visualization
// ============================================================================

export interface NebulaNode {
  id: string
  name: string
  dependencies: string[]
  health: number // 0-1
  critical: boolean
}

export interface LiveDependencyNebulaProps {
  nodes: NebulaNode[]
  onNodeClick?: (node: NebulaNode) => void
  selectedNodeId?: string
  className?: string
}

export function LiveDependencyNebula({
  nodes,
  onNodeClick,
  selectedNodeId,
  className,
}: LiveDependencyNebulaProps) {
  const [time, setTime] = React.useState(0)
  const containerRef = React.useRef<HTMLDivElement>(null)
  const [dimensions, setDimensions] = React.useState({ width: 600, height: 400 })

  React.useEffect(() => {
    if (containerRef.current) {
      const { width, height } = containerRef.current.getBoundingClientRect()
      setDimensions({ width, height })
    }
  }, [])

  React.useEffect(() => {
    const interval = setInterval(() => setTime((t) => t + 1), 50)
    return () => clearInterval(interval)
  }, [])

  const nodePositions = React.useMemo(() => {
    const positions = new Map<string, { x: number; y: number }>()
    const centerX = dimensions.width / 2
    const centerY = dimensions.height / 2
    const radius = Math.min(centerX, centerY) * 0.6

    nodes.forEach((node, i) => {
      const angle = (i / nodes.length) * Math.PI * 2
      const healthRadius = radius * (0.5 + node.health * 0.5)
      positions.set(node.id, {
        x: centerX + Math.cos(angle + time * 0.01) * healthRadius,
        y: centerY + Math.sin(angle + time * 0.01) * healthRadius,
      })
    })
    return positions
  }, [nodes, dimensions, time])

  return (
    <div ref={containerRef} className={cn("relative w-full h-96 rounded-lg overflow-hidden bg-[#050510]", className)}>
      {/* Ambient stars */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        {Array.from({ length: 30 }, (_, i) => (
          <div
            key={i}
            className="absolute rounded-full bg-white"
            style={{
              width: `${1 + Math.random()}px`,
              height: `${1 + Math.random()}px`,
              left: `${Math.random() * 100}%`,
              top: `${Math.random() * 100}%`,
              opacity: 0.2 + Math.random() * 0.3,
            }}
          />
        ))}
      </div>

      <svg className="w-full h-full">
        {/* Dependency lines */}
        {nodes.map((node) => {
          const pos = nodePositions.get(node.id)
          if (!pos) return null

          return node.dependencies.map((depId) => {
            const depPos = nodePositions.get(depId)
            if (!depPos) return null

            const dep = nodes.find((n) => n.id === depId)
            const isCritical = node.critical && dep?.critical

            return (
              <line
                key={`${node.id}-${depId}`}
                x1={pos.x}
                y1={pos.y}
                x2={depPos.x}
                y2={depPos.y}
                stroke={isCritical ? "rgba(239, 68, 68, 0.4)" : "rgba(255, 255, 255, 0.1)"}
                strokeWidth={isCritical ? 2 : 1}
              />
            )
          })
        })}

        {/* Nodes */}
        {nodes.map((node) => {
          const pos = nodePositions.get(node.id)
          if (!pos) return null
          const isSelected = node.id === selectedNodeId
          const size = 6 + node.health * 10
          const color = node.critical ? "#ef4444" : node.health > 0.7 ? "#10b981" : node.health > 0.4 ? "#f59e0b" : "#6b7280"

          return (
            <g
              key={node.id}
              className="cursor-pointer"
              onClick={() => onNodeClick?.(node)}
            >
              {/* Outer glow */}
              <circle cx={pos.x} cy={pos.y} r={size + 8 + Math.sin(time * 0.2) * 3} fill={color} opacity={0.1} />

              {/* Main node */}
              <circle
                cx={pos.x}
                cy={pos.y}
                r={size}
                fill={color}
                stroke={isSelected ? "#ffffff" : "transparent"}
                strokeWidth={2}
              />

              {/* Label */}
              <text
                x={pos.x}
                y={pos.y + size + 12}
                textAnchor="middle"
                fill="rgba(255,255,255,0.6)"
                fontSize="9"
              >
                {node.name}
              </text>
            </g>
          )
        })}
      </svg>

      {/* Stats */}
      <div className="absolute bottom-2 right-2 text-[10px] text-white/40">
        {nodes.length} nodes · {nodes.filter((n) => n.critical).length} critical
      </div>
    </div>
  )
}

// ============================================================================
// RealtimeTopologyGraph - Real-time topology visualization
// ============================================================================

export interface TopologyNode {
  id: string
  label: string
  type: "gateway" | "service" | "database" | "cache" | "queue"
  status: "healthy" | "degraded" | "down"
  x?: number
  y?: number
}

export interface TopologyLink {
  from: string
  to: string
  bandwidth: number
}

export interface RealtimeTopologyGraphProps {
  nodes: TopologyNode[]
  links: TopologyLink[]
  onNodeClick?: (node: TopologyNode) => void
  selectedNodeId?: string
  className?: string
}

export function RealtimeTopologyGraph({
  nodes,
  links,
  onNodeClick,
  selectedNodeId,
  className,
}: RealtimeTopologyGraphProps) {
  const [time, setTime] = React.useState(0)
  const [dimensions, setDimensions] = React.useState({ width: 600, height: 400 })

  React.useEffect(() => {
    const interval = setInterval(() => setTime((t) => t + 1), 50)
    return () => clearInterval(interval)
  }, [])

  const nodePositions = React.useMemo(() => {
    const positions = new Map<string, { x: number; y: number }>()
    const centerX = dimensions.width / 2
    const centerY = dimensions.height / 2

    // Position nodes in a radial layout
    nodes.forEach((node, i) => {
      if (node.x !== undefined && node.y !== undefined) {
        positions.set(node.id, { x: node.x, y: node.y })
      } else {
        const angle = (i / nodes.length) * Math.PI * 2
        const radius = Math.min(centerX, centerY) * 0.5
        positions.set(node.id, {
          x: centerX + Math.cos(angle) * radius,
          y: centerY + Math.sin(angle) * radius,
        })
      }
    })
    return positions
  }, [nodes, dimensions])

  const typeIcons: Record<string, string> = {
    gateway: "🌐",
    service: "⚙️",
    database: "💾",
    cache: "⚡",
    queue: "📬",
  }

  const statusColors = {
    healthy: "#10b981",
    degraded: "#f59e0b",
    down: "#ef4444",
  }

  return (
    <div
      className={cn("relative w-full h-96 rounded-lg overflow-hidden bg-[#08080c]", className)}
      onResize={(e) => {
        const { width, height } = (e.target as HTMLDivElement).getBoundingClientRect()
        setDimensions({ width, height })
      }}
    >
      <svg className="w-full h-full">
        {/* Links */}
        {links.map((link, i) => {
          const fromPos = nodePositions.get(link.from)
          const toPos = nodePositions.get(link.to)
          if (!fromPos || !toPos) return null

          const isActive = Math.sin(time * 0.2 + i) > 0

          return (
            <g key={i}>
              <line
                x1={fromPos.x}
                y1={fromPos.y}
                x2={toPos.x}
                y2={toPos.y}
                stroke="rgba(255,255,255,0.1)"
                strokeWidth={1}
              />
              {isActive && (
                <circle r={3} fill="#00d4ff">
                  <animateMotion
                    path={`M${fromPos.x},${fromPos.y} L${toPos.x},${toPos.y}`}
                    dur="2s"
                    repeatCount="indefinite"
                  />
                </circle>
              )}
            </g>
          )
        })}

        {/* Nodes */}
        {nodes.map((node) => {
          const pos = nodePositions.get(node.id)
          if (!pos) return null
          const isSelected = node.id === selectedNodeId
          const color = statusColors[node.status]
          const size = 24

          return (
            <g key={node.id} className="cursor-pointer" onClick={() => onNodeClick?.(node)}>
              {/* Background */}
              <rect
                x={pos.x - size / 2}
                y={pos.y - size / 2}
                width={size}
                height={size}
                rx={4}
                fill="rgba(20, 20, 35, 0.9)"
                stroke={isSelected ? "#ffffff" : color}
                strokeWidth={isSelected ? 2 : 1}
              />

              {/* Icon */}
              <text x={pos.x} y={pos.y + size / 2 - 4} textAnchor="middle" fontSize="12">
                {typeIcons[node.type]}
              </text>

              {/* Status dot */}
              <circle
                cx={pos.x + size / 2 - 4}
                cy={pos.y - size / 2 + 4}
                r={3}
                fill={color}
              />

              {/* Label */}
              <text
                x={pos.x}
                y={pos.y + size / 2 + 12}
                textAnchor="middle"
                fill="rgba(255,255,255,0.7)"
                fontSize="10"
              >
                {node.label}
              </text>
            </g>
          )
        })}
      </svg>
    </div>
  )
}

// ============================================================================
// ExecutionDensityField - Density visualization for execution patterns
// ============================================================================

export interface DensityPoint {
  x: number
  y: number
  density: number
}

export interface ExecutionDensityFieldProps {
  points: DensityPoint[]
  width?: number
  height?: number
  colorScheme?: "brand" | "thermal" | "ocean"
  onPointClick?: (point: DensityPoint) => void
  className?: string
}

export function ExecutionDensityField({
  points,
  width = 600,
  height = 400,
  colorScheme = "brand",
  onPointClick,
  className,
}: ExecutionDensityFieldProps) {
  const canvasRef = React.useRef<HTMLCanvasElement>(null)

  const getColor = (density: number) => {
    const t = Math.min(density, 1)
    switch (colorScheme) {
      case "thermal":
        return `rgba(${t * 255}, ${t * 100}, ${(1 - t) * 255}, 0.8)`
      case "ocean":
        return `rgba(${t * 50}, ${t * 150}, ${255 - t * 100}, 0.8)`
      default:
        return `rgba(255, ${107 - t * 80}, ${53 - t * 53}, 0.8)`
    }
  }

  React.useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext("2d")
    if (!ctx) return

    // Background
    ctx.fillStyle = "#050510"
    ctx.fillRect(0, 0, width, height)

    // Draw density points
    points.forEach((point) => {
      const gradient = ctx.createRadialGradient(point.x, point.y, 0, point.x, point.y, 50)
      gradient.addColorStop(0, getColor(point.density))
      gradient.addColorStop(1, "transparent")

      ctx.fillStyle = gradient
      ctx.fillRect(point.x - 50, point.y - 50, 100, 100)
    })

    // Draw contour lines (simplified)
    ctx.strokeStyle = "rgba(255,255,255,0.1)"
    ctx.lineWidth = 0.5
    for (let i = 0; i < 5; i++) {
      const threshold = 0.2 + i * 0.15
      ctx.beginPath()
      points
        .filter((p) => Math.abs(p.density - threshold) < 0.05)
        .forEach((p, j) => {
          if (j === 0) ctx.moveTo(p.x, p.y)
          else ctx.lineTo(p.x, p.y)
        })
      ctx.stroke()
    }
  }, [points, width, height, colorScheme])

  return (
    <div className={cn("relative rounded-lg overflow-hidden bg-[#050510]", className)}>
      <canvas
        ref={canvasRef}
        width={width}
        height={height}
        className="w-full h-full cursor-crosshair"
        onClick={(e) => {
          const rect = e.currentTarget.getBoundingClientRect()
          const x = e.clientX - rect.left
          const y = e.clientY - rect.top
          onPointClick?.({ x, y, density: 0.5 })
        }}
      />
      <div className="absolute bottom-2 right-2 text-[10px] text-text-muted">
        {points.length} points
      </div>
    </div>
  )
}

// ============================================================================
// InfrastructureHologram - Holographic 3D-style infrastructure view
// ============================================================================

export interface HologramLayer {
  id: string
  label: string
  items: Array<{
    id: string
    label: string
    value: number
    max: number
  }>
}

export interface InfrastructureHologramProps {
  layers: HologramLayer[]
  width?: number
  height?: number
  onItemClick?: (layerId: string, itemId: string) => void
  className?: string
}

export function InfrastructureHologram({
  layers,
  width = 600,
  height = 400,
  onItemClick,
  className,
}: InfrastructureHologramProps) {
  const [time, setTime] = React.useState(0)
  const [hoveredItem, setHoveredItem] = React.useState<{ layer: string; item: string } | null>(null)

  React.useEffect(() => {
    const interval = setInterval(() => setTime((t) => t + 1), 50)
    return () => clearInterval(interval)
  }, [])

  const layerHeight = (height - 60) / layers.length

  return (
    <div className={cn("relative rounded-lg overflow-hidden bg-[#050510]", className)}>
      {/* Holographic effect background */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div
          className="absolute inset-0 opacity-20"
          style={{
            background: `repeating-linear-gradient(
              0deg,
              transparent,
              transparent 2px,
              rgba(0, 212, 255, 0.1) 2px,
              rgba(0, 212, 255, 0.1) 4px
            )`,
            animation: `scanline ${10 + Math.sin(time * 0.1) * 2}s linear infinite`,
          }}
        />
      </div>

      <svg width={width} height={height} className="w-full h-full">
        {/* Layers */}
        {layers.map((layer, layerIdx) => {
          const layerY = 30 + layerIdx * layerHeight

          return (
            <g key={layer.id}>
              {/* Layer label */}
              <text
                x={10}
                y={layerY + layerHeight / 2 + 4}
                fill="rgba(0, 212, 255, 0.6)"
                fontSize="10"
                fontFamily="monospace"
              >
                {layer.label.toUpperCase()}
              </text>

              {/* Items */}
              {layer.items.map((item, itemIdx) => {
                const itemX = 100 + itemIdx * ((width - 120) / layer.items.length)
                const fillWidth = (item.value / item.max) * 80
                const isHovered = hoveredItem?.layer === layer.id && hoveredItem?.item === item.id

                return (
                  <g
                    key={item.id}
                    className="cursor-pointer"
                    onMouseEnter={() => setHoveredItem({ layer: layer.id, item: item.id })}
                    onMouseLeave={() => setHoveredItem(null)}
                    onClick={() => onItemClick?.(layer.id, item.id)}
                  >
                    {/* Item bar */}
                    <rect
                      x={itemX}
                      y={layerY + 8}
                      width={80}
                      height={12}
                      rx={2}
                      fill="rgba(20, 20, 35, 0.8)"
                      stroke={isHovered ? "#00d4ff" : "rgba(0, 212, 255, 0.2)"}
                      strokeWidth={isHovered ? 1 : 0.5}
                    />

                    {/* Fill */}
                    <rect
                      x={itemX + 2}
                      y={layerY + 10}
                      width={fillWidth}
                      height={8}
                      rx={1}
                      fill="url(#hologram-gradient)"
                      opacity={0.7 + Math.sin(time * 0.2 + itemIdx * 0.5) * 0.2}
                    />

                    {/* Glow effect for active items */}
                    {item.value > item.max * 0.8 && (
                      <rect
                        x={itemX + 2}
                        y={layerY + 10}
                        width={fillWidth}
                        height={8}
                        rx={1}
                        fill="#00d4ff"
                        opacity={0.3}
                      >
                        <animate attributeName="opacity" values="0.3;0.6;0.3" dur="1s" repeatCount="indefinite" />
                      </rect>
                    )}

                    {/* Label */}
                    <text
                      x={itemX}
                      y={layerY + layerHeight - 4}
                      fill="rgba(255,255,255,0.5)"
                      fontSize="8"
                      fontFamily="monospace"
                    >
                      {item.label.slice(0, 8)}
                    </text>

                    {/* Value */}
                    <text
                      x={itemX + 80}
                      y={layerY + layerHeight - 4}
                      fill="rgba(0, 212, 255, 0.8)"
                      fontSize="8"
                      fontFamily="monospace"
                      textAnchor="end"
                    >
                      {item.value.toFixed(0)}
                    </text>
                  </g>
                )
              })}
            </g>
          )
        })}

        {/* Gradient definition */}
        <defs>
          <linearGradient id="hologram-gradient" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stopColor="#00d4ff" />
            <stop offset="100%" stopColor="#00d4ff" stopOffset="100%" />
          </linearGradient>
        </defs>
      </svg>

      {/* Scanline animation */}
      <style>{`
        @keyframes scanline {
          from { transform: translateY(0); }
          to { transform: translateY(100px); }
        }
      `}</style>

      {/* Stats overlay */}
      <div className="absolute top-2 right-2 text-[10px] text-[#00d4ff] font-mono opacity-60">
        SYS://INFRA.MONITOR
      </div>
    </div>
  )
}

