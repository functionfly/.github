import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { ProviderIcon } from '@/components/common/ProviderIcon';
import { AlertCircle, Shield, ArrowRight, Loader2, Check } from 'lucide-react';
import { useState } from 'react';
import type { ProviderConfig } from '../constants/providerMeta';

export interface FailoverConfig {
  enabled: boolean;
  primaryProviderId: string | null;
  fallbackProviderId: string | null;
  autoSwitchThreshold: number; // Error rate % to trigger failover
  switchbackDelay: number; // Minutes before switching back
}

interface AutoFailoverDialogProps {
  providers: ProviderConfig[];
  connectedProviderIds: string[];
  currentConfig: FailoverConfig;
  isOpen: boolean;
  onClose: () => void;
  onSave: (config: FailoverConfig) => Promise<void>;
  isSaving?: boolean;
}

export function AutoFailoverDialog({
  providers,
  connectedProviderIds,
  currentConfig,
  isOpen,
  onClose,
  onSave,
  isSaving = false,
}: AutoFailoverDialogProps) {
  const [config, setConfig] = useState<FailoverConfig>(currentConfig);
  const [showSuccess, setShowSuccess] = useState(false);

  // Filter to only connected providers
  const connectedProviders = providers.filter((p) =>
    connectedProviderIds.includes(p.id)
  );

  const handleClose = () => {
    setConfig(currentConfig);
    setShowSuccess(false);
    onClose();
  };

  const handleSave = async () => {
    try {
      await onSave(config);
      setShowSuccess(true);
      setTimeout(() => {
        handleClose();
      }, 1500);
    } catch (error) {
      console.error('Failed to save failover config:', error);
    }
  };

  // Get available fallback options (exclude primary)
  const fallbackOptions = connectedProviders.filter(
    (p) => p.id !== config.primaryProviderId
  );

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && handleClose()}>
      <DialogContent className="bg-bg-tertiary border-border-subtle sm:max-w-lg">
        <DialogHeader>
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded-xl flex items-center justify-center bg-blue-100 dark:bg-blue-900/30">
              <Shield className="w-5 h-5 text-blue-600 dark:text-blue-400" />
            </div>
            <div>
              <DialogTitle className="text-text-primary text-lg">Auto-Failover Configuration</DialogTitle>
            </div>
          </div>
          <DialogDescription className="text-text-secondary">
            Configure automatic failover between providers when error rates exceed thresholds.
          </DialogDescription>
        </DialogHeader>

        {showSuccess ? (
          <div className="p-4 rounded-lg bg-emerald-50 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-800/50 animate-in slide-in-from-top-2">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-emerald-100 dark:bg-emerald-900/30 flex items-center justify-center">
                <Check className="w-5 h-5 text-emerald-600 dark:text-emerald-400" />
              </div>
              <div>
                <h4 className="font-medium text-emerald-800 dark:text-emerald-400">Configuration Saved</h4>
                <p className="text-sm text-emerald-700 dark:text-emerald-400">
                  Auto-failover settings have been updated.
                </p>
              </div>
            </div>
          </div>
        ) : (
          <div className="space-y-6 pt-2">
            {/* Enable Switch */}
            <div className="flex items-center justify-between p-3 rounded-lg bg-bg-secondary/50 border border-border-subtle">
              <div>
                <Label className="text-text-primary font-medium">Enable Auto-Failover</Label>
                <p className="text-sm text-text-secondary">
                  Automatically switch providers when error rates exceed threshold
                </p>
              </div>
              <Switch
                checked={config.enabled}
                onCheckedChange={(checked) =>
                  setConfig((prev) => ({ ...prev, enabled: checked }))
                }
              />
            </div>

            {config.enabled && (
              <>
                {/* Primary Provider */}
                <div className="space-y-3">
                  <Label className="text-text-primary">Primary Provider</Label>
                  <Select
                    value={config.primaryProviderId || ''}
                    onValueChange={(value) =>
                      setConfig((prev) => ({
                        ...prev,
                        primaryProviderId: value,
                        fallbackProviderId: null, // Reset fallback when primary changes
                      }))
                    }
                  >
                    <SelectTrigger className="bg-bg-secondary border-border-subtle">
                      <SelectValue placeholder="Select primary provider" />
                    </SelectTrigger>
                    <SelectContent>
                      {connectedProviders.map((provider) => (
                        <SelectItem key={provider.id} value={provider.id}>
                          <div className="flex items-center gap-2">
                            <ProviderIcon provider={provider.id} size="sm" />
                            <span>{provider.name}</span>
                          </div>
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                {/* Fallback Provider */}
                <div className="space-y-3">
                  <div className="flex items-center gap-2">
                    <Label className="text-text-primary">Fallback Provider</Label>
                    <ArrowRight className="w-4 h-4 text-text-muted" />
                  </div>
                  <Select
                    value={config.fallbackProviderId || ''}
                    onValueChange={(value) =>
                      setConfig((prev) => ({ ...prev, fallbackProviderId: value }))
                    }
                    disabled={!config.primaryProviderId || fallbackOptions.length === 0}
                  >
                    <SelectTrigger className="bg-bg-secondary border-border-subtle">
                      <SelectValue
                        placeholder={
                          fallbackOptions.length === 0
                            ? 'Connect more providers to enable failover'
                            : 'Select fallback provider'
                        }
                      />
                    </SelectTrigger>
                    <SelectContent>
                      {fallbackOptions.map((provider) => (
                        <SelectItem key={provider.id} value={provider.id}>
                          <div className="flex items-center gap-2">
                            <ProviderIcon provider={provider.id} size="sm" />
                            <span>{provider.name}</span>
                          </div>
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                {/* Threshold Settings */}
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label className="text-text-primary text-sm">Error Rate Threshold</Label>
                    <Select
                      value={String(config.autoSwitchThreshold)}
                      onValueChange={(value) =>
                        setConfig((prev) => ({
                          ...prev,
                          autoSwitchThreshold: Number(value),
                        }))
                      }
                    >
                      <SelectTrigger className="bg-bg-secondary border-border-subtle">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="1">1% error rate</SelectItem>
                        <SelectItem value="5">5% error rate</SelectItem>
                        <SelectItem value="10">10% error rate</SelectItem>
                        <SelectItem value="25">25% error rate</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-2">
                    <Label className="text-text-primary text-sm">Switchback Delay</Label>
                    <Select
                      value={String(config.switchbackDelay)}
                      onValueChange={(value) =>
                        setConfig((prev) => ({
                          ...prev,
                          switchbackDelay: Number(value),
                        }))
                      }
                    >
                      <SelectTrigger className="bg-bg-secondary border-border-subtle">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="5">5 minutes</SelectItem>
                        <SelectItem value="15">15 minutes</SelectItem>
                        <SelectItem value="30">30 minutes</SelectItem>
                        <SelectItem value="60">1 hour</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>

                {connectedProviders.length < 2 && (
                  <Alert className="bg-amber-50 dark:bg-amber-950/30 border-amber-200 dark:border-amber-800/50">
                    <AlertCircle className="h-4 w-4 text-amber-600 dark:text-amber-400" />
                    <AlertDescription className="text-amber-700 dark:text-amber-400 text-sm">
                      You need at least 2 connected providers to configure failover. Connect another
                      provider first.
                    </AlertDescription>
                  </Alert>
                )}
              </>
            )}

            <DialogFooter className="gap-2 sm:gap-0 pt-2">
              <Button
                variant="outline"
                onClick={handleClose}
                disabled={isSaving}
                className="border-border-default"
              >
                Cancel
              </Button>
              <Button
                onClick={handleSave}
                disabled={
                  isSaving ||
                  (config.enabled && (!config.primaryProviderId || !config.fallbackProviderId))
                }
                className="gap-2 bg-blue-600 hover:bg-blue-700"
              >
                {isSaving ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin" />
                    Saving...
                  </>
                ) : (
                  <>
                    <Shield className="w-4 h-4" />
                    Save Configuration
                  </>
                )}
              </Button>
            </DialogFooter>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
