import { useQuery } from "@tanstack/react-query";
import { Calendar, ShieldCheck, Package, ChevronRight } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/api/client";
import { dreApi, type CertificateListItem } from "@/api/dre";
import { LoadingSpinner } from "@/components/common/LoadingSpinner";
import { ErrorMessage } from "@/components/common/ErrorMessage";
import { format } from "date-fns";

export interface FunctionHistoryTabProps {
  author: string;
  name: string;
  onViewCert: (certId: string) => void;
  className?: string;
}

/** Function info as returned by GET /registry/functions/{author}/{name} (flat map) */
interface FunctionInfoResponse {
  author?: string;
  name?: string;
  created_at?: string;
  updated_at?: string;
  published_at?: string;
  version?: string;
  [key: string]: unknown;
}

export function FunctionHistoryTab({
  author,
  name,
  onViewCert,
  className,
}: FunctionHistoryTabProps) {
  const { data: functionInfo, isLoading: loadingFn, error: errorFn } = useQuery({
    queryKey: ["function-info", author, name],
    queryFn: () =>
      apiClient.get<FunctionInfoResponse>(`/v1/registry/functions/${author}/${name}`),
    enabled: !!author && !!name,
  });

  const { data: certsData, isLoading: loadingCerts } = useQuery({
    queryKey: ["certificates", author, name],
    queryFn: () => dreApi.listCertificates(author, name, { limit: 50, offset: 0 }),
    enabled: !!author && !!name,
  });

  if (loadingFn) {
    return (
      <div className={className}>
        <LoadingSpinner />
      </div>
    );
  }

  if (errorFn) {
    return (
      <div className={className}>
        <ErrorMessage
          error={errorFn instanceof Error ? errorFn : new Error("Failed to load function info")}
        />
      </div>
    );
  }

  const created_at = functionInfo?.created_at;
  const certs = certsData?.certs ?? [];
  const firstCert = certs.length > 0 ? certs[certs.length - 1] : null; // oldest first if backend orders DESC

  return (
    <div className={className}>
      <Card className="bg-bg-primary/60 border-border-subtle">
        <CardContent className="p-6">
          <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wide mb-4">
            Timeline
          </h3>
          <ul className="space-y-4">
            {created_at && (
              <li className="flex items-start gap-3">
                <div className="p-2 rounded-lg bg-brand-500/10 shrink-0">
                  <Package className="h-4 w-4 text-brand-500" />
                </div>
                <div>
                  <p className="font-medium text-foreground">Function created</p>
                  <p className="text-sm text-muted-foreground">
                    {format(new Date(created_at), "PPpp")}
                  </p>
                </div>
              </li>
            )}
            {loadingCerts ? (
              <li className="flex items-center gap-3 text-muted-foreground">
                <LoadingSpinner />
                <span className="text-sm">Loading certificates…</span>
              </li>
            ) : firstCert ? (
              <li className="flex items-start gap-3">
                <div className="p-2 rounded-lg bg-emerald-500/10 shrink-0">
                  <ShieldCheck className="h-4 w-4 text-emerald-500" />
                </div>
                <div>
                  <p className="font-medium text-foreground">First FXCERT issued</p>
                  <p className="text-sm text-muted-foreground">
                    {format(new Date(firstCert.created_at), "PPpp")}
                  </p>
                  <p className="text-xs text-muted-foreground mt-1 font-mono">
                    {firstCert.certificate_id}
                  </p>
                </div>
              </li>
            ) : null}
          </ul>
        </CardContent>
      </Card>

      {certs.length > 0 && (
        <div className="mt-6">
          <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wide mb-3">
            FXCERTs after publish
          </h3>
          <ul className="space-y-2">
            {certs.map((cert: CertificateListItem) => (
              <li key={cert.certificate_id}>
                <Card
                  className="bg-bg-primary/60 border-border-subtle hover:border-brand-500/30 transition-colors"
                  onClick={() => onViewCert(cert.certificate_id)}
                >
                  <CardContent className="p-3 flex items-center justify-between gap-2">
                    <div className="min-w-0">
                      <p className="font-mono text-sm truncate">{cert.certificate_id}</p>
                      <p className="text-xs text-muted-foreground">
                        {format(new Date(cert.created_at), "PP")}
                      </p>
                    </div>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="shrink-0 gap-1"
                      onClick={(e) => {
                        e.stopPropagation();
                        onViewCert(cert.certificate_id);
                      }}
                    >
                      View
                      <ChevronRight className="h-4 w-4" />
                    </Button>
                  </CardContent>
                </Card>
              </li>
            ))}
          </ul>
        </div>
      )}

      {!loadingCerts && certs.length === 0 && (
        <Card className="mt-6 bg-bg-primary/60 border-border-subtle">
          <CardContent className="p-6 text-center">
            <Calendar className="h-10 w-10 text-muted-foreground mx-auto mb-2" />
            <p className="text-sm text-muted-foreground">
              No FXCERTs yet. Certificates are generated when this function is executed.
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
