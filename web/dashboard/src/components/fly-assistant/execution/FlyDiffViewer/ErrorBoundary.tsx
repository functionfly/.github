/**
 * Error boundary wrapper for the FlyDiffViewer component
 */

import React from "react";
import { ErrorBoundary as ReactErrorBoundary, FallbackProps } from "react-error-boundary";
import { AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { motion } from "framer-motion";

const ErrorFallback: React.FC<FallbackProps> = ({ error, resetErrorBoundary }) => (
  <motion.div
    initial={{ opacity: 0, y: 10 }}
    animate={{ opacity: 1, y: 0 }}
    className="rounded-xl overflow-hidden bg-bg-primary border border-red-500/20 p-8 text-center"
  >
    <AlertTriangle className="h-8 w-8 mx-auto mb-2 text-red-400" />
    <h3 className="text-lg font-medium text-red-400 mb-2">Something went wrong</h3>
    <p className="text-sm text-text-secondary mb-4">
      The diff viewer encountered an unexpected error.
    </p>
    <details className="text-left mb-4">
      <summary className="cursor-pointer text-sm text-red-300">Error details</summary>
      <pre className="mt-2 text-xs text-red-200 bg-red-950/50 p-2 rounded overflow-auto">
        {error instanceof Error ? error.message : "Unknown error"}
      </pre>
    </details>
    <Button variant="outline" onClick={resetErrorBoundary}>
      Try again
    </Button>
  </motion.div>
);

interface DiffErrorBoundaryProps {
  children: React.ReactNode;
  fallback?: React.ComponentType<FallbackProps>;
}

export const DiffErrorBoundary: React.FC<DiffErrorBoundaryProps> = ({
  children,
  fallback
}) => (
  <ReactErrorBoundary FallbackComponent={fallback || ErrorFallback}>
    {children}
  </ReactErrorBoundary>
);
