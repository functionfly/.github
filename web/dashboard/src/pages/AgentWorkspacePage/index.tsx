import { agentApi } from '@/api/agent';
import { PageGrid, StatusPill } from '@/components/containment';
import { useAgent, useAgentUsage } from '@/hooks/useAgent';
import { normalizeAgentIdentity } from '@/api/agent';
import {
  Activity,
  ArrowLeft,
  Bot,
  Database,
  DollarSign,
  GitBranch,
  Heart,
  Pause,
  Play,
  Settings,
  Shield,
  Sparkles,
  Square,
  Terminal,
  Wrench,
  Zap,
} from 'lucide-react';
import { usePageTitle } from '@/hooks';
import { lazy, Suspense, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useParams } from 'react-router-dom';
import { useAgentWorkspace, type WorkspaceView } from './hooks/useAgentWorkspace';
import { useAgentHealth, useCostRate } from './hooks/useAgentHealth';
import './agent-workspace.css';

const ConsoleView = lazy(() => import('./views/ConsoleView').then(m => ({ default: m.ConsoleView })));
const TracesView = lazy(() => import('./views/TracesView').then(m => ({ default: m.TracesView })));
const HealthView = lazy(() => import('./views/HealthView').then(m => ({ default: m.HealthView })));
const DaemonView = lazy(() => import('./views/DaemonView').then(m => ({ default: m.DaemonView })));
const CostsView = lazy(() => import('./views/CostsView').then(m => ({ default: m.CostsView })));
const ToolsView = lazy(() => import('./views/ToolsView').then(m => ({ default: m.ToolsView })));
const PolicyView = lazy(() => import('./views/PolicyView').then(m => ({ default: m.PolicyView })));
const SwarmView = lazy(() => import('./views/SwarmView').then(m => ({ default: m.SwarmView })));
const MemoryView = lazy(() => import('./views/MemoryView').then(m => ({ default: m.MemoryView })));
const EvolutionView = lazy(() => import('./views/EvolutionView').then(m => ({ default: m.EvolutionView })));
const ConfigView = lazy(() => import('./views/ConfigView').then(m => ({ default: m.ConfigView })));

function ViewSpinner() {
  return (
    <div className="aw-loading">
      <div className="aw-loading__spinner" />
    </div>
  );
}

interface NavItemProps {
  icon: React.ReactNode;
  label: string;
  view: WorkspaceView;
  activeView: WorkspaceView;
  onClick: (v: WorkspaceView) => void;
  badge?: number;
}

function NavItem({ icon, label, view, activeView, onClick, badge }: NavItemProps) {
  return (
    <button
      className={`aw-nav-item ${activeView === view ? 'aw-nav-item--active' : ''}`}
      onClick={() => onClick(view)}
    >
      <span className="aw-nav-item__icon">{icon}</span>
      <span className="aw-nav-item__label">{label}</span>
      {badge !== undefined && badge > 0 && (
        <span className="aw-nav-item__badge">{badge}</span>
      )}
    </button>
  );
}

export function AgentWorkspacePage() {
  const { t } = useTranslation();
  const { id: agentId } = useParams<{ id: string }>();
  const { activeView, setView, rightContext, clearRightContext } = useAgentWorkspace();

  const { data: agentData, isLoading: agentLoading } = useAgent(agentId!);
  const { data: usageData } = useAgentUsage(agentId!);
  const { health, concurrency } = useAgentHealth(agentId!);
  const costRate = useCostRate(agentId!);

  const agent = agentData?.agent ? normalizeAgentIdentity(agentData.agent) : undefined;
  const usage = usageData?.usage;

  usePageTitle(agent ? `${agent.name} / Workspace` : 'Workspace');

  const handleAction = useCallback(async (action: 'start' | 'stop') => {
    if (!agentId) return;
    if (action === 'start') {
      await agentApi.startSession(agentId);
    } else {
      await agentApi.triggerKillSwitch(agentId, 'Manual stop from workspace');
    }
  }, [agentId]);

  if (agentLoading || !agent) {
    return (
      <>
        <PageGrid />
        <div className="aw-shell">
          <div className="aw-loading" style={{ flex: 1 }}>
            <div className="aw-loading__spinner" />
          </div>
        </div>
      </>
    );
  }

  const statusMap: Record<string, 'live' | 'pending' | 'revoked'> = {
    active: 'live',
    suspended: 'revoked',
  };

  const anomalyCount = health?.anomalies?.length ?? 0;

  return (
    <>
      <PageGrid />
      <div className="aw-shell">
        {/* Header */}
        <div className="aw-header">
          <div className="aw-header__left">
            <Link to="/agents" className="aw-header__back">
              <ArrowLeft size={14} />
            </Link>
            <div className="aw-header__icon">
              <Bot />
            </div>
            <span className="aw-header__name">{agent.name}</span>
            <StatusPill status={statusMap[agent.status] ?? 'pending'} label={agent.status} />
          </div>
          <div className="aw-header__right">
            <span className="aw-header__model">{agent.model || 'gpt-4o-mini'}</span>
            <button
              className="aw-header__back"
              onClick={() => handleAction('start')}
              title="Start"
            >
              <Play size={12} />
            </button>
            <button
              className="aw-header__back"
              onClick={() => handleAction('stop')}
              title="Stop"
            >
              <Square size={12} />
            </button>
          </div>
        </div>

        {/* Body */}
        <div className="aw-shell__body">
          {/* Left Nav */}
          <nav className="aw-left-nav">
            <div className="aw-left-nav__section">
              <div className="aw-left-nav__section-label">Operations</div>
              <NavItem icon={<Terminal size={16} />} label="Console" view="console" activeView={activeView} onClick={setView} />
              <NavItem icon={<Activity size={16} />} label="Traces" view="traces" activeView={activeView} onClick={setView} />
              <NavItem icon={<Heart size={16} />} label="Health" view="health" activeView={activeView} onClick={setView} badge={anomalyCount} />
              <NavItem icon={<Zap size={16} />} label="Daemon" view="daemon" activeView={activeView} onClick={setView} />
              <NavItem icon={<DollarSign size={16} />} label="Costs" view="costs" activeView={activeView} onClick={setView} />
            </div>
            <div className="aw-left-nav__section">
              <div className="aw-left-nav__section-label">Configuration</div>
              <NavItem icon={<Wrench size={16} />} label="Tools" view="tools" activeView={activeView} onClick={setView} />
              <NavItem icon={<Shield size={16} />} label="Policy" view="policy" activeView={activeView} onClick={setView} />
              <NavItem icon={<GitBranch size={16} />} label="Swarm" view="swarm" activeView={activeView} onClick={setView} />
              <NavItem icon={<Database size={16} />} label="Memory" view="memory" activeView={activeView} onClick={setView} />
              <NavItem icon={<Sparkles size={16} />} label="Evolution" view="evolution" activeView={activeView} onClick={setView} />
              <NavItem icon={<Settings size={16} />} label="Config" view="config" activeView={activeView} onClick={setView} />
            </div>
          </nav>

          {/* Center Panel */}
          <div className="aw-center">
            <div className="aw-center__scroll">
              <Suspense fallback={<ViewSpinner />}>
                {activeView === 'console' && (
                  <ConsoleView agentId={agentId!} agentName={agent.name} model={agent.model || 'gpt-4o-mini'} />
                )}
                {activeView === 'traces' && (
                  <TracesView agentId={agentId!} setRightContext={() => {}} />
                )}
                {activeView === 'health' && (
                  <HealthView agentId={agentId!} />
                )}
                {activeView === 'daemon' && (
                  <DaemonView agentId={agentId!} />
                )}
                {activeView === 'costs' && (
                  <CostsView agentId={agentId!} />
                )}
                {activeView === 'tools' && (
                  <ToolsView agentId={agentId!} setRightContext={() => {}} />
                )}
                {activeView === 'policy' && (
                  <PolicyView agentId={agentId!} />
                )}
                {activeView === 'swarm' && (
                  <SwarmView agentId={agentId!} setRightContext={() => {}} />
                )}
                {activeView === 'memory' && (
                  <MemoryView agentId={agentId!} />
                )}
                {activeView === 'evolution' && (
                  <EvolutionView agentId={agentId!} />
                )}
                {activeView === 'config' && (
                  <ConfigView agentId={agentId!} />
                )}
              </Suspense>
            </div>
          </div>

          {/* Right Panel */}
          {rightContext && (
            <div className="aw-right">
              <div className="aw-right__header">
                <h3 className="aw-right__title">
                  {rightContext.type === 'execution' && 'Execution Detail'}
                  {rightContext.type === 'trace' && 'Trace Detail'}
                  {rightContext.type === 'tool' && 'Tool Config'}
                  {rightContext.type === 'swarm-node' && 'Agent Detail'}
                  {rightContext.type === 'alert' && 'Alert Detail'}
                  {rightContext.type === 'policy-violation' && 'Violation Detail'}
                </h3>
                <button className="aw-right__close" onClick={clearRightContext}>×</button>
              </div>
              <div className="aw-right__scroll">
                <div className="aw-empty">
                  <span className="aw-empty__title">{rightContext.type}</span>
                  <span className="aw-empty__desc">ID: {(rightContext as { type: string; id: string }).id}</span>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Status Bar */}
        <div className="aw-status-bar">
          <div className="aw-status-bar__item">
            <span className={`aw-status-bar__dot ${agent.is_daemon_running ? 'aw-status-bar__dot--live' : 'aw-status-bar__dot--off'}`} />
            Daemon {agent.is_daemon_running ? 'Running' : 'Stopped'}
          </div>
          <span className="aw-status-bar__separator" />
          <div className="aw-status-bar__item">
            <span className="aw-status-bar__dot aw-status-bar__dot--live" />
            {concurrency?.active_executions ?? 0}/{concurrency?.max_concurrent ?? '—'} concurrent
          </div>
          <span className="aw-status-bar__separator" />
          <div className="aw-status-bar__item">
            ${costRate.toFixed(4)}/min
          </div>
          <span className="aw-status-bar__separator" />
          <div className="aw-status-bar__item">
            <span className={`aw-status-bar__dot ${
              health?.status === 'healthy' ? 'aw-status-bar__dot--live' :
              health?.status === 'degraded' ? 'aw-status-bar__dot--warn' :
              health?.status === 'critical' ? 'aw-status-bar__dot--error' :
              'aw-status-bar__dot--off'
            }`} />
            {health?.status ?? 'Unknown'}
          </div>
          {anomalyCount > 0 && (
            <div className="aw-status-bar__alert">
              <Shield size={12} />
              {anomalyCount} alert{anomalyCount !== 1 ? 's' : ''}
            </div>
          )}
        </div>
      </div>
    </>
  );
}
