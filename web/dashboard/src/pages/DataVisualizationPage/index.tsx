import React, { useState } from 'react';
import {
  StreamingLineChart,
  RealtimeScatterPlot,
  ThreeDTopologyChart,
  ExecutionSunburst,
  DependencyTreemap,
  CircularFlowDiagram,
  RuntimeWaterfallChart,
  CostDistributionGraph,
  SemanticClusterChart,
  AgentInteractionGraph,
} from '@functionfly/ui-data-visualization';
import { useDataVisualizationStore } from '@/stores/dataVisualizationStore';
import { cn } from '@/lib/utils';
import { RefreshCw, Settings, ChevronDown, BarChart3, ScatterChart, Globe, CircleDot, Grid3X3, TrendingDown, PieChart, Sparkles, Network } from 'lucide-react';
import {
  adaptStreamingLineData,
  adaptScatterData,
  adaptTopologyNodes,
  adaptSunburstData,
  adaptTreemapData,
  adaptWaterfallData,
  adaptCostData,
  adaptClusterData,
  extractClusters,
  adaptAgentNodes,
  adaptAgentEdges,
  adaptCircularFlowConnections,
} from '@/adapters';

const chartTypeIcons: Record<string, React.ComponentType<{ className?: string }>> = {
  'streaming-line': BarChart3,
  'realtime-scatter': ScatterChart,
  '3d-topology': Globe,
  'execution-sunburst': PieChart,
  'dependency-treemap': Grid3X3,
  'circular-flow': CircleDot,
  'waterfall': TrendingDown,
  'cost-distribution': PieChart,
  'semantic-cluster': Sparkles,
  'agent-interaction': Network,
};

const chartTypeLabels: Record<string, string> = {
  'streaming-line': 'Streaming Line',
  'realtime-scatter': 'Realtime Scatter',
  '3d-topology': '3D Topology',
  'execution-sunburst': 'Execution Sunburst',
  'dependency-treemap': 'Dependency Treemap',
  'circular-flow': 'Circular Flow',
  'waterfall': 'Waterfall',
  'cost-distribution': 'Cost Distribution',
  'semantic-cluster': 'Semantic Cluster',
  'agent-interaction': 'Agent Interaction',
};

const timeRangeOptions = [
  { label: 'Last 15 Minutes', value: 15 * 60 * 1000 },
  { label: 'Last Hour', value: 60 * 60 * 1000 },
  { label: 'Last 6 Hours', value: 6 * 60 * 60 * 1000 },
  { label: 'Last 24 Hours', value: 24 * 60 * 60 * 1000 },
  { label: 'Last 7 Days', value: 7 * 24 * 60 * 60 * 1000 },
];

const refreshIntervalOptions = [
  { label: '1s', value: 1000 },
  { label: '5s', value: 5000 },
  { label: '15s', value: 15000 },
  { label: '30s', value: 30000 },
  { label: '1m', value: 60000 },
];

export function DataVisualizationPage() {
  const {
    activeView,
    setActiveView,
    timeRange,
    setTimeRange,
    refreshInterval,
    setRefreshInterval,
    displayOptions,
    updateDisplayOptions,
  } = useDataVisualizationStore();

  const [showChartMenu, setShowChartMenu] = useState(false);
  const [showTimeMenu, setShowTimeMenu] = useState(false);
  const [showRefreshMenu, setShowRefreshMenu] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);

  const handleRefresh = () => {
    setIsRefreshing(true);
    setTimeout(() => setIsRefreshing(false), 1000);
  };

  const renderChart = () => {
    const sampleData = Array.from({ length: 20 }, (_, i) => ({
      timestamp: Date.now() - (20 - i) * 1000,
      value: Math.random() * 100,
      label: `Point ${i}`,
    }));

    const scatterSampleData = Array.from({ length: 50 }, () => ({
      x: Math.random() * 100,
      y: Math.random() * 100,
      size: Math.random() * 50,
      category: `cluster_${Math.floor(Math.random() * 5)}`,
      label: `Point`,
    }));

    const sunburstSampleData = {
      name: 'Root',
      value: 100,
      children: [
        { name: 'A', value: 40, children: [] },
        { name: 'B', value: 30, children: [] },
        { name: 'C', value: 30, children: [] },
      ],
    };

    const treemapSampleData = [
      { id: 'treemap-1', name: 'core', value: 30, children: [] },
      { id: 'treemap-2', name: 'utils', value: 25, children: [] },
      { id: 'treemap-3', name: 'api', value: 20, children: [] },
      { id: 'treemap-4', name: 'ui', value: 15, children: [] },
      { id: 'treemap-5', name: 'auth', value: 10, children: [] },
    ];

    const waterfallSampleData = [
      { name: 'Start', start: 0, end: 0, category: 'compute' },
      { name: 'CPU', start: 0, end: 25, category: 'compute' },
      { name: 'Memory', start: 25, end: 40, category: 'memory' },
      { name: 'Network', start: 40, end: 50, category: 'network' },
      { name: 'Storage', start: 50, end: 70, category: 'io' },
      { name: 'End', start: 70, end: 70, category: 'compute' },
    ];

    const costSampleData = [
      { category: 'Compute', value: 45 },
      { category: 'Storage', value: 25 },
      { category: 'Network', value: 15 },
      { category: 'API', value: 10 },
      { category: 'Other', value: 5 },
    ];

    const semanticSampleData = Array.from({ length: 30 }, (_, i) => ({
      id: `cluster-${i}`,
      x: Math.random() * 100,
      y: Math.random() * 100,
      cluster: `Cluster ${Math.floor(Math.random() * 5)}`,
      label: `Item ${i}`,
    }));

    const agentNodes = [
      { id: '1', label: 'Agent A', type: 'agent', connections: ['2', '4'] },
      { id: '2', label: 'Func B', type: 'function', connections: ['3'] },
      { id: '3', label: 'User C', type: 'user', connections: ['1'] },
      { id: '4', label: 'Agent D', type: 'agent', connections: [] },
    ];

    const agentEdges = [
      { source: '1', target: '2', weight: 3, type: 'calls' },
      { source: '2', target: '3', weight: 2, type: 'calls' },
      { source: '3', target: '1', weight: 4, type: 'calls' },
      { source: '1', target: '4', weight: 2, type: 'calls' },
    ];

    switch (activeView) {
      case 'streaming-line':
        return <StreamingLineChart data={sampleData} />;
      case 'realtime-scatter':
        return <RealtimeScatterPlot data={scatterSampleData} />;
      case '3d-topology':
        return (
          <ThreeDTopologyChart
            nodes={[
              { id: '1', label: 'Node A', depth: 0, children: [] },
              { id: '2', label: 'Node B', depth: 1, children: [] },
              { id: '3', label: 'Node C', depth: 0, children: [] },
              { id: '4', label: 'Node D', depth: 1, children: [] },
            ]}
          />
        );
      case 'execution-sunburst':
        return <ExecutionSunburst data={sunburstSampleData} />;
      case 'dependency-treemap':
        return <DependencyTreemap data={treemapSampleData} />;
      case 'circular-flow':
        return (
          <CircularFlowDiagram
            nodes={[
              { id: '1', label: 'API', value: 100, type: 'source' },
              { id: '2', label: 'Auth', value: 80, type: 'processor' },
              { id: '3', label: 'DB', value: 60, type: 'sink' },
              { id: '4', label: 'Cache', value: 40, type: 'processor' },
            ]}
            connections={[
              { source: '1', target: '2', value: 50 },
              { source: '2', target: '3', value: 40 },
              { source: '1', target: '4', value: 30 },
            ]}
          />
        );
      case 'waterfall':
        return <RuntimeWaterfallChart steps={waterfallSampleData} />;
      case 'cost-distribution':
        return <CostDistributionGraph data={costSampleData} />;
      case 'semantic-cluster':
        return <SemanticClusterChart points={semanticSampleData} clusters={['Cluster 0', 'Cluster 1', 'Cluster 2', 'Cluster 3', 'Cluster 4']} />;
      case 'agent-interaction':
        return <AgentInteractionGraph nodes={agentNodes} edges={agentEdges} />;
      default:
        return <StreamingLineChart data={sampleData} />;
    }
  };

  return (
    <div className="aviation-dv-page flex flex-col h-full">
      <header className="aviation-dv-header border-b border-aviation-border-panel px-6 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <h1 className="text-xl font-semibold text-aviation-text-primary">Data Visualization</h1>
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={handleRefresh}
              disabled={isRefreshing}
              className={cn(
                'aviation-btn-ghost p-2 rounded-lg',
                isRefreshing && 'animate-spin'
              )}
            >
              <RefreshCw className="w-4 h-4" />
            </button>
          </div>
        </div>
      </header>

      <div className="aviation-dv-toolbar flex items-center gap-4 px-6 py-3 border-b border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="relative">
          <button
            onClick={() => setShowChartMenu(!showChartMenu)}
            className="aviation-btn-secondary flex items-center gap-2 px-4 py-2 rounded-lg"
          >
            {React.createElement(chartTypeIcons[activeView] || BarChart3, { className: 'w-4 h-4' })}
            <span>{chartTypeLabels[activeView]}</span>
            <ChevronDown className="w-4 h-4" />
          </button>
          {showChartMenu && (
            <div className="absolute top-full left-0 mt-2 w-56 bg-aviation-bg-panel border border-aviation-border-panel rounded-lg shadow-lg z-50">
              {Object.entries(chartTypeLabels).map(([key, label]) => {
                const Icon = chartTypeIcons[key];
                return (
                  <button
                    key={key}
                    onClick={() => {
                      setActiveView(key as any);
                      setShowChartMenu(false);
                    }}
                    className={cn(
                      'w-full flex items-center gap-3 px-4 py-2.5 text-left hover:bg-aviation-bg-instrument transition-colors',
                      activeView === key && 'text-aviation-amber'
                    )}
                  >
                    {Icon && <Icon className="w-4 h-4" />}
                    <span>{label}</span>
                  </button>
                );
              })}
            </div>
          )}
        </div>

        <div className="relative">
          <button
            onClick={() => setShowTimeMenu(!showTimeMenu)}
            className="aviation-btn-ghost flex items-center gap-2 px-3 py-2 rounded-lg"
          >
            <span className="text-sm text-aviation-text-secondary">{timeRange.label || 'Select Range'}</span>
            <ChevronDown className="w-4 h-4" />
          </button>
          {showTimeMenu && (
            <div className="absolute top-full left-0 mt-2 w-48 bg-aviation-bg-panel border border-aviation-border-panel rounded-lg shadow-lg z-50">
              {timeRangeOptions.map((opt) => (
                <button
                  key={opt.value}
                  onClick={() => {
                    setTimeRange({
                      start: new Date(Date.now() - opt.value),
                      end: new Date(),
                      label: opt.label,
                    });
                    setShowTimeMenu(false);
                  }}
                  className="w-full px-4 py-2.5 text-left text-sm hover:bg-aviation-bg-instrument transition-colors"
                >
                  {opt.label}
                </button>
              ))}
            </div>
          )}
        </div>

        <div className="relative">
          <button
            onClick={() => setShowRefreshMenu(!showRefreshMenu)}
            className="aviation-btn-ghost flex items-center gap-2 px-3 py-2 rounded-lg"
          >
            <RefreshCw className="w-4 h-4" />
            <span className="text-sm text-aviation-text-secondary">Every {refreshInterval / 1000}s</span>
            <ChevronDown className="w-4 h-4" />
          </button>
          {showRefreshMenu && (
            <div className="absolute top-full left-0 mt-2 w-32 bg-aviation-bg-panel border border-aviation-border-panel rounded-lg shadow-lg z-50">
              {refreshIntervalOptions.map((opt) => (
                <button
                  key={opt.value}
                  onClick={() => {
                    setRefreshInterval(opt.value);
                    setShowRefreshMenu(false);
                  }}
                  className="w-full px-4 py-2.5 text-left text-sm hover:bg-aviation-bg-instrument transition-colors"
                >
                  {opt.label}
                </button>
              ))}
            </div>
          )}
        </div>

        <div className="flex-1" />

        <button
          onClick={() => updateDisplayOptions({ showGrid: !displayOptions.showGrid })}
          className={cn(
            'aviation-btn-ghost px-3 py-2 rounded-lg',
            displayOptions.showGrid && 'text-aviation-amber'
          )}
        >
          <Settings className="w-4 h-4" />
        </button>
      </div>

      <div className="aviation-dv-content flex-1 p-6 overflow-auto">
        <div className="aviation-dv-grid grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="aviation-dv-chart-card bg-aviation-bg-panel border border-aviation-border-panel rounded-xl p-6 shadow-lg">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-medium text-aviation-text-primary">{chartTypeLabels[activeView]}</h3>
              <span className="aviation-dv-badge text-xs px-2 py-1 rounded-full bg-aviation-bg-instrument text-aviation-text-secondary">
                Live
              </span>
            </div>
            <div className="aviation-dv-chart-container">{renderChart()}</div>
          </div>

          <div className="aviation-dv-chart-card bg-aviation-bg-panel border border-aviation-border-panel rounded-xl p-6 shadow-lg">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-medium text-aviation-text-primary">Overview</h3>
              <span className="aviation-dv-badge text-xs px-2 py-1 rounded-full bg-aviation-bg-instrument text-aviation-text-secondary">
                Summary
              </span>
            </div>
            <div className="aviation-dv-chart-container">
              <StreamingLineChart
                data={Array.from({ length: 30 }, (_, i) => ({ timestamp: Date.now() - (30 - i) * 1000, value: Math.sin(i / 5) * 50 + 50 }))}
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default DataVisualizationPage;
