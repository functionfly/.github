import * as React from "react";
export interface SelectProps {
    value?: string;
    onValueChange?: (value: string) => void;
    children: React.ReactNode;
    className?: string;
}
export declare function Select({ value, onValueChange, children, className }: SelectProps): React.JSX.Element;
export interface SelectTriggerProps {
    children: React.ReactNode;
    className?: string;
    asChild?: boolean;
}
export declare function SelectTrigger({ children, className, asChild }: SelectTriggerProps): React.JSX.Element;
export declare function SelectValue({ placeholder }: {
    placeholder?: string;
}): React.JSX.Element;
export interface SelectContentProps {
    children: React.ReactNode;
    className?: string;
    open?: boolean;
}
export declare function SelectContent({ children, className, open }: SelectContentProps): React.JSX.Element;
export interface SelectItemProps {
    value: string;
    children: React.ReactNode;
    className?: string;
}
export declare function SelectItem({ value, children, className }: SelectItemProps): React.JSX.Element;
//# sourceMappingURL=Select.d.ts.map