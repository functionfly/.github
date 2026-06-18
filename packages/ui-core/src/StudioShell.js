import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
/**
 * @functionfly/ui-core
 * StudioShell - The main application shell layout
 */
import * as React from "react";
import { cn } from "./utils";
import { Tooltip } from "./Tooltip";
import { Zap, LayoutDashboard, Settings, Bell, Search, Maximize2, Minus, X } from "lucide-react";
import { Sheet, SheetContent, SheetHeader, SheetTitle } from "./Sheet";
const StudioShellContext = React.createContext(null);
/**
 * Top-level layout shell for FunctionFly Studio.
 * Provides panel management, adaptive complexity, and workspace layout.
 */
export function StudioShell({ children, className, settingsContent: settingsContentProp }) {
    const [collapsedPanels, setCollapsedPanels] = React.useState({});
    const [complexity, setComplexity] = React.useState("intermediate");
    const [zoom, setZoom] = React.useState(1);
    const [settingsOpen, setSettingsOpen] = React.useState(false);
    const [settingsContentState, setSettingsContent] = React.useState(null);
    const settingsContent = settingsContentProp ?? settingsContentState;
    const togglePanel = (id) => {
        setCollapsedPanels((prev) => ({ ...prev, [id]: !prev[id] }));
    };
    const contextValue = React.useMemo(() => ({ collapsedPanels, togglePanel, complexity, zoom, setZoom, settingsOpen, setSettingsOpen, settingsContent, setSettingsContent }), [collapsedPanels, togglePanel, complexity, zoom, setZoom, settingsOpen, setSettingsOpen, settingsContent, setSettingsContent]);
    console.log('StudioShell render: settingsContent =', settingsContent ? 'set' : 'null');
    return (_jsx(StudioShellContext.Provider, { value: contextValue, children: _jsxs("div", { className: cn("flex flex-col h-screen bg-bg-primary text-text-primary overflow-hidden", className), children: [_jsx(TitleBar, { complexity: complexity, onZoomChange: setZoom }), _jsx("div", { className: "flex-1 flex flex-col overflow-hidden", children: children }), _jsx(Sheet, { open: settingsOpen, onOpenChange: setSettingsOpen, children: _jsxs(SheetContent, { side: "right", className: "w-full max-w-[90vw] sm:max-w-[700px] lg:max-w-[900px] xl:max-w-[1100px] h-full flex flex-col", children: [_jsx(SheetHeader, { className: "shrink-0 px-5 py-4 border-b border-border-subtle", children: _jsx(SheetTitle, { children: "Studio Settings" }) }), _jsx("div", { className: "flex-1 overflow-hidden", children: settingsContent ?? (_jsx("div", { className: "h-full p-6", style: { backgroundColor: '#f59e0b' }, children: _jsx("h1", { className: "text-white text-xl font-bold", children: "FALLBACK: No settings content set" }) })) })] }) })] }) }));
}
/**
 * Window control functions
 */
function minimizeWindow() {
    const win = window;
    if (win.minimize) {
        win.minimize();
    }
}
function toggleMaximize(isFullscreen, setIsFullscreen) {
    if (isFullscreen) {
        document.exitFullscreen().catch(() => { });
        setIsFullscreen(false);
    }
    else {
        document.documentElement.requestFullscreen().catch(() => {
            console.log("Fullscreen not supported or blocked");
        });
        setIsFullscreen(true);
    }
}
function closeWindow() {
    const win = window;
    if (win.close) {
        win.close();
    }
}
/**
 * Title bar with branding, controls, and quick actions
 */
export function TitleBar({ complexity, onZoomChange, }) {
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
    return (_jsxs("div", { className: cn("h-10 flex items-center justify-between px-4", "bg-bg-secondary border-b border-border-subtle", "select-none shrink-0"), style: {
            WebkitAppRegion: "drag",
        }, children: [_jsxs("div", { className: "flex items-center gap-3", children: [_jsxs("div", { className: "flex items-center gap-2", children: [_jsx(Zap, { className: "size-4 text-brand-400 animate-pulse" }), _jsx("span", { className: "text-sm font-bold bg-gradient-to-r from-brand-400 to-brand-500 bg-clip-text text-transparent", children: "FunctionFly Studio" })] }), _jsx("div", { className: "hidden md:flex items-center gap-1 ml-4", children: _jsx(ComplexityToggle, { current: complexity, onChange: () => { } }) }), _jsxs("div", { className: "hidden lg:flex items-center gap-1.5 ml-3 text-[11px] text-text-muted", children: [_jsx("span", { className: "hover:text-text-primary cursor-pointer", children: "Workspace" }), _jsx("span", { className: "opacity-50", children: "/" }), _jsx("span", { className: "text-text-secondary", children: "Active Project" })] })] }), _jsx("div", { className: "flex-1 max-w-md mx-4", children: _jsxs("div", { className: "flex items-center gap-2 px-3 py-1.5 bg-bg-primary rounded-lg border border-border-subtle hover:border-border-default transition-colors cursor-text", onClick: () => { }, children: [_jsx(Search, { className: "size-4 text-text-muted" }), _jsx("span", { className: "text-sm text-text-muted", children: "Quick search... (\u2318K)" }), _jsx("kbd", { className: "ml-auto text-[10px] text-text-muted bg-bg-tertiary px-1.5 py-0.5 rounded hidden md:inline", children: "\u2318K" })] }) }), _jsxs("div", { className: "flex items-center gap-1", children: [_jsx(Tooltip, { content: "Zoom Out", children: _jsx("button", { className: "p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-bg-hover transition-colors", onClick: () => { const next = Math.max(0.5, zoom - 0.1); setZoom(next); onZoomChange(next); }, children: _jsx(Minus, { className: "size-4" }) }) }), _jsx(Tooltip, { content: "Zoom In", children: _jsx("button", { className: "p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-bg-hover transition-colors", onClick: () => { const next = Math.min(2, zoom + 0.1); setZoom(next); onZoomChange(next); }, children: _jsx(Maximize2, { className: "size-4" }) }) }), _jsx("div", { className: "w-px h-5 bg-border-subtle mx-1" }), _jsx(Tooltip, { content: "Notifications", children: _jsxs("button", { className: "p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-bg-hover relative transition-colors", children: [_jsx(Bell, { className: "size-4" }), _jsx("span", { className: "absolute -top-0.5 -right-0.5 size-2 rounded-full bg-brand-500" })] }) }), _jsx(Tooltip, { content: "Settings", children: _jsx("button", { className: "p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-bg-hover transition-colors", onClick: () => context?.setSettingsOpen(true), children: _jsx(Settings, { className: "size-4" }) }) }), _jsxs("div", { className: "flex ml-2", style: { WebkitAppRegion: "no-drag" }, children: [_jsx(Tooltip, { content: "Minimize", children: _jsx("button", { className: "p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-bg-hover transition-colors", onClick: minimizeWindow, children: _jsx(Minus, { className: "size-3" }) }) }), _jsx(Tooltip, { content: isFullscreen ? "Exit Fullscreen" : "Enter Fullscreen", children: _jsx("button", { className: "p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-bg-hover transition-colors", onClick: () => toggleMaximize(isFullscreen, setIsFullscreen), children: _jsx(Maximize2, { className: "size-3" }) }) }), _jsx(Tooltip, { content: "Close", children: _jsx("button", { className: "p-1.5 rounded-md text-text-muted hover:text-error hover:bg-error/10 transition-colors", onClick: closeWindow, children: _jsx(X, { className: "size-3" }) }) })] })] })] }));
}
/**
 * Complexity level toggler
 */
export function ComplexityToggle({ current, onChange }) {
    const levels = [
        { id: "beginner", label: "Beginner", icon: "🟢" },
        { id: "intermediate", label: "Intermediate", icon: "🟡" },
        { id: "expert", label: "Expert", icon: "🔴" },
    ];
    return (_jsx("div", { className: "flex items-center gap-1", children: levels.map((level) => (_jsx(Tooltip, { content: `${level.label} mode`, children: _jsxs("button", { onClick: () => onChange(level.id), className: cn("px-2 py-0.5 text-[10px] rounded transition-all duration-200", current === level.id
                    ? "bg-brand-500/20 text-brand-400 font-medium border border-brand-500/30"
                    : "text-text-muted hover:text-text-secondary"), children: [level.icon, " ", level.label] }) }, level.id))) }));
}
/**
 * Sidebar panels layout
 */
export function LeftSidebar({ panels }) {
    return (_jsxs("div", { className: "w-64 shrink-0 border-r border-border-subtle bg-bg-secondary flex flex-col overflow-hidden", children: [_jsx("div", { className: "p-2 border-b border-border-subtle", children: _jsxs("div", { className: "flex items-center gap-2 px-2", children: [_jsx(LayoutDashboard, { className: "size-4 text-brand-400" }), _jsx("span", { className: "text-xs font-semibold text-text-muted uppercase tracking-wider", children: "Studio" })] }) }), _jsx("div", { className: "flex-1 overflow-y-auto p-1 space-y-1", children: panels.map((panel) => (_jsxs("button", { className: "w-full flex items-center gap-2 px-2.5 py-2 text-sm text-text-secondary rounded-lg hover:bg-bg-hover hover:text-text-primary transition-colors text-left", children: [panel.icon, _jsx("span", { className: "truncate", children: panel.title })] }, panel.id))) })] }));
}
/**
 * Right panel dock
 */
export function RightPanel({ panels }) {
    return (_jsx("div", { className: "w-80 shrink-0 border-l border-border-subtle bg-bg-secondary flex flex-col overflow-hidden", children: panels.map((panel) => (_jsx("div", { className: cn("border-b border-border-subtle", panel.collapsible && "transition-all duration-200"), children: panel.component }, panel.id))) }));
}
export { StudioShellContext };
//# sourceMappingURL=StudioShell.js.map