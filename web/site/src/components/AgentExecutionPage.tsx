import React from 'react'
import { APP_DASHBOARD_ORIGIN, AUTH_ORIGIN } from '../config'
import './homepage.css'
import './pricing.css'

const AgentExecutionPage: React.FC = () => {
  return (
    <div className="ff-homepage">
      <section className="ff-hero-section ff-statefabric-hero">
        <div className="ff-hero-inner">
          <div className="ff-hero-eyebrow">
            <span className="ff-pulse-dot" />
            <span>Agent Execution</span>
          </div>
          <h1 className="ff-hero-headline">
            Ship AI agents that<br />actually work in production
          </h1>
          <p className="ff-hero-sub">
            Stop wrestling with tool calls, state management, and credential rotation. Get infrastructure that just works so you can focus on building agents that ship.
          </p>
          <div className="ff-hero-actions">
            <a className="ff-btn ff-btn-primary" href={`${AUTH_ORIGIN}/signup`}>
              Start free — no credit card
            </a>
            <a className="ff-btn ff-btn-secondary" href="/pricing">View pricing</a>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--alt">
        <div className="ff-container">
          <h2 className="ff-section-title">Why teams choose Agent Execution</h2>
          <div className="ff-addon-grid">
            <div className="ff-addon-card">
              <h4>Stop rebuilding the boring stuff</h4>
              <p>Tool calls, retry logic, state persistence, credential rotation — we handle the plumbing so you can focus on what makes your agent unique.</p>
            </div>
            <div className="ff-addon-card">
              <h4>Scale without rewriting</h4>
              <p>From prototype to production without refactoring. Our infrastructure handles concurrency, rate limiting, and fallback logic automatically.</p>
            </div>
            <div className="ff-addon-card">
              <h4>Credentials that rotate themselves</h4>
              <p>Dynamic credentials mean zero-downtime credential rotation. No redeploys, no incidents, no on-call at 3am.</p>
            </div>
            <div className="ff-addon-card">
              <h4>Debug what actually happened</h4>
              <p>Full execution traces, tool call logs, and state snapshots. Know exactly why your agent did what it did.</p>
            </div>
            <div className="ff-addon-card">
              <h4>Memory that persists</h4>
              <p>Context windows are expensive. Store what matters in persistent memory and retrieve it when relevant.</p>
            </div>
            <div className="ff-addon-card">
              <h4>Security without the headache</h4>
              <p>SOC2-ready audit logs, automatic key rotation, IP allowlisting. Enterprise security without the enterprise procurement process.</p>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <div className="ff-pricing-grid">
            <div className="ff-price-card">
              <h3>Free</h3>
              <div className="ff-price-row">
                <span className="ff-price-amount">$0</span>
                <span className="ff-price-period">/ month</span>
              </div>
              <p className="ff-price-desc">For prototyping and personal projects.</p>
              <ul className="ff-feature-list">
                <li><span className="ff-check">✓</span> 3 AI agents</li>
                <li><span className="ff-check">✓</span> 10,000 calls / month</li>
                <li><span className="ff-check">✓</span> 3 concurrency</li>
                <li><span className="ff-check">✓</span> Dynamic credentials (PostgreSQL)</li>
                <li><span className="ff-check">✓</span> Basic state management</li>
                <li><span className="ff-check">✓</span> Community support</li>
              </ul>
              <a className="ff-btn ff-btn-secondary" href={`${AUTH_ORIGIN}/signup`}>Start free</a>
            </div>

            <div className="ff-price-card">
              <h3>Starter</h3>
              <div className="ff-price-row">
                <span className="ff-price-amount">$24</span>
                <span className="ff-price-period">/ month</span>
              </div>
              <p className="ff-price-desc">For indie devs and small teams shipping agent-powered apps.</p>
              <p className="ff-price-note">Overage: $0.40 per 1K calls</p>
              <ul className="ff-feature-list">
                <li><span className="ff-check">✓</span> 10 AI agents</li>
                <li><span className="ff-check">✓</span> 100,000 calls / month</li>
                <li><span className="ff-check">✓</span> 10 concurrency</li>
                <li><span className="ff-check">✓</span> Dynamic credentials (PostgreSQL)</li>
                <li><span className="ff-check">✓</span> Full state management</li>
                <li><span className="ff-check">✓</span> Memory & context retention</li>
                <li><span className="ff-check">✓</span> Email support</li>
              </ul>
              <a className="ff-btn ff-btn-primary" href={`${APP_DASHBOARD_ORIGIN}/pricing?tab=agents&plan=starter`}>Choose Starter</a>
            </div>

            <div className="ff-price-card ff-price-card--featured">
              <span className="ff-featured-badge">Most Popular</span>
              <h3>Professional</h3>
              <div className="ff-price-row">
                <span className="ff-price-amount">$79</span>
                <span className="ff-price-period">/ month</span>
              </div>
              <p className="ff-price-desc">For growing teams with production agents.</p>
              <p className="ff-price-note">Overage: $0.30 per 1K calls</p>
              <ul className="ff-feature-list">
                <li><span className="ff-check">✓</span> 100 AI agents</li>
                <li><span className="ff-check">✓</span> 1,000,000 calls / month</li>
                <li><span className="ff-check">✓</span> 100 concurrency</li>
                <li><span className="ff-check">✓</span> Dynamic credentials (PostgreSQL + MySQL)</li>
                <li><span className="ff-check">✓</span> Advanced state management</li>
                <li><span className="ff-check">✓</span> Memory & context retention</li>
                <li><span className="ff-check">✓</span> Tool call analytics</li>
                <li><span className="ff-check">✓</span> Priority support</li>
              </ul>
              <a className="ff-btn ff-btn-primary" href={`${APP_DASHBOARD_ORIGIN}/pricing?tab=agents&plan=professional`}>Choose Professional</a>
            </div>

            <div className="ff-price-card">
              <h3>Enterprise</h3>
              <div className="ff-price-row">
                <span className="ff-price-amount">$299</span>
                <span className="ff-price-period">/ month</span>
              </div>
              <p className="ff-price-desc">For large-scale agent deployments.</p>
              <p className="ff-price-note">Overage: $0.20 per 1K calls</p>
              <ul className="ff-feature-list">
                <li><span className="ff-check">✓</span> 500 AI agents</li>
                <li><span className="ff-check">✓</span> 5,000,000 calls / month</li>
                <li><span className="ff-check">✓</span> 500 concurrency</li>
                <li><span className="ff-check">✓</span> Dynamic credentials (PostgreSQL + MySQL + Redis)</li>
                <li><span className="ff-check">✓</span> Full state management</li>
                <li><span className="ff-check">✓</span> Memory & context retention</li>
                <li><span className="ff-check">✓</span> Tool call analytics & cost breakdown</li>
                <li><span className="ff-check">✓</span> 99.99% SLA</li>
                <li><span className="ff-check">✓</span> Dedicated support</li>
              </ul>
              <a className="ff-btn ff-btn-primary" href={`${APP_DASHBOARD_ORIGIN}/pricing?tab=agents&plan=enterprise`}>Choose Enterprise</a>
            </div>

            <div className="ff-price-card">
              <h3>Agent Enterprise</h3>
              <div className="ff-price-row">
                <span className="ff-price-amount ff-price-amount--custom">Custom</span>
              </div>
              <p className="ff-price-desc">For unlimited scale and dedicated infrastructure.</p>
              <ul className="ff-feature-list">
                <li><span className="ff-check">✓</span> Unlimited AI agents</li>
                <li><span className="ff-check">✓</span> Unlimited calls</li>
                <li><span className="ff-check">✓</span> 2,000+ calls/min throughput</li>
                <li><span className="ff-check">✓</span> All database drivers</li>
                <li><span className="ff-check">✓</span> Unlimited state & memory</li>
                <li><span className="ff-check">✓</span> Dedicated infrastructure</li>
                <li><span className="ff-check">✓</span> SSO & SAML support</li>
                <li><span className="ff-check">✓</span> High availability status</li>
                <li><span className="ff-check">✓</span> Custom SLAs</li>
              </ul>
              <a className="ff-btn ff-btn-secondary" href="/contact">Contact sales</a>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--alt">
        <div className="ff-container">
          <h2 className="ff-section-title">Compare plans</h2>
          <div className="ff-comparison-table">
            <div className="ff-comparison-header">
              <div>Feature</div>
              <div>Free</div>
              <div>Starter</div>
              <div>Professional</div>
              <div>Enterprise</div>
              <div>Agent Enterprise</div>
            </div>
            <div className="ff-comparison-row">
              <div>AI agents</div>
              <div>3</div>
              <div>10</div>
              <div>100</div>
              <div>500</div>
              <div>Unlimited</div>
            </div>
            <div className="ff-comparison-row">
              <div>Calls / month</div>
              <div>10K</div>
              <div>100K</div>
              <div>1M</div>
              <div>5M</div>
              <div>Unlimited</div>
            </div>
            <div className="ff-comparison-row">
              <div>Concurrency</div>
              <div>3</div>
              <div>10</div>
              <div>100</div>
              <div>500</div>
              <div>2K+/min</div>
            </div>
            <div className="ff-comparison-row">
              <div>Dynamic credentials</div>
              <div>PostgreSQL</div>
              <div>PostgreSQL</div>
              <div>PostgreSQL + MySQL</div>
              <div>PostgreSQL + MySQL + Redis</div>
              <div>All drivers</div>
            </div>
            <div className="ff-comparison-row">
              <div>State management</div>
              <div>Basic</div>
              <div>Full</div>
              <div>Advanced</div>
              <div>Full</div>
              <div>Unlimited</div>
            </div>
            <div className="ff-comparison-row">
              <div>Memory & context</div>
              <div>—</div>
              <div>Yes</div>
              <div>Yes</div>
              <div>Yes</div>
              <div>Unlimited</div>
            </div>
            <div className="ff-comparison-row">
              <div>Tool call analytics</div>
              <div>—</div>
              <div>—</div>
              <div>Yes</div>
              <div>Advanced</div>
              <div>Custom</div>
            </div>
            <div className="ff-comparison-row">
              <div>SLA</div>
              <div>—</div>
              <div>—</div>
              <div>99.9%</div>
              <div>99.99%</div>
              <div>Custom</div>
            </div>
            <div className="ff-comparison-row">
              <div>Support</div>
              <div>Community</div>
              <div>Email</div>
              <div>Priority</div>
              <div>Dedicated</div>
              <div>Dedicated CSM</div>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <h2 className="ff-section-title">Add-ons</h2>
          <p className="ff-section-lead">
            Expand your plan without upgrading tiers.
          </p>
          <div className="ff-addon-grid">
            <div className="ff-addon-card">
              <h4>Extra Calls Pack</h4>
              <div className="ff-addon-price">$19 <span>/mo per 100K calls</span></div>
              <p>Expand your monthly call limit without upgrading tiers.</p>
            </div>
            <div className="ff-addon-card">
              <h4>Concurrency Booster</h4>
              <div className="ff-addon-price">$49 <span>/mo per 50 concurrent</span></div>
              <p>Handle more simultaneous agent executions.</p>
            </div>
            <div className="ff-addon-card">
              <h4>Advanced Analytics</h4>
              <div className="ff-addon-price">$79 <span>/mo</span></div>
              <p>Cost forecasting, anomaly detection, and agent performance insights.</p>
            </div>
            <div className="ff-addon-card">
              <h4>Security Pack</h4>
              <div className="ff-addon-price">$99 <span>/mo</span></div>
              <p>SOC2-friendly audit logs, key rotation, and IP allowlisting.</p>
            </div>
            <div className="ff-addon-card">
              <h4>Additional Database Drivers</h4>
              <div className="ff-addon-price">$29 <span>/mo per driver</span></div>
              <p>Enable Redis, MongoDB, or other database backends for dynamic credentials.</p>
            </div>
            <div className="ff-addon-card">
              <h4>Extended Memory Retention</h4>
              <div className="ff-addon-price">$149 <span>/mo</span></div>
              <p>Extend agent context and memory retention beyond plan limits.</p>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--alt">
        <div className="ff-container">
          <h2 className="ff-section-title">Frequently asked questions</h2>
          <dl className="pricing-faq">
            <div className="pricing-faq-item">
              <dt className="pricing-faq-question">How is Agent Execution different from just using an LLM API?</dt>
              <dd className="pricing-faq-answer">LLM APIs give you model access. Agent Execution gives you infrastructure — tool call execution, state persistence, credential rotation, retry logic, and observability. It's the difference between a function and a production service.</dd>
            </div>
            <div className="pricing-faq-item">
              <dt className="pricing-faq-question">What counts as a "call"?</dt>
              <dd className="pricing-faq-answer">A call is one execution of your agent's logic, including all tool calls, state operations, and API requests within that execution. Streaming responses count as one call.</dd>
            </div>
            <div className="pricing-faq-item">
              <dt className="pricing-faq-question">Can I switch plans anytime?</dt>
              <dd className="pricing-faq-answer">Yes. Upgrade instantly, downgrade at the end of your billing period. No lock-in, no migration required.</dd>
            </div>
            <div className="pricing-faq-item">
              <dt className="pricing-faq-question">What happens if I exceed my call limit?</dt>
              <dd className="pricing-faq-answer">We'll notify you when you hit 80% of your limit. Overages are billed at the rate shown for each plan. You can also set spending alerts to avoid surprises.</dd>
            </div>
            <div className="pricing-faq-item">
              <dt className="pricing-faq-question">Do you offer a free trial for paid plans?</dt>
              <dd className="pricing-faq-answer">The Free tier is your trial — no credit card required. When you're ready to scale, Starter and above give you more capacity to handle production workloads.</dd>
            </div>
            <div className="pricing-faq-item">
              <dt className="pricing-faq-question">What dynamic credentials are supported?</dt>
              <dd className="pricing-faq-answer">PostgreSQL and MySQL are available on all plans. Redis, MongoDB, and other databases are available as add-ons or included in Enterprise plans.</dd>
            </div>
            <div className="pricing-faq-item">
              <dt className="pricing-faq-question">Is my agent data secure?</dt>
              <dd className="pricing-faq-answer">Yes. All data is encrypted in transit and at rest. We maintain SOC2 Type II compliance, provide audit logs, and support BYOK (bring your own key) on Enterprise plans.</dd>
            </div>
          </dl>
        </div>
      </section>

      <section className="ff-cta-section">
        <div className="ff-container">
          <h2>Start building agents that ship</h2>
          <p>Free tier includes everything you need to prototype. No credit card required.</p>
          <div className="ff-actions">
            <a className="ff-btn ff-btn-primary" href={`${AUTH_ORIGIN}/signup`}>Start free — no credit card</a>
            <a className="ff-btn ff-btn-secondary" href="/contact">Talk to sales</a>
          </div>
        </div>
      </section>
    </div>
  )
}

export default AgentExecutionPage