import { useParams, useNavigate } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import { Button } from "@/components/ui/button";
import { ArrowLeft } from "lucide-react";
import { SettingsContent } from "@/pages/SettingsPage/SettingsContent";

interface UserDashboardSettingsPageProps {
  initialTab?: string;
}

/**
 * Settings page at /u/:username/settings (and /u/:username/settings/billing).
 * Renders the full settings content (Account, Billing with Enterprise/Invoices/Contact Sales,
 * API Keys, Notifications, Privacy) from the old standalone settings page.
 */
export function UserDashboardSettingsPage({ initialTab }: UserDashboardSettingsPageProps) {
  const { username } = useParams<{ username: string }>();
  const navigate = useNavigate();
  const user = useAuthStore((s) => s.user);

  // Only allow viewing own settings
  if (user?.username && username && user.username !== username) {
    navigate(`/u/${user.username}/settings`, { replace: true });
    return null;
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => navigate(`/u/${username}`)}
            className="text-text-secondary hover:text-text-primary"
            aria-label="Back to profile"
          >
            <ArrowLeft className="w-4 h-4" />
          </Button>
          <div>
            <h1 className="font-display text-2xl font-bold bg-gradient-to-r from-brand-500 via-ff-afterburner to-brand-400 bg-clip-text text-transparent">
              Settings
            </h1>
            <p className="text-text-secondary">
              Manage your profile and preferences
            </p>
          </div>
        </div>
      </div>

      <SettingsContent
        showHeader={false}
        profile={undefined}
        initialTab={initialTab}
      />
    </div>
  );
}

export default UserDashboardSettingsPage;
