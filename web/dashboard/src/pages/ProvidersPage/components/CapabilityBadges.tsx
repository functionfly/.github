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

const CAPABILITY_COLORS: Record<string, { bg: string; text: string }> = {
  edge: { bg: 'bg-yellow-200 dark:bg-yellow-900/50', text: 'text-yellow-950 dark:text-yellow-300' },
  serverless: { bg: 'bg-blue-200 dark:bg-blue-900/50', text: 'text-blue-950 dark:text-blue-300' },
  containers: { bg: 'bg-purple-200 dark:bg-purple-900/50', text: 'text-purple-950 dark:text-purple-300' },
  freeTier: { bg: 'bg-green-200 dark:bg-green-900/50', text: 'text-green-950 dark:text-green-300' },
  lowLatency: { bg: 'bg-orange-200 dark:bg-orange-900/50', text: 'text-orange-950 dark:text-orange-300' },
  security: { bg: 'bg-emerald-200 dark:bg-emerald-900/50', text: 'text-emerald-950 dark:text-emerald-300' },
  global: { bg: 'bg-indigo-200 dark:bg-indigo-900/50', text: 'text-indigo-950 dark:text-indigo-300' },
  wasm: { bg: 'bg-pink-200 dark:bg-pink-900/50', text: 'text-pink-950 dark:text-pink-300' },
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
            className={`${colors.bg} ${colors.text} border-0 text-xs font-medium flex items-center gap-1 px-1.5 py-0.5`}
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
