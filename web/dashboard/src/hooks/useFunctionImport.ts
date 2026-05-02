import { useState, useCallback } from 'react';
import { functionsApi } from '../api/functions';
import type { ParsedFunction, ImportConfig } from '../types/codePaste';

interface UseFunctionImportResult {
  status: 'idle' | 'importing' | 'success' | 'error';
  createdFunctions: Array<{ id: string; name: string; status: string }>;
  errors: Array<{ name: string; error: string }>;
  importFunctions: (
    functions: ParsedFunction[],
    selectedIds: Set<string>,
    config: ImportConfig
  ) => Promise<boolean>;
  reset: () => void;
}

export function useFunctionImport(): UseFunctionImportResult {
  const [status, setStatus] = useState<'idle' | 'importing' | 'success' | 'error'>('idle');
  const [createdFunctions, setCreatedFunctions] = useState<Array<{ id: string; name: string; status: string }>>([]);
  const [errors, setErrors] = useState<Array<{ name: string; error: string }>>([]);

  const importFunctions = useCallback(async (
    functions: ParsedFunction[],
    selectedIds: Set<string>,
    config: ImportConfig
  ): Promise<boolean> => {
    const selectedFunctions = functions.filter((f) => selectedIds.has(f.id));

    if (selectedFunctions.length === 0) {
      setErrors([{ name: '', error: 'No functions selected' }]);
      setStatus('error');
      return false;
    }

    setStatus('importing');
    setErrors([]);

    try {
      const result = await functionsApi.createFromCode({
        functions: selectedFunctions.map((f) => ({
          name: f.name,
          code: f.code,
          language: f.language,
        })),
        visibility: config.visibility,
        providers: config.providers,
        region: config.region,
      });

      if (result.created && result.created.length > 0) {
        setCreatedFunctions(result.created);
        setStatus('success');
        return true;
      }

      if (result.failed && result.failed.length > 0) {
        setErrors(result.failed);
        setStatus('error');
        return false;
      }

      setStatus('idle');
      return false;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Import failed';
      setErrors([{ name: 'import', error: errorMessage }]);
      setStatus('error');
      return false;
    }
  }, []);

  const reset = useCallback(() => {
    setStatus('idle');
    setCreatedFunctions([]);
    setErrors([]);
  }, []);

  return {
    status,
    createdFunctions,
    errors,
    importFunctions,
    reset,
  };
}