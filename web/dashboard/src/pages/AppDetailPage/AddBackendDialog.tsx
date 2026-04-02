import { appsApi } from '@/api/apps';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { cn } from '@/lib/utils';
import axios from 'axios';
import { Globe, Key, Loader2, MapPin, Server } from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';
import { toast } from 'sonner';

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
  'arn1',
  'bom1',
  'cdg1',
  'cle1',
  'cpt1',
  'dub1',
  'fra1',
  'gru1',
  'hkg1',
  'hnd1',
  'iad1',
  'icn1',
  'jnb1',
  'lax1',
  'lhr1',
  'pdx1',
  'sfo1',
  'sin1',
  'syd1',
] as const;

/** Keep in sync with internal/adapters/fly GetRegions */
const FLY_REGIONS = [
  'ams',
  'arn',
  'atl',
  'bog',
  'bos',
  'bru',
  'cdg',
  'den',
  'dfw',
  'ewr',
  'eze',
  'fra',
  'gig',
  'gru',
  'hkg',
  'iad',
  'jnb',
  'lax',
  'lhr',
  'mad',
  'mia',
  'nrt',
  'ord',
  'phx',
  'pma',
  'sea',
  'sfo',
  'sin',
  'syd',
  'tsn',
  'waw',
  'yyz',
] as const;

/** Keep in sync with internal/adapters/deno GetRegions */
const DENO_REGIONS = [
  { value: 'us-east4', label: 'US East (N. Virginia)' },
  { value: 'europe-west4', label: 'Europe (Eemshaven)' },
  { value: 'asia-southeast1', label: 'Asia Pacific (Jurong)' },
  { value: 'us-west2', label: 'US West (Los Angeles)' },
] as const;

/** Example URL + default region when user switches provider (passes adapter ValidateConfig shape). */
const PROVIDER_PRESETS: Record<
  AddBackendProvider,
  { region: string; url: string; urlHint: string }
> = {
  'functionfly-edge': {
    region: 'us-east-1',
    url: 'https://edge.functionfly.com',
    urlHint: 'FunctionFly managed edge — use this host unless directed otherwise.',
  },
  workers: {
    region: 'us-east-1',
    url: 'https://my-worker.my-subdomain.workers.dev',
    urlHint: 'Replace with your Worker URL (*.workers.dev or a custom domain).',
  },
  vercel: {
    region: 'iad1',
    url: 'https://my-deployment.vercel.app',
    urlHint: 'Replace with your deployment (*.vercel.app or custom domain).',
  },
  fly: {
    region: 'iad',
    url: 'https://my-app.fly.dev',
    urlHint: 'Replace with your Fly app URL (*.fly.dev, *.internal, or custom domain).',
  },
  'deno-deploy': {
    region: 'us-east4',
    url: 'https://my-project.deno.dev',
    urlHint: 'Replace with your Deno Deploy URL (*.deno.dev or custom domain).',
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
  const [sharedSecret, setSharedSecret] = useState('');
  const [priorityStr, setPriorityStr] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const reset = useCallback(() => {
    const p = PROVIDER_PRESETS['functionfly-edge'];
    setProvider('functionfly-edge');
    setRegion(p.region);
    setUrl(p.url);
    setSharedSecret('');
    setPriorityStr('');
  }, []);

  useEffect(() => {
    if (!open) reset();
  }, [open, reset]);

  const handleProviderChange = (next: string) => {
    const preset = PROVIDER_PRESETS[next as AddBackendProvider];
    if (!preset) return;
    setProvider(next);
    setRegion(preset.region);
    setUrl(preset.url);
  };

  const regionOptions = regionsForProvider(provider);
  const urlHint = PROVIDER_PRESETS[provider as AddBackendProvider]?.urlHint ?? '';

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
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg p-0 overflow-hidden border-border/50 shadow-xl">
        <form onSubmit={handleSubmit}>
          <div className="relative">
            {/* Header gradient background */}
            <div className="absolute inset-0 h-32 bg-gradient-to-br from-brand-500/[0.06] via-purple-500/[0.02] to-transparent dark:from-brand-500/10 dark:via-purple-500/5 pointer-events-none" />

            <DialogHeader className="relative px-6 pt-6 pb-4">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-brand-500/[0.08] to-purple-500/[0.03] border border-brand-500/12 dark:from-brand-500/15 dark:to-purple-500/10 dark:border-brand-500/20 flex items-center justify-center">
                  <Server className="w-5 h-5 text-brand-500" />
                </div>
                <div>
                  <DialogTitle className="text-lg font-semibold tracking-tight">
                    Add Backend
                  </DialogTitle>
                  <DialogDescription className="text-sm text-muted-foreground/80 mt-0.5">
                    Connect a deploy target for this app
                  </DialogDescription>
                </div>
              </div>
            </DialogHeader>
          </div>

          <div className="px-6 py-4 space-y-5">
            {/* Provider Select */}
            <div className="space-y-2.5">
              <Label
                htmlFor="add-backend-provider"
                className="flex items-center gap-2 text-sm font-medium"
              >
                <Server className="w-3.5 h-3.5 text-muted-foreground" />
                Provider
              </Label>
              <Select value={provider} onValueChange={handleProviderChange}>
                <SelectTrigger
                  id="add-backend-provider"
                  className="h-11 bg-background/50 backdrop-blur-sm border-border/50 focus:border-brand-500/50 focus:ring-brand-500/20"
                >
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="border-border/50">
                  {PROVIDERS.map((p) => (
                    <SelectItem key={p.value} value={p.value} className="focus:bg-brand-500/10">
                      {p.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Region — options match backend adapter GetRegions / ValidateConfig */}
            <div className="space-y-2.5">
              <Label
                htmlFor="add-backend-region"
                className="flex items-center gap-2 text-sm font-medium"
              >
                <MapPin className="w-3.5 h-3.5 text-muted-foreground" />
                Region
              </Label>
              {regionOptions && regionOptions.length > 0 ? (
                <Select value={region} onValueChange={setRegion}>
                  <SelectTrigger
                    id="add-backend-region"
                    className="h-11 bg-background/50 backdrop-blur-sm border-border/50 focus:border-brand-500/50 focus:ring-brand-500/20"
                  >
                    <SelectValue placeholder="Select region" />
                  </SelectTrigger>
                  <SelectContent className="border-border/50 max-h-60">
                    {regionOptions.map((r) => (
                      <SelectItem key={r.value} value={r.value} className="focus:bg-brand-500/10">
                        {r.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              ) : (
                <Input
                  id="add-backend-region"
                  value={region}
                  onChange={(e) => setRegion(e.target.value)}
                  placeholder="Region code"
                  autoComplete="off"
                  className="h-11 bg-background/50 backdrop-blur-sm border-border/50 focus:border-brand-500/50 focus:ring-brand-500/20"
                />
              )}
            </div>

            {/* URL Input */}
            <div className="space-y-2.5">
              <Label
                htmlFor="add-backend-url"
                className="flex items-center gap-2 text-sm font-medium"
              >
                <Globe className="w-3.5 h-3.5 text-muted-foreground" />
                Base URL
              </Label>
              <Input
                id="add-backend-url"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                placeholder="https://…"
                autoComplete="off"
                className="h-11 bg-background/50 backdrop-blur-sm border-border/50 focus:border-brand-500/50 focus:ring-brand-500/20 font-mono text-sm"
              />
              {urlHint ? (
                <p className="text-xs text-muted-foreground/80 leading-relaxed">{urlHint}</p>
              ) : null}
            </div>

            {/* Shared Secret Input */}
            <div className="space-y-2.5">
              <Label
                htmlFor="add-backend-secret"
                className="flex items-center gap-2 text-sm font-medium"
              >
                <Key className="w-3.5 h-3.5 text-muted-foreground" />
                Shared Secret{' '}
                <span className="text-muted-foreground/60 font-normal">(optional)</span>
              </Label>
              <Input
                id="add-backend-secret"
                type="password"
                value={sharedSecret}
                onChange={(e) => setSharedSecret(e.target.value)}
                placeholder="Leave blank to auto-generate"
                autoComplete="new-password"
                className="h-11 bg-background/50 backdrop-blur-sm border-border/50 focus:border-brand-500/50 focus:ring-brand-500/20"
              />
              <p className="text-xs text-muted-foreground/70">
                Leave empty to generate one automatically
              </p>
            </div>

            {/* Priority Input */}
            <div className="space-y-2.5">
              <Label
                htmlFor="add-backend-priority"
                className="flex items-center gap-2 text-sm font-medium"
              >
                <span className="w-3.5 h-3.5 flex items-center justify-center text-xs text-muted-foreground">
                  #
                </span>
                Priority <span className="text-muted-foreground/60 font-normal">(optional)</span>
              </Label>
              <Input
                id="add-backend-priority"
                value={priorityStr}
                onChange={(e) => setPriorityStr(e.target.value)}
                placeholder="Lower runs first when routing"
                inputMode="numeric"
                autoComplete="off"
                className="h-11 bg-background/50 backdrop-blur-sm border-border/50 focus:border-brand-500/50 focus:ring-brand-500/20"
              />
            </div>
          </div>

          <DialogFooter className="px-6 py-5 gap-2 sm:gap-3 border-t border-border/30 bg-muted/20">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              className="px-5 hover:border-brand-500/30 hover:bg-brand-500/5 transition-all"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={submitting}
              className={cn(
                'px-5 gap-2 transition-all',
                submitting ? 'opacity-70' : 'hover:shadow-[0_0_20px_rgba(99,102,241,0.3)]'
              )}
            >
              {submitting ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Adding…
                </>
              ) : (
                <>
                  <Server className="w-4 h-4" />
                  Add Backend
                </>
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
