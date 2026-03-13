import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { toast } from "sonner";

const NOTIFICATION_OPTIONS = [
  {
    key: "deploymentSuccess" as const,
    label: "Deployment Success",
    description: "Get notified when a deployment succeeds",
  },
  {
    key: "deploymentFailure" as const,
    label: "Deployment Failure",
    description: "Get notified when a deployment fails",
  },
  {
    key: "failoverEvents" as const,
    label: "Failover Events",
    description: "Get notified when failover is triggered",
  },
  {
    key: "providerIssues" as const,
    label: "Provider Issues",
    description: "Get notified when a provider has issues",
  },
] as const;

const DEFAULT_PREFS = {
  deploymentSuccess: true,
  deploymentFailure: true,
  failoverEvents: true,
  providerIssues: true,
};

function loadPreferences(): typeof DEFAULT_PREFS {
  const saved = localStorage.getItem("notificationPreferences");
  if (saved) {
    try {
      return { ...DEFAULT_PREFS, ...JSON.parse(saved) };
    } catch {
      /* ignore */
    }
  }
  return DEFAULT_PREFS;
}

export function NotificationsSettingsTab() {
  const [notifications, setNotifications] = useState(loadPreferences);

  const handleToggle = (key: keyof typeof DEFAULT_PREFS, checked: boolean) => {
    const updated = { ...notifications, [key]: checked };
    setNotifications(updated);
    localStorage.setItem("notificationPreferences", JSON.stringify(updated));
    toast.success("Notification preference saved");
  };

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Notification Preferences</CardTitle>
          <CardDescription className="text-text-secondary">
            Choose what notifications you want to receive
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {NOTIFICATION_OPTIONS.map((item) => (
              <div key={item.key} className="flex items-center justify-between">
                <div>
                  <h4 className="font-medium text-white">{item.label}</h4>
                  <p className="text-sm text-text-muted">{item.description}</p>
                </div>
                <Switch
                  checked={notifications[item.key]}
                  onCheckedChange={(checked) => handleToggle(item.key, checked)}
                />
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
