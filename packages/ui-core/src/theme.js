/**
 * @functionfly/ui-core
 * Theme configuration for FunctionFly Studio
 */
export const THEME_CONFIG = {
    colors: {
        // Brand
        brand: {
            50: "#fff7ed",
            100: "#ffedd5",
            200: "#fed7aa",
            300: "#fdba74",
            400: "#fb923c",
            500: "#f97316",
            600: "#ea580c",
            700: "#c2410c",
            800: "#9a3412",
            900: "#7c2d12",
        },
        // Semantic
        success: "#10b981",
        error: "#ef4444",
        warning: "#f59e0b",
        info: "#3b82f6",
        // Status
        online: "#10b981",
        offline: "#ef4444",
        degraded: "#f59e0b",
        pending: "#6b7280",
    },
    // Provider colors for runtime indicators
    providers: {
        "functionfly-edge": "#f48120",
        "workers": "#000000",
        vercel: "#000000",
        fly: "#7b68ee",
        "deno-deploy": "#000000",
    },
};
export const ANIMATION_DURATIONS = {
    fast: "150ms",
    normal: "200ms",
    slow: "300ms",
    slower: "500ms",
};
export const Z_INDEX_LAYERS = {
    dropdown: 1000,
    sticky: 1100,
    modal: 1200,
    popover: 1300,
    toast: 1400,
    tooltip: 1500,
    overlay: 1600,
    canvas: 2000,
    orbit: 3000,
    notification: 4000,
};
export const STUDIO_BREAKPOINTS = {
    xs: "480px",
    sm: "640px",
    md: "768px",
    lg: "1024px",
    xl: "1280px",
    "2xl": "1536px",
};
// SVG filter definitions for glow effects
export const SVG_FILTERS = `
<svg style="position:absolute;width:0;height:0">
  <defs>
    <filter id="glow-brand" x="-50%" y="-50%" width="200%" height="200%">
      <feGaussianBlur in="SourceGraphic" stdDeviation="4" result="blur"/>
      <feFlood flood-color="#f97316" flood-opacity="0.3" result="color"/>
      <feComposite in="color" in2="blur" operator="in" result="glow"/>
      <feMerge>
        <feMergeNode in="glow"/>
        <feMergeNode in="SourceGraphic"/>
      </feMerge>
    </filter>
    <filter id="glow-green" x="-50%" y="-50%" width="200%" height="200%">
      <feGaussianBlur in="SourceGraphic" stdDeviation="3" result="blur"/>
      <feFlood flood-color="#10b981" flood-opacity="0.4" result="color"/>
      <feComposite in="color" in2="blur" operator="in" result="glow"/>
      <feMerge>
        <feMergeNode in="glow"/>
        <feMergeNode in="SourceGraphic"/>
      </feMerge>
    </filter>
  </defs>
</svg>
`;
//# sourceMappingURL=theme.js.map