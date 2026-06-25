import React from 'react'
import { SealedButton } from './sc'

const ArrowIcon = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
    <path d="M3 8L13 8M9 4L13 8L9 12" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

const DocsIcon = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
    <path d="M7 1L7 15M1 8L15 8" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
  </svg>
)

interface PricingHeroProps {
  authSignupUrl: string
  docsUrl: string
}

export default function PricingHero({ authSignupUrl, docsUrl }: PricingHeroProps) {
  return (
    <section className="pricing-hero">
      <div className="pricing-hero-inner">
        <div className="pricing-hero-eyebrow">
          <span className="ff-pulse-dot" />
          <span>Infrastructure layer</span>
          <span className="ff-live-badge">LIVE</span>
        </div>
        <h1 className="ff-hero-headline">
          Pricing built for <span className="ff-trust-text">trust</span>
        </h1>
        <p className="pricing-hero-lead">
          Start free. Paid tiers include AI agents with generous call volumes and concurrency. Upgrade as your agent workloads grow.
        </p>
        <div className="ff-hero-actions">
          <a href={authSignupUrl}>
            <SealedButton size="lg" iconRight={<ArrowIcon />}>Get started</SealedButton>
          </a>
          <a href={docsUrl} target="_blank" rel="noopener noreferrer">
            <SealedButton size="lg" iconLeft={<DocsIcon />}>Read the docs</SealedButton>
          </a>
        </div>
      </div>
    </section>
  )
}
