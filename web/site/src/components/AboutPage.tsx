import React from 'react'
import { AUTH_ORIGIN } from '../config'
import './homepage.css'

const AboutPage: React.FC = () => {
  return (
    <div className="ff-homepage">
      <section className="ff-hero-section">
        <div className="ff-hero-inner">
          <div className="ff-hero-eyebrow">
            <span className="ff-pulse-dot" />
            <span>Founded 2026 — Wyoming / Fort Worth, TX</span>
          </div>
          <h1 className="ff-hero-headline">
            Agent tools<br />should be<br />trustworthy.
          </h1>
          <p className="ff-hero-sub">
            We built FunctionFly so AI agents can call any function and know — not hope — that it's safe, signed, and behaving as advertised. Trust is technical, not metaphorical.
          </p>
          <div className="ff-hero-actions">
            <a className="ff-btn ff-btn-primary" href={`${AUTH_ORIGIN}/signup`}>
              Start Building
            </a>
            <a className="ff-btn ff-btn-secondary" href="/trust">
              How Trust Works
            </a>
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <div className="ff-about-content">
            <div className="ff-about-lead">
              <h2>Every tool an agent calls is a leap of faith</h2>
              <p>
                Today, AI agents call functions — email senders, payment processors, database queries — with no way to verify those functions do what they claim. There's no trust infrastructure: no verification, no attestations, no execution history, no trust scores.
              </p>
              <p>
                Existing platforms treat tools as dumb endpoints. Define the API, the agent calls it, whatever comes back is accepted. But in a world where agents act autonomously on your behalf, that gap is not just a technical problem — it's a security and reliability crisis waiting to happen.
              </p>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--alt">
        <div className="ff-container">
          <div className="ff-about-header">
            <div className="ff-section-tag">
              <span className="ff-tag-rule" />
              <span>The Solution</span>
            </div>
            <h2 className="ff-section-title">
              FunctionFly is the<br />
              <span className="ff-title-accent">trust layer</span>
            </h2>
            <p className="ff-section-desc">
              A complete trust infrastructure for AI agent tool usage — from publication to execution to trust score.
            </p>
          </div>

          <div className="ff-about-steps">
            <div className="ff-about-step">
              <div className="ff-about-step-num">01</div>
              <div className="ff-about-step-body">
                <div className="ff-about-step-icon">
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round"><path d="M9 12l2 2 4-4"/><path d="M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z"/></svg>
                </div>
                <h3>Publish & Verify</h3>
                <p>
                  Publish a function with a manifest describing its behavior, capabilities, and intended use. Submit to automated verification — static analysis, fuzzing, signature checks — and human attestation for higher trust tiers.
                </p>
              </div>
            </div>

            <div className="ff-about-step">
              <div className="ff-about-step-num">02</div>
              <div className="ff-about-step-body">
                <div className="ff-about-step-icon ff-about-step-icon--green">
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round"><path d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/></svg>
                </div>
                <h3>Sandbox Execution</h3>
                <p>
                  Every function call runs in an isolated WebAssembly sandbox with declared capabilities enforced at runtime. No filesystem access, no network egress, no privilege escalation — unless explicitly granted in the manifest.
                </p>
              </div>
            </div>

            <div className="ff-about-step">
              <div className="ff-about-step-num">03</div>
              <div className="ff-about-step-body">
                <div className="ff-about-step-icon ff-about-step-icon--amber">
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round"><path d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/></svg>
                </div>
                <h3>Trust Scores</h3>
                <p>
                  Every execution feeds into a live trust score. Functions that run successfully thousands of times earn higher scores. Functions that fail, time out, or behave unexpectedly are penalized. Agents can set policy thresholds and refuse to call functions below a given tier.
                </p>
              </div>
            </div>

            <div className="ff-about-step">
              <div className="ff-about-step-num">04</div>
              <div className="ff-about-step-body">
                <div className="ff-about-step-icon ff-about-step-icon--purple">
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg>
                </div>
                <h3>Zero-Knowledge Vault</h3>
                <p>
                  Store API keys, secrets, and sensitive config in the FunctionFly Vault. Encryption and decryption happen client-side — the server never sees your plaintext or your passphrase. Your secrets never leave your control.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <div className="ff-about-header">
            <div className="ff-section-tag">
              <span className="ff-tag-rule" />
              <span>How We Build</span>
            </div>
            <h2 className="ff-section-title">The principles behind every decision</h2>
          </div>

          <div className="ff-principles">
            <div className="ff-principle">
              <div className="ff-principle-num">01</div>
              <div className="ff-principle-body">
                <h3>Deny-by-default capabilities</h3>
                <p>
                  Every capability — network access, filesystem read/write, environment variable exposure — must be declared at publish time and is enforced at runtime by the sandbox. A function that hasn't declared network access cannot make outbound requests. No ambient authority, ever.
                </p>
              </div>
            </div>

            <div className="ff-principle">
              <div className="ff-principle-num">02</div>
              <div className="ff-principle-body">
                <h3>Trust is earned, not claimed</h3>
                <p>
                  Trust scores are computed from real execution outcomes. A function that has run 50,000 times with a 99.97% success rate has a measurably different trust profile than one that's never been called. Agents can query trust scores before making calls, and operators can set minimum thresholds per policy.
                </p>
              </div>
            </div>

            <div className="ff-principle">
              <div className="ff-principle-num">03</div>
              <div className="ff-principle-body">
                <h3>Multi-level verification</h3>
                <p>
                  Functions can submit to automated checks — static analysis, fuzzing, signature validation, sandbox behavior tests — and human attestation for higher trust tiers. Higher tiers unlock broader agent policies and higher call volume limits. Verification is ongoing, not one-time.
                </p>
              </div>
            </div>

            <div className="ff-principle">
              <div className="ff-principle-num">04</div>
              <div className="ff-principle-body">
                <h3>Observability is not optional</h3>
                <p>
                  Every function call is recorded: input, output, latency, cost, trust score delta, and failure mode. Execution traces, audit logs, and metrics dashboards give operators complete visibility. If something goes wrong, you know exactly what happened and why — not just that it failed.
                </p>
              </div>
            </div>

            <div className="ff-principle">
              <div className="ff-principle-num">05</div>
              <div className="ff-principle-body">
                <h3>Multi-language, no rewrites</h3>
                <p>
                  Write functions in Python, Go, Rust, JavaScript, TypeScript, or WebAssembly. One SDK, every runtime. Agents discover and call functions through a single registry interface — language is an implementation detail, not an integration burden. Switch languages without rewriting your tool integrations.
                </p>
              </div>
            </div>

            <div className="ff-principle">
              <div className="ff-principle-num">06</div>
              <div className="ff-principle-body">
                <h3>Zero-knowledge by design</h3>
                <p>
                  Secrets in the FunctionFly Vault are encrypted client-side with a key derived from your passphrase. The server stores only ciphertext, IV, salt, and auth tag — it cannot decrypt your secrets and cannot perform key recovery. If you lose your passphrase, your secrets are unrecoverable. That's the point.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--alt">
        <div className="ff-container">
          <div className="ff-about-header">
            <div className="ff-section-tag">
              <span className="ff-tag-rule" />
              <span>Our Values</span>
            </div>
            <h2 className="ff-section-title">What we stand for</h2>
            <p className="ff-section-desc">
              Six beliefs that shape every product decision, every architectural choice, and every line of code we ship.
            </p>
          </div>

          <div className="ff-values-grid">
            <div className="ff-value-card">
              <div className="ff-value-icon">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round"><path d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"/></svg>
              </div>
              <h3>Trust is technical</h3>
              <p>We build measurable signals — verification levels, execution history, signed attestations, trust scores — that agents and operators can audit programmatically.</p>
            </div>

            <div className="ff-value-card">
              <div className="ff-value-icon ff-value-icon--red">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round"><path d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/></svg>
              </div>
              <h3>Safe by default</h3>
              <p>Every function runs with the minimum capabilities it needs — nothing more. No ambient authority, no implicit permissions. The sandbox enforces this.</p>
            </div>

            <div className="ff-value-card">
              <div className="ff-value-icon ff-value-icon--green">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round"><path d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
              </div>
              <h3>Agent-native from day one</h3>
              <p>FunctionFly was designed for agents to discover, evaluate, and call functions — with manifests, trust filters, capability declarations, and policy routing built in from the start.</p>
            </div>

            <div className="ff-value-card">
              <div className="ff-value-icon ff-value-icon--amber">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round"><path d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z"/></svg>
              </div>
              <h3>Measurable, not magical</h3>
              <p>Trust tiers are earned through verifiable evidence — not self-attestation or marketing language. Automated tests, execution history, and human reviews are the inputs.</p>
            </div>

            <div className="ff-value-card">
              <div className="ff-value-icon ff-value-icon--purple">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round"><path d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"/></svg>
              </div>
              <h3>Multi-language, no rewrites</h3>
              <p>Write in Python, Go, Rust, JavaScript, TypeScript, or WASM. One SDK, every runtime. Agents discover functions through a single registry regardless of the language they're written in.</p>
            </div>

            <div className="ff-value-card">
              <div className="ff-value-icon ff-value-icon--cyan">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round"><path d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/></svg>
              </div>
              <h3>Full observability</h3>
              <p>Execution traces, audit logs, latency breakdowns, cost attribution, and trust score history. Every function call produces a complete record.</p>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <div className="ff-about-header">
            <div className="ff-section-tag">
              <span className="ff-tag-rule" />
              <span>Milestones</span>
            </div>
            <h2 className="ff-section-title">How we got here</h2>
          </div>

          <div className="ff-timeline">
            <div className="ff-timeline-item">
              <div className="ff-timeline-marker">
                <div className="ff-timeline-dot" />
                <span className="ff-timeline-year">Q1 2026</span>
              </div>
              <div className="ff-timeline-content">
                <h4>Founded</h4>
                <p>FunctionFly incorporated in Wyoming, headquartered in Fort Worth, TX. Began building the trust layer for AI agents.</p>
              </div>
            </div>

            <div className="ff-timeline-item">
              <div className="ff-timeline-marker">
                <div className="ff-timeline-dot" />
                <span className="ff-timeline-year">Q2 2026</span>
              </div>
              <div className="ff-timeline-content">
                <h4>Beta Launch</h4>
                <p>Public beta launch with multi-language SDKs for Python, Go, Rust, JavaScript, and TypeScript. Released the function registry API and Zero-knowledge Vault.</p>
              </div>
            </div>

            <div className="ff-timeline-item">
              <div className="ff-timeline-marker">
                <div className="ff-timeline-dot ff-timeline-dot--live" />
                <span className="ff-timeline-year">Q2 2026</span>
              </div>
              <div className="ff-timeline-content">
                <h4>Today</h4>
                <p>FunctionFly powers trust for agents across production deployments. 5 language runtimes, 4 trust tiers, 12 global regions, billions of function calls processed.</p>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--cta">
        <div className="ff-container">
          <div className="ff-cta">
            <div className="ff-section-tag">
              <span className="ff-tag-rule" />
              <span>Start Today</span>
            </div>
            <h2 className="ff-cta-title">
              Ready to build with<br />
              <span className="ff-cta-accent">trust infrastructure</span>?
            </h2>
            <p className="ff-cta-desc">
              Whether you're building agent tooling, evaluating trust infrastructure, or just want to understand how FunctionFly works — get started in minutes.
            </p>
            <div className="ff-cta-actions">
              <a className="ff-btn ff-btn-primary" href={`${AUTH_ORIGIN}/signup`}>
                Create free account
              </a>
              <a className="ff-btn ff-btn-outline" href="/pricing">
                View pricing
              </a>
            </div>
            <p className="ff-cta-footnote">No credit card required. Free tier includes 25K requests/month and 3 functions.</p>
          </div>
        </div>
      </section>
    </div>
  )
}

export default AboutPage