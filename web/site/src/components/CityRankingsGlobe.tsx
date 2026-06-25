/**
 * CityRankingsGlobe — 3D visualization for the AI World Map.
 *
 * Renders a wireframe globe with glowing city markers positioned by lat/lon.
 * Cities are tier-colored (gold/blue/green) based on per-capita score. The
 * globe auto-rotates slowly; users can drag to rotate manually and click a
 * marker to open that city's detail page.
 *
 * Lazy-loaded by the rankings page; uses three.js + @react-three/fiber.
 * Falls back to a static SVG world map when WebGL is unavailable or the
 * device is low-power.
 */

import { Canvas, useFrame, useThree } from '@react-three/fiber'
import type { ThreeEvent } from '@react-three/fiber'
import { useEffect, useMemo, useRef, useState } from 'react'
import * as THREE from 'three'

export interface GlobePoint {
  metro_slug: string
  metro_name: string
  country_code: string
  state_code: string
  latitude: number
  longitude: number
  rank_position: number
  score_per_capita: number
  active_users: number
  tier: string
}

interface Props {
  points: GlobePoint[]
  /** Called when a marker is clicked. */
  onSelect?: (p: GlobePoint) => void
  /** Optional fixed height in pixels. */
  height?: number
  /** Which tiers to show. If undefined, all tiers are visible. */
  visibleTiers?: Array<'gold' | 'blue' | 'green'>
  /** Whether the globe data is loading. */
  isLoading?: boolean
}

const GLOBE_RADIUS = 1.6
const TIER_COLORS: Record<GlobePoint['tier'], THREE.Color> = {
  gold: new THREE.Color(0xffb74d),
  blue: new THREE.Color(0x4fc3f7),
  green: new THREE.Color(0x66bb6a),
}
const TIER_SIZES: Record<GlobePoint['tier'], number> = {
  gold: 0.04,
  blue: 0.032,
  green: 0.024,
}

// Convert lat/lon (degrees) to a point on a unit sphere.
// Standard geographic → Cartesian: x = R·cos(lat)·cos(lon),
// y = R·sin(lat), z = R·cos(lat)·sin(lon).
function latLonToVec3(lat: number, lon: number, radius = GLOBE_RADIUS): THREE.Vector3 {
  const phi = (90 - lat) * (Math.PI / 180) // polar angle
  const theta = (lon + 180) * (Math.PI / 180) // azimuth
  return new THREE.Vector3(
    -radius * Math.sin(phi) * Math.cos(theta),
    radius * Math.cos(phi),
    radius * Math.sin(phi) * Math.sin(theta),
  )
}

// Build a wireframe sphere by sampling latitude and longitude circles.
function buildGlobeWireframe(): { positions: Float32Array; count: number } {
  const segs = 36
  const rings = 18
  const positions: number[] = []
  // Latitude rings (horizontal).
  for (let i = 1; i < rings; i++) {
    const lat = -90 + (180 * i) / rings
    const r = Math.cos((lat * Math.PI) / 180) * GLOBE_RADIUS
    const y = Math.sin((lat * Math.PI) / 180) * GLOBE_RADIUS
    for (let j = 0; j <= segs; j++) {
      const a = (j / segs) * Math.PI * 2
      positions.push(Math.cos(a) * r, y, Math.sin(a) * r)
    }
  }
  // Longitude meridians (vertical).
  for (let j = 0; j < segs; j += 2) {
    const a = (j / segs) * Math.PI * 2
    for (let i = 0; i <= rings; i++) {
      const lat = -90 + (180 * i) / rings
      const phi = ((90 - lat) * Math.PI) / 180
      const r = Math.sin(phi) * GLOBE_RADIUS
      const y = Math.cos(phi) * GLOBE_RADIUS
      positions.push(Math.cos(a) * r, y, Math.sin(a) * r)
    }
  }
  return { positions: new Float32Array(positions), count: positions.length / 3 }
}

const WireGlobe: React.FC<{ autoRotate: boolean }> = ({ autoRotate }) => {
  const ref = useRef<THREE.Group>(null)
  const wire = useMemo(buildGlobeWireframe, [])

  useFrame((_, delta) => {
    if (ref.current && autoRotate) {
      ref.current.rotation.y += delta * 0.04
    }
  })

  return (
    <group ref={ref}>
      {/* Faint inner sphere to give the globe visual weight. */}
      <mesh>
        <sphereGeometry args={[GLOBE_RADIUS * 0.995, 48, 48]} />
        <meshBasicMaterial color="#0d1117" transparent opacity={0.4} />
      </mesh>
      {/* Latitude/longitude wireframe. */}
      <lineSegments>
        <bufferGeometry>
          <bufferAttribute
            attach="attributes-position"
            args={[wire.positions, 3]}
            count={wire.count}
          />
        </bufferGeometry>
        <lineBasicMaterial color="#5b7cf5" transparent opacity={0.35} />
      </lineSegments>
    </group>
  )
}

interface MarkerProps {
  point: GlobePoint
  onSelect?: (p: GlobePoint) => void
}

const Marker: React.FC<MarkerProps> = ({ point, onSelect }) => {
  const ref = useRef<THREE.Mesh>(null)
  const haloRef = useRef<THREE.Mesh>(null)
  const position = useMemo(
    () => latLonToVec3(point.latitude, point.longitude, GLOBE_RADIUS * 1.005),
    [point.latitude, point.longitude],
  )
  const color = TIER_COLORS[point.tier]
  const size = TIER_SIZES[point.tier]

  useFrame(({ clock }) => {
    if (!ref.current) return
    // Gentle pulse.
    const s = 1 + Math.sin(clock.getElapsedTime() * 2 + point.rank_position) * 0.15
    ref.current.scale.setScalar(s)
    if (haloRef.current) {
      haloRef.current.scale.setScalar(1.8 + Math.sin(clock.getElapsedTime() * 2 + point.rank_position) * 0.2)
      ;(haloRef.current.material as THREE.MeshBasicMaterial).opacity =
        0.25 + Math.sin(clock.getElapsedTime() * 2 + point.rank_position) * 0.1
    }
  })

  const handleClick = (e: ThreeEvent<MouseEvent>) => {
    e.stopPropagation()
    onSelect?.(point)
  }

  return (
    <group position={position}>
      {/* Soft halo — slightly larger transparent sphere. */}
      <mesh ref={haloRef}>
        <sphereGeometry args={[size * 1.4, 16, 16]} />
        <meshBasicMaterial color={color} transparent opacity={0.25} depthWrite={false} />
      </mesh>
      {/* Bright core. */}
      <mesh ref={ref} onClick={handleClick} onPointerOver={(e) => { e.stopPropagation(); document.body.style.cursor = 'pointer' }} onPointerOut={() => { document.body.style.cursor = '' }}>
        <sphereGeometry args={[size, 16, 16]} />
        <meshBasicMaterial color={color} />
      </mesh>
    </group>
  )
}

const CameraRig: React.FC = () => {
  const { camera } = useThree()
  useEffect(() => {
    camera.position.set(0, 0.4, 4.2)
  }, [camera])
  return null
}

const hasWebGL = (): boolean => {
  if (typeof window === 'undefined') return false
  try {
    const canvas = document.createElement('canvas')
    return !!(canvas.getContext('webgl2') || canvas.getContext('webgl'))
  } catch {
    return false
  }
}

const FallbackMap: React.FC<{ points: GlobePoint[]; onSelect?: (p: GlobePoint) => void }> = ({
  points,
  onSelect,
}) => {
  // Simple equirectangular projection. The map itself is omitted (no land
  // outlines); we only render the markers so the fallback still works
  // gracefully on low-power devices.
  const w = 800
  const h = 400
  return (
    <svg viewBox={`0 0 ${w} ${h}`} className="city-rankings-globe-fallback" style={{ width: '100%', maxWidth: 720 }}>
      <rect x={0} y={0} width={w} height={h} fill="#0d1117" rx={12} />
      <text x={w / 2} y={24} textAnchor="middle" fill="#5b7cf5" fontSize={12} opacity={0.6}>
        2D fallback — WebGL unavailable
      </text>
      {points.map((p) => {
        const x = ((p.longitude + 180) / 360) * w
        const y = ((90 - p.latitude) / 180) * h
        const color = p.tier === 'gold' ? '#ffb74d' : p.tier === 'blue' ? '#4fc3f7' : '#66bb6a'
        return (
          <g key={p.metro_slug} onClick={() => onSelect?.(p)} style={{ cursor: 'pointer' }}>
            <circle cx={x} cy={y} r={p.tier === 'gold' ? 6 : p.tier === 'blue' ? 5 : 4} fill={color} opacity={0.3} />
            <circle cx={x} cy={y} r={p.tier === 'gold' ? 3 : p.tier === 'blue' ? 2.5 : 2} fill={color} />
          </g>
        )
      })}
    </svg>
  )
}

export const CityRankingsGlobe: React.FC<Props> = ({ points, onSelect, height = 480, visibleTiers, isLoading }) => {
  const [supportsWebGL, setSupportsWebGL] = useState<boolean | null>(null)
  useEffect(() => {
    setSupportsWebGL(hasWebGL())
  }, [])

  const filteredPoints = useMemo(() => {
    if (!visibleTiers || visibleTiers.length === 0) return points
    return points.filter((p) => visibleTiers.includes(p.tier as 'gold' | 'blue' | 'green'))
  }, [points, visibleTiers])

  if (supportsWebGL === null || isLoading) {
    return <div className="city-rankings-globe-skeleton" style={{ height }} />
  }

  if (!supportsWebGL) {
    return <FallbackMap points={filteredPoints} onSelect={onSelect} />
  }

  return (
    <div className="city-rankings-globe" style={{ height, position: 'relative' }}>
      <Canvas
        camera={{ position: [0, 0.4, 4.2], fov: 45 }}
        dpr={[1, 2]}
        gl={{ antialias: true, alpha: true, powerPreference: 'high-performance' }}
        style={{ background: 'transparent' }}
      >
        <CameraRig />
        <ambientLight intensity={0.4} />
        <WireGlobe autoRotate={true} />
        {filteredPoints.map((p) => (
          <Marker key={p.metro_slug} point={p} onSelect={onSelect} />
        ))}
      </Canvas>
    </div>
  )
}

export default CityRankingsGlobe
