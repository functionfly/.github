import { useEffect, useState, useCallback } from "react";
import { Link, useNavigate, useLocation } from "react-router-dom";
import { motion, AnimatePresence } from "framer-motion";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Home,
  ArrowLeft,
  Search,
  Compass,
  FunctionSquare,
  MessageSquare,
  Command,
  Ghost,
  HelpCircle,
  Rocket,
  Binary,
  Terminal,
  Clock,
  Activity,
  AlertTriangle,
  Eye,
  Zap
} from "lucide-react";
import { SpotlightCard } from "@/components/ui/SpotlightCard";
import { ParticleBackground } from "@/components/ui/ParticleBackground";
import { cn } from "@/lib/utils";
import "./styles.css";

// Animated glitch text component
function GlitchText({ text }: { text: string }) {
  return (
    <div className="relative inline-block glitch-text" data-text={text}>
      <motion.span 
        className="text-9xl font-black text-brand-500 relative z-10 glow-text"
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
      >
        {text}
      </motion.span>
      <motion.span 
        className="absolute top-0 left-0 text-9xl font-black text-red-500/50 -z-10"
        animate={{ 
          x: [0, -4, 4, -2, 0],
          opacity: [1, 0.5, 1, 0.5, 1]
        }}
        transition={{ 
          duration: 0.5,
          repeat: Infinity,
          repeatType: "reverse",
          times: [0, 0.2, 0.4, 0.6, 1]
        }}
      >
        {text}
      </motion.span>
      <motion.span 
        className="absolute top-0 left-0 text-9xl font-black text-cyan-500/50 -z-10"
        animate={{ 
          x: [0, 4, -4, 2, 0],
          opacity: [1, 0.5, 1, 0.5, 1]
        }}
        transition={{ 
          duration: 0.5,
          repeat: Infinity,
          repeatType: "reverse",
          times: [0, 0.2, 0.4, 0.6, 1],
          delay: 0.1
        }}
      >
        {text}
      </motion.span>
    </div>
  );
}

// Floating ghost animation component
function FloatingGhost() {
  return (
    <motion.div
      className="relative ghost-float"
      animate={{
        y: [0, -20, 0],
        rotate: [0, 5, -5, 0],
      }}
      transition={{
        duration: 4,
        repeat: Infinity,
        ease: "easeInOut"
      }}
    >
      <Ghost className="w-32 h-32 text-brand-500/30" />
      <motion.div
        className="absolute -bottom-2 left-1/2 -translate-x-1/2 w-20 h-4 bg-brand-500/20 rounded-full blur-xl ghost-shadow"
        animate={{
          scale: [1, 0.8, 1],
          opacity: [0.5, 0.3, 0.5]
        }}
        transition={{
          duration: 4,
          repeat: Infinity,
          ease: "easeInOut"
        }}
      />
    </motion.div>
  );
}

// Terminal typing effect
function TerminalText() {
  const [displayText, setDisplayText] = useState("");
  const fullText = "$ ERROR: Page not found in the void...\n$ Searching alternate dimensions...\n$ Nope, definitely not here.";
  
  useEffect(() => {
    let index = 0;
    const timer = setInterval(() => {
      if (index < fullText.length) {
        setDisplayText(fullText.slice(0, index + 1));
        index++;
      } else {
        clearInterval(timer);
      }
    }, 50);
    return () => clearInterval(timer);
  }, []);

  return (
    <div className="font-mono text-xs text-text-secondary/60 bg-black/20 p-4 rounded-lg border border-border/50 max-w-md mx-auto glass">
      <div className="flex items-center gap-2 mb-2 border-b border-border/50 pb-2">
        <div className="w-3 h-3 rounded-full bg-red-500/50" />
        <div className="w-3 h-3 rounded-full bg-yellow-500/50" />
        <div className="w-3 h-3 rounded-full bg-green-500/50" />
        <span className="text-xs ml-2">terminal</span>
      </div>
      <pre className="whitespace-pre-wrap">{displayText}</pre>
      <span className="terminal-cursor" />
    </div>
  );
}

// Quick action card
function QuickActionCard({ 
  icon: Icon, 
  title, 
  description, 
  to, 
  color,
  delay 
}: { 
  icon: React.ComponentType<{ className?: string }>;
  title: string;
  description: string;
  to: string;
  color: string;
  delay: number;
}) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay, duration: 0.4 }}
    >
      <Link to={to}>
        <SpotlightCard 
          className="h-full cursor-pointer group holographic-card spotlight-hover"
          spotlightColor={`rgba(var(--${color}), 0.15)`}
        >
          <div className="flex items-start gap-4 relative z-10">
            <div className={cn(
              "p-3 rounded-xl transition-colors pulse-ring",
              `bg-${color}/10 group-hover:bg-${color}/20`
            )}>
              <Icon className={cn("w-6 h-6", `text-${color}`)} />
            </div>
            <div>
              <h3 className="font-semibold text-text-primary group-hover:text-brand-400 transition-colors">
                {title}
              </h3>
              <p className="text-sm text-text-secondary mt-1">
                {description}
              </p>
            </div>
          </div>
        </SpotlightCard>
      </Link>
    </motion.div>
  );
}

// Search suggestions based on common paths
const SEARCH_SUGGESTIONS = [
  "/overview",
  "/functions",
  "/registry",
  "/agents",
  "/docs",
  "/settings",
  "/profile",
  "/wallet"
];

const QUICK_ACTIONS = [
  {
    icon: FunctionSquare,
    title: "Functions",
    description: "Browse and discover serverless functions",
    to: "/functions",
    color: "brand-500"
  },
  {
    icon: Command,
    title: "Registry",
    description: "Explore the function registry",
    to: "/registry",
    color: "info"
  },
  {
    icon: Zap,
    title: "Agents",
    description: "Manage your AI agents",
    to: "/agents",
    color: "warning"
  },
  {
    icon: MessageSquare,
    title: "Messages",
    description: "Check your conversations",
    to: "/conversations",
    color: "success"
  }
];

export function NotFoundPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchQuery, setSearchQuery] = useState("");
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [randomQuote, setRandomQuote] = useState(0);

  const quotes = [
    "Looks like this function returned null",
    "This page has been garbage collected",
    "404: Page not in cache",
    "The route you're looking for is undefined",
    "This page has been moved to /dev/null",
    "Exception: PageNotFoundException",
    "HTTP/1.1 404 Not Found",
    "console.log('page'): undefined"
  ];

  useEffect(() => {
    setRandomQuote(Math.floor(Math.random() * quotes.length));
  }, []);

  const filteredSuggestions = SEARCH_SUGGESTIONS.filter(path => 
    path.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (searchQuery.trim()) {
      // Try to navigate if it looks like a path
      if (searchQuery.startsWith('/')) {
        navigate(searchQuery);
      } else {
        // Search in registry
        navigate(`/registry?q=${encodeURIComponent(searchQuery)}`);
      }
    }
  };

  return (
    <div className="min-h-screen flex flex-col bg-bg-primary relative overflow-hidden scanlines crt-flicker noise-overlay">
      {/* Background Effects */}
      <div className="absolute inset-0 opacity-30">
        <ParticleBackground 
          particleCount={30}
          color="rgba(var(--brand-500), 0.4)"
          speed={0.5}
        />
      </div>
      
      {/* Matrix Rain Background */}
      <div className="matrix-bg" />
      
      {/* Grid Pattern Overlay */}
      <div 
        className="absolute inset-0 opacity-[0.03] grid-lines"
        style={{
          backgroundImage: `linear-gradient(to right, var(--border-subtle) 1px, transparent 1px),
                           linear-gradient(to bottom, var(--border-subtle) 1px, transparent 1px)`,
          backgroundSize: '40px 40px'
        }}
      />

      <main className="flex-1 flex flex-col items-center justify-center p-4 relative z-10">
        {/* Main Content */}
        <div className="text-center space-y-8 max-w-4xl w-full">
          
          {/* 404 Header with Ghost */}
          <div className="relative">
            <div className="absolute -top-20 left-1/2 -translate-x-1/2">
              <FloatingGhost />
            </div>
            <GlitchText text="404" />
            <motion.p 
              className="text-lg text-text-secondary mt-4 font-mono"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: 0.5 }}
            >
              {quotes[randomQuote]}
            </motion.p>
          </div>

          {/* Terminal Output */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.7 }}
          >
            <TerminalText />
          </motion.div>

          {/* Search Box */}
          <motion.form 
            onSubmit={handleSearch}
            className="max-w-md mx-auto relative"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.9 }}
          >
            <div className="relative neon-border">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-text-secondary" />
              <Input
                type="text"
                placeholder="Search for pages or functions..."
                value={searchQuery}
                onChange={(e) => {
                  setSearchQuery(e.target.value);
                  setShowSuggestions(true);
                }}
                onBlur={() => setTimeout(() => setShowSuggestions(false), 200)}
                className="pl-10 h-12 bg-bg-secondary/50 border-border/50 focus:border-brand-500 glass"
              />
              <Button 
                type="submit"
                size="sm"
                className="absolute right-2 top-1/2 -translate-y-1/2"
              >
                Search
              </Button>
            </div>
            
            {/* Search Suggestions */}
            <AnimatePresence>
              {showSuggestions && searchQuery && filteredSuggestions.length > 0 && (
                <motion.div
                  initial={{ opacity: 0, y: -10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -10 }}
                  className="absolute top-full left-0 right-0 mt-2 bg-bg-secondary border border-border rounded-lg shadow-lg overflow-hidden z-50"
                >
                  {filteredSuggestions.map((suggestion) => (
                    <button
                      key={suggestion}
                      type="button"
                      onClick={() => navigate(suggestion)}
                      className="w-full px-4 py-2 text-left text-sm text-text-secondary hover:bg-brand-500/10 hover:text-text-primary transition-colors flex items-center gap-2"
                    >
                      <Compass className="w-4 h-4" />
                      {suggestion}
                    </button>
                  ))}
                </motion.div>
              )}
            </AnimatePresence>
          </motion.form>

          {/* Navigation Buttons */}
          <motion.div 
            className="flex flex-col sm:flex-row gap-3 justify-center"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 1.0 }}
          >
            <Button 
              variant="outline" 
              className="gap-2 holographic-card" 
              onClick={() => navigate(-1)}
              size="lg"
            >
              <ArrowLeft className="w-4 h-4" />
              Go Back
            </Button>
            <Button 
              asChild 
              className="gap-2 neon-border"
              size="lg"
            >
              <Link to="/">
                <Home className="w-4 h-4" />
                Go Home
              </Link>
            </Button>
            <Button 
              variant="secondary"
              asChild 
              className="gap-2 holographic-card"
              size="lg"
            >
              <a href="/docs" target="_blank" rel="noopener noreferrer">
                <HelpCircle className="w-4 h-4" />
                Help & Docs
              </a>
            </Button>
          </motion.div>

          {/* Quick Actions Grid */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 1.2 }}
            className="mt-12"
          >
            <h2 className="text-sm font-medium text-text-secondary mb-4 uppercase tracking-wider fade-in-up stagger-1">
              Popular Destinations
            </h2>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
              {QUICK_ACTIONS.map((action, index) => (
                <QuickActionCard
                  key={action.title}
                  {...action}
                  delay={1.3 + index * 0.1}
                />
              ))}
            </div>
          </motion.div>

          {/* Fun Zone */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 1.6 }}
            className="mt-16 pt-8 border-t border-border/30"
          >
            {/* Section Header */}
            <motion.div 
              className="flex items-center justify-center gap-3 mb-8"
              initial={{ scale: 0.9 }}
              animate={{ scale: 1 }}
              transition={{ delay: 1.7, type: "spring" }}
            >
              <div className="h-px w-12 bg-gradient-to-r from-transparent via-border to-border" />
              <div className="flex items-center gap-2 px-4 py-2 rounded-full bg-bg-secondary/50 border border-border/50">
                <Rocket className="w-4 h-4 text-brand-400" />
                <span className="text-sm font-medium text-text-primary">While you're here...</span>
                <Rocket className="w-4 h-4 text-brand-400" />
              </div>
              <div className="h-px w-12 bg-gradient-to-l from-transparent via-border to-border" />
            </motion.div>

            {/* Cards Grid */}
            <div className="max-w-xl mx-auto">
              {/* Lost Path Stats Card */}
              <motion.div 
                className="group"
                whileHover={{ scale: 1.02 }}
                transition={{ type: "spring", stiffness: 300 }}
              >
                <SpotlightCard 
                  className="h-full holographic-card spotlight-hover"
                  spotlightColor="rgba(var(--info), 0.1)"
                >
                  {/* Card Header */}
                  <div className="flex items-center gap-3 mb-4">
                    <div className="p-2.5 rounded-xl bg-info/10 border border-info/20">
                      <Terminal className="w-5 h-5 text-info" />
                    </div>
                    <div>
                      <h3 className="font-semibold text-text-primary flex items-center gap-2">
                        Lost Path Stats
                        <span className="text-xs font-normal text-text-secondary px-2 py-0.5 rounded-full bg-info/10 border border-info/20">
                          Debug Info
                        </span>
                      </h3>
                      <p className="text-xs text-text-secondary">Error diagnostics & metadata</p>
                    </div>
                  </div>

                  {/* Terminal-style Stats Display */}
                  <div className="relative bg-black/40 rounded-xl border border-border/50 overflow-hidden">
                    {/* Terminal Header */}
                    <div className="flex items-center gap-2 px-3 py-2 bg-bg-secondary/30 border-b border-border/30">
                      <div className="flex gap-1.5">
                        <div className="w-2.5 h-2.5 rounded-full bg-red-500/60" />
                        <div className="w-2.5 h-2.5 rounded-full bg-yellow-500/60" />
                        <div className="w-2.5 h-2.5 rounded-full bg-green-500/60" />
                      </div>
                      <span className="text-[10px] text-text-secondary/60 font-mono ml-2">debug.log — 404 error analysis</span>
                    </div>
                    
                    {/* Terminal Body */}
                    <div className="p-4 space-y-3 font-mono text-xs">
                      {/* Path */}
                      <div className="group/path">
                        <div className="flex items-center gap-2 text-text-secondary/70 mb-1">
                          <Compass className="w-3 h-3" />
                          <span>Current Path</span>
                        </div>
                        <div className="flex items-center gap-2">
                          <span className="text-success">→</span>
                          <code 
                            className="text-brand-400 bg-brand-500/10 px-2 py-1.5 rounded border border-brand-500/20 text-xs font-mono break-all flex-1 min-w-0"
                            title={location.pathname}
                          >
                            {location.pathname}
                          </code>
                          <button 
                            onClick={() => navigator.clipboard.writeText(location.pathname)}
                            className="opacity-70 hover:opacity-100 transition-opacity text-text-secondary hover:text-brand-400 flex-shrink-0 p-1 rounded hover:bg-brand-500/10"
                            title="Copy path"
                          >
                            <span className="text-[10px] font-mono">copy</span>
                          </button>
                        </div>
                      </div>

                      {/* Error Code */}
                      <div className="flex items-center justify-between py-2 border-y border-border/20 border-dashed">
                        <div className="flex items-center gap-2 text-text-secondary/70">
                          <AlertTriangle className="w-3 h-3" />
                          <span>Error Code</span>
                        </div>
                        <div className="flex items-center gap-2">
                          <span className="text-error animate-pulse">●</span>
                          <span className="text-error font-bold">404_NOT_FOUND</span>
                        </div>
                      </div>

                      {/* Timestamp */}
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2 text-text-secondary/70">
                          <Clock className="w-3 h-3" />
                          <span>Timestamp</span>
                        </div>
                        <code className="text-text-primary/80 text-[10px]">
                          {new Date().toISOString()}
                        </code>
                      </div>

                      {/* Status with visual indicator */}
                      <div className="pt-2 mt-2 border-t border-border/20">
                        <div className="flex items-center justify-between mb-2">
                          <div className="flex items-center gap-2 text-text-secondary/70">
                            <Activity className="w-3 h-3" />
                            <span>Status</span>
                          </div>
                          <span className="text-warning flex items-center gap-1.5">
                            <span className="w-1.5 h-1.5 rounded-full bg-warning animate-pulse" />
                            Lost in the void
                          </span>
                        </div>
                        {/* Progress bar visual */}
                        <div className="h-1.5 w-full bg-bg-secondary rounded-full overflow-hidden">
                          <motion.div 
                            className="h-full bg-gradient-to-r from-error via-warning to-info"
                            initial={{ width: "0%" }}
                            animate={{ width: "100%" }}
                            transition={{ duration: 2, delay: 2 }}
                          />
                        </div>
                        <div className="flex justify-between text-[10px] text-text-secondary/50 mt-1">
                          <span>Searching...</span>
                          <span>∞</span>
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Footer hint */}
                  <div className="mt-3 flex items-center justify-between text-xs">
                    <span className="flex items-center gap-1.5 text-text-secondary/60">
                      <Eye className="w-3 h-3" />
                      Only visible to you
                    </span>
                    <span className="font-mono text-[10px] text-text-secondary/40">
                      trace_id: {Math.random().toString(36).substring(2, 10)}
                    </span>
                  </div>
                </SpotlightCard>
              </motion.div>
            </div>
          </motion.div>

          {/* Binary Art Footer */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 1.8 }}
            className="pt-8 text-center"
          >
            <div className="flex items-center justify-center gap-4 text-text-secondary/40 text-xs font-mono binary-stream">
              <Binary className="w-4 h-4" />
              <span className="typing-text">01000100 01100101 01100001 01100100</span>
              <span className="text-brand-500/40 glow-text">|</span>
              <span className="typing-text" style={{ animationDelay: '0.5s' }}>01100101 01101110 01100100</span>
              <Terminal className="w-4 h-4" />
            </div>
          </motion.div>
        </div>
      </main>
    </div>
  );
}

export default NotFoundPage;
