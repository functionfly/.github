import { useEffect, useRef } from "react";

const COLORS = {
  indigo: "99, 102, 241",
  violet: "139, 92, 246",
  fuchsia: "217, 70, 239",
  pink: "244, 114, 182",
};

interface Vec2 {
  x: number;
  y: number;
}

interface Particle {
  edgeIndex: number;
  t: number;
  speed: number;
  trail: Vec2[];
  size: number;
}

interface FloatOrb {
  x: number;
  y: number;
  vx: number;
  vy: number;
  radius: number;
  phase: number;
  hue: 0 | 1 | 2; // indigo, violet, fuchsia
}

interface Spark {
  x: number;
  y: number;
  phase: number;
  radius: number;
}

function quadraticPoint(p0: Vec2, p1: Vec2, p2: Vec2, t: number): Vec2 {
  const mt = 1 - t;
  return {
    x: mt * mt * p0.x + 2 * mt * t * p1.x + t * t * p2.x,
    y: mt * mt * p0.y + 2 * mt * t * p1.y + t * t * p2.y,
  };
}

export function AuthHeroAnimation() {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    let animationId: number;
    let particles: Particle[] = [];
    let floatOrbs: FloatOrb[] = [];
    let sparks: Spark[] = [];
    const nodes: Vec2[] = [];
    const nodeDrift: Vec2[] = [];
    const edges: [number, number][] = [];
    const controlOffsets: Vec2[] = [];
    const edgeSpeeds: number[] = [];
    const TRAIL_LENGTH = 16;
    const NODE_COUNT = 26;
    const EDGE_PARTICLES_PER_EDGE = 4;
    const pad = 0.04;

    function buildGraph(w: number, h: number) {
      const cx = w * 0.5;
      const cy = h * 0.5;
      const rx = w * (0.5 - pad);
      const ry = h * (0.5 - pad);

      nodes.length = 0;
      nodeDrift.length = 0;
      const positions: Vec2[] = [
        { x: w * pad + 0, y: h * pad },
        { x: w * (0.25), y: h * pad + ry * 0.1 },
        { x: cx, y: h * pad },
        { x: w * (0.75), y: h * pad + ry * 0.1 },
        { x: w * (1 - pad), y: h * pad + ry * 0.3 },
        { x: w * (1 - pad), y: cy },
        { x: w * (1 - pad), y: h * (1 - pad) - ry * 0.2 },
        { x: w * 0.78, y: h * (1 - pad) },
        { x: cx + rx * 0.2, y: h * (1 - pad) },
        { x: cx - rx * 0.15, y: h * (1 - pad) },
        { x: w * 0.22, y: h * (1 - pad) },
        { x: w * pad, y: h * (1 - pad) - ry * 0.2 },
        { x: w * pad, y: cy },
        { x: w * pad, y: h * pad + ry * 0.4 },
        { x: cx - rx * 0.6, y: cy - ry * 0.5 },
        { x: cx - rx * 0.2, y: cy - ry * 0.3 },
        { x: cx + rx * 0.25, y: cy - ry * 0.4 },
        { x: cx + rx * 0.6, y: cy - ry * 0.2 },
        { x: cx + rx * 0.5, y: cy + ry * 0.3 },
        { x: cx, y: cy + ry * 0.2 },
        { x: cx - rx * 0.4, y: cy + ry * 0.4 },
        { x: cx - rx * 0.35, y: cy },
        { x: cx, y: cy },
        { x: cx + rx * 0.2, y: cy },
        { x: cx - rx * 0.5, y: cy - ry * 0.1 },
        { x: cx + rx * 0.35, y: cy + ry * 0.1 },
      ];
      for (let i = 0; i < NODE_COUNT; i++) {
        nodes.push(positions[i] ?? { x: cx, y: cy });
        nodeDrift.push({
          x: (Math.random() - 0.5) * 10,
          y: (Math.random() - 0.5) * 10,
        });
      }

      edges.length = 0;
      controlOffsets.length = 0;
      edgeSpeeds.length = 0;
      const pairs: [number, number][] = [
        [0, 1], [1, 2], [2, 3], [3, 4], [4, 5], [5, 6], [6, 7], [7, 8], [8, 9], [9, 10], [10, 11], [11, 12], [12, 13], [13, 0],
        [2, 16], [3, 17], [5, 18], [6, 8], [12, 14], [13, 1],
        [14, 15], [15, 16], [16, 17], [17, 18], [18, 19], [19, 20], [20, 21], [21, 14],
        [15, 22], [16, 23], [17, 23], [18, 23], [19, 22], [20, 22], [21, 22],
        [22, 23], [22, 24], [23, 25], [24, 25], [24, 20], [25, 19],
        [1, 14], [4, 17], [7, 19], [10, 20], [13, 21],
      ];
      for (const [a, b] of pairs) {
        if (a < NODE_COUNT && b < NODE_COUNT) {
          edges.push([a, b]);
          controlOffsets.push({
            x: (Math.random() - 0.5) * w * 0.14,
            y: (Math.random() - 0.5) * h * 0.14,
          });
          edgeSpeeds.push(0.0005 + Math.random() * 0.001);
        }
      }

      particles = [];
      for (let ei = 0; ei < edges.length; ei++) {
        for (let k = 0; k < EDGE_PARTICLES_PER_EDGE; k++) {
          particles.push({
            edgeIndex: ei,
            t: Math.random(),
            speed: edgeSpeeds[ei] * (0.6 + Math.random() * 0.8),
            trail: [],
            size: 2.5 + Math.random() * 2,
          });
        }
      }

      if (floatOrbs.length === 0) {
        for (let i = 0; i < 45; i++) {
          floatOrbs.push({
            x: Math.random() * w,
            y: Math.random() * h,
            vx: (Math.random() - 0.5) * 0.2,
            vy: (Math.random() - 0.5) * 0.2,
            radius: 4 + Math.random() * 18,
            phase: Math.random() * Math.PI * 2,
            hue: (i % 3) as 0 | 1 | 2,
          });
        }
      }

      if (sparks.length === 0) {
        for (let i = 0; i < 60; i++) {
          const [a, b] = edges[Math.floor(Math.random() * edges.length)];
          const p0 = nodes[a];
          const p1 = nodes[b];
          const mt = Math.random();
          sparks.push({
            x: p0.x + (p1.x - p0.x) * mt + (Math.random() - 0.5) * 30,
            y: p0.y + (p1.y - p0.y) * mt + (Math.random() - 0.5) * 30,
            phase: Math.random() * Math.PI * 2,
            radius: 1.5 + Math.random() * 2,
          });
        }
      }
    }

    function resize() {
      const dpr = Math.min(window.devicePixelRatio ?? 1, 2);
      const rect = canvas.getBoundingClientRect();
      const w = rect.width;
      const h = rect.height;
      canvas.width = w * dpr;
      canvas.height = h * dpr;
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      buildGraph(w, h);
    }

    resize();
    window.addEventListener("resize", resize);

    let start: number | null = null;
    function frame(now: number) {
      if (start == null) start = now;
      const elapsed = now - start;
      const w = canvas.getBoundingClientRect().width;
      const h = canvas.getBoundingClientRect().height;
      ctx.clearRect(0, 0, w, h);

      const t = elapsed * 0.001;
      const gridOffset = (elapsed * 0.025) % 32;
      const cx = w * 0.5;
      const cy = h * 0.5;

      // ---- Layer 1: Moving grid ----
      ctx.strokeStyle = `rgba(${COLORS.indigo}, 0.05)`;
      ctx.lineWidth = 0.8;
      const gridStep = 32;
      for (let x = -gridOffset; x < w + gridStep; x += gridStep) {
        ctx.beginPath();
        ctx.moveTo(x, 0);
        ctx.lineTo(x, h);
        ctx.stroke();
      }
      for (let y = -gridOffset; y < h + gridStep; y += gridStep) {
        ctx.beginPath();
        ctx.moveTo(0, y);
        ctx.lineTo(w, y);
        ctx.stroke();
      }

      // ---- Layer 2: Corner & edge ambient orbs (fill the panel) ----
      const cornerSpots: Vec2[] = [
        { x: 0, y: 0 }, { x: w, y: 0 }, { x: w, y: h }, { x: 0, y: h },
        { x: w * 0.5, y: 0 }, { x: w, y: h * 0.5 }, { x: w * 0.5, y: h }, { x: 0, y: h * 0.5 },
      ];
      const orbRadius = Math.max(w, h) * 0.35;
      for (let i = 0; i < cornerSpots.length; i++) {
        const spot = cornerSpots[i];
        const pulse = 0.5 + 0.35 * Math.sin(t * 0.7 + i);
        const g = ctx.createRadialGradient(
          spot.x, spot.y, 0,
          spot.x, spot.y, orbRadius,
        );
        g.addColorStop(0, `rgba(${COLORS.violet}, ${0.08 * pulse})`);
        g.addColorStop(0.5, `rgba(${COLORS.fuchsia}, 0.03)`);
        g.addColorStop(1, "transparent");
        ctx.fillStyle = g;
        ctx.beginPath();
        ctx.arc(spot.x, spot.y, orbRadius, 0, Math.PI * 2);
        ctx.fill();
      }

      // ---- Layer 3: Concentric rotating arcs ----
      for (let ring = 0; ring < 3; ring++) {
        const R = 55 + ring * 38 + 12 * Math.sin(t + ring * 0.8);
        const startAngle = (t * 0.3 + ring * 0.5) % (Math.PI * 2);
        const span = Math.PI * 0.7 + 0.2 * Math.sin(t * 0.5 + ring);
        ctx.beginPath();
        ctx.arc(cx, cy, R, startAngle, startAngle + span);
        ctx.strokeStyle = `rgba(${COLORS.indigo}, ${0.12 + 0.06 * Math.sin(t + ring)})`;
        ctx.lineWidth = 2;
        ctx.stroke();
      }

      // ---- Layer 4: Central hex (larger, breathing) ----
      const corePulse = 0.88 + 0.12 * Math.sin(t * 1.2);
      const coreR = 52 * corePulse;
      ctx.beginPath();
      for (let i = 0; i < 6; i++) {
        const a = (i / 6) * Math.PI * 2 - Math.PI / 6;
        const x = cx + coreR * Math.cos(a);
        const y = cy + coreR * Math.sin(a);
        if (i === 0) ctx.moveTo(x, y);
        else ctx.lineTo(x, y);
      }
      ctx.closePath();
      const coreGrad = ctx.createRadialGradient(cx, cy, 0, cx, cy, coreR * 2.2);
      coreGrad.addColorStop(0, `rgba(${COLORS.indigo}, ${0.15 * corePulse})`);
      coreGrad.addColorStop(0.4, `rgba(${COLORS.violet}, 0.06)`);
      coreGrad.addColorStop(1, "transparent");
      ctx.fillStyle = coreGrad;
      ctx.fill();
      ctx.strokeStyle = `rgba(${COLORS.indigo}, ${0.25 * corePulse})`;
      ctx.lineWidth = 2;
      ctx.stroke();

      // ---- Layer 5: Edge glow (thick soft pass) ----
      for (let i = 0; i < edges.length; i++) {
        const [a, b] = edges[i];
        const driftA = 0.5 + 0.5 * Math.sin(t + a * 0.5);
        const driftB = 0.5 + 0.5 * Math.sin(t * 1.1 + b * 0.5);
        const p0 = { x: nodes[a].x + nodeDrift[a].x * driftA, y: nodes[a].y + nodeDrift[a].y * driftA };
        const p1 = { x: nodes[b].x + nodeDrift[b].y * driftB, y: nodes[b].y + nodeDrift[b].x * driftB };
        const ctrl = controlOffsets[i];
        const wave = Math.sin(t * 0.8 + i * 0.7) * 18;
        const mid = { x: (p0.x + p1.x) * 0.5 + ctrl.x + wave, y: (p0.y + p1.y) * 0.5 + ctrl.y + wave * 0.6 };
        ctx.beginPath();
        ctx.moveTo(p0.x, p0.y);
        ctx.quadraticCurveTo(mid.x, mid.y, p1.x, p1.y);
        ctx.strokeStyle = `rgba(${COLORS.violet}, 0.08)`;
        ctx.lineWidth = 12;
        ctx.lineCap = "round";
        ctx.stroke();
      }

      // ---- Layer 6: Edges (main stroke) ----
      for (let i = 0; i < edges.length; i++) {
        const [a, b] = edges[i];
        const driftA = 0.5 + 0.5 * Math.sin(t + a * 0.5);
        const driftB = 0.5 + 0.5 * Math.sin(t * 1.1 + b * 0.5);
        const p0 = { x: nodes[a].x + nodeDrift[a].x * driftA, y: nodes[a].y + nodeDrift[a].y * driftA };
        const p1 = { x: nodes[b].x + nodeDrift[b].y * driftB, y: nodes[b].y + nodeDrift[b].x * driftB };
        const ctrl = controlOffsets[i];
        const wave = Math.sin(t * 0.8 + i * 0.7) * 18;
        const mid = { x: (p0.x + p1.x) * 0.5 + ctrl.x + wave, y: (p0.y + p1.y) * 0.5 + ctrl.y + wave * 0.6 };
        const gradient = ctx.createLinearGradient(p0.x, p0.y, p1.x, p1.y);
        gradient.addColorStop(0, `rgba(${COLORS.indigo}, 0.4)`);
        gradient.addColorStop(0.5, `rgba(${COLORS.violet}, 0.45)`);
        gradient.addColorStop(1, `rgba(${COLORS.fuchsia}, 0.4)`);
        ctx.beginPath();
        ctx.moveTo(p0.x, p0.y);
        ctx.quadraticCurveTo(mid.x, mid.y, p1.x, p1.y);
        ctx.strokeStyle = gradient;
        ctx.lineWidth = 2;
        ctx.globalAlpha = 0.7 + 0.2 * Math.sin(t + i * 0.3);
        ctx.stroke();
        ctx.beginPath();
        ctx.moveTo(p0.x, p0.y);
        ctx.quadraticCurveTo(mid.x, mid.y, p1.x, p1.y);
        ctx.strokeStyle = "rgba(255,255,255,0.08)";
        ctx.lineWidth = 0.8;
        ctx.stroke();
        ctx.globalAlpha = 1;
      }

      // ---- Layer 7: Particle trails and particles ----
      for (const p of particles) {
        const [a, b] = edges[p.edgeIndex];
        const driftA = 0.5 + 0.5 * Math.sin(t + a * 0.5);
        const driftB = 0.5 + 0.5 * Math.sin(t * 1.1 + b * 0.5);
        const p0 = { x: nodes[a].x + nodeDrift[a].x * driftA, y: nodes[a].y + nodeDrift[a].y * driftA };
        const p1 = { x: nodes[b].x + nodeDrift[b].y * driftB, y: nodes[b].y + nodeDrift[b].x * driftB };
        const ctrl = controlOffsets[p.edgeIndex];
        const wave = Math.sin(t * 0.8 + p.edgeIndex * 0.7) * 18;
        const mid = { x: (p0.x + p1.x) * 0.5 + ctrl.x + wave, y: (p0.y + p1.y) * 0.5 + ctrl.y + wave * 0.6 };
        const pos = quadraticPoint(p0, mid, p1, p.t);

        p.trail.push({ ...pos });
        if (p.trail.length > TRAIL_LENGTH) p.trail.shift();

        for (let i = 0; i < p.trail.length; i++) {
          const pt = p.trail[i];
          const alpha = (i / p.trail.length) * 0.4;
          const r = (i / p.trail.length) * 10 + 2;
          const trailGrad = ctx.createRadialGradient(pt.x, pt.y, 0, pt.x, pt.y, r);
          trailGrad.addColorStop(0, `rgba(${COLORS.fuchsia}, ${alpha})`);
          trailGrad.addColorStop(1, "transparent");
          ctx.beginPath();
          ctx.arc(pt.x, pt.y, r, 0, Math.PI * 2);
          ctx.fillStyle = trailGrad;
          ctx.fill();
        }

        const gradient = ctx.createRadialGradient(pos.x, pos.y, 0, pos.x, pos.y, 24);
        gradient.addColorStop(0, `rgba(${COLORS.fuchsia}, 0.75)`);
        gradient.addColorStop(0.4, `rgba(${COLORS.violet}, 0.2)`);
        gradient.addColorStop(1, "transparent");
        ctx.beginPath();
        ctx.arc(pos.x, pos.y, 24, 0, Math.PI * 2);
        ctx.fillStyle = gradient;
        ctx.fill();
        ctx.beginPath();
        ctx.arc(pos.x, pos.y, p.size, 0, Math.PI * 2);
        ctx.fillStyle = `rgba(255,255,255,${0.8 + 0.15 * Math.sin(t * 2)})`;
        ctx.fill();

        p.t += p.speed;
        if (p.t > 1) p.t = 0;
      }

      // ---- Layer 8: Twinkling sparks (small dots) ----
      const hueKeys = [COLORS.indigo, COLORS.violet, COLORS.fuchsia] as const;
      for (const s of sparks) {
        const twinkle = 0.3 + 0.5 * Math.sin(t * 2.5 + s.phase);
        const hue = hueKeys[Math.floor(s.phase * 3) % 3];
        ctx.beginPath();
        ctx.arc(s.x, s.y, s.radius, 0, Math.PI * 2);
        ctx.fillStyle = `rgba(${hue}, ${twinkle})`;
        ctx.fill();
      }

      // ---- Layer 9: Floating orbs (no text) ----
      for (const o of floatOrbs) {
        o.x += o.vx;
        o.y += o.vy;
        if (o.x < -50 || o.x > w + 50) o.vx *= -1;
        if (o.y < -50 || o.y > h + 50) o.vy *= -1;
        const pulse = 0.4 + 0.35 * Math.sin(t * 1.2 + o.phase);
        const hue = [COLORS.indigo, COLORS.violet, COLORS.fuchsia][o.hue];
        const g = ctx.createRadialGradient(o.x, o.y, 0, o.x, o.y, o.radius * 2);
        g.addColorStop(0, `rgba(${hue}, ${0.25 * pulse})`);
        g.addColorStop(0.6, `rgba(${hue}, 0.05)`);
        g.addColorStop(1, "transparent");
        ctx.beginPath();
        ctx.arc(o.x, o.y, o.radius * 2, 0, Math.PI * 2);
        ctx.fillStyle = g;
        ctx.fill();
      }

      // ---- Layer 10: Main nodes (glow + core) ----
      for (let i = 0; i < nodes.length; i++) {
        const drift = 0.5 + 0.5 * Math.sin(t + i * 0.4);
        const n = {
          x: nodes[i].x + nodeDrift[i].x * drift,
          y: nodes[i].y + nodeDrift[i].y * drift,
        };
        const phase = t * 1.5 + i * 0.6;
        const nodePulse = 0.65 + 0.35 * Math.sin(phase);

        const rOuter = 32 + 12 * nodePulse;
        const gradient = ctx.createRadialGradient(n.x, n.y, 0, n.x, n.y, rOuter);
        gradient.addColorStop(0, `rgba(${COLORS.indigo}, ${0.55 * nodePulse})`);
        gradient.addColorStop(0.35, `rgba(${COLORS.violet}, ${0.22 * nodePulse})`);
        gradient.addColorStop(0.7, `rgba(${COLORS.fuchsia}, 0.06)`);
        gradient.addColorStop(1, "transparent");
        ctx.beginPath();
        ctx.arc(n.x, n.y, rOuter, 0, Math.PI * 2);
        ctx.fillStyle = gradient;
        ctx.fill();

        ctx.beginPath();
        ctx.arc(n.x, n.y, 7, 0, Math.PI * 2);
        ctx.fillStyle = `rgba(255,255,255,${0.65 + 0.25 * Math.sin(phase)})`;
        ctx.fill();
      }

      animationId = requestAnimationFrame(frame);
    }

    animationId = requestAnimationFrame(frame);

    return () => {
      window.removeEventListener("resize", resize);
      cancelAnimationFrame(animationId);
    };
  }, []);

  return (
    <canvas
      ref={canvasRef}
      className="absolute inset-0 w-full h-full"
      aria-hidden
    />
  );
}
