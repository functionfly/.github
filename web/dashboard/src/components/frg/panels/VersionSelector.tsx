/**
 * VersionSelector Panel
 * Switch graph versions and compare changes
 */

import { useState, useCallback, useEffect } from 'react';
import {
  GitBranch,
  GitCommit,
  GitMerge,
  GitCompare,
  ArrowRightLeft,
  Clock,
  User,
  CheckCircle,
  Circle,
  MoreHorizontal,
  Download,
  Upload,
  Copy,
  Trash2,
  Plus,
  Loader2,
} from 'lucide-react';
import { useQuery } from '@tanstack/react-query';

import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

import { useFRGStore } from '@/stores/frgStore';
import { frgApi } from '@/api/frg';
import type { GraphDefinition } from '@/types/frg';

interface VersionItem {
  version: string;
  createdAt: string;
  isPublished: boolean;
  author?: string;
  message?: string;
}

export function VersionSelector() {
  const store = useFRGStore();
  const { graphAuthor, graphName, selectedVersion, setSelectedVersion, compareVersions, setCompareVersions } = store;

  const [activeTab, setActiveTab] = useState('versions');
  const [compareMode, setCompareMode] = useState(false);
  const [compareFrom, setCompareFrom] = useState<string | null>(null);
  const [compareTo, setCompareTo] = useState<string | null>(null);

  // Fetch versions from API
  const { data: versionData, isLoading: versionsLoading } = useQuery({
    queryKey: ['frg', 'versions', graphAuthor, graphName],
    queryFn: () => {
      if (!graphAuthor || !graphName) return { versions: [] };
      return frgApi.listVersions(graphAuthor, graphName);
    },
    enabled: !!graphAuthor && !!graphName,
  });

  const versions: VersionItem[] = versionData?.versions || [];

  const handleVersionSelect = useCallback((version: string) => {
    if (compareMode) {
      if (!compareFrom || (compareFrom && compareTo)) {
        setCompareFrom(version);
        setCompareTo(null);
      } else if (compareFrom && !compareTo) {
        setCompareTo(version);
        setCompareVersions([compareFrom, version]);
      }
    } else {
      setSelectedVersion(version);
    }
  }, [compareMode, compareFrom, compareTo, setSelectedVersion, setCompareVersions]);

  const clearComparison = useCallback(() => {
    setCompareFrom(null);
    setCompareTo(null);
    setCompareVersions(null);
  }, [setCompareVersions]);

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const handlePublish = async () => {
    if (!graphAuthor || !graphName) return;
    try {
      await frgApi.publishGraph(graphAuthor, graphName);
    } catch (err) {
      console.error('Failed to publish graph:', err);
    }
  };

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="p-3 border-b border-[var(--border-subtle)]">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-lg bg-[var(--bg-tertiary)] flex items-center justify-center">
              <GitBranch className="w-4 h-4 text-[var(--text-secondary)]" />
            </div>
            <div>
              <h3 className="font-semibold text-sm text-[var(--text-primary)]">Versions</h3>
              <p className="text-[10px] text-[var(--text-secondary)]">
                {versionsLoading ? 'Loading...' : `${versions.length} version${versions.length !== 1 ? 's' : ''} available`}
              </p>
            </div>
          </div>
          <div className="flex gap-1">
            <Button
              variant={compareMode ? 'default' : 'ghost'}
              size="sm"
              className="h-7 text-xs"
              onClick={() => {
                setCompareMode(!compareMode);
                if (compareMode) clearComparison();
              }}
            >
              <GitCompare className="w-3 h-3 mr-1" />
              Compare
            </Button>
          </div>
        </div>

        {/* Compare Banner */}
        {compareMode && (
          <div className="mt-3 p-2 bg-brand-500/10 border border-brand-500/20 rounded-lg text-xs">
            <div className="flex items-center justify-between mb-1">
              <span className="font-medium">Compare Mode</span>
              <Button variant="ghost" size="sm" className="h-5 text-[10px]" onClick={clearComparison}>
                Clear
              </Button>
            </div>
            <div className="flex items-center gap-2 text-[10px] text-[var(--text-secondary)]">
              <span className={cn(
                "px-1.5 py-0.5 rounded",
                compareFrom ? "bg-brand-500 text-white" : "bg-[var(--bg-tertiary)]"
              )}>
                {compareFrom || 'select'}
              </span>
              <ArrowRightLeft className="w-3 h-3" />
              <span className={cn(
                "px-1.5 py-0.5 rounded",
                compareTo ? "bg-brand-500 text-white" : "bg-[var(--bg-tertiary)]"
              )}>
                {compareTo || 'select'}
              </span>
            </div>
          </div>
        )}
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col min-h-0">
        <TabsList className="w-full rounded-none border-b border-[var(--border-subtle)] bg-transparent p-0 h-9 px-3">
          <TabsTrigger
            value="versions"
            className="rounded-none data-[state=active]:bg-transparent data-[state=active]:border-b-2 data-[state=active]:border-brand-500 text-xs"
          >
            Versions
          </TabsTrigger>
          <TabsTrigger
            value="forks"
            className="rounded-none data-[state=active]:bg-transparent data-[state=active]:border-b-2 data-[state=active]:border-brand-500 text-xs"
          >
            Forks
          </TabsTrigger>
          <TabsTrigger
            value="history"
            className="rounded-none data-[state=active]:bg-transparent data-[state=active]:border-b-2 data-[state=active]:border-brand-500 text-xs"
          >
            History
          </TabsTrigger>
        </TabsList>

        <ScrollArea className="flex-1">
          <TabsContent value="versions" className="m-0 p-0">
            {versionsLoading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="w-6 h-6 animate-spin text-brand-500" />
              </div>
            ) : versions.length === 0 ? (
              <div className="p-4 text-center">
                <p className="text-sm text-[var(--text-secondary)]">No versions yet</p>
                <p className="text-xs text-[var(--text-muted)] mt-1">
                  Save your graph to create the first version
                </p>
              </div>
            ) : (
              <div className="divide-y divide-[var(--border-subtle)]">
                {versions.map((v, index) => (
                  <div
                    key={v.version}
                    className={cn(
                      "p-3 hover:bg-[var(--bg-tertiary)] transition-colors cursor-pointer",
                      selectedVersion === v.version && !compareMode && "bg-brand-500/10",
                      compareFrom === v.version && compareMode && "bg-blue-500/10",
                      compareTo === v.version && compareMode && "bg-purple-500/10"
                    )}
                    onClick={() => handleVersionSelect(v.version)}
                  >
                    <div className="flex items-start gap-3">
                      <div className="flex flex-col items-center gap-1">
                        {index === 0 ? (
                          <GitCommit className="w-4 h-4 text-brand-500" />
                        ) : (
                          <Circle className="w-4 h-4 text-[var(--text-muted)]" />
                        )}
                        {index < versions.length - 1 && (
                          <div className="w-px h-6 bg-[var(--border-subtle)]" />
                        )}
                      </div>

                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <span className="font-mono text-sm font-medium">v{v.version}</span>
                          {v.isPublished && (
                            <Badge variant="default" className="text-[9px] h-4">
                              Published
                            </Badge>
                          )}
                          {index === 0 && (
                            <Badge variant="secondary" className="text-[9px] h-4">
                              Latest
                            </Badge>
                          )}
                        </div>

                        {v.message && (
                          <p className="text-xs text-[var(--text-secondary)] mt-1 truncate">
                            {v.message}
                          </p>
                        )}

                        <div className="flex items-center gap-2 mt-2 text-[10px] text-[var(--text-muted)]">
                          {v.author && (
                            <span className="flex items-center gap-1">
                              <User className="w-3 h-3" />
                              {v.author}
                            </span>
                          )}
                          <span className="flex items-center gap-1">
                            <Clock className="w-3 h-3" />
                            {formatDate(v.createdAt)}
                          </span>
                        </div>
                      </div>

                      <DropdownMenu>
                        <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
                          <Button variant="ghost" size="icon" className="h-6 w-6">
                            <MoreHorizontal className="w-3 h-3" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem>
                            <Copy className="w-4 h-4 mr-2" />
                            Copy Version
                          </DropdownMenuItem>
                          <DropdownMenuItem>
                            <Download className="w-4 h-4 mr-2" />
                            Export
                          </DropdownMenuItem>
                          <DropdownMenuItem>
                            <GitMerge className="w-4 h-4 mr-2" />
                            Fork
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem className="text-red-500">
                            <Trash2 className="w-4 h-4 mr-2" />
                            Delete
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </TabsContent>

          <TabsContent value="forks" className="m-0 p-4">
            <div className="text-center py-8">
              <div className="w-12 h-12 rounded-full bg-[var(--bg-tertiary)] flex items-center justify-center mx-auto mb-3">
                <GitBranch className="w-6 h-6 text-[var(--text-muted)]" />
              </div>
              <p className="text-sm text-[var(--text-secondary)]">No forks yet</p>
              <p className="text-xs text-[var(--text-muted)] mt-1">
                Fork this graph to create your own version
              </p>
              <Button className="mt-4" size="sm">
                <GitMerge className="w-4 h-4 mr-2" />
                Fork Graph
              </Button>
            </div>
          </TabsContent>

          <TabsContent value="history" className="m-0 p-0">
            <div className="p-4 text-center">
              <p className="text-xs text-[var(--text-muted)]">
                Execution history will appear here
              </p>
            </div>
          </TabsContent>
        </ScrollArea>
      </Tabs>

      {/* Footer */}
      <div className="p-3 border-t border-[var(--border-subtle)]">
        <div className="flex gap-2">
          <Button variant="outline" size="sm" className="flex-1">
            <Download className="w-3 h-3 mr-1" />
            Export
          </Button>
          <Button size="sm" className="flex-1" onClick={handlePublish}>
            <Upload className="w-3 h-3 mr-1" />
            Publish
          </Button>
        </div>
      </div>

      {/* Comparison Dialog */}
      {compareVersions && (
        <Dialog open onOpenChange={() => clearComparison()}>
          <DialogContent className="max-w-2xl">
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2">
                <GitCompare className="w-5 h-5" />
                Version Comparison
                <Badge variant="secondary">
                  {compareVersions[0]} → {compareVersions[1]}
                </Badge>
              </DialogTitle>
              <DialogDescription>
                Compare changes between the selected versions
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4 mt-4">
              <div className="bg-[var(--bg-tertiary)] rounded-lg p-4">
                <h4 className="text-sm font-medium mb-2">Changes</h4>
                <div className="space-y-2 text-sm">
                  <div className="flex items-center gap-2 text-green-500">
                    <Plus className="w-4 h-4" />
                    <span>Version comparison details would appear here</span>
                  </div>
                </div>
              </div>
            </div>

            <div className="flex justify-end gap-2 mt-4">
              <Button variant="outline" onClick={clearComparison}>
                Close
              </Button>
              <Button onClick={clearComparison}>
                <CheckCircle className="w-4 h-4 mr-2" />
                Apply Changes
              </Button>
            </div>
          </DialogContent>
        </Dialog>
      )}
    </div>
  );
}

export default VersionSelector;