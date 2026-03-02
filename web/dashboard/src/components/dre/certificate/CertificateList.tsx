import { useQuery } from "@tanstack/react-query";
import { Shield, ChevronLeft, ChevronRight } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { CertificateCard } from "./CertificateCard";
import { dreApi, type CertificateListItem } from "@/api/dre";
import { LoadingSpinner } from "@/components/common/LoadingSpinner";
import { ErrorMessage } from "@/components/common/ErrorMessage";

const PAGE_SIZE = 20;

export interface CertificateListProps {
  author: string;
  name: string;
  onViewCert: (certId: string) => void;
  className?: string;
}

export function CertificateList({
  author,
  name,
  onViewCert,
  className,
}: CertificateListProps) {
  const { data, isLoading, error } = useQuery({
    queryKey: ["certificates", author, name],
    queryFn: () => dreApi.listCertificates(author, name, { limit: PAGE_SIZE, offset: 0 }),
    enabled: !!author && !!name,
  });

  if (isLoading) {
    return (
      <div className={className}>
        <LoadingSpinner />
      </div>
    );
  }

  if (error) {
    return (
      <div className={className}>
        <ErrorMessage error={error instanceof Error ? error : new Error("Failed to load certificates")} />
      </div>
    );
  }

  const certs = data?.certs ?? [];

  if (certs.length === 0) {
    return (
      <Card className={`bg-bg-primary/60 border-border-subtle ${className ?? ""}`}>
        <CardContent className="p-12">
          <div className="text-center">
            <Shield className="w-12 h-12 text-muted-foreground mx-auto mb-4" />
            <h3 className="text-lg font-medium mb-2">No certificates yet</h3>
            <p className="text-muted-foreground">
              FXCERTs are created when this function is executed. Run the function to generate
              execution certificates.
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className={`space-y-4 ${className ?? ""}`}>
      {certs.map((cert: CertificateListItem) => (
        <CertificateCard
          key={cert.certificate_id}
          cert={cert}
          onView={() => onViewCert(cert.certificate_id)}
        />
      ))}
      {certs.length >= PAGE_SIZE && (
        <div className="flex justify-center gap-2 pt-4">
          <Button variant="outline" size="sm" disabled>
            <ChevronLeft className="h-4 w-4" />
            Previous
          </Button>
          <Button variant="outline" size="sm" disabled>
            Next
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      )}
    </div>
  );
}
