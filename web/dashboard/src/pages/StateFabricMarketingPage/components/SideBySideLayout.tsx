import { ReactNode } from "react";

interface SideBySideLayoutProps {
  left: ReactNode;
  right: ReactNode;
  className?: string;
  reverse?: boolean;
  gap?: string;
}

export function SideBySideLayout({
  left,
  right,
  className = "",
  reverse = false,
  gap = "gap-8"
}: SideBySideLayoutProps) {
  return (
    <div className={`flex flex-col lg:flex-row items-center ${gap} ${className}`}>
      <div className={`flex-1 ${reverse ? 'lg:order-2' : 'lg:order-1'}`}>
        {left}
      </div>
      <div className={`flex-1 ${reverse ? 'lg:order-1' : 'lg:order-2'}`}>
        {right}
      </div>
    </div>
  );
}