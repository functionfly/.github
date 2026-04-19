import { ProviderIcon } from '@/components/common/ProviderIcon';
import { PROVIDER_CAPABILITIES, PROVIDER_COMPARISON, getProviderRegionGroup } from '../constants/providerMeta';
import type { ProviderConfig } from '../constants/providerMeta';
import { Zap, Server, Container, DollarSign, Clock, Shield, Globe, Cpu, ChevronRight } from 'lucide-react';
import { useState, useRef, useEffect } from 'react';
import { createPortal } from 'react-dom';

interface ProviderComparisonTooltipProps {
  provider: ProviderConfig;
  children: React.ReactNode;
}

const CAPABILITY_ICONS: Record<string, React.ReactNode> = {
  edge: <Zap className="w-3.5 h-3.5" />,
  serverless: <Server className="w-3.5 h-3.5" />,
  containers: <Container className="w-3.5 h-3.5" />,
  freeTier: <DollarSign className="w-3.5 h-3.5" />,
  lowLatency: <Clock className="w-3.5 h-3.5" />,
  security: <Shield className="w-3.5 h-3.5" />,
  global: <Globe className="w-3.5 h-3.5" />,
  wasm: <Cpu className="w-3.5 h-3.5" />,
};

export function ProviderComparisonTooltip({ provider, children }: ProviderComparisonTooltipProps) {
  const comparison = PROVIDER_COMPARISON[provider.id];
  const capabilities = PROVIDER_CAPABILITIES[provider.id] || [];
  const [isVisible, setIsVisible] = useState(false);
  const [position, setPosition] = useState({ top: 0, left: 0 });
  const triggerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (isVisible && triggerRef.current) {
      const rect = triggerRef.current.getBoundingClientRect();
      const tooltipWidth = 288; // w-72 = 18rem = 288px
      const tooltipHeight = 200; // approximate height
      
      // Position to the right, vertically centered
      let left = rect.right + 12;
      let top = rect.top + rect.height / 2 - tooltipHeight / 2;
      
      // Check if tooltip would go off-screen to the right
      if (left + tooltipWidth > window.innerWidth) {
        // Position to the left instead
        left = rect.left - tooltipWidth - 12;
      }
      
      // Check vertical bounds
      if (top < 10) top = 10;
      if (top + tooltipHeight > window.innerHeight - 10) {
        top = window.innerHeight - tooltipHeight - 10;
      }
      
      setPosition({ top, left });
    }
  }, [isVisible]);

  const tooltipContent = (
    <div 
      className="fixed z-[9999] pointer-events-none"
      style={{ top: position.top, left: position.left }}
    >
      <div className="w-72 bg-bg-tertiary border border-border-subtle rounded-lg shadow-2xl p-4 animate-in fade-in duration-200">
        {/* Header */}
        <div className="flex items-center gap-3 mb-3">
          <div
            className="w-10 h-10 rounded-xl flex items-center justify-center"
            style={{ backgroundColor: `${provider.color}15` }}
          >
            <ProviderIcon provider={provider.id} size="md" />
          </div>
          <div>
            <h4 className="font-semibold text-text-primary">{provider.name}</h4>
            <p className="text-xs text-text-secondary">Provider Comparison</p>
          </div>
        </div>

        {/* Quick Stats */}
        {comparison && (
          <div className="grid grid-cols-3 gap-2 mb-3">
            <div className="p-2 rounded bg-bg-secondary/50 text-center">
              <p className="text-xs text-text-tertiary">Cold Start</p>
              <p className="text-sm font-medium text-text-primary">{comparison.coldStartMs}ms</p>
            </div>
            <div className="p-2 rounded bg-bg-secondary/50 text-center">
              <p className="text-xs text-text-tertiary">Pricing</p>
              <p className="text-sm font-medium text-text-primary">{comparison.pricingTier}</p>
            </div>
            <div className="p-2 rounded bg-bg-secondary/50 text-center">
              <p className="text-xs text-text-tertiary">Regions</p>
              <p className="text-sm font-medium text-text-primary">{comparison.regionCount}</p>
            </div>
          </div>
        )}

        {/* Capabilities */}
        <div className="mb-3">
          <p className="text-xs text-text-tertiary mb-2">Key Features</p>
          <div className="flex flex-wrap gap-1.5">
            {capabilities.slice(0, 4).map((cap) => (
              <span
                key={cap}
                className="inline-flex items-center gap-1 px-2 py-1 rounded bg-bg-secondary text-xs text-text-secondary"
              >
                {CAPABILITY_ICONS[cap]}
                <span className="capitalize">{cap.replace(/([A-Z])/g, ' $1').trim()}</span>
              </span>
            ))}
          </div>
        </div>

        {/* Comparison Note */}
        {comparison && (
          <div className="pt-3 border-t border-border-subtle">
            <p className="text-xs text-text-secondary leading-relaxed">{comparison.bestFor}</p>
          </div>
        )}
      </div>
    </div>
  );

  return (
    <div 
      ref={triggerRef}
      className="relative cursor-help"
      onMouseEnter={() => setIsVisible(true)}
      onMouseLeave={() => setIsVisible(false)}
    >
      {children}
      {isVisible && typeof document !== 'undefined' && createPortal(tooltipContent, document.body)}
    </div>
  );
}

// Comparison table for side-by-side view
interface ProviderComparisonTableProps {
  providers: ProviderConfig[];
}

export function ProviderComparisonTable({ providers }: ProviderComparisonTableProps) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-border-subtle">
            <th className="text-left py-2 px-3 text-text-tertiary font-medium">Provider</th>
            <th className="text-center py-2 px-3 text-text-tertiary font-medium">Cold Start</th>
            <th className="text-center py-2 px-3 text-text-tertiary font-medium">Pricing</th>
            <th className="text-center py-2 px-3 text-text-tertiary font-medium">Regions</th>
            <th className="text-left py-2 px-3 text-text-tertiary font-medium">Best For</th>
          </tr>
        </thead>
        <tbody>
          {providers.map((provider) => {
            const comparison = PROVIDER_COMPARISON[provider.id];
            if (!comparison) return null;

            return (
              <tr key={provider.id} className="border-b border-border-subtle/50">
                <td className="py-3 px-3">
                  <div className="flex items-center gap-2">
                    <ProviderIcon provider={provider.id} size="sm" />
                    <span className="font-medium text-text-primary">{provider.name}</span>
                  </div>
                </td>
                <td className="text-center py-3 px-3 text-text-secondary">{comparison.coldStartMs}ms</td>
                <td className="text-center py-3 px-3">
                  <span
                    className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
                      comparison.pricingTier === 'Free'
                        ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                        : 'bg-bg-secondary text-text-secondary'
                    }`}
                  >
                    {comparison.pricingTier}
                  </span>
                </td>
                <td className="text-center py-3 px-3 text-text-secondary">{comparison.regionCount}</td>
                <td className="py-3 px-3 text-text-secondary text-xs max-w-xs truncate">
                  {comparison.bestFor}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
