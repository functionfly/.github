import React from 'react'
import { APP_DASHBOARD_ORIGIN, AUTH_ORIGIN } from '../config'
import './homepage.css'

const StateFabricPage: React.FC = () => {
  return (
    <div className="ff-homepage">
      <section className="ff-hero-section ff-statefabric-hero">
        <div className="ff-hero-inner">
          <div className="ff-hero-eyebrow">
            <span className="ff-pulse-dot" />
            <span>State Fabric</span>
          </div>
          <h1 className="ff-hero-headline">
            Stateful capabilities<br />for serverless
          </h1>
          <p className="ff-hero-sub">
            From sandbox experimentation to enterprise-grade state management. Start free and scale up as your application grows.
          </p>
          <div className="ff-hero-actions">
            <a className="ff-btn ff-btn-primary" href={`${AUTH_ORIGIN}/signup`}>Try now — free</a>
            <a className="ff-btn ff-btn-secondary" href="/pricing">Back to pricing</a>
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <div className="ff-pricing-grid">
            <div className="ff-price-card">
              <h3>Sandbox</h3>
              <div className="ff-price-row">
                <span className="ff-price-amount">$0</span>
                <span className="ff-price-period">/ month</span>
              </div>
              <p className="ff-price-desc">Great for experimentation & onboarding.</p>
              <ul className="ff-feature-list">
                <li><span className="ff-check">✓</span> Stateless usage</li>
                <li><span className="ff-check">✓</span> 1 state object</li>
                <li><span className="ff-check">✓</span> 1 GB event storage</li>
                <li><span className="ff-check">✓</span> 10,000 state ops</li>
                <li><span className="ff-check">✓</span> 7-day snapshot retention</li>
                <li><span className="ff-check">✓</span> Community support</li>
              </ul>
              <a className="ff-btn ff-btn-secondary" href={`${AUTH_ORIGIN}/signup`}>Try now — free</a>
            </div>

            <div className="ff-price-card">
              <h3>Starter</h3>
              <div className="ff-price-row">
                <span className="ff-price-amount">$19</span>
                <span className="ff-price-period">/ month</span>
              </div>
              <p className="ff-price-desc">Internal dev & small side projects.</p>
              <p className="ff-price-note">Storage: $0.10/GB · Ops: $0.50/100k</p>
              <ul className="ff-feature-list">
                <li><span className="ff-check">✓</span> Up to 5 state objects</li>
                <li><span className="ff-check">✓</span> 10 GB event storage</li>
                <li><span className="ff-check">✓</span> 100,000 state ops</li>
                <li><span className="ff-check">✓</span> 30-day snapshot retention</li>
                <li><span className="ff-check">✓</span> Basic snapshot scheduling</li>
                <li><span className="ff-check">✓</span> Replay engine (limited)</li>
              </ul>
              <a className="ff-btn ff-btn-primary" href={`${APP_DASHBOARD_ORIGIN}/pricing?tab=state-fabric&plan=sf_starter`}>Choose Starter</a>
            </div>

            <div className="ff-price-card ff-price-card--featured">
              <h3>Pro</h3>
              <div className="ff-price-row">
                <span className="ff-price-amount">$99</span>
                <span className="ff-price-period">/ month</span>
              </div>
              <p className="ff-price-desc">Production apps & team projects.</p>
              <p className="ff-price-note">Storage: $0.08/GB · Ops: $0.35/100k</p>
              <ul className="ff-feature-list">
                <li><span className="ff-check">✓</span> Up to 50 state objects</li>
                <li><span className="ff-check">✓</span> 100 GB event storage</li>
                <li><span className="ff-check">✓</span> 1M state ops</li>
                <li><span className="ff-check">✓</span> 90-day snapshot retention</li>
                <li><span className="ff-check">✓</span> Hot cache tier</li>
                <li><span className="ff-check">✓</span> Fast deterministic replay</li>
                <li><span className="ff-check">✓</span> Replay API + billing analytics</li>
              </ul>
              <a className="ff-btn ff-btn-primary" href={`${APP_DASHBOARD_ORIGIN}/pricing?tab=state-fabric&plan=sf_pro`}>Choose Pro</a>
            </div>

            <div className="ff-price-card">
              <h3>Business</h3>
              <div className="ff-price-row">
                <span className="ff-price-amount">$499</span>
                <span className="ff-price-period">/ month</span>
              </div>
              <p className="ff-price-desc">Business-critical systems.</p>
              <p className="ff-price-note">Storage: $0.06/GB · Archive: $30/TB/mo</p>
              <ul className="ff-feature-list">
                <li><span className="ff-check">✓</span> Up to 500 state objects</li>
                <li><span className="ff-check">✓</span> 1 TB event storage</li>
                <li><span className="ff-check">✓</span> 10M state ops</li>
                <li><span className="ff-check">✓</span> 180-day retention</li>
                <li><span className="ff-check">✓</span> Multi-region replication</li>
                <li><span className="ff-check">✓</span> Dedicated hot cache</li>
                <li><span className="ff-check">✓</span> SLA + advanced analytics</li>
                <li><span className="ff-check">✓</span> Event subscription streams</li>
              </ul>
              <a className="ff-btn ff-btn-primary" href={`${APP_DASHBOARD_ORIGIN}/pricing?tab=state-fabric&plan=sf_business`}>Choose Business</a>
            </div>

            <div className="ff-price-card">
              <h3>Enterprise</h3>
              <div className="ff-price-row">
                <span className="ff-price-amount ff-price-amount--custom">Custom</span>
              </div>
              <p className="ff-price-desc">Large enterprises & regulated industries.</p>
              <ul className="ff-feature-list">
                <li><span className="ff-check">✓</span> Unlimited state objects</li>
                <li><span className="ff-check">✓</span> Unlimited event storage</li>
                <li><span className="ff-check">✓</span> Unlimited ops</li>
                <li><span className="ff-check">✓</span> 365-day retention</li>
                <li><span className="ff-check">✓</span> Multi-region + edge replication</li>
                <li><span className="ff-check">✓</span> Enterprise key management (BYOK)</li>
                <li><span className="ff-check">✓</span> Immutable audit logs</li>
                <li><span className="ff-check">✓</span> Replay export · Dedicated support</li>
              </ul>
              <a className="ff-btn ff-btn-secondary" href="/contact">Contact sales</a>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--alt">
        <div className="ff-container">
          <h2 className="ff-section-title">Compare State Fabric plans</h2>
          <div className="ff-comparison-table">
            <div className="ff-comparison-header">
              <div>Feature</div>
              <div>Sandbox</div>
              <div>Starter</div>
              <div>Pro</div>
              <div>Business</div>
              <div>Enterprise</div>
            </div>
            <div className="ff-comparison-row">
              <div>State objects</div>
              <div>1</div>
              <div>5</div>
              <div>50</div>
              <div>500</div>
              <div>Unlimited</div>
            </div>
            <div className="ff-comparison-row">
              <div>Event storage</div>
              <div>1 GB</div>
              <div>10 GB</div>
              <div>100 GB</div>
              <div>1 TB</div>
              <div>Unlimited</div>
            </div>
            <div className="ff-comparison-row">
              <div>State ops</div>
              <div>10K</div>
              <div>100K</div>
              <div>1M</div>
              <div>10M</div>
              <div>Unlimited</div>
            </div>
            <div className="ff-comparison-row">
              <div>Snapshot retention</div>
              <div>7 days</div>
              <div>30 days</div>
              <div>90 days</div>
              <div>180 days</div>
              <div>365 days</div>
            </div>
            <div className="ff-comparison-row">
              <div>Replay</div>
              <div>—</div>
              <div>Limited</div>
              <div>Fast + API</div>
              <div>Full</div>
              <div>Export</div>
            </div>
            <div className="ff-comparison-row">
              <div>Multi-region</div>
              <div>—</div>
              <div>—</div>
              <div>—</div>
              <div>Yes</div>
              <div>Yes + edge</div>
            </div>
            <div className="ff-comparison-row">
              <div>Hot cache</div>
              <div>—</div>
              <div>—</div>
              <div>Yes</div>
              <div>Dedicated</div>
              <div>Dedicated</div>
            </div>
            <div className="ff-comparison-row">
              <div>Support</div>
              <div>Community</div>
              <div>Email</div>
              <div>Priority</div>
              <div>SLA</div>
              <div>Dedicated</div>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <h2 className="ff-section-title">Optional add-ons</h2>
          <p className="ff-section-lead">
            Enhance any State Fabric plan with performance, security, or analytics add-ons.
          </p>
          <div className="ff-addon-grid">
            <div className="ff-addon-card">
              <h4>Hot Cache Booster</h4>
              <div className="ff-addon-price">$49 <span>/mo per 5GB</span></div>
              <p>Reduces replay and read costs.</p>
            </div>
            <div className="ff-addon-card">
              <h4>Advanced Security Pack</h4>
              <div className="ff-addon-price">$99 <span>/mo</span></div>
              <p>SOC2-friendly logs, key rotation, audit streams.</p>
            </div>
            <div className="ff-addon-card">
              <h4>AI Memory Pack</h4>
              <div className="ff-addon-price">$149 <span>/mo</span></div>
              <p>Vector index, embeddings storage, fast read engine.</p>
            </div>
            <div className="ff-addon-card">
              <h4>Advanced Insights</h4>
              <div className="ff-addon-price">$79 <span>/mo</span></div>
              <p>Cost forecasting, anomaly detection, hot path alerts.</p>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-cta-section">
        <div className="ff-container">
          <h2>Ready to add state to your functions?</h2>
          <p>Start with the free Sandbox or choose a plan that fits your scale.</p>
          <div className="ff-actions">
            <a className="ff-btn ff-btn-primary" href={`${AUTH_ORIGIN}/signup`}>Try now — free</a>
            <a className="ff-btn ff-btn-secondary" href="/contact">Talk to sales</a>
          </div>
        </div>
      </section>
    </div>
  )
}

export default StateFabricPage
