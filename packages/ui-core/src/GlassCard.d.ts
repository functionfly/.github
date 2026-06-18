/**
 * @functionfly/ui-core
 * Glass morphism card component
 */
import * as React from "react";
export interface GlassCardProps extends React.HTMLAttributes<HTMLDivElement> {
    glowColor?: string;
    intensity?: "low" | "medium" | "high";
    animated?: boolean;
    glass?: boolean;
}
export declare function GlassCard({ className, glowColor, intensity, animated, glass, children, ...props }: GlassCardProps): React.JSX.Element;
//# sourceMappingURL=GlassCard.d.ts.map