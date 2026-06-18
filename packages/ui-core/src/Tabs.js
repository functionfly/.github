import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
/**
 * @functionfly/ui-core
 * Tab component for Workspace sections
 */
import * as React from "react";
import { cn } from "./utils";
const TabsContext = React.createContext(null);
export function Tabs({ value, defaultValue, onValueChange, children, className }) {
    const [internalValue, setInternalValue] = React.useState(defaultValue || value);
    const currentValue = value ?? internalValue;
    const handleValueChange = (newValue) => {
        setInternalValue(newValue);
        onValueChange?.(newValue);
    };
    const contextValue = React.useMemo(() => ({ value: currentValue ?? "", onValueChange: handleValueChange }), [currentValue]);
    return (_jsx(TabsContext.Provider, { value: contextValue, children: _jsx("div", { className: cn("flex flex-col", className), children: children }) }));
}
export function TabsList({ children, className }) {
    return (_jsx("div", { className: cn("flex items-center gap-1 rounded-t-lg border-b border-border-subtle bg-bg-secondary/50 p-1", className), children: children }));
}
export function TabsTrigger({ value, children, disabled, className, icon }) {
    const context = React.useContext(TabsContext);
    if (!context)
        throw new Error("TabsTrigger must be used within Tabs");
    const isActive = context.value === value;
    return (_jsxs("button", { disabled: disabled, onClick: () => context.onValueChange(value), className: cn("inline-flex items-center gap-2 rounded-md px-3 py-1.5 text-sm font-medium transition-all duration-200", "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-border-focus focus-visible:ring-offset-2", isActive
            ? "bg-bg-primary text-text-primary shadow-sm border border-border-default"
            : "text-text-secondary hover:text-text-primary hover:bg-bg-hover", disabled && "opacity-50 cursor-not-allowed", className), children: [icon && _jsx("span", { className: "size-4", children: icon }), children] }));
}
export function TabsContent({ value, children, className }) {
    const context = React.useContext(TabsContext);
    if (!context)
        throw new Error("TabsContent must be used within Tabs");
    if (context.value !== value)
        return null;
    return _jsx("div", { className: cn("mt-2", className), children: children });
}
//# sourceMappingURL=Tabs.js.map