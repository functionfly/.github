/**
 * @functionfly/ui-core
 * Badge component
 */
import * as React from "react";
export type BadgeVariant = "default" | "brand" | "success" | "error" | "warning" | "info" | "ghost" | "outline";
export interface BadgeProps {
    variant?: BadgeVariant;
    size?: "sm" | "md" | "lg";
    dot?: boolean;
    pulse?: boolean;
    className?: string;
    children?: React.ReactNode;
    [key: string]: unknown;
}
export declare function Badge({ className, variant, size, dot, pulse, children, ...props }: BadgeProps): React.JSX.Element;
//# sourceMappingURL=Badge.d.ts.map