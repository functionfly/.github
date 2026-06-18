import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
/**
 * @functionfly/ui-core
 * Tooltip component
 */
import * as React from "react";
import { cn } from "./utils";
export function Tooltip({ content, children, side = "top", delayMs = 200 }) {
    const [visible, setVisible] = React.useState(false);
    const timerRef = React.useRef(null);
    const wrapperRef = React.useRef(null);
    const show = () => {
        timerRef.current = setTimeout(() => setVisible(true), delayMs);
    };
    const hide = () => {
        if (timerRef.current)
            clearTimeout(timerRef.current);
        setVisible(false);
    };
    React.useEffect(() => {
        return () => {
            if (timerRef.current)
                clearTimeout(timerRef.current);
        };
    }, []);
    const sideClasses = {
        top: "top-full left-1/2 -translate-x-1/2 mb-2",
        bottom: "bottom-full left-1/2 -translate-x-1/2 mt-2",
        left: "top-1/2 left-full -translate-y-1/2 mr-2",
        right: "top-1/2 right-full -translate-y-1/2 ml-2",
    };
    return (_jsxs("div", { ref: wrapperRef, className: "relative inline-flex", onMouseEnter: show, onMouseLeave: hide, onFocus: show, onBlur: hide, children: [children, visible && (_jsx("div", { className: cn("absolute whitespace-nowrap rounded-md bg-bg-tertiary px-3 py-1.5 text-sm text-text-primary shadow-lg border border-border-subtle z-50 animate-in fade-in slide-in-from-bottom-2", sideClasses[side]), role: "tooltip", children: content }))] }));
}
//# sourceMappingURL=Tooltip.js.map