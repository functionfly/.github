import { useState } from "react";
import { Link } from "react-router-dom";
import { Key, RefreshCw, Trash2, ExternalLink, AlertTriangle } from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Textarea } from "@/components/ui/textarea";
import { toast } from "sonner";
import { apiKeysApi } from "@/api/apikeys";
import type { APIKey, CreateAPIKeyRequest } from "@/types/api-key";
import { useQuery } from "@tanstack/react-query";
import { formatDate, getApiErrorMessage } from "../settings-utils";

export function ApiKeysSettingsTab() {
  const [generatingKey, setGeneratingKey] = useState(false);
  const [createKeyModalOpen, setCreateKeyModalOpen] = useState(false);
  const [createKeyName, setCreateKeyName] = useState("");
  const [createKeyDescription, setCreateKeyDescription] = useState("");
  const [createKeyExpiresAt, setCreateKeyExpiresAt] = useState("");
  const [createKeyStep, setCreateKeyStep] = useState<"form" | "key">("form");
  const [createKeyPlaintext, setCreateKeyPlaintext] = useState<string | null>(null);

  const [rotateKeyId, setRotateKeyId] = useState<string | null>(null);
  const [rotateKeyName, setRotateKeyName] = useState("");
  const [rotateNewPlaintext, setRotateNewPlaintext] = useState<string | null>(null);
  const [rotating, setRotating] = useState(false);

  const [revokeKeyId, setRevokeKeyId] = useState<string | null>(null);
  const [revokeKeyName, setRevokeKeyName] = useState("");
  const [revoking, setRevoking] = useState(false);

  const { data: apiKeysData, isLoading: apiKeysLoading, refetch: refetchApiKeys } = useQuery({
    queryKey: ["api-keys"],
    queryFn: async () => {
      try {
        const res = await apiKeysApi.list();
        return res.data as { data?: APIKey[]; meta?: { total?: number } };
      } catch (e: unknown) {
        const status = (e as { response?: { status?: number } })?.response?.status;
        if (status === 404) return { data: [] };
        throw e;
      }
    },
    enabled: true,
    retry: false,
  });

  const apiKeys: APIKey[] = Array.isArray((apiKeysData as { data?: APIKey[] })?.data)
    ? (apiKeysData as { data: APIKey[] }).data
    : [];

  const handleOpenCreateKeyModal = () => {
    setCreateKeyStep("form");
    setCreateKeyPlaintext(null);
    setCreateKeyName("");
    setCreateKeyDescription("");
    setCreateKeyExpiresAt("");
    setCreateKeyModalOpen(true);
  };

  const handleCreateKeySubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!createKeyName.trim()) {
      toast.error("Name is required");
      return;
    }
    setGeneratingKey(true);
    try {
      const body: CreateAPIKeyRequest = {
        name: createKeyName.trim(),
        key_type: "platform",
      };
      if (createKeyDescription.trim()) body.description = createKeyDescription.trim();
      if (createKeyExpiresAt) body.expires_at = new Date(createKeyExpiresAt).toISOString();
      const res = await apiKeysApi.create(body);
      const plaintext = (res.data as { data?: { plaintext?: string } })?.data?.plaintext ?? null;
      setCreateKeyPlaintext(plaintext);
      setCreateKeyStep("key");
      refetchApiKeys();
      toast.success("API key created");
    } catch (err: unknown) {
      const msg = getApiErrorMessage(err, {
        401: "Please sign in to generate API keys.",
        403: "You don't have permission to create API keys.",
        404: "API keys route not found. Run the orchestrator API on 8080.",
        default: "Failed to generate API key.",
      });
      toast.error(msg);
    } finally {
      setGeneratingKey(false);
    }
  };

  const handleCreateKeyDone = () => {
    setCreateKeyModalOpen(false);
    setCreateKeyStep("form");
    setCreateKeyPlaintext(null);
    setCreateKeyName("");
    setCreateKeyDescription("");
    setCreateKeyExpiresAt("");
  };

  const handleCopyKey = (value: string) => {
    navigator.clipboard.writeText(value);
    toast.success("Copied to clipboard");
  };

  const handleRotateClick = (key: APIKey) => {
    setRotateKeyId(key.id);
    setRotateKeyName(key.name);
    setRotateNewPlaintext(null);
  };

  const handleRotateConfirm = async () => {
    if (!rotateKeyId) return;
    setRotating(true);
    try {
      const res = await apiKeysApi.rotate(rotateKeyId, { reason: "manual" });
      const plaintext = (res.data as { data?: { plaintext?: string } })?.data?.plaintext ?? null;
      setRotateNewPlaintext(plaintext);
      refetchApiKeys();
      toast.success("API key rotated. Copy the new key below.");
    } catch (err: unknown) {
      const msg = getApiErrorMessage(err, { default: "Failed to rotate API key." });
      toast.error(msg);
    } finally {
      setRotating(false);
    }
  };

  const handleRotateClose = () => {
    setRotateKeyId(null);
    setRotateKeyName("");
    setRotateNewPlaintext(null);
  };

  const handleRevokeClick = (key: APIKey) => {
    setRevokeKeyId(key.id);
    setRevokeKeyName(key.name);
  };

  const handleRevokeConfirm = async () => {
    if (!revokeKeyId) return;
    setRevoking(true);
    try {
      await apiKeysApi.delete(revokeKeyId);
      refetchApiKeys();
      toast.success("API key revoked");
      setRevokeKeyId(null);
      setRevokeKeyName("");
    } catch (err: unknown) {
      const msg = getApiErrorMessage(err, { default: "Failed to revoke API key." });
      toast.error(msg);
    } finally {
      setRevoking(false);
    }
  };

  return (
    <div className="space-y-6">
      <Card className="ff-card-velocity">
        <CardHeader>
          <CardTitle className="font-display">API Keys</CardTitle>
          <CardDescription className="text-text-secondary">
            Manage your API keys for programmatic access. Treat keys like passwords—do not share them, and rotate if
            compromised.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {apiKeysLoading ? (
              <p className="text-text-muted text-sm">Loading API keys...</p>
            ) : apiKeys.length === 0 ? (
              <div className="rounded-lg border border-border-default bg-bg-secondary/50 p-6 text-center">
                <p className="text-text-muted text-sm">No API keys yet.</p>
                <p className="mt-1 text-sm text-text-secondary">
                  Generate your first API key to get started with programmatic access.
                </p>
                <Button className="ff-btn-velocity mt-4 gap-2" onClick={handleOpenCreateKeyModal} disabled={generatingKey}>
                  <Key className="h-4 w-4" />
                  {generatingKey ? "Generating…" : "Generate your first API key"}
                </Button>
              </div>
            ) : (
              <>
                {apiKeys.map((key: APIKey) => {
                  const prefix = key.prefix ?? key.key_prefix ?? "ffp_";
                  const lastUsed = key.last_used_at ? formatDate(key.last_used_at) : "Never";
                  const expiresAt = key.expires_at ? formatDate(key.expires_at) : null;
                  return (
                    <div
                      key={key.id}
                      className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-border-default bg-bg-secondary p-4"
                    >
                      <div className="min-w-0">
                        <h4 className="font-medium text-text-primary">{key.name || "API Key"}</h4>
                        <p className="text-sm text-text-muted">
                          Created {key.created_at ? formatDate(key.created_at) : "—"} · Last used {lastUsed}
                          {expiresAt && ` · Expires ${expiresAt}`}
                        </p>
                      </div>
                      <div className="flex flex-wrap items-center gap-2">
                        <Badge
                          variant={key.is_active ? "default" : "secondary"}
                          className={
                            key.is_active ? "ff-badge-success" : ""
                          }
                        >
                          {key.is_active ? "Active" : "Inactive"}
                        </Badge>
                        <code className="rounded bg-bg-tertiary px-3 py-1 text-sm text-text-secondary">
                          {prefix}••••••••••••
                        </code>
                        <Button variant="ghost" size="sm" onClick={() => handleCopyKey(prefix)} title="Copy prefix">
                          Copy prefix
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleRotateClick(key)}
                          title="Rotate key"
                          disabled={rotating}
                        >
                          <RefreshCw className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleRevokeClick(key)}
                          title="Revoke key"
                          className="text-destructive hover:text-destructive"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                  );
                })}
                <Button className="ff-btn-velocity gap-2" onClick={handleOpenCreateKeyModal} disabled={generatingKey}>
                  <Key className="h-4 w-4" />
                  {generatingKey ? "Generating…" : "Generate New Key"}
                </Button>
              </>
            )}
            <p className="text-muted-foreground text-xs">
              For permissions, environments, and full management, go to{" "}
              <Link to="/dashboard/api-keys" className="text-primary underline hover:no-underline">
                API Keys in Dashboard
              </Link>
              <ExternalLink className="ml-0.5 inline h-3 w-3" />
            </p>
          </div>
        </CardContent>
      </Card>

      <Dialog open={createKeyModalOpen} onOpenChange={(open) => !open && handleCreateKeyDone()}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{createKeyStep === "form" ? "Create API key" : "Copy your API key"}</DialogTitle>
            <DialogDescription>
              {createKeyStep === "form"
                ? "Give the key a name and optional description. You can set an expiration date if needed."
                : "This key is shown only once. Copy it now and store it securely."}
            </DialogDescription>
          </DialogHeader>
          {createKeyStep === "form" ? (
            <form onSubmit={handleCreateKeySubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="api-key-name">Name</Label>
                <Input
                  id="api-key-name"
                  value={createKeyName}
                  onChange={(e) => setCreateKeyName(e.target.value)}
                  placeholder="e.g. CI pipeline, Mobile app"
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="api-key-desc">Description (optional)</Label>
                <Textarea
                  id="api-key-desc"
                  value={createKeyDescription}
                  onChange={(e) => setCreateKeyDescription(e.target.value)}
                  placeholder="What this key is used for"
                  rows={2}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="api-key-expires">Expires at (optional)</Label>
                <Input
                  id="api-key-expires"
                  type="date"
                  value={createKeyExpiresAt}
                  onChange={(e) => setCreateKeyExpiresAt(e.target.value)}
                />
              </div>
              <DialogFooter>
                <Button type="button" variant="outline" onClick={handleCreateKeyDone}>
                  Cancel
                </Button>
                <Button type="submit" disabled={generatingKey}>
                  {generatingKey ? "Creating…" : "Create key"}
                </Button>
              </DialogFooter>
            </form>
          ) : (
            <div className="space-y-4">
              <div className="flex items-center gap-2 rounded-lg border border-amber-500/50 bg-amber-500/10 p-3 text-sm text-amber-600 dark:text-amber-400">
                <AlertTriangle className="h-5 w-5 shrink-0" />
                <span>Copy this key now. You won&apos;t be able to see it again.</span>
              </div>
              {createKeyPlaintext && (
                <div className="flex gap-2">
                  <Input readOnly value={createKeyPlaintext} className="font-mono text-sm" />
                  <Button type="button" onClick={() => handleCopyKey(createKeyPlaintext)}>
                    Copy
                  </Button>
                </div>
              )}
              <DialogFooter>
                <Button onClick={handleCreateKeyDone}>Done</Button>
              </DialogFooter>
            </div>
          )}
        </DialogContent>
      </Dialog>

      <Dialog open={!!rotateKeyId} onOpenChange={(open) => !open && handleRotateClose()}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Rotate API key</DialogTitle>
            <DialogDescription>
              {rotateNewPlaintext
                ? "Your new key is below. Copy it now; the old key is already invalid."
                : `Rotate "${rotateKeyName}". The current key will stop working immediately.`}
            </DialogDescription>
          </DialogHeader>
          {rotateNewPlaintext ? (
            <div className="space-y-4">
              <div className="flex items-center gap-2 rounded-lg border border-amber-500/50 bg-amber-500/10 p-3 text-sm text-amber-600 dark:text-amber-400">
                <AlertTriangle className="h-5 w-5 shrink-0" />
                <span>Copy the new key now. It won&apos;t be shown again.</span>
              </div>
              <div className="flex gap-2">
                <Input readOnly value={rotateNewPlaintext} className="font-mono text-sm" />
                <Button type="button" onClick={() => handleCopyKey(rotateNewPlaintext)}>
                  Copy
                </Button>
              </div>
              <DialogFooter>
                <Button onClick={handleRotateClose}>Done</Button>
              </DialogFooter>
            </div>
          ) : (
            <DialogFooter>
              <Button variant="outline" onClick={handleRotateClose}>
                Cancel
              </Button>
              <Button onClick={handleRotateConfirm} disabled={rotating}>
                {rotating ? "Rotating…" : "Rotate key"}
              </Button>
            </DialogFooter>
          )}
        </DialogContent>
      </Dialog>

      <Dialog
        open={!!revokeKeyId}
        onOpenChange={(open) => !open && (setRevokeKeyId(null), setRevokeKeyName(""))}
      >
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Revoke API key</DialogTitle>
            <DialogDescription>
              Are you sure you want to revoke &quot;{revokeKeyName}&quot;? This key will stop working immediately and
              cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => { setRevokeKeyId(null); setRevokeKeyName(""); }}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleRevokeConfirm} disabled={revoking}>
              {revoking ? "Revoking…" : "Revoke key"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
