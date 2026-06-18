/**
 * @functionfly/ui-core
 * Tooltip component
 */
import * as React from "react";
export interface TooltipProps {
    content: React.ReactNode;
    children: React.ReactNode;
    side?: "top" | "right" | "bottom" | "left";
    delayMs?: number;
}
export declare function Tooltip({ content, children, side, delayMs }: TooltipProps): React.JSX.Element;
//# sourceMappingURL=Tooltip.d.ts.map