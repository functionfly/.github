import {
  connectorsApi,
  type Connector,
  type UserConnector,
} from "@/api/connectors";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { EmptyState } from "@/components/ui/empty-state";
import { Skeleton } from "@/components/ui/skeleton";
import { AnimatePresence, motion } from "framer-motion";
import {
  AlertCircle,
  CheckCircle2,
  FileText,
  Link2,
  Loader2,
  Mail,
  MessageSquare,
  Plus,
  RefreshCw,
  Shield,
  Unlink,
} from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { siGithub, siLinear } from "simple-icons";

const connectorIcons: Record<string, React.ReactNode> = {
  github: (
    <svg viewBox="0 0 24 24" className="w-5 h-5" fill="currentColor" aria-hidden="true">
      <path d={siGithub.path} />
    </svg>
  ),
  notion: <FileText className="w-5 h-5" />,
  slack: <MessageSquare className="w-5 h-5" />,
  gmail: <Mail className="w-5 h-5" />,
  linear: (
    <svg viewBox="0 0 24 24" className="w-5 h-5" fill="currentColor" aria-hidden="true">
      <path d={siLinear.path} />
    </svg>
  ),
};

const connectorColors: Record<string, string> = {
  github: "from-gray-700 to-gray-900",
  notion: "from-gray-800 to-black",
  slack: "from-purple-600 to-purple-800",
  gmail: "from-red-600 to-red-800",
  linear: "from-blue-600 to-blue-800",
};

const statusConfig: Record<string, { label: string; color: string; icon: React.ReactNode }> = {
  active: { label: "Active", color: "bg-green-500/10 text-green-400 border-green-500/20", icon: <CheckCircle2 className="w-3.5 h-3.5" /> },
  sync_error: { label: "Sync Error", color: "bg-red-500/10 text-red-400 border-red-500/20", icon: <AlertCircle className="w-3.5 h-3.5" /> },
  revoked: { label: "Revoked", color: "bg-yellow-500/10 text-yellow-400 border-yellow-500/20", icon: <AlertCircle className="w-3.5 h-3.5" /> },
};

function formatRelativeTime(dateStr?: string): string {
  if (!dateStr) return "Never";
  const date = new Date(dateStr);
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return "Just now";
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

function ConnectorCard({
  userConnector,
  onUnlink,
  onSync,
}: {
  userConnector: UserConnector;
  onUnlink: (id: string) => void;
  onSync: (id: string) => void;
}) {
  const [syncing, setSyncing] = useState(false);
  const status = statusConfig[userConnector.status] || statusConfig.active;
  const icon = connectorIcons[userConnector.connector_slug] || <Link2 className="w-5 h-5" />;
  const gradient = connectorColors[userConnector.connector_slug] || "from-gray-600 to-gray-800";

  const handleSync = async () => {
    setSyncing(true);
    await onSync(userConnector.id);
    setSyncing(false);
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -20 }}
      layout
    >
      <Card className="group relative overflow-hidden border-white/[0.06] bg-white/[0.02] hover:bg-white/[0.04] transition-all duration-300">
        <div className={`absolute top-0 left-0 right-0 h-1 bg-gradient-to-r ${gradient}`} />
        <CardHeader className="pb-3">
          <div className="flex items-start justify-between">
            <div className="flex items-center gap-3">
              <div className={`flex items-center justify-center w-10 h-10 rounded-lg bg-gradient-to-br ${gradient} text-white`}>
                {icon}
              </div>
              <div>
                <CardTitle className="text-base text-text-primary">{userConnector.display_name || userConnector.connector_name}</CardTitle>
                <CardDescription className="text-xs mt-0.5">
                  Last synced: {formatRelativeTime(userConnector.last_sync_at)}
                </CardDescription>
              </div>
            </div>
            <Badge variant="outline" className={status.color}>
              <span className="flex items-center gap-1.5">
                {status.icon}
                {status.label}
              </span>
            </Badge>
          </div>
        </CardHeader>
        <CardContent>
          {userConnector.sync_error && (
            <div className="mb-3 p-2 rounded-md bg-red-500/5 border border-red-500/10 text-xs text-red-400">
              {userConnector.sync_error}
            </div>
          )}
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={handleSync}
              disabled={syncing}
              className="flex-1 text-xs border-white/10 hover:bg-white/5"
            >
              {syncing ? (
                <Loader2 className="w-3.5 h-3.5 mr-1.5 animate-spin" />
              ) : (
                <RefreshCw className="w-3.5 h-3.5 mr-1.5" />
              )}
              Sync Now
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onUnlink(userConnector.id)}
              className="text-xs text-red-400 hover:text-red-300 hover:bg-red-500/10"
            >
              <Unlink className="w-3.5 h-3.5 mr-1.5" />
              Unlink
            </Button>
          </div>
        </CardContent>
      </Card>
    </motion.div>
  );
}

function CatalogCard({
  connector,
  isLinked,
  onLink,
}: {
  connector: Connector;
  isLinked: boolean;
  onLink: (slug: string) => void;
}) {
  const icon = connectorIcons[connector.slug] || <Link2 className="w-5 h-5" />;
  const gradient = connectorColors[connector.slug] || "from-gray-600 to-gray-800";

  return (
    <motion.div initial={{ opacity: 0, scale: 0.95 }} animate={{ opacity: 1, scale: 1 }} whileHover={{ scale: 1.02 }}>
      <Card className={`relative overflow-hidden border-white/[0.06] bg-white/[0.02] transition-all duration-300 ${isLinked ? "opacity-60" : "hover:bg-white/[0.04] hover:border-white/[0.1]"}`}>
        <CardContent className="p-4">
          <div className="flex items-center gap-3">
            <div className={`flex items-center justify-center w-10 h-10 rounded-lg bg-gradient-to-br ${gradient} text-white`}>
              {icon}
            </div>
            <div className="flex-1 min-w-0">
              <h3 className="text-sm font-medium text-text-primary truncate">{connector.name}</h3>
              <p className="text-xs text-text-secondary mt-0.5 truncate">
                {connector.scopes.split(",").length} permissions
              </p>
            </div>
            {isLinked ? (
              <Badge variant="outline" className="bg-green-500/10 text-green-400 border-green-500/20 text-xs">
                Linked
              </Badge>
            ) : (
              <Button
                size="sm"
                onClick={() => onLink(connector.slug)}
                className="text-xs"
              >
                <Plus className="w-3.5 h-3.5 mr-1" />
                Link
              </Button>
            )}
          </div>
        </CardContent>
      </Card>
    </motion.div>
  );
}

export function IntegrationsSettingsTab() {
  const [catalog, setCatalog] = useState<Connector[]>([]);
  const [userConnectors, setUserConnectors] = useState<UserConnector[]>([]);
  const [loading, setLoading] = useState(true);
  const [linkDialogOpen, setLinkDialogOpen] = useState(false);
  const [selectedSlug, setSelectedSlug] = useState<string | null>(null);
  const [linking, setLinking] = useState(false);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const [cat, uc] = await Promise.all([
        connectorsApi.listCatalog(),
        connectorsApi.listUserConnectors(),
      ]);
      setCatalog(cat);
      setUserConnectors(uc);
    } catch (err) {
      console.error("Failed to load connectors:", err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleLink = async (slug: string) => {
    setSelectedSlug(slug);
    setLinkDialogOpen(true);
  };

  const confirmLink = async () => {
    if (!selectedSlug) return;
    setLinking(true);
    try {
      await connectorsApi.linkConnector({ connector_slug: selectedSlug });
      await loadData();
      setLinkDialogOpen(false);
    } catch (err) {
      console.error("Failed to link connector:", err);
    } finally {
      setLinking(false);
    }
  };

  const handleUnlink = async (id: string) => {
    try {
      await connectorsApi.unlinkConnector(id);
      setUserConnectors((prev) => prev.filter((c) => c.id !== id));
    } catch (err) {
      console.error("Failed to unlink connector:", err);
    }
  };

  const handleSync = async (id: string) => {
    try {
      await connectorsApi.triggerSync(id);
      await loadData();
    } catch (err) {
      console.error("Failed to trigger sync:", err);
    }
  };

  const linkedSlugs = new Set(userConnectors.map((c) => c.connector_slug));

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-32 rounded-lg" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Security notice */}
      <Card className="border-indigo-500/20 bg-indigo-500/5">
        <CardContent className="p-4 flex items-center gap-3">
          <Shield className="w-5 h-5 text-indigo-400 flex-shrink-0" />
          <div className="text-sm">
            <span className="text-indigo-300 font-medium">Zero-knowledge encryption.</span>{" "}
            <span className="text-text-secondary">
              OAuth tokens are encrypted client-side before storage. The server never sees your plaintext credentials.
            </span>
          </div>
        </CardContent>
      </Card>

      {/* Linked Connectors */}
      {userConnectors.length > 0 && (
        <div>
          <h2 className="text-lg font-semibold text-text-primary mb-4 flex items-center gap-2">
            <CheckCircle2 className="w-5 h-5 text-green-400" />
            Linked Accounts
            <Badge variant="outline" className="ml-auto bg-indigo-500/10 text-indigo-400 border-indigo-500/20">
              {userConnectors.length} linked
            </Badge>
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <AnimatePresence mode="popLayout">
              {userConnectors.map((uc) => (
                <ConnectorCard
                  key={uc.id}
                  userConnector={uc}
                  onUnlink={handleUnlink}
                  onSync={handleSync}
                />
              ))}
            </AnimatePresence>
          </div>
        </div>
      )}

      {/* Empty state */}
      {userConnectors.length === 0 && (
        <EmptyState
          variant="card"
          icon={<Link2 className="w-8 h-8" />}
          title="No connectors linked"
          description="Link your first connector to start feeding your Brain with signals from your tools."
        />
      )}

      {/* Available Connectors */}
      <div>
        <h2 className="text-lg font-semibold text-text-primary mb-4 flex items-center gap-2">
          <Link2 className="w-5 h-5 text-indigo-400" />
          Available Connectors
        </h2>
        <p className="text-sm text-text-secondary mb-4">
          Link external accounts to feed your Brain with real-time signals from your tools.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {catalog.map((connector) => (
            <CatalogCard
              key={connector.id}
              connector={connector}
              isLinked={linkedSlugs.has(connector.slug)}
              onLink={handleLink}
            />
          ))}
        </div>
      </div>

      {/* Link confirmation dialog */}
      <Dialog open={linkDialogOpen} onOpenChange={setLinkDialogOpen}>
        <DialogContent className="bg-slate-900 border-white/10">
          <DialogHeader>
            <DialogTitle className="text-text-primary">
              Link {selectedSlug && catalog.find((c) => c.slug === selectedSlug)?.name}
            </DialogTitle>
          </DialogHeader>
          <div className="py-4">
            <p className="text-sm text-text-secondary">
              This will initiate an OAuth flow to connect your account. Your credentials will be
              encrypted client-side using AES-256-GCM before being stored.
            </p>
            <div className="mt-4 p-3 rounded-lg bg-white/[0.03] border border-white/[0.06]">
              <h4 className="text-xs font-medium text-text-primary mb-2">Requested permissions:</h4>
              <div className="flex flex-wrap gap-1.5">
                {selectedSlug &&
                  catalog
                    .find((c) => c.slug === selectedSlug)
                    ?.scopes.split(",")
                    .map((scope) => (
                      <Badge key={scope} variant="outline" className="text-xs bg-white/5 border-white/10">
                        {scope.trim()}
                      </Badge>
                    ))}
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setLinkDialogOpen(false)} className="border-white/10">
              Cancel
            </Button>
            <Button onClick={confirmLink} disabled={linking}>
              {linking ? (
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
              ) : (
                <Link2 className="w-4 h-4 mr-2" />
              )}
              Link Account
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
