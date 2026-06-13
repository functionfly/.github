/**
 * EnterpriseTab — Phase 4.3 / 4.4 / 4.5 / 4.2 (partial)
 *
 * Sections:
 *   - Namespaces (Phase 4.3)
 *   - Cross-tenant shares (Phase 4.4)
 *   - SSO config (Phase 4.5)
 *   - SIEM webhooks (Phase 4.2)
 *
 * Plan-gating: SSO is Enterprise-only. SIEM webhooks + cross-tenant
 * shares require Team+. Namespaces are available to all plans.
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
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Share2,
  Building2,
  KeyRound,
  Webhook,
  Plus,
  Trash2,
  Network,
  ShieldCheck,
} from "lucide-react";
import {
  useCreateNamespace,
  useCreateSIEMWebhook,
  useDeleteNamespace,
  useDeleteSIEMWebhook,
  useIncomingShares,
  useNamespaces,
  useRevokeShare,
  useSIEMWebhooks,
  useSSOConfig,
  useUpdateSSOConfig,
} from "@/api/vault";
import { PlanGate } from "@/components/VaultEnterprise/PlanGate";
import { formatPlan, formatRelativeTime } from "@/components/VaultEnterprise/utils";
import type { CreateSIEMWebhookRequest, UpdateSSORequest, VaultPlan } from "@/types/vault-enterprise";

export function EnterpriseTab({ plan }: { plan: VaultPlan }) {
  return (
    <div className="space-y-6">
      <NamespacesCard plan={plan} />
      <SharesCard plan={plan} />
      <SSOCard plan={plan} />
      <SIEMCard plan={plan} />
    </div>
  );
}

function NamespacesCard({ plan }: { plan: VaultPlan }) {
  const { data, isLoading } = useNamespaces();
  const create = useCreateNamespace();
  const del = useDeleteNamespace();
  const [path, setPath] = useState("");
  const [description, setDescription] = useState("");
  return (
    <PlanGate feature="namespaces" plan={plan} title="Namespaces" description="Hierarchical secret organization. Available to all plans.">
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Network className="h-5 w-5" />
            <CardTitle>Namespaces (Phase 4.3)</CardTitle>
          </div>
          <CardDescription>
            Hierarchical paths (lowercase, <code>/</code>-separated). Used by the
            secret browser's tree view.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <Skeleton className="h-12 w-full" />
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Path</TableHead>
                    <TableHead>Description</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data?.namespaces?.length ? (
                    data.namespaces.map((n) => (
                      <TableRow key={n.id}>
                        <TableCell className="font-mono text-xs">{n.path}</TableCell>
                        <TableCell className="text-xs">{n.description}</TableCell>
                        <TableCell className="text-xs">{formatRelativeTime(n.created_at)}</TableCell>
                        <TableCell>
                          <Button size="sm" variant="ghost" onClick={() => del.mutate(n.id)}>
                            <Trash2 className="h-3 w-3" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))
                  ) : (
                    <TableRow>
                      <TableCell colSpan={4} className="text-center text-muted-foreground">
                        No namespaces yet. The <code>default</code> namespace is implicit.
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
              <form
                className="mt-4 grid gap-2 md:grid-cols-3"
                onSubmit={(e) => {
                  e.preventDefault();
                  create.mutate({ path, description });
                  setPath("");
                  setDescription("");
                }}
              >
                <Input placeholder="path (e.g. production/api)" value={path} onChange={(e) => setPath(e.target.value)} required />
                <Input placeholder="description" value={description} onChange={(e) => setDescription(e.target.value)} />
                <Button type="submit" disabled={create.isPending}>
                  <Plus className="mr-2 h-4 w-4" />
                  Add
                </Button>
              </form>
            </>
          )}
        </CardContent>
      </Card>
    </PlanGate>
  );
}

function SharesCard({ plan }: { plan: VaultPlan }) {
  const { data, isLoading } = useIncomingShares();
  const revoke = useRevokeShare();
  return (
    <PlanGate feature="shares" plan={plan} title="Cross-tenant shares" description="Read or read-write access to secrets owned by other tenants in your org.">
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Share2 className="h-5 w-5" />
            <CardTitle>Shared with me (Phase 4.4)</CardTitle>
            <Badge variant="secondary" className="ml-2">{formatPlan(plan)}+</Badge>
          </div>
          <CardDescription>
            The grantee sees the secret under a <code>shared/</code> namespace and can
            decrypt locally with their own passphrase.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <Skeleton className="h-12 w-full" />
          ) : data?.shares?.length ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Secret</TableHead>
                  <TableHead>Source tenant</TableHead>
                  <TableHead>Permissions</TableHead>
                  <TableHead>Expires</TableHead>
                  <TableHead></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data.shares.map((s) => (
                  <TableRow key={s.id}>
                    <TableCell className="font-mono text-xs">{s.secret_id.slice(0, 8)}…</TableCell>
                    <TableCell className="font-mono text-xs">{s.source_tenant_id.slice(0, 8)}…</TableCell>
                    <TableCell>
                      <Badge>{s.permissions}</Badge>
                    </TableCell>
                    <TableCell className="text-xs">{s.expires_at ?? "—"}</TableCell>
                    <TableCell>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => revoke.mutate(s.id)}
                        disabled={revoke.isPending}
                      >
                        <Trash2 className="h-3 w-3" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : (
            <p className="text-sm text-muted-foreground">No shares with this tenant.</p>
          )}
        </CardContent>
      </Card>
    </PlanGate>
  );
}

function SSOCard({ plan }: { plan: VaultPlan }) {
  const { data, isLoading } = useSSOConfig();
  const update = useUpdateSSOConfig();
  return (
    <PlanGate feature="sso" plan={plan} title="SAML SSO" description="Connect your IdP. Just-in-time provisioning maps SAML attributes to vault roles.">
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Building2 className="h-5 w-5" />
            <CardTitle>SAML SSO (Phase 4.5)</CardTitle>
            {data?.enabled && <Badge variant="default" className="ml-2">Enabled</Badge>}
          </div>
          <CardDescription>
            Map SAML attributes to roles via the <code>attribute_role_mapping</code> JSON map.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading || !data ? (
            <Skeleton className="h-24 w-full" />
          ) : (
            <div className="grid gap-3 md:grid-cols-2">
              <BoolField
                label="Enable SSO"
                description="When disabled, users sign in with email + password only."
                checked={data.enabled}
                onChange={(v) => update.mutate({ enabled: v })}
              />
              <BoolField
                label="JIT provisioning"
                description="Create user accounts on first SAML sign-in."
                checked={data.jit_provisioning_enabled}
                onChange={(v) => update.mutate({ jit_provisioning_enabled: v })}
              />
              <Field label="SAML metadata URL">
                <Input
                  value={data.saml_metadata_url ?? ""}
                  onChange={(e) => update.mutate({ saml_metadata_url: e.target.value })}
                  placeholder="https://idp.example.com/metadata"
                />
              </Field>
              <Field label="SAML entity ID">
                <Input
                  value={data.saml_entity_id ?? ""}
                  onChange={(e) => update.mutate({ saml_entity_id: e.target.value })}
                  placeholder="urn:functionfly:vault"
                />
              </Field>
              <Field label="SAML SSO URL">
                <Input
                  value={data.saml_sso_url ?? ""}
                  onChange={(e) => update.mutate({ saml_sso_url: e.target.value })}
                  placeholder="https://idp.example.com/sso"
                />
              </Field>
              <Field label="SAML SLO URL">
                <Input
                  value={data.saml_slo_url ?? ""}
                  onChange={(e) => update.mutate({ saml_slo_url: e.target.value })}
                  placeholder="https://idp.example.com/slo"
                />
              </Field>
              <Field label="X.509 cert">
                <Textarea
                  className="font-mono text-xs"
                  rows={4}
                  placeholder="-----BEGIN CERTIFICATE-----&#10;…&#10;-----END CERTIFICATE-----"
                  onChange={(e) => update.mutate({ saml_x509_cert: e.target.value })}
                />
              </Field>
            </div>
          )}
        </CardContent>
      </Card>
    </PlanGate>
  );
}

function SIEMCard({ plan }: { plan: VaultPlan }) {
  const { data, isLoading } = useSIEMWebhooks();
  const create = useCreateSIEMWebhook();
  const del = useDeleteSIEMWebhook();
  const [name, setName] = useState("");
  const [url, setUrl] = useState("");
  const [format, setFormat] = useState<"json" | "cef">("json");
  return (
    <PlanGate feature="siemWebhooks" plan={plan} title="SIEM webhooks" description="Stream audit events to your SIEM in real time. Each payload is HMAC-signed.">
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Webhook className="h-5 w-5" />
            <CardTitle>SIEM webhooks (Phase 4.2)</CardTitle>
          </div>
          <CardDescription>
            Each webhook receives every audit event in JSON or CEF. The server signs
            each delivery with HMAC-SHA-256 using the secret returned on create.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <Skeleton className="h-12 w-full" />
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>URL</TableHead>
                    <TableHead>Format</TableHead>
                    <TableHead>Last status</TableHead>
                    <TableHead></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data?.webhooks?.length ? (
                    data.webhooks.map((w) => (
                      <TableRow key={w.id}>
                        <TableCell className="font-medium">{w.name}</TableCell>
                        <TableCell className="font-mono text-xs">{w.url}</TableCell>
                        <TableCell>
                          <Badge variant="outline">{w.format}</Badge>
                        </TableCell>
                        <TableCell>
                          {w.last_delivery_status ? (
                            <Badge variant={w.last_delivery_status >= 200 && w.last_delivery_status < 300 ? "default" : "destructive"}>
                              {w.last_delivery_status}
                            </Badge>
                          ) : (
                            "—"
                          )}
                        </TableCell>
                        <TableCell>
                          <Button size="sm" variant="ghost" onClick={() => del.mutate(w.id)}>
                            <Trash2 className="h-3 w-3" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))
                  ) : (
                    <TableRow>
                      <TableCell colSpan={5} className="text-center text-muted-foreground">
                        No SIEM webhooks configured.
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
              <form
                className="mt-4 grid gap-2 md:grid-cols-4"
                onSubmit={(e) => {
                  e.preventDefault();
                  create.mutate({ name, url, format } as CreateSIEMWebhookRequest, {
                    onSuccess: (res) => {
                      setName("");
                      setUrl("");
                      if (res.secret_hmac) {
                        alert(`Webhook created.\n\nSave this HMAC secret — you will not see it again:\n\n${res.secret_hmac}`);
                      }
                    },
                  });
                }}
              >
                <Input placeholder="name" value={name} onChange={(e) => setName(e.target.value)} required />
                <Input placeholder="https://siem.example.com/hook" value={url} onChange={(e) => setUrl(e.target.value)} required />
                <div className="flex gap-1">
                  <Button
                    type="button"
                    variant={format === "json" ? "default" : "outline"}
                    size="sm"
                    onClick={() => setFormat("json")}
                  >
                    JSON
                  </Button>
                  <Button
                    type="button"
                    variant={format === "cef" ? "default" : "outline"}
                    size="sm"
                    onClick={() => setFormat("cef")}
                  >
                    CEF
                  </Button>
                </div>
                <Button type="submit" disabled={create.isPending}>
                  <Plus className="mr-2 h-4 w-4" />
                  Add
                </Button>
              </form>
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
        <p className="text-xs text-muted-foreground mt-1">{description}</p>
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
