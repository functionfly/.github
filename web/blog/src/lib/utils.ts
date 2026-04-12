import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

/**
 * Utility to merge Tailwind classes with clsx
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

/**
 * Blog utility functions
 */

/**
 * Format a date string for display
 */
export function formatDate(dateString: string | null | undefined): string {
  if (!dateString) return '';
  try {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  } catch {
    return '';
  }
}

/**
 * Format a date string for datetime attribute
 */
export function formatDateISO(dateString: string | null | undefined): string {
  if (!dateString) return '';
  try {
    return new Date(dateString).toISOString();
  } catch {
    return '';
  }
}

/**
 * Create a URL-friendly slug from a string
 */
export function slugify(text: string): string {
  return text
    .toLowerCase()
    .replace(/[^\w\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
    .trim();
}

/**
 * Truncate text to a specified length with ellipsis
 */
export function truncate(text: string, length: number = 160): string {
  if (text.length <= length) return text;
  return text.slice(0, length).trim() + '...';
}

/**
 * Calculate pagination range
 */
export function getPaginationRange(
  currentPage: number,
  totalPages: number,
  delta: number = 2
): (number | string)[] {
  const range: (number | string)[] = [];
  
  for (let i = Math.max(2, currentPage - delta); i <= Math.min(totalPages - 1, currentPage + delta); i++) {
    range.push(i);
  }
  
  if (currentPage - delta > 2) {
    range.unshift('...');
  }
  if (currentPage + delta < totalPages - 1) {
    range.push('...');
  }
  
  range.unshift(1);
  if (totalPages > 1) {
    range.push(totalPages);
  }
  
  return range;
}

/**
 * Generate reading time text
 */
export function readingTimeText(minutes: number): string {
  if (minutes < 1) return 'Less than 1 min read';
  if (minutes === 1) return '1 min read';
  return `${minutes} min read`;
}

/**
 * Get category color with fallback
 */
export function getCategoryColor(color?: string): string {
  return color || '#6366f1';
}
