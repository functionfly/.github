import React from 'react'
import { DOCS_ORIGIN, SITE_ORIGIN } from '../config'
import './homepage.css'

const McpPage: React.FC = () => {
  return (
    <div className="ff-homepage">
      <section className="ff-hero-section ff-mcp-hero">
        <div className="ff-hero-inner">
          <div className="ff-hero-eyebrow">
            <span className="ff-pulse-dot" />
            <span>Model Context Protocol</span>
            <span className="ff-live-badge">MCP 2025-03-26</span>
          </div>
          <h1 className="ff-hero-headline">
            Functions that agents<br />can actually find
          </h1>
          <p className="ff-hero-sub">
            MCP is the standard that makes your functions discoverable by Claude, Cursor, Continue, Windsurf, Cline, and every other agent platform. FunctionFly makes publishing MCP-compatible functions effortless.
          </p>
          <div className="ff-hero-actions">
            <a className="ff-btn ff-btn-primary" href="/registry">Browse functions</a>
            <a className="ff-btn ff-btn-secondary" href={`${DOCS_ORIGIN}/mcp`}>Read the docs</a>
            <a className="ff-btn ff-btn-secondary" href="/pricing">Pricing</a>
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <div className="ff-pillars">
            <div className="ff-pillar-card">
              <div className="ff-pillar-icon">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                  <path d="M12 6v6l4 2" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                </svg>
              </div>
              <h2>Zero glue code</h2>
              <p>
                Your function runs on FunctionFly. Flip a toggle. It immediately appears in the MCP Function Registry and becomes callable from any MCP-compatible agent.
              </p>
            </div>

            <div className="ff-pillar-card ff-pillar-card--accent">
              <div className="ff-pillar-icon">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M12 3 4 7v6c0 5 3.5 8.5 8 9 4.5-.5 8-4 8-9V7l-8-4Z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"/>
                  <path d="m9 12 2 2 4-4" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </div>
              <h2>Trust built in</h2>
              <p>
                Every MCP function carries its FunctionFly trust score, verification tier, and capability declarations. Agents can make informed decisions before they call.
              </p>
            </div>

            <div className="ff-pillar-card">
              <div className="ff-pillar-icon">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </div>
              <h2>Universal discovery</h2>
              <p>
                MCP is adopted by Claude, Cursor, Continue, Windsurf, Cline, and dozens more. One publish step reaches every agent platform at once.
              </p>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--alt">
        <div className="ff-container">
          <h2 className="ff-section-title">How it works</h2>
          <div className="ff-steps-grid">
            <div className="ff-step-card">
              <div className="ff-step-num">01</div>
              <h3>Publish your function</h3>
              <p>Deploy your function via the dashboard or CLI. FunctionFly bundles, signs, and verifies it.</p>
            </div>
            <div className="ff-step-card">
              <div className="ff-step-num">02</div>
              <h3>Enable MCP</h3>
              <p>Flip the "MCP-compatible" toggle in the dashboard. Your function gets an MCP manifest at <code>/v1/mcp/manifest</code>.</p>
            </div>
            <div className="ff-step-card">
              <div className="ff-step-num">03</div>
              <h3>Agents discover it</h3>
              <p>Your function appears in the registry and is callable via the FunctionFly MCP server from any MCP client.</p>
            </div>
            <div className="ff-step-card">
              <div className="ff-step-num">04</div>
              <h3>Call with confidence</h3>
              <p>Agents call over streamable-HTTP or stdio. FunctionFly enforces rate limits, logs every invocation, and updates trust scores.</p>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <h2 className="ff-section-title">Works with every agent</h2>
          <div className="ff-compat-grid">
            <div className="ff-compat-card">
              <h3>Claude Desktop</h3>
              <p>Add the FunctionFly MCP server to your Claude Desktop config and start calling functions.</p>
              <a href={`${DOCS_ORIGIN}/mcp/server-setup`}>Setup guide →</a>
            </div>
            <div className="ff-compat-card">
              <h3>Cursor</h3>
              <p>Enable the MCP server in Cursor's settings to unlock FunctionFly functions in your workflow.</p>
              <a href={`${DOCS_ORIGIN}/mcp/server-setup`}>Setup guide →</a>
            </div>
            <div className="ff-compat-card">
              <h3>VS Code</h3>
              <p>Install the FunctionFly MCP extension or configure it manually in VS Code's MCP settings.</p>
              <a href={`${DOCS_ORIGIN}/mcp/server-setup`}>Setup guide →</a>
            </div>
            <div className="ff-compat-card">
              <h3>Continue</h3>
              <p>Add FunctionFly as an MCP server in Continue's config to surface functions in your IDE.</p>
              <a href={`${DOCS_ORIGIN}/mcp/server-setup`}>Setup guide →</a>
            </div>
            <div className="ff-compat-card">
              <h3>Windsurf</h3>
              <p>Configure FunctionFly as an MCP server in Windsurf to access functions directly from the editor.</p>
              <a href={`${DOCS_ORIGIN}/mcp/server-setup`}>Setup guide →</a>
            </div>
            <div className="ff-compat-card">
              <h3>Cline</h3>
              <p>Add FunctionFly MCP server to Cline's configuration for AI-powered function calls.</p>
              <a href={`${DOCS_ORIGIN}/mcp/server-setup`}>Setup guide →</a>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section ff-section--alt">
        <div className="ff-container">
          <h2 className="ff-section-title">Verified MCP badge</h2>
          <p className="ff-section-lead">
            Functions can earn a Verified MCP badge by meeting standards that agents can trust at a glance.
          </p>
          <div className="ff-pillars">
            <div className="ff-pillar-card">
              <h2>Trust score 80+</h2>
              <p>Functions must have a trust score of 80 or higher, based on execution history and verification level.</p>
            </div>
            <div className="ff-pillar-card">
              <h2>Malware scan clean</h2>
              <p>Automated and manual security review ensures no malicious code can reach agent toolchains.</p>
            </div>
            <div className="ff-pillar-card">
              <h2>100+ invocations</h2>
              <p>Real usage history proves the function works correctly in production at scale.</p>
            </div>
            <div className="ff-pillar-card">
              <h2>30+ days old</h2>
              <p>Stability over time gives agents confidence the function is maintained and reliable.</p>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-section">
        <div className="ff-container">
          <h2 className="ff-section-title">Technical details</h2>
          <div className="ff-tech-grid">
            <div className="ff-tech-card">
              <h3>Transport protocols</h3>
              <ul>
                <li><strong>streamable-HTTP</strong> — Default, per MCP spec 2025-03-26</li>
                <li><strong>stdio</strong> — Via <code>@functionfly/mcp-server</code> npm package</li>
              </ul>
            </div>
            <div className="ff-tech-card">
              <h3>Rate limiting</h3>
              <ul>
                <li>1–10,000 calls/min per function</li>
                <li>Default: 60 calls/min</li>
                <li>Configurable per-function in dashboard</li>
              </ul>
            </div>
            <div className="ff-tech-card">
              <h3>Security</h3>
              <ul>
                <li>Bearer auth (JWT session or API key)</li>
                <li>CORS allowlist per function</li>
                <li>Input schema validation</li>
                <li>256 KiB max payload</li>
              </ul>
            </div>
            <div className="ff-tech-card">
              <h3>JSON-RPC 2.0 methods</h3>
              <ul>
                <li><code>initialize</code>, <code>ping</code></li>
                <li><code>tools/list</code>, <code>tools/call</code></li>
                <li><code>resources/list</code>, <code>prompts/list</code></li>
              </ul>
            </div>
          </div>
        </div>
      </section>

      <section className="ff-cta-section">
        <div className="ff-container">
          <h2>Ready to publish?</h2>
          <p>Join the registry and make your functions discoverable by every MCP agent.</p>
          <div className="ff-actions">
            <a className="ff-btn ff-btn-primary" href="/registry">Browse the registry</a>
            <a className="ff-btn ff-btn-secondary" href={`${DOCS_ORIGIN}/mcp/publish-mcp`}>Publish a function</a>
          </div>
        </div>
      </section>
    </div>
  )
}

export default McpPage
