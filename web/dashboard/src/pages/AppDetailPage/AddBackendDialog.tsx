import { appsApi } from '@/api/apps';
import { providersApi } from '@/api/providers';
import { vaultApi } from '@/api/vault';
import { Modal } from '@/components/containment/Modal';
import { SealedButton } from '@/components/containment/SealedButton';
import { FrameButton } from '@/components/containment/FrameButton';
import { Input } from '@/components/containment/Input';
import { VaultSetupDialog } from '@/components/api-keys/VaultSetupDialog';
import { isVaultPassphraseSet, getVaultPassphrase } from '@/services/vault-api-key-storage';
import { VaultCrypto } from '@/utils/vault-crypto';
import axios from 'axios';
import { Globe, Key, Loader2, MapPin, Server, Shield } from 'lucide-react';
import { useCallback, useEffect, useRef, useState } from 'react';
import { toast } from 'sonner';

interface DiscoveredWorker {
  id: string;
  name: string;
  url: string;
}

const PROVIDERS = [
  { value: 'functionfly-edge', label: 'FunctionFly Edge' },
  { value: 'workers', label: 'Cloudflare Workers' },
  { value: 'vercel', label: 'Vercel' },
  { value: 'fly', label: 'Fly.io' },
  { value: 'deno-deploy', label: 'Deno Deploy' },
] as const;

type AddBackendProvider = (typeof PROVIDERS)[number]['value'];

/** Must stay in sync with internal/adapters/functionfly GetRegions (US + Germany). */
const FUNCTIONFLY_EDGE_REGIONS = [
  { value: 'us-east-1', label: 'United States (us-east-1)' },
  { value: 'eu-central-1', label: 'Germany — Frankfurt (eu-central-1)' },
] as const;

/** Keep in sync with internal/adapters/cloudflare GetRegions */
const WORKERS_REGIONS = [
  { value: 'us-east-1', label: 'US East (N. Virginia)' },
  { value: 'us-west-1', label: 'US West (N. California)' },
  { value: 'eu-west-1', label: 'EU West (Ireland)' },
  { value: 'ap-southeast-1', label: 'Asia Pacific (Singapore)' },
  { value: 'ap-northeast-1', label: 'Asia Pacific (Tokyo)' },
  { value: 'ap-southeast-2', label: 'Asia Pacific (Sydney)' },
  { value: 'sa-east-1', label: 'South America (São Paulo)' },
  { value: 'af-south-1', label: 'Africa (Cape Town)' },
] as const;

/** Keep in sync with internal/adapters/vercel GetRegions */
const VERCEL_REGIONS = [
  'arn1', 'bom1', 'cdg1', 'cle1', 'cpt1', 'dub1', 'fra1', 'gru1',
  'hkg1', 'hnd1', 'iad1', 'icn1', 'jnb1', 'lax1', 'lhr1', 'pdx1',
  'sfo1', 'sin1', 'syd1',
] as const;

/** Keep in sync with internal/adapters/fly GetRegions */
const FLY_REGIONS = [
  'ams', 'arn', 'atl', 'bog', 'bos', 'bru', 'cdg', 'den', 'dfw', 'ewr',
  'eze', 'fra', 'gig', 'gru', 'hkg', 'iad', 'jnb', 'lax', 'lhr', 'mad',
  'mia', 'nrt', 'ord', 'phx', 'pma', 'sea', 'sfo', 'sin', 'syd', 'tsn',
  'waw', 'yyz',
] as const;

/** Keep in sync with internal/adapters/deno GetRegions */
const DENO_REGIONS = [
  { value: 'us-east4', label: 'US East (N. Virginia)' },
  { value: 'europe-west4', label: 'Europe (Eemshaven)' },
  { value: 'asia-southeast1', label: 'Asia Pacific (Jurong)' },
  { value: 'us-west2', label: 'US West (Los Angeles)' },
] as const;

/** Example URL + default region when user switches provider. */
const PROVIDER_PRESETS: Record<
  AddBackendProvider,
  { region: string; url: string; urlHint: string; apiKeyHint: string; apiKeyPlaceholder: string }
> = {
  'functionfly-edge': {
    region: 'us-east-1',
    url: 'https://edge.functionfly.com',
    urlHint: 'FunctionFly managed edge — use this host unless directed otherwise.',
    apiKeyHint: '',
    apiKeyPlaceholder: '',
  },
  workers: {
    region: 'us-east-1',
    url: '',
    urlHint: 'Your Worker URL from the Cloudflare dashboard → Workers & Pages → your worker → Settings. Looks like: https://my-worker.my-account.workers.dev',
    apiKeyHint: 'Cloudflare API token with Workers permissions. Create at dash.cloudflare.com → API Tokens.',
    apiKeyPlaceholder: 'xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx',
  },
  vercel: {
    region: 'iad1',
    url: '',
    urlHint: 'Your deployment URL from vercel.com → your project → Settings → Domains. Looks like: https://my-project.vercel.app',
    apiKeyHint: 'Vercel personal access token. Create at vercel.com/account/tokens.',
    apiKeyPlaceholder: 'vercel_xxxxxxxxxxxxxxxxxxxxxxxxxxxx',
  },
  fly: {
    region: 'iad',
    url: '',
    urlHint: 'Your Fly app URL from fly.io → your app → Dashboard. Looks like: https://my-app.fly.dev',
    apiKeyHint: 'Fly.io API token. Run `fly auth token` or create at fly.io/user/personal_access_tokens.',
    apiKeyPlaceholder: 'fo1_xxxxxxxxxxxxxxxxxxxxxxxxxxxx',
  },
  'deno-deploy': {
    region: 'us-east4',
    url: '',
    urlHint: 'Your Deno Deploy URL from dash.deno.com → your project → Settings. Looks like: https://my-project.deno.dev',
    apiKeyHint: 'Deno Deploy access token. Create at dash.deno.com/account#access-tokens.',
    apiKeyPlaceholder: 'ddp_xxxxxxxxxxxxxxxxxxxxxxxxxxxx',
  },
};

function regionsForProvider(
  provider: string
): readonly { readonly value: string; readonly label: string }[] | null {
  switch (provider) {
    case 'functionfly-edge':
      return FUNCTIONFLY_EDGE_REGIONS;
    case 'workers':
      return WORKERS_REGIONS;
    case 'vercel':
      return VERCEL_REGIONS.map((v) => ({ value: v, label: v }));
    case 'fly':
      return FLY_REGIONS.map((v) => ({ value: v, label: v }));
    case 'deno-deploy':
      return DENO_REGIONS;
    default:
      return null;
  }
}

export interface AddBackendDialogProps {
  appId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function AddBackendDialog({ appId, open, onOpenChange, onSuccess }: AddBackendDialogProps) {
  const [provider, setProvider] = useState<string>('functionfly-edge');
  const [region, setRegion] = useState('us-east-1');
  const [url, setUrl] = useState('https://edge.functionfly.com');
  const [apiKey, setApiKey] = useState('');
  const [saveToVault, setSaveToVault] = useState(true);
  const [sharedSecret, setSharedSecret] = useState('');
  const [priorityStr, setPriorityStr] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [vaultDialogOpen, setVaultDialogOpen] = useState(false);
  const [vaultUnlocked, setVaultUnlocked] = useState(isVaultPassphraseSet());
  const [discoveredWorkers, setDiscoveredWorkers] = useState<DiscoveredWorker[]>([]);
  const [discovering, setDiscovering] = useState(false);
  const discoverAbortRef = useRef<AbortController | null>(null);

  const reset = useCallback(() => {
    const p = PROVIDER_PRESETS['functionfly-edge'];
    setProvider('functionfly-edge');
    setRegion(p.region);
    setUrl(p.url);
    setApiKey('');
    setSaveToVault(true);
    setSharedSecret('');
    setPriorityStr('');
  }, []);

  useEffect(() => {
    if (!open) reset();
  }, [open, reset]);

  const regionOptions = regionsForProvider(provider);
  const preset = PROVIDER_PRESETS[provider as AddBackendProvider];
  const urlHint = preset?.urlHint ?? '';
  const apiKeyHint = preset?.apiKeyHint ?? '';
  const apiKeyPlaceholder = preset?.apiKeyPlaceholder ?? '';
  const needsApiKey = provider !== 'functionfly-edge';

  // Auto-discover provider resources (Workers, projects) when an API key is entered.
  useEffect(() => {
    if (!needsApiKey || apiKey.trim().length < 10) {
      setDiscoveredWorkers([]);
      return;
    }

    discoverAbortRef.current?.abort();
    const ctrl = new AbortController();
    discoverAbortRef.current = ctrl;

    const timeout = setTimeout(async () => {
      setDiscovering(true);
      try {
        const resources = await providersApi.discoverResources(provider, apiKey.trim());
        if (!ctrl.aborted) {
          setDiscoveredWorkers(resources);
          if (resources.length > 0 && !url) {
            setUrl(resources[0].url);
          }
        }
      } catch {
        // Silently fail — user can still enter URL manually
      } finally {
        if (!ctrl.aborted) setDiscovering(false);
      }
    }, 800);

    return () => {
      clearTimeout(timeout);
      ctrl.abort();
    };
  }, [provider, apiKey, needsApiKey]); // eslint-disable-line react-hooks/exhaustive-deps

  const handleProviderChange = (next: string) => {
    const preset = PROVIDER_PRESETS[next as AddBackendProvider];
    if (!preset) return;
    setProvider(next);
    setRegion(preset.region);
    setUrl(preset.url);
    setApiKey('');
    setDiscoveredWorkers([]);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    let priority: number | undefined;
    if (priorityStr.trim() !== '') {
      const n = Number(priorityStr);
      if (!Number.isFinite(n) || !Number.isInteger(n)) {
        toast.error('Priority must be a whole number.');
        return;
      }
      priority = n;
    }

    setSubmitting(true);
    try {
      // If an API key was provided, connect the provider credentials first.
      if (needsApiKey && apiKey.trim() !== '') {
        await providersApi.connectProvider({
          providerId: provider,
          apiKey: apiKey.trim(),
        });

        // Auto-save the provider API key to the user's vault so they can
        // view and manage it later from the Secrets Vault page.
        if (saveToVault && vaultUnlocked) {
          try {
            const passphrase = await getVaultPassphrase();
            if (passphrase) {
              const encrypted = await VaultCrypto.encryptWithPassphrase(apiKey.trim(), passphrase);
              const providerLabel = PROVIDERS.find((p) => p.value === provider)?.label ?? provider;
              const secretName = `${providerLabel} API Key`;
              try {
                await vaultApi.createSecret({
                  name: secretName,
                  description: `Provider credential for ${providerLabel} backend`,
                  secret_type: 'api_key',
                  encrypted_data: VaultCrypto.toPayload(encrypted),
                  metadata: { provider, source: 'add-backend' },
                });
                toast.success('API key saved to vault');
              } catch (vaultErr) {
                if (axios.isAxiosError(vaultErr) && vaultErr.response?.status === 409) {
                  toast.info(`Vault already has a secret named "${secretName}"`);
                } else {
                  toast.info('Backend added, but could not save key to vault');
                }
              }
            }
          } catch {
            // Passphrase retrieval failed — non-fatal.
          }
        }
      }

      await appsApi.createBackend(appId, {
        provider,
        region: region.trim(),
        url: url.trim(),
        ...(sharedSecret.trim() !== '' ? { sharedSecret: sharedSecret.trim() } : {}),
        ...(priority !== undefined ? { priority } : {}),
      });
      toast.success('Backend added');
      onOpenChange(false);
      onSuccess?.();
    } catch (err) {
      const ax = axios.isAxiosError(err) ? err : null;
      let message = 'Could not add backend. Check provider, region, and URL.';
      if (ax) {
        const d = ax.response?.data;
        if (typeof d === 'string' && d.trim()) message = d.trim();
      }
      toast.error(message);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <>
    <Modal open={open && !vaultDialogOpen} onClose={() => onOpenChange(false)} title="Add Backend" className="sc-add-backend__modal">
      <form onSubmit={handleSubmit} className="sc-add-backend">
        <p className="sc-add-backend__desc">Connect a deploy target for this app</p>

        {/* Provider */}
        <div className="sc-add-backend__field">
          <label htmlFor="add-backend-provider" className="sc-add-backend__label">
            <Server size={12} />
            Provider
          </label>
          <select
            id="add-backend-provider"
            className="sc-add-backend__select"
            value={provider}
            onChange={(e) => handleProviderChange(e.target.value)}
          >
            {PROVIDERS.map((p) => (
              <option key={p.value} value={p.value}>{p.label}</option>
            ))}
          </select>
        </div>

        {/* Region */}
        <div className="sc-add-backend__field">
          <label htmlFor="add-backend-region" className="sc-add-backend__label">
            <MapPin size={12} />
            Region
          </label>
          {regionOptions && regionOptions.length > 0 ? (
            <select
              id="add-backend-region"
              className="sc-add-backend__select"
              value={region}
              onChange={(e) => setRegion(e.target.value)}
            >
              {regionOptions.map((r) => (
                <option key={r.value} value={r.value}>{r.label}</option>
              ))}
            </select>
          ) : (
            <Input
              id="add-backend-region"
              value={region}
              onChange={(e) => setRegion(e.target.value)}
              placeholder="Region code"
              autoComplete="off"
            />
          )}
        </div>

        {/* API Key (provider credentials) — shown before URL so discovery can populate it */}
        {needsApiKey && (
          <div className="sc-add-backend__field">
            <label htmlFor="add-backend-apikey" className="sc-add-backend__label">
              <Key size={12} />
              API Key <span className="sc-add-backend__optional">(optional)</span>
            </label>
            <Input
              id="add-backend-apikey"
              type="password"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              placeholder={apiKeyPlaceholder}
              autoComplete="new-password"
            />
            {apiKeyHint && <p className="sc-add-backend__hint">{apiKeyHint}</p>}
            {/* Save to Vault toggle */}
            <label
              className="sc-add-backend__checkbox-row"
              onClick={!vaultUnlocked ? (e) => { e.preventDefault(); setVaultDialogOpen(true); } : undefined}
              style={!vaultUnlocked ? { cursor: 'pointer' } : undefined}
            >
              <input
                type="checkbox"
                checked={saveToVault && vaultUnlocked}
                onChange={(e) => setSaveToVault(e.target.checked)}
                disabled={!vaultUnlocked}
              />
              <Shield size={12} />
              <span>
                {vaultUnlocked
                  ? 'Save key to Secrets Vault'
                  : 'Unlock vault to save key'}
              </span>
            </label>
            {!vaultUnlocked && (
              <p className="sc-add-backend__hint" style={{ marginTop: '4px' }}>
                Set up your vault at <strong>Secrets → Vault</strong> first, or click here to unlock.
              </p>
            )}
          </div>
        )}

        {/* URL */}
        <div className="sc-add-backend__field">
          <label htmlFor="add-backend-url" className="sc-add-backend__label">
            <Globe size={12} />
            Base URL
          </label>
          {discoveredWorkers.length > 0 ? (
            <select
              id="add-backend-url"
              className="sc-add-backend__select"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
            >
              <option value="">Select a Worker…</option>
              {discoveredWorkers.map((w) => (
                <option key={w.id} value={w.url}>{w.name} — {w.url}</option>
              ))}
              <option value="__manual__">Enter URL manually…</option>
            </select>
          ) : (
            <Input
              id="add-backend-url"
              value={url === '__manual__' ? '' : url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder={discovering ? 'Discovering Workers…' : needsApiKey ? 'https://your-deployment-url.com' : 'https://…'}
              autoComplete="off"
              disabled={discovering}
            />
          )}
          {discovering && <p className="sc-add-backend__hint">Looking up your Workers…</p>}
          {urlHint && !discovering && <p className="sc-add-backend__hint">{urlHint}</p>}
        </div>

        {/* Shared Secret */}
        <div className="sc-add-backend__field">
          <label htmlFor="add-backend-secret" className="sc-add-backend__label">
            <Key size={12} />
            Shared Secret <span className="sc-add-backend__optional">(optional)</span>
          </label>
          <Input
            id="add-backend-secret"
            type="password"
            value={sharedSecret}
            onChange={(e) => setSharedSecret(e.target.value)}
            placeholder="Leave blank to auto-generate"
            autoComplete="new-password"
          />
          <p className="sc-add-backend__hint">Leave empty to generate one automatically</p>
        </div>

        {/* Priority */}
        <div className="sc-add-backend__field">
          <label htmlFor="add-backend-priority" className="sc-add-backend__label">
            <span className="sc-add-backend__label-hash">#</span>
            Priority <span className="sc-add-backend__optional">(optional)</span>
          </label>
          <Input
            id="add-backend-priority"
            value={priorityStr}
            onChange={(e) => setPriorityStr(e.target.value)}
            placeholder="Lower runs first when routing"
            inputMode="numeric"
            autoComplete="off"
          />
        </div>

        {/* Footer */}
        <div className="sc-add-backend__footer">
          <FrameButton type="button" onClick={() => onOpenChange(false)}>
            Cancel
          </FrameButton>
          <SealedButton
            type="submit"
            iconLeft={submitting ? <Loader2 size={14} className="sc-community-spinner" /> : <Server size={14} />}
            loading={submitting}
          >
            Add Backend
          </SealedButton>
        </div>
      </form>
    </Modal>

    {/* Vault unlock dialog — rendered outside Modal so it layers on top */}
    <VaultSetupDialog
      open={vaultDialogOpen}
      onOpenChange={setVaultDialogOpen}
      mode="unlock"
      onSuccess={() => {
        setVaultUnlocked(true);
        setSaveToVault(true);
      }}
    />
    </>
  );
}
