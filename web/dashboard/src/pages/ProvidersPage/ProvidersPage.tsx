import { ProviderIcon } from '@/components/common/ProviderIcon';
import { providersApi } from '@/api';
import { StatusBadge } from '@/components/common/StatusBadge';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { ScrollArea } from '@/components/ui/scroll-area';
import { PROVIDERS, PROVIDER_EXTERNAL_DASHBOARD_URL, ROUTES } from '@/lib/constants';
import { useProvidersStore } from '@/stores/providersStore';
import type { ConnectedProvider } from '@/types';
import {
  AlertCircle,
  Check,
  ExternalLink,
  History,
  LayoutGrid,
  List,
  Loader2,
  Maximize2,
  Minimize2,
  Plus,
  RefreshCw,
  RotateCw,
  Search,
  Shield,
  Sparkles,
  Star,
  Trash2,
  X,
} from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { ProviderCard } from './components/ProviderCard';
import { ProviderCardSkeleton } from './components/ProviderCardSkeleton';
import { ConnectDialog } from './components/ConnectDialog';
import { ConnectAWSDialog } from './components/ConnectAWSDialog';
import { DisconnectConfirmationDialog } from './components/DisconnectConfirmationDialog';
import { ProviderSearchFilter } from './components/ProviderSearchFilter';
import { ApiKeyRotationDialog } from './components/ApiKeyRotationDialog';
import { AutoFailoverDialog } from './components/AutoFailoverDialog';
import {
  ConnectionAuditLog,
  generateMockAuditLog,
  AuditLogEntry,
} from './components/ConnectionAuditLog';
import {
  generateMockHealthData,
} from './components/ConnectionHealthSparkline';
import {
  ProviderComparisonTable,
} from './components/ProviderComparisonTooltip';
import {
  getAllProviderConfigs,
  getProviderConfig,
} from './constants/providerMeta';
import type { FailoverConfig } from './components/AutoFailoverDialog';
import type { ProviderConfig } from './constants/providerMeta';
import './styles.css';

// Provider brand colors for accents (keys match PROVIDERS constant IDs)
const providerAccents: Record<string, { border: string; glow: string; text: string }> = {
  workers: { border: '#f48120', glow: 'rgba(244, 129, 32, 0.15)', text: '#f48120' },
  vercel: { border: '#171717', glow: 'rgba(23, 23, 23, 0.15)', text: '#171717' },
  fly: { border: '#7b68ee', glow: 'rgba(123, 104, 238, 0.15)', text: '#7b68ee' },
  deno: { border: '#0a0a0a', glow: 'rgba(10, 10, 10, 0.15)', text: '#3c3c3c' },
  'functionfly-edge': { border: '#f97316', glow: 'rgba(249, 115, 22, 0.25)', text: '#f97316' },
  'aws-lambda': { border: '#FF9900', glow: 'rgba(255, 153, 0, 0.20)', text: '#FF9900' },
};

// Extended provider data with mock stats for demonstration
interface ExtendedProviderData extends ConnectedProvider {
  functionCount?: number;
  healthData?: ReturnType<typeof generateMockHealthData>;
  last24hUptime?: number;
}

type FilterStatus = 'all' | 'connected' | 'available' | 'degraded';
type SortOption = 'name' | 'status' | 'recent' | 'regions';
type ViewMode = 'grid' | 'list';
type DataDensity = 'compact' | 'comfortable' | 'dashboard';

export function ProvidersPage() {
  // State for new features
  const [searchQuery, setSearchQuery] = useState('');
  const [filterStatus, setFilterStatus] = useState<FilterStatus>('all');
  const [sortBy, setSortBy] = useState<SortOption>('name');
  const [viewMode, setViewMode] = useState<ViewMode>('grid');
  const [dataDensity, setDataDensity] = useState<DataDensity>('comfortable');
  const [glassMorphism, setGlassMorphism] = useState(true);
  const [statusGlow, setStatusGlow] = useState(true);
  const [defaultProviderId, setDefaultProviderId] = useState<string | null>(null);
  const [settingDefault, setSettingDefault] = useState<string | null>(null);
  const [isRotating, setIsRotating] = useState(false);
  const [isSavingFailover, setIsSavingFailover] = useState(false);
  const [showAuditLog, setShowAuditLog] = useState(false);
  const [showComparisonTable, setShowComparisonTable] = useState(false);
  const [auditLogEntries] = useState<AuditLogEntry[]>(generateMockAuditLog());

  // Failover configuration
  const [failoverConfig, setFailoverConfig] = useState<FailoverConfig>({
    enabled: false,
    primaryProviderId: null,
    fallbackProviderId: null,
    autoSwitchThreshold: 10,
    switchbackDelay: 15,
  });
  const [failoverDialogOpen, setFailoverDialogOpen] = useState(false);

  // Dialog states
  const [rotationDialogOpen, setRotationDialogOpen] = useState(false);
  const [rotatingProvider, setRotatingProvider] = useState<ProviderConfig | null>(null);

  // Original state from store
  const [disconnecting, setDisconnecting] = useState<string | null>(null);
  const [disconnectConfirmOpen, setDisconnectConfirmOpen] = useState(false);
  const [disconnectingProvider, setDisconnectingProvider] = useState<{
    id: string;
    name: string;
  } | null>(null);
  const [testingProvider, setTestingProvider] = useState<string | null>(null);
  const [connectionTestResults, setConnectionTestResults] = useState<
    Record<string, 'success' | 'error' | null>
  >({});

  const {
    providers,
    error,
    isLoading,
    fetchProviders,
    connectProvider,
    disconnectProvider,
    testConnection,
    clearError,
    startHealthCheckPolling,
  } = useProvidersStore();

  // Extend providers with mock data
  const extendedProviders = useMemo<ExtendedProviderData[]>(() => {
    return providers.map((p) => ({
      ...p,
      functionCount: Math.floor(Math.random() * 15),
      healthData: generateMockHealthData(24, p.status === 'online'),
      last24hUptime: p.status === 'online' ? 99 + Math.random() : 95 + Math.random() * 4,
    }));
  }, [providers]);

  useEffect(() => {
    fetchProviders();
  }, [fetchProviders]);

  // Start health check polling when providers are loaded
  useEffect(() => {
    if (providers.length > 0) {
      const stopPolling = startHealthCheckPolling(5 * 60 * 1000); // 5 minute interval
      return () => {
        stopPolling();
      };
    }
  }, [providers.length, startHealthCheckPolling]);

  // Filter and sort providers
  const filteredProviders = useMemo(() => {
    const allProviders = getAllProviderConfigs();

    return allProviders.filter((provider) => {
      // Search filter
      if (searchQuery) {
        const query = searchQuery.toLowerCase();
        if (!provider.name.toLowerCase().includes(query) &&
            !provider.id.toLowerCase().includes(query)) {
          return false;
        }
      }

      // Status filter
      if (filterStatus !== 'all') {
        const isConnected = providers.some((p) => p.name === provider.id);
        const providerData = extendedProviders.find((p) => p.name === provider.id);
        const isDegraded = providerData?.status === 'degraded' || providerData?.status === 'offline';

        if (filterStatus === 'connected' && !isConnected) return false;
        if (filterStatus === 'available' && isConnected) return false;
        if (filterStatus === 'degraded' && (!isConnected || !isDegraded)) return false;
      }

      return true;
    }).sort((a, b) => {
      switch (sortBy) {
        case 'name':
          return a.name.localeCompare(b.name);
        case 'regions':
          return b.regions.length - a.regions.length;
        case 'status': {
          const aConnected = providers.some((p) => p.name === a.id);
          const bConnected = providers.some((p) => p.name === b.id);
          return Number(bConnected) - Number(aConnected);
        }
        case 'recent': {
          const aData = extendedProviders.find((p) => p.name === a.id);
          const bData = extendedProviders.find((p) => p.name === b.id);
          if (!aData?.lastUsedAt) return 1;
          if (!bData?.lastUsedAt) return -1;
          return new Date(bData.lastUsedAt).getTime() - new Date(aData.lastUsedAt).getTime();
        }
        default:
          return 0;
      }
    });
  }, [searchQuery, filterStatus, sortBy, providers, extendedProviders]);

  // Stats for filter bar
  const connectedCount = providers.length;
  const degradedCount = providers.filter((p) => p.status === 'degraded' || p.status === 'offline').length;
  const totalCount = Object.keys(PROVIDERS).length;
  const availableCount = totalCount - connectedCount;

  const handleConnect = async (providerId: string, key?: string) => {
    const isFunctionFly = providerId === 'functionfly-edge';
    const providerKey = key ?? '';

    try {
      const result = await connectProvider({ providerId, apiKey: providerKey });
      await fetchProviders();
      setConnectionTestResults((prev) => ({ ...prev, [providerId]: null }));
      return result;
    } catch (error) {
      console.error('Failed to connect provider:', error);
      throw error;
    }
  };

  const handleTestConnection = async (providerId: string) => {
    setTestingProvider(providerId);
    setConnectionTestResults((prev) => ({ ...prev, [providerId]: null }));

    try {
      const isSuccess = await testConnection(providerId);
      const status = isSuccess ? 'success' : 'error';
      setConnectionTestResults((prev) => ({ ...prev, [providerId]: status }));

      if (isSuccess) {
        setTimeout(() => {
          setConnectionTestResults((prev) => ({ ...prev, [providerId]: null }));
        }, 3000);
      }
    } catch (error) {
      setConnectionTestResults((prev) => ({ ...prev, [providerId]: 'error' }));
    } finally {
      setTestingProvider(null);
    }
  };

  const handleSetDefault = async (providerId: string) => {
    setSettingDefault(providerId);
    // Simulate API call
    await new Promise((resolve) => setTimeout(resolve, 500));
    setDefaultProviderId(providerId);
    setSettingDefault(null);
  };

  const openDisconnectConfirm = (catalogProviderId: string) => {
    const row = providers.find((p) => p.name === catalogProviderId);
    if (!row) return;
    setDisconnectingProvider({ id: row.id, name: row.name });
    setDisconnectConfirmOpen(true);
  };

  const handleDisconnectConfirm = async () => {
    if (!disconnectingProvider) return;

    setDisconnecting(disconnectingProvider.name);
    clearError();

    try {
      await disconnectProvider(disconnectingProvider.id);
      setDisconnectConfirmOpen(false);
      setDisconnectingProvider(null);
    } catch (error) {
      console.error('Failed to disconnect provider:', error);
    } finally {
      setDisconnecting(null);
    }
  };

  const handleDisconnectCancel = () => {
    setDisconnectConfirmOpen(false);
    setDisconnectingProvider(null);
  };

  const openRotationDialog = (provider: ProviderConfig) => {
    setRotatingProvider(provider);
    setRotationDialogOpen(true);
  };

  const handleRotateKey = async (providerId: string, newApiKey: string) => {
    setIsRotating(true);
    try {
      await providersApi.rotateKey(providerId, newApiKey);
    } finally {
      setIsRotating(false);
    }
  };

  const handleSaveFailover = async (config: FailoverConfig) => {
    setIsSavingFailover(true);
    // Simulate API call
    await new Promise((resolve) => setTimeout(resolve, 800));
    setFailoverConfig(config);
    setIsSavingFailover(false);
  };

  const isConnected = (catalogProviderId: string) =>
    providers.some((p) => p.name === catalogProviderId);

  const getProviderStatus = (catalogProviderId: string) => {
    const connected = providers.find((p) => p.name === catalogProviderId);
    return connected?.status || 'pending';
  };

  const getProviderData = (catalogProviderId: string) => {
    return extendedProviders.find((p) => p.name === catalogProviderId);
  };

  const getAccent = (providerId: string) => {
    return providerAccents[providerId] || {
      border: '#f97316',
      glow: 'rgba(249, 115, 22, 0.25)',
      text: '#f97316',
    };
  };

  const renderConnectDialog = (provider: ProviderConfig, accent: { border: string; glow: string; text: string }) => {
    if (provider.id === 'aws-lambda') {
      return (
        <ConnectAWSDialog
          provider={provider}
          accent={accent}
          onConnect={async (pid, key) => {
            await handleConnect(pid, key);
          }}
        />
      );
    }
    return (
      <ConnectDialog
        provider={provider}
        accent={accent}
        onConnect={async (pid, key) => handleConnect(pid, key)}
      />
    );
  };

  // Get grid columns based on density mode
  const getGridColumns = () => {
    switch (dataDensity) {
      case 'dashboard':
        return 'grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5';
      case 'compact':
        return 'grid-cols-1 lg:grid-cols-2 xl:grid-cols-3';
      case 'comfortable':
      default:
        return 'grid-cols-1 lg:grid-cols-2 xl:grid-cols-3';
    }
  };

  // Density mode icons
  const DensityIcon = () => {
    switch (dataDensity) {
      case 'compact':
        return <Minimize2 className="w-4 h-4" />;
      case 'dashboard':
        return <LayoutGrid className="w-4 h-4" />;
      case 'comfortable':
      default:
        return <Maximize2 className="w-4 h-4" />;
    }
  };

  return (
    <div className="space-y-8 animate-fade-in">
      {/* Header */}
      <div className="provider-header">
        <div>
          <h1 className="provider-title tracking-tight v-section-header">
            Providers
          </h1>
          <p className="provider-subtitle">Connect and manage your deployment targets</p>
        </div>
        <div className="provider-header-actions">
          {/* Data Density Toggle */}
          <div className="provider-density-toggle">
            <Button
              variant={dataDensity === 'compact' ? 'default' : 'ghost'}
              size="sm"
              onClick={() => setDataDensity('compact')}
              className={`provider-density-btn ${dataDensity === 'compact' ? 'provider-density-btn-active' : ''}`}
              title="Compact view"
            >
              <Minimize2 className="w-3.5 h-3.5" />
              <span className="hidden sm:inline text-xs">Compact</span>
            </Button>
            <Button
              variant={dataDensity === 'comfortable' ? 'default' : 'ghost'}
              size="sm"
              onClick={() => setDataDensity('comfortable')}
              className={`provider-density-btn ${dataDensity === 'comfortable' ? 'provider-density-btn-active' : ''}`}
              title="Comfortable view"
            >
              <Maximize2 className="w-3.5 h-3.5" />
              <span className="hidden sm:inline text-xs">Comfort</span>
            </Button>
            <Button
              variant={dataDensity === 'dashboard' ? 'default' : 'ghost'}
              size="sm"
              onClick={() => setDataDensity('dashboard')}
              className={`provider-density-btn ${dataDensity === 'dashboard' ? 'provider-density-btn-active' : ''}`}
              title="Dashboard view"
            >
              <LayoutGrid className="w-3.5 h-3.5" />
              <span className="hidden sm:inline text-xs">Dashboard</span>
            </Button>

            <div className="w-px h-6 bg-border-subtle" />

            <Button
              variant={glassMorphism ? 'default' : 'ghost'}
              size="sm"
              onClick={() => setGlassMorphism(!glassMorphism)}
              className={`provider-status-btn ${glassMorphism ? 'provider-status-btn-active' : ''}`}
              title={glassMorphism ? 'Disable glass effect' : 'Enable glass effect'}
            >
              <Sparkles className="w-3.5 h-3.5" />
              <span className="hidden sm:inline text-xs">Glass</span>
            </Button>

            <Button
              variant={statusGlow ? 'default' : 'ghost'}
              size="sm"
              onClick={() => setStatusGlow(!statusGlow)}
              className={`provider-status-btn ${statusGlow ? 'provider-status-btn-active' : ''}`}
              title={statusGlow ? 'Disable status glow' : 'Enable status glow'}
            >
              <span className="w-2 h-2 rounded-full bg-current animate-pulse" />
              <span className="hidden sm:inline text-xs">Glow</span>
            </Button>
          </div>

          <div className="w-px h-6 bg-border-subtle mx-1" />

          <Button
            variant={showAuditLog ? 'default' : 'ghost'}
            size="sm"
            onClick={() => setShowAuditLog(!showAuditLog)}
            className="gap-2"
          >
            <History className="w-3.5 h-3.5" />
            <span className="hidden sm:inline text-xs">Audit</span>
          </Button>
          <Button
            variant={failoverConfig.enabled ? 'default' : 'ghost'}
            size="sm"
            onClick={() => setFailoverDialogOpen(true)}
            className="gap-2"
          >
            <Shield className="w-3.5 h-3.5" />
            <span className="hidden sm:inline text-xs">Failover</span>
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowComparisonTable(!showComparisonTable)}
            className={`gap-2 ${showComparisonTable ? 'bg-bg-secondary' : ''}`}
          >
            <LayoutGrid className="w-4 h-4" />
            <span className="hidden sm:inline">Compare</span>
          </Button>
        </div>
      </div>

      {/* Search and Filter Bar */}
      <ProviderSearchFilter
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        filterStatus={filterStatus}
        onFilterStatusChange={setFilterStatus}
        sortBy={sortBy}
        onSortChange={setSortBy}
        viewMode={viewMode}
        onViewModeChange={setViewMode}
        connectedCount={connectedCount}
        availableCount={availableCount}
        degradedCount={degradedCount}
        totalCount={totalCount}
      />

      {/* Theme Status Bar */}
      <div className="provider-theme-status">
        <span className="provider-theme-status-item provider-theme-status-item-density">
          <DensityIcon />
          Density: <span className="capitalize">{dataDensity}</span>
        </span>
        {glassMorphism && (
          <span className="provider-theme-status-item provider-theme-status-item-glass">
            <Sparkles className="w-3 h-3" />
            Glass morphism enabled
          </span>
        )}
        {statusGlow && (
          <span className="provider-theme-status-item provider-theme-status-item-glow">
            <span className="w-2 h-2 rounded-full bg-current animate-pulse" />
            Status glow enabled
          </span>
        )}
      </div>

      {/* Provider Comparison Table */}
      {showComparisonTable && (
        <Card className="provider-comparison-card">
          <CardHeader>
            <CardTitle className="provider-comparison-title">
              <LayoutGrid className="w-5 h-5" />
              Provider Comparison
            </CardTitle>
          </CardHeader>
          <CardContent>
            <ProviderComparisonTable providers={getAllProviderConfigs()} />
          </CardContent>
        </Card>
      )}

      {/* Audit Log Panel */}
      {showAuditLog && (
        <Card className="provider-audit-card">
          <CardHeader className="pb-3">
            <div className="flex items-center justify-between">
              <CardTitle className="provider-audit-title">
                <History className="w-5 h-5" />
                Connection Audit Log
              </CardTitle>
              <Button variant="ghost" size="sm" onClick={() => setShowAuditLog(false)}>
                <X className="w-4 h-4" />
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <ConnectionAuditLog entries={auditLogEntries} maxHeight={300} />
          </CardContent>
        </Card>
      )}

      {/* Stats Summary */}
      <div className="provider-stats-summary">
        {isLoading && (
          <div className="provider-stats-loading">
            <Loader2 className="w-4 h-4 animate-spin" />
            <span>Loading...</span>
          </div>
        )}
        <div className="flex items-center gap-4 text-sm">
          <span className="provider-stats-item">
            <span className="provider-status-dot provider-status-dot-online" />
            <span className="text-text-secondary">{connectedCount} connected</span>
          </span>
          <span className="provider-stats-item">
            <span className="provider-status-dot provider-status-dot-degraded" />
            <span className="text-text-secondary">{availableCount} available</span>
          </span>
          {defaultProviderId && (
            <span className="provider-stats-default">
              <Star className="w-3 h-3" />
              <span className="text-xs">
                Default: {getProviderConfig(defaultProviderId)?.name}
              </span>
            </span>
          )}
          {failoverConfig.enabled && (
            <span className="provider-stats-failover">
              <Shield className="w-3 h-3" />
              <span className="text-xs">Auto-failover enabled</span>
            </span>
          )}
        </div>
      </div>

      {/* Error Message with Retry */}
      {error && (
        <div className="provider-error">
          <div className="flex items-start gap-3">
            <AlertCircle className="provider-error-icon" />
            <div className="flex-1">
              <p className="provider-error-title">Failed to load providers</p>
              <p className="provider-error-message">{error}</p>
              <div className="provider-error-actions">
                <Button
                  variant="outline"
                  size="sm"
                  className="provider-error-btn-retry"
                  onClick={() => fetchProviders()}
                  disabled={isLoading}
                >
                  <RefreshCw className={`w-4 h-4 mr-1.5 ${isLoading ? 'animate-spin' : ''}`} />
                  Retry
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  className="provider-error-btn-dismiss"
                  onClick={clearError}
                >
                  Dismiss
                </Button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Providers Grid/List */}
      {viewMode === 'grid' ? (
        <div className={`provider-grid ${getGridColumns()} gap-5`}>
          {isLoading && providers.length === 0
            ? Array.from({ length: 5 }).map((_, i) => <ProviderCardSkeleton key={i} />)
            : filteredProviders.map((provider) => {
                const connected = isConnected(provider.id);
                const status = getProviderStatus(provider.id);
                const providerData = getProviderData(provider.id);
                const accent = getAccent(provider.id);

                return (
                  <ProviderCard
                    key={provider.id}
                    provider={provider}
                    connected={connected}
                    status={status}
                    isDefault={defaultProviderId === provider.id}
                    onSetDefault={connected ? () => handleSetDefault(provider.id) : undefined}
                    onDisconnect={() => openDisconnectConfirm(provider.id)}
                    onConnect={async (pid, key) => handleConnect(pid, key)}
                    onTestConnection={
                      connected ? () => handleTestConnection(provider.id) : undefined
                    }
                    onRotateKey={connected ? () => openRotationDialog(provider) : undefined}
                    isDisconnecting={disconnecting === provider.id}
                    isTestingConnection={testingProvider === provider.id}
                    isSettingDefault={settingDefault === provider.id}
                    lastUsedAt={providerData?.lastUsedAt}
                    isStale={providerData?.isStale}
                    connectionTestResult={connectionTestResults[provider.id]}
                    healthData={providerData?.healthData}
                    last24hUptime={providerData?.last24hUptime}
                    functionCount={providerData?.functionCount}
                    accent={accent}
                    connectDialog={renderConnectDialog(provider, accent)}
                    glassMorphism={glassMorphism}
                    density={dataDensity}
                    statusGlow={statusGlow}
                  />
                );
              })}
        </div>
      ) : (
        // List View
        <Card className="provider-list-card overflow-hidden">
          <ScrollArea className="provider-list-scroll h-auto max-h-[600px]">
            <div className="divide-y divide-border-subtle">
              {filteredProviders.map((provider) => {
                const connected = isConnected(provider.id);
                const status = getProviderStatus(provider.id);
                const providerData = getProviderData(provider.id);
                const accent = getAccent(provider.id);

                return (
                  <div
                    key={provider.id}
                    className="provider-list-item"
                  >
                    <div
                      className="provider-avatar"
                      style={{ backgroundColor: `${accent.border}15` }}
                    >
                      <ProviderIcon provider={provider.id} size="md" />
                    </div>
                    <div className="provider-list-info flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <h4 className="font-medium text-text-primary">{provider.name}</h4>
                        {defaultProviderId === provider.id && (
                          <Badge variant="outline" className="text-xs">
                            <Star className="w-3 h-3 mr-1 fill-amber-400 text-amber-400" />
                            Default
                          </Badge>
                        )}
                        {connected ? (
                          <Badge
                            variant="outline"
                            className="provider-badge-connected text-xs"
                          >
                            Connected
                          </Badge>
                        ) : (
                          <Badge
                            variant="outline"
                            className="provider-badge-not-connected text-xs"
                          >
                            Not Connected
                          </Badge>
                        )}
                      </div>
                      <p className="text-sm text-text-secondary truncate">
                        {provider.regions.length} regions • {provider.regions.slice(0, 3).join(', ')}
                        {provider.regions.length > 3 && ` +${provider.regions.length - 3} more`}
                      </p>
                    </div>
                    <div className="provider-list-actions flex items-center gap-2">
                      {connected ? (
                        <>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleTestConnection(provider.id)}
                            disabled={testingProvider === provider.id}
                          >
                            {testingProvider === provider.id ? (
                              <Loader2 className="w-4 h-4 animate-spin" />
                            ) : (
                              <Check className="w-4 h-4" />
                            )}
                            <span className="ml-2 hidden sm:inline">Test</span>
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => openRotationDialog(provider)}
                          >
                            <RotateCw className="w-4 h-4" />
                            <span className="ml-2 hidden sm:inline">Rotate</span>
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => openDisconnectConfirm(provider.id)}
                            disabled={disconnecting === provider.id}
                            className="text-error hover:text-error hover:bg-error/10"
                          >
                            {disconnecting === provider.id ? (
                              <Loader2 className="w-4 h-4 animate-spin" />
                            ) : (
                              <Trash2 className="w-4 h-4" />
                            )}
                          </Button>
                        </>
                      ) : (
                        renderConnectDialog(provider, accent)
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          </ScrollArea>
        </Card>
      )}

      {/* Disconnect Confirmation Dialog */}
      {disconnectingProvider && (
        <DisconnectConfirmationDialog
          providerName={
            PROVIDERS[disconnectingProvider.name.toUpperCase() as keyof typeof PROVIDERS]?.name ||
            disconnectingProvider.name
          }
          isOpen={disconnectConfirmOpen}
          onClose={handleDisconnectCancel}
          onConfirm={handleDisconnectConfirm}
          isDisconnecting={!!disconnecting}
        />
      )}

      {/* API Key Rotation Dialog */}
      {rotatingProvider && (
        <ApiKeyRotationDialog
          provider={rotatingProvider}
          accent={getAccent(rotatingProvider.id)}
          isOpen={rotationDialogOpen}
          onClose={() => {
            setRotationDialogOpen(false);
            setRotatingProvider(null);
          }}
          onRotate={handleRotateKey}
          isRotating={isRotating}
        />
      )}

      {/* Auto-Failover Dialog */}
      <AutoFailoverDialog
        providers={getAllProviderConfigs()}
        connectedProviderIds={providers.map((p) => p.name)}
        currentConfig={failoverConfig}
        isOpen={failoverDialogOpen}
        onClose={() => setFailoverDialogOpen(false)}
        onSave={handleSaveFailover}
        isSaving={isSavingFailover}
      />
    </div>
  );
}
