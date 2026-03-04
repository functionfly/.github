/**
 * FlyDiffViewer with built-in error boundary
 */

import React from "react";
import { FlyDiffViewer } from "./FlyDiffViewer";
import { DiffErrorBoundary } from "./ErrorBoundary";
import type { FlyDiffViewerWithBoundaryProps } from "./types";

export const FlyDiffViewerWithBoundary = React.memo<FlyDiffViewerWithBoundaryProps>(
  ({ enableErrorBoundary = true, ...props }) => {
    if (!enableErrorBoundary) {
      return <FlyDiffViewer {...props} />;
    }

    return (
      <DiffErrorBoundary>
        <FlyDiffViewer {...props} />
      </DiffErrorBoundary>
    );
  }
);

FlyDiffViewerWithBoundary.displayName = "FlyDiffViewerWithBoundary";