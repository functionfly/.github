/**
 * FuturisticIntegration
 * Wires together all futuristic sub-components into a showcase layout
 */

import React, { useEffect, useState } from 'react';
import { useFuturisticStore, type ThoughtWave, type Token } from '@/stores/futuristicStore';
import {
  HolographicPanel,
  OrbitCommandLayer,
  QuantumWorkspaceTransition,
  AIThoughtWave,
  TokenStormRenderer,
  SwarmMindVisualizer,
  AmbientTelemetryLayer,
  DigitalTwinViewport,
  CinematicFocusMode,
} from '@functionfly/ui-futuristic';
import { Sparkles, Brain, Zap, Activity, Bot, Settings } from 'lucide-react';
import { cn } from '@/lib/utils';

export const FuturisticIntegration: React.FC = () => {
  const {
    preferences,
    ambientEnabled,
    thoughtWaves,
    tokenStream,
    swarmAgents,
    twinEntities,
  } = useFuturisticStore();

  const [isActive, setIsActive] = useState(ambientEnabled);

  useEffect(() => {
    setIsActive(ambientEnabled);
  }, [ambientEnabled]);

  const thoughtWavePoints = thoughtWaves.map(w => ({
    timestamp: w.timestamp,
    amplitude: w.amplitude,
    frequency: w.frequency,
  }));

  const tokenEvents = tokenStream.slice(-30).map(t => ({
    id: t.id,
    type: (t.type === 'text' ? 'output' : t.type === 'tool' ? 'input' : 'thought') as 'input' | 'output' | 'thought',
    content: t.value,
    timestamp: t.timestamp,
  }));

  const swarmVisualizerAgents = swarmAgents.map(a => ({
    id: a.id,
    x: 0,
    y: 0,
    velocity: { dx: 0, dy: 0 },
    state: 'exploring' as const,
  }));

  return (
    <CinematicFocusMode
      mode="theater"
      isActive={isActive}
      onActivate={() => setIsActive(true)}
      onDeactivate={() => setIsActive(false)}
      content={
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
              <button
                onClick={() => setIsActive(!isActive)}
                className={cn(
                  'p-2 rounded-lg border transition-colors',
                  isActive
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
              <HolographicPanel
                effect="cyan"
                intensity={preferences.intensity === 'low' ? 0.3 : preferences.intensity === 'medium' ? 0.6 : 0.9}
              >
                <div className="text-center py-4">
                  <h3 className="text-lg font-semibold text-aviation-text-primary mb-1">Holographic Core</h3>
                  <p className="text-xs text-aviation-text-muted">Premium visualization layer</p>
                  <div className="mt-4 flex items-center justify-center gap-2">
                    <div className="w-2 h-2 rounded-full bg-aviation-cyan animate-pulse" />
                    <span className="text-xs text-aviation-cyan">ACTIVE</span>
                  </div>
                </div>
              </HolographicPanel>
            </div>

            <div className="col-span-12 lg:col-span-8 h-[300px]">
              <div className="h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel p-4">
                <div className="flex items-center gap-2 mb-4">
                  <Zap className="w-4 h-4 text-aviation-cyan" />
                  <span className="text-sm font-medium text-aviation-text-primary">Orbit Command Center</span>
                </div>
                <div className="h-[220px]">
                  <OrbitCommandLayer
                    layers={[{
                      radius: 80,
                      speed: 1,
                      items: [
                        { id: '1', label: 'Core', icon: 'sparkles', angle: 0 },
                        { id: '2', label: 'AI', icon: 'brain', angle: 90 },
                        { id: '3', label: 'Swarm', icon: 'bot', angle: 180 },
                        { id: '4', label: 'Data', icon: 'activity', angle: 270 },
                      ],
                    }]}
                    activeItemId="1"
                    isOpen={true}
                  />
                </div>
              </div>
            </div>

            <div className="col-span-12 lg:col-span-6 h-[250px]">
              <AIThoughtWave
                points={thoughtWavePoints}
                isActive={isActive}
                showGrid={true}
              />
            </div>

            <div className="col-span-12 lg:col-span-6 h-[250px]">
              <TokenStormRenderer
                events={tokenEvents}
                isStreaming={true}
              />
            </div>

            <div className="col-span-12 lg:col-span-6 h-[280px]">
              <SwarmMindVisualizer
                agents={swarmVisualizerAgents}
              />
            </div>

          <div className="col-span-12 lg:col-span-6 h-[280px]">
            <DigitalTwinViewport
              state={{
                entities: twinEntities.map((e, i) => ({
                  id: e.id,
                  type: e.type,
                  status: e.status,
                  position: { x: i * 50, y: i * 30, z: 0 },
                })),
                connections: [],
                timestamp: Date.now(),
              }}
              isAnimating={true}
            />
          </div>

            <div className="col-span-12 h-[200px]">
              <AmbientTelemetryLayer isActive={isActive} />
            </div>

            <div className="col-span-12 h-[200px]">
              <QuantumWorkspaceTransition
                phase="expand"
                progress={0.8}
              />
            </div>
          </div>
        </div>
      }
    />
  );
};
