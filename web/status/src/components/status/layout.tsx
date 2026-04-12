import { Logo } from "@/components/Logo";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
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
  Command,
  Activity,
} from "lucide-react";
import { useEffect, useState, useCallback } from "react";
import { StatusOrbital } from "./backgrounds";
import { HeroSkeleton } from "./skeletons";
import type { HeaderProps, HeroStatusProps } from "./types";

// Quick action shortcuts
const QUICK_ACTIONS = [
  { key: 'r', label: 'Refresh Status', icon: RefreshCw },
  { key: 's', label: 'Search', icon: Search },
];

export function Header({
  onRefresh,
  isRefreshing,
  lastUpdated,
  isRealtimeConnected,
}: HeaderProps) {
  const [isScrolled, setIsScrolled] = useState(false);
  const [showCommandPalette, setShowCommandPalette] = useState(false);

  useEffect(() => {
    const handleScroll = () => setIsScrolled(window.scrollY > 50);
    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  // Keyboard shortcuts
  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    // Command palette: Cmd/Ctrl + K
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault();
      setShowCommandPalette(true);
    }

    // Quick refresh: Cmd/Ctrl + R
    if ((e.metaKey || e.ctrlKey) && e.key === 'r') {
      e.preventDefault();
      onRefresh();
    }
  }, [onRefresh]);

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);

  return (
    <TooltipProvider delayDuration={0}>
      <>
        <motion.header
          className={cn(
            "fixed top-0 left-0 right-0 z-50 transition-all duration-300",
            isScrolled
              ? "bg-aviation-bg-primary/95 backdrop-blur-xl border-b border-aviation-border-panel shadow-lg"
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
                <span className="text-aviation-text-muted">/</span>
                <span className="font-semibold text-aviation-text-primary">Status</span>
              </div>

              <div className="flex items-center gap-2 sm:gap-4">
                {/* Command Palette Trigger */}
                <Tooltip>
                  <TooltipTrigger asChild>
                    <button
                      onClick={() => setShowCommandPalette(true)}
                      className="hidden md:flex items-center gap-2 px-3 py-1.5 text-sm text-aviation-text-muted bg-aviation-bg-instrument/50 border border-aviation-border-instrument rounded-lg hover:text-aviation-text-primary hover:border-aviation-amber/30 transition-all"
                    >
                      <Command className="w-3.5 h-3.5" />
                      <span>Quick Actions</span>
                      <kbd className="text-[10px] font-mono text-aviation-text-dim bg-aviation-bg-instrument px-1.5 py-0.5 rounded">
                        ⌘K
                      </kbd>
                    </button>
                  </TooltipTrigger>
                  <TooltipContent side="bottom">
                    <p>Command Palette</p>
                  </TooltipContent>
                </Tooltip>

                {/* Live Status Indicator */}
                <div className="hidden md:flex items-center gap-2 text-sm text-aviation-text-secondary">
                  <span
                    className={cn(
                      "w-2 h-2 rounded-full animate-pulse",
                      isRealtimeConnected ? "bg-aviation-green" : "bg-aviation-yellow",
                    )}
                  />
                  <span>
                    {isRealtimeConnected ? "Live Updates" : "Monitoring Active"}
                  </span>
                </div>

                {/* Refresh Button */}
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={onRefresh}
                      className="group text-aviation-text-secondary hover:text-aviation-text-primary hover:bg-aviation-bg-instrument"
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
                  </TooltipTrigger>
                  <TooltipContent side="bottom">
                    <p>Refresh Status (⌘R)</p>
                  </TooltipContent>
                </Tooltip>

                {/* Search */}
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="hidden sm:flex text-aviation-text-secondary hover:text-aviation-text-primary hover:bg-aviation-bg-instrument"
                    >
                      <Search className="w-4 h-4" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent side="bottom">
                    <p>Search</p>
                  </TooltipContent>
                </Tooltip>

                {/* Notifications */}
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="hidden sm:flex text-aviation-text-secondary hover:text-aviation-text-primary hover:bg-aviation-bg-instrument"
                    >
                      <Bell className="w-4 h-4" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent side="bottom">
                    <p>Notifications</p>
                  </TooltipContent>
                </Tooltip>
              </div>
            </div>
          </div>
        </motion.header>

        {/* Command Palette Overlay */}
        <AnimatePresence>
          {showCommandPalette && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-start justify-center pt-[20vh]"
              onClick={() => setShowCommandPalette(false)}
            >
              <motion.div
                initial={{ opacity: 0, y: -20, scale: 0.95 }}
                animate={{ opacity: 1, y: 0, scale: 1 }}
                exit={{ opacity: 0, y: -20, scale: 0.95 }}
                transition={{ type: 'spring', stiffness: 300, damping: 30 }}
                className="w-full max-w-2xl mx-4 bg-aviation-bg-primary border border-aviation-border-panel rounded-xl shadow-2xl overflow-hidden"
                onClick={(e) => e.stopPropagation()}
              >
                {/* Search Input */}
                <div className="flex items-center gap-3 px-4 py-4 border-b border-aviation-border-panel">
                  <Command className="w-5 h-5 text-aviation-text-muted" />
                  <input
                    type="text"
                    placeholder="Search services, incidents, components..."
                    className="flex-1 text-base text-aviation-text-primary placeholder:text-aviation-text-dim bg-transparent focus:outline-none"
                    autoFocus
                  />
                  <kbd className="text-[10px] font-mono text-aviation-text-dim bg-aviation-bg-instrument px-2 py-1 rounded">
                    ESC
                  </kbd>
                </div>

                {/* Quick Actions */}
                <div className="p-2">
                  <p className="px-3 py-2 text-xs font-semibold text-aviation-text-muted uppercase tracking-wider">
                    Quick Actions
                  </p>
                  <div className="space-y-1">
                    {QUICK_ACTIONS.map((action) => (
                      <button
                        key={action.key}
                        onClick={() => {
                          setShowCommandPalette(false);
                          if (action.key === 'r') onRefresh();
                        }}
                        className="w-full flex items-center justify-between px-3 py-2.5 rounded-lg text-sm text-aviation-text-secondary hover:text-aviation-text-primary hover:bg-aviation-bg-instrument transition-colors"
                      >
                        <div className="flex items-center gap-3">
                          <action.icon className="w-4 h-4" />
                          <span>{action.label}</span>
                        </div>
                        <kbd className="text-[10px] font-mono text-aviation-text-dim bg-aviation-bg-instrument px-1.5 py-0.5 rounded">
                          ⌘{action.key.toUpperCase()}
                        </kbd>
                      </button>
                    ))}
                  </div>
                </div>

                {/* Footer */}
                <div className="px-4 py-3 bg-aviation-bg-secondary border-t border-aviation-border-panel text-xs text-aviation-text-muted">
                  <p className="flex items-center gap-2">
                    <span>Use</span>
                    <kbd className="font-mono bg-aviation-bg-instrument px-1 py-0.5 rounded">↑</kbd>
                    <kbd className="font-mono bg-aviation-bg-instrument px-1 py-0.5 rounded">↓</kbd>
                    <span>to navigate,</span>
                    <kbd className="font-mono bg-aviation-bg-instrument px-1 py-0.5 rounded">↵</kbd>
                    <span>to select</span>
                  </p>
                </div>
              </motion.div>
            </motion.div>
          )}
        </AnimatePresence>
      </>
    </TooltipProvider>
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
          ? "bg-aviation-green/10 border-aviation-green/30"
          : "bg-aviation-yellow/10 border-aviation-yellow/30",
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
              ? "bg-gradient-to-r from-aviation-green/20 via-aviation-green/10 to-transparent"
              : "bg-gradient-to-r from-aviation-yellow/20 via-aviation-yellow/10 to-transparent",
          )}
        />
      </div>

      <div className="relative z-10 flex flex-col md:flex-row items-center gap-6 md:gap-8">
        <StatusOrbital status={overallStatus} size="xl" />

        <div className="flex-1 text-center md:text-left">
          <motion.h2
            className={cn(
              "text-3xl md:text-4xl font-bold mb-2",
              isOperational ? "text-aviation-green" : "text-aviation-yellow",
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
            className="text-aviation-text-secondary text-lg"
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: 0.3 }}
          >
            FunctionFly services are running normally. Our team is monitoring
            24/7 for any issues.
          </motion.p>
        </div>

        <motion.div
          className="flex items-center gap-2 text-sm text-aviation-text-muted"
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
      icon: Activity,
    },
    {
      label: "GitHub",
      href: "https://github.com/functionfly",
      icon: ExternalLink,
    },
  ];

  return (
    <footer className="border-t border-aviation-border-panel bg-aviation-bg-secondary mt-16">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          <div>
            <Logo size="md" showText={true} />
            <p className="mt-4 text-sm text-aviation-text-secondary">
              Real-time status monitoring for FunctionFly platform services.
              Built with reliability at the edge.
            </p>
          </div>

          <div>
            <h4 className="font-semibold text-aviation-text-primary mb-4">Resources</h4>
            <ul className="space-y-2">
              {links.map((link) => (
                <li key={link.label}>
                  <a
                    href={link.href}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-sm text-aviation-text-secondary hover:text-aviation-text-primary transition-colors inline-flex items-center gap-2 group"
                  >
                    {link.label}
                    <link.icon className="w-3 h-3 opacity-0 group-hover:opacity-100 transition-opacity" />
                  </a>
                </li>
              ))}
            </ul>
          </div>

          <div>
            <h4 className="font-semibold text-aviation-text-primary mb-4">Legal</h4>
            <ul className="space-y-2">
              <li>
                <a
                  href="https://functionfly.com/privacy"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-aviation-text-secondary hover:text-aviation-text-primary transition-colors inline-flex items-center gap-2 group"
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
                  className="text-sm text-aviation-text-secondary hover:text-aviation-text-primary transition-colors inline-flex items-center gap-2 group"
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
                  className="text-sm text-aviation-text-secondary hover:text-aviation-text-primary transition-colors inline-flex items-center gap-2 group"
                >
                  Security
                  <ExternalLink className="w-3 h-3 opacity-0 group-hover:opacity-100 transition-opacity" />
                </a>
              </li>
            </ul>
          </div>
        </div>

        <div className="mt-12 pt-8 border-t border-aviation-border-panel flex flex-col md:flex-row items-center justify-between gap-4">
          <p className="text-sm text-aviation-text-muted">
            © {new Date().getFullYear()} FunctionFly. All rights reserved.
          </p>
          <div className="flex items-center gap-2 text-sm text-aviation-text-muted">
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
        "bg-aviation-amber/10 border border-aviation-amber/30",
      )}
    >
      <div className="relative z-10 flex flex-col md:flex-row items-center justify-between gap-6">
        <div className="text-center md:text-left">
          <h3 className="text-xl font-semibold text-aviation-text-primary mb-2">
            Subscribe to Status Updates
          </h3>
          <p className="text-aviation-text-secondary">
            Get notified when there are service disruptions or maintenance
            windows.
          </p>
        </div>
        <div className="flex gap-3">
          <Button
            variant="outline"
            className="border-aviation-border-instrument hover:bg-aviation-bg-instrument text-aviation-text-secondary"
            onClick={() => setShowEmailSubscription(!showEmailSubscription)}
          >
            <Bell className="w-4 h-4 mr-2" />
            Email Alerts
          </Button>
          <Button
            className="aviation-button-primary"
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
            <div className="mt-6 pt-6 border-t border-aviation-border-panel">
              {subscribed ? (
                <div className="flex items-center gap-3 text-aviation-green">
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
                    className="flex-1 w-full px-4 py-2 rounded-lg bg-aviation-bg-instrument border border-aviation-border-instrument text-aviation-text-primary placeholder:text-aviation-text-dim focus:outline-none focus:border-aviation-amber"
                  />
                  <Button
                    className="w-full sm:w-auto aviation-button-primary"
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
