/**
 * @functionfly/ui-core
 * Panel component for Studio layout
 */
import * as React from "react";
import { PanelProps } from "./types";
export declare function Panel({ id, title, children, icon, collapsible, defaultOpen, className, headerClassName, bodyClassName, }: PanelProps): React.JSX.Element;
export declare function usePanel(): {
    id: string;
    isOpen: boolean;
    toggle: () => void;
};
//# sourceMappingURL=Panel.d.ts.map