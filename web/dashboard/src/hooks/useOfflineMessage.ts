import { useMutation, useQueryClient } from '@tanstack/react-query';
import { conversationsApi } from '@/api/conversations';
import { offlineStore } from '@/lib/offline-store';
import { offlineSyncQueue } from '@/lib/offline-sync';
import { conversationKeys } from './useConversations';
import { useAuthStore } from '@/stores/authStore';
import { toast } from 'sonner';
import { useCallback, useEffect, useState } from 'react';

export interface PendingMessageIndicator {
  id: string;
  conversationId: string;
  content: string;
  timestamp: number;
  status: 'pending' | 'sending' | 'failed';
  error?: string;
}

export function useOfflineMessage() {
  const queryClient = useQueryClient();
  const currentUser = useAuthStore((state) => state.user);
  const [pendingMessages, setPendingMessages] = useState<PendingMessageIndicator[]>([]);

  useEffect(() => {
    if (currentUser?.username) {
      offlineSyncQueue.setUsername(currentUser.username);
    }
  }, [currentUser?.username]);

  useEffect(() => {
    const loadPending = async () => {
      const items = await offlineStore.getPendingMessages();
      const pending = items
        .filter((item) => (item as unknown as { type?: string }).type === 'create_message')
        .map((item) => {
          const i = item as unknown as { id: string; conversationId: string; content: string; timestamp: number; lastError?: string };
          return {
            id: i.id,
            conversationId: i.conversationId,
            content: i.content,
            timestamp: i.timestamp,
            status: 'pending' as const,
            error: i.lastError,
          };
        });
      setPendingMessages(pending);
    };

    loadPending();

    const interval = setInterval(loadPending, 5000);
    return () => clearInterval(interval);
  }, []);

  const sendMessageOptimistic = useCallback(
    async (conversationId: string, content: string): Promise<string> => {
      const localId = crypto.randomUUID();

      const optimisticMessage = {
        id: localId,
        conversationId,
        content,
        authorId: currentUser?.id || '',
        createdAt: new Date().toISOString(),
        embeddings: {},
        synced: false,
      };

      await offlineStore.cacheMessage(optimisticMessage);

      setPendingMessages((prev) => [
        ...prev,
        {
          id: localId,
          conversationId,
          content,
          timestamp: Date.now(),
          status: 'pending',
        },
      ]);

      try {
        setPendingMessages((prev) =>
          prev.map((m) => (m.id === localId ? { ...m, status: 'sending' } : m))
        );

        await offlineSyncQueue.enqueueMessage('create_message', {
          conversationId,
          content,
          localId,
        });

        queryClient.invalidateQueries({ queryKey: conversationKeys.messages(conversationId) });

        return localId;
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Failed to send';
        setPendingMessages((prev) =>
          prev.map((m) => (m.id === localId ? { ...m, status: 'failed', error: errorMessage } : m))
        );
        throw error;
      }
    },
    [currentUser?.id, queryClient]
  );

  const retryMessage = useCallback(async (localId: string) => {
    setPendingMessages((prev) =>
      prev.map((m) => (m.id === localId ? { ...m, status: 'pending', error: undefined } : m))
    );

    const items = await offlineStore.getPendingMessages();
    const item = items.find((i: { id: string }) => i.id === localId);

    if (item) {
      await offlineSyncQueue.enqueueMessage(
        item.type,
        item.payload
      );
    }
  }, []);

  const dismissFailedMessage = useCallback(async (localId: string) => {
    setPendingMessages((prev) => prev.filter((m) => m.id !== localId));
  }, []);

  return {
    pendingMessages,
    sendMessageOptimistic,
    retryMessage,
    dismissFailedMessage,
    isOnline: typeof navigator !== 'undefined' ? navigator.onLine : true,
  };
}

export function useOfflineStatus() {
  const [isOnline, setIsOnline] = useState(
    typeof navigator !== 'undefined' ? navigator.onLine : true
  );

  useEffect(() => {
    const handleOnline = () => setIsOnline(true);
    const handleOffline = () => setIsOnline(false);

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, []);

  return isOnline;
}