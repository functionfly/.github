import gsap from "gsap";
import React, { useCallback, useRef, useState, useEffect } from "react";
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

const API_BASE = import.meta.env.PUBLIC_API_URL || "https://api.functionfly.com";

interface PlatformStats {
  executions_per_day: string;
  cold_start_ms: string;
  receipt_coverage: string;
}

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
  const [platformStats, setPlatformStats] = useState<PlatformStats | null>(null);

  useEffect(() => {
    const loadStats = async () => {
      try {
        const res = await fetch(`${API_BASE}/v1/platform-stats`);
        if (res.ok) {
          const data = await res.json();
          setPlatformStats(data);
        }
      } catch {
        // silently ignore - fallback to hardcoded values
      }
    };
    loadStats();
  }, []);

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
                  { v: platformStats?.executions_per_day || "248K+", l: "executions / day" },
                  { v: platformStats?.cold_start_ms || "<2ms", l: "cold start" },
                  { v: platformStats?.receipt_coverage || "100%", l: "receipt coverage" },
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
              {[
                { name: "Python", icon: <svg width="24" height="24" viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg"><path fill-rule="evenodd" clip-rule="evenodd" d="M13.0164 2C10.8193 2 9.03825 3.72453 9.03825 5.85185V8.51852H15.9235V9.25926H5.97814C3.78107 9.25926 2 10.9838 2 13.1111L2 18.8889C2 21.0162 3.78107 22.7407 5.97814 22.7407H8.27322V19.4815C8.27322 17.3542 10.0543 15.6296 12.2514 15.6296H19.5956C21.4547 15.6296 22.9617 14.1704 22.9617 12.3704V5.85185C22.9617 3.72453 21.1807 2 18.9836 2H13.0164ZM12.0984 6.74074C12.8589 6.74074 13.4754 6.14378 13.4754 5.40741C13.4754 4.67103 12.8589 4.07407 12.0984 4.07407C11.3378 4.07407 10.7213 4.67103 10.7213 5.40741C10.7213 6.14378 11.3378 6.74074 12.0984 6.74074Z" fill="url(#paint0_linear_87_8204)"/><path fill-rule="evenodd" clip-rule="evenodd" d="M18.9834 30C21.1805 30 22.9616 28.2755 22.9616 26.1482V23.4815L16.0763 23.4815L16.0763 22.7408L26.0217 22.7408C28.2188 22.7408 29.9998 21.0162 29.9998 18.8889V13.1111C29.9998 10.9838 28.2188 9.25928 26.0217 9.25928L23.7266 9.25928V12.5185C23.7266 14.6459 21.9455 16.3704 19.7485 16.3704L12.4042 16.3704C10.5451 16.3704 9.03809 17.8296 9.03809 19.6296L9.03809 26.1482C9.03809 28.2755 10.8192 30 13.0162 30H18.9834ZM19.9015 25.2593C19.1409 25.2593 18.5244 25.8562 18.5244 26.5926C18.5244 27.329 19.1409 27.9259 19.9015 27.9259C20.662 27.9259 21.2785 27.329 21.2785 26.5926C21.2785 25.8562 20.662 25.2593 19.9015 25.2593Z" fill="url(#paint1_linear_87_8204)"/><defs><linearGradient id="paint0_linear_87_8204" x1="12.4809" y1="2" x2="12.4809" y2="22.7407" gradientUnits="userSpaceOnUse"><stop stop-color="#327EBD"/><stop offset="1" stop-color="#1565A7"/></linearGradient><linearGradient id="paint1_linear_87_8204" x1="19.519" y1="9.25928" x2="19.519" y2="30" gradientUnits="userSpaceOnUse"><stop stop-color="#FFDA4B"/><stop offset="1" stop-color="#F9C600"/></linearGradient></defs></svg> },
                { name: "Go", icon: <svg width="24" height="24" viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg"><path fill-rule="evenodd" clip-rule="evenodd" d="M18.1177 14.0442C17.7408 14.1497 17.3586 14.2566 16.9162 14.3768C16.7001 14.438 16.6509 14.4519 16.4498 14.2074C16.2086 13.9194 16.0317 13.7331 15.6939 13.5636C14.6807 13.0384 13.6996 13.1909 12.7829 13.8178C11.6893 14.5632 11.1264 15.6644 11.1425 17.0367C11.1585 18.3921 12.0431 19.5103 13.3137 19.6966C14.4073 19.8491 15.324 19.4425 16.0477 18.5785C16.1924 18.3922 16.3212 18.1887 16.482 17.9516H13.378C13.0402 17.9516 12.9598 17.7314 13.0724 17.4433C13.2815 16.9181 13.6675 16.0372 13.8926 15.5967C13.9409 15.495 14.0535 15.3256 14.2947 15.3256H19.4702C19.7027 14.5496 20.0799 13.8164 20.5831 13.1226C21.7572 11.4961 23.1725 10.649 25.0863 10.2933C26.7268 9.9883 28.2707 10.1577 29.6699 11.1573C30.9405 12.0722 31.7285 13.3089 31.9376 14.9354C32.211 17.2225 31.5838 19.0862 30.0881 20.6787C29.0266 21.8138 27.7239 22.5254 26.2282 22.8473C25.9429 22.9029 25.6576 22.9293 25.3768 22.9553C25.2303 22.9689 25.085 22.9823 24.9416 22.9998C23.478 22.9659 22.1432 22.5254 21.0173 21.5089C20.2256 20.7879 19.6803 19.9019 19.4092 18.8705C19.2211 19.2707 18.9962 19.6539 18.7336 20.0185C17.5756 21.628 16.0638 22.6276 14.15 22.8987C12.5738 23.1189 11.1103 22.797 9.82366 21.7805C8.63353 20.8317 7.95805 19.578 7.78114 18.0194C7.57206 16.1727 8.08671 14.5124 9.14818 13.0554C10.2901 11.4798 11.8019 10.4802 13.6514 10.1244C15.1632 9.8364 16.6106 10.0228 17.9134 10.9546C18.7657 11.5475 19.3769 12.3608 19.779 13.3434C19.8755 13.4959 19.8111 13.5806 19.6181 13.6314C19.0545 13.7822 18.5903 13.9121 18.1177 14.0442ZM28.7581 15.974C28.7613 16.0309 28.7646 16.0909 28.7693 16.1552C28.6889 17.6122 27.9973 18.6965 26.7268 19.3911C25.8744 19.8485 24.9898 19.8994 24.1053 19.4928C22.9473 18.9506 22.3361 17.6122 22.6256 16.2907C22.9795 14.6982 23.9444 13.6986 25.4401 13.3428C26.968 12.9701 28.4316 13.9188 28.7211 15.5961C28.7438 15.7161 28.7505 15.836 28.7581 15.974Z" fill="#00ACD7"/><path d="M2.44461 13.8517C2.41244 13.9025 2.42852 13.9364 2.49285 13.9364L7.2826 13.9534C7.33085 13.9534 7.41126 13.9025 7.44343 13.8517L7.71684 13.4112C7.749 13.3604 7.73292 13.3096 7.66859 13.3096H2.95926C2.89493 13.3096 2.81451 13.3435 2.78235 13.3943L2.44461 13.8517Z" fill="#00ACD7"/><path d="M0.0160829 15.4103C-0.0160829 15.4611 7.45058e-09 15.495 0.0643316 15.495L6.63928 15.4781C6.70361 15.4781 6.76794 15.4442 6.78402 15.3764L6.91269 14.9698C6.92877 14.919 6.8966 14.8682 6.83227 14.8682H0.530735C0.466404 14.8682 0.385989 14.902 0.353823 14.9529L0.0160829 15.4103Z" fill="#00ACD7"/><path d="M3.90813 16.9521C3.87596 17.0029 3.89204 17.0537 3.95638 17.0537L6.43019 17.0707C6.47843 17.0707 6.54277 17.0199 6.54277 16.9521L6.57493 16.5455C6.57493 16.4777 6.54277 16.4269 6.47843 16.4269H4.29412C4.22978 16.4269 4.16545 16.4777 4.13329 16.5285L3.90813 16.9521Z" fill="#00ACD7"/></svg> },
                { name: "Rust", icon: <svg width="24" height="24" viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg"><path fill="#000000" d="M25.763 12.291c0.099-0.098 0.235-0.159 0.385-0.159 0.301 0 0.545 0.244 0.545 0.545s-0.244 0.545-0.545 0.545c-0.301 0-0.545-0.244-0.545-0.545 0-0.15 0.061-0.286 0.159-0.385v0zM15.342 3.906c0.163-0.171 0.392-0.278 0.647-0.278 0.493 0 0.892 0.399 0.892 0.892s-0.399 0.892-0.892 0.892c-0.413 0-0.76-0.28-0.862-0.661l-0.001-0.006c-0.018-0.068-0.029-0.145-0.029-0.225 0-0.238 0.093-0.454 0.245-0.614l-0 0zM26.965 15.901c0 0.273-0.010 0.545-0.030 0.814h-1.125c-0.112 0-0.158 0.073-0.158 0.185v0.516c0 1.216-0.686 1.48-1.287 1.547-0.572 0.065-1.206-0.239-1.284-0.589-0.089-1.265-0.773-2.353-1.772-2.994l-0.015-0.009c1.206-0.593 2.062-1.736 2.246-3.093l0.002-0.021c-0.049-1.229-0.72-2.29-1.705-2.884l-0.016-0.009c-0.667-0.438-1.472-0.719-2.339-0.773l-0.014-0.001h-11.63c1.569-1.749 3.678-2.987 6.063-3.45l0.069-0.011 1.371 1.438c0.148 0.155 0.356 0.252 0.587 0.252 0.218 0 0.415-0.086 0.561-0.226l-0 0 1.533-1.467c3.233 0.622 5.913 2.593 7.475 5.291l0.028 0.053-1.050 2.372c-0.044 0.097-0.069 0.21-0.069 0.329 0 0.329 0.195 0.613 0.477 0.742l0.005 0.002 2.022 0.898c0.035 0.359 0.053 0.721 0.053 1.089zM13.522 14.044v-2.063h3.699c0.191 0 1.349 0.221 1.349 1.087 0 0.719-0.888 0.977-1.618 0.977zM5.106 14.723l1.918-0.853c0.287-0.13 0.483-0.413 0.483-0.742 0-0.12-0.026-0.233-0.072-0.335l0.002 0.005-0.395-0.893h1.554v7.001h-3.134c-0.266-0.899-0.418-1.931-0.418-3 0-0.417 0.023-0.829 0.069-1.234l-0.005 0.050zM6.15 12.247c-0-0.295-0.239-0.534-0.534-0.534s-0.534 0.239-0.534 0.534c0 0.295 0.239 0.534 0.534 0.534s0.534-0.239 0.534-0.534v0zM9.548 26.027c-0.061 0.015-0.13 0.023-0.202 0.023-0.493 0-0.892-0.399-0.892-0.892s0.399-0.892 0.892-0.892c0.493 0 0.892 0.399 0.892 0.892 0 0.096-0.015 0.189-0.043 0.276l0.002-0.006c-0.097 0.3-0.34 0.526-0.643 0.599l-0.006 0.001zM21.937 23.178c-0.051-0.012-0.11-0.018-0.171-0.018-0.388 0-0.713 0.273-0.792 0.638l-0.001 0.005-0.447 2.085c-1.329 0.615-2.884 0.974-4.523 0.974-1.675 0-3.263-0.375-4.684-1.046l0.067 0.028-0.447-2.085c-0.080-0.369-0.404-0.642-0.792-0.642-0.061 0-0.12 0.007-0.177 0.019l0.005-0.001-1.841 0.395c-0.332-0.341-0.644-0.707-0.931-1.093l-0.021-0.029h8.957c0.101 0 0.169-0.018 0.169-0.11v-3.169c0-0.092-0.067-0.11-0.169-0.11h-2.62v-2.009h2.834c0.881 0.009 1.607 0.656 1.741 1.5l0.001 0.010c0.113 0.441 0.359 1.88 0.529 2.34 0.168 0.516 0.854 1.547 1.585 1.547h4.463c0.058-0.001 0.114-0.007 0.168-0.017l-0.006 0.001c-0.326 0.44-0.658 0.828-1.016 1.192l0.001-0.001zM22.365 26.070c1.13-0 2.046-0.917 2.046-2.047s-0.916-2.047-2.047-2.047-2.047 0.916-2.047 2.047c0 1.13 0.916 2.046 2.046 2.047h0zM30.789 15.629l-1.259-0.779q-0.016-0.184-0.035-0.367l1.082-1.008c0.084-0.079 0.136-0.192 0.136-0.316 0-0.184-0.115-0.342-0.277-0.406l-0.003-0.001-1.383-0.517q-0.051-0.179-0.109-0.357l0.863-1.198c0.051-0.070 0.082-0.158 0.082-0.253 0-0.215-0.156-0.393-0.361-0.427l-0.003-0-1.458-0.237q-0.085-0.165-0.176-0.328l0.612-1.345c0.025-0.053 0.039-0.114 0.039-0.18 0-0.090-0.027-0.173-0.075-0.242l0.001 0.001c-0.079-0.117-0.212-0.193-0.362-0.193-0.005 0-0.010 0-0.015 0l0.001-0-1.48 0.052q-0.114-0.144-0.234-0.283l0.34-1.441c0.007-0.030 0.011-0.064 0.011-0.099 0-0.24-0.194-0.434-0.434-0.434-0.035 0-0.069 0.004-0.102 0.012l0.003-0.001-1.441 0.34q-0.141-0.119-0.285-0.234l0.052-1.48c0-0.006 0-0.013 0-0.021 0-0.238-0.193-0.43-0.43-0.43-0.066 0-0.129 0.015-0.185 0.042l0.003-0.001-1.345 0.614c-0.109-0.059-0.218-0.119-0.328-0.176l-0.238-1.459c-0.036-0.207-0.215-0.362-0.429-0.362-0.094 0-0.182 0.030-0.253 0.081l0.001-0.001-1.199 0.863q-0.177-0.057-0.357-0.107l-0.517-1.383c-0.064-0.165-0.222-0.28-0.407-0.28-0.124 0-0.236 0.052-0.316 0.136l-0 0-1.008 1.083q-0.183-0.021-0.367-0.035l-0.779-1.259c-0.078-0.124-0.213-0.205-0.368-0.205s-0.291 0.081-0.367 0.204l-0.001 0.002-0.779 1.259q-0.184 0.016-0.367 0.035l-1.010-1.083c-0.079-0.085-0.192-0.138-0.317-0.138-0.185 0-0.343 0.116-0.405 0.279l-0.001 0.003-0.517 1.383c-0.120 0.034-0.238 0.071-0.357 0.107l-1.198-0.863c-0.070-0.050-0.157-0.080-0.252-0.080-0.215 0-0.393 0.155-0.429 0.36l-0 0.003-0.238 1.459q-0.165 0.085-0.328 0.176l-1.345-0.614c-0.053-0.025-0.115-0.039-0.18-0.039-0.239 0-0.433 0.194-0.433 0.433 0 0.006 0 0.012 0 0.017l-0-0.001 0.052 1.48q-0.144 0.114-0.285 0.234l-1.441-0.341c-0.030-0.007-0.064-0.011-0.099-0.011-0.240 0-0.434 0.194-0.434 0.434 0 0.035 0.004 0.069 0.012 0.102l-0.001-0.003 0.339 1.441c-0.078 0.094-0.157 0.189-0.233 0.283l-1.48-0.052c-0.005-0-0.011-0-0.017-0-0.239 0-0.433 0.194-0.433 0.433 0 0.065 0.014 0.127 0.040 0.183l-0.001-0.003 0.614 1.345q-0.091 0.162-0.176 0.328l-1.457 0.237c-0.207 0.036-0.362 0.214-0.362 0.429 0 0.094 0.030 0.182 0.081 0.253l-0.001-0.001 0.863 1.198q-0.056 0.178-0.109 0.357l-1.383 0.517c-0.165 0.064-0.280 0.222-0.280 0.407 0 0.124 0.052 0.236 0.135 0.316l0 0 1.082 1.008q-0.021 0.183-0.035 0.367l-1.259 0.779c-0.125 0.077-0.208 0.213-0.208 0.368s0.082 0.292 0.206 0.367l0.002 0.001 1.259 0.779c0.010 0.123 0.023 0.245 0.035 0.367l-1.082 1.010c-0.085 0.079-0.138 0.192-0.138 0.317 0 0.185 0.116 0.343 0.279 0.405l0.003 0.001 1.383 0.517c0.034 0.120 0.071 0.239 0.109 0.357l-0.863 1.198c-0.052 0.070-0.083 0.159-0.083 0.254 0 0.215 0.158 0.394 0.364 0.426l0.002 0 1.457 0.237c0.057 0.110 0.115 0.219 0.176 0.328l-0.614 1.345c-0.025 0.053-0.039 0.115-0.039 0.18 0 0.239 0.194 0.433 0.433 0.433 0.006 0 0.012-0 0.017-0l-0.001 0 1.479-0.052c0.077 0.096 0.154 0.191 0.234 0.285l-0.339 1.442c-0.007 0.030-0.011 0.064-0.011 0.099 0 0.239 0.194 0.433 0.433 0.433 0.036 0 0.070-0.004 0.103-0.012l-0.003 0.001 1.441-0.339c0.094 0.080 0.189 0.157 0.285 0.233l-0.052 1.48c-0 0.006-0 0.012-0 0.019 0 0.238 0.193 0.43 0.43 0.43 0.066 0 0.129-0.015 0.185-0.042l-0.003 0.001 1.345-0.612c0.109 0.061 0.218 0.119 0.328 0.176l0.238 1.457c0.036 0.207 0.214 0.363 0.429 0.363 0.094 0 0.181-0.030 0.253-0.081l-0.001 0.001 1.198-0.863q0.178 0.057 0.357 0.109l0.517 1.383c0.062 0.167 0.220 0.283 0.405 0.283 0.125 0 0.238-0.053 0.317-0.139l0-0 1.010-1.082c0.121 0.014 0.244 0.025 0.367 0.037l0.779 1.259c0.078 0.123 0.214 0.204 0.368 0.204s0.29-0.081 0.367-0.203l0.001-0.002 0.779-1.259c0.123-0.011 0.245-0.023 0.367-0.037l1.008 1.082c0.079 0.084 0.192 0.136 0.316 0.136 0.184 0 0.342-0.115 0.406-0.277l0.001-0.003 0.517-1.383q0.179-0.051 0.357-0.109l1.198 0.863c0.070 0.052 0.159 0.083 0.254 0.083 0.215 0 0.394-0.158 0.426-0.364l0-0.002 0.238-1.457c0.110-0.057 0.219-0.116 0.328-0.176l1.345 0.612c0.052 0.024 0.114 0.038 0.179 0.038 0.240 0 0.434-0.194 0.434-0.434 0-0.005-0-0.009-0-0.014l0 0.001-0.052-1.48q0.144-0.113 0.283-0.233l1.441 0.339c0.030 0.007 0.064 0.011 0.098 0.011 0.240 0 0.434-0.194 0.434-0.434 0-0.035-0.004-0.068-0.012-0.1l0.001 0.003-0.339-1.442c0.078-0.094 0.157-0.188 0.233-0.285l1.48 0.052c0.006 0 0.013 0 0.020 0 0.238 0 0.43-0.193 0.43-0.43 0-0.066-0.015-0.129-0.042-0.185l0.001 0.003-0.612-1.345c0.059-0.109 0.119-0.218 0.176-0.328l1.457-0.237c0.207-0.036 0.362-0.215 0.362-0.429 0-0.094-0.030-0.182-0.081-0.253l0.001 0.001-0.863-1.198 0.109-0.357 1.383-0.517c0.166-0.063 0.282-0.221 0.282-0.406 0-0.125-0.053-0.238-0.138-0.317l-0-0-1.082-1.010c0.013-0.121 0.025-0.244 0.035-0.367l1.259-0.779c0.125-0.077 0.208-0.213 0.208-0.368s-0.082-0.291-0.206-0.367l-0.002-0.001z"/></svg> },
                { name: "JavaScript", icon: <svg width="24" height="24" viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg"><rect x="2" y="2" width="28" height="28" fill="#FFCA28"/><path d="M19 25.2879L21.0615 23.9237C21.2231 24.4313 22.2462 25.6368 23.5385 25.6368C24.8308 25.6368 25.4308 24.931 25.4308 24.463C25.4308 23.1878 24.1112 22.7382 23.4774 22.5223C23.374 22.4871 23.289 22.4581 23.2308 22.4328C23.2009 22.4198 23.1558 22.4025 23.0979 22.3804C22.393 22.1111 19.7923 21.1175 19.7923 18.2373C19.7923 15.065 22.8538 14.7002 23.5462 14.7002C23.9991 14.7002 26.1769 14.7557 27.2615 16.7939L25.2615 18.1898C24.8231 17.3015 24.0946 17.0081 23.6462 17.0081C22.5385 17.0081 22.3077 17.8201 22.3077 18.1898C22.3077 19.227 23.5112 19.6919 24.5273 20.0844C24.7932 20.1871 25.0462 20.2848 25.2615 20.3866C26.3692 20.91 28 21.7666 28 24.463C28 25.8136 26.8672 28.0002 24.0154 28.0002C20.1846 28.0002 19.1692 25.7003 19 25.2879Z" fill="#3E3E3E"/><path d="M9 25.5587L11.1487 24.1953C11.317 24.7026 11.9713 25.638 12.9205 25.638C13.8698 25.638 14.3557 24.663 14.3557 24.1953V15.0002H16.9982V24.1953C17.041 25.4636 16.3376 28.0002 13.2332 28.0002C10.379 28.0002 9.19242 26.3039 9 25.5587Z" fill="#3E3E3E"/></svg> },
                { name: "WebAssembly", icon: <svg width="24" height="24" viewBox="0 0 512 512" xmlns="http://www.w3.org/2000/svg"><rect width="512" height="512" rx="15%" fill="#ffffff"/><path fill="#654ff0" d="m159.1 270.1h24l16.5 87.2 19.8-87.2h22.5l17.9 88.3 18.9-88.3h23.5l-30.6 128.2h-23.8L230 311l-19.1 87.3h-24.3zm170.2 0h37.8l37.5 128.2h-24.7l-8.2-28.6h-43.1l-6.3 28.6h-24.1zm14.4 31.6-10.5 47h32.6l-12.1-47zM297.4 75c0 .6 0 1.3 0 2c0 22.9-18.6 41.5-41.5 41.5c-22.9 0-41.5-18.6-41.5-41.5c0-.7 0-1.4 0-2H75V437H437V75z"/></svg> },
              ].map(({ name, icon }) => (
                <Card key={name}>
                  <div style={{ display: "flex", alignItems: "center", gap: "var(--space-3)" }}>
                    {icon}
                    <span
                      style={{
                        fontFamily: "var(--font-mono)",
                        fontSize: 14,
                        fontWeight: 500,
                        color: "var(--text)",
                      }}
                    >
                      {name}
                    </span>
                  </div>
                </Card>
              ))}
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
                        border: "1px solid",
                        borderColor: "rgba(128, 128, 128, 0.3)",
                        borderRadius: "var(--radius-sm)",
                        color: "var(--text)",
                        backgroundColor: "rgba(128, 128, 128, 0.1)",
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
