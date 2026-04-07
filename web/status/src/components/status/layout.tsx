import { Logo } from "@/components/Logo";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { format, formatDistanceToNow } from "date-fns";
import { AnimatePresence, motion } from "framer-motion";
import {
  Bell,
  CheckCircle,
  Clock,
  ExternalLink,
  Globe,
  Mail,
  RefreshCw,
  Search,
} from "lucide-react";
import { useEffect, useState } from "react";
import { StatusOrbital } from "./backgrounds";
import { HeroSkeleton } from "./skeletons";
import type { HeaderProps, HeroStatusProps } from "./types";

export function Header({
  onRefresh,
  isRefreshing,
  lastUpdated,
  isRealtimeConnected,
}: HeaderProps) {
  const [isScrolled, setIsScrolled] = useState(false);

  useEffect(() => {
    const handleScroll = () => setIsScrolled(window.scrollY > 50);
    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  return (
    <motion.header
      className={cn(
        "fixed top-0 left-0 right-0 z-50 transition-all duration-300",
        isScrolled
          ? "bg-bg-glass-strong/90 backdrop-blur-xl border-b border-border-subtle shadow-lg"
          : "bg-transparent",
      )}
      initial={{ y: -100 }}
      animate={{ y: 0 }}
      transition={{ duration: 0.5 }}
    >
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex h-16 items-center justify-between">
          <div className="flex items-center gap-3">
            <Logo size="sm" showText={true} animated={true} />
            <span className="text-text-muted">/</span>
            <span className="font-semibold text-text-primary">Status</span>
          </div>

          <div className="flex items-center gap-4">
            <div className="hidden md:flex items-center gap-2 text-sm text-text-muted">
              <span
                className={cn(
                  "w-2 h-2 rounded-full animate-pulse",
                  isRealtimeConnected ? "bg-emerald-500" : "bg-amber-500",
                )}
              />
              <span>
                {isRealtimeConnected ? "Live Updates" : "Monitoring Active"}
              </span>
            </div>

            <Button
              variant="ghost"
              size="sm"
              onClick={onRefresh}
              className="group"
            >
              <RefreshCw
                className={cn(
                  "w-4 h-4 transition-transform duration-500",
                  isRefreshing && "animate-spin",
                )}
              />
              <span className="hidden sm:inline ml-2">
                {format(lastUpdated, "HH:mm:ss")}
              </span>
            </Button>

            <Button variant="ghost" size="icon" className="hidden sm:flex">
              <Search className="w-4 h-4 text-text-secondary" />
            </Button>

            <Button variant="ghost" size="icon" className="hidden sm:flex">
              <Bell className="w-4 h-4 text-text-secondary" />
            </Button>
          </div>
        </div>
      </div>
    </motion.header>
  );
}

export function HeroStatus({
  overallStatus,
  isLoading,
  lastUpdated,
}: HeroStatusProps) {
  const isOperational = overallStatus === "operational";

  if (isLoading) {
    return <HeroSkeleton />;
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.6, ease: [0.4, 0, 0.2, 1] }}
      className={cn(
        "relative overflow-hidden rounded-2xl p-8 md:p-12",
        "border backdrop-blur-xl",
        isOperational
          ? "bg-gradient-to-br from-emerald-500/10 via-emerald-500/5 to-transparent border-emerald-500/20"
          : "bg-gradient-to-br from-amber-500/10 via-amber-500/5 to-transparent border-amber-500/20",
      )}
    >
      <div
        className={cn(
          "absolute inset-0 opacity-30",
          isOperational ? "animate-pulse" : "animate-pulse",
        )}
      >
        <div
          className={cn(
            "absolute inset-0",
            isOperational
              ? "bg-gradient-to-r from-emerald-500/20 via-emerald-400/10 to-transparent"
              : "bg-gradient-to-r from-amber-500/20 via-amber-400/10 to-transparent",
          )}
        />
      </div>

      <div className="relative z-10 flex flex-col md:flex-row items-center gap-6 md:gap-8">
        <StatusOrbital status={overallStatus} size="xl" />

        <div className="flex-1 text-center md:text-left">
          <motion.h2
            className={cn(
              "text-3xl md:text-4xl font-bold mb-2",
              isOperational ? "text-emerald-400" : "text-amber-400",
            )}
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: 0.2 }}
          >
            {isOperational
              ? "All Systems Operational"
              : "Partial Service Disruption"}
          </motion.h2>
          <motion.p
            className="text-text-secondary text-lg"
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: 0.3 }}
          >
            FunctionFly services are running normally. Our team is monitoring
            24/7 for any issues.
          </motion.p>
        </div>

        <motion.div
          className="flex items-center gap-2 text-sm text-text-muted"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.4 }}
        >
          <Clock className="w-4 h-4" />
          <span>
            Updated {formatDistanceToNow(lastUpdated, { addSuffix: true })}
          </span>
        </motion.div>
      </div>

      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <motion.div
          className="absolute inset-0 -translate-x-full"
          style={{
            background:
              "linear-gradient(90deg, transparent, rgba(255,255,255,0.05), transparent)",
          }}
          animate={{ x: ["-100%", "200%"] }}
          transition={{
            duration: 3,
            repeat: Infinity,
            ease: "linear",
            repeatDelay: 5,
          }}
        />
      </div>
    </motion.div>
  );
}

export function Footer() {
  const links = [
    {
      label: "Documentation",
      href: "https://docs.functionfly.com",
      icon: ExternalLink,
    },
    {
      label: "Support",
      href: "mailto:support@functionfly.com",
      icon: ExternalLink,
    },
    {
      label: "API Status",
      href: "https://api.functionfly.com/health",
      icon: ExternalLink,
    },
    {
      label: "GitHub",
      href: "https://github.com/functionfly",
      icon: ExternalLink,
    },
  ];

  return (
    <footer className="border-t border-border-subtle bg-bg-secondary/50 mt-16">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          <div>
            <Logo size="md" showText={true} />
            <p className="mt-4 text-sm text-text-secondary">
              Real-time status monitoring for FunctionFly platform services.
              Built with reliability at the edge.
            </p>
          </div>

          <div>
            <h4 className="font-semibold text-text-primary mb-4">Resources</h4>
            <ul className="space-y-2">
              {links.map((link) => (
                <li key={link.label}>
                  <a
                    href={link.href}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-sm text-text-secondary hover:text-text-primary transition-colors inline-flex items-center gap-2 group"
                  >
                    {link.label}
                    <link.icon className="w-3 h-3 opacity-0 group-hover:opacity-100 transition-opacity" />
                  </a>
                </li>
              ))}
            </ul>
          </div>

          <div>
            <h4 className="font-semibold text-text-primary mb-4">Legal</h4>
            <ul className="space-y-2">
              <li>
                <a
                  href="https://functionfly.com/privacy"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-text-secondary hover:text-text-primary transition-colors inline-flex items-center gap-2 group"
                >
                  Privacy Policy
                  <ExternalLink className="w-3 h-3 opacity-0 group-hover:opacity-100 transition-opacity" />
                </a>
              </li>
              <li>
                <a
                  href="https://functionfly.com/terms"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-text-secondary hover:text-text-primary transition-colors inline-flex items-center gap-2 group"
                >
                  Terms of Service
                  <ExternalLink className="w-3 h-3 opacity-0 group-hover:opacity-100 transition-opacity" />
                </a>
              </li>
              <li>
                <a
                  href="https://functionfly.com/security"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-text-secondary hover:text-text-primary transition-colors inline-flex items-center gap-2 group"
                >
                  Security
                  <ExternalLink className="w-3 h-3 opacity-0 group-hover:opacity-100 transition-opacity" />
                </a>
              </li>
            </ul>
          </div>
        </div>

        <div className="mt-12 pt-8 border-t border-border-subtle flex flex-col md:flex-row items-center justify-between gap-4">
          <p className="text-sm text-text-muted">
            © {new Date().getFullYear()} FunctionFly. All rights reserved.
          </p>
          <div className="flex items-center gap-2 text-sm text-text-muted">
            <Globe className="w-4 h-4" />
            <span>Global Infrastructure</span>
          </div>
        </div>
      </div>
    </footer>
  );
}

export function SubscribeSection() {
  const [showEmailSubscription, setShowEmailSubscription] = useState(false);
  const [email, setEmail] = useState("");
  const [subscribed, setSubscribed] = useState(false);

  return (
    <motion.section
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: 0.8 }}
      className={cn(
        "relative overflow-hidden rounded-2xl p-8",
        "bg-gradient-to-br from-brand-500/10 via-purple-500/5 to-transparent",
        "border border-brand-500/20",
      )}
    >
      <div className="relative z-10 flex flex-col md:flex-row items-center justify-between gap-6">
        <div className="text-center md:text-left">
          <h3 className="text-xl font-semibold text-text-primary mb-2">
            Subscribe to Status Updates
          </h3>
          <p className="text-text-secondary">
            Get notified when there are service disruptions or maintenance
            windows.
          </p>
        </div>
        <div className="flex gap-3">
          <Button
            variant="outline"
            className="border-border-default hover:bg-bg-hover"
            onClick={() => setShowEmailSubscription(!showEmailSubscription)}
          >
            <Bell className="w-4 h-4 mr-2" />
            Email Alerts
          </Button>
          <Button
            className="bg-brand-500 hover:bg-brand-600 text-white shadow-lg shadow-brand-500/25"
            onClick={() => window.open("/api/v1/status/rss", "_blank")}
          >
            <ExternalLink className="w-4 h-4 mr-2" />
            RSS Feed
          </Button>
        </div>
      </div>

      <AnimatePresence>
        {showEmailSubscription && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.3 }}
            className="overflow-hidden"
          >
            <div className="mt-6 pt-6 border-t border-border-subtle">
              {subscribed ? (
                <div className="flex items-center gap-3 text-emerald-400">
                  <CheckCircle className="w-5 h-5" />
                  <span>Thanks! You'll receive status updates at {email}</span>
                </div>
              ) : (
                <div className="flex flex-col sm:flex-row items-center gap-3">
                  <input
                    type="email"
                    placeholder="Enter your email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="flex-1 w-full px-4 py-2 rounded-lg bg-bg-secondary border border-border-subtle text-text-primary placeholder:text-text-muted focus:outline-none focus:ring-2 focus:ring-brand-500"
                  />
                  <Button
                    className="w-full sm:w-auto bg-brand-500 hover:bg-brand-600 text-white"
                    onClick={() => {
                      if (email.includes("@")) {
                        setSubscribed(true);
                        setTimeout(() => {
                          setShowEmailSubscription(false);
                          setSubscribed(false);
                          setEmail("");
                        }, 3000);
                      }
                    }}
                    disabled={!email.includes("@")}
                  >
                    <Mail className="w-4 h-4 mr-2" />
                    Subscribe
                  </Button>
                </div>
              )}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </motion.section>
  );
}
