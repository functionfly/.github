import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import {
  type PlatformStatus,
  type ComponentHealth,
  type ProviderStatus,
  type Incident,
  type MaintenanceWindow,
  type UptimeMetrics,
} from '@/api/status';

interface StatusState {
  // Status data
  platformStatus: PlatformStatus | null;
  components: ComponentHealth[];
  providers: ProviderStatus[];
  incidents: Incident[];
  maintenanceWindows: MaintenanceWindow[];
  uptimeMetrics: UptimeMetrics | null;

  // UI State
  selectedTimeRange: 30 | 90 | 365;
  selectedProvider: string | null;
  selectedRegion: string | null;
  lastUpdated: string | null;

  // Loading states
  isLoading: boolean;
  isInitialLoad: boolean;

  // Error state
  error: string | null;
}

interface StatusActions {
  // Data setters
  setPlatformStatus: (status: PlatformStatus) => void;
  setComponents: (components: ComponentHealth[]) => void;
  setProviders: (providers: ProviderStatus[]) => void;
  setIncidents: (incidents: Incident[]) => void;
  addIncident: (incident: Incident) => void;
  updateIncident: (id: string, updates: Partial<Incident>) => void;
  removeIncident: (id: string) => void;
  setMaintenanceWindows: (windows: MaintenanceWindow[]) => void;
  setUptimeMetrics: (metrics: UptimeMetrics) => void;

  // UI actions
  setSelectedTimeRange: (range: 30 | 90 | 365) => void;
  setSelectedProvider: (provider: string | null) => void;
  setSelectedRegion: (region: string | null) => void;

  // Loading actions
  setIsLoading: (isLoading: boolean) => void;
  setIsInitialLoad: (isInitial: boolean) => void;

  // Error actions
  setError: (error: string | null) => void;

  // Bulk updates
  updateFromWebSocket: (data: PlatformStatus) => void;
  refreshAll: (data: {
    platform: PlatformStatus;
    components: ComponentHealth[];
    providers: ProviderStatus[];
  }) => void;

  // Computed helpers
  getComponentById: (id: string) => ComponentHealth | undefined;
  getProviderById: (id: string) => ProviderStatus | undefined;
  getActiveIncidents: () => Incident[];
  getResolvedIncidents: () => Incident[];
}

const initialState: StatusState = {
  platformStatus: null,
  components: [],
  providers: [],
  incidents: [],
  maintenanceWindows: [],
  uptimeMetrics: null,
  selectedTimeRange: 30,
  selectedProvider: null,
  selectedRegion: null,
  lastUpdated: null,
  isLoading: false,
  isInitialLoad: true,
  error: null,
};

export const useStatusStore = create<StatusState & StatusActions>()(
  devtools(
    (set, get) => ({
      ...initialState,

      // Data setters
      setPlatformStatus: (status) =>
        set({
          platformStatus: status,
          lastUpdated: new Date().toISOString(),
        }),

      setComponents: (components) =>
        set({
          components,
          lastUpdated: new Date().toISOString(),
        }),

      setProviders: (providers) =>
        set({
          providers,
          lastUpdated: new Date().toISOString(),
        }),

      setIncidents: (incidents) =>
        set({
          incidents,
          lastUpdated: new Date().toISOString(),
        }),

      addIncident: (incident) =>
        set((state) => ({
          incidents: [incident, ...state.incidents],
          lastUpdated: new Date().toISOString(),
        })),

      updateIncident: (id, updates) =>
        set((state) => ({
          incidents: state.incidents.map((incident) =>
            incident.id === id ? { ...incident, ...updates } : incident
          ),
          lastUpdated: new Date().toISOString(),
        })),

      removeIncident: (id) =>
        set((state) => ({
          incidents: state.incidents.filter((incident) => incident.id !== id),
        })),

      setMaintenanceWindows: (windows) =>
        set({
          maintenanceWindows: windows,
        }),

      setUptimeMetrics: (metrics) =>
        set({
          uptimeMetrics: metrics,
        }),

      // UI actions
      setSelectedTimeRange: (range) =>
        set({
          selectedTimeRange: range,
        }),

      setSelectedProvider: (provider) =>
        set({
          selectedProvider: provider,
          selectedRegion: null, // Reset region when provider changes
        }),

      setSelectedRegion: (region) =>
        set({
          selectedRegion: region,
        }),

      // Loading actions
      setIsLoading: (isLoading) =>
        set({
          isLoading,
        }),

      setIsInitialLoad: (isInitialLoad) =>
        set({
          isInitialLoad,
        }),

      // Error actions
      setError: (error) =>
        set({
          error,
        }),

      // Bulk updates
      updateFromWebSocket: (data) =>
        set({
          platformStatus: data,
          components: data.components,
          lastUpdated: new Date().toISOString(),
        }),

      refreshAll: (data) =>
        set({
          platformStatus: data.platform,
          components: data.components,
          providers: data.providers,
          lastUpdated: new Date().toISOString(),
          isInitialLoad: false,
        }),

      // Computed helpers
      getComponentById: (id) => {
        return get().components.find((c) => c.id === id);
      },

      getProviderById: (id) => {
        return get().providers.find((p) => p.id === id);
      },

      getActiveIncidents: () => {
        return get().incidents.filter(
          (i) => i.status !== 'resolved'
        );
      },

      getResolvedIncidents: () => {
        return get().incidents.filter(
          (i) => i.status === 'resolved'
        );
      },
    }),
    {
      name: 'status-store',
    }
  )
);

// ============================================================================
// Selectors for better performance
// ============================================================================

export const selectPlatformStatus = (state: StatusState) => state.platformStatus;
export const selectComponents = (state: StatusState) => state.components;
export const selectProviders = (state: StatusState) => state.providers;
export const selectIncidents = (state: StatusState) => state.incidents;
export const selectActiveIncidents = (state: StatusState) =>
  state.incidents.filter((i) => i.status !== 'resolved');
export const selectIsLoading = (state: StatusState) => state.isLoading;
export const selectLastUpdated = (state: StatusState) => state.lastUpdated;

// ============================================================================
// Status summary helpers
// ============================================================================

export function useStatusSummary() {
  const store = useStatusStore();

  const getOverallStatus = () => {
    if (!store.platformStatus) return 'unknown';
    return store.platformStatus.status;
  };

  const getAffectedComponents = () => {
    return store.components.filter(
      (c) => c.status !== 'operational'
    );
  };

  const getUptimeSummary = () => {
    const metrics = store.uptimeMetrics;
    if (!metrics) return null;

    return {
      overall: metrics.overall_uptime,
      worstComponent: Object.entries(metrics.by_component).sort(
        ([, a], [, b]) => a - b
      )[0],
      bestComponent: Object.entries(metrics.by_component).sort(
        ([, a], [, b]) => b - a
      )[0],
    };
  };

  return {
    overallStatus: getOverallStatus(),
    affectedComponents: getAffectedComponents(),
    activeIncidents: store.getActiveIncidents(),
    uptimeSummary: getUptimeSummary(),
    lastUpdated: store.lastUpdated,
  };
}
