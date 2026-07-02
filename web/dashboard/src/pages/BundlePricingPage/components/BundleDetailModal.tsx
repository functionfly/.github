import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Check, Rocket, X, CreditCard, Zap } from 'lucide-react';
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

interface BundleDetailModalProps {
  bundle: Bundle | null;
  icon: LucideIcon;
  colorClass: string;
  isOpen: boolean;
  onClose: () => void;
  onSelect: (bundle: Bundle) => void;
  onPayNow: (bundle: Bundle) => void;
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

// Sentinel value for unlimited - use -1 consistently
const UNLIMITED = -1;

export function BundleDetailModal({ bundle, icon: Icon, colorClass, isOpen, onClose, onSelect, onPayNow }: BundleDetailModalProps) {
  if (!bundle) return null;

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-2xl max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <div className={`inline-flex items-center gap-3 mb-2`}>
            <div className={`w-12 h-12 rounded-xl flex items-center justify-center bg-gradient-to-r ${colorClass} text-white`}>
              <Icon className="w-6 h-6" />
            </div>
            <DialogTitle className="text-2xl">{bundle.display_name}</DialogTitle>
          </div>
          <DialogDescription className="text-base">
            {bundle.description}
          </DialogDescription>
        </DialogHeader>

        {/* Pricing */}
        <div className="flex items-baseline gap-2 py-4 border-y">
          <span className="text-4xl font-bold">{bundle.price_usd}</span>
          <span className="text-text-secondary">/{bundle.billing_interval}</span>
          {bundle.price_cents === 0 && (
            <span className="ml-2 text-sm text-success font-medium">(Free while building)</span>
          )}
        </div>

        {/* All Features */}
        <div className="py-4">
          <p className="text-sm font-semibold text-text-muted uppercase tracking-wide mb-4">
            Everything included:
          </p>
          <div className="grid grid-cols-2 gap-3">
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
        </div>

        {/* Resource Limits */}
        <div className="bg-bg-tertiary rounded-lg p-4">
          <p className="text-sm font-semibold text-text-muted mb-3">
            Resource limits:
          </p>
          <div className="flex flex-wrap gap-2">
            {Object.entries(bundle.feature_limits).map(([key, value]) => (
              <span
                key={key}
                className="text-sm bg-bg-primary px-3 py-1.5 rounded"
              >
                {value === UNLIMITED ? 'Unlimited' : value.toLocaleString()} {key}
              </span>
            ))}
          </div>
        </div>

        {/* Provisioning Steps */}
        {bundle.provisioning_steps.length > 0 && (
          <div className="py-4">
            <p className="text-sm font-semibold text-text-muted uppercase tracking-wide mb-3">
              What happens next:
            </p>
            <ol className="space-y-2">
              {bundle.provisioning_steps.map((step, index) => (
                <li key={index} className="flex items-start gap-3">
                  <span className="flex-shrink-0 w-6 h-6 rounded-full bg-brand-500/20 text-brand-400 text-sm flex items-center justify-center font-medium">
                    {index + 1}
                  </span>
                  <span className="text-text-secondary pt-0.5">{step}</span>
                </li>
              ))}
            </ol>
          </div>
        )}

        {/* Actions */}
        <div className="flex gap-3 pt-4">
          {bundle.price_cents > 0 && (
            <Button
              onClick={() => {
                onPayNow(bundle);
                onClose();
              }}
              className="flex-1 bg-gradient-to-r from-brand-500 to-brand-400 hover:from-ff-flame hover:to-ff-afterburner text-white transition-all duration-300"
            >
              <CreditCard className="w-4 h-4 mr-2" />
              Pay Now — {bundle.price_usd}/{bundle.billing_interval}
            </Button>
          )}
          <Button
            onClick={() => {
              onSelect(bundle);
              onClose();
            }}
            variant={bundle.price_cents > 0 ? 'outline' : 'default'}
            className={`flex-1 ${
              bundle.price_cents > 0
                ? 'border-brand-500/30 text-brand-500 hover:bg-brand-500/10'
                : 'bg-gradient-to-r from-brand-500 to-brand-400 hover:from-ff-flame hover:to-ff-afterburner text-white transition-all duration-300'
            }`}
          >
            <Zap className="w-4 h-4 mr-2" />
            {bundle.price_cents > 0 ? 'Start Free (Founder Mode)' : 'Start Free'}
          </Button>
          <Button variant="ghost" onClick={onClose} className="px-4">
            <X className="w-4 h-4" />
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
