import { Button } from '@/components/ui/button';
import { motion } from 'framer-motion';
import { Check, Rocket, Sparkles, CreditCard, Zap, Loader2 } from 'lucide-react';
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
  onPayNow: (bundle: Bundle) => void;
  onViewDetails: (bundle: Bundle) => void;
  delay?: number;
  loading?: boolean;
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

// Sentinel value for unlimited resources
const UNLIMITED = -1;

export function BundleCard({ bundle, icon: Icon, colorClass, onSelect, onPayNow, onViewDetails, delay = 0, loading = false }: BundleCardProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay, duration: 0.4 }}
      className="relative group"
    >
      <div className="h-full bg-bg-secondary rounded-2xl shadow-lg hover:shadow-xl transition-shadow border border-border overflow-hidden">
        {/* Most Popular Badge - floats above card */}
        {bundle.is_popular && (
          <div className="absolute -top-0 left-1/2 -translate-x-1/2 -translate-y-1/2 z-20">
            <div className="most-popular-badge text-white text-xs font-bold px-5 py-1.5 rounded-full flex items-center gap-1.5 shadow-lg border border-white/30 whitespace-nowrap animate-pulse-glow">
              <Sparkles className="w-3.5 h-3.5" />
              Most Popular
            </div>
          </div>
        )}

        {/* Popular glow effect */}
        {bundle.is_popular && (
          <div className="absolute inset-0 bg-gradient-to-br from-brand-500/5 via-transparent to-transparent pointer-events-none" />
        )}

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
          <p className="text-text-secondary mb-6">
            {bundle.description}
          </p>

          {/* Features */}
          <div className="space-y-3 mb-6">
            <p className="text-sm font-semibold text-text-muted uppercase tracking-wide">
              Includes:
            </p>
            {bundle.features_included.map((feature) => (
              <div key={feature} className="flex items-center gap-3">
                <div className="w-5 h-5 rounded-full bg-success/10 dark:bg-success/20 flex items-center justify-center flex-shrink-0">
                  <Check className="w-3 h-3 text-success" />
                </div>
                <span className="text-text-primary">
                  {featureLabels[feature] || feature}
                </span>
              </div>
            ))}
          </div>

          {/* Resource limits preview */}
          <div className="bg-bg-tertiary rounded-lg p-4 mb-6">
            <p className="text-xs text-text-muted mb-2">
              Resources included:
            </p>
            <div className="flex flex-wrap gap-2">
              {Object.entries(bundle.feature_limits).slice(0, 3).map(([key, value]) => (
                <span
                  key={key}
                  className="text-xs bg-bg-primary px-2 py-1 rounded"
                >
                  {value === UNLIMITED ? 'Unlimited' : value.toLocaleString()} {key}
                </span>
              ))}
            </div>
          </div>

          {/* CTA Buttons */}
          <div className="space-y-2">
            {bundle.price_cents > 0 && (
              <Button
                onClick={() => onPayNow(bundle)}
                disabled={loading}
                className="w-full bg-gradient-to-r from-brand-500 to-brand-400 hover:from-ff-flame hover:to-ff-afterburner text-white transition-all duration-300 hover:shadow-[0_4px_20px_rgba(255,107,53,0.4)]"
              >
                {loading ? <Loader2 className="w-4 h-4 mr-2 animate-spin" /> : <CreditCard className="w-4 h-4 mr-2" />}
                {loading ? 'Redirecting to Stripe...' : `Pay Now — ${bundle.price_usd}/${bundle.billing_interval}`}
              </Button>
            )}

            <Button
              onClick={() => onSelect(bundle)}
              disabled={loading}
              variant={bundle.price_cents > 0 ? 'outline' : 'default'}
              className={`w-full ${
                bundle.price_cents > 0
                  ? 'border-brand-500/30 text-brand-500 hover:bg-brand-500/10'
                  : 'bg-gradient-to-r from-brand-500 to-brand-400 hover:from-ff-flame hover:to-ff-afterburner text-white transition-all duration-300 hover:shadow-[0_4px_20px_rgba(255,107,53,0.4)]'
              }`}
            >
              <Zap className="w-4 h-4 mr-2" />
              {bundle.price_cents > 0 ? 'Start Free (Founder Mode)' : 'Start Free'}
            </Button>

            <Button
              variant="ghost"
              onClick={() => onViewDetails(bundle)}
              disabled={loading}
              className="w-full text-text-secondary hover:text-ff-flame hover:bg-ff-flame/5 transition-all"
            >
              View Details
            </Button>
          </div>
        </div>
      </div>
    </motion.div>
  );
}
