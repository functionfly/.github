import { useState, useEffect } from "react";
import { BarChart3, Eye, EyeOff, Settings, CheckCircle, AlertCircle, Loader2 } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { getAnalyticsSettings, updateAnalyticsSettings } from "@/api";
import type { AnalyticsSettings } from "@/types";

export function AnalyticsManagement() {
  const [settings, setSettings] = useState<AnalyticsSettings | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Form state
  const [googleAnalyticsId, setGoogleAnalyticsId] = useState("");
  const [googleAnalyticsEnabled, setGoogleAnalyticsEnabled] = useState(false);
  const [hotjarSiteId, setHotjarSiteId] = useState("");
  const [hotjarEnabled, setHotjarEnabled] = useState(false);

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      setLoading(true);
      const data = await getAnalyticsSettings();
      setSettings(data);

      // Initialize form values
      if (data.googleAnalytics) {
        setGoogleAnalyticsId(data.googleAnalytics.measurementId || "");
        setGoogleAnalyticsEnabled(data.googleAnalytics.enabled || false);
      }
      if (data.hotjar) {
        setHotjarSiteId(data.hotjar.siteId || "");
        setHotjarEnabled(data.hotjar.enabled || false);
      }
    } catch (err) {
      setError("Failed to load analytics settings");
      console.error("Error loading analytics settings:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    try {
      setSaving(true);
      setError(null);
      setSuccess(null);

      const updateData = {
        googleAnalytics: googleAnalyticsEnabled ? {
          measurementId: googleAnalyticsId,
          enabled: true,
        } : undefined,
        hotjar: hotjarEnabled ? {
          siteId: hotjarSiteId,
          enabled: true,
        } : undefined,
      };

      const response = await updateAnalyticsSettings(updateData);
      setSuccess(response.message);

      // Reload settings to get updated data
      await loadSettings();
    } catch (err) {
      setError("Failed to update analytics settings");
      console.error("Error updating analytics settings:", err);
    } finally {
      setSaving(false);
    }
  };

  const getServiceStatusColor = (status: string) => {
    switch (status) {
      case "loaded":
        return "bg-emerald-500/10 text-emerald-400 border-emerald-500/20";
      case "loading":
        return "bg-blue-500/10 text-blue-400 border-blue-500/20";
      case "error":
        return "bg-red-500/10 text-red-400 border-red-500/20";
      case "disabled":
        return "bg-gray-500/10 text-gray-400 border-gray-500/20";
      default:
        return "bg-gray-500/10 text-gray-400 border-gray-500/20";
    }
  };

  const getServiceStatusIcon = (status: string) => {
    switch (status) {
      case "loaded":
        return <CheckCircle className="w-4 h-4" />;
      case "loading":
        return <Loader2 className="w-4 h-4 animate-spin" />;
      case "error":
        return <AlertCircle className="w-4 h-4" />;
      case "disabled":
        return <EyeOff className="w-4 h-4" />;
      default:
        return <AlertCircle className="w-4 h-4" />;
    }
  };

  if (loading) {
    return (
      <Card className="card">
        <CardContent className="card-content">
          <div className="flex items-center justify-center py-8">
            <Loader2 className="w-6 h-6 animate-spin mr-2" />
            <span>Loading analytics settings...</span>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold text-white flex items-center gap-2">
            <BarChart3 className="w-5 h-5" />
            Analytics Management
          </h2>
          <p className="text-text-secondary">Configure Google Analytics and Hotjar tracking</p>
        </div>
      </div>

      {/* Success/Error Messages */}
      {error && (
        <Alert className="border-red-500/20 bg-red-500/10">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription className="text-red-400">{error}</AlertDescription>
        </Alert>
      )}

      {success && (
        <Alert className="border-emerald-500/20 bg-emerald-500/10">
          <CheckCircle className="h-4 w-4" />
          <AlertDescription className="text-emerald-400">{success}</AlertDescription>
        </Alert>
      )}

      {/* Service Status */}
      <Card className="card">
        <CardHeader className="card-header">
          <CardTitle className="flex items-center gap-2">
            <Eye className="w-5 h-5" />
            Running Services
          </CardTitle>
        </CardHeader>
        <CardContent className="card-content">
          {settings?.services && settings.services.length > 0 ? (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {settings.services.map((service, index) => (
                <div key={index} className="flex items-center justify-between p-4 border border-border-default rounded-lg">
                  <div className="flex items-center gap-3">
                    {getServiceStatusIcon(service.status)}
                    <div>
                      <p className="font-medium text-white">{service.name}</p>
                      <p className="text-sm text-text-secondary capitalize">{service.status}</p>
                    </div>
                  </div>
                  <Badge className={getServiceStatusColor(service.status)}>
                    {service.status}
                  </Badge>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8">
              <BarChart3 className="w-12 h-12 text-text-secondary mx-auto mb-4" />
              <p className="text-white font-medium">No analytics services configured</p>
              <p className="text-text-secondary">Configure services below to start tracking</p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Configuration */}
      <Card className="card">
        <CardHeader className="card-header">
          <CardTitle className="flex items-center gap-2">
            <Settings className="w-5 h-5" />
            Service Configuration
          </CardTitle>
        </CardHeader>
        <CardContent className="card-content space-y-6">
          {/* Google Analytics */}
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 bg-blue-500/10 rounded-lg flex items-center justify-center">
                  <span className="text-blue-400 font-bold text-sm">GA</span>
                </div>
                <div>
                  <h3 className="font-medium text-white">Google Analytics 4</h3>
                  <p className="text-sm text-text-secondary">Track user behavior and website analytics</p>
                </div>
              </div>
              <Switch
                checked={googleAnalyticsEnabled}
                onCheckedChange={setGoogleAnalyticsEnabled}
              />
            </div>

            {googleAnalyticsEnabled && (
              <div className="ml-11 space-y-2">
                <Label htmlFor="ga-id" className="text-text-secondary">
                  Measurement ID (G-XXXXXXXXXX)
                </Label>
                <Input
                  id="ga-id"
                  value={googleAnalyticsId}
                  onChange={(e) => setGoogleAnalyticsId(e.target.value)}
                  placeholder="G-XXXXXXXXXX"
                  className="input"
                />
                <p className="text-xs text-text-muted">
                  Get your Measurement ID from Google Analytics
                </p>
              </div>
            )}
          </div>

          {/* Hotjar */}
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 bg-orange-500/10 rounded-lg flex items-center justify-center">
                  <span className="text-orange-400 font-bold text-sm">HJ</span>
                </div>
                <div>
                  <h3 className="font-medium text-white">Hotjar</h3>
                  <p className="text-sm text-text-secondary">Record user sessions and heatmaps</p>
                </div>
              </div>
              <Switch
                checked={hotjarEnabled}
                onCheckedChange={setHotjarEnabled}
              />
            </div>

            {hotjarEnabled && (
              <div className="ml-11 space-y-2">
                <Label htmlFor="hj-id" className="text-text-secondary">
                  Site ID (numeric)
                </Label>
                <Input
                  id="hj-id"
                  value={hotjarSiteId}
                  onChange={(e) => setHotjarSiteId(e.target.value)}
                  placeholder="1234567"
                  className="input"
                />
                <p className="text-xs text-text-muted">
                  Get your Site ID from Hotjar settings
                </p>
              </div>
            )}
          </div>

          {/* Save Button */}
          <div className="flex justify-end pt-4 border-t border-border-default">
            <Button
              onClick={handleSave}
              disabled={saving}
              className="btn-primary"
            >
              {saving ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Saving...
                </>
              ) : (
                <>
                  <CheckCircle className="w-4 h-4 mr-2" />
                  Save Changes
                </>
              )}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Note */}
      <Alert className="border-blue-500/20 bg-blue-500/10">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription className="text-blue-400">
          Changes will take effect on the next deployment. Analytics scripts are loaded conditionally based on user cookie consent preferences.
        </AlertDescription>
      </Alert>
    </div>
  );
}