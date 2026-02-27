import { ReactNode } from "react";

interface IconWithTitleProps {
  icon: ReactNode;
  title: string;
  className?: string;
}

export function IconWithTitle({ icon, title, className = "" }: IconWithTitleProps) {
  return (
    <div className={`flex flex-col items-center text-center ${className}`}>
      <div className="w-12 h-12 rounded-lg glass-light flex items-center justify-center mb-4 animate-float">
        {icon}
      </div>
      <h3 className="text-xl font-semibold text-slate-900 dark:text-white">{title}</h3>
    </div>
  );
}