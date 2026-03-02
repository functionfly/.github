import { Shield, ExternalLink, Anchor } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { HashBlock } from "../primitives/HashBlock";
import { cn } from "@/lib/utils";
import type { CertificateListItem } from "@/api/dre";
import { format } from "date-fns";

export interface CertificateCardProps {
  cert: CertificateListItem;
  onView: () => void;
  className?: string;
}

const levelColors: Record<string, string> = {
  standard: "bg-blue-500/10 text-blue-500 border-blue-500/20",
  lite: "bg-slate-500/10 text-slate-500 border-slate-500/20",
  legal_grade: "bg-amber-500/10 text-amber-500 border-amber-500/20",
};

export function CertificateCard({ cert, onView, className }: CertificateCardProps) {
  const levelColor =
    levelColors[cert.cert_level?.toLowerCase() ?? ""] ??
    "bg-muted text-muted-foreground border-border-subtle";

  return (
    <Card
      className={cn(
        "bg-bg-primary/60 border-border-subtle hover:border-brand-500/30 transition-colors cursor-pointer",
        className
      )}
      onClick={onView}
    >
      <CardContent className="p-4">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div className="flex items-start gap-3 min-w-0">
            <div className="p-2 rounded-lg bg-brand-500/10 shrink-0">
              <Shield className="h-5 w-5 text-brand-500" />
            </div>
            <div className="min-w-0">
              <p className="font-mono text-sm truncate" title={cert.certificate_id}>
                {cert.certificate_id}
              </p>
              <p className="text-xs text-muted-foreground mt-0.5">
                {format(new Date(cert.created_at), "PPp")}
              </p>
              <div className="flex flex-wrap gap-2 mt-2">
                <Badge variant="outline" className={levelColor}>
                  {cert.cert_level || "standard"}
                </Badge>
                {cert.anchored && (
                  <Badge variant="outline" className="gap-1 bg-emerald-500/10 text-emerald-600 border-emerald-500/20">
                    <Anchor className="h-3 w-3" />
                    Anchored
                  </Badge>
                )}
              </div>
            </div>
          </div>
          <div className="flex items-center gap-2 shrink-0">
            <HashBlock
              hash={cert.execution_root_hash}
              label="Root hash"
              truncate
              truncateChars={8}
              className="text-xs"
            />
            <Button
              variant="ghost"
              size="sm"
              className="gap-1"
              onClick={(e) => {
                e.stopPropagation();
                onView();
              }}
            >
              View
              <ExternalLink className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
