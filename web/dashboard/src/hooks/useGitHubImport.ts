import { useEffect, useState, useRef, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { githubApi } from '@/api/github';
import { githubKeys } from '@/hooks/useGitHubConnection';
import type {
  ListImportsParams,
  StartImportRequest,
  BulkImportRequest,
  BulkImportResult,
  ImportProgressEvent,
  ImportCompleteEvent,
  ImportErrorEvent,
} from '@/types/github';

export function useGitHubImports(params?: ListImportsParams) {
  return useQuery({
    queryKey: githubKeys.imports(params as Record<string, unknown>),
    queryFn: () => githubApi.listImports(params),
    refetchInterval: (query) => {
      const hasInProgress = query.state.data?.imports.some(
        (imp) =>
          imp.status === 'pending' ||
          imp.status === 'scanning' ||
          imp.status === 'configuring' ||
          imp.status === 'fetching' ||
          imp.status === 'building' ||
          imp.status === 'publishing'
      );
      return hasInProgress ? 3000 : false;
    },
    staleTime: 0,
  });
}

export function useGitHubImport(importId: string) {
  return useQuery({
    queryKey: githubKeys.importItem(importId),
    queryFn: () => githubApi.getImport(importId),
    enabled: !!importId,
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      return status === 'pending' ||
        status === 'scanning' ||
        status === 'configuring' ||
        status === 'fetching' ||
        status === 'building' ||
        status === 'publishing'
        ? 2000
        : false;
    },
    staleTime: 1000 * 15,
  });
}

export function useStartImport() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: StartImportRequest) => githubApi.startImport(data),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['github', 'imports'] });
      toast.success(`Import started: ${data.function_name}`);
    },
    onError: (error: Error) => {
      toast.error(`Failed to start import: ${error.message}`);
    },
  });
}

export function usePreviewImport() {
  return useMutation({
    mutationFn: (data: StartImportRequest) => githubApi.previewImport(data),
    onError: (error: Error) => {
      console.error('[usePreviewImport] Error details:', {
        message: error.message,
        cause: (error as unknown as { cause?: unknown }).cause,
        stack: error.stack,
      });
      toast.error(`Preview failed: ${error.message}`);
    },
  });
}

export function useBulkImport() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: BulkImportRequest) => githubApi.bulkImport(data),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['github', 'imports'] });
      toast.success(`Started ${data.length} imports`);
    },
    onError: (error: Error) => {
      toast.error(`Bulk import failed: ${error.message}`);
    },
  });
}

export function useCancelImport() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (importId: string) => githubApi.cancelImport(importId),
    onSuccess: (_data, importId) => {
      queryClient.invalidateQueries({ queryKey: ['github', 'imports'] });
      queryClient.invalidateQueries({ queryKey: githubKeys.importItem(importId) });
      toast.success('Import cancelled');
    },
    onError: (error: Error) => {
      toast.error(`Failed to cancel import: ${error.message}`);
    },
  });
}

export function useRetryImport() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (importId: string) => githubApi.retryImport(importId),
    onSuccess: (_data, importId) => {
      queryClient.invalidateQueries({ queryKey: ['github', 'imports'] });
      queryClient.invalidateQueries({ queryKey: githubKeys.importItem(importId) });
      toast.success('Import restarted');
    },
    onError: (error: Error) => {
      toast.error(`Failed to retry import: ${error.message}`);
    },
  });
}

export function useResyncImport() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (importId: string) => githubApi.resyncImport(importId),
    onSuccess: (_data, importId) => {
      queryClient.invalidateQueries({ queryKey: ['github', 'imports'] });
      queryClient.invalidateQueries({ queryKey: githubKeys.importItem(importId) });
      toast.success('Resync started');
    },
    onError: (error: Error) => {
      toast.error(`Failed to resync: ${error.message}`);
    },
  });
}

export function useImportProgress(importId: string | null) {
  const [progress, setProgress] = useState<ImportProgressEvent | null>(null);
  const [complete, setComplete] = useState<ImportCompleteEvent | null>(null);
  const [error, setError] = useState<ImportErrorEvent | null>(null);
  const [status, setStatus] = useState<'idle' | 'connecting' | 'streaming' | 'completed' | 'failed'>('idle');
  const eventSourceRef = useRef<EventSource | null>(null);
  const retryCountRef = useRef(0);
  const maxRetries = 5;
  const retryTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const cleanup = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
    if (retryTimeoutRef.current) {
      clearTimeout(retryTimeoutRef.current);
      retryTimeoutRef.current = null;
    }
  }, []);

  const connect = useCallback(() => {
    if (!importId) return;

    const token = localStorage.getItem('ff-access-token');
    if (!token) {
      setStatus('idle');
      return;
    }

    setStatus('connecting');

    const eventSource = githubApi.streamImportProgress(
      importId,
      (event) => {
        setProgress(event);
        setStatus('streaming');
        retryCountRef.current = 0;
      },
      (event) => {
        setComplete(event);
        setStatus('completed');
        cleanup();
      },
      (event) => {
        setError(event);
        setStatus('failed');
        cleanup();
      }
    );

    eventSourceRef.current = eventSource;

    eventSource.onopen = () => {
      setStatus('streaming');
      retryCountRef.current = 0;
    };

    eventSource.onerror = () => {
      if (status === 'completed' || status === 'failed') {
        return;
      }
      eventSource.close();
      eventSourceRef.current = null;

      if (retryCountRef.current < maxRetries) {
        retryCountRef.current += 1;
        const delay = Math.min(1000 * Math.pow(2, retryCountRef.current - 1), 10000);
        retryTimeoutRef.current = setTimeout(() => {
          connect();
        }, delay);
      } else {
        setError({
          stage: 'error',
          progress: progress?.progress ?? 0,
          message: 'Connection lost. Please refresh to retry.',
        });
        setStatus('failed');
      }
    };
  }, [importId, cleanup]);

  useEffect(() => {
    if (!importId) {
      setStatus('idle');
      setProgress(null);
      setComplete(null);
      setError(null);
      retryCountRef.current = 0;
      return;
    }

    setProgress(null);
    setComplete(null);
    setError(null);
    retryCountRef.current = 0;
    connect();

    return cleanup;
  }, [importId, connect, cleanup]);

  return { progress, complete, error, status };
}
