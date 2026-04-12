import { useState, useCallback, useRef, useEffect } from 'react';
import { toast } from 'sonner';
import { create } from 'zustand';

export interface UndoableAction<T = unknown> {
  id: string;
  type: string;
  description: string;
  data: T;
  timestamp: number;
  undo: () => Promise<void> | void;
  redo?: () => Promise<void> | void;
}

interface UndoRedoState {
  past: UndoableAction[];
  present: UndoableAction | null;
  future: UndoableAction[];
}

interface UseUndoRedoOptions {
  maxHistory?: number;
  showToastOnUndo?: boolean;
  showToastOnRedo?: boolean;
  toastDuration?: number;
}

interface UseUndoRedoReturn {
  actions: {
    undo: () => Promise<boolean>;
    redo: () => Promise<boolean>;
    canUndo: boolean;
    canRedo: boolean;
    clearHistory: () => void;
    addAction: (action: Omit<UndoableAction, 'timestamp'>) => void;
    getHistory: () => { past: UndoableAction[]; present: UndoableAction | null; future: UndoableAction[] };
  };
  state: {
    historyLength: number;
    pastCount: number;
    futureCount: number;
  };
}

export function useUndoRedo(options: UseUndoRedoOptions = {}): UseUndoRedoReturn {
  const {
    maxHistory = 50,
    showToastOnUndo = true,
    showToastOnRedo = false,
    toastDuration = 5000,
  } = options;

  const [state, setState] = useState<UndoRedoState>({
    past: [],
    present: null,
    future: [],
  });

  const stateRef = useRef(state);
  stateRef.current = state;

  // Cleanup timeout ref
  const undoTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    return () => {
      if (undoTimeoutRef.current) {
        clearTimeout(undoTimeoutRef.current);
      }
    };
  }, []);

  const addAction = useCallback((action: Omit<UndoableAction, 'timestamp'>) => {
    const fullAction: UndoableAction = {
      ...action,
      timestamp: Date.now(),
    };

    setState((prev) => {
      const newPast = prev.present 
        ? [...prev.past.slice(-(maxHistory - 1)), prev.present]
        : prev.past.slice(-(maxHistory - 1));
      
      return {
        past: newPast,
        present: fullAction,
        future: [],
      };
    });
  }, [maxHistory]);

  const undo = useCallback(async (): Promise<boolean> => {
    const currentState = stateRef.current;
    
    if (currentState.past.length === 0) {
      return false;
    }

    const previous = currentState.past[currentState.past.length - 1];
    const newPast = currentState.past.slice(0, -1);

    try {
      await previous.undo();

      setState({
        past: newPast,
        present: previous,
        future: currentState.present 
          ? [currentState.present, ...currentState.future].slice(0, maxHistory)
          : currentState.future.slice(0, maxHistory),
      });

      if (showToastOnUndo) {
        toast.success(`Undone: ${previous.description}`, {
          duration: toastDuration,
        });
      }

      return true;
    } catch (error) {
      toast.error('Failed to undo action', {
        description: error instanceof Error ? error.message : 'Unknown error',
        duration: toastDuration,
      });
      return false;
    }
  }, [maxHistory, showToastOnUndo, toastDuration]);

  const redo = useCallback(async (): Promise<boolean> => {
    const currentState = stateRef.current;
    
    if (currentState.future.length === 0) {
      return false;
    }

    const next = currentState.future[0];
    const newFuture = currentState.future.slice(1);

    try {
      if (next.redo) {
        await next.redo();
      }

      setState({
        past: currentState.present 
          ? [...currentState.past, currentState.present]
          : currentState.past,
        present: next,
        future: newFuture,
      });

      if (showToastOnRedo) {
        toast.success(`Redone: ${next.description}`, {
          duration: toastDuration,
        });
      }

      return true;
    } catch (error) {
      toast.error('Failed to redo action', {
        description: error instanceof Error ? error.message : 'Unknown error',
        duration: toastDuration,
      });
      return false;
    }
  }, [showToastOnRedo, toastDuration]);

  const clearHistory = useCallback(() => {
    setState({
      past: [],
      present: null,
      future: [],
    });
  }, []);

  const getHistory = useCallback(() => ({
    past: stateRef.current.past,
    present: stateRef.current.present,
    future: stateRef.current.future,
  }), []);

  return {
    actions: {
      undo,
      redo,
      canUndo: state.past.length > 0,
      canRedo: state.future.length > 0,
      clearHistory,
      addAction,
      getHistory,
    },
    state: {
      historyLength: state.past.length + (state.present ? 1 : 0) + state.future.length,
      pastCount: state.past.length,
      futureCount: state.future.length,
    },
  };
}

// Hook for managing toast-based undo for destructive actions
interface UseUndoToastOptions {
  duration?: number;
  onUndo?: () => void;
  onConfirm?: () => void;
}

export function useUndoToast() {
  const pendingUndoRef = useRef<{
    id: string;
    timeoutId: ReturnType<typeof setTimeout>;
    onUndo: () => void;
    onConfirm: () => void;
  } | null>(null);

  const showUndoToast = useCallback((
    message: string,
    options: UseUndoToastOptions = {}
  ): { dismiss: () => void; confirm: () => void } => {
    const { duration = 5000, onUndo, onConfirm } = options;
    
    // Clear any existing pending undo
    if (pendingUndoRef.current) {
      clearTimeout(pendingUndoRef.current.timeoutId);
      pendingUndoRef.current.onConfirm();
      pendingUndoRef.current = null;
    }

    const id = `undo-${Date.now()}`;
    let isUndone = false;

    const dismiss = () => {
      toast.dismiss(id);
      if (pendingUndoRef.current?.id === id) {
        clearTimeout(pendingUndoRef.current.timeoutId);
        pendingUndoRef.current = null;
      }
    };

    const confirm = () => {
      if (!isUndone && pendingUndoRef.current?.id === id) {
        clearTimeout(pendingUndoRef.current.timeoutId);
        onConfirm?.();
        pendingUndoRef.current = null;
      }
    };

    const timeoutId = setTimeout(() => {
      if (!isUndone && pendingUndoRef.current?.id === id) {
        onConfirm?.();
        pendingUndoRef.current = null;
      }
    }, duration);

    pendingUndoRef.current = {
      id,
      timeoutId,
      onUndo: () => {
        isUndone = true;
        onUndo?.();
      },
      onConfirm: () => {
        if (!isUndone) {
          onConfirm?.();
        }
      },
    };

    toast(message, {
      id,
      duration,
      action: {
        label: 'Undo',
        onClick: () => {
          isUndone = true;
          clearTimeout(timeoutId);
          onUndo?.();
          pendingUndoRef.current = null;
          toast.dismiss(id);
          toast.success('Action undone');
        },
      },
    });

    return { dismiss, confirm };
  }, []);

  return { showUndoToast };
}

// Store-based undo/redo for application-wide state
interface GlobalUndoState {
  undoableActions: Map<string, UndoableAction>;
  registerUndoable: (action: UndoableAction) => void;
  unregisterUndoable: (id: string) => void;
  executeUndo: (id: string) => Promise<boolean>;
  getUndoable: (id: string) => UndoableAction | undefined;
  clearAll: () => void;
}

export const useGlobalUndoStore = create<GlobalUndoState>()((set, get) => ({
  undoableActions: new Map(),
  
  registerUndoable: (action) => {
    set((state) => {
      const newMap = new Map(state.undoableActions);
      newMap.set(action.id, action);
      return { undoableActions: newMap };
    });
  },
  
  unregisterUndoable: (id) => {
    set((state) => {
      const newMap = new Map(state.undoableActions);
      newMap.delete(id);
      return { undoableActions: newMap };
    });
  },
  
  executeUndo: async (id) => {
    const action = get().undoableActions.get(id);
    if (!action) return false;
    
    try {
      await action.undo();
      set((state) => {
        const newMap = new Map(state.undoableActions);
        newMap.delete(id);
        return { undoableActions: newMap };
      });
      return true;
    } catch (error) {
      toast.error('Failed to undo action');
      return false;
    }
  },
  
  getUndoable: (id) => get().undoableActions.get(id),
  
  clearAll: () => set({ undoableActions: new Map() }),
}));
