import { isVaultPassphraseSet } from '@/services/vault-api-key-storage';

export function useVaultStatus() {
  return {
    isVaultPassphraseSet: isVaultPassphraseSet(),
  };
}
