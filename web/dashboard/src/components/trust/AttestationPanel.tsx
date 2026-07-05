import React, { useState, useEffect, useCallback } from 'react';
import {
  Chamber,
  CornerBrace,
  StatusPill,
  GaugeStrip,
  TrustSeal,
  SealedButton,
  FrameButton,
  AnnotationTag,
  Spinner,
} from '@/components/sc';
import {
  listAttestations,
  verifyAttestationChain,
  revokeAttestation,
  getAttestationPublicKey,
  getMerkleTreeHead,
  ATTESTATION_TYPE_LABELS,
  ATTESTATION_TYPE_COLORS,
  type Attestation,
  type AttestationType,
  type ChainVerificationResult,
  type AttestationPublicKey,
  type MerkleTreeHead,
} from '@/api/trustapi';
import {
  canViewAttestations,
  canCreateAttestations,
  canRevokeAttestations,
} from '@/lib/plan-utils';

interface AttestationPanelProps {
  functionId: string;
  functionName?: string;
  /** User's current plan tier — gates create/revoke actions */
  plan?: string;
  className?: string;
}

const TYPE_ICONS: Record<AttestationType, string> = {
  verification: '\u2713',
  security_scan: '\u26a0',
  code_review: '\u2022',
  execution: '\u25b6',
  compliance: '\u2611',
  signature: '\u2709',
  delegation: '\u2192',
};

function truncateHash(hash: string, len = 12): string {
  if (!hash) return '-';
  if (hash.length <= len * 2) return hash;
  return `${hash.slice(0, len)}...${hash.slice(-len)}`;
}

function formatDate(iso: string): string {
  if (!iso) return '-';
  return new Date(iso).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function attestationStatus(status: string): 'live' | 'pending' | 'revoked' {
  switch (status) {
    case 'valid':
      return 'live';
    case 'expired':
      return 'pending';
    case 'revoked':
      return 'revoked';
    default:
      return 'pending';
  }
}

export const AttestationPanel: React.FC<AttestationPanelProps> = ({
  functionId,
  functionName,
  plan,
  className = '',
}) => {
  const [attestations, setAttestations] = useState<Attestation[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [chainResult, setChainResult] = useState<ChainVerificationResult | null>(null);
  const [publicKey, setPublicKey] = useState<AttestationPublicKey | null>(null);
  const [merkleHead, setMerkleHead] = useState<MerkleTreeHead | null>(null);
  const [loading, setLoading] = useState(true);
  const [verifying, setVerifying] = useState(false);
  const [selectedType, setSelectedType] = useState<AttestationType | ''>('');
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [revokeTarget, setRevokeTarget] = useState<string | null>(null);
  const [revokeReason, setRevokeReason] = useState('');
  const [revoking, setRevoking] = useState(false);

  const hasReadAccess = canViewAttestations(plan);
  const hasCreateAccess = canCreateAttestations(plan);
  const hasRevokeAccess = canRevokeAttestations(plan);

  const fetchData = useCallback(async () => {
    if (!hasReadAccess) {
      setLoading(false);
      return;
    }
    setLoading(true);
    try {
      const params: Record<string, unknown> = { page_size: 50 };
      if (selectedType) params.type = selectedType;

      const [listRes, pkRes, merkleRes] = await Promise.all([
        listAttestations(functionId, params as { type?: string; page_size?: number }),
        getAttestationPublicKey().catch(() => null),
        getMerkleTreeHead().catch(() => null),
      ]);
      setAttestations(listRes.attestations ?? []);
      setTotalCount(listRes.total_count ?? 0);
      if (pkRes) setPublicKey(pkRes);
      if (merkleRes && merkleRes.tree_size > 0) setMerkleHead(merkleRes);
    } catch {
      setAttestations([]);
    } finally {
      setLoading(false);
    }
  }, [functionId, selectedType]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleVerifyChain = async () => {
    setVerifying(true);
    try {
      const result = await verifyAttestationChain(functionId);
      setChainResult(result);
    } catch {
      setChainResult(null);
    } finally {
      setVerifying(false);
    }
  };

  const handleRevoke = async (attestationId: string) => {
    if (!revokeReason.trim()) return;
    setRevoking(true);
    try {
      await revokeAttestation(attestationId, revokeReason);
      setRevokeTarget(null);
      setRevokeReason('');
      await fetchData();
    } catch {
      // error handled by caller
    } finally {
      setRevoking(false);
    }
  };

  const gauges = [
    { value: totalCount.toString(), label: 'Total' },
    {
      value: attestations.filter((a) => a.status === 'valid').length.toString(),
      label: 'Valid',
    },
    {
      value: attestations.filter((a) => a.status === 'revoked').length.toString(),
      label: 'Revoked',
    },
    {
      value: attestations.filter((a) => a.signature).length.toString(),
      label: 'Signed',
    },
  ];

  const types: Array<AttestationType | ''> = [
    '',
    'verification',
    'security_scan',
    'code_review',
    'execution',
    'compliance',
    'signature',
  ];

  return (
    <div className={`space-y-4 ${className}`}>
      {/* Upgrade prompt for plans without attestation access */}
      {!hasReadAccess && (
        <Chamber className="relative p-5">
          <CornerBrace position="tl" />
          <CornerBrace position="tr" />
          <CornerBrace position="bl" />
          <CornerBrace position="br" />
          <div className="text-center py-4">
            <p className="font-mono text-sm font-semibold text-[var(--text-primary)] uppercase tracking-wider mb-2">
              Attestation Ledger
            </p>
            <p className="font-mono text-xs text-[var(--text-faint)] mb-4">
              Upgrade to Starter or higher to view function attestations and verification chains.
            </p>
            <SealedButton size="sm" onClick={() => window.location.href = '/pricing'}>
              View Plans
            </SealedButton>
          </div>
        </Chamber>
      )}

      {/* Header Chamber */}
      {hasReadAccess && (<>
      <Chamber className="relative p-5">
        <CornerBrace position="tl" />
        <CornerBrace position="tr" />
        <CornerBrace position="bl" />
        <CornerBrace position="br" />
        <AnnotationTag position="tl">attestation ledger</AnnotationTag>

        <div className="flex items-start justify-between mb-4 pt-2">
          <div className="flex items-center gap-3">
            <TrustSeal
              label={
                chainResult?.chain_valid
                  ? 'Sealed'
                  : attestations.length > 0
                    ? 'Active'
                    : 'None'
              }
              variant={
                chainResult?.chain_valid
                  ? 'verified'
                  : attestations.length > 0
                    ? 'trust'
                    : 'live'
              }
              size="lg"
            />
            <div>
              <h3 className="font-mono text-sm font-semibold text-[var(--text-primary)] uppercase tracking-wider">
                {functionName ?? 'Function'} Attestations
              </h3>
              <p className="font-mono text-[10px] text-[var(--text-faint)] uppercase tracking-widest mt-0.5">
                Cryptographic proof chain \u00b7 Ed25519 signed
              </p>
            </div>
          </div>

          <FrameButton
            size="sm"
            loading={verifying}
            onClick={handleVerifyChain}
          >
            Verify Chain
          </FrameButton>
        </div>

        <GaugeStrip gauges={gauges} className="mb-3" />

        {chainResult && (
          <div
            className={`flex items-center gap-2 p-2 rounded font-mono text-xs ${
              chainResult.chain_valid
                ? 'bg-[var(--status-ok)]/10 text-[var(--status-ok)]'
                : 'bg-[var(--status-revoked)]/10 text-[var(--status-revoked)]'
            }`}
          >
            <span className="font-semibold">
              {chainResult.chain_valid ? '\u2713 Chain Verified' : '\u2717 Chain Broken'}
            </span>
            <span className="text-[var(--text-faint)]">
              \u00b7 {chainResult.chain_length} attestation{chainResult.chain_length !== 1 ? 's' : ''} \u00b7 {chainResult.algorithm}
            </span>
          </div>
        )}

        {/* Merkle Audit Trail Status */}
        {merkleHead && (
          <div className="flex items-center gap-2 p-2 rounded font-mono text-xs bg-[var(--status-ok)]/5 text-[var(--text-secondary)]">
            <span className="text-[var(--status-ok)] font-semibold">{'\u2713'} Merkle Log</span>
            <span className="text-[var(--text-faint)]">
              \u00b7 {merkleHead.tree_size} leaves \u00b7 root {truncateHash(merkleHead.root_hash, 8)}
            </span>
            {merkleHead.signature && (
              <span className="text-[var(--status-ok)]">\u00b7 signed</span>
            )}
            <span className="text-[var(--text-faint)]">\u00b7 {formatDate(merkleHead.timestamp)}</span>
          </div>
        )}
      </Chamber>

      {/* Type Filter */}
      <div className="flex items-center gap-1 overflow-x-auto pb-1">
        {types.map((t) => (
          <button
            key={t ?? 'all'}
            onClick={() => setSelectedType(t)}
            className={`px-3 py-1 rounded font-mono text-[10px] uppercase tracking-widest transition-colors whitespace-nowrap ${
              selectedType === t
                ? 'bg-[var(--steel-light)] text-[var(--text-primary)]'
                : 'text-[var(--text-faint)] hover:text-[var(--text-secondary)]'
            }`}
          >
            {t ? ATTESTATION_TYPE_LABELS[t] : 'All'}
          </button>
        ))}
      </div>

      {/* Attestation List */}
      {loading ? (
        <Chamber className="p-8 flex items-center justify-center">
          <Spinner size="lg" />
        </Chamber>
      ) : attestations.length === 0 ? (
        <Chamber className="relative p-6 text-center">
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <p className="font-mono text-xs text-[var(--text-faint)] uppercase tracking-widest">
            No attestations recorded
          </p>
        </Chamber>
      ) : (
        <div className="space-y-2">
          {attestations.map((att) => {
            const isExpanded = expandedId === att.attestation_id;
            const typeColor =
              ATTESTATION_TYPE_COLORS[att.type as AttestationType] ?? 'var(--text-faint)';

            return (
              <Chamber key={att.attestation_id} className="relative">
                <CornerBrace position="tl" />
                <CornerBrace position="br" />

                {/* Main Row */}
                <button
                  className="w-full text-left p-4 flex items-center gap-3"
                  onClick={() =>
                    setExpandedId(isExpanded ? null : att.attestation_id)
                  }
                >
                  {/* Type Icon */}
                  <span
                    className="w-7 h-7 rounded flex items-center justify-center font-mono text-sm shrink-0"
                    style={{
                      background: `${typeColor}15`,
                      color: typeColor,
                      border: `1px solid ${typeColor}30`,
                    }}
                  >
                    {TYPE_ICONS[att.type as AttestationType] ?? '?'}
                  </span>

                  {/* Info */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="font-mono text-xs font-semibold text-[var(--text-primary)] truncate">
                        {att.title}
                      </span>
                      <StatusPill
                        status={attestationStatus(att.status)}
                        label={att.status}
                      />
                    </div>
                    <div className="flex items-center gap-3 mt-1">
                      <span className="font-mono text-[10px] text-[var(--text-faint)] uppercase tracking-widest">
                        {ATTESTATION_TYPE_LABELS[att.type as AttestationType] ?? att.type}
                      </span>
                      <span className="font-mono text-[10px] text-[var(--text-faint)]">
                        {formatDate(att.attested_at)}
                      </span>
                      <span className="font-mono text-[10px] text-[var(--text-faint)]">
                        by {att.attester_name ?? att.attester_type}
                      </span>
                    </div>
                  </div>

                  {/* Integrity Indicators */}
                  <div className="flex items-center gap-2 shrink-0">
                    {att.signature && (
                      <span
                        className={`font-mono text-[10px] font-semibold uppercase tracking-widest ${
                          att.signature_valid
                            ? 'text-[var(--status-ok)]'
                            : 'text-[var(--status-revoked)]'
                        }`}
                      >
                        {att.signature_valid ? '\u2713 Sig' : '\u2717 Sig'}
                      </span>
                    )}
                    <span
                      className={`font-mono text-[10px] font-semibold uppercase tracking-widest ${
                        att.is_valid
                          ? 'text-[var(--status-ok)]'
                          : 'text-[var(--status-revoked)]'
                      }`}
                    >
                      {att.is_valid ? '\u2713 Hash' : '\u2717 Hash'}
                    </span>
                  </div>
                </button>

                {/* Expanded Details */}
                {isExpanded && (
                  <div className="px-4 pb-4 border-t border-[var(--steel)] pt-3 space-y-3">
                    {att.description && (
                      <p className="text-xs text-[var(--text-secondary)] leading-relaxed">
                        {att.description}
                      </p>
                    )}

                    <div className="grid grid-cols-2 gap-2">
                      <div>
                        <span className="font-mono text-[9px] text-[var(--text-faint)] uppercase tracking-widest block mb-0.5">
                          Attestation ID
                        </span>
                        <span className="font-mono text-[11px] text-[var(--text-primary)]">
                          {att.attestation_id}
                        </span>
                      </div>
                      <div>
                        <span className="font-mono text-[9px] text-[var(--text-faint)] uppercase tracking-widest block mb-0.5">
                          Function Version
                        </span>
                        <span className="font-mono text-[11px] text-[var(--text-primary)]">
                          {att.function_version ?? '-'}
                        </span>
                      </div>
                      <div>
                        <span className="font-mono text-[9px] text-[var(--text-faint)] uppercase tracking-widest block mb-0.5">
                          Proof Hash
                        </span>
                        <span className="font-mono text-[11px] text-[var(--text-primary)] break-all">
                          {truncateHash(att.proof_hash, 16)}
                        </span>
                      </div>
                      <div>
                        <span className="font-mono text-[9px] text-[var(--text-faint)] uppercase tracking-widest block mb-0.5">
                          Previous Hash
                        </span>
                        <span className="font-mono text-[11px] text-[var(--text-primary)] break-all">
                          {truncateHash(att.previous_hash, 16)}
                        </span>
                      </div>
                      {att.signature && (
                        <div>
                          <span className="font-mono text-[9px] text-[var(--text-faint)] uppercase tracking-widest block mb-0.5">
                            Signature
                          </span>
                          <span className="font-mono text-[11px] text-[var(--text-primary)] break-all">
                            {truncateHash(att.signature, 16)}
                          </span>
                        </div>
                      )}
                      {att.public_key_id && (
                        <div>
                          <span className="font-mono text-[9px] text-[var(--text-faint)] uppercase tracking-widest block mb-0.5">
                            Key ID
                          </span>
                          <span className="font-mono text-[11px] text-[var(--text-primary)]">
                            {att.public_key_id}
                          </span>
                        </div>
                      )}
                      {att.source_data_hash && (
                        <div>
                          <span className="font-mono text-[9px] text-[var(--text-faint)] uppercase tracking-widest block mb-0.5">
                            Source Hash
                          </span>
                          <span className="font-mono text-[11px] text-[var(--text-primary)] break-all">
                            {truncateHash(att.source_data_hash, 16)}
                          </span>
                        </div>
                      )}
                      {att.code_hash && (
                        <div>
                          <span className="font-mono text-[9px] text-[var(--text-faint)] uppercase tracking-widest block mb-0.5">
                            Code Hash
                          </span>
                          <span className="font-mono text-[11px] text-[var(--text-primary)] break-all">
                            {truncateHash(att.code_hash, 16)}
                          </span>
                        </div>
                      )}
                      {att.input_hash && (
                        <div>
                          <span className="font-mono text-[9px] text-[var(--text-faint)] uppercase tracking-widest block mb-0.5">
                            Input Hash
                          </span>
                          <span className="font-mono text-[11px] text-[var(--text-primary)] break-all">
                            {truncateHash(att.input_hash, 16)}
                          </span>
                        </div>
                      )}
                      {att.output_hash && (
                        <div>
                          <span className="font-mono text-[9px] text-[var(--text-faint)] uppercase tracking-widest block mb-0.5">
                            Output Hash
                          </span>
                          <span className="font-mono text-[11px] text-[var(--text-primary)] break-all">
                            {truncateHash(att.output_hash, 16)}
                          </span>
                        </div>
                      )}
                      {att.valid_until && (
                        <div>
                          <span className="font-mono text-[9px] text-[var(--text-faint)] uppercase tracking-widest block mb-0.5">
                            Valid Until
                          </span>
                          <span className="font-mono text-[11px] text-[var(--text-primary)]">
                            {formatDate(att.valid_until)}
                          </span>
                        </div>
                      )}
                    </div>

                    {/* Results JSON */}
                    {att.results && Object.keys(att.results).length > 0 && (
                      <div>
                        <span className="font-mono text-[9px] text-[var(--text-faint)] uppercase tracking-widest block mb-1">
                          Results
                        </span>
                        <pre className="font-mono text-[10px] text-[var(--text-secondary)] bg-[var(--chamber-bg)] rounded p-2 overflow-x-auto max-h-32">
                          {JSON.stringify(att.results, null, 2)}
                        </pre>
                      </div>
                    )}

                    {/* Revocation Info */}
                    {att.status === 'revoked' && (
                      <div className="p-2 rounded bg-[var(--status-revoked)]/10">
                        <span className="font-mono text-[10px] text-[var(--status-revoked)] font-semibold uppercase tracking-widest">
                          Revoked
                        </span>
                        {att.revoke_reason && (
                          <p className="font-mono text-[11px] text-[var(--text-secondary)] mt-1">
                            {att.revoke_reason}
                          </p>
                        )}
                        {att.revoked_at && (
                          <p className="font-mono text-[10px] text-[var(--text-faint)] mt-0.5">
                            {formatDate(att.revoked_at)}
                          </p>
                        )}
                      </div>
                    )}

                    {/* Revoke Action — Enterprise+ only */}
                    {att.status === 'valid' && hasRevokeAccess && (
                      <div className="pt-2 border-t border-[var(--steel)]">
                        {revokeTarget === att.attestation_id ? (
                          <div className="space-y-2">
                            <input
                              type="text"
                              value={revokeReason}
                              onChange={(e) => setRevokeReason(e.target.value)}
                              placeholder="Reason for revocation..."
                              className="w-full px-3 py-1.5 rounded bg-[var(--chamber-bg)] border border-[var(--steel)] font-mono text-xs text-[var(--text-primary)] placeholder:text-[var(--text-faint)] focus:outline-none focus:border-[var(--steel-light)]"
                            />
                            <div className="flex items-center gap-2">
                              <SealedButton
                                size="sm"
                                loading={revoking}
                                onClick={() => handleRevoke(att.attestation_id)}
                              >
                                Confirm Revocation
                              </SealedButton>
                              <FrameButton
                                size="sm"
                                onClick={() => {
                                  setRevokeTarget(null);
                                  setRevokeReason('');
                                }}
                              >
                                Cancel
                              </FrameButton>
                            </div>
                          </div>
                        ) : (
                          <FrameButton
                            size="sm"
                            onClick={() => setRevokeTarget(att.attestation_id)}
                          >
                            Revoke Attestation
                          </FrameButton>
                        )}
                      </div>
                    )}
                  </div>
                )}
              </Chamber>
            );
          })}
        </div>
      )}

      {/* Public Key Info */}
      {publicKey && (
        <Chamber className="relative p-4">
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <AnnotationTag position="tr">signing key</AnnotationTag>

          <div className="flex items-center justify-between">
            <div>
              <span className="font-mono text-[10px] text-[var(--text-faint)] uppercase tracking-widest block mb-0.5">
                Verification Key ({publicKey.algorithm})
              </span>
              <span className="font-mono text-[11px] text-[var(--text-primary)] break-all">
                {truncateHash(publicKey.public_key, 20)}
              </span>
            </div>
            <span className="font-mono text-[10px] text-[var(--text-faint)]">
              {publicKey.key_id}
            </span>
          </div>
        </Chamber>
      )}
      </>)}
    </div>
  );
};

AttestationPanel.displayName = 'AttestationPanel';
