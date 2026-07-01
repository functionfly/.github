/**
 * SecretBrowser — Phase 6.1
 *
 * Tree-view secret browser with namespace grouping. Built on
 * top of the namespace list + secret list endpoints.
 *
 * Layout:
 *   [Tree panel]              [List panel]
 *   ├── production/            • STRIPE_API_KEY
 *   │   ├── api-gateway        • DATABASE_URL
 *   │   └── billing            ...
 *   ├── staging
 *   └── shared
 */

import { useCallback, useMemo, useState } from "react";
import {
  ChevronRight,
  FileKey,
  Folder,
  FolderOpen,
  KeyRound,
  ShieldCheck,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import { useNamespaces, useCreateNamespace, vaultApi } from "@/api/vault";
import type { VaultPlan, VaultNamespace } from "@/types/vault-enterprise";
import type { SecretMetadata } from "@/types/vault";
import { PlanGate } from "@/components/VaultEnterprise/PlanGate";
import { useQuery } from "@tanstack/react-query";

interface SecretBrowserProps {
  plan: VaultPlan;
}

interface TreeNode {
  name: string;
  fullPath: string;
  isNamespace: boolean;
  children: TreeNode[];
  namespace?: VaultNamespace;
}

function buildTree(namespaces: VaultNamespace[]): TreeNode {
  const root: TreeNode = { name: "", fullPath: "", isNamespace: false, children: [] };
  const sortAndNest = (nodes: TreeNode[]) => {
    nodes.sort((a, b) => a.name.localeCompare(b.name));
    nodes.forEach((n) => sortAndNest(n.children));
  };
  for (const ns of namespaces) {
    const segs = ns.path.split("/");
    let cur = root;
    let acc = "";
    for (const seg of segs) {
      acc = acc ? `${acc}/${seg}` : seg;
      let child = cur.children.find((c) => c.name === seg && c.isNamespace);
      if (!child) {
        child = {
          name: seg,
          fullPath: acc,
          isNamespace: true,
          children: [],
          namespace: ns.path === acc ? ns : undefined,
        };
        cur.children.push(child);
      }
      cur = child;
    }
  }
  sortAndNest(root.children);
  return root;
}

export function SecretBrowser({ plan }: SecretBrowserProps) {
  const { data: nsResp, isLoading: nsLoading } = useNamespaces();
  const createNamespace = useCreateNamespace();
  const [selectedPath, setSelectedPath] = useState<string>("default");
  const [filter, setFilter] = useState<string>("");
  const [expanded, setExpanded] = useState<Set<string>>(new Set(["default"]));

  const tree = useMemo(() => buildTree(nsResp?.namespaces ?? []), [nsResp]);
  const visibleNamespaces = useMemo(
    () => (nsResp?.namespaces ?? []).filter((n) => n.path.includes(filter.toLowerCase())),
    [nsResp, filter],
  );

  const toggle = (path: string) => {
    setExpanded((s) => {
      const n = new Set(s);
      if (n.has(path)) n.delete(path);
      else n.add(path);
      return n;
    });
  };

  const handleCreateNamespace = useCallback(() => {
    const path = window.prompt("New namespace path (lowercase, dash/underscore, /-separated):");
    if (path && path.trim()) {
      createNamespace.mutate(
        { path: path.trim(), description: "" },
        {
          onSuccess: () => {
            setSelectedPath(path.trim());
            setExpanded((s) => new Set([...s, path.trim()]));
          },
        },
      );
    }
  }, [createNamespace]);

  return (
    <PlanGate feature="namespaces" plan={plan} title="Namespaces require Professional plan or higher">
      <div className="grid gap-4 md:grid-cols-[260px_1fr]">
        <div className="border rounded-md p-3 space-y-2">
          <Input
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            placeholder="Filter namespaces…"
            className="h-8 text-sm"
          />
          {nsLoading ? (
            <div className="space-y-1">
              <Skeleton className="h-6 w-full" />
              <Skeleton className="h-6 w-full" />
              <Skeleton className="h-6 w-full" />
            </div>
          ) : (
            <TreeView
              node={{ ...tree, children: filter ? visibleNamespaces.map(nsToNode) : tree.children }}
              selectedPath={selectedPath}
              onSelect={setSelectedPath}
              expanded={expanded}
              onToggle={toggle}
            />
          )}
          <Button
            size="sm"
            variant="outline"
            className="w-full"
            onClick={handleCreateNamespace}
          >
            + New namespace
          </Button>
        </div>

        <div className="border rounded-md p-3 space-y-2">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium">{selectedPath || "default"}</div>
              <div className="text-xs text-[var(--text-faint)]">
                Secrets in this namespace
              </div>
            </div>
            <Button size="sm" variant="outline">
              + New secret
            </Button>
          </div>
          <SecretListForNamespace path={selectedPath} plan={plan} />
        </div>
      </div>
    </PlanGate>
  );
}

function nsToNode(ns: VaultNamespace): TreeNode {
  return {
    name: ns.path.split("/").pop() ?? ns.path,
    fullPath: ns.path,
    isNamespace: true,
    children: [],
    namespace: ns,
  };
}

function TreeView({
  node,
  selectedPath,
  onSelect,
  expanded,
  onToggle,
  depth = 0,
}: {
  node: TreeNode;
  selectedPath: string;
  onSelect: (path: string) => void;
  expanded: Set<string>;
  onToggle: (path: string) => void;
  depth?: number;
}) {
  if (!node.children.length) return null;
  return (
    <ul className="space-y-0.5 text-sm">
      {node.children.map((c) => {
        const open = expanded.has(c.fullPath);
        const isSel = c.fullPath === selectedPath;
        return (
          <li key={c.fullPath}>
            <button
              type="button"
              onClick={() => {
                onSelect(c.fullPath);
                if (c.children.length) onToggle(c.fullPath);
              }}
              className={`flex w-full items-center gap-1 rounded px-1.5 py-1 text-left ${
                isSel ? "bg-accent" : "hover:bg-[var(--panel-raised)]/50"
              }`}
              style={{ paddingLeft: `${depth * 12 + 6}px` }}
            >
              {c.children.length ? (
                <ChevronRight
                  className={`h-3 w-3 transition-transform ${open ? "rotate-90" : ""}`}
                />
              ) : (
                <span className="w-3" />
              )}
              {open ? (
                <FolderOpen className="h-4 w-4 text-[var(--status-pending)]" />
              ) : (
                <Folder className="h-4 w-4 text-[var(--status-pending)]" />
              )}
              <span className="truncate flex-1">{c.name}</span>
              {c.namespace && (
                <Badge variant="outline" className="ml-auto text-[10px]">
                  {c.namespace.path.split("/").length}
                </Badge>
              )}
            </button>
            {open && (
              <TreeView
                node={c}
                selectedPath={selectedPath}
                onSelect={onSelect}
                expanded={expanded}
                onToggle={onToggle}
                depth={depth + 1}
              />
            )}
          </li>
        );
      })}
    </ul>
  );
}

function SecretListForNamespace({ path, plan }: { path: string; plan: VaultPlan }) {
  const namespaceParam = path === "default" ? "" : path;

  const { data: secretsData, isLoading: secretsLoading } = useQuery({
    queryKey: ["vault", "secrets", "namespace", namespaceParam || "default"],
    queryFn: () => vaultApi.listSecretsWithNamespace(namespaceParam),
    enabled: true,
  });

  if (secretsLoading) {
    return (
      <div className="space-y-2">
        <Skeleton className="h-8 w-full" />
        <Skeleton className="h-8 w-full" />
        <Skeleton className="h-8 w-full" />
      </div>
    );
  }

  const secrets = secretsData?.secrets ?? [];
  const isDefaultNamespace = path === "default";

  if (secrets.length === 0 && isDefaultNamespace) {
    return (
      <div className="text-center py-8 text-[var(--text-faint)]">
        <KeyRound className="h-8 w-8 mx-auto mb-2 opacity-50" />
        <p className="text-sm">No secrets yet</p>
        <p className="text-xs mt-1">
          Create your first secret using the button above
        </p>
      </div>
    );
  }

  if (secrets.length === 0) {
    return (
      <div className="text-center py-8 text-[var(--text-faint)]">
        <Folder className="h-8 w-8 mx-auto mb-2 opacity-50" />
        <p className="text-sm">No secrets in this namespace</p>
        <p className="text-xs mt-1">
          Create a secret and assign it to this namespace
        </p>
      </div>
    );
  }

  return (
    <div className="divide-y">
      {secrets.map((s: SecretMetadata) => (
        <div key={s.id} className="flex items-center justify-between py-2 text-sm">
          <div className="flex items-center gap-2">
            {s.secret_type === "api_key" ? (
              <KeyRound className="h-4 w-4 text-[var(--text-faint)]" />
            ) : s.secret_type === "oauth_token" ? (
              <ShieldCheck className="h-4 w-4 text-[var(--text-faint)]" />
            ) : (
              <FileKey className="h-4 w-4 text-[var(--text-faint)]" />
            )}
            <span className="font-mono">{s.name}</span>
          </div>
          <div className="flex items-center gap-2">
            <Badge variant="outline">{s.secret_type}</Badge>
            {s.expires_at && (
              <Badge variant="secondary">
                {new Date(s.expires_at).toLocaleDateString()}
              </Badge>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}
