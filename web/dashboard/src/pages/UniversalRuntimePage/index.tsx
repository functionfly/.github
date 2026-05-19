import { useUniversalRuntimeStore } from '@/stores/universalRuntimeStore';
import { cn } from '@/lib/utils';
import {
  Activity,
  Box,
  Cpu,
  Globe,
  Layers,
  Network,
  Settings,
  Zap,
} from 'lucide-react';

const VIEW_TABS = [
  { id: 'overview', label: 'Overview', icon: Activity },
  { id: 'wasm', label: 'WASM Execution', icon: Box },
  { id: 'gpu', label: 'GPU Kernel', icon: Cpu },
  { id: 'serverless', label: 'Serverless', icon: Zap },
  { id: 'browser', label: 'Browser Agent', icon: Globe },
  { id: 'edge', label: 'Edge Runtime', icon: Globe },
  { id: 'orchestrator', label: 'Hybrid Orchestrator', icon: Layers },
  { id: 'topology', label: 'Cross-Cloud', icon: Network },
  { id: 'routing', label: 'Model Routing', icon: Layers },
  { id: 'inference', label: 'Inference', icon: Settings },
] as const;

const RuntimePlaceholder = ({ view }: { view: string }) => (
  <div className="aviation-panel p-6">
    <div className="flex flex-col items-center justify-center h-[400px] text-center">
      <div className="aviation-icon-container mb-4">
        <Settings className="w-8 h-8 text-aviation-cyan" />
      </div>
      <h3 className="text-lg font-semibold text-aviation-text-primary mb-2">
        {view} Runtime
      </h3>
      <p className="text-sm text-aviation-text-secondary max-w-md">
        The {view.toLowerCase()} runtime interface is being loaded from the @functionfly/ui-universal-runtime package.
        Components will render here once the package is fully populated.
      </p>
    </div>
  </div>
);

export function UniversalRuntimePage() {
  const { activeView, setActiveView, executionMode, setExecutionMode, runtimes, metrics } =
    useUniversalRuntimeStore();

  const activeTab = VIEW_TABS.find((t) => t.id === activeView) || VIEW_TABS[0];

  return (
    <div className="aviation-layout-container p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="aviation-heading-1">Universal Runtime</h1>
          <p className="aviation-subtitle">
            Manage WebAssembly, GPU, Serverless, Browser, Edge, and Hybrid execution environments
          </p>
        </div>
        <div className="aviation-metric-card">
          <span className="aviation-metric-label">Total Throughput</span>
          <span className="aviation-metric-value">{metrics.throughput.toFixed(1)}</span>
          <span className="aviation-metric-unit">req/s</span>
        </div>
      </div>

      <div className="flex gap-2 overflow-x-auto pb-2">
        {VIEW_TABS.map((tab) => {
          const Icon = tab.icon;
          return (
            <button
              key={tab.id}
              onClick={() => setActiveView(tab.id)}
              className={cn(
                'aviation-tab',
                activeView === tab.id && 'active'
              )}
            >
              <Icon className="w-4 h-4 mr-2" />
              {tab.label}
              {tab.id !== 'overview' && (
                <span className="aviation-tab-badge">
                  {runtimes.filter((r) => r.type === tab.id).length || ''}
                </span>
              )}
            </button>
          );
        })}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <div className="aviation-panel p-4">
          <h3 className="text-sm font-semibold text-aviation-text-secondary mb-3">Execution Mode</h3>
          <div className="space-y-2">
            {(['wasm', 'gpu', 'serverless', 'browser', 'edge', 'hybrid'] as const).map((mode) => (
              <button
                key={mode}
                onClick={() => setExecutionMode(mode)}
                className={cn(
                  'w-full text-left px-3 py-2 rounded-lg transition-colors',
                  'aviation-list-item',
                  executionMode === mode && 'active'
                )}
              >
                <span className="capitalize">{mode}</span>
                <span className="aviation-badge ml-2">
                  {runtimes.filter((r) => r.type === mode).length}
                </span>
              </button>
            ))}
          </div>
        </div>

        <div className="aviation-panel p-4 lg:col-span-2">
          <h3 className="text-sm font-semibold text-aviation-text-secondary mb-3">Runtime Status</h3>
          <div className="grid grid-cols-2 gap-3">
            <div className="aviation-stat">
              <span className="aviation-stat-label">Ready</span>
              <span className="aviation-stat-value">
                {runtimes.filter((r) => r.status === 'ready').length}
              </span>
            </div>
            <div className="aviation-stat">
              <span className="aviation-stat-label">Active</span>
              <span className="aviation-stat-value">
                {runtimes.filter((r) => r.status === 'active').length}
              </span>
            </div>
            <div className="aviation-stat">
              <span className="aviation-stat-label">Total Requests</span>
              <span className="aviation-stat-value">{metrics.totalRequests}</span>
            </div>
            <div className="aviation-stat">
              <span className="aviation-stat-label">Avg Latency</span>
              <span className="aviation-stat-value">{metrics.averageLatency.toFixed(0)}ms</span>
            </div>
          </div>
        </div>
      </div>

      <div className="aviation-panel">
        <div className="p-4 border-b border-aviation-border-panel">
          <div className="flex items-center gap-2">
            <activeTab.icon className="w-5 h-5 text-aviation-cyan" />
            <h2 className="text-lg font-semibold text-aviation-text-primary">{activeTab.label}</h2>
          </div>
        </div>
        <div className="p-4">
          <RuntimePlaceholder view={activeTab.label} />
        </div>
      </div>
    </div>
  );
}

export default UniversalRuntimePage;
