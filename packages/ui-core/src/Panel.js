import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
/**
 * @functionfly/ui-core
 * Panel component for Studio layout
 */
import * as React from "react";
import { cn } from "./utils";
const PanelContext = React.createContext(null);
export function Panel({ id, title, children, icon, collapsible = true, defaultOpen = true, className, headerClassName, bodyClassName, }) {
    const [isOpen, setIsOpen] = React.useState(defaultOpen);
    const contextValue = React.useMemo(() => ({ id, isOpen, toggle: () => setIsOpen((v) => !v) }), [id, isOpen]);
    return (_jsx(PanelContext.Provider, { value: contextValue, children: _jsxs("div", { className: cn("flex flex-col overflow-hidden", className), children: [collapsible && (_jsxs("button", { className: cn("flex items-center gap-2 px-4 py-2 text-sm font-medium transition-colors duration-200", "text-text-secondary hover:text-text-primary hover:bg-bg-hover rounded-t-lg", "border-b border-border-subtle", headerClassName), onClick: () => setIsOpen((v) => !v), "aria-expanded": isOpen, "aria-controls": `panel-body-${id}`, children: [icon && _jsx("span", { className: "shrink-0", children: icon }), _jsx("span", { className: "flex-1 text-left truncate", children: title }), _jsx("svg", { className: cn("size-4 shrink-0 transition-transform duration-200", isOpen ? "rotate-0" : "-rotate-90"), viewBox: "0 0 24 24", fill: "none", stroke: "currentColor", strokeWidth: "2", strokeLinecap: "round", strokeLinejoin: "round", children: _jsx("polyline", { points: "6 9 12 15 18 9" }) })] })), _jsx("div", { id: `panel-body-${id}`, className: cn("flex-1 overflow-auto transition-all duration-200", !isOpen && collapsible ? "h-0 p-0 opacity-0" : "opacity-100", bodyClassName), children: children })] }) }));
}
export function usePanel() {
    const context = React.useContext(PanelContext);
    if (!context) {
        throw new Error("usePanel must be used within a Panel component");
    }
    return context;
}
//# sourceMappingURL=Panel.js.map