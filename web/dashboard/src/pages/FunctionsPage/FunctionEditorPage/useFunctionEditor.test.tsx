import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] ?? null,
    setItem: (key: string, value: string) => { store[key] = value; },
    removeItem: (key: string) => { delete store[key]; },
    clear: () => { store = {}; },
  };
})();

Object.defineProperty(window, 'localStorage', { value: localStorageMock });

const mockNavigate = vi.fn();
const mockParams = { id: undefined as string | undefined };

vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
  useParams: () => mockParams,
}));

vi.mock('@/hooks/useVault', () => ({
  useVaultSecrets: () => ({ data: [] }),
}));

vi.mock('@/api', () => ({
  functionsApi: {
    get: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    deploy: vi.fn(),
    test: vi.fn(),
  },
}));

vi.mock('@/api/apps', () => ({
  fetchDeployBackendOptions: vi.fn().mockResolvedValue([]),
}));

vi.mock('@/api/providers', () => ({
  providersApi: {
    getConnectedProviders: vi.fn().mockResolvedValue([]),
  },
}));

vi.mock('@/api/vault', () => ({
  vaultApi: {
    decryptSecret: vi.fn(),
  },
}));

vi.mock('sonner', () => ({
  toast: {
    error: vi.fn(),
    success: vi.fn(),
    info: vi.fn(),
  },
}));

vi.mock('@tanstack/react-query', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@tanstack/react-query')>();
  return {
    ...actual,
    useQuery: ({ queryFn }: any) => {
      let data: any;
      if (queryFn) {
        const result = queryFn();
        if (result && typeof result.then === 'function') {
          data = undefined;
        } else {
          data = result;
        }
      }
      return { data, isLoading: false };
    },
  };
});

import { useFunctionEditor } from './useFunctionEditor';

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  return function Wrapper({ children }: { children: any }) {
    return (
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    );
  };
};

describe('useFunctionEditor', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorageMock.clear();
    mockParams.id = undefined;
    mockNavigate.mockClear();
  });

  afterEach(() => {
    window.localStorage.clear();
  });

  describe('draft persistence', () => {
    it('saves draft to localStorage when dirty', async () => {
      const { result } = renderHook(() => useFunctionEditor(), {
        wrapper: createWrapper(),
      });

      await act(async () => {
        result.current.setFunctionName('test-fn');
        result.current.markDirty();
      });

      await act(async () => {
        await new Promise((r) => setTimeout(r, 1600));
      });

      const saved = localStorageMock.getItem('functionfly:new-function-draft');
      expect(saved).toBeTruthy();
      const parsed = JSON.parse(saved as string);
      expect(parsed.functionName).toBe('test-fn');
    });

    it('restores draft from localStorage on new page load', async () => {
      const draft = {
        functionName: 'restored-fn',
        slug: 'restored-fn',
        description: 'restored',
        runtime: 'typescript',
        runtimeVersion: 'ES2022',
        code: 'code',
        visibility: 'private',
        tags: [] as string[],
        envVars: [] as any[],
        resources: { memoryMb: 128, timeoutMs: 30000, maxConcurrency: 10 },
        httpTrigger: { enabled: true, method: 'ANY', path: '/' },
        scheduleTrigger: { enabled: false, cron: '0 * * * *', timezone: 'UTC' },
        retryPolicy: { maxRetries: 3, backoffMs: 1000, backoffStrategy: 'exponential' },
        warmInstances: 0,
        savedAt: new Date().toISOString(),
      };
      localStorageMock.setItem('functionfly:new-function-draft', JSON.stringify(draft));

      const { result } = renderHook(() => useFunctionEditor(), {
        wrapper: createWrapper(),
      });

      await act(async () => {
        await new Promise((r) => setTimeout(r, 100));
      });

      expect(result.current.showDraftRestorePrompt).toBe(true);
    });

    it('discards draft and removes from localStorage', async () => {
      const draft = {
        functionName: 'discard-fn',
        slug: 'discard-fn',
        description: '',
        runtime: 'typescript',
        runtimeVersion: 'ES2022',
        code: '',
        visibility: 'private',
        tags: [] as string[],
        envVars: [] as any[],
        resources: { memoryMb: 128, timeoutMs: 30000, maxConcurrency: 10 },
        httpTrigger: { enabled: true, method: 'ANY', path: '/' },
        scheduleTrigger: { enabled: false, cron: '0 * * * *', timezone: 'UTC' },
        retryPolicy: { maxRetries: 3, backoffMs: 1000, backoffStrategy: 'exponential' },
        warmInstances: 0,
        savedAt: new Date().toISOString(),
      };
      localStorageMock.setItem('functionfly:new-function-draft', JSON.stringify(draft));

      const { result } = renderHook(() => useFunctionEditor(), {
        wrapper: createWrapper(),
      });

      await act(async () => {
        await new Promise((r) => setTimeout(r, 100));
      });

      await act(async () => {
        result.current.handleDiscardDraft();
      });

      expect(localStorageMock.getItem('functionfly:new-function-draft')).toBeNull();
      expect(result.current.showDraftRestorePrompt).toBe(false);
    });
  });

  describe('layout persistence', () => {
    it('restores activeTab from sessionStorage', async () => {
      sessionStorage.setItem('functionfly:editor-layout', JSON.stringify({ activeTab: 'logs' }));

      const { result } = renderHook(() => useFunctionEditor(), {
        wrapper: createWrapper(),
      });

      await act(async () => {
        await new Promise((r) => setTimeout(r, 100));
      });

      expect(result.current.activeTab).toBe('logs');
    });

    it('persists activeTab to sessionStorage on change', async () => {
      const { result } = renderHook(() => useFunctionEditor(), {
        wrapper: createWrapper(),
      });

      await act(async () => {
        result.current.setActiveTab('logs');
      });

      const saved = sessionStorage.getItem('functionfly:editor-layout');
      expect(saved).toBeTruthy();
      expect(JSON.parse(saved as string).activeTab).toBe('logs');
    });
  });
});
