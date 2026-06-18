import { jsx as _jsx } from "react/jsx-runtime";
import * as React from "react";
import { cn } from "./utils";
export function Select({ value, onValueChange, children, className }) {
    const [open, setOpen] = React.useState(false);
    return (_jsx("div", { className: cn("relative", className), children: React.Children.map(children, (child) => {
            if (React.isValidElement(child)) {
                return React.cloneElement(child, {
                    open,
                    onOpenChange: setOpen,
                    value,
                    onValueChange,
                });
            }
            return child;
        }) }));
}
export function SelectTrigger({ children, className, asChild }) {
    return (_jsx("button", { type: "button", className: cn("flex h-9 w-full items-center justify-between whitespace-nowrap rounded-md border border-border-subtle bg-bg-primary px-3 py-2 text-sm text-text-primary hover:bg-bg-hover focus:outline-none focus:ring-2 focus:ring-brand-500 disabled:cursor-not-allowed disabled:opacity-50", className), children: children }));
}
export function SelectValue({ placeholder }) {
    return _jsx("span", { className: "text-text-muted", children: placeholder });
}
export function SelectContent({ children, className, open }) {
    const ref = React.useRef(null);
    React.useEffect(() => {
        const handleClickOutside = (e) => {
            if (ref.current && !ref.current.contains(e.target)) {
                // Would trigger close
            }
        };
        if (open) {
            document.addEventListener("mousedown", handleClickOutside);
        }
        return () => document.removeEventListener("mousedown", handleClickOutside);
    }, [open]);
    if (!open)
        return null;
    return (_jsx("div", { ref: ref, className: cn("absolute z-50 mt-1 min-w-[8rem] overflow-hidden rounded-md border border-border-subtle bg-bg-primary p-1 shadow-lg", className), children: children }));
}
export function SelectItem({ value, children, className }) {
    return (_jsx("button", { type: "button", className: cn("relative flex w-full cursor-pointer select-none items-center rounded-sm py-1.5 px-2 text-sm text-text-primary hover:bg-bg-hover", className), children: children }));
}
//# sourceMappingURL=Select.js.map