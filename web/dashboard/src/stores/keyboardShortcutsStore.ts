import { create } from 'zustand';

export interface KeyboardShortcut {
  key: string;
  displayKey: string;
  ctrlKey?: boolean;
  metaKey?: boolean;
  shiftKey?: boolean;
  altKey?: boolean;
  description: string;
  category: 'navigation' | 'actions' | 'editor' | 'playground' | 'global';
  scope?: string;
}

interface KeyboardShortcutsState {
  isHelpOpen: boolean;
  shortcuts: KeyboardShortcut[];
  globalShortcuts: KeyboardShortcut[];
  setHelpOpen: (open: boolean) => void;
  toggleHelp: () => void;
  registerShortcut: (shortcut: KeyboardShortcut) => void;
  unregisterShortcut: (key: string, scope?: string) => void;
  registerGlobalShortcut: (shortcut: KeyboardShortcut) => void;
  unregisterGlobalShortcut: (key: string) => void;
  getShortcutsByCategory: (category: KeyboardShortcut['category']) => KeyboardShortcut[];
  getShortcutsForScope: (scope: string) => KeyboardShortcut[];
}

const DEFAULT_GLOBAL_SHORTCUTS: KeyboardShortcut[] = [
  {
    key: '?',
    displayKey: '?',
    shiftKey: true,
    description: 'Show keyboard shortcuts help',
    category: 'global',
  },
  {
    key: 'g',
    displayKey: 'g',
    description: 'Go to... (press g then another key)',
    category: 'navigation',
  },
  {
    key: 'gd',
    displayKey: 'g → d',
    description: 'Go to Dashboard',
    category: 'navigation',
  },
  {
    key: 'go',
    displayKey: 'g → o',
    description: 'Go to Overview',
    category: 'navigation',
  },
  {
    key: 'gf',
    displayKey: 'g → f',
    description: 'Go to Functions',
    category: 'navigation',
  },
  {
    key: 'gp',
    displayKey: 'g → p',
    description: 'Go to Providers',
    category: 'navigation',
  },
  {
    key: 'ga',
    displayKey: 'g → a',
    description: 'Go to Analytics',
    category: 'navigation',
  },
  {
    key: 'gs',
    displayKey: 'g → s',
    description: 'Go to Settings',
    category: 'navigation',
  },
  {
    key: 'k',
    displayKey: '⌘K / Ctrl+K',
    ctrlKey: true,
    metaKey: true,
    description: 'Open command palette',
    category: 'actions',
  },
  {
    key: 'n',
    displayKey: '⌘N / Ctrl+N',
    ctrlKey: true,
    metaKey: true,
    description: 'Create new item',
    category: 'actions',
  },
  {
    key: 'Escape',
    displayKey: 'Esc',
    description: 'Close modals / Cancel',
    category: 'global',
  },
  {
    key: 'Enter',
    displayKey: 'Enter',
    description: 'Submit / Confirm',
    category: 'global',
  },
];

const DEFAULT_SCOPE_SHORTCUTS: KeyboardShortcut[] = [
  // Playground shortcuts
  {
    key: 'Enter',
    displayKey: '⌘Enter / Ctrl+Enter',
    ctrlKey: true,
    metaKey: true,
    description: 'Run function',
    category: 'playground',
    scope: 'playground',
  },
  {
    key: 'f',
    displayKey: '⌘⇧F / Ctrl+Shift+F',
    ctrlKey: true,
    metaKey: true,
    shiftKey: true,
    description: 'Format JSON',
    category: 'playground',
    scope: 'playground',
  },
  {
    key: '[',
    displayKey: '⌘[ / Ctrl+[',
    ctrlKey: true,
    metaKey: true,
    description: 'Previous execution',
    category: 'playground',
    scope: 'playground',
  },
  {
    key: ']',
    displayKey: '⌘] / Ctrl+]',
    ctrlKey: true,
    metaKey: true,
    description: 'Next execution',
    category: 'playground',
    scope: 'playground',
  },
  // Editor shortcuts
  {
    key: 's',
    displayKey: '⌘S / Ctrl+S',
    ctrlKey: true,
    metaKey: true,
    description: 'Save changes',
    category: 'editor',
    scope: 'editor',
  },
  {
    key: 'z',
    displayKey: '⌘Z / Ctrl+Z',
    ctrlKey: true,
    metaKey: true,
    description: 'Undo',
    category: 'editor',
    scope: 'editor',
  },
  {
    key: 'y',
    displayKey: '⌘Y / Ctrl+Y',
    ctrlKey: true,
    metaKey: true,
    description: 'Redo',
    category: 'editor',
    scope: 'editor',
  },
  {
    key: 'z',
    displayKey: '⌘⇧Z / Ctrl+Shift+Z',
    ctrlKey: true,
    metaKey: true,
    shiftKey: true,
    description: 'Redo (alternative)',
    category: 'editor',
    scope: 'editor',
  },
  {
    key: 'f',
    displayKey: '⌘F / Ctrl+F',
    ctrlKey: true,
    metaKey: true,
    description: 'Find',
    category: 'editor',
    scope: 'editor',
  },
];

export const useKeyboardShortcutsStore = create<KeyboardShortcutsState>()((set, get) => ({
  isHelpOpen: false,
  shortcuts: DEFAULT_SCOPE_SHORTCUTS,
  globalShortcuts: DEFAULT_GLOBAL_SHORTCUTS,
  setHelpOpen: (open) => set({ isHelpOpen: open }),
  toggleHelp: () => set((state) => ({ isHelpOpen: !state.isHelpOpen })),
  registerShortcut: (shortcut) =>
    set((state) => ({
      shortcuts: [...state.shortcuts.filter(s => s.key !== shortcut.key || s.scope !== shortcut.scope), shortcut],
    })),
  unregisterShortcut: (key, scope) =>
    set((state) => ({
      shortcuts: state.shortcuts.filter(
        (s) => !(s.key === key && (!scope || s.scope === scope))
      ),
    })),
  registerGlobalShortcut: (shortcut) =>
    set((state) => ({
      globalShortcuts: [...state.globalShortcuts.filter(s => s.key !== shortcut.key), shortcut],
    })),
  unregisterGlobalShortcut: (key) =>
    set((state) => ({
      globalShortcuts: state.globalShortcuts.filter((s) => s.key !== key),
    })),
  getShortcutsByCategory: (category) => {
    const state = get();
    return [...state.globalShortcuts, ...state.shortcuts].filter(
      (s) => s.category === category
    );
  },
  getShortcutsForScope: (scope) => {
    const state = get();
    return state.shortcuts.filter((s) => s.scope === scope);
  },
}));
