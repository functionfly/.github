import { useState, useEffect } from "react";
import { Plus, Check, X, ExternalLink, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { StatusBadge } from "@/components/common/StatusBadge";
import { ProviderIcon } from "@/components/common/ProviderIcon";
import { PROVIDERS } from "@/lib/constants";
import { useProvidersStore } from "@/stores/providersStore";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";

export function ProvidersPage() {
  const [apiKey, setApiKey] = useState("");
  const [connecting, setConnecting] = useState(false);
  const [disconnecting, setDisconnecting] = useState<string | null>(null);
  const {
    providers,
    error,
    fetchProviders,
    connectProvider,
    disconnectProvider,
    testConnection,
    clearError,
  } = useProvidersStore();

  useEffect(() => {
    fetchProviders();
  }, [fetchProviders]);

  const handleConnect = async (providerId: string) => {
    if (!apiKey.trim()) return;

    setConnecting(true);
    clearError();

    try {
      await connectProvider({ providerId, apiKey });

      // Test the connection
      const connectionTest = await testConnection(providerId);
      if (!connectionTest) {
        // If connection test fails, we might want to show a warning but keep the provider connected
        console.warn(`Connection test failed for ${providerId}, but provider was connected`);
      }

      setApiKey("");
    } catch (error) {
      // Error is handled by the store
      console.error("Failed to connect provider:", error);
    } finally {
      setConnecting(false);
    }
  };

  const handleDisconnect = async (providerId: string) => {
    setDisconnecting(providerId);
    clearError();

    try {
      await disconnectProvider(providerId);
    } catch (error) {
      // Error is handled by the store
      console.error("Failed to disconnect provider:", error);
    } finally {
      setDisconnecting(null);
    }
  };

  const isConnected = (providerId: string) =>
    providers.some((p) => p.id === providerId);

  const getProviderStatus = (providerId: string) => {
    const connected = providers.find((p) => p.id === providerId);
    return connected?.status || "pending";
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-white">Providers</h1>
          <p className="text-text-secondary">Connect and manage your edge providers</p>
        </div>
      </div>

      {/* Error Message */}
      {error && (
        <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-lg">
          <p className="text-red-400">{error}</p>
          <Button
            variant="ghost"
            size="sm"
            className="mt-2 text-red-400 hover:text-red-400 hover:bg-red-500/10"
            onClick={clearError}
          >
            Dismiss
          </Button>
        </div>
      )}

      {/* Providers Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {Object.values(PROVIDERS).map((provider) => {
          const connected = isConnected(provider.id);
          const status = getProviderStatus(provider.id);

          return (
            <Card key={provider.id} className={connected ? "border-emerald-500/30" : ""}>
              <CardContent className="p-6">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-4">
                    <div
                      className="w-12 h-12 rounded-xl flex items-center justify-center"
                      style={{ backgroundColor: `${provider.color}20` }}
                    >
                      <ProviderIcon provider={provider.id} size="lg" />
                    </div>
                    <div>
                      <h3 className="font-semibold text-white">{provider.name}</h3>
                      <p className="text-sm text-text-muted">
                        {connected ? "Connected" : "Not connected"}
                      </p>
                    </div>
                  </div>
                  {connected && <StatusBadge status={status} />}
                </div>

                <div className="mt-4 pt-4 border-t border-white/8">
                  <p className="text-sm text-text-secondary mb-2">Available regions:</p>
                  <div className="flex flex-wrap gap-2">
                    {provider.regions.slice(0, 4).map((region) => (
                      <span
                        key={region}
                        className="px-2 py-1 rounded-md bg-bg-secondary text-xs text-text-secondary"
                      >
                        {region}
                      </span>
                    ))}
                    {provider.regions.length > 4 && (
                      <span className="px-2 py-1 rounded-md bg-bg-secondary text-xs text-text-secondary">
                        +{provider.regions.length - 4} more
                      </span>
                    )}
                  </div>
                </div>

                <div className="mt-4 flex gap-2">
                  {connected ? (
                    <>
                      <Button variant="outline" className="flex-1 gap-2">
                        <ExternalLink className="w-4 h-4" />
                        Configure
                      </Button>
                      <Button
                        variant="ghost"
                        className="text-red-400 hover:text-red-400 hover:bg-red-500/10"
                        onClick={() => handleDisconnect(provider.id)}
                        disabled={disconnecting === provider.id}
                      >
                        {disconnecting === provider.id ? (
                          <Loader2 className="w-4 h-4 animate-spin" />
                        ) : (
                          <X className="w-4 h-4" />
                        )}
                      </Button>
                    </>
                  ) : (
                    <Dialog>
                      <DialogTrigger asChild>
                        <Button className="w-full gap-2">
                          <Plus className="w-4 h-4" />
                          Connect
                        </Button>
                      </DialogTrigger>
                        <DialogContent className="bg-bg-tertiary border-white/8">
                          <DialogHeader>
                            <DialogTitle className="text-white">Connect {provider.name}</DialogTitle>
                            <DialogDescription className="text-text-secondary">
                              Enter your API key to connect {provider.name} to FunctionFly.
                            </DialogDescription>
                          </DialogHeader>
                        <div className="space-y-4 pt-4">
                          <div className="space-y-2">
                            <Label htmlFor="apiKey">API Key</Label>
                            <Input
                              id="apiKey"
                              type="password"
                              placeholder="Enter your API key"
                              value={apiKey}
                              onChange={(e) => setApiKey(e.target.value)}
                            />
                          </div>
                          <Button
                            className="w-full gap-2"
                            onClick={() => handleConnect(provider.id)}
                            disabled={!apiKey || connecting}
                          >
                            {connecting ? (
                              <Loader2 className="w-4 h-4 animate-spin" />
                            ) : (
                              <Check className="w-4 h-4" />
                            )}
                            {connecting ? "Connecting..." : "Connect Provider"}
                          </Button>
                        </div>
                      </DialogContent>
                    </Dialog>
                  )}
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>
    </div>
  );
}
