import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

/**
 * Combines clsx and tailwind-merge to conditionally join class names
 * while resolving Tailwind CSS conflicts.
 * 
 * Usage:
 *   cn("btn", isActive && "btn-active", isLarge && "btn-lg")
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
