/**
 * @functionfly/ui-core
 * Loading/spinner component
 */
import * as React from "react";
export interface SpinnerProps extends React.HTMLAttributes<HTMLDivElement> {
    size?: "sm" | "md" | "lg" | "xl";
    variant?: "default" | "brand" | "monochrome";
}
export declare function Spinner({ className, size, variant, children, ...props }: SpinnerProps): React.JSX.Element;
//# sourceMappingURL=Spinner.d.ts.map