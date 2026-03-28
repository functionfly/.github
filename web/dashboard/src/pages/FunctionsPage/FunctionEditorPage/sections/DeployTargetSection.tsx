import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Loader2, Server } from 'lucide-react';
import { Link } from 'react-router-dom';
import { FieldError } from '../components/editor-ui';
import type { FunctionEditorModel } from '../useFunctionEditor';

type Props = { editor: FunctionEditorModel };

export function DeployTargetSection({ editor }: Props) {
  const {
    isEditing,
    linkedAppId,
    deployBackendsLoading,
    filteredDeployBackends,
    selectedDeployBackendId,
    setDeployBackendId,
    errors,
  } = editor;

  if (!isEditing) {
    return (
      <Card className="card border-border-subtle/50" style={{ background: 'var(--bg-secondary)' }}>
        <CardHeader className="pb-2 pt-4 px-5">
          <CardTitle className="text-sm font-semibold text-text-primary flex items-center gap-2">
            <Server className="w-4 h-4 text-indigo-400" />
            Deploy target
          </CardTitle>
        </CardHeader>
        <CardContent className="px-5 pb-4">
          <p className="text-xs text-text-muted leading-relaxed">
            Save this function as a draft first. You can then choose an app backend and deploy.
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="card border-border-subtle/50" style={{ background: 'var(--bg-secondary)' }}>
      <CardHeader className="pb-2 pt-4 px-5">
        <CardTitle className="text-sm font-semibold text-text-primary flex items-center gap-2">
          <Server className="w-4 h-4 text-indigo-400" />
          Deploy target
        </CardTitle>
        {linkedAppId ? (
          <p className="text-xs text-text-muted mt-1 font-normal leading-relaxed">
            This function is linked to an app — only backends for that app are shown.
          </p>
        ) : null}
      </CardHeader>
      <CardContent className="px-5 pb-4 space-y-2">
        {deployBackendsLoading ? (
          <div className="flex items-center gap-2 text-xs text-text-muted">
            <Loader2 className="w-3.5 h-3.5 animate-spin shrink-0" />
            Loading backends…
          </div>
        ) : filteredDeployBackends.length === 0 ? (
          <p className="text-xs text-text-muted leading-relaxed">
            No backends found.
            <Link to="/apps" className="text-indigo-400 hover:underline ml-1">
              Open Apps
            </Link>
            to create an app and register a backend.
          </p>
        ) : (
          <>
            <Label htmlFor="deploy-backend" className="text-xs text-text-muted">
              Backend
            </Label>
            <Select value={selectedDeployBackendId} onValueChange={setDeployBackendId}>
              <SelectTrigger id="deploy-backend" className="h-9 text-xs">
                <SelectValue placeholder="Select backend" />
              </SelectTrigger>
              <SelectContent>
                {filteredDeployBackends.map((b) => (
                  <SelectItem key={b.id} value={b.id} className="text-xs">
                    {b.appName} · {b.provider} / {b.region}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <FieldError message={errors.deployBackend} />
          </>
        )}
      </CardContent>
    </Card>
  );
}
