/**
 * Formatting utilities for the diff viewer
 */

import React from "react";
import type { DiffLine } from "../types";

/**
 * Format file size for display
 */
export function formatFileSize(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i];
}

/**
 * Highlight search matches in text
 */
export function highlightSearch(text: string, searchTerm: string): { start: number; end: number }[] {
  if (!searchTerm.trim()) return [];

  const highlights: { start: number; end: number }[] = [];
  const lowerText = text.toLowerCase();
  const lowerSearch = searchTerm.toLowerCase();
  let startIndex = 0;

  while (true) {
    const index = lowerText.indexOf(lowerSearch, startIndex);
    if (index === -1) break;

    highlights.push({ start: index, end: index + searchTerm.length });
    startIndex = index + 1;
  }

  return highlights;
}

/**
 * Render content with search highlights
 */
export function renderContentWithHighlights(
  content: string,
  highlights?: { start: number; end: number }[]
): React.ReactNode {
  if (!highlights || highlights.length === 0) {
    return content;
  }

  const parts: React.ReactNode[] = [];
  let lastIndex = 0;

  highlights.forEach((highlight, index) => {
    // Add text before highlight
    if (highlight.start > lastIndex) {
      parts.push(content.slice(lastIndex, highlight.start));
    }
    // Add highlighted text
    parts.push(
      React.createElement("mark", {
        key: `highlight-${index}`,
        className: "bg-yellow-400/30 text-yellow-900 rounded px-0.5",
        children: content.slice(highlight.start, highlight.end)
      })
    );
    lastIndex = highlight.end;
  });

  // Add remaining text
  if (lastIndex < content.length) {
    parts.push(content.slice(lastIndex));
  }

  return parts;
}

/**
 * Render word-level diffs for modified lines
 */
export function renderWordDiffs(wordDiffs: { text: string; type: string }[]): React.ReactNode[] {
  return wordDiffs.map((word, index) => {
    const styles = {
      added: "bg-emerald-500/10 text-emerald-400",
      removed: "bg-red-500/10 text-red-400",
      modified: "bg-amber-500/10 text-amber-400",
      unchanged: "",
    };

    return (
      <span
        key={index}
        className={`rounded px-0.5 ${word.type !== "unchanged" ? styles[word.type as keyof typeof styles] : ""}`}
      >
        {word.text}
      </span>
    );
  });
}
