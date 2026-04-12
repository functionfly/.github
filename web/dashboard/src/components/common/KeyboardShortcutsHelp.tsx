import { useKeyboardShortcutsStore, type KeyboardShortcut } from '@/stores/keyboardShortcutsStore';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Command, Compass, Edit3, Globe, Zap, ArrowRight } from 'lucide-react';

const CATEGORY_ICONS: Record<KeyboardShortcut['category'], React.ReactNode> = {
  navigation: <Compass className="h-4 w-4" />,
  actions: <Zap className="h-4 w-4" />,
  editor: <Edit3 className="h-4 w-4" />,
  playground: <Command className="h-4 w-4" />,
  global: <Globe className="h-4 w-4" />,
};

const CATEGORY_LABELS: Record<KeyboardShortcut['category'], string> = {
  navigation: 'Navigation',
  actions: 'Actions',
  editor: 'Editor',
  playground: 'Playground',
  global: 'Global',
};

function ShortcutKey({ shortcut }: { shortcut: KeyboardShortcut }) {
  const isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0;
  
  let display = shortcut.displayKey;
  
  // Replace ⌘ with Ctrl on non-Mac platforms for display
  if (!isMac) {
    display = display.replace('⌘', 'Ctrl').replace('⇧', 'Shift').replace('⌥', 'Alt');
  }
  
  return (
    <kbd className="inline-flex items-center gap-1 rounded border bg-muted px-2 py-1 text-xs font-mono font-medium">
      {display}
    </kbd>
  );
}

function ShortcutRow({ shortcut }: { shortcut: KeyboardShortcut }) {
  return (
    <div className="flex items-center justify-between py-2 border-b border-border-subtle last:border-0">
      <span className="text-sm text-text-secondary">{shortcut.description}</span>
      <ShortcutKey shortcut={shortcut} />
    </div>
  );
}

function CategorySection({ 
  category, 
  shortcuts 
}: { 
  category: KeyboardShortcut['category']; 
  shortcuts: KeyboardShortcut[] 
}) {
  if (shortcuts.length === 0) return null;
  
  return (
    <div className="mb-6">
      <div className="flex items-center gap-2 mb-3">
        <div className="p-1.5 rounded bg-accent/10 text-accent-foreground">
          {CATEGORY_ICONS[category]}
        </div>
        <h3 className="font-semibold text-text-primary">{CATEGORY_LABELS[category]}</h3>
        <Badge variant="secondary" className="ml-auto text-xs">
          {shortcuts.length}
        </Badge>
      </div>
      <div className="pl-8">
        {shortcuts.map((shortcut, index) => (
          <ShortcutRow key={`${shortcut.key}-${shortcut.scope}-${index}`} shortcut={shortcut} />
        ))}
      </div>
    </div>
  );
}

export function KeyboardShortcutsHelp() {
  const { isHelpOpen, setHelpOpen, getShortcutsByCategory } = useKeyboardShortcutsStore();

  const globalShortcuts = getShortcutsByCategory('global');
  const navigationShortcuts = getShortcutsByCategory('navigation');
  const actionShortcuts = getShortcutsByCategory('actions');
  const editorShortcuts = getShortcutsByCategory('editor');
  const playgroundShortcuts = getShortcutsByCategory('playground');

  return (
    <Dialog open={isHelpOpen} onOpenChange={setHelpOpen}>
      <DialogContent className="max-w-2xl max-h-[80vh] p-0">
        <DialogHeader className="px-6 pt-6 pb-4 border-b border-border-subtle">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-primary/10">
              <Command className="h-5 w-5 text-primary" />
            </div>
            <div>
              <DialogTitle className="text-xl">Keyboard Shortcuts</DialogTitle>
              <p className="text-sm text-text-secondary mt-1">
                Press <kbd className="px-1.5 py-0.5 rounded bg-muted text-xs">?</kbd> to show this dialog from anywhere
              </p>
            </div>
          </div>
        </DialogHeader>
        
        <ScrollArea className="px-6 py-4 max-h-[60vh]">
          <CategorySection category="global" shortcuts={globalShortcuts} />
          <CategorySection category="navigation" shortcuts={navigationShortcuts} />
          <CategorySection category="actions" shortcuts={actionShortcuts} />
          <CategorySection category="editor" shortcuts={editorShortcuts} />
          <CategorySection category="playground" shortcuts={playgroundShortcuts} />
          
          <div className="mt-6 p-4 rounded-lg bg-muted/50 border border-border-subtle">
            <h4 className="font-medium text-text-primary mb-2 flex items-center gap-2">
              <ArrowRight className="h-4 w-4" />
              Navigation Tips
            </h4>
            <ul className="text-sm text-text-secondary space-y-1.5 list-disc list-inside">
              <li>Press <kbd className="px-1 rounded bg-muted text-xs">g</kbd> followed by a letter to quickly navigate between pages</li>
              <li>Use <kbd className="px-1 rounded bg-muted text-xs">⌘K</kbd> or <kbd className="px-1 rounded bg-muted text-xs">Ctrl+K</kbd> to open the command palette</li>
              <li>Press <kbd className="px-1 rounded bg-muted text-xs">Esc</kbd> to close any open modal or dropdown</li>
              <li>In editors, use <kbd className="px-1 rounded bg-muted text-xs">⌘S</kbd> to save your changes</li>
            </ul>
          </div>
        </ScrollArea>
      </DialogContent>
    </Dialog>
  );
}
