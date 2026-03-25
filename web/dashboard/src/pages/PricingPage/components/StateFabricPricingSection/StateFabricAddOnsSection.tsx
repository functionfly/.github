import {
  createStateFabricAddOnCheckout,
  getStateFabricAddOnEntitlements,
  listStateFabricAddOnCatalog,
} from '@/api/billing';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { useMutation, useQuery } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { BarChart3, Brain, Shield, Zap } from 'lucide-react';
import { toast } from 'sonner';
import { useScrollAnimation } from '../../hooks';
import { FALLBACK_STATE_FABRIC_ADDONS, type StateFabricAddOnCatalogItem } from './data';

const ADDON_ICONS = [Zap, Shield, Brain, BarChart3];

const iconForAddon = (addon: StateFabricAddOnCatalogItem, index: number) => {
  const byId: Record<string, number> = {
    hot_cache_booster: 0,
    advanced_security_pack: 1,
    ai_memory_pack: 2,
    advanced_insights: 3,
  };
  const i = byId[addon.id] ?? index % ADDON_ICONS.length;
  return ADDON_ICONS[i % ADDON_ICONS.length];
};

export function StateFabricAddOnsSection() {
  const { ref, inView } = useScrollAnimation(0.1, false);
  const { data, isError } = useQuery({
    queryKey: ['billing', 'state-fabric-add-ons-catalog'],
    queryFn: listStateFabricAddOnCatalog,
    staleTime: 5 * 60_000,
    retry: 1,
  });
  const addOns: StateFabricAddOnCatalogItem[] =
    data?.add_ons?.length && !isError ? data.add_ons : FALLBACK_STATE_FABRIC_ADDONS;
  const { data: entitlements } = useQuery({
    queryKey: ['billing', 'state-fabric-add-ons-entitlements'],
    queryFn: getStateFabricAddOnEntitlements,
    retry: false,
  });
  const checkoutMutation = useMutation({
    mutationFn: async (addonId: string) => {
      const res = await createStateFabricAddOnCheckout(
        addonId,
        `${window.location.origin}/pricing?sf_addon=success`,
        `${window.location.origin}/pricing?sf_addon=cancel`
      );
      return res;
    },
    onError: () => toast.error('Could not start add-on checkout. Please try again.'),
  });

  return (
    <motion.div
      ref={ref}
      initial={{ opacity: 0, y: 20 }}
      animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
      transition={{ duration: 0.5 }}
      className="mt-16"
    >
      <h3 className="text-xl font-bold text-white mb-4 text-center">Optional add-ons</h3>
      <p className="text-text-secondary text-sm text-center max-w-2xl mx-auto mb-6">
        Enhance any plan with performance, security, or analytics add-ons.
      </p>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 max-w-6xl mx-auto">
        {addOns.map((addon, index) => {
          const Icon = iconForAddon(addon, index);
          return (
            <motion.div
              key={addon.id}
              initial={{ opacity: 0, y: 10 }}
              animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 10 }}
              transition={{ delay: index * 0.08 }}
            >
              <Card className="pricing-state-fabric-addon border-white/8 bg-white/5 h-full">
                <CardContent className="p-4">
                  <div className="flex items-center gap-2 mb-2">
                    <Icon className="w-5 h-5 text-[#6366f1] shrink-0" />
                    <h4 className="text-base font-semibold text-white">{addon.name}</h4>
                  </div>
                  <div className="flex items-baseline gap-1 text-white font-bold mb-1">
                    <span className="pricing-state-fabric-price">{addon.price}</span>
                    <span className="text-text-secondary text-sm font-normal">{addon.period}</span>
                  </div>
                  <p className="text-text-secondary text-sm">{addon.description}</p>
                  <div className="mt-3 flex items-center justify-between gap-2">
                    {entitlements?.addon_ids?.includes(addon.id) ? (
                      <span className="text-xs text-emerald-400 font-medium">Active</span>
                    ) : (
                      <span className="text-xs text-text-secondary">Not active</span>
                    )}
                    <Button
                      size="sm"
                      variant="outline"
                      disabled={checkoutMutation.isPending}
                      onClick={async () => {
                        if (entitlements?.addon_ids?.includes(addon.id)) {
                          toast.message('This add-on is already active.');
                          return;
                        }
                        const res = await checkoutMutation.mutateAsync(addon.id);
                        if (res?.url) {
                          window.location.href = res.url;
                        }
                      }}
                    >
                      Buy add-on
                    </Button>
                  </div>
                </CardContent>
              </Card>
            </motion.div>
          );
        })}
      </div>
    </motion.div>
  );
}
