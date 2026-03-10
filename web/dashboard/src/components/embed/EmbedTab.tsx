import { useState, useEffect } from "react";
import { toast } from "sonner";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { embedApi, type EmbedConfig, type EmbedAnalytics } from "@/api/embed";
import { EmbedCodeGenerator } from "./EmbedCodeGenerator";
import {
  Globe,
  Shield,
  Palette,
  BarChart3,
  Save,
  Loader2,
  RefreshCw,
  AlertCircle,
} from "lucide-react";
import "@/styles/components.css";

interface EmbedTabProps {
  author: string;
  name: string;
  version?: string;
}

const defaultConfig: EmbedConfig = {
  enabled: false,
  allowed_origins: [],
  require_api_key: false,
  ui_enabled: false,
  ui_theme: "auto",
  rate_limit_per_hour: 1000,
};

export function EmbedTab({ author, name, version }: EmbedTabProps) {
  const [config, setConfig] = useState<EmbedConfig>(defaultConfig);
  const [analytics, setAnalytics] = useState<EmbedAnalytics | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [analyticsLoading, setAnalyticsLoading] = useState(false);
  const [allowedOriginsInput, setAllowedOriginsInput] = useState("");

  // Fetch embed config and analytics
  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      try {
        const [configResponse, analyticsResponse] = await Promise.all([
          embedApi.getEmbedConfig(author, name).catch(() => defaultConfig),
          embedApi.getEmbedAnalytics(author, name).catch(() => null),
        ]);
        setConfig(configResponse);
        setAnalytics(analyticsResponse);
        setAllowedOriginsInput(configResponse.allowed_origins.join(", "));
      } catch (error) {
        console.error("Failed to fetch embed data:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [author, name]);

  const handleSave = async () => {
    setSaving(true);
    try {
      // Parse allowed origins from comma-separated string
      const allowedOrigins = allowedOriginsInput
        .split(",")
        .map((origin) => origin.trim())
        .filter((origin) => origin.length > 0);

      const updatedConfig = {
        ...config,
        allowed_origins: allowedOrigins,
      };

      await embedApi.updateEmbedConfig(author, name, updatedConfig);
      setConfig(updatedConfig);
      toast.success("Embed configuration saved successfully!");
    } catch (error) {
      console.error("Failed to save embed config:", error);
      toast.error("Failed to save configuration");
    } finally {
      setSaving(false);
    }
  };

  const handleRefreshAnalytics = async () => {
    setAnalyticsLoading(true);
    try {
      const analyticsResponse = await embedApi.getEmbedAnalytics(author, name);
      setAnalytics(analyticsResponse);
    } catch (error) {
      console.error("Failed to refresh analytics:", error);
      toast.error("Failed to refresh analytics");
    } finally {
      setAnalyticsLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <Loader2 className="w-6 h-6 animate-spin text-text-muted" />
        <span className="ml-2 text-text-muted">Loading embed configuration...</span>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Configuration Panel */}
      <Card className="card">
        <CardHeader className="card-header">
          <CardTitle className="card-title flex items-center gap-2">
            <Shield className="w-5 h-5" />
            Embed Configuration
          </CardTitle>
        </CardHeader>
        <CardContent className="card-content space-y-6">
          {/* Enable/Disable Toggle */}
          <div className="flex items-center justify-between">
            <div className="space-y-1">
              <Label className="text-text-primary font-medium">Enable Embed</Label>
              <p className="text-sm text-text-secondary">
                Allow this function to be embedded on external websites
              </p>
            </div>
            <Switch
              checked={config.enabled}
              onCheckedChange={(checked) => setConfig({ ...config, enabled: checked })}
            />
          </div>

          <div className="border-t border-border" />

          {/* Allowed Origins */}
          <div className="space-y-2">
            <Label htmlFor="allowed-origins" className="text-text-secondary flex items-center gap-2">
              <Globe className="w-4 h-4" />
              Allowed Origins
            </Label>
            <Input
              id="allowed-origins"
              value={allowedOriginsInput}
              onChange={(e) => setAllowedOriginsInput(e.target.value)}
              placeholder="example.com, *.example.com (comma-separated)"
              disabled={!config.enabled}
            />
            <p className="text-xs text-text-muted">
              Comma-separated list of domains. Leave empty to allow all origins.
              Use * for wildcard.
            </p>
          </div>

          {/* API Key Requirement */}
          <div className="flex items-center justify-between">
            <div className="space-y-1">
              <Label className="text-text-primary font-medium">Require API Key</Label>
              <p className="text-sm text-text-secondary">
                Require callers to provide an API key for authentication
              </p>
            </div>
            <Switch
              checked={config.require_api_key}
              onCheckedChange={(checked) =>
                setConfig({ ...config, require_api_key: checked })
              }
              disabled={!config.enabled}
            />
          </div>

          {/* UI Widget Toggle */}
          <div className="flex items-center justify-between">
            <div className="space-y-1">
              <Label className="text-text-primary font-medium flex items-center gap-2">
                <Palette className="w-4 h-4" />
                UI Widget
              </Label>
              <p className="text-sm text-text-secondary">
                Enable the interactive UI widget for this embed
              </p>
            </div>
            <Switch
              checked={config.ui_enabled}
              onCheckedChange={(checked) => setConfig({ ...config, ui_enabled: checked })}
              disabled={!config.enabled}
            />
          </div>

          {/* Theme Selector */}
          <div className="space-y-2">
            <Label htmlFor="ui-theme" className="text-text-secondary">
              UI Theme
            </Label>
            <Select
              value={config.ui_theme}
              onValueChange={(value) =>
                setConfig({ ...config, ui_theme: value as "light" | "dark" | "auto" })
              }
              disabled={!config.enabled || !config.ui_enabled}
            >
              <SelectTrigger id="ui-theme" className="w-full md:w-[200px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="light">Light</SelectItem>
                <SelectItem value="dark">Dark</SelectItem>
                <SelectItem value="auto">Auto (System)</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Rate Limiting */}
          <div className="space-y-2">
            <Label htmlFor="rate-limit" className="text-text-secondary">
              Rate Limit (per hour)
            </Label>
            <Input
              id="rate-limit"
              type="number"
              value={config.rate_limit_per_hour}
              onChange={(e) =>
                setConfig({
                  ...config,
                  rate_limit_per_hour: parseInt(e.target.value) || 0,
                })
              }
              disabled={!config.enabled}
              min={0}
              className="w-full md:w-[200px]"
            />
            <p className="text-xs text-text-muted">
              Maximum requests per hour per origin. Set to 0 for unlimited.
            </p>
          </div>

          {/* Save Button */}
          <div className="flex justify-end pt-4">
            <Button onClick={handleSave} disabled={saving}>
              {saving ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Saving...
                </>
              ) : (
                <>
                  <Save className="w-4 h-4 mr-2" />
                  Save Configuration
                </>
              )}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Embed Code Generator */}
      {config.enabled && (
        <EmbedCodeGenerator author={author} name={name} version={version} />
      )}

      {/* Analytics Section */}
      {config.enabled && (
        <Card className="card">
          <CardHeader className="card-header">
            <CardTitle className="card-title flex items-center justify-between">
              <div className="flex items-center gap-2">
                <BarChart3 className="w-5 h-5" />
                Embed Analytics
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={handleRefreshAnalytics}
                disabled={analyticsLoading}
              >
                <RefreshCw
                  className={`w-4 h-4 mr-2 ${analyticsLoading ? "animate-spin" : ""}`}
                />
                Refresh
              </Button>
            </CardTitle>
          </CardHeader>
          <CardContent className="card-content">
            {analytics ? (
              <div className="space-y-4">
                {/* Summary Stats */}
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                  <div className="bg-bg-secondary rounded-lg p-4">
                    <div className="text-2xl font-bold text-text-primary">
                      {analytics.total_executions.toLocaleString()}
                    </div>
                    <div className="text-sm text-text-secondary">Total Executions</div>
                  </div>
                  <div className="bg-bg-secondary rounded-lg p-4">
                    <div className="text-2xl font-bold text-text-primary">
                      {analytics.unique_origins}
                    </div>
                    <div className="text-sm text-text-secondary">Unique Origins</div>
                  </div>
                  <div className="bg-bg-secondary rounded-lg p-4">
                    <div className="text-2xl font-bold text-text-primary">
                      {analytics.period}
                    </div>
                    <div className="text-sm text-text-secondary">Period</div>
                  </div>
                </div>

                {/* Origin Breakdown */}
                {analytics.origin_stats.length > 0 ? (
                  <div className="space-y-2">
                    <Label className="text-text-secondary">Executions by Origin</Label>
                    <div className="space-y-2 max-h-[300px] overflow-y-auto">
                      {analytics.origin_stats.map((stat, index) => (
                        <div
                          key={index}
                          className="flex items-center justify-between p-3 bg-bg-secondary rounded-lg"
                        >
                          <div className="flex items-center gap-2">
                            <Globe className="w-4 h-4 text-text-muted" />
                            <span className="text-sm text-text-primary font-mono">
                              {stat.origin}
                            </span>
                          </div>
                          <Badge variant="secondary">
                            {stat.count.toLocaleString()} calls
                          </Badge>
                        </div>
                      ))}
                    </div>
                  </div>
                ) : (
                  <div className="flex items-center gap-2 text-text-muted p-4">
                    <AlertCircle className="w-4 h-4" />
                    <span>No embed executions yet</span>
                  </div>
                )}
              </div>
            ) : (
              <div className="flex items-center gap-2 text-text-muted p-4">
                <AlertCircle className="w-4 h-4" />
                <span>Analytics data unavailable</span>
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
