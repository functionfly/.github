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
      className="max-w-2xl mx-auto mb-10"
    >
      <div className="bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700 p-2">
        <div className="flex flex-col sm:flex-row gap-2">
          <button
            onClick={() => onModeChange('immediate')}
            className={`flex-1 flex items-center justify-center gap-2 px-4 py-3 rounded-lg transition-all ${
              mode === 'immediate'
                ? 'bg-slate-900 dark:bg-slate-700 text-white shadow-md'
                : 'hover:bg-slate-50 dark:hover:bg-slate-700/50 text-slate-600 dark:text-slate-400'
            }`}
          >
            <CreditCard className="w-4 h-4" />
            <div className="text-left">
              <div className="font-semibold text-sm">Pay Now</div>
              <div className={`text-xs ${mode === 'immediate' ? 'text-slate-300' : 'text-slate-500'}`}>
                Start immediately with full features
              </div>
            </div>
          </button>

          <button
            onClick={() => onModeChange('deferred')}
            className={`flex-1 flex items-center justify-center gap-2 px-4 py-3 rounded-lg transition-all relative overflow-hidden ${
              mode === 'deferred'
                ? 'bg-gradient-to-r from-violet-600 to-fuchsia-600 text-white shadow-md'
                : 'hover:bg-slate-50 dark:hover:bg-slate-700/50 text-slate-600 dark:text-slate-400'
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
                    : 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400'
                }`}>
                  FREE
                </span>
              </div>
              <div className={`text-xs ${mode === 'deferred' ? 'text-white/80' : 'text-slate-500'}`}>
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
            className="mt-3 p-3 bg-gradient-to-r from-violet-50 to-fuchsia-50 dark:from-violet-900/20 dark:to-fuchsia-900/20 rounded-lg border border-violet-100 dark:border-violet-800"
          >
            <div className="flex items-start gap-3">
              <div className="w-8 h-8 bg-violet-100 dark:bg-violet-800 rounded-lg flex items-center justify-center flex-shrink-0">
                <Clock className="w-4 h-4 text-violet-600 dark:text-violet-400" />
              </div>
              <div>
                <p className="text-sm font-medium text-slate-900 dark:text-white">
                  "Build Now, Pay Later" is active
                </p>
                <p className="text-sm text-slate-600 dark:text-slate-400 mt-1">
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
