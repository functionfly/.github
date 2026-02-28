import { useMemo } from 'react';
import { usePlaygroundStore } from '../store/playgroundStore';

/**
 * Derived state helpers for the playground.
 * Computes commonly needed derived values from the store.
 */
export function usePlaygroundState() {
  const {
    executionResult,
    executionHistory,
    isExecuting,
    inputValue,
    inputJson,
    functionInfo,
  } = usePlaygroundStore();

  const hasResult = executionResult !== null;
  const isSuccess = executionResult?.ok === true;
  const isError = executionResult?.ok === false;
  const isCached = executionResult?.cached === true;

  const latencyHistory = useMemo(
    () =>
      executionHistory
        .slice(0, 10)
        .reverse()
        .map((item, index) => ({
          index,
          duration_ms: item.result.duration_ms,
          ok: item.result.ok,
          timestamp: item.timestamp,
        })),
    [executionHistory]
  );

  const averageLatency = useMemo(() => {
    if (latencyHistory.length === 0) return null;
    const sum = latencyHistory.reduce((acc, item) => acc + item.duration_ms, 0);
    return Math.round(sum / latencyHistory.length);
  }, [latencyHistory]);

  const successRate = useMemo(() => {
    if (executionHistory.length === 0) return null;
    const successes = executionHistory.filter((h) => h.result.ok).length;
    return Math.round((successes / executionHistory.length) * 100);
  }, [executionHistory]);

  const shareableUrl = useMemo(() => {
    if (!functionInfo) return '';
    const serialized = JSON.stringify(inputValue);
    const encoded = btoa(serialized).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
    return `${window.location.origin}/run/${functionInfo.author}/${functionInfo.name}?input=${encoded}`;
  }, [functionInfo, inputValue]);

  const isInputValid = useMemo(() => {
    try {
      JSON.parse(inputJson);
      return true;
    } catch {
      return false;
    }
  }, [inputJson]);

  return {
    hasResult,
    isSuccess,
    isError,
    isCached,
    isExecuting,
    latencyHistory,
    averageLatency,
    successRate,
    shareableUrl,
    isInputValid,
  };
}
