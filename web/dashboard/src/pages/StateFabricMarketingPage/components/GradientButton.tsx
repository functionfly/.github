import { motion } from "framer-motion";
import { ReactNode } from "react";

interface GradientButtonProps {
  children: ReactNode;
  href?: string;
  onClick?: () => void;
  variant?: "primary" | "secondary";
  size?: "sm" | "md" | "lg";
  className?: string;
}

const variants = {
  primary: "bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white shadow-lg hover:shadow-xl",
  secondary: "bg-bg-secondary hover:bg-bg-primary border border-border hover:border-border/80 text-slate-900 dark:text-white"
};

const sizes = {
  sm: "px-4 py-2 text-sm",
  md: "px-6 py-3 text-base",
  lg: "px-8 py-4 text-lg"
};

export function GradientButton({
  children,
  href,
  onClick,
  variant = "primary",
  size = "md",
  className = ""
}: GradientButtonProps) {
  const buttonClasses = `
    inline-flex items-center justify-center rounded-lg font-semibold
    transition-all duration-300 transform hover:scale-105 active:scale-95
    focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2
    ${variants[variant]} ${sizes[size]} ${className}
  `.trim();

  if (href) {
    return (
      <motion.a
        href={href}
        className={buttonClasses}
        whileHover={{ y: -2 }}
        whileTap={{ y: 0 }}
        transition={{ type: "spring", stiffness: 400, damping: 17 }}
      >
        {children}
      </motion.a>
    );
  }

  return (
    <motion.button
      onClick={onClick}
      className={buttonClasses}
      whileHover={{ y: -2 }}
      whileTap={{ y: 0 }}
      transition={{ type: "spring", stiffness: 400, damping: 17 }}
    >
      {children}
    </motion.button>
  );
}