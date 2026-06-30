import { useState } from 'react';
import { useAIKeys, useSupportedProviders, useConnectAIKey, useDisconnectAIKey, useTestAIKey, useRotateAIKey } from '@/api/ai-keys';
import type { AIProviderKey, SupportedProvider } from '@/types/ai-keys';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger } from '@/components/ui/alert-dialog';
import { Eye, EyeOff, Key, Plus, RefreshCw, Shield, TestTube, Trash2 } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';

const PROVIDER_LOGOS: Record<string, string> = {
  openai: '🤖',
  anthropic: '🧠',
  openrouter: '🔀',
  fireworks: '🎆',
  groq: '⚡',
  deepinfra: '🌊',
  together: '🤝',
  mimo: '📱',
  stepfun: '🚶',
};

function StatusBadge({ status }: { status: string }) {
  const variants: Record<string, 'default' | 'secondary' | 'destructive' | 'outline'> = {
    active: 'default',
    degraded: 'secondary',
    expired: 'destructive',
    revoked: 'destructive',
  };
  return (
    <Badge variant={variants[status] ?? 'outline'} className="capitalize">
      {status}
    </Badge>
  );
}

function ConnectedKeyCard({
  keyEntry,
  onTest,
  onRotate,
  onDisconnect,
  isTesting,
}: {
  keyEntry: AIProviderKey;
  onTest: () => void;
  onRotate: (newKey: string) => void;
  onDisconnect: () => void;
  isTesting: boolean;
}) {
  const [showRotate, setShowRotate] = useState(false);
  const [newKey, setNewKey] = useState('');

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span className="text-2xl">{PROVIDER_LOGOS[keyEntry.provider] ?? '🔑'}</span>
            <div>
              <CardTitle className="text-base capitalize">{keyEntry.provider}</CardTitle>
              <p className="text-sm text-muted-foreground font-mono">
                sk-...{keyEntry.key_last4}
              </p>
            </div>
          </div>
          <StatusBadge status={keyEntry.status} />
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        {keyEntry.health_message && keyEntry.status !== 'active' && (
          <p className="text-sm text-yellow-600 dark:text-yellow-400">{keyEntry.health_message}</p>
        )}
        <div className="flex flex-wrap gap-2 text-xs text-muted-foreground">
          {keyEntry.last_health_check && (
            <span>Checked {formatDistanceToNow(new Date(keyEntry.last_health_check), { addSuffix: true })}</span>
          )}
          {keyEntry.last_used_at && (
            <span>Used {formatDistanceToNow(new Date(keyEntry.last_used_at), { addSuffix: true })}</span>
          )}
        </div>

        {showRotate && (
          <div className="flex gap-2">
            <Input
              type="password"
              placeholder="New API key"
              value={newKey}
              onChange={(e) => setNewKey(e.target.value)}
              className="font-mono text-sm"
            />
            <Button
              size="sm"
              onClick={() => {
                onRotate(newKey);
                setNewKey('');
                setShowRotate(false);
              }}
              disabled={!newKey}
            >
              Save
            </Button>
            <Button size="sm" variant="ghost" onClick={() => setShowRotate(false)}>
              Cancel
            </Button>
          </div>
        )}

        <div className="flex gap-2">
          <Button size="sm" variant="outline" onClick={onTest} disabled={isTesting}>
            <TestTube className="h-3.5 w-3.5 mr-1" />
            {isTesting ? 'Testing...' : 'Test'}
          </Button>
          <Button size="sm" variant="outline" onClick={() => setShowRotate(!showRotate)}>
            <RefreshCw className="h-3.5 w-3.5 mr-1" />
            Rotate
          </Button>
          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button size="sm" variant="destructive">
                <Trash2 className="h-3.5 w-3.5 mr-1" />
                Disconnect
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Disconnect {keyEntry.provider} key?</AlertDialogTitle>
                <AlertDialogDescription>
                  This will remove your BYOK key. Future AI calls to {keyEntry.provider} will use platform keys and incur platform pricing.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction onClick={onDisconnect}>Disconnect</AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </div>
      </CardContent>
    </Card>
  );
}

function ConnectKeyDialog({
  providers,
  connectedProviders,
  onConnect,
  isConnecting,
}: {
  providers: SupportedProvider[];
  connectedProviders: Set<string>;
  onConnect: (provider: string, apiKey: string) => void;
  isConnecting: boolean;
}) {
  const [open, setOpen] = useState(false);
  const [selected, setSelected] = useState('');
  const [apiKey, setApiKey] = useState('');
  const [showKey, setShowKey] = useState(false);

  const available = providers.filter((p) => !connectedProviders.has(p.id));
  const selectedProvider = providers.find((p) => p.id === selected);

  const handleSubmit = () => {
    if (selected && apiKey) {
      onConnect(selected, apiKey);
      setApiKey('');
      setSelected('');
      setOpen(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
          <Plus className="h-4 w-4 mr-2" />
          Connect AI Key
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Connect AI Provider Key</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <div>
            <Label>Provider</Label>
            <div className="grid grid-cols-3 gap-2 mt-2">
              {available.map((p) => (
                <button
                  key={p.id}
                  onClick={() => setSelected(p.id)}
                  className={`p-2 rounded-lg border text-center text-sm transition-colors ${
                    selected === p.id
                      ? 'border-primary bg-primary/10'
                      : 'border-border hover:border-primary/50'
                  }`}
                >
                  <span className="text-lg">{PROVIDER_LOGOS[p.id] ?? '🔑'}</span>
                  <p className="text-xs mt-1 capitalize">{p.name}</p>
                </button>
              ))}
            </div>
          </div>

          {selectedProvider && (
            <>
              {selectedProvider.is_meta_provider && (
                <div className="rounded-lg bg-blue-50 dark:bg-blue-950/30 p-3 text-sm text-blue-700 dark:text-blue-300">
                  <strong>{selectedProvider.name}</strong> gives you access to 100+ models through a single key.
                </div>
              )}
              <div>
                <Label htmlFor="api-key">API Key</Label>
                <div className="relative mt-1">
                  <Input
                    id="api-key"
                    type={showKey ? 'text' : 'password'}
                    placeholder={selectedProvider.key_format}
                    value={apiKey}
                    onChange={(e) => setApiKey(e.target.value)}
                    className="font-mono text-sm pr-10"
                  />
                  <button
                    type="button"
                    onClick={() => setShowKey(!showKey)}
                    className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                  >
                    {showKey ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </button>
                </div>
              </div>
              <div className="flex items-start gap-2 rounded-lg bg-muted p-3 text-sm">
                <Shield className="h-4 w-4 mt-0.5 text-muted-foreground shrink-0" />
                <p className="text-muted-foreground">
                  Your key is encrypted with AES-256-GCM. We never store plaintext.
                </p>
              </div>
              <Button onClick={handleSubmit} disabled={!apiKey || isConnecting} className="w-full">
                {isConnecting ? 'Validating...' : 'Test & Connect'}
              </Button>
            </>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}

export function AIKeysSettingsTab() {
  const { data: keys = [], isLoading: keysLoading } = useAIKeys();
  const { data: providers = [] } = useSupportedProviders();
  const connectMutation = useConnectAIKey();
  const disconnectMutation = useDisconnectAIKey();
  const testMutation = useTestAIKey();
  const rotateMutation = useRotateAIKey();

  const connectedProviders = new Set(keys.map((k) => k.provider));

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold">AI Provider Keys</h3>
          <p className="text-sm text-muted-foreground">
            Bring your own API keys to avoid platform AI costs. You pay the provider directly.
          </p>
        </div>
        <ConnectKeyDialog
          providers={providers}
          connectedProviders={connectedProviders}
          onConnect={(provider, apiKey) => connectMutation.mutate({ provider, apiKey })}
          isConnecting={connectMutation.isPending}
        />
      </div>

      <Separator />

      {keysLoading ? (
        <div className="text-center py-8 text-muted-foreground">Loading keys...</div>
      ) : keys.length === 0 ? (
        <div className="text-center py-12">
          <Key className="h-12 w-12 mx-auto text-muted-foreground/50 mb-4" />
          <p className="text-muted-foreground">No AI provider keys connected yet.</p>
          <p className="text-sm text-muted-foreground mt-1">
            Connect your own keys to save on AI costs.
          </p>
        </div>
      ) : (
        <div className="grid gap-4">
          {keys.map((key) => (
            <ConnectedKeyCard
              key={key.id}
              keyEntry={key}
              onTest={() => testMutation.mutate(key.provider)}
              onRotate={(newKey) => rotateMutation.mutate({ provider: key.provider, apiKey: newKey })}
              onDisconnect={() => disconnectMutation.mutate(key.provider)}
              isTesting={testMutation.isPending}
            />
          ))}
        </div>
      )}
    </div>
  );
}
