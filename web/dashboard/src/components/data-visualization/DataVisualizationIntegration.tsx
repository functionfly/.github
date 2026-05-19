import React from 'react';
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

export function DataVisualizationIntegration() {
  const { charts, activeView, displayOptions } = useDataVisualizationStore();

  const renderChartByType = (type: string, index: number) => {
    const sampleData = Array.from({ length: 15 }, (_, i) => ({
      x: i,
      y: Math.random() * 100,
      label: `Point ${i}`,
    }));

    const scatterData = Array.from({ length: 30 }, () => ({
      x: Math.random() * 100,
      y: Math.random() * 100,
      z: Math.random() * 30,
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
              { id: '1', label: 'Core', x: 50, y: 50, z: 80 },
              { id: '2', label: 'Edge A', x: 30, y: 30, z: 60 },
              { id: '3', label: 'Edge B', x: 70, y: 40, z: 70 },
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
            data={{
              name: 'Deps',
              value: 100,
              children: [
                { name: 'A', value: 30 },
                { name: 'B', value: 40 },
                { name: 'C', value: 30 },
              ],
            }}
            className="aviation-chart-treemap"
          />
        );
      case 'circular-flow':
        return (
          <CircularFlow
            key={index}
            nodes={[
              { id: '1', label: 'Input' },
              { id: '2', label: 'Process' },
              { id: '3', label: 'Output' },
            ]}
            className="aviation-chart-circular"
          />
        );
      case 'waterfall':
        return (
          <WaterfallChart
            key={index}
            data={[
              { label: 'Start', value: 0, isTotal: true },
              { label: 'A', value: 20 },
              { label: 'B', value: 30 },
              { label: 'End', value: 50, isTotal: true },
            ]}
            className="aviation-chart-waterfall"
          />
        );
      case 'cost-distribution':
        return (
          <CostDistribution
            key={index}
            data={[
              { label: 'CPU', value: 50 },
              { label: 'Memory', value: 30 },
              { label: 'Network', value: 20 },
            ]}
            className="aviation-chart-cost"
          />
        );
      case 'semantic-cluster':
        return (
          <SemanticCluster
            key={index}
            points={Array.from({ length: 20 }, (_, i) => ({
              x: Math.random() * 100,
              y: Math.random() * 100,
              label: `P${i}`,
              cluster: Math.floor(Math.random() * 3),
            }))}
            className="aviation-chart-cluster"
          />
        );
      case 'agent-interaction':
        return (
          <AgentInteractionGraph
            key={index}
            nodes={[
              { id: '1', label: 'Agent', type: 'agent' },
              { id: '2', label: 'Service', type: 'function' },
              { id: '3', label: 'Client', type: 'user' },
            ]}
            edges={[
              { source: '1', target: '2', strength: 2 },
              { source: '2', target: '3', strength: 3 },
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
