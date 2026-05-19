import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

export type ChartType =
  | 'streaming-line'
  | 'realtime-scatter'
  | '3d-topology'
  | 'execution-sunburst'
  | 'dependency-treemap'
  | 'circular-flow'
  | 'waterfall'
  | 'cost-distribution'
  | 'semantic-cluster'
  | 'agent-interaction';

export interface ChartData {
  id: string;
  type: ChartType;
  data: Array<{ x: number; y: number; z?: number; label?: string }>;
  timestamp: Date;
  metadata?: Record<string, unknown>;
}

export interface TimeRange {
  start: Date;
  end: Date;
  label?: string;
}

export interface DisplayOptions {
  theme: 'dark' | 'light';
  showGrid: boolean;
  showLegend: boolean;
  showTooltip: boolean;
  animationDuration: number;
  colorScheme?: string[];
}

export interface Annotation {
  id: string;
  timestamp: Date;
  label: string;
  description?: string;
  type: 'marker' | 'region' | 'alert';
}

export type ViewType = ChartType;

interface DataVisualizationState {
  charts: ChartData[];
  selectedChartId: string | null;
  activeView: ViewType;
  timeRange: TimeRange;
  refreshInterval: number;
  displayOptions: DisplayOptions;
  annotations: Annotation[];
}

interface DataVisualizationActions {
  selectChart: (chartId: string | null) => void;
  setActiveView: (view: ViewType) => void;
  setTimeRange: (range: TimeRange) => void;
  setRefreshInterval: (interval: number) => void;
  updateDisplayOptions: (options: Partial<DisplayOptions>) => void;
  addAnnotation: (annotation: Annotation) => void;
  removeAnnotation: (annotationId: string) => void;
  pushDataPoint: (chartId: string, point: { x: number; y: number; z?: number; label?: string }) => void;
}

const useDataVisualizationStoreBase = create<DataVisualizationState & DataVisualizationActions>()(
  immer((set) => ({
    charts: [],
    selectedChartId: null,
    activeView: 'streaming-line',
    timeRange: {
      start: new Date(Date.now() - 3600000),
      end: new Date(),
      label: 'Last Hour',
    },
    refreshInterval: 5000,
    displayOptions: {
      theme: 'dark',
      showGrid: true,
      showLegend: true,
      showTooltip: true,
      animationDuration: 300,
    },
    annotations: [],

    selectChart: (chartId) =>
      set((state) => {
        state.selectedChartId = chartId;
      }),

    setActiveView: (view) =>
      set((state) => {
        state.activeView = view;
      }),

    setTimeRange: (range) =>
      set((state) => {
        state.timeRange = range;
      }),

    setRefreshInterval: (interval) =>
      set((state) => {
        state.refreshInterval = interval;
      }),

    updateDisplayOptions: (options) =>
      set((state) => {
        Object.assign(state.displayOptions, options);
      }),

    addAnnotation: (annotation) =>
      set((state) => {
        state.annotations.push(annotation);
      }),

    removeAnnotation: (annotationId) =>
      set((state) => {
        state.annotations = state.annotations.filter((a) => a.id !== annotationId);
      }),

    pushDataPoint: (chartId, point) =>
      set((state) => {
        const chart = state.charts.find((c) => c.id === chartId);
        if (chart) {
          chart.data.push(point);
          chart.timestamp = new Date();
        }
      }),
  }))
);

export const useDataVisualizationStore = useDataVisualizationStoreBase;

export const selectChartByType = (type: ChartType) => (state: DataVisualizationState) =>
  state.charts.filter((c) => c.type === type);

export const selectVisibleCharts = (state: DataVisualizationState) =>
  state.charts.filter((c) => c.data.length > 0);

export const selectStreamingCharts = (state: DataVisualizationState) =>
  state.charts.filter((c) => c.type === 'streaming-line' || c.type === 'realtime-scatter');

export const selectChartData = (chartId: string) => (state: DataVisualizationState) =>
  state.charts.find((c) => c.id === chartId)?.data ?? [];

export const useDataVisualization = () => {
  const store = useDataVisualizationStore();
  return store;
};

export const useStreamingLineChart = () => {
  const store = useDataVisualizationStore();
  const streamingData = store.charts.filter((c) => c.type === 'streaming-line').flatMap((c) => c.data);
  return { data: streamingData, activeView: store.activeView, setActiveView: store.setActiveView };
};

export const useRealtimeScatterPlot = () => {
  const store = useDataVisualizationStore();
  const scatterData = store.charts.filter((c) => c.type === 'realtime-scatter').flatMap((c) => c.data);
  return { data: scatterData, activeView: store.activeView, setActiveView: store.setActiveView };
};

export const use3DTopologyChart = () => {
  const store = useDataVisualizationStore();
  const topologyData = store.charts.filter((c) => c.type === '3d-topology').flatMap((c) => c.data);
  return { data: topologyData, activeView: store.activeView, setActiveView: store.setActiveView };
};

export const useExecutionSunburst = () => {
  const store = useDataVisualizationStore();
  const sunburstData = store.charts.filter((c) => c.type === 'execution-sunburst').flatMap((c) => c.data);
  return { data: sunburstData, activeView: store.activeView, setActiveView: store.setActiveView };
};

export const useDependencyTreemap = () => {
  const store = useDataVisualizationStore();
  const treemapData = store.charts.filter((c) => c.type === 'dependency-treemap').flatMap((c) => c.data);
  return { data: treemapData, activeView: store.activeView, setActiveView: store.setActiveView };
};

export const useCircularFlow = () => {
  const store = useDataVisualizationStore();
  const circularData = store.charts.filter((c) => c.type === 'circular-flow').flatMap((c) => c.data);
  return { data: circularData, activeView: store.activeView, setActiveView: store.setActiveView };
};

export const useWaterfallChart = () => {
  const store = useDataVisualizationStore();
  const waterfallData = store.charts.filter((c) => c.type === 'waterfall').flatMap((c) => c.data);
  return { data: waterfallData, activeView: store.activeView, setActiveView: store.setActiveView };
};

export const useCostDistribution = () => {
  const store = useDataVisualizationStore();
  const costData = store.charts.filter((c) => c.type === 'cost-distribution').flatMap((c) => c.data);
  return { data: costData, activeView: store.activeView, setActiveView: store.setActiveView };
};

export const useSemanticCluster = () => {
  const store = useDataVisualizationStore();
  const clusterData = store.charts.filter((c) => c.type === 'semantic-cluster').flatMap((c) => c.data);
  return { data: clusterData, activeView: store.activeView, setActiveView: store.setActiveView };
};

export const useAgentInteractionGraph = () => {
  const store = useDataVisualizationStore();
  const graphData = store.charts.filter((c) => c.type === 'agent-interaction').flatMap((c) => c.data);
  return { data: graphData, activeView: store.activeView, setActiveView: store.setActiveView };
};
