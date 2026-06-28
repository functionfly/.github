import { ProviderIcon } from '@/components/common/ProviderIcon';
import { usePageTitle } from '@/hooks';
import { providersApi } from '@/api';
import { PROVIDERS, PROVIDER_EXTERNAL_DASHBOARD_URL, ROUTES } from '@/lib/constants';
import { useProvidersStore } from '@/stores/providersStore';
import type { ConnectedProvider, ProviderMaintenanceStatus } from '@/types';
import {
  AlertCircle,
  History,
  LayoutGrid,
  List,
  Loader2,
  Maximize2,
  Minimize2,
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
import {
  PageGrid,
  Chamber,
  CornerBrace,
  TrustSeal,
  SealedButton,
  FrameButton,
  StatusPill,
  GaugeStrip,
  Gauge,
  AnnotationTag,
  Card,
} from '@/components/containment';
import { ProviderCard } from './components/ProviderCard';
import { ProviderCardSkeleton } from './components/ProviderCardSkeleton';
import { ConnectDialog } from './components/ConnectDialog';
import { ConnectAWSDialog } from './components/ConnectAWSDialog';
import { DisconnectConfirmationDialog } from './components/DisconnectConfirmationDialog';
import { ProviderSearchFilter } from './components/ProviderSearchFilter';
import { ApiKeyRotationDialog } from './components/ApiKeyRotationDialog';
import { AutoFailoverDialog } from './components/AutoFailoverDialog';
import { ConnectionAuditLog, generateMockAuditLog, AuditLogEntry } from './components/ConnectionAuditLog';
import { generateMockHealthData } from './components/ConnectionHealthSparkline';
import { ProviderComparisonTable } from './components/ProviderComparisonTooltip';
import { getAllProviderConfigs, getProviderConfig } from './constants/providerMeta';
import type { FailoverConfig } from './components/AutoFailoverDialog';
import type { ProviderConfig } from './constants/providerMeta';
import './styles.css';

const providerAccents: Record<string, { border: string; glow: string; text: string }> = {
  workers: { border: '#f48120', glow: 'rgba(244, 129, 32, 0.15)', text: '#f48120' },
  vercel: { border: '#171717', glow: 'rgba(23, 23, 23, 0.15)', text: '#171717' },
  fly: { border: '#7b68ee', glow: 'rgba(123, 104, 238, 0.15)', text: '#7b68ee' },
  deno: { border: '#0a0a0a', glow: 'rgba(10, 10, 10, 0.15)', text: '#3c3c3c' },
  'functionfly-edge': { border: '#f97316', glow: 'rgba(249, 115, 22, 0.25)', text: '#f97316' },
  'aws-lambda': { border: '#FF9900', glow: 'rgba(255, 153, 0, 0.20)', text: '#FF9900' },
};

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
  const [searchQuery, setSearchQuery] = useState('');
  const [filterStatus, setFilterStatus] = useState<FilterStatus>('all');
  const [sortBy, setSortBy] = useState<SortOption>('name');
  const [viewMode, setViewMode] = useState<ViewMode>('grid');
  const [dataDensity, setDataDensity] = useState<DataDensity>('comfortable');
  const [defaultProviderId, setDefaultProviderId] = useState<string | null>(null);
  const [settingDefault, setSettingDefault] = useState<string | null>(null);
  const [isRotating, setIsRotating] = useState(false);
  const [isSavingFailover, setIsSavingFailover] = useState(false);
  const [showAuditLog, setShowAuditLog] = useState(false);
  const [showComparisonTable, setShowComparisonTable] = useState(false);
  const [auditLogEntries] = useState<AuditLogEntry[]>(generateMockAuditLog());

  const [failoverConfig, setFailoverConfig] = useState<FailoverConfig>({
    enabled: false, primaryProviderId: null, fallbackProviderId: null, autoSwitchThreshold: 10, switchbackDelay: 15,
  });
  const [failoverDialogOpen, setFailoverDialogOpen] = useState(false);
  const [rotationDialogOpen, setRotationDialogOpen] = useState(false);
  const [rotatingProvider, setRotatingProvider] = useState<ProviderConfig | null>(null);

  const [disconnecting, setDisconnecting] = useState<string | null>(null);
  const [disconnectConfirmOpen, setDisconnectConfirmOpen] = useState(false);
  const [disconnectingProvider, setDisconnectingProvider] = useState<{ id: string; name: string } | null>(null);
  const [testingProvider, setTestingProvider] = useState<string | null>(null);
  const [connectionTestResults, setConnectionTestResults] = useState<Record<string, 'success' | 'error' | null>>({});
  const [maintenanceStatus, setMaintenanceStatus] = useState<Record<string, ProviderMaintenanceStatus>>({});

  const { providers, error, isLoading, fetchProviders, connectProvider, disconnectProvider, testConnection, clearError, startHealthCheckPolling } = useProvidersStore();

  usePageTitle('Providers');

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
    providersApi.getProviderMaintenanceStatus().then(setMaintenanceStatus).catch(console.error);
  }, [fetchProviders]);

  useEffect(() => {
    if (providers.length > 0) {
      const stopPolling = startHealthCheckPolling(5 * 60 * 1000);
      return () => { stopPolling(); };
    }
  }, [providers.length, startHealthCheckPolling]);

  const filteredProviders = useMemo(() => {
    return getAllProviderConfigs().filter((provider) => {
      if (searchQuery) {
        const query = searchQuery.toLowerCase();
        if (!provider.name.toLowerCase().includes(query) && !provider.id.toLowerCase().includes(query)) return false;
      }
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
        case 'name': return a.name.localeCompare(b.name);
        case 'regions': return b.regions.length - a.regions.length;
        case 'status': { const ac = providers.some((p) => p.name === a.id); const bc = providers.some((p) => p.name === b.id); return Number(bc) - Number(ac); }
        case 'recent': { const ad = extendedProviders.find((p) => p.name === a.id); const bd = extendedProviders.find((p) => p.name === b.id); if (!ad?.lastUsedAt) return 1; if (!bd?.lastUsedAt) return -1; return new Date(bd.lastUsedAt).getTime() - new Date(ad.lastUsedAt).getTime(); }
        default: return 0;
      }
    });
  }, [searchQuery, filterStatus, sortBy, providers, extendedProviders]);

  const connectedCount = providers.length;
  const degradedCount = providers.filter((p) => p.status === 'degraded' || p.status === 'offline').length;
  const totalCount = Object.keys(PROVIDERS).length;
  const availableCount = totalCount - connectedCount;

  const handleConnect = async (providerId: string, key?: string) => {
    const maintenance = maintenanceStatus[providerId];
    if (maintenance?.disabled) throw new Error(maintenance.reason || 'This provider is currently under maintenance');
    try {
      const result = await connectProvider({ providerId, apiKey: key ?? '' });
      await fetchProviders();
      setConnectionTestResults((prev) => ({ ...prev, [providerId]: null }));
      return result;
    } catch (error) { console.error('Failed to connect provider:', error); throw error; }
  };

  const handleTestConnection = async (providerId: string) => {
    setTestingProvider(providerId);
    setConnectionTestResults((prev) => ({ ...prev, [providerId]: null }));
    try {
      const isSuccess = await testConnection(providerId);
      setConnectionTestResults((prev) => ({ ...prev, [providerId]: isSuccess ? 'success' : 'error' }));
      if (isSuccess) setTimeout(() => { setConnectionTestResults((prev) => ({ ...prev, [providerId]: null })); }, 3000);
    } catch { setConnectionTestResults((prev) => ({ ...prev, [providerId]: 'error' })); } finally { setTestingProvider(null); }
  };

  const handleSetDefault = async (providerId: string) => {
    setSettingDefault(providerId);
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
    try { await disconnectProvider(disconnectingProvider.id); setDisconnectConfirmOpen(false); setDisconnectingProvider(null); }
    catch (error) { console.error('Failed to disconnect provider:', error); } finally { setDisconnecting(null); }
  };

  const openRotationDialog = (provider: ProviderConfig) => { setRotatingProvider(provider); setRotationDialogOpen(true); };

  const handleRotateKey = async (providerId: string, newApiKey: string) => {
    setIsRotating(true);
    try { await providersApi.rotateKey(providerId, newApiKey); } finally { setIsRotating(false); }
  };

  const handleSaveFailover = async (config: FailoverConfig) => {
    setIsSavingFailover(true);
    await new Promise((resolve) => setTimeout(resolve, 800));
    setFailoverConfig(config);
    setIsSavingFailover(false);
  };

  const isConnected = (catalogProviderId: string) => providers.some((p) => p.name === catalogProviderId);
  const getProviderStatus = (catalogProviderId: string) => providers.find((p) => p.name === catalogProviderId)?.status || 'pending';
  const getProviderData = (catalogProviderId: string) => extendedProviders.find((p) => p.name === catalogProviderId);
  const getAccent = (providerId: string) => providerAccents[providerId] || { border: '#f97316', glow: 'rgba(249, 115, 22, 0.25)', text: '#f97316' };

  const renderConnectDialog = (provider: ProviderConfig, accent: { border: string; glow: string; text: string }) => {
    if (provider.id === 'aws-lambda') return <ConnectAWSDialog provider={provider} accent={accent} onConnect={async (pid, key) => { await handleConnect(pid, key); }} />;
    return <ConnectDialog provider={provider} accent={accent} onConnect={async (pid, key) => (await handleConnect(pid, key)) ?? { success: false }} />;
  };

  const getGridColumns = () => {
    switch (dataDensity) {
      case 'dashboard': return 'provider-grid--dashboard';
      case 'compact': return 'provider-grid--compact';
      default: return 'provider-grid--comfortable';
    }
  };

  return (
    <div className="prov-page">
      <PageGrid />

      {/* Hero */}
      <Chamber className="prov-hero" ribs>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary="MODULE PV-01" secondary="Providers" position="top-right" />

        <div className="prov-hero__header">
          <div className="prov-hero__title-row">
            <TrustSeal size="lg" />
            <h1 className="prov-hero__title">Providers</h1>
          </div>
          <p className="prov-hero__subtitle">Connect and manage your deployment targets</p>
          <div className="prov-hero__actions">
            <div className="prov-density-toggle">
              {(['compact', 'comfortable', 'dashboard'] as const).map((d) => (
                <button key={d} className={`prov-density-btn ${dataDensity === d ? 'prov-density-btn--active' : ''}`} onClick={() => setDataDensity(d)}>
                  {d === 'compact' ? <Minimize2 className="prov-icon-xs" /> : d === 'dashboard' ? <LayoutGrid className="prov-icon-xs" /> : <Maximize2 className="prov-icon-xs" />}
                  <span className="prov-density-label">{d === 'compact' ? 'Compact' : d === 'dashboard' ? 'Dashboard' : 'Comfort'}</span>
                </button>
              ))}
            </div>
            <FrameButton size="sm" onClick={() => setShowAuditLog(!showAuditLog)} iconLeft={<History className="prov-icon-xs" />}>Audit</FrameButton>
            <FrameButton size="sm" onClick={() => setFailoverDialogOpen(true)} iconLeft={<Shield className="prov-icon-xs" />}>Failover</FrameButton>
            <FrameButton size="sm" onClick={() => setShowComparisonTable(!showComparisonTable)} iconLeft={<LayoutGrid className="prov-icon-xs" />}>Compare</FrameButton>
          </div>
        </div>

        <GaugeStrip>
          <Gauge isFirst data={{ value: connectedCount, label: 'Connected' }} />
          <Gauge data={{ value: availableCount, label: 'Available' }} />
          <Gauge data={{ value: degradedCount, label: 'Degraded' }} />
          <Gauge data={{ value: totalCount, label: 'Total Providers' }} />
        </GaugeStrip>
      </Chamber>

      {/* Search and Filter */}
      <ProviderSearchFilter
        searchQuery={searchQuery} onSearchChange={setSearchQuery}
        filterStatus={filterStatus} onFilterStatusChange={setFilterStatus}
        sortBy={sortBy} onSortChange={setSortBy}
        viewMode={viewMode} onViewModeChange={setViewMode}
        connectedCount={connectedCount} availableCount={availableCount} degradedCount={degradedCount} totalCount={totalCount}
      />

      {/* Comparison Table */}
      {showComparisonTable && (
        <Chamber className="prov-comparison">
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <h2 className="prov-section-title"><LayoutGrid className="prov-icon-sm" /> Provider Comparison</h2>
          <ProviderComparisonTable providers={getAllProviderConfigs()} />
        </Chamber>
      )}

      {/* Audit Log */}
      {showAuditLog && (
        <Chamber className="prov-audit">
          <div className="prov-audit__header">
            <h2 className="prov-section-title"><History className="prov-icon-sm" /> Connection Audit Log</h2>
            <button className="prov-close-btn" onClick={() => setShowAuditLog(false)}><X className="prov-icon-sm" /></button>
          </div>
          <ConnectionAuditLog entries={auditLogEntries} maxHeight={300} />
        </Chamber>
      )}

      {/* Error */}
      {error && (
        <div className="prov-error">
          <AlertCircle className="prov-error__icon" />
          <div className="prov-error__content">
            <p className="prov-error__title">Failed to load providers</p>
            <p className="prov-error__message">{error}</p>
            <div className="prov-error__actions">
              <FrameButton size="sm" onClick={() => fetchProviders()} iconLeft={<RefreshCw className={`prov-icon-xs ${isLoading ? 'prov-spin' : ''}`} />}>Retry</FrameButton>
              <button className="prov-dismiss-btn" onClick={clearError}>Dismiss</button>
            </div>
          </div>
        </div>
      )}

      {/* Provider Grid */}
      {viewMode === 'grid' ? (
        <div className={`provider-grid ${getGridColumns()}`}>
          {isLoading && providers.length === 0
            ? Array.from({ length: 5 }).map((_, i) => <ProviderCardSkeleton key={i} />)
            : filteredProviders.map((provider) => {
                const connected = isConnected(provider.id);
                const status = getProviderStatus(provider.id);
                const providerData = getProviderData(provider.id);
                const accent = getAccent(provider.id);
                return (
                  <ProviderCard
                    key={provider.id} provider={provider} connected={connected} status={status}
                    isDefault={defaultProviderId === provider.id}
                    onSetDefault={connected ? () => handleSetDefault(provider.id) : undefined}
                    onDisconnect={() => openDisconnectConfirm(provider.id)}
                    onConnect={async (pid, key) => { await handleConnect(pid, key); }}
                    onTestConnection={connected ? () => handleTestConnection(provider.id) : undefined}
                    onRotateKey={connected ? () => openRotationDialog(provider) : undefined}
                    isDisconnecting={disconnecting === provider.id}
                    isTestingConnection={testingProvider === provider.id}
                    isSettingDefault={settingDefault === provider.id}
                    lastUsedAt={providerData?.lastUsedAt} isStale={providerData?.isStale}
                    connectionTestResult={connectionTestResults[provider.id]}
                    healthData={providerData?.healthData} last24hUptime={providerData?.last24hUptime}
                    functionCount={providerData?.functionCount} accent={accent}
                    connectDialog={renderConnectDialog(provider, accent)}
                  />
                );
              })}
        </div>
      ) : (
        <Chamber className="prov-list-chamber">
          <div className="prov-list">
            {filteredProviders.map((provider) => {
              const connected = isConnected(provider.id);
              const status = getProviderStatus(provider.id);
              const accent = getAccent(provider.id);
              return (
                <div key={provider.id} className="prov-list-item">
                  <div className="prov-list-avatar" style={{ backgroundColor: `${accent.border}15` }}>
                    <ProviderIcon provider={provider.id} size="md" />
                  </div>
                  <div className="prov-list-info">
                    <div className="prov-list-name-row">
                      <h4 className="prov-list-name">{provider.name}</h4>
                      {defaultProviderId === provider.id && <StatusPill status="pending" label="Default" />}
                      <StatusPill status={connected ? 'live' : 'pending'} label={connected ? 'Connected' : 'Not Connected'} />
                    </div>
                    <p className="prov-list-meta">{provider.regions.length} regions &middot; {provider.regions.slice(0, 3).join(', ')}{provider.regions.length > 3 ? ` +${provider.regions.length - 3} more` : ''}</p>
                  </div>
                  <div className="prov-list-actions">
                    {connected ? (
                      <>
                        <FrameButton size="sm" onClick={() => handleTestConnection(provider.id)} disabled={testingProvider === provider.id}>
                          {testingProvider === provider.id ? <Loader2 className="prov-icon-xs prov-spin" /> : 'Test'}
                        </FrameButton>
                        <FrameButton size="sm" onClick={() => openRotationDialog(provider)}>Rotate</FrameButton>
                        <button className="prov-delete-btn" onClick={() => openDisconnectConfirm(provider.id)} disabled={disconnecting === provider.id}>
                          {disconnecting === provider.id ? <Loader2 className="prov-icon-xs prov-spin" /> : <Trash2 className="prov-icon-xs" />}
                        </button>
                      </>
                    ) : renderConnectDialog(provider, accent)}
                  </div>
                </div>
              );
            })}
          </div>
        </Chamber>
      )}

      {/* Dialogs */}
      {disconnectingProvider && (
        <DisconnectConfirmationDialog
          providerName={PROVIDERS[disconnectingProvider.name.toUpperCase() as keyof typeof PROVIDERS]?.name || disconnectingProvider.name}
          isOpen={disconnectConfirmOpen} onClose={() => { setDisconnectConfirmOpen(false); setDisconnectingProvider(null); }}
          onConfirm={handleDisconnectConfirm} isDisconnecting={!!disconnecting}
        />
      )}
      {rotatingProvider && (
        <ApiKeyRotationDialog provider={rotatingProvider} accent={getAccent(rotatingProvider.id)} isOpen={rotationDialogOpen}
          onClose={() => { setRotationDialogOpen(false); setRotatingProvider(null); }} onRotate={handleRotateKey} isRotating={isRotating} />
      )}
      <AutoFailoverDialog providers={getAllProviderConfigs()} connectedProviderIds={providers.map((p) => p.name)} currentConfig={failoverConfig}
        isOpen={failoverDialogOpen} onClose={() => setFailoverDialogOpen(false)} onSave={handleSaveFailover} isSaving={isSavingFailover} />
    </div>
  );
}
