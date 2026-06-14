import React from 'react'
import { APP_DASHBOARD_ORIGIN, AUTH_ORIGIN } from '../config'
import './homepage.css'

const TrustPage: React.FC = () => {
  return (
    <div className="ff-homepage">
      <section className="ff-hero-section ff-trust-hero">
        <div className="ff-hero-inner">
          <div className="ff-hero-eyebrow">
            <span className="ff-pulse-dot" />
            <span>Trust, explained</span>
          </div>
          <h1 className="ff-hero-headline">
            The Trust Layer<br />for AI Agents
          </h1>
          <p className="ff-hero-sub">
            Agents do not need "more tools." They need tools they can trust. FunctionFly turns functions into
            verified, signed, auditable building blocks with execution-backed trust scores.
          </p>
          <div className="ff-hero-actions">
            <a className="ff-btn ff-btn-primary" href="#verification-levels">Verification levels</a>
            <a className="ff-btn ff-btn-secondary" href="#trust-api">Trust API</a>
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <div className="ff-trust-grid-2">
            <div className="ff-trust-card">
              <h2 id="verification-levels">Verification levels</h2>
              <p className="ff-muted">Each level adds deeper assurance before a tool can be considered "trusted."</p>
              <div className="ff-trust-steps">
                <div className="ff-trust-step">
                  <div className="ff-trust-step-k">L1</div>
                  <div className="ff-trust-step-v">
                    Format checks
                    <div className="ff-trust-step-s">Validate manifest structure and I/O schema.</div>
                  </div>
                </div>
                <div className="ff-trust-step">
                  <div className="ff-trust-step-k">L2</div>
                  <div className="ff-trust-step-v">
                    Security scans
                    <div className="ff-trust-step-s">Scan for risky behaviors and validate capability constraints.</div>
                  </div>
                </div>
                <div className="ff-trust-step">
                  <div className="ff-trust-step-k">L3</div>
                  <div className="ff-trust-step-v">
                    Code review
                    <div className="ff-trust-step-s">Manual and automated review of safety-relevant aspects.</div>
                  </div>
                </div>
                <div className="ff-trust-step">
                  <div className="ff-trust-step-k">L4</div>
                  <div className="ff-trust-step-v">
                    Platform verified
                    <div className="ff-trust-step-s">Signed and reviewed as official/recommended tooling.</div>
                  </div>
                </div>
              </div>
            </div>

            <div className="ff-trust-card">
              <h2>Signing, attestations, revocation</h2>
              <p className="ff-muted">Trust is portable because verification becomes an immutable record.</p>
              <ul className="ff-trust-list">
                <li><strong>Signing:</strong> Function artifacts include platform signatures.</li>
                <li><strong>Attestations:</strong> Verifiable records of what was checked and when.</li>
                <li><strong>Revocation:</strong> If a tool is flagged, it can be downgraded or removed from trusted pools.</li>
              </ul>

              <div className="ff-trust-callout">
                <div className="ff-trust-callout-title">What agents see</div>
                <div className="ff-trust-callout-body">
                  A trust score and policy-relevant metadata (capabilities, constraints, and verification level).
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--alt">
        <div className="ff-container">
          <div className="ff-trust-grid-3">
            <div className="ff-trust-card">
              <h2>Execution-backed trust</h2>
              <p className="ff-muted">
                Trust scores compound from real execution history so "it ran" and "it ran safely" are both represented.
              </p>
            </div>
            <div className="ff-trust-card">
              <h2>Zero-knowledge vault</h2>
              <p className="ff-muted">
                Secrets are encrypted client-side. FunctionFly stores ciphertext only, enabling agent tools without exposing plaintext secrets.
              </p>
            </div>
            <div className="ff-trust-card">
              <h2>Trust API (usage-based)</h2>
              <p className="ff-muted">
                Query attestations, trust scores, and revocation state. Bill on usage so agent workloads can adapt trust dynamically.
              </p>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <h2 className="ff-section-title" id="trust-lifecycle">Trust score lifecycle</h2>
          <p className="ff-section-lead">
            Trust is not a static label. It starts at publish-time, evolves with execution history, and can be updated through verification, reports, and revocations.
          </p>
          <div className="ff-trust-lifecycle-grid">
            <div className="ff-trust-life-step">
              <div className="ff-trust-life-step-k">1</div>
              <div className="ff-trust-life-step-title">Publish</div>
              <div className="ff-trust-life-step-desc">A function is published with a manifest and a declared I/O contract.</div>
            </div>
            <div className="ff-trust-life-step">
              <div className="ff-trust-life-step-k">2</div>
              <div className="ff-trust-life-step-title">Verify</div>
              <div className="ff-trust-life-step-desc">Verification stages validate structure, security, and safety-relevant behavior.</div>
            </div>
            <div className="ff-trust-life-step">
              <div className="ff-trust-life-step-k">3</div>
              <div className="ff-trust-life-step-title">Sign & attest</div>
              <div className="ff-trust-life-step-desc">Attestations become immutable records that agents can reason about.</div>
            </div>
            <div className="ff-trust-life-step">
              <div className="ff-trust-life-step-k">4</div>
              <div className="ff-trust-life-step-title">Execute</div>
              <div className="ff-trust-life-step-desc">Real execution updates trust score components over time.</div>
            </div>
            <div className="ff-trust-life-step">
              <div className="ff-trust-life-step-k">5</div>
              <div className="ff-trust-life-step-title">Evolve</div>
              <div className="ff-trust-life-step-desc">History, reports, and revocation keep trust current for agents and clients.</div>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--alt">
        <div className="ff-container">
          <div className="ff-trust-grid-2">
            <div className="ff-trust-card">
              <h2 id="trust-api">Trust API for agents</h2>
              <p className="ff-muted">
                Use a scoped API key to read trust scores and trigger verification workflows from your agent runtime.
              </p>
              <ul className="ff-trust-api-endpoints">
                <li>
                  <strong>Get trust score:</strong> <span className="ff-mono">GET /v1/trust/score/&#123;function_id&#125;</span>
                </li>
                <li>
                  <strong>Batch scores:</strong> <span className="ff-mono">POST /v1/trust/batch</span>
                </li>
                <li>
                  <strong>Get history:</strong> <span className="ff-mono">GET /v1/trust/history/&#123;function_id&#125;</span>
                </li>
                <li>
                  <strong>Submit verification:</strong> <span className="ff-mono">POST /v1/trust/verify</span>
                </li>
              </ul>

              <div className="ff-trust-callout">
                <div className="ff-trust-callout-title">Scopes matter</div>
                <div className="ff-trust-callout-body">
                  Trust read endpoints require an API key; verification and reporting endpoints require the corresponding scopes (e.g. <code>verification:request</code> and <code>reports:submit</code>).
                </div>
              </div>
            </div>

            <div className="ff-trust-card">
              <h2>Example: submit verification</h2>
              <p className="ff-muted">
                Trigger a verification run for a function version, then query status until the verification completes.
              </p>

              <div className="ff-codeblock" role="region" aria-label="Trust API example request">
                <pre><code>{`POST /v1/trust/verify
Authorization: Bearer API_KEY
Content-Type: application/json

{
  "function_id": "550e8400-e29b-41d4-a716-446655440002",
  "function_version": "1.2.0",
  "verification_level": "standard",
  "metadata": {
    "use_case": "data_processing",
    "expected_traffic": "high"
  }
}`}</code></pre>
              </div>

              <div className="ff-codeblock ff-codeblock--response" role="region" aria-label="Trust API example response">
                <pre><code>{`{
  "id": "550e8400-e29b-41d4-a716-446655440005",
  "verification_id": "vfy_abc123def456",
  "function_id": "550e8400-e29b-41d4-a716-446655440002",
  "function_author": "alice",
  "function_name": "data-processor",
  "function_version": "1.2.0",
  "verification_level": "standard",
  "status": "pending",
  "created_at": "2026-03-21T10:00:00Z"
}`}</code></pre>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <h2 className="ff-section-title" id="partner-tiers">Partner tiers</h2>
          <p className="ff-section-lead">
            Each tier maps to a verification level, giving your agents the right trust guarantees for the right workload.
          </p>
          <div className="ff-trust-partner-grid">
            <div className="ff-trust-partner-card">
              <div className="ff-trust-partner-header">
                <div className="ff-trust-partner-tier">Developer</div>
                <div className="ff-trust-partner-level">L1–L2</div>
              </div>
              <div className="ff-trust-partner-price">Free</div>
              <ul className="ff-trust-partner-features">
                <li>Format validation (L1)</li>
                <li>Security scan results (L2)</li>
                <li>Basic trust scores</li>
                <li>10k API calls/month</li>
                <li>Community support</li>
              </ul>
              <a href={`${APP_DASHBOARD_ORIGIN}/settings#trust-api`} className="ff-btn ff-btn-secondary">Get Started</a>
            </div>
            <div className="ff-trust-partner-card ff-trust-partner-card--featured">
              <div className="ff-trust-partner-header">
                <div className="ff-trust-partner-tier">Startup</div>
                <div className="ff-trust-partner-level">L2–L3</div>
              </div>
              <div className="ff-trust-partner-price">$49<span>/mo</span></div>
              <ul className="ff-trust-partner-features">
                <li>Security scans (L2)</li>
                <li>Code review attestations (L3)</li>
                <li>Full trust score history</li>
                <li>100k API calls/month</li>
                <li>Email support</li>
              </ul>
              <a href={`${APP_DASHBOARD_ORIGIN}/settings#trust-api?checkout=startup`} className="ff-btn ff-btn-primary">Subscribe</a>
            </div>
            <div className="ff-trust-partner-card">
              <div className="ff-trust-partner-header">
                <div className="ff-trust-partner-tier">Business</div>
                <div className="ff-trust-partner-level">L3–L4</div>
              </div>
              <div className="ff-trust-partner-price">$199<span>/mo</span></div>
              <ul className="ff-trust-partner-features">
                <li>Code review + attestations (L3)</li>
                <li>Platform verification (L4)</li>
                <li>Revocation monitoring</li>
                <li>1M API calls/month</li>
                <li>Priority support + SLAs</li>
              </ul>
              <a href={`${APP_DASHBOARD_ORIGIN}/settings#trust-api?checkout=business`} className="ff-btn ff-btn-secondary">Subscribe</a>
            </div>
            <div className="ff-trust-partner-card">
              <div className="ff-trust-partner-header">
                <div className="ff-trust-partner-tier">Enterprise</div>
                <div className="ff-trust-partner-level">L4 + Custom</div>
              </div>
              <div className="ff-trust-partner-price">Custom</div>
              <ul className="ff-trust-partner-features">
                <li>All Business features</li>
                <li>Custom verification policies</li>
                <li>Dedicated trust consultant</li>
                <li>Unlimited API calls</li>
                <li>On-premise deployment option</li>
              </ul>
              <a href="/contact" className="ff-btn ff-btn-secondary">Contact Sales</a>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--alt">
        <div className="ff-container">
          <h2 className="ff-section-title">Why FunctionFly wins on trust</h2>
          <p className="ff-section-lead">No other platform combines verification, trust scoring, and zero-knowledge security.</p>
          <div className="ff-trust-comparison-table">
            <div className="ff-trust-comparison-header">
              <div className="ff-trust-comparison-cell">Feature</div>
              <div className="ff-trust-comparison-cell ff-trust-comparison-cell--highlight">FunctionFly</div>
              <div className="ff-trust-comparison-cell">RapidAPI</div>
              <div className="ff-trust-comparison-cell">Toolhouse</div>
              <div className="ff-trust-comparison-cell">Theagora</div>
            </div>
            <div className="ff-trust-comparison-row">
              <div className="ff-trust-comparison-cell">Trust scoring</div>
              <div className="ff-trust-comparison-cell ff-trust-comparison-cell--highlight"><span className="ff-check">✓</span></div>
              <div className="ff-trust-comparison-cell"><span className="ff-dash">—</span></div>
              <div className="ff-trust-comparison-cell"><span className="ff-dash">—</span></div>
              <div className="ff-trust-comparison-cell"><span className="ff-partial">Partial</span></div>
            </div>
            <div className="ff-trust-comparison-row">
              <div className="ff-trust-comparison-cell">Multi-level verification</div>
              <div className="ff-trust-comparison-cell ff-trust-comparison-cell--highlight"><span className="ff-check">✓</span></div>
              <div className="ff-trust-comparison-cell"><span className="ff-dash">—</span></div>
              <div className="ff-trust-comparison-cell"><span className="ff-dash">—</span></div>
              <div className="ff-trust-comparison-cell"><span className="ff-check">✓</span></div>
            </div>
            <div className="ff-trust-comparison-row">
              <div className="ff-trust-comparison-cell">Zero-knowledge vault</div>
              <div className="ff-trust-comparison-cell ff-trust-comparison-cell--highlight"><span className="ff-check">✓</span></div>
              <div className="ff-trust-comparison-cell"><span className="ff-dash">—</span></div>
              <div className="ff-trust-comparison-cell"><span className="ff-dash">—</span></div>
              <div className="ff-trust-comparison-cell"><span className="ff-dash">—</span></div>
            </div>
            <div className="ff-trust-comparison-row">
              <div className="ff-trust-comparison-cell">Function execution</div>
              <div className="ff-trust-comparison-cell ff-trust-comparison-cell--highlight"><span className="ff-check">✓</span></div>
              <div className="ff-trust-comparison-cell"><span className="ff-dash">—</span></div>
              <div className="ff-trust-comparison-cell"><span className="ff-dash">—</span></div>
              <div className="ff-trust-comparison-cell"><span className="ff-check">✓</span></div>
            </div>
            <div className="ff-trust-comparison-row">
              <div className="ff-trust-comparison-cell">Agent-native tool discovery</div>
              <div className="ff-trust-comparison-cell ff-trust-comparison-cell--highlight"><span className="ff-check">✓</span></div>
              <div className="ff-trust-comparison-cell"><span className="ff-dash">—</span></div>
              <div className="ff-trust-comparison-cell"><span className="ff-check">✓</span></div>
              <div className="ff-trust-comparison-cell"><span className="ff-check">✓</span></div>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-cta-section">
        <div className="ff-container">
          <h2>Ready to build with trusted tools?</h2>
          <p>Browse trusted functions in the marketplace, or run your own agent trust policy to select tools safely.</p>
          <div className="ff-actions">
            <a className="ff-btn ff-btn-primary" href={`${AUTH_ORIGIN}/signup`}>Get Started Free</a>
            <a className="ff-btn ff-btn-secondary" href="/pricing">See trust-aligned pricing</a>
          </div>
        </div>
      </section>
    </div>
  )
}

export default TrustPage
