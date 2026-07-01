import { useCallback, useMemo } from 'react';
import { useSearchParams } from 'react-router-dom';

export type WorkspaceView =
  | 'console'
  | 'traces'
  | 'tools'
  | 'swarm'
  | 'memory'
  | 'config'
  | 'costs'
  | 'health'
  | 'daemon'
  | 'policy'
  | 'evolution';

export type RightPanelContext =
  | null
  | { type: 'execution'; id: string }
  | { type: 'tool'; name: string }
  | { type: 'swarm-node'; id: string }
  | { type: 'alert'; id: string }
  | { type: 'trace'; id: string }
  | { type: 'policy-violation'; id: string };

export function useAgentWorkspace() {
  const [searchParams, setSearchParams] = useSearchParams();

  const activeView: WorkspaceView = useMemo(() => {
    const view = searchParams.get('view');
    const validViews: WorkspaceView[] = [
      'console', 'traces', 'tools', 'swarm', 'memory',
      'config', 'costs', 'health', 'daemon', 'policy', 'evolution',
    ];
    return validViews.includes(view as WorkspaceView) ? (view as WorkspaceView) : 'console';
  }, [searchParams]);

  const rightContext: RightPanelContext = useMemo(() => {
    const ctx = searchParams.get('ctx');
    if (!ctx) return null;
    const [type, id] = ctx.split(':');
    if (!type || !id) return null;
    return { type, id } as RightPanelContext;
  }, [searchParams]);

  const setView = useCallback((view: WorkspaceView) => {
    setSearchParams(prev => {
      prev.set('view', view);
      prev.delete('ctx');
      return prev;
    }, { replace: true });
  }, [setSearchParams]);

  const setRightContext = useCallback((ctx: RightPanelContext) => {
    setSearchParams(prev => {
      if (ctx) {
        const c = ctx as { type: string; id: string };
        prev.set('ctx', `${c.type}:${c.id}`);
      } else {
        prev.delete('ctx');
      }
      return prev;
    }, { replace: true });
  }, [setSearchParams]);

  const clearRightContext = useCallback(() => {
    setSearchParams(prev => {
      prev.delete('ctx');
      return prev;
    }, { replace: true });
  }, [setSearchParams]);

  return {
    activeView,
    setView,
    rightContext,
    setRightContext,
    clearRightContext,
  };
}
