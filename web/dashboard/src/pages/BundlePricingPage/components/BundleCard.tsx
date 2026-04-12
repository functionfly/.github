import { Button } from '@/components/ui/button';
import { motion } from 'framer-motion';
import { Check, Rocket, Sparkles } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';

interface Bundle {
  id: string;
  slug: string;
  name: string;
  display_name: string;
  description: string;
  short_description: string;
  price_cents: number;
  price_usd: string;
  billing_interval: string;
  icon: string;
  color: string;
  features_included: string[];
  feature_limits: Record<string, number>;
  provisioning_steps: string[];
  is_popular: boolean;
}

interface BundleCardProps {
  bundle: Bundle;
  icon: LucideIcon;
  colorClass: string;
  onSelect: (bundle: Bundle) => void;
  delay?: number;
}

const featureLabels: Record<string, string> = {
  auth: 'Auth',
  payments: 'Payments (Stripe)',
  user_db: 'User DB',
  email: 'Email workflows',
  analytics: 'Analytics',
  listings: 'Listings',
  messaging: 'Messaging',
  notifications: 'Notifications',
  vector_db: 'Vector DB',
  embeddings: 'Embeddings',
  chat: 'Chat workflows',
  memory: 'Memory system',
};

export function BundleCard({ bundle, icon: Icon, colorClass, onSelect, delay = 0 }: BundleCardProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay, duration: 0.4 }}
      className="relative group"
    >
      {bundle.is_popular && (
        <div className="absolute -top-4 left-1/2 -translate-x-1/2 z-10">
          <div className="bg-gradient-to-r from-violet-600 to-fuchsia-600 text-white text-sm font-bold px-4 py-1 rounded-full flex items-center gap-1 shadow-lg">
            <Sparkles className="w-4 h-4" />
            Most Popular
          </div>
        </div>
      )}

      <div className="h-full bg-white dark:bg-slate-800 rounded-2xl shadow-lg hover:shadow-xl transition-shadow border border-slate-200 dark:border-slate-700 overflow-hidden">
        {/* Header with gradient */}
        <div className={`bg-gradient-to-r ${colorClass} p-6 text-white`}>
          <div className="flex items-center gap-3 mb-4">
            <div className="w-12 h-12 bg-white/20 rounded-xl flex items-center justify-center">
              <Icon className="w-6 h-6" />
            </div>
            <div>
              <h3 className="text-xl font-bold">{bundle.display_name}</h3>
              <p className="text-white/80 text-sm">{bundle.short_description}</p>
            </div>
          </div>

          <div className="flex items-baseline gap-1">
            <span className="text-4xl font-bold">{bundle.price_usd}</span>
            <span className="text-white/70">/{bundle.billing_interval}</span>
          </div>
        </div>

        {/* Content */}
        <div className="p-6">
          {/* Tagline */}
          <p className="text-slate-600 dark:text-slate-400 mb-6">
            {bundle.description}
          </p>

          {/* Features */}
          <div className="space-y-3 mb-6">
            <p className="text-sm font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">
              Includes:
            </p>
            {bundle.features_included.map((feature) => (
              <div key={feature} className="flex items-center gap-3">
                <div className="w-5 h-5 rounded-full bg-green-100 dark:bg-green-900/30 flex items-center justify-center flex-shrink-0">
                  <Check className="w-3 h-3 text-green-600 dark:text-green-400" />
                </div>
                <span className="text-slate-700 dark:text-slate-300">
                  {featureLabels[feature] || feature}
                </span>
              </div>
            ))}
          </div>

          {/* Resource limits preview */}
          <div className="bg-slate-50 dark:bg-slate-700/50 rounded-lg p-4 mb-6">
            <p className="text-xs text-slate-500 dark:text-slate-400 mb-2">
              Resources included:
            </p>
            <div className="flex flex-wrap gap-2">
              {Object.entries(bundle.feature_limits).slice(0, 3).map(([key, value]) => (
                <span
                  key={key}
                  className="text-xs bg-slate-200 dark:bg-slate-600 px-2 py-1 rounded"
                >
                  {value === -1 || value === 999999999 ? 'Unlimited' : value.toLocaleString()} {key}
                </span>
              ))}
            </div>
          </div>

          {/* CTA Buttons */}
          <div className="space-y-3">
            <Button
              onClick={() => onSelect(bundle)}
              className="w-full bg-gradient-to-r from-violet-600 to-fuchsia-600 hover:from-violet-700 hover:to-fuchsia-700 text-white"
            >
              <Rocket className="w-4 h-4 mr-2" />
              {bundle.price_cents === 0 ? 'Start Free' : 'Get Started'}
            </Button>

            <Button
              variant="ghost"
              onClick={() => onSelect(bundle)}
              className="w-full text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white"
            >
              View Details
            </Button>
          </div>
        </div>
      </div>
    </motion.div>
  );
}
