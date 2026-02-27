import { ReactNode } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { Download, RefreshCw, Settings, Maximize2, Minimize2 } from "lucide-react";

interface ChartsSectionProps {
  title?: string;
  description?: string;
  children: ReactNode;
  className?: string;
  headerClassName?: string;
  contentClassName?: string;

  // Optional features
  showRefresh?: boolean;
  showDownload?: boolean;
  showSettings?: boolean;
  showExpand?: boolean;
  isExpanded?: boolean;
  isLoading?: boolean;

  // Actions
  onRefresh?: () => void;
  onDownload?: () => void;
  onSettings?: () => void;
  onToggleExpand?: () => void;

  // Metadata
  lastUpdated?: string | Date;
  dataSource?: string;
  tags?: string[];
}

export function ChartsSection({
  title,
  description,
  children,
  className,
  headerClassName,
  contentClassName,
  showRefresh = false,
  showDownload = false,
  showSettings = false,
  showExpand = false,
  isExpanded = false,
  isLoading = false,
  onRefresh,
  onDownload,
  onSettings,
  onToggleExpand,
  lastUpdated,
  dataSource,
  tags,
}: ChartsSectionProps) {
  const hasActions = showRefresh || showDownload || showSettings || showExpand;

  return (
    <Card className={cn("chart-container", className)}>
      {(title || hasActions) && (
        <CardHeader className={cn("chart-header", headerClassName)}>
          <div className="flex items-start justify-between">
            <div className="space-y-1">
              {title && <CardTitle className="chart-title">{title}</CardTitle>}
              {description && (
                <CardDescription className="text-text-secondary">
                  {description}
                </CardDescription>
              )}
              {(lastUpdated || dataSource) && (
                <div className="flex items-center gap-4 text-xs text-text-muted">
                  {lastUpdated && (
                    <span>
                      Last updated: {typeof lastUpdated === "string"
                        ? lastUpdated
                        : new Intl.DateTimeFormat("en-US", {
                            month: "short",
                            day: "numeric",
                            hour: "numeric",
                            minute: "2-digit",
                          }).format(new Date(lastUpdated))
                      }
                    </span>
                  )}
                  {dataSource && (
                    <span>Source: {dataSource}</span>
                  )}
                </div>
              )}
              {tags && tags.length > 0 && (
                <div className="flex flex-wrap gap-1 mt-2">
                  {tags.map((tag) => (
                    <Badge
                      key={tag}
                      variant="secondary"
                      className="text-xs bg-indigo-500/10 text-indigo-400 border-indigo-500/20"
                    >
                      {tag}
                    </Badge>
                  ))}
                </div>
              )}
            </div>

            {hasActions && (
              <div className="flex items-center gap-2">
                {showRefresh && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={onRefresh}
                    disabled={isLoading}
                    className="h-8 w-8 p-0"
                  >
                    <RefreshCw className={cn("h-4 w-4", isLoading && "animate-spin")} />
                  </Button>
                )}
                {showDownload && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={onDownload}
                    className="h-8 w-8 p-0"
                  >
                    <Download className="h-4 w-4" />
                  </Button>
                )}
                {showSettings && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={onSettings}
                    className="h-8 w-8 p-0"
                  >
                    <Settings className="h-4 w-4" />
                  </Button>
                )}
                {showExpand && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={onToggleExpand}
                    className="h-8 w-8 p-0"
                  >
                    {isExpanded ? (
                      <Minimize2 className="h-4 w-4" />
                    ) : (
                      <Maximize2 className="h-4 w-4" />
                    )}
                  </Button>
                )}
              </div>
            )}
          </div>
        </CardHeader>
      )}

      <CardContent className={contentClassName}>
        {children}
      </CardContent>
    </Card>
  );
}

// Specialized chart section for multiple charts in a grid
interface ChartsGridProps extends Omit<ChartsSectionProps, "children"> {
  charts: ReactNode[];
  columns?: number;
  gap?: "sm" | "md" | "lg";
}

export function ChartsGrid({
  charts,
  columns = 2,
  gap = "md",
  className,
  ...props
}: ChartsGridProps) {
  const gapClasses = {
    sm: "gap-3",
    md: "gap-4",
    lg: "gap-6",
  };

  return (
    <ChartsSection {...props} className={className}>
      <div
        className={cn(
          "grid",
          {
            "grid-cols-1": columns === 1,
            "grid-cols-2": columns === 2,
            "grid-cols-3": columns === 3,
            "grid-cols-4": columns === 4,
            "md:grid-cols-2 lg:grid-cols-3": columns === 2,
            "md:grid-cols-3 lg:grid-cols-4": columns === 3,
            "md:grid-cols-2 lg:grid-cols-4 xl:grid-cols-5": columns === 4,
          },
          gapClasses[gap]
        )}
      >
        {charts.map((chart, index) => (
          <div key={index} className="min-w-0">
            {chart}
          </div>
        ))}
      </div>
    </ChartsSection>
  );
}