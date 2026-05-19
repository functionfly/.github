import { useState, useEffect, useCallback } from "react";
import { GlassCard, Badge } from "@functionfly/ui-core";
import { Input } from "@/components/ui/input";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";
import {
  Keyboard, Search, Command, Option, ArrowUp, CornerDownLeft, Delete, MousePointer,
  ArrowDown, ArrowLeft, ArrowRight, List, X, Bookmark, ChevronsUp
} from "lucide-react";

interface ShortcutCategory {
  id: string;
  name: string;
  icon: React.ReactNode;
}

interface Shortcut {
  id: string;
  keys: string[];
  description: string;
  category: string;
  isNew?: boolean;
}

const shortcutCategories: ShortcutCategory[] = [
  { id: "global", name: "Global", icon: <Command className="w-4 h-4" /> },
  { id: "navigation", name: "Navigation", icon: <ArrowUp className="w-4 h-4" /> },
  { id: "editing", name: "Editing", icon: <Delete className="w-4 h-4" /> },
  { id: "graphs", name: "Graphs", icon: <Command className="w-4 h-4" /> },
];

const allShortcuts: Shortcut[] = [
  { id: "s-1", keys: ["⌘", "K"], description: "Open command palette", category: "global", isNew: true },
  { id: "s-2", keys: ["⌘", "S"], description: "Save workspace", category: "global" },
  { id: "s-3", keys: ["⌘", "P"], description: "Quick open files", category: "global" },
  { id: "s-4", keys: ["⌘", "⇧", "P"], description: "Open settings", category: "global" },
  { id: "s-5", keys: ["⌘", "N"], description: "New graph", category: "global", isNew: true },
  { id: "s-6", keys: ["⌘", "Z"], description: "Undo", category: "editing" },
  { id: "s-7", keys: ["⌘", "⇧", "Z"], description: "Redo", category: "editing" },
  { id: "s-8", keys: ["⌘", "C"], description: "Copy", category: "editing" },
  { id: "s-9", keys: ["⌘", "V"], description: "Paste", category: "editing" },
  { id: "s-10", keys: ["⌘", "X"], description: "Cut", category: "editing" },
  { id: "s-11", keys: ["⌘", "D"], description: "Duplicate selection", category: "editing" },
  { id: "s-12", keys: ["⌘", "F"], description: "Find in editor", category: "editing" },
  { id: "s-13", keys: ["⌘", "H"], description: "Find and replace", category: "editing" },
  { id: "s-14", keys: ["⌘", "/"], description: "Toggle comment", category: "editing" },
  { id: "s-15", keys: ["⌘", "["], description: "Indent less", category: "editing" },
  { id: "s-16", keys: ["⌘", "]"], description: "Indent more", category: "editing" },
  { id: "s-17", keys: ["↑", "↓", "←", "→"], description: "Navigate nodes", category: "navigation" },
  { id: "s-18", keys: ["Tab"], description: "Next panel", category: "navigation" },
  { id: "s-19", keys: ["⇧", "Tab"], description: "Previous panel", category: "navigation" },
  { id: "s-20", keys: ["Esc"], description: "Close panel / Cancel", category: "navigation" },
  { id: "s-21", keys: ["⌘", "W"], description: "Close tab", category: "navigation" },
  { id: "s-22", keys: ["⌘", "⇧", "T"], description: "Reopen closed tab", category: "navigation" },
  { id: "s-23", keys: ["⌘", "1-9"], description: "Switch to tab 1-9", category: "navigation" },
  { id: "s-24", keys: ["Space"], description: "Pan mode (hold)", category: "graphs" },
  { id: "s-25", keys: ["Z"], description: "Zoom mode (hold)", category: "graphs", isNew: true },
  { id: "s-26", keys: ["⌘", "Scroll"], description: "Zoom in/out", category: "graphs" },
  { id: "s-27", keys: ["⌘", "0"], description: "Reset zoom", category: "graphs" },
  { id: "s-28", keys: ["⌘", "A"], description: "Select all nodes", category: "graphs" },
  { id: "s-29", keys: ["Delete"], description: "Delete selected", category: "graphs" },
  { id: "s-30", keys: ["⌘", "G"], description: "Group selected nodes", category: "graphs", isNew: true },
  { id: "s-31", keys: ["⌘", "⌥", "G"], description: "Ungroup nodes", category: "graphs", isNew: true },
  { id: "s-32", keys: ["⌘", "E"], description: "Execute graph", category: "graphs" },
];

export function KeyboardShortcutVisualizer() {
  const [searchQuery, setSearchQuery] = useState("");
  const [activeCategory, setActiveCategory] = useState("all");
  const [pressedKeys, setPressedKeys] = useState<string[]>([]);
  const [listening, setListening] = useState(false);

  const filteredShortcuts = allShortcuts.filter((shortcut) => {
    const matchesSearch =
      shortcut.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
      shortcut.keys.join(" ").toLowerCase().includes(searchQuery.toLowerCase());
    const matchesCategory = activeCategory === "all" || shortcut.category === activeCategory;
    return matchesSearch && matchesCategory;
  });

  const keyDisplayMap: Record<string, string> = {
    "⌘": "Cmd",
    "⇧": "Shift",
    "⌥": "Opt",
    "Enter": "Return",
    " ": "Space",
  };

  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if (!listening) return;
    e.preventDefault();
    const key = keyDisplayMap[e.key] || e.key;
    setPressedKeys((prev) => {
      if (prev.includes(key)) return prev;
      return [...prev, key];
    });
  }, [listening]);

  const handleKeyUp = useCallback(() => {
    if (!listening) return;
    setPressedKeys([]);
    setListening(false);
  }, [listening]);

  useEffect(() => {
    window.addEventListener("keydown", handleKeyDown);
    window.addEventListener("keyup", handleKeyUp);
    return () => {
      window.removeEventListener("keydown", handleKeyDown);
      window.removeEventListener("keyup", handleKeyUp);
    };
  }, [handleKeyDown, handleKeyUp]);

  const formatKey = (key: string) => {
    const specialKeys: Record<string, React.ReactNode> = {
      "⌘": <Command className="w-4 h-4" />,
      "⇧": <ChevronsUp className="w-4 h-4" />,
      "⌥": <Option className="w-4 h-4" />,
      "⌃": <MousePointer className="w-4 h-4" />,
      "Enter": <CornerDownLeft className="w-4 h-4" />,
      "Delete": <Delete className="w-4 h-4" />,
      "↑": <ArrowUp className="w-4 h-4" />,
      "↓": <ArrowDown className="w-4 h-4" />,
      "←": <ArrowLeft className="w-4 h-4" />,
      "→": <ArrowRight className="w-4 h-4" />,
      "Tab": <List className="w-4 h-4" />,
      "Esc": <X className="w-4 h-4" />,
    };
    if (specialKeys[key]) return specialKeys[key];
    return <span className="text-xs font-medium">{key}</span>;
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between p-5 border-b border-white/10">
        <div>
          <h2 className="text-xl font-semibold text-white">Keyboard Shortcuts</h2>
          <p className="text-sm text-white/60">View and learn keyboard shortcuts</p>
        </div>
        <div className="flex items-center gap-3">
          <Badge variant="outline" className="text-white/60 border-white/20">
            <Keyboard className="w-3 h-3 mr-1" />
            {allShortcuts.length} shortcuts
          </Badge>
        </div>
      </div>

      <div className="px-5 pt-5">
        <div className="flex items-center gap-4 mb-4">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-white/40" />
            <Input
              placeholder="Search shortcuts..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-9 bg-white/5 border-white/10"
            />
          </div>
          <button
            onClick={() => {
              setListening(true);
              setPressedKeys([]);
            }}
            className={cn(
              "px-4 py-2 rounded-lg border transition-all duration-200",
              listening
                ? "bg-orange-500/20 border-orange-500/50 text-orange-400"
                : "bg-white/5 border-white/10 text-white/60 hover:bg-white/10 hover:text-white"
            )}
          >
            <Keyboard className="w-4 h-4 inline mr-2" />
            {listening ? "Press keys..." : "Record shortcut"}
          </button>
        </div>

        <Tabs
          value={activeCategory}
          onValueChange={setActiveCategory}
          className="w-full"
        >
          <TabsList className="inline-flex h-auto flex-wrap gap-1 rounded-xl border border-white/10 bg-white/5 p-1.5 text-white/60">
            <TabsTrigger
              value="all"
              className="gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              All
            </TabsTrigger>
            {shortcutCategories.map((cat) => (
              <TabsTrigger
                key={cat.id}
                value={cat.id}
                className="gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
              >
                {cat.icon}
                {cat.name}
              </TabsTrigger>
            ))}
          </TabsList>
        </Tabs>
      </div>

      <div className="flex-1 overflow-auto p-4">
        {listening && pressedKeys.length > 0 && (
          <GlassCard className="mb-4 p-4">
            <div className="flex items-center justify-center gap-2">
              {pressedKeys.map((key) => (
                <div
                  key={key}
                  className="w-12 h-12 rounded-lg bg-gradient-to-br from-orange-500 to-red-500 flex items-center justify-center shadow-lg shadow-orange-500/30 animate-pulse"
                >
                  {formatKey(key)}
                </div>
              ))}
            </div>
          </GlassCard>
        )}

        <div className="grid gap-3">
          {filteredShortcuts.map((shortcut) => (
            <div
              key={shortcut.id}
              className="flex items-center justify-between p-4 rounded-xl bg-white/5 border border-white/10 hover:bg-white/10 transition-colors group"
            >
              <div className="flex items-center gap-4">
                <div className="flex items-center gap-1.5">
                  {shortcut.keys.map((key, i) => (
                    <div
                      key={i}
                      className="w-10 h-10 rounded-lg bg-white/10 border border-white/20 flex items-center justify-center text-white/80 group-hover:bg-white/15 transition-colors"
                    >
                      {formatKey(key)}
                    </div>
                  ))}
                </div>
                <div>
                  <p className="text-sm font-medium text-white flex items-center gap-2">
                    {shortcut.description}
                    {shortcut.isNew && (
                      <Badge className="text-[10px] px-1.5 py-0.5 bg-orange-500/20 text-orange-400 border-orange-500/30">
                        NEW
                      </Badge>
                    )}
                  </p>
                  <p className="text-xs text-white/50 capitalize">{shortcut.category}</p>
                </div>
              </div>
              <Bookmark className="w-4 h-4 text-white/30 group-hover:text-white/60 transition-colors cursor-pointer" />
            </div>
          ))}
        </div>

        {filteredShortcuts.length === 0 && (
          <GlassCard className="flex flex-col items-center justify-center h-48 mt-4">
            <Keyboard className="w-10 h-10 text-white/30 mb-3" />
            <p className="text-white/60">No shortcuts found</p>
            <p className="text-sm text-white/40">Try a different search term</p>
          </GlassCard>
        )}
      </div>
    </div>
  );
}