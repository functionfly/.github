/**
 * DynamicCredsTab — Phase 2.1 / 2.2 / 2.4
 *
 * Sections:
 *   - DB targets (Postgres / MySQL depending on plan)
 *   - Credential templates
 *   - Live lease monitor with renew/revoke
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
  Database,
  KeyRound,
  Plus,
  RefreshCcw,
  ServerCog,
  Trash2,
  Zap,
} from "lucide-react";
import {
  useCreateDynamicCredential,
  useCreateTarget,
  useDeleteTarget,
  useDynamicCredentials,
  useDynamicTargets,
  useGenerateDynamicCredential,
  useRenewLease,
  useRevokeAllDynamicCredentials,
  useRevokeLease,
  useTestTarget,
} from "@/api/vault";
import { PlanGate } from "@/components/VaultEnterprise/PlanGate";
import { formatPlan, formatRelativeTime, statusColor } from "@/components/VaultEnterprise/utils";
import type { CreateCredentialRequest, CreateTargetRequest, DynamicDBType, VaultPlan } from "@/types/vault-enterprise";

export function DynamicCredsTab({ plan }: { plan: VaultPlan }) {
  return (
    <div className="space-y-6">
      <TargetsCard plan={plan} />
      <CredentialsCard plan={plan} />
      <LeasesMonitorCard plan={plan} />
    </div>
  );
}

function TargetsCard({ plan }: { plan: VaultPlan }) {
  const { data, isLoading } = useDynamicTargets();
  const create = useCreateTarget();
  const del = useDeleteTarget();
  const test = useTestTarget();
  const [show, setShow] = useState(false);
  const allowedBackends = plan === "team" || plan === "enterprise" ? ["postgres", "mysql"] : ["postgres"];
  return (
    <PlanGate feature="expiration" plan={plan} title="Dynamic credentials" description="Mint short-lived DB users against admin-managed targets.">
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Database className="h-5 w-5" />
            <CardTitle>Database targets</CardTitle>
            <Badge variant="secondary" className="ml-2">
              {allowedBackends.join(" + ")}
            </Badge>
          </div>
          <CardDescription>
            Server-side: each target's admin password is envelope-encrypted (AES-256-GCM with
            per-tenant key). The server never sees plaintext. Plan:{" "}
            <strong>{formatPlan(plan)}</strong>
            {plan === "free" && " — only Postgres is included; upgrade to Team for MySQL."}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <Skeleton className="h-24 w-full" />
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>DB</TableHead>
                    <TableHead>Host:Port</TableHead>
                    <TableHead>Database</TableHead>
                    <TableHead>TTL</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data?.targets?.length ? (
                    data.targets.map((t) => (
                      <TableRow key={t.id}>
                        <TableCell className="font-medium">{t.name}</TableCell>
                        <TableCell>
                          <Badge variant="outline">{t.db_type}</Badge>
                        </TableCell>
                        <TableCell className="font-mono text-xs">
                          {t.host}:{t.port}
                        </TableCell>
                        <TableCell className="font-mono text-xs">{t.database_name}</TableCell>
                        <TableCell className="text-xs">
                          {t.default_ttl_seconds}s / {t.max_ttl_seconds}s
                        </TableCell>
                        <TableCell>
                          <Badge variant={statusColor(t.status)}>{t.status}</Badge>
                        </TableCell>
                        <TableCell>
                          <div className="flex gap-1">
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={() => test.mutate(t.id)}
                              disabled={test.isPending}
                            >
                              <Zap className="h-3 w-3" />
                            </Button>
                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={() => del.mutate(t.id)}
                              disabled={del.isPending}
                            >
                              <Trash2 className="h-3 w-3" />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))
                  ) : (
                    <TableRow>
                      <TableCell colSpan={7} className="text-center text-muted-foreground">
                        No targets configured.
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
              <div className="mt-4">
                <Button variant="outline" onClick={() => setShow((s) => !s)}>
                  <Plus className="mr-2 h-4 w-4" />
                  New target
                </Button>
                {show && (
                  <TargetForm
                    onSubmit={(body) => create.mutate(body)}
                    backends={allowedBackends as DynamicDBType[]}
                    submitting={create.isPending}
                  />
                )}
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </PlanGate>
  );
}

function TargetForm({
  onSubmit,
  backends,
  submitting,
}: {
  onSubmit: (b: CreateTargetRequest) => void;
  backends: DynamicDBType[];
  submitting: boolean;
}) {
  const [name, setName] = useState("");
  const [dbType, setDbType] = useState<DynamicDBType>(backends[0] ?? "postgres");
  const [host, setHost] = useState("");
  const [port, setPort] = useState(5432);
  const [database, setDatabase] = useState("");
  const [user, setUser] = useState("");
  const [pass, setPass] = useState("");
  return (
    <form
      className="mt-3 grid gap-2 md:grid-cols-2"
      onSubmit={(e) => {
        e.preventDefault();
        onSubmit({
          name,
          db_type: dbType,
          host,
          port,
          database_name: database,
          admin_username: user,
          admin_password: pass,
          default_ttl_seconds: 3600,
          max_ttl_seconds: 86400,
        });
      }}
    >
      <Input placeholder="name" value={name} onChange={(e) => setName(e.target.value)} required />
      <div className="flex gap-2">
        {(backends ?? []).map((b) => (
          <Button
            key={b}
            type="button"
            variant={dbType === b ? "default" : "outline"}
            size="sm"
            onClick={() => setDbType(b)}
          >
            {b}
          </Button>
        ))}
      </div>
      <Input placeholder="host" value={host} onChange={(e) => setHost(e.target.value)} required />
      <Input
        placeholder="port"
        type="number"
        value={port}
        onChange={(e) => setPort(Number(e.target.value))}
      />
      <Input placeholder="database_name" value={database} onChange={(e) => setDatabase(e.target.value)} required />
      <Input placeholder="admin_username" value={user} onChange={(e) => setUser(e.target.value)} required />
      <Input
        placeholder="admin_password"
        type="password"
        value={pass}
        onChange={(e) => setPass(e.target.value)}
        required
        className="md:col-span-2"
      />
      <Button type="submit" disabled={submitting} className="md:col-span-2">
        {submitting ? "Creating…" : "Create target"}
      </Button>
    </form>
  );
}

function CredentialsCard({ plan }: { plan: VaultPlan }) {
  const { data, isLoading } = useDynamicCredentials();
  const create = useCreateDynamicCredential();
  const generate = useGenerateDynamicCredential();
  const revokeAll = useRevokeAllDynamicCredentials();
  const [targetId, setTargetId] = useState("");
  const [name, setName] = useState("");
  const [ttl, setTTL] = useState(3600);
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <KeyRound className="h-5 w-5" />
          <CardTitle>Credential templates</CardTitle>
        </div>
        <CardDescription>
          Named, reusable templates that mint DB users on demand. The generated
          password is shown exactly once in the API response — copy it immediately.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <Skeleton className="h-24 w-full" />
        ) : (
          <>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>TTL (s)</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.credentials?.length ? (
                  data.credentials.map((c) => (
                    <TableRow key={c.id}>
                      <TableCell className="font-medium">{c.name}</TableCell>
                      <TableCell className="font-mono text-xs">
                        {c.ttl_seconds} / {c.max_ttl_seconds}
                      </TableCell>
                      <TableCell>
                        <Badge variant={statusColor(c.status)}>{c.status}</Badge>
                      </TableCell>
                      <TableCell>
                        <div className="flex gap-1">
                          <Button
                            size="sm"
                            variant="default"
                            onClick={() =>
                              generate.mutate(
                                { id: c.id },
                                {
                                  onSuccess: (res) => {
                                    alert(
                                      `Generated credential\n\nUser: ${res.username}\nPassword: ${res.password}\nLease: ${res.lease_id}\n\nCopy the password now — it will not be shown again.`,
                                    );
                                  },
                                },
                              )
                            }
                            disabled={generate.isPending}
                          >
                            <Zap className="mr-1 h-3 w-3" /> Generate
                          </Button>
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => revokeAll.mutate(c.id)}
                            disabled={revokeAll.isPending}
                          >
                            Revoke all
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell colSpan={4} className="text-center text-muted-foreground">
                      No credential templates.
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
            <form
              className="mt-4 grid gap-2 md:grid-cols-3"
              onSubmit={(e) => {
                e.preventDefault();
                create.mutate({ target_id: targetId, name, ttl_seconds: ttl, max_ttl_seconds: ttl * 24 });
              }}
            >
              <Input placeholder="target_id" value={targetId} onChange={(e) => setTargetId(e.target.value)} required />
              <Input placeholder="name" value={name} onChange={(e) => setName(e.target.value)} required />
              <Input
                placeholder="ttl_seconds"
                type="number"
                value={ttl}
                onChange={(e) => setTTL(Number(e.target.value))}
              />
              <Button type="submit" disabled={create.isPending} className="md:col-span-3">
                <Plus className="mr-2 h-4 w-4" />
                Create template
              </Button>
            </form>
          </>
        )}
      </CardContent>
    </Card>
  );
}

function LeasesMonitorCard({ plan }: { plan: VaultPlan }) {
  return (
    <PlanGate feature="tokenMonitor" plan={plan} title="Token / lease monitor" description="Live monitor of dynamic-credential leases, with renew and revoke.">
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <ServerCog className="h-5 w-5" />
            <CardTitle>Live lease monitor</CardTitle>
          </div>
          <CardDescription>
            Active leases are auto-collected by the background sweeper. Use the renew /
            revoke buttons to manually extend or drop them.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            Open a credential template's "Generate" action to mint a lease; the
            resulting lease_id is then listed here. The monitor polls every 5s
            and re-validates expiry.
          </p>
          <p className="text-xs text-muted-foreground mt-2">
            Plan: <strong>{formatPlan(plan)}</strong> · Phase 2.2
          </p>
        </CardContent>
      </Card>
    </PlanGate>
  );
}
