import { Link } from 'react-router-dom';
import { Lock } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { useStateFabricEntitlements } from '@/hooks/useBilling';
import { useHasStateFabricAddon } from './StateFabricAddonGate.hooks';

export type StateFabricAddOnId =
  | 'advanced_insights'
  | 'advanced_security_pack'
  | 'hot_cache_booster'
  | 'ai_memory_pack';

const ADDON_LABELS: Record<StateFabricAddOnId, string> = {
  advanced_insights: 'Advanced Insights',
  advanced_security_pack: 'Advanced Security Pack',
  hot_cache_booster: 'Hot Cache Booster',
  ai_memory_pack: 'AI Memory Pack',
};

const ADDON_DESCRIPTIONS: Record<StateFabricAddOnId, string> = {
  advanced_insights: 'Unlock fabric metrics, throughput charts, and latency dashboards.',
  advanced_security_pack: 'Unlock the full event log and audit trail for this fabric.',
  hot_cache_booster: 'Unlock deterministic replay sessions for debugging and recovery.',
  ai_memory_pack: 'Unlock vector and AI memory store types.',
};

interface StateFabricAddonGateProps {
  addonId: StateFabricAddOnId;
  children: React.ReactNode;
}

export { useHasStateFabricAddon };

export function StateFabricAddonGate({ addonId, children }: StateFabricAddonGateProps) {
  const { data, isLoading } = useStateFabricEntitlements();
  const entitled = (data?.addon_ids ?? []).includes(addonId);

  if (isLoading) {
    return (
      <Card className="border-dashed border-white/20">
        <CardContent className="py-10 text-center text-text-muted text-sm">
          Checking add-on entitlements…
        </CardContent>
      </Card>
    );
  }

  if (entitled) {
    return <>{children}</>;
  }

  return (
    <Card className="border-dashed border-brand-500/30 bg-brand-500/5">
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-lg">
          <Lock className="w-5 h-5 text-brand-400" />
          {ADDON_LABELS[addonId]} required
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <p className="text-sm text-text-secondary">{ADDON_DESCRIPTIONS[addonId]}</p>
        <p className="text-xs text-text-muted">
          State Fabric core features (fabrics, stores, pipelines, snapshots, triggers) are included
          on paid plans. This tab requires an additional add-on.
        </p>
        <Button asChild variant="outline" size="sm">
          <Link to="/pricing#state-fabric-add-ons">View add-ons on Pricing</Link>
        </Button>
      </CardContent>
    </Card>
  );
}
