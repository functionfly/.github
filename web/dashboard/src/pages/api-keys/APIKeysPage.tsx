import { APIKeyList, APIKeyRotationModal, CreateAPIKeyModal } from '@/components/api-keys';
import { VaultSetupDialog } from '@/components/api-keys/VaultSetupDialog';
import { Chamber } from '@/components/ui/Chamber';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { getApiBaseUrl } from '@/lib/constants';
import { apiKeysService, getStoredApiKey } from '@/services/api-keys';
import { APIKey, APIKeyFilters, DEFAULT_RATE_LIMIT } from '@/types/api-key';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  AlertTriangle,
  BookOpen,
  Check,
  ChevronDown,
  ChevronRight,
  Copy,
  Key,
  Plus,
  Shield,
} from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { toast } from 'sonner';
import './styles.css';

const PAGE_SIZE = 10;
const EXPIRING_DAYS = 30;

export function APIKeysPage() {
  const queryClient = useQueryClient();
  // Default to active keys only so soft-deleted (revoked) keys don't reappear after refresh
  const [filters, setFilters] = useState<APIKeyFilters>({ is_active: true });
  const [page, setPage] = useState(1);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showRotationModal, setShowRotationModal] = useState(false);
  const [selectedKey, setSelectedKey] = useState<APIKey | null>(null);
  const [deleteKey, setDeleteKey] = useState<APIKey | null>(null);
  const [deletedIds, setDeletedIds] = useState<Set<string>>(() => new Set());
  const [usageOpen, setUsageOpen] = useState(false);
  const [securityOpen, setSecurityOpen] = useState(true);
  const [copiedSnippet, setCopiedSnippet] = useState(false);
  const [showVaultSetup, setShowVaultSetup] = useState(false);

  const toggleUsage = () => setUsageOpen((v) => !v);
  const toggleSecurity = () => setSecurityOpen((v) => !v);

  useEffect(() => {
    const handleOpenVaultSetup = () => setShowVaultSetup(true);
    window.addEventListener('openVaultSetup', handleOpenVaultSetup);
    return () => window.removeEventListener('openVaultSetup', handleOpenVaultSetup);
  }, []);

  // Check for stored newly created API key
  useEffect(() => {
    const stored = getStoredApiKey();
    if (stored) {
      toast.success('API key created successfully', {
        description: "Copy your new API key now. It won't be shown again.",
        duration: 10000,
      });
    }
  }, []);

  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: ['api-keys', filters, page],
    queryFn: () => apiKeysService.listKeys(filters, page, PAGE_SIZE),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => apiKeysService.deleteKey(id),
    onSuccess: (_data, deletedId) => {
      const idStr = String(deletedId);
      setDeletedIds((prev) => new Set(prev).add(idStr));
      queryClient.invalidateQueries({ queryKey: ['api-keys'] });
      toast.success('API key deleted');
      setDeleteKey(null);
    },
    onError: (error) => {
      toast.error('Failed to delete API key', {
        description: error instanceof Error ? error.message : 'Unknown error',
      });
    },
  });

  const rawKeys = data?.data ?? [];
  const total = data?.total ?? 0;
  const apiKeys = useMemo(
    () => rawKeys.filter((k) => !deletedIds.has(String(k.id))),
    [rawKeys, deletedIds]
  );
  const displayTotal = Math.max(0, total - deletedIds.size);

  const stats = useMemo(() => {
    const now = Date.now();
    const expiringThreshold = now + EXPIRING_DAYS * 24 * 60 * 60 * 1000;
    let active = 0;
    let inactive = 0;
    let expiringSoon = 0;
    for (const k of apiKeys) {
      if (k.is_active) active++;
      else inactive++;
      if (k.expires_at && new Date(k.expires_at).getTime() <= expiringThreshold) expiringSoon++;
    }
    return { active, inactive, expiringSoon };
  }, [apiKeys]);

  const handleFiltersChange = (newFilters: APIKeyFilters) => {
    setFilters(newFilters);
    setPage(1);
  };

  const handlePageChange = (newPage: number) => {
    setPage(newPage);
  };

  const handleRotate = (key: APIKey) => {
    setSelectedKey(key);
    setShowRotationModal(true);
  };

  const selectedIds = useMemo(() => new Set(deleteKey ? [deleteKey.id] : []), [deleteKey]);

  const handleDelete = (key: APIKey) => {
    setDeleteKey(key);
  };

  const confirmDelete = () => {
    if (deleteKey) deleteMutation.mutate(deleteKey.id);
  };

  // Bulk delete: iterate single-delete to reuse the existing handler (with
  // admin role checks, rate limiting, and audit logging from the backend).
  const bulkDeleteMutation = useMutation({
    mutationFn: async (ids: string[]) => {
      // Run sequentially so the per-user rate limiter on the backend isn't
      // tripped (30 req/min window). Each individual delete already has full
      // RBAC + audit trail, so this is safe.
      for (const id of ids) {
        await apiKeysService.deleteKey(id);
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['api-keys'] });
      toast.success('Selected API keys deleted');
    },
    onError: (err) => {
      toast.error('Failed to delete some keys', {
        description: err instanceof Error ? err.message : 'Unknown error',
      });
    },
  });

  // Bulk rotate: same reasoning as bulk delete.
  const bulkRotateMutation = useMutation({
    mutationFn: async (ids: string[]) => {
      for (const id of ids) {
        await apiKeysService.rotateKey(id);
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['api-keys'] });
      toast.success('Selected API keys rotated');
    },
    onError: (err) => {
      toast.error('Failed to rotate some keys', {
        description: err instanceof Error ? err.message : 'Unknown error',
      });
    },
  });

  const apiBase = getApiBaseUrl() || (typeof window !== 'undefined' ? window.location.origin : '');
  const curlExample = `curl -X GET "${apiBase}/v1/functions" \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -H "Content-Type: application/json"`;

  const copyCurl = async () => {
    try {
      await navigator.clipboard.writeText(curlExample);
      setCopiedSnippet(true);
      toast.success('Copied to clipboard');
      setTimeout(() => setCopiedSnippet(false), 2000);
    } catch {
      toast.error('Failed to copy. Your browser may block clipboard access.');
    }
  };

  return (
    <div className="apikeys-page">
      {/* Page header */}
      <Chamber className="apikeys-header" ribs>
        <div className="apikeys-header-inner">
          <div>
            <div className="apikeys-breadcrumb">
              <Link to="/dashboard" className="apikeys-breadcrumb-link">
                Dashboard
              </Link>
              <span className="apikeys-breadcrumb-sep">/</span>
              <span>API Keys</span>
            </div>
            <h1 className="apikeys-title">API Keys</h1>
            <p className="apikeys-subtitle">
              Create and manage API keys for programmatic access. Use keys in headers or environment
              variables; rotate them regularly and revoke if compromised.
            </p>
          </div>
          <div className="apikeys-header-actions">
            <button
              onClick={() => setShowCreateModal(true)}
              className="sealed-button sealed-button--lg"
            >
              <Plus className="sealed-button__icon sealed-button__icon--left" />
              Create API Key
            </button>
          </div>
        </div>
      </Chamber>

      {/* Stats row */}
      <div className="gauge-strip">
        <div className="gauge gauge--first">
          <div className="gauge__value">
            <span className="gauge__dot" />
            {displayTotal}
          </div>
          <div className="gauge__label">Total keys</div>
          <p className="gauge__sublabel">Across all pages</p>
        </div>
        <div className="gauge">
          <div className="gauge__value gauge__value--success">
            <span className="gauge__dot" />
            {stats.active}
          </div>
          <div className="gauge__label">Active</div>
          <p className="gauge__sublabel">Can be used for requests</p>
        </div>
        <div className="gauge">
          <div className="gauge__value gauge__value--warning">{stats.inactive}</div>
          <div className="gauge__label">Inactive</div>
          <p className="gauge__sublabel">Revoked or disabled</p>
        </div>
        <div className="gauge">
          <div className="gauge__value gauge__value--accent">{stats.expiringSoon}</div>
          <div className="gauge__label">Expiring soon</div>
          <p className="gauge__sublabel">Within {EXPIRING_DAYS} days</p>
        </div>
      </div>

      {/* How to use & Security */}
      <div className="apikeys-grid-2">
        <Chamber className="apikey-card" annotation="USAGE / HOW TO">
          <button
            type="button"
            onClick={toggleUsage}
            className="apikey-card-toggle"
            aria-expanded={usageOpen}
          >
            <div className="apikey-card-header">
              <div className="flex items-center gap-2">
                <BookOpen className="w-5 h-5" />
                <span className="apikey-card-title">How to use your API key</span>
              </div>
              {usageOpen ? (
                <ChevronDown className="w-4 h-4" />
              ) : (
                <ChevronRight className="w-4 h-4" />
              )}
            </div>
          </button>
          {usageOpen && (
            <div className="apikey-card-content">
              <p className="apikey-card-text">
                Send your API key in the{' '}
                <code className="apikey-code">Authorization: Bearer {'<key>'}</code> header.
              </p>
              <div className="apikey-code-block">
                <button
                  onClick={(e) => {
                    e.preventDefault();
                    copyCurl();
                  }}
                  className="frame-button frame-button--sm apikey-copy-btn"
                >
                  {copiedSnippet ? (
                    <Check className="w-3.5 h-3.5" />
                  ) : (
                    <Copy className="w-3.5 h-3.5" />
                  )}
                </button>
                <pre className="apikey-pre">
                  <code>{curlExample}</code>
                </pre>
              </div>
              <p className="apikey-card-text apikey-card-text--sm">
                Default rate limits: {DEFAULT_RATE_LIMIT.rpm.toLocaleString()} RPM,{' '}
                {DEFAULT_RATE_LIMIT.rph.toLocaleString()} RPH. Adjust when creating a key.
              </p>
            </div>
          )}
        </Chamber>

        <Chamber className="apikey-card" annotation="SECURITY / BEST PRACTICES">
          <button
            type="button"
            onClick={toggleSecurity}
            className="apikey-card-toggle"
            aria-expanded={securityOpen}
          >
            <div className="apikey-card-header">
              <div className="flex items-center gap-2">
                <Shield className="w-5 h-5" />
                <span className="apikey-card-title">Security best practices</span>
              </div>
              {securityOpen ? (
                <ChevronDown className="w-4 h-4" />
              ) : (
                <ChevronRight className="w-4 h-4" />
              )}
            </div>
          </button>
          {securityOpen && (
            <div className="apikey-card-content">
              <ul className="apikey-security-list">
                <li>Never commit keys to git or share them in public.</li>
                <li>Use environment variables or a secrets manager in production.</li>
                <li>Rotate keys periodically and immediately if compromised.</li>
                <li>Create separate keys per environment (e.g. staging vs production).</li>
                <li>Prefer scoped keys (e.g. function type) over platform keys when possible.</li>
              </ul>
            </div>
          )}
        </Chamber>
      </div>

      {/* Keys table */}
      <Chamber className="apikey-table-chamber" corners={['tl', 'br']}>
        <div className="apikey-table-header">
          <div className="flex items-center gap-2">
            <Key className="w-5 h-5" />
            <h2 className="apikey-table-title">Your API Keys</h2>
          </div>
          <p className="apikey-table-description">
            Search, filter, and manage keys. Click a key to edit permissions and environments.
          </p>
        </div>
        <APIKeyList
          apiKeys={apiKeys}
          isLoading={isLoading}
          isError={isError}
          error={error as Error | null}
          onRetry={() => refetch()}
          total={displayTotal}
          page={page}
          pageSize={PAGE_SIZE}
          filters={filters}
          onFiltersChange={handleFiltersChange}
          onPageChange={handlePageChange}
          onCreateNew={() => setShowCreateModal(true)}
          onRotate={handleRotate}
          onDelete={handleDelete}
          onBulkDelete={(ids) => bulkDeleteMutation.mutate(ids)}
          onBulkRotate={(ids) => bulkRotateMutation.mutate(ids)}
        />
      </Chamber>

      <CreateAPIKeyModal
        open={showCreateModal}
        onOpenChange={setShowCreateModal}
        onSuccess={() => queryClient.invalidateQueries({ queryKey: ['api-keys'] })}
      />

      <APIKeyRotationModal
        open={showRotationModal}
        onOpenChange={setShowRotationModal}
        apiKey={selectedKey}
        onSuccess={() => {
          refetch();
          toast.success('API key rotated successfully');
        }}
      />

      <AlertDialog open={!!deleteKey} onOpenChange={() => setDeleteKey(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle className="flex items-center gap-2">
              <AlertTriangle className="w-5 h-5" />
              Delete API Key
            </AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete the API key &quot;{deleteKey?.name}&quot;? This action
              cannot be undone and any applications using this key will stop working.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmDelete}
              className="sealed-button sealed-button--sm"
              disabled={deleteMutation.isPending}
            >
              {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <VaultSetupDialog
        open={showVaultSetup}
        onOpenChange={setShowVaultSetup}
        mode="setup"
        onSuccess={() => {
          toast.success('Vault set up successfully', {
            description: 'You can now encrypt and store API keys securely',
          });
          setShowCreateModal(true);
        }}
      />
    </div>
  );
}
