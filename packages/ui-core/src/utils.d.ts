/**
 * @functionfly/ui-core
 * Shared utility functions for Studio components
 */
import { type ClassValue } from "clsx";
import { type CSSProperties } from "react";
/** Merge class names with tailwind-merge */
export declare function cn(...inputs: ClassValue[]): string;
/** Generate a unique ID */
export declare function generateId(prefix?: string): string;
/** Clamp a number between min and max */
export declare function clamp(value: number, min: number, max: number): number;
/** Linear interpolation */
export declare function lerp(a: number, b: number, t: number): number;
/** Map a value from one range to another */
export declare function mapRange(value: number, inMin: number, inMax: number, outMin: number, outMax: number): number;
/** Convert hex color to CSS variables-friendly format */
export declare function hexToRgb(hex: string): {
    r: number;
    g: number;
    b: number;
} | null;
/** Calculate contrast ratio between two colors */
export declare function contrastRatio(hex1: string, hex2: string): number;
/** Debounce a function */
export declare function debounce<T extends (...args: unknown[]) => unknown>(func: T, wait: number): (...args: Parameters<T>) => void;
/** Throttle a function */
export declare function throttle<T extends (...args: unknown[]) => unknown>(func: T, limit: number): (...args: Parameters<T>) => void;
/** Common animation easing curves */
export declare const EASING: {
    readonly linear: (t: number) => number;
    readonly easeIn: (t: number) => number;
    readonly easeOut: (t: number) => number;
    readonly easeInOut: (t: number) => number;
    readonly cubic: (t: number) => number;
    readonly elastic: (t: number) => number;
};
/** Format numbers for display */
export declare function formatNumber(num: number, options?: {
    decimals?: number;
    compact?: boolean;
}): string;
/** Generate CSS for glassmorphism effect */
export declare function glassEffect(bg?: string, blur?: number, borderAlpha?: number): CSSProperties;
/** Calculate node position on a grid */
export declare function gridPosition(index: number, columns: number, cellWidth: number, cellHeight: number, gap?: number): {
    x: number;
    y: number;
};
/** Safe JSON stringify with error handling */
export declare function safeStringify(obj: unknown, fallback?: string): string;
/** Deep merge two objects */
export declare function deepMerge<T extends Record<string, any>>(target: T, source: Partial<T>): T;
//# sourceMappingURL=utils.d.ts.map