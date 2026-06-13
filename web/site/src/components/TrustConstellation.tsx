/**
 * TrustConstellation — FunctionFly's signature 3D brand scene.
 *
 * A single premium visualization that captures the entire brand:
 *   - Aviation / air-traffic-control metaphor (FunctionFly = flight control for AI agents)
 *   - Central "Trust Core" (the FunctionFly platform)
 *   - Orbiting "Agent Nodes" (connected AI agents / functions)
 *   - Verified "Trust Beams" with signed receipt particles flowing along them
 *   - Radar sweep + altitude rings for the control-tower feel
 *   - Brand palette: flame, cyan, afterburner, success-green, stratosphere
 *
 * Used in the hero section. Self-contained: imports three, @react-three/fiber,
 * @react-three/drei only. Renders responsively and is mouse-reactive.
 */

import { Canvas, useFrame, useThree } from '@react-three/fiber'
import { Stars } from '@react-three/drei'
import { useEffect, useMemo, useRef } from 'react'
import * as THREE from 'three'

// Brand palette (matches design-system.css)
const BRAND = {
  flame: new THREE.Color(0xff6b35),
  flameDark: new THREE.Color(0xe85a2a),
  cyan: new THREE.Color(0x00d4ff),
  cyanDark: new THREE.Color(0x00b8e0),
  afterburner: new THREE.Color(0xff4f5e),
  success: new THREE.Color(0x00ff9d),
  stratosphere: new THREE.Color(0x5b7cf5),
  tarmac: new THREE.Color(0x0d1117),
  cockpit: new THREE.Color(0x161b22),
  flightdeck: new THREE.Color(0x21262d),
  white: new THREE.Color(0xf0f6fc),
}

const AGENT_COUNT = 28
const ORBIT_RINGS = 3
const RECEIPTS_PER_LINK = 2

interface AgentNode {
  basePos: THREE.Vector3
  angle: number
  radius: number
  altitude: number
  speed: number
  phase: number
  color: THREE.Color
  size: number
  trust: number // 0..1 — strength of trust link
}

function makeAgents(): AgentNode[] {
  const palette = [
    BRAND.cyan,
    BRAND.cyan,
    BRAND.flame,
    BRAND.success,
    BRAND.stratosphere,
    BRAND.afterburner,
  ]
  const agents: AgentNode[] = []
  for (let i = 0; i < AGENT_COUNT; i++) {
    const ring = i % ORBIT_RINGS
    const radius = 3.2 + ring * 1.8 + Math.random() * 0.6
    const altitude = (ring - 1) * 1.4 + (Math.random() - 0.5) * 0.8
    const angle = (i / AGENT_COUNT) * Math.PI * 2 + Math.random() * 0.3
    const color = palette[Math.floor(Math.random() * palette.length)]
    agents.push({
      basePos: new THREE.Vector3(
        Math.cos(angle) * radius,
        altitude,
        Math.sin(angle) * radius
      ),
      angle,
      radius,
      altitude,
      speed: 0.04 + Math.random() * 0.06,
      phase: Math.random() * Math.PI * 2,
      color,
      size: 0.08 + Math.random() * 0.07,
      trust: 0.55 + Math.random() * 0.45,
    })
  }
  return agents
}

const TrustCore: React.FC = () => {
  const ref = useRef<THREE.Group>(null)
  const innerRef = useRef<THREE.Mesh>(null)
  const ringRef = useRef<THREE.Mesh>(null)
  const glowRef = useRef<THREE.Mesh>(null)

  useFrame((state) => {
    const t = state.clock.getElapsedTime()
    if (ref.current) {
      ref.current.rotation.y = t * 0.15
      ref.current.rotation.x = Math.sin(t * 0.3) * 0.08
    }
    if (innerRef.current) {
      const s = 1 + Math.sin(t * 2) * 0.04
      innerRef.current.scale.setScalar(s)
    }
    if (ringRef.current) {
      ringRef.current.rotation.z = t * 0.4
      ringRef.current.rotation.x = Math.PI / 2
    }
    if (glowRef.current) {
      const m = glowRef.current.material as THREE.MeshBasicMaterial
      m.opacity = 0.18 + Math.sin(t * 1.6) * 0.06
    }
  })

  return (
    <group ref={ref}>
      {/* Outer glow halo */}
      <mesh ref={glowRef}>
        <sphereGeometry args={[1.6, 32, 32]} />
        <meshBasicMaterial
          color={BRAND.cyan}
          transparent
          opacity={0.18}
          blending={THREE.AdditiveBlending}
          depthWrite={false}
        />
      </mesh>
      {/* Core sphere — the FunctionFly trust platform */}
      <mesh ref={innerRef}>
        <icosahedronGeometry args={[0.55, 2]} />
        <meshStandardMaterial
          color={BRAND.cyan}
          emissive={BRAND.cyan}
          emissiveIntensity={1.4}
          metalness={0.4}
          roughness={0.25}
        />
      </mesh>
      {/* Inner flame heart */}
      <mesh>
        <sphereGeometry args={[0.28, 24, 24]} />
        <meshBasicMaterial color={BRAND.flame} />
      </mesh>
      {/* Orbital ring */}
      <mesh ref={ringRef}>
        <torusGeometry args={[1.05, 0.012, 16, 128]} />
        <meshBasicMaterial color={BRAND.flame} transparent opacity={0.7} />
      </mesh>
      <mesh rotation={[Math.PI / 2.2, 0, Math.PI / 3]}>
        <torusGeometry args={[1.25, 0.006, 12, 128]} />
        <meshBasicMaterial color={BRAND.cyan} transparent opacity={0.5} />
      </mesh>
    </group>
  )
}

const AgentNodeMesh: React.FC<{ agent: AgentNode; index: number }> = ({ agent, index }) => {
  const ref = useRef<THREE.Mesh>(null)
  const glowRef = useRef<THREE.Mesh>(null)
  const orbit = useRef({
    pos: agent.basePos.clone(),
  })

  useFrame((state) => {
    const t = state.clock.getElapsedTime()
    const wobble = Math.sin(t * agent.speed * 4 + agent.phase) * 0.15
    const a = agent.angle + t * agent.speed
    orbit.current.pos.set(
      Math.cos(a) * agent.radius,
      agent.altitude + wobble,
      Math.sin(a) * agent.radius
    )
    if (ref.current) {
      ref.current.position.copy(orbit.current.pos)
    }
    if (glowRef.current) {
      glowRef.current.position.copy(orbit.current.pos)
      const pulse = 0.5 + Math.sin(t * 2.5 + agent.phase) * 0.5
      const s = 0.6 + pulse * 0.5
      glowRef.current.scale.setScalar(s)
      const m = glowRef.current.material as THREE.MeshBasicMaterial
      m.opacity = 0.12 + pulse * 0.12
    }
  })

  return (
    <>
      <mesh ref={glowRef}>
        <sphereGeometry args={[agent.size * 2.4, 16, 16]} />
        <meshBasicMaterial
          color={agent.color}
          transparent
          opacity={0.18}
          blending={THREE.AdditiveBlending}
          depthWrite={false}
        />
      </mesh>
      <mesh ref={ref}>
        <icosahedronGeometry args={[agent.size, 1]} />
        <meshStandardMaterial
          color={agent.color}
          emissive={agent.color}
          emissiveIntensity={0.9}
          metalness={0.3}
          roughness={0.4}
        />
      </mesh>
    </>
  )
}

interface ReceiptFlow {
  startIdx: number
  endIdx: number
  offset: number // 0..1 position along the link
  speed: number
  color: THREE.Color
  outward: boolean // true = agent->core (request), false = core->agent (receipt)
}

const TrustLinks: React.FC<{ agents: AgentNode[] }> = ({ agents }) => {
  const linesRef = useRef<THREE.Group>(null)
  const particlesRef = useRef<THREE.Group>(null)
  const particleStates = useRef<
    { pos: THREE.Vector3; start: THREE.Vector3; end: THREE.Vector3; offset: number; speed: number; outward: boolean; color: THREE.Color }[]
  >([])

  // Build a sparse but visually rich set of links (core -> subset of agents)
  const links = useMemo(() => {
    const out: { from: number; to: number }[] = []
    // Core-to-agent: each agent
    for (let i = 0; i < agents.length; i++) out.push({ from: -1, to: i })
    // Agent-to-agent: a few inter-agent links (agent swarm coordination)
    for (let i = 0; i < agents.length; i++) {
      const j = (i + 7) % agents.length
      if (i < j) out.push({ from: i, to: j })
    }
    return out
  }, [agents])

  // Initialize receipt flows (2 per core link, 1 per agent-agent link)
  const flows = useMemo<ReceiptFlow[]>(() => {
    const arr: ReceiptFlow[] = []
    links.forEach((l, i) => {
      const count = l.from === -1 ? RECEIPTS_PER_LINK : 1
      for (let k = 0; k < count; k++) {
        const isOutward = l.from === -1 && k === 1 // one request, one receipt
        const colorChoices = [BRAND.flame, BRAND.cyan, BRAND.success, BRAND.cyan, BRAND.flame]
        arr.push({
          startIdx: l.from,
          endIdx: l.to,
          offset: Math.random(),
          speed: 0.18 + Math.random() * 0.22,
          color: colorChoices[Math.floor(Math.random() * colorChoices.length)],
          outward: isOutward,
        })
      }
    })
    return arr
  }, [links])

  // Pre-build line geometries (one per link)
  const lineGeoms = useMemo(() => {
    return links.map(() => {
      const g = new THREE.BufferGeometry()
      g.setAttribute('position', new THREE.BufferAttribute(new Float32Array(6), 3))
      return g
    })
  }, [links])

  // Per-link colors
  const linkColors = useMemo(() => {
    return links.map((l) => {
      if (l.from === -1) return BRAND.cyan
      return BRAND.flame
    })
  }, [links])

  // Initialize particle state once
  useEffect(() => {
    particleStates.current = flows.map((f) => ({
      pos: new THREE.Vector3(),
      start: new THREE.Vector3(),
      end: new THREE.Vector3(),
      offset: f.offset,
      speed: f.speed,
      outward: f.outward,
      color: f.color,
    }))
  }, [flows])

  const corePos = useRef(new THREE.Vector3(0, 0, 0))

  useFrame((state) => {
    const t = state.clock.getElapsedTime()
    // Update line positions
    links.forEach((l, idx) => {
      const geom = lineGeoms[idx]
      const pos = geom.attributes.position as THREE.BufferAttribute
      const arr = pos.array as Float32Array
      if (l.from === -1) {
        arr[0] = 0; arr[1] = 0; arr[2] = 0
        const a = agents[l.to]
        const tt = t * a.speed
        const ang = a.angle + tt
        const wobble = Math.sin(tt * 4 + a.phase) * 0.15
        arr[3] = Math.cos(ang) * a.radius
        arr[4] = a.altitude + wobble
        arr[5] = Math.sin(ang) * a.radius
      } else {
        const a = agents[l.from]
        const b = agents[l.to]
        const ta = t * a.speed
        const tb = t * b.speed
        const aa = a.angle + ta
        const ab = b.angle + tb
        const wa = Math.sin(ta * 4 + a.phase) * 0.15
        const wb = Math.sin(tb * 4 + b.phase) * 0.15
        arr[0] = Math.cos(aa) * a.radius
        arr[1] = a.altitude + wa
        arr[2] = Math.sin(aa) * a.radius
        arr[3] = Math.cos(ab) * b.radius
        arr[4] = b.altitude + wb
        arr[5] = Math.sin(ab) * b.radius
      }
      pos.needsUpdate = true
    })

    // Update receipt particles
    if (!particlesRef.current) return
    const meshes = particlesRef.current.children as THREE.Mesh[]
    particleStates.current.forEach((s, i) => {
      if (!meshes[i]) return
      // Compute current endpoints (same math as above)
      let start: THREE.Vector3, end: THREE.Vector3
      const f = flows[i]
      if (f.startIdx === -1) {
        start = corePos.current
        const a = agents[f.endIdx]
        const tt = t * a.speed
        const ang = a.angle + tt
        const wobble = Math.sin(tt * 4 + a.phase) * 0.15
        end = new THREE.Vector3(Math.cos(ang) * a.radius, a.altitude + wobble, Math.sin(ang) * a.radius)
      } else {
        const a = agents[f.startIdx]
        const b = agents[f.endIdx]
        const ta = t * a.speed
        const tb = t * b.speed
        const aa = a.angle + ta
        const ab = b.angle + tb
        const wa = Math.sin(ta * 4 + a.phase) * 0.15
        const wb = Math.sin(tb * 4 + b.phase) * 0.15
        start = new THREE.Vector3(Math.cos(aa) * a.radius, a.altitude + wa, Math.sin(aa) * a.radius)
        end = new THREE.Vector3(Math.cos(ab) * b.radius, b.altitude + wb, Math.sin(ab) * b.radius)
      }
      s.offset += (s.outward ? 1 : -1) * s.speed * 0.012
      if (s.offset > 1) s.offset = 0
      if (s.offset < 0) s.offset = 1
      // Cubic ease for a "packet" feel
      const e = s.offset
      const eased = e < 0.5 ? 4 * e * e * e : 1 - Math.pow(-2 * e + 2, 3) / 2
      s.pos.lerpVectors(start, end, eased)
      // Add a subtle arc
      const arc = Math.sin(eased * Math.PI) * 0.35
      meshes[i].position.set(s.pos.x, s.pos.y + arc, s.pos.z)
      const scale = 0.7 + Math.sin(t * 6 + i) * 0.3
      meshes[i].scale.setScalar(scale)
    })
  })

  return (
    <>
      <group ref={linesRef}>
        {lineGeoms.map((g, i) => (
          // @ts-expect-error - three.js LineBasicMaterial with vertexColors via material props
          <line key={`l-${i}`} geometry={g}>
            <lineBasicMaterial
              color={linkColors[i]}
              transparent
              opacity={0.18}
            />
          </line>
        ))}
      </group>
      <group ref={particlesRef}>
        {flows.map((f, i) => (
          <mesh key={`p-${i}`}>
            <sphereGeometry args={[0.05, 8, 8]} />
            <meshBasicMaterial color={f.color} />
          </mesh>
        ))}
      </group>
    </>
  )
}

const AltitudeRings: React.FC = () => {
  const ref = useRef<THREE.Group>(null)
  useFrame((state) => {
    if (ref.current) {
      ref.current.rotation.y = state.clock.getElapsedTime() * 0.03
    }
  })
  const rings = [-2.2, -0.8, 0.8, 2.2]
  return (
    <group ref={ref}>
      {rings.map((y, i) => (
        <mesh key={i} position={[0, y, 0]} rotation={[Math.PI / 2, 0, 0]}>
          <ringGeometry args={[2.6 + i * 0.3, 2.605 + i * 0.3, 128]} />
          <meshBasicMaterial
            color={i % 2 === 0 ? BRAND.cyan : BRAND.flame}
            transparent
            opacity={0.08}
            side={THREE.DoubleSide}
          />
        </mesh>
      ))}
    </group>
  )
}

const RadarSweep: React.FC = () => {
  const ref = useRef<THREE.Mesh>(null)
  useFrame((state) => {
    if (ref.current) {
      ref.current.rotation.z = -state.clock.getElapsedTime() * 0.6
    }
  })
  return (
    <mesh ref={ref} rotation={[-Math.PI / 2, 0, 0]} position={[0, -2.35, 0]}>
      <ringGeometry args={[0, 4.5, 64, 1, 0, Math.PI * 0.18]} />
      <meshBasicMaterial
        color={BRAND.cyan}
        transparent
        opacity={0.12}
        side={THREE.DoubleSide}
        blending={THREE.AdditiveBlending}
        depthWrite={false}
      />
    </mesh>
  )
}

const GridFloor: React.FC = () => {
  return (
    <gridHelper
      args={[24, 40, BRAND.cyan.getHex(), BRAND.cockpit.getHex()]}
      position={[0, -2.5, 0]}
    />
  )
}

const CameraRig: React.FC = () => {
  const { camera, mouse } = useThree()
  const target = useRef(new THREE.Vector3(0, 0, 0))
  useFrame(() => {
    // Gentle parallax tilt based on mouse
    const desiredX = mouse.x * 0.6
    const desiredY = mouse.y * 0.4 + 0.6
    camera.position.x += (desiredX - camera.position.x) * 0.03
    camera.position.y += (desiredY - camera.position.y) * 0.03
    camera.position.z += (8 - camera.position.z) * 0.03
    camera.lookAt(target.current)
  })
  return null
}

const SceneContents: React.FC = () => {
  const agents = useMemo(() => makeAgents(), [])
  return (
    <>
      <ambientLight intensity={0.35} />
      <pointLight position={[0, 0, 0]} color={BRAND.cyan} intensity={1.6} distance={10} />
      <pointLight position={[5, 4, 3]} color={BRAND.flame} intensity={0.9} distance={14} />
      <pointLight position={[-5, -3, -4]} color={BRAND.stratosphere} intensity={0.6} distance={14} />

      <Stars radius={50} depth={40} count={1200} factor={3} fade speed={0.5} />

      <GridFloor />
      <AltitudeRings />
      <RadarSweep />
      <TrustCore />
      {agents.map((a, i) => (
        <AgentNodeMesh key={i} agent={a} index={i} />
      ))}
      <TrustLinks agents={agents} />

      <CameraRig />
    </>
  )
}

const TrustConstellation: React.FC = () => {
  return (
    <Canvas
      camera={{ position: [0, 0.6, 8], fov: 55, near: 0.1, far: 200 }}
      dpr={[1, 2]}
      gl={{ antialias: true, alpha: true, powerPreference: 'high-performance' }}
      style={{ background: 'transparent' }}
    >
      <SceneContents />
    </Canvas>
  )
}

export default TrustConstellation
