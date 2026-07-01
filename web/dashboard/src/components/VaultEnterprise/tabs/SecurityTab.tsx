/**
 * SecurityTab — Phase 1.1, 1.2, 1.3, 1.4, 1.4b
 *
 * Plan-gated sections:
 *   - MFA config           (Pro+)
 *   - Token IP allowlist   (Pro+)  (shown per-token)
 *   - Secret expiration    (Free)  (basic TTL)
 *   - Break-glass          (Pro+)
 *   - Escrow               (Team+)
 */

import { useState } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import {
  AlertCircle,
  CalendarClock,
  Clock4,
  KeySquare,
  LifeBuoy,
  Lock,
  MapPin,
  ShieldCheck,
  Timer,
} from "lucide-react";
import { PlanGate } from "@/components/VaultEnterprise/PlanGate";
import { formatPlan, formatRelativeTime, statusColor } from "@/components/VaultEnterprise/utils";
import {
  useApproveBreakGlass,
  useBreakGlassConfig,
  useBreakGlassList,
  useDenyBreakGlass,
  useDisableEscrow,
  useEnableEscrow,
  useEscrowStatus,
  useRequestBreakGlass,
  useRevokeBreakGlass,
  useUpdateVaultMFA,
  useVaultMFA,
  useVerifyVaultMFA,
} from "@/api/vault";
import type { VaultPlan } from "@/types/vault-enterprise";

export function SecurityTab({ plan }: { plan: VaultPlan }) {
  return (
    <div className="space-y-6">
      <MFACard plan={plan} />
      <IPAllowlistCard plan={plan} />
      <ExpirationCard plan={plan} />
      <BreakGlassCard plan={plan} />
      <EscrowCard plan={plan} />
    </div>
  );
}

// ============================================================================
// Phase 1.1: MFA
// ============================================================================

function MFACard({ plan }: { plan: VaultPlan }) {
  const { data, isLoading } = useVaultMFA();
  const updateMFA = useUpdateVaultMFA();
  const verifyMFA = useVerifyVaultMFA();

  return (
    <PlanGate feature="mfa" plan={plan} title="MFA for vault access" description="Require a TOTP or WebAuthn step for every vault operation.">
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <ShieldCheck className="h-5 w-5" />
            <CardTitle>Vault MFA (Phase 1.1)</CardTitle>
            {data?.mfa_required && <Badge variant="default" className="ml-2">Required</Badge>}
          </div>
          <CardDescription>
            Gate vault operations on a TOTP / WebAuthn challenge. The session is honored for the
            configured TTL after the user proves possession of a factor.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {isLoading || !data ? (
            <Skeleton className="h-24 w-full" />
          ) : (
            <>
              <div className="grid gap-3 md:grid-cols-2">
                <BoolField
                  label="MFA required"
                  description="Refuse vault operations when the session is unverified."
                  checked={data.mfa_required}
                  onChange={(v) => updateMFA.mutate({ mfa_required: v })}
                />
                <Field label="Method">
                  <Select
                    value={data.mfa_method}
                    onValueChange={(v) =>
                      updateMFA.mutate({ mfa_method: v as "totp" | "webauthn" | "both" })
                    }
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="totp">TOTP (authenticator apps)</SelectItem>
                      <SelectItem value="webauthn">WebAuthn (passkeys, biometrics)</SelectItem>
                      <SelectItem value="both">Both (TOTP + WebAuthn)</SelectItem>
                    </SelectContent>
                  </Select>
                </Field>
                <BoolField
                  label="Enforce for tokens"
                  description="Runtime access tokens must satisfy MFA on each use."
                  checked={data.enforce_for_tokens}
                  onChange={(v) => updateMFA.mutate({ enforce_for_tokens: v })}
                />
                <BoolField
                  label="Enforce for API"
                  description="API-issued requests (machine-to-machine) must also satisfy MFA."
                  checked={data.enforce_for_api}
                  onChange={(v) => updateMFA.mutate({ enforce_for_api: v })}
                />
                <Field label="Session TTL (seconds)">
                  <Input
                    type="number"
                    min={60}
                    max={86400}
                    value={data.mfa_session_ttl_seconds}
                    onChange={(e) =>
                      updateMFA.mutate({
                        mfa_session_ttl_seconds: Number(e.target.value),
                      })
                    }
                  />
                </Field>
              </div>
              <div className="flex gap-2">
                <Button onClick={() => verifyMFA.mutate()} disabled={verifyMFA.isPending}>
                  <KeySquare className="mr-2 h-4 w-4" />
                  Verify now
                </Button>
                {verifyMFA.isSuccess && (
                  <Badge variant="secondary" className="self-center">
                    Verified · expires {formatRelativeTime(verifyMFA.data?.expires_at)}
                  </Badge>
                )}
                {verifyMFA.isError && (
                  <Badge variant="destructive" className="self-center">
                    <AlertCircle className="mr-1 h-3 w-3" />
                    Verification failed
                  </Badge>
                )}
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </PlanGate>
  );
}

function BoolField({
  label,
  description,
  checked,
  onChange,
}: {
  label: string;
  description: string;
  checked: boolean;
  onChange: (v: boolean) => void;
}) {
  return (
    <div className="flex items-start justify-between gap-3 rounded-md border p-3">
      <div>
        <div className="font-medium text-sm">{label}</div>
        <p className="text-xs text-[var(--text-faint)] mt-1">{description}</p>
      </div>
      <Switch checked={checked} onCheckedChange={onChange} />
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1.5">
      <label className="text-sm font-medium">{label}</label>
      {children}
    </div>
  );
}

// ============================================================================
// Phase 1.2: IP allowlist
// ============================================================================

function IPAllowlistCard({ plan }: { plan: VaultPlan }) {
  return (
    <PlanGate feature="ipAllowlist" plan={plan} title="Token IP allowlist" description="Restrict a runtime access token to specific CIDR ranges or block individual IPs.">
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <MapPin className="h-5 w-5" />
            <CardTitle>IP allowlist (Phase 1.2)</CardTitle>
          </div>
          <CardDescription>
            Per-token IP policy. CIDR notation (e.g. <code>10.0.0.0/8</code>) or single IPs.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-[var(--text-faint)]">
            Open any secret and edit its tokens to set an IP policy. The runtime middleware
            denies requests whose source IP doesn't match the allow list and isn't on the deny
            list.
          </p>
        </CardContent>
      </Card>
    </PlanGate>
  );
}

// ============================================================================
// Phase 1.3: Expiration
// ============================================================================

function ExpirationCard({ plan }: { plan: VaultPlan }) {
  return (
    <PlanGate feature="expiration" plan={plan} title="Secret expiration" description="Set a TTL on any secret; the background sweeper revokes tokens and notifies when secrets expire.">
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Timer className="h-5 w-5" />
            <CardTitle>Expiration (Phase 1.3)</CardTitle>
            <Badge variant="secondary">Default</Badge>
          </div>
          <CardDescription>
            Set a TTL in days, or pin a specific expiry timestamp. The dashboard surfaces
            expiring_soon and expired badges in the secret browser.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <ExpirationForm />
        </CardContent>
      </Card>
    </PlanGate>
  );
}

function ExpirationForm() {
  const [days, setDays] = useState(90);
  return (
    <div className="flex flex-wrap items-end gap-3">
      <div className="space-y-1.5">
        <label className="text-sm font-medium">Expire after (days)</label>
        <Input
          type="number"
          min={1}
          max={3650}
          value={days}
          onChange={(e) => setDays(Number(e.target.value))}
          className="w-32"
        />
      </div>
      <Button variant="outline" onClick={() => alert("Set expiration for selected secret")}>
        <CalendarClock className="mr-2 h-4 w-4" />
        Apply
      </Button>
    </div>
  );
}

// ============================================================================
// Phase 1.4: Break-glass
// ============================================================================

function BreakGlassCard({ plan }: { plan: VaultPlan }) {
  const { data: config } = useBreakGlassConfig();
  const { data: listResp, isLoading } = useBreakGlassList();
  const request = useRequestBreakGlass();
  const approve = useApproveBreakGlass();
  const deny = useDenyBreakGlass();
  const revoke = useRevokeBreakGlass();
  const [reason, setReason] = useState("");

  return (
    <PlanGate feature="breakGlass" plan={plan} title="Break-glass emergency access" description="Request short-lived elevated access when normal authentication is unavailable.">
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <LifeBuoy className="h-5 w-5" />
            <CardTitle>Break-glass (Phase 1.4)</CardTitle>
            {config && (
              <Badge variant="outline" className="ml-2">
                {config.max_duration_minutes}m max · {config.required_approver_count} approver(s)
              </Badge>
            )}
          </div>
          <CardDescription>
            Request a one-time emergency grant. Requires approval from a configured approver.
            The grant is bound by <em>max_duration_minutes</em>.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Reason for emergency access</label>
            <Textarea
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder="e.g. Primary auth provider outage; need to rotate leaked key"
            />
            <div className="flex gap-2">
              <Button
                disabled={!reason.trim() || request.isPending}
                onClick={() => request.mutate({ reason, duration_minutes: config?.max_duration_minutes ?? 60 })}
              >
                Request emergency access
              </Button>
            </div>
          </div>

          <div>
            <h4 className="text-sm font-medium mb-2">Recent requests</h4>
            {isLoading ? (
              <Skeleton className="h-12 w-full" />
            ) : listResp && listResp.requests.length ? (
              <div className="divide-y">
                {listResp.requests.map((r) => (
                  <div key={r.id} className="py-2 flex items-center justify-between text-sm">
                    <div>
                      <span className="font-mono text-xs text-[var(--text-faint)]">
                        {r.id.slice(0, 8)}…
                      </span>
                      <p className="text-sm">{r.reason}</p>
                      <p className="text-xs text-[var(--text-faint)]">
                        {formatRelativeTime(r.created_at)} ·{" "}
                        {r.duration_minutes}m · status{" "}
                        <Badge variant={statusColor(r.status)}>{r.status}</Badge>
                      </p>
                    </div>
                    <div className="flex gap-1">
                      {r.status === "pending" && (
                        <>
                          <Button
                            size="sm"
                            variant="default"
                            onClick={() => approve.mutate(r.id)}
                            disabled={approve.isPending}
                          >
                            Approve
                          </Button>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => deny.mutate(r.id)}
                            disabled={deny.isPending}
                          >
                            Deny
                          </Button>
                        </>
                      )}
                      {r.status === "approved" && (
                        <Button
                          size="sm"
                          variant="destructive"
                          onClick={() => revoke.mutate(r.id)}
                          disabled={revoke.isPending}
                        >
                          Revoke
                        </Button>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-[var(--text-faint)]">No requests yet.</p>
            )}
          </div>
        </CardContent>
      </Card>
    </PlanGate>
  );
}

// ============================================================================
// Phase 1.4b: Escrow
// ============================================================================

function EscrowCard({ plan }: { plan: VaultPlan }) {
  const { data, isLoading } = useEscrowStatus();
  const enable = useEnableEscrow();
  const disable = useDisableEscrow();
  const [blob, setBlob] = useState("");
  const [iv, setIv] = useState("");
  const [tag, setTag] = useState("");
  const [salt, setSalt] = useState("");
  const [hashes, setHashes] = useState("");

  return (
    <PlanGate feature="escrow" plan={plan} title="Optional escrowed recovery" description="Enterprise add-on: store an encrypted recovery blob that can reset a lost passphrase.">
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Lock className="h-5 w-5" />
            <CardTitle>Escrow (Phase 1.4b)</CardTitle>
            {data?.enabled && <Badge variant="default" className="ml-2">Enabled</Badge>}
          </div>
          <CardDescription>
            All encryption happens client-side. The server stores only the encrypted
            recovery blob, KDF salt, and the security-question hashes. Plan:{" "}
            <strong>{formatPlan(plan)}</strong>
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading || !data ? (
            <Skeleton className="h-12 w-full" />
          ) : data.enabled ? (
            <div className="space-y-3">
              <p className="text-sm">
                Escrow is active. To reset a passphrase, the user must answer at least
                one security question; the recovery blob is then decrypted client-side.
              </p>
              <Button variant="destructive" onClick={() => disable.mutate()} disabled={disable.isPending}>
                Disable escrow
              </Button>
            </div>
          ) : (
            <div className="space-y-3">
              <p className="text-sm">Provide the encrypted blob and KDF parameters.</p>
              <div className="grid gap-2 md:grid-cols-2">
                <Input placeholder="encrypted_blob (base64)" value={blob} onChange={(e) => setBlob(e.target.value)} />
                <Input placeholder="blob_iv (base64)" value={iv} onChange={(e) => setIv(e.target.value)} />
                <Input placeholder="blob_auth_tag (base64)" value={tag} onChange={(e) => setTag(e.target.value)} />
                <Input placeholder="kdf_salt (base64)" value={salt} onChange={(e) => setSalt(e.target.value)} />
                <Input
                  placeholder="security_question_hashes (comma-separated)"
                  value={hashes}
                  onChange={(e) => setHashes(e.target.value)}
                  className="md:col-span-2"
                />
              </div>
              <Button
                onClick={() =>
                  enable.mutate({
                    encrypted_blob: blob,
                    blob_iv: iv,
                    blob_auth_tag: tag,
                    kdf_salt: salt,
                    security_question_hashes: hashes
                      .split(",")
                      .map((h) => h.trim())
                      .filter(Boolean),
                  })
                }
                disabled={!blob || !iv || !tag || !salt || enable.isPending}
              >
                Enable escrow
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
    </PlanGate>
  );
}
