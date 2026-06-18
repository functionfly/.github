/**
 * @functionfly/ui-core
 * Core layout and adaptive complexity components for FunctionFly Studio
 */
export interface StudioShellProps {
    children: React.ReactNode;
    className?: string;
}
export interface AdaptiveWorkspaceProps {
    children: React.ReactNode;
    className?: string;
}
export type ComplexityLevel = 'beginner' | 'intermediate' | 'expert';
export interface AdaptiveComplexityLayerProps {
    children: React.ReactNode;
    level?: ComplexityLevel;
}
export interface ResizablePanelProps {
    children: React.ReactNode;
    defaultSize?: number;
    minSize?: number;
    maxSize?: number;
    direction?: 'horizontal' | 'vertical';
    className?: string;
    onResize?: (size: number) => void;
}
export interface DockLayoutProps {
    children: React.ReactNode;
    className?: string;
    defaultLayout?: 'grid' | 'tabs' | 'sidebar';
}
export interface WorkspaceViewportProps {
    children: React.ReactNode;
    className?: string;
    zoom?: number;
    onZoomChange?: (zoom: number) => void;
}
export interface PanelProps {
    id: string;
    title: string;
    children: React.ReactNode;
    icon?: React.ReactNode;
    collapsible?: boolean;
    defaultOpen?: boolean;
    className?: string;
    headerClassName?: string;
    bodyClassName?: string;
}
export interface FloatingPanelProps {
    children: React.ReactNode;
    title?: string;
    position?: 'left' | 'right' | 'top' | 'bottom' | {
        x: number;
        y: number;
    };
    size?: 'sm' | 'md' | 'lg' | 'xl' | {
        width: number;
        height: number;
    };
    zIndex?: number;
    onClose?: () => void;
    resizable?: boolean;
    draggable?: boolean;
}
export interface UsePanelOptions {
    defaultOpen?: boolean;
    onOpen?: () => void;
    onClose?: () => void;
    onToggle?: (isOpen: boolean) => void;
}
export interface PanelState {
    isOpen: boolean;
    toggle: () => void;
    open: () => void;
    close: () => void;
}
export interface StudioContextType {
    complexity: ComplexityLevel;
    setComplexity: (level: ComplexityLevel) => void;
    panels: Record<string, boolean>;
    togglePanel: (id: string) => void;
    zoom: number;
    setZoom: (zoom: number) => void;
    theme: 'dark' | 'light';
    setTheme: (theme: 'dark' | 'light') => void;
}
//# sourceMappingURL=types.d.ts.map