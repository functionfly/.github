/**
 * Syntax highlighting utilities
 */

import React from "react";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { oneDark, oneLight } from "react-syntax-highlighter/dist/esm/styles/prism";
import type { ThemeMode } from "../types";
import { SUPPORTED_LANGUAGES } from "../constants";

/**
 * Get the appropriate syntax highlighting theme
 */
export function getSyntaxTheme(theme: ThemeMode) {
  return theme === "dark" ? oneDark : oneLight;
}

/**
 * Check if a language is supported for syntax highlighting
 */
export function isLanguageSupported(language?: string): boolean {
  if (!language) return false;
  return SUPPORTED_LANGUAGES.has(language.toLowerCase());
}

/**
 * Create syntax highlighted content
 */
export function createSyntaxHighlightedContent(
  content: string,
  language: string,
  theme: ThemeMode,
  className?: string
): React.ReactElement {
  if (!isLanguageSupported(language)) {
    return React.createElement("span", { className }, content);
  }

  return React.createElement(SyntaxHighlighter, {
    language,
    style: getSyntaxTheme(theme),
    customStyle: {
      margin: 0,
      padding: 0,
      background: "transparent",
      fontSize: "inherit",
      lineHeight: "inherit",
    },
    codeTagProps: {
      style: {
        fontSize: "inherit",
        lineHeight: "inherit",
        fontFamily: "inherit",
      },
    },
    className,
    children: content,
  });
}