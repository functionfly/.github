/**
 * @functionfly/ui-core
 * Shared utility functions for Studio components
 */
import { clsx } from "clsx";
import { twMerge } from "tailwind-merge";
/** Merge class names with tailwind-merge */
export function cn(...inputs) {
    return twMerge(clsx(inputs));
}
/** Generate a unique ID */
export function generateId(prefix = "studio") {
    return `${prefix}-${Math.random().toString(36).slice(2, 9)}-${Date.now().toString(36)}`;
}
/** Clamp a number between min and max */
export function clamp(value, min, max) {
    return Math.min(Math.max(value, min), max);
}
/** Linear interpolation */
export function lerp(a, b, t) {
    return a + (b - a) * t;
}
/** Map a value from one range to another */
export function mapRange(value, inMin, inMax, outMin, outMax) {
    return ((value - inMin) * (outMax - outMin)) / (inMax - inMin) + outMin;
}
/** Convert hex color to CSS variables-friendly format */
export function hexToRgb(hex) {
    const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
    return result
        ? {
            r: parseInt(result[1], 16),
            g: parseInt(result[2], 16),
            b: parseInt(result[3], 16),
        }
        : null;
}
/** Calculate contrast ratio between two colors */
export function contrastRatio(hex1, hex2) {
    const rgb1 = hexToRgb(hex1);
    const rgb2 = hexToRgb(hex2);
    if (!rgb1 || !rgb2)
        return 1;
    const luminance = (r, g, b) => {
        const [rs, gs, bs] = [r, g, b].map((c) => {
            c /= 255;
            return c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4);
        });
        return 0.2126 * rs + 0.7152 * gs + 0.0722 * bs;
    };
    const l1 = luminance(rgb1.r, rgb1.g, rgb1.b);
    const l2 = luminance(rgb2.r, rgb2.g, rgb2.b);
    const lighter = Math.max(l1, l2);
    const darker = Math.min(l1, l2);
    return (lighter + 0.05) / (darker + 0.05);
}
/** Debounce a function */
export function debounce(func, wait) {
    let timeout = null;
    return (...args) => {
        if (timeout)
            clearTimeout(timeout);
        timeout = setTimeout(() => {
            func(...args);
            timeout = null;
        }, wait);
    };
}
/** Throttle a function */
export function throttle(func, limit) {
    let inThrottle = false;
    return (...args) => {
        if (!inThrottle) {
            func(...args);
            inThrottle = true;
            setTimeout(() => {
                inThrottle = false;
            }, limit);
        }
    };
}
/** Common animation easing curves */
export const EASING = {
    linear: (t) => t,
    easeIn: (t) => t * t,
    easeOut: (t) => t * (2 - t),
    easeInOut: (t) => (t < 0.5 ? 2 * t * t : -1 + (4 - 2 * t) * t),
    cubic: (t) => t * t * t,
    elastic: (t) => t === 0 || t === 1 ? t : Math.pow(2, -10 * t) * Math.sin((t * 10 - 0.75) * ((2 * Math.PI) / 3)) + 1,
};
/** Format numbers for display */
export function formatNumber(num, options) {
    if (options?.compact) {
        if (num >= 1000000)
            return (num / 1000000).toFixed(1) + "M";
        if (num >= 1000)
            return (num / 1000).toFixed(1) + "K";
    }
    return num.toFixed(options?.decimals ?? 0);
}
/** Generate CSS for glassmorphism effect */
export function glassEffect(bg = "rgba(255,255,255,0.05)", blur = 12, borderAlpha = 0.1) {
    return {
        backgroundColor: bg,
        backdropFilter: `blur(${blur}px)`,
        WebkitBackdropFilter: `blur(${blur}px)`,
        border: `1px solid rgba(255, 255, 255, ${borderAlpha})`,
    };
}
/** Calculate node position on a grid */
export function gridPosition(index, columns, cellWidth, cellHeight, gap = 16) {
    const row = Math.floor(index / columns);
    const col = index % columns;
    return {
        x: col * (cellWidth + gap),
        y: row * (cellHeight + gap),
    };
}
/** Safe JSON stringify with error handling */
export function safeStringify(obj, fallback = "{}") {
    try {
        return JSON.stringify(obj);
    }
    catch {
        return fallback;
    }
}
/** Deep merge two objects */
export function deepMerge(target, source) {
    const result = { ...target };
    for (const key of Object.keys(source)) {
        if (source[key] &&
            typeof source[key] === "object" &&
            !Array.isArray(source[key]) &&
            typeof result[key] === "object" &&
            !Array.isArray(result[key])) {
            result[key] = deepMerge(result[key], source[key]);
        }
        else {
            result[key] = source[key];
        }
    }
    return result;
}
//# sourceMappingURL=utils.js.map