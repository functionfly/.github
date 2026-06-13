import React, { useRef, useMemo, useEffect } from 'react'
import { Canvas, useFrame, useThree } from '@react-three/fiber'
import * as THREE from 'three'

interface NodeData {
  x: number
  y: number
  z: number
  vx: number
  vy: number
  vz: number
  pulse: number
  pulseSpeed: number
  c: THREE.Color
}

interface ReceiptData {
  active: boolean
  x: number
  y: number
  z: number
  life: number
  maxLife: number
  vx: number
  vy: number
  vz: number
}

const NODE_COUNT = 100
const EDGE_MAX = 200
const RECEIPT_COUNT = 50
const GLOW_COLORS = [
  new THREE.Color(0xFF6B35),
  new THREE.Color(0x00D4FF),
  new THREE.Color(0x5B7CF5),
  new THREE.Color(0x00FF9D),
  new THREE.Color(0xFF4F5E),
]

export const CustomCursor: React.FC = () => {
  const dotRef = useRef<HTMLDivElement>(null)
  const ringRef = useRef<HTMLDivElement>(null)
  const pos = useRef({ x: 0, y: 0, rx: 0, ry: 0 })

  useEffect(() => {
    const onMove = (e: MouseEvent) => {
      pos.current.x = e.clientX
      pos.current.y = e.clientY
      if (dotRef.current) {
        dotRef.current.style.transform = `translate(${e.clientX}px, ${e.clientY}px) translate(-50%,-50%)`
      }
    }
    window.addEventListener('mousemove', onMove)
    let raf: number
    const animate = () => {
      pos.current.rx += (pos.current.x - pos.current.rx) * 0.1
      pos.current.ry += (pos.current.y - pos.current.ry) * 0.1
      if (ringRef.current) {
        ringRef.current.style.transform = `translate(${pos.current.rx}px, ${pos.current.ry}px) translate(-50%,-50%)`
      }
      raf = requestAnimationFrame(animate)
    }
    raf = requestAnimationFrame(animate)
    return () => {
      window.removeEventListener('mousemove', onMove)
      cancelAnimationFrame(raf)
    }
  }, [])

  useEffect(() => {
    const onEnter = () => document.body.classList.add('ff-cursor-hover')
    const onLeave = () => document.body.classList.remove('ff-cursor-hover')
    const selectors = 'a, button, .ff-feature-card, .ff-protocol-card, .ff-audience-card, .ff-receipt-chip'
    const els = document.querySelectorAll(selectors)
    els.forEach(el => {
      el.addEventListener('mouseenter', onEnter)
      el.addEventListener('mouseleave', onLeave)
    })
    return () => {
      els.forEach(el => {
        el.removeEventListener('mouseenter', onEnter)
        el.removeEventListener('mouseleave', onLeave)
      })
    }
  }, [])

  return (
    <div className="ff-cursor-track">
      <div ref={dotRef} className="ff-cursor-dot" />
      <div ref={ringRef} className="ff-cursor-ring" />
    </div>
  )
}

const ParticleNetwork: React.FC = () => {
  const meshRef = useRef<THREE.Points>(null)
  const linesRef = useRef<THREE.LineSegments>(null)
  const receiptsRef = useRef<THREE.Points>(null)
  const { camera } = useThree()
  const mouse3D = useRef(new THREE.Vector2(0, 0))
  const frameCount = useRef(0)

  const data = useMemo(() => {
    const nodes: NodeData[] = []
    const receipts: ReceiptData[] = []
    for (let i = 0; i < NODE_COUNT; i++) {
      const r = 12 + Math.random() * 6
      const theta = Math.random() * Math.PI * 2
      const phi = Math.acos(2 * Math.random() - 1)
      nodes.push({
        x: r * Math.sin(phi) * Math.cos(theta),
        y: r * Math.sin(phi) * Math.sin(theta) * 0.5,
        z: r * Math.cos(phi) * 0.6,
        vx: (Math.random() - 0.5) * 0.012,
        vy: (Math.random() - 0.5) * 0.008,
        vz: (Math.random() - 0.5) * 0.006,
        pulse: Math.random() * Math.PI * 2,
        pulseSpeed: 0.015 + Math.random() * 0.025,
        c: GLOW_COLORS[Math.floor(Math.random() * GLOW_COLORS.length)],
      })
    }
    for (let i = 0; i < RECEIPT_COUNT; i++) {
      receipts.push({ 
        active: false, x: 0, y: 0, z: 0, life: 0, maxLife: 80 + Math.floor(Math.random() * 40),
        vx: 0, vy: 0, vz: 0
      })
    }
    return { nodes, receipts }
  }, [])

  const nodeGeo = useMemo(() => {
    const g = new THREE.BufferGeometry()
    g.setAttribute('position', new THREE.BufferAttribute(new Float32Array(NODE_COUNT * 3), 3))
    g.setAttribute('color', new THREE.BufferAttribute(new Float32Array(NODE_COUNT * 3), 3))
    return g
  }, [])

  const edgeGeo = useMemo(() => {
    const g = new THREE.BufferGeometry()
    g.setAttribute('position', new THREE.BufferAttribute(new Float32Array(EDGE_MAX * 2 * 3), 3))
    g.setAttribute('color', new THREE.BufferAttribute(new Float32Array(EDGE_MAX * 2 * 3), 3))
    return g
  }, [])

  const receiptGeo = useMemo(() => {
    const g = new THREE.BufferGeometry()
    const rPos = new Float32Array(RECEIPT_COUNT * 3)
    const rCol = new Float32Array(RECEIPT_COUNT * 3)
    for (let i = 0; i < RECEIPT_COUNT; i++) { rPos[i*3] = 9999; rPos[i*3+1] = 9999; rPos[i*3+2] = 9999 }
    g.setAttribute('position', new THREE.BufferAttribute(rPos, 3))
    g.setAttribute('color', new THREE.BufferAttribute(rCol, 3))
    return g
  }, [])

  useEffect(() => {
    camera.position.z = 28
  }, [camera])

  useFrame(() => {
    frameCount.current++
    const posAttr = nodeGeo.attributes.position as THREE.BufferAttribute
    const colAttr = nodeGeo.attributes.color as THREE.BufferAttribute
    const nodes = data.nodes
    const posArr = posAttr.array as Float32Array
    const colArr = colAttr.array as Float32Array

    camera.position.x += (mouse3D.current.x * 4 - camera.position.x) * 0.015
    camera.position.y += (mouse3D.current.y * 3 - camera.position.y) * 0.015
    camera.lookAt(0, 0, 0)

    for (let i = 0; i < NODE_COUNT; i++) {
      const n = nodes[i]
      n.x += n.vx; n.y += n.vy; n.z += n.vz
      const dist = Math.sqrt(n.x * n.x + n.y * n.y + n.z * n.z)
      if (dist > 16) { n.vx *= -1; n.vy *= -1; n.vz *= -1 }
      n.pulse += n.pulseSpeed
      const brightness = 0.5 + 0.5 * Math.sin(n.pulse)
      posArr[i*3] = n.x; posArr[i*3+1] = n.y; posArr[i*3+2] = n.z
      colArr[i*3] = n.c.r * brightness; colArr[i*3+1] = n.c.g * brightness; colArr[i*3+2] = n.c.b * brightness
    }
    posAttr.needsUpdate = true; colAttr.needsUpdate = true

    const edgePosArr = (edgeGeo.attributes.position as THREE.BufferAttribute).array as Float32Array
    const edgeColArr = (edgeGeo.attributes.color as THREE.BufferAttribute).array as Float32Array
    let eIdx = 0
    for (let i = 0; i < NODE_COUNT && eIdx < EDGE_MAX; i++) {
      for (let j = i + 1; j < NODE_COUNT && eIdx < EDGE_MAX; j++) {
        const dx = nodes[i].x - nodes[j].x
        const dy = nodes[i].y - nodes[j].y
        const dz = nodes[i].z - nodes[j].z
        const d = Math.sqrt(dx*dx + dy*dy + dz*dz)
        if (d < 6) {
          const alpha = (1 - d/6) * 0.5
          edgePosArr[eIdx*6] = nodes[i].x; edgePosArr[eIdx*6+1] = nodes[i].y; edgePosArr[eIdx*6+2] = nodes[i].z
          edgePosArr[eIdx*6+3] = nodes[j].x; edgePosArr[eIdx*6+4] = nodes[j].y; edgePosArr[eIdx*6+5] = nodes[j].z
          edgeColArr[eIdx*6] = nodes[i].c.r * alpha; edgeColArr[eIdx*6+1] = nodes[i].c.g * alpha; edgeColArr[eIdx*6+2] = nodes[i].c.b * alpha
          edgeColArr[eIdx*6+3] = nodes[j].c.r * alpha; edgeColArr[eIdx*6+4] = nodes[j].c.g * alpha; edgeColArr[eIdx*6+5] = nodes[j].c.b * alpha
          eIdx++
        }
      }
    }
    while (eIdx < EDGE_MAX) {
      edgePosArr[eIdx*6] = edgePosArr[eIdx*6+1] = edgePosArr[eIdx*6+2] = 9999
      edgePosArr[eIdx*6+3] = edgePosArr[eIdx*6+4] = edgePosArr[eIdx*6+5] = 9999
      eIdx++
    }
    ;(edgeGeo.attributes.position as THREE.BufferAttribute).needsUpdate = true
    ;(edgeGeo.attributes.color as THREE.BufferAttribute).needsUpdate = true

    const rPosArr = (receiptGeo.attributes.position as THREE.BufferAttribute).array as Float32Array
    const rColArr = (receiptGeo.attributes.color as THREE.BufferAttribute).array as Float32Array
    if (frameCount.current % 8 === 0) {
      const ri = data.receipts.findIndex(r => !r.active)
      if (ri !== -1) {
        const ni = Math.floor(Math.random() * NODE_COUNT)
        data.receipts[ri].active = true
        data.receipts[ri].x = nodes[ni].x; data.receipts[ri].y = nodes[ni].y; data.receipts[ri].z = nodes[ni].z
        data.receipts[ri].life = 0
        data.receipts[ri].vx = (Math.random() - 0.5) * 0.08
        data.receipts[ri].vy = 0.03 + Math.random() * 0.05
        data.receipts[ri].vz = (Math.random() - 0.5) * 0.04
      }
    }
    for (let i = 0; i < RECEIPT_COUNT; i++) {
      const r = data.receipts[i]
      if (!r.active) { rPosArr[i*3] = 9999; continue }
      r.life++
      r.x += r.vx; r.y += r.vy; r.z += r.vz
      const t = r.life / r.maxLife
      rPosArr[i*3] = r.x; rPosArr[i*3+1] = r.y; rPosArr[i*3+2] = r.z
      rColArr[i*3] = 0.2 * t; rColArr[i*3+1] = 0.8 * t; rColArr[i*3+2] = 0.6 * t
      if (r.life >= r.maxLife) r.active = false
    }
    ;(receiptGeo.attributes.position as THREE.BufferAttribute).needsUpdate = true
    ;(receiptGeo.attributes.color as THREE.BufferAttribute).needsUpdate = true
  })

  return (
    <>
      <points ref={meshRef} geometry={nodeGeo}>
        <pointsMaterial size={0.3} vertexColors transparent opacity={0.95} sizeAttenuation />
      </points>
      <lineSegments ref={linesRef} geometry={edgeGeo}>
        <lineBasicMaterial vertexColors transparent opacity={0.2} />
      </lineSegments>
      <points ref={receiptsRef} geometry={receiptGeo}>
        <pointsMaterial size={0.15} vertexColors transparent opacity={0.85} sizeAttenuation />
      </points>
      <mesh position={[0, 0, 0]}>
        <sphereGeometry args={[1.4, 32, 32]} />
        <meshBasicMaterial color={0xFF6B35} transparent opacity={0.06} />
      </mesh>
      <mesh rotation={[Math.PI * 0.25, 0, 0]} position={[0, 0, 0]}>
        <torusGeometry args={[9, 0.018, 4, 200]} />
        <meshBasicMaterial color={0x00D4FF} transparent opacity={0.12} />
      </mesh>
      <mesh rotation={[Math.PI * 0.35, 0, Math.PI * 0.15]} position={[0, 0, 0]}>
        <torusGeometry args={[12, 0.012, 4, 200]} />
        <meshBasicMaterial color={0xFF6B35} transparent opacity={0.08} />
      </mesh>
      <mesh rotation={[Math.PI * 0.5, Math.PI * 0.3, 0]} position={[0, 0, 0]}>
        <torusGeometry args={[6, 0.01, 4, 180]} />
        <meshBasicMaterial color={0x5B7CF5} transparent opacity={0.06} />
      </mesh>
    </>
  )
}

export const HeroCanvas: React.FC = () => {
  return (
    <div className="ff-hero-canvas-wrapper">
      <Canvas
        camera={{ position: [0, 0, 28], fov: 60, near: 0.1, far: 1000 }}
        gl={{ antialias: true, alpha: true }}
        style={{ position: 'absolute', top: 0, left: 0, width: '100%', height: '100%', zIndex: 0 }}
        onCreated={({ gl }) => { gl.setClearColor(0x000000, 0) }}
      >
        <ParticleNetwork />
      </Canvas>
    </div>
  )
}
