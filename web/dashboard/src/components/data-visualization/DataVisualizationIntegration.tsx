import React from 'react';
import {
  StreamingLineChart,
  RealtimeScatterPlot,
  ThreeDTopologyChart,
  ExecutionSunburst,
  DependencyTreemap,
  CircularFlow,
  RuntimeWaterfallChart,
  CostDistribution,
  SemanticCluster,
  AgentInteractionGraph,
} from '@functionfly/ui-data-visualization';
import { useDataVisualizationStore } from '@/stores/dataVisualizationStore';
import { cn } from '@/lib/utils';

export function DataVisualizationIntegration() {
  const { charts, activeView, displayOptions } = useDataVisualizationStore();

  const renderChartByType = (type: string, index: number) => {
    const sampleData = Array.from({ length: 15 }, (_, i) => ({
      timestamp: Date.now() - (15 - i) * 1000,
      value: Math.random() * 100,
      label: `Point ${i}`,
    }));

    const scatterData = Array.from({ length: 30 }, () => ({
      x: Math.random() * 100,
      y: Math.random() * 100,
      size: Math.random() * 30,
    }));

    switch (type) {
      case 'streaming-line':
        return <StreamingLineChart key={index} data={sampleData} className="aviation-chart-streaming" />;
      case 'realtime-scatter':
        return <RealtimeScatterPlot key={index} data={scatterData} className="aviation-chart-scatter" />;
      case '3d-topology':
        return (
          <ThreeDTopologyChart
            key={index}
            nodes={[
              { id: '1', label: 'Core', depth: 0 },
              { id: '2', label: 'Edge A', depth: 1 },
              { id: '3', label: 'Edge B', depth: 1 },
            ]}
            className="aviation-chart-topology"
          />
        );
      case 'execution-sunburst':
        return (
          <ExecutionSunburst
            key={index}
            data={{ name: 'Execution', value: 100, children: [{ name: 'Phase 1', value: 40 }, { name: 'Phase 2', value: 60 }] }}
            className="aviation-chart-sunburst"
          />
        );
      case 'dependency-treemap':
        return (
          <DependencyTreemap
            key={index}
            data={[
              { id: '1', name: 'Deps', value: 100, children: [
                { id: '2', name: 'A', value: 30 },
                { id: '3', name: 'B', value: 40 },
                { id: '4', name: 'C', value: 30 },
              ]},
            ]}
            className="aviation-chart-treemap"
          />
        );
      case 'circular-flow':
        return (
          <CircularFlow
            key={index}
            nodes={[
              { id: '1', label: 'Input', value: 100 },
              { id: '2', label: 'Process', value: 100 },
              { id: '3', label: 'Output', value: 100 },
            ]}
            connections={[
              { source: '1', target: '2', value: 1 },
              { source: '2', target: '3', value: 1 },
            ]}
            className="aviation-chart-circular"
          />
        );
      case 'waterfall':
        return (
          <RuntimeWaterfallChart
            key={index}
            steps={[
              { name: 'Start', start: 0, end: 0 },
              { name: 'A', start: 0, end: 20 },
              { name: 'B', start: 20, end: 50 },
              { name: 'End', start: 50, end: 50 },
            ]}
            className="aviation-chart-waterfall"
          />
        );
      case 'cost-distribution':
        return (
          <CostDistribution
            key={index}
            data={[
              { category: 'CPU', value: 50 },
              { category: 'Memory', value: 30 },
              { category: 'Network', value: 20 },
            ]}
            className="aviation-chart-cost"
          />
        );
      case 'semantic-cluster':
        return (
          <SemanticCluster
            key={index}
            points={Array.from({ length: 20 }, (_, i) => ({
              id: `p${i}`,
              x: Math.random() * 100,
              y: Math.random() * 100,
              label: `P${i}`,
              cluster: String(Math.floor(Math.random() * 3)),
            }))}
            className="aviation-chart-cluster"
          />
        );
      case 'agent-interaction':
        return (
          <AgentInteractionGraph
            key={index}
            nodes={[
              { id: '1', label: 'Agent', type: 'agent', connections: ['2'] },
              { id: '2', label: 'Service', type: 'function', connections: ['3'] },
              { id: '3', label: 'Client', type: 'user', connections: [] },
            ]}
            edges={[
              { source: '1', target: '2', weight: 2 },
              { source: '2', target: '3', weight: 3 },
            ]}
            className="aviation-chart-agent"
          />
        );
      default:
        return <StreamingLineChart key={index} data={sampleData} />;
    }
  };

  return (
    <div className="aviation-dv-integration">
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 p-4">
        {charts.length > 0 ? (
          charts.map((chart, index) => (
            <div
              key={chart.id}
              className={cn(
                'aviation-dv-card bg-aviation-bg-panel border border-aviation-border-panel rounded-xl p-4',
                'hover:border-aviation-amber/30 transition-colors'
              )}
            >
              <div className="flex items-center justify-between mb-3">
                <span className="text-sm font-medium text-aviation-text-primary capitalize">
                  {chart.type.replace('-', ' ')}
                </span>
                <span className="aviation-dv-status-dot w-2 h-2 rounded-full bg-aviation-cyan" />
              </div>
              <div className="aviation-dv-chart-wrapper">{renderChartByType(chart.type, index)}</div>
            </div>
          ))
        ) : (
          <div className="col-span-full">
            <div className="aviation-dv-empty-state flex flex-col items-center justify-center py-16">
              <div className="aviation-dv-empty-icon w-16 h-16 mb-4 rounded-xl bg-aviation-bg-instrument flex items-center justify-center">
                <svg
                  className="w-8 h-8 text-aviation-text-muted"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={1.5}
                    d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
                  />
                </svg>
              </div>
              <h3 className="text-lg font-medium text-aviation-text-primary mb-2">No Charts Configured</h3>
              <p className="text-sm text-aviation-text-secondary text-center max-w-sm">
                Add charts to your dashboard to visualize your data in real-time
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

export default DataVisualizationIntegration;
