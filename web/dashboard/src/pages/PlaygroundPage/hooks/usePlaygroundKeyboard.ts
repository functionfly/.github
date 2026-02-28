import { useEffect, useCallback } from 'react';
import { usePlaygroundStore } from '../store/playgroundStore';

interface UsePlaygroundKeyboardOptions {
  author: string;
  name: string;
  onNavigateHistory?: (direction: 'prev' | 'next') => void;
}

export function usePlaygroundKeyboard({ author, name, onNavigateHistory }: UsePlaygroundKeyboardOptions) {
  const {
    execute,
    formatJson,
    resetPlayground,
    setActiveInputTab,
    setActiveOutputTab,
  } = usePlaygroundStore();

  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      const isMac = navigator.platform.toUpperCase().includes('MAC');
      const modifier = isMac ? e.metaKey : e.ctrlKey;

      if (!modifier) return;

      // Cmd+Enter — Run function
      if (e.key === 'Enter') {
        e.preventDefault();
        execute(author, name);
        return;
      }

      // Cmd+Shift+F — Format JSON
      if (e.shiftKey && e.key === 'F') {
        e.preventDefault();
        formatJson();
        return;
      }

      // Cmd+Shift+R — Reset playground
      if (e.shiftKey && e.key === 'R') {
        e.preventDefault();
        resetPlayground();
        return;
      }

      // Cmd+Shift+C — Copy shareable link
      if (e.shiftKey && e.key === 'C') {
        e.preventDefault();
        // Handled by toolbar
        return;
      }

      // Cmd+[ — Previous history item
      if (e.key === '[') {
        e.preventDefault();
        onNavigateHistory?.('prev');
        return;
      }

      // Cmd+] — Next history item
      if (e.key === ']') {
        e.preventDefault();
        onNavigateHistory?.('next');
        return;
      }

      // Cmd+1/2/3 — Switch input tabs
      if (!e.shiftKey) {
        if (e.key === '1') {
          e.preventDefault();
          setActiveInputTab('form');
          return;
        }
        if (e.key === '2') {
          e.preventDefault();
          setActiveInputTab('json');
          return;
        }
        if (e.key === '3') {
          e.preventDefault();
          setActiveInputTab('examples');
          return;
        }
      }

      // Cmd+Shift+1/2/3/4 — Switch output tabs
      if (e.shiftKey) {
        if (e.key === '1') {
          e.preventDefault();
          setActiveOutputTab('response');
          return;
        }
        if (e.key === '2') {
          e.preventDefault();
          setActiveOutputTab('headers');
          return;
        }
        if (e.key === '3') {
          e.preventDefault();
          setActiveOutputTab('timeline');
          return;
        }
        if (e.key === '4') {
          e.preventDefault();
          setActiveOutputTab('diff');
          return;
        }
      }
    },
    [author, name, execute, formatJson, resetPlayground, setActiveInputTab, setActiveOutputTab, onNavigateHistory]
  );

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);
}

export const KEYBOARD_SHORTCUTS = [
  { keys: ['⌘', 'Enter'], description: 'Run function' },
  { keys: ['⌘', '⇧', 'F'], description: 'Format JSON' },
  { keys: ['⌘', '⇧', 'R'], description: 'Reset playground' },
  { keys: ['⌘', '⇧', 'C'], description: 'Copy shareable link' },
  { keys: ['⌘', '['], description: 'Previous history item' },
  { keys: ['⌘', ']'], description: 'Next history item' },
  { keys: ['⌘', '1/2/3'], description: 'Switch input tabs' },
  { keys: ['⌘', '⇧', '1/2/3/4'], description: 'Switch output tabs' },
];
