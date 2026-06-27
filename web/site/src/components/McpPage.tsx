import React from 'react'
import { DOCS_ORIGIN } from '../config'
import {
  PageGrid,
  Chamber,
  CornerBrace,
  SealedButton,
  FrameButton,
  Card,
} from './sc'
import '../styles/sc-main.css';

const Container: React.FC<{ children: React.ReactNode; narrow?: boolean }> = ({ children, narrow }) => (
  <div style={{ maxWidth: narrow ? 720 : 1100, margin: '0 auto', padding: '0 var(--space-4)' }}>
    {children}
  </div>
)
const Section: React.FC<{ children: React.ReactNode; alt?: boolean }> = ({ children, alt }) => (
  <section style={{ padding: 'var(--space-8) 0', background: 'var(--bg)' }}>{children}</section>
)

const CircleIcon: React.FC<{ size?: number }> = ({ size = 28 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
    <path d="M12 6v6l4 2" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
  </svg>
)

const ShieldCheckIcon: React.FC<{ size?: number }> = ({ size = 28 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M12 3 4 7v6c0 5 3.5 8.5 8 9 4.5-.5 8-4 8-9V7l-8-4Z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
    <path d="m9 12 2 2 4-4" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
)

const GlobeIcon: React.FC<{ size?: number }> = ({ size = 28 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
)

const CheckIcon: React.FC<{ size?: number }> = ({ size = 20 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <polyline points="20 6 9 17 4 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
)

const agents = [
  { name: 'Claude Desktop', desc: 'Add the FunctionFly MCP server to your Claude Desktop config and start calling functions.', href: `${DOCS_ORIGIN}/mcp/server-setup` },
  { name: 'Cursor', desc: 'Enable the MCP server in Cursor\'s settings to unlock FunctionFly functions in your workflow.', href: `${DOCS_ORIGIN}/mcp/server-setup` },
  { name: 'VS Code', desc: 'Install the FunctionFly MCP extension or configure it manually in VS Code\'s MCP settings.', href: `${DOCS_ORIGIN}/mcp/server-setup` },
  { name: 'Continue', desc: 'Add FunctionFly as an MCP server in Continue\'s config to surface functions in your IDE.', href: `${DOCS_ORIGIN}/mcp/server-setup` },
  { name: 'Windsurf', desc: 'Configure FunctionFly as an MCP server in Windsurf to access functions directly from the editor.', href: `${DOCS_ORIGIN}/mcp/server-setup` },
  { name: 'Cline', desc: 'Add FunctionFly MCP server to Cline\'s configuration for AI-powered function calls.', href: `${DOCS_ORIGIN}/mcp/server-setup` },
]

const badges = [
  { title: 'Trust score 80+', desc: 'Functions must have a trust score of 80 or higher, based on execution history and verification level.' },
  { title: 'Malware scan clean', desc: 'Automated and manual security review ensures no malicious code can reach agent toolchains.' },
  { title: '100+ invocations', desc: 'Real usage history proves the function works correctly in production at scale.' },
  { title: '30+ days old', desc: 'Stability over time gives agents confidence the function is maintained and reliable.' },
]

const techDetails = [
  { title: 'Transport protocols', items: ['streamable-HTTP — Default, per MCP spec 2025-03-26', 'stdio — Via @functionfly/mcp-server npm package'] },
  { title: 'Rate limiting', items: ['1–10,000 calls/min per function', 'Default: 60 calls/min', 'Configurable per-function in dashboard'] },
  { title: 'Security', items: ['Bearer auth (JWT session or API key)', 'CORS allowlist per function', 'Input schema validation', '256 KiB max payload'] },
  { title: 'JSON-RPC 2.0 methods', items: ['initialize, ping', 'tools/list, tools/all', 'resources/list, prompts/list'] },
]

const McpPage: React.FC = () => {
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
              display: 'inline-flex', alignItems: 'center', gap: 'var(--space-3)',
            }}>
              <span style={{ width: 8, height: 8, borderRadius: '50%', background: 'var(--status-ok)', boxShadow: '0 0 12px var(--status-ok)' }} />
              Model Context Protocol
              <span style={{
                fontFamily: 'var(--font-mono)', fontSize: 10, fontWeight: 500,
                padding: '2px 8px', borderRadius: 'var(--radius-sm)',
                background: 'rgba(255, 122, 61, 0.15)',
                border: '1px solid var(--accent-dim)',
                color: 'var(--accent)',
              }}>
                MCP 2025-03-26
              </span>
            </div>
            <h1 style={{
              fontFamily: 'var(--font-display)', fontSize: '58px', fontWeight: 700,
              letterSpacing: '-0.01em', lineHeight: 1.08, color: 'var(--text)',
              marginBottom: 'var(--space-5)',
            }}>
              Functions that agents<br />can actually find
            </h1>
            <p style={{ fontSize: 17, lineHeight: 1.6, color: 'var(--text-dim)', maxWidth: 720, marginBottom: 'var(--space-6)' }}>
              MCP is the standard that makes your functions discoverable by Claude, Cursor, Continue, Windsurf, Cline, and every other agent platform. FunctionFly makes publishing MCP-compatible functions effortless.
            </p>
            <div style={{ display: 'flex', gap: 'var(--space-3)', flexWrap: 'wrap' }}>
              <SealedButton onClick={() => window.location.href = '/registry'}>
                Browse functions
              </SealedButton>
              <FrameButton onClick={() => window.location.href = `${DOCS_ORIGIN}/mcp`}>
                Read the docs
              </FrameButton>
              <FrameButton onClick={() => window.location.href = '/pricing'}>
                Pricing
              </FrameButton>
            </div>
          </Chamber>
        </Container>
      </Section>

      <Section>
        <Container>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: 'var(--space-5)' }}>
            <Card>
              <div style={{ color: 'var(--accent)', marginBottom: 'var(--space-4)' }}>
                <CircleIcon size={32} />
              </div>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                Zero glue code
              </h3>
              <p style={{ color: 'var(--text-dim)', fontSize: 14, lineHeight: 1.6 }}>
                Your function runs on FunctionFly. Flip a toggle. It immediately appears in the MCP Function Registry and becomes callable from any MCP-compatible agent.
              </p>
            </Card>

            <Card>
              <div style={{ color: 'var(--accent)', marginBottom: 'var(--space-4)' }}>
                <ShieldCheckIcon size={32} />
              </div>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                Trust built in
              </h3>
              <p style={{ color: 'var(--text-dim)', fontSize: 14, lineHeight: 1.6 }}>
                Every MCP function carries its FunctionFly trust score, verification tier, and capability declarations. Agents can make informed decisions before they call.
              </p>
            </Card>

            <Card>
              <div style={{ color: 'var(--accent)', marginBottom: 'var(--space-4)' }}>
                <GlobeIcon size={32} />
              </div>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                Universal discovery
              </h3>
              <p style={{ color: 'var(--text-dim)', fontSize: 14, lineHeight: 1.6 }}>
                MCP is adopted by Claude, Cursor, Continue, Windsurf, Cline, and dozens more. One publish step reaches every agent platform at once.
              </p>
            </Card>
          </div>
        </Container>
      </Section>

      <Section alt>
        <Container>
          <h2 style={{
            fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700,
            letterSpacing: '-0.005em', color: 'var(--text)',
            marginBottom: 'var(--space-7)', textAlign: 'center',
          }}>
            How it works
          </h2>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: 'var(--space-5)' }}>
            {[
              { num: '01', title: 'Publish your function', desc: 'Deploy your function via the dashboard or CLI. FunctionFly bundles, signs, and verifies it.' },
              { num: '02', title: 'Enable MCP', desc: 'Flip the "MCP-compatible" toggle in the dashboard. Your function gets an MCP manifest at /v1/mcp/manifest.' },
              { num: '03', title: 'Agents discover it', desc: 'Your function appears in the registry and is callable via the FunctionFly MCP server from any MCP client.' },
              { num: '04', title: 'Call with confidence', desc: 'Agents call over streamable-HTTP or stdio. FunctionFly enforces rate limits, logs every invocation, and updates trust scores.' },
            ].map((step) => (
              <Card key={step.num}>
                <div style={{ fontFamily: 'var(--font-mono)', fontSize: 28, fontWeight: 500, color: 'var(--accent)', marginBottom: 'var(--space-3)' }}>
                  {step.num}
                </div>
                <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 15, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>
                  {step.title}
                </h3>
                <p style={{ fontSize: 13, color: 'var(--text-dim)', lineHeight: 1.6, margin: 0 }}>
                  {step.desc}
                </p>
              </Card>
            ))}
          </div>
        </Container>
      </Section>

      <Section>
        <Container>
          <h2 style={{
            fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700,
            letterSpacing: '-0.005em', color: 'var(--text)',
            marginBottom: 'var(--space-7)', textAlign: 'center',
          }}>
            Works with every agent
          </h2>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: 'var(--space-4)' }}>
            {agents.map((agent) => (
              <Card key={agent.name}>
                <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 16, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>
                  {agent.name}
                </h3>
                <p style={{ fontSize: 13, color: 'var(--text-dim)', lineHeight: 1.6, marginBottom: 'var(--space-3)' }}>
                  {agent.desc}
                </p>
                <a href={agent.href} style={{ fontSize: 13, color: 'var(--accent)', textDecoration: 'none' }}>
                  Setup guide →
                </a>
              </Card>
            ))}
          </div>
        </Container>
      </Section>

      <Section alt>
        <Container>
          <h2 style={{
            fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700,
            letterSpacing: '-0.005em', color: 'var(--text)',
            marginBottom: 'var(--space-3)', textAlign: 'center',
          }}>
            Verified MCP badge
          </h2>
          <p style={{ color: 'var(--text-dim)', textAlign: 'center', marginBottom: 'var(--space-6)' }}>
            Functions can earn a Verified MCP badge by meeting standards that agents can trust at a glance.
          </p>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 'var(--space-4)' }}>
            {badges.map((badge) => (
              <Card key={badge.title}>
                <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 15, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>
                  {badge.title}
                </h3>
                <p style={{ fontSize: 13, color: 'var(--text-dim)', lineHeight: 1.6, margin: 0 }}>
                  {badge.desc}
                </p>
              </Card>
            ))}
          </div>
        </Container>
      </Section>

      <Section>
        <Container>
          <h2 style={{
            fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700,
            letterSpacing: '-0.005em', color: 'var(--text)',
            marginBottom: 'var(--space-6)', textAlign: 'center',
          }}>
            Technical details
          </h2>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 'var(--space-4)' }}>
            {techDetails.map((tech) => (
              <Card key={tech.title}>
                <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 14, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-3)' }}>
                  {tech.title}
                </h3>
                <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
                  {tech.items.map((item) => (
                    <li key={item} style={{ fontSize: 12, color: 'var(--text-dim)', marginBottom: 'var(--space-1)', lineHeight: 1.5 }}>
                      → {item}
                    </li>
                  ))}
                </ul>
              </Card>
            ))}
          </div>
        </Container>
      </Section>

      <Section>
        <Container narrow>
          <Chamber variant="ribs">
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <div style={{ textAlign: 'center' }}>
              <h2 style={{
                fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700,
                letterSpacing: '-0.005em', color: 'var(--text)',
                marginBottom: 'var(--space-4)',
              }}>
                Ready to publish?
              </h2>
              <p style={{ color: 'var(--text-dim)', lineHeight: 1.6, marginBottom: 'var(--space-6)' }}>
                Join the registry and make your functions discoverable by every MCP agent.
              </p>
              <div style={{ display: 'flex', gap: 'var(--space-3)', justifyContent: 'center', flexWrap: 'wrap' }}>
                <SealedButton onClick={() => window.location.href = '/registry'}>
                  Browse the registry
                </SealedButton>
                <FrameButton onClick={() => window.location.href = `${DOCS_ORIGIN}/mcp/publish-mcp`}>
                  Publish a function
                </FrameButton>
              </div>
            </div>
          </Chamber>
        </Container>
      </Section>
    </>
  )
}

export default McpPage
