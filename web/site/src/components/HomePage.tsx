import gsap from "gsap";
import React, { useCallback, useRef } from "react";
import Typewriter from "typewriter-effect";
import "../styles/sc-main.css";
import { Card } from "./containment/Card";
import { Chamber } from "./containment/Chamber";
import { CornerBrace } from "./containment/CornerBrace";
import { FrameButton } from "./containment/FrameButton";
import { PageGrid } from "./containment/PageGrid";
import { SealedButton } from "./containment/SealedButton";
import { StatusPill } from "./containment/StatusPill";
import { TrustSeal } from "./containment/TrustSeal";
import DemoPlayground from "./DemoPlayground";

const Container: React.FC<{ children: React.ReactNode; narrow?: boolean }> = ({
  children,
  narrow,
}) => (
  <div
    style={{
      maxWidth: narrow ? 720 : 1100,
      margin: "0 auto",
      padding: "0 var(--space-4)",
    }}
  >
    {children}
  </div>
);
const Section: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <section style={{ padding: "var(--space-8) 0", background: "var(--bg)" }}>
    {children}
  </section>
);
const SectionTitle: React.FC<{
  children: React.ReactNode;
  id?: string;
  lead?: React.ReactNode;
}> = ({ children, id, lead }) => (
  <div id={id} style={{ textAlign: "center", marginBottom: "var(--space-7)" }}>
    <h2
      style={{
        fontFamily: "var(--font-display)",
        fontSize: 36,
        fontWeight: 700,
        letterSpacing: "-0.005em",
        color: "var(--text)",
        marginBottom: lead ? "var(--space-4)" : 0,
      }}
    >
      {children}
    </h2>
    {lead && (
      <p
        style={{
          color: "var(--text-dim)",
          maxWidth: 640,
          margin: "0 auto",
          lineHeight: 1.6,
        }}
      >
        {lead}
      </p>
    )}
  </div>
);

const ArrowIcon = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
    <path
      d="M3 8L13 8M9 4L13 8L9 12"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
);

const DocsIcon = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
    <path
      d="M7 1L7 15M1 8L15 8"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
    />
  </svg>
);

const CheckIcon = () => <span style={{ color: "var(--status-ok)" }}>✓</span>;

const sampleExecutions = [
  {
    id: "rx_3f8a91c",
    function: "fetch_weather_api",
    duration: "1.92ms",
    status: "success" as const,
  },
  {
    id: "rx_7c2d4e1",
    function: "send_email",
    duration: "3.21ms",
    status: "success" as const,
  },
  {
    id: "rx_a91f3b8",
    function: "validate_schema",
    duration: "0.84ms",
    status: "success" as const,
  },
  {
    id: "rx_2b7d41e",
    function: "process_payment",
    duration: "1.04ms",
    status: "success" as const,
  },
  {
    id: "rx_5c9f13a",
    function: "call_external_api",
    duration: "2.67ms",
    status: "success" as const,
  },
  {
    id: "rx_e4a82f6",
    function: "ocr_document",
    duration: "12.30ms",
    status: "success" as const,
  },
  {
    id: "rx_9d3c701",
    function: "vector_search",
    duration: "1.76ms",
    status: "success" as const,
  },
  {
    id: "rx_1b6e9a4",
    function: "run_code_sandbox",
    duration: "6.48ms",
    status: "success" as const,
  },
  {
    id: "rx_84f2c1d",
    function: "parse_csv",
    duration: "0.51ms",
    status: "success" as const,
  },
  {
    id: "rx_6a9e3b2",
    function: "generate_embedding",
    duration: "4.13ms",
    status: "success" as const,
  },
  {
    id: "rx_c1f7d83",
    function: "resize_image",
    duration: "2.05ms",
    status: "success" as const,
  },
  {
    id: "rx_3e8b94a",
    function: "translate_text",
    duration: "3.89ms",
    status: "success" as const,
  },
  {
    id: "rx_72d4f6c",
    function: "geocode_address",
    duration: "1.31ms",
    status: "success" as const,
  },
  {
    id: "rx_b5a91e7",
    function: "verify_signature",
    duration: "0.62ms",
    status: "success" as const,
  },
  {
    id: "rx_4f2c8d3",
    function: "webhook_dispatch",
    duration: "1.18ms",
    status: "warning" as const,
  },
  {
    id: "rx_9c3e7a1",
    function: "query_database",
    duration: "2.94ms",
    status: "success" as const,
  },
];

const ReceiptStrip: React.FC = () => {
  // Map sampleExecutions to the receipt format
  const receipts = sampleExecutions.map((e) => ({
    id: e.id,
    fn: e.function,
    status: e.status === "success" ? ("live" as const) : ("warning" as const),
    t: e.status === "success" ? `✓ ${e.duration}` : `⚠ ${e.duration}`,
  }));
  const doubled = [...receipts, ...receipts];

  return (
    <div
      style={{
        background: "var(--panel)",
        borderTop: "1px solid var(--panel-edge)",
        borderBottom: "1px solid var(--panel-edge)",
        padding: "var(--space-3) 0",
        overflow: "hidden",
      }}
    >
      <div
        style={{
          display: "flex",
          gap: "var(--space-4)",
          animation: "receipt-scroll 15s linear infinite",
          width: "max-content",
        }}
      >
        {doubled.map((r, i) => (
          <div
            key={i}
            style={{
              display: "inline-flex",
              alignItems: "center",
              gap: "var(--space-2)",
              padding: "var(--space-2) var(--space-3)",
              background: "var(--panel-raised)",
              border: "1px solid var(--panel-edge)",
              borderRadius: "var(--radius-sm)",
              fontFamily: "var(--font-mono)",
              fontSize: 11,
            }}
          >
            <span
              style={{
                width: 6,
                height: 6,
                borderRadius: "50%",
                background:
                  r.status === "live"
                    ? "var(--status-ok)"
                    : "var(--status-pending)",
                boxShadow:
                  r.status === "live"
                    ? "0 0 8px rgba(143,255,208,0.6)"
                    : "none",
              }}
            />
            <span style={{ color: "var(--text-dim)" }}>{r.id}</span>
            <span style={{ color: "var(--text)" }}>{r.fn}</span>
            <span
              style={{
                color:
                  r.status === "live"
                    ? "var(--status-ok)"
                    : "var(--status-pending)",
              }}
            >
              {r.t}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
};

const HomePage: React.FC<{ authOrigin: string; docsUrl: string }> = ({
  authOrigin,
  docsUrl,
}) => {
  const receiptCardRef = useRef<HTMLDivElement>(null);

  const handleReceiptMouseMove = useCallback(
    (e: React.MouseEvent<HTMLDivElement>) => {
      const card = receiptCardRef.current;
      if (!card) return;
      const rect = card.getBoundingClientRect();
      const x = e.clientX - rect.left;
      const y = e.clientY - rect.top;
      const centerX = rect.width / 2;
      const centerY = rect.height / 2;
      const rotateX = ((y - centerY) / centerY) * -8;
      const rotateY = ((x - centerX) / centerX) * 8;
      gsap.to(card, {
        rotateX,
        rotateY,
        duration: 0.4,
        ease: "power2.out",
        transformPerspective: 1000,
      });
    },
    [],
  );

  const handleReceiptMouseEnter = useCallback(() => {
    const card = receiptCardRef.current;
    if (!card) return;
    gsap.to(card, {
      scale: 1.02,
      duration: 0.3,
      ease: "power2.out",
    });
  }, []);

  const handleReceiptMouseLeave = useCallback(() => {
    const card = receiptCardRef.current;
    if (!card) return;
    gsap.to(card, {
      rotateX: 0,
      rotateY: 0,
      scale: 1,
      duration: 0.5,
      ease: "power3.out",
    });
  }, []);

  return (
    <>
      <PageGrid />

      <main>
        <Section>
          <Container>
            <Chamber variant="ribs" style={{ textAlign: "center" }}>
              <CornerBrace position="tl" />
              <CornerBrace position="br" />
              <div
                style={{
                  fontFamily: "var(--font-mono)",
                  fontSize: 11,
                  letterSpacing: "0.08em",
                  textTransform: "uppercase",
                  color: "var(--status-ok)",
                  marginBottom: "var(--space-5)",
                  display: "inline-flex",
                  alignItems: "center",
                  gap: "var(--space-2)",
                }}
              >
                <span
                  style={{
                    width: 8,
                    height: 8,
                    borderRadius: "50%",
                    background: "var(--status-ok)",
                    boxShadow: "0 0 12px var(--status-ok)",
                  }}
                />
                Infrastructure layer
                <StatusPill status="live" label="Live" />
              </div>
              <h1
                style={{
                  minHeight: "2.6em",
                  lineHeight: "1.2",
                  margin: 0,
                }}
              >
                <Typewriter
                  onInit={(typewriter) => {
                    typewriter
                      .typeString("Agents execute.<br/>")
                      .typeString(
                        '<span style="color: var(--accent)">You verify.</span>',
                      )
                      .pauseFor(1500)
                      .deleteAll()
                      .start();
                  }}
                />
              </h1>
              <p
                style={{
                  fontSize: 17,
                  lineHeight: 1.7,
                  color: "var(--text-dim)",
                  maxWidth: 640,
                  margin: "0 auto var(--space-6)",
                }}
              >
                The{" "}
                <strong style={{ color: "var(--text)" }}>
                  execution and trust layer
                </strong>{" "}
                for AI agents — built on MCP and A2A, backed by a marketplace of
                verified, reusable functions. Every call audited. Every result
                provable.
              </p>
              <div
                style={{
                  display: "flex",
                  gap: "var(--space-3)",
                  flexWrap: "wrap",
                  justifyContent: "center",
                  marginBottom: "var(--space-6)",
                }}
              >
                <SealedButton
                  size="lg"
                  iconRight={<ArrowIcon />}
                  onClick={() =>
                    (window.location.href = `${authOrigin}/signup`)
                  }
                >
                  Deploy your first function
                </SealedButton>
                <FrameButton
                  size="lg"
                  iconLeft={<DocsIcon />}
                  onClick={() => window.open(docsUrl, "_blank")}
                >
                  View Documentation
                </FrameButton>
              </div>
              <div
                style={{
                  display: "flex",
                  justifyContent: "center",
                  gap: "var(--space-8)",
                  borderTop: "1px solid var(--panel-edge)",
                  flexWrap: "wrap",
                }}
              >
                {[
                  { v: "248K+", l: "executions / day" },
                  { v: "<2ms", l: "cold start" },
                  { v: "100%", l: "receipt coverage" },
                ].map((g) => (
                  <div
                    key={g.l}
                    style={{
                      padding: "var(--space-5) var(--space-4) var(--space-6)",
                      textAlign: "center",
                    }}
                  >
                    <div
                      style={{
                        fontFamily: "var(--font-mono)",
                        fontSize: 26,
                        fontWeight: 500,
                        color: "var(--text)",
                        display: "flex",
                        alignItems: "center",
                        gap: "var(--space-2)",
                      }}
                    >
                      <span
                        style={{
                          width: 5,
                          height: 5,
                          borderRadius: "50%",
                          background: "var(--status-ok)",
                          boxShadow: "0 0 6px rgba(143,255,208,0.6)",
                        }}
                      />
                      {g.v}
                    </div>
                    <div
                      style={{
                        fontFamily: "var(--font-mono)",
                        fontSize: 11,
                        fontWeight: 500,
                        letterSpacing: "0.06em",
                        textTransform: "uppercase",
                        color: "var(--text-faint)",
                        marginTop: "var(--space-1)",
                      }}
                    >
                      {g.l}
                    </div>
                  </div>
                ))}
              </div>
            </Chamber>
          </Container>
        </Section>

        <ReceiptStrip />

        <Section>
          <Container>
            <SectionTitle lead="Precision infrastructure. Zero-trust execution.">
              What's under the hood
            </SectionTitle>
            <div
              style={{
                display: "grid",
                gridTemplateColumns: "repeat(auto-fit, minmax(280px, 1fr))",
                gap: "var(--space-5)",
                marginTop: "var(--space-7)",
              }}
            >
              {[
                {
                  icon: "🔒",
                  title: "Sandboxed Execution",
                  desc: "Every function runs in an isolated gVisor container. No shared memory, no escape vectors, no trust assumed.",
                  tag: "gVisor · runsc",
                },
                {
                  icon: "📜",
                  title: "Execution Receipts",
                  desc: "Cryptographically signed receipts for every call. Immutable, timestamped, verifiable by any party downstream.",
                  tag: "SHA-256 · Signed",
                },
                {
                  icon: "⚡",
                  title: "MCP + A2A Native",
                  desc: "First-class Model Context Protocol and Agent-to-Agent support. Drop into any AI orchestration layer without rewiring.",
                  tag: "MCP · A2A",
                },
                {
                  icon: "🛒",
                  title: "Developer Marketplace",
                  desc: "Publish, discover, and monetize sandboxed functions. Every listing is a verified, execution-ready artifact.",
                  tag: "Marketplace",
                },
                {
                  icon: "🌐",
                  title: "Auth Proxy + Routing",
                  desc: "Centralized auth and intelligent routing across agent calls. The most direct, secure path between your agents and the open internet.",
                  tag: "Auth · Routing",
                },
                {
                  icon: "📡",
                  title: "Relay Message Bus",
                  desc: "Async-first message bus for agent-to-agent coordination. Fire and forget, or await — your architecture, your rules.",
                  tag: "Async · Relay",
                },
              ].map((f) => (
                <Card key={f.title}>
                  <div
                    style={{
                      fontSize: 28,
                      lineHeight: 1,
                      marginBottom: "var(--space-3)",
                    }}
                  >
                    {f.icon}
                  </div>
                  <h3
                    style={{
                      fontFamily: "var(--font-display)",
                      fontSize: 17,
                      fontWeight: 600,
                      color: "var(--text)",
                      marginBottom: "var(--space-3)",
                    }}
                  >
                    {f.title}
                  </h3>
                  <p
                    style={{
                      fontSize: 14,
                      color: "var(--text-dim)",
                      lineHeight: 1.6,
                      marginBottom: "auto",
                    }}
                  >
                    {f.desc}
                  </p>
                  <span
                    style={{
                      display: "inline-flex",
                      alignItems: "center",
                      fontFamily: "var(--font-mono)",
                      fontSize: 11,
                      fontWeight: 500,
                      letterSpacing: "0.04em",
                      textTransform: "uppercase",
                      color: "var(--status-ok)",
                      background: "rgba(143, 255, 208, 0.06)",
                      border: "1px solid rgba(143, 255, 208, 0.2)",
                      borderRadius: "var(--radius-sm)",
                      padding: "var(--space-1) var(--space-2)",
                      marginTop: "var(--space-4)",
                      width: "fit-content",
                    }}
                  >
                    {f.tag}
                  </span>
                </Card>
              ))}
            </div>
          </Container>
        </Section>

        <Section>
          <Container>
            <SectionTitle lead="Your language. Your runtime. Our sandbox.">
              Language Support
            </SectionTitle>
            <div
              style={{
                display: "flex",
                flexWrap: "wrap",
                justifyContent: "center",
                gap: "var(--space-3)",
                marginTop: "var(--space-6)",
                marginBottom: "var(--space-6)",
              }}
            >
              {["Python", "Go", "Rust", "JavaScript", "WebAssembly"].map(
                (lang) => (
                  <Card key={lang}>
                    <span
                      style={{
                        fontFamily: "var(--font-mono)",
                        fontSize: 14,
                        fontWeight: 500,
                        color: "var(--text)",
                      }}
                    >
                      {lang}
                    </span>
                  </Card>
                ),
              )}
            </div>
            <div
              style={{
                display: "grid",
                gridTemplateColumns: "repeat(auto-fit, minmax(260px, 1fr))",
                gap: "var(--space-5)",
              }}
            >
              <Card>
                <div style={{ display: "flex", gap: "var(--space-4)" }}>
                  <span style={{ fontSize: 20, flexShrink: 0 }}>🔒</span>
                  <div>
                    <strong
                      style={{
                        fontFamily: "var(--font-display)",
                        fontSize: 15,
                        fontWeight: 600,
                        color: "var(--text)",
                        display: "block",
                        marginBottom: "var(--space-1)",
                      }}
                    >
                      Isolated
                    </strong>
                    <p
                      style={{
                        fontSize: 13,
                        color: "var(--text-dim)",
                        lineHeight: 1.5,
                        margin: 0,
                      }}
                    >
                      Wasm + WASI per-request sandbox. Every invocation gets its
                      own clean execution environment.
                    </p>
                  </div>
                </div>
              </Card>
              <Card>
                <div style={{ display: "flex", gap: "var(--space-4)" }}>
                  <span style={{ fontSize: 20, flexShrink: 0 }}>✅</span>
                  <div>
                    <strong
                      style={{
                        fontFamily: "var(--font-display)",
                        fontSize: 15,
                        fontWeight: 600,
                        color: "var(--text)",
                        display: "block",
                        marginBottom: "var(--space-1)",
                      }}
                    >
                      Trust-scored
                    </strong>
                    <p
                      style={{
                        fontSize: 13,
                        color: "var(--text-dim)",
                        lineHeight: 1.5,
                        margin: 0,
                      }}
                    >
                      Execution and routing with cryptographic verification.
                      Zero-trust by default.
                    </p>
                  </div>
                </div>
              </Card>
            </div>
          </Container>
        </Section>

        <Section>
          <Container>
            <SectionTitle lead="Built for every stage. Trusted by all.">
              Who benefits most
            </SectionTitle>
            <div
              style={{
                display: "grid",
                gridTemplateColumns: "repeat(auto-fit, minmax(300px, 1fr))",
                gap: "var(--space-5)",
              }}
            >
              {[
                {
                  icon: "🎯",
                  badge: "Indie",
                  title: "Indie SaaS Founder",
                  subtitle: "Deploying your first app",
                  desc: "You built an amazing product and now need trustworthy agent tooling without enterprise overhead.",
                  features: [
                    "Python, Go, Rust, JS, WASM support",
                    "Wasm + WASI sandbox isolation",
                    "Trust scores and verification tiers",
                    "Pay as you grow",
                  ],
                  tag: "$0 to start",
                },
                {
                  icon: "🚀",
                  badge: "Growth",
                  title: "Growing Startup",
                  subtitle: "Scaling agent + API workloads",
                  desc: "Your team is shipping fast and agents are calling more tools every week.",
                  features: [
                    "Multi-level verification workflows",
                    "Agent swarm orchestration",
                    "Policy engine for safe execution",
                    "Multi-provider routing",
                  ],
                  tag: "Usage-based",
                },
                {
                  icon: "🏢",
                  badge: "Enterprise",
                  title: "Enterprise",
                  subtitle: "Governance and compliance",
                  desc: "Your organization needs auditability, least-privilege access, and consistent trust policies.",
                  features: [
                    "Firecracker MicroVMs",
                    "Multi-level attestations",
                    "Custom SLAs and audit logs",
                    "Multi-team patterns",
                  ],
                  tag: "Contact sales",
                },
              ].map((a) => (
                <Card key={a.title}>
                  <div
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: "var(--space-3)",
                      marginBottom: "var(--space-3)",
                    }}
                  >
                    <span style={{ fontSize: 24 }}>{a.icon}</span>
                    <span
                      style={{
                        fontFamily: "var(--font-mono)",
                        fontSize: 11,
                        fontWeight: 500,
                        letterSpacing: "0.06em",
                        textTransform: "uppercase",
                        padding: "2px 8px",
                        border: "1px solid var(--panel-edge)",
                        borderRadius: "var(--radius-sm)",
                        color: "var(--text-dim)",
                      }}
                    >
                      {a.badge}
                    </span>
                  </div>
                  <h3
                    style={{
                      fontFamily: "var(--font-display)",
                      fontSize: 17,
                      fontWeight: 600,
                      color: "var(--text)",
                      marginBottom: 4,
                    }}
                  >
                    {a.title}
                  </h3>
                  <p
                    style={{
                      fontFamily: "var(--font-mono)",
                      fontSize: 12,
                      color: "var(--text-dim)",
                      marginBottom: "var(--space-3)",
                    }}
                  >
                    {a.subtitle}
                  </p>
                  <p
                    style={{
                      fontSize: 14,
                      color: "var(--text-dim)",
                      lineHeight: 1.6,
                      marginBottom: "var(--space-4)",
                    }}
                  >
                    {a.desc}
                  </p>
                  <ul
                    style={{
                      listStyle: "none",
                      padding: 0,
                      margin: "0 0 var(--space-4) 0",
                    }}
                  >
                    {a.features.map((f) => (
                      <li
                        key={f}
                        style={{
                          display: "flex",
                          alignItems: "center",
                          gap: "var(--space-2)",
                          color: "var(--text-dim)",
                          fontSize: 13,
                          marginBottom: 6,
                        }}
                      >
                        <CheckIcon /> {f}
                      </li>
                    ))}
                  </ul>
                  <span
                    style={{
                      fontFamily: "var(--font-mono)",
                      fontSize: 11,
                      fontWeight: 500,
                      letterSpacing: "0.04em",
                      textTransform: "uppercase",
                      padding: "var(--space-1) var(--space-2)",
                      borderRadius: "var(--radius-sm)",
                      background: "rgba(143, 255, 208, 0.06)",
                      border: "1px solid rgba(143, 255, 208, 0.2)",
                      color: "var(--status-ok)",
                    }}
                  >
                    {a.tag}
                  </span>
                </Card>
              ))}
            </div>
          </Container>
        </Section>

        <Section>
          <Container>
            <SectionTitle lead="From code to trusted agent tool in 4 steps.">
              How it works
            </SectionTitle>
            <div
              style={{
                display: "grid",
                gridTemplateColumns: "repeat(auto-fit, minmax(240px, 1fr))",
                gap: "var(--space-5)",
              }}
            >
              {[
                {
                  num: "01",
                  icon: "✏️",
                  title: "Write",
                  desc: "Implement your function in Python, Go, Rust, JavaScript, or TypeScript. Your code runs in isolation — it cannot phone home or access the filesystem unless you grant capabilities explicitly.",
                },
                {
                  num: "02",
                  icon: "📤",
                  title: "Publish",
                  desc: "Run ff check to verify your setup, then ff deploy. FunctionFly bundles your code, compiles it for the target runtime, signs the artifact, and runs multi-level verification checks.",
                },
                {
                  num: "03",
                  icon: "🔍",
                  title: "Discover",
                  desc: "Agents find your function through manifests and trust filters. Your function's trust score, verification tier, execution history, and capability declarations are all surfaced automatically.",
                },
                {
                  num: "04",
                  icon: "⚡",
                  title: "Execute",
                  desc: "Agents call your function over HTTPS. FunctionFly routes to the right backend, enforces policy limits, runs in a Wasm sandbox, and records every result for trust scoring.",
                },
              ].map((step) => (
                <Card key={step.num}>
                  <div
                    style={{
                      width: 32,
                      height: 32,
                      display: "flex",
                      alignItems: "center",
                      justifyContent: "center",
                      background: "var(--panel)",
                      border: "1px solid var(--accent)",
                      borderRadius: "50%",
                      fontFamily: "var(--font-mono)",
                      fontSize: 14,
                      fontWeight: 700,
                      color: "var(--accent)",
                      marginBottom: "var(--space-3)",
                    }}
                  >
                    {step.num}
                  </div>
                  <div style={{ fontSize: 24, marginBottom: "var(--space-2)" }}>
                    {step.icon}
                  </div>
                  <h3
                    style={{
                      fontFamily: "var(--font-display)",
                      fontSize: 17,
                      fontWeight: 600,
                      color: "var(--text)",
                      marginBottom: "var(--space-2)",
                    }}
                  >
                    {step.title}
                  </h3>
                  <p
                    style={{
                      fontSize: 13,
                      color: "var(--text-dim)",
                      lineHeight: 1.6,
                    }}
                  >
                    {step.desc}
                  </p>
                </Card>
              ))}
            </div>
          </Container>
        </Section>

        <Section>
          <Container>
            <div
              style={{
                display: "grid",
                gridTemplateColumns: "repeat(auto-fit, minmax(400px, 1fr))",
                gap: "var(--space-6)",
                alignItems: "center",
              }}
            >
              <div>
                <SectionTitle>
                  Every call. Signed. Sealed. Provable.
                </SectionTitle>
                <p
                  style={{
                    color: "var(--text-dim)",
                    lineHeight: 1.7,
                    marginBottom: "var(--space-4)",
                  }}
                >
                  When an AI agent executes a function through FunctionFly, it
                  gets back more than a result. It gets a{" "}
                  <strong style={{ color: "var(--text)" }}>
                    cryptographic receipt
                  </strong>{" "}
                  — a tamper-proof artifact that any downstream system can
                  verify.
                </p>
                <p style={{ color: "var(--text-dim)", lineHeight: 1.7 }}>
                  No more "did that really run?" No more trust gaps between
                  agents.
                  <strong style={{ color: "var(--text)" }}>
                    {" "}
                    The receipt is the proof.
                  </strong>
                </p>
              </div>
              <div
                ref={receiptCardRef}
                onMouseMove={handleReceiptMouseMove}
                onMouseEnter={handleReceiptMouseEnter}
                onMouseLeave={handleReceiptMouseLeave}
                style={{
                  transformStyle: "preserve-3d",
                  perspective: 1000,
                  cursor: "default",
                }}
              >
                <Chamber>
                  <CornerBrace position="tl" />
                  <CornerBrace position="br" />
                  <div
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: "var(--space-2)",
                      marginBottom: "var(--space-4)",
                    }}
                  >
                    <div style={{ display: "flex", gap: 4 }}>
                      <span
                        style={{
                          width: 10,
                          height: 10,
                          borderRadius: "50%",
                          background: "#FF2D55",
                        }}
                      />
                      <span
                        style={{
                          width: 10,
                          height: 10,
                          borderRadius: "50%",
                          background: "#FFB800",
                        }}
                      />
                      <span
                        style={{
                          width: 10,
                          height: 10,
                          borderRadius: "50%",
                          background: "#00FF9D",
                        }}
                      />
                    </div>
                    <span
                      style={{
                        fontFamily: "var(--font-mono)",
                        fontSize: 11,
                        color: "var(--text-dim)",
                      }}
                    >
                      execution.receipt.v1
                    </span>
                    <TrustSeal size="small" label="Verified" />
                  </div>
                  <div
                    style={{
                      display: "flex",
                      flexDirection: "column",
                      gap: "var(--space-2)",
                    }}
                  >
                    {[
                      {
                        k: "receipt_id",
                        v: "rcpt_7f2a1b3c9d4e5f8a1b2c3d4e5f6a7b8c",
                        cls: "ff-id",
                      },
                      { k: "function", v: "process_payment" },
                      { k: "agent_id", v: "agt_claude_prod_7" },
                      {
                        k: "sandbox",
                        v: "gvisor/runsc-secure",
                        cls: "ff-cyan-val",
                      },
                      { k: "executed_at", v: "2026-06-12T14:22:01.384Z" },
                      { k: "duration_ms", v: "1.84" },
                      { k: "exit_code", v: "0 · SUCCESS", cls: "ff-ok" },
                      { k: "memory_used", v: "12.3 MB" },
                      { k: "network_egress", v: "blocked", cls: "ff-ok" },
                    ].map((row) => (
                      <div
                        key={row.k}
                        style={{
                          display: "flex",
                          justifyContent: "space-between",
                          gap: "var(--space-3)",
                          fontSize: 13,
                          padding: "var(--space-1) 0",
                          borderBottom: "1px solid var(--panel-edge)",
                        }}
                      >
                        <span
                          style={{
                            fontFamily: "var(--font-mono)",
                            color: "var(--text-faint)",
                          }}
                        >
                          {row.k}
                        </span>
                        <span
                          style={{
                            fontFamily: "var(--font-mono)",
                            color:
                              row.cls === "ff-ok"
                                ? "var(--status-ok)"
                                : row.cls === "ff-cyan-val"
                                  ? "var(--accent)"
                                  : "var(--text)",
                          }}
                        >
                          {row.v}
                        </span>
                      </div>
                    ))}
                    <div
                      style={{
                        fontFamily: "var(--font-mono)",
                        fontSize: 11,
                        color: "var(--text-faint)",
                        marginTop: "var(--space-2)",
                      }}
                    >
                      SHA-256:
                      e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
                    </div>
                  </div>
                </Chamber>
              </div>
            </div>
          </Container>
        </Section>

        <Section>
          <Container>
            <SectionTitle lead="Built for the agentic stack.">
              Protocol Support
            </SectionTitle>
            <div
              style={{
                display: "grid",
                gridTemplateColumns: "repeat(auto-fit, minmax(220px, 1fr))",
                gap: "var(--space-5)",
              }}
            >
              {[
                {
                  name: "MCP",
                  desc: "Model Context Protocol. Native tool registration, schema validation, and sandboxed handler execution.",
                },
                {
                  name: "A2A",
                  desc: "Agent-to-Agent protocol. Structured inter-agent messaging with receipt-backed delivery guarantees.",
                },
                {
                  name: "REST",
                  desc: "Standard HTTP execution endpoints. Works with any agent framework, any language, any runtime.",
                },
                {
                  name: "gRPC",
                  desc: "High-throughput streaming execution for latency-sensitive agentic workloads at scale.",
                },
              ].map((p) => (
                <Card key={p.name}>
                  <div
                    style={{
                      fontFamily: "var(--font-display)",
                      fontSize: 20,
                      fontWeight: 600,
                      color: "var(--text)",
                      marginBottom: "var(--space-3)",
                    }}
                  >
                    {p.name}
                  </div>
                  <p
                    style={{
                      fontSize: 14,
                      color: "var(--text-dim)",
                      lineHeight: 1.6,
                    }}
                  >
                    {p.desc}
                  </p>
                </Card>
              ))}
            </div>
          </Container>
        </Section>

        <DemoPlayground authOrigin={authOrigin} />

        <Section>
          <Container>
            <Chamber variant="ribs">
              <CornerBrace position="tl" />
              <CornerBrace position="br" />
              <div style={{ textAlign: "center" }}>
                <p
                  style={{
                    fontFamily: "var(--font-mono)",
                    fontSize: 11,
                    letterSpacing: "0.08em",
                    textTransform: "uppercase",
                    color: "var(--status-ok)",
                    marginBottom: "var(--space-4)",
                  }}
                >
                  Ready for takeoff
                </p>
                <h2
                  style={{
                    fontFamily: "var(--font-display)",
                    fontSize: "clamp(28px, 4vw, 48px)",
                    fontWeight: 700,
                    letterSpacing: "-0.03em",
                    color: "var(--text)",
                    marginBottom: "var(--space-4)",
                  }}
                >
                  Your agents deserve
                  <br />
                  <span style={{ color: "var(--accent)" }}>
                    better infrastructure.
                  </span>
                </h2>
                <p
                  style={{
                    fontSize: 17,
                    color: "var(--text-dim)",
                    maxWidth: 560,
                    margin: "0 auto var(--space-6)",
                    lineHeight: 1.6,
                  }}
                >
                  Deploy your first sandboxed function in under 5 minutes. No
                  credit card. Full receipt coverage from day one.
                </p>
                <div
                  style={{
                    display: "inline-flex",
                    gap: "var(--space-3)",
                    flexWrap: "wrap",
                    justifyContent: "center",
                  }}
                >
                  <SealedButton
                    size="lg"
                    iconRight={<ArrowIcon />}
                    onClick={() =>
                      (window.location.href = `${authOrigin}/signup`)
                    }
                  >
                    Start building free
                  </SealedButton>
                  <FrameButton
                    size="lg"
                    iconLeft={<DocsIcon />}
                    onClick={() => window.open(docsUrl, "_blank")}
                  >
                    Read the docs
                  </FrameButton>
                </div>
              </div>
            </Chamber>
          </Container>
        </Section>
      </main>
    </>
  );
};

export default HomePage;
