import { useAuthStore } from '@/stores/authStore';
import { useQueryClient } from '@tanstack/react-query';
import { API_BASE_URL } from '@/lib/constants';
import { tokenVault } from '@/utils/token-vault';
import { useCallback, useEffect, useRef } from 'react';

export interface ExamStatusUpdate {
  examId: string;
  status: string;
}

function wsBaseFromApiUrl(): string {
  const base =
    API_BASE_URL.startsWith('http://') || API_BASE_URL.startsWith('https://')
      ? API_BASE_URL
      : `${typeof window !== 'undefined' ? window.location.origin : ''}${API_BASE_URL}`;
  return base.replace(/^http/, 'ws').replace(/\/$/, '');
}

export function useExamStatusStream(examId: string | undefined) {
  const queryClient = useQueryClient();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectAttempts = useRef(0);
  const examIdRef = useRef(examId);
  examIdRef.current = examId;

  const readStream = useCallback(async () => {
    if (!isAuthenticated || !examIdRef.current) return;

    await tokenVault.initialize();
    const token = await tokenVault.getAccessToken();
    if (!token?.trim()) return;

    const connect = () => {
      const url = new URL(`${wsBaseFromApiUrl()}/v1/notifications/stream`);
      url.searchParams.set('token', token);

      const ws = new WebSocket(url.toString());
      wsRef.current = ws;

      ws.onopen = () => {
        reconnectAttempts.current = 0;
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data as string) as Record<string, unknown>;
          if (data.type !== 'cert_exam_status') return;
          const payload = data.payload as Record<string, string> | undefined;
          if (!payload) return;
          if (payload.exam_id !== examIdRef.current) return;

          queryClient.invalidateQueries({ queryKey: ['certification', 'exam', examIdRef.current] });
          queryClient.invalidateQueries({ queryKey: ['certification', 'exams'] });
        } catch {
          /* ignore malformed frames */
        }
      };

      ws.onclose = () => {
        wsRef.current = null;
        if (!isAuthenticated || !examIdRef.current) return;
        if (reconnectAttempts.current >= 5) return;
        reconnectAttempts.current += 1;
        const delay = Math.min(1000 * 2 ** reconnectAttempts.current, 30000);
        window.setTimeout(connect, delay);
      };
    };

    connect();
  }, [isAuthenticated, queryClient]);

  useEffect(() => {
    if (!isAuthenticated || !examId) return;
    void readStream();
    return () => {
      wsRef.current?.close();
      wsRef.current = null;
      reconnectAttempts.current = 0;
    };
  }, [isAuthenticated, examId, readStream]);

  return {};
}
