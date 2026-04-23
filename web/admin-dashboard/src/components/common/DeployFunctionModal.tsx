import { useState, useEffect } from 'react';
import { useQuery, useMutation } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { Dialog, DialogContent, DialogHeader, DialogFooter } from '@/components/ui/Dialog';
import { Input } from '@/components/ui/Input';
import { Select } from '@/components/ui/Select';
import { Button } from '@/components/ui/Button';
import { cn } from '@/lib/utils';
import { Loader2, Cloud, AlertTriangle } from 'lucide-react';
import { toast } from 'sonner';

interface Backend {
  id: string;
  app_id: string;
  app_name: string;
  tenant_id: string;
  tenant_name: string;
  provider: string;
  region: string;
  url: string;
  enabled: boolean;
  priority: number;
}

interface DeployFunctionModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  functionId?: string;
  functionName?: string;
}

interface DeployFormData {
  functionId: string;
  backendId: string;
  version: string;
  environment: 'dev' | 'staging' | 'prod';
}

const environmentOptions = [
  { value: 'dev', label: 'Development', color: '#10b981' },
  { value: 'staging', label: 'Staging', color: '#f59e0b' },
  { value: 'prod', label: 'Production', color: '#ef4444' },
] as const;

export function DeployFunctionModal({ open, onOpenChange, functionId, functionName }: DeployFunctionModalProps) {
  const [formData, setFormData] = useState<DeployFormData>({
    functionId: functionId || '',
    backendId: '',
    version: '1.0.0',
    environment: 'prod',
  });
  const [backendError, setBackendError] = useState<string | undefined>();

  const { data: backendsData, isLoading: backendsLoading } = useQuery({
    queryKey: ['admin-backends'],
    queryFn: async () => {
      const res = await adminApiClient.get<{ backends: Backend[] }>('/backends');
      return res;
    },
    enabled: open,
    staleTime: 1000 * 60 * 5,
  });

  const deployMutation = useMutation({
    mutationFn: async (data: DeployFormData) => {
      const payload = {
        function_id: data.functionId,
        backend_id: data.backendId,
        version: data.version,
        environment: data.environment,
      };
      return adminApiClient.post('/functions/deploy', payload);
    },
    onSuccess: (_, variables) => {
      toast.success('Function deployed successfully', {
        description: `${variables.functionId || functionName} deployed to backend`,
      });
      onOpenChange(false);
      setFormData({ functionId: '', backendId: '', version: '1.0.0', environment: 'prod' });
    },
    onError: (error: any) => {
      toast.error('Failed to deploy function', {
        description: error?.response?.data?.error || error?.message || 'Unknown error',
      });
    },
  });

  useEffect(() => {
    if (functionId && open) {
      setFormData((prev) => ({ ...prev, functionId }));
    }
  }, [functionId, open]);

  useEffect(() => {
    if (!open) {
      setFormData({ functionId: '', backendId: '', version: '1.0.0', environment: 'prod' });
      setBackendError(undefined);
    }
  }, [open]);

  const backends = backendsData?.backends || [];
  const enabledBackends = backends.filter((b) => b.enabled);

  const backendOptions = enabledBackends.map((b) => ({
    value: b.id,
    label: `${b.app_name} (${b.provider}/${b.region}) - ${b.tenant_name}`,
  }));

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.backendId) {
      setBackendError('Please select a backend');
      return;
    }
    deployMutation.mutate(formData);
  };

  const isLoading = deployMutation.isPending;
  const selectedBackend = backends.find((b) => b.id === formData.backendId);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg max-h-[90vh] overflow-y-auto">
        <DialogHeader
          title="Deploy Function"
          description={functionName ? `Deploy "${functionName}" to a backend` : 'Deploy a function to a backend'}
        />

        <form onSubmit={handleSubmit} className="space-y-5">
          <Input
            label="Function ID"
            placeholder="Enter function ID (UUID)"
            value={formData.functionId}
            onChange={(e) => setFormData({ ...formData, functionId: e.target.value })}
            required
            disabled={!!functionId}
            helperText={!functionId ? 'The UUID of the function to deploy' : undefined}
          />

          <Input
            label="Version"
            placeholder="1.0.0"
            value={formData.version}
            onChange={(e) => setFormData({ ...formData, version: e.target.value })}
            required
            pattern="^\d+\.\d+\.\d+$"
            helperText="Semantic version (x.y.z)"
          />

          <Select
            label="Backend"
            placeholder={backendsLoading ? 'Loading backends...' : 'Select a backend'}
            options={backendOptions}
            value={formData.backendId}
            onChange={(value) => {
              setFormData({ ...formData, backendId: value });
              setBackendError(undefined);
            }}
            error={backendError}
            disabled={backendsLoading || backends.length === 0}
            required
          />

          {backends.length === 0 && !backendsLoading && (
            <div className="flex items-start gap-3 p-3 rounded-lg bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800">
              <AlertTriangle className="w-5 h-5 text-amber-600 dark:text-amber-400 shrink-0 mt-0.5" />
              <div>
                <p className="text-sm font-medium text-amber-800 dark:text-amber-200">No backends available</p>
                <p className="text-sm text-amber-700 dark:text-amber-300 mt-0.5">Create a backend first before deploying functions.</p>
              </div>
            </div>
          )}

          <Select
            label="Environment"
            options={environmentOptions.map((o) => ({ value: o.value, label: o.label }))}
            value={formData.environment}
            onChange={(value) => setFormData({ ...formData, environment: value as any })}
            required
          />

          <div className="rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900/50 p-4">
            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">Deployment Summary</h4>
            <div className="grid grid-cols-2 gap-3 text-sm">
              <div className="flex flex-col gap-0.5">
                <span className="text-gray-500 dark:text-gray-400 text-xs">Function ID</span>
                <span className="font-mono text-gray-900 dark:text-gray-100 text-xs bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded truncate">
                  {formData.functionId || '—'}
                </span>
              </div>
              <div className="flex flex-col gap-0.5">
                <span className="text-gray-500 dark:text-gray-400 text-xs">Version</span>
                <span className="text-gray-900 dark:text-gray-100">{formData.version}</span>
              </div>
              <div className="flex flex-col gap-0.5">
                <span className="text-gray-500 dark:text-gray-400 text-xs">Backend</span>
                <span className="text-gray-900 dark:text-gray-100 truncate">
                  {selectedBackend ? `${selectedBackend.provider}/${selectedBackend.region}` : '—'}
                </span>
              </div>
              <div className="flex flex-col gap-0.5">
                <span className="text-gray-500 dark:text-gray-400 text-xs">Environment</span>
                <span className={cn(
                  'capitalize',
                  formData.environment === 'prod' ? 'text-red-600 dark:text-red-400' :
                  formData.environment === 'staging' ? 'text-amber-600 dark:text-amber-400' :
                  'text-green-600 dark:text-green-400'
                )}>
                  {formData.environment}
                </span>
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={isLoading || !formData.functionId || !formData.backendId || enabledBackends.length === 0}
            >
              {isLoading ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Deploying...
                </>
              ) : (
                <>
                  <Cloud className="w-4 h-4" />
                  Deploy
                </>
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}