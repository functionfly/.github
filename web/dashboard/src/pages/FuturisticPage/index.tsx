/**
 * FuturisticPage
 * Futuristic signature showcase with immersive design
 */

import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useFuturisticStore, type ThoughtWave, type Token, type SwarmAgent, type TelemetryMetric, type TwinEntity } from '@/stores/futuristicStore';
import {
  OrbitCommandLayer as OrbitCommand,
  QuantumWorkspaceTransition as QuantumTransition,
  HolographicPanel as HolographicDisplay,
  CinematicFocusMode as CinematicFocus,
  AIThoughtWave as AIThoughtWaveVisualizer,
  GlassExecutionCard,
  TokenStormRenderer as TokenStreamDisplay,
  SwarmMindVisualizer as SwarmAgentMonitor,
  AmbientTelemetryLayer as TelemetryMetricsPanel,
  DigitalTwinViewport as DigitalTwinView,
  AmbientEffects,
} from '@functionfly/ui-futuristic';
import { Sparkles, Brain, Zap, Activity, Wifi, AudioWaveform, Bot } from 'lucide-react';
import { cn } from '@/lib/utils';

const FuturisticPage: React.FC = () => {
  const { panel } = useParams<{ panel?: string }>();
  const {
    activeView,
    setActiveView,
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

  const [activeLayer, setActiveLayer] = useState('orbit');
  const [isStreaming, setIsStreaming] = useState(true);
  const [isProcessing, setIsProcessing] = useState(true);

  useEffect(() => {
    if (panel) {
      setActiveView(panel);
    }
  }, [panel, setActiveView]);

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
        value: Math.random() > 0.5 ? 'token_' + Math.random().toString(36).slice(2, 6) : ['think', 'act', 'query', 'emit'][Math.floor(Math.random() * 4)],
        type: ['text', 'tool', 'reasoning', 'control'][Math.floor(Math.random() * 4)] as Token['type'],
        timestamp: Date.now(),
      };
      pushToken(newToken);
    }, 500);
    return () => clearInterval(interval);
  }, [pushToken]);

  useEffect(() => {
    if (twinEntities.length === 0) {
      const entities: TwinEntity[] = [
        { id: '1', name: 'Auth Service', type: 'service', status: 'healthy', connections: ['2', '3'] },
        { id: '2', name: 'User API', type: 'function', status: 'healthy', connections: ['1'] },
        { id: '3', name: 'Payment Agent', type: 'agent', status: 'degraded', connections: ['1', '4'] },
        { id: '4', name: 'Database', type: 'infrastructure', status: 'healthy', connections: ['2', '3'] },
      ];
      setTwinState(entities);
    }
  }, [twinEntities.length, setTwinState]);

  useEffect(() => {
    const interval = setInterval(() => {
      const agents: SwarmAgent[] = [
        { id: '1', name: 'Coordinator', status: 'thinking', progress: 0.7, task: 'Orchestrating workflow' },
        { id: '2', name: 'Data Miner', status: 'acting', progress: 0.45, task: 'Extracting features' },
        { id: '3', name: 'Validator', status: 'waiting', progress: 0, task: 'Awaiting input' },
        { id: '4', name: 'Reporter', status: 'idle', progress: 0, task: 'Ready to report' },
      ];
      agents.forEach(agent => updateSwarmAgent(agent));
    }, 2000);
    return () => clearInterval(interval);
  }, [updateSwarmAgent]);

  useEffect(() => {
    const interval = setInterval(() => {
      const metrics: TelemetryMetric[] = [
        { id: '1', name: 'CPU Usage', value: 45 + Math.random() * 30, unit: '%', trend: 'stable', threshold: 80 },
        { id: '2', name: 'Memory', value: 2.5 + Math.random() * 2, unit: 'GB', trend: 'up', threshold: 8 },
        { id: '3', name: 'Throughput', value: 1200 + Math.random() * 400, unit: 'req/s', trend: 'up', threshold: 2000 },
        { id: '4', name: 'Latency', value: 45 + Math.random() * 30, unit: 'ms', trend: 'down', threshold: 100 },
        { id: '5', name: 'Error Rate', value: Math.random() * 2, unit: '%', trend: 'stable', threshold: 5 },
        { id: '6', name: 'Active Connections', value: 150 + Math.random() * 50, unit: 'count', trend: 'stable', threshold: 500 },
      ];
      updateTelemetry(metrics);
    }, 2000);
    return () => clearInterval(interval);
  }, [updateTelemetry]);

  const orbitItems = [
    { id: '1', label: 'Orbit', icon: <Sparkles className="w-4 h-4" />, status: 'active' as const, orbitRadius: 80 },
    { id: '2', label: 'Quantum', icon: <Zap className="w-4 h-4" />, status: 'processing' as const, orbitRadius: 110 },
    { id: '3', label: 'Neural', icon: <Brain className="w-4 h-4" />, status: 'active' as const, orbitRadius: 140 },
    { id: '4', label: 'Swarm', icon: <Bot className="w-4 h-4" />, status: 'inactive' as const, orbitRadius: 170 },
  ];

  const renderContent = () => {
    switch (activeView) {
      case 'holographic':
        return (
          <HolographicDisplay
            content={
              <div className="text-center">
                <h3 className="text-lg font-semibold text-aviation-text-primary mb-2">Holographic Display</h3>
                <p className="text-sm text-aviation-text-muted">Next-generation visualization layer</p>
              </div>
            }
            intensity={preferences.intensity}
            color={preferences.colorScheme}
            showScanline={preferences.showScanlines}
            showGlow={preferences.showGlow}
          />
        );
      case 'orbit':
        return (
          <div className="h-[400px]">
            <OrbitCommand
              items={orbitItems}
              activeItemId={orbitItems.find(item => item.label.toLowerCase() === activeLayer)?.id}
              onItemSelect={(item) => setActiveLayer(item.label.toLowerCase())}
              rotationSpeed={activeLayer === 'orbit' ? 0.2 : 0}
              showTrails
            />
          </div>
        );
      case 'quantum':
        return (
          <QuantumTransition
            states={transitionStates}
            currentStateIndex={currentStateIndex}
            onStateChange={triggerTransition}
          />
        );
      case 'thought-wave':
        return (
          <AIThoughtWaveVisualizer
            waves={thoughtWaves}
            isProcessing={isProcessing}
          />
        );
      case 'token-stream':
        return (
          <TokenStreamDisplay
            tokens={tokenStream}
            isStreaming={isStreaming}
            maxDisplay={100}
          />
        );
      case 'swarm':
        return (
          <SwarmAgentMonitor
            agents={swarmAgents}
            showConnections
          />
        );
      case 'telemetry':
        return (
          <TelemetryMetricsPanel
            metrics={telemetryMetrics}
            refreshInterval={2000}
          />
        );
      case 'twin':
        return (
          <DigitalTwinView
            entities={twinEntities}
          />
        );
      default:
        return (
          <CinematicFocus isActive={ambientEnabled} vignetteIntensity={0.4}>
            <div className="grid grid-cols-2 gap-4 h-full">
              <div className="col-span-2 h-[200px]">
                <OrbitCommand
                  items={orbitItems}
                  activeItemId={orbitItems[0].id}
                  showTrails
                />
              </div>
              <div className="h-[200px]">
                <AIThoughtWaveVisualizer
                  waves={thoughtWaves}
                  isProcessing={isProcessing}
                />
              </div>
              <div className="h-[200px]">
                <SwarmAgentMonitor
                  agents={swarmAgents}
                  showConnections
                />
              </div>
              <div className="h-[200px]">
                <TokenStreamDisplay
                  tokens={tokenStream.slice(-20)}
                  isStreaming={isStreaming}
                />
              </div>
              <div className="h-[200px]">
                <TelemetryMetricsPanel
                  metrics={telemetryMetrics.slice(0, 4)}
                />
              </div>
            </div>
          </CinematicFocus>
        );
    }
  };

  const viewTabs = [
    { id: 'overview', label: 'Overview', icon: <Sparkles className="w-4 h-4" /> },
    { id: 'orbit', label: 'Orbit Command', icon: <Zap className="w-4 h-4" /> },
    { id: 'quantum', label: 'Quantum', icon: <AudioWaveform className="w-4 h-4" /> },
    { id: 'holographic', label: 'Holographic', icon: <Activity className="w-4 h-4" /> },
    { id: 'thought-wave', label: 'Thought Wave', icon: <Brain className="w-4 h-4" /> },
    { id: 'token-stream', label: 'Token Stream', icon: <AudioWaveform className="w-4 h-4" /> },
    { id: 'swarm', label: 'Swarm', icon: <Bot className="w-4 h-4" /> },
    { id: 'telemetry', label: 'Telemetry', icon: <Activity className="w-4 h-4" /> },
    { id: 'twin', label: 'Digital Twin', icon: <Wifi className="w-4 h-4" /> },
  ];

  return (
    <div className="p-6 space-y-6 aviation-scroll">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-aviation-text-primary">Futuristic Signature</h1>
          <p className="text-sm text-aviation-text-muted">Next-generation AI visualization components</p>
        </div>
        <AmbientEffects enabled={ambientEnabled} onToggle={toggleAmbient} />
      </div>

      <div className="flex items-center gap-2 overflow-x-auto pb-2">
        {viewTabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveView(tab.id)}
            className={cn(
              'flex items-center gap-2 px-4 py-2 rounded-lg border transition-all whitespace-nowrap text-sm',
              activeView === tab.id
                ? 'bg-aviation-cyan/10 border-aviation-cyan/30 text-aviation-cyan'
                : 'bg-aviation-bg-panel border-aviation-border-panel text-aviation-text-muted hover:text-aviation-text-primary hover:border-aviation-border-hover'
            )}
          >
            {tab.icon}
            <span>{tab.label}</span>
          </button>
        ))}
      </div>

      <div className="min-h-[400px]">
        {renderContent()}
      </div>
    </div>
  );
};

export default FuturisticPage;
