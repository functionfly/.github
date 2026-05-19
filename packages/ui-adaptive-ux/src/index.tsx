/**
 * @functionfly/ui-adaptive-ux
 * Adaptive UX Components - AI-powered user experience components
 */

import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { cn } from '@functionfly/ui-core';
import {
  Gauge,
  Layers,
  Lightbulb,
  Brain,
  GraduationCap,
  Zap,
  Eye,
  Target,
  Sparkles,
  Compass,
  AlertTriangle,
  Clock,
  TrendingUp,
  ChevronRight,
  ChevronDown,
  Check,
  X,
  Info,
  ArrowRight,
  Keyboard,
  LayoutDashboard,
  Workflow,
  GaugeCircle,
  Activity,
  Coffee,
  EyeOff,
  Focus,
  LightbulbIcon,
  Wand2,
  Timer,
  AlertCircle,
} from 'lucide-react';

// ============================================================================
// Adaptive Complexity Layer
// ============================================================================

type ComplexityLevel = 'simple' | 'standard' | 'advanced';
type UserSkillLevel = 'beginner' | 'intermediate' | 'expert';

interface ComplexityConfig {
  level: ComplexityLevel;
  enabledFeatures: string[];
  disabledFeatures: string[];
  uiDensity: 'compact' | 'normal' | 'spacious';
  animationLevel: 'minimal' | 'standard' | 'enhanced';
}

interface AdaptiveComplexityLayerProps {
  userSkillLevel: UserSkillLevel;
  currentContext: string;
  onComplexityChange?: (config: ComplexityConfig) => void;
  className?: string;
}

export const AdaptiveComplexityLayer: React.FC<AdaptiveComplexityLayerProps> = ({
  userSkillLevel,
  currentContext,
  onComplexityChange,
  className,
}) => {
  const [currentLevel, setCurrentLevel] = useState<ComplexityLevel>('standard');

  const getComplexityConfig = useCallback((level: ComplexityLevel): ComplexityConfig => {
    const featureSets = {
      simple: {
        enabledFeatures: ['basic-navigation', 'primary-actions', 'clear-labels', 'tooltips'],
        disabledFeatures: ['keyboard-shortcuts', 'advanced-filters', 'bulk-operations', 'customizations'],
        uiDensity: 'spacious' as const,
        animationLevel: 'minimal' as const,
      },
      standard: {
        enabledFeatures: ['basic-navigation', 'primary-actions', 'secondary-actions', 'tooltips', 'keyboard-shortcuts'],
        disabledFeatures: ['bulk-operations', 'custom-shortcuts'],
        uiDensity: 'normal' as const,
        animationLevel: 'standard' as const,
      },
      advanced: {
        enabledFeatures: ['all'],
        disabledFeatures: [],
        uiDensity: 'compact' as const,
        animationLevel: 'enhanced' as const,
      },
    };
    return { level, ...featureSets[level] };
  }, []);

  useEffect(() => {
    const config = getComplexityConfig(currentLevel);
    onComplexityChange?.(config);
  }, [currentLevel, getComplexityConfig, onComplexityChange]);

  const getSkillIndicatorColor = (skill: UserSkillLevel) => {
    switch (skill) {
      case 'beginner': return 'text-green-400';
      case 'intermediate': return 'text-amber-400';
      case 'expert': return 'text-aviation-cyan';
    }
  };

  return (
    <div className={cn('bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Layers className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Adaptive Complexity</h3>
          </div>
          <div className="flex items-center gap-2">
            <span className={cn('text-xs font-medium uppercase', getSkillIndicatorColor(userSkillLevel))}>
              {userSkillLevel}
            </span>
            <span className="text-aviation-text-dim">•</span>
            <span className="text-xs text-aviation-text-dim">{currentContext}</span>
          </div>
        </div>
      </div>

      <div className="p-4">
        <div className="flex items-center justify-between mb-4">
          <span className="text-xs text-aviation-text-dim">UI Complexity Level</span>
          <div className="flex items-center gap-1">
            {(['simple', 'standard', 'advanced'] as ComplexityLevel[]).map((level) => (
              <button
                key={level}
                onClick={() => setCurrentLevel(level)}
                className={cn(
                  'px-3 py-1.5 text-xs rounded transition-colors',
                  currentLevel === level
                    ? 'bg-aviation-cyan text-aviation-bg-primary'
                    : 'bg-aviation-bg-secondary text-aviation-text-muted hover:text-aviation-text-primary'
                )}
              >
                {level.charAt(0).toUpperCase() + level.slice(1)}
              </button>
            ))}
          </div>
        </div>

        <div className="space-y-3">
          <div className="flex items-center justify-between text-xs">
            <span className="text-aviation-text-dim">UI Density</span>
            <span className="text-aviation-text-primary">
              {currentLevel === 'simple' ? 'Spacious' : currentLevel === 'standard' ? 'Normal' : 'Compact'}
            </span>
          </div>
          <div className="flex items-center justify-between text-xs">
            <span className="text-aviation-text-dim">Animations</span>
            <span className="text-aviation-text-primary">
              {currentLevel === 'simple' ? 'Minimal' : currentLevel === 'standard' ? 'Standard' : 'Enhanced'}
            </span>
          </div>
          <div className="flex items-center justify-between text-xs">
            <span className="text-aviation-text-dim">Features Enabled</span>
            <span className="text-aviation-cyan">
              {currentLevel === 'simple' ? '4' : currentLevel === 'standard' ? '5' : 'All'}
            </span>
          </div>
        </div>

        <div className="mt-4 pt-4 border-t border-aviation-border-panel">
          <div className="flex items-center gap-2 text-xs text-aviation-text-muted">
            <Brain className="w-4 h-4" />
            <span>Automatically adapts based on your interaction patterns</span>
          </div>
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Context-Aware Toolbar
// ============================================================================

interface ToolbarAction {
  id: string;
  label: string;
  icon: string;
  shortcut?: string;
  enabled: boolean;
  priority: number;
  contextRequirements?: string[];
}

interface ContextAwareToolbarProps {
  availableActions: ToolbarAction[];
  currentContext: string;
  onActionExecute?: (action: ToolbarAction) => void;
  onContextChange?: (context: string) => void;
  className?: string;
}

export const ContextAwareToolbar: React.FC<ContextAwareToolbarProps> = ({
  availableActions,
  currentContext,
  onActionExecute,
  onContextChange,
  className,
}) => {
  const [hoveredAction, setHoveredAction] = useState<string | null>(null);

  const contextOptions = [
    { id: 'editing', label: 'Editing', icon: 'Edit3' },
    { id: 'viewing', label: 'Viewing', icon: 'Eye' },
    { id: 'analyzing', label: 'Analyzing', icon: 'BarChart2' },
    { id: 'debugging', label: 'Debugging', icon: 'Bug' },
  ];

  const sortedActions = useMemo(() => {
    return [...availableActions]
      .filter(a => a.enabled && (!a.contextRequirements || a.contextRequirements.includes(currentContext)))
      .sort((a, b) => b.priority - a.priority);
  }, [availableActions, currentContext]);

  const getIconComponent = (iconName: string) => {
    const icons: Record<string, React.ReactNode> = {
      'Plus': <Plus className="w-4 h-4" />,
      'Trash': <Trash2 className="w-4 h-4" />,
      'Edit3': <Edit3 className="w-4 h-4" />,
      'Copy': <Copy className="w-4 h-4" />,
      'Save': <Save className="w-4 h-4" />,
      'Settings': <Settings className="w-4 h-4" />,
      'Play': <Play className="w-4 h-4" />,
      'Pause': <Pause className="w-4 h-4" />,
    };
    return icons[iconName] || <Square className="w-4 h-4" />;
  };

  return (
    <div className={cn('bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-2 border-b border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Compass className="w-4 h-4 text-aviation-cyan" />
            <span className="text-xs text-aviation-text-dim">Context:</span>
            <select
              value={currentContext}
              onChange={(e) => onContextChange?.(e.target.value)}
              className="bg-transparent text-xs text-aviation-text-primary border-none focus:outline-none cursor-pointer"
            >
              {contextOptions.map(ctx => (
                <option key={ctx.id} value={ctx.id}>{ctx.label}</option>
              ))}
            </select>
          </div>
          <span className="text-[10px] text-aviation-text-dim">{sortedActions.length} actions available</span>
        </div>
      </div>

      <div className="flex items-center gap-1 p-2 overflow-x-auto">
        {sortedActions.map((action) => (
          <button
            key={action.id}
            onClick={() => onActionExecute?.(action)}
            onMouseEnter={() => setHoveredAction(action.id)}
            onMouseLeave={() => setHoveredAction(null)}
            className={cn(
              'flex items-center gap-2 px-3 py-1.5 rounded transition-all',
              'hover:bg-aviation-bg-instrument',
              hoveredAction === action.id && 'bg-aviation-bg-instrument ring-1 ring-aviation-cyan/50'
            )}
          >
            <span className="text-aviation-text-primary">{getIconComponent(action.icon)}</span>
            <span className="text-xs text-aviation-text-primary whitespace-nowrap">{action.label}</span>
            {action.shortcut && (
              <span className="text-[10px] text-aviation-text-dim px-1 py-0.5 bg-aviation-bg-secondary rounded">
                {action.shortcut}
              </span>
            )}
          </button>
        ))}
      </div>
    </div>
  );
};

// ============================================================================
// Predictive Action Bar
// ============================================================================

interface PredictedAction {
  id: string;
  label: string;
  probability: number;
  icon: string;
  expectedOutcome: string;
  confidence: number;
}

interface PredictiveActionBarProps {
  predictions: PredictedAction[];
  onActionSelect?: (action: PredictedAction) => void;
  onFeedback?: (actionId: string, helpful: boolean) => void;
  maxVisible?: number;
  className?: string;
}

export const PredictiveActionBar: React.FC<PredictiveActionBarProps> = ({
  predictions,
  onActionSelect,
  onFeedback,
  maxVisible = 5,
  className,
}) => {
  const [feedbackVisible, setFeedbackVisible] = useState<string | null>(null);

  const topPredictions = useMemo(() => {
    return [...predictions]
      .sort((a, b) => b.probability - a.probability)
      .slice(0, maxVisible);
  }, [predictions, maxVisible]);

  const getProbabilityColor = (prob: number) => {
    if (prob >= 0.8) return 'text-green-400';
    if (prob >= 0.5) return 'text-amber-400';
    return 'text-aviation-text-dim';
  };

  return (
    <div className={cn('bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-2 border-b border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="flex items-center gap-2">
          <Sparkles className="w-4 h-4 text-aviation-cyan" />
          <span className="text-xs text-aviation-text-primary font-medium">Next Predicted Actions</span>
          <span className="text-[10px] text-aviation-text-dim ml-auto">AI-powered predictions</span>
        </div>
      </div>

      <div className="flex items-center gap-2 p-3 overflow-x-auto">
        {topPredictions.map((action, index) => (
          <div
            key={action.id}
            className={cn(
              'relative flex flex-col items-center p-3 rounded-lg border transition-all cursor-pointer',
              'hover:bg-aviation-bg-secondary hover:border-aviation-cyan/50',
              'min-w-[100px]'
            )}
            onClick={() => onActionSelect?.(action)}
          >
            {index === 0 && (
              <div className="absolute -top-1 -right-1 px-1.5 py-0.5 bg-aviation-cyan rounded text-[9px] text-aviation-bg-primary font-medium">
                Top
              </div>
            )}
            <div className="w-8 h-8 rounded-full bg-aviation-bg-instrument flex items-center justify-center mb-2">
              <span className="text-aviation-text-primary">
                <LightbulbIcon className="w-4 h-4" />
              </span>
            </div>
            <span className="text-xs text-aviation-text-primary font-medium mb-1">{action.label}</span>
            <span className={cn('text-lg font-bold', getProbabilityColor(action.probability))}>
              {(action.probability * 100).toFixed(0)}%
            </span>
            <span className="text-[10px] text-aviation-text-dim mt-1">confidence: {action.confidence}</span>

            {feedbackVisible === action.id && (
              <div className="absolute top-full mt-2 flex items-center gap-1 z-10">
                <button
                  onClick={(e) => { e.stopPropagation(); onFeedback?.(action.id, true); setFeedbackVisible(null); }}
                  className="p-1 bg-green-500/20 rounded hover:bg-green-500/40"
                >
                  <ThumbsUp className="w-3 h-3 text-green-400" />
                </button>
                <button
                  onClick={(e) => { e.stopPropagation(); onFeedback?.(action.id, false); setFeedbackVisible(null); }}
                  className="p-1 bg-red-500/20 rounded hover:bg-red-500/40"
                >
                  <ThumbsDown className="w-3 h-3 text-red-400" />
                </button>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
};

// ============================================================================
// Smart Workspace Recommendations
// ============================================================================

type RecommendationType = 'workspace' | 'workflow' | 'shortcut' | 'feature';

interface WorkspaceRecommendation {
  id: string;
  type: RecommendationType;
  title: string;
  description: string;
  reason: string;
  confidence: number;
  impact: 'low' | 'medium' | 'high';
  actions: Array<{ label: string; execute: () => void }>;
}

interface SmartWorkspaceRecommendationsProps {
  recommendations: WorkspaceRecommendation[];
  onAccept?: (id: string) => void;
  onDismiss?: (id: string) => void;
  onApplyInstant?: (id: string) => void;
  className?: string;
}

export const SmartWorkspaceRecommendations: React.FC<SmartWorkspaceRecommendationsProps> = ({
  recommendations,
  onAccept,
  onDismiss,
  onApplyInstant,
  className,
}) => {
  const getTypeIcon = (type: RecommendationType) => {
    switch (type) {
      case 'workspace': return <LayoutDashboard className="w-4 h-4" />;
      case 'workflow': return <Workflow className="w-4 h-4" />;
      case 'shortcut': return <Keyboard className="w-4 h-4" />;
      case 'feature': return <Zap className="w-4 h-4" />;
    }
  };

  const getImpactColor = (impact: string) => {
    switch (impact) {
      case 'high': return 'text-green-400 bg-green-400/10';
      case 'medium': return 'text-amber-400 bg-amber-400/10';
      default: return 'text-aviation-text-dim bg-aviation-bg-secondary';
    }
  };

  return (
    <div className={cn('bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center gap-2">
          <Lightbulb className="w-5 h-5 text-amber-400" />
          <h3 className="text-sm font-medium text-aviation-text-primary">Smart Recommendations</h3>
          <span className="ml-auto text-xs text-aviation-text-dim">{recommendations.length} suggestions</span>
        </div>
      </div>

      <div className="max-h-96 overflow-y-auto">
        {recommendations.map((rec) => (
          <div key={rec.id} className="p-4 border-b border-aviation-border-panel last:border-b-0">
            <div className="flex items-start gap-3">
              <div className="w-8 h-8 rounded-lg bg-aviation-bg-secondary flex items-center justify-center flex-shrink-0">
                <span className="text-aviation-text-primary">{getTypeIcon(rec.type)}</span>
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <h4 className="text-sm font-medium text-aviation-text-primary truncate">{rec.title}</h4>
                  <span className={cn('text-[10px] px-1.5 py-0.5 rounded uppercase font-medium', getImpactColor(rec.impact))}>
                    {rec.impact}
                  </span>
                </div>
                <p className="text-xs text-aviation-text-dim mb-2">{rec.description}</p>
                <div className="flex items-center gap-1 text-[10px] text-aviation-cyan mb-3">
                  <Info className="w-3 h-3" />
                  <span>{rec.reason}</span>
                </div>
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => onApplyInstant?.(rec.id)}
                    className="px-3 py-1.5 bg-aviation-cyan text-aviation-bg-primary rounded text-xs hover:bg-aviation-cyan/90 transition-colors"
                  >
                    Apply
                  </button>
                  <button
                    onClick={() => onAccept?.(rec.id)}
                    className="px-3 py-1.5 bg-aviation-bg-secondary text-aviation-text-primary rounded text-xs hover:bg-aviation-bg-instrument transition-colors"
                  >
                    Accept
                  </button>
                  <button
                    onClick={() => onDismiss?.(rec.id)}
                    className="p-1.5 text-aviation-text-dim hover:text-aviation-text-primary transition-colors ml-auto"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>
          </div>
        ))}

        {recommendations.length === 0 && (
          <div className="flex flex-col items-center justify-center py-12 text-aviation-text-muted">
            <Check className="w-8 h-8 mb-2 text-green-400" />
            <p className="text-sm">All recommendations applied</p>
          </div>
        )}
      </div>
    </div>
  );
};

// ============================================================================
// Learning Mode Overlay
// ============================================================================

type LearningPhase = 'observing' | 'adapting' | 'optimized';

interface LearningPattern {
  action: string;
  frequency: number;
  averageTime: number;
  context: string;
  learnedAt: number;
}

interface LearningModeOverlayProps {
  isActive: boolean;
  phase: LearningPhase;
  observedPatterns: LearningPattern[];
  onPatternLearn?: (pattern: LearningPattern) => void;
  onDismiss?: () => void;
  className?: string;
}

export const LearningModeOverlay: React.FC<LearningModeOverlayProps> = ({
  isActive,
  phase,
  observedPatterns,
  onDismiss,
  className,
}) => {
  const [currentTip, setCurrentTip] = useState(0);

  const getPhaseColor = (p: LearningPhase) => {
    switch (p) {
      case 'observing': return 'text-amber-400';
      case 'adapting': return 'text-aviation-cyan';
      case 'optimized': return 'text-green-400';
    }
  };

  const getPhaseLabel = (p: LearningPhase) => {
    switch (p) {
      case 'observing': return 'Learning your patterns...';
      case 'adapting': return 'Adapting interface...';
      case 'optimized': return 'Optimization complete!';
    }
  };

  if (!isActive) return null;

  return (
    <div className={cn('bg-aviation-bg-panel rounded-lg border border-aviation-cyan/50 overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel bg-aviation-cyan/5">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Brain className="w-5 h-5 text-aviation-cyan animate-pulse" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Learning Mode</h3>
          </div>
          <div className="flex items-center gap-2">
            <span className={cn('text-xs font-medium', getPhaseColor(phase))}>{getPhaseLabel(phase)}</span>
            <button onClick={onDismiss} className="p-1 hover:bg-aviation-bg-instrument rounded">
              <X className="w-4 h-4 text-aviation-text-muted" />
            </button>
          </div>
        </div>
      </div>

      <div className="p-4">
        <div className="mb-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs text-aviation-text-dim">Learning Progress</span>
            <span className="text-xs text-aviation-text-primary">{observedPatterns.length} patterns detected</span>
          </div>
          <div className="h-2 bg-aviation-bg-instrument rounded-full overflow-hidden">
            <div
              className={cn('h-full transition-all duration-500', {
                'bg-amber-400': phase === 'observing',
                'bg-aviation-cyan': phase === 'adapting',
                'bg-green-400': phase === 'optimized',
              })}
              style={{ width: `${phase === 'observing' ? 33 : phase === 'adapting' ? 66 : 100}%` }}
            />
          </div>
        </div>

        {observedPatterns.length > 0 && (
          <div className="space-y-2">
            <span className="text-xs text-aviation-text-dim">Observed Patterns</span>
            {observedPatterns.slice(0, 5).map((pattern, index) => (
              <div key={index} className="flex items-center justify-between p-2 bg-aviation-bg-secondary rounded">
                <span className="text-xs text-aviation-text-primary">{pattern.action}</span>
                <span className="text-[10px] text-aviation-text-dim">{pattern.frequency}x • {pattern.context}</span>
              </div>
            ))}
          </div>
        )}

        <div className="mt-4 pt-4 border-t border-aviation-border-panel">
          <div className="flex items-center gap-2 text-xs text-aviation-cyan">
            <GraduationCap className="w-4 h-4" />
            <span>Keep using the interface to improve recommendations</span>
          </div>
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Beginner Simplification View
// ============================================================================

interface SimplifiedFeature {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
  tooltip?: string;
}

interface BeginnerSimplificationViewProps {
  enabledFeatures: SimplifiedFeature[];
  onFeatureToggle?: (id: string, enabled: boolean) => void;
  onGuidedTourStart?: () => void;
  onHelpRequest?: () => void;
  className?: string;
}

export const BeginnerSimplificationView: React.FC<BeginnerSimplificationViewProps> = ({
  enabledFeatures,
  onFeatureToggle,
  onGuidedTourStart,
  onHelpRequest,
  className,
}) => {
  return (
    <div className={cn('bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel bg-green-500/5">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <GraduationCap className="w-5 h-5 text-green-400" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Beginner Mode</h3>
          </div>
          <button
            onClick={onGuidedTourStart}
            className="flex items-center gap-1.5 px-3 py-1.5 bg-aviation-cyan text-aviation-bg-primary rounded text-xs hover:bg-aviation-cyan/90 transition-colors"
          >
            <Play className="w-3 h-3" />
            Start Tour
          </button>
        </div>
      </div>

      <div className="p-4">
        <div className="mb-4 p-3 bg-aviation-bg-secondary rounded-lg">
          <p className="text-xs text-aviation-text-primary mb-1">Welcome! This simplified view shows essential features.</p>
          <p className="text-xs text-aviation-text-dim">Toggle features below to customize your experience.</p>
        </div>

        <div className="space-y-2">
          {enabledFeatures.map((feature) => (
            <div
              key={feature.id}
              className="flex items-center justify-between p-3 bg-aviation-bg-secondary rounded-lg hover:bg-aviation-bg-instrument transition-colors"
            >
              <div className="flex items-center gap-3">
                <button
                  onClick={() => onFeatureToggle?.(feature.id, !feature.enabled)}
                  className={cn(
                    'w-10 h-6 rounded-full transition-colors relative',
                    feature.enabled ? 'bg-aviation-cyan' : 'bg-aviation-text-dim'
                  )}
                >
                  <span
                    className={cn(
                      'absolute top-1 w-4 h-4 rounded-full bg-white transition-transform',
                      feature.enabled ? 'left-5' : 'left-1'
                    )}
                  />
                </button>
                <div>
                  <span className="text-sm text-aviation-text-primary font-medium">{feature.name}</span>
                  {feature.tooltip && (
                    <p className="text-[10px] text-aviation-text-dim">{feature.tooltip}</p>
                  )}
                </div>
              </div>
              <button
                onClick={onHelpRequest}
                className="p-2 text-aviation-text-dim hover:text-aviation-text-primary transition-colors"
              >
                <HelpCircle className="w-4 h-4" />
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Expert System View
// ============================================================================

interface ExpertFeature {
  id: string;
  name: string;
  shortcut: string;
  category: string;
  advancedOptions?: string[];
}

interface ExpertSystemViewProps {
  availableFeatures: ExpertFeature[];
  onQuickAccess?: (featureId: string) => void;
  onMacroCreate?: (name: string, steps: string[]) => void;
  onCustomization?: (featureId: string, config: any) => void;
  className?: string;
}

export const ExpertSystemView: React.FC<ExpertSystemViewProps> = ({
  availableFeatures,
  onQuickAccess,
  onMacroCreate,
  onCustomization,
  className,
}) => {
  const [activeCategory, setActiveCategory] = useState('all');

  const categories = useMemo(() => {
    const cats = new Set(availableFeatures.map(f => f.category));
    return ['all', ...Array.from(cats)];
  }, [availableFeatures]);

  const filteredFeatures = useMemo(() => {
    if (activeCategory === 'all') return availableFeatures;
    return availableFeatures.filter(f => f.category === activeCategory);
  }, [availableFeatures, activeCategory]);

  return (
    <div className={cn('bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel bg-aviation-cyan/5">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <GaugeCircle className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Expert Mode</h3>
          </div>
          <div className="flex items-center gap-1">
            <Zap className="w-4 h-4 text-aviation-cyan" />
            <span className="text-xs text-aviation-text-dim">Full access enabled</span>
          </div>
        </div>
      </div>

      <div className="px-4 py-2 border-b border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="flex items-center gap-1 overflow-x-auto">
          {categories.map((cat) => (
            <button
              key={cat}
              onClick={() => setActiveCategory(cat)}
              className={cn(
                'px-3 py-1.5 text-xs rounded whitespace-nowrap transition-colors',
                activeCategory === cat
                  ? 'bg-aviation-cyan text-aviation-bg-primary'
                  : 'text-aviation-text-muted hover:text-aviation-text-primary'
              )}
            >
              {cat.charAt(0).toUpperCase() + cat.slice(1)}
            </button>
          ))}
        </div>
      </div>

      <div className="p-2 max-h-64 overflow-y-auto">
        <div className="grid grid-cols-2 gap-2">
          {filteredFeatures.map((feature) => (
            <button
              key={feature.id}
              onClick={() => onQuickAccess?.(feature.id)}
              className="flex items-center justify-between p-3 bg-aviation-bg-secondary rounded-lg hover:bg-aviation-bg-instrument transition-colors text-left"
            >
              <span className="text-sm text-aviation-text-primary">{feature.name}</span>
              <span className="text-[10px] px-1.5 py-0.5 bg-aviation-bg-instrument rounded text-aviation-text-dim">
                {feature.shortcut}
              </span>
            </button>
          ))}
        </div>
      </div>

      <div className="px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
        <button
          onClick={() => onMacroCreate?.('New Macro', [])}
          className="w-full flex items-center justify-center gap-2 px-3 py-2 bg-aviation-bg-instrument text-aviation-text-primary rounded text-xs hover:bg-aviation-bg-panel transition-colors"
        >
          <Wand2 className="w-4 h-4" />
          Create Custom Macro
        </button>
      </div>
    </div>
  );
};

// ============================================================================
// Cognitive Load Balancer
// ============================================================================

interface CognitiveMetrics {
  currentLoad: number;
  maxCapacity: number;
  stressLevel: 'low' | 'medium' | 'high' | 'critical';
  recommendedBreaks: number;
  focusScore: number;
}

interface CognitiveLoadBalancerProps {
  metrics: CognitiveMetrics;
  onBreakSuggestion?: (type: 'micro' | 'short' | 'long') => void;
  onLoadReduction?: (actions: string[]) => void;
  onFocusModeToggle?: (enabled: boolean) => void;
  className?: string;
}

export const CognitiveLoadBalancer: React.FC<CognitiveLoadBalancerProps> = ({
  metrics,
  onBreakSuggestion,
  onLoadReduction,
  onFocusModeToggle,
  className,
}) => {
  const [focusModeEnabled, setFocusModeEnabled] = useState(false);

  const getStressColor = (level: string) => {
    switch (level) {
      case 'low': return 'text-green-400';
      case 'medium': return 'text-amber-400';
      case 'high': return 'text-orange-400';
      case 'critical': return 'text-red-400';
      default: return 'text-aviation-text-dim';
    }
  };

  const loadPercentage = (metrics.currentLoad / metrics.maxCapacity) * 100;

  return (
    <div className={cn('bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Brain className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Cognitive Load</h3>
          </div>
          <span className={cn('text-xs font-medium uppercase', getStressColor(metrics.stressLevel))}>
            {metrics.stressLevel} stress
          </span>
        </div>
      </div>

      <div className="p-4">
        <div className="mb-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs text-aviation-text-dim">Current Load</span>
            <span className="text-sm text-aviation-text-primary font-medium">
              {metrics.currentLoad}/{metrics.maxCapacity}
            </span>
          </div>
          <div className="h-3 bg-aviation-bg-instrument rounded-full overflow-hidden">
            <div
              className={cn('h-full transition-all duration-500', {
                'bg-green-400': loadPercentage < 50,
                'bg-amber-400': loadPercentage >= 50 && loadPercentage < 75,
                'bg-orange-400': loadPercentage >= 75 && loadPercentage < 90,
                'bg-red-400': loadPercentage >= 90,
              })}
              style={{ width: `${Math.min(100, loadPercentage)}%` }}
            />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-3 mb-4">
          <div className="p-3 bg-aviation-bg-secondary rounded-lg">
            <div className="flex items-center gap-1.5 text-[10px] text-aviation-text-dim mb-1">
              <Target className="w-3 h-3" />
              Focus Score
            </div>
            <div className="text-lg font-bold text-aviation-text-primary">
              {metrics.focusScore}%
            </div>
          </div>

          <div className="p-3 bg-aviation-bg-secondary rounded-lg">
            <div className="flex items-center gap-1.5 text-[10px] text-aviation-text-dim mb-1">
              <Coffee className="w-3 h-3" />
              Breaks Recommended
            </div>
            <div className="text-lg font-bold text-aviation-text-primary">
              {metrics.recommendedBreaks}
            </div>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => {
              setFocusModeEnabled(!focusModeEnabled);
              onFocusModeToggle?.(!focusModeEnabled);
            }}
            className={cn(
              'flex-1 flex items-center justify-center gap-2 px-3 py-2 rounded transition-colors',
              focusModeEnabled
                ? 'bg-aviation-cyan text-aviation-bg-primary'
                : 'bg-aviation-bg-secondary text-aviation-text-primary hover:bg-aviation-bg-instrument'
            )}
          >
            {focusModeEnabled ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
            <span className="text-xs">{focusModeEnabled ? 'Exit Focus Mode' : 'Focus Mode'}</span>
          </button>

          <button
            onClick={() => onBreakSuggestion?.('micro')}
            className="px-3 py-2 bg-aviation-bg-secondary text-aviation-text-primary rounded hover:bg-aviation-bg-instrument transition-colors"
          >
            <Coffee className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Attention Focus Overlay
// ============================================================================

type AttentionState = 'focused' | 'distracted' | 'idle';

interface FocusHighlight {
  elementId: string;
  message: string;
  priority: 'low' | 'medium' | 'high' | 'critical';
  position: 'top' | 'bottom' | 'left' | 'right';
}

interface AttentionFocusOverlayProps {
  attentionState: AttentionState;
  highlights: FocusHighlight[];
  onHighlightClick?: (highlight: FocusHighlight) => void;
  onStateChange?: (state: AttentionState) => void;
  className?: string;
}

export const AttentionFocusOverlay: React.FC<AttentionFocusOverlayProps> = ({
  attentionState,
  highlights,
  onHighlightClick,
  onStateChange,
  className,
}) => {
  const [expandedHighlight, setExpandedHighlight] = useState<string | null>(null);

  const getStateColor = (state: AttentionState) => {
    switch (state) {
      case 'focused': return 'text-green-400 bg-green-400/10';
      case 'distracted': return 'text-red-400 bg-red-400/10';
      case 'idle': return 'text-aviation-text-dim bg-aviation-bg-secondary';
    }
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'critical': return 'border-red-500 bg-red-500/10';
      case 'high': return 'border-amber-500 bg-amber-500/10';
      case 'medium': return 'border-aviation-cyan bg-aviation-cyan/10';
      default: return 'border-aviation-text-dim bg-aviation-bg-secondary';
    }
  };

  return (
    <div className={cn('bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Focus className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Attention Monitor</h3>
          </div>
          <div className="flex items-center gap-2">
            {(['idle', 'focused', 'distracted'] as AttentionState[]).map((state) => (
              <button
                key={state}
                onClick={() => onStateChange?.(state)}
                className={cn(
                  'px-2 py-1 text-[10px] rounded uppercase transition-colors',
                  attentionState === state ? getStateColor(state) : 'text-aviation-text-dim hover:text-aviation-text-primary'
                )}
              >
                {state}
              </button>
            ))}
          </div>
        </div>
      </div>

      <div className="p-4">
        <div className="flex items-center gap-3 mb-4 p-3 bg-aviation-bg-secondary rounded-lg">
          <div className={cn('w-12 h-12 rounded-full flex items-center justify-center', getStateColor(attentionState))}>
            {attentionState === 'focused' && <Target className="w-6 h-6" />}
            {attentionState === 'distracted' && <AlertTriangle className="w-6 h-6" />}
            {attentionState === 'idle' && <Coffee className="w-6 h-6" />}
          </div>
          <div>
            <span className="text-sm text-aviation-text-primary font-medium capitalize">{attentionState}</span>
            <p className="text-xs text-aviation-text-dim">
              {attentionState === 'focused' && 'You are highly focused. Great work!'}
              {attentionState === 'distracted' && 'Multiple highlights require attention.'}
              {attentionState === 'idle' && 'Take a break. You have been active.'}
            </p>
          </div>
        </div>

        {highlights.length > 0 && (
          <div className="space-y-2">
            <span className="text-xs text-aviation-text-dim">{highlights.length} active highlights</span>
            {highlights.map((highlight) => (
              <div
                key={highlight.elementId}
                className={cn(
                  'p-3 rounded-lg border transition-all cursor-pointer',
                  getPriorityColor(highlight.priority),
                  expandedHighlight === highlight.elementId && 'ring-2 ring-aviation-cyan/50'
                )}
                onClick={() => {
                  setExpandedHighlight(expandedHighlight === highlight.elementId ? null : highlight.elementId);
                  onHighlightClick?.(highlight);
                }}
              >
                <div className="flex items-center justify-between mb-1">
                  <span className="text-xs text-aviation-text-primary font-medium">{highlight.elementId}</span>
                  <span className="text-[10px] px-1.5 py-0.5 rounded uppercase">{highlight.priority}</span>
                </div>
                <p className="text-xs text-aviation-text-dim">{highlight.message}</p>
                {expandedHighlight === highlight.elementId && (
                  <div className="mt-2 pt-2 border-t border-aviation-border-panel">
                    <button className="text-xs text-aviation-cyan hover:underline">View Details</button>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

// ============================================================================
// Workflow Optimization Hints
// ============================================================================

interface WorkflowHint {
  id: string;
  type: 'tip' | 'warning' | 'optimization';
  title: string;
  description: string;
  potentialSaving?: { time: number; effort: number };
  actionLabel?: string;
}

interface WorkflowOptimizationHintsProps {
  hints: WorkflowHint[];
  onHintApply?: (hint: WorkflowHint) => void;
  onHintDismiss?: (hintId: string) => void;
  onHintSnooze?: (hintId: string, duration: number) => void;
  className?: string;
}

export const WorkflowOptimizationHints: React.FC<WorkflowOptimizationHintsProps> = ({
  hints,
  onHintApply,
  onHintDismiss,
  onHintSnooze,
  className,
}) => {
  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'tip': return <LightbulbIcon className="w-4 h-4 text-amber-400" />;
      case 'warning': return <AlertTriangle className="w-4 h-4 text-orange-400" />;
      case 'optimization': return <TrendingUp className="w-4 h-4 text-aviation-cyan" />;
      default: return <Info className="w-4 h-4 text-aviation-text-muted" />;
    }
  };

  const getTypeStyles = (type: string) => {
    switch (type) {
      case 'tip': return 'border-amber-500/30 bg-amber-500/5';
      case 'warning': return 'border-orange-500/30 bg-orange-500/5';
      case 'optimization': return 'border-aviation-cyan/30 bg-aviation-cyan/5';
      default: return 'border-aviation-border-panel bg-aviation-bg-secondary';
    }
  };

  return (
    <div className={cn('bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center gap-2">
          <TrendingUp className="w-5 h-5 text-aviation-cyan" />
          <h3 className="text-sm font-medium text-aviation-text-primary">Optimization Hints</h3>
          <span className="ml-auto text-xs text-aviation-text-dim">{hints.length} active</span>
        </div>
      </div>

      <div className="max-h-80 overflow-y-auto">
        {hints.map((hint) => (
          <div
            key={hint.id}
            className={cn('p-4 border-b border-aviation-border-panel last:border-b-0', getTypeStyles(hint.type))}
          >
            <div className="flex items-start gap-3">
              <div className="flex-shrink-0 mt-0.5">{getTypeIcon(hint.type)}</div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between mb-1">
                  <h4 className="text-sm font-medium text-aviation-text-primary">{hint.title}</h4>
                  <button
                    onClick={() => onHintDismiss?.(hint.id)}
                    className="p-1 text-aviation-text-dim hover:text-aviation-text-primary transition-colors"
                  >
                    <X className="w-3 h-3" />
                  </button>
                </div>
                <p className="text-xs text-aviation-text-dim mb-2">{hint.description}</p>

                {hint.potentialSaving && (
                  <div className="flex items-center gap-3 mb-3 text-[10px]">
                    <span className="flex items-center gap-1 text-green-400">
                      <Timer className="w-3 h-3" />
                      Save {hint.potentialSaving.time}s
                    </span>
                    <span className="flex items-center gap-1 text-aviation-cyan">
                      <Activity className="w-3 h-3" />
                      {hint.potentialSaving.effort}% effort
                    </span>
                  </div>
                )}

                {hint.actionLabel && (
                  <button
                    onClick={() => onHintApply?.(hint)}
                    className="flex items-center gap-1.5 px-3 py-1.5 bg-aviation-cyan text-aviation-bg-primary rounded text-xs hover:bg-aviation-cyan/90 transition-colors"
                  >
                    {hint.actionLabel}
                    <ArrowRight className="w-3 h-3" />
                  </button>
                )}
              </div>
            </div>
          </div>
        ))}

        {hints.length === 0 && (
          <div className="flex flex-col items-center justify-center py-12 text-aviation-text-muted">
            <Check className="w-8 h-8 mb-2 text-green-400" />
            <p className="text-sm">No optimization hints at this time</p>
          </div>
        )}
      </div>
    </div>
  );
};

// ============================================================================
// Adaptive Layout Engine
// ============================================================================

interface LayoutConfiguration {
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

interface AdaptiveLayoutEngineProps {
  configurations: LayoutConfiguration[];
  currentConfig: LayoutConfiguration | null;
  onConfigSelect?: (config: LayoutConfiguration) => void;
  onAutoAdjust?: (skillLevel: UserSkillLevel) => void;
  className?: string;
}

export const AdaptiveLayoutEngine: React.FC<AdaptiveLayoutEngineProps> = ({
  configurations,
  currentConfig,
  onConfigSelect,
  onAutoAdjust,
  className,
}) => {
  const getSkillColor = (skill: UserSkillLevel) => {
    switch (skill) {
      case 'beginner': return 'bg-green-400';
      case 'intermediate': return 'bg-amber-400';
      case 'expert': return 'bg-aviation-cyan';
    }
  };

  return (
    <div className={cn('bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <LayoutDashboard className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Adaptive Layout</h3>
          </div>
          <button
            onClick={() => onAutoAdjust?.('beginner')}
            className="flex items-center gap-1.5 px-3 py-1.5 bg-aviation-bg-secondary text-aviation-text-primary rounded text-xs hover:bg-aviation-bg-instrument transition-colors"
          >
            <Sparkles className="w-3 h-3" />
            Auto Adjust
          </button>
        </div>
      </div>

      <div className="p-4">
        <div className="grid grid-cols-1 gap-3">
          {configurations.map((config) => {
            const isSelected = currentConfig?.id === config.id;

            return (
              <button
                key={config.id}
                onClick={() => onConfigSelect?.(config)}
                className={cn(
                  'p-4 rounded-lg border text-left transition-all',
                  isSelected
                    ? 'border-aviation-cyan bg-aviation-cyan/10'
                    : 'border-aviation-border-panel hover:border-aviation-text-dim'
                )}
              >
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm font-medium text-aviation-text-primary">{config.name}</span>
                  <div className={cn('w-2 h-2 rounded-full', getSkillColor(config.recommendedFor))} />
                </div>

                <div className="flex items-center gap-2 mb-3">
                  <span className="text-[10px] px-1.5 py-0.5 bg-aviation-bg-secondary rounded uppercase">
                    {config.complexityLevel}
                  </span>
                  <span className="text-[10px] text-aviation-text-dim capitalize">
                    for {config.recommendedFor}
                  </span>
                </div>

                <div className="flex items-center gap-1">
                  {config.panels.map((panel) => (
                    <div
                      key={panel.id}
                      className={cn(
                        'px-1.5 py-0.5 rounded text-[10px]',
                        panel.collapsed ? 'bg-aviation-bg-secondary text-aviation-text-dim' : 'bg-aviation-bg-instrument text-aviation-text-primary'
                      )}
                    >
                      {panel.position}
                    </div>
                  ))}
                </div>
              </button>
            );
          })}
        </div>
      </div>
    </div>
  );
};
