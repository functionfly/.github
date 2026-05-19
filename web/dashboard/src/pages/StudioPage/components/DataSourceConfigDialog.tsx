import { useState } from "react";
import { Button, Input, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@functionfly/ui-core";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Database, HardDrive, Globe, Link2, Key, Plus, Trash2, Loader2, Check, X, RefreshCw, Server, Cloud, Settings } from "lucide-react";
import { cn } from "@/lib/utils";

export interface DataSource {
  id: string;
  name: string;
  type: 'postgresql' | 'mysql' | 'mongodb' | 'redis' | 's3' | 'http' | 'webhook' | 'api';
  config: Record<string, string>;
  enabled: boolean;
  createdAt?: string;
}

interface DataSourceConfigDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: (sources: DataSource[]) => void;
  sources: DataSource[];
}

interface DataSourceConfigSectionProps {
  source: Partial<DataSource>;
  updateConfig: (key: string, value: string) => void;
}

export function DataSourceConfigDialog({ open, onOpenChange, onSave, sources }: DataSourceConfigDialogProps) {
  const [localSources, setLocalSources] = useState<DataSource[]>(sources);
  const [isSaving, setIsSaving] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);

  const handleAddSource = () => {
    const newId = `ds-${Date.now()}`;
    setLocalSources(prev => [...prev, {
      id: newId,
      name: `New Source ${prev.length + 1}`,
      type: 'postgresql',
      config: {},
      enabled: true
    }]);
    setEditingId(newId);
  };

  const handleRemoveSource = (id: string) => {
    setLocalSources(prev => prev.filter(s => s.id !== id));
    if (editingId === id) setEditingId(null);
  };

  const handleUpdateSource = (id: string, updates: Partial<DataSource>) => {
    setLocalSources(prev => prev.map(s => s.id === id ? { ...s, ...updates } : s));
  };

  const handleSave = async () => {
    setIsSaving(true);
    try {
      onSave(localSources);
      onOpenChange(false);
    } finally {
      setIsSaving(false);
    }
  };

  const getSourceIcon = (type: string) => {
    switch (type) {
      case 'postgresql':
      case 'mysql':
      case 'mongodb':
        return <Database className="w-5 h-5" />;
      case 'redis':
        return <Server className="w-5 h-5" />;
      case 's3':
        return <HardDrive className="w-5 h-5" />;
      case 'http':
      case 'api':
        return <Globe className="w-5 h-5" />;
      case 'webhook':
        return <Link2 className="w-5 h-5" />;
      default:
        return <Settings className="w-5 h-5" />;
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="bg-[#1a1a2e] border-white/10 max-w-3xl max-h-[90vh] overflow-hidden flex flex-col">
        <DialogHeader>
          <DialogTitle className="text-white flex items-center gap-2">
            <span className="w-8 h-8 rounded-lg bg-white/10 flex items-center justify-center">
              <Database className="w-5 h-5" />
            </span>
            Configure Data Sources
          </DialogTitle>
          <DialogDescription className="text-white/60">
            Connect external data sources to your workflow
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 overflow-y-auto space-y-4 py-4">
          {localSources.length === 0 && (
            <div className="text-center py-12 text-white/40">
              <Database className="w-12 h-12 mx-auto mb-4 opacity-50" />
              <p>No data sources configured</p>
              <p className="text-sm mt-1">Add a data source to get started</p>
            </div>
          )}

          {localSources.map(source => (
            <div key={source.id} className="rounded-lg border border-white/10 bg-white/5 p-4">
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-3">
                  <span className="w-8 h-8 rounded-lg bg-white/10 flex items-center justify-center">
                    {getSourceIcon(source.type)}
                  </span>
                  <div>
                    <input
                      type="text"
                      value={source.name}
                      onChange={(e) => handleUpdateSource(source.id, { name: e.target.value })}
                      className="bg-transparent text-white font-medium border-none outline-none"
                    />
                    <p className="text-xs text-white/40 capitalize">{source.type}</p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => handleUpdateSource(source.id, { enabled: !source.enabled })}
                    className={cn(
                      "relative w-11 h-6 rounded-full transition-colors",
                      source.enabled ? "bg-green-500" : "bg-white/20"
                    )}
                  >
                    <span className={cn("absolute top-1 left-1 w-4 h-4 rounded-full bg-white transition-transform", source.enabled && "translate-x-5")} />
                  </button>
                  <button
                    onClick={() => handleRemoveSource(source.id)}
                    className="p-1.5 rounded-md text-white/40 hover:text-red-400 hover:bg-white/5"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </div>

              {editingId === source.id ? (
                <DataSourceFields
                  source={source}
                  onUpdate={(key, value) => {
                    const newConfig = { ...source.config, [key]: value };
                    handleUpdateSource(source.id, { config: newConfig });
                  }}
                />
              ) : (
                <button
                  onClick={() => setEditingId(source.id)}
                  className="text-xs text-brand-400 hover:text-brand-300 flex items-center gap-1"
                >
                  <Settings className="w-3 h-3" />
                  Configure
                </button>
              )}
            </div>
          ))}

          <button
            onClick={handleAddSource}
            className="w-full p-4 rounded-lg border border-dashed border-white/20 text-white/60 hover:text-white hover:border-white/40 transition-colors flex items-center justify-center gap-2"
          >
            <Plus className="w-4 h-4" />
            Add Data Source
          </button>
        </div>

        <DialogFooter className="border-t border-white/10 pt-4">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={isSaving}>
            {isSaving ? <Loader2 className="w-4 h-4 animate-spin" /> : null}
            Save Configuration
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function DataSourceFields({ source, onUpdate }: { source: DataSource; onUpdate: (key: string, value: string) => void }) {
  const renderFields = () => {
    switch (source.type) {
      case 'postgresql':
      case 'mysql':
        return (
          <>
            <div className="space-y-2">
              <label className="text-xs text-white/60">Connection String</label>
              <Input
                value={source.config.connection_string || ''}
                onChange={(e) => onUpdate('connection_string', e.target.value)}
                placeholder="postgresql://user:pass@host:5432/db"
                className="bg-white/5 border-white/10 text-white font-mono text-sm"
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2">
                <label className="text-xs text-white/60">Host</label>
                <Input
                  value={source.config.host || ''}
                  onChange={(e) => onUpdate('host', e.target.value)}
                  placeholder="localhost"
                  className="bg-white/5 border-white/10 text-white"
                />
              </div>
              <div className="space-y-2">
                <label className="text-xs text-white/60">Port</label>
                <Input
                  value={source.config.port || ''}
                  onChange={(e) => onUpdate('port', e.target.value)}
                  placeholder="5432"
                  className="bg-white/5 border-white/10 text-white"
                />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2">
                <label className="text-xs text-white/60">Database</label>
                <Input
                  value={source.config.database || ''}
                  onChange={(e) => onUpdate('database', e.target.value)}
                  placeholder="mydb"
                  className="bg-white/5 border-white/10 text-white"
                />
              </div>
              <div className="space-y-2">
                <label className="text-xs text-white/60">Schema</label>
                <Input
                  value={source.config.schema || 'public'}
                  onChange={(e) => onUpdate('schema', e.target.value)}
                  placeholder="public"
                  className="bg-white/5 border-white/10 text-white"
                />
              </div>
            </div>
          </>
        );

      case 'mongodb':
        return (
          <>
            <div className="space-y-2">
              <label className="text-xs text-white/60">Connection URI</label>
              <Input
                value={source.config.connection_uri || ''}
                onChange={(e) => onUpdate('connection_uri', e.target.value)}
                placeholder="mongodb://user:pass@host:27017/db"
                className="bg-white/5 border-white/10 text-white font-mono text-sm"
              />
            </div>
            <div className="space-y-2">
              <label className="text-xs text-white/60">Database Name</label>
              <Input
                value={source.config.database || ''}
                onChange={(e) => onUpdate('database', e.target.value)}
                placeholder="mydb"
                className="bg-white/5 border-white/10 text-white"
              />
            </div>
          </>
        );

      case 'redis':
        return (
          <>
            <div className="space-y-2">
              <label className="text-xs text-white/60">Connection URL</label>
              <Input
                value={source.config.connection_url || ''}
                onChange={(e) => onUpdate('connection_url', e.target.value)}
                placeholder="redis://localhost:6379"
                className="bg-white/5 border-white/10 text-white font-mono text-sm"
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2">
                <label className="text-xs text-white/60">Password</label>
                <Input
                  type="password"
                  value={source.config.password || ''}
                  onChange={(e) => onUpdate('password', e.target.value)}
                  placeholder="Optional"
                  className="bg-white/5 border-white/10 text-white"
                />
              </div>
              <div className="space-y-2">
                <label className="text-xs text-white/60">DB Number</label>
                <Input
                  value={source.config.db || '0'}
                  onChange={(e) => onUpdate('db', e.target.value)}
                  placeholder="0"
                  className="bg-white/5 border-white/10 text-white"
                />
              </div>
            </div>
          </>
        );

      case 's3':
        return (
          <>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2">
                <label className="text-xs text-white/60">Access Key ID</label>
                <Input
                  value={source.config.access_key_id || ''}
                  onChange={(e) => onUpdate('access_key_id', e.target.value)}
                  placeholder="AKIA..."
                  className="bg-white/5 border-white/10 text-white font-mono"
                />
              </div>
              <div className="space-y-2">
                <label className="text-xs text-white/60">Secret Access Key</label>
                <Input
                  type="password"
                  value={source.config.secret_access_key || ''}
                  onChange={(e) => onUpdate('secret_access_key', e.target.value)}
                  placeholder="xxxxxxxxxxxx"
                  className="bg-white/5 border-white/10 text-white font-mono"
                />
              </div>
            </div>
            <div className="space-y-2">
              <label className="text-xs text-white/60">Bucket Name</label>
              <Input
                value={source.config.bucket || ''}
                onChange={(e) => onUpdate('bucket', e.target.value)}
                placeholder="my-bucket"
                className="bg-white/5 border-white/10 text-white"
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2">
                <label className="text-xs text-white/60">Region</label>
                <Select value={source.config.region || 'us-east-1'} onValueChange={(v) => onUpdate('region', v)}>
                  <SelectTrigger className="bg-white/5 border-white/10 text-white">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="bg-[#1a1a2e] border-white/10">
                    {['us-east-1', 'us-west-2', 'eu-west-1', 'eu-central-1', 'ap-southeast-1', 'ap-northeast-1'].map(r => (
                      <SelectItem key={r} value={r}>{r}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <label className="text-xs text-white/60">Prefix</label>
                <Input
                  value={source.config.prefix || ''}
                  onChange={(e) => onUpdate('prefix', e.target.value)}
                  placeholder="path/to/data/"
                  className="bg-white/5 border-white/10 text-white"
                />
              </div>
            </div>
          </>
        );

      case 'http':
      case 'api':
        return (
          <>
            <div className="space-y-2">
              <label className="text-xs text-white/60">Base URL</label>
              <Input
                value={source.config.base_url || ''}
                onChange={(e) => onUpdate('base_url', e.target.value)}
                placeholder="https://api.example.com"
                className="bg-white/5 border-white/10 text-white font-mono"
              />
            </div>
            <div className="space-y-2">
              <label className="text-xs text-white/60">Authentication</label>
              <Select value={source.config.auth_type || 'none'} onValueChange={(v) => onUpdate('auth_type', v)}>
                <SelectTrigger className="bg-white/5 border-white/10 text-white">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="bg-[#1a1a2e] border-white/10">
                  <SelectItem value="none">None</SelectItem>
                  <SelectItem value="api_key">API Key</SelectItem>
                  <SelectItem value="bearer">Bearer Token</SelectItem>
                  <SelectItem value="basic">Basic Auth</SelectItem>
                </SelectContent>
              </Select>
            </div>
            {source.config.auth_type === 'bearer' && (
              <div className="space-y-2">
                <label className="text-xs text-white/60">Bearer Token</label>
                <Input
                  type="password"
                  value={source.config.bearer_token || ''}
                  onChange={(e) => onUpdate('bearer_token', e.target.value)}
                  placeholder="xxxxx"
                  className="bg-white/5 border-white/10 text-white font-mono"
                />
              </div>
            )}
            {source.config.auth_type === 'basic' && (
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-2">
                  <label className="text-xs text-white/60">Username</label>
                  <Input
                    value={source.config.username || ''}
                    onChange={(e) => onUpdate('username', e.target.value)}
                    className="bg-white/5 border-white/10 text-white"
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-xs text-white/60">Password</label>
                  <Input
                    type="password"
                    value={source.config.password || ''}
                    onChange={(e) => onUpdate('password', e.target.value)}
                    className="bg-white/5 border-white/10 text-white"
                  />
                </div>
              </div>
            )}
            {source.config.auth_type === 'api_key' && (
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-2">
                  <label className="text-xs text-white/60">Header Name</label>
                  <Input
                    value={source.config.header_name || 'X-API-Key'}
                    onChange={(e) => onUpdate('header_name', e.target.value)}
                    className="bg-white/5 border-white/10 text-white"
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-xs text-white/60">API Key</label>
                  <Input
                    type="password"
                    value={source.config.api_key || ''}
                    onChange={(e) => onUpdate('api_key', e.target.value)}
                    className="bg-white/5 border-white/10 text-white font-mono"
                  />
                </div>
              </div>
            )}
          </>
        );

      case 'webhook':
        return (
          <>
            <div className="space-y-2">
              <label className="text-xs text-white/60">Webhook URL</label>
              <Input
                value={source.config.webhook_url || ''}
                onChange={(e) => onUpdate('webhook_url', e.target.value)}
                placeholder="https://example.com/webhook"
                className="bg-white/5 border-white/10 text-white font-mono"
              />
            </div>
            <div className="space-y-2">
              <label className="text-xs text-white/60">HTTP Method</label>
              <Select value={source.config.method || 'POST'} onValueChange={(v) => onUpdate('method', v)}>
                <SelectTrigger className="bg-white/5 border-white/10 text-white">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="bg-[#1a1a2e] border-white/10">
                  <SelectItem value="GET">GET</SelectItem>
                  <SelectItem value="POST">POST</SelectItem>
                  <SelectItem value="PUT">PUT</SelectItem>
                  <SelectItem value="PATCH">PATCH</SelectItem>
                  <SelectItem value="DELETE">DELETE</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </>
        );

      default:
        return (
          <div className="space-y-2">
            <label className="text-xs text-white/60">Configuration</label>
            <textarea
              value={JSON.stringify(source.config, null, 2)}
              onChange={(e) => {
                try {
                  const parsed = JSON.parse(e.target.value);
                  Object.entries(parsed).forEach(([key, value]) => {
                    onUpdate(key, String(value));
                  });
                } catch {}
              }}
              className="w-full h-32 bg-white/5 border border-white/10 rounded-lg p-3 text-xs font-mono text-white/80 resize-none"
              placeholder="{}"
            />
          </div>
        );
    }
  };

  return (
    <div className="space-y-3 pt-2 border-t border-white/10">
      <div className="space-y-2">
        <label className="text-xs text-white/60">Source Type</label>
        <Select value={source.type} onValueChange={(v) => onUpdate('type', v)}>
          <SelectTrigger className="bg-white/5 border-white/10 text-white">
            <SelectValue />
          </SelectTrigger>
          <SelectContent className="bg-[#1a1a2e] border-white/10">
            <SelectItem value="postgresql">PostgreSQL</SelectItem>
            <SelectItem value="mysql">MySQL</SelectItem>
            <SelectItem value="mongodb">MongoDB</SelectItem>
            <SelectItem value="redis">Redis</SelectItem>
            <SelectItem value="s3">AWS S3</SelectItem>
            <SelectItem value="http">HTTP API</SelectItem>
            <SelectItem value="webhook">Webhook</SelectItem>
          </SelectContent>
        </Select>
      </div>
      {renderFields()}
    </div>
  );
}