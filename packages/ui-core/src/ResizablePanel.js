import { jsx as _jsx } from "react/jsx-runtime";
/**
 * @functionfly/ui-core
 * Resizable split panel layout
 */
import * as React from "react";
import { cn } from "./utils";
const ResizablePanelContext = React.createContext({
    direction: "horizontal",
    sizes: [],
    setSizes: () => { },
    panelCount: 0,
});
export function ResizablePanelGroup({ children, direction = "horizontal", className, }) {
    let panelCounter = 0;
    let handleCounter = 0;
    const childrenWithIndex = React.Children.map(children, (child) => {
        if (!React.isValidElement(child))
            return child;
        if (child.type === ResizablePanel) {
            const idx = panelCounter++;
            return React.cloneElement(child, { _index: idx });
        }
        if (child.type === ResizableHandle) {
            const idx = handleCounter++;
            return React.cloneElement(child, { _index: idx });
        }
        return child;
    });
    const panels = [];
    React.Children.forEach(childrenWithIndex, (child) => {
        if (React.isValidElement(child) && child.type === ResizablePanel) {
            panels.push(child);
        }
    });
    const panelCount = panels.length;
    const initialSizes = panels.map((child) => child.props.defaultSize ?? 100 / panelCount);
    const [sizes, setSizes] = React.useState(initialSizes);
    const contextValue = React.useMemo(() => ({ direction, sizes, setSizes, panelCount }), [direction, sizes, setSizes, panelCount]);
    return (_jsx(ResizablePanelContext.Provider, { value: contextValue, children: _jsx("div", { "data-panel-group": true, className: cn("flex overflow-hidden h-full w-full", direction === "vertical" ? "flex-col" : "flex-row", className), children: childrenWithIndex }) }));
}
export function ResizablePanel({ children, defaultSize, minSize = 0, maxSize = 100, className, _index, }) {
    const { direction, sizes, panelCount } = React.useContext(ResizablePanelContext);
    const panelIdx = _index ?? -1;
    const size = panelIdx >= 0
        ? (sizes[panelIdx] ?? defaultSize ?? 100 / panelCount)
        : (defaultSize ?? 100 / panelCount);
    return (_jsx("div", { className: cn("relative flex flex-col overflow-hidden h-full w-full transition-[flex] duration-150", className), style: direction === "horizontal"
            ? { flex: `0 0 ${size}%`, width: `${size}%` }
            : { flex: `0 0 ${size}%`, height: `${size}%` }, children: children }));
}
export function ResizableHandle({ withHandle = true, className, _index, }) {
    const { direction, setSizes } = React.useContext(ResizablePanelContext);
    const [dragging, setDragging] = React.useState(false);
    const dragIndex = _index ?? -1;
    const handleRef = React.useRef(null);
    React.useEffect(() => {
        if (!dragging)
            return;
        const handleMouseMove = (e) => {
            const container = handleRef.current?.closest("[data-panel-group]");
            if (!container)
                return;
            const rect = container.getBoundingClientRect();
            const pos = direction === "horizontal" ? e.clientX - rect.left : e.clientY - rect.top;
            const totalSize = direction === "horizontal" ? rect.width : rect.height;
            const percentage = (pos / totalSize) * 100;
            setSizes((prev) => {
                const newSizes = [...prev];
                if (dragIndex >= 0 && dragIndex < newSizes.length - 1) {
                    const total = newSizes[dragIndex] + newSizes[dragIndex + 1];
                    newSizes[dragIndex] = Math.max(10, Math.min(total - 10, percentage));
                    newSizes[dragIndex + 1] = total - newSizes[dragIndex];
                }
                return newSizes;
            });
        };
        const handleMouseUp = () => {
            setDragging(false);
        };
        document.addEventListener("mousemove", handleMouseMove);
        document.addEventListener("mouseup", handleMouseUp);
        return () => {
            document.removeEventListener("mousemove", handleMouseMove);
            document.removeEventListener("mouseup", handleMouseUp);
        };
    }, [dragging, direction, setSizes, dragIndex]);
    const handleMouseDown = (e) => {
        e.preventDefault();
        setDragging(true);
    };
    return (_jsx("div", { ref: handleRef, className: cn("relative flex-shrink-0 transition-colors duration-150", direction === "horizontal" ? "w-1 cursor-col-resize" : "h-1 cursor-row-resize", dragging
            ? "bg-brand-500/40"
            : "bg-transparent hover:bg-brand-500/20", className), onMouseDown: handleMouseDown, role: "separator", "aria-orientation": direction, tabIndex: -1, children: withHandle && (_jsx("div", { className: cn("absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2", "flex gap-0.5", direction === "horizontal" ? "flex-col" : "flex-row"), children: [0, 1, 2].map((i) => (_jsx("span", { className: cn("block rounded-sm", direction === "horizontal" ? "h-1 w-2" : "h-2 w-1", "bg-border-default/50") }, i))) })) }));
}
//# sourceMappingURL=ResizablePanel.js.map