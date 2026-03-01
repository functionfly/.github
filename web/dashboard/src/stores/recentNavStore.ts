import { create } from "zustand";
import { persist } from "zustand/middleware";
import { getCanonicalNavPath } from "@/lib/constants";

const MAX_RECENT = 5;
const STORAGE_KEY = "functionfly-recent-paths";

interface RecentNavState {
  recentPaths: string[];
  record: (pathname: string) => void;
  getRecentPaths: () => string[];
}

export const useRecentNavStore = create<RecentNavState>()(
  persist(
    (set, get) => ({
      recentPaths: [],

      record(pathname: string) {
        const canonical = getCanonicalNavPath(pathname);
        if (!canonical) return;

        set((state) => {
          const next = [
            canonical,
            ...state.recentPaths.filter((p) => p !== canonical),
          ].slice(0, MAX_RECENT);
          return { recentPaths: next };
        });
      },

      getRecentPaths() {
        return get().recentPaths;
      },
    }),
    { name: STORAGE_KEY }
  )
);
