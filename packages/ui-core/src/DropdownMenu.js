import { jsx as _jsx } from "react/jsx-runtime";
import * as React from "react";
import { cn } from "./utils";
const DropdownMenuContext = React.createContext(null);
export function DropdownMenu({ children }) {
    const [open, setOpen] = React.useState(false);
    return (_jsx(DropdownMenuContext.Provider, { value: { open, setOpen }, children: _jsx("div", { className: "relative", children: children }) }));
}
export function DropdownMenuTrigger({ children, asChild, className }) {
    const ctx = React.useContext(DropdownMenuContext);
    if (asChild && React.isValidElement(children)) {
        return React.cloneElement(children, {
            onClick: () => ctx?.setOpen(!ctx.open),
            className,
        });
    }
    return (_jsx("button", { onClick: () => ctx?.setOpen(!ctx.open), className: className, children: children }));
}
export function DropdownMenuContent({ children, align = "end", className }) {
    const ctx = React.useContext(DropdownMenuContext);
    const ref = React.useRef(null);
    React.useEffect(() => {
        const handleClickOutside = (e) => {
            if (ref.current && !ref.current.contains(e.target)) {
                ctx?.setOpen(false);
            }
        };
        if (ctx?.open) {
            document.addEventListener("mousedown", handleClickOutside);
        }
        return () => document.removeEventListener("mousedown", handleClickOutside);
    }, [ctx?.open]);
    if (!ctx?.open)
        return null;
    return (_jsx("div", { ref: ref, className: cn("absolute z-50 mt-1 min-w-[8rem] overflow-hidden rounded-md border border-border-subtle bg-bg-primary p-1 shadow-lg", align === "end" ? "right-0" : "left-0", className), children: children }));
}
export function DropdownMenuItem({ children, className, inset, ...props }) {
    return (_jsx("button", { className: cn("relative flex w-full cursor-pointer select-none items-center gap-2 rounded-sm px-2 py-1.5 text-sm text-text-primary outline-none hover:bg-bg-hover transition-colors", inset && "pl-8", className), ...props, children: children }));
}
export function DropdownMenuSeparator() {
    return _jsx("div", { className: "my-1 h-px bg-border-subtle" });
}
//# sourceMappingURL=DropdownMenu.js.map