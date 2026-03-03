import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { AlertTriangle, CheckCircle, Clock, Power, Settings, History, Info } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Slider } from "@/components/ui/slider";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { LoadingSpinner } from "@/components/common/LoadingSpinner";
import { maintenanceApi, type MaintenanceConfig } from "@/api/admin";

export function MaintenanceMode() {
  const queryClient = useQueryClient();
  const [showEnableDialog, setShowEnableDialog] = useState(false);
  const [showDisableDialog, setShowDisableDialog] = useState(false);
  const [maintenanceConfig, setMaintenanceConfig] = useState<Partial<MaintenanceConfig>>({
    name: "Scheduled Maintenance",
    message: "We're performing scheduled maintenance. We'll be back shortly.",
    retry_after_seconds: 3600,
    rollout_percentage: 100,
  });

  // Fetch maintenance status
  const { data: maintenanceStatus, isLoading, error } = useQuery({
    queryKey: ['maintenance-status'],
    queryFn: () => maintenanceApi.getStatus(),
    refetchInterval: 30000,
  });

  // Enable maintenance mutation
  const enableMutation = useMutation({
    mutationFn: maintenanceApi.enable,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['maintenance-status'] });
      setShowEnableDialog(false);
    },
  });

  // Disable maintenance mutation
  const disableMutation = useMutation({
    mutationFn: maintenanceApi.disable,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['maintenance-status'] });
      setShowDisableDialog(false);
    },
  });

  const handleEnableMaintenance = () => {
    enableMutation.mutate(maintenanceConfig);
  };

  const handleDisableMaintenance = () => {
    disableMutation.mutate();
  };

  const getStatusColor = (enabled: boolean) => {
    return enabled
      ? "bg-red-500/10 text-red-600 dark:text-red-400 border-red-500/20"
      : "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20";
  };

  if (isLoading) {
    return (
      <Card>
        <CardContent className="p-6">
          <div className="flex items-center justify-center h-32">
            <LoadingSpinner size="md" />
          </div>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent className="p-6">
          <Alert variant="destructive">
            <AlertTriangle className="h-4 w-4" />
            <AlertDescription>Failed to load maintenance status</AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    );
  }

  const isEnabled = maintenanceStatus?.enabled || false;

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-amber-500/10 rounded-lg">
                <Power className="w-5 h-5 text-amber-600 dark:text-amber-400" />
              </div>
              <div>
                <CardTitle className="text-lg">Maintenance Mode</CardTitle>
                <CardDescription>Control platform-wide maintenance status</CardDescription>
              </div>
            </div>
            <Badge className={getStatusColor(isEnabled)}>
              {isEnabled ? "Active" : "Inactive"}
            </Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Current Status */}
          <div className={`p-4 rounded-lg border ${isEnabled ? 'bg-red-500/5 border-red-500/20' : 'bg-emerald-500/5 border-emerald-500/20'}`}>
            <div className="flex items-start gap-3">
              {isEnabled ? (
                <AlertTriangle className="w-5 h-5 text-red-500 mt-0.5" />
              ) : (
                <CheckCircle className="w-5 h-5 text-emerald-500 mt-0.5" />
              )}
              <div className="flex-1">
                <h4 className="font-medium text-text-primary">
                  {isEnabled ? "Maintenance Mode is Active" : "Platform is Operational"}
                </h4>
                <p className="text-sm text-text-secondary mt-1">
                  {isEnabled
                    ? `Maintenance "${maintenanceStatus?.name}" is currently active. All visitors will see the maintenance page.`
                    : "The platform is running normally. Users can access all features."
                  }
                </p>
                {isEnabled && maintenanceStatus?.message && (
                  <p className="text-sm text-text-secondary mt-2 italic">
                    Message: "{maintenanceStatus.message}"
                  </p>
                )}
              </div>
            </div>
          </div>

          {/* Action Buttons */}
          <div className="flex flex-wrap gap-3">
            {!isEnabled ? (
              <Button
                onClick={() => setShowEnableDialog(true)}
                variant="destructive"
                className="gap-2"
              >
                <Power className="w-4 h-4" />
                Enable Maintenance Mode
              </Button>
            ) : (
              <Button
                onClick={() => setShowDisableDialog(true)}
                variant="default"
                className="gap-2 bg-emerald-600 hover:bg-emerald-700"
              >
                <CheckCircle className="w-4 h-4" />
                Disable Maintenance Mode
              </Button>
            )}
          </div>

          {/* Configuration Summary */}
          {isEnabled && maintenanceStatus && (
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-4 border-t border-border">
              <div className="flex items-center gap-3">
                <Settings className="w-4 h-4 text-text-muted" />
                <div>
                  <p className="text-xs text-text-muted">Template</p>
                  <p className="text-sm font-medium text-text-primary">{maintenanceStatus.page_template}</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <Clock className="w-4 h-4 text-text-muted" />
                <div>
                  <p className="text-xs text-text-muted">Retry After</p>
                  <p className="text-sm font-medium text-text-primary">{maintenanceStatus.retry_after_seconds} seconds</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <Info className="w-4 h-4 text-text-muted" />
                <div>
                  <p className="text-xs text-text-muted">Rollout</p>
                  <p className="text-sm font-medium text-text-primary">{maintenanceStatus.rollout_percentage}% of traffic</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <History className="w-4 h-4 text-text-muted" />
                <div>
                  <p className="text-xs text-text-muted">Last Updated</p>
                  <p className="text-sm font-medium text-text-primary">
                    {new Date(maintenanceStatus.updated_at).toLocaleString()}
                  </p>
                </div>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Enable Maintenance Dialog */}
      <Dialog open={showEnableDialog} onOpenChange={setShowEnableDialog}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <AlertTriangle className="w-5 h-5 text-amber-500" />
              Enable Maintenance Mode
            </DialogTitle>
            <DialogDescription>
              This will show a maintenance page to all visitors. Make sure to notify users beforehand.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="name">Maintenance Name</Label>
              <Input
                id="name"
                value={maintenanceConfig.name}
                onChange={(e) => setMaintenanceConfig({ ...maintenanceConfig, name: e.target.value })}
                placeholder="e.g., Database Upgrade"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="message">Message to Visitors</Label>
              <Textarea
                id="message"
                value={maintenanceConfig.message}
                onChange={(e) => setMaintenanceConfig({ ...maintenanceConfig, message: e.target.value })}
                placeholder="We'll be back soon..."
                rows={3}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="retry-after">Retry-After (seconds)</Label>
              <Input
                id="retry-after"
                type="number"
                value={maintenanceConfig.retry_after_seconds}
                onChange={(e) => setMaintenanceConfig({ ...maintenanceConfig, retry_after_seconds: parseInt(e.target.value) })}
                min={60}
                max={86400}
              />
              <p className="text-xs text-text-muted">
                Tells browsers and search engines when to retry (in seconds)
              </p>
            </div>

            <div className="space-y-4">
              <div className="flex justify-between items-center">
                <Label>Rollout Percentage</Label>
                <span className="text-sm font-medium">{maintenanceConfig.rollout_percentage}%</span>
              </div>
              <Slider
                value={[maintenanceConfig.rollout_percentage || 100]}
                onValueChange={([value]) => setMaintenanceConfig({ ...maintenanceConfig, rollout_percentage: value })}
                min={0}
                max={100}
                step={1}
              />
              <p className="text-xs text-text-muted">
                Percentage of visitors who will see the maintenance page. Use lower values to test before full rollout.
              </p>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowEnableDialog(false)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleEnableMaintenance}
              disabled={enableMutation.isPending}
              className="gap-2"
            >
              {enableMutation.isPending ? (
                <LoadingSpinner size="sm" />
              ) : (
                <Power className="w-4 h-4" />
              )}
              Enable Maintenance
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Disable Maintenance Dialog */}
      <Dialog open={showDisableDialog} onOpenChange={setShowDisableDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <CheckCircle className="w-5 h-5 text-emerald-500" />
              Disable Maintenance Mode
            </DialogTitle>
            <DialogDescription>
              This will restore normal access to the platform for all visitors.
            </DialogDescription>
          </DialogHeader>

          <div className="py-4">
            <Alert>
              <Info className="h-4 w-4" />
              <AlertDescription>
                Make sure all maintenance work is complete before disabling. Users will immediately regain access.
              </AlertDescription>
            </Alert>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDisableDialog(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleDisableMaintenance}
              disabled={disableMutation.isPending}
              className="gap-2 bg-emerald-600 hover:bg-emerald-700"
            >
              {disableMutation.isPending ? (
                <LoadingSpinner size="sm" />
              ) : (
                <CheckCircle className="w-4 h-4" />
              )}
              Disable Maintenance
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
