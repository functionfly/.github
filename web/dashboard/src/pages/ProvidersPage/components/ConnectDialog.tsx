import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { ProviderIcon } from '@/components/common/ProviderIcon';
import { AlertCircle, Check, Copy, Loader2, Plus, Shield, Zap } from 'lucide-react';
import { useState } from 'react';
import type { ProviderConfig } from '../constants/providerMeta';
import { storeEdgeApiKey, isVaultPassphraseSet, getVaultPassphrase, setupVaultPassphrase } from '@/services/vault-api-key-storage';

export interface ConnectResult {
  success: boolean;
  apiKey?: string;
  apiKeyId?: string;
}

interface ConnectDialogProps {
  provider: ProviderConfig;
  accent: { border: string; text: string };
  onConnect: (providerId: string, apiKey?: string) => Promise<ConnectResult>;
}

export function ConnectDialog({ provider, accent, onConnect }: ConnectDialogProps) {
  const [apiKey, setApiKey] = useState('');
  const [isConnecting, setIsConnecting] = useState(false);
  const [isOpen, setIsOpen] = useState(false);
  const [validationError, setValidationError] = useState<string | null>(null);
  const [generatedApiKey, setGeneratedApiKey] = useState<string | null>(null);
  const [generatedApiKeyId, setGeneratedApiKeyId] = useState<string | null>(null);
  const [isStoringInVault, setIsStoringInVault] = useState(false);
  const [vaultStoreSuccess, setVaultStoreSuccess] = useState(false);
  const isFunctionFly = provider.id === 'functionfly-edge';

  const handleConnect = async () => {
    setValidationError(null);
    setGeneratedApiKey(null);
    setGeneratedApiKeyId(null);
    setVaultStoreSuccess(false);

    if (!isFunctionFly) {
      if (!apiKey.trim()) {
        setValidationError('API key is required');
        return;
      }
      if (apiKey.length < 10) {
        setValidationError('API key must be at least 10 characters');
        return;
      }
    }

    setIsConnecting(true);
    try {
      const result = await onConnect(provider.id, isFunctionFly ? undefined : apiKey);
      if (result.success) {
        if (result.apiKey && result.apiKeyId) {
          setGeneratedApiKey(result.apiKey);
          setGeneratedApiKeyId(result.apiKeyId);
          await storeInVaultIfPossible(result.apiKey, result.apiKeyId);
        } else {
          setIsOpen(false);
          setApiKey('');
        }
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to connect provider';
      setValidationError(errorMessage);
    } finally {
      setIsConnecting(false);
    }
  };

  const storeInVaultIfPossible = async (apiKey: string, apiKeyId: string) => {
    setIsStoringInVault(true);
    try {
      if (!isVaultPassphraseSet()) {
        return;
      }
      const passphrase = await getVaultPassphrase();
      if (!passphrase) {
        return;
      }
      await storeEdgeApiKey({ apiKey, apiKeyId, passphrase });
      setVaultStoreSuccess(true);
    } catch (error) {
      console.error('Failed to store API key in vault:', error);
    } finally {
      setIsStoringInVault(false);
    }
  };

  const copyToClipboard = async () => {
    if (generatedApiKey) {
      await navigator.clipboard.writeText(generatedApiKey);
    }
  };

  const handleClose = () => {
    setIsOpen(false);
    setApiKey('');
    setValidationError(null);
    setGeneratedApiKey(null);
    setGeneratedApiKeyId(null);
    setVaultStoreSuccess(false);
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && handleClose()}>
      <DialogTrigger asChild>
        <Button
          variant="outline"
          className="w-full gap-2 border-border-default hover:border-border-subtle hover:bg-bg-secondary transition-all duration-200"
          onClick={() => setIsOpen(true)}
        >
          <Plus className="w-4 h-4" />
          Connect
        </Button>
      </DialogTrigger>
      <DialogContent className="bg-bg-tertiary border-border-subtle sm:max-w-md">
        <DialogHeader>
          <div className="flex items-center gap-3 mb-2">
            <div
              className="w-10 h-10 rounded-xl flex items-center justify-center"
              style={{ backgroundColor: `${accent.border}15` }}
            >
              <ProviderIcon provider={provider.id} size="md" />
            </div>
            <div>
              <DialogTitle className="text-text-primary text-lg">
                {isFunctionFly ? 'Enable' : 'Connect'} {provider.name}
              </DialogTitle>
            </div>
          </div>
          <DialogDescription className="text-text-secondary">
            {isFunctionFly
              ? 'Enable FunctionFly Edge to deploy to our managed infrastructure. No external account required.'
              : `Enter your API credentials to connect ${provider.name}. Credentials are encrypted with AES-256.`}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 pt-2">
          {validationError && (
            <div className="p-3 rounded-lg bg-error/10 border border-error/20 animate-in slide-in-from-top-1 duration-200">
              <div className="flex items-start gap-2">
                <AlertCircle className="w-4 h-4 text-error mt-0.5 shrink-0" />
                <p className="text-sm text-error">{validationError}</p>
              </div>
            </div>
          )}

          {generatedApiKey ? (
            <div className="space-y-4">
              <div className="p-4 rounded-lg bg-emerald-50 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-800/50">
                <div className="flex items-center gap-2 mb-2">
                  <Check className="w-4 h-4 text-emerald-600 dark:text-emerald-400" />
                  <span className="text-sm font-medium text-emerald-800 dark:text-emerald-300">
                    {isFunctionFly ? 'Provider Enabled' : 'Provider Connected'}
                  </span>
                </div>
                <p className="text-xs text-emerald-700 dark:text-emerald-400">
                  {vaultStoreSuccess
                    ? 'Your API key has been securely stored in your vault.'
                    : 'Your API key has been generated. Store it in your vault for safekeeping.'}
                </p>
              </div>

              <div className="space-y-2">
                <Label className="text-text-primary">Your API Key</Label>
                <div className="flex gap-2">
                  <Input
                    value={generatedApiKey}
                    readOnly
                    className="bg-bg-secondary border-border-subtle font-mono text-sm"
                  />
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={copyToClipboard}
                    title="Copy to clipboard"
                  >
                    <Copy className="w-4 h-4" />
                  </Button>
                </div>
              </div>

              <div className="flex items-start gap-2 p-3 rounded-lg bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-800/50">
                <Shield className="w-4 h-4 text-amber-600 dark:text-amber-400 mt-0.5 shrink-0" />
                <p className="text-xs text-amber-950 dark:text-amber-100">
                  This key will only be shown once. Store it in your vault or keep it safe. You can
                  always find it again in your vault at /api-keys.
                </p>
              </div>

              <Button
                onClick={() => {
                  setIsOpen(false);
                  setGeneratedApiKey(null);
                  setGeneratedApiKeyId(null);
                  setApiKey('');
                }}
                className="w-full"
                style={{ backgroundColor: accent.border, color: 'white' }}
              >
                Done
              </Button>
            </div>
          ) : isFunctionFly ? (
            <div className="space-y-4">
              <div className="p-4 rounded-lg bg-bg-secondary border border-border-subtle">
                <div className="flex items-start gap-3">
                  <div
                    className="w-8 h-8 rounded-lg flex items-center justify-center shrink-0"
                    style={{ backgroundColor: `${accent.border}15` }}
                  >
                    <Zap className="w-4 h-4" style={{ color: accent.text }} />
                  </div>
                  <div>
                    <h4 className="text-sm font-medium text-text-primary mb-1">Ready to Deploy</h4>
                    <p className="text-sm text-text-secondary">
                      FunctionFly Edge is our managed hosting. Select your preferred regions and
                      deploy immediately.
                    </p>
                  </div>
                </div>
              </div>

              <button
                type="button"
                onClick={handleConnect}
                disabled={isConnecting}
                className="w-full flex items-center justify-center gap-2 px-4 py-2.5 rounded-md font-medium transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed hover:opacity-90 active:scale-[0.98]"
                style={{
                  backgroundColor: accent.border,
                  color: 'white',
                }}
              >
                {isConnecting ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin" />
                    Enabling...
                  </>
                ) : (
                  <>
                    <Check className="w-4 h-4" />
                    Enable Provider
                  </>
                )}
              </button>
            </div>
          ) : (
            <>
              <div className="space-y-2">
                <Label htmlFor={`apiKey-${provider.id}`} className="text-text-primary">
                  API Key
                </Label>
                <Input
                  id={`apiKey-${provider.id}`}
                  type="password"
                  placeholder={`Enter your ${provider.name} API key`}
                  value={apiKey}
                  onChange={(e) => {
                    setApiKey(e.target.value);
                    if (validationError) setValidationError(null);
                  }}
                  className="bg-bg-secondary border-border-subtle focus:border-border-default"
                  disabled={isConnecting}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && !isConnecting) {
                      e.preventDefault();
                      handleConnect();
                    }
                  }}
                />
              </div>
              <div className="flex items-start gap-2 p-3 rounded-lg bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-800/50">
                <Shield className="w-4 h-4 text-amber-600 dark:text-amber-400 mt-0.5 shrink-0" />
                <p className="text-xs text-amber-950 dark:text-amber-100">
                  This allows FunctionFly to deploy to your {provider.name} account. You can revoke
                  access anytime. API keys are encrypted with AES-256-GCM.
                </p>
              </div>

              <button
                type="button"
                onClick={handleConnect}
                disabled={!apiKey.trim() || isConnecting}
                className="w-full flex items-center justify-center gap-2 px-4 py-2.5 rounded-md font-medium transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed hover:opacity-90 active:scale-[0.98]"
                style={{
                  backgroundColor: accent.border,
                  color: 'white',
                }}
              >
                {isConnecting ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin" />
                    Connecting...
                  </>
                ) : (
                  <>
                    <Check className="w-4 h-4" />
                    Connect Provider
                  </>
                )}
              </button>
            </>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
