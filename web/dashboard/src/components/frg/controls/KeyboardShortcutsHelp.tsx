/**
 * KeyboardShortcutsHelp Component
 * Modal dialog showing keyboard shortcuts for the FRG canvas
 */

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog';
import { cn } from '@/lib/utils';

interface KeyboardShortcutsHelpProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

interface Shortcut {
  keys: string[];
  description: string;
  category: 'navigation' | 'editing' | 'execution' | 'general';
}

const shortcuts: Shortcut[] = [
  {
    keys: ['Space', '+ Drag'],
    description: 'Pan canvas',
    category: 'navigation',
  },
  {
    keys: ['Scroll'],
    description: 'Zoom in/out',
    category: 'navigation',
  },
  {
    keys: ['Cmd', 'Ctrl', '+ Z'],
    description: 'Undo',
    category: 'editing',
  },
  {
    keys: ['Cmd', 'Ctrl', '+ Shift', '+ Z'],
    description: 'Redo',
    category: 'editing',
  },
  {
    keys: ['Cmd', 'Ctrl', '+ S'],
    description: 'Save graph',
    category: 'editing',
  },
  {
    keys: ['Delete', 'Backspace'],
    description: 'Delete selected node',
    category: 'editing',
  },
  {
    keys: ['Cmd', 'Ctrl', '+ Enter'],
    description: 'Run graph',
    category: 'execution',
  },
  {
    keys: ['F'],
    description: 'Fit view to all nodes',
    category: 'navigation',
  },
  {
    keys: ['?'],
    description: 'Show this help',
    category: 'general',
  },
  {
    keys: ['Esc'],
    description: 'Close dialogs / Deselect',
    category: 'general',
  },
  {
    keys: ['Tab'],
    description: 'Focus next element',
    category: 'general',
  },
  {
    keys: ['Shift', '+ Drag'],
    description: 'Select multiple nodes',
    category: 'editing',
  },
];

function Kbd({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <kbd
      className={cn(
        "inline-flex items-center justify-center px-2 py-1 text-xs font-medium",
        "bg-[var(--bg-tertiary)] border border-[var(--border-subtle)] rounded",
        "text-[var(--text-primary)] shadow-sm",
        className
      )}
    >
      {children}
    </kbd>
  );
}

function ShortcutRow({ shortcut }: { shortcut: Shortcut }) {
  return (
    <tr className="border-b border-[var(--border-subtle)] last:border-0">
      <td className="py-3 pr-4">
        <div className="flex items-center gap-1 flex-wrap">
          {shortcut.keys.map((key, index) => (
            <span key={index} className="flex items-center">
              {key === '+' ? (
                <span className="text-[var(--text-muted)] mx-1">+</span>
              ) : (
                <Kbd>{key}</Kbd>
              )}
            </span>
          ))}
        </div>
      </td>
      <td className="py-3 pl-4 text-sm text-[var(--text-secondary)]">
        {shortcut.description}
      </td>
    </tr>
  );
}

function ShortcutTable({
  title,
  shortcuts,
}: {
  title: string;
  shortcuts: Shortcut[];
}) {
  return (
    <div className="space-y-3">
      <h3 className="text-sm font-semibold text-[var(--text-primary)] uppercase tracking-wide">
        {title}
      </h3>
      <table className="w-full">
        <tbody>
          {shortcuts.map((shortcut, index) => (
            <ShortcutRow key={index} shortcut={shortcut} />
          ))}
        </tbody>
      </table>
    </div>
  );
}

export function KeyboardShortcutsHelp({ open, onOpenChange }: KeyboardShortcutsHelpProps) {
  const navigationShortcuts = shortcuts.filter((s) => s.category === 'navigation');
  const editingShortcuts = shortcuts.filter((s) => s.category === 'editing');
  const executionShortcuts = shortcuts.filter((s) => s.category === 'execution');
  const generalShortcuts = shortcuts.filter((s) => s.category === 'general');

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[80vh] overflow-auto">
        <DialogHeader>
          <DialogTitle className="text-xl font-bold text-[var(--text-primary)]">
            Keyboard Shortcuts
          </DialogTitle>
          <DialogDescription className="text-sm text-[var(--text-secondary)]">
            Master these shortcuts to work faster in the Function Runtime Graph editor
          </DialogDescription>
        </DialogHeader>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mt-4">
          <div className="space-y-6">
            <ShortcutTable title="Navigation" shortcuts={navigationShortcuts} />
            <ShortcutTable title="Execution" shortcuts={executionShortcuts} />
          </div>
          <div className="space-y-6">
            <ShortcutTable title="Editing" shortcuts={editingShortcuts} />
            <ShortcutTable title="General" shortcuts={generalShortcuts} />
          </div>
        </div>

        <div className="mt-6 pt-4 border-t border-[var(--border-subtle)]">
          <p className="text-xs text-[var(--text-muted)]">
            <Kbd>Cmd</Kbd> refers to Command key on Mac, <Kbd>Ctrl</Kbd> refers to Control key on Windows/Linux.
            Shortcuts may vary based on your operating system and browser.
          </p>
        </div>
      </DialogContent>
    </Dialog>
  );
}

export default KeyboardShortcutsHelp;
