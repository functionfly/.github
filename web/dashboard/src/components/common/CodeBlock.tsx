import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Copy, Check, Terminal } from "lucide-react";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { vscDarkPlus } from "react-syntax-highlighter/dist/esm/styles/prism";
import { motion, AnimatePresence } from "framer-motion";

interface CodeBlockProps {
  code: string;
  language?: string;
  className?: string;
  showLineNumbers?: boolean;
  title?: string;
  maxHeight?: string;
}

const languageLabels: Record<string, string> = {
  javascript: "JavaScript",
  typescript: "TypeScript",
  python: "Python",
  bash: "Bash",
  shell: "Shell",
  json: "JSON",
  yaml: "YAML",
  html: "HTML",
  css: "CSS",
  rust: "Rust",
  go: "Go",
  java: "Java",
  sql: "SQL",
  markdown: "Markdown",
  text: "Plain Text",
};

export function CodeBlock({
  code,
  language = "text",
  className = "",
  showLineNumbers = true,
  title,
  maxHeight = "400px",
}: CodeBlockProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy code:", err);
    }
  };

  const displayLanguage = languageLabels[language] || language.toUpperCase();

  return (
    <div className={`functionfly-code-block rounded-lg overflow-hidden border border-border-subtle ${className}`}>
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-2 bg-bg-tertiary border-b border-border-subtle code-block-header">
        <div className="flex items-center gap-2">
          <Terminal className="w-4 h-4 text-text-muted code-block-icon" />
          <span className="text-xs font-medium text-text-secondary code-block-lang">
            {title || displayLanguage}
          </span>
        </div>
        <AnimatePresence mode="wait">
          <motion.div
            key={copied ? "copied" : "copy"}
            initial={{ opacity: 0, scale: 0.8 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.8 }}
            transition={{ duration: 0.15 }}
          >
            <Button
              variant="ghost"
              size="sm"
              onClick={handleCopy}
              className="h-7 gap-1.5 text-xs text-text-muted hover:text-text-primary code-block-copy-btn"
            >
              {copied ? (
                <>
                  <Check className="h-3.5 w-3.5 text-green-500" />
                  <span className="text-green-500">Copied!</span>
                </>
              ) : (
                <>
                  <Copy className="h-3.5 w-3.5" />
                  Copy
                </>
              )}
            </Button>
          </motion.div>
        </AnimatePresence>
      </div>

      {/* Code Content */}
      <div className="relative" style={{ maxHeight }}>
        <SyntaxHighlighter
          language={language}
          style={vscDarkPlus}
          showLineNumbers={showLineNumbers}
          lineNumberStyle={{
            minWidth: "2.5em",
            paddingRight: "1em",
            color: "var(--code-line-number, #6b7280)",
            fontSize: "0.875rem",
          }}
          customStyle={{
            margin: 0,
            padding: "1rem",
            background: "var(--bg-secondary)",
            fontSize: "0.875rem",
            lineHeight: "1.5",
            borderRadius: 0,
            maxHeight,
            overflow: "auto",
          }}
          wrapLines={true}
          wrapLongLines={false}
        >
          {code.trim()}
        </SyntaxHighlighter>
      </div>
    </div>
  );
}
