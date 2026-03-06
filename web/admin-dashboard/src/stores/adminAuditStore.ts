/**
 * Admin Audit Store
 * Tracks local audit events for display
 */

import { create } from 'zustand';
import type { AuditEvent } from '@/types';

interface AdminAuditState {
  events: AuditEvent[];
  addEvent: (event: AuditEvent) => void;
  clearEvents: () => void;
  getRecentEvents: (limit?: number) => AuditEvent[];
}

export const useAdminAuditStore = create<AdminAuditState>((set, get) => ({
  events: [],

  addEvent: (event: AuditEvent) => {
    const state = get();
    // Keep only last 100 events in memory
    const newEvents = [event, ...state.events].slice(0, 100);
    set({ events: newEvents });
  },

  clearEvents: () => {
    set({ events: [] });
  },

  getRecentEvents: (limit = 10) => {
    return get().events.slice(0, limit);
  },
}));
