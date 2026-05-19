/**
 * @functionfly/ui-adaptive-ux
 * Adaptive UX Components - AI-powered user experience components
 */

// ============================================================================
// User & Skill Types
// ============================================================================

export type UserSkillLevel = 'beginner' | 'intermediate' | 'expert';
export type ComplexityLevel = 'simple' | 'standard' | 'advanced';
export type ContextHint = 'action' | 'navigation' | 'information';
export type RecommendationType = 'workspace' | 'workflow' | 'shortcut' | 'feature';
export type AttentionState = 'focused' | 'distracted' | 'idle';
export type LearningPhase = 'observing' | 'adapting' | 'optimized';

// ============================================================================
// Adaptive Complexity Layer
// ============================================================================

export interface ComplexityConfig {
  level: ComplexityLevel;
  enabledFeatures: string[];
  disabledFeatures: string[];
  uiDensity: 'compact' | 'normal' | 'spacious';
  animationLevel: 'minimal' | 'standard' | 'enhanced';
}

export interface AdaptiveComplexityLayerProps {
  userSkillLevel: UserSkillLevel;
  currentContext: string;
  onComplexityChange?: (config: ComplexityConfig) => void;
  className?: string;
}

// ============================================================================
// Context-Aware Toolbar
// ============================================================================

export interface ToolbarAction {
  id: string;
  label: string;
  icon: string;
  shortcut?: string;
  enabled: boolean;
  priority: number;
  contextRequirements?: string[];
}

export interface ContextAwareToolbarProps {
  availableActions: ToolbarAction[];
  currentContext: string;
  onActionExecute?: (action: ToolbarAction) => void;
  onContextChange?: (context: string) => void;
  className?: string;
}

// ============================================================================
// Predictive Action Bar
// ============================================================================

export interface PredictedAction {
  id: string;
  label: string;
  probability: number;
  icon: string;
  expectedOutcome: string;
  confidence: number;
}

export interface PredictiveActionBarProps {
  predictions: PredictedAction[];
  onActionSelect?: (action: PredictedAction) => void;
  onFeedback?: (actionId: string, helpful: boolean) => void;
  maxVisible?: number;
  className?: string;
}

// ============================================================================
// Smart Workspace Recommendations
// ============================================================================

export interface WorkspaceRecommendation {
  id: string;
  type: RecommendationType;
  title: string;
  description: string;
  reason: string;
  confidence: number;
  impact: 'low' | 'medium' | 'high';
  actions: Array<{ label: string; execute: () => void }>;
}

export interface SmartWorkspaceRecommendationsProps {
  recommendations: WorkspaceRecommendation[];
  onAccept?: (id: string) => void;
  onDismiss?: (id: string) => void;
  onApplyInstant?: (id: string) => void;
  className?: string;
}

// ============================================================================
// Learning Mode Overlay
// ============================================================================

export interface LearningPattern {
  action: string;
  frequency: number;
  averageTime: number;
  context: string;
  learnedAt: number;
}

export interface LearningModeOverlayProps {
  isActive: boolean;
  phase: LearningPhase;
  observedPatterns: LearningPattern[];
  onPatternLearn?: (pattern: LearningPattern) => void;
  onDismiss?: () => void;
  className?: string;
}

// ============================================================================
// Beginner Simplification View
// ============================================================================

export interface SimplifiedFeature {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
  tooltip?: string;
}

export interface BeginnerSimplificationViewProps {
  enabledFeatures: SimplifiedFeature[];
  onFeatureToggle?: (id: string, enabled: boolean) => void;
  onGuidedTourStart?: () => void;
  onHelpRequest?: () => void;
  className?: string;
}

// ============================================================================
// Expert System View
// ============================================================================

export interface ExpertFeature {
  id: string;
  name: string;
  shortcut: string;
  category: string;
  advancedOptions?: string[];
}

export interface ExpertSystemViewProps {
  availableFeatures: ExpertFeature[];
  onQuickAccess?: (featureId: string) => void;
  onMacroCreate?: (name: string, steps: string[]) => void;
  onCustomization?: (featureId: string, config: any) => void;
  className?: string;
}

// ============================================================================
// Cognitive Load Balancer
// ============================================================================

export interface CognitiveMetrics {
  currentLoad: number;
  maxCapacity: number;
  stressLevel: 'low' | 'medium' | 'high' | 'critical';
  recommendedBreaks: number;
  focusScore: number;
}

export interface CognitiveLoadBalancerProps {
  metrics: CognitiveMetrics;
  onBreakSuggestion?: (type: 'micro' | 'short' | 'long') => void;
  onLoadReduction?: (actions: string[]) => void;
  onFocusModeToggle?: (enabled: boolean) => void;
  className?: string;
}

// ============================================================================
// Attention Focus Overlay
// ============================================================================

export interface FocusHighlight {
  elementId: string;
  message: string;
  priority: 'low' | 'medium' | 'high' | 'critical';
  position: 'top' | 'bottom' | 'left' | 'right';
}

export interface AttentionFocusOverlayProps {
  attentionState: AttentionState;
  highlights: FocusHighlight[];
  onHighlightClick?: (highlight: FocusHighlight) => void;
  onStateChange?: (state: AttentionState) => void;
  className?: string;
}

// ============================================================================
// Workflow Optimization Hints
// ============================================================================

export interface WorkflowHint {
  id: string;
  type: 'tip' | 'warning' | 'optimization';
  title: string;
  description: string;
  potentialSaving?: { time: number; effort: number };
  actionLabel?: string;
}

export interface WorkflowOptimizationHintsProps {
  hints: WorkflowHint[];
  onHintApply?: (hint: WorkflowHint) => void;
  onHintDismiss?: (hintId: string) => void;
  onHintSnooze?: (hintId: string, duration: number) => void;
  className?: string;
}

// ============================================================================
// Adaptive Layout Engine
// ============================================================================

export interface LayoutConfiguration {
  id: string;
  name: string;
  complexityLevel: ComplexityLevel;
  panels: Array<{
    id: string;
    position: 'left' | 'right' | 'top' | 'bottom' | 'center';
    size: number;
    collapsed: boolean;
  }>;
  recommendedFor: UserSkillLevel;
}

export interface AdaptiveLayoutEngineProps {
  configurations: LayoutConfiguration[];
  currentConfig: LayoutConfiguration | null;
  onConfigSelect?: (config: LayoutConfiguration) => void;
  onAutoAdjust?: (skillLevel: UserSkillLevel) => void;
  className?: string;
}
