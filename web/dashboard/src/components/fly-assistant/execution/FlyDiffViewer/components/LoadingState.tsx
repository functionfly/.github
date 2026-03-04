/**
 * Loading state component for diff computation
 */

import React from "react";
import { Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";

interface LoadingStateProps {
  className?: string;
}

export const LoadingState: React.FC<LoadingStateProps> = ({ className }) => (
  <div className={cn("flex items-center justify-center p-8", className)}>
    <Loader2 className="h-6 w-6 animate-spin mr-2" />
    <span className="text-text-secondary">Computing diff...</span>
  </div>
);