import React, { useState } from 'react';
import {
  StreamingLineChart,
  RealtimeScatterPlot,
  ThreeDTopologyChart,
  ExecutionSunburst,
  DependencyTreemap,
  CircularFlow,
  WaterfallChart,
  CostDistribution,
  SemanticCluster,
  AgentInteractionGraph,
} from '@functionfly/ui-data-visualization';
import { useDataVisualizationStore } from '@/stores/dataVisualizationStore';
import { cn } from '@/lib/utils';
import { RefreshCw, Settings, ChevronDown, BarChart3, ScatterChart, Globe, CircleDot, Grid3X3, TrendingDown, PieChart, Sparkles, Network } from 'lucide-react';

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
      x: i,
      y: Math.random() * 100,
      label: `Point ${i}`,
    }));

    const scatterSampleData = Array.from({ length: 50 }, () => ({
      x: Math.random() * 100,
      y: Math.random() * 100,
      z: Math.random() * 50,
      label: `Cluster ${Math.floor(Math.random() * 5)}`,
    }));

    const sunburstSampleData = {
      name: 'Root',
      value: 100,
      children: [
        { name: 'A', value: 40 },
        { name: 'B', value: 30 },
        { name: 'C', value: 30 },
      ],
    };

    const treemapSampleData = {
      name: 'Dependencies',
      value: 100,
      children: [
        { name: 'core', value: 30 },
        { name: 'utils', value: 25 },
        { name: 'api', value: 20 },
        { name: 'ui', value: 15 },
        { name: 'auth', value: 10 },
      ],
    };

    const waterfallSampleData = [
      { label: 'Start', value: 0, isTotal: true },
      { label: 'CPU', value: 25 },
      { label: 'Memory', value: 15 },
      { label: 'Network', value: 10 },
      { label: 'Storage', value: 20 },
      { label: 'End', value: 70, isTotal: true },
    ];

    const costSampleData = [
      { label: 'Compute', value: 45 },
      { label: 'Storage', value: 25 },
      { label: 'Network', value: 15 },
      { label: 'API', value: 10 },
      { label: 'Other', value: 5 },
    ];

    const semanticSampleData = Array.from({ length: 30 }, (_, i) => ({
      x: Math.random() * 100,
      y: Math.random() * 100,
      label: `Item ${i}`,
      cluster: Math.floor(Math.random() * 5),
    }));

    const agentNodes = [
      { id: '1', label: 'Agent A', type: 'agent' as const },
      { id: '2', label: 'Func B', type: 'function' as const },
      { id: '3', label: 'User C', type: 'user' as const },
      { id: '4', label: 'Agent D', type: 'agent' as const },
    ];

    const agentEdges = [
      { source: '1', target: '2', strength: 3 },
      { source: '2', target: '3', strength: 2 },
      { source: '3', target: '1', strength: 4 },
      { source: '1', target: '4', strength: 2 },
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
              { id: '1', label: 'Node A', x: 30, y: 40, z: 60 },
              { id: '2', label: 'Node B', x: 70, y: 30, z: 40 },
              { id: '3', label: 'Node C', x: 50, y: 70, z: 80 },
              { id: '4', label: 'Node D', x: 20, y: 20, z: 50 },
            ]}
          />
        );
      case 'execution-sunburst':
        return <ExecutionSunburst data={sunburstSampleData} />;
      case 'dependency-treemap':
        return <DependencyTreemap data={treemapSampleData} />;
      case 'circular-flow':
        return (
          <CircularFlow
            nodes={[
              { id: '1', label: 'API' },
              { id: '2', label: 'Auth' },
              { id: '3', label: 'DB' },
              { id: '4', label: 'Cache' },
            ]}
          />
        );
      case 'waterfall':
        return <WaterfallChart data={waterfallSampleData} />;
      case 'cost-distribution':
        return <CostDistribution data={costSampleData} />;
      case 'semantic-cluster':
        return <SemanticCluster points={semanticSampleData} />;
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
                data={Array.from({ length: 30 }, (_, i) => ({ x: i, y: Math.sin(i / 5) * 50 + 50 }))}
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default DataVisualizationPage;
