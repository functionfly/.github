import React, { useState, useEffect, useRef } from 'react'
import {
  Chamber,
  CornerBrace,
  StatusPill,
  SealedButton,
  FrameButton,
} from './sc'
import '../styles/sc-main.css';

const SECTIONS = [
  { id: 'definitions', label: 'Definitions', number: 1 },
  { id: 'scope', label: 'Scope of Processing', number: 2 },
  { id: 'obligations', label: 'Processor Obligations', number: 3 },
  { id: 'subprocessors', label: 'Subprocessors', number: 4 },
  { id: 'transfers', label: 'International Transfers', number: 5 },
  { id: 'rights', label: 'Data Subject Rights', number: 6 },
  { id: 'security', label: 'Security', number: 7 },
  { id: 'breach', label: 'Breach Notification', number: 8 },
  { id: 'audit', label: 'Audits', number: 9 },
]

const KEY_OBLIGATIONS = [
  {
    title: '72-Hour Breach Notification',
    description: 'We notify you within 72 hours of becoming aware of any security incident affecting personal data.',
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
        <line x1="12" y1="9" x2="12" y2="13"></line>
        <line x1="12" y1="17" x2="12.01" y2="17"></line>
      </svg>
    ),
  },
  {
    title: 'SOC 2 Type II Certified',
    description: 'Our security controls are audited annually by independent third parties.',
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
        <path d="M9 12l2 2 4-4"></path>
      </svg>
    ),
  },
  {
    title: 'GDPR Art. 28 Compliant',
    description: 'Full compliance with EU data protection requirements for data processors.',
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <circle cx="12" cy="12" r="10"></circle>
        <path d="M12 16v-4"></path>
        <path d="M12 8h.01"></path>
      </svg>
    ),
  },
  {
    title: 'Subprocessor Transparency',
    description: 'Complete visibility into all third parties processing your data.',
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
        <circle cx="9" cy="7" r="4"></circle>
        <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
        <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
      </svg>
    ),
  },
]

const RELATED_DOCS = [
  {
    title: 'Security Policy',
    description: 'Detailed security architecture, encryption standards, and access controls.',
    href: '/security',
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
        <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
      </svg>
    ),
  },
  {
    title: 'Trust Center',
    description: 'Comprehensive overview of our compliance certifications and controls.',
    href: '/trust',
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
        <path d="M9 12l2 2 4-4"></path>
      </svg>
    ),
  },
  {
    title: 'Privacy Policy',
    description: 'How we collect, use, and protect your personal information.',
    href: '/privacy',
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
      </svg>
    ),
  },
  {
    title: 'Compliance',
    description: 'Our compliance program and certifications including SOC 2, ISO 27001.',
    href: '/compliance',
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
        <polyline points="22 4 12 14.01 9 11.01"></polyline>
      </svg>
    ),
  },
]

const DpaPage: React.FC = () => {
  const [searchQuery, setSearchQuery] = useState('')
  const [expandedSections, setExpandedSections] = useState<Set<string>>(new Set(['definitions', 'scope']))
  const [activeSection, setActiveSection] = useState('definitions')
  const [progress, setProgress] = useState(0)
  const [showSearch, setShowSearch] = useState(false)
  const [sidebarVisible, setSidebarVisible] = useState(false)
  const contentRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handleScroll = () => {
      if (!contentRef.current) return

      const scrollTop = window.scrollY
      const docHeight = document.documentElement.scrollHeight - window.innerHeight
      const scrollProgress = (scrollTop / docHeight) * 100
      setProgress(scrollProgress)

      SECTIONS.forEach((section) => {
        const element = document.getElementById(section.id)
        if (element) {
          const rect = element.getBoundingClientRect()
          if (rect.top <= 150 && rect.bottom >= 150) {
            setActiveSection(section.id)
          }
        }
      })
    }

    const handleResize = () => {
      setSidebarVisible(window.innerWidth >= 1024)
    }

    window.addEventListener('scroll', handleScroll)
    window.addEventListener('resize', handleResize)
    handleResize()

    return () => {
      window.removeEventListener('scroll', handleScroll)
      window.removeEventListener('resize', handleResize)
    }
  }, [])

  const toggleSection = (sectionId: string) => {
    setExpandedSections((prev) => {
      const next = new Set(prev)
      if (next.has(sectionId)) {
        next.delete(sectionId)
      } else {
        next.add(sectionId)
      }
      return next
    })
  }

  const scrollToSection = (sectionId: string) => {
    const element = document.getElementById(sectionId)
    if (element) {
      const top = element.getBoundingClientRect().top + window.scrollY - 100
      window.scrollTo({ top, behavior: 'smooth' })
      window.history.pushState(null, '', `#${sectionId}`)
    }
    setShowSearch(false)
    setSearchQuery('')
  }

  const filteredSections = SECTIONS.filter((s) =>
    s.label.toLowerCase().includes(searchQuery.toLowerCase())
  )

  return (
    <div className="dpa-page" ref={contentRef}>
      {/* Progress Bar */}
      <div className="dpa-progress-bar">
        <div className="dpa-progress-fill" style={{ width: `${progress}%` }} />
      </div>

      {/* Sticky Header with Actions */}
      <div className="dpa-sticky-header">
        <div className="dpa-sticky-inner">
          <div className="dpa-sticky-left">
            <span className="dpa-sticky-label">Data Processing Agreement</span>
            <span className="dpa-sticky-section">{SECTIONS.find(s => s.id === activeSection)?.label}</span>
          </div>
          <div className="dpa-sticky-actions">
            <button className="dpa-search-toggle" onClick={() => setShowSearch(!showSearch)} aria-label="Search">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <circle cx="11" cy="11" r="8"></circle>
                <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
              </svg>
            </button>
            <a href="/dpa.pdf" className="dpa-action-btn" download>
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                <polyline points="7 10 12 15 17 10"></polyline>
                <line x1="12" y1="15" x2="12" y2="3"></line>
              </svg>
              Download PDF
            </a>
          </div>
        </div>
        {showSearch && (
          <div className="dpa-search-dropdown">
            <input
              type="text"
              placeholder="Search sections..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="dpa-search-input"
              autoFocus
            />
            {searchQuery && (
              <div className="dpa-search-results">
                {filteredSections.map((section) => (
                  <button
                    key={section.id}
                    onClick={() => scrollToSection(section.id)}
                    className="dpa-search-result"
                  >
                    Section {section.number}: {section.label}
                  </button>
                ))}
                {filteredSections.length === 0 && (
                  <div className="dpa-search-no-results">No sections found</div>
                )}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Hero Header */}
      <section className="trust-hero">
        <Chamber ribs className="trust-hero-chamber">
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <div className="trust-hero-content">
            <div className="trust-hero-eyebrow">
              <StatusPill status="live" label="GDPR Art. 28" />
              <span className="dpa-last-updated">Last updated: June 2026</span>
            </div>
            <h1 className="trust-hero-title">
              Data Processing<br />Agreement
            </h1>
            <p className="trust-hero-subtitle">
              This Data Processing Agreement ("DPA") forms part of the Terms of Service between FunctionFly™ LLC ("Processor") and the customer ("Controller") for enterprise deployments requiring GDPR compliance.
            </p>
            <div className="dpa-hero-badges">
              <div className="dpa-badge">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
                  <path d="M9 12l2 2 4-4"></path>
                </svg>
                SOC 2 Type II
              </div>
              <div className="dpa-badge">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
                  <polyline points="22 4 12 14.01 9 11.01"></polyline>
                </svg>
                ISO 27001
              </div>
              <div className="dpa-badge">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                  <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
                </svg>
                GDPR Compliant
              </div>
            </div>
            <div className="dpa-hero-actions">
              <a href="/dpa.pdf" className="dpa-hero-btn dpa-hero-btn-primary" download>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                  <polyline points="7 10 12 15 17 10"></polyline>
                  <line x1="12" y1="15" x2="12" y2="3"></line>
                </svg>
                Download PDF
              </a>
              <a href="mailto:privacy@functionfly.com?subject=Executed%20DPA%20Request" className="dpa-hero-btn dpa-hero-btn-secondary">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path>
                  <polyline points="22,6 12,13 2,6"></polyline>
                </svg>
                Request Executed DPA
              </a>
            </div>
          </div>
        </Chamber>
      </section>

      <div className="dpa-layout">
        {/* Sidebar with Sticky TOC */}
        {sidebarVisible && (
          <aside className="dpa-sidebar">
            <div className="dpa-sidebar-inner">
              <h4 className="dpa-sidebar-title">Contents</h4>
              <nav className="dpa-sidebar-nav">
                {SECTIONS.map((section) => (
                  <button
                    key={section.id}
                    onClick={() => scrollToSection(section.id)}
                    className={`dpa-sidebar-link ${activeSection === section.id ? 'active' : ''}`}
                  >
                    <span className="dpa-sidebar-number">{section.number}</span>
                    <span className="dpa-sidebar-label">{section.label}</span>
                  </button>
                ))}
              </nav>
              <div className="dpa-sidebar-cta">
                <p>Need a custom DPA?</p>
                <a href="mailto:privacy@functionfly.com">Contact us</a>
              </div>
            </div>
          </aside>
        )}

        {/* Main Content */}
        <div className="dpa-content">
          {/* Key Obligations */}
          <section className="trust-section">
            <div className="trust-container">
              <div className="dpa-obligations-grid">
                {KEY_OBLIGATIONS.map((obligation, index) => (
                  <div key={index} className="dpa-obligation-card">
                    <div className="dpa-obligation-icon">{obligation.icon}</div>
                    <h3>{obligation.title}</h3>
                    <p>{obligation.description}</p>
                  </div>
                ))}
              </div>
            </div>
          </section>

          {/* Section 1: Definitions */}
          <section className="trust-section" id="definitions">
            <div className="trust-container">
              <Chamber>
                <CornerBrace position="tr" />
                <button
                  className="dpa-section-toggle"
                  onClick={() => toggleSection('definitions')}
                  aria-expanded={expandedSections.has('definitions')}
                >
                  <h2>1. Definitions</h2>
                  <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    className={expandedSections.has('definitions') ? 'expanded' : ''}
                  >
                    <polyline points="6 9 12 15 18 9"></polyline>
                  </svg>
                </button>
                {expandedSections.has('definitions') && (
                  <div className="dpa-section-content">
                    <p>In this DPA, the following terms have the meanings given below:</p>
                    <dl className="dpa-definitions">
                      <dt>"Applicable Data Protection Law"</dt>
                      <dd>means the General Data Protection Regulation (EU) 2016/679 ("GDPR") and applicable national implementing laws.</dd>

                      <dt>"Personal Data"</dt>
                      <dd>means any information relating to an identified or identifiable natural person as defined in Applicable Data Protection Law.</dd>

                      <dt>"Processing"</dt>
                      <dd>has the meaning given in Applicable Data Protection Law.</dd>

                      <dt>"Controller"</dt>
                      <dd>means the natural or legal person which, alone or jointly with others, determines the purposes and means of Processing.</dd>

                      <dt>"Processor"</dt>
                      <dd>means a natural or legal person that processes Personal Data on behalf of the Controller.</dd>

                      <dt>"Subprocessor"</dt>
                      <dd>means any processor engaged by the Processor to process Personal Data on behalf of the Controller.</dd>

                      <dt>"Data Subject"</dt>
                      <dd>means an identified or identifiable natural person whose Personal Data is processed.</dd>

                      <dt>"Security Incident"</dt>
                      <dd>means a breach of security leading to the accidental or unlawful destruction, loss, alteration, unauthorized disclosure of, or access to Personal Data.</dd>
                    </dl>
                  </div>
                )}
              </Chamber>
            </div>
          </section>

          {/* Section 2: Scope of Processing */}
          <section className="trust-section" id="scope">
            <div className="trust-container">
              <Chamber>
                <CornerBrace position="bl" />
                <button
                  className="dpa-section-toggle"
                  onClick={() => toggleSection('scope')}
                  aria-expanded={expandedSections.has('scope')}
                >
                  <h2>2. Subject Matter and Scope of Processing</h2>
                  <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    className={expandedSections.has('scope') ? 'expanded' : ''}
                  >
                    <polyline points="6 9 12 15 18 9"></polyline>
                  </svg>
                </button>
                {expandedSections.has('scope') && (
                  <div className="dpa-section-content">
                    <h3>2.1 Subject Matter</h3>
                    <p>
                      The Processor provides a cloud platform service for publishing, discovering, running, and managing serverless functions and related developer tools. The Processing of Personal Data under this DPA is limited to the data described in this Section 2.
                    </p>

                    <h3>2.2 Nature and Purpose of Processing</h3>
                    <p>The Processor shall process Personal Data for the following purposes:</p>
                    <ul>
                      <li>Providing and maintaining the FunctionFly™ platform service</li>
                      <li>Account management and authentication</li>
                      <li>Billing and subscription management</li>
                      <li>Function execution, logging, and monitoring</li>
                      <li>Security, abuse prevention, and fraud detection</li>
                      <li>Compliance with legal obligations</li>
                    </ul>

                    <h3>2.3 Categories of Personal Data</h3>
                    <p>The Processor may process the following categories of Personal Data:</p>
                    <ul>
                      <li>Account information (name, email, company, job title)</li>
                      <li>Authentication and session data (IP address, user agent, login timestamps)</li>
                      <li>Billing and transaction metadata</li>
                      <li>Function execution metadata (timestamps, function identifiers, error logs)</li>
                      <li>API usage logs and quota events</li>
                      <li>Support communications when you contact us</li>
                    </ul>

                    <h3>2.4 Categories of Data Subjects</h3>
                    <p>This DPA covers Processing of Personal Data relating to:</p>
                    <ul>
                      <li>Controller's users and employees who use the FunctionFly™ platform</li>
                      <li>API consumers and agents acting on behalf of the Controller</li>
                      <li>Any other individuals whose data the Controller submits to the Service</li>
                    </ul>

                    <h3>2.5 Duration</h3>
                    <p>
                      Processing under this DPA shall continue for the duration of the Terms of Service. Upon termination, the Processor will delete or return Personal Data according to Section 4 and the controller's instructions, except where retention is required by law.
                    </p>
                  </div>
                )}
              </Chamber>
            </div>
          </section>

          {/* Section 3: Processor Obligations */}
          <section className="trust-section" id="obligations">
            <div className="trust-container">
              <Chamber>
                <CornerBrace position="tr" />
                <button
                  className="dpa-section-toggle"
                  onClick={() => toggleSection('obligations')}
                  aria-expanded={expandedSections.has('obligations')}
                >
                  <h2>3. Processor Obligations</h2>
                  <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    className={expandedSections.has('obligations') ? 'expanded' : ''}
                  >
                    <polyline points="6 9 12 15 18 9"></polyline>
                  </svg>
                </button>
                {expandedSections.has('obligations') && (
                  <div className="dpa-section-content">
                    <p>The Processor shall:</p>

                    <h3>3.1 Compliance with Law</h3>
                    <p>
                      Process Personal Data only on documented instructions from the Controller, including with regard to transfers of Personal Data to a third country, unless required to do otherwise by Applicable Data Protection Law.
                    </p>

                    <h3>3.2 Personnel</h3>
                    <p>
                      Ensure that personnel authorized to process Personal Data have committed themselves to confidentiality or are under an appropriate statutory obligation of confidentiality.
                    </p>

                    <h3>3.3 Confidentiality</h3>
                    <p>
                      Not disclose Personal Data to third parties except to Subprocessors as specified in this DPA, or as required by law. The Processor shall require Subprocessors to maintain at least the same confidentiality obligations.
                    </p>

                    <h3>3.4 Security Measures</h3>
                    <p>
                      Implement appropriate technical and organizational measures to ensure a level of security appropriate to the risk, as described in Section 7 of this DPA and our <a href="/security">Security Policy</a>.
                    </p>

                    <h3>3.5 Subprocessors</h3>
                    <p>
                      Not engage additional Subprocessors without the Controller's prior general written authorization. The Controller grants general authorization for the engagement of Subprocessors listed in our <a href="/privacy#subprocessors">Privacy Policy subprocessor list</a>.
                    </p>

                    <h3>3.6 Assistance to Controller</h3>
                    <p>
                      Taking into account the nature of the Processing, assist the Controller by appropriate technical and organizational measures for the fulfillment of the Controller's obligations to respond to requests to exercise Data Subject rights.
                    </p>

                    <h3>3.7 Deletion or Return</h3>
                    <p>
                      At the Controller's choice, delete or return all Personal Data after the end of the provision of services, and delete existing copies unless retention is required by Applicable Data Protection Law.
                    </p>
                  </div>
                )}
              </Chamber>
            </div>
          </section>

          {/* Section 4: Subprocessors */}
          <section className="trust-section" id="subprocessors">
            <div className="trust-container">
              <Chamber>
                <CornerBrace position="bl" />
                <button
                  className="dpa-section-toggle"
                  onClick={() => toggleSection('subprocessors')}
                  aria-expanded={expandedSections.has('subprocessors')}
                >
                  <h2>4. Subprocessors</h2>
                  <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    className={expandedSections.has('subprocessors') ? 'expanded' : ''}
                  >
                    <polyline points="6 9 12 15 18 9"></polyline>
                  </svg>
                </button>
                {expandedSections.has('subprocessors') && (
                  <div className="dpa-section-content">
                    <h3>4.1 Authorized Subprocessors</h3>
                    <p>
                      The Controller authorizes the Processor to engage the Subprocessors listed in our <a href="/privacy#subprocessors">Privacy Policy</a> for the processing of Personal Data. This list may be updated from time to time; the Processor will provide notice of material changes.
                    </p>

                    <h3>4.2 Subprocessor Obligations</h3>
                    <p>
                      The Processor shall ensure that Subprocessors are bound by data processing terms no less protective than this DPA. The Processor remains fully liable to the Controller for the performance of Subprocessors' obligations.
                    </p>

                    <h3>4.3 Objection Right</h3>
                    <p>
                      If a Controller has a reasonable objection to a new Subprocessor, the Controller may terminate the affected services by providing written notice within 30 days of being notified of the addition.
                    </p>
                  </div>
                )}
              </Chamber>
            </div>
          </section>

          {/* Section 5: International Transfers */}
          <section className="trust-section" id="transfers">
            <div className="trust-container">
              <Chamber>
                <CornerBrace position="tr" />
                <button
                  className="dpa-section-toggle"
                  onClick={() => toggleSection('transfers')}
                  aria-expanded={expandedSections.has('transfers')}
                >
                  <h2>5. International Data Transfers</h2>
                  <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    className={expandedSections.has('transfers') ? 'expanded' : ''}
                  >
                    <polyline points="6 9 12 15 18 9"></polyline>
                  </svg>
                </button>
                {expandedSections.has('transfers') && (
                  <div className="dpa-section-content">
                    <h3>5.1 Transfer Mechanisms</h3>
                    <p>
                      Personal Data may be transferred internationally. When transferring Personal Data outside the European Economic Area (EEA), the Processor implements the following safeguards:
                    </p>
                    <ul>
                      <li>Standard Contractual Clauses (SCCs) approved by the European Commission</li>
                      <li>Adequacy decisions for transfers to countries with equivalent privacy protections</li>
                      <li>Binding Corporate Rules for intra-group transfers</li>
                      <li>Explicit consent where required by Applicable Data Protection Law</li>
                    </ul>

                    <h3>5.2 Data Location</h3>
                    <p>
                      Personal Data may be stored and processed in data centers located in the United States, European Union, and other jurisdictions as described in our <a href="/privacy">Privacy Policy</a>.
                    </p>
                  </div>
                )}
              </Chamber>
            </div>
          </section>

          {/* Section 6: Data Subject Rights */}
          <section className="trust-section" id="rights">
            <div className="trust-container">
              <Chamber>
                <CornerBrace position="bl" />
                <button
                  className="dpa-section-toggle"
                  onClick={() => toggleSection('rights')}
                  aria-expanded={expandedSections.has('rights')}
                >
                  <h2>6. Data Subject Rights</h2>
                  <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    className={expandedSections.has('rights') ? 'expanded' : ''}
                  >
                    <polyline points="6 9 12 15 18 9"></polyline>
                  </svg>
                </button>
                {expandedSections.has('rights') && (
                  <div className="dpa-section-content">
                    <p>The Processor shall assist the Controller in fulfilling its obligations to respond to Data Subject requests for:</p>
                    <ul>
                      <li>Right of access</li>
                      <li>Right to rectification</li>
                      <li>Right to erasure ("right to be forgotten")</li>
                      <li>Right to restrict processing</li>
                      <li>Right to data portability</li>
                      <li>Right to object</li>
                    </ul>
                    <p>
                      The Controller is responsible for directing Data Subjects to submit requests through the Controller's systems. The Processor will provide reasonable assistance within 30 days of receiving documented instructions from the Controller.
                    </p>
                  </div>
                )}
              </Chamber>
            </div>
          </section>

          {/* Section 7: Security */}
          <section className="trust-section" id="security">
            <div className="trust-container">
              <Chamber>
                <CornerBrace position="tr" />
                <button
                  className="dpa-section-toggle"
                  onClick={() => toggleSection('security')}
                  aria-expanded={expandedSections.has('security')}
                >
                  <h2>7. Security</h2>
                  <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    className={expandedSections.has('security') ? 'expanded' : ''}
                  >
                    <polyline points="6 9 12 15 18 9"></polyline>
                  </svg>
                </button>
                {expandedSections.has('security') && (
                  <div className="dpa-section-content">
                    <h3>7.1 Technical and Organizational Measures</h3>
                    <p>The Processor implements appropriate technical and organizational security measures including:</p>
                    <ul>
                      <li><strong>Encryption:</strong> TLS 1.3 for data in transit; AES-256 for data at rest</li>
                      <li><strong>Access controls:</strong> Role-based access control, MFA for administrative access</li>
                      <li><strong>Monitoring:</strong> 24/7 security monitoring, intrusion detection, SIEM integration</li>
                      <li><strong>Incident response:</strong> Documented procedures for security incidents</li>
                      <li><strong>Vulnerability management:</strong> Regular security assessments and patches</li>
                    </ul>

                    <h3>7.2 Security Documentation</h3>
                    <p>
                      Full security architecture details are available in our <a href="/security">Security Policy</a> and <a href="/trust">Trust Center</a>.
                    </p>
                  </div>
                )}
              </Chamber>
            </div>
          </section>

          {/* Section 8: Breach Notification */}
          <section className="trust-section" id="breach">
            <div className="trust-container">
              <Chamber>
                <CornerBrace position="bl" />
                <button
                  className="dpa-section-toggle"
                  onClick={() => toggleSection('breach')}
                  aria-expanded={expandedSections.has('breach')}
                >
                  <h2>8. Security Incident Notification</h2>
                  <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    className={expandedSections.has('breach') ? 'expanded' : ''}
                  >
                    <polyline points="6 9 12 15 18 9"></polyline>
                  </svg>
                </button>
                {expandedSections.has('breach') && (
                  <div className="dpa-section-content">
                    <h3>8.1 Notification Obligation</h3>
                    <p>
                      The Processor shall notify the Controller without undue delay (and in any event within 72 hours) after becoming aware of a Security Incident affecting Personal Data.
                    </p>

                    <h3>8.2 Information Provided</h3>
                    <p>When notifying the Controller of a Security Incident, the Processor shall provide:</p>
                    <ul>
                      <li>Description of the nature of the Security Incident</li>
                      <li>Categories and approximate number of affected Data Subjects</li>
                      <li>Categories and approximate number of affected Personal Data records</li>
                      <li>Contact point for further information</li>
                      <li>Likely consequences of the Security Incident</li>
                      <li>Measures taken or proposed to address the Security Incident</li>
                    </ul>

                    <h3>8.3 Exclusions</h3>
                    <p>
                      Notification obligations do not apply when the Processor determines that the Security Incident is unlikely to result in a risk to the rights and freedoms of Data Subjects, or when notification would involve disproportionate effort.
                    </p>
                  </div>
                )}
              </Chamber>
            </div>
          </section>

          {/* Section 9: Audits */}
          <section className="trust-section" id="audit">
            <div className="trust-container">
              <Chamber>
                <CornerBrace position="tr" />
                <button
                  className="dpa-section-toggle"
                  onClick={() => toggleSection('audit')}
                  aria-expanded={expandedSections.has('audit')}
                >
                  <h2>9. Audits</h2>
                  <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    className={expandedSections.has('audit') ? 'expanded' : ''}
                  >
                    <polyline points="6 9 12 15 18 9"></polyline>
                  </svg>
                </button>
                {expandedSections.has('audit') && (
                  <div className="dpa-section-content">
                    <h3>9.1 Audit Rights</h3>
                    <p>
                      The Controller may audit the Processor's compliance with this DPA by requesting:
                    </p>
                    <ul>
                      <li>Copies of SOC 2 Type II audit reports (available upon request under NDA)</li>
                      <li>Documentation of security policies and procedures</li>
                      <li>Information about the Processor's security certifications</li>
                    </ul>

                    <h3>9.2 On-Site Audits</h3>
                    <p>
                      For enterprise customers requiring on-site audits, the Processor will cooperate with reasonable audit requests. The Controller shall provide at least 30 days' prior written notice. Audits shall be conducted during business hours without disrupting the Processor's operations.
                    </p>

                    <h3>9.3 Audit Costs</h3>
                    <p>
                      The Controller shall bear the costs of any audit. If an audit reveals material non-compliance, the Processor shall bear its own costs and remedy the non-compliance within a reasonable timeframe.
                    </p>
                  </div>
                )}
              </Chamber>
            </div>
          </section>

          {/* Section 10: General Provisions */}
          <section className="trust-section">
            <div className="trust-container">
              <Chamber>
                <CornerBrace position="bl" />
                <button
                  className="dpa-section-toggle"
                  onClick={() => toggleSection('general')}
                  aria-expanded={expandedSections.has('general')}
                >
                  <h2>10. General Provisions</h2>
                  <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    className={expandedSections.has('general') ? 'expanded' : ''}
                  >
                    <polyline points="6 9 12 15 18 9"></polyline>
                  </svg>
                </button>
                {expandedSections.has('general') && (
                  <div className="dpa-section-content">
                    <h3>10.1 Order of Precedence</h3>
                    <p>
                      This DPA forms part of the Terms of Service. In case of conflict between this DPA and the Terms of Service, this DPA shall prevail for matters related to data protection.
                    </p>

                    <h3>10.2 Governing Law</h3>
                    <p>
                      This DPA shall be governed by the same law as the Terms of Service.
                    </p>

                    <h3>10.3 Contact</h3>
                    <p>
                      For questions about this DPA or to request an executed copy, contact: <a href="mailto:privacy@functionfly.com">privacy@functionfly.com</a>
                    </p>

                    <h3>10.4 Mutual Execution</h3>
                    <p>
                      This DPA may be executed electronically. The Processor will provide a countersigned copy upon request from the Controller.
                    </p>
                  </div>
                )}
              </Chamber>
            </div>
          </section>

          {/* Enterprise Contact CTA */}
          <section className="trust-section">
            <div className="trust-container">
              <Chamber className="dpa-cta-chamber">
                <CornerBrace position="tr" />
                <CornerBrace position="bl" />
                <div className="dpa-cta-content">
                  <h2>Need a Custom DPA?</h2>
                  <p>
                    Enterprise customers may require custom data processing terms, additional security questionnaires, or on-site audit arrangements. Our legal team is ready to help.
                  </p>
                  <div className="dpa-cta-form">
                    <input
                      type="email"
                      placeholder="Work email"
                      className="dpa-cta-input"
                    />
                    <button className="dpa-cta-submit">
                      Contact Us
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <line x1="5" y1="12" x2="19" y2="12"></line>
                        <polyline points="12 5 19 12 12 19"></polyline>
                      </svg>
                    </button>
                  </div>
                  <p className="dpa-cta-note">
                    Typically respond within 1 business day. For urgent matters, email{' '}
                    <a href="mailto:privacy@functionfly.com">privacy@functionfly.com</a> directly.
                  </p>
                </div>
              </Chamber>
            </div>
          </section>

          {/* Related Docs */}
          <section className="trust-section">
            <div className="trust-container">
              <Chamber nested>
                <h3>Related Documentation</h3>
                <div className="dpa-related-docs-grid">
                  {RELATED_DOCS.map((doc, index) => (
                    <a key={index} href={doc.href} className="dpa-related-doc-card">
                      <div className="dpa-related-doc-icon">{doc.icon}</div>
                      <div className="dpa-related-doc-content">
                        <h4>{doc.title}</h4>
                        <p>{doc.description}</p>
                      </div>
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="dpa-related-doc-arrow">
                        <line x1="5" y1="12" x2="19" y2="12"></line>
                        <polyline points="12 5 19 12 12 19"></polyline>
                      </svg>
                    </a>
                  ))}
                </div>
              </Chamber>
            </div>
          </section>
        </div>
      </div>
    </div>
  )
}

export default DpaPage