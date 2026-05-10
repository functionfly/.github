import { motion } from 'framer-motion';
import { Check, Minus, type LucideIcon } from 'lucide-react';
import type { Bundle } from '@/api/billing';

interface BundleComparisonTableProps {
  bundles: Bundle[];
  iconMap: Record<string, LucideIcon>;
  colorMap: Record<string, string>;
}

interface ComparisonRow {
  category: string;
  features: { key: string; label: string }[];
}

const featureCategories: ComparisonRow[] = [
  {
    category: 'Authentication',
    features: [{ key: 'auth', label: 'Auth' }],
  },
  {
    category: 'Payments',
    features: [{ key: 'payments', label: 'Stripe Payments' }],
  },
  {
    category: 'Database',
    features: [
      { key: 'user_db', label: 'User DB' },
      { key: 'vector_db', label: 'Vector DB' },
    ],
  },
  {
    category: 'Communication',
    features: [
      { key: 'email', label: 'Email Workflows' },
      { key: 'notifications', label: 'Notifications' },
      { key: 'messaging', label: 'Messaging' },
    ],
  },
  {
    category: 'AI & Data',
    features: [
      { key: 'analytics', label: 'Analytics' },
      { key: 'embeddings', label: 'Embeddings' },
      { key: 'chat', label: 'Chat Workflows' },
      { key: 'memory', label: 'Memory System' },
    ],
  },
  {
    category: 'Commerce',
    features: [{ key: 'listings', label: 'Listings' }],
  },
];

const UNLIMITED = -1;

export function BundleComparisonTable({ bundles, iconMap, colorMap }: BundleComparisonTableProps) {
  const hasBundleFeature = (bundle: Bundle, featureKey: string): boolean => {
    return bundle.features_included.includes(featureKey);
  };

  const getFeatureValue = (bundle: Bundle, featureKey: string): { value: string; isUnlimited: boolean } | null => {
    const limit = bundle.feature_limits[featureKey];
    if (limit === undefined) return null;
    if (limit === UNLIMITED) return { value: 'Unlimited', isUnlimited: true };
    return { value: limit.toLocaleString(), isUnlimited: false };
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 40 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.8, ease: 'easeOut' }}
      className="mt-20 mb-12"
    >
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ duration: 0.6, delay: 0.2 }}
        className="text-center mb-8"
      >
        <h2 className="text-3xl font-bold text-text-primary mb-4">Compare Bundles</h2>
        <p className="text-text-secondary max-w-2xl mx-auto">
          See what&apos;s included in each bundle to find the perfect fit for your project
        </p>
      </motion.div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, delay: 0.4 }}
        className="bg-bg-secondary rounded-2xl shadow-lg border border-border overflow-hidden"
      >
        <div className="overflow-x-auto">
          <table className="w-full min-w-[600px]">
            <thead>
              <tr className="border-b border-border">
                <th className="text-left p-5 text-text-muted font-semibold uppercase tracking-wide text-sm">
                  Features
                </th>
                {bundles.map((bundle) => {
                  const Icon = iconMap[bundle.icon as keyof typeof iconMap] || Check;
                  return (
                    <th key={bundle.id} className="text-center p-5 min-w-[160px]">
                      <div className="flex flex-col items-center gap-2">
                        <div
                          className={`w-10 h-10 rounded-xl flex items-center justify-center bg-gradient-to-r ${colorMap[bundle.color as keyof typeof colorMap] || colorMap.blue} text-white`}
                        >
                          <Icon className="w-5 h-5" />
                        </div>
                        <span className="font-semibold text-text-primary text-sm">
                          {bundle.display_name}
                        </span>
                      </div>
                    </th>
                  );
                })}
              </tr>
            </thead>
            <tbody>
              {featureCategories.map((category, catIndex) => (
                <>
                  <tr
                    key={`category-${catIndex}`}
                    className="bg-bg-tertiary/50 border-b border-border"
                  >
                    <td colSpan={bundles.length + 1} className="px-5 py-3">
                      <span className="text-sm font-bold text-text-primary uppercase tracking-wide">
                        {category.category}
                      </span>
                    </td>
                  </tr>
                  {category.features.map((feature) => (
                    <tr
                      key={feature.key}
                      className={`border-b border-border/50 ${catIndex % 2 === 0 ? 'bg-white/2 dark:bg-white/[0.02]' : ''}`}
                    >
                      <td className="p-4 pl-8 text-text-secondary">
                        {feature.label}
                      </td>
                      {bundles.map((bundle) => {
                        const hasFeature = hasBundleFeature(bundle, feature.key);
                        const featureValue = getFeatureValue(bundle, feature.key);
                        return (
                          <td key={bundle.id} className="p-4 text-center">
                            {hasFeature ? (
                              featureValue ? (
                                <span className={`text-sm font-medium ${featureValue.isUnlimited ? 'text-success' : 'text-text-primary'}`}>
                                  {featureValue.value}
                                </span>
                              ) : (
                                <div className="w-6 h-6 rounded-full bg-success/10 dark:bg-success/20 flex items-center justify-center mx-auto">
                                  <Check className="w-4 h-4 text-success" />
                                </div>
                              )
                            ) : (
                              <Minus className="w-5 h-5 text-text-muted mx-auto" />
                            )}
                          </td>
                        );
                      })}
                    </tr>
                  ))}
                </>
              ))}
            </tbody>
          </table>
        </div>
      </motion.div>
    </motion.div>
  );
}