import { LoadingSpinner } from '@/components/common/LoadingSpinner';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { useFabricKeys } from '@/hooks/useStateFabric';
import { ROUTES } from '@/lib/constants';
import { ChevronRight, ExternalLink, Key, Search } from 'lucide-react';
import { useState } from 'react';
import { Link } from 'react-router-dom';

interface FabricKeyBrowserProps {
  fabricId: string;
  fabricName?: string;
}

export function FabricKeyBrowser({ fabricId, fabricName }: FabricKeyBrowserProps) {
  const [prefix, setPrefix] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [page, setPage] = useState(0);
  const [selectedKey, setSelectedKey] = useState<{
    key: string;
    value: Record<string, unknown>;
  } | null>(null);

  const pageSize = 20;
  const { data, isLoading, error } = useFabricKeys(fabricId, {
    prefix: prefix || undefined,
    limit: pageSize,
    offset: page * pageSize,
  });

  const keys = data?.keys ?? [];
  const total = data?.total ?? 0;
  const statePath = data?.statePath ?? '';
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  const handleSearch = () => {
    setPrefix(searchInput.trim());
    setPage(0);
  };

  const formatValue = (value: Record<string, unknown>): string => {
    try {
      const str = JSON.stringify(value);
      return str.length > 80 ? `${str.slice(0, 80)}…` : str;
    } catch {
      return String(value);
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h3 className="text-lg font-semibold text-text-primary">Keys</h3>
          <p className="text-sm text-text-muted">
            Browse durable keys stored in {fabricName ?? 'this fabric'}
          </p>
        </div>
        {statePath && (
          <Button variant="outline" size="sm" asChild>
            <Link to={`${ROUTES.STATE}/${encodeURIComponent(statePath)}`}>
              <ExternalLink className="w-4 h-4 mr-2" />
              Open in State
            </Link>
          </Button>
        )}
      </div>

      <div className="flex gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
          <Input
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
            placeholder="Filter by key prefix (e.g. carts/)"
            className="pl-9"
          />
        </div>
        <Button variant="secondary" onClick={handleSearch}>
          Search
        </Button>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <LoadingSpinner />
        </div>
      ) : error ? (
        <Card className="p-8 text-center text-red-400">
          Failed to load keys: {(error as Error).message}
        </Card>
      ) : keys.length === 0 ? (
        <Card className="p-8 text-center">
          <Key className="w-12 h-12 mx-auto mb-4 text-text-muted" />
          <p className="text-text-muted">No keys found for this fabric</p>
          <p className="text-sm text-text-muted mt-2">
            Keys appear when functions write state via Edge State or the HTTP API
          </p>
        </Card>
      ) : (
        <>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-base flex items-center justify-between">
                <span>{total.toLocaleString()} keys</span>
                {prefix && (
                  <Badge variant="secondary" className="font-normal">
                    prefix: {prefix}
                  </Badge>
                )}
              </CardTitle>
            </CardHeader>
            <CardContent className="p-0">
              <div className="divide-y divide-border-subtle">
                {keys.map((entry) => (
                  <button
                    key={entry.key}
                    type="button"
                    className="w-full flex items-center justify-between px-4 py-3 hover:bg-bg-hover text-left transition-colors"
                    onClick={() => setSelectedKey({ key: entry.key, value: entry.value })}
                  >
                    <div className="min-w-0 flex-1">
                      <code className="text-sm font-mono text-text-primary">{entry.key}</code>
                      <p className="text-xs text-text-muted truncate mt-1">
                        {formatValue(entry.value)}
                      </p>
                    </div>
                    <div className="flex items-center gap-2 shrink-0 ml-4">
                      <span className="text-xs text-text-muted hidden sm:inline">
                        {new Date(entry.updatedAt).toLocaleString()}
                      </span>
                      <ChevronRight className="w-4 h-4 text-text-muted" />
                    </div>
                  </button>
                ))}
              </div>
            </CardContent>
          </Card>

          {totalPages > 1 && (
            <div className="flex items-center justify-between">
              <Button
                variant="outline"
                size="sm"
                disabled={page === 0}
                onClick={() => setPage((p) => p - 1)}
              >
                Previous
              </Button>
              <span className="text-sm text-text-muted">
                Page {page + 1} of {totalPages}
              </span>
              <Button
                variant="outline"
                size="sm"
                disabled={page + 1 >= totalPages}
                onClick={() => setPage((p) => p + 1)}
              >
                Next
              </Button>
            </div>
          )}
        </>
      )}

      <Dialog open={!!selectedKey} onOpenChange={(open) => !open && setSelectedKey(null)}>
        <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="font-mono text-base break-all">{selectedKey?.key}</DialogTitle>
          </DialogHeader>
          <pre className="text-sm font-mono bg-bg-secondary p-4 rounded-lg overflow-x-auto">
            {JSON.stringify(selectedKey?.value, null, 2)}
          </pre>
        </DialogContent>
      </Dialog>
    </div>
  );
}
