import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import type {
  UserSkillLevel,
  ComplexityLevel,
  ContextHint,
  RecommendationType,
  AttentionState,
  LearningPhase,
  ComplexityConfig,
  ToolbarAction,
  PredictedAction,
  WorkspaceRecommendation,
  LearningPattern,
  CognitiveMetrics,
  FocusHighlight,
  WorkflowHint,
} from '@functionfly/ui-adaptive-ux';

interface AdaptiveUXState {
  userSkillLevel: UserSkillLevel;
  complexityLayers: Record<ComplexityLevel, boolean>;
  contextHints: ContextHint[];
  recommendations: WorkspaceRecommendation[];
  learningPhase: LearningPhase;
  activeView: 'overview' | 'complexity' | 'toolbar' | 'predictions' | 'workspace' | 'learning' | 'beginner' | 'expert' | 'cognitive' | 'attention' | 'hints';
  cognitiveLoad: CognitiveMetrics;
  attentionState: AttentionState;
  dismissedHints: string[];
  preferences: {
    uiDensity: 'compact' | 'normal' | 'spacious';
    animationLevel: 'minimal' | 'standard' | 'enhanced';
    autoAdjust: boolean;
    learningModeEnabled: boolean;
  };
  toolbarActions: ToolbarAction[];
  predictedActions: PredictedAction[];
  observedPatterns: LearningPattern[];
  focusHighlights: FocusHighlight[];
  availableHints: WorkflowHint[];
  simplifiedFeatures: Array<{ id: string; name: string; description: string; enabled: boolean; tooltip?: string }>;
  expertFeatures: Array<{ id: string; name: string; shortcut: string; category: string; advancedOptions?: string[] }>;
}

interface AdaptiveUXActions {
  setSkillLevel: (level: UserSkillLevel) => void;
  toggleComplexityLayer: (level: ComplexityLevel) => void;
  dismissHint: (hintId: string) => void;
  acceptRecommendation: (id: string) => void;
  setLearningPhase: (phase: LearningPhase) => void;
  updateCognitiveLoad: (metrics: Partial<CognitiveMetrics>) => void;
  setAttentionState: (state: AttentionState) => void;
  setActiveView: (view: AdaptiveUXState['activeView']) => void;
  addToolbarAction: (action: ToolbarAction) => void;
  removeToolbarAction: (actionId: string) => void;
  addPredictedAction: (action: PredictedAction) => void;
  clearPredictedActions: () => void;
  recordPattern: (pattern: LearningPattern) => void;
  setFocusHighlights: (highlights: FocusHighlight[]) => void;
  addHint: (hint: WorkflowHint) => void;
  snoozeHint: (hintId: string, duration: number) => void;
  setSimplifiedFeatures: (features: AdaptiveUXState['simplifiedFeatures']) => void;
  setExpertFeatures: (features: AdaptiveUXState['expertFeatures']) => void;
  updatePreferences: (prefs: Partial<AdaptiveUXState['preferences']>) => void;
  resetStore: () => void;
}

const initialState: AdaptiveUXState = {
  userSkillLevel: 'intermediate',
  complexityLayers: {
    simple: true,
    standard: true,
    advanced: false,
  },
  contextHints: ['action', 'navigation', 'information'],
  recommendations: [],
  learningPhase: 'observing',
  activeView: 'overview',
  cognitiveLoad: {
    currentLoad: 30,
    maxCapacity: 100,
    stressLevel: 'low',
    recommendedBreaks: 0,
    focusScore: 85,
  },
  attentionState: 'focused',
  dismissedHints: [],
  preferences: {
    uiDensity: 'normal',
    animationLevel: 'standard',
    autoAdjust: true,
    learningModeEnabled: true,
  },
  toolbarActions: [],
  predictedActions: [],
  observedPatterns: [],
  focusHighlights: [],
  availableHints: [],
  simplifiedFeatures: [],
  expertFeatures: [],
};

export const useAdaptiveUXStore = create<AdaptiveUXState & AdaptiveUXActions>()(
  immer((set) => ({
    ...initialState,

    setSkillLevel: (level) =>
      set((state) => {
        state.userSkillLevel = level;
      }),

    toggleComplexityLayer: (level) =>
      set((state) => {
        state.complexityLayers[level] = !state.complexityLayers[level];
      }),

    dismissHint: (hintId) =>
      set((state) => {
        if (!state.dismissedHints.includes(hintId)) {
          state.dismissedHints.push(hintId);
        }
      }),

    acceptRecommendation: (id) =>
      set((state) => {
        const rec = state.recommendations.find((r) => r.id === id);
        if (rec) {
          rec.actions[0]?.execute();
        }
        state.recommendations = state.recommendations.filter((r) => r.id !== id);
      }),

    setLearningPhase: (phase) =>
      set((state) => {
        state.learningPhase = phase;
      }),

    updateCognitiveLoad: (metrics) =>
      set((state) => {
        state.cognitiveLoad = { ...state.cognitiveLoad, ...metrics };
      }),

    setAttentionState: (state) =>
      set((s) => {
        s.attentionState = state;
      }),

    setActiveView: (view) =>
      set((s) => {
        s.activeView = view;
      }),

    addToolbarAction: (action) =>
      set((s) => {
        s.toolbarActions.push(action);
      }),

    removeToolbarAction: (actionId) =>
      set((s) => {
        s.toolbarActions = s.toolbarActions.filter((a) => a.id !== actionId);
      }),

    addPredictedAction: (action) =>
      set((s) => {
        s.predictedActions.push(action);
      }),

    clearPredictedActions: () =>
      set((s) => {
        s.predictedActions = [];
      }),

    recordPattern: (pattern) =>
      set((s) => {
        s.observedPatterns.push(pattern);
      }),

    setFocusHighlights: (highlights) =>
      set((s) => {
        s.focusHighlights = highlights;
      }),

    addHint: (hint) =>
      set((s) => {
        s.availableHints.push(hint);
      }),

    snoozeHint: (hintId, _duration) =>
      set((s) => {
        const hint = s.availableHints.find((h) => h.id === hintId);
        if (hint) {
          s.dismissedHints.push(hintId);
        }
      }),

    setSimplifiedFeatures: (features) =>
      set((s) => {
        s.simplifiedFeatures = features;
      }),

    setExpertFeatures: (features) =>
      set((s) => {
        s.expertFeatures = features;
      }),

    updatePreferences: (prefs) =>
      set((s) => {
        s.preferences = { ...s.preferences, ...prefs };
      }),

    resetStore: () =>
      set(() => initialState),
  }))
);

export const selectVisibleHints = (state: AdaptiveUXState): WorkflowHint[] =>
  state.availableHints.filter((h) => !state.dismissedHints.includes(h.id));

export const selectActiveRecommendations = (state: AdaptiveUXState): WorkspaceRecommendation[] =>
  state.recommendations.filter((r) => r.impact === 'high' || r.impact === 'medium');

export const selectCurrentComplexity = (state: AdaptiveUXState): ComplexityConfig => {
  const enabledFeatures = Object.entries(state.complexityLayers)
    .filter(([, enabled]) => enabled)
    .map(([level]) => level);
  
  return {
    level: state.userSkillLevel === 'beginner' ? 'simple' : state.userSkillLevel === 'expert' ? 'advanced' : 'standard',
    enabledFeatures,
    disabledFeatures: Object.entries(state.complexityLayers)
      .filter(([, enabled]) => !enabled)
      .map(([level]) => level),
    uiDensity: state.preferences.uiDensity,
    animationLevel: state.preferences.animationLevel,
  };
};

export const selectAdaptationScore = (state: AdaptiveUXState): number => {
  let score = 0;
  
  if (state.learningPhase === 'optimized') score += 40;
  else if (state.learningPhase === 'adapting') score += 20;
  
  score += Math.round(state.cognitiveLoad.focusScore * 0.3);
  
  const visibleHintsCount = state.availableHints.filter((h) => !state.dismissedHints.includes(h.id)).length;
  score += Math.max(0, 30 - visibleHintsCount * 3);
  
  return Math.min(100, score);
};

export const useAdaptiveUX = () => useAdaptiveUXStore();
export const useComplexityLayers = () => {
  const layers = useAdaptiveUXStore((s) => s.complexityLayers);
  const toggle = useAdaptiveUXStore((s) => s.toggleComplexityLayer);
  const skillLevel = useAdaptiveUXStore((s) => s.userSkillLevel);
  return { layers, toggle, skillLevel };
};
export const useContextToolbar = () => {
  const actions = useAdaptiveUXStore((s) => s.toolbarActions);
  const addAction = useAdaptiveUXStore((s) => s.addToolbarAction);
  const removeAction = useAdaptiveUXStore((s) => s.removeToolbarAction);
  return { actions, addAction, removeAction };
};
export const usePredictiveActions = () => {
  const predictions = useAdaptiveUXStore((s) => s.predictedActions);
  const addPrediction = useAdaptiveUXStore((s) => s.addPredictedAction);
  const clearPredictions = useAdaptiveUXStore((s) => s.clearPredictedActions);
  return { predictions, addPrediction, clearPredictions };
};
export const useWorkspaceRecommendations = () => {
  const recommendations = useAdaptiveUXStore((s) => s.recommendations);
  const accept = useAdaptiveUXStore((s) => s.acceptRecommendation);
  const activeRecs = selectActiveRecommendations(useAdaptiveUXStore.getState());
  return { recommendations, accept, activeRecs };
};
export const useLearningMode = () => {
  const phase = useAdaptiveUXStore((s) => s.learningPhase);
  const patterns = useAdaptiveUXStore((s) => s.observedPatterns);
  const setPhase = useAdaptiveUXStore((s) => s.setLearningPhase);
  const recordPattern = useAdaptiveUXStore((s) => s.recordPattern);
  return { phase, patterns, setPhase, recordPattern };
};
export const useBeginnerView = () => {
  const features = useAdaptiveUXStore((s) => s.simplifiedFeatures);
  const setFeatures = useAdaptiveUXStore((s) => s.setSimplifiedFeatures);
  return { features, setFeatures };
};
export const useExpertView = () => {
  const features = useAdaptiveUXStore((s) => s.expertFeatures);
  const setFeatures = useAdaptiveUXStore((s) => s.setExpertFeatures);
  return { features, setFeatures };
};
export const useCognitiveLoad = () => {
  const metrics = useAdaptiveUXStore((s) => s.cognitiveLoad);
  const update = useAdaptiveUXStore((s) => s.updateCognitiveLoad);
  return { metrics, update };
};
export const useAttentionFocus = () => {
  const state = useAdaptiveUXStore((s) => s.attentionState);
  const highlights = useAdaptiveUXStore((s) => s.focusHighlights);
  const setState = useAdaptiveUXStore((s) => s.setAttentionState);
  const setHighlights = useAdaptiveUXStore((s) => s.setFocusHighlights);
  return { state, highlights, setState, setHighlights };
};
export const useWorkflowHints = () => {
  const hints = useAdaptiveUXStore((s) => s.availableHints);
  const dismissed = useAdaptiveUXStore((s) => s.dismissedHints);
  const dismiss = useAdaptiveUXStore((s) => s.dismissHint);
  const snooze = useAdaptiveUXStore((s) => s.snoozeHint);
  const visibleHints = selectVisibleHints(useAdaptiveUXStore.getState());
  return { hints, dismissed, dismiss, snooze, visibleHints };
};
