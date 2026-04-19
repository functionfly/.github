/**
 * Blog Settings Store
 * Persists blog settings to localStorage with a Zustand-like API.
 * Ready to swap localStorage for API calls once backend supports PATCH /content/blog/settings.
 */

import { CACHE_KEYS } from '@/lib/constants';

export interface BlogSettings {
  blogTitle: string;
  postsPerPage: number;
  metaDescription: string;
}

const DEFAULT_SETTINGS: BlogSettings = {
  blogTitle: 'FunctionFly Blog',
  postsPerPage: 10,
  metaDescription: '',
};

function loadFromStorage(): BlogSettings {
  try {
    const raw = localStorage.getItem(CACHE_KEYS.BLOG_SETTINGS);
    if (!raw) return DEFAULT_SETTINGS;
    const parsed = JSON.parse(raw) as Partial<BlogSettings>;
    return {
      blogTitle: parsed.blogTitle ?? DEFAULT_SETTINGS.blogTitle,
      postsPerPage: parsed.postsPerPage ?? DEFAULT_SETTINGS.postsPerPage,
      metaDescription: parsed.metaDescription ?? DEFAULT_SETTINGS.metaDescription,
    };
  } catch {
    return DEFAULT_SETTINGS;
  }
}

function saveToStorage(settings: BlogSettings): void {
  try {
    localStorage.setItem(CACHE_KEYS.BLOG_SETTINGS, JSON.stringify(settings));
  } catch {
    /* localStorage unavailable — silently ignore */
  }
}

// In-memory state for reactivity (the component manages its own useState,
// this module just provides the localStorage facade)
export const blogSettingsStore = {
  load: loadFromStorage,
  save: saveToStorage,
  default: DEFAULT_SETTINGS,
};
