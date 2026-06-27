import React from 'react'
import { AUTH_ORIGIN, DOCS_ORIGIN } from '../config'
import '../styles/sc-main.css';
import { PageGrid } from './containment/PageGrid'
import { Chamber } from './containment/Chamber'
import { CornerBrace } from './containment/CornerBrace'
import { TrustSeal } from './containment/TrustSeal'
import { SealedButton } from './containment/SealedButton'
import { FrameButton } from './containment/FrameButton'
import { Card } from './containment/Card'
import { StatusPill } from './containment/StatusPill'
import { CodeBlock } from './containment/CodeBlock'
import { GaugeStrip } from './containment/GaugeStrip'
import { Gauge, GaugeValue, GaugeLabel } from './containment/Gauge'

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
  _meta: { functionfly: ToolMeta }
}

interface RegistryStats {
  total_functions: number
  verified_functions: number
  total_executions: string
  trust_tiers: number
  runtimes: number
}

interface Category { slug: string; label: string; count: number }

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
  L1: { bg: 'rgba(100, 116, 139, 0.15)', border: 'rgba(100, 116, 139, 0.3)', text: '#94a3b8', label: 'Automated' },
  L2: { bg: 'rgba(59, 130, 246, 0.15)', border: 'rgba(59, 130, 246, 0.3)', text: '#3b82f6', label: 'Security Scanned' },
  L3: { bg: 'rgba(139, 92, 246, 0.15)', border: 'rgba(139, 92, 246, 0.3)', text: '#8b5cf6', label: 'Code Reviewed' },
  L4: { bg: 'rgba(143, 255, 208, 0.15)', border: 'rgba(143, 255, 208, 0.3)', text: '#8fffd0', label: 'Certified' },
}

function formatNumber(num: number): string {
  if (num >= 1000000000) return `${(num / 1000000000).toFixed(1)}B+`
  if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M+`
  if (num >= 1000) return `${(num / 1000).toFixed(1)}K+`
  return num.toString()
}

const Container: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <div style={{ maxWidth: 1180, margin: '0 auto', padding: '0 var(--space-4)' }}>{children}</div>
)
const Section: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <section style={{ padding: 'var(--space-8) 0', background: 'var(--bg)' }}>{children}</section>
)
const SectionTitle: React.FC<{ children: React.ReactNode; lead?: React.ReactNode }> = ({ children, lead }) => (
  <div style={{ textAlign: 'center', marginBottom: 'var(--space-7)' }}>
    <h2 style={{
      fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700,
      letterSpacing: '-0.005em', color: 'var(--text)',
      marginBottom: lead ? 'var(--space-4)' : 0,
    }}>{children}</h2>
    {lead && <p style={{ color: 'var(--text-dim)', maxWidth: 640, margin: '0 auto', lineHeight: 1.6 }}>{lead}</p>}
  </div>
)

const RegistryPage: React.FC<RegistryPageProps> = ({ stats, trendingTools, categories }) => {
  const displayCategories = categories.length > 0 ? categories : defaultCategories
  const totalFunctions = stats?.total_functions ?? 0
  const verifiedFunctions = stats?.verified_functions ?? 0
  const totalExecutions = stats?.total_executions ?? '0'
  const trustTiers = stats?.trust_tiers ?? 4
  const runtimes = stats?.runtimes ?? 5

  return (
    <>
      <PageGrid />

      <Section>
        <Container>
          <Chamber variant="ribs">
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <div style={{
              fontFamily: 'var(--font-mono)', fontSize: 11, letterSpacing: '0.08em',
              textTransform: 'uppercase', color: 'var(--status-ok)',
              marginBottom: 'var(--space-5)',
              display: 'inline-flex', alignItems: 'center', gap: 'var(--space-2)',
            }}>
              <span style={{ width: 8, height: 8, borderRadius: '50%', background: 'var(--status-ok)', boxShadow: '0 0 12px var(--status-ok)' }} />
              Model Context Protocol
            </div>
            <h1 style={{
              fontFamily: 'var(--font-display)', fontSize: '58px', fontWeight: 700,
              letterSpacing: '-0.01em', lineHeight: 1.08, color: 'var(--text)',
              marginBottom: 'var(--space-5)',
            }}>
              The default directory of<br />
              <span style={{ color: 'var(--accent)' }}>AI agent functions</span>
            </h1>
            <p style={{ fontSize: 17, lineHeight: 1.6, color: 'var(--text-dim)', maxWidth: 720, marginBottom: 'var(--space-6)' }}>
              Searchable, trust-scored directory of MCP-compatible functions. Install in one click for Claude Desktop, Cursor, and VS Code.
            </p>
            <div style={{ display: 'flex', gap: 'var(--space-3)', alignItems: 'center', flexWrap: 'wrap', marginBottom: 'var(--space-6)' }}>
              <SealedButton onClick={() => { window.location.href = `${AUTH_ORIGIN}/signup` }}>Get started free</SealedButton>
              <FrameButton onClick={() => { window.location.href = `${DOCS_ORIGIN}/mcp` }}>Read the docs</FrameButton>
              <StatusPill status="live" label={`${formatNumber(totalFunctions)} functions live`} />
            </div>
          </Chamber>
        </Container>
      </Section>

      <Section>
        <Container>
          <Chamber>
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <GaugeStrip>
              <Gauge>
                <GaugeValue>{formatNumber(verifiedFunctions)}</GaugeValue>
                <GaugeLabel>Verified Functions</GaugeLabel>
              </Gauge>
              <Gauge>
                <GaugeValue>{totalExecutions}</GaugeValue>
                <GaugeLabel>Total Executions</GaugeLabel>
              </Gauge>
              <Gauge>
                <GaugeValue>{trustTiers.toString()}</GaugeValue>
                <GaugeLabel>Trust Tiers</GaugeLabel>
              </Gauge>
              <Gauge>
                <GaugeValue>{runtimes.toString()}</GaugeValue>
                <GaugeLabel>Language Runtimes</GaugeLabel>
              </Gauge>
            </GaugeStrip>
          </Chamber>
        </Container>
      </Section>

      <Section>
        <Container>
          <SectionTitle lead="The Function Registry gives AI agents access to verified, trust-scored functions they can call with confidence.">
            Everything you need to build with agents
          </SectionTitle>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: 'var(--space-5)' }}>
            {[
              { n: '01', t: 'Trust Scores', d: 'Every function has a trust score computed from real execution outcomes. Functions that run successfully earn higher scores.', c: 'var(--status-ok)' },
              { n: '02', t: 'Sandbox Execution', d: 'Every function runs in an isolated WebAssembly sandbox. No filesystem access, no network egress unless explicitly granted.', c: 'var(--status-pending)' },
              { n: '03', t: 'Verification Levels', d: 'From automated L1 checks to human attestation at L4, each function\'s trust tier is clearly displayed.', c: 'var(--accent)' },
              { n: '04', t: 'One-Click Install', d: 'Add functions to your MCP server configuration with a single click. Works with Claude Desktop, Cursor, and VS Code.', c: 'var(--foil-b, #d9c4ff)' },
            ].map((s) => (
              <Card key={s.n}>
                <div style={{
                  fontFamily: 'var(--font-mono)', fontSize: 11, fontWeight: 700,
                  letterSpacing: '0.08em', color: 'var(--text-faint)', marginBottom: 'var(--space-3)',
                }}>{s.n}</div>
                <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 600, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>
                  {s.t}
                </h3>
                <p style={{ color: 'var(--text-dim)', fontSize: 14, lineHeight: 1.6 }}>{s.d}</p>
              </Card>
            ))}
          </div>
        </Container>
      </Section>

      {trendingTools.length > 0 && (
        <Section>
          <Container>
            <SectionTitle>Popular this week</SectionTitle>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(260px, 1fr))', gap: 'var(--space-4)' }}>
              {trendingTools.slice(0, 4).map((fn, i) => {
                const tier = fn._meta?.functionfly?.trust_tier ?? 'L1'
                const tierColor = trustTierColors[tier] ?? trustTierColors.L1
                const trustScore = fn._meta?.functionfly?.trust_score ?? 0
                const author = fn._meta?.functionfly?.author ?? 'unknown'
                const fnName = fn._meta?.functionfly?.name ?? fn.name
                const fnTitle = fn.title || fnName
                return (
                  <Card key={fn.name}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 'var(--space-3)', marginBottom: 'var(--space-3)' }}>
                      <div>
                        <div style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--text-faint)', marginBottom: 4 }}>#{i + 1}</div>
                        <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 16, fontWeight: 600, color: 'var(--text)' }}>{fnTitle}</h3>
                      </div>
                      <span style={{
                        fontFamily: 'var(--font-mono)', fontSize: 10, fontWeight: 600,
                        padding: '3px 8px', borderRadius: 'var(--radius-sm)',
                        background: tierColor.bg, border: `1px solid ${tierColor.border}`, color: tierColor.text,
                        textTransform: 'uppercase', letterSpacing: '0.06em',
                      }}>{tier} · {tierColor.label}</span>
                    </div>
                    <div style={{ fontFamily: 'var(--font-mono)', fontSize: 12, color: 'var(--text-dim)', marginBottom: 'var(--space-3)' }}>@{author}</div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
                      <TrustSeal size="small" label={`${(trustScore * 100).toFixed(1)}%`} />
                      <span style={{ fontFamily: 'var(--font-mono)', fontSize: 13, color: 'var(--status-ok)' }}>{(trustScore * 100).toFixed(1)}% trust</span>
                    </div>
                  </Card>
                )
              })}
            </div>
          </Container>
        </Section>
      )}

      <Section>
        <Container>
          <SectionTitle>Verification levels explained</SectionTitle>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: 'var(--space-4)' }}>
            {Object.entries(trustTierColors).map(([tier, colors]) => (
              <Card key={tier}>
                <div style={{
                  display: 'inline-block',
                  fontFamily: 'var(--font-mono)', fontSize: 12, fontWeight: 700,
                  padding: '4px 10px', borderRadius: 'var(--radius-sm)',
                  background: colors.bg, border: `1px solid ${colors.border}`, color: colors.text,
                  marginBottom: 'var(--space-3)',
                }}>{tier}</div>
                <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 600, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>{colors.label}</h3>
                <p style={{ color: 'var(--text-dim)', fontSize: 14, lineHeight: 1.6 }}>
                  {tier === 'L1' && 'Automated format and schema validation. Every function starts here.'}
                  {tier === 'L2' && 'Security scan for risky behaviors and capability constraints.'}
                  {tier === 'L3' && 'Human code review of safety-relevant aspects.'}
                  {tier === 'L4' && 'Platform-verified with signed attestation. Highest trust.'}
                </p>
              </Card>
            ))}
          </div>
        </Container>
      </Section>

      <Section>
        <Container>
          <SectionTitle>Browse by category</SectionTitle>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 'var(--space-3)' }}>
            {displayCategories.map((cat) => (
              <a key={cat.slug} href={`/registry?category=${cat.slug}`} style={{
                display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                padding: 'var(--space-4) var(--space-5)',
                background: 'var(--panel-raised)',
                border: '1px solid var(--panel-edge)',
                borderRadius: 'var(--radius)',
                textDecoration: 'none', color: 'var(--text)',
                transition: 'border-color var(--duration-fast) var(--ease-out)',
              }}>
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: 14 }}>{cat.label}</span>
                <span style={{ color: 'var(--text-faint)' }}>→</span>
              </a>
            ))}
          </div>
        </Container>
      </Section>

      <Section>
        <Container>
          <SectionTitle>Get started in minutes</SectionTitle>
          <Chamber>
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 'var(--space-6)', alignItems: 'start' }}>
              <div>
                <div style={{ display: 'flex', alignItems: 'flex-start', gap: 'var(--space-3)', marginBottom: 'var(--space-4)' }}>
                  <div style={{
                    flexShrink: 0, width: 32, height: 32, display: 'flex', alignItems: 'center', justifyContent: 'center',
                    border: '1px solid var(--accent)', borderRadius: '50%',
                    fontFamily: 'var(--font-mono)', fontSize: 14, fontWeight: 700, color: 'var(--accent)',
                  }}>1</div>
                  <div>
                    <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 17, fontWeight: 600, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>
                      Install the MCP server
                    </h3>
                    <p style={{ color: 'var(--text-dim)', fontSize: 14, marginBottom: 'var(--space-3)' }}>
                      Add FunctionFly to your MCP server configuration:
                    </p>
                  </div>
                </div>
                <CodeBlock language="json">
{`{
  "mcpServers": {
    "functionfly": {
      "command": "npx",
      "args": ["-y", "@functionfly/mcp-server"],
      "env": {
        "FUNCTIONFLY_API_KEY": "ffp_..."
      }
    }
  }
}`}
                </CodeBlock>
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-5)' }}>
                <div style={{ display: 'flex', alignItems: 'flex-start', gap: 'var(--space-3)' }}>
                  <div style={{
                    flexShrink: 0, width: 32, height: 32, display: 'flex', alignItems: 'center', justifyContent: 'center',
                    border: '1px solid var(--accent)', borderRadius: '50%',
                    fontFamily: 'var(--font-mono)', fontSize: 14, fontWeight: 700, color: 'var(--accent)',
                  }}>2</div>
                  <div>
                    <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 17, fontWeight: 600, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>Browse the registry</h3>
                    <p style={{ color: 'var(--text-dim)', fontSize: 14, lineHeight: 1.5 }}>
                      Explore trust-scored functions and add them to your agent configuration with one click.
                    </p>
                  </div>
                </div>
                <div style={{ display: 'flex', alignItems: 'flex-start', gap: 'var(--space-3)' }}>
                  <div style={{
                    flexShrink: 0, width: 32, height: 32, display: 'flex', alignItems: 'center', justifyContent: 'center',
                    border: '1px solid var(--accent)', borderRadius: '50%',
                    fontFamily: 'var(--font-mono)', fontSize: 14, fontWeight: 700, color: 'var(--accent)',
                  }}>3</div>
                  <div>
                    <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 17, fontWeight: 600, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>Start building</h3>
                    <p style={{ color: 'var(--text-dim)', fontSize: 14, lineHeight: 1.5 }}>
                      Your agent can now call verified functions with full trust scores and execution tracking.
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </Chamber>
        </Container>
      </Section>

      <Section>
        <Container>
          <Chamber>
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <div style={{ textAlign: 'center' }}>
              <h2 style={{
                fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700,
                letterSpacing: '-0.005em', color: 'var(--text)',
                marginBottom: 'var(--space-4)',
              }}>
                Ready to build with<br />
                <span style={{ color: 'var(--accent)' }}>trust infrastructure</span>?
              </h2>
              <p style={{ fontSize: 17, color: 'var(--text-dim)', margin: '0 auto var(--space-6)', lineHeight: 1.6 }}>
                Join thousands of developers building reliable AI agent applications.
              </p>
              <div style={{ display: 'inline-flex', gap: 'var(--space-3)', flexWrap: 'wrap', justifyContent: 'center' }}>
                <SealedButton onClick={() => { window.location.href = `${AUTH_ORIGIN}/signup` }}>Create free account</SealedButton>
                <FrameButton onClick={() => { window.location.href = `${DOCS_ORIGIN}/mcp` }}>Read the docs</FrameButton>
              </div>
            </div>
          </Chamber>
        </Container>
      </Section>
    </>
  )
}

export default RegistryPage
