import { motion } from "framer-motion";

interface CodeSnippetBlockProps {
  title: string;
  code: string;
  language?: string;
  className?: string;
  delay?: number;
}

export function CodeSnippetBlock({ title, code, language = "typescript", className = "", delay = 0 }: CodeSnippetBlockProps) {
  return (
    <motion.div
      className={`bg-slate-900 p-6 rounded-lg border border-border ${className}`}
      initial={{ opacity: 0, y: 20 }}
      whileInView={{ opacity: 1, y: 0 }}
      transition={{ delay, duration: 0.6 }}
      viewport={{ once: true }}
    >
      <div className="flex items-center justify-between mb-4">
        <h4 className="text-sm font-medium text-white dark:text-white">{title}</h4>
        <span className="text-xs text-slate-400 dark:text-slate-500 bg-slate-800 dark:bg-slate-800 px-2 py-1 rounded">
          {language}
        </span>
      </div>
      <pre className="text-sm text-slate-100 dark:text-slate-100 overflow-x-auto">
        <code>{code}</code>
      </pre>
    </motion.div>
  );
}