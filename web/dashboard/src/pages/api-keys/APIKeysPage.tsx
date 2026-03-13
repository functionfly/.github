import { useState, useEffect, useMemo } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import {
  Key,
  Plus,
  AlertTriangle,
  Shield,
  BookOpen,
  ChevronDown,
  ChevronRight,
  Copy,
  Check,
  KeyRound,
  Clock,
  CheckCircle2,
  XCircle,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { toast } from "sonner";
import {
  APIKeyList,
  CreateAPIKeyModal,
  APIKeyRotationModal,
} from "@/components/api-keys";
import { APIKey, APIKeyFilters, DEFAULT_RATE_LIMIT } from "@/types/api-key";
import { apiKeysService, getStoredApiKey } from "@/services/api-keys";
import { getApiBaseUrl } from "@/lib/constants";

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

  const toggleUsage = () => setUsageOpen((v) => !v);
  const toggleSecurity = () => setSecurityOpen((v) => !v);

  // Check for stored newly created API key
  useEffect(() => {
    const stored = getStoredApiKey();
    if (stored) {
      toast.success("API key created successfully", {
        description: "Copy your new API key now. It won't be shown again.",
        duration: 10000,
      });
    }
  }, []);

  const { data, isLoading, refetch } = useQuery({
    queryKey: ["api-keys", filters, page],
    queryFn: () => apiKeysService.listKeys(filters, page, PAGE_SIZE),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => apiKeysService.deleteKey(id),
    onSuccess: (_data, deletedId) => {
      const idStr = String(deletedId);
      setDeletedIds((prev) => new Set(prev).add(idStr));
      queryClient.invalidateQueries({ queryKey: ["api-keys"] });
      toast.success("API key deleted");
      setDeleteKey(null);
    },
    onError: (error) => {
      toast.error("Failed to delete API key", {
        description: error instanceof Error ? error.message : "Unknown error",
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

  const handleDelete = (key: APIKey) => {
    setDeleteKey(key);
  };

  const confirmDelete = () => {
    if (deleteKey) deleteMutation.mutate(deleteKey.id);
  };

  const apiBase = getApiBaseUrl() || (typeof window !== "undefined" ? window.location.origin : "");
  const curlExample = `curl -X GET "${apiBase}/v1/functions" \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -H "Content-Type: application/json"`;

  const copyCurl = async () => {
    await navigator.clipboard.writeText(curlExample);
    setCopiedSnippet(true);
    toast.success("Copied to clipboard");
    setTimeout(() => setCopiedSnippet(false), 2000);
  };

  return (
    <div className="space-y-8">
      {/* Page header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
            <Link to="/dashboard" className="hover:text-foreground">
              Dashboard
            </Link>
            <span>/</span>
            <span>API Keys</span>
          </div>
          <h1 className="text-3xl font-bold tracking-tight">API Keys</h1>
          <p className="text-muted-foreground mt-1 max-w-2xl">
            Create and manage API keys for programmatic access. Use keys in headers or environment
            variables; rotate them regularly and revoke if compromised.
          </p>
        </div>
        <Button onClick={() => setShowCreateModal(true)} size="lg" className="shrink-0">
          <Plus className="w-4 h-4 mr-2" />
          Create API Key
        </Button>
      </div>

      {/* Stats row */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total keys
            </CardTitle>
            <KeyRound className="w-4 h-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{displayTotal}</div>
            <p className="text-xs text-muted-foreground">Across all pages</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Active (this page)
            </CardTitle>
            <CheckCircle2 className="w-4 h-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.active}</div>
            <p className="text-xs text-muted-foreground">Can be used for requests</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Inactive (this page)
            </CardTitle>
            <XCircle className="w-4 h-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.inactive}</div>
            <p className="text-xs text-muted-foreground">Revoked or disabled</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Expiring soon
            </CardTitle>
            <Clock className="w-4 h-4 text-amber-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.expiringSoon}</div>
            <p className="text-xs text-muted-foreground">Within {EXPIRING_DAYS} days</p>
          </CardContent>
        </Card>
      </div>

      {/* How to use & Security */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <button
            type="button"
            onClick={toggleUsage}
            className="w-full text-left"
            aria-expanded={usageOpen}
          >
            <CardHeader className="flex flex-row items-center justify-between hover:bg-muted/50 rounded-t-lg transition-colors">
              <div className="flex items-center gap-2">
                <BookOpen className="w-5 h-5 text-primary" />
                <CardTitle className="text-base">How to use your API key</CardTitle>
              </div>
              {usageOpen ? (
                <ChevronDown className="w-4 h-4 text-muted-foreground" />
              ) : (
                <ChevronRight className="w-4 h-4 text-muted-foreground" />
              )}
            </CardHeader>
          </button>
          {usageOpen && (
            <CardContent className="pt-0">
              <p className="text-sm text-muted-foreground mb-3">
                Send your API key in the{" "}
                <code className="text-xs bg-muted px-1 rounded">
                  Authorization: Bearer &lt;key&gt;
                </code>{" "}
                header.
              </p>
              <pre className="relative text-xs bg-muted/80 rounded-lg p-4 overflow-x-auto">
                <Button
                  variant="ghost"
                  size="icon"
                  className="absolute top-2 right-2 h-7 w-7"
                  onClick={(e) => {
                    e.preventDefault();
                    copyCurl();
                  }}
                >
                  {copiedSnippet ? (
                    <Check className="w-3.5 h-3.5 text-green-600" />
                  ) : (
                    <Copy className="w-3.5 h-3.5" />
                  )}
                </Button>
                <code>{curlExample}</code>
              </pre>
              <p className="text-xs text-muted-foreground mt-2">
                Default rate limits: {DEFAULT_RATE_LIMIT.rpm.toLocaleString()} RPM,{" "}
                {DEFAULT_RATE_LIMIT.rph.toLocaleString()} RPH. Adjust when creating a key.
              </p>
            </CardContent>
          )}
        </Card>

        <Card>
          <button
            type="button"
            onClick={toggleSecurity}
            className="w-full text-left"
            aria-expanded={securityOpen}
          >
            <CardHeader className="flex flex-row items-center justify-between hover:bg-muted/50 rounded-t-lg transition-colors">
              <div className="flex items-center gap-2">
                <Shield className="w-5 h-5 text-primary" />
                <CardTitle className="text-base">Security best practices</CardTitle>
              </div>
              {securityOpen ? (
                <ChevronDown className="w-4 h-4 text-muted-foreground" />
              ) : (
                <ChevronRight className="w-4 h-4 text-muted-foreground" />
              )}
            </CardHeader>
          </button>
          {securityOpen && (
            <CardContent className="pt-0">
              <ul className="text-sm text-muted-foreground space-y-2 list-disc list-inside">
                <li>Never commit keys to git or share them in public.</li>
                <li>Use environment variables or a secrets manager in production.</li>
                <li>Rotate keys periodically and immediately if compromised.</li>
                <li>Create separate keys per environment (e.g. staging vs production).</li>
                <li>Prefer scoped keys (e.g. function type) over platform keys when possible.</li>
              </ul>
            </CardContent>
          )}
        </Card>
      </div>

      {/* Keys table */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Key className="w-5 h-5" />
            Your API Keys
          </CardTitle>
          <CardDescription>
            Search, filter, and manage keys. Click a key to edit permissions and environments.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <APIKeyList
            apiKeys={apiKeys}
            isLoading={isLoading}
            total={displayTotal}
            page={page}
            pageSize={PAGE_SIZE}
            filters={filters}
            onFiltersChange={handleFiltersChange}
            onPageChange={handlePageChange}
            onCreateNew={() => setShowCreateModal(true)}
            onRotate={handleRotate}
            onDelete={handleDelete}
          />
        </CardContent>
      </Card>

      <CreateAPIKeyModal
        open={showCreateModal}
        onOpenChange={setShowCreateModal}
        onSuccess={() => queryClient.invalidateQueries({ queryKey: ["api-keys"] })}
      />

      <APIKeyRotationModal
        open={showRotationModal}
        onOpenChange={setShowRotationModal}
        apiKey={selectedKey}
        onSuccess={() => {
          refetch();
          toast.success("API key rotated successfully");
        }}
      />

      <AlertDialog open={!!deleteKey} onOpenChange={() => setDeleteKey(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle className="flex items-center gap-2">
              <AlertTriangle className="w-5 h-5 text-red-500" />
              Delete API Key
            </AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete the API key &quot;{deleteKey?.name}&quot;? This
              action cannot be undone and any applications using this key will stop working.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmDelete}
              className="bg-red-600 hover:bg-red-700"
              disabled={deleteMutation.isPending}
            >
              {deleteMutation.isPending ? "Deleting..." : "Delete"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
