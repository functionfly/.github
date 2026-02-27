import { motion } from "framer-motion";

interface TagPillProps {
  text: string;
  variant?: "default" | "primary" | "secondary";
  className?: string;
}

const variants = {
  default: "tag-pill-default bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 border-slate-200 dark:border-slate-700",
  primary: "tag-pill-primary bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300 border-blue-200 dark:border-blue-800",
  secondary: "tag-pill-secondary bg-purple-50 dark:bg-purple-900/20 text-purple-700 dark:text-purple-300 border-purple-200 dark:border-purple-800",
};

export function TagPill({ text, variant = "default", className = "" }: TagPillProps) {
  return (
    <motion.span
      className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium border ${variants[variant]} ${className}`}
      whileHover={{ scale: 1.05 }}
      transition={{ type: "spring", stiffness: 400, damping: 17 }}
    >
      {text}
    </motion.span>
  );
}