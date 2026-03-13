import { create } from "zustand";

/**
 * Tracks whether the backend API is reachable. Used to avoid firing optional
 * requests (e.g. follow/me/stats, api-keys, users/me) when the API is down,
 * so the console stays clean (no 404s).
 * - undefined: not yet known (will run health check before optional requests)
 * - true: API responded successfully (optional requests allowed)
 * - false: API unreachable or returned 404/5xx (optional requests skipped)
 */
interface ApiReachableState {
  apiReachable: boolean | undefined;
  setApiReachable: (value: boolean) => void;
}

export const useApiReachableStore = create<ApiReachableState>((set) => ({
  apiReachable: undefined,
  setApiReachable: (value) => set({ apiReachable: value }),
}));
