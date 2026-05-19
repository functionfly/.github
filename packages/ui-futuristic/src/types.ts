/**
 * @functionfly/ui-futuristic
 * Futuristic Signature Components - Branding-Level UI
 */

// ============================================================================
// OrbitCommandLayer - Orbiting command palette with radial navigation
// ============================================================================

export interface OrbitalItem {
  id: string;
  label: string;
  icon?: string;
  angle: number;
}

export interface OrbitalLayer {
  radius: number;
  speed: number;
  items: OrbitalItem[];
}

export interface OrbitCommandLayerProps {
  layers?: OrbitalLayer[];
  activeItemId?: string | null;
  centerLabel?: string;
  isOpen?: boolean;
  onItemSelect?: (item: OrbitalItem, layer: OrbitalLayer) => void;
  onToggle?: () => void;
  className?: string;
}

// ============================================================================
// QuantumWorkspaceTransition - Quantum-inspired workspace switching
// ============================================================================

export type TransitionPhase = 'collapse' | 'teleport' | 'expand';

export interface QuantumWorkspaceTransitionProps {
  phase: TransitionPhase;
  fromWorkspace?: string;
  toWorkspace?: string;
  progress?: number;
  onPhaseComplete?: (phase: TransitionPhase) => void;
  className?: string;
}

// ============================================================================
// HolographicPanel - Holographic display effect panel
// ============================================================================

export type HolographicEffect = 'rainbow' | 'cyan' | 'magenta' | 'white';

export interface HolographicPanelProps {
  effect?: HolographicEffect;
  intensity?: number;
  children?: React.ReactNode;
  className?: string;
}

// ============================================================================
// CinematicFocusMode - Cinematic focus mode for immersive viewing
// ============================================================================

export type FocusMode = 'theater' | 'spotlight' | 'zen';

export interface CinematicFocusModeProps {
  mode?: FocusMode;
  isActive?: boolean;
  content?: React.ReactNode;
  onActivate?: () => void;
  onDeactivate?: () => void;
  className?: string;
}

// ============================================================================
// AIThoughtWave - AI thought wave visualization
// ============================================================================

export interface ThoughtWavePoint {
  timestamp: number;
  amplitude: number;
  frequency: number;
}

export interface AIThoughtWaveProps {
  points?: ThoughtWavePoint[];
  isActive?: boolean;
  color?: string;
  showGrid?: boolean;
  onPointHover?: (point: ThoughtWavePoint | null) => void;
  className?: string;
}

// ============================================================================
// GlassExecutionCard - Glass-morphism execution card
// ============================================================================

export interface ExecutionData {
  id: string;
  name: string;
  status: 'running' | 'completed' | 'failed' | 'pending';
  progress?: number;
  duration?: number;
  startTime?: number;
  metadata?: Record<string, unknown>;
}

export interface GlassExecutionCardProps {
  execution: ExecutionData;
  onClick?: (execution: ExecutionData) => void;
  onCancel?: (executionId: string) => void;
  className?: string;
}

// ============================================================================
// TokenStormRenderer - Token stream visualization
// ============================================================================

export type TokenEventType = 'input' | 'output' | 'thought';

export interface TokenEvent {
  id: string;
  type: TokenEventType;
  content: string;
  timestamp: number;
}

export interface TokenStormRendererProps {
  events?: TokenEvent[];
  isStreaming?: boolean;
  speed?: number;
  onEventClick?: (event: TokenEvent) => void;
  className?: string;
}

// ============================================================================
// SwarmMindVisualizer - Swarm intelligence visualization
// ============================================================================

export type SwarmAgentState = 'exploring' | 'exploiting' | 'returning';

export interface SwarmAgent {
  id: string;
  x: number;
  y: number;
  velocity: { dx: number; dy: number };
  state: SwarmAgentState;
}

export interface SwarmMindVisualizerProps {
  agents?: SwarmAgent[];
  targetX?: number;
  targetY?: number;
  isActive?: boolean;
  onAgentClick?: (agent: SwarmAgent) => void;
  className?: string;
}

// ============================================================================
// AmbientTelemetryLayer - Ambient background telemetry display
// ============================================================================

export type TelemetryTrend = 'up' | 'down' | 'stable';

export interface TelemetryMetric {
  label: string;
  value: number;
  unit: string;
  trend: TelemetryTrend;
}

export interface AmbientTelemetryLayerProps {
  metrics?: TelemetryMetric[];
  isActive?: boolean;
  opacity?: number;
  onMetricClick?: (metric: TelemetryMetric) => void;
  className?: string;
}

// ============================================================================
// DigitalTwinViewport - Digital twin viewport
// ============================================================================

export interface TwinEntity {
  id: string;
  type: string;
  position: { x: number; y: number; z: number };
  status: string;
}

export interface TwinConnection {
  source: string;
  target: string;
  strength: number;
}

export interface TwinState {
  entities: TwinEntity[];
  connections: TwinConnection[];
  timestamp: number;
}

export interface DigitalTwinViewportProps {
  state?: TwinState;
  selectedEntityId?: string | null;
  isAnimating?: boolean;
  cameraRotation?: { x: number; y: number };
  onEntitySelect?: (entity: TwinEntity) => void;
  onConnectionHover?: (connection: TwinConnection | null) => void;
  className?: string;
}
