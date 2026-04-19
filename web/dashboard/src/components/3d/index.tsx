/**
 * 3D Visualization Components for FRG Showcase
 * Wrapped with Suspense boundaries to handle async loading
 */

import { Suspense } from 'react';
import { Loader2 } from 'lucide-react';

// Loading fallback component
function SceneLoader() {
  return (
    <div className="w-full h-full min-h-[500px] flex items-center justify-center bg-black/50">
      <div className="flex flex-col items-center gap-3">
        <Loader2 className="w-8 h-8 text-brand-500 animate-spin" />
        <span className="text-white/60 text-sm">Loading 3D Scene...</span>
      </div>
    </div>
  );
}

// Wrapped exports with Suspense boundaries
import { GraphNetwork3D as GraphNetwork3DBase } from './GraphNetwork';
import { FlowingDataStream as FlowingDataStreamBase } from './FlowingDataStream';
import { CrystalGraph as CrystalGraphBase } from './CrystalGraph';
import { AnimatedNodeCluster as AnimatedNodeClusterBase } from './AnimatedNodeCluster';

export function GraphNetwork3D() {
  return (
    <Suspense fallback={<SceneLoader />}>
      <GraphNetwork3DBase />
    </Suspense>
  );
}

export function FlowingDataStream() {
  return (
    <Suspense fallback={<SceneLoader />}>
      <FlowingDataStreamBase />
    </Suspense>
  );
}

export function CrystalGraph() {
  return (
    <Suspense fallback={<SceneLoader />}>
      <CrystalGraphBase />
    </Suspense>
  );
}

export function AnimatedNodeCluster() {
  return (
    <Suspense fallback={<SceneLoader />}>
      <AnimatedNodeClusterBase />
    </Suspense>
  );
}
