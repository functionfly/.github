import { ReactNode } from "react";
import { motion } from "framer-motion";
import { cn } from "@/lib/utils";

interface PageLayoutProps {
  children: ReactNode;
  className?: string;
  maxWidth?: "sm" | "md" | "lg" | "xl" | "2xl" | "3xl" | "4xl" | "5xl" | "6xl" | "7xl" | "full";
  padding?: "none" | "sm" | "md" | "lg";
  animate?: boolean;
  title?: string;
}

const maxWidthClasses = {
  sm: "max-w-sm",
  md: "max-w-md",
  lg: "max-w-lg",
  xl: "max-w-xl",
  "2xl": "max-w-2xl",
  "3xl": "max-w-3xl",
  "4xl": "max-w-4xl",
  "5xl": "max-w-5xl",
  "6xl": "max-w-6xl",
  "7xl": "max-w-7xl",
  full: "max-w-full",
};

const paddingClasses = {
  none: "",
  sm: "p-4 lg:p-6",
  md: "p-6 lg:p-8",
  lg: "p-8 lg:p-12",
};

export function PageLayout({
  children,
  className,
  maxWidth = "7xl",
  padding = "md",
  animate = true,
  title,
}: PageLayoutProps) {
  const Content = animate ? motion.div : "div";

  const contentProps = animate ? {
    initial: { opacity: 0, y: 20 },
    animate: { opacity: 1, y: 0 },
    transition: { duration: 0.3, ease: "easeOut" as const },
  } : {};

  return (
    <div className={cn("min-h-full", className)}>
      <Content
        {...contentProps}
        className={cn(
          "mx-auto w-full",
          maxWidthClasses[maxWidth],
          paddingClasses[padding]
        )}
      >
        {title && (
          <div className="mb-6">
            <h1 className="text-2xl font-bold text-text-primary">{title}</h1>
          </div>
        )}
        {children}
      </Content>
    </div>
  );
}