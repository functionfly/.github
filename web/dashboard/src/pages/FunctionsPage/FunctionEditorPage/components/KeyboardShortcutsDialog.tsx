import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { HelpCircle, Keyboard, Save, Play, Rocket, ArrowLeft } from 'lucide-react';

interface ShortcutItem {
  keys: string[];
  description: string;
  icon?: React.ReactNode;
}

const SHORTCUTS: ShortcutItem[] = [
  { keys: ['Ctrl', 'S'], description: 'Save draft', icon: <Save className="w-4 h-4" /> },
  { keys: ['Ctrl', 'Enter'], description: 'Test function', icon: <Play className="w-4 h-4" /> },
  { keys: ['Ctrl', 'Shift', 'Enter'], description: 'Deploy function', icon: <Rocket className="w-4 h-4" /> },
  { keys: ['Esc'], description: 'Close modals / Exit fullscreen' },
  { keys: ['Ctrl', 'K'], description: 'Open command palette (future)' },
];

function formatKey(key: string): string {
  const isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0;
  if (key === 'Ctrl') return isMac ? '⌘' : 'Ctrl';
  if (key === 'Alt') return isMac ? '⌥' : 'Alt';
  return key;
}

export function KeyboardShortcutsDialog({ children }: { children?: React.ReactNode }) {
  return (
    <Dialog>
      <DialogTrigger asChild>
        {children || (
          <button
            className="flex items-center gap-1.5 text-xs text-text-muted hover:text-text-primary transition-colors"
            aria-label="Keyboard shortcuts"
          >
            <Keyboard className="w-3.5 h-3.5" />
            <span className="hidden sm:inline">Shortcuts</span>
            <kbd className="hidden md:inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded bg-bg-tertiary border border-border-subtle text-[10px] font-mono">
              ?
            </kbd>
          </button>
        )}
      </DialogTrigger>
      <DialogContent className="sm:max-w-md" style={{ background: 'var(--bg-secondary)' }}>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 text-base font-display">
            <Keyboard className="w-5 h-5 text-[#FF6B35]" />
            Keyboard Shortcuts
          </DialogTitle>
        </DialogHeader>
        <div className="space-y-3 mt-4">
          {SHORTCUTS.map((shortcut, idx) => (
            <div
              key={idx}
              className="flex items-center justify-between py-2 border-b border-border-subtle/30 last:border-0"
            >
              <div className="flex items-center gap-3">
                {shortcut.icon && <span className="text-text-muted">{shortcut.icon}</span>}
                <span className="text-sm text-text-secondary">{shortcut.description}</span>
              </div>
              <div className="flex items-center gap-1">
                {shortcut.keys.map((key, kidx) => (
                  <kbd
                    key={kidx}
                    className="px-2 py-1 rounded bg-bg-tertiary border border-border-subtle text-xs font-mono text-text-primary min-w-[28px] text-center"
                  >
                    {formatKey(key)}
                  </kbd>
                ))}
              </div>
            </div>
          ))}
        </div>
        <div className="mt-4 pt-4 border-t border-border-subtle/50">
          <p className="text-xs text-text-muted flex items-start gap-2">
            <HelpCircle className="w-4 h-4 shrink-0 mt-0.5" />
            Press <kbd className="px-1.5 py-0.5 rounded bg-bg-tertiary border border-border-subtle text-[10px] font-mono">?</kbd> anytime from the editor to open this dialog.
          </p>
        </div>
      </DialogContent>
    </Dialog>
  );
}

export function useKeyboardShortcuts(callback: (shortcut: string) => void) {
  // Hook for listening to keyboard shortcuts
  // Usage in component:
  // useKeyboardShortcuts((shortcut) => {
  //   if (shortcut === 'ctrl+s') handleSave();
  // });
}
