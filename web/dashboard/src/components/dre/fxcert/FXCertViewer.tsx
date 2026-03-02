import { useState } from "react";
import { encode as cborEncode } from "cbor-x";
import { jsPDF } from "jspdf";
import { FileText, Shield, Download, ChevronDown, ChevronRight } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { CollapsibleSection } from "../primitives/CollapsibleSection";
import { SignatureVerificationBadge } from "./SignatureVerificationBadge";
import { AnchorStatusCard } from "./AnchorStatusCard";
import { HashBlock } from "../primitives/HashBlock";
import { cn } from "@/lib/utils";

export interface FXCertData {
  certificate_id: string;
  level: "standard" | "extended" | "enterprise";
  certificate_hash: string;
  execution_root_hash: string;
  issued_at: string;
  expires_at: string;
  signatures: {
    node: { verified: boolean; key_id: string };
    platform: { verified: boolean; key_id: string };
  };
  anchor?: {
    chain: string;
    block_number: number;
    tx_hash: string;
    timestamp: string;
  };
  metadata?: Record<string, string>;
}

export interface FXCertViewerProps {
  /** Certificate data */
  certificate: FXCertData;
  /** Show detailed view */
  showDetails?: boolean;
  /** Custom className */
  className?: string;
  /** Full FXCERT JSON string for download (when opened from getCertificate) */
  fullCertJson?: string;
}

const levelColors = {
  standard: "bg-blue-500/10 text-blue-500 border-blue-500/20",
  extended: "bg-purple-500/10 text-purple-500 border-purple-500/20",
  enterprise: "bg-gold-500/10 text-gold-500 border-gold-500/20",
};

function effectiveKeyId(keyId: string | undefined): string | null {
  const s = (keyId ?? "").trim();
  if (s === "" || s === "—") return null;
  return s;
}

export function FXCertViewer({
  certificate,
  showDetails = false,
  className,
  fullCertJson,
}: FXCertViewerProps) {
  const [expanded, setExpanded] = useState(showDetails);

  const isExpired = new Date(certificate.expires_at) < new Date();

  const handleDownload = (format: "json" | "cbor" | "pdf") => {
    if (format === "json" && fullCertJson) {
      const blob = new Blob([fullCertJson], { type: "application/json" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `${certificate.certificate_id}.fxcert.json`;
      a.click();
      URL.revokeObjectURL(url);
      return;
    }
    if (format === "json" && !fullCertJson) {
      const summary = JSON.stringify(certificate, null, 2);
      const blob = new Blob([summary], { type: "application/json" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `${certificate.certificate_id}-summary.json`;
      a.click();
      URL.revokeObjectURL(url);
      return;
    }
    const payload: object = fullCertJson ? (JSON.parse(fullCertJson) as object) : certificate;

    if (format === "cbor") {
      const bytes = cborEncode(payload);
      const blob = new Blob([new Uint8Array(bytes)], { type: "application/cbor" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `${certificate.certificate_id}.fxcert.cbor`;
      a.click();
      URL.revokeObjectURL(url);
      return;
    }

    if (format === "pdf") {
      const doc = new jsPDF({ format: "a4" });
      const margin = 20;
      let y = 20;
      const lineHeight = 7;
      doc.setFontSize(16);
      doc.text("FXCERT", margin, y);
      y += lineHeight + 4;
      doc.setFontSize(10);
      doc.text(`Certificate ID: ${certificate.certificate_id}`, margin, y);
      y += lineHeight;
      doc.text(`Level: ${certificate.level}`, margin, y);
      y += lineHeight;
      doc.text(`Issued: ${certificate.issued_at}`, margin, y);
      y += lineHeight;
      doc.text(`Expires: ${certificate.expires_at}`, margin, y);
      y += lineHeight;
      doc.text(`Execution root hash: ${certificate.execution_root_hash}`, margin, y);
      y += lineHeight;
      doc.text(`Certificate hash: ${certificate.certificate_hash}`, margin, y);
      y += lineHeight + 4;
      const jsonLines = doc.splitTextToSize(
        fullCertJson ?? JSON.stringify(certificate, null, 2),
        doc.internal.pageSize.getWidth() - margin * 2
      );
      doc.setFontSize(8);
      doc.text("Certificate data (JSON):", margin, y);
      y += lineHeight;
      doc.text(jsonLines, margin, y);
      const blob = doc.output("blob");
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `${certificate.certificate_id}.fxcert.pdf`;
      a.click();
      URL.revokeObjectURL(url);
    }
  };

  return (
    <Card className={cn(className)}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-brand-500/10 rounded-lg">
              <Shield className="h-5 w-5 text-brand-500" />
            </div>
            <div>
              <CardTitle className="text-base flex items-center gap-2">
                FX Certificate
              </CardTitle>
              <p className="text-xs text-muted-foreground font-mono">
                {certificate.certificate_id}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Badge className={levelColors[certificate.level]}>
              {certificate.level.toUpperCase()}
            </Badge>
            {isExpired && (
              <Badge variant="outline" className="text-red-500 border-red-500">
                Expired
              </Badge>
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Quick Info */}
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-muted-foreground">Issued</span>
            <p>{new Date(certificate.issued_at).toLocaleDateString()}</p>
          </div>
          <div>
            <span className="text-muted-foreground">Expires</span>
            <p className={isExpired ? "text-red-500" : ""}>
              {new Date(certificate.expires_at).toLocaleDateString()}
            </p>
          </div>
        </div>

        {/* Execution Root Hash */}
        <HashBlock
          hash={certificate.execution_root_hash}
          label="Execution Root Hash"
          truncate
          truncateChars={12}
        />

        {/* Certificate Hash */}
        <HashBlock
          hash={certificate.certificate_hash}
          label="Certificate Hash"
          truncate
          truncateChars={12}
        />

        {/* Signatures */}
        <div className="space-y-2">
          <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
            Signature Verification
          </span>
          <SignatureVerificationBadge
            nodeVerified={certificate.signatures.node.verified}
            platformVerified={certificate.signatures.platform.verified}
            expired={isExpired}
          />
        </div>

        {/* Expandable Details */}
        <CollapsibleSection
          title="Certificate Details"
          icon={<FileText className="h-4 w-4" />}
          open={expanded}
          onOpenChange={setExpanded}
        >
          <div className="space-y-4">
            {/* Key IDs — show placeholder when cert has no signature (e.g. bootstrap) */}
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span className="text-muted-foreground text-xs">Node Key ID</span>
                <p className="font-mono text-xs">
                  {effectiveKeyId(certificate.signatures.node.key_id) ?? (
                    <span className="text-muted-foreground italic">Not set (e.g. bootstrap cert)</span>
                  )}
                </p>
              </div>
              <div>
                <span className="text-muted-foreground text-xs">Platform Key ID</span>
                <p className="font-mono text-xs">
                  {effectiveKeyId(certificate.signatures.platform.key_id) ?? (
                    <span className="text-muted-foreground italic">Not set</span>
                  )}
                </p>
              </div>
            </div>

            {/* Anchor Info */}
            {certificate.anchor && (
              <AnchorStatusCard
                anchored={true}
                chain={certificate.anchor.chain}
                blockNumber={certificate.anchor.block_number}
                txHash={certificate.anchor.tx_hash}
                timestamp={certificate.anchor.timestamp}
              />
            )}

            {/* Custom Metadata */}
            {certificate.metadata && Object.keys(certificate.metadata).length > 0 && (
              <div className="space-y-2">
                <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                  Metadata
                </span>
                <div className="bg-bg-secondary rounded-md p-3 space-y-1">
                  {Object.entries(certificate.metadata).map(([key, value]) => (
                    <div key={key} className="flex justify-between text-sm">
                      <span className="text-muted-foreground">{key}</span>
                      <span className="font-mono text-xs">{value}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </CollapsibleSection>

        {/* Download Actions */}
        <div className="flex flex-wrap gap-2 pt-2 border-t border-border-subtle">
          <Button
            variant="outline"
            size="sm"
            onClick={() => handleDownload("json")}
            className="gap-2"
          >
            <Download className="h-4 w-4" />
            JSON
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => handleDownload("cbor")}
            className="gap-2"
          >
            <Download className="h-4 w-4" />
            CBOR
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => handleDownload("pdf")}
            className="gap-2"
          >
            <Download className="h-4 w-4" />
            PDF
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
