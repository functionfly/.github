import { ReactNode } from "react";

interface IconCardProps {
  icon: ReactNode;
  title: string;
  className?: string;
}

export function IconCard({ icon, title, className = "" }: IconCardProps) {
  return (
    <div className={`flex flex-col items-center text-center p-6 rounded-lg bg-bg-secondary border border-border shadow-sm hover:shadow-md transition-shadow ${className}`}>
      <div className="w-12 h-12 rounded-lg bg-red-100 dark:bg-red-900 flex items-center justify-center mb-4 animate-pulse-scale problem-card-icon">
        {icon}
      </div>
      <h3 className="text-lg font-semibold text-slate-900 dark:text-white problem-card-title">{title}</h3>
    </div>
  );
}