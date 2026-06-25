import { motion } from 'framer-motion';
import { Globe, MapPin, Check, Zap } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { HelpTooltip } from '@/components/ui/help-tooltip';
import { useOnboardingStore } from '@/stores/onboardingStore';
import { getRegionsLimit, canAddMore } from '@/lib/plan-gating';
import type { PlanTier } from '@/stores/onboardingStore';

const REGIONS = [
  { id: 'us-east-1', name: 'US East', city: 'New York', flag: '🇺🇸', latency: 45 },
  { id: 'us-west-2', name: 'US West', city: 'San Francisco', flag: '🇺🇸', latency: 62 },
  { id: 'eu-central-1', name: 'EU Central', city: 'Frankfurt', flag: '🇪🇺', latency: 120 },
  { id: 'eu-west-1', name: 'EU West', city: 'London', flag: '🇬🇧', latency: 135 },
  { id: 'ap-southeast-1', name: 'Asia Pacific', city: 'Singapore', flag: '🇸🇬', latency: 180 },
  { id: 'ap-northeast-1', name: 'Asia Northeast', city: 'Tokyo', flag: '🇯🇵', latency: 175 },
  { id: 'sa-east-1', name: 'South America', city: 'São Paulo', flag: '🇧🇷', latency: 155 },
  { id: 'ca-central-1', name: 'Canada', city: 'Montreal', flag: '🇨🇦', latency: 55 },
];

const REGION_GROUPS = [
  {
    name: 'Americas',
    regions: REGIONS.filter((r) => r.id.startsWith('us-') || r.id.startsWith('ca-') || r.id.startsWith('sa-')),
  },
  {
    name: 'Europe',
    regions: REGIONS.filter((r) => r.id.startsWith('eu-')),
  },
  {
    name: 'Asia Pacific',
    regions: REGIONS.filter((r) => r.id.startsWith('ap-')),
  },
];

export function RegionSelectionStep() {
  const {
    selectedRegions = [],
    autoSelectRegions = true,
    setSelectedRegions,
    selectedPlan = 'free',
  } = useOnboardingStore();

  const regionsLimit = getRegionsLimit(selectedPlan as PlanTier);
  const isUnlimited = regionsLimit === -1;

  const handleToggleRegion = (regionId: string) => {
    if (selectedRegions.includes(regionId)) {
      setSelectedRegions(selectedRegions.filter((r) => r !== regionId), false);
    } else {
      if (!canAddMore(selectedRegions.length, regionsLimit)) {
        return;
      }
      setSelectedRegions([...selectedRegions, regionId], false);
    }
  };

  const handleAutoSelect = () => {
    const recommended = ['us-east-1', 'eu-central-1', 'ap-southeast-1'];
    const limit = isUnlimited ? recommended.length : Math.min(recommended.length, regionsLimit);
    setSelectedRegions(recommended.slice(0, limit), true);
  };

  const getLatencyColor = (latency: number) => {
    if (latency < 100) return 'text-aviation-green';
    if (latency < 150) return 'text-aviation-amber';
    return 'text-aviation-red';
  };

  return (
    <div className="space-y-6">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="text-center space-y-4"
      >
        <div className="onboarding-step-icon w-16 h-16 rounded-2xl flex items-center justify-center mx-auto bg-blue-500/20">
          <Globe className="w-8 h-8 text-blue-400" />
        </div>
        <p className="text-lg text-aviation-text-secondary font-mono max-w-xl mx-auto">
          Choose deployment regions to reduce latency for your users worldwide.
        </p>
      </motion.div>

      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-sm font-mono text-aviation-text-secondary">
            Selected: {selectedRegions.length}
            {!isUnlimited && ` / ${regionsLimit}`}
            {isUnlimited && ' / Unlimited'}
          </span>
          <HelpTooltip content={`Your plan allows up to ${isUnlimited ? 'unlimited' : regionsLimit} regions`} />
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleAutoSelect}
            className="font-mono text-xs"
          >
            <Zap className="w-3 h-3 mr-1" />
            Auto-Select Best
          </Button>
        </div>
      </div>

      {autoSelectRegions && selectedRegions.length > 0 && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className="bg-aviation-cyan/10 border border-aviation-cyan/30 rounded-lg p-3"
        >
          <div className="flex items-center gap-2 text-aviation-cyan">
            <Zap className="w-4 h-4" />
            <span className="text-sm font-mono">Auto-selected regions for optimal performance</span>
          </div>
        </motion.div>
      )}

      <div className="space-y-6">
        {REGION_GROUPS.map((group) => (
          <div key={group.name} className="space-y-3">
            <h3 className="text-sm font-mono font-semibold text-aviation-text-muted uppercase tracking-wider">
              {group.name}
            </h3>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              {group.regions.map((region, index) => {
                const isSelected = selectedRegions.includes(region.id);
                const isDisabled =
                  !isSelected && !canAddMore(selectedRegions.length, regionsLimit);

                return (
                  <motion.div
                    key={region.id}
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: index * 0.05 }}
                  >
                    <Card
                      className={`onboarding-step-card p-3 cursor-pointer transition-all ${
                        isSelected
                          ? 'border-aviation-amber shadow-lg shadow-aviation-amber/20'
                          : isDisabled
                            ? 'opacity-50 cursor-not-allowed'
                            : 'hover:border-aviation-amber/50'
                      }`}
                      onClick={() => !isDisabled && handleToggleRegion(region.id)}
                    >
                      <div className="flex items-center gap-3">
                        <span className="text-2xl">{region.flag}</span>
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <h4 className="font-mono font-semibold text-aviation-text-primary text-sm">
                              {region.name}
                            </h4>
                          </div>
                          <p className="text-xs text-aviation-text-muted font-mono">
                            {region.city}
                          </p>
                        </div>
                        <div className="flex items-center gap-2">
                          <div className={`text-xs font-mono ${getLatencyColor(region.latency)}`}>
                            {region.latency}ms
                          </div>
                          <div
                            className={`w-5 h-5 rounded-full border-2 flex items-center justify-center ${
                              isSelected
                                ? 'border-aviation-amber bg-aviation-amber'
                                : 'border-aviation-border-panel'
                            }`}
                          >
                            {isSelected && <Check className="w-3 h-3 text-aviation-bg-primary" />}
                          </div>
                        </div>
                      </div>
                    </Card>
                  </motion.div>
                );
              })}
            </div>
          </div>
        ))}
      </div>

      {!canAddMore(selectedRegions.length, regionsLimit) && (
        <div className="bg-aviation-amber/10 border border-aviation-amber/30 rounded-lg p-4">
          <div className="flex items-start gap-3">
            <MapPin className="w-5 h-5 text-aviation-amber flex-shrink-0 mt-0.5" />
            <div>
              <h4 className="font-mono font-semibold text-aviation-amber mb-1">
                Region Limit Reached
              </h4>
              <p className="text-sm text-aviation-text-secondary font-mono">
                Your current plan supports up to {regionsLimit} regions.
                Upgrade to Professional or higher for more regions.
              </p>
            </div>
          </div>
        </div>
      )}

      <div className="bg-aviation-bg-tertiary rounded-lg p-4">
        <h4 className="font-mono font-medium text-aviation-text-primary mb-2">
          Why Region Selection Matters
        </h4>
        <ul className="space-y-1 text-sm text-aviation-text-secondary font-mono">
          <li>• Closer regions = lower latency for your users</li>
          <li>• Multi-region deployment improves reliability</li>
          <li>• FunctionFly automatically routes to the nearest healthy region</li>
        </ul>
      </div>
    </div>
  );
}
