/**
 * @functionfly/ui-core
 * StudioShell - The main application shell layout
 */
import * as React from "react";
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
declare const StudioShellContext: React.Context<StudioShellContextValue>;
interface StudioShellProps {
    children: React.ReactNode;
    className?: string;
    settingsContent?: React.ReactNode;
}
/**
 * Top-level layout shell for FunctionFly Studio.
 * Provides panel management, adaptive complexity, and workspace layout.
 */
export declare function StudioShell({ children, className, settingsContent: settingsContentProp }: StudioShellProps): React.JSX.Element;
/**
 * Title bar with branding, controls, and quick actions
 */
export declare function TitleBar({ complexity, onZoomChange, }: {
    complexity: string;
    onZoomChange: (zoom: number) => void;
}): React.JSX.Element;
/**
 * Complexity level toggler
 */
export declare function ComplexityToggle({ current, onChange }: {
    current: string;
    onChange: (level: string) => void;
}): React.JSX.Element;
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
export declare function LeftSidebar({ panels }: {
    panels: PanelConfig[];
}): React.JSX.Element;
/**
 * Right panel dock
 */
export declare function RightPanel({ panels }: {
    panels: PanelConfig[];
}): React.JSX.Element;
export { StudioShellContext };
//# sourceMappingURL=StudioShell.d.ts.map