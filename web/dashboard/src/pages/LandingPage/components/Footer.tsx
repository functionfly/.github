import { Logo } from '@/components/common/Logo';
import { useNewsletter } from '@/hooks/useNewsletter';
import { BLOG_SITE_URL, getMarketingPageUrl, STATUS_SITE_URL } from '@/lib/constants';
import { isValidEmail } from '@/lib/url-utils';
import {
    GitHubIcon,
    InstagramIcon,
    LinkedInIcon,
    XIcon,
} from '@/pages/LandingPage/components/icons';
import { useAuthStore } from '@/stores/authStore';
import '@/styles/sc-footer.css';
import { AlertCircle, ArrowUp, Check, Heart, Mail } from 'lucide-react';
import { FormEvent, useState } from 'react';

interface FooterProps {
  showScrollToTop?: boolean;
}

export function Footer({ showScrollToTop = true }: FooterProps) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const { subscribe, unsubscribe, isLoading, isSuccess, error, reset } = useNewsletter();
  const [email, setEmail] = useState('');
  const [showUnsubscribeDialog, setShowUnsubscribeDialog] = useState(false);
  const [unsubscribeEmail, setUnsubscribeEmail] = useState('');
  const [unsubscribeLoading, setUnsubscribeLoading] = useState(false);
  const [unsubscribeSuccess, setUnsubscribeSuccess] = useState(false);

  const handleSubscribe = async (e: FormEvent) => {
    e.preventDefault();
    const trimmedEmail = email.trim();
    if (!trimmedEmail || isLoading) return;
    if (!isValidEmail(trimmedEmail)) return;
    const success = await subscribe({ email: trimmedEmail });
    if (success) {
      setEmail('');
      setTimeout(() => { reset(); }, 5000);
    }
  };

  const handleUnsubscribe = async (e: FormEvent) => {
    e.preventDefault();
    const trimmedEmail = unsubscribeEmail.trim();
    if (!trimmedEmail || unsubscribeLoading) return;
    if (!isValidEmail(trimmedEmail)) return;
    setUnsubscribeLoading(true);
    const success = await unsubscribe(trimmedEmail);
    setUnsubscribeLoading(false);
    if (success) {
      setUnsubscribeSuccess(true);
      setTimeout(() => {
        setShowUnsubscribeDialog(false);
        setUnsubscribeEmail('');
        setUnsubscribeSuccess(false);
      }, 3000);
    }
  };

  const footerNavSections = [
    {
      title: 'Product',
      links: [
        { label: 'Features', href: '/features' },
        { label: 'Pricing', href: '/pricing' },
        { label: 'State Fabric', href: isAuthenticated ? '/state-fabric' : '/products/state-fabric' },
        { label: 'AI Agents', href: '/agents' },
        { label: 'Function Registry', href: '/registry' },
        { label: 'Security', href: '/security' },
        { label: 'System Status', href: STATUS_SITE_URL, external: true },
        { label: 'Changelog', href: '/changelog' },
      ],
    },
    {
      title: 'Resources',
      links: [
        { label: 'System Status', href: STATUS_SITE_URL, external: true },
        { label: 'API Reference', href: '#' },
        { label: 'MCP Protocol', href: '#' },
        { label: 'Guides', href: '#' },
        { label: 'Blog', href: BLOG_SITE_URL, external: true },
        { label: 'FAQ', href: '/faq' },
      ],
    },
    {
      title: 'Support',
      links: [
        { label: 'Help Center', href: '/help' },
        { label: 'Community', href: '/community' },
        { label: 'Contact Us', href: '/contact' },
        { label: 'System Status', href: STATUS_SITE_URL, external: true },
        { label: 'Feedback', href: '/feedback' },
      ],
    },
    {
      title: 'Community',
      links: [
        { label: 'Discord', href: 'https://discord.gg/functionfly', external: true },
        { label: 'GitHub Discussions', href: 'https://github.com/functionfly/functionfly/discussions', external: true },
        { label: 'X (Twitter)', href: 'https://x.com/functionflycom', external: true },
        { label: 'LinkedIn', href: 'https://linkedin.com/company/functionfly', external: true },
      ],
    },
  ];

  const socialLinks = [
    { icon: GitHubIcon, href: 'https://github.com/functionfly', label: 'GitHub' },
    { icon: XIcon, href: 'https://x.com/functionflycom', label: 'X' },
    { icon: InstagramIcon, href: 'https://instagram.com/functionfly', label: 'Instagram' },
    { icon: LinkedInIcon, href: '#', label: 'LinkedIn' },
  ];

  return (
    <footer className="sc-footer" role="contentinfo" aria-label="Site footer">
      <div className="sc-footer__inner">
        {/* Main Grid */}
        <div className="sc-footer__grid">
          {/* Company Info */}
          <div className="sc-footer__brand">
            <div className="sc-footer__logo">
              <Logo size="md" />
            </div>
            <p className="sc-footer__desc">
              Serverless functions & AI agent infrastructure. Deploy to edge, build AI agents with
              built-in cost controls, and scale with confidence.
            </p>
            <div className="sc-footer__socials" role="group" aria-label="Social links">
              {socialLinks.map((social) => (
                <a
                  key={social.label}
                  href={social.href}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="sc-footer-social"
                  aria-label={`FunctionFly on ${social.label}`}
                >
                  <social.icon className="sc-footer-social__icon" />
                </a>
              ))}
            </div>
          </div>

          {/* Nav Sections */}
          {footerNavSections.map((section) => (
            <div key={section.title} className="sc-footer__nav-col">
              <h3 className="sc-footer__nav-title">{section.title}</h3>
              <ul className="sc-footer__nav-list">
                {section.links.map((link) => (
                  <li key={link.label}>
                    <a
                      href={link.href}
                      className="sc-footer-nav-link"
                      {...(link.external ? { target: '_blank', rel: 'noopener noreferrer' } : {})}
                    >
                      {link.label}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          ))}

          {/* Newsletter */}
          <div className="sc-footer__newsletter">
            <h3 className="sc-footer__nav-title">Stay Updated</h3>
            <p className="sc-footer__newsletter-desc">
              Get the latest updates and news about FunctionFly.
            </p>
            <form onSubmit={handleSubscribe} className="sc-footer__form">
              <div className="sc-footer__form-row">
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="Enter your email"
                  disabled={isLoading || isSuccess}
                  className="sc-footer-input"
                />
                <button
                  type="submit"
                  disabled={isLoading || isSuccess || !email.trim()}
                  className="sc-footer-btn"
                >
                  {isLoading ? (
                    <span className="sc-footer__spinner" />
                  ) : isSuccess ? (
                    <Check className="sc-footer__btn-icon" />
                  ) : (
                    <Mail className="sc-footer__btn-icon" />
                  )}
                </button>
              </div>
              {error && (
                <div className="sc-footer__alert sc-footer__alert--error">
                  <AlertCircle className="sc-footer__alert-icon" />
                  <span>{error}</span>
                </div>
              )}
              {isSuccess && (
                <div className="sc-footer__alert sc-footer__alert--ok">
                  <Check className="sc-footer__alert-icon" />
                  <span>Successfully subscribed! Check your email for confirmation.</span>
                </div>
              )}
              {!error && !isSuccess && (
                <p className="sc-footer__form-hint">
                  We respect your privacy.{' '}
                  <button
                    type="button"
                    onClick={() => setShowUnsubscribeDialog(true)}
                    className="sc-footer__form-link"
                  >
                    Unsubscribe
                  </button>{' '}
                  at any time.
                </p>
              )}
            </form>
          </div>
        </div>

        {/* Bottom Bar */}
        <div className="sc-footer__bottom">
          <div className="sc-footer__bottom-inner">
            <div className="sc-footer__copyright">
              <span className="sc-footer-text">
                &copy; {new Date().getFullYear()} FunctionFly LLC. Made with
              </span>
              <Heart className="sc-footer__heart" />
              <span className="sc-footer-text">for indie developers.</span>
            </div>
            <div className="sc-footer__legal">
              <a href={getMarketingPageUrl('/terms')} className="sc-footer-nav-link" rel="noopener noreferrer">
                Terms of Service
              </a>
              <a href={getMarketingPageUrl('/privacy')} className="sc-footer-nav-link" rel="noopener noreferrer">
                Privacy Policy
              </a>
              <a href={getMarketingPageUrl('/privacy#cookies')} className="sc-footer-nav-link" rel="noopener noreferrer">
                Cookie Policy
              </a>
            </div>
          </div>
        </div>

        {/* Back to Top */}
        {showScrollToTop && (
          <button
            onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })}
            className="sc-footer__scroll-top"
            aria-label="Scroll to top"
          >
            <ArrowUp className="sc-footer__scroll-icon" />
          </button>
        )}
      </div>

      {/* Unsubscribe Dialog */}
      {showUnsubscribeDialog && (
        <div className="sc-footer__modal-overlay" onClick={() => setShowUnsubscribeDialog(false)}>
          <div className="sc-footer__modal" onClick={(e) => e.stopPropagation()}>
            <h2 className="sc-footer__modal-title">Unsubscribe from Newsletter</h2>
            <p className="sc-footer__modal-desc">
              Enter your email address to unsubscribe from the FunctionFly newsletter.
            </p>
            <form onSubmit={handleUnsubscribe} className="sc-footer__modal-form">
              <input
                type="email"
                placeholder="your.email@example.com"
                value={unsubscribeEmail}
                onChange={(e) => setUnsubscribeEmail(e.target.value)}
                disabled={unsubscribeLoading || unsubscribeSuccess}
                className="sc-footer-input"
              />
              {unsubscribeSuccess ? (
                <div className="sc-footer__alert sc-footer__alert--ok">
                  <Check className="sc-footer__alert-icon" />
                  <span>Successfully unsubscribed!</span>
                </div>
              ) : (
                <div className="sc-footer__modal-actions">
                  <button
                    type="button"
                    onClick={() => setShowUnsubscribeDialog(false)}
                    disabled={unsubscribeLoading}
                    className="sc-footer__modal-cancel"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    disabled={!unsubscribeEmail.trim() || unsubscribeLoading || !isValidEmail(unsubscribeEmail.trim())}
                    className="sc-footer-btn"
                  >
                    {unsubscribeLoading ? <span className="sc-footer__spinner" /> : 'Unsubscribe'}
                  </button>
                </div>
              )}
            </form>
          </div>
        </div>
      )}
    </footer>
  );
}
