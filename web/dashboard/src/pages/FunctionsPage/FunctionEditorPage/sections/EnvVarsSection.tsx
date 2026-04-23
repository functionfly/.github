import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { Layers, Eye, EyeOff, Key, Plus, Shield, Upload, X } from 'lucide-react';
import { useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { EnvPresetsPicker, type EnvPreset } from '../components/EnvPresetsPicker';
import { SectionCard } from '../components/editor-ui';
import type { FunctionEditorModel } from '../useFunctionEditor';

type Props = { editor: FunctionEditorModel };

function parseEnvFile(content: string): Array<{ key: string; value: string }> {
  const lines = content.split('\n');
  const result: Array<{ key: string; value: string }> = [];
  for (const line of lines) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith('#')) continue;
    const eqIdx = trimmed.indexOf('=');
    if (eqIdx === -1) continue;
    const key = trimmed.slice(0, eqIdx).trim();
    let value = trimmed.slice(eqIdx + 1).trim();
    // Strip surrounding quotes
    if (
      (value.startsWith('"') && value.endsWith('"')) ||
      (value.startsWith("'") && value.endsWith("'"))
    ) {
      value = value.slice(1, -1);
    }
    if (key) result.push({ key, value });
  }
  return result;
}

export function EnvVarsSection({ editor }: Props) {
  const { t } = useTranslation();
  const {
    envVars,
    newEnvKey,
    setNewEnvKey,
    newEnvValue,
    setNewEnvValue,
    isNewEnvSecret,
    setIsNewEnvSecret,
    showEnvValues,
    setShowEnvValues,
    addEnvironmentVariable,
    removeEnvironmentVariable,
    setEnvVars,
    markDirty,
    setVaultPickerOpen,
  } = editor;

  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleImportEnv = () => {
    fileInputRef.current?.click();
  };

  const handleApplyPreset = (preset: EnvPreset) => {
    const newVars = preset.variables.map((v, i) => ({
      id: `env-preset-${preset.id}-${Date.now()}-${i}`,
      key: v.key,
      value: v.value,
      isSecret: v.isSecret,
    }));

    setEnvVars((prev) => {
      const existingKeys = new Set(prev.map((v) => v.key));
      const toAdd = newVars.filter((v) => !existingKeys.has(v.key));
      const updated = [...prev];
      for (const v of newVars) {
        if (existingKeys.has(v.key)) {
          const idx = updated.findIndex((u) => u.key === v.key);
          if (idx !== -1) updated[idx] = { ...updated[idx], value: v.value, isSecret: v.isSecret };
        } else {
          updated.push(v);
        }
      }
      return updated;
    });
    markDirty();
    toast.success(
        t('funcEditor.toastPresetAdded', { name: preset.name, count: newVars.length })
      );
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = (ev) => {
      const content = ev.target?.result as string;
      const parsed = parseEnvFile(content);
      if (parsed.length === 0) {
        toast.error(t('funcEditor.toastNoValidEnvPairs'));
        return;
      }
      const newVars = parsed.map((p) => ({
        id: `env-import-${Date.now()}-${Math.random()}`,
        key: p.key,
        value: p.value,
        isSecret: false,
      }));
      setEnvVars((prev) => {
        const existingKeys = new Set(prev.map((v) => v.key));
        const toAdd = newVars.filter((v) => !existingKeys.has(v.key));
        const updated = [...prev];
        for (const v of newVars) {
          if (existingKeys.has(v.key)) {
            const idx = updated.findIndex((u) => u.key === v.key);
            if (idx !== -1) updated[idx] = { ...updated[idx], value: v.value };
          } else {
            updated.push(v);
          }
        }
        return updated;
      });
      markDirty();
      toast.success(
          t('funcEditor.toastImportedVars', { count: parsed.length })
        );
    };
    reader.readAsText(file);
    // Reset input so same file can be re-imported
    e.target.value = '';
  };

  return (
    <SectionCard
      icon={<Key className="w-4 h-4" />}
      title={t('funcEditor.envVars')}
      step={4}
      description={t('funcEditor.envVarsDescription')}
    >
      {envVars.length > 0 && (
        <div className="space-y-2">
          {envVars.map((envVar) => (
            <div
              key={envVar.id}
              className="flex items-center gap-3 p-3 rounded-lg border border-border-subtle/30"
               style={{ background: 'var(--bg-tertiary)' }}
            >
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <code className="text-sm font-medium text-text-primary">{envVar.key}</code>
                  {envVar.isSecret && <EyeOff className="w-3 h-3 text-text-muted shrink-0" />}
                </div>
                <div className="text-sm text-text-secondary mt-0.5 truncate font-mono">
                  {envVar.isSecret
                    ? showEnvValues[envVar.id]
                      ? envVar.value
                      : '••••••••'
                    : envVar.value}
                </div>
              </div>
              <div className="flex items-center gap-1 shrink-0">
                {envVar.isSecret && (
                  <Button
                    variant="ghost"
                    size="sm"
                    className="text-text-muted hover:text-text-primary h-7 w-7 p-0"
                    onClick={() => setShowEnvValues((p) => ({ ...p, [envVar.id]: !p[envVar.id] }))}
                    aria-label={showEnvValues[envVar.id] ? t('funcEditor.hideValue') : t('funcEditor.showValue')}
                  >
                    {showEnvValues[envVar.id] ? (
                      <EyeOff className="w-3.5 h-3.5" />
                    ) : (
                      <Eye className="w-3.5 h-3.5" />
                    )}
                  </Button>
                )}
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => removeEnvironmentVariable(envVar.id)}
                  className="text-text-muted hover:text-red-400 h-7 w-7"
                  aria-label={`Remove ${envVar.key}`}
                >
                  <X className="w-3.5 h-3.5" />
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}

      <Separator className="opacity-30" />

      <div className="space-y-3">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
          <div>
            <Label
              htmlFor="env-key"
              className="text-xs font-medium text-text-secondary mb-1.5 block"
            >
              {t('funcEditor.key')}
            </Label>
            <Input
              id="env-key"
              placeholder="MY_VARIABLE"
              value={newEnvKey}
              onChange={(e) => setNewEnvKey(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') addEnvironmentVariable();
              }}
              className="input font-mono text-sm"
              autoComplete="off"
            />
          </div>
          <div>
            <Label
              htmlFor="env-value"
              className="text-xs font-medium text-text-secondary mb-1.5 block"
            >
              {t('funcEditor.value')}
            </Label>
            <Input
              id="env-value"
              placeholder="value"
              value={newEnvValue}
              onChange={(e) => setNewEnvValue(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') addEnvironmentVariable();
              }}
              className="input font-mono text-sm"
              autoComplete="off"
            />
          </div>
        </div>
        <div className="flex items-center justify-between gap-2 flex-wrap">
          <div className="flex items-center gap-2">
            <Checkbox
              id="env-secret"
              checked={isNewEnvSecret}
              onCheckedChange={(v) => setIsNewEnvSecret(!!v)}
            />
            <Label
              htmlFor="env-secret"
              className="text-xs text-text-secondary cursor-pointer flex items-center gap-1"
            >
              <Shield className="w-3 h-3" />
              {t('funcEditor.markAsSecret')}
            </Label>
          </div>
          <div className="flex items-center gap-2 flex-wrap">
            <EnvPresetsPicker onSelect={handleApplyPreset}>
              <Button
                variant="outline"
                size="sm"
                className="text-xs gap-1.5"
                type="button"
              >
                <Layers className="w-3 h-3" />
                {t('funcEditor.presets')}
              </Button>
            </EnvPresetsPicker>
            <Button
              variant="outline"
              size="sm"
              className="text-xs gap-1.5"
              onClick={() => setVaultPickerOpen(true)}
              type="button"
            >
              <Key className="w-3 h-3" />
              {t('funcEditor.fromVault')}
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="text-xs gap-1.5 hidden sm:flex"
              onClick={handleImportEnv}
              type="button"
            >
              <Upload className="w-3 h-3" />
              {t('funcEditor.importEnv')}
            </Button>
            <input
              ref={fileInputRef}
              type="file"
              accept=".env,text/plain"
              className="hidden"
              onChange={handleFileChange}
            />
            <Button
              size="sm"
              className="text-xs gap-1.5"
              onClick={addEnvironmentVariable}
              disabled={!newEnvKey.trim()}
              type="button"
            >
              <Plus className="w-3 h-3" />
              {t('funcEditor.add')}
            </Button>
          </div>
        </div>
      </div>
    </SectionCard>
  );
}
