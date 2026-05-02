import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { VaultCrypto } from '@/utils/vault-crypto';
import { setupVaultPassphrase, isVaultPassphraseSet, setVaultPassphrase } from '@/services/vault-api-key-storage';
import { toast } from 'sonner';
import { Loader2, Lock, Shield, Eye, EyeOff } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { PasswordStrengthIndicator } from '@/components/common/PasswordStrengthIndicator';
import { cn } from '@/lib/utils';

interface VaultSetupDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
  mode?: 'setup' | 'unlock';
}

export function VaultSetupDialog({
  open,
  onOpenChange,
  onSuccess,
  mode: initialMode = 'setup',
}: VaultSetupDialogProps) {
  const { t } = useTranslation();
  const [mode, setMode] = useState<'setup' | 'unlock'>(initialMode);
  const [passphrase, setPassphrase] = useState('');
  const [confirmPassphrase, setConfirmPassphrase] = useState('');
  const [showPassphrase, setShowPassphrase] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isSetup = mode === 'setup';

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!passphrase) {
      setError('Passphrase is required');
      return;
    }

    if (isSetup) {
      if (passphrase.length < 12) {
        setError('Passphrase must be at least 12 characters');
        return;
      }
      if (!/[A-Z]/.test(passphrase) || !/[a-z]/.test(passphrase) || !/[0-9]/.test(passphrase)) {
        setError('Passphrase must contain uppercase, lowercase, and a digit');
        return;
      }
      if (passphrase !== confirmPassphrase) {
        setError('Passphrases do not match');
        return;
      }
    }

    setIsLoading(true);
    try {
      if (isSetup) {
        const result = await setupVaultPassphrase(passphrase);
        if (result.success) {
          toast.success('Vault passphrase set', {
            description: 'Your vault is now ready to securely store API keys',
          });
          onSuccess?.();
          onOpenChange(false);
          resetForm();
        } else {
          setError(result.error || 'Failed to setup vault');
        }
      } else {
        const testPayload = await VaultCrypto.encryptWithPassphrase('__vault_test__', passphrase);
        const decrypted = await VaultCrypto.decryptWithPassphrase(
          VaultCrypto.fromPayload(VaultCrypto.toPayload(testPayload)),
          passphrase
        );
        if (decrypted === '__vault_test__') {
          await setVaultPassphrase(passphrase);
          toast.success('Vault unlocked');
          onSuccess?.();
          onOpenChange(false);
          resetForm();
        } else {
          setError('Invalid passphrase');
        }
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setIsLoading(false);
    }
  };

  const resetForm = () => {
    setPassphrase('');
    setConfirmPassphrase('');
    setShowPassphrase(false);
    setError(null);
  };

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) resetForm();
    onOpenChange(newOpen);
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[480px]">
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-brand-500/20 flex items-center justify-center">
              <Shield className="w-5 h-5 text-brand-500" />
            </div>
            <div>
              <DialogTitle>
                {isSetup ? 'Set Up Vault Passphrase' : 'Unlock Vault'}
              </DialogTitle>
              <DialogDescription>
                {isSetup
                  ? 'Create a passphrase to encrypt your API keys. This passphrase will be required to view your keys.'
                  : 'Enter your vault passphrase to access your stored API keys.'}
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="bg-red-50 dark:bg-red-950 border border-red-200 dark:border-red-800 rounded-lg p-3 text-sm text-red-600 dark:text-red-400">
              {error}
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="vault-passphrase">
              {isSetup ? 'Create Passphrase' : 'Passphrase'} <span className="text-red-500">*</span>
            </Label>
            <div className="relative">
              <Input
                id="vault-passphrase"
                type={showPassphrase ? 'text' : 'password'}
                placeholder={isSetup ? 'Create a strong passphrase' : 'Enter your passphrase'}
                value={passphrase}
                onChange={(e) => setPassphrase(e.target.value)}
                className={cn(error && 'border-red-500', 'pr-10')}
                disabled={isLoading}
              />
              <button
                type="button"
                onClick={() => setShowPassphrase(!showPassphrase)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-text-muted hover:text-text-primary"
              >
                {showPassphrase ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              </button>
            </div>
            {isSetup && passphrase && <PasswordStrengthIndicator password={passphrase} />}
          </div>

          {isSetup && (
            <div className="space-y-2">
              <Label htmlFor="confirm-passphrase">
                Confirm Passphrase <span className="text-red-500">*</span>
              </Label>
              <Input
                id="confirm-passphrase"
                type={showPassphrase ? 'text' : 'password'}
                placeholder="Confirm your passphrase"
                value={confirmPassphrase}
                onChange={(e) => setConfirmPassphrase(e.target.value)}
                className={cn(error && 'border-red-500')}
                disabled={isLoading}
              />
            </div>
          )}

          {isSetup && (
            <div className="rounded-lg bg-amber-50 dark:bg-amber-950 border border-amber-200 dark:border-amber-800 p-4">
              <div className="flex items-start gap-3">
                <Lock className="w-5 h-5 text-amber-600 mt-0.5" />
                <div className="text-sm text-amber-800 dark:text-amber-200">
                  <p className="font-medium mb-1">Important: Save Your Passphrase</p>
                  <p>
                    Your passphrase is used to encrypt your API keys locally. We cannot recover
                    this passphrase if you lose it. Store it securely in a password manager.
                  </p>
                </div>
              </div>
            </div>
          )}

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => handleOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  {isSetup ? 'Setting up...' : 'Unlocking...'}
                </>
              ) : isSetup ? (
                'Set Up Vault'
              ) : (
                'Unlock'
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

export function useVaultStatus() {
  return {
    isVaultPassphraseSet: isVaultPassphraseSet(),
  };
}
