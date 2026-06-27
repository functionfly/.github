import React from 'react'
import {
  PageGrid,
  Chamber,
  CornerBrace,
  StatusPill,
} from './sc'
import '../styles/sc-main.css';

const CookiePolicyPage: React.FC = () => {
  return (
    <>
      <PageGrid />
      {/* Hero Header */}
      <section className="trust-hero">
        <Chamber ribs className="trust-hero-chamber">
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <div className="trust-hero-content">
            <div className="trust-hero-eyebrow">
              <StatusPill status="live" label="Updated June 2026" />
            </div>
            <h1 className="trust-hero-title">
              Cookie<br />Policy
            </h1>
            <p className="trust-hero-subtitle">
              This Cookie Policy explains how FunctionFly™ LLC ("FunctionFly," "we," or "us") uses cookies and similar tracking technologies when you visit our website or use our platform services.
            </p>
          </div>
        </Chamber>
      </section>

      {/* Table of Contents */}
      <section className="trust-section">
        <div className="trust-container">
          <Chamber>
            <nav className="dpa-toc" aria-label="Page sections">
              <a href="#what-are-cookies">What Are Cookies</a>
              <a href="#how-we-use">How We Use Cookies</a>
              <a href="#types">Types of Cookies</a>
              <a href="#third-party">Third-Party Cookies</a>
              <a href="#manage">Managing Cookies</a>
              <a href="#updates">Policy Updates</a>
              <a href="#contact">Contact</a>
            </nav>
          </Chamber>
        </div>
      </section>

      {/* Section 1: What Are Cookies */}
      <section className="trust-section" id="what-are-cookies">
        <div className="trust-container">
          <Chamber>
            <CornerBrace position="tr" />
            <h2>1. What Are Cookies</h2>
            <p>
              Cookies are small text files that are stored on your device (computer, tablet, or mobile phone) when you visit a website. They are widely used to make websites work more efficiently, provide a better user experience, and give website owners useful information.
            </p>
            <p>
              Similar technologies include web beacons, pixel tags, local storage, and other tracking technologies. When we refer to "cookies" in this policy, we mean all these technologies collectively.
            </p>
          </Chamber>
        </div>
      </section>

      {/* Section 2: How We Use Cookies */}
      <section className="trust-section" id="how-we-use">
        <div className="trust-container">
          <Chamber>
            <CornerBrace position="bl" />
            <h2>2. How We Use Cookies</h2>
            <p>We use cookies for several purposes:</p>
            <ul>
              <li><strong>Authentication:</strong> To recognize you when you sign in and keep you logged in during your session</li>
              <li><strong>Security:</strong> To detect unauthorized access and protect your account</li>
              <li><strong>Analytics:</strong> To understand how visitors use our website and platform</li>
              <li><strong>Performance:</strong> To monitor and improve website loading times and functionality</li>
              <li><strong>Preferences:</strong> To remember your settings and preferences (such as language and theme)</li>
              <li><strong>Marketing:</strong> To deliver relevant advertisements (where applicable)</li>
            </ul>
          </Chamber>
        </div>
      </section>

      {/* Section 3: Types of Cookies */}
      <section className="trust-section" id="types">
        <div className="trust-container">
          <Chamber>
            <CornerBrace position="tr" />
            <h2>3. Types of Cookies We Use</h2>

            <h3>3.1 Essential Cookies</h3>
            <p>
              These cookies are necessary for the website to function properly. They enable core functionality such as security, account access, session management, and language preferences. You cannot opt out of essential cookies as the service would not work without them.
            </p>
            <table className="cookie-table">
              <thead>
                <tr>
                  <th>Cookie Name</th>
                  <th>Purpose</th>
                  <th>Duration</th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td><code>session_id</code></td>
                  <td>Maintains your authenticated session</td>
                  <td>Session</td>
                </tr>
                <tr>
                  <td><code>csrf_token</code></td>
                  <td>Cross-Site Request Forgery protection</td>
                  <td>Session</td>
                </tr>
                <tr>
                  <td><code>locale</code></td>
                  <td>Stores your language preference</td>
                  <td>1 year</td>
                </tr>
              </tbody>
            </table>

            <h3>3.2 Analytics Cookies</h3>
            <p>
              These cookies help us understand how visitors interact with our website by collecting and reporting information anonymously. This helps us improve our website and services.
            </p>
            <table className="cookie-table">
              <thead>
                <tr>
                  <th>Cookie Name</th>
                  <th>Purpose</th>
                  <th>Duration</th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td><code>_ga</code></td>
                  <td>Google Analytics - distinguishes users</td>
                  <td>2 years</td>
                </tr>
                <tr>
                  <td><code>_gid</code></td>
                  <td>Google Analytics - distinguishes users</td>
                  <td>24 hours</td>
                </tr>
                <tr>
                  <td><code>mp_*</code></td>
                  <td>Mixpanel - product analytics</td>
                  <td>1 year</td>
                </tr>
              </tbody>
            </table>

            <h3>3.3 Functional Cookies</h3>
            <p>
              These cookies enable enhanced functionality and personalization, such as remembering your preferences and settings.
            </p>
            <table className="cookie-table">
              <thead>
                <tr>
                  <th>Cookie Name</th>
                  <th>Purpose</th>
                  <th>Duration</th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td><code>theme</code></td>
                  <td>Stores your light/dark theme preference</td>
                  <td>1 year</td>
                </tr>
                <tr>
                  <td><code>recent_functions</code></td>
                  <td>Remembers recently viewed functions</td>
                  <td>30 days</td>
                </tr>
              </tbody>
            </table>
          </Chamber>
        </div>
      </section>

      {/* Section 4: Third-Party Cookies */}
      <section className="trust-section" id="third-party">
        <div className="trust-container">
          <Chamber>
            <CornerBrace position="bl" />
            <h2>4. Third-Party Cookies</h2>
            <p>Some cookies are placed by third-party services that appear on our pages. We use the following third-party services:</p>

            <h3>4.1 Google Analytics</h3>
            <p>
              We use Google Analytics to understand how visitors use our website. Google Analytics collects information about page visits, time spent on pages, and navigation paths. This data is aggregated and anonymous. You can opt out of Google Analytics by installing the <a href="https://tools.google.com/dlpage/gaoptout" target="_blank" rel="noopener">Google Analytics Opt-out Browser Add-on</a>.
            </p>

            <h3>4.2 Mixpanel</h3>
            <p>
              We use Mixpanel for product analytics to understand how users interact with our platform. This helps us improve user experience. You can opt out of Mixpanel tracking by contacting us at <a href="mailto:privacy@functionfly.com">privacy@functionfly.com</a>.
            </p>

            <h3>4.3 Cloudflare</h3>
            <p>
              Cloudflare is used for security and performance optimization. Cloudflare cookies help detect malicious traffic and improve page load times. For more information, see <a href="https://www.cloudflare.com/cookie-policy/" target="_blank" rel="noopener">Cloudflare's Cookie Policy</a>.
            </p>

            <h3>4.4 Customer Support Chat</h3>
            <p>
              If we offer live chat support, that service may place cookies to maintain your session and remember conversation context.
            </p>
          </Chamber>
        </div>
      </section>

      {/* Section 5: Managing Cookies */}
      <section className="trust-section" id="manage">
        <div className="trust-container">
          <Chamber>
            <CornerBrace position="tr" />
            <h2>5. Managing Your Cookie Preferences</h2>

            <h3>5.1 Browser Controls</h3>
            <p>
              Most web browsers allow you to control cookies through their settings. You can:
            </p>
            <ul>
              <li>View what cookies are stored on your device</li>
              <li>Delete all or specific cookies</li>
              <li>Block cookies from all or certain websites</li>
              <li>Block third-party cookies</li>
              <li>Clear all cookies when you close your browser</li>
            </ul>
            <p>
              Instructions for managing cookies in popular browsers:
            </p>
            <ul>
              <li><a href="https://support.google.com/chrome/answer/95647" target="_blank" rel="noopener">Google Chrome</a></li>
              <li><a href="https://support.mozilla.org/en-US/kb/cookies-information-websites-store-on-your-computer" target="_blank" rel="noopener">Mozilla Firefox</a></li>
              <li><a href="https://support.apple.com/guide/safari/manage-cookies-sfri11471/" target="_blank" rel="noopener">Apple Safari</a></li>
              <li><a href="https://support.microsoft.com/en-us/microsoft-edge/delete-cookies-in-microsoft-edge-63947406-40ac-c3b8-57b9-2a946a29ae09" target="_blank" rel="noopener">Microsoft Edge</a></li>
            </ul>

            <h3>5.2 Cookie Consent Banner</h3>
            <p>
              When you first visit our website, you will see a cookie consent banner that allows you to accept or decline non-essential cookies. Your preferences will be remembered for future visits.
            </p>

            <h3>5.3 Mobile Devices</h3>
            <p>
              On mobile devices, you can manage cookies through your device's settings or browser settings. Some browsers offer additional controls such as limiting cookies to HTTPS only.
            </p>

            <h3>5.4 Impact of Disabling Cookies</h3>
            <p>
              If you choose to disable cookies, please note that some features of our website may not work properly. Essential cookies cannot be disabled as they are required for the service to function.
            </p>
          </Chamber>
        </div>
      </section>

      {/* Section 6: Policy Updates */}
      <section className="trust-section" id="updates">
        <div className="trust-container">
          <Chamber>
            <CornerBrace position="bl" />
            <h2>6. Updates to This Policy</h2>
            <p>
              We may update this Cookie Policy from time to time to reflect changes in our practices, technology, or legal requirements. When we make significant changes, we will notify you by:
            </p>
            <ul>
              <li>Posting a prominent notice on our website</li>
              <li>Updating the "Last Updated" date at the top of this policy</li>
              <li>Sending an email notification to registered users</li>
            </ul>
            <p>
              We encourage you to review this policy periodically to stay informed about our use of cookies.
            </p>
          </Chamber>
        </div>
      </section>

      {/* Section 7: Contact */}
      <section className="trust-section" id="contact">
        <div className="trust-container">
          <Chamber>
            <CornerBrace position="tr" />
            <h2>7. Contact Us</h2>
            <p>
              If you have questions about this Cookie Policy or our use of cookies, please contact us:
            </p>
            <ul>
              <li><strong>Email:</strong> <a href="mailto:privacy@functionfly.com">privacy@functionfly.com</a></li>
              <li><strong>Address:</strong> FunctionFly LLC, Privacy Team</li>
            </ul>
            <p>
              For more information about how we protect your privacy, please see our <a href="/privacy">Privacy Policy</a>.
            </p>
          </Chamber>
        </div>
      </section>

      {/* Related Links */}
      <section className="trust-section">
        <div className="trust-container">
          <Chamber nested>
            <h3>Related pages</h3>
            <div className="dpa-related-links">
              <a href="/privacy">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
                </svg>
                Privacy Policy
              </a>
              <a href="/terms">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                  <path d="M14 2v6h6"></path>
                </svg>
                Terms of Service
              </a>
              <a href="/dpa">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                  <path d="M14 2v6h6"></path>
                  <path d="M9 15l2 2 4-4"></path>
                </svg>
                Data Processing Agreement
              </a>
              <a href="/security">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                  <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
                </svg>
                Security Policy
              </a>
            </div>
          </Chamber>
        </div>
      </section>
    </>
  )
}

export default CookiePolicyPage