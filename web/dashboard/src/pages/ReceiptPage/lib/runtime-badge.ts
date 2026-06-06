// Runtime badge color/icon mapping. Keeps visual style consistent across
// the receipt page and other surfaces (function page, playground).
import {
  Boxes,
  Code2,
  Database,
  Globe,
  type LucideIcon,
  Sparkles,
  Terminal,
  Zap,
} from "lucide-react";

export interface RuntimeStyle {
  icon: LucideIcon;
  label: string;
  className: string; // tailwind class for the badge background
}

const styles: Record<string, RuntimeStyle> = {
  python3:    { icon: Code2,      label: "Python",      className: "bg-blue-500/10 text-blue-400 border-blue-500/20" },
  python3_11: { icon: Code2,      label: "Python 3.11", className: "bg-blue-500/10 text-blue-400 border-blue-500/20" },
  python3_12: { icon: Code2,      label: "Python 3.12", className: "bg-blue-500/10 text-blue-400 border-blue-500/20" },
  python3_13: { icon: Code2,      label: "Python 3.13", className: "bg-blue-500/10 text-blue-400 border-blue-500/20" },
  python_microvm: { icon: Boxes, label: "Python (microVM)", className: "bg-violet-500/10 text-violet-400 border-violet-500/20" },
  node18:     { icon: Terminal,   label: "Node 18",     className: "bg-green-500/10 text-green-400 border-green-500/20" },
  node20:     { icon: Terminal,   label: "Node 20",     className: "bg-green-500/10 text-green-400 border-green-500/20" },
  node22:     { icon: Terminal,   label: "Node 22",     className: "bg-green-500/10 text-green-400 border-green-500/20" },
  deno:       { icon: Globe,      label: "Deno",        className: "bg-cyan-500/10 text-cyan-400 border-cyan-500/20" },
  wasm:       { icon: Sparkles,   label: "WASM",        className: "bg-amber-500/10 text-amber-400 border-amber-500/20" },
  sql:        { icon: Database,   label: "SQL",         className: "bg-pink-500/10 text-pink-400 border-pink-500/20" },
  default:    { icon: Zap,        label: "Function",    className: "bg-zinc-500/10 text-zinc-300 border-zinc-500/20" },
};

/**
 * Get a stable style for a runtime identifier. Falls back to a neutral
 * "Function" badge for unknown runtimes so the UI never breaks.
 */
export function getRuntimeStyle(runtime: string): RuntimeStyle {
  if (!runtime) return styles.default;
  const key = runtime.toLowerCase().replace(/[^a-z0-9_]/g, "_");
  if (styles[key]) return styles[key];
  // Fuzzy prefix match — e.g. "python3.12" → "python3_12"
  for (const k of Object.keys(styles)) {
    if (key.startsWith(k)) return styles[k];
  }
  return {
    ...styles.default,
    label: runtime,
  };
}
