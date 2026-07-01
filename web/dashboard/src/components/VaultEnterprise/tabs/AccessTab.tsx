/**
 * AccessTab — Phase 4.1
 *
 * RBAC: roles, permissions, and per-user assignments. Built-in
 * roles (admin / operator / reader) are seeded on first use.
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
  ShieldCheck,
  Plus,
  Trash2,
  UserCog,
} from "lucide-react";
import {
  useAssignRole,
  useCreateRole,
  useDeleteRole,
  useMyAssignments,
  useRoles,
  useUnassignRole,
} from "@/api/vault";
import { PlanGate } from "@/components/VaultEnterprise/PlanGate";
import { formatPlan, formatRelativeTime, statusColor } from "@/components/VaultEnterprise/utils";
import type { CreateRoleRequest, VaultPlan, VaultRole } from "@/types/vault-enterprise";

export function AccessTab({ plan }: { plan: VaultPlan }) {
  return (
    <PlanGate feature="rbac" plan={plan} title="Role-based access control" description="Manage roles, JSONB permission sets, and per-user assignments.">
      <div className="space-y-6">
        <RolesCard />
        <AssignmentsCard />
      </div>
    </PlanGate>
  );
}

function RolesCard() {
  const { data, isLoading } = useRoles();
  const create = useCreateRole();
  const del = useDeleteRole();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [permsText, setPermsText] = useState('"secrets:read": true, "secrets:create": true');

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <ShieldCheck className="h-5 w-5" />
          <CardTitle>Roles (Phase 4.1)</CardTitle>
        </div>
        <CardDescription>
          Built-in roles (<code>admin</code>, <code>operator</code>, <code>reader</code>) are
          seeded lazily on first read. Custom roles can layer additional permissions on top.
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
                  <TableHead>Permissions</TableHead>
                  <TableHead>Built-in</TableHead>
                  <TableHead>Updated</TableHead>
                  <TableHead></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.roles?.length ? (
                  data.roles.map((r) => (
                    <TableRow key={r.id}>
                      <TableCell className="font-medium">{r.name}</TableCell>
                      <TableCell>
                        <PermissionChips role={r} />
                      </TableCell>
                      <TableCell>{r.is_builtin ? <Badge>built-in</Badge> : <Badge variant="outline">custom</Badge>}</TableCell>
                      <TableCell className="text-xs">{formatRelativeTime(r.updated_at)}</TableCell>
                      <TableCell>
                        {!r.is_builtin && (
                          <Button size="sm" variant="ghost" onClick={() => del.mutate(r.id)}>
                            <Trash2 className="h-3 w-3" />
                          </Button>
                        )}
                      </TableCell>
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell colSpan={5} className="text-center text-[var(--text-faint)]">
                      No roles yet.
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
            <form
              className="mt-4 grid gap-2 md:grid-cols-2"
              onSubmit={(e) => {
                e.preventDefault();
                try {
                  const perms = JSON.parse(`{${permsText}}`);
                  create.mutate({ name, description, permissions: perms } as CreateRoleRequest);
                  setName("");
                  setDescription("");
                } catch (err) {
                  alert("permissions must be valid JSON object");
                }
              }}
            >
              <Input placeholder="name (e.g. auditor)" value={name} onChange={(e) => setName(e.target.value)} required />
              <Input placeholder="description" value={description} onChange={(e) => setDescription(e.target.value)} />
              <Input
                placeholder='permissions JSON (e.g. "secrets:read": true)'
                value={permsText}
                onChange={(e) => setPermsText(e.target.value)}
                className="md:col-span-2 font-mono text-xs"
              />
              <Button type="submit" disabled={create.isPending} className="md:col-span-2">
                <Plus className="mr-2 h-4 w-4" />
                Create role
              </Button>
            </form>
          </>
        )}
      </CardContent>
    </Card>
  );
}

function PermissionChips({ role }: { role: VaultRole }) {
  const perms = Object.keys(role.permissions).filter((k) => {
    const v = role.permissions[k];
    return v === true || (typeof v === "string" && v === "true");
  });
  const displayed = perms.slice(0, 6);
  return (
    <div className="flex flex-wrap gap-1">
      {displayed.map((p) => (
        <Badge key={p} variant="secondary" className="text-[10px] font-mono">
          {p}
        </Badge>
      ))}
      {perms.length > displayed.length && (
        <Badge variant="outline" className="text-[10px]">+{perms.length - displayed.length}</Badge>
      )}
    </div>
  );
}

function AssignmentsCard() {
  const { data, isLoading } = useMyAssignments();
  const assign = useAssignRole();
  const unassign = useUnassignRole();
  const [userId, setUserId] = useState("");
  const [scope, setScope] = useState("all");

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <UserCog className="h-5 w-5" />
          <CardTitle>My assignments</CardTitle>
        </div>
        <CardDescription>
          Roles assigned to <em>you</em> in this tenant. Admins can assign to
          other users via the API.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <Skeleton className="h-12 w-full" />
        ) : data?.assignments?.length ? (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Role</TableHead>
                <TableHead>Scope</TableHead>
                <TableHead>Created</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.assignments.map((a) => (
                <TableRow key={a.id}>
                  <TableCell className="font-mono text-xs">{a.role_id.slice(0, 8)}…</TableCell>
                  <TableCell>
                    <Badge variant="outline">{a.scope}</Badge>
                  </TableCell>
                  <TableCell className="text-xs">{formatRelativeTime(a.created_at)}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : (
          <p className="text-sm text-[var(--text-faint)]">No role assignments yet.</p>
        )}
      </CardContent>
    </Card>
  );
}
