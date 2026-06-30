/**
 * MCP Center - Settings Tab
 * Global MCP configuration defaults
 */

import { useState } from 'react';
import { Settings, Save, ExternalLink, Loader2 } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Textarea } from '@/components/ui/textarea';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { useMCPSettings } from '../hooks';
import { DEFAULT_MCP_SETTINGS } from '../constants';

export function SettingsTab() {
  const { settings, isLoading, isUpdating, updateSettings } = useMCPSettings();
  const effectiveSettings = settings ?? DEFAULT_MCP_SETTINGS;
  const [formState, setFormState] = useState({
    default_transport: effectiveSettings.default_transport,
    default_rate_limit: effectiveSettings.default_rate_limit,
    default_expose_input: effectiveSettings.default_expose_input,
    default_expose_output: effectiveSettings.default_expose_output,
    auto_add_to_registry: effectiveSettings.auto_add_to_registry,
    require_verification: effectiveSettings.require_verification,
    public_listing: effectiveSettings.public_listing,
    cors_allowlist: effectiveSettings.cors_allowlist.join('\n'),
    rate_limit_multiplier: effectiveSettings.rate_limit_multiplier,
  });
  const [saved, setSaved] = useState(false);

  const handleSave = () => {
    updateSettings({
      ...formState,
      cors_allowlist: formState.cors_allowlist
        .split('\n')
        .map((s) => s.trim())
        .filter(Boolean),
    });
    setSaved(true);
    setTimeout(() => setSaved(false), 3000);
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-2">
          <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
          <span className="text-text-secondary">Loading MCP settings...</span>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold text-text-primary flex items-center gap-2">
            <Settings className="h-5 w-5 text-[var(--status-ok)]" />
            MCP Settings
          </h2>
          <p className="text-sm text-text-secondary mt-1">
            Configure global defaults for Model Context Protocol
          </p>
        </div>
        <Button onClick={handleSave} disabled={isUpdating}>
          {isUpdating ? (
            <Loader2 className="h-4 w-4 animate-spin mr-2" />
          ) : (
            <Save className="h-4 w-4 mr-2" />
          )}
          Save Changes
        </Button>
      </div>

      {/* Success Message */}
      {saved && (
        <Alert>
          <AlertDescription>MCP settings saved successfully!</AlertDescription>
        </Alert>
      )}

      {/* Default Settings */}
      <Card className="mcp-panel">
        <CardHeader>
          <CardTitle className="mcp-table-title">Default Settings</CardTitle>
          <CardDescription>
            These defaults apply to new functions when MCP is enabled
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-2">
            {/* Default Transport */}
            <div className="space-y-2">
              <Label htmlFor="default-transport">Default Transport</Label>
              <Select
                value={formState.default_transport}
                onValueChange={(v) =>
                  setFormState((s) => ({
                    ...s,
                    default_transport: v as typeof formState.default_transport,
                  }))
                }
              >
                <SelectTrigger id="default-transport">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="streamable-http">Streamable HTTP</SelectItem>
                  <SelectItem value="stdio">STDIO</SelectItem>
                  <SelectItem value="both">Both</SelectItem>
                </SelectContent>
              </Select>
              <p className="text-xs text-muted-foreground">
                Recommended: Streamable HTTP for most use cases
              </p>
            </div>

            {/* Default Rate Limit */}
            <div className="space-y-2">
              <Label htmlFor="default-rate">Default Rate Limit (calls/min)</Label>
              <Input
                id="default-rate"
                type="number"
                min={1}
                max={10000}
                value={formState.default_rate_limit}
                onChange={(e) =>
                  setFormState((s) => ({
                    ...s,
                    default_rate_limit: parseInt(e.target.value) || 60,
                  }))
                }
              />
              <p className="text-xs text-muted-foreground">Per-caller, per-function. Default: 60</p>
            </div>
          </div>

          {/* Default Schema Exposure */}
          <div className="space-y-2">
            <Label>Default Schema Exposure</Label>
            <div className="flex gap-6">
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={formState.default_expose_input}
                  onChange={(e) =>
                    setFormState((s) => ({
                      ...s,
                      default_expose_input: e.target.checked,
                    }))
                  }
                />
                <span className="text-sm">Expose input schema to MCP clients</span>
              </label>
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={formState.default_expose_output}
                  onChange={(e) =>
                    setFormState((s) => ({
                      ...s,
                      default_expose_output: e.target.checked,
                    }))
                  }
                />
                <span className="text-sm">Expose output schema to MCP clients</span>
              </label>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Registry Settings */}
      <Card className="mcp-panel">
        <CardHeader>
          <CardTitle className="mcp-table-title">Registry Settings</CardTitle>
          <CardDescription>Configure MCP registry behavior</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-3">
            <label className="flex items-start justify-between gap-4">
              <div>
                <span className="text-sm font-medium">Auto-add new functions</span>
                <p className="text-xs text-muted-foreground mt-1">
                  Automatically enable MCP for newly created functions
                </p>
              </div>
              <Switch
                checked={formState.auto_add_to_registry}
                onCheckedChange={(v) => setFormState((s) => ({ ...s, auto_add_to_registry: v }))}
              />
            </label>

            <label className="flex items-start justify-between gap-4">
              <div>
                <span className="text-sm font-medium">Require verification</span>
                <p className="text-xs text-muted-foreground mt-1">
                  New MCP functions must be verified before becoming publicly discoverable
                </p>
              </div>
              <Switch
                checked={formState.require_verification}
                onCheckedChange={(v) => setFormState((s) => ({ ...s, require_verification: v }))}
              />
            </label>

            <label className="flex items-start justify-between gap-4">
              <div>
                <span className="text-sm font-medium">Public MCP listing</span>
                <p className="text-xs text-muted-foreground mt-1">
                  Include your functions in the public MCP registry for AI clients to discover
                </p>
              </div>
              <Switch
                checked={formState.public_listing}
                onCheckedChange={(v) => setFormState((s) => ({ ...s, public_listing: v }))}
              />
            </label>
          </div>
        </CardContent>
      </Card>

      {/* Security Settings */}
      <Card className="mcp-panel">
        <CardHeader>
          <CardTitle className="mcp-table-title">Security Settings</CardTitle>
          <CardDescription>Configure MCP security options</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Rate Limit Multiplier */}
          <div className="space-y-2">
            <Label htmlFor="rate-multiplier">Rate Limit Multiplier</Label>
            <Select
              value={String(formState.rate_limit_multiplier)}
              onValueChange={(v) =>
                setFormState((s) => ({
                  ...s,
                  rate_limit_multiplier: parseInt(v),
                }))
              }
            >
              <SelectTrigger id="rate-multiplier" className="w-[200px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="1">1x (Default)</SelectItem>
                <SelectItem value="2">2x</SelectItem>
                <SelectItem value="5">5x</SelectItem>
                <SelectItem value="10">10x</SelectItem>
              </SelectContent>
            </Select>
            <p className="text-xs text-muted-foreground">
              Multiply all MCP rate limits by this factor
            </p>
          </div>

          {/* CORS Allowlist */}
          <div className="space-y-2">
            <Label htmlFor="cors-allowlist">CORS Allowlist (one origin per line)</Label>
            <Textarea
              id="cors-allowlist"
              value={formState.cors_allowlist}
              onChange={(e) => setFormState((s) => ({ ...s, cors_allowlist: e.target.value }))}
              placeholder={'https://my-frontend.example.com\nhttps://another.example.com'}
              rows={4}
            />
            <p className="text-xs text-muted-foreground">
              Leave empty to allow all origins. Restricted origins will be blocked.
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Documentation Links */}
      <Card className="mcp-panel">
        <CardHeader>
          <CardTitle className="mcp-table-title">Documentation</CardTitle>
          <CardDescription>Learn more about MCP configuration</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-3">
            <Button variant="outline" asChild className="mcp-export-btn">
              <a
                href="https://modelcontextprotocol.io/docs"
                target="_blank"
                rel="noopener noreferrer"
              >
                MCP Documentation
                <ExternalLink className="h-4 w-4 ml-2" />
              </a>
            </Button>
            <Button variant="outline" asChild className="mcp-export-btn">
              <a href="https://docs.functionfly.com/mcp" target="_blank" rel="noopener noreferrer">
                FunctionFly MCP Guide
                <ExternalLink className="h-4 w-4 ml-2" />
              </a>
            </Button>
            <Button variant="outline" asChild className="mcp-export-btn">
              <a
                href="https://docs.functionfly.com/integrations"
                target="_blank"
                rel="noopener noreferrer"
              >
                Integration Guides
                <ExternalLink className="h-4 w-4 ml-2" />
              </a>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
