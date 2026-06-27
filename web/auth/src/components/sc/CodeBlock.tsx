import React, { useState } from "react";

interface CodeBlockProps {
  code: string;
  language?: string;
  filename?: string;
}

export const CodeBlock: React.FC<CodeBlockProps> = ({
  code,
  language = "text",
  filename,
}) => {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Clipboard API not available
    }
  };

  return (
    <div
      className="code-block"
      style={{
        background: "var(--panel)",
        border: "1px solid var(--panel-edge)",
        borderRadius: "var(--radius)",
        overflow: "hidden",
      }}
    >
      {(filename || language !== "text") && (
        <div
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            padding: "var(--space-2) var(--space-3)",
            borderBottom: "1px solid var(--panel-edge)",
            background: "var(--panel-raised)",
          }}
        >
          <span
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: "11px",
              fontWeight: 500,
              textTransform: "uppercase",
              letterSpacing: "0.06em",
              color: "var(--text-faint)",
            }}
          >
            {filename || language}
          </span>
          <button
            onClick={handleCopy}
            aria-label={copied ? "Copied" : "Copy code"}
            className="code-block-copy"
            style={{
              background: "none",
              border: "1px solid var(--steel)",
              borderRadius: "var(--radius-sm)",
              padding: "var(--space-1) var(--space-2)",
              fontFamily: "var(--font-mono)",
              fontSize: "11px",
              color: copied ? "var(--status-ok)" : "var(--text-dim)",
              cursor: "pointer",
              transition:
                "color var(--duration-fast) var(--ease-out), border-color var(--duration-fast) var(--ease-out)",
            }}
          >
            {copied ? "Copied!" : "Copy"}
          </button>
        </div>
      )}
      <pre
        style={{
          margin: 0,
          padding: "var(--space-4)",
          overflow: "auto",
          fontFamily: "var(--font-mono)",
          fontSize: "13px",
          lineHeight: 1.7,
          color: "var(--text-dim)",
        }}
      >
        <code>{code}</code>
      </pre>
    </div>
  );
};
