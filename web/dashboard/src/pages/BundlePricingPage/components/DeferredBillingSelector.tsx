import { Button } from '@/components/ui/button';
import { motion } from 'framer-motion';
import { Clock, CreditCard, Sparkles } from 'lucide-react';

type PricingMode = 'immediate' | 'deferred';

interface DeferredBillingSelectorProps {
  mode: PricingMode;
  onModeChange: (mode: PricingMode) => void;
}

export function DeferredBillingSelector({ mode, onModeChange }: DeferredBillingSelectorProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: 0.3 }}
      className="sticky top-16 z-30 max-w-2xl mx-auto mb-10 py-3 bg-bg-primary/80 backdrop-blur-lg -mx-4 px-4 border-b border-border/50"
    >
      <div className="bg-bg-secondary rounded-xl shadow-sm border border-border p-2">
        <div className="flex flex-col sm:flex-row gap-2">
          <button
            onClick={() => onModeChange('immediate')}
            className={`flex-1 flex items-center justify-center gap-2 px-4 py-3 rounded-lg transition-all duration-300 ${
              mode === 'immediate'
                ? 'bg-text-primary text-white shadow-md'
                : 'hover:bg-ff-flame/10 text-text-secondary hover:text-ff-flame'
            }`}
          >
            <CreditCard className="w-4 h-4" />
            <div className="text-left">
              <div className="font-semibold text-sm">Pay Now</div>
              <div className={`text-xs ${mode === 'immediate' ? 'text-white/70' : 'text-text-muted'}`}>
                Start immediately with full features
              </div>
            </div>
          </button>

          <button
            onClick={() => onModeChange('deferred')}
            className={`flex-1 flex items-center justify-center gap-2 px-4 py-3 rounded-lg transition-all duration-300 relative overflow-hidden ${
              mode === 'deferred'
                ? 'bg-gradient-to-r from-brand-500 to-brand-400 text-white shadow-md'
                : 'hover:bg-ff-flame/10 text-text-secondary hover:text-ff-flame'
            }`}
          >
            {mode === 'deferred' && (
              <div className="absolute top-0 right-0 w-16 h-16 bg-white/10 rounded-bl-full" />
            )}
            <Sparkles className="w-4 h-4" />
            <div className="text-left relative z-10">
              <div className="font-semibold text-sm flex items-center gap-2">
                Founder Mode
                <span className={`text-xs px-2 py-0.5 rounded-full ${
                  mode === 'deferred'
                    ? 'bg-white/20 text-white'
                    : 'bg-success/10 text-success'
                }`}>
                  FREE
                </span>
              </div>
              <div className={`text-xs ${mode === 'deferred' ? 'text-white/80' : 'text-text-muted'}`}>
                Pay only when you hit 100 users or $1K MRR
              </div>
            </div>
          </button>
        </div>

        {mode === 'deferred' && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            className="mt-3 p-3 bg-brand-500/5 dark:bg-brand-500/10 rounded-lg border border-brand-500/20"
          >
            <div className="flex items-start gap-3">
              <div className="w-8 h-8 bg-brand-500/10 rounded-lg flex items-center justify-center flex-shrink-0">
                <Clock className="w-4 h-4 text-brand-400" />
              </div>
              <div>
                <p className="text-sm font-medium text-text-primary">
                  "Build Now, Pay Later" is active
                </p>
                <p className="text-sm text-text-secondary mt-1">
                  We'll track your growth and email you when you're approaching the billing threshold.
                  You'll have a 7-day grace period to add payment info.
                </p>
              </div>
            </div>
          </motion.div>
        )}
      </div>
    </motion.div>
  );
}
