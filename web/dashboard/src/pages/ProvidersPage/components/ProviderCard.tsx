import {
  AlertCircle,
  AlertTriangle,
  Check,
  Clock,
  ExternalLink,
  FunctionSquare,
  Loader2,
  RefreshCw,
  Shield,
  X,
  Zap,
  Settings,
  Star,
} from 'lucide-react';
import { Link } from 'react-router-dom';
import { ProviderIcon } from '@/components/common/ProviderIcon';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { ConnectionStatus } from './ConnectionStatus';
import { RegionBadges } from './RegionBadges';
import { CapabilityBadges } from './CapabilityBadges';
import { HealthSparklineCard, type HealthDataPoint } from './ConnectionHealthSparkline';
import { ProviderComparisonTooltip } from './ProviderComparisonTooltip';
import { PROVIDER_EXTERNAL_DASHBOARD_URL, ROUTES } from '@/lib/constants';
import type { ProviderConfig } from '../constants/providerMeta';

interface ProviderCardProps {
  provider: ProviderConfig;
  connected: boolean;
  status: string;
  isDefault?: boolean;
  onSetDefault?: () => void;
  onDisconnect: () => void;
  onConnect: (providerId: string, apiKey?: string) => Promise<void>;
  onTestConnection?: () => Promise<void>;
  onRotateKey?: () => void;
  isDisconnecting: boolean;
  isTestingConnection?: boolean;
  isSettingDefault?: boolean;
  lastUsedAt?: string;
  isStale?: boolean;
  connectionTestResult?: 'success' | 'error' | null;
  healthData?: HealthDataPoint[];
  last24hUptime?: number;
  functionCount?: number;
  accent: { border: string; glow: string; text: string };
  connectDialog: React.ReactNode;
  /** Glass morphism effect */
  glassMorphism?: boolean;
  /** Data density mode */
  density?: 'compact' | 'comfortable' | 'dashboard';
  /** Enable status-based glow effect */
  statusGlow?: boolean;
}

function formatRelativeTime(dateString?: string): string | null {
  if (!dateString) return null;
  const date = new Date(dateString);
  const now = new Date();
  const diffInMs = now.getTime() - date.getTime();
  const diffInDays = Math.floor(diffInMs / (1000 * 60 * 60 * 24));
  const diffInHours = Math.floor(diffInMs / (1000 * 60 * 60));
  const diffInMinutes = Math.floor(diffInMs / (1000 * 60));

  if (diffInMinutes < 1) return 'Just now';
  if (diffInMinutes < 60) return `${diffInMinutes}m ago`;
  if (diffInHours < 24) return `${diffInHours}h ago`;
  if (diffInDays < 7) return `${diffInDays}d ago`;
  return date.toLocaleDateString();
}

export function ProviderCard({
  provider,
  connected,
  status,
  isDefault = false,
  onSetDefault,
  onDisconnect,
  onConnect,
  onTestConnection,
  onRotateKey,
  isDisconnecting,
  isTestingConnection,
  isSettingDefault,
  lastUsedAt,
  isStale,
  connectionTestResult,
  healthData,
  last24hUptime,
  functionCount,
  accent,
  connectDialog,
  glassMorphism = false,
  density = 'comfortable',
  statusGlow = true,
}: ProviderCardProps) {
  const isFunctionFly = provider.id === 'functionfly-edge';

  // Get status class for glow effect
  const getStatusGlowClass = () => {
    if (!statusGlow || !connected) return '';

    switch (status) {
      case 'online':
        return 'status-online';
      case 'degraded':
        return 'status-degraded';
      case 'offline':
        return 'status-offline';
      default:
        return 'status-pending';
    }
  };

  // Get provider brand accent class
  const getProviderAccentClass = () => {
    const accentMap: Record<string, string> = {
      workers: 'provider-accent-workers',
      vercel: 'provider-accent-vercel',
      fly: 'provider-accent-fly',
      deno: 'provider-accent-deno',
      'functionfly-edge': 'provider-accent-functionfly',
    };
    return accentMap[provider.id] || '';
  };

  // Calculate density-based classes
  const getDensityClasses = () => {
    const baseClasses = 'provider-card';

    switch (density) {
      case 'compact':
        return `${baseClasses} provider-density-compact`;
      case 'dashboard':
        return `${baseClasses} provider-density-dashboard relative`;
      case 'comfortable':
      default:
        return `${baseClasses} provider-density-comfortable`;
    }
  };

  // Get card style classes
  const getCardClasses = () => {
    const classes = [
      'group',
      'relative',
      'transition-all',
      'duration-300',
      getDensityClasses(),
      getProviderAccentClass(),
      getStatusGlowClass(),
    ];

    if (glassMorphism) {
      classes.push('provider-card-glass');
    } else if (statusGlow) {
      classes.push('provider-card-glow');
    }

    if (connected) {
      classes.push('provider-status-strip');
    }

    if (!connected) {
      classes.push('opacity-90 hover:opacity-100');
    }

    if (isDefault) {
      classes.push('ring-2 ring-amber-400/50');
    }

    return classes.join(' ');
  };

  // Get content padding based on density
  const getContentPadding = () => {
    switch (density) {
      case 'compact':
        return 'p-4';
      case 'dashboard':
        return 'p-3';
      case 'comfortable':
      default:
        return 'p-6';
    }
  };

  // Get icon size based on density
  const getIconSize = () => {
    switch (density) {
      case 'compact':
        return 'w-10 h-10';
      case 'dashboard':
        return 'w-8 h-8';
      case 'comfortable':
      default:
        return 'w-14 h-14';
    }
  };

  // Get icon container size
  const getIconContainerClass = () => {
    const baseClass = 'rounded-2xl flex items-center justify-center transition-transform duration-300 group-hover:scale-105';
    return `${getIconSize()} ${baseClass}`;
  };

  return (
    <Card
      className={getCardClasses()}
      style={{
        borderColor: connected ? `${accent.border}40` : undefined,
        '--provider-brand-color': accent.border,
        '--provider-glow': accent.glow,
      } as React.CSSProperties}
    >
      {/* Brand accent top border - Aviation Status Strip style */}
      <div
        className="absolute top-0 left-0 right-0 h-1 transition-all duration-300 provider-brand-accent"
        style={{
          backgroundColor: accent.border,
          opacity: connected ? 1 : 0.3,
          boxShadow: connected ? `0 0 10px ${accent.glow}` : 'none',
        }}
      />

      {/* Default Provider Indicator */}
      {isDefault && (
        <div className="absolute top-3 right-3 z-10">
          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400 text-xs font-medium">
            <Star className="w-3 h-3 fill-current" />
            Default
          </span>
        </div>
      )}

      {/* Status indicator dot for dashboard density */}
      {density === 'dashboard' && connected && (
        <div
          className={`absolute top-2 right-2 w-2 h-2 rounded-full status-dot-runway ${status}`}
          style={{ zIndex: 10 }}
        />
      )}

      {/* Subtle brand glow on hover */}
      <div
        className="absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity duration-500 pointer-events-none"
        style={{
          background: `radial-gradient(600px circle at 50% 0%, ${accent.glow}, transparent 50%)`,
        }}
      />

      <CardContent className={`relative ${getContentPadding()}`}>
        {/* Header: Icon + Name + Status */}
        <div className={`flex items-start justify-between ${density === 'dashboard' ? 'mb-2' : 'mb-5'}`}>
          <ProviderComparisonTooltip provider={provider}>
            <div className={`flex items-center gap-4 cursor-help ${density === 'dashboard' ? 'flex-row' : ''}`}>
              <div
                className={getIconContainerClass()}
                style={{
                  backgroundColor: `${accent.border}15`,
                }}
              >
                <ProviderIcon provider={provider.id} size={density === 'dashboard' ? 'sm' : 'lg'} />
              </div>
              <div>
                <h3 className={`font-semibold text-text-primary mb-0.5 ${density === 'dashboard' ? 'text-sm' : 'text-lg'}`}>
                  {provider.name}
                </h3>
                {density !== 'dashboard' && (
                  <ConnectionStatus connected={connected} status={status} />
                )}
              </div>
            </div>
          </ProviderComparisonTooltip>
        </div>

        {/* Capability Badges - Hidden in dashboard mode */}
        {density !== 'dashboard' && (
          <div className="mb-4">
            <CapabilityBadges providerId={provider.id} maxDisplay={density === 'compact' ? 2 : 3} />
          </div>
        )}

        {/* Description for FunctionFly Edge - Hidden in compact/dashboard */}
        {isFunctionFly && density === 'comfortable' && (
          <p className="text-sm text-text-secondary mb-4 leading-relaxed">
            {provider.description ||
              "Host your edge functions on FunctionFly's infrastructure — no deployment required"}
          </p>
        )}

        {/* Visual Region Badges - Minimal in compact, hidden in dashboard */}
        {density === 'comfortable' && (
          <div className="mb-4">
            <RegionBadges regions={provider.regions} providerId={provider.id} />
          </div>
        )}
        {density === 'compact' && provider.regions.length > 0 && (
          <div className="mb-3 text-xs text-text-muted">
            {provider.regions.slice(0, 2).join(', ')}
            {provider.regions.length > 2 && ` +${provider.regions.length - 2}`}
          </div>
        )}

        {/* Health Bar for dashboard density */}
        {density === 'dashboard' && connected && healthData && (
          <div className="provider-health-bar mt-2">
            <div
              className="provider-health-bar-fill"
              style={{
                width: `${last24hUptime || 0}%`,
                backgroundColor: (last24hUptime || 0) > 95 ? '#00FF9D' : (last24hUptime || 0) > 90 ? '#FFB800' : '#FF2D55',
                boxShadow: `0 0 8px ${(last24hUptime || 0) > 95 ? 'rgba(0, 255, 157, 0.5)' : (last24hUptime || 0) > 90 ? 'rgba(255, 184, 0, 0.5)' : 'rgba(255, 45, 85, 0.5)'}`,
              }}
            />
          </div>
        )}

        {/* Connection Details (only when connected, hidden in dashboard) */}
        {connected && density === 'comfortable' && (
          <div className="mb-5 p-3 rounded-lg bg-bg-secondary/50 border border-border-subtle">
            <div className="flex items-center gap-2 text-sm">
              <Zap className="w-4 h-4" style={{ color: accent.text }} />
              <span className="text-text-secondary">
                Ready to deploy • Primary region:{' '}
                <span className="font-medium text-text-primary">{provider.regions[0]}</span>
              </span>
            </div>
            {functionCount !== undefined && functionCount > 0 && (
              <div className="mt-2 pt-2 border-t border-border-subtle/50 flex items-center gap-2 text-sm">
                <span className="text-text-secondary">Active functions:</span>
                <span className="font-medium text-text-primary">{functionCount}</span>
              </div>
            )}
          </div>
        )}

        {/* Compact metrics */}
        {connected && density === 'compact' && (
          <div className="mb-3 flex items-center gap-3 text-xs text-text-muted">
            {functionCount !== undefined && (
              <span>{functionCount} functions</span>
            )}
            {last24hUptime !== undefined && (
              <span className={last24hUptime > 95 ? 'text-taxiway' : 'text-beacon'}>
                {last24hUptime.toFixed(1)}% uptime
              </span>
            )}
          </div>
        )}

        {/* Health Sparkline - Hidden in compact/dashboard */}
        {connected && healthData && density === 'comfortable' && (
          <div className="mb-4 p-3 rounded-lg bg-bg-secondary/30 border border-border-subtle/50">
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs text-text-tertiary uppercase tracking-wide">24h Health</span>
              {last24hUptime !== undefined && (
                <span
                  className={`text-xs font-medium ${
                    last24hUptime > 99 ? 'text-taxiway' : last24hUptime > 95 ? 'text-beacon' : 'text-afterburner'
                  }`}
                >
                  {last24hUptime.toFixed(1)}% uptime
                </span>
              )}
            </div>
            <HealthSparklineCard providerId={provider.id} data={healthData} />
          </div>
        )}

        {/* Stale Connection Warning - Only in comfortable */}
        {connected && (status === 'degraded' || isStale) && density === 'comfortable' && (
          <div className="mb-5 p-3 rounded-lg bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-800/50">
            <div className="flex items-start gap-2">
              <AlertTriangle className="w-4 h-4 text-amber-600 dark:text-amber-400 mt-0.5 shrink-0" />
              <div className="text-sm">
                <p className="text-amber-800 dark:text-amber-300 font-medium">
                  Connection may be stale
                </p>
                <p className="text-amber-700 dark:text-amber-400 text-xs mt-0.5">
                  This provider hasn't been used recently or the connection test failed. Click
                  "Test" to verify your credentials.
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Last Used Timestamp - Hidden in dashboard */}
        {connected && lastUsedAt && density !== 'dashboard' && (
          <div className="mb-5 flex items-center gap-2 text-xs text-text-tertiary">
            <Clock className="w-3.5 h-3.5" />
            <span>Last used {formatRelativeTime(lastUsedAt)}</span>
          </div>
        )}

        {/* Connection Test Result - Only in comfortable */}
        {connectionTestResult && density === 'comfortable' && (
          <div
            className={`mb-4 p-2.5 rounded-lg text-sm animate-in slide-in-from-top-1 duration-200 ${
              connectionTestResult === 'success'
                ? 'bg-emerald-50 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-800/50'
                : 'bg-error/10 border border-error/20'
            }`}
          >
            <div className="flex items-center gap-2">
              {connectionTestResult === 'success' ? (
                <>
                  <Check className="w-4 h-4 text-taxiway" />
                  <span className="text-taxiway font-medium">
                    Connection successful
                  </span>
                </>
              ) : (
                <>
                  <AlertCircle className="w-4 h-4 text-error" />
                  <span className="text-error font-medium">Connection test failed</span>
                </>
              )}
            </div>
          </div>
        )}

        {/* Action Buttons */}
        <div className={`flex flex-wrap gap-2 ${density === 'dashboard' ? 'provider-actions' : ''}`}>
          {connected ? (
            <>
              {/* Test Connection Button */}
              {onTestConnection && density !== 'dashboard' && (
                <Button
                  variant="outline"
                  size="sm"
                  className={`flex-1 gap-1.5 border-border-default hover:border-border-subtle hover:bg-bg-secondary transition-colors ${
                    connectionTestResult === 'success'
                      ? 'border-taxiway/30 bg-taxiway/5'
                      : connectionTestResult === 'error'
                        ? 'border-error/30 bg-error/5'
                        : ''
                  } ${density === 'compact' ? 'text-xs py-1' : ''}`}
                  onClick={onTestConnection}
                  disabled={isTestingConnection}
                  title="Test provider connection"
                >
                  {isTestingConnection ? (
                    <Loader2 className="w-3.5 h-3.5 animate-spin" />
                  ) : connectionTestResult === 'success' ? (
                    <Check className="w-3.5 h-3.5 text-taxiway" />
                  ) : (
                    <Shield className="w-3.5 h-3.5" />
                  )}
                  Test
                </Button>
              )}

              {/* Set Default Button (if not default) */}
              {onSetDefault && !isDefault && density !== 'dashboard' && (
                <Button
                  variant="outline"
                  size="sm"
                  className={`flex-1 gap-1.5 border-border-default hover:border-beacon/50 hover:bg-beacon/10 transition-colors ${density === 'compact' ? 'text-xs py-1' : ''}`}
                  onClick={onSetDefault}
                  disabled={isSettingDefault}
                >
                  {isSettingDefault ? (
                    <Loader2 className="w-3.5 h-3.5 animate-spin" />
                  ) : (
                    <Star className="w-3.5 h-3.5" />
                  )}
                  Set Default
                </Button>
              )}

              {/* Configure Button */}
              {isFunctionFly ? (
                <Button
                  variant="outline"
                  size="sm"
                  className={`flex-1 gap-1.5 border-border-default hover:border-border-subtle hover:bg-bg-secondary transition-colors ${density === 'compact' ? 'text-xs py-1' : ''}`}
                  asChild
                >
                  <Link to={ROUTES.FUNCTIONS}>
                    <FunctionSquare className="w-3.5 h-3.5" />
                    {density !== 'compact' && 'Configure'}
                  </Link>
                </Button>
              ) : (
                <Button
                  variant="outline"
                  size="sm"
                  className={`flex-1 gap-1.5 border-border-default hover:border-border-subtle hover:bg-bg-secondary transition-colors ${density === 'compact' ? 'text-xs py-1' : ''}`}
                  asChild
                >
                  <a
                    href={
                      PROVIDER_EXTERNAL_DASHBOARD_URL[
                        provider.id as keyof typeof PROVIDER_EXTERNAL_DASHBOARD_URL
                      ]
                    }
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    <ExternalLink className="w-3.5 h-3.5" />
                    {density !== 'compact' && 'Configure'}
                  </a>
                </Button>
              )}

              {/* Rotate Key Button (only for non-managed providers) */}
              {!isFunctionFly && onRotateKey && density !== 'dashboard' && (
                <Button
                  variant="outline"
                  size="icon"
                  className="border-border-default text-text-secondary hover:text-beacon hover:border-beacon/50 hover:bg-beacon/10 transition-colors shrink-0"
                  onClick={onRotateKey}
                  title="Rotate API key"
                  aria-label={`Rotate API key for ${provider.name}`}
                >
                  <Settings className="w-4 h-4" />
                </Button>
              )}

              {/* Disconnect Button */}
              <Button
                variant="outline"
                size={density === 'compact' ? 'sm' : 'icon'}
                className={`border-border-default text-text-secondary hover:text-error hover:border-error/30 hover:bg-error/5 transition-colors shrink-0 ${density === 'compact' ? 'text-xs' : ''}`}
                onClick={onDisconnect}
                disabled={isDisconnecting}
                title="Disconnect provider"
                aria-label={`Disconnect ${provider.name} provider`}
              >
                {isDisconnecting ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  <>
                    <X className="w-4 h-4" />
                    {density === 'compact' && 'Disconnect'}
                  </>
                )}
              </Button>
            </>
          ) : (
            connectDialog
          )}
        </div>
      </CardContent>
    </Card>
  );
}
