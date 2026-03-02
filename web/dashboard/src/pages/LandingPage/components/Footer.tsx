import { motion } from "framer-motion";
import { Github, Twitter, Linkedin, MessageCircle, Mail, ArrowUp, Heart } from "lucide-react";
import { Logo } from "@/components/common/Logo";
import { DOCS_SITE_URL } from "@/lib/constants";

export function Footer() {

  return (
    <footer className="relative overflow-hidden footer-enhanced" style={{ backgroundColor: 'var(--bg-primary)' }}>
      {/* Background Effects */}
      <div className="absolute top-0 left-1/4 w-96 h-96 bg-brand-500/5 rounded-full blur-[128px] light-mode-enhanced" />
      <div className="absolute bottom-0 right-1/4 w-96 h-96 bg-purple-500/5 rounded-full blur-[128px] light-mode-enhanced" />

      <div className="relative border-t border-border-subtle">
        {/* Main Footer Content */}
        <div className="max-w-7xl mx-auto px-4 lg:px-6 py-16">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-12 gap-8 lg:gap-12">
            {/* Company Info */}
            <div className="lg:col-span-4">
              <div className="mb-6">
                <Logo size="md" />
              </div>
              <p className="text-text-secondary mb-6 max-w-sm">
                Serverless functions & AI agent infrastructure. Deploy to edge,
                build AI agents with built-in cost controls, and scale with confidence.
              </p>
              <div className="flex items-center gap-4">
                <motion.a
                  href="#"
                  className="w-10 h-10 rounded-xl bg-bg-hover border border-border-subtle flex items-center justify-center text-text-secondary hover:text-text-primary hover:bg-bg-hover hover:border-brand-500/30 transition-all duration-200 shine-effect shine-effect"
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                >
                  <Github className="w-5 h-5" />
                </motion.a>
                <motion.a
                  href="#"
                  className="w-10 h-10 rounded-xl bg-bg-hover border border-border-subtle flex items-center justify-center text-text-secondary hover:text-text-primary hover:bg-bg-hover hover:border-brand-500/30 transition-all duration-200 shine-effect"
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                >
                  <Twitter className="w-5 h-5" />
                </motion.a>
                <motion.a
                  href="#"
                  className="w-10 h-10 rounded-xl bg-bg-hover border border-border-subtle flex items-center justify-center text-text-secondary hover:text-text-primary hover:bg-bg-hover hover:border-brand-500/30 transition-all duration-200 shine-effect"
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                >
                  <Linkedin className="w-5 h-5" />
                </motion.a>
                <motion.a
                  href="#"
                  className="w-10 h-10 rounded-xl bg-bg-hover border border-border-subtle flex items-center justify-center text-text-secondary hover:text-text-primary hover:bg-bg-hover hover:border-brand-500/30 transition-all duration-200 shine-effect"
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                >
                  <MessageCircle className="w-5 h-5" />
                </motion.a>
              </div>
            </div>

            {/* Product */}
            <div className="lg:col-span-2">
              <h3 className="text-text-primary font-semibold mb-4">Product</h3>
              <ul className="space-y-3">
                <li>
                  <a
                    href="/features"
                    className="text-text-secondary hover:text-text-primary transition-colors underline-animation text-sm underline-animation underline-animation"
                  >
                    Features
                  </a>
                </li>
                <li>
                  <a
                    href="/pricing"
                    className="text-text-secondary hover:text-text-primary transition-colors underline-animation text-sm underline-animation"
                  >
                    Pricing
                  </a>
                </li>
                <li>
                  <a
                    href="/state-fabric"
                    className="text-text-secondary hover:text-text-primary transition-colors underline-animation text-sm underline-animation"
                  >
                    State Fabric
                  </a>
                </li>
                <li>
                  <a
                    href="/agents"
                    className="text-text-secondary hover:text-text-primary transition-colors underline-animation text-sm underline-animation"
                  >
                    AI Agents
                  </a>
                </li>
                <li>
                  <a
                    href="/registry"
                    className="text-text-secondary hover:text-text-primary transition-colors underline-animation text-sm underline-animation"
                  >
                    Function Registry
                  </a>
                </li>
                <li>
                  <a
                    href="/security"
                    className="text-text-secondary hover:text-text-primary transition-colors underline-animation text-sm underline-animation"
                  >
                    Security
                  </a>
                </li>
                <li>
                  <a
                    href="/integrations"
                    className="text-text-secondary hover:text-text-primary transition-colors underline-animation text-sm underline-animation"
                  >
                    Integrations
                  </a>
                </li>
                <li>
                  <a
                    href="/changelog"
                    className="text-text-secondary hover:text-text-primary transition-colors underline-animation text-sm underline-animation"
                  >
                    Changelog
                  </a>
                </li>
              </ul>
            </div>

            {/* Resources */}
            <div className="lg:col-span-2">
              <h3 className="text-text-primary font-semibold mb-4">Resources</h3>
              <ul className="space-y-3">
                <li>
                  <a
                    href={DOCS_SITE_URL}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-text-secondary hover:text-text-primary transition-colors underline-animation text-sm underline-animation"
                  >
                    Documentation
                  </a>
                </li>
                <li>
                  <a
                    href="#"
                    className="text-text-secondary hover:text-text-primary transition-colors underline-animation text-sm underline-animation"
                  >
                    API Reference
                  </a>
                </li>
                <li>
                  <a
                    href="#"
                    className="text-text-secondary hover:text-text-primary transition-colors underline-animation text-sm underline-animation"
                  >
                    Guides
                  </a>
                </li>
                <li>
                  <a
                    href="/blog"
                    className="text-text-secondary hover:text-text-primary transition-colors underline-animation text-sm underline-animation"
                  >
                    Blog
                  </a>
                </li>
                <li>
                  <a
                    href="/faq"
                    className="text-text-secondary hover:text-text-primary transition-colors underline-animation text-sm underline-animation"
                  >
                    FAQ
                  </a>
                </li>
                <li>
                  <a
                    href="#"
                    className="text-text-secondary hover:text-text-primary transition-colors underline-animation text-sm underline-animation"
                  >
                    Community
                  </a>
                </li>
              </ul>
            </div>

            {/* Support */}
            <div className="lg:col-span-2">
              <h3 className="text-text-primary font-semibold mb-4">Support</h3>
              <ul className="space-y-3">
                <li>
                  <a
                    href="#"
                    className="text-text-secondary hover:text-text-primary transition-colors underline-animation text-sm underline-animation"
                  >
                    Help Center
                  </a>
                </li>
                <li>
                  <a
                    href="/contact"
                    className="text-text-secondary hover:text-text-primary transition-colors underline-animation text-sm underline-animation"
                  >
                    Contact Us
                  </a>
                </li>
                <li>
                  <a
                    href="#"
                    className="text-text-secondary hover:text-text-primary transition-colors underline-animation text-sm underline-animation"
                  >
                    Status Page
                  </a>
                </li>
                <li>
                  <a
                    href="#"
                    className="text-text-secondary hover:text-text-primary transition-colors underline-animation text-sm underline-animation"
                  >
                    System Status
                  </a>
                </li>
                <li>
                  <a
                    href="/feedback"
                    className="text-text-secondary hover:text-text-primary transition-colors underline-animation text-sm underline-animation"
                  >
                    Feedback
                  </a>
                </li>
              </ul>
            </div>

            {/* Newsletter */}
            <div className="lg:col-span-2">
              <h3 className="text-text-primary font-semibold mb-4">Stay Updated</h3>
              <p className="text-text-secondary text-sm mb-4">
                Get the latest updates and news about FunctionFly.
              </p>
              <div className="space-y-3">
                <div className="flex gap-2">
                  <input
                    type="email"
                    placeholder="Enter your email"
                    className="newsletter-input flex-1 px-3 py-2 bg-bg-secondary border border-border-subtle rounded-lg text-text-primary placeholder-text-muted text-sm focus:outline-none focus:border-brand-500/50 focus:ring-1 focus:ring-brand-500/20"
                  />
                  <motion.button
                    className="px-4 py-2 bg-gradient-to-r from-brand-500 to-purple-500 rounded-lg text-white text-sm font-medium hover:shadow-lg hover:shadow-brand-500/25 transition-all duration-200 glow hover-lift"
                    whileHover={{ scale: 1.02 }}
                    whileTap={{ scale: 0.98 }}
                  >
                    <Mail className="w-4 h-4" />
                  </motion.button>
                </div>
                <p className="text-text-muted text-xs">
                  We respect your privacy. Unsubscribe at any time.
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* Bottom Bar */}
        <div className="border-t border-border-subtle">
          <div className="max-w-7xl mx-auto px-4 lg:px-6 py-6">
            <div className="flex flex-col md:flex-row items-center justify-between gap-4">
              <div className="flex items-center gap-2 text-sm text-text-muted">
                <span>© 2026 FunctionFly. Made with</span>
                <Heart className="w-4 h-4 text-error fill-current" />
                <span>for indie developers.</span>
              </div>
              <div className="flex items-center gap-6 text-sm">
                <a
                  href="#"
                  className="text-text-secondary hover:text-text-primary transition-colors underline-animation"
                >
                  Terms of Service
                </a>
                <a
                  href="/privacy"
                  className="text-text-secondary hover:text-text-primary transition-colors underline-animation"
                >
                  Privacy Policy
                </a>
                <a
                  href="#"
                  className="text-text-secondary hover:text-text-primary transition-colors underline-animation"
                >
                  Cookie Policy
                </a>
              </div>
            </div>
          </div>
        </div>

        {/* Back to Top Button */}
        <motion.button
          onClick={() => window.scrollTo({ top: 0, behavior: "smooth" })}
          className="fixed bottom-8 right-8 w-12 h-12 bg-gradient-to-r from-brand-500 to-purple-500 rounded-full flex items-center justify-center text-white shadow-lg hover:shadow-xl hover:shadow-brand-500/25 transition-all duration-200 z-50 glow hover-lift"
          whileHover={{ scale: 1.05 }}
          whileTap={{ scale: 0.95 }}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
        >
          <ArrowUp className="w-5 h-5" />
        </motion.button>
      </div>
    </footer>
  );
}
