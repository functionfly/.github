/**
 * SecretsTab — Namespace tree + full SecretList
 *
 * Desktop: 2-column layout with namespace tree on left, secrets list on right
 * Mobile (< 768px): namespace dropdown above secrets list
 */

import { useMemo, useState } from "react";
import {
  ChevronRight,
  Folder,
  FolderOpen,
  Plus,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useNamespaces } from "@/api/vault";
import { SecretList } from "@/components/SecretsVault/SecretList";
import { QuotaSummaryBar } from "@/components/VaultPage/components/QuotaSummaryBar";
import type { VaultPlan, VaultNamespace } from "@/types/vault-enterprise";

interface SecretsTabProps {
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
                isSel ? "bg-accent" : "hover:bg-muted/50"
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
                <FolderOpen className="h-4 w-4 text-amber-500" />
              ) : (
                <Folder className="h-4 w-4 text-amber-500" />
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

export function SecretsTab({ plan }: SecretsTabProps) {
  const { data: nsResp, isLoading } = useNamespaces();
  const [selectedPath, setSelectedPath] = useState<string>("default");
  const [filter, setFilter] = useState<string>("");
  const [expanded, setExpanded] = useState<Set<string>>(new Set(["default", "production"]));

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

  return (
    <div className="space-y-4">
      <QuotaSummaryBar plan={plan} />

      <div className="block md:hidden space-y-3">
        <Select value={selectedPath} onValueChange={setSelectedPath}>
          <SelectTrigger>
            <SelectValue placeholder="Select namespace" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="default">default</SelectItem>
            {(nsResp?.namespaces ?? []).map((ns) => (
              <SelectItem key={ns.path} value={ns.path}>
                {ns.path}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <SecretList />
      </div>

      <div className="hidden md:block border rounded-md">
        <div className="grid md:grid-cols-[260px_1fr]">
          <div className="border-r p-3 space-y-2">
            <Input
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
              placeholder="Filter namespaces…"
              className="h-8 text-sm"
            />
            {isLoading ? (
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
              onClick={() => {
                const path = window.prompt("New namespace path (lowercase, dash/underscore, /-separated):");
                if (path) {
                  console.log("create namespace", path);
                }
              }}
            >
              + New namespace
            </Button>
          </div>

          <div className="p-3 space-y-2">
            <div className="flex items-center justify-between">
              <div>
                <div className="text-sm font-medium">{selectedPath || "default"}</div>
                <div className="text-xs text-muted-foreground">
                  Secrets in this namespace
                </div>
              </div>
            </div>
            <SecretList />
          </div>
        </div>
      </div>
    </div>
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
