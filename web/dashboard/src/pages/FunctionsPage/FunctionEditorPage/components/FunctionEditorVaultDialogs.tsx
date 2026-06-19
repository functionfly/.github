import { SecretRevealGate } from '@/components/SecretsVault';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { ScrollArea } from '@/components/ui/scroll-area';
import { useTranslation } from 'react-i18next';
import type { FunctionEditorModel } from '../useFunctionEditor';
import { Key, Loader2, Shield } from 'lucide-react';

type Props = { editor: FunctionEditorModel };

export function FunctionEditorVaultDialogs({ editor }: Props) {
  const { t } = useTranslation();
  const {
    vaultSecrets,
    vaultPickerOpen,
    setVaultPickerOpen,
    pickingSecretId,
    pendingSecretForDecrypt,
    setPendingSecretForDecrypt,
    vaultDecryptPassphrase,
    setVaultDecryptPassphrase,
    handleSelectVaultSecret,
    handleConfirmVaultPassphrase,
    revealEnvVarId,
    setRevealEnvVarId,
    revealGateOpen,
    setRevealGateOpen,
    handleRevealVerified,
    envVars,
  } = editor;

  const revealTarget = revealEnvVarId
    ? envVars.find((e) => e.id === revealEnvVarId)
    : undefined;

  return (
    <>
      <Dialog open={vaultPickerOpen} onOpenChange={setVaultPickerOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Key className="w-5 h-5" />
              {t('funcEditor.useSecretFromVault')}
            </DialogTitle>
            <DialogDescription>{t('funcEditor.chooseSecretDescription')}</DialogDescription>
          </DialogHeader>
          <ScrollArea className="max-h-[280px] rounded-md border border-border-subtle">
            <div className="p-2 space-y-1">
              {!vaultSecrets?.secrets?.length && (
                <p className="text-sm text-text-secondary py-4 text-center">
                  {t('funcEditor.noSecretsInVault')}
                </p>
              )}
              {vaultSecrets?.secrets?.map((secret) => (
                <button
                  key={secret.id}
                  type="button"
                  onClick={() => handleSelectVaultSecret(secret)}
                  disabled={pickingSecretId !== null}
                  className="w-full flex items-center gap-3 rounded-lg px-3 py-2.5 text-left transition-colors hover:bg-bg-hover disabled:opacity-50"
                >
                  <Shield className="w-4 h-4 text-text-muted shrink-0" />
                  <div className="flex-1 min-w-0">
                    <span className="text-sm font-medium text-text-primary block truncate">
                      {secret.name}
                    </span>
                    {secret.description && (
                      <span className="text-xs text-text-secondary block truncate">
                        {secret.description}
                      </span>
                    )}
                  </div>
                  {pickingSecretId === secret.id && (
                    <Loader2 className="w-4 h-4 animate-spin text-text-muted shrink-0" />
                  )}
                </button>
              ))}
            </div>
          </ScrollArea>
        </DialogContent>
      </Dialog>

      <Dialog
        open={!!pendingSecretForDecrypt}
        onOpenChange={(open) => {
          if (!open) {
            setPendingSecretForDecrypt(null);
            setVaultDecryptPassphrase('');
          }
        }}
      >
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Key className="w-5 h-5" />
              {t('funcEditor.vaultPassphrase')}
            </DialogTitle>
            <DialogDescription>
              {t('funcEditor.enterVaultPassphrase', { name: pendingSecretForDecrypt?.name })}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <Label htmlFor="vault-passphrase">{t('funcEditor.passphrase')}</Label>
              <Input
                id="vault-passphrase"
                type="password"
                placeholder={t('funcEditor.enterEncryptionPassphrase')}
                value={vaultDecryptPassphrase}
                onChange={(e) => setVaultDecryptPassphrase(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && void handleConfirmVaultPassphrase()}
                autoComplete="off"
                className="font-mono"
              />
            </div>
            <div className="flex justify-end gap-2">
              <Button
                variant="outline"
                onClick={() => {
                  setPendingSecretForDecrypt(null);
                  setVaultDecryptPassphrase('');
                }}
              >
                {t('funcEditor.cancel')}
              </Button>
              <Button
                onClick={() => void handleConfirmVaultPassphrase()}
                disabled={!vaultDecryptPassphrase.trim() || pickingSecretId !== null}
              >
                {pickingSecretId ? <Loader2 className="w-4 h-4 animate-spin" /> : t('funcEditor.useSecret')}
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {revealTarget?.isSecret && (
        <SecretRevealGate
          trigger={<span className="hidden" aria-hidden />}
          isOpen={revealGateOpen}
          onOpenChange={(open) => {
            setRevealGateOpen(open);
            if (!open) setRevealEnvVarId(null);
          }}
          onVerified={() => handleRevealVerified(revealTarget)}
          onCancelled={() => {
            setRevealGateOpen(false);
            setRevealEnvVarId(null);
          }}
          title={t('funcEditor.revealSecretValue')}
          description={t('funcEditor.revealSecretDescription', { key: revealTarget.key })}
          requiredLevel="medium"
        />
      )}
    </>
  );
}
