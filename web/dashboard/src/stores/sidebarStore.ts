import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';

interface SidebarState {
  // Collapsed state
  isCollapsed: boolean;
  toggleCollapsed: () => void;
  setCollapsed: (value: boolean) => void;

  // Section expansion state (persisted)
  expandedSections: Set<string>;
  toggleSection: (sectionId: string) => void;
  setSectionExpanded: (sectionId: string, expanded: boolean) => void;
  isSectionExpanded: (sectionId: string) => boolean;

  // Section order (for drag-to-reorder)
  sectionOrder: string[];
  moveSection: (fromIndex: number, toIndex: number) => void;
  setSectionOrder: (order: string[]) => void;

  // Favorites
  favorites: string[]; // Array of paths
  addFavorite: (path: string) => void;
  removeFavorite: (path: string) => void;
  isFavorite: (path: string) => boolean;
  toggleFavorite: (path: string) => void;

  // Workspace/Environment
  currentWorkspace: string;
  currentEnvironment: 'production' | 'staging' | 'development';
  setWorkspace: (workspace: string) => void;
  setEnvironment: (env: 'production' | 'staging' | 'development') => void;

  // Onboarding progress
  showOnboardingHints: boolean;
  completedOnboardingSteps: string[];
  markOnboardingStepCompleted: (step: string) => void;
  toggleOnboardingHints: () => void;
}

const STORAGE_KEY = 'sidebar-preferences-v1';
const DEFAULT_SECTIONS = ['discover', 'build', 'deploy', 'operate', 'advanced', 'account'];

export const useSidebarStore = create<SidebarState>()(
  persist(
    immer((set, get) => ({
      // Collapsed state
      isCollapsed: false,
      toggleCollapsed: () =>
        set((state) => {
          state.isCollapsed = !state.isCollapsed;
        }),
      setCollapsed: (value) =>
        set((state) => {
          state.isCollapsed = value;
        }),

      // Section expansion
      expandedSections: new Set(DEFAULT_SECTIONS),
      toggleSection: (sectionId) =>
        set((state) => {
          if (state.expandedSections.has(sectionId)) {
            state.expandedSections.delete(sectionId);
          } else {
            state.expandedSections.add(sectionId);
          }
        }),
      setSectionExpanded: (sectionId, expanded) =>
        set((state) => {
          if (expanded) {
            state.expandedSections.add(sectionId);
          } else {
            state.expandedSections.delete(sectionId);
          }
        }),
      isSectionExpanded: (sectionId) => {
        return get().expandedSections.has(sectionId);
      },

      // Section order
      sectionOrder: DEFAULT_SECTIONS,
      moveSection: (fromIndex, toIndex) =>
        set((state) => {
          const [moved] = state.sectionOrder.splice(fromIndex, 1);
          state.sectionOrder.splice(toIndex, 0, moved);
        }),
      setSectionOrder: (order) =>
        set((state) => {
          state.sectionOrder = order;
        }),

      // Favorites
      favorites: [],
      addFavorite: (path) =>
        set((state) => {
          if (!state.favorites.includes(path)) {
            state.favorites.unshift(path);
            // Keep max 8 favorites
            if (state.favorites.length > 8) {
              state.favorites.pop();
            }
          }
        }),
      removeFavorite: (path) =>
        set((state) => {
          state.favorites = state.favorites.filter((p) => p !== path);
        }),
      isFavorite: (path) => {
        return get().favorites.includes(path);
      },
      toggleFavorite: (path) => {
        const { isFavorite, addFavorite, removeFavorite } = get();
        if (isFavorite(path)) {
          removeFavorite(path);
        } else {
          addFavorite(path);
        }
      },

      // Workspace/Environment
      currentWorkspace: 'personal',
      currentEnvironment: 'production',
      setWorkspace: (workspace) =>
        set((state) => {
          state.currentWorkspace = workspace;
        }),
      setEnvironment: (env) =>
        set((state) => {
          state.currentEnvironment = env;
        }),

      // Onboarding
      showOnboardingHints: true,
      completedOnboardingSteps: [],
      markOnboardingStepCompleted: (step) =>
        set((state) => {
          if (!state.completedOnboardingSteps.includes(step)) {
            state.completedOnboardingSteps.push(step);
          }
        }),
      toggleOnboardingHints: () =>
        set((state) => {
          state.showOnboardingHints = !state.showOnboardingHints;
        }),
    })),
    {
      name: STORAGE_KEY,
      partialize: (state) => ({
        isCollapsed: state.isCollapsed,
        expandedSections: Array.from(state.expandedSections),
        sectionOrder: state.sectionOrder,
        favorites: state.favorites,
        currentWorkspace: state.currentWorkspace,
        currentEnvironment: state.currentEnvironment,
        showOnboardingHints: state.showOnboardingHints,
        completedOnboardingSteps: state.completedOnboardingSteps,
      }),
      merge: (persistedState, currentState) => {
        // Restore Set from array
        const state = { ...currentState, ...persistedState } as SidebarState;
        if (Array.isArray(persistedState.expandedSections)) {
          state.expandedSections = new Set(persistedState.expandedSections);
        }
        return state;
      },
    }
  )
);
