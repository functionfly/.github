import { useCallback } from 'react';
import { useCodePasteStore } from '../stores/codePasteStore';
import { functionsApi } from '../api/functions';
import type { ParsedFunction } from '../types/codePaste';

export function useCodeAnalyzer() {
  const {
    code,
    language,
    confidence,
    functions,
    status,
    error,
    setCode,
    setLanguage,
    setFunctions,
    setStatus,
    setError,
    reset,
  } = useCodePasteStore();

  const parseCode = useCallback(async (codeToParse?: string) => {
    const actualCode = codeToParse ?? code;

    if (!actualCode || actualCode.trim().length === 0) {
      setError('Please enter some code to parse');
      return;
    }

    setStatus('parsing');
    setError(null);

    try {
      const result = await functionsApi.parseCode(
        actualCode,
        language && language !== 'auto' ? language : undefined
      );

      if (result.functions && result.functions.length > 0) {
        const parsedFunctions: ParsedFunction[] = result.functions.map((f) => ({
          id: f.id,
          name: f.name,
          language: f.language,
          signature: f.signature,
          parameters: f.parameters || [],
          return_type: f.return_type,
          docstring: f.docstring,
          code: f.code,
          start_line: f.start_line,
          end_line: f.end_line,
        }));

        setFunctions(parsedFunctions);

        if (result.language) {
          setLanguage(result.language, result.confidence);
        }
      } else {
        setError('No functions detected in the code. Make sure your code contains valid function definitions.');
        setFunctions([]);
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to parse code';
      setError(errorMessage);
      console.error('Code parsing error:', err);
    }
  }, [code, language, setFunctions, setLanguage, setStatus, setError]);

  const clearAnalysis = useCallback(() => {
    reset();
  }, [reset]);

  return {
    code,
    language,
    confidence,
    functions,
    status,
    error,
    setCode,
    setLanguage,
    parseCode,
    clearAnalysis,
  };
}