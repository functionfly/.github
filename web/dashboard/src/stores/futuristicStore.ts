/**
 * Futuristic Store
 * Global state management for futuristic signature components
 */

import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

// ============================================================================
// Types
// ============================================================================

export type TransitionPhase = 'idle' | 'entering' | 'stable' | 'exiting';
export type FocusMode = 'cinematic' | 'ambient' | 'deep-focus';
export type AmbientStatus = 'inactive' | 'active' | 'transitional';

export interface ThoughtWave {
  id: string;
  amplitude: number;
  frequency: number;
  phase: number;
  timestamp: number;
}

export interface Token {
  id: string;
  value: string;
  type: 'text' | 'tool' | 'reasoning' | 'control';
  timestamp: number;
}

export interface SwarmAgent {
  id: string;
  name: string;
  status: 'idle' | 'thinking' | 'acting' | 'waiting';
  progress: number;
  task?: string;
}

export interface TelemetryMetric {
  id: string;
  name: string;
  value: number;
  unit: string;
  trend?: 'up' | 'down' | 'stable';
  threshold?: number;
}

export interface TwinEntity {
  id: string;
  name: string;
  type: 'function' | 'agent' | 'service' | 'infrastructure';
  status: 'healthy' | 'degraded' | 'failed' | 'unknown';
  metrics?: Record<string, number>;
  connections?: string[];
}

export interface TransitionState {
  phase: TransitionPhase;
  progress: number;
  content?: string;
}

export interface FuturisticPreferences {
  intensity: 'low' | 'medium' | 'high';
  colorScheme: 'cyan' | 'amber' | 'purple' | 'green';
  showScanlines: boolean;
  showGlow: boolean;
  animationSpeed: 'slow' | 'normal' | 'fast';
}

// ============================================================================
// State Interface
// ============================================================================

export interface FuturisticState {
  // Layer State
  activeLayer: string;
  
  // Transition State
  transitionPhase: TransitionPhase;
  transitionStates: TransitionState[];
  currentStateIndex: number;
  
  // Focus Mode
  focusMode: FocusMode;
  
  // Thought Waves
  thoughtWaves: ThoughtWave[];
  
  // Token Stream
  tokenStream: Token[];
  
  // Swarm Agents
  swarmAgents: SwarmAgent[];
  
  // Telemetry
  telemetryMetrics: TelemetryMetric[];
  
  // Digital Twin
  twinEntities: TwinEntity[];
  selectedTwinEntityId: string | null;
  
  // Ambient State
  ambientEnabled: boolean;
  ambientStatus: AmbientStatus;
  
  // Preferences
  preferences: FuturisticPreferences;
  
  // UI State
  activeView: string;
}

// ============================================================================
// Actions
// ============================================================================

interface FuturisticActions {
  // Layer Actions
  setActiveLayer: (layer: string) => void;
  
  // Transition Actions
  setTransitionPhase: (phase: TransitionPhase) => void;
  triggerTransition: (targetState: number) => void;
  
  // Focus Mode Actions
  setFocusMode: (mode: FocusMode) => void;
  
  // Thought Wave Actions
  addThoughtWave: (wave: ThoughtWave) => void;
  clearThoughtWaves: () => void;
  
  // Token Stream Actions
  pushToken: (token: Token) => void;
  clearTokens: () => void;
  
  // Swarm Agent Actions
  updateSwarmAgent: (agent: SwarmAgent) => void;
  removeSwarmAgent: (agentId: string) => void;
  
  // Telemetry Actions
  updateTelemetry: (metrics: TelemetryMetric[]) => void;
  
  // Digital Twin Actions
  setTwinState: (entities: TwinEntity[]) => void;
  selectTwinEntity: (entityId: string | null) => void;
  
  // Ambient Actions
  toggleAmbient: () => void;
  setAmbientStatus: (status: AmbientStatus) => void;
  
  // Preference Actions
  setPreference: <K extends keyof FuturisticPreferences>(
    key: K,
    value: FuturisticPreferences[K]
  ) => void;
  
  // UI Actions
  setActiveView: (view: string) => void;
}

export type FuturisticStore = FuturisticState & FuturisticActions;

// ============================================================================
// Store
// ============================================================================

export const useFuturisticStore = create<FuturisticStore>()(
  immer((set) => ({
    // ============================================================================
    // Initial State
    // ============================================================================

    activeLayer: 'orbit',
    transitionPhase: 'idle',
    transitionStates: [
      { phase: 'idle', progress: 0 },
      { phase: 'entering', progress: 0.5 },
      { phase: 'stable', progress: 1 },
      { phase: 'exiting', progress: 0.5 },
    ],
    currentStateIndex: 0,
    focusMode: 'ambient',
    thoughtWaves: [],
    tokenStream: [],
    swarmAgents: [],
    telemetryMetrics: [],
    twinEntities: [],
    selectedTwinEntityId: null,
    ambientEnabled: true,
    ambientStatus: 'active',
    preferences: {
      intensity: 'medium',
      colorScheme: 'cyan',
      showScanlines: true,
      showGlow: true,
      animationSpeed: 'normal',
    },
    activeView: 'overview',

    // ============================================================================
    // Layer Actions
    // ============================================================================

    setActiveLayer: (layer) =>
      set((state) => {
        state.activeLayer = layer;
      }),

    // ============================================================================
    // Transition Actions
    // ============================================================================

    setTransitionPhase: (phase) =>
      set((state) => {
        state.transitionPhase = phase;
        state.transitionStates[state.currentStateIndex].phase = phase;
      }),

    triggerTransition: (targetState) =>
      set((state) => {
        state.transitionPhase = 'entering';
        state.transitionStates[state.currentStateIndex].phase = 'exiting';
        state.transitionStates[state.currentStateIndex].progress = 0;
        state.currentStateIndex = targetState;
        state.transitionStates[targetState].phase = 'entering';
        state.transitionStates[targetState].progress = 0.5;
      }),

    // ============================================================================
    // Focus Mode Actions
    // ============================================================================

    setFocusMode: (mode) =>
      set((state) => {
        state.focusMode = mode;
      }),

    // ============================================================================
    // Thought Wave Actions
    // ============================================================================

    addThoughtWave: (wave) =>
      set((state) => {
        state.thoughtWaves.push(wave);
        if (state.thoughtWaves.length > 10) {
          state.thoughtWaves = state.thoughtWaves.slice(-10);
        }
      }),

    clearThoughtWaves: () =>
      set((state) => {
        state.thoughtWaves = [];
      }),

    // ============================================================================
    // Token Stream Actions
    // ============================================================================

    pushToken: (token) =>
      set((state) => {
        state.tokenStream.push(token);
        if (state.tokenStream.length > 200) {
          state.tokenStream = state.tokenStream.slice(-200);
        }
      }),

    clearTokens: () =>
      set((state) => {
        state.tokenStream = [];
      }),

    // ============================================================================
    // Swarm Agent Actions
    // ============================================================================

    updateSwarmAgent: (agent) =>
      set((state) => {
        const index = state.swarmAgents.findIndex((a) => a.id === agent.id);
        if (index >= 0) {
          state.swarmAgents[index] = agent;
        } else {
          state.swarmAgents.push(agent);
        }
      }),

    removeSwarmAgent: (agentId) =>
      set((state) => {
        state.swarmAgents = state.swarmAgents.filter((a) => a.id !== agentId);
      }),

    // ============================================================================
    // Telemetry Actions
    // ============================================================================

    updateTelemetry: (metrics) =>
      set((state) => {
        state.telemetryMetrics = metrics;
      }),

    // ============================================================================
    // Digital Twin Actions
    // ============================================================================

    setTwinState: (entities) =>
      set((state) => {
        state.twinEntities = entities;
      }),

    selectTwinEntity: (entityId) =>
      set((state) => {
        state.selectedTwinEntityId = entityId;
      }),

    // ============================================================================
    // Ambient Actions
    // ============================================================================

    toggleAmbient: () =>
      set((state) => {
        state.ambientEnabled = !state.ambientEnabled;
        state.ambientStatus = state.ambientEnabled ? 'active' : 'inactive';
      }),

    setAmbientStatus: (status) =>
      set((state) => {
        state.ambientStatus = status;
      }),

    // ============================================================================
    // Preference Actions
    // ============================================================================

    setPreference: (key, value) =>
      set((state) => {
        state.preferences[key] = value;
      }),

    // ============================================================================
    // UI Actions
    // ============================================================================

    setActiveView: (view) =>
      set((state) => {
        state.activeView = view;
      }),
  }))
);

// ============================================================================
// Selectors
// ============================================================================

export const selectActiveComponents = (state: FuturisticStore) => ({
  activeLayer: state.activeLayer,
  transitionPhase: state.transitionPhase,
  focusMode: state.focusMode,
  ambientEnabled: state.ambientEnabled,
});

export const selectAmbientMetrics = (state: FuturisticStore) => ({
  enabled: state.ambientEnabled,
  status: state.ambientStatus,
  metrics: state.telemetryMetrics,
});

export const selectSwarmStatus = (state: FuturisticStore) => ({
  agents: state.swarmAgents,
  agentCount: state.swarmAgents.length,
});

export const selectTwinEntities = (state: FuturisticStore) => ({
  entities: state.twinEntities,
  selectedEntityId: state.selectedTwinEntityId,
});

// ============================================================================
// Custom Hooks
// ============================================================================

export const useFuturistic = () =>
  useFuturisticStore((state) => ({
    activeLayer: state.activeLayer,
    transitionPhase: state.transitionPhase,
    focusMode: state.focusMode,
    ambientEnabled: state.ambientEnabled,
    preferences: state.preferences,
    activeView: state.activeView,
    setActiveLayer: state.setActiveLayer,
    triggerTransition: state.triggerTransition,
    setFocusMode: state.setFocusMode,
    toggleAmbient: state.toggleAmbient,
    setActiveView: state.setActiveView,
  }));

export const useOrbitCommand = () =>
  useFuturisticStore((state) => ({
    activeLayer: state.activeLayer,
    setActiveLayer: state.setActiveLayer,
  }));

export const useQuantumTransition = () =>
  useFuturisticStore((state) => ({
    transitionPhase: state.transitionPhase,
    transitionStates: state.transitionStates,
    currentStateIndex: state.currentStateIndex,
    setTransitionPhase: state.setTransitionPhase,
    triggerTransition: state.triggerTransition,
  }));

export const useHolographic = () =>
  useFuturisticStore((state) => ({
    preferences: state.preferences,
    setPreference: state.setPreference,
  }));

export const useCinematicFocus = () =>
  useFuturisticStore((state) => ({
    focusMode: state.focusMode,
    setFocusMode: state.setFocusMode,
  }));

export const useAIThoughtWave = () =>
  useFuturisticStore((state) => ({
    thoughtWaves: state.thoughtWaves,
    addThoughtWave: state.addThoughtWave,
    clearThoughtWaves: state.clearThoughtWaves,
  }));

export const useTokenStorm = () =>
  useFuturisticStore((state) => ({
    tokenStream: state.tokenStream,
    pushToken: state.pushToken,
    clearTokens: state.clearTokens,
  }));

export const useSwarmMind = () =>
  useFuturisticStore((state) => ({
    swarmAgents: state.swarmAgents,
    updateSwarmAgent: state.updateSwarmAgent,
    removeSwarmAgent: state.removeSwarmAgent,
  }));

export const useAmbientTelemetry = () =>
  useFuturisticStore((state) => ({
    telemetryMetrics: state.telemetryMetrics,
    updateTelemetry: state.updateTelemetry,
    ambientEnabled: state.ambientEnabled,
    ambientStatus: state.ambientStatus,
    toggleAmbient: state.toggleAmbient,
  }));

export const useDigitalTwin = () =>
  useFuturisticStore((state) => ({
    twinEntities: state.twinEntities,
    selectedTwinEntityId: state.selectedTwinEntityId,
    setTwinState: state.setTwinState,
    selectTwinEntity: state.selectTwinEntity,
  }));
