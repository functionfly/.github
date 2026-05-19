/**
 * @functionfly/ui-core
 * StudioShell - The main application shell layout
 */

import * as React from "react";
import { cn } from "./utils";
import { ResizablePanelGroup, ResizablePanel, ResizableHandle } from "./ResizablePanel";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "./Tabs";
import { Panel } from "./Panel";
import { GlassCard } from "./GlassCard";
import { Badge } from "./Badge";
import { Tooltip } from "./Tooltip";
import { Spinner } from "./Spinner";
import { THEME_CONFIG, Z_INDEX_LAYERS } from "./theme";
import { Zap, LayoutDashboard, Settings, Menu, Bell, Search, Maximize2, Minus, X } from "lucide-react";
import { Sheet, SheetContent, SheetHeader, SheetTitle } from "./Sheet";

interface StudioShellContextValue {
  collapsedPanels: Record<string, boolean>;
  togglePanel: (id: string) => void;
  complexity: "beginner" | "intermediate" | "expert";
  zoom: number;
  setZoom: (zoom: number) => void;
  settingsOpen: boolean;
  setSettingsOpen: (open: boolean) => void;
  settingsContent: React.ReactNode;
  setSettingsContent: (content: React.ReactNode) => void;
}

const StudioShellContext = React.createContext<StudioShellContextValue | null>(null);

interface StudioShellProps {
  children: React.ReactNode;
  className?: string;
  settingsContent?: React.ReactNode;
}

/**
 * Top-level layout shell for FunctionFly Studio.
 * Provides panel management, adaptive complexity, and workspace layout.
 */
export function StudioShell({ children, className, settingsContent: settingsContentProp }: StudioShellProps) {
  const [collapsedPanels, setCollapsedPanels] = React.useState<Record<string, boolean>>({});
  const [complexity, setComplexity] = React.useState<"beginner" | "intermediate" | "expert">("intermediate");
  const [zoom, setZoom] = React.useState(1);
  const [settingsOpen, setSettingsOpen] = React.useState(false);
  const [settingsContentState, setSettingsContent] = React.useState<React.ReactNode>(null);
  const settingsContent = settingsContentProp ?? settingsContentState;

  const togglePanel = (id: string) => {
    setCollapsedPanels((prev) => ({ ...prev, [id]: !prev[id] }));
  };

  const contextValue = React.useMemo(
    () => ({ collapsedPanels, togglePanel, complexity, zoom, setZoom, settingsOpen, setSettingsOpen, settingsContent, setSettingsContent }),
    [collapsedPanels, togglePanel, complexity, zoom, setZoom, settingsOpen, setSettingsOpen, settingsContent, setSettingsContent]
  );

  console.log('StudioShell render: settingsContent =', settingsContent ? 'set' : 'null');

  return (
    <StudioShellContext.Provider value={contextValue}>
      <div
        className={cn(
          "flex flex-col h-screen bg-bg-primary text-text-primary overflow-hidden",
          className
        )}
      >
        {/* Title Bar */}
        <TitleBar complexity={complexity} onZoomChange={setZoom} />

        {/* Main Workspace */}
        <div className="flex-1 flex flex-col overflow-hidden">
          {children}
        </div>

        {/* Settings Sheet */}
        <Sheet open={settingsOpen} onOpenChange={setSettingsOpen}>
          <SheetContent side="right" className="w-full max-w-[90vw] sm:max-w-[700px] lg:max-w-[900px] xl:max-w-[1100px] h-full flex flex-col">
            <SheetHeader className="shrink-0 px-5 py-4 border-b border-border-subtle">
              <SheetTitle>Studio Settings</SheetTitle>
            </SheetHeader>
            <div className="flex-1 overflow-hidden">
              {settingsContent ?? (
                <div className="h-full p-6" style={{ backgroundColor: '#f59e0b' }}>
                  <h1 className="text-white text-xl font-bold">FALLBACK: No settings content set</h1>
                </div>
              )}
            </div>
          </SheetContent>
        </Sheet>
      </div>
    </StudioShellContext.Provider>
  );
}

/**
 * Window control functions
 */
function minimizeWindow() {
  const win = window as unknown as { minimize?: () => void; electron?: boolean };
  if (win.minimize) {
    win.minimize();
  }
}

function toggleMaximize(isFullscreen: boolean, setIsFullscreen: (v: boolean) => void) {
  if (isFullscreen) {
    document.exitFullscreen().catch(() => {});
    setIsFullscreen(false);
  } else {
    document.documentElement.requestFullscreen().catch(() => {
      console.log("Fullscreen not supported or blocked");
    });
    setIsFullscreen(true);
  }
}

function closeWindow() {
  const win = window as unknown as { close?: () => void };
  if (win.close) {
    win.close();
  }
}

/**
 * Title bar with branding, controls, and quick actions
 */
export function TitleBar({
  complexity,
  onZoomChange,
}: {
  complexity: string;
  onZoomChange: (zoom: number) => void;
}) {
  const [zoom, setZoom] = React.useState(1);
  const [isFullscreen, setIsFullscreen] = React.useState(false);
  const context = React.useContext(StudioShellContext);

  // Listen for fullscreen changes to sync state
  React.useEffect(() => {
    const handleFullscreenChange = () => {
      setIsFullscreen(!!document.fullscreenElement);
    };
    document.addEventListener('fullscreenchange', handleFullscreenChange);
    return () => document.removeEventListener('fullscreenchange', handleFullscreenChange);
  }, []);

  return (
    <div
      className={cn(
        "h-10 flex items-center justify-between px-4",
        "bg-bg-secondary border-b border-border-subtle",
        "select-none shrink-0"
      )}
      style={
        {
          WebkitAppRegion: "drag",
        } as React.CSSProperties
      }
    >
      <div className="flex items-center gap-3">
        {/* Logo */}
        <div className="flex items-center gap-2">
          <Zap className="size-4 text-brand-400 animate-pulse" />
          <span className="text-sm font-bold bg-gradient-to-r from-brand-400 to-brand-500 bg-clip-text text-transparent">
            FunctionFly Studio
          </span>
        </div>

        {/* Complexity toggles */}
        <div className="hidden md:flex items-center gap-1 ml-4">
          <ComplexityToggle current={complexity} onChange={() => {}} />
        </div>

        {/* Breadcrumb / File path */}
        <div className="hidden lg:flex items-center gap-1.5 ml-3 text-[11px] text-text-muted">
          <span className="hover:text-text-primary cursor-pointer">Workspace</span>
          <span className="opacity-50">/</span>
          <span className="text-text-secondary">Active Project</span>
        </div>
      </div>

      {/* Center - Search */}
      <div className="flex-1 max-w-md mx-4">
        <div
          className="flex items-center gap-2 px-3 py-1.5 bg-bg-primary rounded-lg border border-border-subtle hover:border-border-default transition-colors cursor-text"
          onClick={() => {}}
        >
          <Search className="size-4 text-text-muted" />
          <span className="text-sm text-text-muted">Quick search... (⌘K)</span>
          <kbd className="ml-auto text-[10px] text-text-muted bg-bg-tertiary px-1.5 py-0.5 rounded hidden md:inline">
            ⌘K
          </kbd>
        </div>
      </div>

      {/* Right controls */}
      <div className="flex items-center gap-1">
        <Tooltip content="Zoom Out">
          <button
            className="p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-bg-hover transition-colors"
            onClick={() => { const next = Math.max(0.5, zoom - 0.1); setZoom(next); onZoomChange(next); }}
          >
            <Minus className="size-4" />
          </button>
        </Tooltip>
        <Tooltip content="Zoom In">
          <button
            className="p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-bg-hover transition-colors"
            onClick={() => { const next = Math.min(2, zoom + 0.1); setZoom(next); onZoomChange(next); }}
          >
            <Maximize2 className="size-4" />
          </button>
        </Tooltip>

        <div className="w-px h-5 bg-border-subtle mx-1" />

        <Tooltip content="Notifications">
          <button className="p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-bg-hover relative transition-colors">
            <Bell className="size-4" />
            <span className="absolute -top-0.5 -right-0.5 size-2 rounded-full bg-brand-500" />
          </button>
        </Tooltip>
        <Tooltip content="Settings">
          <button
            className="p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-bg-hover transition-colors"
            onClick={() => context?.setSettingsOpen(true)}
          >
            <Settings className="size-4" />
          </button>
        </Tooltip>

        {/* Window controls */}
        <div className="flex ml-2" style={{ WebkitAppRegion: "no-drag" } as React.CSSProperties}>
          <Tooltip content="Minimize">
            <button
              className="p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-bg-hover transition-colors"
              onClick={minimizeWindow}
            >
              <Minus className="size-3" />
            </button>
          </Tooltip>
          <Tooltip content={isFullscreen ? "Exit Fullscreen" : "Enter Fullscreen"}>
            <button
              className="p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-bg-hover transition-colors"
              onClick={() => toggleMaximize(isFullscreen, setIsFullscreen)}
            >
              <Maximize2 className="size-3" />
            </button>
          </Tooltip>
          <Tooltip content="Close">
            <button
              className="p-1.5 rounded-md text-text-muted hover:text-error hover:bg-error/10 transition-colors"
              onClick={closeWindow}
            >
              <X className="size-3" />
            </button>
          </Tooltip>
        </div>
      </div>
    </div>
  );
}

/**
 * Complexity level toggler
 */
export function ComplexityToggle({ current, onChange }: { current: string; onChange: (level: string) => void }) {
  const levels = [
    { id: "beginner", label: "Beginner", icon: "🟢" },
    { id: "intermediate", label: "Intermediate", icon: "🟡" },
    { id: "expert", label: "Expert", icon: "🔴" },
  ];

  return (
    <div className="flex items-center gap-1">
      {levels.map((level) => (
        <Tooltip key={level.id} content={`${level.label} mode`}>
          <button
            onClick={() => onChange(level.id)}
            className={cn(
              "px-2 py-0.5 text-[10px] rounded transition-all duration-200",
              current === level.id
                ? "bg-brand-500/20 text-brand-400 font-medium border border-brand-500/30"
                : "text-text-muted hover:text-text-secondary"
            )}
          >
            {level.icon} {level.label}
          </button>
        </Tooltip>
      ))}
    </div>
  );
}

// --- Panel Config (for sidebar/content layout) ---
export interface PanelConfig {
  id: string;
  title: string;
  icon: React.ReactNode;
  component: React.ReactNode;
  defaultOpen?: boolean;
  side?: "left" | "right";
  width?: string;
  collapsible?: boolean;
}

/**
 * Sidebar panels layout
 */
export function LeftSidebar({ panels }: { panels: PanelConfig[] }) {
  return (
    <div className="w-64 shrink-0 border-r border-border-subtle bg-bg-secondary flex flex-col overflow-hidden">
      <div className="p-2 border-b border-border-subtle">
        <div className="flex items-center gap-2 px-2">
          <LayoutDashboard className="size-4 text-brand-400" />
          <span className="text-xs font-semibold text-text-muted uppercase tracking-wider">Studio</span>
        </div>
      </div>
      <div className="flex-1 overflow-y-auto p-1 space-y-1">
        {panels.map((panel) => (
          <button
            key={panel.id}
            className="w-full flex items-center gap-2 px-2.5 py-2 text-sm text-text-secondary rounded-lg hover:bg-bg-hover hover:text-text-primary transition-colors text-left"
          >
            {panel.icon}
            <span className="truncate">{panel.title}</span>
          </button>
        ))}
      </div>
    </div>
  );
}

/**
 * Right panel dock
 */
export function RightPanel({ panels }: { panels: PanelConfig[] }) {
  return (
    <div className="w-80 shrink-0 border-l border-border-subtle bg-bg-secondary flex flex-col overflow-hidden">
      {panels.map((panel) => (
        <div
          key={panel.id}
          className={cn(
            "border-b border-border-subtle",
            panel.collapsible && "transition-all duration-200"
          )}
        >
          {panel.component}
        </div>
      ))}
    </div>
  );
}

export { StudioShellContext };
