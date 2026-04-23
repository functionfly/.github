import { Badge } from '@/components/ui/badge';
import { PROVIDER_CAPABILITIES, type ProviderCapability } from '../constants/providerMeta';
import { Zap, Server, Container, DollarSign, Clock, Shield, Globe, Cpu } from 'lucide-react';

interface CapabilityBadgesProps {
  providerId: string;
  showAll?: boolean;
  maxDisplay?: number;
}

const CAPABILITY_ICONS: Record<string, React.ReactNode> = {
  edge: <Zap className="w-3 h-3" />,
  serverless: <Server className="w-3 h-3" />,
  containers: <Container className="w-3 h-3" />,
  freeTier: <DollarSign className="w-3 h-3" />,
  lowLatency: <Clock className="w-3 h-3" />,
  security: <Shield className="w-3 h-3" />,
  global: <Globe className="w-3 h-3" />,
  wasm: <Cpu className="w-3 h-3" />,
};

const CAPABILITY_COLORS: Record<string, { bg: string; text: string; border: string }> = {
  edge: { bg: 'bg-amber-100 dark:bg-yellow-900/50', text: 'text-amber-900 dark:text-yellow-300', border: 'border-amber-200 dark:border-yellow-800' },
  serverless: { bg: 'bg-blue-100 dark:bg-blue-900/50', text: 'text-blue-900 dark:text-blue-300', border: 'border-blue-200 dark:border-blue-800' },
  containers: { bg: 'bg-purple-100 dark:bg-purple-900/50', text: 'text-purple-900 dark:text-purple-300', border: 'border-purple-200 dark:border-purple-800' },
  freeTier: { bg: 'bg-green-100 dark:bg-green-900/50', text: 'text-green-900 dark:text-green-300', border: 'border-green-200 dark:border-green-800' },
  lowLatency: { bg: 'bg-orange-100 dark:bg-orange-900/50', text: 'text-orange-900 dark:text-orange-300', border: 'border-orange-200 dark:border-orange-800' },
  security: { bg: 'bg-emerald-100 dark:bg-emerald-900/50', text: 'text-emerald-900 dark:text-emerald-300', border: 'border-emerald-200 dark:border-emerald-800' },
  global: { bg: 'bg-indigo-100 dark:bg-indigo-900/50', text: 'text-indigo-900 dark:text-indigo-300', border: 'border-indigo-200 dark:border-indigo-800' },
  wasm: { bg: 'bg-pink-100 dark:bg-pink-900/50', text: 'text-pink-900 dark:text-pink-300', border: 'border-pink-200 dark:border-pink-800' },
};

const CAPABILITY_LABELS: Record<string, string> = {
  edge: 'Edge',
  serverless: 'Serverless',
  containers: 'Containers',
  freeTier: 'Free Tier',
  lowLatency: 'Low Latency',
  security: 'Enterprise Security',
  global: 'Global CDN',
  wasm: 'WebAssembly',
};

export function CapabilityBadges({ providerId, showAll = false, maxDisplay = 3 }: CapabilityBadgesProps) {
  const capabilities = PROVIDER_CAPABILITIES[providerId] || [];

  if (capabilities.length === 0) return null;

  const displayCapabilities = showAll ? capabilities : capabilities.slice(0, maxDisplay);
  const remaining = capabilities.length - maxDisplay;

  return (
    <div className="flex flex-wrap gap-1.5">
      {displayCapabilities.map((capability) => {
        const colors = CAPABILITY_COLORS[capability] || CAPABILITY_COLORS.serverless;
        return (
          <Badge
            key={capability}
            variant="outline"
            className={`${colors.bg} ${colors.text} ${colors.border} border text-xs font-medium flex items-center gap-1 px-1.5 py-0.5`}
          >
            {CAPABILITY_ICONS[capability]}
            <span>{CAPABILITY_LABELS[capability]}</span>
          </Badge>
        );
      })}
      {!showAll && remaining > 0 && (
        <Badge variant="outline" className="text-xs font-medium px-1.5 py-0.5">
          +{remaining} more
        </Badge>
      )}
    </div>
  );
}
