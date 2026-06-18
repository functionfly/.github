/**
 * @functionfly/ui-core
 * Theme configuration for FunctionFly Studio
 */
export declare const THEME_CONFIG: {
    readonly colors: {
        readonly brand: {
            readonly 50: "#fff7ed";
            readonly 100: "#ffedd5";
            readonly 200: "#fed7aa";
            readonly 300: "#fdba74";
            readonly 400: "#fb923c";
            readonly 500: "#f97316";
            readonly 600: "#ea580c";
            readonly 700: "#c2410c";
            readonly 800: "#9a3412";
            readonly 900: "#7c2d12";
        };
        readonly success: "#10b981";
        readonly error: "#ef4444";
        readonly warning: "#f59e0b";
        readonly info: "#3b82f6";
        readonly online: "#10b981";
        readonly offline: "#ef4444";
        readonly degraded: "#f59e0b";
        readonly pending: "#6b7280";
    };
    readonly providers: {
        readonly "functionfly-edge": "#f48120";
        readonly workers: "#000000";
        readonly vercel: "#000000";
        readonly fly: "#7b68ee";
        readonly "deno-deploy": "#000000";
    };
};
export declare const ANIMATION_DURATIONS: {
    readonly fast: "150ms";
    readonly normal: "200ms";
    readonly slow: "300ms";
    readonly slower: "500ms";
};
export declare const Z_INDEX_LAYERS: {
    readonly dropdown: 1000;
    readonly sticky: 1100;
    readonly modal: 1200;
    readonly popover: 1300;
    readonly toast: 1400;
    readonly tooltip: 1500;
    readonly overlay: 1600;
    readonly canvas: 2000;
    readonly orbit: 3000;
    readonly notification: 4000;
};
export declare const STUDIO_BREAKPOINTS: {
    readonly xs: "480px";
    readonly sm: "640px";
    readonly md: "768px";
    readonly lg: "1024px";
    readonly xl: "1280px";
    readonly "2xl": "1536px";
};
export declare const SVG_FILTERS = "\n<svg style=\"position:absolute;width:0;height:0\">\n  <defs>\n    <filter id=\"glow-brand\" x=\"-50%\" y=\"-50%\" width=\"200%\" height=\"200%\">\n      <feGaussianBlur in=\"SourceGraphic\" stdDeviation=\"4\" result=\"blur\"/>\n      <feFlood flood-color=\"#f97316\" flood-opacity=\"0.3\" result=\"color\"/>\n      <feComposite in=\"color\" in2=\"blur\" operator=\"in\" result=\"glow\"/>\n      <feMerge>\n        <feMergeNode in=\"glow\"/>\n        <feMergeNode in=\"SourceGraphic\"/>\n      </feMerge>\n    </filter>\n    <filter id=\"glow-green\" x=\"-50%\" y=\"-50%\" width=\"200%\" height=\"200%\">\n      <feGaussianBlur in=\"SourceGraphic\" stdDeviation=\"3\" result=\"blur\"/>\n      <feFlood flood-color=\"#10b981\" flood-opacity=\"0.4\" result=\"color\"/>\n      <feComposite in=\"color\" in2=\"blur\" operator=\"in\" result=\"glow\"/>\n      <feMerge>\n        <feMergeNode in=\"glow\"/>\n        <feMergeNode in=\"SourceGraphic\"/>\n      </feMerge>\n    </filter>\n  </defs>\n</svg>\n";
//# sourceMappingURL=theme.d.ts.map