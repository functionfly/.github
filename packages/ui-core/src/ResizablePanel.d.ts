/**
 * @functionfly/ui-core
 * Resizable split panel layout
 */
import * as React from "react";
interface ResizablePanelGroupProps {
    children: React.ReactNode;
    direction?: "horizontal" | "vertical";
    className?: string;
}
interface ResizablePanelProps {
    children: React.ReactNode;
    defaultSize?: number;
    minSize?: number;
    maxSize?: number;
    className?: string;
    _index?: number;
}
interface ResizableHandleProps {
    className?: string;
    withHandle?: boolean;
    _index?: number;
}
export declare function ResizablePanelGroup({ children, direction, className, }: ResizablePanelGroupProps): React.JSX.Element;
export declare function ResizablePanel({ children, defaultSize, minSize, maxSize, className, _index, }: ResizablePanelProps): React.JSX.Element;
export declare function ResizableHandle({ withHandle, className, _index, }: ResizableHandleProps): React.JSX.Element;
export {};
//# sourceMappingURL=ResizablePanel.d.ts.map