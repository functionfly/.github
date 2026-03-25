/**
 * Export Utilities
 * Functions for exporting data to CSV and JSON formats
 */

/**
 * Convert an array of objects to CSV format
 */
export function toCSV<T extends Record<string, any>>(
  data: T[],
  columns?: { key: keyof T; label: string }[]
): string {
  if (data.length === 0) return '';

  // If columns are not specified, use all keys from the first object
  const cols = columns || Object.keys(data[0]).map((key) => ({ key, label: key }));

  // Header row
  const header = cols.map((col) => `"${col.label}"`).join(',');

  // Data rows
  const rows = data.map((row) => {
    return cols
      .map((col) => {
        const value = row[col.key];
        // Handle null/undefined
        if (value === null || value === undefined) return '';
        // Escape quotes and wrap in quotes
        const stringValue = String(value).replace(/"/g, '""');
        return `"${stringValue}"`;
      })
      .join(',');
  });

  return [header, ...rows].join('\n');
}

/**
 * Download data as a file
 */
export function downloadFile(content: string, filename: string, mimeType: string): void {
  const blob = new Blob([content], { type: mimeType });
  const url = URL.createObjectURL(blob);

  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);

  URL.revokeObjectURL(url);
}

/**
 * Export data to CSV file
 */
export function exportToCSV<T extends Record<string, any>>(
  data: T[],
  filename: string,
  columns?: { key: keyof T; label: string }[]
): void {
  const csv = toCSV(data, columns);
  downloadFile(csv, `${filename}.csv`, 'text/csv;charset=utf-8;');
}

/**
 * Export data to JSON file
 */
export function exportToJSON<T>(data: T, filename: string): void {
  const json = JSON.stringify(data, null, 2);
  downloadFile(json, `${filename}.json`, 'application/json');
}

/**
 * Format data for export with custom transformations
 */
export function formatDataForExport<T extends Record<string, any>>(
  data: T[],
  formatters: Partial<Record<keyof T, (value: any) => string>>
): T[] {
  return data.map((row) => {
    const formatted = { ...row };
    Object.entries(formatters).forEach(([key, formatter]) => {
      if (key in formatted && formatter) {
        (formatted as any)[key] = formatter(formatted[key as keyof T]);
      }
    });
    return formatted;
  });
}

/**
 * Common column definitions for export
 */
export const COMMON_EXPORT_COLUMNS = {
  // Date formatting
  date: (value: any) => {
    if (!value) return '';
    const date = new Date(value);
    return isNaN(date.getTime()) ? String(value) : date.toISOString().split('T')[0];
  },

  // DateTime formatting
  datetime: (value: any) => {
    if (!value) return '';
    const date = new Date(value);
    return isNaN(date.getTime()) ? String(value) : date.toLocaleString();
  },

  // Boolean formatting
  boolean: (value: any) => {
    if (value === true || value === 'true') return 'Yes';
    if (value === false || value === 'false') return 'No';
    return String(value);
  },

  // Status formatting
  status: (value: any) => {
    if (!value) return '';
    return String(value).charAt(0).toUpperCase() + String(value).slice(1);
  },

  // Currency formatting
  currency: (value: any) => {
    if (value === null || value === undefined) return '';
    const num = Number(value);
    if (isNaN(num)) return String(value);
    return `$${(num / 100).toFixed(2)}`;
  },
};

export default {
  toCSV,
  downloadFile,
  exportToCSV,
  exportToJSON,
  formatDataForExport,
  COMMON_EXPORT_COLUMNS,
};
