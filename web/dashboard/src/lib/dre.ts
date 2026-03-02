import type { CertificateDetailResponse } from "@/api/dre";
import type { FXCertData } from "@/components/dre/fxcert/FXCertViewer";

/** Truncate base64 public key for display as key_id (first N chars). */
function truncateKeyId(publicKeyB64: string, chars: number): string {
  if (!publicKeyB64 || chars <= 0) return "";
  return publicKeyB64.length <= chars
    ? publicKeyB64
    : publicKeyB64.slice(0, chars) + "…";
}

/**
 * Maps backend certificate detail (full FXCERT) to the FXCertViewer display format.
 */
export function mapCertificateDetailToFXCertData(
  res: CertificateDetailResponse
): FXCertData {
  const cert = res.cert;
  const sig = cert.signatures;
  const nodeSig = sig?.node_signature;
  const platformSig = sig?.platform_signature;
  const issuedAt = res.created_at;
  const expiresAt = new Date(issuedAt);
  expiresAt.setFullYear(expiresAt.getFullYear() + 1);

  return {
    certificate_id: res.certificate_id,
    level: (res.cert_level?.toLowerCase() === "legal_grade"
      ? "enterprise"
      : res.cert_level?.toLowerCase() === "lite"
        ? "standard"
        : "standard") as FXCertData["level"],
    certificate_hash: res.certificate_hash,
    execution_root_hash: res.execution_root_hash,
    issued_at: issuedAt,
    expires_at: expiresAt.toISOString(),
    signatures: {
      node: {
        verified: !!nodeSig,
        key_id: nodeSig?.public_key
          ? truncateKeyId(nodeSig.public_key, 16)
          : "",
      },
      platform: {
        verified: !!platformSig,
        key_id: platformSig?.public_key
          ? truncateKeyId(platformSig.public_key, 16)
          : "",
      },
    },
    anchor: res.anchored && cert.anchoring
      ? {
          chain: cert.anchoring.anchor_chain ?? "—",
          block_number: cert.anchoring.anchor_block_number ?? 0,
          tx_hash: cert.anchoring.anchor_tx_hash ?? "—",
          timestamp: cert.anchoring.anchored_at ?? "",
        }
      : undefined,
    metadata: cert.execution
      ? {
          function_id: cert.execution.function_id,
          node_id: cert.execution.node_id,
          region: cert.execution.region,
          protocol_version: cert.execution.protocol_version ?? "",
        }
      : undefined,
  };
}
