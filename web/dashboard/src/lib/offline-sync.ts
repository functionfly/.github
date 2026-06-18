import { conversationsApi } from '@/api/conversations';
import { offlineStore, type OfflineQueueItem, type CachedMessage } from './offline-store';
import { toast } from 'sonner';

const MAX_RETRIES = 3;
const RETRY_DELAY = 5000;

class OfflineSyncQueue {
  private isProcessing = false;
  private isOnline = typeof navigator !== 'undefined' ? navigator.onLine : true;
  private username = '';
  private retryTimeouts = new Map<string, ReturnType<typeof setTimeout>>();

  constructor() {
    if (typeof window !== 'undefined') {
      window.addEventListener('online', () => this.handleOnline());
      window.addEventListener('offline', () => this.handleOffline());
    }
  }

  setUsername(username: string) {
    this.username = username;
  }

  private handleOnline() {
    this.isOnline = true;
    toast.success('Back online - syncing messages');
    this.processQueue();
  }

  private handleOffline() {
    this.isOnline = false;
    toast.error('You are offline - messages will be sent when you reconnect');
  }

  async enqueueMessage(
    type: OfflineQueueItem['type'],
    payload: Record<string, unknown>
  ): Promise<string> {
    const id = crypto.randomUUID();
    const item: OfflineQueueItem = {
      id,
      type,
      payload,
      timestamp: Date.now(),
      retries: 0,
    };

    await offlineStore.addToQueue(item);

    if (this.isOnline) {
      this.processQueue();
    }

    return id;
  }

  async processQueue(): Promise<void> {
    if (this.isProcessing || !this.isOnline) return;

    this.isProcessing = true;

    try {
      const pendingItems = await offlineStore.getPendingMessages();

      for (const rawItem of pendingItems) {
        const item = rawItem as unknown as OfflineQueueItem;
        if (this.retryTimeouts.has(item.id)) continue;

        try {
          await this.processItem(item);
          await offlineStore.removeFromQueue(item.id);
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Unknown error';
          item.retries++;
          item.lastError = errorMessage;

          if (item.retries >= MAX_RETRIES) {
            await offlineStore.removeFromQueue(item.id);
            toast.error(`Failed to sync: ${typeToHuman(item.type)}`);
          } else {
            await offlineStore.updateQueueItem(item);
            this.scheduleRetry(item.id, item.retries);
          }
        }
      }
    } finally {
      this.isProcessing = false;
    }
  }

  private async processItem(item: OfflineQueueItem): Promise<void> {
    switch (item.type) {
      case 'create_message':
        await this.processCreateMessage(item);
        break;
      case 'edit_message':
        await this.processEditMessage(item);
        break;
      case 'delete_message':
        await this.processDeleteMessage(item);
        break;
      case 'add_reaction':
        await this.processAddReaction(item);
        break;
      case 'remove_reaction':
        await this.processRemoveReaction(item);
        break;
    }
  }

  private async processCreateMessage(item: OfflineQueueItem): Promise<void> {
    const { conversationId, content, localId, embeddings } = item.payload as {
      conversationId: string;
      content: string;
      localId: string;
      embeddings?: Record<string, unknown>;
    };

    const message = await conversationsApi.createMessage(this.username, conversationId, {
      content,
      embeddings,
    });

    const cachedMessage: CachedMessage = {
      id: message.id,
      conversationId,
      content: message.content,
      authorId: message.author_id,
      createdAt: message.created_at,
      editedAt: message.edited_at ?? undefined,
      deletedAt: message.deleted_at ?? undefined,
      embeddings: message.embeddings as Record<string, unknown>,
      reactions: message.reactions,
      attachments: message.attachments,
      synced: true,
    };

    await offlineStore.cacheMessage(cachedMessage);

    toast.success('Message sent');
  }

  private async processEditMessage(item: OfflineQueueItem): Promise<void> {
    const { conversationId, messageId, content } = item.payload as {
      conversationId: string;
      messageId: string;
      content: string;
    };

    await conversationsApi.editMessage(this.username, conversationId, messageId, { content });
    await offlineStore.markMessageSynced(messageId);

    toast.success('Message updated');
  }

  private async processDeleteMessage(item: OfflineQueueItem): Promise<void> {
    const { conversationId, messageId } = item.payload as {
      conversationId: string;
      messageId: string;
    };

    await conversationsApi.deleteMessage(this.username, conversationId, messageId);

    toast.success('Message deleted');
  }

  private async processAddReaction(item: OfflineQueueItem): Promise<void> {
    const { conversationId, messageId, reaction } = item.payload as {
      conversationId: string;
      messageId: string;
      reaction: string;
    };

    await conversationsApi.addReaction(this.username, conversationId, messageId, reaction);
  }

  private async processRemoveReaction(item: OfflineQueueItem): Promise<void> {
    const { conversationId, messageId, reaction } = item.payload as {
      conversationId: string;
      messageId: string;
      reaction: string;
    };

    await conversationsApi.removeReaction(
      this.username,
      conversationId,
      messageId,
      reaction
    );
  }

  private scheduleRetry(itemId: string, retryCount: number): void {
    const delay = RETRY_DELAY * Math.pow(2, retryCount - 1);

    const timeout = setTimeout(async () => {
      this.retryTimeouts.delete(itemId);
      if (this.isOnline) {
        this.processQueue();
      }
    }, delay);

    this.retryTimeouts.set(itemId, timeout);
  }

  async getQueueSize(): Promise<number> {
    const items = await offlineStore.getPendingMessages();
    return items.length;
  }

  async clearQueue(): Promise<void> {
    for (const timeout of this.retryTimeouts.values()) {
      clearTimeout(timeout);
    }
    this.retryTimeouts.clear();

    const items = await offlineStore.getPendingMessages();
    for (const item of items) {
      await offlineStore.removeFromQueue(item.id);
    }
  }
}

function typeToHuman(type: OfflineQueueItem['type']): string {
  switch (type) {
    case 'create_message':
      return 'new message';
    case 'edit_message':
      return 'edited message';
    case 'delete_message':
      return 'deleted message';
    case 'add_reaction':
      return 'reaction';
    case 'remove_reaction':
      return 'reaction removal';
    default:
      return 'operation';
  }
}

export const offlineSyncQueue = new OfflineSyncQueue();