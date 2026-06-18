import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
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
import {
  useCreateLicense,
  useMarketplaceFunctions,
  useMarketplaceLicenses,
  useRevokeLicense,
  type MarketplaceLicense,
} from '@/hooks/useStudioMarketplace';
import {
  LicenseManager as EconomyLicenseManager,
} from '@functionfly/ui-marketplace-economy/components';
import type { License } from '@functionfly/ui-marketplace-economy/types';
import { useMemo, useState } from 'react';
import { toast } from 'sonner';

function toUiLicense(license: MarketplaceLicense): License {
  return {
    id: license.id,
    key: license.key ?? '',
    type: license.type,
    functionId: license.functionId,
    functionName: license.functionName,
    purchaserId: license.purchaserId,
    purchaserName: license.purchaserName,
    issuedAt: license.issuedAt ? new Date(license.issuedAt).getTime() : Date.now(),
    expiresAt: license.expiresAt ? new Date(license.expiresAt).getTime() : null,
    maxActivations: license.maxActivations ?? null,
    activationCount: license.activationCount ?? 0,
    revoked: license.revoked ?? false,
  };
}

interface LicenseManagerPanelProps {
  className?: string;
}

export function LicenseManagerPanel({ className }: LicenseManagerPanelProps) {
  const { data, isLoading } = useMarketplaceLicenses();
  const { data: functionsData } = useMarketplaceFunctions();
  const createLicense = useCreateLicense();
  const revokeLicense = useRevokeLicense();

  const [generateOpen, setGenerateOpen] = useState(false);
  const [generatedKey, setGeneratedKey] = useState<string | null>(null);
  const [form, setForm] = useState({
    functionId: '',
    functionName: '',
    licenseType: 'commercial' as 'open' | 'restricted' | 'commercial',
    purchaserId: '',
    purchaserName: '',
    maxActivations: '',
    expiresAt: '',
  });

  const licenses = useMemo(() => (data?.licenses ?? []).map(toUiLicense), [data?.licenses]);

  const functions = functionsData?.functions ?? [];

  const openGenerate = () => {
    const first = functions[0];
    setForm({
      functionId: first?.id ?? '',
      functionName: first?.name ?? '',
      licenseType: 'commercial',
      purchaserId: '',
      purchaserName: '',
      maxActivations: '',
      expiresAt: '',
    });
    setGeneratedKey(null);
    setGenerateOpen(true);
  };

  const handleGenerate = async () => {
    if (!form.functionId) {
      toast.error('Select a function');
      return;
    }

    const selected = functions.find((fn) => fn.id === form.functionId);
    const result = await createLicense.mutateAsync({
      functionId: form.functionId,
      functionName: form.functionName || selected?.name || form.functionId,
      type: form.licenseType,
      purchaserId: form.purchaserId || 'anonymous',
      purchaserName: form.purchaserName || undefined,
      maxActivations: form.maxActivations ? Number(form.maxActivations) : undefined,
      expiresAt: form.expiresAt ? new Date(form.expiresAt).toISOString() : undefined,
    });

    if (result.license?.key) {
      setGeneratedKey(result.license.key);
      toast.success('Copy the license key now — it will not be shown again.');
    } else {
      setGenerateOpen(false);
    }
  };

  const handleRevoke = (licenseId: string) => {
    revokeLicense.mutate(licenseId);
  };

  if (isLoading) {
    return (
      <div
        className={`flex items-center justify-center h-64 text-aviation-text-muted ${className ?? ''}`}
      >
        Loading licenses…
      </div>
    );
  }

  return (
    <>
      <EconomyLicenseManager
        licenses={licenses}
        totalActive={data?.totalActive ?? 0}
        totalRevoked={data?.totalRevoked ?? 0}
        onLicenseRevoke={handleRevoke}
        onLicenseGenerate={openGenerate}
        className={className}
      />

      <Dialog open={generateOpen} onOpenChange={setGenerateOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{generatedKey ? 'License generated' : 'Generate license'}</DialogTitle>
          </DialogHeader>

          {generatedKey ? (
            <div className="space-y-3">
              <p className="text-sm text-muted-foreground">
                Share this key with the purchaser. It is only shown once.
              </p>
              <code className="block w-full break-all rounded-md bg-muted p-3 text-xs font-mono">
                {generatedKey}
              </code>
              <Button
                type="button"
                variant="secondary"
                className="w-full"
                onClick={() => {
                  void navigator.clipboard.writeText(generatedKey);
                  toast.success('License key copied');
                }}
              >
                Copy key
              </Button>
            </div>
          ) : (
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="license-function">Function</Label>
                <Select
                  value={form.functionId}
                  onValueChange={(value) => {
                    const fn = functions.find((item) => item.id === value);
                    setForm((prev) => ({
                      ...prev,
                      functionId: value,
                      functionName: fn?.name ?? prev.functionName,
                    }));
                  }}
                >
                  <SelectTrigger id="license-function">
                    <SelectValue placeholder="Select function" />
                  </SelectTrigger>
                  <SelectContent>
                    {functions.map((fn) => (
                      <SelectItem key={fn.id} value={fn.id}>
                        {fn.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="license-type">License type</Label>
                <Select
                  value={form.licenseType}
                  onValueChange={(value) =>
                    setForm((prev) => ({
                      ...prev,
                      licenseType: value as 'open' | 'restricted' | 'commercial',
                    }))
                  }
                >
                  <SelectTrigger id="license-type">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="open">Open</SelectItem>
                    <SelectItem value="restricted">Restricted</SelectItem>
                    <SelectItem value="commercial">Commercial</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="license-purchaser">Purchaser name</Label>
                <Input
                  id="license-purchaser"
                  value={form.purchaserName}
                  onChange={(e) => setForm((prev) => ({ ...prev, purchaserName: e.target.value }))}
                  placeholder="Acme Corp"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-2">
                  <Label htmlFor="license-max">Max activations</Label>
                  <Input
                    id="license-max"
                    type="number"
                    min={1}
                    value={form.maxActivations}
                    onChange={(e) =>
                      setForm((prev) => ({ ...prev, maxActivations: e.target.value }))
                    }
                    placeholder="Optional"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="license-expires">Expires</Label>
                  <Input
                    id="license-expires"
                    type="date"
                    value={form.expiresAt}
                    onChange={(e) => setForm((prev) => ({ ...prev, expiresAt: e.target.value }))}
                  />
                </div>
              </div>
            </div>
          )}

          <DialogFooter>
            {generatedKey ? (
              <Button type="button" onClick={() => setGenerateOpen(false)}>
                Done
              </Button>
            ) : (
              <>
                <Button type="button" variant="ghost" onClick={() => setGenerateOpen(false)}>
                  Cancel
                </Button>
                <Button
                  type="button"
                  onClick={() => void handleGenerate()}
                  disabled={createLicense.isPending}
                >
                  {createLicense.isPending ? 'Generating…' : 'Generate'}
                </Button>
              </>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
