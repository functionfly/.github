import { motion } from 'framer-motion';
import { Link, useParams } from 'react-router-dom';
import {
  AlertTriangle,
  ArrowLeft,
  Binary,
  Compass,
  Ghost,
  Home,
  Search,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { ParticleBackground } from '@/components/ui/ParticleBackground';
import { SpotlightCard } from '@/components/ui/SpotlightCard';
import { cn } from '@/lib/utils';
import './styles.css';

interface FunctionNotFoundProps {
  author?: string;
  name?: string;
  onSearch?: (query: string) => void;
}

export function FunctionNotFound({ author, name, onSearch }: FunctionNotFoundProps) {
  const { author: urlAuthor, name: urlName } = useParams<{ author: string; name: string }>();
  const fnAuthor = author || urlAuthor;
  const fnName = name || urlName;
  const functionPath = fnAuthor && fnName ? `${fnAuthor}/${fnName}` : null;

  const quotes = [
    'This function returned null',
    'Function garbage collected',
    '404: Function not in cache',
    'The function you seek is undefined',
    'This function moved to /dev/null',
    'Exception: FunctionNotFoundException',
  ];

  const randomQuote = quotes[Math.floor(Math.random() * quotes.length)];

  return (
    <div className="min-h-screen flex flex-col bg-bg-primary relative overflow-hidden scanlines noise-overlay">
      <div className="absolute inset-0 opacity-20">
        <ParticleBackground
          particleCount={20}
          color="rgba(var(--brand-500), 0.3)"
          speed={0.3}
        />
      </div>

      <main className="flex-1 flex flex-col items-center justify-center p-4 relative z-10">
        <div className="text-center space-y-8 max-w-2xl w-full">
          <motion.div
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.5 }}
            className="relative"
          >
            <div className="relative inline-block">
              <Ghost className="h-24 w-24 text-brand-500/30 mx-auto mb-4 animate-float" />
              <h1 className="text-8xl font-black text-brand-500 glow-text">404</h1>
            </div>
            <motion.p
              className="text-lg text-text-secondary mt-4 font-mono"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: 0.3 }}
            >
              {randomQuote}
            </motion.p>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.5 }}
            className="glass rounded-xl p-6 border border-border/50"
          >
            <div className="flex items-center gap-3 mb-4">
              <AlertTriangle className="h-5 w-5 text-warning" />
              <span className="font-semibold text-text-primary">Function Not Found</span>
            </div>
            {functionPath ? (
              <p className="text-sm text-text-secondary">
                The function <code className="text-brand-400 bg-brand-500/10 px-2 py-0.5 rounded">{functionPath}</code> could not be found in the registry.
              </p>
            ) : (
              <p className="text-sm text-text-secondary">
                The function you are looking for does not exist or has been removed.
              </p>
            )}
          </motion.div>

          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.7 }}
            className="flex flex-col sm:flex-row gap-3 justify-center"
          >
            <Button
              variant="outline"
              onClick={() => window.history.back()}
              className="gap-2"
              size="lg"
            >
              <ArrowLeft className="w-4 h-4" />
              Go Back
            </Button>
            <Button asChild className="gap-2" size="lg">
              <Link to="/registry">
                <Search className="w-4 h-4" />
                Browse Registry
              </Link>
            </Button>
            <Button asChild variant="secondary" className="gap-2" size="lg">
              <Link to="/">
                <Home className="w-4 h-4" />
                Go Home
              </Link>
            </Button>
          </motion.div>

          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.9 }}
            className="pt-6 border-t border-border/30"
          >
            <SpotlightCard
              className="text-left"
              spotlightColor="rgba(var(--info), 0.1)"
            >
              <div className="flex items-start gap-3">
                <Compass className="h-5 w-5 text-info mt-0.5" />
                <div>
                  <h3 className="font-semibold text-text-primary mb-1">Looking for something specific?</h3>
                  <p className="text-xs text-text-secondary mb-3">
                    Try searching the registry for alternative functions or browse by category.
                  </p>
                  <div className="flex flex-wrap gap-2">
                    <Link
                      to="/registry?category=api-tools"
                      className="text-xs px-3 py-1.5 rounded-lg bg-bg-tertiary/60 border border-border-subtle text-text-secondary hover:text-text-primary hover:bg-bg-hover transition-all"
                    >
                      API Tools
                    </Link>
                    <Link
                      to="/registry?category=utilities"
                      className="text-xs px-3 py-1.5 rounded-lg bg-bg-tertiary/60 border border-border-subtle text-text-secondary hover:text-text-primary hover:bg-bg-hover transition-all"
                    >
                      Utilities
                    </Link>
                    <Link
                      to="/registry?category=data-format"
                      className="text-xs px-3 py-1.5 rounded-lg bg-bg-tertiary/60 border border-border-subtle text-text-secondary hover:text-text-primary hover:bg-bg-hover transition-all"
                    >
                      Data Format
                    </Link>
                    <Link
                      to="/registry?sort=popularity"
                      className="text-xs px-3 py-1.5 rounded-lg bg-bg-tertiary/60 border border-border-subtle text-text-secondary hover:text-text-primary hover:bg-bg-hover transition-all"
                    >
                      Popular Functions
                    </Link>
                  </div>
                </div>
              </div>
            </SpotlightCard>
          </motion.div>

          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 1.1 }}
            className="pt-4 text-center"
          >
            <div className="flex items-center justify-center gap-3 text-text-secondary/40 text-xs font-mono">
              <Binary className="w-4 h-4" />
              <span>01000100 01100101 01100001 01100100</span>
              <span className="text-brand-500/40">|</span>
              <span>01100101 01101110 01100100</span>
            </div>
          </motion.div>
        </div>
      </main>
    </div>
  );
}

export default FunctionNotFound;