import { openDB, type IDBPDatabase } from 'idb';

const DB_NAME = 'functionfly-offline';
const DB_VERSION = 1;

export interface PendingMessage {
  id: string;
  conversationId: string;
  content: string;
  timestamp: number;
  retries: number;
  lastError?: string;
}

export interface CachedMessage {
  id: string;
  conversationId: string;
  content: string;
  authorId: string;
  createdAt: string;
  editedAt?: string;
  deletedAt?: string;
  embeddings: Record<string, unknown>;
  reactions?: { reaction: string; count: number; user_ids: string[] }[];
  attachments?: unknown[];
  synced: boolean;
}

export interface OfflineQueueItem {
  id: string;
  type: 'create_message' | 'edit_message' | 'delete_message' | 'add_reaction' | 'remove_reaction';
  payload: Record<string, unknown>;
  timestamp: number;
  retries: number;
  lastError?: string;
}

class OfflineStore {
  private db: IDBPDatabase | null = null;
  private dbPromise: Promise<IDBPDatabase> | null = null;

  async init(): Promise<IDBPDatabase> {
    if (this.db) return this.db;
    if (this.dbPromise) return this.dbPromise;

    this.dbPromise = openDB(DB_NAME, DB_VERSION, {
      upgrade(db) {
        if (!db.objectStoreNames.contains('messages')) {
          const messagesStore = db.createObjectStore('messages', { keyPath: 'id' });
          messagesStore.createIndex('conversationId', 'conversationId');
          messagesStore.createIndex('synced', 'synced');
        }

        if (!db.objectStoreNames.contains('queue')) {
          const queueStore = db.createObjectStore('queue', { keyPath: 'id' });
          queueStore.createIndex('timestamp', 'timestamp');
        }

        if (!db.objectStoreNames.contains('conversations')) {
          db.createObjectStore('conversations', { keyPath: 'id' });
        }
      },
    });

    this.db = await this.dbPromise;
    return this.db;
  }

  async getMessages(conversationId: string): Promise<CachedMessage[]> {
    const db = await this.init();
    const index = db.transaction('messages').store.index('conversationId');
    const messages = await index.getAll(conversationId);
    return messages.filter((m) => !m.deletedAt);
  }

  async getPendingMessages(): Promise<OfflineQueueItem[]> {
    const db = await this.init();
    const queue = db.transaction('queue').store;
    const items = await queue.getAll();
    return items as OfflineQueueItem[];
  }

  async addToQueue(item: OfflineQueueItem): Promise<void> {
    const db = await this.init();
    await db.put('queue', item);
  }

  async removeFromQueue(id: string): Promise<void> {
    const db = await this.init();
    await db.delete('queue', id);
  }

  async updateQueueItem(item: OfflineQueueItem): Promise<void> {
    const db = await this.init();
    await db.put('queue', item);
  }

  async cacheMessage(message: CachedMessage): Promise<void> {
    const db = await this.init();
    await db.put('messages', { ...message, synced: true });
  }

  async cacheMessages(messages: CachedMessage[]): Promise<void> {
    const db = await this.init();
    const tx = db.transaction('messages', 'readwrite');
    await Promise.all(messages.map((m) => tx.store.put({ ...m, synced: true })));
    await tx.done;
  }

  async getUnsyncedMessages(): Promise<CachedMessage[]> {
    const db = await this.init();
    const index = db.transaction('messages').store.index('synced');
    const messages = await index.getAll(IDBKeyRange.bound([false], [false]));
    return messages;
  }

  async markMessageSynced(id: string): Promise<void> {
    const db = await this.init();
    const message = await db.get('messages', id);
    if (message) {
      await db.put('messages', { ...message, synced: true });
    }
  }

  async getCachedConversations(): Promise<unknown[]> {
    const db = await this.init();
    return db.getAll('conversations');
  }

  async cacheConversation(conversation: unknown): Promise<void> {
    const db = await this.init();
    await db.put('conversations', conversation);
  }

  async clearAll(): Promise<void> {
    const db = await this.init();
    await db.clear('messages');
    await db.clear('queue');
    await db.clear('conversations');
  }
}

export const offlineStore = new OfflineStore();