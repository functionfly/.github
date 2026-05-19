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
  useAdaptiveUXStore,
  selectVisibleHints,
  selectActiveRecommendations,
  selectCurrentComplexity,
  selectAdaptationScore,
} from '@/stores/adaptiveUXStore';
import { cn } from '@/lib/utils';

export function AdaptiveUXIntegration() {
  const store = useAdaptiveUXStore();
  const visibleHints = selectVisibleHints(store);
  const activeRecommendations = selectActiveRecommendations(store);
  const currentComplexity = selectCurrentComplexity(store);
  const adaptationScore = selectAdaptationScore(store);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold text-text-primary">Adaptive UX</h3>
          <p className="text-sm text-text-secondary">
            Adaptation Score: {adaptationScore}%
          </p>
        </div>
        <div className="text-sm text-text-muted">
          Skill Level: <span className="capitalize">{store.userSkillLevel}</span>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <div className="rounded-lg border border-aviation-700 bg-aviation-800/50 p-4">
          <h4 className="mb-3 text-sm font-medium text-text-primary">Complexity Layers</h4>
          <AdaptiveComplexityLayer
            userSkillLevel={store.userSkillLevel}
            currentContext={store.activeView}
            onComplexityChange={(config) => {
              console.log('Complexity changed:', config);
            }}
            className="bg-aviation-900/50 rounded-lg p-3"
          />
        </div>

        <div className="rounded-lg border border-aviation-700 bg-aviation-800/50 p-4">
          <h4 className="mb-3 text-sm font-medium text-text-primary">Context Toolbar</h4>
          <ContextAwareToolbar
            availableActions={store.toolbarActions}
            currentContext={store.activeView}
            onActionExecute={(action) => {
              console.log('Action executed:', action);
            }}
            className="bg-aviation-900/50 rounded-lg p-3"
          />
        </div>

        <div className="rounded-lg border border-aviation-700 bg-aviation-800/50 p-4">
          <h4 className="mb-3 text-sm font-medium text-text-primary">Predicted Actions</h4>
          <PredictiveActionBar
            predictions={store.predictedActions}
            onActionSelect={(action) => {
              console.log('Action selected:', action);
            }}
            onFeedback={(actionId, helpful) => {
              console.log('Feedback:', actionId, helpful);
            }}
            maxVisible={5}
            className="bg-aviation-900/50 rounded-lg p-3"
          />
        </div>

        <div className="rounded-lg border border-aviation-700 bg-aviation-800/50 p-4">
          <h4 className="mb-3 text-sm font-medium text-text-primary">Workspace Recommendations</h4>
          <SmartWorkspaceRecommendations
            recommendations={activeRecommendations}
            onAccept={(id) => store.acceptRecommendation(id)}
            onDismiss={(id) => store.dismissHint(id)}
            onApplyInstant={(id) => store.acceptRecommendation(id)}
            className="bg-aviation-900/50 rounded-lg p-3"
          />
        </div>

        <div className="rounded-lg border border-aviation-700 bg-aviation-800/50 p-4">
          <h4 className="mb-3 text-sm font-medium text-text-primary">Learning Mode</h4>
          <LearningModeOverlay
            isActive={store.preferences.learningModeEnabled}
            phase={store.learningPhase}
            observedPatterns={store.observedPatterns}
            onPatternLearn={(pattern) => store.recordPattern(pattern)}
            onDismiss={() => store.updatePreferences({ learningModeEnabled: false })}
            className="bg-aviation-900/50 rounded-lg p-3"
          />
        </div>

        <div className="rounded-lg border border-aviation-700 bg-aviation-800/50 p-4">
          <h4 className="mb-3 text-sm font-medium text-text-primary">Cognitive Load</h4>
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
            className="bg-aviation-900/50 rounded-lg p-3"
          />
        </div>

        <div className="rounded-lg border border-aviation-700 bg-aviation-800/50 p-4">
          <h4 className="mb-3 text-sm font-medium text-text-primary">Attention Focus</h4>
          <AttentionFocusOverlay
            attentionState={store.attentionState}
            highlights={store.focusHighlights}
            onHighlightClick={(highlight) => {
              console.log('Highlight clicked:', highlight);
            }}
            onStateChange={(state) => store.setAttentionState(state)}
            className="bg-aviation-900/50 rounded-lg p-3"
          />
        </div>

        <div className="rounded-lg border border-aviation-700 bg-aviation-800/50 p-4">
          <h4 className="mb-3 text-sm font-medium text-text-primary">Workflow Hints</h4>
          <WorkflowOptimizationHints
            hints={visibleHints}
            onHintApply={(hint) => {
              console.log('Hint applied:', hint);
            }}
            onHintDismiss={(hintId) => store.dismissHint(hintId)}
            onHintSnooze={(hintId, duration) => store.snoozeHint(hintId, duration)}
            className="bg-aviation-900/50 rounded-lg p-3"
          />
        </div>

        {store.userSkillLevel === 'beginner' && (
          <div className="rounded-lg border border-aviation-700 bg-aviation-800/50 p-4 lg:col-span-2">
            <h4 className="mb-3 text-sm font-medium text-text-primary">Beginner View</h4>
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
              className="bg-aviation-900/50 rounded-lg p-3"
            />
          </div>
        )}

        {store.userSkillLevel === 'expert' && (
          <div className="rounded-lg border border-aviation-700 bg-aviation-800/50 p-4 lg:col-span-2">
            <h4 className="mb-3 text-sm font-medium text-text-primary">Expert View</h4>
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
              className="bg-aviation-900/50 rounded-lg p-3"
            />
          </div>
        )}

        <div className="rounded-lg border border-aviation-700 bg-aviation-800/50 p-4 lg:col-span-2">
          <h4 className="mb-3 text-sm font-medium text-text-primary">Adaptive Layout</h4>
          <AdaptiveLayoutEngine
            configurations={[]}
            currentConfig={null}
            onConfigSelect={(config) => {
              console.log('Layout config selected:', config);
            }}
            onAutoAdjust={(skillLevel) => {
              store.setSkillLevel(skillLevel);
            }}
            className="bg-aviation-900/50 rounded-lg p-3"
          />
        </div>
      </div>
    </div>
  );
}
