/**
 * @functionfly/ui-core
 * Tab component for Workspace sections
 */
import * as React from "react";
export interface TabsProps {
    value?: string;
    defaultValue?: string;
    onValueChange?: (value: string) => void;
    children: React.ReactNode;
    className?: string;
}
export interface TabsListProps {
    children: React.ReactNode;
    className?: string;
}
export interface TabsTriggerProps {
    value: string;
    children: React.ReactNode;
    disabled?: boolean;
    className?: string;
    icon?: React.ReactNode;
}
export interface TabsContentProps {
    value: string;
    children: React.ReactNode;
    className?: string;
}
export declare function Tabs({ value, defaultValue, onValueChange, children, className }: TabsProps): React.JSX.Element;
export declare function TabsList({ children, className }: TabsListProps): React.JSX.Element;
export declare function TabsTrigger({ value, children, disabled, className, icon }: TabsTriggerProps): React.JSX.Element;
export declare function TabsContent({ value, children, className }: TabsContentProps): React.JSX.Element;
//# sourceMappingURL=Tabs.d.ts.map