import React from 'react'
import { API_ORIGIN, DOCS_ORIGIN, AUTH_ORIGIN } from '../config'
import './homepage.css'

interface ToolMeta {
  author: string
  name: string
  version: string
  trust_score: number
  trust_tier: string
  verified_mcp: boolean
  homepage: string
  tags: string[]
  runtime: string
}

interface ToolItem {
  name: string
  title: string
  description: string
  inputSchema: unknown
  annotations: { category?: string; readOnlyHint?: boolean; openWorldHint?: boolean }
  _meta: {
    functionfly: ToolMeta
  }
}

interface RegistryStats {
  total_functions: number
  verified_functions: number
  total_executions: string
  trust_tiers: number
  runtimes: number
}

interface Category {
  slug: string
  label: string
  count: number
}

interface RegistryPageProps {
  stats: RegistryStats | null
  trendingTools: ToolItem[]
  categories: Category[]
}

const defaultCategories: Category[] = [
  { slug: 'document-processing', label: 'Documents', count: 0 },
  { slug: 'data-extraction', label: 'Data Extraction', count: 0 },
  { slug: 'ai', label: 'AI & ML', count: 0 },
  { slug: 'communication', label: 'Communication', count: 0 },
  { slug: 'finance', label: 'Finance', count: 0 },
  { slug: 'developer-tools', label: 'Developer Tools', count: 0 },
]

const trustTierColors: Record<string, { bg: string; border: string; text: string; label: string }> = {
  L1: { bg: 'rgba(100, 116, 139, 0.15)', border: 'rgba(100, 116, 139, 0.3)', text: '#64748b', label: 'Automated' },
  L2: { bg: 'rgba(59, 130, 246, 0.15)', border: 'rgba(59, 130, 246, 0.3)', text: '#3b82f6', label: 'Security Scanned' },
  L3: { bg: 'rgba(139, 92, 246, 0.15)', border: 'rgba(139, 92, 246, 0.3)', text: '#8b5cf6', label: 'Code Reviewed' },
  L4: { bg: 'rgba(16, 185, 129, 0.15)', border: 'rgba(16, 185, 129, 0.3)', text: '#10b981', label: 'Certified' },
}

function formatNumber(num: number): string {
  if (num >= 1000000000) return `${(num / 1000000000).toFixed(1)}B+`
  if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M+`
  if (num >= 1000) return `${(num / 1000).toFixed(1)}K+`
  return num.toString()
}

const RegistryPage: React.FC<RegistryPageProps> = ({ stats, trendingTools, categories }) => {
  const displayCategories = categories.length > 0 ? categories : defaultCategories
  const totalFunctions = stats?.total_functions ?? 0
  const verifiedFunctions = stats?.verified_functions ?? 0
  const totalExecutions = stats?.total_executions ?? '0'
  const trustTiers = stats?.trust_tiers ?? 4
  const runtimes = stats?.runtimes ?? 5

  return (
    <div className="ff-homepage">
      <section className="ff-hero-section ff-registry-hero">
        <div className="ff-hero-bg-art" aria-hidden="true">
          <div className="ff-bg-grid" />
          <div className="ff-bg-glow ff-bg-glow--1" />
          <div className="ff-bg-glow ff-bg-glow--2" />
        </div>
        <div className="ff-hero-inner">
          <div className="ff-hero-eyebrow">
            <span className="ff-pulse-dot" />
            <span>Model Context Protocol</span>
          </div>
          <h1 className="ff-hero-headline">
            The default directory of<br />
            <span className="ff-hero-accent">AI agent functions</span>
          </h1>
          <p className="ff-hero-sub">
            Searchable, trust-scored directory of MCP-compatible functions. Install in one click for Claude Desktop, Cursor, and VS Code.
          </p>

          <div className="ff-registry-search">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <circle cx="11" cy="11" r="8"/>
              <path d="M21 21l-4.35-4.35"/>
            </svg>
            <input type="text" placeholder="Search functions..." className="ff-registry-search-input" />
            <div className="ff-registry-search-badge">{formatNumber(totalFunctions)} functions</div>
          </div>
        </div>
      </section>

      <div className="ff-registry-stats">
        <div className="ff-container">
          <div className="ff-stats-row">
            <div className="ff-stat-item">
              <span className="ff-stat-value">{formatNumber(verifiedFunctions)}</span>
              <span className="ff-stat-label">Verified Functions</span>
            </div>
            <div className="ff-stat-divider" />
            <div className="ff-stat-item">
              <span className="ff-stat-value">{totalExecutions}</span>
              <span className="ff-stat-label">Total Executions</span>
            </div>
            <div className="ff-stat-divider" />
            <div className="ff-stat-item">
              <span className="ff-stat-value">{trustTiers}</span>
              <span className="ff-stat-label">Trust Tiers</span>
            </div>
            <div className="ff-stat-divider" />
            <div className="ff-stat-item">
              <span className="ff-stat-value">{runtimes}</span>
              <span className="ff-stat-label">Language Runtimes</span>
            </div>
          </div>
        </div>
      </div>

      <section className="ff-section">
        <div className="ff-container">
          <div className="ff-about-header">
            <div className="ff-section-tag">
              <span className="ff-tag-rule" />
              <span>Features</span>
            </div>
            <h2 className="ff-section-title">Everything you need to build with agents</h2>
            <p className="ff-section-desc">
              The Function Registry gives AI agents access to verified, trust-scored functions they can call with confidence.
            </p>
          </div>

          <div className="ff-about-steps">
            <div className="ff-about-step">
              <div className="ff-about-step-num">01</div>
              <div className="ff-about-step-body">
                <div className="ff-about-step-icon">
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round"><path d="M9 12l2 2 4-4"/><path d="M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z"/></svg>
                </div>
                <h3>Trust Scores</h3>
                <p>
                  Every function has a trust score computed from real execution outcomes. Functions that run successfully earn higher scores.
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
                  Every function runs in an isolated WebAssembly sandbox. No filesystem access, no network egress unless explicitly granted.
                </p>
              </div>
            </div>

            <div className="ff-about-step">
              <div className="ff-about-step-num">03</div>
              <div className="ff-about-step-body">
                <div className="ff-about-step-icon ff-about-step-icon--amber">
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round"><path d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/></svg>
                </div>
                <h3>Verification Levels</h3>
                <p>
                  From automated L1 checks to human attestation at L4, each function's trust tier is clearly displayed.
                </p>
              </div>
            </div>

            <div className="ff-about-step">
              <div className="ff-about-step-num">04</div>
              <div className="ff-about-step-body">
                <div className="ff-about-step-icon ff-about-step-icon--purple">
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round"><path d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"/></svg>
                </div>
                <h3>One-Click Install</h3>
                <p>
                  Add functions to your MCP server configuration with a single click. Works with Claude Desktop, Cursor, and VS Code.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {trendingTools.length > 0 && (
        <section className="ff-section ff-section--alt">
          <div className="ff-container">
            <div className="ff-about-header">
              <div className="ff-section-tag">
                <span className="ff-tag-rule" />
                <span>Trending</span>
              </div>
              <h2 className="ff-section-title">Popular this week</h2>
            </div>

            <div className="ff-trending-grid">
              {trendingTools.slice(0, 4).map((fn, i) => {
                const tier = fn._meta?.functionfly?.trust_tier ?? 'L1'
                const tierColor = trustTierColors[tier] ?? trustTierColors.L1
                const trustScore = fn._meta?.functionfly?.trust_score ?? 0
                const author = fn._meta?.functionfly?.author ?? 'unknown'
                const fnName = fn._meta?.functionfly?.name ?? fn.name
                const fnTitle = fn.title || fnName

                return (
                  <a key={fn.name} href={`/@${author}/v1/fx/${fnName}`} className="ff-trending-card">
                    <div className="ff-trending-rank">#{i + 1}</div>
                    <div className="ff-trending-content">
                      <div className="ff-trending-header">
                        <h3 className="ff-trending-title">{fnTitle}</h3>
                        <span
                          className="ff-trending-tier"
                          style={{
                            background: tierColor.bg,
                            borderColor: tierColor.border,
                            color: tierColor.text,
                          }}
                        >
                          {tier} · {tierColor.label}
                        </span>
                      </div>
                      <p className="ff-trending-author">@{author}</p>
                      <div className="ff-trending-stats">
                        <span className="ff-trending-score">
                          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                            <path d="M9 12l2 2 4-4"/>
                          </svg>
                          {(trustScore * 100).toFixed(1)}%
                        </span>
                      </div>
                    </div>
                  </a>
                )
              })}
            </div>
          </div>
        </section>
      )}

      <section className="ff-section">
        <div className="ff-container">
          <div className="ff-about-header">
            <div className="ff-section-tag">
              <span className="ff-tag-rule" />
              <span>Trust Tiers</span>
            </div>
            <h2 className="ff-section-title">Verification levels explained</h2>
          </div>

          <div className="ff-tiers-grid">
            {Object.entries(trustTierColors).map(([tier, colors]) => (
              <div key={tier} className="ff-tier-card">
                <div
                  className="ff-tier-badge"
                  style={{ background: colors.bg, borderColor: colors.border, color: colors.text }}
                >
                  {tier}
                </div>
                <h3>{colors.label}</h3>
                <p>
                  {tier === 'L1' && 'Automated format and schema validation. Every function starts here.'}
                  {tier === 'L2' && 'Security scan for risky behaviors and capability constraints.'}
                  {tier === 'L3' && 'Human code review of safety-relevant aspects.'}
                  {tier === 'L4' && 'Platform-verified with signed attestation. Highest trust.'}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--alt">
        <div className="ff-container">
          <div className="ff-about-header">
            <div className="ff-section-tag">
              <span className="ff-tag-rule" />
              <span>Categories</span>
            </div>
            <h2 className="ff-section-title">Browse by category</h2>
          </div>

          <div className="ff-categories-grid">
            {displayCategories.map((cat) => (
              <a key={cat.slug} href={`/registry?category=${cat.slug}`} className="ff-category-card">
                <span className="ff-category-name">{cat.label}</span>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M5 12h14M12 5l7 7-7 7"/>
                </svg>
              </a>
            ))}
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <div className="ff-about-header">
            <div className="ff-section-tag">
              <span className="ff-tag-rule" />
              <span>Quick Start</span>
            </div>
            <h2 className="ff-section-title">Get started in minutes</h2>
          </div>

          <div className="ff-install-steps">
            <div className="ff-install-step">
              <div className="ff-install-step-num">1</div>
              <div className="ff-install-step-content">
                <h3>Install the MCP server</h3>
                <p>Add FunctionFly to your MCP server configuration:</p>
                <div className="ff-code-block">
                  <div className="ff-code-header">
                    <span className="ff-code-lang">JSON</span>
                    <button className="ff-code-copy">Copy</button>
                  </div>
                  <pre className="ff-install-code">{`{
  "mcpServers": {
    "functionfly": {
      "command": "npx",
      "args": ["-y", "@functionfly/mcp-server"],
      "env": {
        "FUNCTIONFLY_API_KEY": "ffp_..."
      }
    }
  }
}`}</pre>
                </div>
              </div>
            </div>

            <div className="ff-install-step">
              <div className="ff-install-step-num">2</div>
              <div className="ff-install-step-content">
                <h3>Browse the registry</h3>
                <p>Explore trust-scored functions and add them to your agent configuration with one click.</p>
              </div>
            </div>

            <div className="ff-install-step">
              <div className="ff-install-step-num">3</div>
              <div className="ff-install-step-content">
                <h3>Start building</h3>
                <p>Your agent can now call verified functions with full trust scores and execution tracking.</p>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--cta">
        <div className="ff-container">
          <div className="ff-cta">
            <h2 className="ff-cta-title">
              Ready to build with<br />
              <span className="ff-cta-accent">trust infrastructure</span>?
            </h2>
            <p className="ff-cta-desc">
              Join thousands of developers building reliable AI agent applications.
            </p>
            <div className="ff-cta-actions">
              <a className="ff-btn ff-btn-primary" href={`${AUTH_ORIGIN}/signup`}>
                Create free account
              </a>
              <a className="ff-btn ff-btn-outline" href={`${DOCS_ORIGIN}/mcp`}>
                Read the docs
              </a>
            </div>
          </div>
        </div>
      </section>
    </div>
  )
}

export default RegistryPage