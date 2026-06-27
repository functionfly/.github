import { Logo } from '@/components/common/Logo';
import { Button } from '@/components/ui/button';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
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
import { motion } from 'framer-motion';
import { AlertCircle, ArrowUp, Check, Heart, Mail } from 'lucide-react';
import { FormEvent, useState } from 'react';

interface FooterProps {
  /** Set to false when another fixed bottom-right element (e.g. fly guide) is shown to avoid overlap. Default true. */
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

    if (!isValidEmail(trimmedEmail)) {
      return;
    }

    const success = await subscribe({ email: trimmedEmail });
    if (success) {
      setEmail('');
      setTimeout(() => {
        reset();
      }, 5000);
    }
  };

  const handleUnsubscribe = async (e: FormEvent) => {
    e.preventDefault();
    const trimmedEmail = unsubscribeEmail.trim();
    if (!trimmedEmail || unsubscribeLoading) return;

    if (!isValidEmail(trimmedEmail)) {
      return;
    }

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

  return (
    <footer
      className="relative overflow-hidden sc-footer"
      role="contentinfo"
      aria-label="Site footer"
    >
      {/* Background Effects */}
      <div className="absolute top-0 left-1/4 w-96 h-96 bg-brand-500/5 rounded-full blur-[128px]" />
      <div className="absolute bottom-0 right-1/4 w-96 h-96 bg-purple-500/5 rounded-full blur-[128px]" />

      <div className="relative border-t border-border-subtle">
        {/* Main Footer Content */}
        <div className="max-w-7xl mx-auto px-4 lg:px-6 py-16">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-12 gap-8 lg:gap-12">
            {/* Company Info */}
            <div className="lg:col-span-4">
              <div className="mb-6">
                <Logo size="md" />
              </div>
              <p className="sc-footer-text mb-6 max-w-sm">
                Serverless functions & AI agent infrastructure. Deploy to edge, build AI agents with
                built-in cost controls, and scale with confidence.
              </p>
              <div className="flex items-center gap-4" role="group" aria-label="Social links">
                <motion.a
                  href="https://github.com/functionfly"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="sc-footer-social"
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  aria-label="FunctionFly on GitHub"
                >
                  <GitHubIcon className="w-5 h-5" />
                </motion.a>
                <motion.a
                  href="https://x.com/functionflycom"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="sc-footer-social"
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  aria-label="FunctionFly on X"
                >
                  <XIcon className="w-5 h-5" />
                </motion.a>
                <motion.a
                  href="https://instagram.com/functionfly"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="sc-footer-social"
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  aria-label="FunctionFly on Instagram"
                >
                  <InstagramIcon className="w-5 h-5" />
                </motion.a>
                <motion.a
                  href="#"
                  className="sc-footer-social"
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  aria-label="FunctionFly on LinkedIn"
                >
                  <LinkedInIcon className="w-5 h-5" />
                </motion.a>
              </div>
            </div>

            {/* Product */}
            <div className="lg:col-span-2">
              <h3 className="sc-footer-text-primary font-semibold mb-4">Product</h3>
              <ul className="space-y-3">
                <li>
                  <a href="/features" className="sc-footer-nav-link text-sm">
                    Features
                  </a>
                </li>
                <li>
                  <a href="/pricing" className="sc-footer-nav-link text-sm">
                    Pricing
                  </a>
                </li>
                <li>
                  <a
                    href={isAuthenticated ? '/state-fabric' : '/products/state-fabric'}
                    className="sc-footer-nav-link text-sm"
                  >
                    State Fabric
                  </a>
                </li>
                <li>
                  <a href="/agents" className="sc-footer-nav-link text-sm">
                    AI Agents
                  </a>
                </li>
                <li>
                  <a href="/registry" className="sc-footer-nav-link text-sm">
                    Function Registry
                  </a>
                </li>
                <li>
                  <a href="/security" className="sc-footer-nav-link text-sm">
                    Security
                  </a>
                </li>
                <li>
                  <a
                    href={STATUS_SITE_URL}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="sc-footer-nav-link text-sm"
                  >
                    System Status
                  </a>
                </li>
                <li>
                  <a href="/changelog" className="sc-footer-nav-link text-sm">
                    Changelog
                  </a>
                </li>
              </ul>
            </div>

            {/* Resources */}
            <div className="lg:col-span-2">
              <h3 className="sc-footer-text-primary font-semibold mb-4">Resources</h3>
              <ul className="space-y-3">
                <li>
                  <a
                    href={STATUS_SITE_URL}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="sc-footer-nav-link text-sm"
                  >
                    System Status
                  </a>
                </li>
                <li>
                  <a href="#" className="sc-footer-nav-link text-sm">
                    API Reference
                  </a>
                </li>
                <li>
                  <a href="#" className="sc-footer-nav-link text-sm">
                    MCP Protocol
                  </a>
                </li>
                <li>
                  <a href="#" className="sc-footer-nav-link text-sm">
                    Guides
                  </a>
                </li>
                <li>
                  <a
                    href={BLOG_SITE_URL}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="sc-footer-nav-link text-sm"
                  >
                    Blog
                  </a>
                </li>
                <li>
                  <a href="/faq" className="sc-footer-nav-link text-sm">
                    FAQ
                  </a>
                </li>
              </ul>
            </div>

            {/* Support */}
            <div className="lg:col-span-2">
              <h3 className="sc-footer-text-primary font-semibold mb-4">Support</h3>
              <ul className="space-y-3">
                <li>
                  <a href="/help" className="sc-footer-nav-link text-sm">
                    Help Center
                  </a>
                </li>
                <li>
                  <a href="/community" className="sc-footer-nav-link text-sm">
                    Community
                  </a>
                </li>
                <li>
                  <a href="/contact" className="sc-footer-nav-link text-sm">
                    Contact Us
                  </a>
                </li>
                <li>
                  <a
                    href={STATUS_SITE_URL}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="sc-footer-nav-link text-sm"
                  >
                    System Status
                  </a>
                </li>
                <li>
                  <a href="/feedback" className="sc-footer-nav-link text-sm">
                    Feedback
                  </a>
                </li>
              </ul>
            </div>

            {/* Community */}
            <div className="lg:col-span-2">
              <h3 className="sc-footer-text-primary font-semibold mb-4">Community</h3>
              <ul className="space-y-3">
                <li>
                  <a
                    href="https://discord.gg/functionfly"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="sc-footer-nav-link text-sm"
                  >
                    Discord
                  </a>
                </li>
                <li>
                  <a
                    href="https://github.com/functionfly/functionfly/discussions"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="sc-footer-nav-link text-sm"
                  >
                    GitHub Discussions
                  </a>
                </li>
                <li>
                  <a
                    href="https://x.com/functionflycom"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="sc-footer-nav-link text-sm"
                  >
                    X (Twitter)
                  </a>
                </li>
                <li>
                  <a
                    href="https://linkedin.com/company/functionfly"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="sc-footer-nav-link text-sm"
                  >
                    LinkedIn
                  </a>
                </li>
              </ul>
            </div>

            {/* Newsletter */}
            <div className="lg:col-span-2">
              <h3 className="sc-footer-text-primary font-semibold mb-4">Stay Updated</h3>
              <p className="sc-footer-text text-sm mb-4">
                Get the latest updates and news about FunctionFly™.
              </p>
              <form onSubmit={handleSubscribe} className="space-y-3">
                <div className="flex gap-2">
                  <input
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="Enter your email"
                    disabled={isLoading || isSuccess}
                    className="sc-footer-input flex-1 px-3 py-2 rounded-lg text-sm focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed min-w-[200px]"
                  />
                  <motion.button
                    type="submit"
                    disabled={isLoading || isSuccess || !email.trim()}
                    className="sc-footer-btn"
                    whileHover={!isLoading && !isSuccess ? { scale: 1.02 } : {}}
                    whileTap={!isLoading && !isSuccess ? { scale: 0.98 } : {}}
                  >
                    {isLoading ? (
                      <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                    ) : isSuccess ? (
                      <Check className="w-4 h-4" />
                    ) : (
                      <Mail className="w-4 h-4" />
                    )}
                  </motion.button>
                </div>
                {error && (
                  <div className="flex items-center gap-2 text-error text-xs">
                    <AlertCircle className="w-3 h-3" />
                    <span>{error}</span>
                  </div>
                )}
                {isSuccess && (
                  <div className="flex items-center gap-2 text-success text-xs">
                    <Check className="w-3 h-3" />
                    <span>Successfully subscribed! Check your email for confirmation.</span>
                  </div>
                )}
                {!error && !isSuccess && (
                  <p className="sc-footer-text text-xs">
                    We respect your privacy.{' '}
                    <button
                      type="button"
                      onClick={() => setShowUnsubscribeDialog(true)}
                      className="text-brand-500 hover:text-brand-400 underline"
                    >
                      Unsubscribe
                    </button>{' '}
                    at any time.
                  </p>
                )}
              </form>
            </div>
          </div>
        </div>

        {/* Bottom Bar */}
        <div className="border-t border-border-subtle">
          <div className="max-w-7xl mx-auto px-4 lg:px-6 py-6">
            <div className="flex flex-col md:flex-row items-center justify-between gap-4">
              <div className="flex items-center gap-2 text-sm text-text-muted">
                <span className="sc-footer-text">
                  © {new Date().getFullYear()} FunctionFly™ LLC. Made with
                </span>
                <Heart className="w-4 h-4 text-error fill-current" />
                <span className="sc-footer-text">for indie developers.</span>
              </div>
              <div className="flex items-center gap-6 text-sm">
                <a
                  href={getMarketingPageUrl('/terms')}
                  className="sc-footer-nav-link"
                  rel="noopener noreferrer"
                >
                  Terms of Service
                </a>
                <a
                  href={getMarketingPageUrl('/privacy')}
                  className="sc-footer-nav-link"
                  rel="noopener noreferrer"
                >
                  Privacy Policy
                </a>
                <a
                  href={getMarketingPageUrl('/privacy#cookies')}
                  className="sc-footer-nav-link"
                  rel="noopener noreferrer"
                >
                  Cookie Policy
                </a>
              </div>
            </div>
          </div>
        </div>

        {/* Back to Top Button - hidden in dashboard layout so fly guide bot is visible */}
        {showScrollToTop && (
          <motion.button
            onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })}
            className="fixed bottom-8 right-8 w-12 h-12 bg-gradient-to-r from-brand-500 to-purple-500 rounded-full flex items-center justify-center text-white shadow-lg hover:shadow-xl hover:shadow-brand-500/25 transition-all duration-200 z-50 glow hover-lift"
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
          >
            <ArrowUp className="w-5 h-5" />
          </motion.button>
        )}

        {/* Unsubscribe Dialog */}
        <Dialog open={showUnsubscribeDialog} onOpenChange={setShowUnsubscribeDialog}>
          <DialogContent className="sm:max-w-[425px]">
            <DialogHeader>
              <DialogTitle>Unsubscribe from Newsletter</DialogTitle>
              <DialogDescription>
                Enter your email address to unsubscribe from the FunctionFly newsletter.
              </DialogDescription>
            </DialogHeader>
            <form onSubmit={handleUnsubscribe} className="space-y-4">
              <div>
                <Input
                  type="email"
                  placeholder="your.email@example.com"
                  value={unsubscribeEmail}
                  onChange={(e) => setUnsubscribeEmail(e.target.value)}
                  disabled={unsubscribeLoading || unsubscribeSuccess}
                  className="w-full"
                />
              </div>
              {unsubscribeSuccess ? (
                <div className="flex items-center gap-2 text-success text-sm">
                  <Check className="w-4 h-4" />
                  <span>Successfully unsubscribed!</span>
                </div>
              ) : (
                <DialogFooter>
                  <Button
                    type="button"
                    variant="ghost"
                    onClick={() => setShowUnsubscribeDialog(false)}
                    disabled={unsubscribeLoading}
                  >
                    Cancel
                  </Button>
                  <Button
                    type="submit"
                    disabled={
                      !unsubscribeEmail.trim() ||
                      unsubscribeLoading ||
                      !isValidEmail(unsubscribeEmail.trim())
                    }
                  >
                    {unsubscribeLoading ? (
                      <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                    ) : (
                      'Unsubscribe'
                    )}
                  </Button>
                </DialogFooter>
              )}
            </form>
          </DialogContent>
        </Dialog>
      </div>
    </footer>
  );
}
