import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { apiKeysService } from '@/services/api-keys';
import { APIKeyEnvironment, AvailableEnvironment } from '@/types/api-key';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { AlertCircle, Globe, Loader2, Plus, Trash2 } from 'lucide-react';
import { useState } from 'react';

interface APIKeyEnvironmentsProps {
  keyId: string;
}

export function APIKeyEnvironments({ keyId }: APIKeyEnvironmentsProps) {
  const queryClient = useQueryClient();
  const [showAddForm, setShowAddForm] = useState(false);
  const [selectedEnvId, setSelectedEnvId] = useState('');

  const {
    data: environments,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['api-key-environments', keyId],
    queryFn: () => apiKeysService.getEnvironments(keyId),
  });

  const { data: availableEnvironmentsList = [] } = useQuery({
    queryKey: ['api-keys-available-environments'],
    queryFn: () => apiKeysService.getAvailableEnvironments(),
  });

  const linkEnvironmentMutation = useMutation({
    mutationFn: (env: { id: string; name: string }) =>
      apiKeysService.linkEnvironment(keyId, { environment_id: env.id, environment_name: env.name }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['api-key-environments', keyId] });
      setShowAddForm(false);
      setSelectedEnvId('');
    },
  });

  const unlinkEnvironmentMutation = useMutation({
    mutationFn: (environmentId: string) => apiKeysService.unlinkEnvironment(keyId, environmentId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['api-key-environments', keyId] });
    },
  });

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  const handleLinkEnvironment = () => {
    if (!selectedEnvId) return;
    const env = availableEnvironments.find((e) => e.id === selectedEnvId);
    if (env) linkEnvironmentMutation.mutate({ id: env.id, name: env.name });
  };

  // Get already linked environment IDs
  const linkedEnvIds =
    (environments as APIKeyEnvironment[] | undefined)?.map((e) => e.environment_id) || [];

  // Environments that can still be linked (from API, excluding already linked)
  const availableEnvironments = (availableEnvironmentsList as AvailableEnvironment[]).filter(
    (env) => !linkedEnvIds.includes(env.id)
  );

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center py-8">
        <AlertCircle className="w-6 h-6 text-red-500 mx-auto mb-2" />
        <p className="text-sm text-muted-foreground">Failed to load environments</p>
      </div>
    );
  }

  const envs = environments as APIKeyEnvironment[] | undefined;

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle className="flex items-center gap-2">
          <Globe className="w-5 h-5" />
          Environments
        </CardTitle>
        <Button
          variant="outline"
          size="sm"
          onClick={() => setShowAddForm(!showAddForm)}
          disabled={availableEnvironments.length === 0}
        >
          <Plus className="w-4 h-4 mr-2" />
          Link Environment
        </Button>
      </CardHeader>
      <CardContent className="space-y-4">
        {showAddForm && (
          <div className="p-4 border rounded-lg space-y-4">
            <div className="space-y-2">
              <Label>Select Environment</Label>
              <select
                className="w-full px-3 py-2 border rounded-lg bg-background"
                value={selectedEnvId}
                onChange={(e) => setSelectedEnvId(e.target.value)}
              >
                <option value="">Select an environment...</option>
                {availableEnvironments.map((env) => (
                  <option key={env.id} value={env.id}>
                    {env.name}
                  </option>
                ))}
              </select>
            </div>
            <div className="flex justify-end gap-2">
              <Button variant="outline" size="sm" onClick={() => setShowAddForm(false)}>
                Cancel
              </Button>
              <Button
                size="sm"
                onClick={handleLinkEnvironment}
                disabled={!selectedEnvId || linkEnvironmentMutation.isPending}
              >
                {linkEnvironmentMutation.isPending ? 'Linking...' : 'Link'}
              </Button>
            </div>
          </div>
        )}

        {envs && envs.length > 0 ? (
          <div className="space-y-2">
            {envs.map((env) => (
              <div key={env.id} className="flex items-center justify-between p-3 border rounded-lg">
                <div className="flex items-center gap-3">
                  <Globe className="w-4 h-4 text-muted-foreground" />
                  <span className="font-medium">{env.environment_name}</span>
                </div>
                <div className="flex items-center gap-3">
                  <span className="text-xs text-muted-foreground">
                    Linked {formatDate(env.created_at)}
                  </span>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 text-red-600 hover:text-red-600"
                    onClick={() => unlinkEnvironmentMutation.mutate(env.environment_id)}
                    disabled={unlinkEnvironmentMutation.isPending}
                  >
                    <Trash2 className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-center py-8 text-muted-foreground">No environments linked</p>
        )}
      </CardContent>
    </Card>
  );
}
