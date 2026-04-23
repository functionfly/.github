import { useState, useCallback, useEffect } from 'react';
import type { FunctionGenerationResult } from '@/api/composer';

export interface GenerationHistoryItem {
  id: string;
  timestamp: number;
  description: string;
  runtime: string;
  constraints?: string;
  result: FunctionGenerationResult;
  refinementHistory: Array<{ id: string; request: string; timestamp: Date }>;
}

const HISTORY_KEY = 'ai-composer-generation-history';
const MAX_HISTORY_ITEMS = 20;

export function useGenerationHistory() {
  const [history, setHistory] = useState<GenerationHistoryItem[]>([]);

  // Load history from localStorage on mount
  useEffect(() => {
    try {
      const stored = localStorage.getItem(HISTORY_KEY);
      if (stored) {
        const parsed = JSON.parse(stored);
        // Convert timestamp strings back to Date objects for refinement history
        const restored = parsed.map((item: GenerationHistoryItem) => ({
          ...item,
          refinementHistory: (item.refinementHistory || []).map((ref: { id: string; request: string; timestamp: string }) => ({
            ...ref,
            timestamp: new Date(ref.timestamp),
          })),
        }));
        setHistory(restored);
      }
    } catch (error) {
      console.error('Failed to load generation history:', error);
    }
  }, []);

  // Save history to localStorage whenever it changes
  useEffect(() => {
    try {
      localStorage.setItem(HISTORY_KEY, JSON.stringify(history));
    } catch (error) {
      console.error('Failed to save generation history:', error);
    }
  }, [history]);

  const addToHistory = useCallback((item: Omit<GenerationHistoryItem, 'id' | 'timestamp'>) => {
    setHistory((prev) => {
      const newItem: GenerationHistoryItem = {
        ...item,
        id: crypto.randomUUID(),
        timestamp: Date.now(),
      };
      const updated = [newItem, ...prev].slice(0, MAX_HISTORY_ITEMS);
      return updated;
    });
  }, []);

  const removeFromHistory = useCallback((id: string) => {
    setHistory((prev) => prev.filter((item) => item.id !== id));
  }, []);

  const clearHistory = useCallback(() => {
    setHistory([]);
  }, []);

  const revertToGeneration = useCallback((id: string) => {
    const item = history.find((h) => h.id === id);
    if (!item) {
      throw new Error(`Generation ${id} not found in history`);
    }
    return item;
  }, [history]);

  const forkFromGeneration = useCallback((id: string): GenerationHistoryItem => {
    const item = history.find((h) => h.id === id);
    if (!item) {
      throw new Error(`Generation ${id} not found in history`);
    }
    // Create a new history item based on the forked one with a new ID
    const forked: GenerationHistoryItem = {
      ...item,
      id: crypto.randomUUID(),
      timestamp: Date.now(),
      refinementHistory: [], // Start fresh with no refinement history
    };
    // Add the forked item to history
    setHistory((prev) => [forked, ...prev].slice(0, MAX_HISTORY_ITEMS));
    return forked;
  }, [history]);

  const getGenerationById = useCallback((id: string) => {
    return history.find((h) => h.id === id);
  }, [history]);

  return {
    history,
    addToHistory,
    removeFromHistory,
    clearHistory,
    revertToGeneration,
    forkFromGeneration,
    getGenerationById,
  };
}
