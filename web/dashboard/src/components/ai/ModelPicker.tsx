import { aiModelsApi, type ModelSelection } from '@/api/aiModels';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { useResponsivePageSize } from '@/hooks/useResponsivePageSize';
import { cn } from '@/lib/utils';
import { useMutation, useQuery } from '@tanstack/react-query';
import { AlertTriangle, CheckCircle, ChevronLeft, ChevronRight, Loader2, XCircle } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';

type Props = {
  feature: string;
  value?: ModelSelection;
  onChange: (next: ModelSelection | undefined) => void;
  capability?: string;
  compact?: boolean;
  disabled?: boolean;
  showOrgDefaultOption?: boolean;
  showDefaultBadge?: boolean;
};

export function ModelPicker({
  feature,
  value,
  onChange,
  capability = 'code',
  compact = false,
  disabled = false,
  showOrgDefaultOption = true,
}: Props) {
  const [open, setOpen] = useState(false);
  const [page, setPage] = useState(1);
  const pageSize = useResponsivePageSize({ min: 8, max: 16, rowHeight: 36, reservedHeight: 320 });

  const {
    data = [],
    isLoading,
    isError,
  } = useQuery({
    queryKey: ['ai-model-catalog', feature, capability],
    queryFn: () => aiModelsApi.getCatalog(capability, feature),
  });

  const totalPages = Math.max(1, Math.ceil(data.length / pageSize));

  useEffect(() => {
    setPage(1);
  }, [data.length, pageSize]);

  useEffect(() => {
    if (page > totalPages) {
      setPage(totalPages);
    }
  }, [page, totalPages]);

  const jumpToSelectedPage = () => {
    if (!value?.provider || !value.model_id) {
      setPage(1);
      return;
    }
    const idx = data.findIndex((m) => m.provider === value.provider && m.id === value.model_id);
    if (idx >= 0) {
      setPage(Math.floor(idx / pageSize) + 1);
    } else {
      setPage(1);
    }
  };

  const paginated = useMemo(() => {
    const start = (page - 1) * pageSize;
    const slice = data.slice(start, start + pageSize);
    if (!value?.provider || !value.model_id) return slice;

    const selectedKey = `${value.provider}:${value.model_id}`;
    if (slice.some((m) => `${m.provider}:${m.id}` === selectedKey)) {
      return slice;
    }

    const selected = data.find((m) => `${m.provider}:${m.id}` === selectedKey);
    if (!selected) return slice;
    return [selected, ...slice.slice(0, Math.max(0, pageSize - 1))];
  }, [data, page, pageSize, value]);

  const current = value ? `${value.provider}:${value.model_id}` : 'default';

  const checkMutation = useMutation({
    mutationFn: () => aiModelsApi.checkModel(value!.provider, value!.model_id),
  });

  if (isLoading) {
    return (
      <div className="flex h-10 items-center gap-2 rounded-md border border-border-subtle px-3 text-sm text-text-muted">
        <Loader2 className="h-4 w-4 animate-spin" />
        Loading models…
      </div>
    );
  }

  return (
    <div className="space-y-2">
      <Select
        open={open}
        onOpenChange={(next) => {
          setOpen(next);
          if (next) {
            jumpToSelectedPage();
          }
        }}
        value={current}
        onValueChange={(next) => {
          if (next === 'default') {
            onChange(undefined);
            return;
          }
          const [provider, ...rest] = next.split(':');
          onChange({ provider, model_id: rest.join(':') });
          checkMutation.reset();
        }}
        disabled={disabled || (data.length === 0 && !showOrgDefaultOption)}
      >
        <SelectTrigger className={compact ? 'h-8 text-xs' : ''}>
          <SelectValue
            placeholder={
              isError
                ? 'Failed to load models'
                : data.length === 0
                  ? 'No models available'
                  : 'Select model'
            }
          />
        </SelectTrigger>
        <SelectContent>
          {showOrgDefaultOption && <SelectItem value="default">Using org default</SelectItem>}
          {paginated.map((m) => (
            <SelectItem key={`${m.provider}:${m.id}`} value={`${m.provider}:${m.id}`}>
              <span className="flex items-center gap-2">
                <span>{m.display_name}</span>
                <span className="text-text-muted">({m.provider_label || m.provider})</span>
              </span>
            </SelectItem>
          ))}
          {data.length === 0 && !showOrgDefaultOption && (
            <SelectItem value="default" disabled>
              No models in catalog
            </SelectItem>
          )}
          {data.length > pageSize && (
            <div
              className="flex items-center justify-between gap-2 border-t border-border-subtle px-2 py-1.5"
              onPointerDown={(e) => e.preventDefault()}
            >
              <button
                type="button"
                className={cn(
                  'inline-flex items-center gap-1 rounded px-2 py-1 text-xs text-text-muted hover:bg-bg-secondary hover:text-text-primary',
                  page <= 1 && 'pointer-events-none opacity-40'
                )}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page <= 1}
              >
                <ChevronLeft className="h-3.5 w-3.5" />
                Prev
              </button>
              <span className="text-xs text-text-muted">
                {page} / {totalPages}
              </span>
              <button
                type="button"
                className={cn(
                  'inline-flex items-center gap-1 rounded px-2 py-1 text-xs text-text-muted hover:bg-bg-secondary hover:text-text-primary',
                  page >= totalPages && 'pointer-events-none opacity-40'
                )}
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page >= totalPages}
              >
                Next
                <ChevronRight className="h-3.5 w-3.5" />
              </button>
            </div>
          )}
        </SelectContent>
      </Select>

      {value?.provider && value?.model_id && (() => {
        const selectedModel = data.find((m) => m.provider === value.provider && m.id === value.model_id);
        const isTokenPlan = selectedModel?.key_source === 'token-plan';
        const isByok = selectedModel?.key_source === 'byok';
        const providerLabel = selectedModel?.provider_label || value.provider;
        return (
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-xs text-text-muted">{providerLabel}</span>
            {isTokenPlan && (
              <span className="inline-flex items-center gap-1 rounded-md bg-blue-500/10 border border-blue-500/20 px-2 py-0.5 text-xs font-medium text-blue-400">
                Token Plan
              </span>
            )}
            {isByok && (
              <span className="inline-flex items-center gap-1 rounded-md bg-green-500/10 border border-green-500/20 px-2 py-0.5 text-xs font-medium text-green-400">
                BYOK
              </span>
            )}
            <button
              type="button"
              onClick={() => checkMutation.mutate()}
              disabled={checkMutation.isPending}
              className="inline-flex items-center gap-1.5 rounded-md border border-border-subtle px-2.5 py-1 text-xs text-text-muted hover:bg-bg-secondary hover:text-text-primary transition-colors disabled:opacity-50"
            >
              {checkMutation.isPending ? (
                <Loader2 className="h-3 w-3 animate-spin" />
              ) : (
                <AlertTriangle className="h-3 w-3" />
              )}
              {checkMutation.isPending ? 'Checking…' : 'Check model'}
            </button>

            {checkMutation.isSuccess && (
              <span className={cn(
                'inline-flex items-center gap-1 text-xs',
                checkMutation.data.available ? 'text-green-500' : checkMutation.data.deprecated ? 'text-red-500' : 'text-yellow-500'
              )}>
                {checkMutation.data.available ? (
                  <><CheckCircle className="h-3 w-3" /> Available {checkMutation.data.latency_ms ? `(${checkMutation.data.latency_ms}ms)` : ''}</>
                ) : checkMutation.data.deprecated ? (
                  <><XCircle className="h-3 w-3" /> Deprecated</>
                ) : (
                  <><AlertTriangle className="h-3 w-3" /> {checkMutation.data.message || 'Unavailable'}</>
                )}
              </span>
            )}

            {checkMutation.isError && (
              <span className="inline-flex items-center gap-1 text-xs text-red-500">
                <XCircle className="h-3 w-3" /> Check failed
              </span>
            )}
          </div>
        );
      })()}
    </div>
  );
}
