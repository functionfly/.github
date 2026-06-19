import { useParams } from 'react-router-dom';
import { useAdaptiveUXStore } from '@/stores/adaptiveUXStore';
import { AdaptiveUXIntegration } from '@/components/adaptive-ux/AdaptiveUXIntegration';
import {
  AdaptiveComplexityLayer,
  ContextAwareToolbar,
  PredictiveActionBar,
  SmartWorkspaceRecommendations,
  LearningModeOverlay,
  BeginnerSimplificationView,
  ExpertSystemView,
  CognitiveLoadBalancer,
  AttentionFocusOverlay,
  WorkflowOptimizationHints,
  AdaptiveLayoutEngine,
} from '@functionfly/ui-adaptive-ux';
import {
  selectVisibleHints,
  selectActiveRecommendations,
  selectCurrentComplexity,
  selectAdaptationScore,
} from '@/stores/adaptiveUXStore';
import { cn } from '@/lib/utils';
import {
  Layers,
  PanelLeft,
  Sparkles,
  LayoutDashboard,
  GraduationCap,
  Zap,
  Brain,
  Eye,
  Lightbulb,
  Settings2,
} from 'lucide-react';
type LocalUserSkillLevel = 'beginner' | 'intermediate' | 'expert';

const viewTabs = [
  { id: 'overview', label: 'Overview', icon: LayoutDashboard },
  { id: 'complexity', label: 'Complexity', icon: Layers },
  { id: 'toolbar', label: 'Toolbar', icon: PanelLeft },
  { id: 'predictions', label: 'Predictions', icon: Sparkles },
  { id: 'workspace', label: 'Workspace', icon: LayoutDashboard },
  { id: 'learning', label: 'Learning', icon: GraduationCap },
  { id: 'beginner', label: 'Beginner', icon: GraduationCap },
  { id: 'expert', label: 'Expert', icon: Zap },
  { id: 'cognitive', label: 'Cognitive', icon: Brain },
  { id: 'attention', label: 'Attention', icon: Eye },
  { id: 'hints', label: 'Hints', icon: Lightbulb },
] as const;

type ViewTab = (typeof viewTabs)[number]['id'];

const skillLevels: LocalUserSkillLevel[] = ['beginner', 'intermediate', 'expert'];

function OverviewView() {
  const store = useAdaptiveUXStore();
  const visibleHints = selectVisibleHints(useAdaptiveUXStore.getState());
  const activeRecommendations = selectActiveRecommendations(useAdaptiveUXStore.getState());
  const adaptationScore = selectAdaptationScore(useAdaptiveUXStore.getState());

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
        <div className="rounded-lg border border-aviation-700 bg-aviation-800/50 p-4">
          <div className="text-2xl font-bold text-text-primary">{adaptationScore}%</div>
          <div className="text-sm text-text-muted">Adaptation Score</div>
        </div>
        <div className="rounded-lg border border-aviation-700 bg-aviation-800/50 p-4">
          <div className="text-2xl font-bold text-text-primary capitalize">{store.userSkillLevel}</div>
          <div className="text-sm text-text-muted">Skill Level</div>
        </div>
        <div className="rounded-lg border border-aviation-700 bg-aviation-800/50 p-4">
          <div className="text-2xl font-bold text-text-primary">{activeRecommendations.length}</div>
          <div className="text-sm text-text-muted">Active Recommendations</div>
        </div>
      </div>
      <AdaptiveUXIntegration />
    </div>
  );
}

function ComplexityView() {
  const store = useAdaptiveUXStore();
  const currentComplexity = selectCurrentComplexity(useAdaptiveUXStore.getState());

  return (
    <div className="space-y-6">
      <div className="rounded-lg border border-aviation-700 bg-aviation-800/50 p-6">
        <h3 className="mb-4 text-lg font-semibold text-text-primary">Adaptive Complexity Layer</h3>
        <AdaptiveComplexityLayer
          userSkillLevel={store.userSkillLevel}
          currentContext={store.activeView}
          onComplexityChange={(config) => {
            console.log('Complexity changed:', config);
          }}
          className="bg-aviation-900/50 rounded-lg p-4"
        />
      </div>
    </div>
  );
}

function ToolbarView() {
  const store = useAdaptiveUXStore();

  return (
    <div className="space-y-6">
      <div className="rounded-lg border border-aviation-700 bg-aviation-800/50 p-6">
        <h3 className="mb-4 text-lg font-semibold text-text-primary">Context-Aware Toolbar</h3>
        <ContextAwareToolbar
          availableActions={store.toolbarActions}
          currentContext={store.activeView}
          onActionExecute={(action) => {
            console.log('Action executed:', action);
          }}
          onContextChange={(context) => {
            console.log('Context changed:', context);
          }}
          className="bg-aviation-900/50 rounded-lg p-4"
        />
      </div>
    </div>
  );
}

function PredictionsView() {
  const store = useAdaptiveUXStore();

  return (
    <div className="space-y-6">
      <div className="rounded-lg border border-aviation-700 bg-aviation-800/50 p-6">
        <h3 className="mb-4 text-lg font-semibold text-text-primary">Predictive Action Bar</h3>
        <PredictiveActionBar
          predictions={store.predictedActions}
          onActionSelect={(action) => {
            console.log('Action selected:', action);
          }}
          onFeedback={(actionId, helpful) => {
            console.log('Feedback:', actionId, helpful);
          }}
          maxVisible={5}
          className="bg-aviation-900/50 rounded-lg p-4"
        />
      </div>
    </div>
  );
}

function WorkspaceView() {
  const store = useAdaptiveUXStore();
  const activeRecommendations = selectActiveRecommendations(useAdaptiveUXStore.getState());

  return (
    <div className="space-y-6">
      <div className="rounded-lg border border-aviation-700 bg-aviation-800/50 p-6">
        <h3 className="mb-4 text-lg font-semibold text-text-primary">Smart Workspace Recommendations</h3>
        <SmartWorkspaceRecommendations
          recommendations={activeRecommendations}
          onAccept={(id) => store.acceptRecommendation(id)}
          onDismiss={(id) => store.dismissHint(id)}
          onApplyInstant={(id) => store.acceptRecommendation(id)}
          className="bg-aviation-900/50 rounded-lg p-4"
        />
      </div>
    </div>
  );
}

function LearningView() {
  const store = useAdaptiveUXStore();

  return (
    <div className="space-y-6">
      <div className="rounded-lg border border-aviation-700 bg-aviation-800/50 p-6">
        <h3 className="mb-4 text-lg font-semibold text-text-primary">Learning Mode Overlay</h3>
        <LearningModeOverlay
          isActive={store.preferences.learningModeEnabled}
          phase={store.learningPhase}
          observedPatterns={store.observedPatterns}
          onPatternLearn={(pattern) => store.recordPattern(pattern)}
          onDismiss={() => store.updatePreferences({ learningModeEnabled: false })}
          className="bg-aviation-900/50 rounded-lg p-4"
        />
      </div>
    </div>
  );
}

function BeginnerView() {
  const store = useAdaptiveUXStore();

  return (
    <div className="space-y-6">
      <div className="rounded-lg border border-aviation-700 bg-aviation-800/50 p-6">
        <h3 className="mb-4 text-lg font-semibold text-text-primary">Beginner Simplification View</h3>
        <BeginnerSimplificationView
          enabledFeatures={store.simplifiedFeatures}
          onFeatureToggle={(id, enabled) => {
            const features = store.simplifiedFeatures.map((f) =>
              f.id === id ? { ...f, enabled } : f
            );
            store.setSimplifiedFeatures(features);
          }}
          onGuidedTourStart={() => {
            console.log('Starting guided tour');
          }}
          onHelpRequest={() => {
            console.log('Help requested');
          }}
          className="bg-aviation-900/50 rounded-lg p-4"
        />
      </div>
    </div>
  );
}

function ExpertView() {
  const store = useAdaptiveUXStore();

  return (
    <div className="space-y-6">
      <div className="rounded-lg border border-aviation-700 bg-aviation-800/50 p-6">
        <h3 className="mb-4 text-lg font-semibold text-text-primary">Expert System View</h3>
        <ExpertSystemView
          availableFeatures={store.expertFeatures}
          onQuickAccess={(featureId) => {
            console.log('Quick access:', featureId);
          }}
          onMacroCreate={(name, steps) => {
            console.log('Macro created:', name, steps);
          }}
          onCustomization={(featureId, config) => {
            console.log('Customization:', featureId, config);
          }}
          className="bg-aviation-900/50 rounded-lg p-4"
        />
      </div>
    </div>
  );
}

function CognitiveView() {
  const store = useAdaptiveUXStore();

  return (
    <div className="space-y-6">
      <div className="rounded-lg border border-aviation-700 bg-aviation-800/50 p-6">
        <h3 className="mb-4 text-lg font-semibold text-text-primary">Cognitive Load Balancer</h3>
        <CognitiveLoadBalancer
          metrics={store.cognitiveLoad}
          onBreakSuggestion={(type) => {
            console.log('Break suggestion:', type);
          }}
          onLoadReduction={(actions) => {
            console.log('Load reduction actions:', actions);
          }}
          onFocusModeToggle={(enabled) => {
            store.updatePreferences({ uiDensity: enabled ? 'compact' : 'normal' });
          }}
          className="bg-aviation-900/50 rounded-lg p-4"
        />
      </div>
    </div>
  );
}

function AttentionView() {
  const store = useAdaptiveUXStore();

  return (
    <div className="space-y-6">
      <div className="rounded-lg border border-aviation-700 bg-aviation-800/50 p-6">
        <h3 className="mb-4 text-lg font-semibold text-text-primary">Attention Focus Overlay</h3>
        <AttentionFocusOverlay
          attentionState={store.attentionState}
          highlights={store.focusHighlights}
          onHighlightClick={(highlight) => {
            console.log('Highlight clicked:', highlight);
          }}
          onStateChange={(state) => store.setAttentionState(state)}
          className="bg-aviation-900/50 rounded-lg p-4"
        />
      </div>
    </div>
  );
}

function HintsView() {
  const store = useAdaptiveUXStore();
  const visibleHints = selectVisibleHints(useAdaptiveUXStore.getState());

  return (
    <div className="space-y-6">
      <div className="rounded-lg border border-aviation-700 bg-aviation-800/50 p-6">
        <h3 className="mb-4 text-lg font-semibold text-text-primary">Workflow Optimization Hints</h3>
        <WorkflowOptimizationHints
          hints={visibleHints}
          onHintApply={(hint) => {
            console.log('Hint applied:', hint);
          }}
          onHintDismiss={(hintId) => store.dismissHint(hintId)}
          onHintSnooze={(hintId, duration) => store.snoozeHint(hintId, duration)}
          className="bg-aviation-900/50 rounded-lg p-4"
        />
      </div>
    </div>
  );
}

export function AdaptiveUXPage() {
  const { panel } = useParams<{ panel?: string }>();
  const store = useAdaptiveUXStore();
  
  const currentView = (panel as ViewTab) || store.activeView || 'overview';

  const renderView = () => {
    switch (currentView) {
      case 'complexity':
        return <ComplexityView />;
      case 'toolbar':
        return <ToolbarView />;
      case 'predictions':
        return <PredictionsView />;
      case 'workspace':
        return <WorkspaceView />;
      case 'learning':
        return <LearningView />;
      case 'beginner':
        return <BeginnerView />;
      case 'expert':
        return <ExpertView />;
      case 'cognitive':
        return <CognitiveView />;
      case 'attention':
        return <AttentionView />;
      case 'hints':
        return <HintsView />;
      default:
        return <OverviewView />;
    }
  };

  return (
    <div className="container mx-auto max-w-7xl space-y-6 p-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-text-primary">Adaptive UX</h1>
          <p className="text-sm text-text-muted">
            AI-powered adaptive user experience components
          </p>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-sm text-text-muted">Skill Level:</span>
          <div className="flex rounded-lg border border-aviation-700 bg-aviation-800">
            {skillLevels.map((level) => (
              <button
                key={level}
                onClick={() => store.setSkillLevel(level)}
                className={cn(
                  'px-3 py-1.5 text-sm capitalize transition-colors',
                  store.userSkillLevel === level
                    ? 'bg-aviation-600 text-white'
                    : 'text-text-secondary hover:text-text-primary hover:bg-aviation-700'
                )}
              >
                {level}
              </button>
            ))}
          </div>
        </div>
      </div>

      <div className="border-b border-aviation-700">
        <nav className="-mb-px flex space-x-1 overflow-x-auto">
          {viewTabs.map((tab) => {
            const Icon = tab.icon;
            const isActive = currentView === tab.id;
            return (
              <button
                key={tab.id}
                onClick={() => store.setActiveView(tab.id as ViewTab)}
                className={cn(
                  'flex items-center gap-2 whitespace-nowrap border-b-2 px-4 py-3 text-sm transition-colors',
                  isActive
                    ? 'border-aviation-500 text-aviation-400'
                    : 'border-transparent text-text-muted hover:border-aviation-600 hover:text-text-primary'
                )}
              >
                <Icon className="h-4 w-4" />
                {tab.label}
              </button>
            );
          })}
        </nav>
      </div>

      <div className="min-h-[400px]">{renderView()}</div>
    </div>
  );
}

export default AdaptiveUXPage;
