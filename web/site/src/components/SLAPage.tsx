import React from 'react'
import {
  PageGrid,
  Chamber,
  CornerBrace,
  StatusPill,
  GaugeStrip,
  Gauge,
  GaugeValue,
  GaugeLabel,
  AnnotationTag,
} from './containment'
import '../styles/sc-main.css';

const ShieldIcon = () => (
  <svg width="32" height="32" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
    <path d="m9 12 2 2 4-4" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
)

const SLAPage: React.FC = () => {
  return (
    <>
      <PageGrid />

      {/* Hero Section */}
      <section className="trust-hero">
        <Chamber variant="ribs" className="trust-hero-chamber">
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <AnnotationTag title="LEGAL / SLA" subtitle="Page 01" />
          <div className="trust-hero-content">
            <div className="trust-hero-eyebrow">
              <span className="trust-pulse-dot" />
              <span>Service Level Agreement</span>
            </div>
            <h1 className="trust-hero-title">
              Uptime commitments<br />and service credits
            </h1>
            <p className="trust-hero-subtitle">
              Uptime commitments for eligible FunctionFly platform subscriptions, how we measure availability, and service credits when we miss those targets.
            </p>
          </div>
          <div className="trust-hero-visual">
            <ShieldIcon />
          </div>
        </Chamber>
      </section>

      {/* SLA Overview */}
      <section className="trust-section">
        <div className="trust-container">
          <h2 className="trust-section-title">Plan uptime targets</h2>
          <p className="trust-section-lead">
            Free, Starter, and other plans are provided on a best-effort basis. Professional and Enterprise plans include the following uptime commitments.
          </p>
          <GaugeStrip className="trust-lifecycle-strip">
            <Gauge>
              <GaugeValue><span className="gauge-tick" />99.9%</GaugeValue>
              <GaugeLabel>Professional</GaugeLabel>
              <p className="trust-lifecycle-desc">For Professional plan workspaces. 10% credit if below 99.9%, 25% if below 99.0%.</p>
            </Gauge>
            <Gauge>
              <GaugeValue><span className="gauge-tick" />99.99%</GaugeValue>
              <GaugeLabel>Enterprise</GaugeLabel>
              <p className="trust-lifecycle-desc">For Enterprise plan workspaces. 10% credit if below 99.99%, 25% if below 99.9%.</p>
            </Gauge>
          </GaugeStrip>
        </div>
      </section>

      {/* Definitions */}
      <section className="trust-section trust-section--alt">
        <div className="trust-container">
          <div className="trust-grid-2">
            <Chamber className="trust-card">
              <CornerBrace position="tr" />
              <h2>Who this applies to</h2>
              <p className="trust-muted">
                This SLA is between you and FunctionFly LLC. It supplements our <a href="/terms" className="trust-link">Terms of Service</a> for customers with an active paid subscription that explicitly includes an uptime SLA.
              </p>
              <p className="trust-muted">
                If you signed an order form, statement of work, or other written agreement that includes different uptime targets, <strong>that document controls</strong>.
              </p>
            </Chamber>

            <Chamber className="trust-card">
              <CornerBrace position="bl" />
              <h2>Definitions</h2>
              <ul className="trust-list">
                <li><strong>Covered Services:</strong> The generally available FunctionFly control plane including dashboard and documented HTTPS API endpoints.</li>
                <li><strong>Downtime:</strong> Minutes when Covered Services return consecutive HTTP 5xx errors or fail to establish a TLS connection.</li>
                <li><strong>Monthly Uptime Percentage:</strong> (Total minutes − Downtime minutes) ÷ Total minutes × 100.</li>
                <li><strong>Service Credit:</strong> A percentage of subscription fees applied as a credit for failing to meet uptime targets.</li>
              </ul>
            </Chamber>
          </div>
        </div>
      </section>

      {/* Professional Credits */}
      <section className="trust-section">
        <div className="trust-container">
          <h2 className="trust-section-title">Service credits — Professional (99.9% target)</h2>
          <div className="trust-grid-3">
            <Chamber nested className="trust-partner-card">
              <div className="trust-partner-header">
                <span className="trust-partner-tier">10% Credit</span>
              </div>
              <p className="trust-muted">If Monthly Uptime Percentage is:</p>
              <code className="trust-code">&lt; 99.9% but ≥ 99.0%</code>
            </Chamber>
            <Chamber nested className="trust-partner-card">
              <div className="trust-partner-header">
                <span className="trust-partner-tier">25% Credit</span>
              </div>
              <p className="trust-muted">If Monthly Uptime Percentage is:</p>
              <code className="trust-code">&lt; 99.0% but ≥ 95.0%</code>
            </Chamber>
            <Chamber nested className="trust-partner-card">
              <div className="trust-partner-header">
                <span className="trust-partner-tier">50% Credit</span>
              </div>
              <p className="trust-muted">If Monthly Uptime Percentage is:</p>
              <code className="trust-code">&lt; 95.0%</code>
            </Chamber>
          </div>
        </div>
      </section>

      {/* Enterprise Credits */}
      <section className="trust-section trust-section--alt">
        <div className="trust-container">
          <h2 className="trust-section-title">Service credits — Enterprise (99.99% target)</h2>
          <div className="trust-grid-3">
            <Chamber nested className="trust-partner-card">
              <div className="trust-partner-header">
                <span className="trust-partner-tier">10% Credit</span>
              </div>
              <p className="trust-muted">If Monthly Uptime Percentage is:</p>
              <code className="trust-code">&lt; 99.99% but ≥ 99.9%</code>
            </Chamber>
            <Chamber nested className="trust-partner-card">
              <div className="trust-partner-header">
                <span className="trust-partner-tier">25% Credit</span>
              </div>
              <p className="trust-muted">If Monthly Uptime Percentage is:</p>
              <code className="trust-code">&lt; 99.9% but ≥ 99.0%</code>
            </Chamber>
            <Chamber nested className="trust-partner-card">
              <div className="trust-partner-header">
                <span className="trust-partner-tier">50% Credit</span>
              </div>
              <p className="trust-muted">If Monthly Uptime Percentage is:</p>
              <code className="trust-code">&lt; 99.0%</code>
            </Chamber>
          </div>
        </div>
      </section>

      {/* Exclusions */}
      <section className="trust-section">
        <div className="trust-container">
          <h2 className="trust-section-title">Exclusions</h2>
          <p className="trust-section-lead">
            Downtime does not include unavailability caused by or resulting from:
          </p>
          <div className="trust-grid-2">
            <Chamber className="trust-card">
              <CornerBrace position="tr" />
              <h3>Maintenance</h3>
              <ul className="trust-list">
                <li>Scheduled maintenance with 48+ hours notice</li>
                <li>Emergency maintenance for security or stability issues</li>
              </ul>
            </Chamber>

            <Chamber className="trust-card">
              <CornerBrace position="bl" />
              <h3>External factors</h3>
              <ul className="trust-list">
                <li>Your misuse or violations of Terms</li>
                <li>Your networks, devices, DNS, or third-party services</li>
                <li>Force majeure events or internet outages</li>
              </ul>
            </Chamber>

            <Chamber className="trust-card">
              <CornerBrace position="tr" />
              <h3>Not covered</h3>
              <ul className="trust-list">
                <li>Beta, preview, or experimental features</li>
                <li>Your function execution on third-party runtimes</li>
                <li>Suspension or termination permitted under Terms</li>
              </ul>
            </Chamber>

            <Chamber className="trust-card">
              <CornerBrace position="bl" />
              <h3>Measurement</h3>
              <p className="trust-muted">
                We measure using our internal monitoring systems. Enterprise customers may use the SLA dashboard as a convenience. In case of conflicts, our internal records govern.
              </p>
            </Chamber>
          </div>
        </div>
      </section>

      {/* How to Request */}
      <section className="trust-section trust-section--alt">
        <div className="trust-container">
          <h2 className="trust-section-title">How to request a credit</h2>
          <div className="trust-grid-2">
            <Chamber className="trust-card">
              <CornerBrace position="tr" />
              <h3>Requirements</h3>
              <ul className="trust-list">
                <li>Email <a href="mailto:support@functionfly.com" className="trust-link">support@functionfly.com</a> (copy <a href="mailto:legal@functionfly.com" className="trust-link">legal@functionfly.com</a> if applicable)</li>
                <li>Include your workspace/account identifier and the affected calendar month</li>
                <li>Submit within <strong>30 days</strong> after the affected month</li>
              </ul>
            </Chamber>

            <Chamber className="trust-card">
              <CornerBrace position="bl" />
              <h3>Sole remedy</h3>
              <p className="trust-muted">
                Service Credits are your <strong>sole and exclusive</strong> financial remedy for any failure to meet the Monthly Uptime Percentage. Credits will be applied within two billing cycles.
              </p>
              <p className="trust-muted">
                Maximum credit in a single month: <strong>100%</strong> of subscription fees for Covered Services.
              </p>
            </Chamber>
          </div>
        </div>
      </section>

      {/* Contact */}
      <section className="trust-section">
        <div className="trust-container">
          <h2 className="trust-section-title">Contact</h2>
          <div className="trust-grid-3">
            <Chamber nested className="trust-contact-card">
              <h4>Sales</h4>
              <a href="mailto:sales@functionfly.com" className="trust-link">sales@functionfly.com</a>
            </Chamber>
            <Chamber nested className="trust-contact-card">
              <h4>Legal</h4>
              <a href="mailto:legal@functionfly.com" className="trust-link">legal@functionfly.com</a>
            </Chamber>
            <Chamber nested className="trust-contact-card">
              <h4>Support</h4>
              <a href="mailto:support@functionfly.com" className="trust-link">support@functionfly.com</a>
            </Chamber>
          </div>
          <p className="trust-section-lead" style={{ marginTop: 'var(--space-6)' }}>
            Last updated: March 25, 2026
          </p>
        </div>
      </section>
    </>
  )
}

export default SLAPage