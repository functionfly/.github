import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { cn } from "./utils";
export function Spinner({ className, size = "md", variant = "default", children, ...props }) {
    const sizeClasses = {
        sm: "size-4",
        md: "size-6",
        lg: "size-8",
        xl: "size-10",
    };
    const variantClasses = {
        default: "text-brand-500",
        brand: "text-brand-500",
        monochrome: "text-text-primary",
    };
    return (_jsxs("div", { className: cn("flex flex-col items-center gap-2", className), ...props, children: [_jsx("div", { className: cn("rounded-full border-2 border-border-subtle border-t-brand-500 animate-spin", sizeClasses[size], variantClasses[variant]), style: { borderTopColor: "currentColor" }, role: "progressbar", "aria-label": "Loading" }), children && _jsx("span", { className: "text-sm text-text-secondary", children: children })] }));
}
//# sourceMappingURL=Spinner.js.map