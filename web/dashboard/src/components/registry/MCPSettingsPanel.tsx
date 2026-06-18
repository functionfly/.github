/**
 * MCPSettingsPanel — the "MCP" tab in the function settings page.
 *
 * Lets the owner flip a function on/off the MCP Function Registry and
 * tune per-function settings (transports, rate limit, tool name override,
 * origin allowlist, expose input/output schema).
 *
 * Backed by:
 *   GET  /v1/functions/{author}/{name}/mcp
 *   PATCH /v1/functions/{author}/{name}/mcp
 *   POST /v1/functions/publish  (with `mcp` block; legacy path)
 */
import React, { useCallback, useEffect, useState } from "react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import { Loader2, Save, ShieldCheck } from "lucide-react";

export interface MCPSettings {
  function_id?: string;
  enabled: boolean;
  transports: string[];
  expose_input_schema: boolean;
  expose_output_schema: boolean;
  tool_name_override: string;
  rate_limit_per_min: number;
  allowlist_origins: string[];
  verified_mcp?: boolean;
  invocation_count?: number;
  last_invoked_at?: string | null;
}

interface MCPSettingsPanelProps {
  author: string;
  name: string;
  apiBase?: string;
  onSaved?: (s: MCPSettings) => void;
}

const DEFAULTS: MCPSettings = {
  enabled: false,
  transports: ["streamable-http"],
  expose_input_schema: true,
  expose_output_schema: false,
  tool_name_override: "",
  rate_limit_per_min: 60,
  allowlist_origins: [],
};

export function MCPSettingsPanel({ author, name, apiBase, onSaved }: MCPSettingsPanelProps) {
  const [settings, setSettings] = useState<MCPSettings>(DEFAULTS);
  const [origins, setOrigins] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [info, setInfo] = useState<string | null>(null);

  const base = apiBase ?? (import.meta.env.VITE_API_URL || "");

  const fetchSettings = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const resp = await fetch(`${base}/v1/functions/${author}/${name}/mcp`, {
        credentials: "include",
        headers: { accept: "application/json" },
      });
      if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
      const data = (await resp.json()) as MCPSettings;
      setSettings({ ...DEFAULTS, ...data });
      setOrigins((data.allowlist_origins ?? []).join("\n"));
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load MCP settings");
    } finally {
      setLoading(false);
    }
  }, [author, name, base]);

  useEffect(() => {
    void fetchSettings();
  }, [fetchSettings]);

  const save = useCallback(async () => {
    setSaving(true);
    setError(null);
    setInfo(null);
    try {
      const body: Partial<MCPSettings> = {
        enabled: settings.enabled,
        transports: settings.transports,
        expose_input_schema: settings.expose_input_schema,
        expose_output_schema: settings.expose_output_schema,
        tool_name_override: settings.tool_name_override,
        rate_limit_per_min: settings.rate_limit_per_min,
        allowlist_origins: origins
          .split("\n")
          .map((s) => s.trim())
          .filter(Boolean),
      };
      const resp = await fetch(`${base}/v1/functions/${author}/${name}/mcp`, {
        method: "PATCH",
        credentials: "include",
        headers: { "content-type": "application/json", accept: "application/json" },
        body: JSON.stringify(body),
      });
      if (!resp.ok) {
        const txt = await resp.text();
        throw new Error(`HTTP ${resp.status}: ${txt}`);
      }
      const data = (await resp.json()) as MCPSettings;
      setSettings({ ...DEFAULTS, ...data });
      setInfo("Saved. Changes are live on the registry within ~30s.");
      onSaved?.(data);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to save");
    } finally {
      setSaving(false);
    }
  }, [author, name, base, settings, origins, onSaved]);

  if (loading) {
    return (
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <Loader2 className="h-4 w-4 animate-spin" /> Loading MCP settings…
      </div>
    );
  }

  const toolNamePreview = settings.tool_name_override || `${author}__${name}`;

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              MCP Function Registry
              {settings.verified_mcp && (
                <Badge variant="default" className="bg-emerald-600">
                  <ShieldCheck className="h-3 w-3 mr-1" /> Verified
                </Badge>
              )}
            </CardTitle>
            <CardDescription>
              Make this function callable from Claude Desktop, Cursor, VS Code, and every MCP-compatible AI agent.
            </CardDescription>
          </div>
          <div className="flex items-center gap-2">
            <Label htmlFor="mcp-enabled" className="text-sm">
              Enabled
            </Label>
            <Switch
              id="mcp-enabled"
              checked={settings.enabled}
              onCheckedChange={(v) => setSettings((s) => ({ ...s, enabled: v }))}
            />
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        {info && (
          <Alert>
            <AlertDescription>{info}</AlertDescription>
          </Alert>
        )}

        <div className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-2">
            <Label htmlFor="mcp-tool-name">Tool name override</Label>
            <Input
              id="mcp-tool-name"
              value={settings.tool_name_override}
              onChange={(e) => setSettings((s) => ({ ...s, tool_name_override: e.target.value }))}
              placeholder={`${author}__${name}`}
              maxLength={64}
              pattern="[a-zA-Z0-9_-]+"
            />
            <p className="text-xs text-muted-foreground">
              Defaults to <code className="font-mono">{`${author}__${name}`}</code>.{" "}
              1-64 chars, <code className="font-mono">[a-zA-Z0-9_-]</code> only.
            </p>
            <p className="text-xs text-muted-foreground">
              Effective name:{" "}
              <code className="font-mono rounded bg-muted px-1">{toolNamePreview}</code>
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="mcp-rate">Rate limit (calls/min)</Label>
            <Input
              id="mcp-rate"
              type="number"
              min={1}
              max={10000}
              value={settings.rate_limit_per_min}
              onChange={(e) =>
                setSettings((s) => ({ ...s, rate_limit_per_min: Math.max(1, Number(e.target.value || 1)) }))
              }
            />
            <p className="text-xs text-muted-foreground">Per-caller, per-function. Default 60.</p>
          </div>
        </div>

        <div className="space-y-2">
          <Label>Transports</Label>
          <div className="flex flex-wrap gap-3">
            {(["streamable-http", "stdio"] as const).map((t) => {
              const checked = settings.transports.includes(t);
              return (
                <label key={t} className="flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    checked={checked}
                    onChange={() => {
                      setSettings((s) => ({
                        ...s,
                        transports: checked ? s.transports.filter((x) => x !== t) : [...s.transports, t],
                      }));
                    }}
                  />
                  <code className="font-mono text-xs">{t}</code>
                </label>
              );
            })}
          </div>
        </div>

        <div className="space-y-2">
          <Label>Schema exposure</Label>
          <div className="flex flex-col gap-2 text-sm">
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={settings.expose_input_schema}
                onChange={(e) => setSettings((s) => ({ ...s, expose_input_schema: e.target.checked }))}
              />
              Expose input schema to MCP clients
            </label>
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={settings.expose_output_schema}
                onChange={(e) => setSettings((s) => ({ ...s, expose_output_schema: e.target.checked }))}
              />
              Expose output schema to MCP clients
            </label>
          </div>
        </div>

        <div className="space-y-2">
          <Label htmlFor="mcp-origins">CORS allowlist (one origin per line)</Label>
          <Textarea
            id="mcp-origins"
            value={origins}
            onChange={(e) => setOrigins(e.target.value)}
            placeholder={"https://my-frontend.example\nhttps://other.example"}
            rows={4}
          />
          <p className="text-xs text-muted-foreground">
            Reserved for future per-origin CORS gating. Leave empty to allow all origins.
          </p>
        </div>

        {typeof settings.invocation_count === "number" && (
          <div className="rounded-md bg-muted/50 p-3 text-sm">
            <div className="font-medium">MCP usage</div>
            <div className="text-muted-foreground">
              {settings.invocation_count.toLocaleString()} invocations
              {settings.last_invoked_at && (
                <> · last called {new Date(settings.last_invoked_at).toLocaleString()}</>
              )}
            </div>
          </div>
        )}

        <div className="flex justify-end">
          <Button onClick={save} disabled={saving}>
            {saving ? <Loader2 className="h-4 w-4 mr-2 animate-spin" /> : <Save className="h-4 w-4 mr-2" />}
            Save MCP settings
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

export default MCPSettingsPanel;
