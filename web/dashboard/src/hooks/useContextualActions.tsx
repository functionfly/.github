import { useLocation } from "react-router-dom";
import { Plus, Link as LinkIcon, Rocket, Download } from "lucide-react";
import { ROUTES } from "@/lib/constants";
import { useNavigationStatus } from "./useNavigationStatus";

export interface QuickAction {
  id: string;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
  variant?: "default" | "secondary" | "outline";
  onClick: () => void;
}

export function useContextualActions(): QuickAction[] {
  const location = useLocation();
  const status = useNavigationStatus();

  const actions: QuickAction[] = [];

  // Functions page actions
  if (location.pathname === ROUTES.FUNCTIONS || location.pathname.startsWith('/functions/')) {
    actions.push({
      id: 'new-function',
      label: 'New Function',
      icon: Plus,
      variant: 'default' as const,
      onClick: () => {
        // In real app, navigate to new function page or open modal
        console.log('Navigate to new function');
        window.location.href = '/functions/new';
      }
    });
  }

  // Providers page actions
  if (location.pathname === ROUTES.PROVIDERS) {
    actions.push({
      id: 'connect-provider',
      label: 'Connect Provider',
      icon: LinkIcon,
      variant: 'default' as const,
      onClick: () => {
        // In real app, open connect provider modal
        console.log('Open connect provider modal');
      }
    });
  }

  // Global actions based on status
  if (status.functions.pendingDeployments > 0) {
    actions.push({
      id: 'deploy-all',
      label: `Deploy All (${status.functions.pendingDeployments})`,
      icon: Rocket,
      variant: 'outline' as const,
      onClick: () => {
        // In real app, trigger deploy all action
        console.log('Deploy all pending functions');
      }
    });
  }

  // Analytics page actions
  if (location.pathname === ROUTES.ANALYTICS) {
    actions.push({
      id: 'export-data',
      label: 'Export Data',
      icon: Download,
      variant: 'outline' as const,
      onClick: () => {
        // In real app, trigger export
        console.log('Export analytics data');
      }
    });
  }

  return actions;
}