import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { VaultCrypto } from '@/utils/vault-crypto';
import { setupVaultPassphrase, setVaultPassphrase } from '@/services/vault-api-key-storage';
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
      <DialogContent className="sm:max-w-[480px]" style={{ background: 'var(--panel)', borderColor: 'var(--panel-edge)', borderRadius: 'var(--radius-lg)', boxShadow: 'var(--shadow-chamber)' }}>
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full flex items-center justify-center" style={{ background: 'rgba(143, 255, 208, 0.08)', border: '1px solid rgba(143, 255, 208, 0.15)' }}>
              <Shield className="w-5 h-5" style={{ color: 'var(--status-ok)' }} />
            </div>
            <div>
              <DialogTitle style={{ fontFamily: 'var(--font-display)' }}>{isSetup ? 'Set Up Vault Passphrase' : 'Unlock Vault'}</DialogTitle>
              <DialogDescription style={{ color: 'var(--text-dim)' }}>
                {isSetup
                  ? 'Create a passphrase to encrypt your API keys. This passphrase will be required to view your keys.'
                  : 'Enter your vault passphrase to access your stored API keys.'}
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="p-3 text-sm rounded-[var(--radius)]" style={{ background: 'rgba(255, 107, 107, 0.06)', border: '1px solid rgba(255, 107, 107, 0.2)', color: 'var(--status-revoked)' }}>
              {error}
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="vault-passphrase" style={{ color: 'var(--text)' }}>
              {isSetup ? 'Create Passphrase' : 'Passphrase'} <span style={{ color: 'var(--status-revoked)' }}>*</span>
            </Label>
            <div className="relative">
              <Input
                id="vault-passphrase"
                type={showPassphrase ? 'text' : 'password'}
                placeholder={isSetup ? 'Create a strong passphrase' : 'Enter your passphrase'}
                value={passphrase}
                onChange={(e) => setPassphrase(e.target.value)}
                style={error ? { borderColor: 'var(--status-revoked)' } : undefined}
                className="pr-10"
                disabled={isLoading}
              />
              <button
                type="button"
                onClick={() => setShowPassphrase(!showPassphrase)}
                className="absolute right-3 top-1/2 -translate-y-1/2 transition-colors"
                style={{ color: 'var(--text-faint)' }}
              >
                {showPassphrase ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              </button>
            </div>
            {isSetup && passphrase && <PasswordStrengthIndicator password={passphrase} />}
          </div>

          {isSetup && (
            <div className="space-y-2">
              <Label htmlFor="confirm-passphrase" style={{ color: 'var(--text)' }}>
                Confirm Passphrase <span style={{ color: 'var(--status-revoked)' }}>*</span>
              </Label>
              <Input
                id="confirm-passphrase"
                type={showPassphrase ? 'text' : 'password'}
                placeholder="Confirm your passphrase"
                value={confirmPassphrase}
                onChange={(e) => setConfirmPassphrase(e.target.value)}
                style={error ? { borderColor: 'var(--status-revoked)' } : undefined}
                disabled={isLoading}
              />
            </div>
          )}

          {isSetup && (
            <div className="p-4 rounded-[var(--radius)]" style={{ background: 'rgba(232, 196, 104, 0.04)', border: '1px solid rgba(232, 196, 104, 0.15)' }}>
              <div className="flex items-start gap-3">
                <Lock className="w-5 h-5 mt-0.5" style={{ color: 'var(--status-pending)' }} />
                <div className="text-sm" style={{ color: 'var(--status-pending)' }}>
                  <p className="font-medium mb-1">Important: Save Your Passphrase</p>
                  <p style={{ color: 'var(--text-dim)' }}>
                    Your passphrase is used to encrypt your API keys locally. We cannot recover this
                    passphrase if you lose it. Store it securely in a password manager.
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
