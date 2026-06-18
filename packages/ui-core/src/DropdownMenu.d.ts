import * as React from "react";
export interface DropdownMenuProps {
    children: React.ReactNode;
}
export declare function DropdownMenu({ children }: DropdownMenuProps): React.JSX.Element;
export interface DropdownMenuTriggerProps {
    children: React.ReactNode;
    asChild?: boolean;
    className?: string;
}
export declare function DropdownMenuTrigger({ children, asChild, className }: DropdownMenuTriggerProps): React.JSX.Element;
export interface DropdownMenuContentProps {
    children: React.ReactNode;
    align?: "start" | "end";
    className?: string;
}
export declare function DropdownMenuContent({ children, align, className }: DropdownMenuContentProps): React.JSX.Element;
export interface DropdownMenuItemProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    children: React.ReactNode;
    inset?: boolean;
}
export declare function DropdownMenuItem({ children, className, inset, ...props }: DropdownMenuItemProps): React.JSX.Element;
export declare function DropdownMenuSeparator(): React.JSX.Element;
//# sourceMappingURL=DropdownMenu.d.ts.map