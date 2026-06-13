import React, { useEffect, useRef, useState } from 'react'
import { CustomCursor } from './HeroNetwork'
import TrustConstellation from './TrustConstellation'
import DemoPlayground from './DemoPlayground'
import Typewriter from 'typewriter-effect'
import './homepage.css'

interface StatProps {
  end: number
  suffix?: string
  label: string
  duration?: number
}

const AnimatedStat: React.FC<StatProps> = ({ end, suffix = '', label, duration = 2000 }) => {
  const [count, setCount] = useState(0)
  const ref = useRef<HTMLDivElement>(null)
  const started = useRef(false)

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && !started.current) {
          started.current = true
          const startTime = performance.now()
          const animate = (now: number) => {
            const elapsed = now - startTime
            const progress = Math.min(elapsed / duration, 1)
            const eased = 1 - Math.pow(1 - progress, 3)
            setCount(Math.floor(eased * end))
            if (progress < 1) requestAnimationFrame(animate)
          }
          requestAnimationFrame(animate)
        }
      },
      { threshold: 0.5 }
    )
    if (ref.current) observer.observe(ref.current)
    return () => observer.disconnect()
  }, [end, duration])

  return (
    <div ref={ref} className="ff-stat-item">
      <div className="ff-stat-num">{count.toLocaleString()}{suffix}</div>
      <div className="ff-stat-label">{label}</div>
    </div>
  )
}

const ReceiptStrip: React.FC = () => {
  const receipts = [
    { id: 'rx_8f3a92c1', fn: 'process_payment', status: 'ok', t: '1.84ms' },
    { id: 'rx_2b7d41e8', fn: 'validate_schema', status: 'ok', t: '0.92ms' },
    { id: 'rx_5c9f13a0', fn: 'fetch_weather_api', status: 'run', t: '—' },
    { id: 'rx_1a4e87b2', fn: 'send_email', status: 'ok', t: '3.21ms' },
    { id: 'rx_9d2c58f7', fn: 'run_code_sandbox', status: 'ok', t: '6.48ms' },
    { id: 'rx_3e6b90d4', fn: 'ocr_document', status: 'warn', t: '12.1ms' },
    { id: 'rx_7f1a23c9', fn: 'vector_search', status: 'ok', t: '2.17ms' },
    { id: 'rx_4b8e62f1', fn: 'call_external_api', status: 'run', t: '—' },
    { id: 'rx_0c5d74a3', fn: 'transform_data', status: 'ok', t: '0.44ms' },
    { id: 'rx_6a9f18b5', fn: 'generate_report', status: 'ok', t: '4.93ms' },
  ]
  const doubled = [...receipts, ...receipts, ...receipts, ...receipts]

  return (
    <div className="ff-receipt-strip">
      <div className="ff-receipt-track">
        {doubled.map((r, i) => (
          <div key={i} className={`ff-receipt-chip ff-chip-${r.status}`}>
            <span className="ff-chip-status" />
            <span className="ff-chip-id">{r.id}</span>
            <span>{r.fn}</span>
            <span className="ff-chip-time">{r.t}</span>
          </div>
        ))}
      </div>
    </div>
  )
}

const ScrollReveal: React.FC<React.PropsWithChildren<{ className?: string; delay?: number }>> = ({
  children,
  className = '',
  delay = 0,
}) => {
  const ref = useRef<HTMLDivElement>(null)
  const [visible, setVisible] = useState(false)

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) {
          setVisible(true)
        }
      },
      { threshold: 0.1, rootMargin: '0px 0px -60px 0px' }
    )
    if (ref.current) observer.observe(ref.current)
    return () => observer.disconnect()
  }, [])

  return (
    <div
      ref={ref}
      className={`${className} ${visible ? 'ff-visible' : 'ff-hidden-anim'}`}
      style={{ transitionDelay: `${delay}ms` }}
    >
      {children}
    </div>
  )
}

const HomePage: React.FC<{ authOrigin: string; docsUrl: string }> = ({ authOrigin, docsUrl }) => {
  const [navScrolled, setNavScrolled] = useState(false)
  const heroRef = useRef<HTMLElement>(null)
  const canvasRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const onScroll = () => {
      setNavScrolled(window.scrollY > 40)
      if (canvasRef.current) {
        const pct = Math.min(window.scrollY / window.innerHeight, 1)
        canvasRef.current.style.opacity = String(1 - pct * 1.4)
      }
    }
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])

  useEffect(() => {
    document.querySelectorAll('.ff-hero-eyebrow, .ff-hero-headline, .ff-hero-sub, .ff-hero-actions, .ff-hero-stats, .ff-scroll-indicator').forEach((el, i) => {
      ;(el as HTMLElement).style.transitionDelay = `${0.3 + i * 0.2}s`
    })
  }, [])

  return (
    <>
      <CustomCursor />

      <div ref={canvasRef} className="ff-hero-canvas-container">
        <TrustConstellation />
      </div>

      <div className="ff-homepage">
        <section ref={heroRef} className="ff-hero-section">
          <div className="ff-hero-inner">
            <div className="ff-hero-eyebrow">
              <span className="ff-pulse-dot" />
              <span>Infrastructure layer</span>
              <span className="ff-live-badge">LIVE</span>
            </div>
            <h1 className="ff-hero-headline">
              <Typewriter
                onInit={(typewriter) => {
                  typewriter
                    .typeString('Agents ')
                    .typeString('<span class="ff-flame-text">execute.</span>')
                    .typeString(' You ')
                    .typeString('<span class="ff-cyan-text">trust.</span>')
                    .start()
                }}
                options={{
                  wrapperClassName: 'ff-typewriter-wrapper',
                  cursorClassName: 'ff-typewriter-cursor',
                  html: true,
                }}
              />
            </h1>
            <p className="ff-hero-sub">
              The <strong>sandboxed execution and trust layer</strong> built for AI agents.
              MCP, A2A, verifiable receipts — every function call audited, every result provable.
            </p>
            <div className="ff-hero-actions">
              <a href={`${authOrigin}/signup`} className="ff-btn ff-btn-primary ff-btn-xl ff-glow">
                <span>Deploy your first agent</span>
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
                  <path d="M3 8L13 8M9 4L13 8L9 12" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </a>
              <a href={docsUrl} className="ff-btn ff-btn-secondary ff-btn-xl" target="_blank" rel="noopener noreferrer">
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
                  <path d="M7 1L7 15M1 8L15 8" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
                </svg>
                <span>View Documentation</span>
              </a>
            </div>
            <div className="ff-hero-stats-row">
              <AnimatedStat end={248000} suffix="+" label="executions / day" />
              <AnimatedStat end={2} suffix="ms" label="cold start" />
              <AnimatedStat end={100} suffix="%" label="receipt coverage" />
            </div>
          </div>
          <div className="ff-scroll-indicator">
            <div className="ff-scroll-line" />
            <span className="ff-scroll-text">Scroll</span>
          </div>
        </section>

        <ReceiptStrip />

        <section className="ff-features-section">
          <ScrollReveal>
            <p className="ff-section-label">What's under the hood</p>
          </ScrollReveal>
          <ScrollReveal delay={100}>
            <h2 className="ff-section-title">
              Precision infrastructure.<br />Zero-trust execution.
            </h2>
          </ScrollReveal>
          <div className="ff-features-grid">
            {[
              { icon: '🔒', title: 'Sandboxed Execution', desc: 'Every function runs in an isolated gVisor container. No shared memory, no escape vectors, no trust assumed.', tag: 'gVisor · runsc', color: 'flame' },
              { icon: '📜', title: 'Execution Receipts', desc: 'Cryptographically signed receipts for every call. Immutable, timestamped, verifiable by any party downstream.', tag: 'SHA-256 · Signed', color: 'cyan' },
              { icon: '⚡', title: 'MCP + A2A Native', desc: 'First-class Model Context Protocol and Agent-to-Agent support. Drop into any AI orchestration layer without rewiring.', tag: 'MCP · A2A', color: 'strat' },
              { icon: '🛒', title: 'Developer Marketplace', desc: 'Publish, discover, and monetize sandboxed functions. Every listing is a verified, execution-ready artifact.', tag: 'Marketplace', color: 'taxiway' },
              { icon: '🌐', title: 'Auth Proxy + Routing', desc: 'Centralized auth and intelligent routing across agent calls. The moat between your agents and the open internet.', tag: 'Auth · Routing', color: 'beacon' },
              { icon: '📡', title: 'Relay Message Bus', desc: 'Async-first message bus for agent-to-agent coordination. Fire and forget, or await — your architecture, your rules.', tag: 'Async · Relay', color: 'afterburner' },
            ].map((f, i) => (
              <ScrollReveal key={f.title} delay={i * 80}>
                <div className="ff-feature-card" data-accent={f.color}>
                  <div className="ff-feature-card-glow" />
                  <div className="ff-feature-icon">{f.icon}</div>
                  <h3>{f.title}</h3>
                  <p>{f.desc}</p>
                  <span className="ff-feature-tag">{f.tag}</span>
                </div>
              </ScrollReveal>
            ))}
          </div>
        </section>

        <section className="ff-languages-section">
          <ScrollReveal>
            <p className="ff-section-label">Language Support</p>
          </ScrollReveal>
          <ScrollReveal delay={100}>
            <h2 className="ff-section-title">
              Your language.<br />Your runtime.<br />Our sandbox.
            </h2>
          </ScrollReveal>
          <div className="ff-languages-grid">
            {[
              {
                name: 'Python',
                svg: <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 128 128" width="28" height="28"><linearGradient id="python-a" gradientUnits="userSpaceOnUse" x1="70.252" y1="1237.476" x2="170.659" y2="1151.089" gradientTransform="matrix(.563 0 0 -.568 -29.215 707.817)"><stop offset="0" stopColor="#5A9FD4"/><stop offset="1" stopColor="#306998"/></linearGradient><linearGradient id="python-b" gradientUnits="userSpaceOnUse" x1="209.474" y1="1098.811" x2="173.62" y2="1149.537" gradientTransform="matrix(.563 0 0 -.568 -29.215 707.817)"><stop offset="0" stopColor="#FFD43B"/><stop offset="1" stopColor="#FFE873"/></linearGradient><path fill="url(#python-a)" d="M63.391 1.988c-4.222.02-8.252.379-11.8 1.007-10.45 1.846-12.346 5.71-12.346 12.837v9.411h24.693v3.137H29.977c-7.176 0-13.46 4.313-15.426 12.521-2.268 9.405-2.368 15.275 0 25.096 1.755 7.311 5.947 12.519 13.124 12.519h8.491V67.234c0-8.151 7.051-15.34 15.426-15.34h24.665c6.866 0 12.346-5.654 12.346-12.548V15.833c0-6.693-5.646-11.72-12.346-12.837-4.244-.706-8.645-1.027-12.866-1.008zM50.037 9.557c2.55 0 4.634 2.117 4.634 4.721 0 2.593-2.083 4.69-4.634 4.69-2.56 0-4.633-2.097-4.633-4.69-.001-2.604 2.073-4.721 4.633-4.721z" transform="translate(0 10.26)"/><path fill="url(#python-b)" d="M91.682 28.38v10.966c0 8.5-7.208 15.655-15.426 15.655H51.591c-6.756 0-12.346 5.783-12.346 12.549v23.515c0 6.691 5.818 10.628 12.346 12.547 7.816 2.297 15.312 2.713 24.665 0 6.216-1.801 12.346-5.423 12.346-12.547v-9.412H63.938v-3.138h37.012c7.176 0 9.852-5.005 12.348-12.519 2.578-7.735 2.467-15.174 0-25.096-1.774-7.145-5.161-12.521-12.348-12.521h-9.268zM77.809 87.927c2.561 0 4.634 2.097 4.634 4.692 0 2.602-2.074 4.719-4.634 4.719-2.55 0-4.633-2.117-4.633-4.719 0-2.595 2.083-4.692 4.633-4.692z" transform="translate(0 10.26)"/><radialGradient id="python-c" cx="1825.678" cy="444.45" r="26.743" gradientTransform="matrix(0 -.24 -1.055 0 532.979 557.576)" gradientUnits="userSpaceOnUse"><stop offset="0" stopColor="#B8B8B8" stopOpacity=".498"/><stop offset="1" stopColor="#7F7F7F" stopOpacity="0"/></radialGradient><path opacity=".444" fill="url(#python-c)" d="M97.309 119.597c0 3.543-14.816 6.416-33.091 6.416-18.276 0-33.092-2.873-33.092-6.416 0-3.544 14.815-6.417 33.092-6.417 18.275 0 33.091 2.872 33.091 6.417z"/></svg>,
              },
              {
                name: 'Go',
                svg: <svg xmlns="http://www.w3.org/2000/svg" xmlnsXlink="http://www.w3.org/1999/xlink" viewBox="0 0 128 128" width="28" height="28"><path fill="#00ADD8" d="M62.8 4c13.6 0 26.3 1.9 33 15 6 14.6 3.8 30.4 4.8 45.9.8 13.3 2.5 28.6-3.6 40.9-6.5 12.9-22.7 16.2-36 15.7-10.5-.4-23.1-3.8-29.1-13.4-6.9-11.2-3.7-27.9-3.2-40.4.6-14.8-4-29.7.9-44.1C34.5 8.5 48.1 5.1 62.8 4"/><path fill="#fff" d="M65.2 22.2c2.4 14.2 25.6 10.4 22.3-3.9-3-12.8-23.1-9.2-22.3 3.9"/><path fill="#fff" d="M37.5 24.5c3.2 12.3 22.9 9.2 22.2-3.2-.9-14.8-25.3-12-22.2 3.2"/><path fill="#fff" d="M68 39.2c0 1.8.4 3.9.1 5.9-.5.9-1.4 1-2.2 1.3-1.1-.2-2-.9-2.5-1.9-.3-2.2.1-4.4.2-6.6l4.4 1.3z"/><path fill="#fff" d="M58.4 39c-1.5 3.5.8 10.6 4.8 5.4-.3-2.2.1-4.4.2-6.6l-5 1.2z"/><path fill="#F6D2A2" d="M58.9 32.2c-2.7.2-4.9 3.5-3.5 6 1.9 3.4 6-.3 8.6 0 3 .1 5.4 3.2 7.8.6 2.7-2.9-1.2-5.7-4.1-7l-8.8.4z"/><path fill="#fff" d="M58.6 32.1c-.2-4.7 8.8-5.3 9.8-1.4 1.1 4-9.4 4.9-9.8 1.4"/></svg>,
              },
              {
                name: 'Rust',
                svg: <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 128 128" width="28" height="28"><path d="M62.96.242c-.232.135-1.203 1.528-2.16 3.097-2.4 3.94-2.426 3.942-5.65.55-2.098-2.208-2.605-2.612-3.28-2.607-.44.002-.995.152-1.235.332-.24.18-.916 1.612-1.504 3.183-1.346 3.6-1.41 3.715-2.156 3.86-.46.086-1.343-.407-3.463-1.929-1.565-1.125-3.1-2.045-3.411-2.045-1.291 0-1.655.706-2.27 4.4-.78 4.697-.754 4.681-4.988 2.758-1.71-.776-3.33-1.41-3.603-1.41-.274 0-.792.293-1.15.652-.652.652-.653.655-.475 4.246l.178 3.595-.68.364c-.602.322-1.017.283-3.684-.348-3.48-.822-4.216-.8-4.92.15l-.516.693.692 2.964c.38 1.63.745 3.2.814 3.487.067.287-.05.746-.26 1.02-.348.448-.717.49-3.94.44-5.452-.086-5.761.382-3.51 5.3.718 1.56 1.305 2.98 1.305 3.15 0 .898-.717 1.224-3.794 1.727-1.722.28-3.218.51-3.326.51-.107 0-.43.235-.717.522-.937.936-.671 1.816 1.453 4.814 2.646 3.735 2.642 3.75-1.73 5.421-4.971 1.902-5.072 2.37-1.287 5.96 3.525 3.344 3.53 3.295-.461 5.804C.208 62.8.162 62.846.085 63.876c-.093 1.253-.071 1.275 3.538 3.48 3.57 2.18 3.57 2.246.067 5.56C-.078 76.48.038 77 5.013 78.877c4.347 1.64 4.353 1.66 1.702 5.394-1.502 2.117-1.981 3-1.981 3.653 0 1.223.637 1.535 4.44 2.174 3.206.54 3.92.857 3.92 1.741 0 .182-.588 1.612-1.307 3.177-2.236 4.87-1.981 5.275 3.31 5.275 4.93 0 4.799-.15 3.737 4.294-.8 3.35-.813 3.992-.088 4.715.554.556 1.6.494 4.87-.289 2.499-.596 2.937-.637 3.516-.328l.66.354-.177 3.594c-.178 3.593-.177 3.595.475 4.248.358.36.884.652 1.165.652.282 0 1.903-.63 3.604-1.404 4.22-1.916 4.194-1.932 4.973 2.75.617 3.711.977 4.4 2.294 4.4.327 0 1.83-.88 3.34-1.958 2.654-1.893 3.342-2.19 4.049-1.74.182.115.89 1.67 1.572 3.455 1.003 2.625 1.37 3.31 1.929 3.576 1.062.51 1.72.1 4.218-2.62 3.016-3.286 3.14-3.27 5.602.72 2.72 4.406 3.424 4.396 6.212-.089 2.402-3.864 2.374-3.862 5.621-.47 2.157 2.25 2.616 2.61 3.343 2.61.464 0 1.019-.175 1.23-.388.214-.213.92-1.786 1.568-3.496.649-1.71 1.321-3.2 1.495-3.31.687-.436 1.398-.13 4.048 1.752 1.56 1.108 3.028 1.96 3.377 1.96 1.296 0 1.764-.92 2.302-4.535.46-3.082.554-3.378 1.16-3.685.596-.302.954-.2 3.75 1.07 1.701.77 3.323 1.402 3.604 1.402.282 0 .816-.302 1.184-.672l.672-.67-.184-3.448c-.177-3.29-.16-3.468.364-3.943.54-.488.596-.486 3.615.204 3.656.835 4.338.857 5.025.17.671-.67.664-.818-.254-4.69-1.03-4.346-1.168-4.19 3.78-4.19 3.374 0 3.75-.049 4.18-.523.718-.793.547-1.702-.896-4.779-.729-1.55-1.32-2.96-1.315-3.135.024-.914.743-1.227 4.065-1.767 2.033-.329 3.553-.71 3.829-.96.923-.833.584-1.918-1.523-4.873-2.642-3.703-2.63-3.738 1.599-5.297 5.064-1.866 5.209-2.488 1.419-6.09-3.51-3.335-3.512-3.317.333-5.677 4.648-2.853 4.655-3.496.082-6.335-3.933-2.44-3.93-2.406-.405-5.753 3.78-3.593 3.678-4.063-1.295-5.965-4.388-1.679-4.402-1.72-1.735-5.38 1.588-2.18 1.982-2.903 1.982-3.65 0-1.306-.586-1.598-4.436-2.22-3.216-.52-3.924-.835-3.924-1.75 0-.174.588-1.574 1.307-3.113 1.406-3.013 1.604-4.22.808-4.94-.428-.387-1-.443-4.067-.392-3.208.054-3.618.008-4.063-.439-.486-.488-.48-.557.278-3.725.931-3.88.935-3.975.17-4.694-.777-.73-1.262-.718-4.826.121-2.597.612-3.027.653-3.617.337l-.67-.36.185-3.582.186-3.58-.67-.67c-.369-.37-.891-.67-1.163-.67-.27 0-1.884.64-3.583 1.421-2.838 1.306-3.143 1.393-3.757 1.072-.612-.32-.714-.637-1.237-3.829-.603-3.693-.977-4.412-2.288-4.412-.311 0-1.853.925-3.426 2.055-2.584 1.856-2.93 2.032-3.574 1.807-.533-.186-.843-.59-1.221-1.599-.28-.742-.817-2.172-1.194-3.177-.762-2.028-1.187-2.482-2.328-2.482-.637 0-1.213.458-3.28 2.604-3.25 3.375-3.261 3.374-5.65-.545C66.073 1.78 65.075.382 64.81.24c-.597-.32-1.3-.32-1.85.002m2.96 11.798c2.83 2.014 1.326 6.75-2.144 6.75-3.368 0-5.064-4.057-2.66-6.36 1.358-1.3 3.304-1.459 4.805-.39m-3.558 12.507c1.855.705 2.616.282 6.852-3.8l3.182-3.07 1.347.18c4.225.56 12.627 4.25 17.455 7.666 4.436 3.14 10.332 9.534 12.845 13.93l.537.942-2.38 5.364c-1.31 2.95-2.382 5.673-2.382 6.053 0 .878.576 2.267 1.13 2.726.234.195 2.457 1.265 4.939 2.378l4.51 2.025.178 1.148c.23 1.495.26 5.167.052 6.21l-.163.816h-2.575c-2.987 0-2.756-.267-2.918 3.396-.118 2.656-.76 4.124-2.22 5.075-2.377 1.551-6.304 1.27-7.97-.57-.255-.284-.752-1.705-1.105-3.16-1.03-4.254-2.413-6.64-5.193-8.965-.878-.733-1.595-1.418-1.595-1.522 0-.102.965-.915 2.145-1.803 4.298-3.24 6.77-7.012 7.04-10.747.519-7.126-5.158-13.767-13.602-15.92-2.002-.51-2.857-.526-27.624-.526-14.057 0-25.56-.092-25.56-.204 0-.263 3.125-3.295 4.965-4.816 5.054-4.178 11.618-7.465 18.417-9.22l2.35-.61 3.34 3.387c1.839 1.863 3.64 3.5 4.003 3.637M20.3 46.34c1.539 1.008 2.17 3.54 1.26 5.062-1.405 2.356-4.966 2.455-6.373.178-2.046-3.309 1.895-7.349 5.113-5.24m90.672.13c4.026 2.454.906 8.493-3.404 6.586-2.877-1.273-2.97-5.206-.155-6.64 1.174-.6 2.523-.579 3.56.053M32.163 61.5v15.02h-13.28l-.526-2.285c-1.036-4.5-1.472-9.156-1.211-12.969l.182-2.679 4.565-2.047c2.864-1.283 4.706-2.262 4.943-2.625 1.038-1.584.94-2.715-.518-5.933l-.68-1.502h6.523V61.5M70.39 47.132c2.843.74 4.345 2.245 4.349 4.355.002 1.55-.765 2.52-2.67 3.38-1.348.61-1.562.625-10.063.708l-8.686.084v-8.92h7.782c6.078 0 8.112.086 9.288.393m-2.934 21.554c1.41.392 3.076 1.616 3.93 2.888.898 1.337 1.423 3.076 2.667 8.836 1.05 4.87 1.727 6.46 3.62 8.532 2.345 2.566 1.8 2.466 13.514 2.466 5.61 0 10.198.09 10.198.2 0 .197-3.863 4.764-4.03 4.764-.048 0-2.066-.422-4.484-.939-6.829-1.458-7.075-1.287-8.642 6.032l-1.008 4.702-.91.448c-1.518.75-6.453 2.292-9.01 2.82-4.228.87-8.828 1.162-12.871.821-6.893-.585-16.02-3.259-16.377-4.8-.075-.327-.535-2.443-1.018-4.704-.485-2.26-1.074-4.404-1.31-4.764-1.13-1.724-2.318-1.83-7.547-.674-1.98.44-3.708.796-3.84.796-.248 0-3.923-4.249-3.923-4.535 0-.09 8.728-.194 19.396-.23l19.395-.066.07-6.89c.05-4.865-.018-6.997-.23-7.25-.234-.284-1.485-.358-6.011-.358H53.32v-8.36l6.597.001c3.626.002 7.02.12 7.539.264M37.57 100.02c3.084 1.88 1.605 6.804-2.043 6.8-3.74 0-5.127-4.88-1.94-6.826 1.055-.643 2.908-.63 3.983.026m56.48.206c1.512 1.108 2.015 3.413 1.079 4.95-2.46 4.034-8.612.827-6.557-3.419 1.01-2.085 3.695-2.837 5.478-1.53"/></svg>,
              },
              {
                name: 'JS',
                svg: <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 128 128" width="28" height="28"><path fill="#F0DB4F" d="M1.408 1.408h125.184v125.185H1.408z"/><path fill="#323330" d="M116.347 96.736c-.917-5.711-4.641-10.508-15.672-14.981-3.832-1.761-8.104-3.022-9.377-5.926-.452-1.69-.512-2.642-.226-3.665.821-3.32 4.784-4.355 7.925-3.403 2.023.678 3.938 2.237 5.093 4.724 5.402-3.498 5.391-3.475 9.163-5.879-1.381-2.141-2.118-3.129-3.022-4.045-3.249-3.629-7.676-5.498-14.756-5.355l-3.688.477c-3.534.893-6.902 2.748-8.877 5.235-5.926 6.724-4.236 18.492 2.975 23.335 7.104 5.332 17.54 6.545 18.873 11.531 1.297 6.104-4.486 8.08-10.234 7.378-4.236-.881-6.592-3.034-9.139-6.949-4.688 2.713-4.688 2.713-9.508 5.485 1.143 2.499 2.344 3.63 4.26 5.795 9.068 9.198 31.76 8.746 35.83-5.176.165-.478 1.261-3.666.38-8.581zM69.462 58.943H57.753l-.048 30.272c0 6.438.333 12.34-.714 14.149-1.713 3.558-6.152 3.117-8.175 2.427-2.059-1.012-3.106-2.451-4.319-4.485-.333-.584-.583-1.036-.667-1.071l-9.52 5.83c1.583 3.249 3.915 6.069 6.902 7.901 4.462 2.678 10.459 3.499 16.731 2.059 4.082-1.189 7.604-3.652 9.448-7.401 2.666-4.915 2.094-10.864 2.07-17.444.06-10.735.001-21.468.001-32.237z"/></svg>,
              },
              {
                name: 'WASM',
                svg: <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="28" height="28"><title>WebAssembly</title><path fill="#654FF0" d="M14.745,0c0,0.042,0,0.085,0,0.129c0,1.52-1.232,2.752-2.752,2.752c-1.52,0-2.752-1.232-2.752-2.752 c0-0.045,0-0.087,0-0.129H0v24h24V0H14.745z M11.454,21.431l-1.169-5.783h-0.02l-1.264,5.783H7.39l-1.824-8.497h1.59l1.088,5.783 h0.02l1.311-5.783h1.487l1.177,5.854h0.02l1.242-5.854h1.561l-2.027,8.497H11.454z M20.209,21.431l-0.542-1.891h-2.861l-0.417,1.891 h-1.59l2.056-8.497h2.509l2.5,8.497H20.209z M17.812,15.028l-0.694,3.118h2.159l-0.796-3.118H17.812z"/></svg>,
              },
            ].map((lang, i) => (
              <ScrollReveal key={lang.name} delay={i * 80}>
                <div className="ff-language-chip">
                  <span className="ff-language-icon">{lang.svg}</span>
                  <span className="ff-language-name">{lang.name}</span>
                </div>
              </ScrollReveal>
            ))}
          </div>
          <ScrollReveal delay={200}>
            <div className="ff-languages-features">
              <div className="ff-lang-feature">
                <span className="ff-lang-feature-icon">🔒</span>
                <div>
                  <strong>Isolated</strong>
                  <p>Wasm + WASI per-request sandbox. Every invocation gets its own clean execution environment.</p>
                </div>
              </div>
              <div className="ff-lang-feature">
                <span className="ff-lang-feature-icon">✅</span>
                <div>
                  <strong>Trust-scored</strong>
                  <p>Execution and routing with cryptographic verification. Zero-trust by default.</p>
                </div>
              </div>
            </div>
          </ScrollReveal>
        </section>

        <section className="ff-audience-section">
          <ScrollReveal>
            <p className="ff-section-label">Who benefits most</p>
          </ScrollReveal>
          <ScrollReveal delay={100}>
            <h2 className="ff-section-title">
              Built for every stage.<br />Trusted by all.
            </h2>
          </ScrollReveal>
          <div className="ff-audience-grid">
            {[
              {
                icon: '🎯',
                badge: 'Indie',
                title: 'Indie SaaS Founder',
                subtitle: 'Deploying your first app',
                desc: 'You built an amazing product and now need trustworthy agent tooling without enterprise overhead.',
                features: ['Python, Go, Rust, JS, WASM support', 'Wasm + WASI sandbox isolation', 'Trust scores and verification tiers', 'Pay as you grow'],
                tag: '$0 to start',
                color: 'indie'
              },
              {
                icon: '🚀',
                badge: 'Growth',
                title: 'Growing Startup',
                subtitle: 'Scaling agent + API workloads',
                desc: 'Your team is shipping fast and agents are calling more tools every week.',
                features: ['Multi-level verification workflows', 'Agent swarm orchestration', 'Policy engine for safe execution', 'Multi-provider routing'],
                tag: 'Usage-based',
                color: 'startup'
              },
              {
                icon: '🏢',
                badge: 'Enterprise',
                title: 'Enterprise',
                subtitle: 'Governance and compliance',
                desc: 'Your organization needs auditability, least-privilege access, and consistent trust policies.',
                features: ['Firecracker MicroVMs', 'Multi-level attestations', 'Custom SLAs and audit logs', 'Multi-team patterns'],
                tag: 'Contact sales',
                color: 'enterprise'
              },
            ].map((a, i) => (
              <ScrollReveal key={a.title} delay={i * 100}>
                <div className={`ff-audience-card ff-audience-card--${a.color}`}>
                  <div className="ff-audience-card-glow" />
                  <div className="ff-audience-card-header">
                    <span className="ff-audience-icon">{a.icon}</span>
                    <span className={`ff-audience-badge ff-audience-badge--${a.color}`}>{a.badge}</span>
                  </div>
                  <h3 className="ff-audience-title">{a.title}</h3>
                  <p className="ff-audience-subtitle">{a.subtitle}</p>
                  <p className="ff-audience-desc">{a.desc}</p>
                  <ul className="ff-audience-features">
                    {a.features.map(f => (
                      <li key={f}>
                        <span className="ff-check-icon">✓</span>
                        <span>{f}</span>
                      </li>
                    ))}
                  </ul>
                  <div className="ff-audience-footer">
                    <span className={`ff-audience-tag ff-audience-tag--${a.color}`}>{a.tag}</span>
                  </div>
                </div>
              </ScrollReveal>
            ))}
          </div>
        </section>

        <section className="ff-how-it-works">
          <ScrollReveal>
            <p className="ff-section-label">How it works</p>
          </ScrollReveal>
          <ScrollReveal delay={100}>
            <h2 className="ff-section-title">
              From code to trusted<br />agent tool in 4 steps.
            </h2>
          </ScrollReveal>
          <div className="ff-steps-grid">
            {[
              { num: '01', icon: '✏️', title: 'Write', desc: 'Implement your function in Python, Go, Rust, JavaScript, or TypeScript. Your code runs in isolation — it cannot phone home or access the filesystem unless you grant capabilities explicitly.' },
              { num: '02', icon: '📤', title: 'Publish', desc: 'Run ff check to verify your setup, then ff deploy. FunctionFly bundles your code, compiles it for the target runtime, signs the artifact, and runs multi-level verification checks.' },
              { num: '03', icon: '🔍', title: 'Discover', desc: 'Agents find your function through manifests and trust filters. Your function\'s trust score, verification tier, execution history, and capability declarations are all surfaced automatically.' },
              { num: '04', icon: '⚡', title: 'Execute', desc: 'Agents call your function over HTTPS. FunctionFly routes to the right backend, enforces policy limits, runs in a Wasm sandbox, and records every result for trust scoring.' },
            ].map((step, i) => (
              <ScrollReveal key={step.num} delay={i * 120}>
                <div className="ff-step-card">
                  <div className="ff-step-num">{step.num}</div>
                  <div className="ff-step-icon">{step.icon}</div>
                  <h3>{step.title}</h3>
                  <p>{step.desc}</p>
                </div>
              </ScrollReveal>
            ))}
          </div>
        </section>

        <section className="ff-receipt-demo">
          <div className="ff-receipt-demo-text">
            <ScrollReveal>
              <p className="ff-section-label">Execution Receipts</p>
            </ScrollReveal>
            <ScrollReveal delay={100}>
              <h2 className="ff-section-title">
                Every call.<br />Signed. Sealed.<br />Provable.
              </h2>
            </ScrollReveal>
            <ScrollReveal delay={200}>
              <p>
                When an AI agent executes a function through FunctionFly, it gets back more than a result.
                It gets a <strong>cryptographic receipt</strong> — a tamper-proof artifact that any downstream system can verify.
              </p>
            </ScrollReveal>
            <ScrollReveal delay={300}>
              <p>
                No more "did that really run?" No more trust gaps between agents.
                <strong> The receipt is the proof.</strong>
              </p>
            </ScrollReveal>
          </div>
          <ScrollReveal delay={200}>
            <div className="ff-receipt-card">
              <div className="ff-receipt-card-header">
                <div className="ff-receipt-dots">
                  <span style={{ background: '#FF2D55' }} />
                  <span style={{ background: '#FFB800' }} />
                  <span style={{ background: '#00FF9D' }} />
                </div>
                <span>execution.receipt.v1</span>
              </div>
              <div className="ff-receipt-card-body">
                {[
                  { k: 'receipt_id', v: 'rcpt_7f2a1b3c9d4e5f8a1b2c3d4e5f6a7b8c', cls: 'ff-id' },
                  { k: 'function', v: 'process_payment' },
                  { k: 'agent_id', v: 'agt_claude_prod_7' },
                  { k: 'sandbox', v: 'gvisor/runsc-secure', cls: 'ff-cyan-val' },
                  { k: 'executed_at', v: '2026-06-12T14:22:01.384Z' },
                  { k: 'duration_ms', v: '1.84' },
                  { k: 'exit_code', v: '0 · SUCCESS', cls: 'ff-ok' },
                  { k: 'memory_used', v: '12.3 MB' },
                  { k: 'network_egress', v: 'blocked', cls: 'ff-ok' },
                ].map(row => (
                  <div key={row.k} className="ff-receipt-row">
                    <span className="ff-receipt-key">{row.k}</span>
                    <span className={`ff-receipt-value ${row.cls || ''}`}>{row.v}</span>
                  </div>
                ))}
                <div className="ff-receipt-hash">
                  <span className="ff-hash-label">SHA-256 signature</span>
                  e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
                </div>
              </div>
            </div>
          </ScrollReveal>
        </section>

        <section className="ff-protocols-section">
          <ScrollReveal>
            <p className="ff-section-label">Protocol Support</p>
          </ScrollReveal>
          <ScrollReveal delay={100}>
            <h2 className="ff-section-title">
              Built for the<br />agentic stack.
            </h2>
          </ScrollReveal>
          <div className="ff-protocols-grid">
            {[
              { name: 'MCP', desc: 'Model Context Protocol. Native tool registration, schema validation, and sandboxed handler execution.' },
              { name: 'A2A', desc: 'Agent-to-Agent protocol. Structured inter-agent messaging with receipt-backed delivery guarantees.' },
              { name: 'REST', desc: 'Standard HTTP execution endpoints. Works with any agent framework, any language, any runtime.' },
              { name: 'gRPC', desc: 'High-throughput streaming execution for latency-sensitive agentic workloads at scale.' },
            ].map((p, i) => (
              <ScrollReveal key={p.name} delay={i * 100}>
                <div className="ff-protocol-card">
                  <div className="ff-protocol-glow" />
                  <div className="ff-protocol-name">{p.name}</div>
                  <p>{p.desc}</p>
                </div>
              </ScrollReveal>
            ))}
          </div>
        </section>

        <DemoPlayground authOrigin={authOrigin} />

        <section className="ff-cta-section">
          <div className="ff-cta-bg-glow" />
          <ScrollReveal>
            <p className="ff-cta-eyebrow">Ready for takeoff</p>
          </ScrollReveal>
          <ScrollReveal delay={100}>
            <h2 className="ff-cta-title">
              Your agents deserve<br />
              <span className="ff-gradient-text">better infrastructure.</span>
            </h2>
          </ScrollReveal>
          <ScrollReveal delay={200}>
            <p className="ff-cta-sub">
              Deploy your first sandboxed function in under 5 minutes. No credit card. Full receipt coverage from day one.
            </p>
          </ScrollReveal>
          <ScrollReveal delay={300}>
            <div className="ff-cta-actions">
              <a href={`${authOrigin}/signup`} className="ff-btn ff-btn-primary ff-btn-xl ff-glow">
                <span>Start building free</span>
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
                  <path d="M3 8L13 8M9 4L13 8L9 12" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </a>
              <a href={docsUrl} className="ff-btn ff-btn-secondary ff-btn-xl" target="_blank" rel="noopener noreferrer">
                <span>Read the docs</span>
                <span className="ff-arrow">→</span>
              </a>
            </div>
          </ScrollReveal>
        </section>
      </div>
    </>
  )
}

export default HomePage
