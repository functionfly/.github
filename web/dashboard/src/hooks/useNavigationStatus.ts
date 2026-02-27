import { useMemo } from "react";
import { useLocation } from "react-router-dom";

interface NavigationStatus {
  functions: {
    hasIssues: boolean;
    pendingDeployments: number;
    totalCount: number;
  };
  providers: {
    hasOffline: boolean;
    totalCount: number;
  };
  analytics: {
    hasAlerts: boolean;
  };
  settings: {
    hasWarnings: boolean;
  };
}

// Mock data - in real app this would come from API/websockets
const mockStatusData: NavigationStatus = {
  functions: {
    hasIssues: true, // Some functions have deployment issues
    pendingDeployments: 2,
    totalCount: 5,
  },
  providers: {
    hasOffline: false,
    totalCount: 3,
  },
  analytics: {
    hasAlerts: false,
  },
  settings: {
    hasWarnings: true, // Billing issues or something
  },
};

export function useNavigationStatus(): NavigationStatus {
  const location = useLocation();

  // In a real app, this would fetch from API with polling/websockets
  // For now, return mock data
  return useMemo(() => mockStatusData, []);
}