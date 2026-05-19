/**
 * FuturisticIntegration
 * Wires together all futuristic sub-components into a showcase layout
 */

import React, { useEffect, useState } from 'react';
import { useFuturisticStore, type ThoughtWave, type Token, type SwarmAgent, type TelemetryMetric, type TwinEntity } from '@/stores/futuristicStore';
import {
  HolographicDisplay,
  OrbitCommand,
  QuantumTransition,
  AIThoughtWaveVisualizer,
  TokenStreamDisplay,
  SwarmAgentMonitor,
  TelemetryMetricsPanel,
  DigitalTwinView,
  AmbientEffects,
  CinematicFocus,
} from '@functionfly/ui-futuristic';
import { Sparkles, Brain, Zap, Activity, Wifi, AudioWaveform, Bot, Settings } from 'lucide-react';
import { cn } from '@/lib/utils';

export const FuturisticIntegration: React.FC = () => {
  const {
    preferences,
    ambientEnabled,
    toggleAmbient,
    thoughtWaves,
    tokenStream,
    swarmAgents,
    telemetryMetrics,
    twinEntities,
    setTwinState,
    addThoughtWave,
    pushToken,
    updateSwarmAgent,
    updateTelemetry,
    transitionStates,
    currentStateIndex,
    triggerTransition,
  } = useFuturisticStore();

  const [isProcessing, setIsProcessing] = useState(true);
  const [isStreaming, setIsStreaming] = useState(true);
  const [showSettings, setShowSettings] = useState(false);

  useEffect(() => {
    const interval = setInterval(() => {
      const newWave: ThoughtWave = {
        id: Date.now().toString(),
        amplitude: 0.5 + Math.random() * 0.5,
        frequency: 2 + Math.random() * 4,
        phase: Math.random() * Math.PI * 2,
        timestamp: Date.now(),
      };
      addThoughtWave(newWave);
    }, 1000);
    return () => clearInterval(interval);
  }, [addThoughtWave]);

  useEffect(() => {
    const interval = setInterval(() => {
      const newToken: Token = {
        id: Date.now().toString(),
        value: Math.random() > 0.5 ? 'tok_' + Math.random().toString(36).slice(2, 5) : ['think', 'act', 'query', 'emit'][Math.floor(Math.random() * 4)],
        type: ['text', 'tool', 'reasoning', 'control'][Math.floor(Math.random() * 4)] as Token['type'],
        timestamp: Date.now(),
      };
      pushToken(newToken);
    }, 300);
    return () => clearInterval(interval);
  }, [pushToken]);

  useEffect(() => {
    if (twinEntities.length === 0) {
      const entities: TwinEntity[] = [
        { id: '1', name: 'Core API', type: 'service', status: 'healthy', connections: ['2', '3'] },
        { id: '2', name: 'Auth Module', type: 'function', status: 'healthy', connections: ['1'] },
        { id: '3', name: 'AI Engine', type: 'agent', status: 'degraded', connections: ['1', '4'] },
        { id: '4', name: 'Data Layer', type: 'infrastructure', status: 'healthy', connections: ['2'] },
        { id: '5', name: 'Cache', type: 'infrastructure', status: 'healthy', connections: ['1'] },
      ];
      setTwinState(entities);
    }
  }, [twinEntities.length, setTwinState]);

  useEffect(() => {
    const interval = setInterval(() => {
      const agents: SwarmAgent[] = [
        { id: '1', name: 'Orchestrator', status: 'thinking', progress: 0.8, task: 'Managing workflow' },
        { id: '2', name: 'Executor', status: 'acting', progress: 0.6, task: 'Running tasks' },
        { id: '3', name: 'Monitor', status: 'waiting', progress: 0, task: 'Watching metrics' },
        { id: '4', name: 'Reporter', status: 'idle', progress: 0, task: 'Preparing reports' },
        { id: '5', name: 'Guardian', status: 'acting', progress: 0.3, task: 'Validating inputs' },
      ];
      agents.forEach(agent => updateSwarmAgent(agent));
    }, 1500);
    return () => clearInterval(interval);
  }, [updateSwarmAgent]);

  useEffect(() => {
    const interval = setInterval(() => {
      const metrics: TelemetryMetric[] = [
        { id: '1', name: 'CPU', value: 45 + Math.random() * 30, unit: '%', trend: 'stable', threshold: 80 },
        { id: '2', name: 'Memory', value: 2.5 + Math.random() * 2, unit: 'GB', trend: 'up', threshold: 8 },
        { id: '3', name: 'Throughput', value: 1200 + Math.random() * 400, unit: 'req/s', trend: 'up', threshold: 2000 },
        { id: '4', name: 'Latency', value: 45 + Math.random() * 30, unit: 'ms', trend: 'down', threshold: 100 },
        { id: '5', name: 'Errors', value: Math.random() * 2, unit: '%', trend: 'stable', threshold: 5 },
        { id: '6', name: 'Connections', value: 150 + Math.random() * 50, unit: 'count', trend: 'stable', threshold: 500 },
      ];
      updateTelemetry(metrics);
    }, 1500);
    return () => clearInterval(interval);
  }, [updateTelemetry]);

  const orbitItems = [
    { id: '1', label: 'Core', icon: <Sparkles className="w-4 h-4" />, status: 'active' as const, orbitRadius: 60 },
    { id: '2', label: 'AI', icon: <Brain className="w-4 h-4" />, status: 'processing' as const, orbitRadius: 90 },
    { id: '3', label: 'Swarm', icon: <Bot className="w-4 h-4" />, status: 'active' as const, orbitRadius: 120 },
    { id: '4', label: 'Data', icon: <Activity className="w-4 h-4" />, status: 'inactive' as const, orbitRadius: 150 },
  ];

  return (
    <CinematicFocus isActive={ambientEnabled} vignetteIntensity={0.3}>
      <div className="p-6 space-y-6 aviation-scroll">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-xl font-bold text-aviation-text-primary flex items-center gap-2">
              <Sparkles className="w-5 h-5 text-aviation-cyan" />
              Futuristic Integration
            </h2>
            <p className="text-sm text-aviation-text-muted">Complete futuristic component showcase</p>
          </div>
          <div className="flex items-center gap-3">
            <AmbientEffects enabled={ambientEnabled} onToggle={toggleAmbient} />
            <button
              onClick={() => setShowSettings(!showSettings)}
              className={cn(
                'p-2 rounded-lg border transition-colors',
                showSettings
                  ? 'bg-aviation-cyan/10 border-aviation-cyan/30 text-aviation-cyan'
                  : 'bg-aviation-bg-panel border-aviation-border-panel text-aviation-text-muted hover:text-aviation-text-primary'
              )}
            >
              <Settings className="w-4 h-4" />
            </button>
          </div>
        </div>

        <div className="grid grid-cols-12 gap-4">
          <div className="col-span-12 lg:col-span-4 h-[300px]">
            <HolographicDisplay
              content={
                <div className="text-center py-4">
                  <h3 className="text-lg font-semibold text-aviation-text-primary mb-1">Holographic Core</h3>
                  <p className="text-xs text-aviation-text-muted">Premium visualization layer</p>
                  <div className="mt-4 flex items-center justify-center gap-2">
                    <div className="w-2 h-2 rounded-full bg-aviation-cyan animate-pulse" />
                    <span className="text-xs text-aviation-cyan">ACTIVE</span>
                  </div>
                </div>
              }
              intensity={preferences.intensity}
              color={preferences.colorScheme}
              showScanline={preferences.showScanlines}
              showGlow={preferences.showGlow}
            />
          </div>

          <div className="col-span-12 lg:col-span-8 h-[300px]">
            <div className="h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel p-4">
              <div className="flex items-center gap-2 mb-4">
                <Zap className="w-4 h-4 text-aviation-cyan" />
                <span className="text-sm font-medium text-aviation-text-primary">Orbit Command Center</span>
              </div>
              <div className="h-[220px]">
                <OrbitCommand
                  items={orbitItems}
                  activeItemId={orbitItems[0].id}
                  showTrails
                />
              </div>
            </div>
          </div>

          <div className="col-span-12 lg:col-span-6 h-[250px]">
            <AIThoughtWaveVisualizer
              waves={thoughtWaves}
              isProcessing={isProcessing}
            />
          </div>

          <div className="col-span-12 lg:col-span-6 h-[250px]">
            <TokenStreamDisplay
              tokens={tokenStream.slice(-30)}
              isStreaming={isStreaming}
              maxDisplay={50}
            />
          </div>

          <div className="col-span-12 lg:col-span-6 h-[280px]">
            <SwarmAgentMonitor
              agents={swarmAgents}
              showConnections
            />
          </div>

          <div className="col-span-12 lg:col-span-6 h-[280px]">
            <DigitalTwinView
              entities={twinEntities}
            />
          </div>

          <div className="col-span-12 h-[200px]">
            <TelemetryMetricsPanel
              metrics={telemetryMetrics}
              refreshInterval={1500}
            />
          </div>

          <div className="col-span-12 h-[200px]">
            <QuantumTransition
              states={transitionStates}
              currentStateIndex={currentStateIndex}
              onStateChange={triggerTransition}
            />
          </div>
        </div>
      </div>
    </CinematicFocus>
  );
};
