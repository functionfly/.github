import { Navigate, useSearchParams } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import { getApiBaseUrl } from "@/lib/constants";
import { useQuery } from "@tanstack/react-query";
import { SettingsContent } from "./SettingsContent";

/**
 * Standalone Settings page. When the user is logged in, redirect to their
 * profile Settings tab (/u/{username}?tab=settings) so settings live in one place.
 * When not logged in, show the shared settings content (e.g. for layout consistency).
 */
export function SettingsPage() {
  const user = useAuthStore((state) => state.user);
  const [searchParams] = useSearchParams();

  if (user?.username) {
    const subtab = searchParams.get("subtab");
    const success = searchParams.get("success");
    const path = subtab === "billing"
      ? `/u/${user.username}/settings/billing`
      : `/u/${user.username}/settings`;
    const q = new URLSearchParams();
    if (success === "true") q.set("success", "true");
    const queryString = q.toString();
    return (
      <Navigate
        to={queryString ? `${path}?${queryString}` : path}
        replace
      />
    );
  }

  return (
    <div className="space-y-6">
      <SettingsContent showHeader />
    </div>
  );
}
