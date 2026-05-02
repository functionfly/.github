import { motion } from 'framer-motion';
import { Link } from 'react-router-dom';
import {
  AlertTriangle,
  ArrowLeft,
  Binary,
  Home,
  Search,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { ParticleBackground } from '@/components/ui/ParticleBackground';
import { SpotlightCard } from '@/components/ui/SpotlightCard';
import './styles.css';

interface FunctionDetailNotFoundProps {
  id?: string;
  errorMessage?: string;
}

export function FunctionDetailNotFound({ id, errorMessage }: FunctionDetailNotFoundProps) {
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
    <div className="min-h-[60vh] flex flex-col bg-bg-primary relative overflow-hidden">
      <div className="absolute inset-0 opacity-10">
        <ParticleBackground
          particleCount={15}
          color="rgba(var(--brand-500), 0.3)"
        />
      </div>

      <main className="flex-1 flex flex-col items-center justify-center p-4 relative z-10">
        <div className="text-center space-y-6 max-w-lg w-full">
          <motion.div
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.5 }}
          >
            <div className="text-7xl font-black text-brand-500/40 mb-2">404</div>
            <p className="text-lg text-text-secondary font-mono">{randomQuote}</p>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.3 }}
            className="glass rounded-xl p-6 border border-border/50"
          >
            <div className="flex items-center gap-3 mb-4">
              <AlertTriangle className="h-5 w-5 text-warning" />
              <span className="font-semibold text-text-primary">Function Not Found</span>
            </div>
            {id ? (
              <p className="text-sm text-text-secondary">
                The function with ID <code className="text-brand-400 bg-brand-500/10 px-2 py-0.5 rounded">{id}</code> could not be found or may have been deleted.
              </p>
            ) : (
              <p className="text-sm text-text-secondary">
                {errorMessage || 'The function you are looking for does not exist or has been removed.'}
              </p>
            )}
          </motion.div>

          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.5 }}
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
              <Link to="/functions">
                <Search className="w-4 h-4" />
                My Functions
              </Link>
            </Button>
            <Button asChild variant="secondary" className="gap-2" size="lg">
              <Link to="/">
                <Home className="w-4 h-4" />
                Dashboard
              </Link>
            </Button>
          </motion.div>

          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.7 }}
          >
            <SpotlightCard
              className="text-left"
              spotlightColor="rgba(var(--info), 0.1)"
            >
              <div className="flex items-start gap-3">
                <Search className="h-5 w-5 text-info mt-0.5" />
                <div>
                  <h3 className="font-semibold text-text-primary mb-1">Looking for a published function?</h3>
                  <p className="text-xs text-text-secondary mb-3">
                    This page shows your deployed functions. Published registry functions have different URLs.
                  </p>
                  <Link
                    to="/registry"
                    className="inline-flex items-center gap-2 text-sm px-4 py-2 rounded-lg bg-gradient-to-r from-brand-500 to-purple-500 text-white hover:brightness-110 transition-all"
                  >
                    Browse Registry
                  </Link>
                </div>
              </div>
            </SpotlightCard>
          </motion.div>

          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.9 }}
            className="pt-4"
          >
            <div className="flex items-center justify-center gap-3 text-text-secondary/40 text-xs font-mono">
              <Binary className="w-4 h-4" />
              <span>01000100 01100101 01100001 01100100</span>
            </div>
          </motion.div>
        </div>
      </main>
    </div>
  );
}

export default FunctionDetailNotFound;